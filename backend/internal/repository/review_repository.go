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

type ReviewRepository struct {
	db *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, review *models.TaskReview) error {
	review.ID = uuid.New()
	review.CreatedAt = time.Now()

	query := `
		INSERT INTO task_reviews (
			id, task_id, prediction_accuracy_rating, prediction_feedback,
			lessons_learned, would_approach_differently, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query,
		review.ID, review.TaskID, review.PredictionAccuracyRating, review.PredictionFeedback,
		review.LessonsLearned, review.WouldApproachDifferently, review.CreatedBy, review.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}
	return nil
}

func (r *ReviewRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID) (*models.TaskReview, error) {
	query := `
		SELECT id, task_id, prediction_accuracy_rating, prediction_feedback,
			lessons_learned, would_approach_differently, created_by, created_at
		FROM task_reviews WHERE task_id = $1`

	review := &models.TaskReview{}
	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&review.ID, &review.TaskID, &review.PredictionAccuracyRating, &review.PredictionFeedback,
		&review.LessonsLearned, &review.WouldApproachDifferently, &review.CreatedBy, &review.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}
	return review, nil
}

func (r *ReviewRepository) GetAverageAccuracy(ctx context.Context, orgID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(AVG(tr.prediction_accuracy_rating), 0)
		FROM task_reviews tr
		JOIN tasks t ON tr.task_id = t.id
		WHERE t.organization_id = $1`

	var avg float64
	err := r.db.QueryRow(ctx, query, orgID).Scan(&avg)
	if err != nil {
		return 0, fmt.Errorf("failed to get average accuracy: %w", err)
	}
	return avg, nil
}
