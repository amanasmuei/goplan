package repository

import (
	"context"

	"github.com/goplan/goplan/internal/domain/task"
)

// CommentRepository defines comment data access operations.
type CommentRepository interface {
	// Create creates a new comment.
	Create(ctx context.Context, c *task.Comment) error

	// GetByID retrieves a comment by ID.
	GetByID(ctx context.Context, id string) (*task.Comment, error)

	// Update updates comment content.
	Update(ctx context.Context, id string, content string, mentions []string) (*task.Comment, error)

	// Delete deletes a comment by ID.
	Delete(ctx context.Context, id string) error

	// ListByTask retrieves all comments for a task with user info.
	ListByTask(ctx context.Context, taskID string) ([]*task.Comment, error)

	// ListByTaskPaginated retrieves paginated comments for a task.
	ListByTaskPaginated(ctx context.Context, taskID string, pagination Pagination) (*PaginatedResult[task.Comment], error)

	// CountByTask counts comments for a task.
	CountByTask(ctx context.Context, taskID string) (int, error)
}
