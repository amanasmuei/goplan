package repository

import (
	"context"

	"github.com/goplan/goplan/internal/domain/plan"
)

// PhaseRepository defines phase data access operations.
type PhaseRepository interface {
	// Create creates a new phase.
	Create(ctx context.Context, p *plan.Phase) error

	// GetByID retrieves a phase by ID.
	GetByID(ctx context.Context, id string) (*plan.Phase, error)

	// Update updates phase fields.
	Update(ctx context.Context, id string, name string, description *string) (*plan.Phase, error)

	// Delete deletes a phase by ID.
	Delete(ctx context.Context, id string) error

	// ListByPlan retrieves all phases for a plan ordered by order field.
	ListByPlan(ctx context.Context, planID string) ([]*plan.Phase, error)

	// Reorder updates the order of phases in a plan.
	Reorder(ctx context.Context, planID string, phaseIDs []string) error

	// UpdateDates updates phase start and end dates.
	UpdateDates(ctx context.Context, id string, startDate, endDate *string) error
}
