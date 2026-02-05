// Package audit provides audit logging for compliance and security monitoring.
package audit

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/goplan/goplan/internal/logging"
)

// ActionType represents the type of audited action.
type ActionType string

// Action types for audit logging.
const (
	// Authentication actions
	ActionLogin         ActionType = "auth.login"
	ActionLogout        ActionType = "auth.logout"
	ActionTokenRefresh  ActionType = "auth.token_refresh"
	ActionPasswordChange ActionType = "auth.password_change"
	ActionPasswordReset ActionType = "auth.password_reset"

	// User actions
	ActionUserCreate ActionType = "user.create"
	ActionUserUpdate ActionType = "user.update"
	ActionUserDelete ActionType = "user.delete"
	ActionUserInvite ActionType = "user.invite"

	// Workspace actions
	ActionWorkspaceCreate ActionType = "workspace.create"
	ActionWorkspaceUpdate ActionType = "workspace.update"
	ActionWorkspaceDelete ActionType = "workspace.delete"
	ActionMemberAdd       ActionType = "workspace.member_add"
	ActionMemberRemove    ActionType = "workspace.member_remove"
	ActionMemberRoleChange ActionType = "workspace.member_role_change"

	// Plan actions
	ActionPlanCreate ActionType = "plan.create"
	ActionPlanUpdate ActionType = "plan.update"
	ActionPlanDelete ActionType = "plan.delete"
	ActionPlanArchive ActionType = "plan.archive"

	// Task actions
	ActionTaskCreate   ActionType = "task.create"
	ActionTaskUpdate   ActionType = "task.update"
	ActionTaskDelete   ActionType = "task.delete"
	ActionTaskAssign   ActionType = "task.assign"
	ActionTaskComplete ActionType = "task.complete"

	// Data access actions
	ActionDataExport ActionType = "data.export"
	ActionDataImport ActionType = "data.import"

	// Admin actions
	ActionSettingsChange ActionType = "admin.settings_change"
	ActionAPIKeyCreate   ActionType = "admin.api_key_create"
	ActionAPIKeyRevoke   ActionType = "admin.api_key_revoke"
)

// Severity represents the severity level of an audit event.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// AuditEvent represents an audit log entry.
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Action      ActionType             `json:"action"`
	Severity    Severity               `json:"severity"`
	ActorID     string                 `json:"actor_id"`
	ActorType   string                 `json:"actor_type"` // user, system, api_key
	ActorEmail  string                 `json:"actor_email,omitempty"`
	TargetType  string                 `json:"target_type"`  // user, workspace, plan, task, etc.
	TargetID    string                 `json:"target_id"`
	TargetName  string                 `json:"target_name,omitempty"`
	WorkspaceID string                 `json:"workspace_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Success     bool                   `json:"success"`
	FailReason  string                 `json:"fail_reason,omitempty"`
}

// AuditLogger provides audit logging functionality.
type AuditLogger struct {
	logger  *logging.Logger
	store   AuditStore
	enabled bool
}

// AuditStore interface for persisting audit events.
type AuditStore interface {
	Save(ctx context.Context, event *AuditEvent) error
	List(ctx context.Context, filter AuditFilter, limit, offset int) ([]*AuditEvent, int64, error)
}

// AuditFilter for querying audit events.
type AuditFilter struct {
	ActorID     string
	TargetType  string
	TargetID    string
	WorkspaceID string
	Action      ActionType
	StartTime   *time.Time
	EndTime     *time.Time
	Success     *bool
}

// Config holds audit logger configuration.
type Config struct {
	Enabled bool
	Logger  *logging.Logger
	Store   AuditStore
}

// New creates a new audit logger.
func New(config Config) *AuditLogger {
	return &AuditLogger{
		logger:  config.Logger,
		store:   config.Store,
		enabled: config.Enabled,
	}
}

// Log logs an audit event.
func (a *AuditLogger) Log(ctx context.Context, event *AuditEvent) error {
	if !a.enabled {
		return nil
	}

	// Set defaults
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Severity == "" {
		event.Severity = SeverityInfo
	}
	if event.ActorType == "" {
		event.ActorType = "user"
	}

	// Extract request ID from context if not set
	if event.RequestID == "" {
		event.RequestID = logging.GetRequestID(ctx)
	}

	// Log to structured logger
	if a.logger != nil {
		a.logToLogger(event)
	}

	// Persist to store if available
	if a.store != nil {
		if err := a.store.Save(ctx, event); err != nil {
			if a.logger != nil {
				a.logger.WithError(err).Error("failed to save audit event")
			}
			return err
		}
	}

	return nil
}

// logToLogger logs the event to the structured logger.
func (a *AuditLogger) logToLogger(event *AuditEvent) {
	args := []any{
		"audit_id", event.ID,
		"action", event.Action,
		"severity", event.Severity,
		"actor_id", event.ActorID,
		"actor_type", event.ActorType,
		"target_type", event.TargetType,
		"target_id", event.TargetID,
		"success", event.Success,
	}

	if event.WorkspaceID != "" {
		args = append(args, "workspace_id", event.WorkspaceID)
	}
	if event.IPAddress != "" {
		args = append(args, "ip_address", event.IPAddress)
	}
	if event.RequestID != "" {
		args = append(args, "request_id", event.RequestID)
	}
	if !event.Success && event.FailReason != "" {
		args = append(args, "fail_reason", event.FailReason)
	}

	a.logger.Info("audit_event", args...)
}

// LogSuccess logs a successful audit event.
func (a *AuditLogger) LogSuccess(ctx context.Context, action ActionType, actorID, targetType, targetID string, details map[string]interface{}) error {
	event := &AuditEvent{
		Action:     action,
		ActorID:    actorID,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
		Success:    true,
	}
	return a.Log(ctx, event)
}

// LogFailure logs a failed audit event.
func (a *AuditLogger) LogFailure(ctx context.Context, action ActionType, actorID, targetType, targetID, reason string, details map[string]interface{}) error {
	event := &AuditEvent{
		Action:     action,
		ActorID:    actorID,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
		Success:    false,
		FailReason: reason,
	}
	return a.Log(ctx, event)
}

// LogFromRequest logs an audit event with information from the HTTP request.
func (a *AuditLogger) LogFromRequest(r *http.Request, event *AuditEvent) error {
	// Extract IP address
	event.IPAddress = getClientIP(r)
	event.UserAgent = r.UserAgent()

	// Extract request ID from context
	if event.RequestID == "" {
		event.RequestID = logging.GetRequestID(r.Context())
	}

	return a.Log(r.Context(), event)
}

// AuthLogin logs an authentication login attempt.
func (a *AuditLogger) AuthLogin(ctx context.Context, userID, email, ipAddress string, success bool, failReason string) error {
	event := &AuditEvent{
		Action:     ActionLogin,
		ActorID:    userID,
		ActorEmail: email,
		TargetType: "user",
		TargetID:   userID,
		IPAddress:  ipAddress,
		Success:    success,
		FailReason: failReason,
	}
	if !success {
		event.Severity = SeverityWarning
	}
	return a.Log(ctx, event)
}

// AuthLogout logs an authentication logout.
func (a *AuditLogger) AuthLogout(ctx context.Context, userID string) error {
	return a.LogSuccess(ctx, ActionLogout, userID, "user", userID, nil)
}

// UserCreate logs a user creation.
func (a *AuditLogger) UserCreate(ctx context.Context, actorID, newUserID, newUserEmail string) error {
	return a.LogSuccess(ctx, ActionUserCreate, actorID, "user", newUserID, map[string]interface{}{
		"email": newUserEmail,
	})
}

// WorkspaceMemberAdd logs adding a member to a workspace.
func (a *AuditLogger) WorkspaceMemberAdd(ctx context.Context, actorID, workspaceID, memberID, role string) error {
	event := &AuditEvent{
		Action:      ActionMemberAdd,
		ActorID:     actorID,
		TargetType:  "user",
		TargetID:    memberID,
		WorkspaceID: workspaceID,
		Details: map[string]interface{}{
			"role": role,
		},
		Success: true,
	}
	return a.Log(ctx, event)
}

// PlanCreate logs a plan creation.
func (a *AuditLogger) PlanCreate(ctx context.Context, actorID, workspaceID, planID, planName string) error {
	event := &AuditEvent{
		Action:      ActionPlanCreate,
		ActorID:     actorID,
		TargetType:  "plan",
		TargetID:    planID,
		TargetName:  planName,
		WorkspaceID: workspaceID,
		Success:     true,
	}
	return a.Log(ctx, event)
}

// TaskCreate logs a task creation.
func (a *AuditLogger) TaskCreate(ctx context.Context, actorID, workspaceID, planID, taskID, taskTitle string) error {
	event := &AuditEvent{
		Action:      ActionTaskCreate,
		ActorID:     actorID,
		TargetType:  "task",
		TargetID:    taskID,
		TargetName:  taskTitle,
		WorkspaceID: workspaceID,
		Details: map[string]interface{}{
			"plan_id": planID,
		},
		Success: true,
	}
	return a.Log(ctx, event)
}

// TaskComplete logs a task completion.
func (a *AuditLogger) TaskComplete(ctx context.Context, actorID, workspaceID, taskID string) error {
	event := &AuditEvent{
		Action:      ActionTaskComplete,
		ActorID:     actorID,
		TargetType:  "task",
		TargetID:    taskID,
		WorkspaceID: workspaceID,
		Success:     true,
	}
	return a.Log(ctx, event)
}

// DataExport logs a data export.
func (a *AuditLogger) DataExport(ctx context.Context, actorID, workspaceID, exportType string, success bool) error {
	event := &AuditEvent{
		Action:      ActionDataExport,
		Severity:    SeverityCritical,
		ActorID:     actorID,
		TargetType:  "workspace",
		TargetID:    workspaceID,
		WorkspaceID: workspaceID,
		Details: map[string]interface{}{
			"export_type": exportType,
		},
		Success: success,
	}
	return a.Log(ctx, event)
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
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

// Query queries audit events.
func (a *AuditLogger) Query(ctx context.Context, filter AuditFilter, limit, offset int) ([]*AuditEvent, int64, error) {
	if a.store == nil {
		return nil, 0, nil
	}
	return a.store.List(ctx, filter, limit, offset)
}

// MarshalJSON marshals an audit event to JSON.
func (e *AuditEvent) MarshalJSON() ([]byte, error) {
	type Alias AuditEvent
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: e.Timestamp.Format(time.RFC3339),
		Alias:     (*Alias)(e),
	})
}

// Global audit logger instance.
var defaultAuditLogger *AuditLogger

// SetDefault sets the default audit logger.
func SetDefault(a *AuditLogger) {
	defaultAuditLogger = a
}

// Default returns the default audit logger.
func Default() *AuditLogger {
	return defaultAuditLogger
}

// Log logs an audit event using the default logger.
func Log(ctx context.Context, event *AuditEvent) error {
	if defaultAuditLogger == nil {
		return nil
	}
	return defaultAuditLogger.Log(ctx, event)
}
