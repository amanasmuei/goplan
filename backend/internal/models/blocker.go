package models

import (
	"time"

	"github.com/google/uuid"
)

type BlockerType string

const (
	BlockerTypeApproval     BlockerType = "approval"
	BlockerTypeExternalTeam BlockerType = "external_team"
	BlockerTypeVendor       BlockerType = "vendor"
	BlockerTypeTechnical    BlockerType = "technical"
	BlockerTypeResource     BlockerType = "resource"
	BlockerTypeRequirements BlockerType = "requirements"
)

type TaskBlocker struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	TaskID      uuid.UUID   `json:"task_id" db:"task_id"`
	BlockerType BlockerType `json:"blocker_type" db:"blocker_type"`
	Description string      `json:"description" db:"description"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty" db:"resolved_at"`
	DaysBlocked *float64    `json:"days_blocked,omitempty" db:"days_blocked"`
	CreatedBy   uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}

type CreateBlockerRequest struct {
	BlockerType BlockerType `json:"blocker_type" validate:"required,oneof=approval external_team vendor technical resource requirements"`
	Description string      `json:"description" validate:"required,min=10"`
}

type ResolveBlockerRequest struct {
	DaysBlocked float64 `json:"days_blocked" validate:"required,gt=0"`
}
