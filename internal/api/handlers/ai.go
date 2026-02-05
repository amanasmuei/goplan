// Package handlers provides HTTP handlers for the REST API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/api/middleware"
	"github.com/goplan/goplan/internal/claude"
)

// AIHandler handles AI-related HTTP requests.
type AIHandler struct {
	*BaseHandler
	aiService *claude.Service
}

// NewAIHandler creates a new AIHandler.
func NewAIHandler(aiService *claude.Service) *AIHandler {
	return &AIHandler{
		BaseHandler: NewBaseHandler(),
		aiService:   aiService,
	}
}

// ServeHTTP routes AI requests to the appropriate handler.
func (h *AIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")

	switch {
	// POST /api/v1/ai/chat - Chat with AI
	case path == "/api/v1/ai/chat" && r.Method == http.MethodPost:
		h.handleChat(w, r)

	// POST /api/v1/ai/plans/generate - Generate plan from description
	case path == "/api/v1/ai/plans/generate" && r.Method == http.MethodPost:
		h.handleGeneratePlan(w, r)

	// POST /api/v1/ai/tasks/:id/breakdown - Break down task
	case strings.HasPrefix(path, "/api/v1/ai/tasks/") && strings.HasSuffix(path, "/breakdown") && r.Method == http.MethodPost:
		h.handleBreakdownTask(w, r)

	// GET /api/v1/ai/suggestions/:planId - Get AI suggestions
	case strings.HasPrefix(path, "/api/v1/ai/suggestions/") && r.Method == http.MethodGet:
		h.handleGetSuggestions(w, r)

	// POST /api/v1/ai/approve/:requestId - Approve pending AI action
	case strings.HasPrefix(path, "/api/v1/ai/approve/") && r.Method == http.MethodPost:
		h.handleApproveAction(w, r)

	// POST /api/v1/ai/reject/:requestId - Reject pending AI action
	case strings.HasPrefix(path, "/api/v1/ai/reject/") && r.Method == http.MethodPost:
		h.handleRejectAction(w, r)

	// GET /api/v1/ai/approvals - Get pending approvals for user
	case path == "/api/v1/ai/approvals" && r.Method == http.MethodGet:
		h.handleGetPendingApprovals(w, r)

	// POST /api/v1/ai/execute/:requestId - Execute approved action
	case strings.HasPrefix(path, "/api/v1/ai/execute/") && r.Method == http.MethodPost:
		h.handleExecuteApproved(w, r)

	default:
		h.WriteNotFound(w, "endpoint")
	}
}

// ChatRequest represents the request body for the chat endpoint.
type ChatRequest struct {
	Message   string  `json:"message"`
	SessionID *string `json:"sessionId,omitempty"`
	PlanID    *string `json:"planId,omitempty"`
}

// handleChat handles POST /api/v1/ai/chat
func (h *AIHandler) handleChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := middleware.GetAuthContext(ctx)

	// Parse request body
	var req ChatRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	// Validate request
	if req.Message == "" {
		h.WriteBadRequest(w, "message is required")
		return
	}

	// Build chat request
	chatReq := &claude.ChatRequest{
		UserID:      auth.UserID,
		WorkspaceID: auth.WorkspaceID,
		Message:     req.Message,
	}

	if req.SessionID != nil {
		chatReq.SessionID = *req.SessionID
	}

	if req.PlanID != nil {
		chatReq.PlanID = req.PlanID
	}

	// Call AI service
	response, err := h.aiService.Chat(ctx, chatReq)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, response)
}

// GeneratePlanRequest represents the request body for plan generation.
type GeneratePlanRequest struct {
	Description string  `json:"description"`
	Domain      *string `json:"domain,omitempty"`
}

// handleGeneratePlan handles POST /api/v1/ai/plans/generate
func (h *AIHandler) handleGeneratePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := middleware.GetAuthContext(ctx)

	// Parse request body
	var req GeneratePlanRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	// Validate request
	if req.Description == "" {
		h.WriteBadRequest(w, "description is required")
		return
	}

	// Build service request
	genReq := &claude.GeneratePlanRequest{
		UserID:      auth.UserID,
		WorkspaceID: auth.WorkspaceID,
		Description: req.Description,
	}

	if req.Domain != nil {
		genReq.Domain = *req.Domain
	}

	// Call AI service
	response, err := h.aiService.CreatePlanFromDescription(ctx, genReq)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteCreated(w, response)
}

// handleBreakdownTask handles POST /api/v1/ai/tasks/:id/breakdown
func (h *AIHandler) handleBreakdownTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := middleware.GetAuthContext(ctx)

	// Extract task ID from path
	path := r.URL.Path
	// Path format: /api/v1/ai/tasks/{taskId}/breakdown
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		h.WriteBadRequest(w, "invalid path")
		return
	}
	taskID := parts[5] // Index 5 is the task ID

	if taskID == "" {
		h.WriteBadRequest(w, "task ID is required")
		return
	}

	// Build service request
	breakdownReq := &claude.BreakdownTaskRequest{
		UserID:      auth.UserID,
		WorkspaceID: auth.WorkspaceID,
		TaskID:      taskID,
	}

	// Call AI service
	response, err := h.aiService.BreakdownTask(ctx, breakdownReq)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, response)
}

// handleGetSuggestions handles GET /api/v1/ai/suggestions/:planId
func (h *AIHandler) handleGetSuggestions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := middleware.GetAuthContext(ctx)

	// Extract plan ID from path
	path := r.URL.Path
	planID := strings.TrimPrefix(path, "/api/v1/ai/suggestions/")

	if planID == "" {
		h.WriteBadRequest(w, "plan ID is required")
		return
	}

	// Build service request
	suggestReq := &claude.SuggestNextStepsRequest{
		UserID:      auth.UserID,
		WorkspaceID: auth.WorkspaceID,
		PlanID:      planID,
	}

	// Call AI service
	response, err := h.aiService.SuggestNextSteps(ctx, suggestReq)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, response)
}

// ApproveActionRequest represents the request body for approval.
type ApproveActionRequest struct {
	// No additional fields needed - user info comes from auth context
}

// handleApproveAction handles POST /api/v1/ai/approve/:requestId
func (h *AIHandler) handleApproveAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := middleware.GetAuthContext(ctx)

	// Extract request ID from path
	path := r.URL.Path
	requestID := strings.TrimPrefix(path, "/api/v1/ai/approve/")

	if requestID == "" {
		h.WriteBadRequest(w, "request ID is required")
		return
	}

	// Approve the action
	approval, err := h.aiService.ApproveAction(ctx, requestID, auth.UserID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, approval)
}

// RejectActionRequest represents the request body for rejection.
type RejectActionRequest struct {
	Note string `json:"note,omitempty"`
}

// handleRejectAction handles POST /api/v1/ai/reject/:requestId
func (h *AIHandler) handleRejectAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	auth := middleware.GetAuthContext(ctx)

	// Extract request ID from path
	path := r.URL.Path
	requestID := strings.TrimPrefix(path, "/api/v1/ai/reject/")

	if requestID == "" {
		h.WriteBadRequest(w, "request ID is required")
		return
	}

	// Parse request body (optional)
	var req RejectActionRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Reject the action
	rejection, err := h.aiService.RejectAction(ctx, requestID, auth.UserID, req.Note)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, rejection)
}

// handleGetPendingApprovals handles GET /api/v1/ai/approvals
func (h *AIHandler) handleGetPendingApprovals(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r.Context())

	// Get pending approvals for the user
	approvals := h.aiService.GetPendingApprovals(auth.UserID)

	h.WriteSuccess(w, map[string]interface{}{
		"approvals": approvals,
		"count":     len(approvals),
	})
}

// handleExecuteApproved handles POST /api/v1/ai/execute/:requestId
func (h *AIHandler) handleExecuteApproved(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract request ID from path
	path := r.URL.Path
	requestID := strings.TrimPrefix(path, "/api/v1/ai/execute/")

	if requestID == "" {
		h.WriteBadRequest(w, "request ID is required")
		return
	}

	// Execute the approved action
	result, err := h.aiService.ExecuteApprovedAction(ctx, requestID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, map[string]interface{}{
		"executed":  true,
		"requestId": requestID,
		"result":    result,
	})
}

// AIRoutes returns a handler function that registers AI routes.
func AIRoutes(aiHandler *AIHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		aiHandler.ServeHTTP(w, r)
	})
}

// ValidateAIConfig validates that AI configuration is properly set up.
func ValidateAIConfig(apiKey string) error {
	if apiKey == "" {
		return &aiConfigError{message: "Claude API key is not configured"}
	}
	return nil
}

// aiConfigError represents an AI configuration error.
type aiConfigError struct {
	message string
}

func (e *aiConfigError) Error() string {
	return e.message
}
