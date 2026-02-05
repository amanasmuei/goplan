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

type AcknowledgmentRepository struct {
	db *pgxpool.Pool
}

func NewAcknowledgmentRepository(db *pgxpool.Pool) *AcknowledgmentRepository {
	return &AcknowledgmentRepository{db: db}
}

func (r *AcknowledgmentRepository) Create(ctx context.Context, ack *models.TaskAcknowledgment) error {
	ack.ID = uuid.New()
	ack.CreatedAt = time.Now()

	query := `
		INSERT INTO task_acknowledgments (
			id, task_id, action, original_estimate, modified_estimate,
			predicted_low, predicted_high, disagreement_notes,
			acknowledged_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(ctx, query,
		ack.ID, ack.TaskID, ack.Action, ack.OriginalEstimate,
		ack.ModifiedEstimate, ack.PredictedLow, ack.PredictedHigh,
		ack.DisagreementNotes, ack.AcknowledgedBy, ack.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create acknowledgment: %w", err)
	}
	return nil
}

func (r *AcknowledgmentRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*models.TaskAcknowledgment, error) {
	query := `
		SELECT id, task_id, action, original_estimate, modified_estimate,
			   predicted_low, predicted_high, disagreement_notes,
			   acknowledged_by, created_at
		FROM task_acknowledgments
		WHERE task_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	ack := &models.TaskAcknowledgment{}
	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&ack.ID, &ack.TaskID, &ack.Action, &ack.OriginalEstimate,
		&ack.ModifiedEstimate, &ack.PredictedLow, &ack.PredictedHigh,
		&ack.DisagreementNotes, &ack.AcknowledgedBy, &ack.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get acknowledgment: %w", err)
	}
	return ack, nil
}

func (r *AcknowledgmentRepository) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]models.TaskAcknowledgment, error) {
	query := `
		SELECT id, task_id, action, original_estimate, modified_estimate,
			   predicted_low, predicted_high, disagreement_notes,
			   acknowledged_by, created_at
		FROM task_acknowledgments
		WHERE task_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list acknowledgments: %w", err)
	}
	defer rows.Close()

	var acks []models.TaskAcknowledgment
	for rows.Next() {
		var ack models.TaskAcknowledgment
		err := rows.Scan(
			&ack.ID, &ack.TaskID, &ack.Action, &ack.OriginalEstimate,
			&ack.ModifiedEstimate, &ack.PredictedLow, &ack.PredictedHigh,
			&ack.DisagreementNotes, &ack.AcknowledgedBy, &ack.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan acknowledgment: %w", err)
		}
		acks = append(acks, ack)
	}
	return acks, nil
}
