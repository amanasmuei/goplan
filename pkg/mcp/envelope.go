// Package mcp provides public MCP types for external use.
package mcp

import (
	"encoding/json"
	"time"
)

// IntentEnvelope represents an MCP intent envelope for external consumers.
// This is the public API type that can be used by clients.
type IntentEnvelope struct {
	IntentID              string                 `json:"intentId"`
	IntentType            string                 `json:"intentType"`
	Confidence            float64                `json:"confidence"`
	NeedsClarification    bool                   `json:"needsClarification"`
	ClarificationQuestion *string                `json:"clarificationQuestion,omitempty"`
	Source                string                 `json:"source"`
	Actor                 Actor                  `json:"actor"`
	Context               Context                `json:"context"`
	Entities              map[string]interface{} `json:"entities"`
	ProposedActions       []Action               `json:"proposedActions"`
	Timestamp             time.Time              `json:"timestamp"`
}

// Actor represents the user performing the action.
type Actor struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// Context represents the context for an MCP request.
type Context struct {
	WorkspaceID   string        `json:"workspaceId"`
	PlanID        *string       `json:"planId,omitempty"`
	ExistingTasks []TaskSummary `json:"existingTasks,omitempty"`
}

// TaskSummary represents a summary of a task.
type TaskSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
}

// Action represents a proposed action.
type Action struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Response represents the response from an MCP operation.
type Response struct {
	IntentID         string                 `json:"intentId"`
	Status           string                 `json:"status"`
	RequiresApproval bool                   `json:"requiresApproval"`
	Agent            string                 `json:"agent,omitempty"`
	Result           map[string]interface{} `json:"result,omitempty"`
	Error            *ErrorResponse         `json:"error,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ClarificationRequest represents a request for clarification.
type ClarificationRequest struct {
	IntentType            string        `json:"intentType"`
	Confidence            float64       `json:"confidence"`
	NeedsClarification    bool          `json:"needsClarification"`
	ClarificationQuestion string        `json:"clarificationQuestion"`
	ClarificationOptions  []Option      `json:"clarificationOptions,omitempty"`
}

// Option represents a clarification option.
type Option struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// NewIntentEnvelope creates a new IntentEnvelope with default values.
func NewIntentEnvelope(intentID, intentType string, actor Actor, ctx Context) *IntentEnvelope {
	return &IntentEnvelope{
		IntentID:        intentID,
		IntentType:      intentType,
		Confidence:      0.0,
		Source:          "api",
		Actor:           actor,
		Context:         ctx,
		Entities:        make(map[string]interface{}),
		ProposedActions: []Action{},
		Timestamp:       time.Now().UTC(),
	}
}

// SetConfidence sets the confidence level.
func (e *IntentEnvelope) SetConfidence(confidence float64) *IntentEnvelope {
	e.Confidence = confidence
	return e
}

// SetSource sets the source.
func (e *IntentEnvelope) SetSource(source string) *IntentEnvelope {
	e.Source = source
	return e
}

// AddEntity adds an entity to the envelope.
func (e *IntentEnvelope) AddEntity(key string, value interface{}) *IntentEnvelope {
	e.Entities[key] = value
	return e
}

// AddAction adds a proposed action.
func (e *IntentEnvelope) AddAction(tool string, args map[string]interface{}) *IntentEnvelope {
	e.ProposedActions = append(e.ProposedActions, Action{
		Tool:      tool,
		Arguments: args,
	})
	return e
}

// RequestClarification marks the envelope as needing clarification.
func (e *IntentEnvelope) RequestClarification(question string) *IntentEnvelope {
	e.NeedsClarification = true
	e.ClarificationQuestion = &question
	return e
}

// ToJSON serializes the envelope to JSON.
func (e *IntentEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes an envelope from JSON.
func FromJSON(data []byte) (*IntentEnvelope, error) {
	var envelope IntentEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

// ValidateIntentType checks if the intent type is valid.
func ValidateIntentType(intentType string) bool {
	validTypes := map[string]bool{
		"CREATE_PLAN":   true,
		"ADD_TASK":      true,
		"UPDATE_TASK":   true,
		"SUGGEST_TASKS": true,
		"ASK_SUMMARY":   true,
		"MOVE_TASK":     true,
		"ASSIGN_TASK":   true,
	}
	return validTypes[intentType]
}

// ConfidenceLevel returns a human-readable confidence level.
func (e *IntentEnvelope) ConfidenceLevel() string {
	switch {
	case e.Confidence >= 0.8:
		return "high"
	case e.Confidence >= 0.6:
		return "medium"
	default:
		return "low"
	}
}

// ShouldProceed returns true if confidence is high enough to proceed.
func (e *IntentEnvelope) ShouldProceed() bool {
	return e.Confidence >= 0.6
}

// ShouldAskClarification returns true if clarification is needed.
func (e *IntentEnvelope) ShouldAskClarification() bool {
	return e.NeedsClarification || e.Confidence < 0.6
}
