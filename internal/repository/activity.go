package repository

import (
	"context"
	"time"

	"github.com/goplan/goplan/internal/domain/task"
)

// ActivityFilterOptions defines activity log filtering options.
type ActivityFilterOptions struct {
	WorkspaceID *string
	PlanID      *string
	TaskID      *string
	UserID      *string
	Action      *string
	Since       *time.Time
	Until       *time.Time
}

// ActivityLogRepository defines activity log data access operations.
type ActivityLogRepository interface {
	// Create creates a new activity log entry.
	Create(ctx context.Context, a *task.ActivityLog) error

	// List retrieves activity logs with filtering and pagination.
	List(ctx context.Context, filter ActivityFilterOptions, pagination Pagination) (*PaginatedResult[task.ActivityLog], error)

	// ListByWorkspace retrieves activity logs for a workspace.
	ListByWorkspace(ctx context.Context, workspaceID string, pagination Pagination) (*PaginatedResult[task.ActivityLog], error)

	// ListByPlan retrieves activity logs for a plan.
	ListByPlan(ctx context.Context, planID string, pagination Pagination) (*PaginatedResult[task.ActivityLog], error)

	// ListByTask retrieves activity logs for a task.
	ListByTask(ctx context.Context, taskID string, pagination Pagination) (*PaginatedResult[task.ActivityLog], error)

	// GetRecent retrieves recent activity for a user.
	GetRecent(ctx context.Context, userID string, limit int) ([]*task.ActivityLog, error)
}
