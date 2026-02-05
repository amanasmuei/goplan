package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

type BlockerHandler struct {
	blockerRepo *repository.BlockerRepository
	taskRepo    *repository.TaskRepository
}

func NewBlockerHandler(blockerRepo *repository.BlockerRepository, taskRepo *repository.TaskRepository) *BlockerHandler {
	return &BlockerHandler{blockerRepo: blockerRepo, taskRepo: taskRepo}
}

// @Summary Log a blocker
// @Description Log a blocker for a task
// @Tags blockers
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body models.CreateBlockerRequest true "Blocker creation request"
// @Success 201 {object} models.TaskBlocker
// @Router /api/v1/tasks/{id}/blockers [post]
func (h *BlockerHandler) CreateBlocker(c *fiber.Ctx) error {
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	var req models.CreateBlockerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
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

	// Update task status to blocked
	task.Status = models.TaskStatusBlocked
	if err := h.taskRepo.Update(c.Context(), task); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	blocker := &models.TaskBlocker{
		TaskID:      taskID,
		BlockerType: req.BlockerType,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := h.blockerRepo.Create(c.Context(), blocker); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(blocker)
}

// @Summary List task blockers
// @Description Get all blockers for a task
// @Tags blockers
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {array} models.TaskBlocker
// @Router /api/v1/tasks/{id}/blockers [get]
func (h *BlockerHandler) ListBlockers(c *fiber.Ctx) error {
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

	blockers, err := h.blockerRepo.ListByTaskID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(fiber.Map{"blockers": blockers})
}

// @Summary Resolve blocker
// @Description Mark a blocker as resolved
// @Tags blockers
// @Accept json
// @Produce json
// @Param id path string true "Blocker ID"
// @Param request body models.ResolveBlockerRequest true "Resolve request"
// @Success 200 {object} models.TaskBlocker
// @Router /api/v1/blockers/{id}/resolve [put]
func (h *BlockerHandler) ResolveBlocker(c *fiber.Ctx) error {
	blockerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid blocker ID"})
	}

	var req models.ResolveBlockerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	_, _, err = getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	blocker, err := h.blockerRepo.GetByID(c.Context(), blockerID)
	if err != nil || blocker == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Blocker not found"})
	}

	if err := h.blockerRepo.Resolve(c.Context(), blockerID, req.DaysBlocked); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	// Get task and check if there are other unresolved blockers
	blockers, err := h.blockerRepo.ListByTaskID(c.Context(), blocker.TaskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	hasUnresolved := false
	for _, b := range blockers {
		if b.ID != blockerID && b.ResolvedAt == nil {
			hasUnresolved = true
			break
		}
	}

	// If no more unresolved blockers, set task back to active
	if !hasUnresolved {
		task, err := h.taskRepo.GetByID(c.Context(), blocker.TaskID)
		if err == nil && task != nil && task.Status == models.TaskStatusBlocked {
			task.Status = models.TaskStatusActive
			h.taskRepo.Update(c.Context(), task)
		}
	}

	// Refresh blocker data
	blocker, _ = h.blockerRepo.GetByID(c.Context(), blockerID)
	return c.JSON(blocker)
}
