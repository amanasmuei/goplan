package models

import "fmt"

// ValidTransitions defines allowed state transitions
var ValidTransitions = map[TaskStatus][]TaskStatus{
	TaskStatusDraft:               {TaskStatusPendingAcknowledgment, TaskStatusCancelled},
	TaskStatusPendingAcknowledgment: {TaskStatusAcknowledged, TaskStatusCancelled},
	TaskStatusAcknowledged:        {TaskStatusActive, TaskStatusCancelled},
	TaskStatusActive:              {TaskStatusBlocked, TaskStatusPendingReview, TaskStatusCancelled},
	TaskStatusBlocked:             {TaskStatusActive, TaskStatusCancelled},
	TaskStatusPendingReview:       {TaskStatusCompleted, TaskStatusActive, TaskStatusCancelled},
	TaskStatusCompleted:           {}, // Terminal state
	TaskStatusCancelled:           {}, // Terminal state
}

// CanTransition checks if a state transition is valid
func CanTransition(from, to TaskStatus) bool {
	allowedStates, exists := ValidTransitions[from]
	if !exists {
		return false
	}
	for _, state := range allowedStates {
		if state == to {
			return true
		}
	}
	return false
}

// ValidateTransition returns an error if the transition is invalid
func ValidateTransition(from, to TaskStatus) error {
	if !CanTransition(from, to) {
		return fmt.Errorf("invalid state transition from '%s' to '%s'", from, to)
	}
	return nil
}

// GetAllowedTransitions returns all valid next states for a given status
func GetAllowedTransitions(from TaskStatus) []TaskStatus {
	return ValidTransitions[from]
}

// IsTerminalState checks if a status is a terminal state
func IsTerminalState(status TaskStatus) bool {
	return status == TaskStatusCompleted || status == TaskStatusCancelled
}

// RequiresAcknowledgment checks if task needs acknowledgment before starting
func RequiresAcknowledgment(status TaskStatus) bool {
	return status == TaskStatusDraft || status == TaskStatusPendingAcknowledgment
}

// RequiresReview checks if task needs review before completing
func RequiresReview(status TaskStatus) bool {
	return status == TaskStatusPendingReview
}

// AcknowledgmentAction represents how user acknowledged predictions
type AcknowledgmentAction string

const (
	AcknowledgmentAccept   AcknowledgmentAction = "accept"
	AcknowledgmentModify   AcknowledgmentAction = "modify"
	AcknowledgmentDisagree AcknowledgmentAction = "disagree"
)

// AcknowledgmentRequest represents the acknowledgment payload
type AcknowledgmentRequest struct {
	Action            AcknowledgmentAction `json:"action" validate:"required,oneof=accept modify disagree"`
	ModifiedEstimate  *float64             `json:"modified_estimate,omitempty"`
	DisagreementNotes string               `json:"disagreement_notes,omitempty"`
}

// Validate validates the acknowledgment request
func (r *AcknowledgmentRequest) Validate() error {
	switch r.Action {
	case AcknowledgmentModify:
		if r.ModifiedEstimate == nil || *r.ModifiedEstimate <= 0 {
			return fmt.Errorf("modified_estimate is required when action is 'modify'")
		}
	case AcknowledgmentDisagree:
		if len(r.DisagreementNotes) < 20 {
			return fmt.Errorf("disagreement_notes must be at least 20 characters when action is 'disagree'")
		}
	}
	return nil
}
