// Package middleware provides HTTP middleware for the REST API.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/logging"
)

// Context keys for storing auth information.
type contextKey string

const (
	// ContextKeyUserID is the context key for the user ID.
	ContextKeyUserID contextKey = "userID"
	// ContextKeyWorkspaceID is the context key for the workspace ID.
	ContextKeyWorkspaceID contextKey = "workspaceID"
	// ContextKeyUserRole is the context key for the user role.
	ContextKeyUserRole contextKey = "userRole"
	// ContextKeyRequestID is the context key for the request ID.
	ContextKeyRequestID contextKey = "requestID"
)

// Header names for development auth.
const (
	HeaderUserID      = "X-User-ID"
	HeaderWorkspaceID = "X-Workspace-ID"
	HeaderUserRole    = "X-User-Role"
)

// AuthContext contains authentication information extracted from the request.
type AuthContext struct {
	UserID      string
	WorkspaceID string
	Role        string
}

// GetAuthContext extracts the auth context from the request context.
func GetAuthContext(ctx context.Context) *AuthContext {
	userID, _ := ctx.Value(ContextKeyUserID).(string)
	workspaceID, _ := ctx.Value(ContextKeyWorkspaceID).(string)
	role, _ := ctx.Value(ContextKeyUserRole).(string)

	return &AuthContext{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        role,
	}
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(ContextKeyUserID).(string)
	return userID
}

// GetWorkspaceID extracts the workspace ID from the request context.
func GetWorkspaceID(ctx context.Context) string {
	workspaceID, _ := ctx.Value(ContextKeyWorkspaceID).(string)
	return workspaceID
}

// GetUserRole extracts the user role from the request context.
func GetUserRole(ctx context.Context) string {
	role, _ := ctx.Value(ContextKeyUserRole).(string)
	return role
}

// Auth is the authentication middleware that extracts auth information from headers.
// For development, it uses X-User-ID, X-Workspace-ID, and X-User-Role headers.
// In production, this would be replaced with proper JWT validation.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from header (required)
		userID := r.Header.Get(HeaderUserID)
		if userID == "" {
			writeAuthError(w, "authentication required", http.StatusUnauthorized)
			return
		}

		// Extract workspace ID from header (optional - may be in URL)
		workspaceID := r.Header.Get(HeaderWorkspaceID)

		// Extract role from header (optional - defaults to viewer)
		role := r.Header.Get(HeaderUserRole)
		if role == "" {
			role = shared.RoleViewer
		}

		// Validate role if provided
		if !shared.IsValidMemberRole(role) {
			writeAuthError(w, "invalid role", http.StatusBadRequest)
			return
		}

		// Add auth info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyUserID, userID)
		ctx = context.WithValue(ctx, ContextKeyWorkspaceID, workspaceID)
		ctx = context.WithValue(ctx, ContextKeyUserRole, role)

		// Also add to logging context for request correlation
		ctx = logging.ContextWithUserID(ctx, userID)
		if workspaceID != "" {
			ctx = logging.ContextWithWorkspaceID(ctx, workspaceID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth is middleware that requires authentication.
func RequireAuth(next http.Handler) http.Handler {
	return Auth(next)
}

// RequireRole returns middleware that requires a specific role or higher.
func RequireRole(minRole string) func(http.Handler) http.Handler {
	roleHierarchy := map[string]int{
		shared.RoleOwner:  4,
		shared.RoleAdmin:  3,
		shared.RoleMember: 2,
		shared.RoleViewer: 1,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRole(r.Context())

			currentLevel, ok := roleHierarchy[role]
			if !ok {
				writeAuthError(w, "invalid role", http.StatusForbidden)
				return
			}

			requiredLevel, ok := roleHierarchy[minRole]
			if !ok {
				writeAuthError(w, "invalid required role configuration", http.StatusInternalServerError)
				return
			}

			if currentLevel < requiredLevel {
				writeAuthError(w, "insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin returns middleware that requires admin role or higher.
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(shared.RoleAdmin)(next)
}

// RequireMember returns middleware that requires member role or higher.
func RequireMember(next http.Handler) http.Handler {
	return RequireRole(shared.RoleMember)(next)
}

// CanView returns true if the user has at least viewer permissions.
func CanView(role string) bool {
	return role == shared.RoleOwner || role == shared.RoleAdmin ||
		role == shared.RoleMember || role == shared.RoleViewer
}

// CanEdit returns true if the user has at least member permissions.
func CanEdit(role string) bool {
	return role == shared.RoleOwner || role == shared.RoleAdmin || role == shared.RoleMember
}

// CanManage returns true if the user has at least admin permissions.
func CanManage(role string) bool {
	return role == shared.RoleOwner || role == shared.RoleAdmin
}

// IsOwner returns true if the user is the owner.
func IsOwner(role string) bool {
	return role == shared.RoleOwner
}

// writeAuthError writes an authentication/authorization error response.
func writeAuthError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	code := "UNAUTHORIZED"
	if status == http.StatusForbidden {
		code = "FORBIDDEN"
	}

	response := map[string]interface{}{
		"error":   true,
		"code":    code,
		"message": message,
	}

	_ = json.NewEncoder(w).Encode(response)
}
