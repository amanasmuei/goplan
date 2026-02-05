package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/user"
	"github.com/goplan/goplan/internal/postgres/sqlc"
	"github.com/goplan/goplan/internal/repository"
)

// UserRepository implements repository.UserRepository using PostgreSQL.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, u *user.UserWithPassword) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: textPtr(u.PasswordHash),
		AvatarUrl:    textPtrFromPtr(u.AvatarURL),
	})
	if err != nil {
		return MapError(err, "user")
	}

	// Update the user with the generated ID and timestamps
	u.ID = result.ID
	u.CreatedAt = result.CreatedAt.Time
	u.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetUserByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "user")
	}

	return sqlcUserToDomain(result), nil
}

// GetByEmail retrieves a user by email (includes password hash for auth).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.UserWithPassword, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, MapError(err, "user")
	}

	return sqlcUserToDomainWithPassword(result), nil
}

// Update updates user fields.
func (r *UserRepository) Update(ctx context.Context, id string, input *user.UpdateUserInput) (*user.User, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	params := sqlc.UpdateUserParams{
		ID: id,
	}
	if input.Name != nil {
		params.Name = textPtr(*input.Name)
	}
	if input.AvatarURL != nil {
		params.AvatarUrl = textPtr(*input.AvatarURL)
	}

	result, err := q.UpdateUser(ctx, params)
	if err != nil {
		return nil, MapError(err, "user")
	}

	return sqlcUserToDomain(result), nil
}

// Delete deletes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.DeleteUser(ctx, id)
	if err != nil {
		return MapError(err, "user")
	}

	return nil
}

// ExistsByEmail checks if a user with the given email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	exists, err := q.ExistsByEmail(ctx, email)
	if err != nil {
		return false, MapError(err, "user")
	}

	return exists, nil
}

// List retrieves users with pagination.
func (r *UserRepository) List(ctx context.Context, pagination repository.Pagination) (*repository.PaginatedResult[user.User], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	users, err := q.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "user")
	}

	count, err := q.CountUsers(ctx)
	if err != nil {
		return nil, MapError(err, "user")
	}

	items := make([]user.User, len(users))
	for i, u := range users {
		items[i] = *sqlcUserToDomain(u)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// GetByIDs retrieves multiple users by their IDs.
func (r *UserRepository) GetByIDs(ctx context.Context, ids []string) ([]*user.User, error) {
	if len(ids) == 0 {
		return []*user.User{}, nil
	}

	q := sqlc.New(GetConn(ctx, r.pool))

	users, err := q.GetUsersByIDs(ctx, ids)
	if err != nil {
		return nil, MapError(err, "user")
	}

	result := make([]*user.User, len(users))
	for i, u := range users {
		result[i] = sqlcUserToDomain(u)
	}

	return result, nil
}

// Helper functions for type conversion

func sqlcUserToDomain(u sqlc.User) *user.User {
	return &user.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		AvatarURL: textToPtr(u.AvatarUrl),
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func sqlcUserToDomainWithPassword(u sqlc.User) *user.UserWithPassword {
	return &user.UserWithPassword{
		User:         *sqlcUserToDomain(u),
		PasswordHash: textToString(u.PasswordHash),
	}
}

// textPtr creates a pgtype.Text from a string.
func textPtr(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// textPtrFromPtr creates a pgtype.Text from a *string.
func textPtrFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// textToPtr converts a pgtype.Text to a *string.
func textToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// textToString converts a pgtype.Text to a string.
func textToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}
