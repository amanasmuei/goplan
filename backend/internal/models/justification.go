package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskJustification struct {
	ID                      uuid.UUID `json:"id" db:"id"`
	TaskID                  uuid.UUID `json:"task_id" db:"task_id"`
	CheckedSameProject      bool      `json:"checked_same_project" db:"checked_same_project"`
	CheckedSameStakeholders bool      `json:"checked_same_stakeholders" db:"checked_same_stakeholders"`
	CheckedSameDependencies bool      `json:"checked_same_dependencies" db:"checked_same_dependencies"`
	JustificationText       string    `json:"justification_text" db:"justification_text"`
	CreatedBy               uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
}

type CreateJustificationRequest struct {
	CheckedSameProject      bool   `json:"checked_same_project" validate:"required"`
	CheckedSameStakeholders bool   `json:"checked_same_stakeholders" validate:"required"`
	CheckedSameDependencies bool   `json:"checked_same_dependencies" validate:"required"`
	JustificationText       string `json:"justification_text" validate:"required,min=50"`
}
