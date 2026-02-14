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

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, name, role, organization_id, created_at, updated_at
		FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role,
		&user.OrganizationID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, name, role, organization_id, created_at, updated_at
		FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role,
		&user.OrganizationID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// GetByEmailWithPassword retrieves a user by email including their password hash for authentication
func (r *UserRepository) GetByEmailWithPassword(ctx context.Context, email string) (*models.User, string, error) {
	query := `
		SELECT id, email, name, role, organization_id, created_at, updated_at, COALESCE(password_hash, '')
		FROM users WHERE email = $1`

	user := &models.User{}
	var passwordHash string
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Role,
		&user.OrganizationID, &user.CreatedAt, &user.UpdatedAt, &passwordHash,
	)
	if err == pgx.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, passwordHash, nil
}

// Create creates a new user with a password
func (r *UserRepository) Create(ctx context.Context, user *models.User, passwordHash string) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, email, name, role, organization_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.Name, user.Role,
		user.OrganizationID, passwordHash, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// UpdatePassword updates a user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`
	result, err := r.db.Exec(ctx, query, userID, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ListByOrganization lists users in an organization with pagination.
// limit defaults to 100, max 200. offset defaults to 0.
func (r *UserRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]models.User, error) {
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
		SELECT id, email, name, role, organization_id, created_at, updated_at
		FROM users WHERE organization_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.Role,
			&user.OrganizationID, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

// EmailExists checks if an email is already registered
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

// GetDefaultOrganization gets the default organization for new users
func (r *UserRepository) GetDefaultOrganization(ctx context.Context) (uuid.UUID, error) {
	query := `SELECT id FROM organizations ORDER BY created_at ASC LIMIT 1`
	var orgID uuid.UUID
	err := r.db.QueryRow(ctx, query).Scan(&orgID)
	if err == pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("no organization found")
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get default organization: %w", err)
	}
	return orgID, nil
}
