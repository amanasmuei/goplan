package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goplan/backend/internal/models"
)

type GenerationLogRepository struct {
	db *pgxpool.Pool
}

func NewGenerationLogRepository(db *pgxpool.Pool) *GenerationLogRepository {
	return &GenerationLogRepository{db: db}
}

func (r *GenerationLogRepository) Create(ctx context.Context, log *models.GenerationLog) error {
	log.ID = uuid.New()
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO generation_log (
			id, plan_id, user_id, action, section_type, status,
			prompt_tokens, completion_tokens, model, duration_ms,
			error_message, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		log.ID, log.PlanID, log.UserID, log.Action, log.SectionType, log.Status,
		log.PromptTokens, log.CompletionTokens, log.Model, log.DurationMs,
		log.ErrorMessage, log.Metadata, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create generation log: %w", err)
	}
	return nil
}

func (r *GenerationLogRepository) ListByPlan(ctx context.Context, planID uuid.UUID) ([]models.GenerationLog, error) {
	query := `
		SELECT id, plan_id, user_id, action, section_type, status,
			prompt_tokens, completion_tokens, model, duration_ms,
			error_message, metadata, created_at
		FROM generation_log WHERE plan_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to list generation logs by plan: %w", err)
	}
	defer rows.Close()

	var logs []models.GenerationLog
	for rows.Next() {
		var l models.GenerationLog
		err := rows.Scan(
			&l.ID, &l.PlanID, &l.UserID, &l.Action, &l.SectionType, &l.Status,
			&l.PromptTokens, &l.CompletionTokens, &l.Model, &l.DurationMs,
			&l.ErrorMessage, &l.Metadata, &l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generation log: %w", err)
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (r *GenerationLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]models.GenerationLog, error) {
	query := `
		SELECT id, plan_id, user_id, action, section_type, status,
			prompt_tokens, completion_tokens, model, duration_ms,
			error_message, metadata, created_at
		FROM generation_log WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list generation logs by user: %w", err)
	}
	defer rows.Close()

	var logs []models.GenerationLog
	for rows.Next() {
		var l models.GenerationLog
		err := rows.Scan(
			&l.ID, &l.PlanID, &l.UserID, &l.Action, &l.SectionType, &l.Status,
			&l.PromptTokens, &l.CompletionTokens, &l.Model, &l.DurationMs,
			&l.ErrorMessage, &l.Metadata, &l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generation log: %w", err)
		}
		logs = append(logs, l)
	}

	return logs, nil
}
