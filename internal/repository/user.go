package repository

import (
	"context"

	"github.com/goplan/goplan/internal/domain/user"
)

// UserRepository defines user data access operations.
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, u *user.UserWithPassword) error

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id string) (*user.User, error)

	// GetByEmail retrieves a user by email (includes password hash for auth).
	GetByEmail(ctx context.Context, email string) (*user.UserWithPassword, error)

	// Update updates user fields.
	Update(ctx context.Context, id string, input *user.UpdateUserInput) (*user.User, error)

	// Delete deletes a user by ID.
	Delete(ctx context.Context, id string) error

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// List retrieves users with pagination.
	List(ctx context.Context, pagination Pagination) (*PaginatedResult[user.User], error)

	// GetByIDs retrieves multiple users by their IDs.
	GetByIDs(ctx context.Context, ids []string) ([]*user.User, error)
}
