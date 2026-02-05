package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

type JustificationHandler struct {
	justificationRepo *repository.JustificationRepository
	taskRepo          *repository.TaskRepository
}

func NewJustificationHandler(justificationRepo *repository.JustificationRepository, taskRepo *repository.TaskRepository) *JustificationHandler {
	return &JustificationHandler{justificationRepo: justificationRepo, taskRepo: taskRepo}
}

// @Summary Submit justification
// @Description Submit justification when no links exist
// @Tags justifications
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body models.CreateJustificationRequest true "Justification request"
// @Success 201 {object} models.TaskJustification
// @Router /api/v1/tasks/{id}/justify [post]
func (h *JustificationHandler) CreateJustification(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	var req models.CreateJustificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	// Validate all checkboxes are true
	if !req.CheckedSameProject || !req.CheckedSameStakeholders || !req.CheckedSameDependencies {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "All confirmation checkboxes must be checked"})
	}

	if len(req.JustificationText) < 50 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Justification must be at least 50 characters"})
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

	// Check if justification already exists
	existing, err := h.justificationRepo.GetByTaskID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}
	if existing != nil {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: "Justification already submitted for this task"})
	}

	justification := &models.TaskJustification{
		TaskID:                  taskID,
		CheckedSameProject:      req.CheckedSameProject,
		CheckedSameStakeholders: req.CheckedSameStakeholders,
		CheckedSameDependencies: req.CheckedSameDependencies,
		JustificationText:       req.JustificationText,
		CreatedBy:               userID,
	}

	if err := h.justificationRepo.Create(c.Context(), justification); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(justification)
}

// @Summary Get task justification
// @Description Get justification for a task
// @Tags justifications
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.TaskJustification
// @Router /api/v1/tasks/{id}/justify [get]
func (h *JustificationHandler) GetJustification(c *fiber.Ctx) error {
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

	justification, err := h.justificationRepo.GetByTaskID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}
	if justification == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "No justification found"})
	}

	return c.JSON(justification)
}
