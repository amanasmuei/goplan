package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	config := Config{
		DefaultRequestsPerSecond: 10,
		DefaultBurstSize:         10,
		IPRequestsPerSecond:      5,
		IPBurstSize:              5,
		UserRequestsPerSecond:    10,
		UserBurstSize:            10,
		CleanupInterval:          1 * time.Minute,
	}

	limiter := New(config, nil, nil)
	defer limiter.Stop()

	t.Run("Allow within rate limit", func(t *testing.T) {
		key := "test-within-limit"
		for i := 0; i < 5; i++ {
			if !limiter.Allow(key, 10, 10) {
				t.Errorf("Request %d should be allowed", i)
			}
		}
	})

	t.Run("Block when rate exceeded", func(t *testing.T) {
		key := "test-exceed-limit"
		// Use all tokens
		for i := 0; i < 3; i++ {
			limiter.Allow(key, 2, 3)
		}

		// This should be blocked
		if limiter.Allow(key, 2, 3) {
			t.Error("Request should be blocked when rate exceeded")
		}
	})

	t.Run("Tokens replenish over time", func(t *testing.T) {
		key := "test-replenish"
		rate := float64(100) // 100 per second
		burst := 5

		// Use all tokens
		for i := 0; i < burst; i++ {
			limiter.Allow(key, rate, burst)
		}

		// Wait for tokens to replenish
		time.Sleep(100 * time.Millisecond)

		// Should have new tokens now
		if !limiter.Allow(key, rate, burst) {
			t.Error("Request should be allowed after tokens replenish")
		}
	})
}

func TestLimiter_AllowIP(t *testing.T) {
	config := Config{
		IPRequestsPerSecond: 2,
		IPBurstSize:         3,
		CleanupInterval:     1 * time.Minute,
	}

	limiter := New(config, nil, nil)
	defer limiter.Stop()

	ip := "192.168.1.1"

	// Should allow burst
	for i := 0; i < 3; i++ {
		if !limiter.AllowIP(ip) {
			t.Errorf("IP request %d should be allowed", i)
		}
	}

	// Should block after burst
	if limiter.AllowIP(ip) {
		t.Error("IP request should be blocked after burst")
	}
}

func TestLimiter_AllowUser(t *testing.T) {
	config := Config{
		UserRequestsPerSecond: 2,
		UserBurstSize:         3,
		CleanupInterval:       1 * time.Minute,
	}

	limiter := New(config, nil, nil)
	defer limiter.Stop()

	userID := "user-123"

	// Should allow burst
	for i := 0; i < 3; i++ {
		if !limiter.AllowUser(userID) {
			t.Errorf("User request %d should be allowed", i)
		}
	}

	// Should block after burst
	if limiter.AllowUser(userID) {
		t.Error("User request should be blocked after burst")
	}
}

func TestLimiter_AllowEndpoint(t *testing.T) {
	config := Config{
		DefaultRequestsPerSecond: 5,
		DefaultBurstSize:         5,
		EndpointLimits: map[string]EndpointLimit{
			"/api/v1/login": {
				RequestsPerSecond: 2,
				BurstSize:         2,
			},
		},
		CleanupInterval: 1 * time.Minute,
	}

	limiter := New(config, nil, nil)
	defer limiter.Stop()

	t.Run("Endpoint with custom limit", func(t *testing.T) {
		endpoint := "/api/v1/login"
		identifier := "192.168.1.1"

		// Should allow custom burst
		for i := 0; i < 2; i++ {
			if !limiter.AllowEndpoint(endpoint, identifier) {
				t.Errorf("Endpoint request %d should be allowed", i)
			}
		}

		// Should block after custom burst
		if limiter.AllowEndpoint(endpoint, identifier) {
			t.Error("Endpoint request should be blocked after burst")
		}
	})

	t.Run("Endpoint with default limit", func(t *testing.T) {
		endpoint := "/api/v1/tasks"
		identifier := "192.168.1.2"

		// Should allow default burst
		for i := 0; i < 5; i++ {
			if !limiter.AllowEndpoint(endpoint, identifier) {
				t.Errorf("Endpoint request %d should be allowed", i)
			}
		}
	})
}

func TestLimiter_RemainingRequests(t *testing.T) {
	config := Config{
		DefaultRequestsPerSecond: 10,
		DefaultBurstSize:         10,
		CleanupInterval:          1 * time.Minute,
	}

	limiter := New(config, nil, nil)
	defer limiter.Stop()

	key := "test-remaining"

	// Initial remaining should be burst size
	remaining := limiter.RemainingRequests(key)
	if remaining != 10 {
		t.Errorf("Expected 10 remaining, got %d", remaining)
	}

	// Use some tokens
	for i := 0; i < 5; i++ {
		limiter.Allow(key, 10, 10)
	}

	remaining = limiter.RemainingRequests(key)
	if remaining != 5 {
		t.Errorf("Expected 5 remaining, got %d", remaining)
	}
}

func TestLimiter_SetEndpointLimit(t *testing.T) {
	config := Config{
		DefaultRequestsPerSecond: 10,
		DefaultBurstSize:         10,
		EndpointLimits:           make(map[string]EndpointLimit),
		CleanupInterval:          1 * time.Minute,
	}

	limiter := New(config, nil, nil)
	defer limiter.Stop()

	// Set new endpoint limit
	limiter.SetEndpointLimit("/api/v1/new", 5, 5)

	// Verify limit is applied
	for i := 0; i < 5; i++ {
		if !limiter.AllowEndpoint("/api/v1/new", "192.168.1.1") {
			t.Errorf("Request %d should be allowed", i)
		}
	}

	if limiter.AllowEndpoint("/api/v1/new", "192.168.1.1") {
		t.Error("Request should be blocked after limit")
	}
}
