package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Embedding EmbeddingConfig
	AI        AIConfig
	Stripe    StripeConfig
}

type ServerConfig struct {
	Port           string
	Environment    string
	AllowOrigins   string
	TrustedProxies []string
	ProxyHeader    string
}

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	MaxConns    int32
	MinConns    int32
	MaxIdleTime string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type EmbeddingConfig struct {
	ServiceURL string
}

type AIConfig struct {
	ClaudeAPIKey string
	ClaudeModel  string
	MaxTokens    int
	Temperature  float64
	TimeoutSec   int
	RateLimitRPM int
	Enabled      bool
}

type StripeConfig struct {
	SecretKey      string
	WebhookSecret  string
	ProPriceID     string
	ProPlusPriceID string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           getEnv("SERVER_PORT", "8080"),
			Environment:    getEnv("ENVIRONMENT", "development"),
			AllowOrigins:   getEnv("ALLOW_ORIGINS", "http://localhost:3000"),
			TrustedProxies: getEnvAsSlice("TRUSTED_PROXIES", "127.0.0.1"),
			ProxyHeader:    getEnv("PROXY_HEADER", "X-Forwarded-For"),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnv("DB_PORT", "5432"),
			User:        getEnv("DB_USER", "goplan"),
			Password:    getEnv("DB_PASSWORD", "goplan"),
			DBName:      getEnv("DB_NAME", "goplan"),
			SSLMode:     getEnv("DB_SSLMODE", "disable"),
			MaxConns:    int32(getEnvInt("DB_MAX_CONNS", 20)),
			MinConns:    int32(getEnvInt("DB_MIN_CONNS", 5)),
			MaxIdleTime: getEnv("DB_MAX_IDLE_TIME", "30m"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		},
		Embedding: EmbeddingConfig{
			ServiceURL: getEnv("EMBEDDING_SERVICE_URL", "http://localhost:8000"),
		},
		AI: AIConfig{
			ClaudeAPIKey: getEnv("CLAUDE_API_KEY", ""),
			ClaudeModel:  getEnv("CLAUDE_MODEL", "claude-sonnet-4-20250514"),
			MaxTokens:    getEnvInt("AI_MAX_TOKENS", 8192),
			Temperature:  getEnvFloat("AI_TEMPERATURE", 0.7),
			TimeoutSec:   getEnvInt("AI_TIMEOUT_SEC", 120),
			RateLimitRPM: getEnvInt("AI_RATE_LIMIT_RPM", 30),
			Enabled:      getEnvBool("AI_ENABLED", true),
		},
		Stripe: StripeConfig{
			SecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
			ProPriceID:     getEnv("STRIPE_PRO_PRICE_ID", ""),
			ProPlusPriceID: getEnv("STRIPE_PRO_PLUS_PRICE_ID", ""),
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=" + c.SSLMode
}

// Validate checks required configuration values.
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.Server.Environment == "production" {
		if c.Database.Password == "" || c.Database.Password == "goplan" {
			return fmt.Errorf("DB_PASSWORD must be set to a secure value in production")
		}
		if c.Redis.Password == "" || c.Redis.Password == "changeme" {
			return fmt.Errorf("REDIS_PASSWORD must be set to a secure value in production")
		}
		if c.Database.SSLMode == "disable" {
			return fmt.Errorf("DB_SSLMODE must not be 'disable' in production")
		}
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsSlice(key, defaultValue string) []string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
