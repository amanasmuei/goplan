package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/postgres/sqlc"
	"github.com/goplan/goplan/internal/repository"
)

// PlanRepository implements repository.PlanRepository using PostgreSQL.
type PlanRepository struct {
	pool *pgxpool.Pool
}

// NewPlanRepository creates a new PlanRepository.
func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{pool: pool}
}

// Create creates a new plan.
func (r *PlanRepository) Create(ctx context.Context, p *plan.Plan) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	customStatusesJSON, err := json.Marshal(p.CustomStatuses)
	if err != nil {
		return err
	}
	customFieldsJSON, err := json.Marshal(p.CustomFields)
	if err != nil {
		return err
	}

	result, err := q.CreatePlan(ctx, sqlc.CreatePlanParams{
		WorkspaceID:    p.WorkspaceID,
		Name:           p.Name,
		Description:    textPtrFromPtr(p.Description),
		Domain:         p.Domain,
		Status:         p.Status,
		OwnerID:        p.OwnerID,
		StartDate:      dateFromPtr(p.StartDate),
		EndDate:        dateFromPtr(p.EndDate),
		CustomStatuses: customStatusesJSON,
		CustomFields:   customFieldsJSON,
		Tags:           p.Tags,
	})
	if err != nil {
		return MapError(err, "plan")
	}

	p.ID = result.ID
	p.CreatedAt = result.CreatedAt.Time
	p.UpdatedAt = result.UpdatedAt.Time

	return nil
}

// GetByID retrieves a plan by ID.
func (r *PlanRepository) GetByID(ctx context.Context, id string) (*plan.Plan, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	result, err := q.GetPlanByID(ctx, id)
	if err != nil {
		return nil, MapError(err, "plan")
	}

	return sqlcPlanToDomain(result)
}

// Update updates plan fields.
func (r *PlanRepository) Update(ctx context.Context, id string, input *plan.UpdatePlanInput) (*plan.Plan, error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	params := sqlc.UpdatePlanParams{
		ID: id,
	}
	if input.Name != nil {
		params.Name = textPtr(*input.Name)
	}
	if input.Description != nil {
		params.Description = textPtr(*input.Description)
	}
	if input.Status != nil {
		params.Status = textPtr(*input.Status)
	}
	if input.StartDate != nil {
		params.StartDate = dateFromPtr(input.StartDate)
	}
	if input.EndDate != nil {
		params.EndDate = dateFromPtr(input.EndDate)
	}
	if len(input.CustomStatuses) > 0 {
		customStatusesJSON, err := json.Marshal(input.CustomStatuses)
		if err != nil {
			return nil, err
		}
		params.CustomStatuses = customStatusesJSON
	}
	if len(input.CustomFields) > 0 {
		customFieldsJSON, err := json.Marshal(input.CustomFields)
		if err != nil {
			return nil, err
		}
		params.CustomFields = customFieldsJSON
	}
	if input.Tags != nil {
		params.Tags = input.Tags
	}

	result, err := q.UpdatePlan(ctx, params)
	if err != nil {
		return nil, MapError(err, "plan")
	}

	return sqlcPlanToDomain(result)
}

// Delete deletes a plan by ID.
func (r *PlanRepository) Delete(ctx context.Context, id string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	err := q.DeletePlan(ctx, id)
	if err != nil {
		return MapError(err, "plan")
	}

	return nil
}

// List retrieves plans with filtering, sorting, and pagination.
func (r *PlanRepository) List(ctx context.Context, filter repository.PlanFilterOptions, sort repository.PlanSortOptions, pagination repository.Pagination) (*repository.PaginatedResult[plan.Plan], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	if filter.WorkspaceID == nil {
		return nil, nil
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	params := sqlc.ListPlansByWorkspaceFilteredParams{
		WorkspaceID: *filter.WorkspaceID,
		Limit:       int32(pagination.PageSize),
		Offset:      int32(offset),
	}

	if filter.Status != nil {
		params.Status = textPtr(*filter.Status)
	}
	if filter.OwnerID != nil {
		params.OwnerID = uuidFromPtr(filter.OwnerID)
	}
	if filter.Domain != nil {
		params.Domain = textPtr(*filter.Domain)
	}

	sortField := string(sort.Field)
	if sortField == "" {
		sortField = "created_at"
	}
	sortOrder := string(sort.Order)
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	params.SortField = textPtr(sortField)
	params.SortOrder = textPtr(sortOrder)

	plans, err := q.ListPlansByWorkspaceFiltered(ctx, params)
	if err != nil {
		return nil, MapError(err, "plan")
	}

	countParams := sqlc.CountPlansByWorkspaceFilteredParams{
		WorkspaceID: *filter.WorkspaceID,
	}
	if filter.Status != nil {
		countParams.Status = textPtr(*filter.Status)
	}
	if filter.OwnerID != nil {
		countParams.OwnerID = uuidFromPtr(filter.OwnerID)
	}
	if filter.Domain != nil {
		countParams.Domain = textPtr(*filter.Domain)
	}

	count, err := q.CountPlansByWorkspaceFiltered(ctx, countParams)
	if err != nil {
		return nil, MapError(err, "plan")
	}

	items := make([]plan.Plan, len(plans))
	for i, p := range plans {
		domainPlan, err := sqlcPlanToDomain(p)
		if err != nil {
			return nil, err
		}
		items[i] = *domainPlan
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// GetByWorkspace retrieves all plans in a workspace.
func (r *PlanRepository) GetByWorkspace(ctx context.Context, workspaceID string, pagination repository.Pagination) (*repository.PaginatedResult[plan.Plan], error) {
	q := sqlc.New(GetConn(ctx, r.pool))

	offset := (pagination.Page - 1) * pagination.PageSize

	plans, err := q.ListPlansByWorkspace(ctx, sqlc.ListPlansByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       int32(pagination.PageSize),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, MapError(err, "plan")
	}

	count, err := q.CountPlansByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, MapError(err, "plan")
	}

	items := make([]plan.Plan, len(plans))
	for i, p := range plans {
		domainPlan, err := sqlcPlanToDomain(p)
		if err != nil {
			return nil, err
		}
		items[i] = *domainPlan
	}

	return repository.NewPaginatedResult(items, count, pagination), nil
}

// UpdateStatus updates plan status.
func (r *PlanRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	_, err := q.UpdatePlanStatus(ctx, sqlc.UpdatePlanStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return MapError(err, "plan")
	}

	return nil
}

// UpdateTags updates plan tags.
func (r *PlanRepository) UpdateTags(ctx context.Context, id string, tags []string) error {
	q := sqlc.New(GetConn(ctx, r.pool))

	_, err := q.UpdatePlan(ctx, sqlc.UpdatePlanParams{
		ID:   id,
		Tags: tags,
	})
	if err != nil {
		return MapError(err, "plan")
	}

	return nil
}

// Helper functions

func sqlcPlanToDomain(p sqlc.Plan) (*plan.Plan, error) {
	var customStatuses []plan.StatusDefinition
	if len(p.CustomStatuses) > 0 {
		if err := json.Unmarshal(p.CustomStatuses, &customStatuses); err != nil {
			customStatuses = plan.DefaultStatuses()
		}
	} else {
		customStatuses = plan.DefaultStatuses()
	}

	var customFields []plan.FieldDefinition
	if len(p.CustomFields) > 0 {
		if err := json.Unmarshal(p.CustomFields, &customFields); err != nil {
			customFields = []plan.FieldDefinition{}
		}
	}

	tags := p.Tags
	if tags == nil {
		tags = []string{}
	}

	return &plan.Plan{
		ID:             p.ID,
		WorkspaceID:    p.WorkspaceID,
		Name:           p.Name,
		Description:    textToPtr(p.Description),
		Domain:         p.Domain,
		Status:         p.Status,
		OwnerID:        p.OwnerID,
		StartDate:      dateToPtr(p.StartDate),
		EndDate:        dateToPtr(p.EndDate),
		CustomStatuses: customStatuses,
		CustomFields:   customFields,
		Tags:           tags,
		CreatedAt:      p.CreatedAt.Time,
		UpdatedAt:      p.UpdatedAt.Time,
	}, nil
}

func dateFromPtr(s *string) pgtype.Date {
	if s == nil || *s == "" {
		return pgtype.Date{Valid: false}
	}
	// Parse date string (YYYY-MM-DD format)
	t, err := parseDate(*s)
	if err != nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: t, Valid: true}
}

func dateToPtr(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func uuidFromPtr(s *string) pgtype.UUID {
	if s == nil || *s == "" {
		return pgtype.UUID{Valid: false}
	}
	var uuid pgtype.UUID
	if err := uuid.Scan(*s); err != nil {
		return pgtype.UUID{Valid: false}
	}
	return uuid
}
