package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
	teamRepo    *repository.TeamRepository
}

func NewProjectHandler(projectRepo *repository.ProjectRepository, teamRepo *repository.TeamRepository) *ProjectHandler {
	return &ProjectHandler{
		projectRepo: projectRepo,
		teamRepo:    teamRepo,
	}
}

// @Summary Create a new project
// @Description Create a new project in the organization
// @Tags projects
// @Accept json
// @Produce json
// @Param request body models.CreateProjectRequest true "Project creation request"
// @Success 201 {object} models.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	var req models.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project := &models.Project{
		Name:           req.Name,
		Description:    req.Description,
		Status:         models.ProjectStatusActive,
		OrganizationID: orgID,
		CreatedBy:      &userID,
	}

	if err := h.projectRepo.Create(c.Context(), project); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	// Assign teams if specified
	if len(req.TeamIDs) > 0 {
		if err := h.projectRepo.AssignTeams(c.Context(), project.ID, req.TeamIDs); err != nil {
			// Log error but don't fail the request
			// The project was created successfully
		}
	}

	teams, _ := h.projectRepo.GetTeams(c.Context(), project.ID)

	return c.Status(fiber.StatusCreated).JSON(models.ProjectResponse{
		Project:   project,
		Teams:     teams,
		TaskCount: 0,
	})
}

// @Summary List projects
// @Description List all projects in the organization with filtering
// @Tags projects
// @Produce json
// @Param status query string false "Filter by status (active, archived)"
// @Param team_id query string false "Filter by team ID"
// @Param search query string false "Search in name/description"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} models.ProjectListResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	pageSize := c.QueryInt("page_size", 20)
	if pageSize > 200 {
		pageSize = 200
	}
	if pageSize < 1 {
		pageSize = 10
	}

	filters := models.ProjectFilters{
		OrganizationID: orgID,
		Page:           c.QueryInt("page", 1),
		PageSize:       pageSize,
		Search:         c.Query("search"),
	}

	if status := c.Query("status"); status != "" {
		s := models.ProjectStatus(status)
		if s.IsValid() {
			filters.Status = &s
		}
	}

	if teamID := c.Query("team_id"); teamID != "" {
		if id, err := uuid.Parse(teamID); err == nil {
			filters.TeamID = &id
		}
	}

	projects, total, err := h.projectRepo.ListByOrganization(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	var response []models.ProjectResponse
	for _, project := range projects {
		teams, _ := h.projectRepo.GetTeams(c.Context(), project.ID)
		taskCount, _ := h.projectRepo.CountTasks(c.Context(), project.ID)
		p := project
		response = append(response, models.ProjectResponse{
			Project:   &p,
			Teams:     teams,
			TaskCount: taskCount,
		})
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	return c.JSON(models.ProjectListResponse{
		Projects:   response,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	})
}

// @Summary Get project by ID
// @Description Get project details including assigned teams
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.ProjectResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid project ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project, err := h.projectRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if project == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	// Verify project belongs to organization
	if project.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	teams, _ := h.projectRepo.GetTeams(c.Context(), id)
	taskCount, _ := h.projectRepo.CountTasks(c.Context(), id)

	return c.JSON(models.ProjectResponse{
		Project:   project,
		Teams:     teams,
		TaskCount: taskCount,
	})
}

// @Summary Update project
// @Description Update project details
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body models.UpdateProjectRequest true "Project update request"
// @Success 200 {object} models.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid project ID"})
	}

	var req models.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project, err := h.projectRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if project == nil || project.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Status != nil && req.Status.IsValid() {
		project.Status = *req.Status
	}

	if err := h.projectRepo.Update(c.Context(), project); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	teams, _ := h.projectRepo.GetTeams(c.Context(), id)
	taskCount, _ := h.projectRepo.CountTasks(c.Context(), id)

	return c.JSON(models.ProjectResponse{
		Project:   project,
		Teams:     teams,
		TaskCount: taskCount,
	})
}

// @Summary Delete project
// @Description Delete a project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid project ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project, err := h.projectRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if project == nil || project.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	// Check if project has tasks
	taskCount, _ := h.projectRepo.CountTasks(c.Context(), id)
	if taskCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Cannot delete project with existing tasks. Archive it instead."})
	}

	if err := h.projectRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary Archive project
// @Description Archive a project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 200 {object} models.ProjectResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/projects/{id}/archive [post]
func (h *ProjectHandler) ArchiveProject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid project ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project, err := h.projectRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if project == nil || project.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	if err := h.projectRepo.Archive(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	project.Status = models.ProjectStatusArchived
	teams, _ := h.projectRepo.GetTeams(c.Context(), id)
	taskCount, _ := h.projectRepo.CountTasks(c.Context(), id)

	return c.JSON(models.ProjectResponse{
		Project:   project,
		Teams:     teams,
		TaskCount: taskCount,
	})
}

// @Summary Assign teams to project
// @Description Assign one or more teams to a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body models.AssignTeamsRequest true "Assign teams request"
// @Success 200 {object} models.ProjectResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/projects/{id}/teams [post]
func (h *ProjectHandler) AssignTeams(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid project ID"})
	}

	var req models.AssignTeamsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	if len(req.TeamIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "At least one team ID is required"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project, err := h.projectRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if project == nil || project.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	// Verify all teams belong to the same organization
	for _, teamID := range req.TeamIDs {
		team, err := h.teamRepo.GetByID(c.Context(), teamID)
		if err != nil || team == nil || team.OrganizationID != orgID {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID: " + teamID.String()})
		}
	}

	if err := h.projectRepo.AssignTeams(c.Context(), id, req.TeamIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	teams, _ := h.projectRepo.GetTeams(c.Context(), id)
	taskCount, _ := h.projectRepo.CountTasks(c.Context(), id)

	return c.JSON(models.ProjectResponse{
		Project:   project,
		Teams:     teams,
		TaskCount: taskCount,
	})
}

// @Summary Remove team from project
// @Description Remove a team assignment from a project
// @Tags projects
// @Param id path string true "Project ID"
// @Param teamId path string true "Team ID"
// @Success 204
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/projects/{id}/teams/{teamId} [delete]
func (h *ProjectHandler) RemoveTeam(c *fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid project ID"})
	}

	teamID, err := uuid.Parse(c.Params("teamId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project, err := h.projectRepo.GetByID(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if project == nil || project.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	if err := h.projectRepo.RemoveTeam(c.Context(), projectID, teamID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary Get project teams
// @Description Get all teams assigned to a project
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.TeamListResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/projects/{id}/teams [get]
func (h *ProjectHandler) GetProjectTeams(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid project ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	project, err := h.projectRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if project == nil || project.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Project not found"})
	}

	teams, err := h.projectRepo.GetTeams(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	var response []models.TeamResponse
	for _, team := range teams {
		memberCount, _ := h.teamRepo.CountMembers(c.Context(), team.ID)
		t := team
		response = append(response, models.TeamResponse{
			Team:        &t,
			MemberCount: memberCount,
		})
	}

	return c.JSON(models.TeamListResponse{
		Teams: response,
		Total: int64(len(teams)),
	})
}
