package main

import (
	"context"
	"log/slog"
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
	"github.com/goplan/backend/internal/claude"
	"github.com/goplan/backend/internal/config"
	"github.com/goplan/backend/internal/database"
	"github.com/goplan/backend/internal/handlers"
	"github.com/goplan/backend/internal/middleware"
	"github.com/goplan/backend/internal/repository"
	"github.com/goplan/backend/internal/services"

	_ "github.com/goplan/backend/docs"
)

// @title GoPlan API
// @version 2.0
// @description AI-Powered Strategic Planning API
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}

	// Startup validation warnings
	if cfg.Server.AllowOrigins == "*" {
		slog.Warn("ALLOW_ORIGINS is set to '*' (wildcard), should not be used in production")
	}
	if cfg.Server.Environment == "production" && strings.Contains(cfg.Server.AllowOrigins, "localhost") {
		slog.Warn("ALLOW_ORIGINS contains 'localhost' in production environment")
	}
	if cfg.Server.Environment == "production" && cfg.Database.SSLMode == "disable" {
		slog.Error("DB_SSLMODE must not be 'disable' in production")
		os.Exit(1)
	}

	// Connect to database
	db, err := database.New(&cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize Claude client
	var claudeClient *claude.Client
	if cfg.AI.Enabled && cfg.AI.ClaudeAPIKey != "" {
		claudeClient = claude.NewClient(
			cfg.AI.ClaudeAPIKey,
			cfg.AI.ClaudeModel,
			cfg.AI.MaxTokens,
			time.Duration(cfg.AI.TimeoutSec)*time.Second,
			cfg.AI.RateLimitRPM,
		)
		slog.Info("Claude AI client initialized", "model", cfg.AI.ClaudeModel)
	} else {
		slog.Warn("Claude AI client not configured (AI_ENABLED=false or CLAUDE_API_KEY empty)")
	}

	// Initialize embedding client
	var embeddingClient *services.EmbeddingClient
	if cfg.Embedding.ServiceURL != "" {
		embeddingClient = services.NewEmbeddingClient(cfg.Embedding.ServiceURL)
		slog.Info("Embedding service configured", "url", cfg.Embedding.ServiceURL)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Pool)
	planRepo := repository.NewPlanRepository(db.Pool)
	sectionRepo := repository.NewSectionRepository(db.Pool)
	versionRepo := repository.NewVersionRepository(db.Pool)
	subRepo := repository.NewSubscriptionRepository(db.Pool)
	genLogRepo := repository.NewGenerationLogRepository(db.Pool)

	// Initialize services
	strategyService := services.NewStrategyService(
		claudeClient, planRepo, sectionRepo, versionRepo, genLogRepo, subRepo, embeddingClient,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, &cfg.JWT)
	strategyHandler := handlers.NewStrategyHandler(strategyService, versionRepo, sectionRepo, planRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(&cfg.JWT)
	subscriptionMw := middleware.NewSubscriptionMiddleware(subRepo, planRepo)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:                 "GoPlan API",
		ErrorHandler:            customErrorHandler,
		BodyLimit:               1 * 1024 * 1024, // 1MB
		EnableTrustedProxyCheck: true,
		TrustedProxies:          cfg.Server.TrustedProxies,
		ProxyHeader:             cfg.Server.ProxyHeader,
		ReadTimeout:             15 * time.Second,
		WriteTimeout:            120 * time.Second, // Increased for AI generation
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
		slog.Warn("CORS: AllowOrigins is '*' with AllowCredentials, falling back to localhost")
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
			"version": "2.0.0",
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

	// Strategy routes
	strategies := api.Group("/strategies")
	strategies.Post("/", subscriptionMw.CheckPlanLimit(), strategyHandler.CreateStrategy)
	strategies.Get("/", strategyHandler.ListStrategies)
	strategies.Get("/:id", strategyHandler.GetStrategy)
	strategies.Delete("/:id", strategyHandler.ArchiveStrategy)
	strategies.Post("/:id/sections/:type/regenerate", subscriptionMw.RequireSubscription("regeneration"), strategyHandler.RegenerateSection)
	strategies.Post("/:id/sections/:type/refine", subscriptionMw.RequireSubscription("refine"), strategyHandler.RefineSection)
	strategies.Get("/:id/versions", subscriptionMw.RequireSubscription("version_history"), strategyHandler.ListVersions)
	strategies.Get("/:id/versions/:version", strategyHandler.GetVersion)
	strategies.Get("/:id/sections/:type/versions", strategyHandler.ListSectionVersions)
	strategies.Get("/:id/similar", strategyHandler.GetSimilarStrategies)
	strategies.Get("/:id/export", subscriptionMw.RequireSubscription("export"), strategyHandler.ExportStrategy)

	// Subscription routes
	api.Get("/subscription", subscriptionHandler.GetSubscription)
	api.Post("/subscription/upgrade", subscriptionHandler.UpgradeSubscription)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.Server.Port); err != nil {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	// Explicitly close database connection pool
	db.Close()

	slog.Info("Server exited properly")
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
