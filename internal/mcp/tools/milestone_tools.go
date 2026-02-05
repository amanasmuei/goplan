package tools

import (
	"context"

	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
	"github.com/google/uuid"
)

// Milestone tool names
const (
	ToolMilestoneList   = "milestone.list"
	ToolMilestoneCreate = "milestone.create"
	ToolMilestoneUpdate = "milestone.update"
)

// ListMilestonesTool lists milestones for a plan.
type ListMilestonesTool struct {
	milestoneRepo repository.MilestoneRepository
}

// NewListMilestonesTool creates a new ListMilestonesTool.
func NewListMilestonesTool(milestoneRepo repository.MilestoneRepository) *ListMilestonesTool {
	return &ListMilestonesTool{milestoneRepo: milestoneRepo}
}

// Name returns the tool name.
func (t *ListMilestonesTool) Name() string {
	return ToolMilestoneList
}

// Description returns the tool description.
func (t *ListMilestonesTool) Description() string {
	return "List all milestones for a plan"
}

// Execute executes the tool.
func (t *ListMilestonesTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	planID, err := getRequiredString(args, "planId")
	if err != nil {
		// Try from execution context
		if execCtx.PlanID != nil && *execCtx.PlanID != "" {
			planID = *execCtx.PlanID
		} else {
			return nil, err
		}
	}

	// Check if we should filter upcoming only
	withinDays := getOptionalInt(args, "withinDays")

	var milestones []*plan.Milestone
	if withinDays != nil && *withinDays > 0 {
		milestones, err = t.milestoneRepo.GetUpcoming(ctx, planID, *withinDays)
	} else {
		milestones, err = t.milestoneRepo.ListByPlan(ctx, planID)
	}
	if err != nil {
		return nil, err
	}

	// Convert to response format
	results := make([]*plan.MilestoneResponse, len(milestones))
	for i, m := range milestones {
		results[i] = m.ToResponse()
	}

	return map[string]interface{}{
		"milestones": results,
		"count":      len(results),
	}, nil
}

// CreateMilestoneTool creates a new milestone.
type CreateMilestoneTool struct {
	milestoneRepo repository.MilestoneRepository
	planRepo      repository.PlanRepository
}

// NewCreateMilestoneTool creates a new CreateMilestoneTool.
func NewCreateMilestoneTool(milestoneRepo repository.MilestoneRepository, planRepo repository.PlanRepository) *CreateMilestoneTool {
	return &CreateMilestoneTool{
		milestoneRepo: milestoneRepo,
		planRepo:      planRepo,
	}
}

// Name returns the tool name.
func (t *CreateMilestoneTool) Name() string {
	return ToolMilestoneCreate
}

// Description returns the tool description.
func (t *CreateMilestoneTool) Description() string {
	return "Create a new milestone in a plan"
}

// Execute executes the tool.
func (t *CreateMilestoneTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	planID, err := getRequiredString(args, "planId")
	if err != nil {
		// Try from execution context
		if execCtx.PlanID != nil && *execCtx.PlanID != "" {
			planID = *execCtx.PlanID
		} else {
			return nil, err
		}
	}

	// Verify plan exists
	_, err = t.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	name, err := getRequiredString(args, "name")
	if err != nil {
		return nil, err
	}

	dueDate, err := getRequiredString(args, "dueDate")
	if err != nil {
		return nil, err
	}

	description := getOptionalString(args, "description")

	// Create milestone
	m := plan.NewMilestone(
		uuid.New().String(),
		planID,
		name,
		dueDate,
		description,
	)

	// Validate
	if validErrs := m.Validate(); validErrs != nil {
		return nil, shared.NewValidationError("milestone", validErrs.Error())
	}

	if err := t.milestoneRepo.Create(ctx, m); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"milestone": m.ToResponse(),
		"message":   "Milestone created successfully",
	}, nil
}

// UpdateMilestoneTool updates a milestone.
type UpdateMilestoneTool struct {
	milestoneRepo repository.MilestoneRepository
}

// NewUpdateMilestoneTool creates a new UpdateMilestoneTool.
func NewUpdateMilestoneTool(milestoneRepo repository.MilestoneRepository) *UpdateMilestoneTool {
	return &UpdateMilestoneTool{milestoneRepo: milestoneRepo}
}

// Name returns the tool name.
func (t *UpdateMilestoneTool) Name() string {
	return ToolMilestoneUpdate
}

// Description returns the tool description.
func (t *UpdateMilestoneTool) Description() string {
	return "Update a milestone's name, due date, description, or status"
}

// Execute executes the tool.
func (t *UpdateMilestoneTool) Execute(ctx context.Context, execCtx mcp.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	milestoneID, err := getRequiredString(args, "milestoneId")
	if err != nil {
		return nil, err
	}

	// Get existing milestone
	existing, err := t.milestoneRepo.GetByID(ctx, milestoneID)
	if err != nil {
		return nil, err
	}

	// Handle status update separately
	if status := getOptionalString(args, "status"); status != nil {
		validStatuses := map[string]bool{
			shared.MilestoneStatusPending: true,
			shared.MilestoneStatusReached: true,
			shared.MilestoneStatusMissed:  true,
		}
		if !validStatuses[*status] {
			return nil, shared.NewValidationError("status", "invalid milestone status. Valid values: pending, reached, missed")
		}
		if err := t.milestoneRepo.UpdateStatus(ctx, milestoneID, *status); err != nil {
			return nil, err
		}
		existing.Status = *status
	}

	// Update name/dueDate/description if provided
	name := existing.Name
	if n := getOptionalString(args, "name"); n != nil {
		name = *n
	}

	dueDate := existing.DueDate
	if d := getOptionalString(args, "dueDate"); d != nil {
		dueDate = *d
	}

	description := existing.Description
	if desc := getOptionalString(args, "description"); desc != nil {
		description = desc
	}

	// Update if any field changed
	if name != existing.Name || dueDate != existing.DueDate || description != existing.Description {
		updated, err := t.milestoneRepo.Update(ctx, milestoneID, name, dueDate, description)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"milestone": updated.ToResponse(),
			"message":   "Milestone updated successfully",
		}, nil
	}

	// If only status was updated, return the existing with updated status
	return map[string]interface{}{
		"milestone": existing.ToResponse(),
		"message":   "Milestone updated successfully",
	}, nil
}

// RegisterMilestoneTools registers all milestone tools with the registry.
func RegisterMilestoneTools(registry *mcp.ToolRegistry, milestoneRepo repository.MilestoneRepository, planRepo repository.PlanRepository) error {
	tools := []mcp.Tool{
		NewListMilestonesTool(milestoneRepo),
		NewCreateMilestoneTool(milestoneRepo, planRepo),
		NewUpdateMilestoneTool(milestoneRepo),
	}

	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			return err
		}
	}
	return nil
}
