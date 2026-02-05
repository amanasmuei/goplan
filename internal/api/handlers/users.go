// Package handlers provides HTTP handlers for the REST API.
package handlers

import (
	"net/http"

	"github.com/goplan/goplan/internal/api/types"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/user"
	"github.com/goplan/goplan/internal/repository"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	*BaseHandler
	userRepo      repository.UserRepository
	workspaceRepo repository.WorkspaceRepository
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userRepo repository.UserRepository, workspaceRepo repository.WorkspaceRepository) *UserHandler {
	return &UserHandler{
		BaseHandler:   NewBaseHandler(),
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
	}
}

// GetCurrentUser handles GET /api/v1/users/me
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.WriteMethodNotAllowed(w, []string{http.MethodGet})
		return
	}

	ctx := r.Context()
	userID := auth.GetUserID(ctx)

	if userID == "" {
		h.WriteUnauthorized(w, "authentication required")
		return
	}

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, u.ToResponse())
}

// UpdateCurrentUser handles PUT /api/v1/users/me
func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.WriteMethodNotAllowed(w, []string{http.MethodPut})
		return
	}

	ctx := r.Context()
	userID := auth.GetUserID(ctx)

	if userID == "" {
		h.WriteUnauthorized(w, "authentication required")
		return
	}

	var req types.UpdateUserRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to domain input
	input := &user.UpdateUserInput{
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
	}

	updatedUser, err := h.userRepo.Update(ctx, userID, input)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, updatedUser.ToResponse())
}

// GetUserWorkspaces handles GET /api/v1/users/me/workspaces
func (h *UserHandler) GetUserWorkspaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.WriteMethodNotAllowed(w, []string{http.MethodGet})
		return
	}

	ctx := r.Context()
	userID := auth.GetUserID(ctx)

	if userID == "" {
		h.WriteUnauthorized(w, "authentication required")
		return
	}

	workspaces, err := h.workspaceRepo.GetUserWorkspaces(ctx, userID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, workspaces)
}

// ServeHTTP routes requests to the appropriate handler method.
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch path {
	case "/api/v1/users/me":
		switch r.Method {
		case http.MethodGet:
			h.GetCurrentUser(w, r)
		case http.MethodPut:
			h.UpdateCurrentUser(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPut})
		}
	case "/api/v1/users/me/workspaces":
		h.GetUserWorkspaces(w, r)
	default:
		h.WriteNotFound(w, "endpoint")
	}
}
