package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type TaskStatus string

const (
	TaskStatusDraft               TaskStatus = "draft"
	TaskStatusPendingAcknowledgment TaskStatus = "pending_acknowledgment"
	TaskStatusAcknowledged        TaskStatus = "acknowledged"
	TaskStatusActive              TaskStatus = "active"
	TaskStatusBlocked             TaskStatus = "blocked"
	TaskStatusPendingReview       TaskStatus = "pending_review"
	TaskStatusCompleted           TaskStatus = "completed"
	TaskStatusCancelled           TaskStatus = "cancelled"
)

type Task struct {
	ID                    uuid.UUID       `json:"id" db:"id"`
	Title                 string          `json:"title" db:"title"`
	Description           string          `json:"description" db:"description"`
	DescriptionEmbedding  pgvector.Vector `json:"-" db:"description_embedding"`
	Status                TaskStatus      `json:"status" db:"status"`
	CreatedBy             uuid.UUID       `json:"created_by" db:"created_by"`
	AssignedTo            *uuid.UUID      `json:"assigned_to,omitempty" db:"assigned_to"`
	ProjectID             uuid.UUID       `json:"project_id" db:"project_id"`
	OrganizationID        uuid.UUID       `json:"organization_id" db:"organization_id"`
	EstimatedDays         *float64        `json:"estimated_days,omitempty" db:"estimated_days"`
	PredictedDaysLow      *float64        `json:"predicted_days_low,omitempty" db:"predicted_days_low"`
	PredictedDaysHigh     *float64        `json:"predicted_days_high,omitempty" db:"predicted_days_high"`
	PredictionConfidence  *float64        `json:"prediction_confidence,omitempty" db:"prediction_confidence"`
	PlanningQualityScore  *float64        `json:"planning_quality_score,omitempty" db:"planning_quality_score"`
	AcknowledgedAt        *time.Time      `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	StartedAt             *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt           *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	ActualDays            *float64        `json:"actual_days,omitempty" db:"actual_days"`
	Tags                  []string        `json:"tags,omitempty" db:"tags"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
}

type CreateTaskRequest struct {
	Title         string    `json:"title" validate:"required,min=5,max=500"`
	Description   string    `json:"description" validate:"required,min=50"`
	ProjectID     uuid.UUID `json:"project_id" validate:"required"`
	EstimatedDays *float64  `json:"estimated_days" validate:"omitempty,gt=0"`
	AssignedTo    *uuid.UUID `json:"assigned_to"`
	Tags          []string  `json:"tags"`
}

type UpdateTaskRequest struct {
	Title         *string    `json:"title" validate:"omitempty,min=5,max=500"`
	Description   *string    `json:"description" validate:"omitempty,min=50"`
	EstimatedDays *float64   `json:"estimated_days" validate:"omitempty,gt=0"`
	AssignedTo    *uuid.UUID `json:"assigned_to"`
	Tags          []string   `json:"tags"`
	Status        *TaskStatus `json:"status"`
}

type TaskResponse struct {
	Task               *Task           `json:"task"`
	SimilarTasks       []SimilarTask   `json:"similar_tasks,omitempty"`
	Predictions        *Predictions    `json:"predictions,omitempty"`
	PlanningAssessment *Assessment     `json:"planning_assessment,omitempty"`
}

type SimilarTask struct {
	ID                   uuid.UUID  `json:"id"`
	Title                string     `json:"title"`
	Status               TaskStatus `json:"status"`
	SimilarityScore      float64    `json:"similarity_score"`
	ActualDays           *float64   `json:"actual_days,omitempty"`
	EstimatedDays        *float64   `json:"estimated_days,omitempty"`
	BlockersSummary      string     `json:"blockers_summary,omitempty"`
	LessonsLearnedExcerpt string    `json:"lessons_learned_excerpt,omitempty"`
}

type Predictions struct {
	PredictedDaysLow     float64       `json:"predicted_days_low"`
	PredictedDaysHigh    float64       `json:"predicted_days_high"`
	Confidence           float64       `json:"confidence"`
	BlockerRisks         []BlockerRisk `json:"blocker_risks"`
}

type BlockerRisk struct {
	Type        string  `json:"type"`
	Probability float64 `json:"probability"`
	Examples    []string `json:"examples,omitempty"`
}

type Assessment struct {
	Score       float64             `json:"score"`
	Breakdown   map[string]int      `json:"breakdown"`
	Suggestions []string            `json:"suggestions"`
}

type TaskFilters struct {
	ProjectID      *uuid.UUID
	Status         *TaskStatus
	AssignedTo     *uuid.UUID
	CreatedBy      *uuid.UUID
	OrganizationID uuid.UUID
	Search         string
	Page           int
	PageSize       int
}

type TaskListResponse struct {
	Tasks      []Task `json:"tasks"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}
