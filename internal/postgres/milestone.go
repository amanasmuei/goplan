package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/postgres/sqlc"
)

// MilestoneRepository implements repository.MilestoneRepository using PostgreSQL.
type MilestoneRepository struct {
	pool *pgxpool.Pool
}

// NewMilestoneRepository creates a new MilestoneRepository.
func NewMilestoneRepository(pool *pgxpool.Pool) *MilestoneRepository {
	return &MilestoneRepository{pool: pool}
}

// Create creates a new milestone.
func (r *MilestoneRepository) Create(ctx context.Context, m *plan.Milestone) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	dueDate, err := parseDate(m.DueDate)
	if err != nil {
		return err
	}

	result, err := q.CreateMilestone(ctx, sqlc.CreateMilestoneParams{
		PlanID:        m.PlanID,
		Name:          m.Name,
		Description:   textPtrFromPtr(m.Description),
		DueDate:       dateFromTime(dueDate),
		Status:        textPtr(m.Status),
		LinkedTaskIds: m.LinkedTaskIDs,
	})
	if err != nil {
		return MapError(err, "milestone")
	}

	m.ID = result.ID
	m.CreatedAt = result.CreatedAt.Time
	m.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID retrieves a milestone by ID.
func (r *MilestoneRepository) GetByID(ctx context.Context, id string) (*plan.Milestone, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetMilestoneByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "milestone")
	}

	return sqlcMilestoneToDomain(result), nil
}

// Update updates milestone fields.
func (r *MilestoneRepository) Update(ctx context.Context, id string, name, dueDate string, description *string) (*plan.Milestone, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	params := sqlc.UpdateMilestoneParams{
		ID:   id,
		Name: textPtr(name),
	}
	if description != nil {
		params.Description = textPtr(*description)
	}
	if dueDate != "" {
		params.DueDate = dateFromPtr(&dueDate)
	}

	result, err := q.UpdateMilestone(ctx, params)
	if err != nil {
		return nil, MapError(err, "milestone")
	}

	return sqlcMilestoneToDomain(result), nil
}

// Delete deletes a milestone by ID.
func (r *MilestoneRepository) Delete(ctx context.Context, id string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.DeleteMilestone(ctx, id)
	if err != nil {
		return MapError(err, "milestone")
	}

	return nil
}

// ListByPlan retrieves all milestones for a plan.
func (r *MilestoneRepository) ListByPlan(ctx context.Context, planID string) ([]*plan.Milestone, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	milestones, err := q.ListMilestonesByPlan(ctx, planID)
	if err != nil {
		return nil, MapError(err, "milestone")
	}

	result := make([]*plan.Milestone, len(milestones))
	for i, m := range milestones {
		result[i] = sqlcMilestoneToDomain(m)
	}

	return result, nil
}

// UpdateStatus updates milestone status.
func (r *MilestoneRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	_, err := q.UpdateMilestoneStatus(ctx, sqlc.UpdateMilestoneStatusParams{
		ID:     id,
		Status: textPtr(status),
	})
	if err != nil {
		return MapError(err, "milestone")
	}

	return nil
}

// LinkTasks links tasks to a milestone.
func (r *MilestoneRepository) LinkTasks(ctx context.Context, milestoneID string, taskIDs []string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	_, err := q.LinkTasksToMilestone(ctx, sqlc.LinkTasksToMilestoneParams{
		ID:            milestoneID,
		LinkedTaskIds: taskIDs,
	})
	if err != nil {
		return MapError(err, "milestone")
	}

	return nil
}

// UnlinkTask unlinks a task from a milestone.
func (r *MilestoneRepository) UnlinkTask(ctx context.Context, milestoneID, taskID string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	// Get current milestone
	milestone, err := q.GetMilestoneByID(ctx, milestoneID)
	if err != nil {
		return MapError(err, "milestone")
	}

	// Filter out the task ID
	newTaskIDs := make([]string, 0, len(milestone.LinkedTaskIds))
	for _, id := range milestone.LinkedTaskIds {
		if id != taskID {
			newTaskIDs = append(newTaskIDs, id)
		}
	}

	_, err = q.LinkTasksToMilestone(ctx, sqlc.LinkTasksToMilestoneParams{
		ID:            milestoneID,
		LinkedTaskIds: newTaskIDs,
	})
	if err != nil {
		return MapError(err, "milestone")
	}

	return nil
}

// GetUpcoming retrieves milestones due within the specified days.
func (r *MilestoneRepository) GetUpcoming(ctx context.Context, planID string, withinDays int) ([]*plan.Milestone, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	milestones, err := q.GetUpcomingMilestones(ctx, sqlc.GetUpcomingMilestonesParams{
		PlanID: planID,
		Limit:  int32(withinDays),
	})
	if err != nil {
		return nil, MapError(err, "milestone")
	}

	result := make([]*plan.Milestone, len(milestones))
	for i, m := range milestones {
		result[i] = sqlcMilestoneToDomain(m)
	}

	return result, nil
}

// Helper functions

func sqlcMilestoneToDomain(m sqlc.Milestone) *plan.Milestone {
	linkedTaskIDs := m.LinkedTaskIds
	if linkedTaskIDs == nil {
		linkedTaskIDs = []string{}
	}

	status := textToString(m.Status)
	if status == "" {
		status = shared.MilestoneStatusPending
	}

	return &plan.Milestone{
		ID:            m.ID,
		PlanID:        m.PlanID,
		Name:          m.Name,
		Description:   textToPtr(m.Description),
		DueDate:       m.DueDate.Time.Format("2006-01-02"),
		Status:        status,
		LinkedTaskIDs: linkedTaskIDs,
		CreatedAt:     m.CreatedAt.Time,
		UpdatedAt:     m.UpdatedAt.Time,
	}
}
