package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Embedding EmbeddingConfig
}

type ServerConfig struct {
	Port           string
	Environment    string
	AllowOrigins   string
	TrustedProxies []string
	ProxyHeader    string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
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
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "goplan"),
			Password: getEnv("DB_PASSWORD", "goplan"),
			DBName:   getEnv("DB_NAME", "goplan"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
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
