package tools

import (
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
)

// Repositories holds all repository dependencies for MCP tools.
type Repositories struct {
	Workspace   repository.WorkspaceRepository
	Plan        repository.PlanRepository
	Phase       repository.PhaseRepository
	Milestone   repository.MilestoneRepository
	Task        repository.TaskRepository
	Comment     repository.CommentRepository
	ActivityLog repository.ActivityLogRepository
}

// RegisterAllTools registers all MCP tools with the given registry.
func RegisterAllTools(registry *mcp.ToolRegistry, repos Repositories) error {
	// Register workspace tools
	if err := RegisterWorkspaceTools(registry, repos.Workspace); err != nil {
		return err
	}

	// Register plan tools
	if err := RegisterPlanTools(registry, repos.Plan, repos.Phase, repos.Milestone); err != nil {
		return err
	}

	// Register task tools
	if err := RegisterTaskTools(registry, repos.Task, repos.Plan, repos.Comment); err != nil {
		return err
	}

	// Register milestone tools
	if err := RegisterMilestoneTools(registry, repos.Milestone, repos.Plan); err != nil {
		return err
	}

	// Register activity tools
	if err := RegisterActivityTools(registry, repos.ActivityLog); err != nil {
		return err
	}

	return nil
}

// GetToolDefinitions returns definitions for all tools (for documentation/discovery).
func GetToolDefinitions() []mcp.ToolDefinition {
	return []mcp.ToolDefinition{
		// Workspace tools
		{
			Name:        ToolWorkspaceList,
			Description: "List all workspaces the current user is a member of",
			Arguments:   []mcp.ArgumentDef{},
		},
		{
			Name:        ToolWorkspaceGet,
			Description: "Get workspace details by ID, including members",
			Arguments: []mcp.ArgumentDef{
				{Name: "workspaceId", Type: "string", Required: false, Description: "Workspace ID (uses context if not provided)"},
			},
		},
		{
			Name:        ToolWorkspaceCreate,
			Description: "Create a new workspace",
			Arguments: []mcp.ArgumentDef{
				{Name: "name", Type: "string", Required: true, Description: "Workspace name"},
				{Name: "slug", Type: "string", Required: true, Description: "Workspace slug (URL-friendly identifier)"},
				{Name: "settings", Type: "object", Required: false, Description: "Workspace settings (defaultView, aiEnabled)"},
			},
		},
		// Plan tools
		{
			Name:        ToolPlanList,
			Description: "List all plans in a workspace with optional filtering",
			Arguments: []mcp.ArgumentDef{
				{Name: "workspaceId", Type: "string", Required: false, Description: "Workspace ID (uses context if not provided)"},
				{Name: "status", Type: "string", Required: false, Description: "Filter by status (draft, active, on_hold, completed, archived)"},
				{Name: "domain", Type: "string", Required: false, Description: "Filter by domain (software, event, ops, collection, generic)"},
				{Name: "tags", Type: "array", Required: false, Description: "Filter by tags (match any)"},
				{Name: "page", Type: "number", Required: false, Description: "Page number (default: 1)"},
				{Name: "pageSize", Type: "number", Required: false, Description: "Items per page (default: 20, max: 100)"},
			},
		},
		{
			Name:        ToolPlanGet,
			Description: "Get plan details including phases and milestones",
			Arguments: []mcp.ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
			},
		},
		{
			Name:        ToolPlanCreate,
			Description: "Create a new plan in a workspace",
			Arguments: []mcp.ArgumentDef{
				{Name: "name", Type: "string", Required: true, Description: "Plan name"},
				{Name: "domain", Type: "string", Required: true, Description: "Plan domain (software, event, ops, collection, generic)"},
				{Name: "workspaceId", Type: "string", Required: false, Description: "Workspace ID (uses context if not provided)"},
				{Name: "description", Type: "string", Required: false, Description: "Plan description"},
				{Name: "startDate", Type: "string", Required: false, Description: "Start date (YYYY-MM-DD)"},
				{Name: "endDate", Type: "string", Required: false, Description: "End date (YYYY-MM-DD)"},
				{Name: "tags", Type: "array", Required: false, Description: "Plan tags"},
			},
		},
		{
			Name:        ToolPlanUpdate,
			Description: "Update an existing plan's details",
			Arguments: []mcp.ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "name", Type: "string", Required: false, Description: "New plan name"},
				{Name: "description", Type: "string", Required: false, Description: "New description"},
				{Name: "status", Type: "string", Required: false, Description: "New status"},
				{Name: "startDate", Type: "string", Required: false, Description: "New start date"},
				{Name: "endDate", Type: "string", Required: false, Description: "New end date"},
				{Name: "tags", Type: "array", Required: false, Description: "New tags"},
			},
		},
		// Task tools
		{
			Name:        ToolTaskList,
			Description: "List tasks with optional filtering",
			Arguments: []mcp.ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "phaseId", Type: "string", Required: false, Description: "Filter by phase"},
				{Name: "status", Type: "string", Required: false, Description: "Filter by status"},
				{Name: "priority", Type: "string", Required: false, Description: "Filter by priority (low, medium, high, critical)"},
				{Name: "assigneeId", Type: "string", Required: false, Description: "Filter by assignee"},
				{Name: "tags", Type: "array", Required: false, Description: "Filter by tags"},
				{Name: "dueBefore", Type: "string", Required: false, Description: "Filter tasks due before date (YYYY-MM-DD)"},
				{Name: "dueAfter", Type: "string", Required: false, Description: "Filter tasks due after date (YYYY-MM-DD)"},
				{Name: "sortBy", Type: "string", Required: false, Description: "Sort field (position, createdAt, updatedAt, dueDate, priority)"},
				{Name: "sortOrder", Type: "string", Required: false, Description: "Sort order (asc, desc)"},
				{Name: "page", Type: "number", Required: false, Description: "Page number"},
				{Name: "pageSize", Type: "number", Required: false, Description: "Items per page"},
			},
		},
		{
			Name:        ToolTaskGet,
			Description: "Get task details including subtasks, dependencies, and comments",
			Arguments: []mcp.ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
			},
		},
		{
			Name:        ToolTaskCreate,
			Description: "Create a new task in a plan",
			Arguments: []mcp.ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "title", Type: "string", Required: true, Description: "Task title"},
				{Name: "description", Type: "string", Required: false, Description: "Task description"},
				{Name: "status", Type: "string", Required: false, Description: "Task status (defaults to plan default)"},
				{Name: "priority", Type: "string", Required: false, Description: "Priority (low, medium, high, critical)"},
				{Name: "phaseId", Type: "string", Required: false, Description: "Phase ID"},
				{Name: "parentId", Type: "string", Required: false, Description: "Parent task ID (for subtasks)"},
				{Name: "assigneeId", Type: "string", Required: false, Description: "Assignee user ID"},
				{Name: "dueDate", Type: "string", Required: false, Description: "Due date (YYYY-MM-DD)"},
				{Name: "estimatedHours", Type: "number", Required: false, Description: "Estimated hours"},
				{Name: "tags", Type: "array", Required: false, Description: "Task tags"},
				{Name: "customFieldValues", Type: "object", Required: false, Description: "Custom field values"},
			},
		},
		{
			Name:        ToolTaskUpdate,
			Description: "Update a task's properties",
			Arguments: []mcp.ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
				{Name: "title", Type: "string", Required: false, Description: "New title"},
				{Name: "description", Type: "string", Required: false, Description: "New description"},
				{Name: "status", Type: "string", Required: false, Description: "New status"},
				{Name: "priority", Type: "string", Required: false, Description: "New priority"},
				{Name: "phaseId", Type: "string", Required: false, Description: "New phase ID"},
				{Name: "assigneeId", Type: "string", Required: false, Description: "New assignee ID"},
				{Name: "dueDate", Type: "string", Required: false, Description: "New due date"},
				{Name: "estimatedHours", Type: "number", Required: false, Description: "New estimated hours"},
				{Name: "tags", Type: "array", Required: false, Description: "New tags"},
				{Name: "customFieldValues", Type: "object", Required: false, Description: "New custom field values"},
				{Name: "position", Type: "number", Required: false, Description: "New position"},
			},
		},
		{
			Name:        ToolTaskSearch,
			Description: "Search tasks by title and description",
			Arguments: []mcp.ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
				{Name: "page", Type: "number", Required: false, Description: "Page number"},
				{Name: "pageSize", Type: "number", Required: false, Description: "Items per page"},
			},
		},
		{
			Name:        ToolTaskMove,
			Description: "Move a task to a different status/position (Kanban)",
			Arguments: []mcp.ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
				{Name: "status", Type: "string", Required: true, Description: "New status"},
				{Name: "position", Type: "number", Required: false, Description: "New position in column"},
			},
		},
		// Milestone tools
		{
			Name:        ToolMilestoneList,
			Description: "List all milestones for a plan",
			Arguments: []mcp.ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "withinDays", Type: "number", Required: false, Description: "Filter upcoming milestones within N days"},
			},
		},
		{
			Name:        ToolMilestoneCreate,
			Description: "Create a new milestone in a plan",
			Arguments: []mcp.ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "name", Type: "string", Required: true, Description: "Milestone name"},
				{Name: "dueDate", Type: "string", Required: true, Description: "Due date (YYYY-MM-DD)"},
				{Name: "description", Type: "string", Required: false, Description: "Milestone description"},
			},
		},
		{
			Name:        ToolMilestoneUpdate,
			Description: "Update a milestone's properties",
			Arguments: []mcp.ArgumentDef{
				{Name: "milestoneId", Type: "string", Required: true, Description: "Milestone ID"},
				{Name: "name", Type: "string", Required: false, Description: "New name"},
				{Name: "dueDate", Type: "string", Required: false, Description: "New due date"},
				{Name: "description", Type: "string", Required: false, Description: "New description"},
				{Name: "status", Type: "string", Required: false, Description: "New status (pending, reached, missed)"},
			},
		},
		// Activity tools
		{
			Name:        ToolActivityGet,
			Description: "Get recent activity for a workspace, plan, or task",
			Arguments: []mcp.ArgumentDef{
				{Name: "workspaceId", Type: "string", Required: false, Description: "Workspace ID (uses context if not provided)"},
				{Name: "planId", Type: "string", Required: false, Description: "Filter by plan ID"},
				{Name: "taskId", Type: "string", Required: false, Description: "Filter by task ID"},
				{Name: "userId", Type: "string", Required: false, Description: "Filter by user ID"},
				{Name: "action", Type: "string", Required: false, Description: "Filter by action type"},
				{Name: "since", Type: "string", Required: false, Description: "Filter activities since timestamp (RFC3339)"},
				{Name: "until", Type: "string", Required: false, Description: "Filter activities until timestamp (RFC3339)"},
				{Name: "page", Type: "number", Required: false, Description: "Page number"},
				{Name: "pageSize", Type: "number", Required: false, Description: "Items per page"},
			},
		},
	}
}
