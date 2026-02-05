package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     TaskStatus
		to       TaskStatus
		expected bool
	}{
		// Valid transitions from Draft
		{"Draft to PendingAcknowledgment", TaskStatusDraft, TaskStatusPendingAcknowledgment, true},
		{"Draft to Cancelled", TaskStatusDraft, TaskStatusCancelled, true},
		{"Draft to Active (invalid)", TaskStatusDraft, TaskStatusActive, false},
		{"Draft to Completed (invalid)", TaskStatusDraft, TaskStatusCompleted, false},

		// Valid transitions from PendingAcknowledgment
		{"PendingAck to Acknowledged", TaskStatusPendingAcknowledgment, TaskStatusAcknowledged, true},
		{"PendingAck to Cancelled", TaskStatusPendingAcknowledgment, TaskStatusCancelled, true},
		{"PendingAck to Active (invalid)", TaskStatusPendingAcknowledgment, TaskStatusActive, false},

		// Valid transitions from Acknowledged
		{"Acknowledged to Active", TaskStatusAcknowledged, TaskStatusActive, true},
		{"Acknowledged to Cancelled", TaskStatusAcknowledged, TaskStatusCancelled, true},
		{"Acknowledged to Completed (invalid)", TaskStatusAcknowledged, TaskStatusCompleted, false},

		// Valid transitions from Active
		{"Active to Blocked", TaskStatusActive, TaskStatusBlocked, true},
		{"Active to PendingReview", TaskStatusActive, TaskStatusPendingReview, true},
		{"Active to Cancelled", TaskStatusActive, TaskStatusCancelled, true},
		{"Active to Completed (invalid)", TaskStatusActive, TaskStatusCompleted, false},
		{"Active to Draft (invalid)", TaskStatusActive, TaskStatusDraft, false},

		// Valid transitions from Blocked
		{"Blocked to Active", TaskStatusBlocked, TaskStatusActive, true},
		{"Blocked to Cancelled", TaskStatusBlocked, TaskStatusCancelled, true},
		{"Blocked to Completed (invalid)", TaskStatusBlocked, TaskStatusCompleted, false},

		// Valid transitions from PendingReview
		{"PendingReview to Completed", TaskStatusPendingReview, TaskStatusCompleted, true},
		{"PendingReview to Active", TaskStatusPendingReview, TaskStatusActive, true},
		{"PendingReview to Cancelled", TaskStatusPendingReview, TaskStatusCancelled, true},
		{"PendingReview to Blocked (invalid)", TaskStatusPendingReview, TaskStatusBlocked, false},

		// Terminal states
		{"Completed to anything (invalid)", TaskStatusCompleted, TaskStatusActive, false},
		{"Cancelled to anything (invalid)", TaskStatusCancelled, TaskStatusActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result, "CanTransition(%s, %s) should be %v", tt.from, tt.to, tt.expected)
		})
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    TaskStatus
		to      TaskStatus
		wantErr bool
	}{
		{"Valid transition", TaskStatusDraft, TaskStatusPendingAcknowledgment, false},
		{"Invalid transition", TaskStatusDraft, TaskStatusCompleted, true},
		{"From terminal state", TaskStatusCompleted, TaskStatusActive, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransition(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid state transition")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetAllowedTransitions(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected []TaskStatus
	}{
		{
			"Draft allowed transitions",
			TaskStatusDraft,
			[]TaskStatus{TaskStatusPendingAcknowledgment, TaskStatusCancelled},
		},
		{
			"Active allowed transitions",
			TaskStatusActive,
			[]TaskStatus{TaskStatusBlocked, TaskStatusPendingReview, TaskStatusCancelled},
		},
		{
			"Completed has no transitions",
			TaskStatusCompleted,
			[]TaskStatus{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAllowedTransitions(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTerminalState(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected bool
	}{
		{"Completed is terminal", TaskStatusCompleted, true},
		{"Cancelled is terminal", TaskStatusCancelled, true},
		{"Draft is not terminal", TaskStatusDraft, false},
		{"Active is not terminal", TaskStatusActive, false},
		{"Blocked is not terminal", TaskStatusBlocked, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTerminalState(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequiresAcknowledgment(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected bool
	}{
		{"Draft requires acknowledgment", TaskStatusDraft, true},
		{"PendingAck requires acknowledgment", TaskStatusPendingAcknowledgment, true},
		{"Acknowledged does not require", TaskStatusAcknowledged, false},
		{"Active does not require", TaskStatusActive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequiresAcknowledgment(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequiresReview(t *testing.T) {
	tests := []struct {
		name     string
		status   TaskStatus
		expected bool
	}{
		{"PendingReview requires review", TaskStatusPendingReview, true},
		{"Active does not require review", TaskStatusActive, false},
		{"Completed does not require review", TaskStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequiresReview(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAcknowledgmentRequest_Validate(t *testing.T) {
	estimate := 5.0
	tests := []struct {
		name    string
		req     AcknowledgmentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Accept action - valid",
			req:     AcknowledgmentRequest{Action: AcknowledgmentAccept},
			wantErr: false,
		},
		{
			name:    "Modify action with estimate - valid",
			req:     AcknowledgmentRequest{Action: AcknowledgmentModify, ModifiedEstimate: &estimate},
			wantErr: false,
		},
		{
			name:    "Modify action without estimate - invalid",
			req:     AcknowledgmentRequest{Action: AcknowledgmentModify},
			wantErr: true,
			errMsg:  "modified_estimate is required",
		},
		{
			name:    "Disagree action with notes - valid",
			req:     AcknowledgmentRequest{Action: AcknowledgmentDisagree, DisagreementNotes: "This is a detailed explanation of why I disagree with the prediction."},
			wantErr: false,
		},
		{
			name:    "Disagree action with short notes - invalid",
			req:     AcknowledgmentRequest{Action: AcknowledgmentDisagree, DisagreementNotes: "Too short"},
			wantErr: true,
			errMsg:  "disagreement_notes must be at least 20 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskLifecycleFlow(t *testing.T) {
	// Simulate a complete task lifecycle
	t.Run("Happy path: Draft -> PendingAck -> Acknowledged -> Active -> PendingReview -> Completed", func(t *testing.T) {
		currentStatus := TaskStatusDraft

		// Draft -> PendingAcknowledgment
		assert.True(t, CanTransition(currentStatus, TaskStatusPendingAcknowledgment))
		currentStatus = TaskStatusPendingAcknowledgment

		// PendingAcknowledgment -> Acknowledged
		assert.True(t, CanTransition(currentStatus, TaskStatusAcknowledged))
		currentStatus = TaskStatusAcknowledged

		// Acknowledged -> Active
		assert.True(t, CanTransition(currentStatus, TaskStatusActive))
		currentStatus = TaskStatusActive

		// Active -> PendingReview
		assert.True(t, CanTransition(currentStatus, TaskStatusPendingReview))
		currentStatus = TaskStatusPendingReview

		// PendingReview -> Completed
		assert.True(t, CanTransition(currentStatus, TaskStatusCompleted))
		currentStatus = TaskStatusCompleted

		// Verify terminal state
		assert.True(t, IsTerminalState(currentStatus))
	})

	t.Run("With blocker: Active -> Blocked -> Active -> Complete", func(t *testing.T) {
		currentStatus := TaskStatusActive

		// Active -> Blocked
		assert.True(t, CanTransition(currentStatus, TaskStatusBlocked))
		currentStatus = TaskStatusBlocked

		// Blocked -> Active (after resolving blocker)
		assert.True(t, CanTransition(currentStatus, TaskStatusActive))
		currentStatus = TaskStatusActive

		// Active -> PendingReview -> Completed
		assert.True(t, CanTransition(currentStatus, TaskStatusPendingReview))
	})

	t.Run("Cancellation from any non-terminal state", func(t *testing.T) {
		nonTerminalStates := []TaskStatus{
			TaskStatusDraft,
			TaskStatusPendingAcknowledgment,
			TaskStatusAcknowledged,
			TaskStatusActive,
			TaskStatusBlocked,
			TaskStatusPendingReview,
		}

		for _, status := range nonTerminalStates {
			assert.True(t, CanTransition(status, TaskStatusCancelled),
				"Should be able to cancel from %s", status)
		}
	})
}

func TestInvalidTransitionPaths(t *testing.T) {
	tests := []struct {
		name string
		from TaskStatus
		to   TaskStatus
	}{
		{"Skip acknowledgment", TaskStatusDraft, TaskStatusActive},
		{"Skip active", TaskStatusAcknowledged, TaskStatusPendingReview},
		{"Complete without review", TaskStatusActive, TaskStatusCompleted},
		{"Reverse from completed", TaskStatusCompleted, TaskStatusActive},
		{"Reverse from cancelled", TaskStatusCancelled, TaskStatusDraft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, CanTransition(tt.from, tt.to),
				"Should not allow transition from %s to %s", tt.from, tt.to)
		})
	}
}
