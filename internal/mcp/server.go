// Package mcp provides the MCP HTTP server implementation.
package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/shared"
)

// Server handles MCP HTTP requests.
type Server struct {
	registry  *ToolRegistry
	auditRepo AuditRepository
	// In-memory audit log kept as fallback when auditRepo is nil
	auditLog []AuditRecord
}

// NewServer creates a new MCP server.
func NewServer(registry *ToolRegistry, auditRepo AuditRepository) *Server {
	return &Server{
		registry:  registry,
		auditRepo: auditRepo,
		auditLog:  make([]AuditRecord, 0),
	}
}

// HandleIntent handles incoming intent requests.
func (s *Server) HandleIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Validate Content-Type
	if ct := r.Header.Get("Content-Type"); ct != "" {
		if !strings.HasPrefix(ct, "application/json") {
			http.Error(w, `{"error":"Content-Type must be application/json"}`, http.StatusUnsupportedMediaType)
			return
		}
	}

	var envelope MCPIntentEnvelope
	if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request body")
		return
	}

	// Validate the intent
	if err := s.validateIntent(&envelope); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_INTENT", err.Error())
		return
	}

	// Get execution context from headers
	execCtx := s.getExecutionContext(r)

	// Record audit
	s.recordAudit(execCtx, &envelope, nil, nil, "received")

	// Route to appropriate agent
	agent := s.routeIntent(envelope.IntentType)

	// Check confidence level
	behavior := envelope.GetConfidenceBehavior()

	response := MCPResponse{
		IntentID:         envelope.IntentID,
		Status:           "intent_received",
		RequiresApproval: true,
		Agent:            agent,
		Result: map[string]interface{}{
			"confidenceBehavior": behavior,
			"proposedActions":    envelope.ProposedActions,
		},
	}

	// If clarification needed, add clarification info
	if envelope.NeedsClarification || envelope.RequiresClarification() {
		response.Status = "needs_clarification"
		if envelope.ClarificationQuestion != nil {
			response.Result["clarificationQuestion"] = *envelope.ClarificationQuestion
		}
	}

	s.writeJSON(w, http.StatusOK, response)
}

// HandleToolExecute handles tool execution requests.
func (s *Server) HandleToolExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Validate Content-Type
	if ct := r.Header.Get("Content-Type"); ct != "" {
		if !strings.HasPrefix(ct, "application/json") {
			http.Error(w, `{"error":"Content-Type must be application/json"}`, http.StatusUnsupportedMediaType)
			return
		}
	}

	var action MCPAction
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request body")
		return
	}

	if action.Tool == "" {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tool name is required")
		return
	}

	// Get execution context from headers
	execCtx := s.getExecutionContext(r)

	// Execute the tool
	ctx := r.Context()
	result, err := s.registry.ExecuteTool(ctx, action.Tool, execCtx, action.Arguments)

	if err != nil {
		// Record audit
		s.recordAudit(execCtx, nil, &action, nil, AuditStatusFailed)

		// Check error type
		var domainErr *shared.DomainError
		if errors.As(err, &domainErr) {
			s.writeError(w, s.domainErrorToStatus(domainErr), domainErr.Code, domainErr.Message)
			return
		}

		s.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
		return
	}

	// Record audit
	resultMap, _ := result.(map[string]interface{})
	s.recordAudit(execCtx, nil, &action, resultMap, AuditStatusSuccess)

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "executed",
		"result": result,
	})
}

// HandleListTools lists all available tools.
func (s *Server) HandleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.registry.ListTools()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tools": tools,
	})
}

// HandleApprove handles draft approval requests.
func (s *Server) HandleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Validate Content-Type
	if ct := r.Header.Get("Content-Type"); ct != "" {
		if !strings.HasPrefix(ct, "application/json") {
			http.Error(w, `{"error":"Content-Type must be application/json"}`, http.StatusUnsupportedMediaType)
			return
		}
	}

	var req struct {
		IntentID string                 `json:"intentId"`
		Action   MCPAction              `json:"action"`
		Edits    map[string]interface{} `json:"edits,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request body")
		return
	}

	// Merge edits into action arguments if provided
	if req.Edits != nil {
		for k, v := range req.Edits {
			req.Action.Arguments[k] = v
		}
	}

	// Get execution context
	execCtx := s.getExecutionContext(r)

	// Execute the approved action
	ctx := r.Context()
	result, err := s.registry.ExecuteTool(ctx, req.Action.Tool, execCtx, req.Action.Arguments)

	if err != nil {
		s.recordAudit(execCtx, nil, &req.Action, nil, AuditStatusFailed)

		var domainErr *shared.DomainError
		if errors.As(err, &domainErr) {
			s.writeError(w, s.domainErrorToStatus(domainErr), domainErr.Code, domainErr.Message)
			return
		}

		s.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred")
		return
	}

	resultMap, _ := result.(map[string]interface{})
	s.recordAudit(execCtx, nil, &req.Action, resultMap, AuditStatusSuccess)

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "approved",
		"intentId": req.IntentID,
		"result":   result,
	})
}

// HandleReject handles draft rejection requests.
func (s *Server) HandleReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// Validate Content-Type
	if ct := r.Header.Get("Content-Type"); ct != "" {
		if !strings.HasPrefix(ct, "application/json") {
			http.Error(w, `{"error":"Content-Type must be application/json"}`, http.StatusUnsupportedMediaType)
			return
		}
	}

	var req struct {
		IntentID string `json:"intentId"`
		Reason   string `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to decode request body")
		return
	}

	execCtx := s.getExecutionContext(r)
	s.recordAudit(execCtx, nil, nil, map[string]interface{}{"reason": req.Reason}, AuditStatusRejected)

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "rejected",
		"intentId": req.IntentID,
	})
}

// Routes returns the HTTP handler for all MCP routes.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp/intent", s.HandleIntent)
	mux.HandleFunc("/mcp/tool/execute", s.HandleToolExecute)
	mux.HandleFunc("/mcp/tools", s.HandleListTools)
	mux.HandleFunc("/mcp/intent/approve", s.HandleApprove)
	mux.HandleFunc("/mcp/intent/reject", s.HandleReject)
	return mux
}

// validateIntent validates the intent envelope.
func (s *Server) validateIntent(env *MCPIntentEnvelope) error {
	if !IsValidIntentType(env.IntentType) {
		return fmt.Errorf("invalid intent type: %s", env.IntentType)
	}

	if env.Confidence < 0 || env.Confidence > 1 {
		return errors.New("confidence must be between 0 and 1")
	}

	if env.Actor.UserID == "" {
		return errors.New("actor userId is required")
	}

	if env.Context.WorkspaceID == "" {
		return errors.New("context workspaceId is required")
	}

	return nil
}

// routeIntent routes the intent to the appropriate agent.
func (s *Server) routeIntent(intentType string) string {
	switch intentType {
	case IntentCreatePlan, IntentSuggestTask:
		return AgentPlanner
	case IntentAddTask, IntentUpdateTask, IntentMoveTask, IntentAssignTask:
		return AgentExecutor
	case IntentAskSummary:
		return AgentAnalyst
	default:
		return AgentAnalyst
	}
}

// getExecutionContext extracts execution context from the authenticated user
// stored in the request context by the auth middleware. Falls back to raw
// HTTP headers only when no auth context is available (e.g. dev mode without JWT).
func (s *Server) getExecutionContext(r *http.Request) ExecutionContext {
	// First, try to extract from the authenticated context set by auth middleware
	if user := auth.GetUser(r.Context()); user != nil {
		return ExecutionContext{
			UserID:      user.ID,
			WorkspaceID: user.WorkspaceID,
			Role:        user.Role,
		}
	}

	// Fall back to headers only if auth context is not available
	return ExecutionContext{
		UserID:      r.Header.Get("X-User-ID"),
		WorkspaceID: r.Header.Get("X-Workspace-ID"),
		Role:        r.Header.Get("X-User-Role"),
	}
}

// recordAudit records an audit entry.
func (s *Server) recordAudit(execCtx ExecutionContext, intent *MCPIntentEnvelope, action *MCPAction, result map[string]interface{}, status string) {
	record := AuditRecord{
		ID:          generateID(),
		Timestamp:   time.Now().UTC(),
		UserID:      execCtx.UserID,
		WorkspaceID: execCtx.WorkspaceID,
		Result:      result,
		Status:      status,
	}

	if intent != nil {
		record.IntentID = &intent.IntentID
		record.IntentType = &intent.IntentType
		record.IntentEnvelope = intent
	}

	if action != nil {
		record.ActionTool = &action.Tool
		record.ActionArgs = action.Arguments
	}

	// Persist to database if audit repository is available
	if s.auditRepo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.auditRepo.Create(ctx, &record); err != nil {
			log.Printf("failed to persist audit record: %v", err)
		}
	}

	// Only keep in-memory copy when no persistent audit repo is available
	if s.auditRepo == nil {
		s.auditLog = append(s.auditLog, record)
		// Cap in-memory audit log at 10000 entries to prevent unbounded growth
		if len(s.auditLog) > 10000 {
			s.auditLog = s.auditLog[len(s.auditLog)-10000:]
		}
	}
}

// writeJSON writes a JSON response.
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func (s *Server) writeError(w http.ResponseWriter, status int, code, message string) {
	s.writeJSON(w, status, MCPErrorResponse{
		Error:   true,
		Code:    code,
		Message: message,
	})
}

// domainErrorToStatus converts a domain error to an HTTP status code.
func (s *Server) domainErrorToStatus(err *shared.DomainError) int {
	switch {
	case errors.Is(err.Err, shared.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err.Err, shared.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err.Err, shared.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err.Err, shared.ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err.Err, shared.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err.Err, shared.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// generateID generates a unique ID using UUID v4.
func generateID() string {
	return uuid.New().String()
}

// WithRegistry returns a handler function that injects the registry.
func WithRegistry(registry *ToolRegistry, handler func(context.Context, *ToolRegistry, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(r.Context(), registry, w, r)
	}
}
