// Package shared provides shared error types used across domain entities.
package shared

import (
	"errors"
	"fmt"
)

// Domain error types
var (
	// ErrNotFound is returned when an entity is not found.
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists is returned when an entity already exists.
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when the user is not authorized.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user is not allowed to perform the action.
	ErrForbidden = errors.New("forbidden")

	// ErrConflict is returned when there is a conflict with existing data.
	ErrConflict = errors.New("conflict")

	// ErrInternal is returned when an internal error occurs.
	ErrInternal = errors.New("internal error")
)

// DomainError represents a domain-specific error with additional context.
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(entityType, identifier string) *DomainError {
	return &DomainError{
		Code:    "ENTITY_NOT_FOUND",
		Message: fmt.Sprintf("%s with identifier '%s' not found", entityType, identifier),
		Err:     ErrNotFound,
	}
}

// NewAlreadyExistsError creates a new already exists error.
func NewAlreadyExistsError(entityType, field, value string) *DomainError {
	return &DomainError{
		Code:    "ENTITY_ALREADY_EXISTS",
		Message: fmt.Sprintf("%s with %s '%s' already exists", entityType, field, value),
		Field:   field,
		Err:     ErrAlreadyExists,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
		Err:     ErrInvalidInput,
	}
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Err:     ErrUnauthorized,
	}
}

// NewForbiddenError creates a new forbidden error.
func NewForbiddenError(action, resource string) *DomainError {
	return &DomainError{
		Code:    "PERMISSION_DENIED",
		Message: fmt.Sprintf("you don't have permission to %s this %s", action, resource),
		Err:     ErrForbidden,
	}
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string) *DomainError {
	return &DomainError{
		Code:    "CONFLICT",
		Message: message,
		Err:     ErrConflict,
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Err:     err,
	}
}

// ValidationErrors represents a collection of validation errors.
type ValidationErrors struct {
	Errors []*DomainError `json:"errors"`
}

// Add adds a validation error to the collection.
func (v *ValidationErrors) Add(field, message string) {
	v.Errors = append(v.Errors, NewValidationError(field, message))
}

// HasErrors returns true if there are validation errors.
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// Error implements the error interface.
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "no validation errors"
	}
	if len(v.Errors) == 1 {
		return v.Errors[0].Error()
	}
	return fmt.Sprintf("multiple validation errors: %d errors", len(v.Errors))
}

// MCP-specific error codes
const (
	MCPErrorPermissionDenied = "PERMISSION_DENIED"
	MCPErrorEntityNotFound   = "ENTITY_NOT_FOUND"
	MCPErrorRateLimited      = "RATE_LIMITED"
	MCPErrorInvalidIntent    = "INVALID_INTENT"
	MCPErrorLowConfidence    = "LOW_CONFIDENCE"
)

// MCPError represents an MCP-specific error.
type MCPError struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewMCPError creates a new MCP error.
func NewMCPError(code, message string) *MCPError {
	return &MCPError{
		Error:   true,
		Code:    code,
		Message: message,
	}
}
