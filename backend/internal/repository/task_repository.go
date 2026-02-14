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

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	if task.Status == "" {
		task.Status = models.TaskStatusDraft
	}

	query := `
		INSERT INTO tasks (
			id, title, description, description_embedding, status,
			created_by, assigned_to, project_id, organization_id,
			estimated_days, tags, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	// Handle empty embedding - pass nil instead of empty vector
	var embedding interface{}
	if len(task.DescriptionEmbedding.Slice()) > 0 {
		embedding = task.DescriptionEmbedding
	}

	_, err := r.db.Exec(ctx, query,
		task.ID, task.Title, task.Description, embedding, task.Status,
		task.CreatedBy, task.AssignedTo, task.ProjectID, task.OrganizationID,
		task.EstimatedDays, task.Tags, task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	query := `
		SELECT id, title, description, status, created_by, assigned_to,
			project_id, organization_id, estimated_days, predicted_days_low,
			predicted_days_high, prediction_confidence, planning_quality_score,
			acknowledged_at, started_at, completed_at, actual_days, tags,
			created_at, updated_at
		FROM tasks WHERE id = $1`

	task := &models.Task{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedBy,
		&task.AssignedTo, &task.ProjectID, &task.OrganizationID, &task.EstimatedDays,
		&task.PredictedDaysLow, &task.PredictedDaysHigh, &task.PredictionConfidence,
		&task.PlanningQualityScore, &task.AcknowledgedAt, &task.StartedAt,
		&task.CompletedAt, &task.ActualDays, &task.Tags, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now()

	query := `
		UPDATE tasks SET
			title = $2, description = $3, status = $4, assigned_to = $5,
			estimated_days = $6, predicted_days_low = $7, predicted_days_high = $8,
			prediction_confidence = $9, planning_quality_score = $10,
			acknowledged_at = $11, started_at = $12, completed_at = $13,
			actual_days = $14, tags = $15, updated_at = $16
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.AssignedTo,
		task.EstimatedDays, task.PredictedDaysLow, task.PredictedDaysHigh,
		task.PredictionConfidence, task.PlanningQualityScore,
		task.AcknowledgedAt, task.StartedAt, task.CompletedAt,
		task.ActualDays, task.Tags, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tasks SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, models.TaskStatusCancelled, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (r *TaskRepository) List(ctx context.Context, filters models.TaskFilters) ([]models.Task, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIdx))
	args = append(args, filters.OrganizationID)
	argIdx++

	if filters.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIdx))
		args = append(args, *filters.ProjectID)
		argIdx++
	}
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filters.Status)
		argIdx++
	}
	if filters.AssignedTo != nil {
		conditions = append(conditions, fmt.Sprintf("assigned_to = $%d", argIdx))
		args = append(args, *filters.AssignedTo)
		argIdx++
	}
	if filters.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIdx))
		args = append(args, *filters.CreatedBy)
		argIdx++
	}
	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, '')) @@ plainto_tsquery('english', $%d))", argIdx))
		args = append(args, filters.Search)
		argIdx++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Main query with pagination
	offset := (filters.Page - 1) * filters.PageSize
	query := fmt.Sprintf(`
		SELECT id, title, description, status, created_by, assigned_to,
			project_id, organization_id, estimated_days, predicted_days_low,
			predicted_days_high, prediction_confidence, planning_quality_score,
			acknowledged_at, started_at, completed_at, actual_days, tags,
			created_at, updated_at
		FROM tasks WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)

	args = append(args, filters.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedBy,
			&task.AssignedTo, &task.ProjectID, &task.OrganizationID, &task.EstimatedDays,
			&task.PredictedDaysLow, &task.PredictedDaysHigh, &task.PredictionConfidence,
			&task.PlanningQualityScore, &task.AcknowledgedAt, &task.StartedAt,
			&task.CompletedAt, &task.ActualDays, &task.Tags, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

func (r *TaskRepository) FindSimilar(ctx context.Context, embedding pgvector.Vector, orgID uuid.UUID, excludeID *uuid.UUID, limit int) ([]models.SimilarTask, error) {
	return r.FindSimilarWithBoost(ctx, embedding, orgID, excludeID, nil, limit)
}

// FindSimilarWithBoost returns similar tasks with boost factors applied for better ranking
// Boost factors:
// - Same project: +0.1
// - Failed/delayed outcome (actual > estimated * 1.5): +0.15
// - Recent (< 6 months): +0.05
// - Has documented lessons: +0.1
func (r *TaskRepository) FindSimilarWithBoost(ctx context.Context, embedding pgvector.Vector, orgID uuid.UUID, excludeID *uuid.UUID, projectID *uuid.UUID, limit int) ([]models.SimilarTask, error) {
	query := `
		WITH base_similarity AS (
			SELECT
				t.id, t.title, t.status, t.project_id,
				t.actual_days, t.estimated_days, t.created_at,
				1 - (t.description_embedding <=> $1) as base_score,
				COALESCE((
					SELECT string_agg(DISTINCT tb.blocker_type::text, ', ')
					FROM task_blockers tb WHERE tb.task_id = t.id
				), '') as blockers_summary,
				COALESCE((
					SELECT LEFT(tr.lessons_learned, 200)
					FROM task_reviews tr WHERE tr.task_id = t.id LIMIT 1
				), '') as lessons_learned_excerpt,
				EXISTS(SELECT 1 FROM task_reviews tr WHERE tr.task_id = t.id AND tr.lessons_learned IS NOT NULL AND tr.lessons_learned != '') as has_lessons
			FROM tasks t
			WHERE t.organization_id = $2
				AND t.status != 'cancelled'
				AND t.description_embedding IS NOT NULL
				AND ($3::uuid IS NULL OR t.id != $3)
				AND 1 - (t.description_embedding <=> $1) > 0.6
		)
		SELECT
			id, title, status,
			LEAST(1.0,
				base_score
				+ CASE WHEN $4::uuid IS NOT NULL AND project_id = $4 THEN 0.1 ELSE 0 END
				+ CASE WHEN actual_days IS NOT NULL AND estimated_days IS NOT NULL
				       AND actual_days > estimated_days * 1.5 THEN 0.15 ELSE 0 END
				+ CASE WHEN created_at > NOW() - INTERVAL '6 months' THEN 0.05 ELSE 0 END
				+ CASE WHEN has_lessons THEN 0.1 ELSE 0 END
			) as similarity_score,
			actual_days, estimated_days,
			blockers_summary, lessons_learned_excerpt
		FROM base_similarity
		WHERE base_score > 0.6
		ORDER BY similarity_score DESC
		LIMIT $5`

	rows, err := r.db.Query(ctx, query, embedding, orgID, excludeID, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find similar tasks: %w", err)
	}
	defer rows.Close()

	var similar []models.SimilarTask
	for rows.Next() {
		var s models.SimilarTask
		err := rows.Scan(&s.ID, &s.Title, &s.Status, &s.SimilarityScore,
			&s.ActualDays, &s.EstimatedDays, &s.BlockersSummary, &s.LessonsLearnedExcerpt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan similar task: %w", err)
		}
		similar = append(similar, s)
	}

	return similar, nil
}

func (r *TaskRepository) UpdateEmbedding(ctx context.Context, id uuid.UUID, embedding pgvector.Vector) error {
	query := `UPDATE tasks SET description_embedding = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, embedding, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update embedding: %w", err)
	}
	return nil
}

func (r *TaskRepository) UpdatePredictions(ctx context.Context, id uuid.UUID, low, high, confidence float64) error {
	query := `UPDATE tasks SET predicted_days_low = $2, predicted_days_high = $3,
			  prediction_confidence = $4, updated_at = $5 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, low, high, confidence, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update predictions: %w", err)
	}
	return nil
}

// GetByIDs retrieves multiple tasks by their IDs in a single query.
// This avoids N+1 query patterns when fetching multiple tasks.
func (r *TaskRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Task, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, title, description, status, created_by, assigned_to,
			project_id, organization_id, estimated_days, predicted_days_low,
			predicted_days_high, prediction_confidence, planning_quality_score,
			acknowledged_at, started_at, completed_at, actual_days, tags,
			created_at, updated_at
		FROM tasks WHERE id = ANY($1)`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by IDs: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedBy,
			&task.AssignedTo, &task.ProjectID, &task.OrganizationID, &task.EstimatedDays,
			&task.PredictedDaysLow, &task.PredictedDaysHigh, &task.PredictionConfidence,
			&task.PlanningQualityScore, &task.AcknowledgedAt, &task.StartedAt,
			&task.CompletedAt, &task.ActualDays, &task.Tags, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskRepository) UpdatePlanningScore(ctx context.Context, id uuid.UUID, score float64) error {
	query := `UPDATE tasks SET planning_quality_score = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, score, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update planning score: %w", err)
	}
	return nil
}
