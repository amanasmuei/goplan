package models

import (
	"time"

	"github.com/google/uuid"
)

// TaskAcknowledgment records how a user acknowledged task predictions
type TaskAcknowledgment struct {
	ID                uuid.UUID            `json:"id" db:"id"`
	TaskID            uuid.UUID            `json:"task_id" db:"task_id"`
	Action            AcknowledgmentAction `json:"action" db:"action"`
	OriginalEstimate  *float64             `json:"original_estimate,omitempty" db:"original_estimate"`
	ModifiedEstimate  *float64             `json:"modified_estimate,omitempty" db:"modified_estimate"`
	PredictedLow      *float64             `json:"predicted_low,omitempty" db:"predicted_low"`
	PredictedHigh     *float64             `json:"predicted_high,omitempty" db:"predicted_high"`
	DisagreementNotes string               `json:"disagreement_notes,omitempty" db:"disagreement_notes"`
	AcknowledgedBy    uuid.UUID            `json:"acknowledged_by" db:"acknowledged_by"`
	CreatedAt         time.Time            `json:"created_at" db:"created_at"`
}
