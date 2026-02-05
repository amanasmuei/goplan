package repository

import (
	"context"

	"github.com/goplan/goplan/internal/domain/plan"
)

// MilestoneFilterOptions defines milestone filtering options.
type MilestoneFilterOptions struct {
	PlanID *string
	Status *string
}

// MilestoneRepository defines milestone data access operations.
type MilestoneRepository interface {
	// Create creates a new milestone.
	Create(ctx context.Context, m *plan.Milestone) error

	// GetByID retrieves a milestone by ID.
	GetByID(ctx context.Context, id string) (*plan.Milestone, error)

	// Update updates milestone fields.
	Update(ctx context.Context, id string, name, dueDate string, description *string) (*plan.Milestone, error)

	// Delete deletes a milestone by ID.
	Delete(ctx context.Context, id string) error

	// ListByPlan retrieves all milestones for a plan.
	ListByPlan(ctx context.Context, planID string) ([]*plan.Milestone, error)

	// UpdateStatus updates milestone status.
	UpdateStatus(ctx context.Context, id string, status string) error

	// LinkTasks links tasks to a milestone.
	LinkTasks(ctx context.Context, milestoneID string, taskIDs []string) error

	// UnlinkTask unlinks a task from a milestone.
	UnlinkTask(ctx context.Context, milestoneID, taskID string) error

	// GetUpcoming retrieves milestones due within the specified days.
	GetUpcoming(ctx context.Context, planID string, withinDays int) ([]*plan.Milestone, error)
}
