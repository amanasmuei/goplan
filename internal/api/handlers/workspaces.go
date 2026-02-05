package handlers

import (
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/api/middleware"
	"github.com/goplan/goplan/internal/api/types"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/domain/workspace"
	"github.com/goplan/goplan/internal/repository"
)

// WorkspaceHandler handles workspace-related HTTP requests.
type WorkspaceHandler struct {
	*BaseHandler
	workspaceRepo repository.WorkspaceRepository
	userRepo      repository.UserRepository
}

// NewWorkspaceHandler creates a new workspace handler.
func NewWorkspaceHandler(workspaceRepo repository.WorkspaceRepository, userRepo repository.UserRepository) *WorkspaceHandler {
	return &WorkspaceHandler{
		BaseHandler:   NewBaseHandler(),
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

// ListWorkspaces handles GET /api/v1/workspaces
func (h *WorkspaceHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)

	workspaces, err := h.workspaceRepo.GetUserWorkspaces(ctx, userID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to response format
	responses := make([]*workspace.WorkspaceResponse, len(workspaces))
	for i, ws := range workspaces {
		responses[i] = ws.ToResponse()
	}

	h.WriteSuccess(w, responses)
}

// CreateWorkspace handles POST /api/v1/workspaces
func (h *WorkspaceHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)

	var req types.CreateWorkspaceRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Check if slug already exists
	exists, err := h.workspaceRepo.ExistsBySlug(ctx, req.Slug)
	if err != nil {
		h.WriteError(w, err)
		return
	}
	if exists {
		h.WriteError(w, shared.NewAlreadyExistsError("workspace", "slug", req.Slug))
		return
	}

	// Generate a new ID (in production, use UUID)
	id := generateID()

	// Convert settings if provided
	var settings *workspace.WorkspaceSettings
	if req.Settings != nil {
		settings = &workspace.WorkspaceSettings{
			DefaultView: req.Settings.DefaultView,
			AIEnabled:   req.Settings.AIEnabled,
		}
	}

	// Create workspace
	ws := workspace.NewWorkspace(id, req.Name, req.Slug, userID, settings)

	if err := h.workspaceRepo.Create(ctx, ws); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteCreated(w, ws.ToResponse())
}

// GetWorkspace handles GET /api/v1/workspaces/:id
func (h *WorkspaceHandler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := extractWorkspaceID(r.URL.Path)

	if workspaceID == "" {
		h.WriteBadRequest(w, "workspace ID is required")
		return
	}

	ws, err := h.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, ws.ToResponse())
}

// UpdateWorkspace handles PUT /api/v1/workspaces/:id
func (h *WorkspaceHandler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	role := auth.GetUserRole(ctx)
	workspaceID := extractWorkspaceID(r.URL.Path)

	if workspaceID == "" {
		h.WriteBadRequest(w, "workspace ID is required")
		return
	}

	// Check permissions
	if !middleware.CanManage(role) {
		h.WriteForbidden(w, "you don't have permission to update this workspace")
		return
	}

	// Verify user is a member of the workspace
	member, err := h.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		h.WriteError(w, err)
		return
	}
	if member == nil {
		h.WriteForbidden(w, "you are not a member of this workspace")
		return
	}

	var req types.UpdateWorkspaceRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to domain input
	input := &workspace.UpdateWorkspaceInput{
		Name: req.Name,
	}

	if req.Settings != nil {
		input.Settings = &workspace.WorkspaceSettings{
			DefaultView: req.Settings.DefaultView,
			AIEnabled:   req.Settings.AIEnabled,
		}
	}

	updatedWs, err := h.workspaceRepo.Update(ctx, workspaceID, input)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, updatedWs.ToResponse())
}

// ListMembers handles GET /api/v1/workspaces/:id/members
func (h *WorkspaceHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceID := extractWorkspaceIDForMembers(r.URL.Path)

	if workspaceID == "" {
		h.WriteBadRequest(w, "workspace ID is required")
		return
	}

	members, err := h.workspaceRepo.ListMembers(ctx, workspaceID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to response format
	responses := make([]*workspace.WorkspaceMemberResponse, len(members))
	for i, m := range members {
		responses[i] = m.ToResponse()
	}

	h.WriteSuccess(w, responses)
}

// AddMember handles POST /api/v1/workspaces/:id/members
func (h *WorkspaceHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	role := auth.GetUserRole(ctx)
	workspaceID := extractWorkspaceIDForMembers(r.URL.Path)

	if workspaceID == "" {
		h.WriteBadRequest(w, "workspace ID is required")
		return
	}

	// Check permissions - only admins and owners can add members
	if !middleware.CanManage(role) {
		h.WriteForbidden(w, "you don't have permission to add members")
		return
	}

	// Verify user is a member of the workspace
	member, err := h.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		h.WriteError(w, err)
		return
	}
	if member == nil || !member.CanManageMembers() {
		h.WriteForbidden(w, "you don't have permission to manage members")
		return
	}

	var req types.AddMemberRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Verify the user to be added exists
	_, err = h.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		h.WriteError(w, shared.NewNotFoundError("user", req.UserID))
		return
	}

	// Check if user is already a member
	existingMember, err := h.workspaceRepo.GetMember(ctx, workspaceID, req.UserID)
	if err != nil && !isNotFoundError(err) {
		h.WriteError(w, err)
		return
	}
	if existingMember != nil {
		h.WriteError(w, shared.NewAlreadyExistsError("member", "userId", req.UserID))
		return
	}

	// Create member
	newMember := workspace.NewWorkspaceMember(workspaceID, req.UserID, req.Role)

	if err := h.workspaceRepo.AddMember(ctx, newMember); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteCreated(w, newMember.ToResponse())
}

// ServeHTTP routes requests to the appropriate handler method.
func (h *WorkspaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle /api/v1/workspaces
	if path == "/api/v1/workspaces" || path == "/api/v1/workspaces/" {
		switch r.Method {
		case http.MethodGet:
			h.ListWorkspaces(w, r)
		case http.MethodPost:
			h.CreateWorkspace(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPost})
		}
		return
	}

	// Handle /api/v1/workspaces/:id/members
	if strings.Contains(path, "/members") {
		switch r.Method {
		case http.MethodGet:
			h.ListMembers(w, r)
		case http.MethodPost:
			h.AddMember(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPost})
		}
		return
	}

	// Handle /api/v1/workspaces/:id
	switch r.Method {
	case http.MethodGet:
		h.GetWorkspace(w, r)
	case http.MethodPut:
		h.UpdateWorkspace(w, r)
	default:
		h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPut})
	}
}

// extractWorkspaceID extracts workspace ID from path like /api/v1/workspaces/{id}
func extractWorkspaceID(path string) string {
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

// extractWorkspaceIDForMembers extracts workspace ID from path like /api/v1/workspaces/{id}/members
func extractWorkspaceIDForMembers(path string) string {
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

// isNotFoundError checks if the error is a not found error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var domainErr *shared.DomainError
	if de, ok := err.(*shared.DomainError); ok {
		domainErr = de
		return domainErr.Code == "ENTITY_NOT_FOUND"
	}
	return strings.Contains(err.Error(), "not found")
}
