// Package health provides health check endpoints for the GoPlan backend.
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// ComponentCheck represents a health check for a single component.
type ComponentCheck struct {
	Name        string        `json:"name"`
	Status      Status        `json:"status"`
	Message     string        `json:"message,omitempty"`
	Duration    time.Duration `json:"duration_ms"`
	LastChecked time.Time     `json:"last_checked"`
}

// HealthResponse represents the overall health response.
type HealthResponse struct {
	Status     Status                     `json:"status"`
	Version    string                     `json:"version,omitempty"`
	Uptime     string                     `json:"uptime,omitempty"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]*ComponentCheck `json:"components,omitempty"`
}

// LivenessResponse represents the liveness probe response.
type LivenessResponse struct {
	Status    Status    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ReadinessResponse represents the readiness probe response.
type ReadinessResponse struct {
	Ready      bool                       `json:"ready"`
	Status     Status                     `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]*ComponentCheck `json:"components,omitempty"`
}

// Checker defines a health check function.
type Checker interface {
	Name() string
	Check(ctx context.Context) *ComponentCheck
}

// CheckerFunc is a function adapter for Checker.
type CheckerFunc struct {
	name      string
	checkFunc func(ctx context.Context) (Status, string, error)
}

// Name returns the checker name.
func (f *CheckerFunc) Name() string {
	return f.name
}

// Check performs the health check.
func (f *CheckerFunc) Check(ctx context.Context) *ComponentCheck {
	start := time.Now()
	status, msg, err := f.checkFunc(ctx)
	duration := time.Since(start)

	check := &ComponentCheck{
		Name:        f.name,
		Status:      status,
		Duration:    duration / time.Millisecond,
		LastChecked: time.Now(),
	}

	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = err.Error()
	} else if msg != "" {
		check.Message = msg
	}

	return check
}

// NewChecker creates a new checker from a function.
func NewChecker(name string, fn func(ctx context.Context) (Status, string, error)) Checker {
	return &CheckerFunc{name: name, checkFunc: fn}
}

// Handler provides health check HTTP handlers.
type Handler struct {
	mu         sync.RWMutex
	checkers   []Checker
	version    string
	startTime  time.Time
	timeout    time.Duration
}

// Config holds health handler configuration.
type Config struct {
	Version string
	Timeout time.Duration
}

// NewHandler creates a new health handler.
func NewHandler(config Config) *Handler {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	return &Handler{
		checkers:  make([]Checker, 0),
		version:   config.Version,
		startTime: time.Now(),
		timeout:   config.Timeout,
	}
}

// AddChecker adds a health checker.
func (h *Handler) AddChecker(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

// AddCheckerFunc adds a health checker function.
func (h *Handler) AddCheckerFunc(name string, fn func(ctx context.Context) (Status, string, error)) {
	h.AddChecker(NewChecker(name, fn))
}

// LivenessHandler returns the liveness probe handler.
// This checks if the process is running and responsive.
func (h *Handler) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := LivenessResponse{
			Status:    StatusHealthy,
			Timestamp: time.Now().UTC(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

// ReadinessHandler returns the readiness probe handler.
// This checks if the service can accept traffic (all dependencies are healthy).
func (h *Handler) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
		defer cancel()

		// Run all health checks
		components := h.runChecks(ctx)

		// Determine overall status
		overallStatus := StatusHealthy
		allHealthy := true

		for _, check := range components {
			if check.Status == StatusUnhealthy {
				overallStatus = StatusUnhealthy
				allHealthy = false
				break
			} else if check.Status == StatusDegraded {
				overallStatus = StatusDegraded
			}
		}

		response := ReadinessResponse{
			Ready:      allHealthy,
			Status:     overallStatus,
			Timestamp:  time.Now().UTC(),
			Components: components,
		}

		w.Header().Set("Content-Type", "application/json")

		if allHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// HealthHandler returns the full health check handler.
// This provides detailed health information including version and uptime.
func (h *Handler) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
		defer cancel()

		// Run all health checks
		components := h.runChecks(ctx)

		// Determine overall status
		overallStatus := StatusHealthy
		for _, check := range components {
			if check.Status == StatusUnhealthy {
				overallStatus = StatusUnhealthy
				break
			} else if check.Status == StatusDegraded {
				overallStatus = StatusDegraded
			}
		}

		response := HealthResponse{
			Status:     overallStatus,
			Version:    h.version,
			Uptime:     formatDuration(time.Since(h.startTime)),
			Timestamp:  time.Now().UTC(),
			Components: components,
		}

		w.Header().Set("Content-Type", "application/json")

		if overallStatus == StatusHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// runChecks runs all health checks concurrently.
func (h *Handler) runChecks(ctx context.Context) map[string]*ComponentCheck {
	h.mu.RLock()
	checkers := make([]Checker, len(h.checkers))
	copy(checkers, h.checkers)
	h.mu.RUnlock()

	if len(checkers) == 0 {
		return nil
	}

	results := make(map[string]*ComponentCheck)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()

			check := c.Check(ctx)

			mu.Lock()
			results[c.Name()] = check
			mu.Unlock()
		}(checker)
	}

	wg.Wait()
	return results
}

// formatDuration formats a duration as a human-readable string.
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return formatInt(days) + "d " + formatInt(hours) + "h " + formatInt(minutes) + "m"
	}
	if hours > 0 {
		return formatInt(hours) + "h " + formatInt(minutes) + "m " + formatInt(seconds) + "s"
	}
	if minutes > 0 {
		return formatInt(minutes) + "m " + formatInt(seconds) + "s"
	}
	return formatInt(seconds) + "s"
}

// formatInt formats an integer as a string.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte(n%10) + '0'}, digits...)
		n /= 10
	}

	return string(digits)
}

// Common health checkers

// DatabaseChecker creates a database health checker.
func DatabaseChecker(name string, pingFn func(ctx context.Context) error) Checker {
	return NewChecker(name, func(ctx context.Context) (Status, string, error) {
		if err := pingFn(ctx); err != nil {
			return StatusUnhealthy, "", err
		}
		return StatusHealthy, "connection ok", nil
	})
}

// RedisChecker creates a Redis health checker.
func RedisChecker(name string, pingFn func(ctx context.Context) error) Checker {
	return NewChecker(name, func(ctx context.Context) (Status, string, error) {
		if err := pingFn(ctx); err != nil {
			return StatusUnhealthy, "", err
		}
		return StatusHealthy, "connection ok", nil
	})
}

// MemoryChecker creates a memory usage health checker.
func MemoryChecker(maxBytes uint64) Checker {
	return NewChecker("memory", func(ctx context.Context) (Status, string, error) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		if m.Alloc > maxBytes {
			return StatusDegraded, "high memory usage", nil
		}
		return StatusHealthy, "", nil
	})
}

// GoroutineChecker creates a goroutine count health checker.
func GoroutineChecker(maxGoroutines int) Checker {
	return NewChecker("goroutines", func(ctx context.Context) (Status, string, error) {
		count := runtime.NumGoroutine()
		if count > maxGoroutines {
			return StatusDegraded, "high goroutine count", nil
		}
		return StatusHealthy, "", nil
	})
}

// ExternalServiceChecker creates a checker for an external service.
func ExternalServiceChecker(name, url string, timeout time.Duration) Checker {
	return NewChecker(name, func(ctx context.Context) (Status, string, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return StatusUnhealthy, "", err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return StatusUnhealthy, "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return StatusUnhealthy, "service unavailable", nil
		}
		if resp.StatusCode >= 400 {
			return StatusDegraded, "service responding with errors", nil
		}

		return StatusHealthy, "", nil
	})
}
