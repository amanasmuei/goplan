// Package claude provides high-level AI service for GoPlan.
package claude

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/repository"
)

// Service provides high-level AI functionality for the application.
type Service struct {
	client        *Client
	bridge        *Bridge
	safetyChecker *SafetyChecker
	registry      *mcp.ToolRegistry

	// Repositories for context injection
	workspaceRepo repository.WorkspaceRepository
	planRepo      repository.PlanRepository
	taskRepo      repository.TaskRepository

	// Conversation management
	conversations map[string]*Conversation
	convMu        sync.RWMutex
}

// ServiceConfig holds configuration for the AI service.
type ServiceConfig struct {
	Client        *Client
	Registry      *mcp.ToolRegistry
	SafetyChecker *SafetyChecker
	WorkspaceRepo repository.WorkspaceRepository
	PlanRepo      repository.PlanRepository
	TaskRepo      repository.TaskRepository
}

// NewService creates a new AI service.
func NewService(cfg ServiceConfig) *Service {
	bridge := NewBridge(cfg.Registry, cfg.SafetyChecker)

	return &Service{
		client:        cfg.Client,
		bridge:        bridge,
		safetyChecker: cfg.SafetyChecker,
		registry:      cfg.Registry,
		workspaceRepo: cfg.WorkspaceRepo,
		planRepo:      cfg.PlanRepo,
		taskRepo:      cfg.TaskRepo,
		conversations: make(map[string]*Conversation),
	}
}

// ChatRequest represents a chat request.
type ChatRequest struct {
	UserID      string  `json:"userId"`
	WorkspaceID string  `json:"workspaceId"`
	PlanID      *string `json:"planId,omitempty"`
	Message     string  `json:"message"`
	SessionID   string  `json:"sessionId,omitempty"`
}

// ChatResponse represents a chat response.
type ChatResponse struct {
	Message          string             `json:"message"`
	ToolsUsed        []string           `json:"toolsUsed,omitempty"`
	PendingApprovals []*ApprovalRequest `json:"pendingApprovals,omitempty"`
	SessionID        string             `json:"sessionId"`
}

// Chat processes a chat message with AI and returns a response.
func (s *Service) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Build execution context
	execCtx := mcp.ExecutionContext{
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		PlanID:      req.PlanID,
		Role:        "member", // Default role
	}

	// Get or create conversation
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	conversation := s.getOrCreateConversation(sessionID, execCtx)

	// Build system prompt with context
	systemPrompt := s.buildSystemPrompt(ctx, execCtx)
	conversation.system = systemPrompt

	// Add user message
	conversation.AddUserMessage(req.Message)

	// Process the conversation turn
	result, err := conversation.ProcessTurnWithContinuation(ctx, execCtx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to process chat: %w", err)
	}

	// Build response
	response := &ChatResponse{
		Message:          result.TextContent,
		SessionID:        sessionID,
		PendingApprovals: result.PendingApprovals,
	}

	// Collect tools used
	if result.ToolResults != nil {
		for _, tr := range result.ToolResults {
			response.ToolsUsed = append(response.ToolsUsed, tr.ToolName)
		}
	}

	log.Printf("Chat completed: session=%s, tools_used=%v, pending_approvals=%d",
		sessionID, response.ToolsUsed, len(response.PendingApprovals))

	return response, nil
}

// GeneratePlanRequest represents a request to generate a plan from description.
type GeneratePlanRequest struct {
	UserID      string `json:"userId"`
	WorkspaceID string `json:"workspaceId"`
	Description string `json:"description"`
	Domain      string `json:"domain,omitempty"`
}

// GeneratePlanResponse represents the response from plan generation.
type GeneratePlanResponse struct {
	PlanID           string             `json:"planId,omitempty"`
	PlanName         string             `json:"planName,omitempty"`
	TasksCreated     int                `json:"tasksCreated"`
	MilestonesCreated int               `json:"milestonesCreated"`
	Message          string             `json:"message"`
	PendingApprovals []*ApprovalRequest `json:"pendingApprovals,omitempty"`
}

// CreatePlanFromDescription uses AI to create a plan from a natural language description.
func (s *Service) CreatePlanFromDescription(ctx context.Context, req *GeneratePlanRequest) (*GeneratePlanResponse, error) {
	execCtx := mcp.ExecutionContext{
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		Role:        "member",
	}

	domain := req.Domain
	if domain == "" {
		domain = "generic"
	}

	// Build specialized prompt for plan creation
	prompt := fmt.Sprintf(`You are a project planning assistant. The user wants to create a new project plan.

User's description: %s

Based on this description:
1. First, create a plan using the plan.create tool with an appropriate name, description, and domain (%s).
2. Then, create relevant tasks using the task.create tool. Break down the project into actionable tasks with clear titles and descriptions.
3. If applicable, create milestones using the milestone.create tool for major deliverables or checkpoints.

Be thorough but practical. Create tasks that are specific and actionable. Set appropriate priorities based on the project goals.`, req.Description, domain)

	// Create a temporary conversation for this request
	conversation := NewConversation(s.client, s.bridge, s.buildSystemPrompt(ctx, execCtx))
	conversation.AddUserMessage(prompt)

	// Process with continuation to handle multiple tool calls
	result, err := conversation.ProcessTurnWithContinuation(ctx, execCtx, 15)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Build response
	response := &GeneratePlanResponse{
		Message:          result.TextContent,
		PendingApprovals: result.PendingApprovals,
	}

	// Count created entities
	for _, tr := range result.ToolResults {
		if tr.Error != nil || tr.RequiresApproval {
			continue
		}

		switch tr.ToolName {
		case "plan.create":
			if resultMap, ok := tr.Result.(map[string]interface{}); ok {
				if planID, ok := resultMap["id"].(string); ok {
					response.PlanID = planID
				}
				if planName, ok := resultMap["name"].(string); ok {
					response.PlanName = planName
				}
			}
		case "task.create":
			response.TasksCreated++
		case "milestone.create":
			response.MilestonesCreated++
		}
	}

	log.Printf("Plan generation completed: plan_id=%s, tasks=%d, milestones=%d",
		response.PlanID, response.TasksCreated, response.MilestonesCreated)

	return response, nil
}

// BreakdownTaskRequest represents a request to break down a task.
type BreakdownTaskRequest struct {
	UserID      string `json:"userId"`
	WorkspaceID string `json:"workspaceId"`
	TaskID      string `json:"taskId"`
}

// BreakdownTaskResponse represents the response from task breakdown.
type BreakdownTaskResponse struct {
	SubtasksCreated  int                `json:"subtasksCreated"`
	SubtaskIDs       []string           `json:"subtaskIds"`
	Message          string             `json:"message"`
	PendingApprovals []*ApprovalRequest `json:"pendingApprovals,omitempty"`
}

// BreakdownTask uses AI to break down a complex task into subtasks.
func (s *Service) BreakdownTask(ctx context.Context, req *BreakdownTaskRequest) (*BreakdownTaskResponse, error) {
	execCtx := mcp.ExecutionContext{
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		Role:        "member",
	}

	// First, get the task details
	prompt := fmt.Sprintf(`You are a project management assistant helping to break down a complex task into smaller, actionable subtasks.

First, retrieve the task details using task.get with taskId "%s".

Then, based on the task's title and description, break it down into 3-7 smaller, specific subtasks using task.create. Each subtask should:
- Have a clear, actionable title
- Be completable in a reasonable timeframe
- Use the same planId as the parent task
- Set the parentId to "%s" to make them subtasks
- Have appropriate priorities

Provide a summary of what you created.`, req.TaskID, req.TaskID)

	// Create conversation
	conversation := NewConversation(s.client, s.bridge, s.buildSystemPrompt(ctx, execCtx))
	conversation.AddUserMessage(prompt)

	// Process with continuation
	result, err := conversation.ProcessTurnWithContinuation(ctx, execCtx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to break down task: %w", err)
	}

	// Build response
	response := &BreakdownTaskResponse{
		Message:          result.TextContent,
		PendingApprovals: result.PendingApprovals,
		SubtaskIDs:       []string{},
	}

	// Count created subtasks
	for _, tr := range result.ToolResults {
		if tr.Error != nil || tr.RequiresApproval {
			continue
		}

		if tr.ToolName == "task.create" {
			response.SubtasksCreated++
			if resultMap, ok := tr.Result.(map[string]interface{}); ok {
				if taskID, ok := resultMap["id"].(string); ok {
					response.SubtaskIDs = append(response.SubtaskIDs, taskID)
				}
			}
		}
	}

	log.Printf("Task breakdown completed: task_id=%s, subtasks=%d", req.TaskID, response.SubtasksCreated)

	return response, nil
}

// SuggestNextStepsRequest represents a request for AI suggestions.
type SuggestNextStepsRequest struct {
	UserID      string `json:"userId"`
	WorkspaceID string `json:"workspaceId"`
	PlanID      string `json:"planId"`
}

// SuggestNextStepsResponse represents the response with AI suggestions.
type SuggestNextStepsResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Summary     string       `json:"summary"`
}

// Suggestion represents a single AI suggestion.
type Suggestion struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Reasoning   string `json:"reasoning"`
}

// SuggestNextSteps uses AI to analyze a plan and suggest next steps.
func (s *Service) SuggestNextSteps(ctx context.Context, req *SuggestNextStepsRequest) (*SuggestNextStepsResponse, error) {
	execCtx := mcp.ExecutionContext{
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		PlanID:      &req.PlanID,
		Role:        "member",
	}

	prompt := fmt.Sprintf(`You are a project management advisor analyzing a plan to suggest next steps.

Please:
1. Use plan.get to retrieve the plan details for plan ID "%s"
2. Use task.list to get the current tasks in the plan
3. Optionally use activity.get to see recent activity

Then provide recommendations in this JSON format:
{
  "suggestions": [
    {
      "type": "task|priority|milestone|process",
      "title": "Short title",
      "description": "What to do",
      "priority": "high|medium|low",
      "reasoning": "Why this is important"
    }
  ],
  "summary": "Overall assessment of the plan's current state"
}

Focus on:
- Tasks that might be blocked or overdue
- Missing tasks that seem necessary
- Priority adjustments needed
- Upcoming milestones and preparation
- Process improvements

Provide 3-5 actionable suggestions.`, req.PlanID)

	// Create conversation
	conversation := NewConversation(s.client, s.bridge, s.buildSystemPrompt(ctx, execCtx))
	conversation.AddUserMessage(prompt)

	// Process conversation
	result, err := conversation.ProcessTurnWithContinuation(ctx, execCtx, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	// Parse the AI response for structured suggestions
	response := &SuggestNextStepsResponse{
		Summary:     result.TextContent,
		Suggestions: []Suggestion{},
	}

	// Try to extract JSON from the response
	if jsonStr := extractJSON(result.TextContent); jsonStr != "" {
		var parsed struct {
			Suggestions []Suggestion `json:"suggestions"`
			Summary     string       `json:"summary"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
			response.Suggestions = parsed.Suggestions
			if parsed.Summary != "" {
				response.Summary = parsed.Summary
			}
		}
	}

	log.Printf("Suggestions generated: plan_id=%s, suggestions=%d", req.PlanID, len(response.Suggestions))

	return response, nil
}

// ApproveAction approves a pending AI action.
func (s *Service) ApproveAction(ctx context.Context, requestID, approverID string) (*ApprovalRequest, error) {
	return s.safetyChecker.ApproveRequest(ctx, requestID, approverID)
}

// RejectAction rejects a pending AI action.
func (s *Service) RejectAction(ctx context.Context, requestID, rejecterID, note string) (*ApprovalRequest, error) {
	return s.safetyChecker.RejectRequest(ctx, requestID, rejecterID, note)
}

// GetPendingApprovals gets all pending approvals for a user.
func (s *Service) GetPendingApprovals(userID string) []*ApprovalRequest {
	return s.safetyChecker.GetPendingRequestsForUser(userID)
}

// ExecuteApprovedAction executes an approved action.
func (s *Service) ExecuteApprovedAction(ctx context.Context, requestID string) (interface{}, error) {
	// Get the approved request
	request, err := s.safetyChecker.GetPendingRequest(requestID)
	if err != nil {
		return nil, err
	}

	if request.Status != ApprovalStatusApproved {
		return nil, fmt.Errorf("request is not approved: %s", request.Status)
	}

	// Execute the tool
	execCtx := mcp.ExecutionContext{
		UserID:      request.UserID,
		WorkspaceID: request.WorkspaceID,
		Role:        "member",
	}

	result, err := s.registry.ExecuteTool(ctx, request.ToolName, execCtx, request.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to execute approved action: %w", err)
	}

	log.Printf("Approved action executed: request_id=%s, tool=%s", requestID, request.ToolName)

	return result, nil
}

// buildSystemPrompt builds the system prompt with context.
func (s *Service) buildSystemPrompt(ctx context.Context, execCtx mcp.ExecutionContext) string {
	basePrompt := `You are GoPlan AI, an intelligent project management assistant. You help users manage their projects, tasks, and milestones efficiently.

Your capabilities include:
- Creating and managing plans, tasks, and milestones
- Searching and listing project information
- Moving tasks between statuses
- Providing project insights and suggestions

Guidelines:
- Be concise and helpful
- When creating tasks, use clear, actionable titles
- Consider dependencies and priorities when making suggestions
- If unsure about something, ask for clarification
- Always confirm destructive actions with the user

`

	// Add workspace context if available
	if s.workspaceRepo != nil && execCtx.WorkspaceID != "" {
		workspace, err := s.workspaceRepo.GetByID(ctx, execCtx.WorkspaceID)
		if err == nil && workspace != nil {
			basePrompt += fmt.Sprintf("Current Workspace: %s\n", workspace.Name)
		}
	}

	// Add plan context if available
	if s.planRepo != nil && execCtx.PlanID != nil && *execCtx.PlanID != "" {
		plan, err := s.planRepo.GetByID(ctx, *execCtx.PlanID)
		if err == nil && plan != nil {
			basePrompt += fmt.Sprintf("Current Plan: %s (Status: %s, Domain: %s)\n", plan.Name, plan.Status, plan.Domain)
		}
	}

	basePrompt += "\nCurrent time: " + time.Now().Format(time.RFC3339)

	return basePrompt
}

// getOrCreateConversation gets or creates a conversation for a session.
func (s *Service) getOrCreateConversation(sessionID string, execCtx mcp.ExecutionContext) *Conversation {
	s.convMu.Lock()
	defer s.convMu.Unlock()

	if conv, ok := s.conversations[sessionID]; ok {
		return conv
	}

	conv := NewConversation(s.client, s.bridge, "")
	s.conversations[sessionID] = conv
	return conv
}

// ClearSession clears a conversation session.
func (s *Service) ClearSession(sessionID string) {
	s.convMu.Lock()
	defer s.convMu.Unlock()
	delete(s.conversations, sessionID)
}

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return "sess_" + hex.EncodeToString(bytes)
}

// extractJSON extracts JSON from a text that might contain other content.
func extractJSON(text string) string {
	// Find the first { and last }
	start := -1
	end := -1
	depth := 0

	for i, c := range text {
		if c == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 && start != -1 {
				end = i + 1
				break
			}
		}
	}

	if start != -1 && end != -1 {
		return text[start:end]
	}

	return ""
}

// StartBackgroundWorkers starts background workers for the service.
func (s *Service) StartBackgroundWorkers(ctx context.Context) {
	// Start approval cleanup worker
	s.safetyChecker.StartCleanupWorker(ctx, time.Minute)

	// Start conversation cleanup worker
	go s.cleanupOldConversations(ctx)
}

// cleanupOldConversations removes old conversation sessions.
func (s *Service) cleanupOldConversations(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// For now, just clear all conversations older than 1 hour
			// In a production system, you'd track timestamps per conversation
			s.convMu.Lock()
			// Clear all - in production, you'd be more selective
			if len(s.conversations) > 100 {
				s.conversations = make(map[string]*Conversation)
			}
			s.convMu.Unlock()
		}
	}
}
