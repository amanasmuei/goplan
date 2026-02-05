// Package metrics provides Prometheus metrics for the GoPlan backend.
package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Metrics holds all application metrics.
type Metrics struct {
	// HTTP metrics
	httpRequestsTotal    *Counter
	httpRequestDuration  *Histogram
	httpResponseSize     *Histogram
	httpRequestsInFlight *Gauge

	// Database metrics
	dbQueryDuration     *Histogram
	dbConnectionsActive *Gauge
	dbConnectionsIdle   *Gauge
	dbConnectionsTotal  *Gauge
	dbQueryTotal        *Counter
	dbQueryErrors       *Counter

	// Business metrics
	activeUsers     *Gauge
	plansCreated    *Counter
	tasksCreated    *Counter
	tasksCompleted  *Counter
	commentsCreated *Counter

	// Auth metrics
	authAttempts      *Counter
	authFailures      *Counter
	tokensIssued      *Counter
	tokensRefreshed   *Counter
	rateLimitExceeded *Counter

	// Custom metrics registry
	mu       sync.RWMutex
	counters map[string]*Counter
	gauges   map[string]*Gauge
	histos   map[string]*Histogram
}

// Counter represents a counter metric.
type Counter struct {
	name   string
	help   string
	labels map[string]string
	value  int64
	mu     sync.Mutex
}

// Gauge represents a gauge metric.
type Gauge struct {
	name   string
	help   string
	labels map[string]string
	value  float64
	mu     sync.Mutex
}

// Histogram represents a histogram metric.
type Histogram struct {
	name    string
	help    string
	labels  map[string]string
	buckets []float64
	counts  []int64
	sum     float64
	count   int64
	mu      sync.Mutex
}

// DefaultBuckets provides default histogram buckets for latency in milliseconds.
var DefaultBuckets = []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// BytesBuckets provides histogram buckets for response size in bytes.
var BytesBuckets = []float64{100, 1000, 10000, 100000, 1000000, 10000000}

// New creates a new Metrics instance with all standard metrics initialized.
func New() *Metrics {
	m := &Metrics{
		counters: make(map[string]*Counter),
		gauges:   make(map[string]*Gauge),
		histos:   make(map[string]*Histogram),
	}

	// Initialize HTTP metrics
	m.httpRequestsTotal = m.NewCounter("http_requests_total", "Total number of HTTP requests")
	m.httpRequestDuration = m.NewHistogram("http_request_duration_ms", "HTTP request duration in milliseconds", DefaultBuckets)
	m.httpResponseSize = m.NewHistogram("http_response_size_bytes", "HTTP response size in bytes", BytesBuckets)
	m.httpRequestsInFlight = m.NewGauge("http_requests_in_flight", "Number of HTTP requests currently being processed")

	// Initialize database metrics
	m.dbQueryDuration = m.NewHistogram("db_query_duration_ms", "Database query duration in milliseconds", DefaultBuckets)
	m.dbConnectionsActive = m.NewGauge("db_connections_active", "Number of active database connections")
	m.dbConnectionsIdle = m.NewGauge("db_connections_idle", "Number of idle database connections")
	m.dbConnectionsTotal = m.NewGauge("db_connections_total", "Total number of database connections")
	m.dbQueryTotal = m.NewCounter("db_queries_total", "Total number of database queries")
	m.dbQueryErrors = m.NewCounter("db_query_errors_total", "Total number of database query errors")

	// Initialize business metrics
	m.activeUsers = m.NewGauge("active_users", "Number of active users")
	m.plansCreated = m.NewCounter("plans_created_total", "Total number of plans created")
	m.tasksCreated = m.NewCounter("tasks_created_total", "Total number of tasks created")
	m.tasksCompleted = m.NewCounter("tasks_completed_total", "Total number of tasks completed")
	m.commentsCreated = m.NewCounter("comments_created_total", "Total number of comments created")

	// Initialize auth metrics
	m.authAttempts = m.NewCounter("auth_attempts_total", "Total number of authentication attempts")
	m.authFailures = m.NewCounter("auth_failures_total", "Total number of authentication failures")
	m.tokensIssued = m.NewCounter("tokens_issued_total", "Total number of tokens issued")
	m.tokensRefreshed = m.NewCounter("tokens_refreshed_total", "Total number of tokens refreshed")
	m.rateLimitExceeded = m.NewCounter("rate_limit_exceeded_total", "Total number of rate limit exceeded events")

	return m
}

// NewCounter creates a new counter metric.
func (m *Metrics) NewCounter(name, help string) *Counter {
	c := &Counter{
		name:   name,
		help:   help,
		labels: make(map[string]string),
	}
	m.mu.Lock()
	m.counters[name] = c
	m.mu.Unlock()
	return c
}

// NewGauge creates a new gauge metric.
func (m *Metrics) NewGauge(name, help string) *Gauge {
	g := &Gauge{
		name:   name,
		help:   help,
		labels: make(map[string]string),
	}
	m.mu.Lock()
	m.gauges[name] = g
	m.mu.Unlock()
	return g
}

// NewHistogram creates a new histogram metric.
func (m *Metrics) NewHistogram(name, help string, buckets []float64) *Histogram {
	if buckets == nil {
		buckets = DefaultBuckets
	}
	h := &Histogram{
		name:    name,
		help:    help,
		labels:  make(map[string]string),
		buckets: buckets,
		counts:  make([]int64, len(buckets)+1), // +1 for +Inf bucket
	}
	m.mu.Lock()
	m.histos[name] = h
	m.mu.Unlock()
	return h
}

// Counter methods

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

// Add adds the given value to the counter.
func (c *Counter) Add(n int64) {
	c.mu.Lock()
	c.value += n
	c.mu.Unlock()
}

// Value returns the current counter value.
func (c *Counter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Gauge methods

// Set sets the gauge to the given value.
func (g *Gauge) Set(v float64) {
	g.mu.Lock()
	g.value = v
	g.mu.Unlock()
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	g.mu.Lock()
	g.value++
	g.mu.Unlock()
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	g.mu.Lock()
	g.value--
	g.mu.Unlock()
}

// Add adds the given value to the gauge.
func (g *Gauge) Add(v float64) {
	g.mu.Lock()
	g.value += v
	g.mu.Unlock()
}

// Value returns the current gauge value.
func (g *Gauge) Value() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.value
}

// Histogram methods

// Observe records a value in the histogram.
func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += v
	h.count++

	for i, bucket := range h.buckets {
		if v <= bucket {
			h.counts[i]++
			return
		}
	}
	// +Inf bucket
	h.counts[len(h.buckets)]++
}

// ObserveDuration records a duration in milliseconds.
func (h *Histogram) ObserveDuration(d time.Duration) {
	h.Observe(float64(d.Milliseconds()))
}

// HTTP metric methods

// RecordHTTPRequest records an HTTP request.
func (m *Metrics) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, responseSize int64) {
	m.httpRequestsTotal.Inc()
	m.httpRequestDuration.ObserveDuration(duration)
	m.httpResponseSize.Observe(float64(responseSize))
}

// IncHTTPRequestsInFlight increments the in-flight request count.
func (m *Metrics) IncHTTPRequestsInFlight() {
	m.httpRequestsInFlight.Inc()
}

// DecHTTPRequestsInFlight decrements the in-flight request count.
func (m *Metrics) DecHTTPRequestsInFlight() {
	m.httpRequestsInFlight.Dec()
}

// Database metric methods

// RecordDBQuery records a database query.
func (m *Metrics) RecordDBQuery(duration time.Duration, err error) {
	m.dbQueryTotal.Inc()
	m.dbQueryDuration.ObserveDuration(duration)
	if err != nil {
		m.dbQueryErrors.Inc()
	}
}

// SetDBConnectionStats sets database connection statistics.
func (m *Metrics) SetDBConnectionStats(active, idle, total int) {
	m.dbConnectionsActive.Set(float64(active))
	m.dbConnectionsIdle.Set(float64(idle))
	m.dbConnectionsTotal.Set(float64(total))
}

// Business metric methods

// SetActiveUsers sets the number of active users.
func (m *Metrics) SetActiveUsers(count int) {
	m.activeUsers.Set(float64(count))
}

// IncPlansCreated increments the plans created counter.
func (m *Metrics) IncPlansCreated() {
	m.plansCreated.Inc()
}

// IncTasksCreated increments the tasks created counter.
func (m *Metrics) IncTasksCreated() {
	m.tasksCreated.Inc()
}

// IncTasksCompleted increments the tasks completed counter.
func (m *Metrics) IncTasksCompleted() {
	m.tasksCompleted.Inc()
}

// IncCommentsCreated increments the comments created counter.
func (m *Metrics) IncCommentsCreated() {
	m.commentsCreated.Inc()
}

// Auth metric methods

// RecordAuthAttempt records an authentication attempt.
func (m *Metrics) RecordAuthAttempt(success bool) {
	m.authAttempts.Inc()
	if !success {
		m.authFailures.Inc()
	}
}

// IncTokensIssued increments the tokens issued counter.
func (m *Metrics) IncTokensIssued() {
	m.tokensIssued.Inc()
}

// IncTokensRefreshed increments the tokens refreshed counter.
func (m *Metrics) IncTokensRefreshed() {
	m.tokensRefreshed.Inc()
}

// IncRateLimitExceeded increments the rate limit exceeded counter.
func (m *Metrics) IncRateLimitExceeded() {
	m.rateLimitExceeded.Inc()
}

// Handler returns an HTTP handler for the /metrics endpoint.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		// Write counters
		m.mu.RLock()
		for _, c := range m.counters {
			c.mu.Lock()
			_, _ = w.Write([]byte("# HELP " + c.name + " " + c.help + "\n"))
			_, _ = w.Write([]byte("# TYPE " + c.name + " counter\n"))
			_, _ = w.Write([]byte(c.name + " " + strconv.FormatInt(c.value, 10) + "\n"))
			c.mu.Unlock()
		}

		// Write gauges
		for _, g := range m.gauges {
			g.mu.Lock()
			_, _ = w.Write([]byte("# HELP " + g.name + " " + g.help + "\n"))
			_, _ = w.Write([]byte("# TYPE " + g.name + " gauge\n"))
			_, _ = w.Write([]byte(g.name + " " + strconv.FormatFloat(g.value, 'f', -1, 64) + "\n"))
			g.mu.Unlock()
		}

		// Write histograms
		for _, h := range m.histos {
			h.mu.Lock()
			_, _ = w.Write([]byte("# HELP " + h.name + " " + h.help + "\n"))
			_, _ = w.Write([]byte("# TYPE " + h.name + " histogram\n"))

			cumulative := int64(0)
			for i, bucket := range h.buckets {
				cumulative += h.counts[i]
				_, _ = w.Write([]byte(h.name + "_bucket{le=\"" + strconv.FormatFloat(bucket, 'f', -1, 64) + "\"} " + strconv.FormatInt(cumulative, 10) + "\n"))
			}
			cumulative += h.counts[len(h.buckets)]
			_, _ = w.Write([]byte(h.name + "_bucket{le=\"+Inf\"} " + strconv.FormatInt(cumulative, 10) + "\n"))
			_, _ = w.Write([]byte(h.name + "_sum " + strconv.FormatFloat(h.sum, 'f', -1, 64) + "\n"))
			_, _ = w.Write([]byte(h.name + "_count " + strconv.FormatInt(h.count, 10) + "\n"))
			h.mu.Unlock()
		}
		m.mu.RUnlock()
	})
}

// Global metrics instance
var defaultMetrics *Metrics
var metricsOnce sync.Once

// Default returns the default metrics instance.
func Default() *Metrics {
	metricsOnce.Do(func() {
		defaultMetrics = New()
	})
	return defaultMetrics
}

// SetDefault sets the default metrics instance.
func SetDefault(m *Metrics) {
	defaultMetrics = m
}
