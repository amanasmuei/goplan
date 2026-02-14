package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/goplan/backend/internal/config"
	"github.com/goplan/backend/internal/database"
	"github.com/goplan/backend/internal/handlers"
	"github.com/goplan/backend/internal/middleware"
	"github.com/goplan/backend/internal/repository"
	"github.com/goplan/backend/internal/services"
	"github.com/goplan/backend/internal/workers"

	_ "github.com/goplan/backend/docs"
)

// @title GoPlan API
// @version 1.0
// @description Planning-First Task Management API
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Startup validation warnings
	if cfg.Server.AllowOrigins == "*" {
		log.Println("[WARNING] ALLOW_ORIGINS is set to '*' (wildcard). This allows any origin and should not be used in production.")
	}
	if cfg.Server.Environment == "production" && strings.Contains(cfg.Server.AllowOrigins, "localhost") {
		log.Println("[WARNING] ALLOW_ORIGINS contains 'localhost' in production environment. This is likely a misconfiguration.")
	}
	if cfg.Server.Environment == "production" && cfg.Database.SSLMode == "disable" {
		log.Fatalf("DB_SSLMODE must not be 'disable' in production")
	}

	// Connect to database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(db.Pool)
	linkRepo := repository.NewTaskLinkRepository(db.Pool)
	justificationRepo := repository.NewJustificationRepository(db.Pool)
	blockerRepo := repository.NewBlockerRepository(db.Pool)
	reviewRepo := repository.NewReviewRepository(db.Pool)
	ackRepo := repository.NewAcknowledgmentRepository(db.Pool)
	teamRepo := repository.NewTeamRepository(db.Pool)
	projectRepo := repository.NewProjectRepository(db.Pool)

	// Initialize embedding client
	var embeddingClient *services.EmbeddingClient
	if cfg.Embedding.ServiceURL != "" {
		embeddingClient = services.NewEmbeddingClient(cfg.Embedding.ServiceURL)
		log.Printf("Embedding service configured: %s", cfg.Embedding.ServiceURL)
	}

	// Initialize services
	taskService := services.NewTaskService(
		taskRepo, linkRepo, justificationRepo, blockerRepo, reviewRepo, ackRepo, embeddingClient, db.Pool,
	)

	// Start embedding worker for background processing
	var embeddingWorker *workers.EmbeddingWorker
	if embeddingClient != nil {
		embeddingWorker = workers.NewEmbeddingWorker(db.Pool, embeddingClient)
		embeddingWorker.Start()
	}

	// Initialize user repository
	userRepo := repository.NewUserRepository(db.Pool)

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(taskService)
	linkHandler := handlers.NewLinkHandler(linkRepo, taskRepo)
	justificationHandler := handlers.NewJustificationHandler(justificationRepo, taskRepo)
	blockerHandler := handlers.NewBlockerHandler(blockerRepo, taskRepo)
	reviewHandler := handlers.NewReviewHandler(reviewRepo, taskRepo)
	teamHandler := handlers.NewTeamHandler(teamRepo, projectRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo, teamRepo)
	authHandler := handlers.NewAuthHandler(userRepo, &cfg.JWT)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(&cfg.JWT)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:                 "GoPlan API",
		ErrorHandler:            customErrorHandler,
		BodyLimit:               1 * 1024 * 1024, // 1MB
		EnableTrustedProxyCheck: true,
		TrustedProxies:          cfg.Server.TrustedProxies,
		ProxyHeader:             cfg.Server.ProxyHeader,
		ReadTimeout:             15 * time.Second,
		WriteTimeout:            30 * time.Second,
		IdleTimeout:             120 * time.Second,
	})

	// Global middleware
	app.Use(recover.New())

	// Security headers middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Cache-Control", "no-store")
		return c.Next()
	})

	// Request ID middleware for request tracing (before logger so IDs appear in logs)
	app.Use(requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	app.Use(logger.New())

	// CORS hardening: reject wildcard origins when credentials are enabled
	allowOrigins := cfg.Server.AllowOrigins
	if allowOrigins == "*" {
		log.Println("[WARNING] CORS: AllowOrigins is '*' but AllowCredentials is true. Wildcard origin with credentials is a security violation. Falling back to 'http://localhost:3000'.")
		allowOrigins = "http://localhost:3000"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Global rate limiter: 100 requests per minute
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many requests, please try again later",
			})
		},
	}))

	// Health check endpoint (no auth required)
	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Health(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
			})
		}
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	// Swagger documentation (no auth required, disabled in production)
	if cfg.Server.Environment != "production" {
		app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (no authentication required)
	auth := api.Group("/auth")
	// Stricter rate limiting for auth endpoints: 5 requests per minute
	authLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many requests, please try again later",
			})
		},
	})
	auth.Post("/login", authLimiter, authHandler.Login)
	auth.Post("/register", authLimiter, authHandler.Register)

	// Apply auth middleware to all other API routes
	api.Use(authMiddleware.Authenticate())

	// Auth routes that require authentication
	api.Get("/auth/me", authHandler.GetMe)

	// User routes
	api.Get("/users", authHandler.ListUsers)

	// Task routes
	tasks := api.Group("/tasks")
	tasks.Post("/", taskHandler.CreateTask)
	tasks.Get("/", taskHandler.ListTasks)
	tasks.Get("/:id", taskHandler.GetTask)
	tasks.Put("/:id", taskHandler.UpdateTask)
	tasks.Delete("/:id", taskHandler.DeleteTask)
	tasks.Get("/:id/similar", taskHandler.GetSimilarTasks)
	tasks.Post("/:id/acknowledge", taskHandler.AcknowledgeTask)
	tasks.Post("/:id/start", taskHandler.StartTask)
	tasks.Post("/:id/complete", taskHandler.CompleteTask)

	// Task link routes
	tasks.Post("/:id/links", linkHandler.CreateLink)
	tasks.Get("/:id/links", linkHandler.ListLinks)
	tasks.Delete("/:id/links/:linkId", linkHandler.DeleteLink)

	// Justification routes
	tasks.Post("/:id/justify", justificationHandler.CreateJustification)
	tasks.Get("/:id/justify", justificationHandler.GetJustification)

	// Blocker routes
	tasks.Post("/:id/blockers", blockerHandler.CreateBlocker)
	tasks.Get("/:id/blockers", blockerHandler.ListBlockers)

	// Review routes
	tasks.Post("/:id/review", reviewHandler.CreateReview)
	tasks.Get("/:id/review", reviewHandler.GetReview)

	// Blocker resolution (separate route)
	api.Put("/blockers/:id/resolve", blockerHandler.ResolveBlocker)

	// Team routes
	teams := api.Group("/teams")
	teams.Post("/", teamHandler.CreateTeam)
	teams.Get("/", teamHandler.ListTeams)
	teams.Get("/:id", teamHandler.GetTeam)
	teams.Put("/:id", teamHandler.UpdateTeam)
	teams.Delete("/:id", teamHandler.DeleteTeam)
	teams.Post("/:id/members", teamHandler.AddMember)
	teams.Get("/:id/members", teamHandler.ListMembers)
	teams.Put("/:id/members/:userId", teamHandler.UpdateMemberRole)
	teams.Delete("/:id/members/:userId", teamHandler.RemoveMember)
	teams.Get("/:id/projects", teamHandler.ListTeamProjects)

	// Project routes
	projects := api.Group("/projects")
	projects.Post("/", projectHandler.CreateProject)
	projects.Get("/", projectHandler.ListProjects)
	projects.Get("/:id", projectHandler.GetProject)
	projects.Put("/:id", projectHandler.UpdateProject)
	projects.Delete("/:id", projectHandler.DeleteProject)
	projects.Post("/:id/archive", projectHandler.ArchiveProject)
	projects.Post("/:id/teams", projectHandler.AssignTeams)
	projects.Delete("/:id/teams/:teamId", projectHandler.RemoveTeam)
	projects.Get("/:id/teams", projectHandler.GetProjectTeams)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.Server.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop embedding worker
	if embeddingWorker != nil {
		embeddingWorker.Stop()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Explicitly close database connection pool
	db.Close()

	log.Println("Server exited properly")
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "an internal error occurred"
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}
	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
