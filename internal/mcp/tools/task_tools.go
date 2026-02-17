package tools

import (
	"context"

	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
	"github.com/google/uuid"
)

// Task tool names
const (
	ToolTaskList   = "task.list"
	ToolTaskGet    = "task.get"
	ToolTaskCreate = "task.create"
	ToolTaskUpdate = "task.update"
	ToolTaskSearch = "task.search"
	ToolTaskMove   = "task.move"
)

// ListTasksTool lists tasks with filtering.
type ListTasksTool struct {
	taskRepo repository.TaskRepository
	planRepo repository.PlanRepository
}

// NewListTasksTool creates a new ListTasksTool.
func NewListTasksTool(taskRepo repository.TaskRepository, planRepo repository.PlanRepository) *ListTasksTool {
	return &ListTasksTool{taskRepo: taskRepo, planRepo: planRepo}
}

// Name returns the tool name.
func (t *ListTasksTool) Name() string {
	return ToolTaskList
}

// Description returns the tool description.
func (t *ListTasksTool) Description() string {
	return "List tasks with optional filtering by plan, phase, status, priority, assignee"
}

// Execute executes the tool.
func (t *ListTasksTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	planID, err := getRequiredString(args, "planId")
	if err != nil {
		// Try from execution context
		if execCtx.PlanID != nil && *execCtx.PlanID != "" {
			planID = *execCtx.PlanID
		} else {
			return nil, err
		}
	}

	// Verify plan belongs to the user's workspace
	p, err := t.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	if p.WorkspaceID != execCtx.WorkspaceID {
		return nil, shared.NewForbiddenError("access", "plan")
	}

	// Parse pagination
	page := getInt(args, "page", 1)
	pageSize := getInt(args, "pageSize", 50)
	pagination := repository.Pagination{Page: page, PageSize: pageSize}.Normalize()

	// Parse filters
	filter := repository.TaskFilterOptions{
		PlanID: &planID,
	}
	if phaseID := getOptionalString(args, "phaseId"); phaseID != nil {
		filter.PhaseID = phaseID
	}
	if status := getOptionalString(args, "status"); status != nil {
		filter.Status = status
	}
	if priority := getOptionalString(args, "priority"); priority != nil {
		filter.Priority = priority
	}
	if assigneeID := getOptionalString(args, "assigneeId"); assigneeID != nil {
		filter.AssigneeID = assigneeID
	}
	if tags := getStringSlice(args, "tags"); len(tags) > 0 {
		filter.Tags = tags
	}
	if dueBefore := getOptionalString(args, "dueBefore"); dueBefore != nil {
		filter.DueBefore = dueBefore
	}
	if dueAfter := getOptionalString(args, "dueAfter"); dueAfter != nil {
		filter.DueAfter = dueAfter
	}

	// Parse sort options
	sort := repository.DefaultTaskSort()
	if sortField := getOptionalString(args, "sortBy"); sortField != nil {
		switch *sortField {
		case "position":
			sort.Field = repository.TaskSortByPosition
		case "createdAt":
			sort.Field = repository.TaskSortByCreatedAt
		case "updatedAt":
			sort.Field = repository.TaskSortByUpdatedAt
		case "dueDate":
			sort.Field = repository.TaskSortByDueDate
		case "priority":
			sort.Field = repository.TaskSortByPriority
		}
	}
	if sortOrder := getOptionalString(args, "sortOrder"); sortOrder != nil {
		if *sortOrder == "desc" {
			sort.Order = repository.SortDesc
		}
	}

	// Get tasks
	result, err := t.taskRepo.List(ctx, filter, sort, pagination)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	tasks := make([]*task.TaskResponse, len(result.Items))
	for i, tk := range result.Items {
		tasks[i] = tk.ToResponse()
	}

	return map[string]interface{}{
		"tasks":      tasks,
		"totalCount": result.TotalCount,
		"page":       result.Page,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
	}, nil
}

// GetTaskTool retrieves task details with comments.
type GetTaskTool struct {
	taskRepo    repository.TaskRepository
	planRepo    repository.PlanRepository
	commentRepo repository.CommentRepository
}

// NewGetTaskTool creates a new GetTaskTool.
func NewGetTaskTool(taskRepo repository.TaskRepository, planRepo repository.PlanRepository, commentRepo repository.CommentRepository) *GetTaskTool {
	return &GetTaskTool{
		taskRepo:    taskRepo,
		planRepo:    planRepo,
		commentRepo: commentRepo,
	}
}

// Name returns the tool name.
func (t *GetTaskTool) Name() string {
	return ToolTaskGet
}

// Description returns the tool description.
func (t *GetTaskTool) Description() string {
	return "Get task details including subtasks, dependencies, and comments"
}

// Execute executes the tool.
func (t *GetTaskTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	taskID, err := getRequiredString(args, "taskId")
	if err != nil {
		return nil, err
	}

	// Get task with details (subtasks, dependencies)
	tk, err := t.taskRepo.GetByIDWithDetails(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Verify workspace ownership via plan
	p, err := t.planRepo.GetByID(ctx, tk.PlanID)
	if err != nil {
		return nil, err
	}
	if p.WorkspaceID != execCtx.WorkspaceID {
		return nil, shared.NewForbiddenError("access", "task")
	}

	// Get comments
	comments, err := t.commentRepo.ListByTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Add comments to task
	tk.Comments = make([]task.Comment, len(comments))
	for i, c := range comments {
		tk.Comments[i] = *c
	}

	return map[string]interface{}{
		"task": tk.ToResponse(),
	}, nil
}

// CreateTaskTool creates a new task.
type CreateTaskTool struct {
	taskRepo repository.TaskRepository
	planRepo repository.PlanRepository
}

// NewCreateTaskTool creates a new CreateTaskTool.
func NewCreateTaskTool(taskRepo repository.TaskRepository, planRepo repository.PlanRepository) *CreateTaskTool {
	return &CreateTaskTool{
		taskRepo: taskRepo,
		planRepo: planRepo,
	}
}

// Name returns the tool name.
func (t *CreateTaskTool) Name() string {
	return ToolTaskCreate
}

// Description returns the tool description.
func (t *CreateTaskTool) Description() string {
	return "Create a new task in a plan"
}

// Execute executes the tool.
func (t *CreateTaskTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	planID, err := getRequiredString(args, "planId")
	if err != nil {
		// Try from execution context
		if execCtx.PlanID != nil && *execCtx.PlanID != "" {
			planID = *execCtx.PlanID
		} else {
			return nil, err
		}
	}

	title, err := getRequiredString(args, "title")
	if err != nil {
		return nil, err
	}

	// Get plan to determine default status
	p, err := t.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Verify workspace ownership
	if p.WorkspaceID != execCtx.WorkspaceID {
		return nil, shared.NewForbiddenError("access", "plan")
	}

	// Determine status
	status := p.GetDefaultStatus()
	if s := getOptionalString(args, "status"); s != nil {
		if p.HasStatus(*s) {
			status = *s
		} else {
			return nil, shared.NewValidationError("status", "invalid status for this plan")
		}
	}

	// Determine priority
	priority := shared.TaskPriorityMedium
	if pr := getOptionalString(args, "priority"); pr != nil {
		if shared.IsValidTaskPriority(*pr) {
			priority = *pr
		} else {
			return nil, shared.NewValidationError("priority", "invalid priority. Valid values: low, medium, high, critical")
		}
	}

	// Create task
	tk := task.NewTask(
		uuid.New().String(),
		planID,
		title,
		status,
		priority,
	)

	// Set optional fields
	tk.Description = getOptionalString(args, "description")
	tk.PhaseID = getOptionalString(args, "phaseId")
	tk.ParentID = getOptionalString(args, "parentId")
	tk.AssigneeID = getOptionalString(args, "assigneeId")
	tk.DueDate = getOptionalString(args, "dueDate")
	tk.EstimatedHours = getOptionalFloat(args, "estimatedHours")

	if tags := getStringSlice(args, "tags"); len(tags) > 0 {
		tk.Tags = tags
	}

	if customFields := getMap(args, "customFieldValues"); customFields != nil {
		tk.CustomFieldValues = customFields
	}

	// Validate
	if validErrs := tk.Validate(); validErrs != nil {
		return nil, shared.NewValidationError("task", validErrs.Error())
	}

	if err := t.taskRepo.Create(ctx, tk); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"task":    tk.ToResponse(),
		"message": "Task created successfully",
	}, nil
}

// UpdateTaskTool updates a task.
type UpdateTaskTool struct {
	taskRepo repository.TaskRepository
	planRepo repository.PlanRepository
}

// NewUpdateTaskTool creates a new UpdateTaskTool.
func NewUpdateTaskTool(taskRepo repository.TaskRepository, planRepo repository.PlanRepository) *UpdateTaskTool {
	return &UpdateTaskTool{
		taskRepo: taskRepo,
		planRepo: planRepo,
	}
}

// Name returns the tool name.
func (t *UpdateTaskTool) Name() string {
	return ToolTaskUpdate
}

// Description returns the tool description.
func (t *UpdateTaskTool) Description() string {
	return "Update a task's properties (title, description, status, priority, assignee, due date, etc.)"
}

// Execute executes the tool.
func (t *UpdateTaskTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	taskID, err := getRequiredString(args, "taskId")
	if err != nil {
		return nil, err
	}

	// Get existing task to validate status against plan
	existingTask, err := t.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Verify workspace ownership via plan
	{
		p, err := t.planRepo.GetByID(ctx, existingTask.PlanID)
		if err != nil {
			return nil, err
		}
		if p.WorkspaceID != execCtx.WorkspaceID {
			return nil, shared.NewForbiddenError("access", "task")
		}
	}

	// Build update input
	input := &task.UpdateTaskInput{}
	hasUpdates := false

	if title := getOptionalString(args, "title"); title != nil {
		input.Title = title
		hasUpdates = true
	}
	if description := getOptionalString(args, "description"); description != nil {
		input.Description = description
		hasUpdates = true
	}
	if status := getOptionalString(args, "status"); status != nil {
		// Validate status against plan
		p, err := t.planRepo.GetByID(ctx, existingTask.PlanID)
		if err != nil {
			return nil, err
		}
		if !p.HasStatus(*status) {
			return nil, shared.NewValidationError("status", "invalid status for this plan")
		}
		input.Status = status
		hasUpdates = true
	}
	if priority := getOptionalString(args, "priority"); priority != nil {
		if !shared.IsValidTaskPriority(*priority) {
			return nil, shared.NewValidationError("priority", "invalid priority. Valid values: low, medium, high, critical")
		}
		input.Priority = priority
		hasUpdates = true
	}
	if phaseID := getOptionalString(args, "phaseId"); phaseID != nil {
		input.PhaseID = phaseID
		hasUpdates = true
	}
	if assigneeID := getOptionalString(args, "assigneeId"); assigneeID != nil {
		input.AssigneeID = assigneeID
		hasUpdates = true
	}
	if dueDate := getOptionalString(args, "dueDate"); dueDate != nil {
		input.DueDate = dueDate
		hasUpdates = true
	}
	if estimatedHours := getOptionalFloat(args, "estimatedHours"); estimatedHours != nil {
		input.EstimatedHours = estimatedHours
		hasUpdates = true
	}
	if tags := getStringSlice(args, "tags"); tags != nil {
		input.Tags = tags
		hasUpdates = true
	}
	if customFields := getMap(args, "customFieldValues"); customFields != nil {
		input.CustomFieldValues = customFields
		hasUpdates = true
	}
	if position := getOptionalInt(args, "position"); position != nil {
		input.Position = position
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, shared.NewValidationError("fields", "at least one field must be provided for update")
	}

	// Update task
	updated, err := t.taskRepo.Update(ctx, taskID, input)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"task":    updated.ToResponse(),
		"message": "Task updated successfully",
	}, nil
}

// SearchTasksTool performs full-text search on tasks.
type SearchTasksTool struct {
	taskRepo repository.TaskRepository
	planRepo repository.PlanRepository
}

// NewSearchTasksTool creates a new SearchTasksTool.
func NewSearchTasksTool(taskRepo repository.TaskRepository, planRepo repository.PlanRepository) *SearchTasksTool {
	return &SearchTasksTool{taskRepo: taskRepo, planRepo: planRepo}
}

// Name returns the tool name.
func (t *SearchTasksTool) Name() string {
	return ToolTaskSearch
}

// Description returns the tool description.
func (t *SearchTasksTool) Description() string {
	return "Search tasks by title and description using full-text search"
}

// Execute executes the tool.
func (t *SearchTasksTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	planID, err := getRequiredString(args, "planId")
	if err != nil {
		// Try from execution context
		if execCtx.PlanID != nil && *execCtx.PlanID != "" {
			planID = *execCtx.PlanID
		} else {
			return nil, err
		}
	}

	// Verify plan belongs to the user's workspace
	planCheck, err := t.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	if planCheck.WorkspaceID != execCtx.WorkspaceID {
		return nil, shared.NewForbiddenError("access", "plan")
	}

	query, err := getRequiredString(args, "query")
	if err != nil {
		return nil, err
	}

	// Parse pagination
	page := getInt(args, "page", 1)
	pageSize := getInt(args, "pageSize", 20)
	pagination := repository.Pagination{Page: page, PageSize: pageSize}.Normalize()

	// Search tasks
	result, err := t.taskRepo.Search(ctx, planID, query, pagination)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	tasks := make([]*task.TaskResponse, len(result.Items))
	for i, tk := range result.Items {
		tasks[i] = tk.ToResponse()
	}

	return map[string]interface{}{
		"tasks":      tasks,
		"query":      query,
		"totalCount": result.TotalCount,
		"page":       result.Page,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
	}, nil
}

// MoveTaskTool moves a task to a different status/position.
type MoveTaskTool struct {
	taskRepo repository.TaskRepository
	planRepo repository.PlanRepository
}

// NewMoveTaskTool creates a new MoveTaskTool.
func NewMoveTaskTool(taskRepo repository.TaskRepository, planRepo repository.PlanRepository) *MoveTaskTool {
	return &MoveTaskTool{
		taskRepo: taskRepo,
		planRepo: planRepo,
	}
}

// Name returns the tool name.
func (t *MoveTaskTool) Name() string {
	return ToolTaskMove
}

// Description returns the tool description.
func (t *MoveTaskTool) Description() string {
	return "Move a task to a different status column and/or position (for Kanban board)"
}

// Execute executes the tool.
func (t *MoveTaskTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	taskID, err := getRequiredString(args, "taskId")
	if err != nil {
		return nil, err
	}

	status, err := getRequiredString(args, "status")
	if err != nil {
		return nil, err
	}

	position := getInt(args, "position", 0)

	// Get task to validate status
	tk, err := t.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Validate status against plan
	p, err := t.planRepo.GetByID(ctx, tk.PlanID)
	if err != nil {
		return nil, err
	}

	// Verify workspace ownership
	if p.WorkspaceID != execCtx.WorkspaceID {
		return nil, shared.NewForbiddenError("access", "task")
	}

	if !p.HasStatus(status) {
		return nil, shared.NewValidationError("status", "invalid status for this plan")
	}

	// Move task
	if err := t.taskRepo.Move(ctx, taskID, status, position); err != nil {
		return nil, err
	}

	// Get updated task
	updated, err := t.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"task":    updated.ToResponse(),
		"message": "Task moved successfully",
	}, nil
}

// RegisterTaskTools registers all task tools with the registry.
func RegisterTaskTools(registry *mcp.ToolRegistry, taskRepo repository.TaskRepository, planRepo repository.PlanRepository, commentRepo repository.CommentRepository) error {
	tools := []mcp.Tool{
		NewListTasksTool(taskRepo, planRepo),
		NewGetTaskTool(taskRepo, planRepo, commentRepo),
		NewCreateTaskTool(taskRepo, planRepo),
		NewUpdateTaskTool(taskRepo, planRepo),
		NewSearchTasksTool(taskRepo, planRepo),
		NewMoveTaskTool(taskRepo, planRepo),
	}

	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}
