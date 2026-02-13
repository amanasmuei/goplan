package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

type ReviewHandler struct {
	reviewRepo *repository.ReviewRepository
	taskRepo   *repository.TaskRepository
}

func NewReviewHandler(reviewRepo *repository.ReviewRepository, taskRepo *repository.TaskRepository) *ReviewHandler {
	return &ReviewHandler{reviewRepo: reviewRepo, taskRepo: taskRepo}
}

// @Summary Submit completion review
// @Description Submit mandatory review when completing a task
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body models.CreateReviewRequest true "Review request"
// @Success 201 {object} models.TaskReview
// @Router /api/v1/tasks/{id}/review [post]
func (h *ReviewHandler) CreateReview(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	var req models.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	if req.PredictionAccuracyRating < 1 || req.PredictionAccuracyRating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Rating must be between 1 and 5"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	// Verify task exists and belongs to org
	task, err := h.taskRepo.GetByID(c.Context(), taskID)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Task not found"})
	}

	// Task must be in pending_review status
	if task.Status != models.TaskStatusPendingReview {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Task must be in pending review status"})
	}

	// Check if review already exists
	existing, err := h.reviewRepo.GetByTaskID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: "Review already submitted for this task"})
	}

	review := &models.TaskReview{
		TaskID:                   taskID,
		PredictionAccuracyRating: req.PredictionAccuracyRating,
		PredictionFeedback:       req.PredictionFeedback,
		LessonsLearned:           req.LessonsLearned,
		WouldApproachDifferently: req.WouldApproachDifferently,
		CreatedBy:                userID,
	}

	if err := h.reviewRepo.Create(c.Context(), review); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	// Update task to completed and calculate actual days
	now := time.Now()
	task.CompletedAt = &now
	task.Status = models.TaskStatusCompleted

	if task.StartedAt != nil {
		actualDays := now.Sub(*task.StartedAt).Hours() / 24
		task.ActualDays = &actualDays
	}

	if err := h.taskRepo.Update(c.Context(), task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.Status(fiber.StatusCreated).JSON(review)
}

// @Summary Get task review
// @Description Get completion review for a task
// @Tags reviews
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.TaskReview
// @Router /api/v1/tasks/{id}/review [get]
func (h *ReviewHandler) GetReview(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	// Verify task belongs to org
	task, err := h.taskRepo.GetByID(c.Context(), taskID)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Task not found"})
	}

	review, err := h.reviewRepo.GetByTaskID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if review == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "No review found"})
	}

	return c.JSON(review)
}
