// Package main is the entry point for the GoPlan backend server.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/goplan/goplan/internal/api"
	"github.com/goplan/goplan/internal/api/handlers"
	"github.com/goplan/goplan/internal/api/middleware"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/audit"
	"github.com/goplan/goplan/internal/claude"
	"github.com/goplan/goplan/internal/config"
	apierrors "github.com/goplan/goplan/internal/errors"
	"github.com/goplan/goplan/internal/health"
	"github.com/goplan/goplan/internal/logging"
	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/mcp/tools"
	"github.com/goplan/goplan/internal/metrics"
	"github.com/goplan/goplan/internal/postgres"
	"github.com/goplan/goplan/internal/ratelimit"
	"github.com/goplan/goplan/internal/repository"
	"github.com/goplan/goplan/internal/validation"
)

// Version information (set at build time)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// Repositories holds all repository instances.
type Repositories struct {
	User        repository.UserRepository
	Workspace   repository.WorkspaceRepository
	Plan        repository.PlanRepository
	Phase       repository.PhaseRepository
	Milestone   repository.MilestoneRepository
	Task        repository.TaskRepository
	Comment     repository.CommentRepository
	ActivityLog repository.ActivityLogRepository
	TxManager   repository.TxManager
}

// initRepositories creates all repository instances.
func initRepositories(db *postgres.DB) *Repositories {
	pool := db.Pool()
	return &Repositories{
		User:        postgres.NewUserRepository(pool),
		Workspace:   postgres.NewWorkspaceRepository(pool),
		Plan:        postgres.NewPlanRepository(pool),
		Phase:       postgres.NewPhaseRepository(pool),
		Milestone:   postgres.NewMilestoneRepository(pool),
		Task:        postgres.NewTaskRepository(pool),
		Comment:     postgres.NewCommentRepository(pool),
		ActivityLog: postgres.NewActivityLogRepository(pool),
		TxManager:   db.TxManager(),
	}
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	logger := logging.New(logging.Config{
		Level:  cfg.Server.LogLevel,
		Format: cfg.Server.LogFormat,
	})
	logging.SetDefault(logger)

	// Skip validation in development mode for easier local setup
	if !cfg.IsDevelopment() {
		if err := cfg.Validate(); err != nil {
			logger.Error("Configuration validation failed", "error", err)
			os.Exit(1)
		}
	}

	// Initialize metrics
	appMetrics := metrics.New()
	metrics.SetDefault(appMetrics)

	// Initialize validator
	validator := validation.New(validation.DefaultConfig())
	validation.SetDefault(validator)

	// Initialize context for startup
	ctx := context.Background()

	// Initialize database
	db, err := postgres.NewDB(ctx, cfg.Database)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Connected to database")

	// Create repositories
	repos := initRepositories(db)

	// Initialize JWT auth
	jwtAuth := auth.NewJWT(auth.JWTConfig{
		Secret:          cfg.Auth.JWTSecret,
		AccessTokenTTL:  cfg.Auth.JWTAccessExpiry,
		RefreshTokenTTL: cfg.Auth.JWTRefreshExpiry,
		Issuer:          "goplan",
	})

	// Initialize auth middleware
	authMiddleware := auth.NewMiddleware(auth.MiddlewareConfig{
		JWT:     jwtAuth,
		Logger:  logger,
		Metrics: appMetrics,
		SkipPaths: []string{
			"/health",
			"/ready",
			"/metrics",
			"/version",
		},
		DevMode:   cfg.IsDevelopment(),
		DevUserID: "",
	})

	// Initialize rate limiter
	rateLimiter := ratelimit.New(ratelimit.Config{
		DefaultRequestsPerSecond: cfg.Security.RateLimitRequests,
		DefaultBurstSize:         cfg.Security.RateLimitRequests * 2,
		IPRequestsPerSecond:      50,
		IPBurstSize:              100,
		UserRequestsPerSecond:    100,
		UserBurstSize:            200,
		CleanupInterval:          5 * time.Minute,
	}, logger, appMetrics)
	defer rateLimiter.Stop()

	// Initialize audit logger
	auditLogger := audit.New(audit.Config{
		Enabled: true,
		Logger:  logger,
	})
	audit.SetDefault(auditLogger)

	// Initialize health handler
	healthHandler := health.NewHandler(health.Config{
		Version: Version,
		Timeout: 5 * time.Second,
	})

	// Add health checkers
	healthHandler.AddChecker(health.DatabaseChecker("database", func(ctx context.Context) error {
		return db.Ping(ctx)
	}))
	healthHandler.AddChecker(health.MemoryChecker(1024 * 1024 * 1024)) // 1GB
	healthHandler.AddChecker(health.GoroutineChecker(10000))

	// Create API router with repositories, JWT auth, and auth middleware
	apiRouter := api.NewRouter(&api.Repositories{
		User:      repos.User,
		Workspace: repos.Workspace,
		Plan:      repos.Plan,
		Task:      repos.Task,
		Comment:   repos.Comment,
	}, jwtAuth, authMiddleware)

	// Create MCP tool registry
	registry := mcp.NewToolRegistry()

	// Register all MCP tools
	mcpRepos := tools.Repositories{
		Workspace:   repos.Workspace,
		Plan:        repos.Plan,
		Phase:       repos.Phase,
		Milestone:   repos.Milestone,
		Task:        repos.Task,
		Comment:     repos.Comment,
		ActivityLog: repos.ActivityLog,
	}
	if err := tools.RegisterAllTools(registry, mcpRepos); err != nil {
		logger.Error("Failed to register MCP tools", "error", err)
		os.Exit(1)
	}
	logger.Info("Registered MCP tools", "count", len(registry.List()))

	// Create MCP server
	mcpServer := mcp.NewServer(registry)

	// Initialize AI service if enabled
	var aiHandler *handlers.AIHandler
	if cfg.AI.Enabled {
		aiHandler = initAIService(ctx, cfg, registry, repos, logger)
		if aiHandler != nil {
			logger.Info("AI service initialized")
		}
	}

	// Create main router
	mux := http.NewServeMux()

	// Health and monitoring endpoints (no auth required)
	mux.HandleFunc("/health", healthHandler.LivenessHandler())
	mux.HandleFunc("/ready", healthHandler.ReadinessHandler())
	mux.HandleFunc("/healthz", healthHandler.HealthHandler()) // Detailed health
	mux.Handle("/metrics", appMetrics.Handler())
	mux.HandleFunc("/version", versionHandler())

	// Register API v1 routes
	apiRouter.RegisterRoutes(mux)

	// Mount MCP routes
	mux.Handle("/mcp/", mcpServer.Routes())

	// Mount AI routes if enabled
	if aiHandler != nil {
		mux.Handle("/api/v1/ai/", handlers.AIRoutes(aiHandler))
		logger.Info("AI routes registered at /api/v1/ai/")
	}

	// Build middleware chain
	securityHeaders := middleware.NewSecurityHeaders(middleware.APISecurityConfig())

	// Apply middleware chain (order matters - first middleware wraps innermost)
	handler := buildMiddlewareChain(
		mux,
		cfg,
		logger,
		appMetrics,
		authMiddleware,
		rateLimiter,
		securityHeaders,
	)

	// Create HTTP server with proper timeouts
	server := &http.Server{
		Addr:              cfg.GetListenAddr(),
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	// Start server in a goroutine
	go func() {
		logger.Startup(Version, cfg.Server.Env, cfg.GetListenAddr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-quit

	logger.Info("Received shutdown signal", "signal", sig.String())

	// Create shutdown context with timeout
	shutdownTimeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Graceful shutdown
	logger.Info("Starting graceful shutdown", "timeout", shutdownTimeout.String())

	// Stop accepting new requests
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	// Clean up resources
	logger.Info("Closing database connections")
	db.Close()

	logger.Shutdown("graceful shutdown complete")
}

// buildMiddlewareChain creates the middleware chain.
func buildMiddlewareChain(
	handler http.Handler,
	cfg *config.Config,
	logger *logging.Logger,
	appMetrics *metrics.Metrics,
	authMiddleware *auth.Middleware,
	rateLimiter *ratelimit.Limiter,
	securityHeaders *middleware.SecurityHeaders,
) http.Handler {
	// Apply middleware in reverse order (last applied = first executed)

	// 1. Recovery middleware (innermost - catches panics)
	handler = recoveryMiddleware(logger, cfg.IsDevelopment())(handler)

	// 2. Security headers
	handler = securityHeaders.Handler(handler)

	// 3. Authentication (optional for some routes)
	handler = authMiddleware.Optional(handler)

	// 4. Rate limiting
	if cfg.Security.RateLimitEnabled {
		handler = rateLimiter.Middleware(handler)
	}

	// 5. Request logging and metrics
	handler = loggingMiddleware(logger, appMetrics)(handler)

	// 6. Request ID injection
	handler = requestIDMiddleware(handler)

	// 7. CORS (outermost)
	handler = corsMiddleware(cfg)(handler)

	return handler
}

// requestIDMiddleware adds a request ID to the context.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID is provided in header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID based on timestamp
			requestID = time.Now().Format("20060102150405.000000")
		}

		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := logging.ContextWithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggingMiddleware logs incoming requests with structured logging.
func loggingMiddleware(logger *logging.Logger, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Increment in-flight requests
			m.IncHTTPRequestsInFlight()
			defer m.DecHTTPRequestsInFlight()

			// Wrap the response writer
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process the request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Record metrics
			m.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration, wrapped.bytesWritten)

			// Log the request
			logger.HTTPRequest(r.Context(), r.Method, r.URL.Path, wrapped.statusCode, duration, wrapped.bytesWritten)
		})
	}
}

// recoveryMiddleware recovers from panics.
func recoveryMiddleware(logger *logging.Logger, showDetails bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					logger.Panic(r.Context(), err, stack)

					apiErr := apierrors.InternalError("")
					if showDetails {
						apiErr.Message = fmt.Sprintf("panic: %v", err)
					}
					apiErr.WriteResponse(w)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware adds CORS headers.
func corsMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			origin := r.Header.Get("Origin")
			if origin != "" {
				allowed := false
				for _, o := range cfg.Security.CORSOrigins {
					if o == "*" || o == origin {
						allowed = true
						break
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-User-ID, X-Workspace-ID, X-User-Role, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// versionHandler returns the version handler.
func versionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"version":   Version,
			"buildTime": BuildTime,
		}

		writeJSON(w, http.StatusOK, response)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	written      bool
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write writes the response body and tracks bytes written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logging.Default().Error("Error encoding JSON response", "error", err)
	}
}

// initAIService initializes the AI service if configured.
func initAIService(ctx context.Context, cfg *config.Config, registry *mcp.ToolRegistry, repos *Repositories, logger *logging.Logger) *handlers.AIHandler {
	// Check if API key is configured
	if cfg.AI.ClaudeAPIKey == "" {
		logger.Info("AI service disabled: CLAUDE_API_KEY not configured")
		return nil
	}

	// Create Claude client
	claudeClient := claude.NewClient(cfg.AI,
		claude.WithTemperature(cfg.AI.Temperature),
		claude.WithRateLimit(cfg.AI.RateLimitRPM, time.Minute),
	)

	// Create safety checker
	safetyChecker := claude.NewSafetyChecker()

	// Create AI service
	aiService := claude.NewService(claude.ServiceConfig{
		Client:        claudeClient,
		Registry:      registry,
		SafetyChecker: safetyChecker,
		WorkspaceRepo: repos.Workspace,
		PlanRepo:      repos.Plan,
		TaskRepo:      repos.Task,
	})

	// Start background workers
	aiService.StartBackgroundWorkers(ctx)

	// Create and return the handler
	return handlers.NewAIHandler(aiService)
}

// init prints startup banner.
func init() {
	fmt.Print(`
  ____       ____  _
 / ___| ___ |  _ \| | __ _ _ __
| |  _ / _ \| |_) | |/ _' | '_ \
| |_| | (_) |  __/| | (_| | | | |
 \____|\___/|_|   |_|\__,_|_| |_|

  Project Management Made Simple
`)
}
