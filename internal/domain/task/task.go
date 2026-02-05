// Package task provides the Task domain entity and related types.
package task

import (
	"time"

	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/user"
)

// Task represents a task in the system.
type Task struct {
	ID                string                 `json:"id"`
	PlanID            string                 `json:"planId"`
	PhaseID           *string                `json:"phaseId,omitempty"`
	ParentID          *string                `json:"parentId,omitempty"`
	Title             string                 `json:"title"`
	Description       *string                `json:"description,omitempty"`
	Status            string                 `json:"status"`
	Priority          string                 `json:"priority"`
	AssigneeID        *string                `json:"assigneeId,omitempty"`
	DueDate           *string                `json:"dueDate,omitempty"`
	EstimatedHours    *float64               `json:"estimatedHours,omitempty"`
	CustomFieldValues map[string]interface{} `json:"customFieldValues"`
	Tags              []string               `json:"tags"`
	Position          int                    `json:"position"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`

	// Joined fields (not stored directly in tasks table)
	Subtasks     []Task           `json:"subtasks,omitempty"`
	Dependencies []TaskDependency `json:"dependencies,omitempty"`
	Comments     []Comment        `json:"comments,omitempty"`
}

// TaskDependency represents a dependency between tasks.
type TaskDependency struct {
	TaskID      string `json:"taskId"`
	DependsOnID string `json:"dependsOnId"`
	Type        string `json:"type"`
}

// Comment represents a comment on a task.
type Comment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"taskId"`
	UserID    string    `json:"userId"`
	Content   string    `json:"content"`
	Mentions  []string  `json:"mentions"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Joined field
	User *user.User `json:"user,omitempty"`
}

// ActivityLog represents an activity log entry.
type ActivityLog struct {
	ID          string                 `json:"id"`
	WorkspaceID string                 `json:"workspaceId"`
	PlanID      *string                `json:"planId,omitempty"`
	TaskID      *string                `json:"taskId,omitempty"`
	UserID      string                 `json:"userId"`
	Action      string                 `json:"action"`
	Details     map[string]interface{} `json:"details"`
	CreatedAt   time.Time              `json:"createdAt"`
}

// Validate validates the task fields.
func (t *Task) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	// Validate title
	if t.Title == "" {
		errs.Add("title", "title is required")
	} else if len(t.Title) > 500 {
		errs.Add("title", "title must be at most 500 characters")
	}

	// Validate plan ID
	if t.PlanID == "" {
		errs.Add("planId", "plan ID is required")
	}

	// Validate priority
	if !shared.IsValidTaskPriority(t.Priority) {
		errs.Add("priority", "invalid task priority")
	}

	// Validate status (basic check - actual validation against plan statuses happens at service level)
	if t.Status == "" {
		errs.Add("status", "status is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewTask creates a new Task with the given parameters.
func NewTask(id, planID, title, status string, priority string) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:                id,
		PlanID:            planID,
		Title:             title,
		Status:            status,
		Priority:          priority,
		CustomFieldValues: make(map[string]interface{}),
		Tags:              []string{},
		Position:          0,
		CreatedAt:         now,
		UpdatedAt:         now,
		Subtasks:          []Task{},
		Dependencies:      []TaskDependency{},
		Comments:          []Comment{},
	}
}

// IsSubtask returns true if the task is a subtask.
func (t *Task) IsSubtask() bool {
	return t.ParentID != nil
}

// HasSubtasks returns true if the task has subtasks.
func (t *Task) HasSubtasks() bool {
	return len(t.Subtasks) > 0
}

// Validate validates the comment fields.
func (c *Comment) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	if c.TaskID == "" {
		errs.Add("taskId", "task ID is required")
	}

	if c.UserID == "" {
		errs.Add("userId", "user ID is required")
	}

	if c.Content == "" {
		errs.Add("content", "content is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewComment creates a new Comment with the given parameters.
func NewComment(id, taskID, userID, content string, mentions []string) *Comment {
	now := time.Now().UTC()
	if mentions == nil {
		mentions = []string{}
	}
	return &Comment{
		ID:        id,
		TaskID:    taskID,
		UserID:    userID,
		Content:   content,
		Mentions:  mentions,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates the activity log fields.
func (a *ActivityLog) Validate() *shared.ValidationErrors {
	errs := &shared.ValidationErrors{}

	if a.WorkspaceID == "" {
		errs.Add("workspaceId", "workspace ID is required")
	}

	if a.UserID == "" {
		errs.Add("userId", "user ID is required")
	}

	if a.Action == "" {
		errs.Add("action", "action is required")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// NewActivityLog creates a new ActivityLog entry.
func NewActivityLog(id, workspaceID, userID, action string, planID, taskID *string, details map[string]interface{}) *ActivityLog {
	if details == nil {
		details = make(map[string]interface{})
	}
	return &ActivityLog{
		ID:          id,
		WorkspaceID: workspaceID,
		PlanID:      planID,
		TaskID:      taskID,
		UserID:      userID,
		Action:      action,
		Details:     details,
		CreatedAt:   time.Now().UTC(),
	}
}

// CreateTaskInput represents the input for creating a new task.
type CreateTaskInput struct {
	Title             string                 `json:"title"`
	Description       *string                `json:"description,omitempty"`
	PhaseID           *string                `json:"phaseId,omitempty"`
	ParentID          *string                `json:"parentId,omitempty"`
	Priority          string                 `json:"priority"`
	AssigneeID        *string                `json:"assigneeId,omitempty"`
	DueDate           *string                `json:"dueDate,omitempty"`
	EstimatedHours    *float64               `json:"estimatedHours,omitempty"`
	CustomFieldValues map[string]interface{} `json:"customFieldValues,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
}

// UpdateTaskInput represents the input for updating a task.
type UpdateTaskInput struct {
	Title             *string                `json:"title,omitempty"`
	Description       *string                `json:"description,omitempty"`
	Status            *string                `json:"status,omitempty"`
	Priority          *string                `json:"priority,omitempty"`
	PhaseID           *string                `json:"phaseId,omitempty"`
	AssigneeID        *string                `json:"assigneeId,omitempty"`
	DueDate           *string                `json:"dueDate,omitempty"`
	EstimatedHours    *float64               `json:"estimatedHours,omitempty"`
	CustomFieldValues map[string]interface{} `json:"customFieldValues,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	Position          *int                   `json:"position,omitempty"`
}

// MoveTaskInput represents the input for moving a task.
type MoveTaskInput struct {
	Status   string `json:"status"`
	Position int    `json:"position"`
}

// TaskResponse represents the task response for API.
type TaskResponse struct {
	ID                string                 `json:"id"`
	PlanID            string                 `json:"planId"`
	PhaseID           *string                `json:"phaseId,omitempty"`
	ParentID          *string                `json:"parentId,omitempty"`
	Title             string                 `json:"title"`
	Description       *string                `json:"description,omitempty"`
	Status            string                 `json:"status"`
	Priority          string                 `json:"priority"`
	AssigneeID        *string                `json:"assigneeId,omitempty"`
	DueDate           *string                `json:"dueDate,omitempty"`
	EstimatedHours    *float64               `json:"estimatedHours,omitempty"`
	CustomFieldValues map[string]interface{} `json:"customFieldValues"`
	Tags              []string               `json:"tags"`
	Position          int                    `json:"position"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
	Subtasks          []TaskResponse         `json:"subtasks,omitempty"`
	Dependencies      []TaskDependency       `json:"dependencies,omitempty"`
	Comments          []CommentResponse      `json:"comments,omitempty"`
}

// ToResponse converts a Task to TaskResponse with ISO 8601 timestamps.
func (t *Task) ToResponse() *TaskResponse {
	resp := &TaskResponse{
		ID:                t.ID,
		PlanID:            t.PlanID,
		PhaseID:           t.PhaseID,
		ParentID:          t.ParentID,
		Title:             t.Title,
		Description:       t.Description,
		Status:            t.Status,
		Priority:          t.Priority,
		AssigneeID:        t.AssigneeID,
		DueDate:           t.DueDate,
		EstimatedHours:    t.EstimatedHours,
		CustomFieldValues: t.CustomFieldValues,
		Tags:              t.Tags,
		Position:          t.Position,
		CreatedAt:         t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         t.UpdatedAt.Format(time.RFC3339),
		Dependencies:      t.Dependencies,
	}

	// Convert subtasks
	if len(t.Subtasks) > 0 {
		resp.Subtasks = make([]TaskResponse, len(t.Subtasks))
		for i, st := range t.Subtasks {
			resp.Subtasks[i] = *st.ToResponse()
		}
	}

	// Convert comments
	if len(t.Comments) > 0 {
		resp.Comments = make([]CommentResponse, len(t.Comments))
		for i, c := range t.Comments {
			resp.Comments[i] = *c.ToResponse()
		}
	}

	return resp
}

// CommentResponse represents the comment response for API.
type CommentResponse struct {
	ID        string             `json:"id"`
	TaskID    string             `json:"taskId"`
	UserID    string             `json:"userId"`
	Content   string             `json:"content"`
	Mentions  []string           `json:"mentions"`
	CreatedAt string             `json:"createdAt"`
	UpdatedAt string             `json:"updatedAt"`
	User      *user.UserResponse `json:"user,omitempty"`
}

// ToResponse converts a Comment to CommentResponse.
func (c *Comment) ToResponse() *CommentResponse {
	resp := &CommentResponse{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Content:   c.Content,
		Mentions:  c.Mentions,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}
	if c.User != nil {
		resp.User = c.User.ToResponse()
	}
	return resp
}

// ActivityLogResponse represents the activity log response for API.
type ActivityLogResponse struct {
	ID          string                 `json:"id"`
	WorkspaceID string                 `json:"workspaceId"`
	PlanID      *string                `json:"planId,omitempty"`
	TaskID      *string                `json:"taskId,omitempty"`
	UserID      string                 `json:"userId"`
	Action      string                 `json:"action"`
	Details     map[string]interface{} `json:"details"`
	CreatedAt   string                 `json:"createdAt"`
}

// ToResponse converts an ActivityLog to ActivityLogResponse.
func (a *ActivityLog) ToResponse() *ActivityLogResponse {
	return &ActivityLogResponse{
		ID:          a.ID,
		WorkspaceID: a.WorkspaceID,
		PlanID:      a.PlanID,
		TaskID:      a.TaskID,
		UserID:      a.UserID,
		Action:      a.Action,
		Details:     a.Details,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
	}
}

// TaskSummary represents a summary of a task for MCP context.
type TaskSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

// ToSummary converts a Task to TaskSummary.
func (t *Task) ToSummary() *TaskSummary {
	return &TaskSummary{
		ID:       t.ID,
		Title:    t.Title,
		Status:   t.Status,
		Priority: t.Priority,
	}
}
