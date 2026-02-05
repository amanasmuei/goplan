// Package mcp provides MCP (Model Context Protocol) tool registry.
package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/goplan/goplan/internal/domain/shared"
)

// Tool represents an MCP tool that can be executed.
type Tool interface {
	// Name returns the tool name (e.g., "plan.create", "task.update").
	Name() string

	// Description returns a human-readable description of the tool.
	Description() string

	// Execute executes the tool with the given context and arguments.
	Execute(ctx context.Context, execCtx ExecutionContext, args map[string]interface{}) (interface{}, error)
}

// ExecutionContext provides context for tool execution.
type ExecutionContext struct {
	UserID      string
	WorkspaceID string
	PlanID      *string
	Role        string
}

// ToolRegistry manages registered MCP tools.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewToolRegistry creates a new ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool with the registry.
func (r *ToolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name.
func (r *ToolRegistry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, shared.NewNotFoundError("tool", name)
	}

	return tool, nil
}

// ExecuteTool executes a tool by name with the given arguments.
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, execCtx ExecutionContext, args map[string]interface{}) (interface{}, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	return tool.Execute(ctx, execCtx, args)
}

// List returns all registered tool names.
func (r *ToolRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// ListTools returns all registered tools with their descriptions.
func (r *ToolRegistry) ListTools() []ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		infos = append(infos, ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
		})
	}
	return infos
}

// ToolInfo provides basic information about a tool.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Has checks if a tool is registered.
func (r *ToolRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.tools[name]
	return exists
}

// Unregister removes a tool from the registry.
func (r *ToolRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; !exists {
		return shared.NewNotFoundError("tool", name)
	}

	delete(r.tools, name)
	return nil
}

// Standard tool names
const (
	// Workspace tools
	ToolWorkspaceList   = "workspace.list"
	ToolWorkspaceGet    = "workspace.get"
	ToolWorkspaceCreate = "workspace.create"

	// Plan tools
	ToolPlanCreate  = "plan.create"
	ToolPlanUpdate  = "plan.update"
	ToolPlanGet     = "plan.get"
	ToolPlanList    = "plan.list"
	ToolPlanArchive = "plan.archive"

	// Task tools
	ToolTaskCreate = "task.create"
	ToolTaskUpdate = "task.update"
	ToolTaskMove   = "task.move"
	ToolTaskAssign = "task.assign"
	ToolTaskGet    = "task.get"
	ToolTaskList   = "task.list"
	ToolTaskSearch = "task.search"
	ToolTaskDelete = "task.delete"

	// Milestone tools
	ToolMilestoneList   = "milestone.list"
	ToolMilestoneCreate = "milestone.create"
	ToolMilestoneUpdate = "milestone.update"

	// Activity tools
	ToolActivityGet = "activity.get"

	// AI tools
	ToolAISuggestTasks = "ai.suggest_tasks"
	ToolAIGeneratePlan = "ai.generate_plan"
	ToolAISummarize    = "ai.summarize"
)

// ToolDefinition provides metadata about a tool for documentation.
type ToolDefinition struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Arguments   []ArgumentDef     `json:"arguments"`
}

// ArgumentDef defines an argument for a tool.
type ArgumentDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// GetToolDefinitions returns the definitions of all standard tools.
func GetToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolPlanCreate,
			Description: "Create a new plan",
			Arguments: []ArgumentDef{
				{Name: "name", Type: "string", Required: true, Description: "Plan name"},
				{Name: "domain", Type: "string", Required: true, Description: "Plan domain (software, event, ops, collection, generic)"},
				{Name: "description", Type: "string", Required: false, Description: "Plan description"},
				{Name: "startDate", Type: "string", Required: false, Description: "Start date (YYYY-MM-DD)"},
				{Name: "endDate", Type: "string", Required: false, Description: "End date (YYYY-MM-DD)"},
			},
		},
		{
			Name:        ToolPlanUpdate,
			Description: "Update an existing plan",
			Arguments: []ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "fields", Type: "object", Required: true, Description: "Fields to update"},
			},
		},
		{
			Name:        ToolPlanGet,
			Description: "Get plan details",
			Arguments: []ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
			},
		},
		{
			Name:        ToolPlanList,
			Description: "List plans in a workspace",
			Arguments: []ArgumentDef{
				{Name: "workspaceId", Type: "string", Required: true, Description: "Workspace ID"},
				{Name: "filters", Type: "object", Required: false, Description: "Filter options"},
			},
		},
		{
			Name:        ToolPlanArchive,
			Description: "Archive a plan",
			Arguments: []ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
			},
		},
		{
			Name:        ToolTaskCreate,
			Description: "Create a new task",
			Arguments: []ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "title", Type: "string", Required: true, Description: "Task title"},
				{Name: "description", Type: "string", Required: false, Description: "Task description"},
				{Name: "priority", Type: "string", Required: false, Description: "Task priority"},
				{Name: "dueDate", Type: "string", Required: false, Description: "Due date"},
				{Name: "assigneeId", Type: "string", Required: false, Description: "Assignee user ID"},
			},
		},
		{
			Name:        ToolTaskUpdate,
			Description: "Update an existing task",
			Arguments: []ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
				{Name: "fields", Type: "object", Required: true, Description: "Fields to update"},
			},
		},
		{
			Name:        ToolTaskMove,
			Description: "Move a task to a different status",
			Arguments: []ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
				{Name: "status", Type: "string", Required: true, Description: "New status"},
				{Name: "position", Type: "number", Required: false, Description: "Position in the status column"},
			},
		},
		{
			Name:        ToolTaskAssign,
			Description: "Assign a task to a user",
			Arguments: []ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
				{Name: "assigneeId", Type: "string", Required: true, Description: "User ID to assign"},
			},
		},
		{
			Name:        ToolTaskGet,
			Description: "Get task details",
			Arguments: []ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
			},
		},
		{
			Name:        ToolTaskList,
			Description: "List tasks in a plan",
			Arguments: []ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "filters", Type: "object", Required: false, Description: "Filter options"},
			},
		},
		{
			Name:        ToolTaskDelete,
			Description: "Delete a task",
			Arguments: []ArgumentDef{
				{Name: "taskId", Type: "string", Required: true, Description: "Task ID"},
			},
		},
		{
			Name:        ToolAISuggestTasks,
			Description: "Generate task suggestions using AI",
			Arguments: []ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "basedOn", Type: "string", Required: false, Description: "Base suggestions on: goal, existing_tasks, template"},
				{Name: "limit", Type: "number", Required: false, Description: "Maximum number of suggestions"},
			},
		},
		{
			Name:        ToolAIGeneratePlan,
			Description: "Generate a plan structure using AI",
			Arguments: []ArgumentDef{
				{Name: "goal", Type: "string", Required: true, Description: "Plan goal"},
				{Name: "planType", Type: "string", Required: true, Description: "Plan type"},
				{Name: "timeline", Type: "string", Required: false, Description: "Expected timeline"},
			},
		},
		{
			Name:        ToolAISummarize,
			Description: "Generate a summary of a plan",
			Arguments: []ArgumentDef{
				{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
				{Name: "timeRange", Type: "string", Required: false, Description: "Time range: week, month, all"},
			},
		},
	}
}
