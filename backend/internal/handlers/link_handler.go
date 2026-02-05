package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

type LinkHandler struct {
	linkRepo *repository.TaskLinkRepository
	taskRepo *repository.TaskRepository
}

func NewLinkHandler(linkRepo *repository.TaskLinkRepository, taskRepo *repository.TaskRepository) *LinkHandler {
	return &LinkHandler{linkRepo: linkRepo, taskRepo: taskRepo}
}

// @Summary Create task link
// @Description Link a task to another historical task
// @Tags links
// @Accept json
// @Produce json
// @Param id path string true "Source Task ID"
// @Param request body models.CreateTaskLinkRequest true "Link creation request"
// @Success 201 {object} models.TaskLink
// @Router /api/v1/tasks/{id}/links [post]
func (h *LinkHandler) CreateLink(c *fiber.Ctx) error {
	sourceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	var req models.CreateTaskLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	// Verify source task exists and belongs to org
	sourceTask, err := h.taskRepo.GetByID(c.Context(), sourceID)
	if err != nil || sourceTask == nil || sourceTask.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Source task not found"})
	}

	// Verify target task exists and belongs to org
	targetTask, err := h.taskRepo.GetByID(c.Context(), req.TargetTaskID)
	if err != nil || targetTask == nil || targetTask.OrganizationID != orgID {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Target task not found"})
	}

	// Check for duplicate link
	exists, err := h.linkRepo.Exists(c.Context(), sourceID, req.TargetTaskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: "Link already exists"})
	}

	// Check for circular dependency when creating a dependent link
	if req.LinkType == models.LinkTypeDependent {
		// Prevent self-dependency
		if sourceID == req.TargetTaskID {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "A task cannot depend on itself"})
		}

		// Check if adding this dependency would create a cycle
		hasCircular, err := h.linkRepo.HasCircularDependency(c.Context(), sourceID, req.TargetTaskID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
		}
		if hasCircular {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Cannot create dependency: would create a circular dependency"})
		}
	}

	link := &models.TaskLink{
		SourceTaskID: sourceID,
		TargetTaskID: req.TargetTaskID,
		LinkType:     req.LinkType,
		CreatedBy:    userID,
		Notes:        req.Notes,
	}

	if err := h.linkRepo.Create(c.Context(), link); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(link)
}

// @Summary List task links
// @Description Get all links for a task
// @Tags links
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {array} models.TaskLink
// @Router /api/v1/tasks/{id}/links [get]
func (h *LinkHandler) ListLinks(c *fiber.Ctx) error {
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

	links, err := h.linkRepo.ListByTaskID(c.Context(), taskID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(fiber.Map{"links": links})
}

// @Summary Delete task link
// @Description Remove a link between tasks
// @Tags links
// @Param id path string true "Task ID"
// @Param linkId path string true "Link ID"
// @Success 204
// @Router /api/v1/tasks/{id}/links/{linkId} [delete]
func (h *LinkHandler) DeleteLink(c *fiber.Ctx) error {
	linkID, err := uuid.Parse(c.Params("linkId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid link ID"})
	}

	_, _, err = getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	if err := h.linkRepo.Delete(c.Context(), linkID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
