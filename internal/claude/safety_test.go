package claude

import (
	"context"
	"testing"
	"time"

	"github.com/goplan/goplan/internal/mcp"
)

func TestSafetyChecker_CheckToolCall(t *testing.T) {
	checker := NewSafetyChecker()

	execCtx := mcp.ExecutionContext{
		UserID:      "user-123",
		WorkspaceID: "ws-456",
		Role:        "member",
	}

	tests := []struct {
		name             string
		toolName         string
		args             map[string]interface{}
		wantApproval     bool
		wantDescription  string
	}{
		{
			name:         "read operation - no approval needed",
			toolName:     "task.list",
			args:         map[string]interface{}{"planId": "plan-1"},
			wantApproval: false,
		},
		{
			name:         "get operation - no approval needed",
			toolName:     "plan.get",
			args:         map[string]interface{}{"planId": "plan-1"},
			wantApproval: false,
		},
		{
			name:         "search operation - no approval needed",
			toolName:     "task.search",
			args:         map[string]interface{}{"planId": "plan-1", "query": "test"},
			wantApproval: false,
		},
		{
			name:         "delete operation - approval required",
			toolName:     "task.delete",
			args:         map[string]interface{}{"taskId": "task-1"},
			wantApproval: true,
		},
		{
			name:         "archive operation - approval required",
			toolName:     "plan.archive",
			args:         map[string]interface{}{"planId": "plan-1"},
			wantApproval: true,
		},
		{
			name:         "task move - approval required",
			toolName:     "task.move",
			args:         map[string]interface{}{"taskId": "task-1", "status": "done"},
			wantApproval: true,
		},
		{
			name:         "plan create - approval required",
			toolName:     "plan.create",
			args:         map[string]interface{}{"name": "Test Plan", "domain": "generic"},
			wantApproval: true,
		},
		{
			name:         "task create - no approval needed",
			toolName:     "task.create",
			args:         map[string]interface{}{"planId": "plan-1", "title": "New Task"},
			wantApproval: false,
		},
		{
			name:         "plan update with status - approval required",
			toolName:     "plan.update",
			args:         map[string]interface{}{"planId": "plan-1", "status": "completed"},
			wantApproval: true,
		},
		{
			name:         "plan update without status - approval required",
			toolName:     "plan.update",
			args:         map[string]interface{}{"planId": "plan-1", "name": "New Name"},
			wantApproval: false, // Only status changes require approval
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, requiresApproval := checker.CheckToolCall(tt.toolName, tt.args, execCtx)

			if requiresApproval != tt.wantApproval {
				t.Errorf("CheckToolCall() requiresApproval = %v, want %v", requiresApproval, tt.wantApproval)
			}

			if tt.wantApproval {
				if request == nil {
					t.Error("CheckToolCall() request should not be nil when approval is required")
					return
				}
				if request.ToolName != tt.toolName {
					t.Errorf("CheckToolCall() request.ToolName = %v, want %v", request.ToolName, tt.toolName)
				}
				if request.Status != ApprovalStatusPending {
					t.Errorf("CheckToolCall() request.Status = %v, want %v", request.Status, ApprovalStatusPending)
				}
				if request.UserID != execCtx.UserID {
					t.Errorf("CheckToolCall() request.UserID = %v, want %v", request.UserID, execCtx.UserID)
				}
			}
		})
	}
}

func TestSafetyChecker_ApproveRequest(t *testing.T) {
	checker := NewSafetyChecker()
	ctx := context.Background()

	execCtx := mcp.ExecutionContext{
		UserID:      "user-123",
		WorkspaceID: "ws-456",
		Role:        "member",
	}

	// Create a pending request
	request, requiresApproval := checker.CheckToolCall("task.delete", map[string]interface{}{
		"taskId": "task-1",
	}, execCtx)

	if !requiresApproval || request == nil {
		t.Fatal("Expected approval to be required for task.delete")
	}

	// Approve the request
	approved, err := checker.ApproveRequest(ctx, request.RequestID, "admin-1")
	if err != nil {
		t.Fatalf("ApproveRequest() error = %v", err)
	}

	if approved.Status != ApprovalStatusApproved {
		t.Errorf("ApproveRequest() status = %v, want %v", approved.Status, ApprovalStatusApproved)
	}

	if approved.ApprovedBy == nil || *approved.ApprovedBy != "admin-1" {
		t.Error("ApproveRequest() ApprovedBy should be set")
	}

	if approved.ApprovedAt == nil {
		t.Error("ApproveRequest() ApprovedAt should be set")
	}
}

func TestSafetyChecker_RejectRequest(t *testing.T) {
	checker := NewSafetyChecker()
	ctx := context.Background()

	execCtx := mcp.ExecutionContext{
		UserID:      "user-123",
		WorkspaceID: "ws-456",
		Role:        "member",
	}

	// Create a pending request
	request, _ := checker.CheckToolCall("plan.archive", map[string]interface{}{
		"planId": "plan-1",
	}, execCtx)

	// Reject the request
	rejected, err := checker.RejectRequest(ctx, request.RequestID, "admin-1", "Not now")
	if err != nil {
		t.Fatalf("RejectRequest() error = %v", err)
	}

	if rejected.Status != ApprovalStatusRejected {
		t.Errorf("RejectRequest() status = %v, want %v", rejected.Status, ApprovalStatusRejected)
	}

	if rejected.RejectionNote == nil || *rejected.RejectionNote != "Not now" {
		t.Error("RejectRequest() RejectionNote should be set")
	}
}

func TestSafetyChecker_GetPendingRequestsForUser(t *testing.T) {
	checker := NewSafetyChecker()

	execCtx := mcp.ExecutionContext{
		UserID:      "user-123",
		WorkspaceID: "ws-456",
		Role:        "member",
	}

	// Create some pending requests
	checker.CheckToolCall("task.delete", map[string]interface{}{"taskId": "task-1"}, execCtx)
	checker.CheckToolCall("plan.archive", map[string]interface{}{"planId": "plan-1"}, execCtx)

	// Create request for different user
	otherCtx := mcp.ExecutionContext{
		UserID:      "user-other",
		WorkspaceID: "ws-456",
		Role:        "member",
	}
	checker.CheckToolCall("task.delete", map[string]interface{}{"taskId": "task-2"}, otherCtx)

	// Get pending requests for user-123
	requests := checker.GetPendingRequestsForUser("user-123")

	if len(requests) != 2 {
		t.Errorf("GetPendingRequestsForUser() returned %d requests, want 2", len(requests))
	}

	for _, req := range requests {
		if req.UserID != "user-123" {
			t.Errorf("GetPendingRequestsForUser() returned request for wrong user: %s", req.UserID)
		}
	}
}

func TestSafetyChecker_RequestExpiration(t *testing.T) {
	// Create checker with short timeout rules for testing
	rules := []ApprovalRule{
		{ToolPattern: "task.delete", Category: CategoryDelete, RequiresApproval: true, ApprovalTimeout: 1 * time.Millisecond},
	}
	checker := NewSafetyCheckerWithRules(rules)
	ctx := context.Background()

	execCtx := mcp.ExecutionContext{
		UserID:      "user-123",
		WorkspaceID: "ws-456",
		Role:        "member",
	}

	// Create a request
	request, _ := checker.CheckToolCall("task.delete", map[string]interface{}{
		"taskId": "task-1",
	}, execCtx)

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	// Try to approve - should fail
	_, err := checker.ApproveRequest(ctx, request.RequestID, "admin-1")
	if err == nil {
		t.Error("ApproveRequest() should fail for expired request")
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"*", "anything", true},
		{"*.list", "task.list", true},
		{"*.list", "plan.list", true},
		{"*.list", "task.get", false},
		{"task.*", "task.create", true},
		{"task.*", "task.update", true},
		{"task.*", "plan.create", false},
		{"task.create", "task.create", true},
		{"task.create", "task.update", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.name)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}

func TestIsReadOperation(t *testing.T) {
	tests := []struct {
		toolName string
		want     bool
	}{
		{"task.list", true},
		{"plan.get", true},
		{"task.search", true},
		{"activity.get", true},
		{"task.create", false},
		{"task.update", false},
		{"task.delete", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			got := IsReadOperation(tt.toolName)
			if got != tt.want {
				t.Errorf("IsReadOperation(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestIsDestructiveOperation(t *testing.T) {
	tests := []struct {
		toolName string
		want     bool
	}{
		{"task.delete", true},
		{"plan.archive", true},
		{"task.create", false},
		{"task.update", false},
		{"task.list", false},
	}

	for _, tt := range tests {
		t.Run(tt.toolName, func(t *testing.T) {
			got := IsDestructiveOperation(tt.toolName)
			if got != tt.want {
				t.Errorf("IsDestructiveOperation(%q) = %v, want %v", tt.toolName, got, tt.want)
			}
		})
	}
}
