package api

import (
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/api/handlers"
	"github.com/goplan/goplan/internal/api/middleware"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/repository"
)

// jwtMiddleware is the JWT authentication middleware instance
var jwtMiddleware *auth.Middleware

// Router handles routing for the API.
type Router struct {
	authHandler      *handlers.AuthHandler
	userHandler      *handlers.UserHandler
	workspaceHandler *handlers.WorkspaceHandler
	planHandler      *handlers.PlanHandler
	taskHandler      *handlers.TaskHandler
	commentHandler   *handlers.CommentHandler
}

// Repositories contains all repository dependencies for the API.
type Repositories struct {
	User      repository.UserRepository
	Workspace repository.WorkspaceRepository
	Plan      repository.PlanRepository
	Task      repository.TaskRepository
	Comment   repository.CommentRepository
}

// NewRouter creates a new API router with all handlers.
func NewRouter(repos *Repositories, jwt *auth.JWT, authMw *auth.Middleware) *Router {
	// Store the JWT middleware for use in RegisterRoutes
	jwtMiddleware = authMw
	return &Router{
		authHandler:      handlers.NewAuthHandler(repos.User, jwt),
		userHandler:      handlers.NewUserHandler(repos.User, repos.Workspace),
		workspaceHandler: handlers.NewWorkspaceHandler(repos.Workspace, repos.User),
		planHandler:      handlers.NewPlanHandler(repos.Plan, repos.Workspace),
		taskHandler:      handlers.NewTaskHandler(repos.Task, repos.Plan, repos.Workspace),
		commentHandler:   handlers.NewCommentHandler(repos.Comment, repos.Task),
	}
}

// Handler returns the main HTTP handler for the API with middleware applied.
func (rt *Router) Handler() http.Handler {
	// Create the main API mux
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/api/v1/", rt.routeRequest)

	// Apply middleware chain
	handler := middleware.Chain(
		middleware.Recovery,
		middleware.Logging,
		middleware.RequestID,
		middleware.Auth,
	)(mux)

	return handler
}

// routeRequest routes requests to the appropriate handler based on the path.
func (rt *Router) routeRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Route based on path prefix
	switch {
	// Auth routes: /api/v1/auth/... (no auth required)
	case strings.HasPrefix(path, "/api/v1/auth"):
		rt.authHandler.ServeHTTP(w, r)

	// User routes: /api/v1/users/...
	case strings.HasPrefix(path, "/api/v1/users"):
		rt.userHandler.ServeHTTP(w, r)

	// Workspace routes: /api/v1/workspaces/...
	case strings.HasPrefix(path, "/api/v1/workspaces"):
		// Check if this is a plans subroute
		if strings.Contains(path, "/plans") {
			rt.planHandler.ServeHTTP(w, r)
			return
		}
		rt.workspaceHandler.ServeHTTP(w, r)

	// Plan routes: /api/v1/plans/...
	case strings.HasPrefix(path, "/api/v1/plans"):
		// Check if this is a tasks subroute
		if strings.Contains(path, "/tasks") {
			rt.taskHandler.ServeHTTP(w, r)
			return
		}
		rt.planHandler.ServeHTTP(w, r)

	// Task routes: /api/v1/tasks/...
	case strings.HasPrefix(path, "/api/v1/tasks"):
		// Check if this is a comments subroute
		if strings.Contains(path, "/comments") {
			rt.commentHandler.ServeHTTP(w, r)
			return
		}
		rt.taskHandler.ServeHTTP(w, r)

	// Comment routes: /api/v1/comments/...
	case strings.HasPrefix(path, "/api/v1/comments"):
		rt.commentHandler.ServeHTTP(w, r)

	default:
		// Return 404 for unknown routes
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":true,"code":"NOT_FOUND","message":"endpoint not found"}`))
	}
}

// routeProtectedRequest routes requests that require authentication.
func (rt *Router) routeProtectedRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Route based on path prefix (excluding auth routes)
	switch {
	// User routes: /api/v1/users/...
	case strings.HasPrefix(path, "/api/v1/users"):
		rt.userHandler.ServeHTTP(w, r)

	// Workspace routes: /api/v1/workspaces/...
	case strings.HasPrefix(path, "/api/v1/workspaces"):
		if strings.Contains(path, "/plans") {
			rt.planHandler.ServeHTTP(w, r)
			return
		}
		rt.workspaceHandler.ServeHTTP(w, r)

	// Plan routes: /api/v1/plans/...
	case strings.HasPrefix(path, "/api/v1/plans"):
		if strings.Contains(path, "/tasks") {
			rt.taskHandler.ServeHTTP(w, r)
			return
		}
		rt.planHandler.ServeHTTP(w, r)

	// Task routes: /api/v1/tasks/...
	case strings.HasPrefix(path, "/api/v1/tasks"):
		if strings.Contains(path, "/comments") {
			rt.commentHandler.ServeHTTP(w, r)
			return
		}
		rt.taskHandler.ServeHTTP(w, r)

	// Comment routes: /api/v1/comments/...
	case strings.HasPrefix(path, "/api/v1/comments"):
		rt.commentHandler.ServeHTTP(w, r)

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":true,"code":"NOT_FOUND","message":"endpoint not found"}`))
	}
}

// RegisterRoutes registers all API routes on the given ServeMux.
// Auth routes do not require authentication, all other routes do.
func (rt *Router) RegisterRoutes(mux *http.ServeMux) {
	// Auth routes - no auth middleware
	authRoutesHandler := middleware.Recovery(middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rt.authHandler.ServeHTTP(w, r)
	})))
	mux.Handle("/api/v1/auth/", authRoutesHandler)

	// Protected routes - with JWT auth middleware
	var protectedHandler http.Handler = http.HandlerFunc(rt.routeProtectedRequest)
	if jwtMiddleware != nil {
		protectedHandler = jwtMiddleware.Authenticate(protectedHandler)
	}
	protectedHandler = middleware.Recovery(middleware.Logging(protectedHandler))

	// Use Go 1.22+ pattern syntax with method prefixes to handle POST without trailing slash.
	// Registering "/api/v1/workspaces/" creates an automatic redirect from "/api/v1/workspaces"
	// which converts POST to GET, losing the request body.
	// Solution: Use method-specific patterns and wildcard sub-paths instead of trailing slash.

	// POST requests to exact base paths (for creating resources)
	mux.Handle("POST /api/v1/users", protectedHandler)
	mux.Handle("POST /api/v1/workspaces", protectedHandler)
	mux.Handle("POST /api/v1/plans", protectedHandler)
	mux.Handle("POST /api/v1/tasks", protectedHandler)
	mux.Handle("POST /api/v1/comments", protectedHandler)

	// GET requests to exact base paths (for listing resources)
	mux.Handle("GET /api/v1/users", protectedHandler)
	mux.Handle("GET /api/v1/workspaces", protectedHandler)
	mux.Handle("GET /api/v1/plans", protectedHandler)
	mux.Handle("GET /api/v1/tasks", protectedHandler)
	mux.Handle("GET /api/v1/comments", protectedHandler)

	// All methods for sub-paths (e.g., /api/v1/workspaces/{id}, /api/v1/workspaces/{id}/plans)
	mux.Handle("/api/v1/users/{path...}", protectedHandler)
	mux.Handle("/api/v1/workspaces/{path...}", protectedHandler)
	mux.Handle("/api/v1/plans/{path...}", protectedHandler)
	mux.Handle("/api/v1/tasks/{path...}", protectedHandler)
	mux.Handle("/api/v1/comments/{path...}", protectedHandler)
}
