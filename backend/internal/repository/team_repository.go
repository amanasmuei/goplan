package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/goplan/backend/internal/models"
)

type TeamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create creates a new team and adds the creator as owner
func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	team.ID = uuid.New()
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()

	query := `
		INSERT INTO teams (id, name, description, organization_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = tx.Exec(ctx, query,
		team.ID, team.Name, team.Description, team.OrganizationID,
		team.CreatedBy, team.CreatedAt, team.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	// Add the creator as owner
	memberQuery := `
		INSERT INTO team_members (id, team_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, memberQuery,
		uuid.New(), team.ID, team.CreatedBy, models.TeamRoleOwner, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to add creator as owner: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a team by its ID
func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	query := `
		SELECT id, name, description, organization_id, created_by, created_at, updated_at
		FROM teams WHERE id = $1`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Description, &team.OrganizationID,
		&team.CreatedBy, &team.CreatedAt, &team.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

// Update updates an existing team
func (r *TeamRepository) Update(ctx context.Context, team *models.Team) error {
	team.UpdatedAt = time.Now()

	query := `
		UPDATE teams SET name = $2, description = $3, updated_at = $4
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query, team.ID, team.Name, team.Description, team.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}

// Delete deletes a team by ID
func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}

// ListByOrganization lists all teams for an organization
func (r *TeamRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]models.Team, error) {
	query := `
		SELECT id, name, description, organization_id, created_by, created_at, updated_at
		FROM teams WHERE organization_id = $1
		ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.OrganizationID,
			&team.CreatedBy, &team.CreatedAt, &team.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}
	return teams, nil
}

// AddMember adds a user to a team with the specified role
func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID uuid.UUID, role models.TeamRole) error {
	query := `
		INSERT INTO team_members (id, team_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = $4`

	_, err := r.db.Exec(ctx, query, uuid.New(), teamID, userID, role, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

// RemoveMember removes a user from a team
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	result, err := r.db.Exec(ctx, query, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found in team")
	}
	return nil
}

// UpdateMemberRole updates a member's role in a team
func (r *TeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, role models.TeamRole) error {
	query := `UPDATE team_members SET role = $3 WHERE team_id = $1 AND user_id = $2`
	result, err := r.db.Exec(ctx, query, teamID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found in team")
	}
	return nil
}

// GetMembers gets all members of a team with their user details
func (r *TeamRepository) GetMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.role, tm.joined_at,
			   u.id, u.email, u.name, u.role, u.organization_id, u.created_at, u.updated_at
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
		ORDER BY tm.role, u.name`

	rows, err := r.db.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var m models.TeamMember
		var u models.User
		err := rows.Scan(
			&m.ID, &m.TeamID, &m.UserID, &m.Role, &m.JoinedAt,
			&u.ID, &u.Email, &u.Name, &u.Role, &u.OrganizationID, &u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		m.User = &u
		members = append(members, m)
	}
	return members, nil
}

// GetMemberRole gets a user's role in a specific team
func (r *TeamRepository) GetMemberRole(ctx context.Context, teamID, userID uuid.UUID) (models.TeamRole, error) {
	query := `SELECT role FROM team_members WHERE team_id = $1 AND user_id = $2`
	var role models.TeamRole
	err := r.db.QueryRow(ctx, query, teamID, userID).Scan(&role)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("user is not a member of this team")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get member role: %w", err)
	}
	return role, nil
}

// IsUserInTeam checks if a user is a member of a team
func (r *TeamRepository) IsUserInTeam(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, teamID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team membership: %w", err)
	}
	return exists, nil
}

// CountMembers returns the number of members in a team
func (r *TeamRepository) CountMembers(ctx context.Context, teamID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM team_members WHERE team_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, teamID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count members: %w", err)
	}
	return count, nil
}

// GetUserTeams gets all teams a user belongs to
func (r *TeamRepository) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]models.Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.organization_id, t.created_by, t.created_at, t.updated_at
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1
		ORDER BY t.name ASC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.OrganizationID,
			&team.CreatedBy, &team.CreatedAt, &team.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}
	return teams, nil
}

// CountOwners returns the number of owners in a team
func (r *TeamRepository) CountOwners(ctx context.Context, teamID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND role = $2`
	var count int
	err := r.db.QueryRow(ctx, query, teamID, models.TeamRoleOwner).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count owners: %w", err)
	}
	return count, nil
}
