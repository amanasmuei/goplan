package tools

import (
	"context"

	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
	"github.com/google/uuid"
)

// Plan tool names
const (
	ToolPlanList   = "plan.list"
	ToolPlanGet    = "plan.get"
	ToolPlanCreate = "plan.create"
	ToolPlanUpdate = "plan.update"
)

// ListPlansTool lists plans in a workspace.
type ListPlansTool struct {
	planRepo  repository.PlanRepository
	phaseRepo repository.PhaseRepository
}

// NewListPlansTool creates a new ListPlansTool.
func NewListPlansTool(planRepo repository.PlanRepository, phaseRepo repository.PhaseRepository) *ListPlansTool {
	return &ListPlansTool{
		planRepo:  planRepo,
		phaseRepo: phaseRepo,
	}
}

// Name returns the tool name.
func (t *ListPlansTool) Name() string {
	return ToolPlanList
}

// Description returns the tool description.
func (t *ListPlansTool) Description() string {
	return "List all plans in a workspace with optional filtering"
}

// Execute executes the tool.
func (t *ListPlansTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	// Always use the authenticated user's workspace; ignore client-supplied workspaceId
	workspaceID := execCtx.WorkspaceID

	if workspaceID == "" {
		return nil, shared.NewValidationError("workspaceId", "workspace ID is required")
	}

	// Parse pagination
	page := getInt(args, "page", 1)
	pageSize := getInt(args, "pageSize", 20)
	pagination := repository.Pagination{Page: page, PageSize: pageSize}.Normalize()

	// Parse filters
	filter := repository.PlanFilterOptions{
		WorkspaceID: &workspaceID,
	}
	if status := getOptionalString(args, "status"); status != nil {
		filter.Status = status
	}
	if domain := getOptionalString(args, "domain"); domain != nil {
		filter.Domain = domain
	}
	if tags := getStringSlice(args, "tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	// Get plans
	result, err := t.planRepo.List(ctx, filter, repository.DefaultPlanSort(), pagination)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	plans := make([]*plan.PlanResponse, len(result.Items))
	for i, p := range result.Items {
		plans[i] = p.ToResponse()
	}

	return map[string]interface{}{
		"plans":      plans,
		"totalCount": result.TotalCount,
		"page":       result.Page,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
	}, nil
}

// GetPlanTool retrieves plan details with phases and milestones.
type GetPlanTool struct {
	planRepo      repository.PlanRepository
	phaseRepo     repository.PhaseRepository
	milestoneRepo repository.MilestoneRepository
}

// NewGetPlanTool creates a new GetPlanTool.
func NewGetPlanTool(planRepo repository.PlanRepository, phaseRepo repository.PhaseRepository, milestoneRepo repository.MilestoneRepository) *GetPlanTool {
	return &GetPlanTool{
		planRepo:      planRepo,
		phaseRepo:     phaseRepo,
		milestoneRepo: milestoneRepo,
	}
}

// Name returns the tool name.
func (t *GetPlanTool) Name() string {
	return ToolPlanGet
}

// Description returns the tool description.
func (t *GetPlanTool) Description() string {
	return "Get plan details including phases and milestones"
}

// Execute executes the tool.
func (t *GetPlanTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	planID, err := getRequiredString(args, "planId")
	if err != nil {
		// Try from execution context
		if execCtx.PlanID != nil && *execCtx.PlanID != "" {
			planID = *execCtx.PlanID
		} else {
			return nil, err
		}
	}

	// Get plan
	p, err := t.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Verify workspace ownership
	if p.WorkspaceID != execCtx.WorkspaceID {
		return nil, shared.NewForbiddenError("access", "plan")
	}

	// Get phases
	phases, err := t.phaseRepo.ListByPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	phaseResponses := make([]*plan.PhaseResponse, len(phases))
	for i, ph := range phases {
		phaseResponses[i] = ph.ToResponse()
	}

	// Get milestones
	milestones, err := t.milestoneRepo.ListByPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	milestoneResponses := make([]*plan.MilestoneResponse, len(milestones))
	for i, m := range milestones {
		milestoneResponses[i] = m.ToResponse()
	}

	return map[string]interface{}{
		"plan":       p.ToResponse(),
		"phases":     phaseResponses,
		"milestones": milestoneResponses,
	}, nil
}

// CreatePlanTool creates a new plan.
type CreatePlanTool struct {
	planRepo repository.PlanRepository
}

// NewCreatePlanTool creates a new CreatePlanTool.
func NewCreatePlanTool(planRepo repository.PlanRepository) *CreatePlanTool {
	return &CreatePlanTool{planRepo: planRepo}
}

// Name returns the tool name.
func (t *CreatePlanTool) Name() string {
	return ToolPlanCreate
}

// Description returns the tool description.
func (t *CreatePlanTool) Description() string {
	return "Create a new plan in a workspace"
}

// Execute executes the tool.
func (t *CreatePlanTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	if execCtx.UserID == "" {
		return nil, shared.NewUnauthorizedError("user ID is required")
	}

	// Always use the authenticated user's workspace; ignore client-supplied workspaceId
	workspaceID := execCtx.WorkspaceID
	if workspaceID == "" {
		return nil, shared.NewValidationError("workspaceId", "workspace ID is required")
	}

	name, err := getRequiredString(args, "name")
	if err != nil {
		return nil, err
	}

	domain, err := getRequiredString(args, "domain")
	if err != nil {
		return nil, err
	}
	if !shared.IsValidPlanDomain(domain) {
		return nil, shared.NewValidationError("domain", "invalid plan domain. Valid values: software, event, ops, collection, generic")
	}

	description := getOptionalString(args, "description")

	// Create plan
	p := plan.NewPlan(
		uuid.New().String(),
		workspaceID,
		name,
		execCtx.UserID,
		domain,
		description,
	)

	// Set optional dates
	if startDate := getOptionalString(args, "startDate"); startDate != nil {
		p.StartDate = startDate
	}
	if endDate := getOptionalString(args, "endDate"); endDate != nil {
		p.EndDate = endDate
	}

	// Set tags if provided
	if tags := getStringSlice(args, "tags"); len(tags) > 0 {
		p.Tags = tags
	}

	// Validate
	if validErrs := p.Validate(); validErrs != nil {
		return nil, shared.NewValidationError("plan", validErrs.Error())
	}

	if err := t.planRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"plan":    p.ToResponse(),
		"message": "Plan created successfully",
	}, nil
}

// UpdatePlanTool updates a plan.
type UpdatePlanTool struct {
	planRepo repository.PlanRepository
}

// NewUpdatePlanTool creates a new UpdatePlanTool.
func NewUpdatePlanTool(planRepo repository.PlanRepository) *UpdatePlanTool {
	return &UpdatePlanTool{planRepo: planRepo}
}

// Name returns the tool name.
func (t *UpdatePlanTool) Name() string {
	return ToolPlanUpdate
}

// Description returns the tool description.
func (t *UpdatePlanTool) Description() string {
	return "Update an existing plan's details"
}

// Execute executes the tool.
func (t *UpdatePlanTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	planID, err := getRequiredString(args, "planId")
	if err != nil {
		return nil, err
	}

	// Verify workspace ownership
	p, err := t.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	if p.WorkspaceID != execCtx.WorkspaceID {
		return nil, shared.NewForbiddenError("access", "plan")
	}

	// Build update input
	input := &plan.UpdatePlanInput{}
	hasUpdates := false

	if name := getOptionalString(args, "name"); name != nil {
		input.Name = name
		hasUpdates = true
	}
	if description := getOptionalString(args, "description"); description != nil {
		input.Description = description
		hasUpdates = true
	}
	if status := getOptionalString(args, "status"); status != nil {
		if !shared.IsValidPlanStatus(*status) {
			return nil, shared.NewValidationError("status", "invalid plan status. Valid values: draft, active, on_hold, completed, archived")
		}
		input.Status = status
		hasUpdates = true
	}
	if startDate := getOptionalString(args, "startDate"); startDate != nil {
		input.StartDate = startDate
		hasUpdates = true
	}
	if endDate := getOptionalString(args, "endDate"); endDate != nil {
		input.EndDate = endDate
		hasUpdates = true
	}
	if tags := getStringSlice(args, "tags"); tags != nil {
		input.Tags = tags
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, shared.NewValidationError("fields", "at least one field must be provided for update")
	}

	// Update plan
	updated, err := t.planRepo.Update(ctx, planID, input)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"plan":    updated.ToResponse(),
		"message": "Plan updated successfully",
	}, nil
}

// RegisterPlanTools registers all plan tools with the registry.
func RegisterPlanTools(registry *mcp.ToolRegistry, planRepo repository.PlanRepository, phaseRepo repository.PhaseRepository, milestoneRepo repository.MilestoneRepository) error {
	tools := []mcp.Tool{
		NewListPlansTool(planRepo, phaseRepo),
		NewGetPlanTool(planRepo, phaseRepo, milestoneRepo),
		NewCreatePlanTool(planRepo),
		NewUpdatePlanTool(planRepo),
	}

	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}
