package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/postgres/sqlc"
	"github.com/goplan/goplan/internal/repository"
)

// ActivityLogRepository implements repository.ActivityLogRepository using PostgreSQL.
type ActivityLogRepository struct {
	pool *pgxpool.Pool
}

// NewActivityLogRepository creates a new ActivityLogRepository.
func NewActivityLogRepository(pool *pgxpool.Pool) *ActivityLogRepository {
	return &ActivityLogRepository{pool: pool}
}

// Create creates a new activity log entry.
func (r *ActivityLogRepository) Create(ctx context.Context, a *task.ActivityLog) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	detailsJSON, err := json.Marshal(a.Details)
	if err != nil {
		return err
	}

	result, err := q.CreateActivityLog(ctx, sqlc.CreateActivityLogParams{
		WorkspaceID: a.WorkspaceID,
		PlanID:      uuidFromPtr(a.PlanID),
		TaskID:      uuidFromPtr(a.TaskID),
		UserID:      a.UserID,
		Action:      a.Action,
		Details:     detailsJSON,
	})
	if err != nil {
		return MapError(err, "activity_log")
	}

	a.ID = result.ID
	a.CreatedAt = result.CreatedAt.Time

	return nil
}

// List retrieves activity logs with filtering and pagination.
func (r *ActivityLogRepository) List(ctx context.Context, filter repository.ActivityFilterOptions, pagination repository.Pagination) (*repository.PaginatedResult[task.ActivityLog], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	params := sqlc.ListActivityFilteredParams{
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	}

	if filter.WorkspaceID != nil {
		params.WorkspaceID = uuidFromPtr(filter.WorkspaceID)
	}
	if filter.PlanID != nil {
		params.PlanID = uuidFromPtr(filter.PlanID)
	}
	if filter.TaskID != nil {
		params.TaskID = uuidFromPtr(filter.TaskID)
	}
	if filter.UserID != nil {
		params.UserID = uuidFromPtr(filter.UserID)
	}
	if filter.Action != nil {
		params.Action = textPtr(*filter.Action)
	}
	if filter.Since != nil {
		params.Since = timestamptzFromTime(*filter.Since)
	}
	if filter.Until != nil {
		params.Until = timestamptzFromTime(*filter.Until)
	}

	activities, err := q.ListActivityFiltered(ctx, params)
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	countParams := sqlc.CountActivityFilteredParams{}
	if filter.WorkspaceID != nil {
		countParams.WorkspaceID = uuidFromPtr(filter.WorkspaceID)
	}
	if filter.PlanID != nil {
		countParams.PlanID = uuidFromPtr(filter.PlanID)
	}
	if filter.TaskID != nil {
		countParams.TaskID = uuidFromPtr(filter.TaskID)
	}
	if filter.UserID != nil {
		countParams.UserID = uuidFromPtr(filter.UserID)
	}
	if filter.Action != nil {
		countParams.Action = textPtr(*filter.Action)
	}
	if filter.Since != nil {
		countParams.Since = timestamptzFromTime(*filter.Since)
	}
	if filter.Until != nil {
		countParams.Until = timestamptzFromTime(*filter.Until)
	}

	count, err := q.CountActivityFiltered(ctx, countParams)
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	items := make([]task.ActivityLog, len(activities))
	for i, a := range activities {
		items[i] = *sqlcActivityFilteredRowToDomain(a)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// ListByWorkspace retrieves activity logs for a workspace.
func (r *ActivityLogRepository) ListByWorkspace(ctx context.Context, workspaceID string, pagination repository.Pagination) (*repository.PaginatedResult[task.ActivityLog], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	activities, err := q.ListActivityByWorkspace(ctx, sqlc.ListActivityByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       int32(pagination.PageSize),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	count, err := q.CountActivityByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	items := make([]task.ActivityLog, len(activities))
	for i, a := range activities {
		items[i] = *sqlcActivityWorkspaceRowToDomain(a)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// ListByPlan retrieves activity logs for a plan.
func (r *ActivityLogRepository) ListByPlan(ctx context.Context, planID string, pagination repository.Pagination) (*repository.PaginatedResult[task.ActivityLog], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	activities, err := q.ListActivityByPlan(ctx, sqlc.ListActivityByPlanParams{
		PlanID: uuidFromString(planID),
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	count, err := q.CountActivityByPlan(ctx, uuidFromString(planID))
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	items := make([]task.ActivityLog, len(activities))
	for i, a := range activities {
		items[i] = *sqlcActivityPlanRowToDomain(a)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// ListByTask retrieves activity logs for a task.
func (r *ActivityLogRepository) ListByTask(ctx context.Context, taskID string, pagination repository.Pagination) (*repository.PaginatedResult[task.ActivityLog], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	activities, err := q.ListActivityByTask(ctx, sqlc.ListActivityByTaskParams{
		TaskID: uuidFromString(taskID),
		Limit:  int32(pagination.PageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	count, err := q.CountActivityByTask(ctx, uuidFromString(taskID))
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	items := make([]task.ActivityLog, len(activities))
	for i, a := range activities {
		items[i] = *sqlcActivityTaskRowToDomain(a)
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// GetRecent retrieves recent activity for a user.
func (r *ActivityLogRepository) GetRecent(ctx context.Context, userID string, limit int) ([]*task.ActivityLog, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	activities, err := q.GetRecentActivity(ctx, sqlc.GetRecentActivityParams{
		UserID: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, MapError(err, "activity_log")
	}

	result := make([]*task.ActivityLog, len(activities))
	for i, a := range activities {
		result[i] = sqlcActivityRecentRowToDomain(a)
	}

	return result, nil
}

// Helper functions

func sqlcActivityFilteredRowToDomain(a sqlc.ListActivityFilteredRow) *task.ActivityLog {
	var details map[string]interface{}
	if len(a.Details) > 0 {
		if err := json.Unmarshal(a.Details, &details); err != nil {
			details = make(map[string]interface{})
		}
	} else {
		details = make(map[string]interface{})
	}

	return &task.ActivityLog{
		ID:          a.ID,
		WorkspaceID: a.WorkspaceID,
		PlanID:      uuidToPtr(a.PlanID),
		TaskID:      uuidToPtr(a.TaskID),
		UserID:      a.UserID,
		Action:      a.Action,
		Details:     details,
		CreatedAt:   a.CreatedAt.Time,
	}
}

func sqlcActivityWorkspaceRowToDomain(a sqlc.ListActivityByWorkspaceRow) *task.ActivityLog {
	var details map[string]interface{}
	if len(a.Details) > 0 {
		if err := json.Unmarshal(a.Details, &details); err != nil {
			details = make(map[string]interface{})
		}
	} else {
		details = make(map[string]interface{})
	}

	return &task.ActivityLog{
		ID:          a.ID,
		WorkspaceID: a.WorkspaceID,
		PlanID:      uuidToPtr(a.PlanID),
		TaskID:      uuidToPtr(a.TaskID),
		UserID:      a.UserID,
		Action:      a.Action,
		Details:     details,
		CreatedAt:   a.CreatedAt.Time,
	}
}

func sqlcActivityPlanRowToDomain(a sqlc.ListActivityByPlanRow) *task.ActivityLog {
	var details map[string]interface{}
	if len(a.Details) > 0 {
		if err := json.Unmarshal(a.Details, &details); err != nil {
			details = make(map[string]interface{})
		}
	} else {
		details = make(map[string]interface{})
	}

	return &task.ActivityLog{
		ID:          a.ID,
		WorkspaceID: a.WorkspaceID,
		PlanID:      uuidToPtr(a.PlanID),
		TaskID:      uuidToPtr(a.TaskID),
		UserID:      a.UserID,
		Action:      a.Action,
		Details:     details,
		CreatedAt:   a.CreatedAt.Time,
	}
}

func sqlcActivityTaskRowToDomain(a sqlc.ListActivityByTaskRow) *task.ActivityLog {
	var details map[string]interface{}
	if len(a.Details) > 0 {
		if err := json.Unmarshal(a.Details, &details); err != nil {
			details = make(map[string]interface{})
		}
	} else {
		details = make(map[string]interface{})
	}

	return &task.ActivityLog{
		ID:          a.ID,
		WorkspaceID: a.WorkspaceID,
		PlanID:      uuidToPtr(a.PlanID),
		TaskID:      uuidToPtr(a.TaskID),
		UserID:      a.UserID,
		Action:      a.Action,
		Details:     details,
		CreatedAt:   a.CreatedAt.Time,
	}
}

func sqlcActivityRecentRowToDomain(a sqlc.GetRecentActivityRow) *task.ActivityLog {
	var details map[string]interface{}
	if len(a.Details) > 0 {
		if err := json.Unmarshal(a.Details, &details); err != nil {
			details = make(map[string]interface{})
		}
	} else {
		details = make(map[string]interface{})
	}

	return &task.ActivityLog{
		ID:          a.ID,
		WorkspaceID: a.WorkspaceID,
		PlanID:      uuidToPtr(a.PlanID),
		TaskID:      uuidToPtr(a.TaskID),
		UserID:      a.UserID,
		Action:      a.Action,
		Details:     details,
		CreatedAt:   a.CreatedAt.Time,
	}
}
