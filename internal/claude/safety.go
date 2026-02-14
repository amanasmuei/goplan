// Package claude provides human-in-the-loop safety mechanisms for AI operations.
package claude

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/goplan/goplan/internal/mcp"
)

// ApprovalStatus represents the status of an approval request.
type ApprovalStatus string

const (
	// ApprovalStatusPending indicates the request is awaiting approval.
	ApprovalStatusPending ApprovalStatus = "pending"
	// ApprovalStatusApproved indicates the request was approved.
	ApprovalStatusApproved ApprovalStatus = "approved"
	// ApprovalStatusRejected indicates the request was rejected.
	ApprovalStatusRejected ApprovalStatus = "rejected"
	// ApprovalStatusExpired indicates the request has expired.
	ApprovalStatusExpired ApprovalStatus = "expired"
)

// ApprovalCategory represents the category of action requiring approval.
type ApprovalCategory string

const (
	// CategoryDelete indicates a delete operation.
	CategoryDelete ApprovalCategory = "delete"
	// CategoryBulkUpdate indicates a bulk update operation.
	CategoryBulkUpdate ApprovalCategory = "bulk_update"
	// CategoryStatusChange indicates a status change operation.
	CategoryStatusChange ApprovalCategory = "status_change"
	// CategoryCreate indicates a create operation.
	CategoryCreate ApprovalCategory = "create"
	// CategoryUpdate indicates a single update operation.
	CategoryUpdate ApprovalCategory = "update"
	// CategoryRead indicates a read operation.
	CategoryRead ApprovalCategory = "read"
)

// ApprovalRequest represents a pending approval request.
type ApprovalRequest struct {
	RequestID     string                 `json:"requestId"`
	UserID        string                 `json:"userId"`
	WorkspaceID   string                 `json:"workspaceId"`
	ToolName      string                 `json:"toolName"`
	Arguments     map[string]interface{} `json:"arguments"`
	Category      ApprovalCategory       `json:"category"`
	Description   string                 `json:"description"`
	Status        ApprovalStatus         `json:"status"`
	CreatedAt     time.Time              `json:"createdAt"`
	ExpiresAt     time.Time              `json:"expiresAt"`
	ApprovedAt    *time.Time             `json:"approvedAt,omitempty"`
	ApprovedBy    *string                `json:"approvedBy,omitempty"`
	RejectedAt    *time.Time             `json:"rejectedAt,omitempty"`
	RejectedBy    *string                `json:"rejectedBy,omitempty"`
	RejectionNote *string                `json:"rejectionNote,omitempty"`
}

// ApprovalRule defines when approval is required.
type ApprovalRule struct {
	ToolPattern       string           // Tool name pattern (supports wildcards)
	Category          ApprovalCategory // Category of the operation
	RequiresApproval  bool             // Whether approval is required
	ApprovalTimeout   time.Duration    // How long before the request expires
	AllowAutoApprove  bool             // Whether auto-approval is allowed
}

// DefaultApprovalRules returns the default set of approval rules.
func DefaultApprovalRules() []ApprovalRule {
	return []ApprovalRule{
		// Delete operations always require approval
		{ToolPattern: "task.delete", Category: CategoryDelete, RequiresApproval: true, ApprovalTimeout: 5 * time.Minute},
		{ToolPattern: "plan.archive", Category: CategoryDelete, RequiresApproval: true, ApprovalTimeout: 5 * time.Minute},

		// Status changes require approval for important states
		{ToolPattern: "task.move", Category: CategoryStatusChange, RequiresApproval: true, ApprovalTimeout: 5 * time.Minute},
		{ToolPattern: "plan.update", Category: CategoryStatusChange, RequiresApproval: true, ApprovalTimeout: 5 * time.Minute},
		{ToolPattern: "milestone.update", Category: CategoryStatusChange, RequiresApproval: true, ApprovalTimeout: 5 * time.Minute},

		// Create operations require approval for important entities
		{ToolPattern: "plan.create", Category: CategoryCreate, RequiresApproval: true, ApprovalTimeout: 10 * time.Minute},
		{ToolPattern: "workspace.create", Category: CategoryCreate, RequiresApproval: true, ApprovalTimeout: 10 * time.Minute},

		// Updates require approval
		{ToolPattern: "task.update", Category: CategoryUpdate, RequiresApproval: true, ApprovalTimeout: 5 * time.Minute},

		// Task creation is auto-approved (more common, less destructive)
		{ToolPattern: "task.create", Category: CategoryCreate, RequiresApproval: false, AllowAutoApprove: true},
		{ToolPattern: "milestone.create", Category: CategoryCreate, RequiresApproval: false, AllowAutoApprove: true},

		// Read operations are auto-approved
		{ToolPattern: "*.list", Category: CategoryRead, RequiresApproval: false, AllowAutoApprove: true},
		{ToolPattern: "*.get", Category: CategoryRead, RequiresApproval: false, AllowAutoApprove: true},
		{ToolPattern: "*.search", Category: CategoryRead, RequiresApproval: false, AllowAutoApprove: true},
		{ToolPattern: "activity.get", Category: CategoryRead, RequiresApproval: false, AllowAutoApprove: true},
	}
}

// SafetyChecker checks tool calls for safety and manages approvals.
type SafetyChecker struct {
	rules    []ApprovalRule
	pending  map[string]*ApprovalRequest
	mu       sync.RWMutex
}

// NewSafetyChecker creates a new SafetyChecker with default rules.
func NewSafetyChecker() *SafetyChecker {
	return &SafetyChecker{
		rules:   DefaultApprovalRules(),
		pending: make(map[string]*ApprovalRequest),
	}
}

// NewSafetyCheckerWithRules creates a new SafetyChecker with custom rules.
func NewSafetyCheckerWithRules(rules []ApprovalRule) *SafetyChecker {
	return &SafetyChecker{
		rules:   rules,
		pending: make(map[string]*ApprovalRequest),
	}
}

// CheckToolCall checks if a tool call requires approval.
func (s *SafetyChecker) CheckToolCall(toolName string, args map[string]interface{}, execCtx mcp.ExecutionContext) (*ApprovalRequest, bool) {
	rule := s.findMatchingRule(toolName, args)
	if rule == nil || !rule.RequiresApproval {
		return nil, false
	}

	// Create approval request
	request := s.createApprovalRequest(toolName, args, execCtx, rule)

	// Store in pending
	s.mu.Lock()
	s.pending[request.RequestID] = request
	s.mu.Unlock()

	return request, true
}

// findMatchingRule finds the first matching rule for a tool call.
func (s *SafetyChecker) findMatchingRule(toolName string, args map[string]interface{}) *ApprovalRule {
	for i := range s.rules {
		if matchPattern(s.rules[i].ToolPattern, toolName) {
			// Check for specific conditions
			if s.shouldApplyRule(&s.rules[i], toolName, args) {
				return &s.rules[i]
			}
		}
	}
	return nil
}

// matchPattern matches a tool name against a pattern with wildcards.
func matchPattern(pattern, name string) bool {
	if pattern == "*" {
		return true
	}

	// Handle prefix wildcard (e.g., "*.list")
	if len(pattern) > 1 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
	}

	// Handle suffix wildcard (e.g., "task.*")
	if len(pattern) > 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(name) >= len(prefix) && name[:len(prefix)] == prefix
	}

	// Exact match
	return pattern == name
}

// shouldApplyRule checks if a rule should be applied based on arguments.
func (s *SafetyChecker) shouldApplyRule(rule *ApprovalRule, toolName string, args map[string]interface{}) bool {
	// For status change rules, only apply if status is being changed
	if rule.Category == CategoryStatusChange {
		if toolName == "plan.update" {
			_, hasStatus := args["status"]
			return hasStatus
		}
		if toolName == "milestone.update" {
			_, hasStatus := args["status"]
			return hasStatus
		}
	}

	return true
}

// createApprovalRequest creates a new approval request.
func (s *SafetyChecker) createApprovalRequest(toolName string, args map[string]interface{}, execCtx mcp.ExecutionContext, rule *ApprovalRule) *ApprovalRequest {
	requestID := generateRequestID()
	now := time.Now()

	description := generateDescription(toolName, args, rule.Category)

	return &ApprovalRequest{
		RequestID:   requestID,
		UserID:      execCtx.UserID,
		WorkspaceID: execCtx.WorkspaceID,
		ToolName:    toolName,
		Arguments:   args,
		Category:    rule.Category,
		Description: description,
		Status:      ApprovalStatusPending,
		CreatedAt:   now,
		ExpiresAt:   now.Add(rule.ApprovalTimeout),
	}
}

// generateRequestID generates a unique request ID.
func generateRequestID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return "apr_" + hex.EncodeToString(bytes)
}

// generateDescription generates a human-readable description of the action.
func generateDescription(toolName string, args map[string]interface{}, category ApprovalCategory) string {
	switch toolName {
	case "task.delete":
		taskID, _ := args["taskId"].(string)
		return fmt.Sprintf("Delete task (ID: %s)", taskID)
	case "task.move":
		taskID, _ := args["taskId"].(string)
		status, _ := args["status"].(string)
		return fmt.Sprintf("Move task %s to status '%s'", taskID, status)
	case "task.update":
		taskID, _ := args["taskId"].(string)
		return fmt.Sprintf("Update task (ID: %s)", taskID)
	case "plan.create":
		name, _ := args["name"].(string)
		return fmt.Sprintf("Create new plan: %s", name)
	case "plan.update":
		planID, _ := args["planId"].(string)
		if status, ok := args["status"].(string); ok {
			return fmt.Sprintf("Change plan %s status to '%s'", planID, status)
		}
		return fmt.Sprintf("Update plan (ID: %s)", planID)
	case "plan.archive":
		planID, _ := args["planId"].(string)
		return fmt.Sprintf("Archive plan (ID: %s)", planID)
	case "workspace.create":
		name, _ := args["name"].(string)
		return fmt.Sprintf("Create new workspace: %s", name)
	case "milestone.update":
		milestoneID, _ := args["milestoneId"].(string)
		if status, ok := args["status"].(string); ok {
			return fmt.Sprintf("Change milestone %s status to '%s'", milestoneID, status)
		}
		return fmt.Sprintf("Update milestone (ID: %s)", milestoneID)
	default:
		return fmt.Sprintf("Execute %s with %d arguments", toolName, len(args))
	}
}

// GetPendingRequest retrieves a pending approval request by ID.
func (s *SafetyChecker) GetPendingRequest(requestID string) (*ApprovalRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	request, ok := s.pending[requestID]
	if !ok {
		return nil, fmt.Errorf("approval request not found: %s", requestID)
	}

	// Check if expired
	if time.Now().After(request.ExpiresAt) {
		request.Status = ApprovalStatusExpired
		return request, nil
	}

	return request, nil
}

// GetPendingRequestsForUser retrieves all pending approval requests for a user.
func (s *SafetyChecker) GetPendingRequestsForUser(userID string) []*ApprovalRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var requests []*ApprovalRequest
	now := time.Now()

	for _, request := range s.pending {
		if request.UserID == userID && request.Status == ApprovalStatusPending {
			// Check if expired
			if now.After(request.ExpiresAt) {
				request.Status = ApprovalStatusExpired
			} else {
				requests = append(requests, request)
			}
		}
	}

	return requests
}

// ApproveRequest approves a pending request.
func (s *SafetyChecker) ApproveRequest(ctx context.Context, requestID, approverID string) (*ApprovalRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	request, ok := s.pending[requestID]
	if !ok {
		return nil, fmt.Errorf("approval request not found: %s", requestID)
	}

	// Check if expired
	if time.Now().After(request.ExpiresAt) {
		request.Status = ApprovalStatusExpired
		return nil, fmt.Errorf("approval request has expired")
	}

	// Check if already processed
	if request.Status != ApprovalStatusPending {
		return nil, fmt.Errorf("approval request already processed: %s", request.Status)
	}

	now := time.Now()
	request.Status = ApprovalStatusApproved
	request.ApprovedAt = &now
	request.ApprovedBy = &approverID

	return request, nil
}

// RejectRequest rejects a pending request.
func (s *SafetyChecker) RejectRequest(ctx context.Context, requestID, rejecterID, note string) (*ApprovalRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	request, ok := s.pending[requestID]
	if !ok {
		return nil, fmt.Errorf("approval request not found: %s", requestID)
	}

	// Check if already processed
	if request.Status != ApprovalStatusPending {
		return nil, fmt.Errorf("approval request already processed: %s", request.Status)
	}

	now := time.Now()
	request.Status = ApprovalStatusRejected
	request.RejectedAt = &now
	request.RejectedBy = &rejecterID
	if note != "" {
		request.RejectionNote = &note
	}

	return request, nil
}

// RemoveExpiredRequests removes expired requests from the pending map.
func (s *SafetyChecker) RemoveExpiredRequests() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	now := time.Now()

	for id, request := range s.pending {
		if now.After(request.ExpiresAt) || request.Status != ApprovalStatusPending {
			delete(s.pending, id)
			removed++
		}
	}

	return removed
}

// StartCleanupWorker starts a background worker to clean up expired requests.
func (s *SafetyChecker) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				removed := s.RemoveExpiredRequests()
				if removed > 0 {
					slog.Info("cleaned up expired approval requests", "count", removed)
				}
			}
		}
	}()
}

// IsReadOperation checks if a tool name represents a read operation.
func IsReadOperation(toolName string) bool {
	readSuffixes := []string{".list", ".get", ".search"}
	for _, suffix := range readSuffixes {
		if len(toolName) > len(suffix) && toolName[len(toolName)-len(suffix):] == suffix {
			return true
		}
	}
	return toolName == "activity.get"
}

// IsDestructiveOperation checks if a tool name represents a destructive operation.
func IsDestructiveOperation(toolName string) bool {
	destructive := []string{"task.delete", "plan.archive"}
	for _, d := range destructive {
		if toolName == d {
			return true
		}
	}
	return false
}
