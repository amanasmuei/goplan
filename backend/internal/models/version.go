package models

import (
	"time"

	"github.com/google/uuid"
)

type SectionVersion struct {
	ID                uuid.UUID `json:"id" db:"id"`
	SectionID         uuid.UUID `json:"section_id" db:"section_id"`
	PlanID            uuid.UUID `json:"plan_id" db:"plan_id"`
	Version           int       `json:"version" db:"version"`
	Content           any       `json:"content" db:"content"`
	RefinementContext *string   `json:"refinement_context,omitempty" db:"refinement_context"`
	GeneratedBy       string    `json:"generated_by" db:"generated_by"`
	TokenUsage        any       `json:"token_usage,omitempty" db:"token_usage"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type PlanVersion struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PlanID        uuid.UUID `json:"plan_id" db:"plan_id"`
	Version       int       `json:"version" db:"version"`
	Snapshot      any       `json:"snapshot" db:"snapshot"`
	ChangeSummary *string   `json:"change_summary,omitempty" db:"change_summary"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type GenerationLog struct {
	ID               uuid.UUID    `json:"id" db:"id"`
	PlanID           *uuid.UUID   `json:"plan_id,omitempty" db:"plan_id"`
	UserID           uuid.UUID    `json:"user_id" db:"user_id"`
	Action           string       `json:"action" db:"action"`
	SectionType      *SectionType `json:"section_type,omitempty" db:"section_type"`
	Status           string       `json:"status" db:"status"`
	PromptTokens     *int         `json:"prompt_tokens,omitempty" db:"prompt_tokens"`
	CompletionTokens *int         `json:"completion_tokens,omitempty" db:"completion_tokens"`
	Model            *string      `json:"model,omitempty" db:"model"`
	DurationMs       *int         `json:"duration_ms,omitempty" db:"duration_ms"`
	ErrorMessage     *string      `json:"error_message,omitempty" db:"error_message"`
	Metadata         any          `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
}
