// Package errors provides standardized error types and handling for the GoPlan backend.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

// Error codes for API responses.
const (
	// Client errors (4xx)
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	CodeConflict            = "CONFLICT"
	CodeValidationFailed    = "VALIDATION_FAILED"
	CodeTooManyRequests     = "TOO_MANY_REQUESTS"
	CodePayloadTooLarge     = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMedia    = "UNSUPPORTED_MEDIA_TYPE"

	// Server errors (5xx)
	CodeInternalError       = "INTERNAL_ERROR"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeDatabaseError       = "DATABASE_ERROR"
	CodeExternalServiceError = "EXTERNAL_SERVICE_ERROR"
)

// APIError represents a standardized API error.
type APIError struct {
	// Public fields (returned to client)
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`

	// Internal fields (not returned to client)
	HTTPStatus int    `json:"-"`
	Internal   error  `json:"-"`
	Stack      string `json:"-"`
}

// ErrorDetail provides additional error information.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the internal error.
func (e *APIError) Unwrap() error {
	return e.Internal
}

// WithInternal adds an internal error (not exposed to client).
func (e *APIError) WithInternal(err error) *APIError {
	e.Internal = err
	return e
}

// WithStack captures the stack trace.
func (e *APIError) WithStack() *APIError {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	e.Stack = string(buf[:n])
	return e
}

// WithDetail adds an error detail.
func (e *APIError) WithDetail(field, message string) *APIError {
	e.Details = append(e.Details, ErrorDetail{Field: field, Message: message})
	return e
}

// WithDetails adds multiple error details.
func (e *APIError) WithDetails(details []ErrorDetail) *APIError {
	e.Details = append(e.Details, details...)
	return e
}

// ToResponse converts the error to a response payload.
func (e *APIError) ToResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"error":   true,
		"code":    e.Code,
		"message": e.Message,
	}
	if len(e.Details) > 0 {
		resp["details"] = e.Details
	}
	return resp
}

// WriteResponse writes the error response to the HTTP response writer.
func (e *APIError) WriteResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.HTTPStatus)
	_ = json.NewEncoder(w).Encode(e.ToResponse())
}

// Client error constructors

// BadRequest creates a 400 Bad Request error.
func BadRequest(message string) *APIError {
	return &APIError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(message string) *APIError {
	if message == "" {
		message = "authentication required"
	}
	return &APIError{
		Code:       CodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// Forbidden creates a 403 Forbidden error.
func Forbidden(message string) *APIError {
	if message == "" {
		message = "access denied"
	}
	return &APIError{
		Code:       CodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NotFound creates a 404 Not Found error.
func NotFound(resource string) *APIError {
	return &APIError{
		Code:       CodeNotFound,
		Message:    resource + " not found",
		HTTPStatus: http.StatusNotFound,
	}
}

// MethodNotAllowed creates a 405 Method Not Allowed error.
func MethodNotAllowed(allowed []string) *APIError {
	return &APIError{
		Code:       CodeMethodNotAllowed,
		Message:    "method not allowed",
		HTTPStatus: http.StatusMethodNotAllowed,
	}
}

// Conflict creates a 409 Conflict error.
func Conflict(message string) *APIError {
	return &APIError{
		Code:       CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// ValidationFailed creates a 422 Unprocessable Entity error.
func ValidationFailed(message string) *APIError {
	return &APIError{
		Code:       CodeValidationFailed,
		Message:    message,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

// TooManyRequests creates a 429 Too Many Requests error.
func TooManyRequests(message string) *APIError {
	if message == "" {
		message = "rate limit exceeded"
	}
	return &APIError{
		Code:       CodeTooManyRequests,
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// PayloadTooLarge creates a 413 Payload Too Large error.
func PayloadTooLarge(maxSize string) *APIError {
	return &APIError{
		Code:       CodePayloadTooLarge,
		Message:    "request body too large, maximum size is " + maxSize,
		HTTPStatus: http.StatusRequestEntityTooLarge,
	}
}

// UnsupportedMediaType creates a 415 Unsupported Media Type error.
func UnsupportedMediaType(expected string) *APIError {
	return &APIError{
		Code:       CodeUnsupportedMedia,
		Message:    "unsupported media type, expected " + expected,
		HTTPStatus: http.StatusUnsupportedMediaType,
	}
}

// Server error constructors

// InternalError creates a 500 Internal Server Error.
func InternalError(message string) *APIError {
	if message == "" {
		message = "an internal error occurred"
	}
	return &APIError{
		Code:       CodeInternalError,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// ServiceUnavailable creates a 503 Service Unavailable error.
func ServiceUnavailable(message string) *APIError {
	if message == "" {
		message = "service temporarily unavailable"
	}
	return &APIError{
		Code:       CodeServiceUnavailable,
		Message:    message,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// DatabaseError creates a database error (500).
func DatabaseError(err error) *APIError {
	return &APIError{
		Code:       CodeDatabaseError,
		Message:    "database error occurred",
		HTTPStatus: http.StatusInternalServerError,
		Internal:   err,
	}
}

// ExternalServiceError creates an external service error (502).
func ExternalServiceError(service string, err error) *APIError {
	return &APIError{
		Code:       CodeExternalServiceError,
		Message:    service + " service error",
		HTTPStatus: http.StatusBadGateway,
		Internal:   err,
	}
}

// Wrap wraps an error with an API error.
func Wrap(err error, code string, message string, status int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
		Internal:   err,
	}
}

// FromError converts a standard error to an APIError.
func FromError(err error) *APIError {
	if err == nil {
		return nil
	}

	// Check if it's already an APIError
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}

	// Return a generic internal error
	return InternalError("").WithInternal(err)
}

// Is checks if an error is of a specific type.
func Is(err error, code string) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == code
	}
	return false
}

// ErrorHandler is middleware that handles errors and converts them to proper responses.
type ErrorHandler struct {
	logger      interface{ Error(string, ...any) }
	showDetails bool // Whether to show internal error details (development only)
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(logger interface{ Error(string, ...any) }, showDetails bool) *ErrorHandler {
	return &ErrorHandler{
		logger:      logger,
		showDetails: showDetails,
	}
}

// Handle handles an error and writes the response.
func (h *ErrorHandler) Handle(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := FromError(err)

	// Log internal errors
	if apiErr.HTTPStatus >= 500 && h.logger != nil {
		h.logger.Error("internal error",
			"path", r.URL.Path,
			"method", r.Method,
			"error", apiErr.Error(),
			"stack", apiErr.Stack,
		)
	}

	// In production, hide internal error details
	if apiErr.HTTPStatus >= 500 && !h.showDetails {
		apiErr.Message = "an internal error occurred"
		apiErr.Details = nil
	}

	apiErr.WriteResponse(w)
}

// RecoverMiddleware returns middleware that recovers from panics.
func RecoverMiddleware(logger interface{ Error(string, ...any) }, showDetails bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					stack := string(buf[:n])

					if logger != nil {
						logger.Error("panic recovered",
							"panic", rec,
							"path", r.URL.Path,
							"method", r.Method,
							"stack", stack,
						)
					}

					apiErr := InternalError("")
					if showDetails {
						apiErr.Message = fmt.Sprintf("panic: %v", rec)
					}
					apiErr.WriteResponse(w)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
