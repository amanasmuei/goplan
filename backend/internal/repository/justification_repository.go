package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/goplan/backend/internal/models"
)

type JustificationRepository struct {
	db *pgxpool.Pool
}

func NewJustificationRepository(db *pgxpool.Pool) *JustificationRepository {
	return &JustificationRepository{db: db}
}

func (r *JustificationRepository) Create(ctx context.Context, j *models.TaskJustification) error {
	j.ID = uuid.New()
	j.CreatedAt = time.Now()

	query := `
		INSERT INTO task_justifications (
			id, task_id, checked_same_project, checked_same_stakeholders,
			checked_same_dependencies, justification_text, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		j.ID, j.TaskID, j.CheckedSameProject, j.CheckedSameStakeholders,
		j.CheckedSameDependencies, j.JustificationText, j.CreatedBy, j.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create justification: %w", err)
	}
	return nil
}

func (r *JustificationRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*models.TaskJustification, error) {
	query := `
		SELECT id, task_id, checked_same_project, checked_same_stakeholders,
			checked_same_dependencies, justification_text, created_by, created_at
		FROM task_justifications WHERE task_id = $1`

	j := &models.TaskJustification{}
	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&j.ID, &j.TaskID, &j.CheckedSameProject, &j.CheckedSameStakeholders,
		&j.CheckedSameDependencies, &j.JustificationText, &j.CreatedBy, &j.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get justification: %w", err)
	}
	return j, nil
}

func (r *JustificationRepository) Exists(ctx context.Context, taskID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM task_justifications WHERE task_id = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, taskID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check justification existence: %w", err)
	}
	return exists, nil
}
