// Package mcp provides MCP (Model Context Protocol) types and interfaces.
package mcp

import (
	"context"
	"time"

	"github.com/goplan/goplan/internal/domain/task"
)

// Intent Types
const (
	IntentCreatePlan  = "CREATE_PLAN"
	IntentAddTask     = "ADD_TASK"
	IntentUpdateTask  = "UPDATE_TASK"
	IntentSuggestTask = "SUGGEST_TASKS"
	IntentAskSummary  = "ASK_SUMMARY"
	IntentMoveTask    = "MOVE_TASK"
	IntentAssignTask  = "ASSIGN_TASK"
)

// IntentTypes is a list of all valid intent types.
var IntentTypes = []string{
	IntentCreatePlan,
	IntentAddTask,
	IntentUpdateTask,
	IntentSuggestTask,
	IntentAskSummary,
	IntentMoveTask,
	IntentAssignTask,
}

// Source Types
const (
	SourceChat = "chat"
	SourceUI   = "ui"
	SourceAPI  = "api"
)

// Actor Roles
const (
	ActorRolePlanner     = "planner"
	ActorRoleContributor = "contributor"
	ActorRoleObserver    = "observer"
)

// Confidence Thresholds
const (
	ConfidenceHighThreshold = 0.8
	ConfidenceLowThreshold  = 0.6
)

// MCPIntentEnvelope represents the complete intent envelope for AI requests.
type MCPIntentEnvelope struct {
	IntentID              string                 `json:"intentId"`
	IntentType            string                 `json:"intentType"`
	Confidence            float64                `json:"confidence"`
	NeedsClarification    bool                   `json:"needsClarification"`
	ClarificationQuestion *string                `json:"clarificationQuestion,omitempty"`
	Source                string                 `json:"source"`
	Actor                 MCPActor               `json:"actor"`
	Context               MCPContext             `json:"context"`
	Entities              map[string]interface{} `json:"entities"`
	ProposedActions       []MCPAction            `json:"proposedActions"`
	Timestamp             time.Time              `json:"timestamp"`
}

// MCPActor represents the actor performing the action.
type MCPActor struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// MCPContext represents the context for an MCP request.
type MCPContext struct {
	WorkspaceID   string                `json:"workspaceId"`
	PlanID        *string               `json:"planId,omitempty"`
	ExistingTasks []task.TaskSummary    `json:"existingTasks,omitempty"`
	HistoryWindow []ConversationMessage `json:"historyWindow,omitempty"`
}

// ConversationMessage represents a message in the conversation history.
type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MCPAction represents a proposed action to be executed.
type MCPAction struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPResponse represents the response from an MCP operation.
type MCPResponse struct {
	IntentID         string                 `json:"intentId"`
	Status           string                 `json:"status"`
	RequiresApproval bool                   `json:"requiresApproval"`
	Agent            string                 `json:"agent"`
	Result           map[string]interface{} `json:"result,omitempty"`
	Error            *MCPErrorResponse      `json:"error,omitempty"`
}

// MCPErrorResponse represents an error response from MCP.
type MCPErrorResponse struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ClarificationResponse represents a clarification request to the user.
type ClarificationResponse struct {
	IntentType            string                `json:"intentType"`
	Confidence            float64               `json:"confidence"`
	NeedsClarification    bool                  `json:"needsClarification"`
	ClarificationQuestion string                `json:"clarificationQuestion"`
	ClarificationOptions  []ClarificationOption `json:"clarificationOptions,omitempty"`
}

// ClarificationOption represents an option for clarification.
type ClarificationOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Draft represents a draft action pending approval.
type Draft struct {
	DraftID   string                 `json:"draftId"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	ExpiresAt time.Time              `json:"expiresAt"`
}

// DraftStatus constants
const (
	DraftStatusPendingApproval = "pending_approval"
	DraftStatusApproved        = "approved"
	DraftStatusRejected        = "rejected"
	DraftStatusExpired         = "expired"
)

// AuditRecord represents an audit log entry for MCP operations.
type AuditRecord struct {
	ID             string                 `json:"id"`
	Timestamp      time.Time              `json:"timestamp"`
	UserID         string                 `json:"userId"`
	WorkspaceID    string                 `json:"workspaceId"`
	IntentID       *string                `json:"intentId,omitempty"`
	IntentType     *string                `json:"intentType,omitempty"`
	IntentEnvelope *MCPIntentEnvelope     `json:"intentEnvelope,omitempty"`
	ActionTool     *string                `json:"actionTool,omitempty"`
	ActionArgs     map[string]interface{} `json:"actionArguments,omitempty"`
	Result         map[string]interface{} `json:"result,omitempty"`
	Status         string                 `json:"status"`
	ErrorMessage   *string                `json:"errorMessage,omitempty"`
}

// AuditRepository defines MCP audit log data access operations.
type AuditRepository interface {
	// Create persists an audit record to the database.
	Create(ctx context.Context, record *AuditRecord) error
}

// AuditStatus constants
const (
	AuditStatusSuccess  = "success"
	AuditStatusFailed   = "failed"
	AuditStatusRejected = "rejected"
	AuditStatusExpired  = "expired"
)

// TaskSuggestion represents a suggested task from AI.
type TaskSuggestion struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	Confidence  float64 `json:"confidence"`
	Reasoning   string  `json:"reasoning"`
}

// SuggestTasksResponse represents the response for task suggestions.
type SuggestTasksResponse struct {
	Suggestions []TaskSuggestion `json:"suggestions"`
}

// PlanSummary represents a summary of a plan.
type PlanSummary struct {
	Progress       float64  `json:"progress"`
	TasksCompleted int      `json:"tasksCompleted"`
	TasksPending   int      `json:"tasksPending"`
	Overdue        int      `json:"overdue"`
	Blockers       []string `json:"blockers"`
	Highlights     []string `json:"highlights"`
	Upcoming       []string `json:"upcoming"`
}

// SummaryResponse represents the response for a summary request.
type SummaryResponse struct {
	Summary PlanSummary `json:"summary"`
}

// Agent Types
const (
	AgentPlanner  = "PlannerAgent"
	AgentExecutor = "ExecutorAgent"
	AgentAnalyst  = "AnalystAgent"
)

// IsValidIntentType checks if the given intent type is valid.
func IsValidIntentType(intentType string) bool {
	for _, it := range IntentTypes {
		if it == intentType {
			return true
		}
	}
	return false
}

// RequiresHighConfidence returns true if the confidence is high enough for the intent.
func (e *MCPIntentEnvelope) RequiresHighConfidence() bool {
	return e.Confidence >= ConfidenceHighThreshold
}

// RequiresClarification returns true if confidence is too low.
func (e *MCPIntentEnvelope) RequiresClarification() bool {
	return e.Confidence < ConfidenceLowThreshold
}

// GetConfidenceBehavior returns the behavior based on confidence level.
func (e *MCPIntentEnvelope) GetConfidenceBehavior() string {
	if e.Confidence >= ConfidenceHighThreshold {
		return "proceed_with_confirmation"
	}
	if e.Confidence >= ConfidenceLowThreshold {
		return "proceed_with_uncertainty"
	}
	return "ask_clarification"
}
