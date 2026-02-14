package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const cacheKeyPrefix = "cache:"

// Cache provides a simple key-value cache backed by Redis.
// Values are serialized as JSON. Each entry has a TTL after which
// Redis automatically evicts it.
type Cache struct {
	rdb           *goredis.Client
	defaultTTL    time.Duration
}

// CacheConfig holds configuration for the cache.
type CacheConfig struct {
	// DefaultTTL is the default time-to-live for cache entries.
	// If zero, defaults to 5 minutes.
	DefaultTTL time.Duration
}

// NewCache creates a new Redis-backed cache.
func NewCache(client *Client, cfg CacheConfig) *Cache {
	ttl := cfg.DefaultTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &Cache{
		rdb:        client.rdb,
		defaultTTL: ttl,
	}
}

// Get retrieves a cached value by key and unmarshals it into dest.
// Returns false if the key does not exist or an error occurs.
func (c *Cache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	fullKey := cacheKeyPrefix + key
	val, err := c.rdb.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == goredis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("redis cache get: %w", err)
	}

	if err := json.Unmarshal(val, dest); err != nil {
		return false, fmt.Errorf("redis cache unmarshal: %w", err)
	}
	return true, nil
}

// Set stores a value in the cache with the default TTL.
func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with a custom TTL.
func (c *Cache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("redis cache marshal: %w", err)
	}

	fullKey := cacheKeyPrefix + key
	if err := c.rdb.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis cache set: %w", err)
	}
	return nil
}

// Delete removes a value from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
	fullKey := cacheKeyPrefix + key
	if err := c.rdb.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("redis cache delete: %w", err)
	}
	return nil
}

// DeletePattern removes all keys matching a glob pattern.
// For example, DeletePattern(ctx, "users:*") removes all user cache entries.
func (c *Cache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := cacheKeyPrefix + pattern
	iter := c.rdb.Scan(ctx, 0, fullPattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("redis cache delete pattern: %w", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis cache scan: %w", err)
	}
	return nil
}
