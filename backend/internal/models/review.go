package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskReview struct {
	ID                       uuid.UUID `json:"id" db:"id"`
	TaskID                   uuid.UUID `json:"task_id" db:"task_id"`
	PredictionAccuracyRating int       `json:"prediction_accuracy_rating" db:"prediction_accuracy_rating"`
	PredictionFeedback       string    `json:"prediction_feedback,omitempty" db:"prediction_feedback"`
	LessonsLearned           string    `json:"lessons_learned,omitempty" db:"lessons_learned"`
	WouldApproachDifferently string    `json:"would_approach_differently,omitempty" db:"would_approach_differently"`
	CreatedBy                uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
}

type CreateReviewRequest struct {
	PredictionAccuracyRating int    `json:"prediction_accuracy_rating" validate:"required,min=1,max=5"`
	PredictionFeedback       string `json:"prediction_feedback"`
	LessonsLearned           string `json:"lessons_learned"`
	WouldApproachDifferently string `json:"would_approach_differently"`
}
