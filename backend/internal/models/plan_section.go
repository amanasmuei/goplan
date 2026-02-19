package models

import (
	"time"

	"github.com/google/uuid"
)

type SectionType string

const (
	SectionTypeExecutiveBrief      SectionType = "executive_brief"
	SectionTypeStrategicContext     SectionType = "strategic_context"
	SectionTypeRecommendedApproach SectionType = "recommended_approach"
	SectionTypePhasedExecution     SectionType = "phased_execution"
	SectionTypeImmediateAction     SectionType = "immediate_action"
)

// SectionOrder defines the display order for each section type.
var SectionOrder = map[SectionType]int{
	SectionTypeExecutiveBrief:      1,
	SectionTypeStrategicContext:     2,
	SectionTypeRecommendedApproach: 3,
	SectionTypePhasedExecution:     4,
	SectionTypeImmediateAction:     5,
}

// DefaultSectionTitles provides the default human-readable title for each section type.
var DefaultSectionTitles = map[SectionType]string{
	SectionTypeExecutiveBrief:      "Executive Brief",
	SectionTypeStrategicContext:     "Strategic Context",
	SectionTypeRecommendedApproach: "Recommended Approach",
	SectionTypePhasedExecution:     "Phased Execution",
	SectionTypeImmediateAction:     "Immediate Action",
}

type PlanSection struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	PlanID       uuid.UUID   `json:"plan_id" db:"plan_id"`
	SectionType  SectionType `json:"section_type" db:"section_type"`
	SectionOrder int         `json:"section_order" db:"section_order"`
	Title        string      `json:"title" db:"title"`
	Content      any         `json:"content" db:"content"`
	Version      int         `json:"version" db:"version"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

type RegenerateSectionRequest struct {
	AdditionalContext *string `json:"additional_context,omitempty" validate:"omitempty,max=2000"`
	Depth             *string `json:"depth,omitempty" validate:"omitempty,oneof=standard deep"`
}

type RefineSectionRequest struct {
	RefinementPrompt string `json:"refinement_prompt" validate:"required,min=10,max=2000"`
}
