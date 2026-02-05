package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/api/types"
	"github.com/goplan/goplan/internal/domain/shared"
	"github.com/goplan/goplan/internal/logging"
)

// BaseHandler provides base handler functionality with common utilities.
type BaseHandler struct{}

// NewBaseHandler creates a new base handler.
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// WriteJSON writes a JSON response with the given status code.
func (h *BaseHandler) WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logging.Default().Error("Error encoding JSON response", "error", err)
	}
}

// WriteSuccess writes a success response.
func (h *BaseHandler) WriteSuccess(w http.ResponseWriter, data interface{}) {
	h.WriteJSON(w, http.StatusOK, types.NewSuccessResponse(data))
}

// WriteCreated writes a created response.
func (h *BaseHandler) WriteCreated(w http.ResponseWriter, data interface{}) {
	h.WriteJSON(w, http.StatusCreated, types.NewSuccessResponse(data))
}

// WriteNoContent writes a no content response.
func (h *BaseHandler) WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WritePaginated writes a paginated response.
func (h *BaseHandler) WritePaginated(w http.ResponseWriter, items interface{}, totalCount int64, page, pageSize int) {
	h.WriteJSON(w, http.StatusOK, types.NewPaginatedResponse(items, totalCount, page, pageSize))
}

// WriteError writes an error response.
func (h *BaseHandler) WriteError(w http.ResponseWriter, err error) {
	status := types.HTTPStatusFromError(err)
	code := types.ErrorCodeFromError(err)
	details := types.ErrorDetailsFromError(err)
	message := err.Error()

	// Extract message from domain error if available
	if domainErr, ok := err.(*shared.DomainError); ok {
		message = domainErr.Message
	}

	if validationErrs, ok := err.(*shared.ValidationErrors); ok {
		if len(validationErrs.Errors) == 1 {
			message = validationErrs.Errors[0].Message
		} else {
			message = "validation failed"
		}
	}

	response := types.NewErrorResponseWithDetails(code, message, details)
	h.WriteJSON(w, status, response)
}

// WriteErrorWithStatus writes an error response with a specific status code.
func (h *BaseHandler) WriteErrorWithStatus(w http.ResponseWriter, status int, code, message string) {
	h.WriteJSON(w, status, types.NewErrorResponse(code, message))
}

// WriteNotFound writes a not found error response.
func (h *BaseHandler) WriteNotFound(w http.ResponseWriter, resource string) {
	h.WriteJSON(w, http.StatusNotFound, types.NewErrorResponse("NOT_FOUND", resource+" not found"))
}

// WriteBadRequest writes a bad request error response.
func (h *BaseHandler) WriteBadRequest(w http.ResponseWriter, message string) {
	h.WriteJSON(w, http.StatusBadRequest, types.NewErrorResponse("BAD_REQUEST", message))
}

// WriteUnauthorized writes an unauthorized error response.
func (h *BaseHandler) WriteUnauthorized(w http.ResponseWriter, message string) {
	h.WriteJSON(w, http.StatusUnauthorized, types.NewErrorResponse("UNAUTHORIZED", message))
}

// WriteForbidden writes a forbidden error response.
func (h *BaseHandler) WriteForbidden(w http.ResponseWriter, message string) {
	h.WriteJSON(w, http.StatusForbidden, types.NewErrorResponse("FORBIDDEN", message))
}

// WriteInternalError writes an internal server error response.
func (h *BaseHandler) WriteInternalError(w http.ResponseWriter) {
	h.WriteJSON(w, http.StatusInternalServerError, types.NewErrorResponse("INTERNAL_ERROR", "an internal error occurred"))
}

// WriteMethodNotAllowed writes a method not allowed error response.
func (h *BaseHandler) WriteMethodNotAllowed(w http.ResponseWriter, allowed []string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	h.WriteJSON(w, http.StatusMethodNotAllowed, types.NewErrorResponse("METHOD_NOT_ALLOWED", "method not allowed"))
}

// DecodeJSON decodes JSON from the request body.
func (h *BaseHandler) DecodeJSON(r *http.Request, target interface{}) error {
	return types.DecodeJSON(r, target)
}

// ParsePagination extracts pagination parameters from the request.
func (h *BaseHandler) ParsePagination(r *http.Request) types.PaginationParams {
	return types.ParsePagination(r)
}
