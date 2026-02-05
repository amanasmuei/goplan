// Package logging provides structured JSON logging for the GoPlan backend.
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Context keys for logging.
type contextKey string

const (
	// ContextKeyRequestID is the context key for the request ID.
	ContextKeyRequestID contextKey = "requestID"
	// ContextKeyUserID is the context key for the user ID.
	ContextKeyUserID contextKey = "userID"
	// ContextKeyWorkspaceID is the context key for the workspace ID.
	ContextKeyWorkspaceID contextKey = "workspaceID"
	// ContextKeyTraceID is the context key for trace ID.
	ContextKeyTraceID contextKey = "traceID"
)

// Logger wraps slog.Logger with additional functionality.
type Logger struct {
	*slog.Logger
	level slog.Level
}

// Config holds logger configuration.
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	Output     io.Writer
	AddSource  bool
	TimeFormat string
}

// DefaultConfig returns the default logger configuration.
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "json",
		Output:     os.Stdout,
		AddSource:  false,
		TimeFormat: time.RFC3339,
	}
}

// New creates a new structured logger with the given configuration.
func New(cfg Config) *Logger {
	level := parseLevel(cfg.Level)

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Customize time format
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					return slog.String(slog.TimeKey, t.Format(cfg.TimeFormat))
				}
			}
			return a
		},
	}

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	if cfg.Format == "text" {
		handler = slog.NewTextHandler(output, opts)
	} else {
		handler = slog.NewJSONHandler(output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		level:  level,
	}
}

// parseLevel parses a log level string into a slog.Level.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext creates a new logger with context fields extracted.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	attrs := make([]slog.Attr, 0, 4)

	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok && requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok && userID != "" {
		attrs = append(attrs, slog.String("user_id", userID))
	}
	if workspaceID, ok := ctx.Value(ContextKeyWorkspaceID).(string); ok && workspaceID != "" {
		attrs = append(attrs, slog.String("workspace_id", workspaceID))
	}
	if traceID, ok := ctx.Value(ContextKeyTraceID).(string); ok && traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}

	if len(attrs) == 0 {
		return l
	}

	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}

	return &Logger{
		Logger: l.Logger.With(args...),
		level:  l.level,
	}
}

// WithFields creates a new logger with additional fields.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
		level:  l.level,
	}
}

// WithError creates a new logger with an error field.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
		level:  l.level,
	}
}

// WithComponent creates a new logger with a component field.
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", component),
		level:  l.level,
	}
}

// WithOperation creates a new logger with an operation field.
func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With("operation", operation),
		level:  l.level,
	}
}

// HTTPRequest logs an HTTP request.
func (l *Logger) HTTPRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration, bytesWritten int64) {
	l.WithContext(ctx).Info("http_request",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"bytes_written", bytesWritten,
	)
}

// HTTPError logs an HTTP error.
func (l *Logger) HTTPError(ctx context.Context, method, path string, statusCode int, err error) {
	l.WithContext(ctx).Error("http_error",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"error", err.Error(),
	)
}

// DBQuery logs a database query.
func (l *Logger) DBQuery(ctx context.Context, query string, duration time.Duration, err error) {
	logger := l.WithContext(ctx)
	if err != nil {
		logger.Error("db_query_error",
			"query", truncateQuery(query, 200),
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
		)
	} else {
		logger.Debug("db_query",
			"query", truncateQuery(query, 200),
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// Audit logs an audit event.
func (l *Logger) Audit(ctx context.Context, action, resource, resourceID string, details map[string]interface{}) {
	args := []any{
		"action", action,
		"resource", resource,
		"resource_id", resourceID,
	}
	for k, v := range details {
		args = append(args, k, v)
	}
	l.WithContext(ctx).Info("audit", args...)
}

// Panic logs a panic with stack trace.
func (l *Logger) Panic(ctx context.Context, panicValue interface{}, stack []byte) {
	l.WithContext(ctx).Error("panic",
		"panic", panicValue,
		"stack", string(stack),
	)
}

// Startup logs application startup information.
func (l *Logger) Startup(version, env, listenAddr string) {
	l.Info("application_startup",
		"version", version,
		"environment", env,
		"listen_addr", listenAddr,
		"go_version", runtime.Version(),
		"num_cpu", runtime.NumCPU(),
	)
}

// Shutdown logs application shutdown.
func (l *Logger) Shutdown(reason string) {
	l.Info("application_shutdown",
		"reason", reason,
	)
}

// truncateQuery truncates a query string to the specified length.
func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

// ContextWithRequestID adds a request ID to the context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// ContextWithUserID adds a user ID to the context.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// ContextWithWorkspaceID adds a workspace ID to the context.
func ContextWithWorkspaceID(ctx context.Context, workspaceID string) context.Context {
	return context.WithValue(ctx, ContextKeyWorkspaceID, workspaceID)
}

// ContextWithTraceID adds a trace ID to the context.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ContextKeyTraceID, traceID)
}

// GetRequestID gets the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// GetUserID gets the user ID from context.
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return id
	}
	return ""
}

// GetWorkspaceID gets the workspace ID from context.
func GetWorkspaceID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyWorkspaceID).(string); ok {
		return id
	}
	return ""
}

// Global logger instance.
var defaultLogger *Logger

// SetDefault sets the default global logger.
func SetDefault(l *Logger) {
	defaultLogger = l
	slog.SetDefault(l.Logger)
}

// Default returns the default global logger.
func Default() *Logger {
	if defaultLogger == nil {
		defaultLogger = New(DefaultConfig())
	}
	return defaultLogger
}

// Debug logs at debug level using the default logger.
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// Info logs at info level using the default logger.
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Warn logs at warn level using the default logger.
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Error logs at error level using the default logger.
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}
