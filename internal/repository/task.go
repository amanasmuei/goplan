package repository

import (
	"context"

	"github.com/goplan/goplan/internal/domain/task"
)

// TaskFilterOptions defines task filtering options.
type TaskFilterOptions struct {
	PlanID      *string
	PhaseID     *string
	ParentID    *string  // nil = root tasks only, set to empty string for all tasks
	Status      *string
	Priority    *string
	AssigneeID  *string
	Tags        []string // Match any of these tags
	DueBefore   *string  // ISO date
	DueAfter    *string  // ISO date
	SearchQuery *string  // Full-text search
}

// TaskSortField defines sortable fields for tasks.
type TaskSortField string

const (
	TaskSortByPosition  TaskSortField = "position"
	TaskSortByCreatedAt TaskSortField = "created_at"
	TaskSortByUpdatedAt TaskSortField = "updated_at"
	TaskSortByDueDate   TaskSortField = "due_date"
	TaskSortByPriority  TaskSortField = "priority"
)

// TaskSortOptions defines sorting for tasks.
type TaskSortOptions struct {
	Field TaskSortField
	Order SortOrder
}

// DefaultTaskSort returns default task sorting options.
func DefaultTaskSort() TaskSortOptions {
	return TaskSortOptions{
		Field: TaskSortByPosition,
		Order: SortAsc,
	}
}

// TaskRepository defines task data access operations.
type TaskRepository interface {
	// Create creates a new task.
	Create(ctx context.Context, t *task.Task) error

	// GetByID retrieves a task by ID.
	GetByID(ctx context.Context, id string) (*task.Task, error)

	// GetByIDWithDetails retrieves a task with subtasks, dependencies, and comments.
	GetByIDWithDetails(ctx context.Context, id string) (*task.Task, error)

	// Update updates task fields.
	Update(ctx context.Context, id string, input *task.UpdateTaskInput) (*task.Task, error)

	// Delete deletes a task by ID (cascades to subtasks).
	Delete(ctx context.Context, id string) error

	// List retrieves tasks with filtering, sorting, and pagination.
	List(ctx context.Context, filter TaskFilterOptions, sort TaskSortOptions, pagination Pagination) (*PaginatedResult[task.Task], error)

	// ListByPlan retrieves all tasks for a plan (flat list).
	ListByPlan(ctx context.Context, planID string, pagination Pagination) (*PaginatedResult[task.Task], error)

	// GetSubtasks retrieves direct subtasks of a task.
	GetSubtasks(ctx context.Context, parentID string) ([]*task.Task, error)

	// Search performs full-text search on tasks.
	Search(ctx context.Context, planID, query string, pagination Pagination) (*PaginatedResult[task.Task], error)

	// Move updates task status and position (for Kanban).
	Move(ctx context.Context, id string, status string, position int) error

	// BatchUpdatePositions updates positions for multiple tasks.
	BatchUpdatePositions(ctx context.Context, updates map[string]int) error

	// AddDependency adds a task dependency.
	AddDependency(ctx context.Context, dep *task.TaskDependency) error

	// RemoveDependency removes a task dependency.
	RemoveDependency(ctx context.Context, taskID, dependsOnID string) error

	// GetDependencies retrieves all dependencies for a task.
	GetDependencies(ctx context.Context, taskID string) ([]*task.TaskDependency, error)

	// GetDependents retrieves all tasks that depend on this task.
	GetDependents(ctx context.Context, taskID string) ([]*task.TaskDependency, error)

	// GetTasksByStatus retrieves tasks grouped by status for a plan.
	GetTasksByStatus(ctx context.Context, planID string) (map[string][]*task.Task, error)

	// CountByStatus counts tasks by status for a plan.
	CountByStatus(ctx context.Context, planID string) (map[string]int, error)

	// GetOverdue retrieves overdue tasks.
	GetOverdue(ctx context.Context, planID string) ([]*task.Task, error)
}
