package handlers

import (
	"net/http"
	"strings"

	"github.com/goplan/goplan/internal/api/middleware"
	"github.com/goplan/goplan/internal/api/types"
	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/repository"
)

// CommentHandler handles comment-related HTTP requests.
type CommentHandler struct {
	*BaseHandler
	commentRepo repository.CommentRepository
	taskRepo    repository.TaskRepository
}

// NewCommentHandler creates a new comment handler.
func NewCommentHandler(commentRepo repository.CommentRepository, taskRepo repository.TaskRepository) *CommentHandler {
	return &CommentHandler{
		BaseHandler: NewBaseHandler(),
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
	}
}

// ListCommentsByTask handles GET /api/v1/tasks/:tid/comments
func (h *CommentHandler) ListCommentsByTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := extractTaskIDFromCommentPath(r.URL.Path)

	if taskID == "" {
		h.WriteBadRequest(w, "task ID is required")
		return
	}

	// Verify task exists
	_, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	pagination := h.ParsePagination(r)

	result, err := h.commentRepo.ListByTaskPaginated(ctx, taskID, repository.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	})
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Convert to response format
	responses := make([]*task.CommentResponse, len(result.Items))
	for i, c := range result.Items {
		responses[i] = c.ToResponse()
	}

	h.WritePaginated(w, responses, result.TotalCount, result.Page, result.PageSize)
}

// CreateComment handles POST /api/v1/tasks/:tid/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	role := auth.GetUserRole(ctx)
	taskID := extractTaskIDFromCommentPath(r.URL.Path)

	if taskID == "" {
		h.WriteBadRequest(w, "task ID is required")
		return
	}

	// Check permissions - viewers can also comment
	if !middleware.CanView(role) {
		h.WriteForbidden(w, "you don't have permission to comment")
		return
	}

	// Verify task exists
	_, err := h.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	var req types.CreateCommentRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	// Generate a new ID
	id := generateID()

	// Create comment
	c := task.NewComment(id, taskID, userID, req.Content, req.Mentions)

	if err := h.commentRepo.Create(ctx, c); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteCreated(w, c.ToResponse())
}

// UpdateComment handles PUT /api/v1/comments/:id
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	commentID := extractCommentID(r.URL.Path)

	if commentID == "" {
		h.WriteBadRequest(w, "comment ID is required")
		return
	}

	// Verify comment exists and user is the author
	existingComment, err := h.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Only the author can update their comment
	if existingComment.UserID != userID {
		h.WriteForbidden(w, "you can only update your own comments")
		return
	}

	var req types.UpdateCommentRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteError(w, err)
		return
	}

	if err := req.Validate(); err != nil {
		h.WriteError(w, err)
		return
	}

	updatedComment, err := h.commentRepo.Update(ctx, commentID, req.Content, req.Mentions)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteSuccess(w, updatedComment.ToResponse())
}

// DeleteComment handles DELETE /api/v1/comments/:id
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.GetUserID(ctx)
	role := auth.GetUserRole(ctx)
	commentID := extractCommentID(r.URL.Path)

	if commentID == "" {
		h.WriteBadRequest(w, "comment ID is required")
		return
	}

	// Verify comment exists
	existingComment, err := h.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	// Only the author or admins can delete comments
	if existingComment.UserID != userID && !middleware.CanManage(role) {
		h.WriteForbidden(w, "you don't have permission to delete this comment")
		return
	}

	if err := h.commentRepo.Delete(ctx, commentID); err != nil {
		h.WriteError(w, err)
		return
	}

	h.WriteNoContent(w)
}

// ServeHTTP routes requests to the appropriate handler method.
func (h *CommentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle /api/v1/tasks/:tid/comments
	if strings.Contains(path, "/tasks/") && strings.HasSuffix(path, "/comments") {
		switch r.Method {
		case http.MethodGet:
			h.ListCommentsByTask(w, r)
		case http.MethodPost:
			h.CreateComment(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodGet, http.MethodPost})
		}
		return
	}

	// Handle /api/v1/comments/:id
	if strings.HasPrefix(path, "/api/v1/comments/") {
		switch r.Method {
		case http.MethodPut:
			h.UpdateComment(w, r)
		case http.MethodDelete:
			h.DeleteComment(w, r)
		default:
			h.WriteMethodNotAllowed(w, []string{http.MethodPut, http.MethodDelete})
		}
		return
	}

	h.WriteNotFound(w, "endpoint")
}

// extractTaskIDFromCommentPath extracts task ID from path like /api/v1/tasks/{tid}/comments
func extractTaskIDFromCommentPath(path string) string {
	prefix := "/api/v1/tasks/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// extractCommentID extracts comment ID from path like /api/v1/comments/{id}
func extractCommentID(path string) string {
	prefix := "/api/v1/comments/"
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}
