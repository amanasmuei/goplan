package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/services"
)

type TaskHandler struct {
	taskService *services.TaskService
}

func NewTaskHandler(taskService *services.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// @Summary Create a new task
// @Description Create a new task with similarity search and predictions
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body models.CreateTaskRequest true "Task creation request"
// @Success 201 {object} models.TaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/tasks [post]
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	var req models.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	response, err := h.taskService.CreateTask(c.Context(), req, userID, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// @Summary Get task by ID
// @Description Get detailed task information including context
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} models.Task
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	task, err := h.taskService.GetTask(c.Context(), id, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}
	if task == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Task not found"})
	}

	return c.JSON(task)
}

// @Summary Update task
// @Description Update task details
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body models.UpdateTaskRequest true "Task update request"
// @Success 200 {object} models.Task
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	var req models.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	task, err := h.taskService.UpdateTask(c.Context(), id, req, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(task)
}

// @Summary Delete task
// @Description Soft delete a task
// @Tags tasks
// @Param id path string true "Task ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	if err := h.taskService.DeleteTask(c.Context(), id, orgID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary List tasks
// @Description List tasks with filtering and pagination
// @Tags tasks
// @Produce json
// @Param project_id query string false "Filter by project ID"
// @Param status query string false "Filter by status"
// @Param assigned_to query string false "Filter by assignee"
// @Param search query string false "Search in title/description"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(50)
// @Success 200 {object} models.TaskListResponse
// @Router /api/v1/tasks [get]
func (h *TaskHandler) ListTasks(c *fiber.Ctx) error {
	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	filters := models.TaskFilters{
		OrganizationID: orgID,
		Page:           c.QueryInt("page", 1),
		PageSize:       c.QueryInt("page_size", 50),
		Search:         c.Query("search"),
	}

	if projectID := c.Query("project_id"); projectID != "" {
		if id, err := uuid.Parse(projectID); err == nil {
			filters.ProjectID = &id
		}
	}
	if status := c.Query("status"); status != "" {
		s := models.TaskStatus(status)
		filters.Status = &s
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		if id, err := uuid.Parse(assignedTo); err == nil {
			filters.AssignedTo = &id
		}
	}

	response, err := h.taskService.ListTasks(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(response)
}

// @Summary Get similar tasks
// @Description Get tasks similar to the specified task
// @Tags tasks
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {array} models.SimilarTask
// @Router /api/v1/tasks/{id}/similar [get]
func (h *TaskHandler) GetSimilarTasks(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	similar, err := h.taskService.GetSimilarTasks(c.Context(), id, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(fiber.Map{"similar_tasks": similar})
}

// @Summary Acknowledge task predictions
// @Description Acknowledge system predictions with accept/modify/disagree options
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param request body models.AcknowledgmentRequest false "Acknowledgment options"
// @Success 200 {object} models.Task
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/tasks/{id}/acknowledge [post]
func (h *TaskHandler) AcknowledgeTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	// Parse optional acknowledgment request
	var req *models.AcknowledgmentRequest
	if len(c.Body()) > 0 {
		req = &models.AcknowledgmentRequest{}
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
		}
	}

	task, err := h.taskService.AcknowledgeTask(c.Context(), id, userID, orgID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(task)
}

// @Summary Start task
// @Description Mark task as started
// @Tags tasks
// @Param id path string true "Task ID"
// @Success 200 {object} models.Task
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/tasks/{id}/start [post]
func (h *TaskHandler) StartTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	task, err := h.taskService.StartTask(c.Context(), id, userID, orgID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(task)
}

// @Summary Complete task
// @Description Initiate task completion flow
// @Tags tasks
// @Param id path string true "Task ID"
// @Success 200 {object} models.Task
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/tasks/{id}/complete [post]
func (h *TaskHandler) CompleteTask(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid task ID"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	task, err := h.taskService.CompleteTask(c.Context(), id, userID, orgID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(task)
}
