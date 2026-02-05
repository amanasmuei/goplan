package tools

import (
	"context"

	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/workspace"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
	"github.com/google/uuid"
)

// Workspace tool names
const (
	ToolWorkspaceList   = "workspace.list"
	ToolWorkspaceGet    = "workspace.get"
	ToolWorkspaceCreate = "workspace.create"
)

// ListWorkspacesTool lists workspaces for a user.
type ListWorkspacesTool struct {
	repo repository.WorkspaceRepository
}

// NewListWorkspacesTool creates a new ListWorkspacesTool.
func NewListWorkspacesTool(repo repository.WorkspaceRepository) *ListWorkspacesTool {
	return &ListWorkspacesTool{repo: repo}
}

// Name returns the tool name.
func (t *ListWorkspacesTool) Name() string {
	return ToolWorkspaceList
}

// Description returns the tool description.
func (t *ListWorkspacesTool) Description() string {
	return "List all workspaces the current user is a member of"
}

// Execute executes the tool.
func (t *ListWorkspacesTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	if execCtx.UserID == "" {
		return nil, shared.NewUnauthorizedError("user ID is required")
	}

	workspaces, err := t.repo.GetUserWorkspaces(ctx, execCtx.UserID)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	results := make([]*workspace.WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		results[i] = ws.ToResponse()
	}

	return map[string]interface{}{
		"workspaces": results,
		"count":      len(results),
	}, nil
}

// GetWorkspaceTool retrieves workspace details.
type GetWorkspaceTool struct {
	repo repository.WorkspaceRepository
}

// NewGetWorkspaceTool creates a new GetWorkspaceTool.
func NewGetWorkspaceTool(repo repository.WorkspaceRepository) *GetWorkspaceTool {
	return &GetWorkspaceTool{repo: repo}
}

// Name returns the tool name.
func (t *GetWorkspaceTool) Name() string {
	return ToolWorkspaceGet
}

// Description returns the tool description.
func (t *GetWorkspaceTool) Description() string {
	return "Get workspace details by ID, including members"
}

// Execute executes the tool.
func (t *GetWorkspaceTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	workspaceID, err := getRequiredString(args, "workspaceId")
	if err != nil {
		// Try from execution context
		if execCtx.WorkspaceID != "" {
			workspaceID = execCtx.WorkspaceID
		} else {
			return nil, err
		}
	}

	// Get workspace
	ws, err := t.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Get members
	members, err := t.repo.ListMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	memberResponses := make([]*workspace.WorkspaceMemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = m.ToResponse()
	}

	return map[string]interface{}{
		"workspace": ws.ToResponse(),
		"members":   memberResponses,
	}, nil
}

// CreateWorkspaceTool creates a new workspace.
type CreateWorkspaceTool struct {
	repo repository.WorkspaceRepository
}

// NewCreateWorkspaceTool creates a new CreateWorkspaceTool.
func NewCreateWorkspaceTool(repo repository.WorkspaceRepository) *CreateWorkspaceTool {
	return &CreateWorkspaceTool{repo: repo}
}

// Name returns the tool name.
func (t *CreateWorkspaceTool) Name() string {
	return ToolWorkspaceCreate
}

// Description returns the tool description.
func (t *CreateWorkspaceTool) Description() string {
	return "Create a new workspace"
}

// Execute executes the tool.
func (t *CreateWorkspaceTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	if execCtx.UserID == "" {
		return nil, shared.NewUnauthorizedError("user ID is required")
	}

	name, err := getRequiredString(args, "name")
	if err != nil {
		return nil, err
	}

	slug, err := getRequiredString(args, "slug")
	if err != nil {
		return nil, err
	}

	// Check if slug already exists
	exists, err := t.repo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, shared.NewAlreadyExistsError("workspace", "slug", slug)
	}

	// Parse optional settings
	var settings *workspace.WorkspaceSettings
	if settingsMap := getMap(args, "settings"); settingsMap != nil {
		settings = &workspace.WorkspaceSettings{
			DefaultView: shared.ViewKanban,
			AIEnabled:   true,
		}
		if view := getOptionalString(settingsMap, "defaultView"); view != nil {
			if shared.IsValidDefaultView(*view) {
				settings.DefaultView = *view
			}
		}
		if aiEnabled, ok := settingsMap["aiEnabled"].(bool); ok {
			settings.AIEnabled = aiEnabled
		}
	}

	// Create workspace
	ws := workspace.NewWorkspace(
		uuid.New().String(),
		name,
		slug,
		execCtx.UserID,
		settings,
	)

	// Validate
	if validErrs := ws.Validate(); validErrs != nil {
		return nil, shared.NewValidationError("workspace", validErrs.Error())
	}

	if err := t.repo.Create(ctx, ws); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"workspace": ws.ToResponse(),
		"message":   "Workspace created successfully",
	}, nil
}

// RegisterWorkspaceTools registers all workspace tools with the registry.
func RegisterWorkspaceTools(registry *mcp.ToolRegistry, workspaceRepo repository.WorkspaceRepository) error {
	tools := []mcp.Tool{
		NewListWorkspacesTool(workspaceRepo),
		NewGetWorkspaceTool(workspaceRepo),
		NewCreateWorkspaceTool(workspaceRepo),
	}

	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}
