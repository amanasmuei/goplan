// Package types provides request/response DTOs for the REST API.
package types

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/goplan/goplan/internal/domain/shared"
)

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Error   bool           `json:"error"`
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details []ErrorDetail  `json:"details,omitempty"`
}

// ErrorDetail provides additional error information.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// SuccessResponse represents a standardized success response.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated response wrapper.
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"totalCount"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}

// PaginationParams represents pagination query parameters.
type PaginationParams struct {
	Page     int
	PageSize int
}

// ParsePagination extracts pagination parameters from the request.
func ParsePagination(r *http.Request) PaginationParams {
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset calculates the database offset for pagination.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// NewErrorResponse creates an error response.
func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:   true,
		Code:    code,
		Message: message,
	}
}

// NewErrorResponseWithDetails creates an error response with field details.
func NewErrorResponseWithDetails(code, message string, details []ErrorDetail) *ErrorResponse {
	return &ErrorResponse{
		Error:   true,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewSuccessResponse creates a success response.
func NewSuccessResponse(data interface{}) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Data:    data,
	}
}

// NewSuccessResponseWithMessage creates a success response with a message.
func NewSuccessResponseWithMessage(data interface{}, message string) *SuccessResponse {
	return &SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// NewPaginatedResponse creates a paginated response.
func NewPaginatedResponse(items interface{}, totalCount int64, page, pageSize int) *PaginatedResponse {
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 1 {
		totalPages = 1
	}
	return &PaginatedResponse{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// HTTPStatusFromError determines the HTTP status code from a domain error.
func HTTPStatusFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if errors.Is(err, shared.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, shared.ErrAlreadyExists) {
		return http.StatusConflict
	}
	if errors.Is(err, shared.ErrInvalidInput) {
		return http.StatusBadRequest
	}
	if errors.Is(err, shared.ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	if errors.Is(err, shared.ErrForbidden) {
		return http.StatusForbidden
	}
	if errors.Is(err, shared.ErrConflict) {
		return http.StatusConflict
	}

	return http.StatusInternalServerError
}

// ErrorCodeFromError determines the error code from a domain error.
func ErrorCodeFromError(err error) string {
	var domainErr *shared.DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}

	var validationErrs *shared.ValidationErrors
	if errors.As(err, &validationErrs) {
		return "VALIDATION_ERROR"
	}

	if errors.Is(err, shared.ErrNotFound) {
		return "NOT_FOUND"
	}
	if errors.Is(err, shared.ErrAlreadyExists) {
		return "ALREADY_EXISTS"
	}
	if errors.Is(err, shared.ErrInvalidInput) {
		return "INVALID_INPUT"
	}
	if errors.Is(err, shared.ErrUnauthorized) {
		return "UNAUTHORIZED"
	}
	if errors.Is(err, shared.ErrForbidden) {
		return "FORBIDDEN"
	}
	if errors.Is(err, shared.ErrConflict) {
		return "CONFLICT"
	}

	return "INTERNAL_ERROR"
}

// ErrorDetailsFromError extracts error details from a validation error.
func ErrorDetailsFromError(err error) []ErrorDetail {
	var validationErrs *shared.ValidationErrors
	if errors.As(err, &validationErrs) {
		details := make([]ErrorDetail, 0, len(validationErrs.Errors))
		for _, e := range validationErrs.Errors {
			details = append(details, ErrorDetail{
				Field:   e.Field,
				Message: e.Message,
			})
		}
		return details
	}

	var domainErr *shared.DomainError
	if errors.As(err, &domainErr) && domainErr.Field != "" {
		return []ErrorDetail{{
			Field:   domainErr.Field,
			Message: domainErr.Message,
		}}
	}

	return nil
}

// User request/response types

// UpdateUserRequest represents the request to update the current user.
type UpdateUserRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
}

// Validate validates the update user request.
func (r *UpdateUserRequest) Validate() error {
	if r.Name != nil && *r.Name == "" {
		return shared.NewValidationError("name", "name cannot be empty")
	}
	if r.Name != nil && len(*r.Name) > 255 {
		return shared.NewValidationError("name", "name must be at most 255 characters")
	}
	return nil
}

// Workspace request/response types

// CreateWorkspaceRequest represents the request to create a workspace.
type CreateWorkspaceRequest struct {
	Name     string                  `json:"name"`
	Slug     string                  `json:"slug"`
	Settings *WorkspaceSettingsInput `json:"settings,omitempty"`
}

// WorkspaceSettingsInput represents workspace settings input.
type WorkspaceSettingsInput struct {
	DefaultView string `json:"defaultView"`
	AIEnabled   bool   `json:"aiEnabled"`
}

// Validate validates the create workspace request.
func (r *CreateWorkspaceRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Name == "" {
		errs.Add("name", "name is required")
	} else if len(r.Name) > 255 {
		errs.Add("name", "name must be at most 255 characters")
	}

	if r.Slug == "" {
		errs.Add("slug", "slug is required")
	} else if len(r.Slug) < 3 || len(r.Slug) > 50 {
		errs.Add("slug", "slug must be between 3 and 50 characters")
	}

	if r.Settings != nil && !shared.IsValidDefaultView(r.Settings.DefaultView) {
		errs.Add("settings.defaultView", "invalid default view")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateWorkspaceRequest represents the request to update a workspace.
type UpdateWorkspaceRequest struct {
	Name     *string                 `json:"name,omitempty"`
	Settings *WorkspaceSettingsInput `json:"settings,omitempty"`
}

// Validate validates the update workspace request.
func (r *UpdateWorkspaceRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Name != nil {
		if *r.Name == "" {
			errs.Add("name", "name cannot be empty")
		} else if len(*r.Name) > 255 {
			errs.Add("name", "name must be at most 255 characters")
		}
	}

	if r.Settings != nil && !shared.IsValidDefaultView(r.Settings.DefaultView) {
		errs.Add("settings.defaultView", "invalid default view")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// AddMemberRequest represents the request to add a member to a workspace.
type AddMemberRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// Validate validates the add member request.
func (r *AddMemberRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.UserID == "" {
		errs.Add("userId", "user ID is required")
	}

	if !shared.IsValidMemberRole(r.Role) {
		errs.Add("role", "invalid role; must be one of: owner, admin, member, viewer")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Plan request/response types

// CreatePlanRequest represents the request to create a plan.
type CreatePlanRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Domain      string  `json:"domain"`
	StartDate   *string `json:"startDate,omitempty"`
	EndDate     *string `json:"endDate,omitempty"`
}

// Validate validates the create plan request.
func (r *CreatePlanRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Name == "" {
		errs.Add("name", "name is required")
	} else if len(r.Name) > 255 {
		errs.Add("name", "name must be at most 255 characters")
	}

	if !shared.IsValidPlanDomain(r.Domain) {
		errs.Add("domain", "invalid domain; must be one of: software, event, ops, collection, generic")
	}

	if r.StartDate != nil {
		if _, err := time.Parse("2006-01-02", *r.StartDate); err != nil {
			errs.Add("startDate", "invalid date format; use YYYY-MM-DD")
		}
	}

	if r.EndDate != nil {
		if _, err := time.Parse("2006-01-02", *r.EndDate); err != nil {
			errs.Add("endDate", "invalid date format; use YYYY-MM-DD")
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdatePlanRequest represents the request to update a plan.
type UpdatePlanRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Status      *string  `json:"status,omitempty"`
	StartDate   *string  `json:"startDate,omitempty"`
	EndDate     *string  `json:"endDate,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Validate validates the update plan request.
func (r *UpdatePlanRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Name != nil {
		if *r.Name == "" {
			errs.Add("name", "name cannot be empty")
		} else if len(*r.Name) > 255 {
			errs.Add("name", "name must be at most 255 characters")
		}
	}

	if r.Status != nil && !shared.IsValidPlanStatus(*r.Status) {
		errs.Add("status", "invalid status; must be one of: draft, active, on_hold, completed, archived")
	}

	if r.StartDate != nil {
		if _, err := time.Parse("2006-01-02", *r.StartDate); err != nil {
			errs.Add("startDate", "invalid date format; use YYYY-MM-DD")
		}
	}

	if r.EndDate != nil {
		if _, err := time.Parse("2006-01-02", *r.EndDate); err != nil {
			errs.Add("endDate", "invalid date format; use YYYY-MM-DD")
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// Task request/response types

// CreateTaskRequest represents the request to create a task.
type CreateTaskRequest struct {
	Title             string                 `json:"title"`
	Description       *string                `json:"description,omitempty"`
	Status            *string                `json:"status,omitempty"`
	PhaseID           *string                `json:"phaseId,omitempty"`
	ParentID          *string                `json:"parentId,omitempty"`
	Priority          string                 `json:"priority"`
	AssigneeID        *string                `json:"assigneeId,omitempty"`
	DueDate           *string                `json:"dueDate,omitempty"`
	EstimatedHours    *float64               `json:"estimatedHours,omitempty"`
	CustomFieldValues map[string]interface{} `json:"customFieldValues,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
}

// Validate validates the create task request.
func (r *CreateTaskRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Title == "" {
		errs.Add("title", "title is required")
	} else if len(r.Title) > 500 {
		errs.Add("title", "title must be at most 500 characters")
	}

	if r.Priority == "" {
		r.Priority = shared.TaskPriorityMedium // Default priority
	} else if !shared.IsValidTaskPriority(r.Priority) {
		errs.Add("priority", "invalid priority; must be one of: low, medium, high, critical")
	}

	if r.DueDate != nil {
		if _, err := time.Parse("2006-01-02", *r.DueDate); err != nil {
			errs.Add("dueDate", "invalid date format; use YYYY-MM-DD")
		}
	}

	if r.EstimatedHours != nil && *r.EstimatedHours < 0 {
		errs.Add("estimatedHours", "estimated hours cannot be negative")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateTaskRequest represents the request to update a task.
type UpdateTaskRequest struct {
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

// Validate validates the update task request.
func (r *UpdateTaskRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Title != nil {
		if *r.Title == "" {
			errs.Add("title", "title cannot be empty")
		} else if len(*r.Title) > 500 {
			errs.Add("title", "title must be at most 500 characters")
		}
	}

	if r.Priority != nil && !shared.IsValidTaskPriority(*r.Priority) {
		errs.Add("priority", "invalid priority; must be one of: low, medium, high, critical")
	}

	if r.DueDate != nil {
		if _, err := time.Parse("2006-01-02", *r.DueDate); err != nil {
			errs.Add("dueDate", "invalid date format; use YYYY-MM-DD")
		}
	}

	if r.EstimatedHours != nil && *r.EstimatedHours < 0 {
		errs.Add("estimatedHours", "estimated hours cannot be negative")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// MoveTaskRequest represents the request to move a task.
type MoveTaskRequest struct {
	Status   string `json:"status"`
	Position int    `json:"position"`
}

// Validate validates the move task request.
func (r *MoveTaskRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Status == "" {
		errs.Add("status", "status is required")
	}

	if r.Position < 0 {
		errs.Add("position", "position cannot be negative")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// TaskSearchParams represents task search query parameters.
type TaskSearchParams struct {
	Query      string
	PlanID     string
	Status     string
	Priority   string
	AssigneeID string
	DueBefore  string
	DueAfter   string
	Tags       []string
}

// ParseTaskSearchParams extracts task search parameters from the request.
func ParseTaskSearchParams(r *http.Request) TaskSearchParams {
	q := r.URL.Query()
	return TaskSearchParams{
		Query:      q.Get("q"),
		PlanID:     q.Get("planId"),
		Status:     q.Get("status"),
		Priority:   q.Get("priority"),
		AssigneeID: q.Get("assigneeId"),
		DueBefore:  q.Get("dueBefore"),
		DueAfter:   q.Get("dueAfter"),
		Tags:       q["tags"],
	}
}

// Comment request/response types

// CreateCommentRequest represents the request to create a comment.
type CreateCommentRequest struct {
	Content  string   `json:"content"`
	Mentions []string `json:"mentions,omitempty"`
}

// Validate validates the create comment request.
func (r *CreateCommentRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Content == "" {
		errs.Add("content", "content is required")
	} else if len(r.Content) > 10000 {
		errs.Add("content", "content must be at most 10000 characters")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// UpdateCommentRequest represents the request to update a comment.
type UpdateCommentRequest struct {
	Content  string   `json:"content"`
	Mentions []string `json:"mentions,omitempty"`
}

// Validate validates the update comment request.
func (r *UpdateCommentRequest) Validate() error {
	errs := &shared.ValidationErrors{}

	if r.Content == "" {
		errs.Add("content", "content is required")
	} else if len(r.Content) > 10000 {
		errs.Add("content", "content must be at most 10000 characters")
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// DecodeJSON decodes JSON from the request body into the target.
func DecodeJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return shared.NewValidationError("body", "request body is required")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return shared.NewValidationError("body", "invalid JSON: "+err.Error())
	}

	return nil
}
