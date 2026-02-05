// Package docs GoPlan API.
//
// Documentation for GoPlan API - An intelligent task management system
// that prevents repeated failures through contextual awareness.
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /api/v1
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	SecurityDefinitions:
//	  Bearer:
//	    type: apiKey
//	    name: Authorization
//	    in: header
//	    description: JWT Authorization header using the Bearer scheme. Example "Bearer {token}"
//
// swagger:meta
package docs

// swagger:model Task
type TaskResponse struct {
	// The unique identifier of the task
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// The task title
	// required: true
	// example: Implement user authentication
	Title string `json:"title"`

	// Detailed description of the task
	// required: true
	Description string `json:"description"`

	// Current status of the task
	// enum: draft,pending_acknowledgment,acknowledged,active,blocked,pending_review,completed,cancelled
	// example: active
	Status string `json:"status"`

	// User estimated days to complete
	// example: 5
	EstimatedDays *float64 `json:"estimated_days,omitempty"`

	// System predicted minimum days
	// example: 4
	PredictedDaysLow *float64 `json:"predicted_days_low,omitempty"`

	// System predicted maximum days
	// example: 7
	PredictedDaysHigh *float64 `json:"predicted_days_high,omitempty"`

	// Confidence level of prediction (0-1)
	// example: 0.85
	PredictionConfidence *float64 `json:"prediction_confidence,omitempty"`

	// Planning quality score (0-100)
	// example: 75
	PlanningQualityScore *int `json:"planning_quality_score,omitempty"`

	// Actual days taken to complete
	// example: 6
	ActualDays *float64 `json:"actual_days,omitempty"`

	// When the task was created
	// example: 2024-01-15T10:30:00Z
	CreatedAt string `json:"created_at"`

	// When the task was started
	// example: 2024-01-16T09:00:00Z
	StartedAt *string `json:"started_at,omitempty"`

	// When the task was completed
	// example: 2024-01-22T17:00:00Z
	CompletedAt *string `json:"completed_at,omitempty"`
}

// swagger:model CreateTaskRequest
type CreateTaskRequest struct {
	// The task title
	// required: true
	// min length: 5
	// max length: 500
	// example: Implement OAuth2 authentication
	Title string `json:"title"`

	// Detailed description including objectives, risks, and acceptance criteria
	// required: true
	// min length: 50
	// example: Implement OAuth2 authentication flow with Google provider. Dependencies: existing user model. Risks: API rate limits. Acceptance: users can login via Google.
	Description string `json:"description"`

	// Project ID to associate the task with
	// required: true
	// example: 550e8400-e29b-41d4-a716-446655440001
	ProjectID string `json:"project_id"`

	// Estimated days to complete
	// minimum: 0.5
	// example: 5
	EstimatedDays *float64 `json:"estimated_days,omitempty"`

	// Tags for categorization
	// example: ["backend", "auth"]
	Tags []string `json:"tags,omitempty"`
}

// swagger:model TaskListResponse
type TaskListResponse struct {
	// List of tasks
	Tasks []TaskResponse `json:"tasks"`

	// Total number of tasks matching filter
	// example: 42
	Total int64 `json:"total"`

	// Current page number
	// example: 1
	Page int `json:"page"`

	// Items per page
	// example: 10
	PageSize int `json:"page_size"`

	// Total number of pages
	// example: 5
	TotalPages int `json:"total_pages"`
}

// swagger:model SimilarTask
type SimilarTask struct {
	// Task ID
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// Task title
	// example: Implement SSO integration
	Title string `json:"title"`

	// Task status
	// example: completed
	Status string `json:"status"`

	// Similarity score (0-1)
	// example: 0.87
	SimilarityScore float64 `json:"similarity_score"`

	// Days the task actually took
	// example: 6
	ActualDays *float64 `json:"actual_days,omitempty"`

	// Originally estimated days
	// example: 5
	EstimatedDays *float64 `json:"estimated_days,omitempty"`

	// Summary of blockers encountered
	// example: dependency_delay, technical_debt
	BlockersSummary string `json:"blockers_summary,omitempty"`

	// Key lessons from the task
	// example: OAuth implementation requires careful state management
	LessonsLearnedExcerpt string `json:"lessons_learned_excerpt,omitempty"`
}

// swagger:model TaskLink
type TaskLink struct {
	// Link ID
	// example: 550e8400-e29b-41d4-a716-446655440002
	ID string `json:"id"`

	// Source task ID
	// example: 550e8400-e29b-41d4-a716-446655440000
	SourceTaskID string `json:"source_task_id"`

	// Target task ID
	// example: 550e8400-e29b-41d4-a716-446655440001
	TargetTaskID string `json:"target_task_id"`

	// Type of link
	// enum: similar,dependent,retry,related
	// example: similar
	LinkType string `json:"link_type"`

	// When the link was created
	// example: 2024-01-15T10:30:00Z
	CreatedAt string `json:"created_at"`
}

// swagger:model Blocker
type Blocker struct {
	// Blocker ID
	// example: 550e8400-e29b-41d4-a716-446655440003
	ID string `json:"id"`

	// Task ID this blocker affects
	// example: 550e8400-e29b-41d4-a716-446655440000
	TaskID string `json:"task_id"`

	// Type of blocker
	// enum: technical_debt,dependency_delay,unclear_requirements,resource_unavailable,external_blocker
	// example: dependency_delay
	BlockerType string `json:"blocker_type"`

	// Description of the blocker
	// example: Waiting for API team to complete endpoint implementation
	Description string `json:"description"`

	// When the blocker was reported
	// example: 2024-01-17T14:00:00Z
	ReportedAt string `json:"reported_at"`

	// When the blocker was resolved (if resolved)
	// example: 2024-01-19T10:00:00Z
	ResolvedAt *string `json:"resolved_at,omitempty"`

	// Total days blocked
	// example: 2
	DaysBlocked *int `json:"days_blocked,omitempty"`
}

// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message
	// example: Validation failed
	Error string `json:"error"`

	// Detailed error information
	// example: title must be at least 5 characters
	Details string `json:"details,omitempty"`
}

// swagger:model TaskStats
type TaskStats struct {
	// Total number of tasks
	// example: 156
	Total int `json:"total"`

	// Count by status
	ByStatus map[string]int `json:"by_status"`

	// Average days to complete
	// example: 4.5
	AvgDaysToComplete *float64 `json:"avg_days_to_complete,omitempty"`

	// Prediction accuracy percentage
	// example: 78.5
	PredictionAccuracy *float64 `json:"prediction_accuracy,omitempty"`
}

// swagger:model Team
type Team struct {
	// The unique identifier of the team
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// The team name
	// required: true
	// example: Engineering
	Name string `json:"name"`

	// Team description
	// example: Core engineering team
	Description string `json:"description,omitempty"`

	// Organization this team belongs to
	// example: 550e8400-e29b-41d4-a716-446655440001
	OrganizationID string `json:"organization_id"`

	// User who created the team
	// example: 550e8400-e29b-41d4-a716-446655440002
	CreatedBy string `json:"created_by"`

	// When the team was created
	// example: 2024-01-15T10:30:00Z
	CreatedAt string `json:"created_at"`

	// When the team was last updated
	// example: 2024-01-15T10:30:00Z
	UpdatedAt string `json:"updated_at"`
}

// swagger:model TeamMember
type TeamMember struct {
	// The unique identifier of the membership
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// The team ID
	// example: 550e8400-e29b-41d4-a716-446655440001
	TeamID string `json:"team_id"`

	// The user ID
	// example: 550e8400-e29b-41d4-a716-446655440002
	UserID string `json:"user_id"`

	// The user's role in the team
	// enum: owner,manager,member,viewer
	// example: member
	Role string `json:"role"`

	// When the user joined the team
	// example: 2024-01-15T10:30:00Z
	JoinedAt string `json:"joined_at"`
}

// swagger:model CreateTeamRequest
type CreateTeamRequest struct {
	// The team name
	// required: true
	// min length: 2
	// max length: 255
	// example: Engineering
	Name string `json:"name"`

	// Team description
	// max length: 1000
	// example: Core engineering team
	Description string `json:"description,omitempty"`
}

// swagger:model AddTeamMemberRequest
type AddTeamMemberRequest struct {
	// The user ID to add
	// required: true
	// example: 550e8400-e29b-41d4-a716-446655440000
	UserID string `json:"user_id"`

	// The role to assign
	// required: true
	// enum: owner,manager,member,viewer
	// example: member
	Role string `json:"role"`
}

// swagger:model TeamResponse
type TeamResponse struct {
	// The team details
	Team Team `json:"team"`

	// Number of members in the team
	// example: 5
	MemberCount int `json:"member_count"`

	// List of team members (included when fetching single team)
	Members []TeamMember `json:"members,omitempty"`
}

// swagger:model TeamListResponse
type TeamListResponse struct {
	// List of teams
	Teams []TeamResponse `json:"teams"`

	// Total number of teams
	// example: 3
	Total int64 `json:"total"`
}

// swagger:model Project
type Project struct {
	// The unique identifier of the project
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// The project name
	// required: true
	// example: GoPlan MVP
	Name string `json:"name"`

	// Project description
	// example: Main product development project
	Description string `json:"description,omitempty"`

	// Project status
	// enum: active,archived
	// example: active
	Status string `json:"status"`

	// Organization this project belongs to
	// example: 550e8400-e29b-41d4-a716-446655440001
	OrganizationID string `json:"organization_id"`

	// User who created the project
	// example: 550e8400-e29b-41d4-a716-446655440002
	CreatedBy string `json:"created_by,omitempty"`

	// When the project was created
	// example: 2024-01-15T10:30:00Z
	CreatedAt string `json:"created_at"`

	// When the project was last updated
	// example: 2024-01-15T10:30:00Z
	UpdatedAt string `json:"updated_at"`
}

// swagger:model CreateProjectRequest
type CreateProjectRequest struct {
	// The project name
	// required: true
	// min length: 2
	// max length: 255
	// example: GoPlan MVP
	Name string `json:"name"`

	// Project description
	// max length: 2000
	// example: Main product development project
	Description string `json:"description,omitempty"`

	// Team IDs to assign to the project
	// example: ["550e8400-e29b-41d4-a716-446655440000"]
	TeamIDs []string `json:"team_ids,omitempty"`
}

// swagger:model AssignTeamsRequest
type AssignTeamsRequest struct {
	// Team IDs to assign to the project
	// required: true
	// example: ["550e8400-e29b-41d4-a716-446655440000"]
	TeamIDs []string `json:"team_ids"`
}

// swagger:model ProjectResponse
type ProjectResponse struct {
	// The project details
	Project Project `json:"project"`

	// Teams assigned to this project
	Teams []Team `json:"teams,omitempty"`

	// Number of tasks in the project
	// example: 12
	TaskCount int `json:"task_count"`
}

// swagger:model ProjectListResponse
type ProjectListResponse struct {
	// List of projects
	Projects []ProjectResponse `json:"projects"`

	// Total number of projects matching filter
	// example: 5
	Total int64 `json:"total"`

	// Current page number
	// example: 1
	Page int `json:"page"`

	// Items per page
	// example: 20
	PageSize int `json:"page_size"`

	// Total number of pages
	// example: 1
	TotalPages int `json:"total_pages"`
}
