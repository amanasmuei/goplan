package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/postgres/sqlc"
)

// PhaseRepository implements repository.PhaseRepository using PostgreSQL.
type PhaseRepository struct {
	pool *pgxpool.Pool
}

// NewPhaseRepository creates a new PhaseRepository.
func NewPhaseRepository(pool *pgxpool.Pool) *PhaseRepository {
	return &PhaseRepository{pool: pool}
}

// Create creates a new phase.
func (r *PhaseRepository) Create(ctx context.Context, p *plan.Phase) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	// Get max order for the plan
	maxOrder, err := q.GetMaxPhaseOrder(ctx, p.PlanID)
	if err != nil {
		return MapError(err, "phase")
	}

	result, err := q.CreatePhase(ctx, sqlc.CreatePhaseParams{
		PlanID:      p.PlanID,
		Name:        p.Name,
		Description: textPtrFromPtr(p.Description),
		Order:       int4FromInt(int(maxOrder + 1)),
		StartDate:   dateFromPtr(p.StartDate),
		EndDate:     dateFromPtr(p.EndDate),
	})
	if err != nil {
		return MapError(err, "phase")
	}

	p.ID = result.ID
	p.Order = int4ToInt(result.Order)
	p.CreatedAt = result.CreatedAt.Time
	p.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID retrieves a phase by ID.
func (r *PhaseRepository) GetByID(ctx context.Context, id string) (*plan.Phase, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetPhaseByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "phase")
	}

	return sqlcPhaseToDomain(result), nil
}

// Update updates phase fields.
func (r *PhaseRepository) Update(ctx context.Context, id string, name string, description *string) (*plan.Phase, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	params := sqlc.UpdatePhaseParams{
		ID:   id,
		Name: textPtr(name),
	}
	if description != nil {
		params.Description = textPtr(*description)
	}

	result, err := q.UpdatePhase(ctx, params)
	if err != nil {
		return nil, MapError(err, "phase")
	}

	return sqlcPhaseToDomain(result), nil
}

// Delete deletes a phase by ID.
func (r *PhaseRepository) Delete(ctx context.Context, id string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.DeletePhase(ctx, id)
	if err != nil {
		return MapError(err, "phase")
	}

	return nil
}

// ListByPlan retrieves all phases for a plan ordered by order field.
func (r *PhaseRepository) ListByPlan(ctx context.Context, planID string) ([]*plan.Phase, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	phases, err := q.ListPhasesByPlan(ctx, planID)
	if err != nil {
		return nil, MapError(err, "phase")
	}

	result := make([]*plan.Phase, len(phases))
	for i, p := range phases {
		result[i] = sqlcPhaseToDomain(p)
	}

	return result, nil
}

// Reorder updates the order of phases in a plan.
func (r *PhaseRepository) Reorder(ctx context.Context, planID string, phaseIDs []string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.ReorderPhases(ctx, phaseIDs)
	if err != nil {
		return MapError(err, "phase")
	}

	return nil
}

// UpdateDates updates phase start and end dates.
func (r *PhaseRepository) UpdateDates(ctx context.Context, id string, startDate, endDate *string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	params := sqlc.UpdatePhaseParams{
		ID: id,
	}
	if startDate != nil {
		params.StartDate = dateFromPtr(startDate)
	}
	if endDate != nil {
		params.EndDate = dateFromPtr(endDate)
	}

	_, err := q.UpdatePhase(ctx, params)
	if err != nil {
		return MapError(err, "phase")
	}

	return nil
}

// Helper functions

func sqlcPhaseToDomain(p sqlc.Phase) *plan.Phase {
	return &plan.Phase{
		ID:          p.ID,
		PlanID:      p.PlanID,
		Name:        p.Name,
		Description: textToPtr(p.Description),
		Order:       int4ToInt(p.Order),
		StartDate:   dateToPtr(p.StartDate),
		EndDate:     dateToPtr(p.EndDate),
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}
