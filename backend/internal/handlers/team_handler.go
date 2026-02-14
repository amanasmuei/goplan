package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

type TeamHandler struct {
	teamRepo    *repository.TeamRepository
	projectRepo *repository.ProjectRepository
}

func NewTeamHandler(teamRepo *repository.TeamRepository, projectRepo *repository.ProjectRepository) *TeamHandler {
	return &TeamHandler{
		teamRepo:    teamRepo,
		projectRepo: projectRepo,
	}
}

// @Summary Create a new team
// @Description Create a new team in the organization
// @Tags teams
// @Accept json
// @Produce json
// @Param request body models.CreateTeamRequest true "Team creation request"
// @Success 201 {object} models.TeamResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/teams [post]
func (h *TeamHandler) CreateTeam(c *fiber.Ctx) error {
	var req models.CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team := &models.Team{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: orgID,
		CreatedBy:      userID,
	}

	if err := h.teamRepo.Create(c.Context(), team); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	memberCount, _ := h.teamRepo.CountMembers(c.Context(), team.ID)

	return c.Status(fiber.StatusCreated).JSON(models.TeamResponse{
		Team:        team,
		MemberCount: memberCount,
	})
}

// @Summary List teams
// @Description List all teams in the organization
// @Tags teams
// @Produce json
// @Success 200 {object} models.TeamListResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/teams [get]
func (h *TeamHandler) ListTeams(c *fiber.Ctx) error {
	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)
	teams, err := h.teamRepo.ListByOrganization(c.Context(), orgID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	var response []models.TeamResponse
	for _, team := range teams {
		memberCount, _ := h.teamRepo.CountMembers(c.Context(), team.ID)
		t := team // create copy to avoid pointer issues
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

// @Summary Get team by ID
// @Description Get team details including members
// @Tags teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} models.TeamResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/teams/{id} [get]
func (h *TeamHandler) GetTeam(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	// Verify team belongs to organization
	if team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	members, _ := h.teamRepo.GetMembers(c.Context(), id)

	return c.JSON(models.TeamResponse{
		Team:        team,
		MemberCount: len(members),
		Members:     members,
	})
}

// @Summary Update team
// @Description Update team details
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body models.UpdateTeamRequest true "Team update request"
// @Success 200 {object} models.TeamResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/teams/{id} [put]
func (h *TeamHandler) UpdateTeam(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	var req models.UpdateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil || team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	// Check if user has permission to manage team
	role, err := h.teamRepo.GetMemberRole(c.Context(), id, userID)
	if err != nil || !role.CanManageTeam() {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "You don't have permission to update this team"})
	}

	if req.Name != nil {
		team.Name = *req.Name
	}
	if req.Description != nil {
		team.Description = *req.Description
	}

	if err := h.teamRepo.Update(c.Context(), team); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	memberCount, _ := h.teamRepo.CountMembers(c.Context(), id)

	return c.JSON(models.TeamResponse{
		Team:        team,
		MemberCount: memberCount,
	})
}

// @Summary Delete team
// @Description Delete a team
// @Tags teams
// @Param id path string true "Team ID"
// @Success 204
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/teams/{id} [delete]
func (h *TeamHandler) DeleteTeam(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil || team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	// Only owners can delete teams
	role, err := h.teamRepo.GetMemberRole(c.Context(), id, userID)
	if err != nil || role != models.TeamRoleOwner {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Only team owners can delete the team"})
	}

	if err := h.teamRepo.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary Add team member
// @Description Add a user to the team
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param request body models.AddTeamMemberRequest true "Add member request"
// @Success 201 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/teams/{id}/members [post]
func (h *TeamHandler) AddMember(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	var req models.AddTeamMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	if !req.Role.IsValid() {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid role"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil || team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	// Check if user has permission to add members
	role, err := h.teamRepo.GetMemberRole(c.Context(), teamID, userID)
	if err != nil || !role.CanEditMembers() {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "You don't have permission to add members"})
	}

	// Only owners can add other owners
	if req.Role == models.TeamRoleOwner && role != models.TeamRoleOwner {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Only owners can add other owners"})
	}

	if err := h.teamRepo.AddMember(c.Context(), teamID, req.UserID, req.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse{Message: "Member added successfully"})
}

// @Summary List team members
// @Description List all members of a team
// @Tags teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} models.TeamMemberListResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/teams/{id}/members [get]
func (h *TeamHandler) ListMembers(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil || team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	members, err := h.teamRepo.GetMembers(c.Context(), teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.JSON(models.TeamMemberListResponse{
		Members: members,
		Total:   int64(len(members)),
	})
}

// @Summary Update member role
// @Description Update a team member's role
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Param request body models.UpdateMemberRoleRequest true "Update role request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/teams/{id}/members/{userId} [put]
func (h *TeamHandler) UpdateMemberRole(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	memberUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid user ID"})
	}

	var req models.UpdateMemberRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}

	if !req.Role.IsValid() {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid role"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil || team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	// Check if user has permission
	currentRole, err := h.teamRepo.GetMemberRole(c.Context(), teamID, userID)
	if err != nil || !currentRole.CanEditMembers() {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "You don't have permission to update member roles"})
	}

	// Check current role of target member
	targetRole, err := h.teamRepo.GetMemberRole(c.Context(), teamID, memberUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Member not found in team"})
	}

	// Only owners can modify owner roles
	if (targetRole == models.TeamRoleOwner || req.Role == models.TeamRoleOwner) && currentRole != models.TeamRoleOwner {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Only owners can modify owner roles"})
	}

	// Prevent removing the last owner
	if targetRole == models.TeamRoleOwner && req.Role != models.TeamRoleOwner {
		ownerCount, _ := h.teamRepo.CountOwners(c.Context(), teamID)
		if ownerCount <= 1 {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Cannot demote the last owner"})
		}
	}

	if err := h.teamRepo.UpdateMemberRole(c.Context(), teamID, memberUserID, req.Role); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.JSON(SuccessResponse{Message: "Role updated successfully"})
}

// @Summary Remove team member
// @Description Remove a user from the team
// @Tags teams
// @Param id path string true "Team ID"
// @Param userId path string true "User ID"
// @Success 204
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/teams/{id}/members/{userId} [delete]
func (h *TeamHandler) RemoveMember(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	memberUserID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid user ID"})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil || team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	// Users can remove themselves, or managers/owners can remove others
	if memberUserID != userID {
		role, err := h.teamRepo.GetMemberRole(c.Context(), teamID, userID)
		if err != nil || !role.CanEditMembers() {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "You don't have permission to remove members"})
		}

		// Check target member's role
		targetRole, _ := h.teamRepo.GetMemberRole(c.Context(), teamID, memberUserID)
		if targetRole == models.TeamRoleOwner && role != models.TeamRoleOwner {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Only owners can remove other owners"})
		}
	}

	// Check if removing the last owner
	targetRole, _ := h.teamRepo.GetMemberRole(c.Context(), teamID, memberUserID)
	if targetRole == models.TeamRoleOwner {
		ownerCount, _ := h.teamRepo.CountOwners(c.Context(), teamID)
		if ownerCount <= 1 {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Cannot remove the last owner"})
		}
	}

	if err := h.teamRepo.RemoveMember(c.Context(), teamID, memberUserID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// @Summary List team projects
// @Description List all projects assigned to this team
// @Tags teams
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} models.ProjectListResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/teams/{id}/projects [get]
func (h *TeamHandler) ListTeamProjects(c *fiber.Ctx) error {
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid team ID"})
	}

	_, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	team, err := h.teamRepo.GetByID(c.Context(), teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}
	if team == nil || team.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Team not found"})
	}

	projLimit := c.QueryInt("limit", 100)
	projOffset := c.QueryInt("offset", 0)
	projects, err := h.projectRepo.ListByTeam(c.Context(), teamID, projLimit, projOffset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "an internal error occurred"})
	}

	var response []models.ProjectResponse
	for _, project := range projects {
		taskCount, _ := h.projectRepo.CountTasks(c.Context(), project.ID)
		p := project
		response = append(response, models.ProjectResponse{
			Project:   &p,
			TaskCount: taskCount,
		})
	}

	return c.JSON(models.ProjectListResponse{
		Projects:   response,
		Total:      int64(len(projects)),
		Page:       1,
		PageSize:   len(projects),
		TotalPages: 1,
	})
}
