package models

import (
	"time"

	"github.com/google/uuid"
)

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

// Project represents a project within an organization
type Project struct {
	ID             uuid.UUID     `json:"id" db:"id"`
	Name           string        `json:"name" db:"name"`
	Description    string        `json:"description" db:"description"`
	Status         ProjectStatus `json:"status" db:"status"`
	OrganizationID uuid.UUID     `json:"organization_id" db:"organization_id"`
	CreatedBy      *uuid.UUID    `json:"created_by,omitempty" db:"created_by"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at" db:"updated_at"`
}

// ProjectTeam represents the association between a project and a team
type ProjectTeam struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ProjectID  uuid.UUID `json:"project_id" db:"project_id"`
	TeamID     uuid.UUID `json:"team_id" db:"team_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}

// CreateProjectRequest is the request body for creating a new project
type CreateProjectRequest struct {
	Name        string      `json:"name" validate:"required,min=2,max=255"`
	Description string      `json:"description" validate:"omitempty,max=2000"`
	TeamIDs     []uuid.UUID `json:"team_ids" validate:"omitempty"`
}

// UpdateProjectRequest is the request body for updating a project
type UpdateProjectRequest struct {
	Name        *string        `json:"name" validate:"omitempty,min=2,max=255"`
	Description *string        `json:"description" validate:"omitempty,max=2000"`
	Status      *ProjectStatus `json:"status" validate:"omitempty,oneof=active archived"`
}

// AssignTeamsRequest is the request body for assigning teams to a project
type AssignTeamsRequest struct {
	TeamIDs []uuid.UUID `json:"team_ids" validate:"required,min=1"`
}

// ProjectResponse is the response for a single project with additional data
type ProjectResponse struct {
	Project   *Project `json:"project"`
	Teams     []Team   `json:"teams,omitempty"`
	TaskCount int      `json:"task_count"`
}

// ProjectListResponse is the response for listing projects
type ProjectListResponse struct {
	Projects   []ProjectResponse `json:"projects"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// ProjectFilters defines the filters for listing projects
type ProjectFilters struct {
	OrganizationID uuid.UUID
	TeamID         *uuid.UUID
	Status         *ProjectStatus
	Search         string
	Page           int
	PageSize       int
}

// IsValid returns true if the project status is a valid enum value
func (s ProjectStatus) IsValid() bool {
	switch s {
	case ProjectStatusActive, ProjectStatusArchived:
		return true
	default:
		return false
	}
}
