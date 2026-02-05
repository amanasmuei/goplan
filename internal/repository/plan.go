package repository

import (
	"context"

	"github.com/goplan/goplan/internal/domain/plan"
)

// PlanFilterOptions defines plan filtering options.
type PlanFilterOptions struct {
	WorkspaceID *string
	OwnerID     *string
	Status      *string
	Domain      *string
	Tags        []string // Match any of these tags
}

// PlanSortField defines sortable fields for plans.
type PlanSortField string

const (
	PlanSortByCreatedAt PlanSortField = "created_at"
	PlanSortByUpdatedAt PlanSortField = "updated_at"
	PlanSortByName      PlanSortField = "name"
	PlanSortByStartDate PlanSortField = "start_date"
)

// PlanSortOptions defines sorting for plans.
type PlanSortOptions struct {
	Field PlanSortField
	Order SortOrder
}

// DefaultPlanSort returns default plan sorting options.
func DefaultPlanSort() PlanSortOptions {
	return PlanSortOptions{
		Field: PlanSortByCreatedAt,
		Order: SortDesc,
	}
}

// PlanRepository defines plan data access operations.
type PlanRepository interface {
	// Create creates a new plan.
	Create(ctx context.Context, p *plan.Plan) error

	// GetByID retrieves a plan by ID.
	GetByID(ctx context.Context, id string) (*plan.Plan, error)

	// Update updates plan fields.
	Update(ctx context.Context, id string, input *plan.UpdatePlanInput) (*plan.Plan, error)

	// Delete deletes a plan by ID.
	Delete(ctx context.Context, id string) error

	// List retrieves plans with filtering, sorting, and pagination.
	List(ctx context.Context, filter PlanFilterOptions, sort PlanSortOptions, pagination Pagination) (*PaginatedResult[plan.Plan], error)

	// GetByWorkspace retrieves all plans in a workspace.
	GetByWorkspace(ctx context.Context, workspaceID string, pagination Pagination) (*PaginatedResult[plan.Plan], error)

	// UpdateStatus updates plan status.
	UpdateStatus(ctx context.Context, id string, status string) error

	// UpdateTags updates plan tags.
	UpdateTags(ctx context.Context, id string, tags []string) error
}
