// Package config provides configuration loading for the GoPlan backend.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Auth        AuthConfig
	OAuth       OAuthConfig
	AI          AIConfig
	Email       EmailConfig
	Storage     StorageConfig
	Security    SecurityConfig
	Observability ObservabilityConfig
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port      string
	Host      string
	Env       string
	LogLevel  string
	LogFormat string
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	URL         string
	MaxConns    int
	MinConns    int
	MaxIdleTime time.Duration
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string
	BcryptCost       int
}

// OAuthConfig holds OAuth provider configuration.
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
}

// AIConfig holds AI/LLM configuration.
type AIConfig struct {
	ClaudeAPIKey     string
	ClaudeModel      string
	MaxTokens        int
	Temperature      float64
	Timeout          time.Duration
	RateLimitRPM     int  // Requests per minute
	Enabled          bool
	RequireApproval  bool // Whether AI actions require approval by default
}

// EmailConfig holds email configuration.
type EmailConfig struct {
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPassword   string
	SMTPFrom       string
	SendGridAPIKey string
}

// StorageConfig holds storage configuration.
type StorageConfig struct {
	Type       string
	Path       string
	S3Bucket   string
	S3Region   string
	S3AccessKey string
	S3SecretKey string
	S3Endpoint  string
}

// SecurityConfig holds security configuration.
type SecurityConfig struct {
	CORSOrigins       []string
	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   time.Duration
}

// ObservabilityConfig holds observability configuration.
type ObservabilityConfig struct {
	OTelEnabled     bool
	OTelEndpoint    string
	OTelServiceName string
	SentryDSN       string
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	cfg := &Config{}

	// Server configuration
	cfg.Server = ServerConfig{
		Port:      getEnvOrDefault("PORT", "8080"),
		Host:      getEnvOrDefault("HOST", "0.0.0.0"),
		Env:       getEnvOrDefault("ENV", "development"),
		LogLevel:  getEnvOrDefault("LOG_LEVEL", "info"),
		LogFormat: getEnvOrDefault("LOG_FORMAT", "json"),
	}

	// Validate required environment
	if cfg.Server.Env != "development" && cfg.Server.Env != "staging" && cfg.Server.Env != "production" {
		return nil, fmt.Errorf("invalid ENV value: %s", cfg.Server.Env)
	}

	// Database configuration
	cfg.Database = DatabaseConfig{
		URL:         os.Getenv("DATABASE_URL"),
		MaxConns:    getEnvAsIntOrDefault("DATABASE_MAX_CONNS", 25),
		MinConns:    getEnvAsIntOrDefault("DATABASE_MIN_CONNS", 5),
		MaxIdleTime: getEnvAsDurationOrDefault("DATABASE_MAX_IDLE_TIME", 5*time.Minute),
	}

	// Redis configuration
	cfg.Redis = RedisConfig{
		URL:      os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       getEnvAsIntOrDefault("REDIS_DB", 0),
	}

	// Auth configuration
	cfg.Auth = AuthConfig{
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTAccessExpiry:  getEnvAsDurationOrDefault("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getEnvAsDurationOrDefault("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		JWTIssuer:        getEnvOrDefault("JWT_ISSUER", "goplan"),
		BcryptCost:       getEnvAsIntOrDefault("BCRYPT_COST", 12),
	}

	// OAuth configuration
	cfg.OAuth = OAuthConfig{
		GoogleClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		GitHubClientID:     os.Getenv("OAUTH_GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("OAUTH_GITHUB_CLIENT_SECRET"),
	}

	// AI configuration
	cfg.AI = AIConfig{
		ClaudeAPIKey:    os.Getenv("CLAUDE_API_KEY"),
		ClaudeModel:     getEnvOrDefault("CLAUDE_MODEL", "claude-sonnet-4-20250514"),
		MaxTokens:       getEnvAsIntOrDefault("AI_MAX_TOKENS", 4096),
		Temperature:     getEnvAsFloatOrDefault("AI_TEMPERATURE", 0.7),
		Timeout:         getEnvAsDurationOrDefault("AI_TIMEOUT", 30*time.Second),
		RateLimitRPM:    getEnvAsIntOrDefault("AI_RATE_LIMIT_RPM", 60),
		Enabled:         getEnvAsBoolOrDefault("AI_ENABLED", true),
		RequireApproval: getEnvAsBoolOrDefault("AI_REQUIRE_APPROVAL", true),
	}

	// Email configuration
	cfg.Email = EmailConfig{
		SMTPHost:       os.Getenv("SMTP_HOST"),
		SMTPPort:       getEnvAsIntOrDefault("SMTP_PORT", 587),
		SMTPUser:       os.Getenv("SMTP_USER"),
		SMTPPassword:   os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:       getEnvOrDefault("SMTP_FROM", "noreply@goplan.io"),
		SendGridAPIKey: os.Getenv("SENDGRID_API_KEY"),
	}

	// Storage configuration
	cfg.Storage = StorageConfig{
		Type:        getEnvOrDefault("STORAGE_TYPE", "local"),
		Path:        getEnvOrDefault("STORAGE_PATH", "./uploads"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3Region:    os.Getenv("S3_REGION"),
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
		S3Endpoint:  os.Getenv("S3_ENDPOINT"),
	}

	// Security configuration
	cfg.Security = SecurityConfig{
		CORSOrigins:       getEnvAsSliceOrDefault("CORS_ORIGINS", []string{"*"}),
		RateLimitEnabled:  getEnvAsBoolOrDefault("RATE_LIMIT_ENABLED", true),
		RateLimitRequests: getEnvAsIntOrDefault("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvAsDurationOrDefault("RATE_LIMIT_WINDOW", 1*time.Minute),
	}

	// Observability configuration
	cfg.Observability = ObservabilityConfig{
		OTelEnabled:     getEnvAsBoolOrDefault("OTEL_ENABLED", false),
		OTelEndpoint:    os.Getenv("OTEL_ENDPOINT"),
		OTelServiceName: getEnvOrDefault("OTEL_SERVICE_NAME", "goplan-api"),
		SentryDSN:       os.Getenv("SENTRY_DSN"),
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Database URL is required
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	// Redis URL is required
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}

	// JWT Secret is required
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	// JWT Secret must be at least 32 characters
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	// Claude API Key is required if AI is enabled
	if c.AI.Enabled && c.AI.ClaudeAPIKey == "" {
		return fmt.Errorf("CLAUDE_API_KEY is required when AI_ENABLED is true")
	}

	return nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// GetListenAddr returns the full listen address.
func (c *Config) GetListenAddr() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
