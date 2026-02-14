// Package redis provides a Redis client and adapters for the GoPlan backend.
package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/goplan/goplan/internal/ratelimit"
)

// Client wraps the go-redis client and provides convenience methods.
type Client struct {
	rdb *goredis.Client
}

// Config holds Redis client configuration.
type Config struct {
	URL      string
	Password string
	DB       int
}

// New creates a new Redis client from the given configuration.
// If URL is empty, it defaults to localhost:6379.
func New(cfg Config) (*Client, error) {
	addr := cfg.URL
	if addr == "" {
		addr = "localhost:6379"
	}

	// Try to parse as a Redis URL first (redis://...)
	opts, err := goredis.ParseURL(addr)
	if err != nil {
		// Fall back to treating it as a host:port address
		opts = &goredis.Options{
			Addr: addr,
		}
	}

	// Override password and DB if explicitly provided
	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	if cfg.DB != 0 {
		opts.DB = cfg.DB
	}

	rdb := goredis.NewClient(opts)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping checks the Redis connection.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Underlying returns the underlying go-redis client for advanced usage.
func (c *Client) Underlying() *goredis.Client {
	return c.rdb
}

// RateLimitAdapter returns an adapter that implements the ratelimit.RedisClient interface.
func (c *Client) RateLimitAdapter() ratelimit.RedisClient {
	return &rateLimitAdapter{rdb: c.rdb}
}

// rateLimitAdapter adapts go-redis to the ratelimit.RedisClient interface.
type rateLimitAdapter struct {
	rdb *goredis.Client
}

// Incr increments the integer value of a key by one.
func (a *rateLimitAdapter) Incr(ctx context.Context, key string) (int64, error) {
	return a.rdb.Incr(ctx, key).Result()
}

// Expire sets a timeout on a key.
func (a *rateLimitAdapter) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return a.rdb.Expire(ctx, key, ttl).Err()
}

// Get returns the value of a key.
func (a *rateLimitAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.rdb.Get(ctx, key).Result()
}

// Set sets the value of a key with an expiration.
func (a *rateLimitAdapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return a.rdb.Set(ctx, key, value, ttl).Err()
}
