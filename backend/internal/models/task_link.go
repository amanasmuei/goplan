package models

import (
	"time"

	"github.com/google/uuid"
)

type LinkType string

const (
	LinkTypeSimilar   LinkType = "similar"
	LinkTypeDependent LinkType = "dependent"
	LinkTypeRetry     LinkType = "retry"
	LinkTypeRelated   LinkType = "related"
)

type TaskLink struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SourceTaskID uuid.UUID `json:"source_task_id" db:"source_task_id"`
	TargetTaskID uuid.UUID `json:"target_task_id" db:"target_task_id"`
	LinkType     LinkType  `json:"link_type" db:"link_type"`
	CreatedBy    uuid.UUID `json:"created_by" db:"created_by"`
	Notes        string    `json:"notes,omitempty" db:"notes"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type CreateTaskLinkRequest struct {
	TargetTaskID uuid.UUID `json:"target_task_id" validate:"required"`
	LinkType     LinkType  `json:"link_type" validate:"required,oneof=similar dependent retry related"`
	Notes        string    `json:"notes"`
}

type TaskLinkResponse struct {
	Link       TaskLink `json:"link"`
	TargetTask *Task    `json:"target_task,omitempty"`
}
