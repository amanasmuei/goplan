package tools

import (
	"context"
	"time"

	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
)

// Activity tool names
const (
	ToolActivityGet = "activity.get"
)

// GetActivityTool retrieves recent activity.
type GetActivityTool struct {
	activityRepo repository.ActivityLogRepository
}

// NewGetActivityTool creates a new GetActivityTool.
func NewGetActivityTool(activityRepo repository.ActivityLogRepository) *GetActivityTool {
	return &GetActivityTool{activityRepo: activityRepo}
}

// Name returns the tool name.
func (t *GetActivityTool) Name() string {
	return ToolActivityGet
}

// Description returns the tool description.
func (t *GetActivityTool) Description() string {
	return "Get recent activity for a workspace, plan, or task"
}

// Execute executes the tool.
func (t *GetActivityTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	// Parse pagination
	page := getInt(args, "page", 1)
	pageSize := getInt(args, "pageSize", 20)
	pagination := repository.Pagination{Page: page, PageSize: pageSize}.Normalize()

	// Build filter
	filter := repository.ActivityFilterOptions{}

	// Determine scope: task > plan > workspace
	if taskID := getOptionalString(args, "taskId"); taskID != nil {
		filter.TaskID = taskID
		result, err := t.activityRepo.ListByTask(ctx, *taskID, pagination)
		if err != nil {
			return nil, err
		}
		return t.formatResult(result), nil
	}

	if planID := getOptionalString(args, "planId"); planID != nil {
		filter.PlanID = planID
		result, err := t.activityRepo.ListByPlan(ctx, *planID, pagination)
		if err != nil {
			return nil, err
		}
		return t.formatResult(result), nil
	}

	// Use workspace from args or execution context
	workspaceID := execCtx.WorkspaceID
	if ws := getOptionalString(args, "workspaceId"); ws != nil {
		workspaceID = *ws
	}
	if workspaceID == "" {
		return nil, shared.NewValidationError("scope", "at least one of workspaceId, planId, or taskId is required")
	}

	// Optional action filter
	if action := getOptionalString(args, "action"); action != nil {
		filter.Action = action
	}

	// Optional user filter
	if userID := getOptionalString(args, "userId"); userID != nil {
		filter.UserID = userID
	}

	// Optional time range
	if since := getOptionalString(args, "since"); since != nil {
		if t, err := time.Parse(time.RFC3339, *since); err == nil {
			filter.Since = &t
		}
	}
	if until := getOptionalString(args, "until"); until != nil {
		if t, err := time.Parse(time.RFC3339, *until); err == nil {
			filter.Until = &t
		}
	}

	filter.WorkspaceID = &workspaceID
	result, err := t.activityRepo.List(ctx, filter, pagination)
	if err != nil {
		return nil, err
	}

	return t.formatResult(result), nil
}

func (t *GetActivityTool) formatResult(result *repository.PaginatedResult[task.ActivityLog]) map[string]interface{} {
	activities := make([]*task.ActivityLogResponse, len(result.Items))
	for i, a := range result.Items {
		activities[i] = a.ToResponse()
	}

	return map[string]interface{}{
		"activities": activities,
		"totalCount": result.TotalCount,
		"page":       result.Page,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
	}
}

// RegisterActivityTools registers all activity tools with the registry.
func RegisterActivityTools(registry *mcp.ToolRegistry, activityRepo repository.ActivityLogRepository) error {
	tools := []mcp.Tool{
		NewGetActivityTool(activityRepo),
	}

	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}
