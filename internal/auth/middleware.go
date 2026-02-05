// Package auth provides authentication middleware for the GoPlan backend.
package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/logging"
	"github.com/goplan/goplan/internal/metrics"
)

// Context keys for auth information.
type contextKey string

const (
	// ContextKeyUser is the context key for the authenticated user.
	ContextKeyUser contextKey = "user"
	// ContextKeyClaims is the context key for JWT claims.
	ContextKeyClaims contextKey = "claims"
)

// User represents an authenticated user.
type User struct {
	ID          string
	Email       string
	Name        string
	WorkspaceID string
	Role        string
}

// Middleware provides JWT authentication middleware.
type Middleware struct {
	jwt          *JWT
	logger       *logging.Logger
	metrics      *metrics.Metrics
	skipPaths    []string
	devMode      bool
	devUserID    string
}

// MiddlewareConfig holds middleware configuration.
type MiddlewareConfig struct {
	JWT       *JWT
	Logger    *logging.Logger
	Metrics   *metrics.Metrics
	SkipPaths []string
	DevMode   bool
	DevUserID string
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware(config MiddlewareConfig) *Middleware {
	return &Middleware{
		jwt:       config.JWT,
		logger:    config.Logger,
		metrics:   config.Metrics,
		skipPaths: config.SkipPaths,
		devMode:   config.DevMode,
		devUserID: config.DevUserID,
	}
}

// Authenticate is the main authentication middleware.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should be skipped
		if m.shouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// In dev mode, check for dev headers first
		if m.devMode {
			if user := m.extractDevUser(r); user != nil {
				ctx := context.WithValue(r.Context(), ContextKeyUser, user)
				ctx = logging.ContextWithUserID(ctx, user.ID)
				if user.WorkspaceID != "" {
					ctx = logging.ContextWithWorkspaceID(ctx, user.WorkspaceID)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Extract JWT token
		authHeader := r.Header.Get("Authorization")
		token, err := ExtractBearerToken(authHeader)
		if err != nil {
			m.recordAuthFailure()
			m.writeAuthError(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.jwt.ValidateAccessToken(token)
		if err != nil {
			m.recordAuthFailure()

			status := http.StatusUnauthorized
			message := "invalid token"

			switch err {
			case ErrTokenExpired:
				message = "token expired"
			case ErrInvalidSignature:
				message = "invalid token signature"
			case ErrMissingClaims:
				message = "invalid token claims"
			}

			m.writeAuthError(w, message, status)
			return
		}

		// Create user from claims
		user := &User{
			ID:          claims.UserID,
			Email:       claims.Email,
			Name:        claims.Name,
			WorkspaceID: claims.WorkspaceID,
			Role:        claims.Role,
		}

		// Add user and claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyUser, user)
		ctx = context.WithValue(ctx, ContextKeyClaims, claims)
		ctx = logging.ContextWithUserID(ctx, user.ID)
		if user.WorkspaceID != "" {
			ctx = logging.ContextWithWorkspaceID(ctx, user.WorkspaceID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns middleware that requires a specific role or higher.
func (m *Middleware) RequireRole(minRole string) func(http.Handler) http.Handler {
	roleHierarchy := map[string]int{
		"owner":  4,
		"admin":  3,
		"member": 2,
		"viewer": 1,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r.Context())
			if user == nil {
				m.writeAuthError(w, "authentication required", http.StatusUnauthorized)
				return
			}

			currentLevel, ok := roleHierarchy[user.Role]
			if !ok {
				m.writeAuthError(w, "invalid role", http.StatusForbidden)
				return
			}

			requiredLevel, ok := roleHierarchy[minRole]
			if !ok {
				m.writeAuthError(w, "invalid required role configuration", http.StatusInternalServerError)
				return
			}

			if currentLevel < requiredLevel {
				m.writeAuthError(w, "insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin returns middleware that requires admin role or higher.
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole("admin")(next)
}

// RequireMember returns middleware that requires member role or higher.
func (m *Middleware) RequireMember(next http.Handler) http.Handler {
	return m.RequireRole("member")(next)
}

// Optional allows unauthenticated requests but extracts user if token is present.
func (m *Middleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In dev mode, check for dev headers
		if m.devMode {
			if user := m.extractDevUser(r); user != nil {
				ctx := context.WithValue(r.Context(), ContextKeyUser, user)
				ctx = logging.ContextWithUserID(ctx, user.ID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Try to extract and validate token
		authHeader := r.Header.Get("Authorization")
		if token, err := ExtractBearerToken(authHeader); err == nil {
			if claims, err := m.jwt.ValidateAccessToken(token); err == nil {
				user := &User{
					ID:          claims.UserID,
					Email:       claims.Email,
					Name:        claims.Name,
					WorkspaceID: claims.WorkspaceID,
					Role:        claims.Role,
				}
				ctx := context.WithValue(r.Context(), ContextKeyUser, user)
				ctx = context.WithValue(ctx, ContextKeyClaims, claims)
				ctx = logging.ContextWithUserID(ctx, user.ID)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// shouldSkip checks if a path should skip authentication.
func (m *Middleware) shouldSkip(path string) bool {
	for _, skip := range m.skipPaths {
		if skip == path || strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// extractDevUser extracts user from development headers.
func (m *Middleware) extractDevUser(r *http.Request) *User {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		if m.devUserID != "" {
			userID = m.devUserID
		} else {
			return nil
		}
	}

	return &User{
		ID:          userID,
		Email:       r.Header.Get("X-User-Email"),
		Name:        r.Header.Get("X-User-Name"),
		WorkspaceID: r.Header.Get("X-Workspace-ID"),
		Role:        r.Header.Get("X-User-Role"),
	}
}

// writeAuthError writes an authentication error response.
func (m *Middleware) writeAuthError(w http.ResponseWriter, message string, status int) {
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

// recordAuthFailure records an authentication failure metric.
func (m *Middleware) recordAuthFailure() {
	if m.metrics != nil {
		m.metrics.RecordAuthAttempt(false)
	}
}

// GetUser extracts the user from the context.
func GetUser(ctx context.Context) *User {
	user, _ := ctx.Value(ContextKeyUser).(*User)
	return user
}

// GetClaims extracts the JWT claims from the context.
func GetClaims(ctx context.Context) *Claims {
	claims, _ := ctx.Value(ContextKeyClaims).(*Claims)
	return claims
}

// GetUserID extracts the user ID from the context.
func GetUserID(ctx context.Context) string {
	if user := GetUser(ctx); user != nil {
		return user.ID
	}
	return ""
}

// GetWorkspaceID extracts the workspace ID from the context.
func GetWorkspaceID(ctx context.Context) string {
	if user := GetUser(ctx); user != nil {
		return user.WorkspaceID
	}
	return ""
}

// GetUserRole extracts the user role from the context.
func GetUserRole(ctx context.Context) string {
	if user := GetUser(ctx); user != nil {
		return user.Role
	}
	return ""
}

// IsAuthenticated checks if the request is authenticated.
func IsAuthenticated(ctx context.Context) bool {
	return GetUser(ctx) != nil
}

// HasRole checks if the user has at least the specified role.
func HasRole(ctx context.Context, minRole string) bool {
	user := GetUser(ctx)
	if user == nil {
		return false
	}

	roleHierarchy := map[string]int{
		"owner":  4,
		"admin":  3,
		"member": 2,
		"viewer": 1,
	}

	currentLevel, ok := roleHierarchy[user.Role]
	if !ok {
		return false
	}

	requiredLevel, ok := roleHierarchy[minRole]
	if !ok {
		return false
	}

	return currentLevel >= requiredLevel
}
