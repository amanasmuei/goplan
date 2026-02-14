package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const blacklistKeyPrefix = "token:blacklist:"

// TokenBlacklist provides JWT token blacklisting backed by Redis.
// Blacklisted tokens are stored with a TTL matching the token's remaining
// expiry time, so entries are automatically cleaned up.
type TokenBlacklist struct {
	rdb *goredis.Client
}

// NewTokenBlacklist creates a new Redis-backed token blacklist.
func NewTokenBlacklist(client *Client) *TokenBlacklist {
	return &TokenBlacklist{rdb: client.rdb}
}

// Add adds a token ID to the blacklist. The entry expires at the given
// expiry time so that it is automatically removed once the token would
// have expired anyway.
func (b *TokenBlacklist) Add(ctx context.Context, tokenID string, expiry time.Time) error {
	ttl := time.Until(expiry)
	if ttl <= 0 {
		// Token is already expired; no need to blacklist.
		return nil
	}

	key := blacklistKeyPrefix + tokenID
	if err := b.rdb.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("redis blacklist set: %w", err)
	}
	return nil
}

// IsBlacklisted checks whether a token ID has been blacklisted.
func (b *TokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) bool {
	key := blacklistKeyPrefix + tokenID
	val, err := b.rdb.Exists(ctx, key).Result()
	if err != nil {
		// On Redis error, treat token as not blacklisted (fail open).
		return false
	}
	return val > 0
}
