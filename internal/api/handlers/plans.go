package handlers

import (
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/api/middleware"
	"github.com/goplan/goplan/internal/api/types"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/repository"
)

// PlanHandler handles plan-related HTTP requests.
type PlanHandler struct {
	*BaseHandler
	planRepo      repository.PlanRepository
	workspaceRepo repository.WorkspaceRepository
}

// NewPlanHandler creates a new plan handler.
func NewPlanHandler(planRepo repository.PlanRepository, workspaceRepo repository.WorkspaceRepository) *PlanHandler {
	return &PlanHandler{
		BaseHandler:   NewBaseHandler(),
		planRepo:      planRepo,
		workspaceRepo: workspaceRepo,
	}
}

// ListPlansByWorkspace handles GET /api/v1/workspaces/:wid/plans
func (h *PlanHandler) ListPlansByWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := extractWorkspaceIDFromPlanPath(r.URL.Path)

	if workspaceID == "" {
		h.WriteBadRequest(w, "workspace ID is required")
		return
	}

	pagination := h.ParsePagination(r)

	result, err := h.planRepo.GetByWorkspace(ctx, workspaceID, repository.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to response format
	responses := make([]*plan.PlanResponse, len(result.Items))
	for i, p := range result.Items {
		responses[i] = p.ToResponse()
	}

	h.WritePaginated(w, responses, result.TotalCount, result.Page, result.PageSize)
}

// CreatePlan handles POST /api/v1/workspaces/:wid/plans
func (h *PlanHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	workspaceID := extractWorkspaceIDFromPlanPath(r.URL.Path)

	if workspaceID == "" {
		h.WriteBadRequest(w, "workspace ID is required")
		return
	}

	// Verify workspace exists and get workspace data
	ws, err := h.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Get user's role in the workspace
	// First check if user is the workspace owner
	role := auth.GetUserRole(ctx)
	if role == "" {
		if ws.OwnerID == userID {
			role = "owner"
		} else {
			// Check if user is a member of the workspace
			member, err := h.workspaceRepo.GetMember(ctx, workspaceID, userID)
			if err == nil && member != nil {
				role = member.Role
			}
		}
	}

	// Check permissions - members and above can create plans
	if !middleware.CanEdit(role) {
		h.WriteForbidden(w, "you don't have permission to create plans")
		return
	}

	var req types.CreatePlanRequest
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

	// Create plan
	p := plan.NewPlan(id, workspaceID, req.Name, userID, req.Domain, req.Description)

	// Set optional dates
	if req.StartDate != nil {
		p.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		p.EndDate = req.EndDate
	}

	if err := h.planRepo.Create(ctx, p); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteCreated(w, p.ToResponse())
}

// GetPlan handles GET /api/v1/plans/:id
func (h *PlanHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	planID := extractPlanID(r.URL.Path)

	if planID == "" {
		h.WriteBadRequest(w, "plan ID is required")
		return
	}

	p, err := h.planRepo.GetByID(ctx, planID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, p.ToResponse())
}

// UpdatePlan handles PUT /api/v1/plans/:id
func (h *PlanHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	planID := extractPlanID(r.URL.Path)

	if planID == "" {
		h.WriteBadRequest(w, "plan ID is required")
		return
	}

	// Verify plan exists
	existingPlan, err := h.planRepo.GetByID(ctx, planID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Get user's role in the workspace
	role := auth.GetUserRole(ctx)
	if role == "" {
		ws, err := h.workspaceRepo.GetByID(ctx, existingPlan.WorkspaceID)
		if err == nil {
			if ws.OwnerID == userID {
				role = "owner"
			} else {
				member, err := h.workspaceRepo.GetMember(ctx, existingPlan.WorkspaceID, userID)
				if err == nil && member != nil {
					role = member.Role
				}
			}
		}
	}

	// Check permissions
	if !middleware.CanEdit(role) {
		h.WriteForbidden(w, "you don't have permission to update plans")
		return
	}

	var req types.UpdatePlanRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to domain input
	input := &plan.UpdatePlanInput{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Tags:        req.Tags,
	}

	updatedPlan, err := h.planRepo.Update(ctx, planID, input)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, updatedPlan.ToResponse())
}

// DeletePlan handles DELETE /api/v1/plans/:id
func (h *PlanHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	planID := extractPlanID(r.URL.Path)

	if planID == "" {
		h.WriteBadRequest(w, "plan ID is required")
		return
	}

	// Verify plan exists
	existingPlan, err := h.planRepo.GetByID(ctx, planID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Get user's role in the workspace
	role := auth.GetUserRole(ctx)
	if role == "" {
		ws, err := h.workspaceRepo.GetByID(ctx, existingPlan.WorkspaceID)
		if err == nil {
			if ws.OwnerID == userID {
				role = "owner"
			} else {
				member, err := h.workspaceRepo.GetMember(ctx, existingPlan.WorkspaceID, userID)
				if err == nil && member != nil {
					role = member.Role
				}
			}
		}
	}

	// Check permissions - only admins and owners can delete plans
	if !middleware.CanManage(role) {
		h.WriteForbidden(w, "you don't have permission to delete plans")
		return
	}

	if err := h.planRepo.Delete(ctx, planID); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteNoContent(w)
}

// ServeHTTP routes requests to the appropriate handler method.
func (h *PlanHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle /api/v1/workspaces/:wid/plans
	if strings.Contains(path, "/workspaces/") && strings.HasSuffix(path, "/plans") {
		switch r.Method {
		case http.MethodGet:
			h.ListPlansByWorkspace(w, r)
		case http.MethodPost:
			h.CreatePlan(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPost})
		}
		return
	}

	// Handle /api/v1/plans/:id
	if strings.HasPrefix(path, "/api/v1/plans/") {
		switch r.Method {
		case http.MethodGet:
			h.GetPlan(w, r)
		case http.MethodPut:
			h.UpdatePlan(w, r)
		case http.MethodDelete:
			h.DeletePlan(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPut, http.MethodDelete})
		}
		return
	}

	h.WriteNotFound(w, "endpoint")
}

// extractWorkspaceIDFromPlanPath extracts workspace ID from path like /api/v1/workspaces/{wid}/plans
func extractWorkspaceIDFromPlanPath(path string) string {
	prefix := "/api/v1/workspaces/"
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

// extractPlanID extracts plan ID from path like /api/v1/plans/{id}
func extractPlanID(path string) string {
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

