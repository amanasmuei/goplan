// Package ratelimit provides rate limiting functionality for the GoPlan backend.
package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/goplan/goplan/internal/logging"
	"github.com/goplan/goplan/internal/metrics"
)

// Config holds rate limiter configuration.
type Config struct {
	// Default rate limits
	DefaultRequestsPerSecond int
	DefaultBurstSize         int

	// Per-IP rate limits
	IPRequestsPerSecond int
	IPBurstSize         int

	// Per-user rate limits
	UserRequestsPerSecond int
	UserBurstSize         int

	// Endpoint-specific limits
	EndpointLimits map[string]EndpointLimit

	// Cleanup interval for expired entries
	CleanupInterval time.Duration

	// Whether to use Redis (if nil, uses in-memory)
	RedisClient RedisClient
}

// EndpointLimit defines rate limits for a specific endpoint.
type EndpointLimit struct {
	RequestsPerSecond int
	BurstSize         int
}

// RedisClient interface for Redis-based rate limiting.
type RedisClient interface {
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// DefaultConfig returns the default rate limiter configuration.
func DefaultConfig() Config {
	return Config{
		DefaultRequestsPerSecond: 100,
		DefaultBurstSize:         200,
		IPRequestsPerSecond:      50,
		IPBurstSize:              100,
		UserRequestsPerSecond:    100,
		UserBurstSize:            200,
		CleanupInterval:          5 * time.Minute,
		EndpointLimits:           make(map[string]EndpointLimit),
	}
}

// Limiter provides rate limiting functionality.
type Limiter struct {
	config  Config
	logger  *logging.Logger
	metrics *metrics.Metrics

	// In-memory storage
	mu      sync.RWMutex
	buckets map[string]*bucket
	stopCh  chan struct{}
}

// bucket represents a token bucket for rate limiting.
type bucket struct {
	tokens     float64
	lastUpdate time.Time
	rate       float64 // tokens per second
	burst      int     // maximum tokens
}

// New creates a new rate limiter.
func New(config Config, logger *logging.Logger, metrics *metrics.Metrics) *Limiter {
	l := &Limiter{
		config:  config,
		logger:  logger,
		metrics: metrics,
		buckets: make(map[string]*bucket),
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine for in-memory rate limiter
	if config.RedisClient == nil {
		go l.cleanup()
	}

	return l
}

// Stop stops the rate limiter cleanup goroutine.
func (l *Limiter) Stop() {
	close(l.stopCh)
}

// cleanup periodically removes expired buckets.
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(l.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for key, b := range l.buckets {
				// Remove buckets that haven't been used in 10 minutes
				if now.Sub(b.lastUpdate) > 10*time.Minute {
					delete(l.buckets, key)
				}
			}
			l.mu.Unlock()
		case <-l.stopCh:
			return
		}
	}
}

// Allow checks if a request should be allowed based on rate limits.
func (l *Limiter) Allow(key string, rate float64, burst int) bool {
	if l.config.RedisClient != nil {
		return l.allowRedis(key, rate, burst)
	}
	return l.allowMemory(key, rate, burst)
}

// allowMemory implements in-memory token bucket rate limiting.
func (l *Limiter) allowMemory(key string, rate float64, burst int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	b, exists := l.buckets[key]
	if !exists {
		l.buckets[key] = &bucket{
			tokens:     float64(burst) - 1, // Use one token
			lastUpdate: now,
			rate:       rate,
			burst:      burst,
		}
		return true
	}

	// Calculate tokens to add based on time passed
	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > float64(b.burst) {
		b.tokens = float64(b.burst)
	}
	b.lastUpdate = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

// allowRedis implements Redis-based rate limiting using sliding window.
func (l *Limiter) allowRedis(key string, rate float64, burst int) bool {
	ctx := context.Background()

	windowKey := fmt.Sprintf("ratelimit:%s:%d", key, time.Now().Unix())

	count, err := l.config.RedisClient.Incr(ctx, windowKey)
	if err != nil {
		// On Redis error, allow the request (fail open)
		if l.logger != nil {
			l.logger.WithError(err).Warn("rate limit redis error, allowing request")
		}
		return true
	}

	// Set expiry on first increment
	if count == 1 {
		_ = l.config.RedisClient.Expire(ctx, windowKey, 2*time.Second)
	}

	return int(count) <= int(rate)
}

// AllowIP checks rate limit for an IP address.
func (l *Limiter) AllowIP(ip string) bool {
	key := "ip:" + ip
	return l.Allow(key, float64(l.config.IPRequestsPerSecond), l.config.IPBurstSize)
}

// AllowUser checks rate limit for a user.
func (l *Limiter) AllowUser(userID string) bool {
	key := "user:" + userID
	return l.Allow(key, float64(l.config.UserRequestsPerSecond), l.config.UserBurstSize)
}

// AllowEndpoint checks rate limit for a specific endpoint.
func (l *Limiter) AllowEndpoint(endpoint, identifier string) bool {
	limit, exists := l.config.EndpointLimits[endpoint]
	if !exists {
		limit = EndpointLimit{
			RequestsPerSecond: l.config.DefaultRequestsPerSecond,
			BurstSize:         l.config.DefaultBurstSize,
		}
	}

	key := fmt.Sprintf("endpoint:%s:%s", endpoint, identifier)
	return l.Allow(key, float64(limit.RequestsPerSecond), limit.BurstSize)
}

// Middleware returns rate limiting middleware.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := l.getClientIP(r)

		// Check IP rate limit
		if !l.AllowIP(ip) {
			l.handleRateLimited(w, r, "IP rate limit exceeded")
			return
		}

		// Check user rate limit if authenticated
		if userID := r.Header.Get("X-User-ID"); userID != "" {
			if !l.AllowUser(userID) {
				l.handleRateLimited(w, r, "User rate limit exceeded")
				return
			}
		}

		// Check endpoint-specific rate limit
		if !l.AllowEndpoint(r.URL.Path, ip) {
			l.handleRateLimited(w, r, "Endpoint rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MiddlewareForEndpoint returns rate limiting middleware for a specific endpoint.
func (l *Limiter) MiddlewareForEndpoint(endpoint string, requestsPerSecond, burstSize int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := l.getClientIP(r)
			key := fmt.Sprintf("endpoint:%s:%s", endpoint, ip)

			if !l.Allow(key, float64(requestsPerSecond), burstSize) {
				l.handleRateLimited(w, r, "Rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request.
func (l *Limiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := len(xff); idx > 0 {
			ips := xff
			for i, c := range xff {
				if c == ',' {
					ips = xff[:i]
					break
				}
			}
			if ip := net.ParseIP(ips); ip != nil {
				return ip.String()
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// handleRateLimited handles a rate-limited request.
func (l *Limiter) handleRateLimited(w http.ResponseWriter, r *http.Request, message string) {
	if l.metrics != nil {
		l.metrics.IncRateLimitExceeded()
	}

	if l.logger != nil {
		l.logger.WithContext(r.Context()).Warn("rate_limit_exceeded",
			"path", r.URL.Path,
			"method", r.Method,
			"remote_addr", r.RemoteAddr,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "1")
	w.WriteHeader(http.StatusTooManyRequests)

	response := map[string]interface{}{
		"error":   true,
		"code":    "RATE_LIMIT_EXCEEDED",
		"message": message,
	}

	_ = json.NewEncoder(w).Encode(response)
}

// RemainingRequests returns the number of remaining requests for a key.
func (l *Limiter) RemainingRequests(key string) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	b, exists := l.buckets[key]
	if !exists {
		return l.config.DefaultBurstSize
	}

	// Update tokens based on time
	elapsed := time.Since(b.lastUpdate).Seconds()
	tokens := b.tokens + elapsed*b.rate
	if tokens > float64(b.burst) {
		tokens = float64(b.burst)
	}

	return int(tokens)
}

// SetEndpointLimit sets rate limits for a specific endpoint.
func (l *Limiter) SetEndpointLimit(endpoint string, requestsPerSecond, burstSize int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.EndpointLimits[endpoint] = EndpointLimit{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
	}
}
