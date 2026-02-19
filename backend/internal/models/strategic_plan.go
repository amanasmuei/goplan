package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type PlanStatus string

const (
	PlanStatusDraft      PlanStatus = "draft"
	PlanStatusGenerating PlanStatus = "generating"
	PlanStatusComplete   PlanStatus = "complete"
	PlanStatusArchived   PlanStatus = "archived"
)

type PlanCategory string

const (
	PlanCategoryBusiness   PlanCategory = "business"
	PlanCategorySaaS       PlanCategory = "saas"
	PlanCategoryEvent      PlanCategory = "event"
	PlanCategoryNonprofit  PlanCategory = "nonprofit"
	PlanCategoryPersonal   PlanCategory = "personal"
	PlanCategoryEducation  PlanCategory = "education"
	PlanCategoryRealEstate PlanCategory = "real_estate"
	PlanCategoryGeneric    PlanCategory = "generic"
)

type StrategicPlan struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	UserID           uuid.UUID       `json:"user_id" db:"user_id"`
	OrganizationID   uuid.UUID       `json:"organization_id" db:"organization_id"`
	Title            string          `json:"title" db:"title"`
	OriginalInput    string          `json:"original_input" db:"original_input"`
	Category         PlanCategory    `json:"category" db:"category"`
	SubCategory      *string         `json:"sub_category,omitempty" db:"sub_category"`
	Complexity       *string         `json:"complexity,omitempty" db:"complexity"`
	Status           PlanStatus      `json:"status" db:"status"`
	CurrentVersion   int             `json:"current_version" db:"current_version"`
	ContentEmbedding pgvector.Vector `json:"-" db:"content_embedding"`
	Metadata         map[string]any  `json:"metadata,omitempty" db:"metadata"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at" db:"updated_at"`
}

type CreatePlanRequest struct {
	Input    string        `json:"input" validate:"required,min=20,max=5000"`
	Category *PlanCategory `json:"category,omitempty" validate:"omitempty,oneof=business saas event nonprofit personal education real_estate generic"`
}

type PlanResponse struct {
	Plan     StrategicPlan `json:"plan"`
	Sections []PlanSection `json:"sections"`
}

type PlanListResponse struct {
	Plans      []StrategicPlan `json:"plans"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

type PlanFilters struct {
	UserID         *uuid.UUID
	OrganizationID *uuid.UUID
	Status         *PlanStatus
	Category       *PlanCategory
	Search         string
	Page           int
	PageSize       int
}
