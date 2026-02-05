package handlers

import (
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/api/middleware"
	"github.com/goplan/goplan/internal/api/types"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/repository"
)

// TaskHandler handles task-related HTTP requests.
type TaskHandler struct {
	*BaseHandler
	taskRepo      repository.TaskRepository
	planRepo      repository.PlanRepository
	workspaceRepo repository.WorkspaceRepository
}

// NewTaskHandler creates a new task handler.
func NewTaskHandler(taskRepo repository.TaskRepository, planRepo repository.PlanRepository, workspaceRepo repository.WorkspaceRepository) *TaskHandler {
	return &TaskHandler{
		BaseHandler:   NewBaseHandler(),
		taskRepo:      taskRepo,
		planRepo:      planRepo,
		workspaceRepo: workspaceRepo,
	}
}

// ListTasksByPlan handles GET /api/v1/plans/:pid/tasks
func (h *TaskHandler) ListTasksByPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planID := extractPlanIDFromTaskPath(r.URL.Path)

	if planID == "" {
		h.WriteBadRequest(w, "plan ID is required")
		return
	}

	pagination := h.ParsePagination(r)

	result, err := h.taskRepo.ListByPlan(ctx, planID, repository.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to response format
	responses := make([]*task.TaskResponse, len(result.Items))
	for i, t := range result.Items {
		responses[i] = t.ToResponse()
	}

	h.WritePaginated(w, responses, result.TotalCount, result.Page, result.PageSize)
}

// CreateTask handles POST /api/v1/plans/:pid/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	planID := extractPlanIDFromTaskPath(r.URL.Path)

	if planID == "" {
		h.WriteBadRequest(w, "plan ID is required")
		return
	}

	// Verify plan exists and get default status
	p, err := h.planRepo.GetByID(ctx, planID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Get user's role in the workspace
	role := auth.GetUserRole(ctx)
	if role == "" {
		ws, err := h.workspaceRepo.GetByID(ctx, p.WorkspaceID)
		if err == nil {
			if ws.OwnerID == userID {
				role = "owner"
			} else {
				member, err := h.workspaceRepo.GetMember(ctx, p.WorkspaceID, userID)
				if err == nil && member != nil {
					role = member.Role
				}
			}
		}
	}

	// Check permissions
	if !middleware.CanEdit(role) {
		h.WriteForbidden(w, "you don't have permission to create tasks")
		return
	}

	var req types.CreateTaskRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Generate a new ID
	id := generateID()

	// Determine status: use provided status or plan's default
	var taskStatus string
	if req.Status != nil && *req.Status != "" {
		// Validate status against plan's custom statuses
		if !p.HasStatus(*req.Status) {
			h.WriteError(w, shared.NewValidationError("status", "invalid status for this plan"))
			return
		}
		taskStatus = *req.Status
	} else {
		taskStatus = p.GetDefaultStatus()
	}

	t := task.NewTask(id, planID, req.Title, taskStatus, req.Priority)

	// Set optional fields
	t.Description = req.Description
	t.PhaseID = req.PhaseID
	t.ParentID = req.ParentID
	t.AssigneeID = req.AssigneeID
	t.DueDate = req.DueDate
	t.EstimatedHours = req.EstimatedHours
	if req.CustomFieldValues != nil {
		t.CustomFieldValues = req.CustomFieldValues
	}
	if req.Tags != nil {
		t.Tags = req.Tags
	}

	if err := h.taskRepo.Create(ctx, t); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteCreated(w, t.ToResponse())
}

// GetTask handles GET /api/v1/tasks/:id
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := extractTaskID(r.URL.Path)

	if taskID == "" {
		h.WriteBadRequest(w, "task ID is required")
		return
	}

	// Check if details are requested
	withDetails := r.URL.Query().Get("details") == "true"

	var t *task.Task
	var err error

	if withDetails {
		t, err = h.taskRepo.GetByIDWithDetails(ctx, taskID)
	} else {
		t, err = h.taskRepo.GetByID(ctx, taskID)
	}

	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, t.ToResponse())
}

// UpdateTask handles PUT /api/v1/tasks/:id
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	taskID := extractTaskID(r.URL.Path)

	if taskID == "" {
		h.WriteBadRequest(w, "task ID is required")
		return
	}

	// Verify task exists
	existingTask, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Get user's role in the workspace via the plan
	role := auth.GetUserRole(ctx)
	if role == "" {
		p, err := h.planRepo.GetByID(ctx, existingTask.PlanID)
		if err == nil {
			ws, err := h.workspaceRepo.GetByID(ctx, p.WorkspaceID)
			if err == nil {
				if ws.OwnerID == userID {
					role = "owner"
				} else {
					member, err := h.workspaceRepo.GetMember(ctx, p.WorkspaceID, userID)
					if err == nil && member != nil {
						role = member.Role
					}
				}
			}
		}
	}

	// Check permissions
	if !middleware.CanEdit(role) {
		h.WriteForbidden(w, "you don't have permission to update tasks")
		return
	}

	var req types.UpdateTaskRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Validate status against plan's custom statuses if status is being updated
	if req.Status != nil {
		p, err := h.planRepo.GetByID(ctx, existingTask.PlanID)
		if err != nil {
			h.WriteError(w, err)
			return
		}
		if !p.HasStatus(*req.Status) {
			h.WriteError(w, shared.NewValidationError("status", "invalid status for this plan"))
			return
		}
	}

	// Convert to domain input
	input := &task.UpdateTaskInput{
		Title:             req.Title,
		Description:       req.Description,
		Status:            req.Status,
		Priority:          req.Priority,
		PhaseID:           req.PhaseID,
		AssigneeID:        req.AssigneeID,
		DueDate:           req.DueDate,
		EstimatedHours:    req.EstimatedHours,
		CustomFieldValues: req.CustomFieldValues,
		Tags:              req.Tags,
		Position:          req.Position,
	}

	updatedTask, err := h.taskRepo.Update(ctx, taskID, input)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, updatedTask.ToResponse())
}

// DeleteTask handles DELETE /api/v1/tasks/:id
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	taskID := extractTaskID(r.URL.Path)

	if taskID == "" {
		h.WriteBadRequest(w, "task ID is required")
		return
	}

	// Verify task exists
	existingTask, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Get user's role in the workspace via the plan
	role := auth.GetUserRole(ctx)
	if role == "" {
		p, err := h.planRepo.GetByID(ctx, existingTask.PlanID)
		if err == nil {
			ws, err := h.workspaceRepo.GetByID(ctx, p.WorkspaceID)
			if err == nil {
				if ws.OwnerID == userID {
					role = "owner"
				} else {
					member, err := h.workspaceRepo.GetMember(ctx, p.WorkspaceID, userID)
					if err == nil && member != nil {
						role = member.Role
					}
				}
			}
		}
	}

	// Check permissions
	if !middleware.CanEdit(role) {
		h.WriteForbidden(w, "you don't have permission to delete tasks")
		return
	}

	if err := h.taskRepo.Delete(ctx, taskID); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteNoContent(w)
}

// MoveTask handles POST /api/v1/tasks/:id/move
func (h *TaskHandler) MoveTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	taskID := extractTaskIDForMove(r.URL.Path)

	if taskID == "" {
		h.WriteBadRequest(w, "task ID is required")
		return
	}

	// Verify task exists
	existingTask, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Get user's role in the workspace via the plan
	role := auth.GetUserRole(ctx)
	if role == "" {
		p, err := h.planRepo.GetByID(ctx, existingTask.PlanID)
		if err == nil {
			ws, err := h.workspaceRepo.GetByID(ctx, p.WorkspaceID)
			if err == nil {
				if ws.OwnerID == userID {
					role = "owner"
				} else {
					member, err := h.workspaceRepo.GetMember(ctx, p.WorkspaceID, userID)
					if err == nil && member != nil {
						role = member.Role
					}
				}
			}
		}
	}

	// Check permissions
	if !middleware.CanEdit(role) {
		h.WriteForbidden(w, "you don't have permission to move tasks")
		return
	}

	var req types.MoveTaskRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Validate status against plan's custom statuses
	p, err := h.planRepo.GetByID(ctx, existingTask.PlanID)
	if err != nil {
		h.WriteError(w, err)
		return
	}
	if !p.HasStatus(req.Status) {
		h.WriteError(w, shared.NewValidationError("status", "invalid status for this plan"))
		return
	}

	if err := h.taskRepo.Move(ctx, taskID, req.Status, req.Position); err != nil {
		h.WriteError(w, err)
		return
	}

	// Get updated task
	updatedTask, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, updatedTask.ToResponse())
}

// SearchTasks handles GET /api/v1/tasks/search
func (h *TaskHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := types.ParseTaskSearchParams(r)
	pagination := h.ParsePagination(r)

	// Search requires either a plan ID or a query
	if params.PlanID == "" && params.Query == "" {
		h.WriteBadRequest(w, "either planId or query parameter is required")
		return
	}

	// Build filter options
	filter := repository.TaskFilterOptions{
		SearchQuery: &params.Query,
	}

	if params.PlanID != "" {
		filter.PlanID = &params.PlanID
	}
	if params.Status != "" {
		filter.Status = &params.Status
	}
	if params.Priority != "" {
		filter.Priority = &params.Priority
	}
	if params.AssigneeID != "" {
		filter.AssigneeID = &params.AssigneeID
	}
	if params.DueBefore != "" {
		filter.DueBefore = &params.DueBefore
	}
	if params.DueAfter != "" {
		filter.DueAfter = &params.DueAfter
	}
	if len(params.Tags) > 0 {
		filter.Tags = params.Tags
	}

	result, err := h.taskRepo.List(ctx, filter, repository.DefaultTaskSort(), repository.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to response format
	responses := make([]*task.TaskResponse, len(result.Items))
	for i, t := range result.Items {
		responses[i] = t.ToResponse()
	}

	h.WritePaginated(w, responses, result.TotalCount, result.Page, result.PageSize)
}

// ServeHTTP routes requests to the appropriate handler method.
func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle GET /api/v1/tasks/search
	if path == "/api/v1/tasks/search" {
		if r.Method != http.MethodGet {
			h.WriteMethodNotAllowed(w, []string{http.MethodGet})
			return
		}
		h.SearchTasks(w, r)
		return
	}

	// Handle /api/v1/plans/:pid/tasks
	if strings.Contains(path, "/plans/") && strings.HasSuffix(path, "/tasks") {
		switch r.Method {
		case http.MethodGet:
			h.ListTasksByPlan(w, r)
		case http.MethodPost:
			h.CreateTask(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPost})
		}
		return
	}

	// Handle POST /api/v1/tasks/:id/move
	if strings.HasPrefix(path, "/api/v1/tasks/") && strings.HasSuffix(path, "/move") {
		if r.Method != http.MethodPost {
			h.WriteMethodNotAllowed(w, []string{http.MethodPost})
			return
		}
		h.MoveTask(w, r)
		return
	}

	// Handle /api/v1/tasks/:id
	if strings.HasPrefix(path, "/api/v1/tasks/") {
		switch r.Method {
		case http.MethodGet:
			h.GetTask(w, r)
		case http.MethodPut:
			h.UpdateTask(w, r)
		case http.MethodDelete:
			h.DeleteTask(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPut, http.MethodDelete})
		}
		return
	}

	h.WriteNotFound(w, "endpoint")
}

// extractPlanIDFromTaskPath extracts plan ID from path like /api/v1/plans/{pid}/tasks
func extractPlanIDFromTaskPath(path string) string {
	prefix := "/api/v1/plans/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// extractTaskID extracts task ID from path like /api/v1/tasks/{id}
func extractTaskID(path string) string {
	prefix := "/api/v1/tasks/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && parts[0] != "" && parts[0] != "search" {
		return parts[0]
	}
	return ""
}

// extractTaskIDForMove extracts task ID from path like /api/v1/tasks/{id}/move
func extractTaskIDForMove(path string) string {
	prefix := "/api/v1/tasks/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(path, prefix)
	remainder = strings.TrimSuffix(remainder, "/move")
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}
