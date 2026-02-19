package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"github.com/goplan/backend/internal/models"
)

type PlanRepository struct {
	db *pgxpool.Pool
}

func NewPlanRepository(db *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) Create(ctx context.Context, plan *models.StrategicPlan) error {
	plan.ID = uuid.New()
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	if plan.Status == "" {
		plan.Status = models.PlanStatusDraft
	}
	if plan.CurrentVersion == 0 {
		plan.CurrentVersion = 1
	}

	query := `
		INSERT INTO strategic_plans (
			id, user_id, organization_id, title, original_input,
			category, sub_category, complexity, status, current_version,
			metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err := r.db.Exec(ctx, query,
		plan.ID, plan.UserID, plan.OrganizationID, plan.Title, plan.OriginalInput,
		plan.Category, plan.SubCategory, plan.Complexity, plan.Status, plan.CurrentVersion,
		plan.Metadata, plan.CreatedAt, plan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.StrategicPlan, error) {
	query := `
		SELECT id, user_id, organization_id, title, original_input,
			category, sub_category, complexity, status, current_version,
			metadata, created_at, updated_at
		FROM strategic_plans WHERE id = $1`

	plan := &models.StrategicPlan{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&plan.ID, &plan.UserID, &plan.OrganizationID, &plan.Title, &plan.OriginalInput,
		&plan.Category, &plan.SubCategory, &plan.Complexity, &plan.Status, &plan.CurrentVersion,
		&plan.Metadata, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	return plan, nil
}

func (r *PlanRepository) List(ctx context.Context, filters models.PlanFilters) ([]models.StrategicPlan, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filters.UserID)
		argIdx++
	}
	if filters.OrganizationID != nil {
		conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIdx))
		args = append(args, *filters.OrganizationID)
		argIdx++
	}
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filters.Status)
		argIdx++
	}
	if filters.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *filters.Category)
		argIdx++
	}
	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIdx))
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	whereClause := "TRUE"
	if len(conditions) > 0 {
		whereClause = strings.Join(conditions, " AND ")
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM strategic_plans WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count plans: %w", err)
	}

	// Pagination defaults
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT id, user_id, organization_id, title, original_input,
			category, sub_category, complexity, status, current_version,
			metadata, created_at, updated_at
		FROM strategic_plans WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	var plans []models.StrategicPlan
	for rows.Next() {
		var plan models.StrategicPlan
		err := rows.Scan(
			&plan.ID, &plan.UserID, &plan.OrganizationID, &plan.Title, &plan.OriginalInput,
			&plan.Category, &plan.SubCategory, &plan.Complexity, &plan.Status, &plan.CurrentVersion,
			&plan.Metadata, &plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan plan: %w", err)
		}
		plans = append(plans, plan)
	}

	return plans, total, nil
}

func (r *PlanRepository) Update(ctx context.Context, plan *models.StrategicPlan) error {
	plan.UpdatedAt = time.Now()

	query := `
		UPDATE strategic_plans SET
			title = $2, status = $3, current_version = $4,
			metadata = $5, updated_at = $6
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		plan.ID, plan.Title, plan.Status, plan.CurrentVersion,
		plan.Metadata, plan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE strategic_plans SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, models.PlanStatusArchived, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM strategic_plans WHERE user_id = $1 AND status != 'archived'`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count plans by user: %w", err)
	}
	return count, nil
}

func (r *PlanRepository) UpdateEmbedding(ctx context.Context, id uuid.UUID, embedding pgvector.Vector) error {
	query := `UPDATE strategic_plans SET content_embedding = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, embedding, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update plan embedding: %w", err)
	}
	return nil
}

func (r *PlanRepository) FindWithoutEmbedding(ctx context.Context, limit int) ([]models.StrategicPlan, error) {
	query := `
		SELECT id, user_id, organization_id, title, original_input,
			category, sub_category, complexity, status, current_version,
			metadata, created_at, updated_at
		FROM strategic_plans
		WHERE content_embedding IS NULL AND status != 'archived'
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find plans without embedding: %w", err)
	}
	defer rows.Close()

	var plans []models.StrategicPlan
	for rows.Next() {
		var plan models.StrategicPlan
		err := rows.Scan(
			&plan.ID, &plan.UserID, &plan.OrganizationID, &plan.Title, &plan.OriginalInput,
			&plan.Category, &plan.SubCategory, &plan.Complexity, &plan.Status, &plan.CurrentVersion,
			&plan.Metadata, &plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}
		plans = append(plans, plan)
	}

	return plans, rows.Err()
}

func (r *PlanRepository) FindSimilar(ctx context.Context, embedding pgvector.Vector, orgID uuid.UUID, excludeID *uuid.UUID, limit int) ([]models.StrategicPlan, error) {
	query := `
		SELECT id, user_id, organization_id, title, original_input,
			category, sub_category, complexity, status, current_version,
			metadata, created_at, updated_at
		FROM strategic_plans
		WHERE organization_id = $1
			AND status != 'archived'
			AND content_embedding IS NOT NULL
			AND ($2::uuid IS NULL OR id != $2)
			AND (content_embedding <=> $3) < 0.4
		ORDER BY content_embedding <=> $3
		LIMIT $4`

	rows, err := r.db.Query(ctx, query, orgID, excludeID, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar plans: %w", err)
	}
	defer rows.Close()

	var plans []models.StrategicPlan
	for rows.Next() {
		var plan models.StrategicPlan
		err := rows.Scan(
			&plan.ID, &plan.UserID, &plan.OrganizationID, &plan.Title, &plan.OriginalInput,
			&plan.Category, &plan.SubCategory, &plan.Complexity, &plan.Status, &plan.CurrentVersion,
			&plan.Metadata, &plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan similar plan: %w", err)
		}
		plans = append(plans, plan)
	}

	return plans, nil
}
