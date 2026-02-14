package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/goplan/backend/internal/models"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create creates a new project
func (r *ProjectRepository) Create(ctx context.Context, project *models.Project) error {
	project.ID = uuid.New()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	if project.Status == "" {
		project.Status = models.ProjectStatusActive
	}

	query := `
		INSERT INTO projects (id, name, description, status, organization_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		project.ID, project.Name, project.Description, project.Status,
		project.OrganizationID, project.CreatedBy, project.CreatedAt, project.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}

// GetByID retrieves a project by its ID
func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	query := `
		SELECT id, name, description, COALESCE(status, 'active'), organization_id, created_by, created_at, updated_at
		FROM projects WHERE id = $1`

	project := &models.Project{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.Status,
		&project.OrganizationID, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return project, nil
}

// Update updates an existing project
func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	project.UpdatedAt = time.Now()

	query := `
		UPDATE projects SET name = $2, description = $3, status = $4, updated_at = $5
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		project.ID, project.Name, project.Description, project.Status, project.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

// Delete soft-deletes a project by setting its status to archived.
func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE projects SET status = 'archived', updated_at = NOW() WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

// Archive sets a project's status to archived
func (r *ProjectRepository) Archive(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE projects SET status = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id, models.ProjectStatusArchived, time.Now())
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

// ListByOrganization lists projects for an organization with filters
func (r *ProjectRepository) ListByOrganization(ctx context.Context, filters models.ProjectFilters) ([]models.Project, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("p.organization_id = $%d", argIdx))
	args = append(args, filters.OrganizationID)
	argIdx++

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("p.status = $%d", argIdx))
		args = append(args, *filters.Status)
		argIdx++
	}

	if filters.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("EXISTS(SELECT 1 FROM project_teams pt WHERE pt.project_id = p.id AND pt.team_id = $%d)", argIdx))
		args = append(args, *filters.TeamID)
		argIdx++
	}

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM projects p WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Main query with pagination
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	offset := (filters.Page - 1) * filters.PageSize

	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, COALESCE(p.status, 'active'), p.organization_id, p.created_by, p.created_at, p.updated_at
		FROM projects p WHERE %s
		ORDER BY p.name ASC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, filters.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.Status,
			&project.OrganizationID, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, total, nil
}

// ListByTeam lists projects assigned to a team with pagination.
// limit defaults to 100, max 200. offset defaults to 0.
func (r *ProjectRepository) ListByTeam(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]models.Project, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT p.id, p.name, p.description, COALESCE(p.status, 'active'), p.organization_id, p.created_by, p.created_at, p.updated_at
		FROM projects p
		JOIN project_teams pt ON p.id = pt.project_id
		WHERE pt.team_id = $1
		ORDER BY p.name ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, teamID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list team projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.Status,
			&project.OrganizationID, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}
	return projects, nil
}

// AssignTeams assigns teams to a project
func (r *ProjectRepository) AssignTeams(ctx context.Context, projectID uuid.UUID, teamIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, teamID := range teamIDs {
		query := `
			INSERT INTO project_teams (id, project_id, team_id, assigned_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (project_id, team_id) DO NOTHING`

		_, err := tx.Exec(ctx, query, uuid.New(), projectID, teamID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to assign team: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// RemoveTeam removes a team from a project
func (r *ProjectRepository) RemoveTeam(ctx context.Context, projectID, teamID uuid.UUID) error {
	query := `DELETE FROM project_teams WHERE project_id = $1 AND team_id = $2`
	result, err := r.db.Exec(ctx, query, projectID, teamID)
	if err != nil {
		return fmt.Errorf("failed to remove team: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("team assignment not found")
	}
	return nil
}

// GetTeams gets all teams assigned to a project
func (r *ProjectRepository) GetTeams(ctx context.Context, projectID uuid.UUID) ([]models.Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.organization_id, t.created_by, t.created_at, t.updated_at
		FROM teams t
		JOIN project_teams pt ON t.id = pt.team_id
		WHERE pt.project_id = $1
		ORDER BY t.name ASC`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project teams: %w", err)
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

// CountTasks returns the number of tasks in a project
func (r *ProjectRepository) CountTasks(ctx context.Context, projectID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE project_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, projectID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	return count, nil
}

// IsTeamAssigned checks if a team is assigned to a project
func (r *ProjectRepository) IsTeamAssigned(ctx context.Context, projectID, teamID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM project_teams WHERE project_id = $1 AND team_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, projectID, teamID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check team assignment: %w", err)
	}
	return exists, nil
}
