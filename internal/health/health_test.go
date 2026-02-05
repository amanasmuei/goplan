package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler_LivenessHandler(t *testing.T) {
	handler := NewHandler(Config{
		Version: "1.0.0",
		Timeout: 5 * time.Second,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.LivenessHandler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response LivenessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("Expected status healthy, got %s", response.Status)
	}
}

func TestHealthHandler_ReadinessHandler(t *testing.T) {
	t.Run("All checks pass", func(t *testing.T) {
		handler := NewHandler(Config{
			Version: "1.0.0",
			Timeout: 5 * time.Second,
		})

		// Add a passing checker
		handler.AddCheckerFunc("database", func(ctx context.Context) (Status, string, error) {
			return StatusHealthy, "connected", nil
		})

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		handler.ReadinessHandler()(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ReadinessResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if !response.Ready {
			t.Error("Expected ready to be true")
		}
		if response.Status != StatusHealthy {
			t.Errorf("Expected status healthy, got %s", response.Status)
		}
	})

	t.Run("One check fails", func(t *testing.T) {
		handler := NewHandler(Config{
			Version: "1.0.0",
			Timeout: 5 * time.Second,
		})

		// Add a failing checker
		handler.AddCheckerFunc("database", func(ctx context.Context) (Status, string, error) {
			return StatusUnhealthy, "", errors.New("connection refused")
		})

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		handler.ReadinessHandler()(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected status 503, got %d", w.Code)
		}

		var response ReadinessResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response.Ready {
			t.Error("Expected ready to be false")
		}
		if response.Status != StatusUnhealthy {
			t.Errorf("Expected status unhealthy, got %s", response.Status)
		}
	})

	t.Run("Degraded status", func(t *testing.T) {
		handler := NewHandler(Config{
			Version: "1.0.0",
			Timeout: 5 * time.Second,
		})

		handler.AddCheckerFunc("database", func(ctx context.Context) (Status, string, error) {
			return StatusHealthy, "", nil
		})
		handler.AddCheckerFunc("cache", func(ctx context.Context) (Status, string, error) {
			return StatusDegraded, "high latency", nil
		})

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()

		handler.ReadinessHandler()(w, req)

		var response ReadinessResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if response.Status != StatusDegraded {
			t.Errorf("Expected status degraded, got %s", response.Status)
		}
	})
}

func TestHealthHandler_HealthHandler(t *testing.T) {
	handler := NewHandler(Config{
		Version: "1.0.0",
		Timeout: 5 * time.Second,
	})

	handler.AddCheckerFunc("database", func(ctx context.Context) (Status, string, error) {
		return StatusHealthy, "connected", nil
	})
	handler.AddCheckerFunc("redis", func(ctx context.Context) (Status, string, error) {
		return StatusHealthy, "connected", nil
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	handler.HealthHandler()(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("Expected status healthy, got %s", response.Status)
	}
	if response.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", response.Version)
	}
	if len(response.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(response.Components))
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHandler(Config{
		Version: "1.0.0",
		Timeout: 5 * time.Second,
	})

	// Test POST method
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	handler.LivenessHandler()(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestDatabaseChecker(t *testing.T) {
	t.Run("Successful ping", func(t *testing.T) {
		checker := DatabaseChecker("db", func(ctx context.Context) error {
			return nil
		})

		check := checker.Check(context.Background())
		if check.Status != StatusHealthy {
			t.Errorf("Expected healthy, got %s", check.Status)
		}
	})

	t.Run("Failed ping", func(t *testing.T) {
		checker := DatabaseChecker("db", func(ctx context.Context) error {
			return errors.New("connection refused")
		})

		check := checker.Check(context.Background())
		if check.Status != StatusUnhealthy {
			t.Errorf("Expected unhealthy, got %s", check.Status)
		}
		if check.Message == "" {
			t.Error("Expected error message")
		}
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m 0s"},
		{2 * time.Hour, "2h 0m 0s"},
		{25 * time.Hour, "1d 1h 0m"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
