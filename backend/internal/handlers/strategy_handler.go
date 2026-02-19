package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/goplan/backend/internal/middleware"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
	"github.com/goplan/backend/internal/services"
)

// StrategyHandler handles HTTP requests for strategic plan operations.
type StrategyHandler struct {
	strategyService *services.StrategyService
	versionRepo     *repository.VersionRepository
	sectionRepo     *repository.SectionRepository
	planRepo        *repository.PlanRepository
}

// NewStrategyHandler creates a new StrategyHandler.
func NewStrategyHandler(
	strategyService *services.StrategyService,
	versionRepo *repository.VersionRepository,
	sectionRepo *repository.SectionRepository,
	planRepo *repository.PlanRepository,
) *StrategyHandler {
	return &StrategyHandler{
		strategyService: strategyService,
		versionRepo:     versionRepo,
		sectionRepo:     sectionRepo,
		planRepo:        planRepo,
	}
}

// CreateStrategy creates a new AI-generated strategic plan.
// @Summary Create a strategic plan
// @Description Generate a new AI-powered strategic plan from user input
// @Tags Strategy
// @Accept json
// @Produce json
// @Param request body models.CreatePlanRequest true "Plan creation request"
// @Success 201 {object} models.PlanResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies [post]
func (h *StrategyHandler) CreateStrategy(c *fiber.Ctx) error {
	var req models.CreatePlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}
	if err := middleware.Validate(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
	}

	userID, orgID, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	response, err := h.strategyService.GenerateStrategy(c.Context(), userID, orgID, req)
	if err != nil {
		if strings.Contains(err.Error(), "plan limit reached") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to generate strategy"})
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// ListStrategies lists strategic plans with filtering and pagination.
// @Summary List strategic plans
// @Description List strategic plans with optional filtering by status, category, and search
// @Tags Strategy
// @Produce json
// @Param status query string false "Filter by status (draft, generating, complete, archived)"
// @Param category query string false "Filter by category"
// @Param search query string false "Search in title"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.PlanListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies [get]
func (h *StrategyHandler) ListStrategies(c *fiber.Ctx) error {
	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	pageSize := c.QueryInt("page_size", 10)
	if pageSize > 50 {
		pageSize = 50
	}
	if pageSize < 1 {
		pageSize = 10
	}

	filters := models.PlanFilters{
		UserID:   &userID,
		Page:     c.QueryInt("page", 1),
		PageSize: pageSize,
		Search:   c.Query("search"),
	}

	if status := c.Query("status"); status != "" {
		s := models.PlanStatus(status)
		filters.Status = &s
	}
	if category := c.Query("category"); category != "" {
		cat := models.PlanCategory(category)
		filters.Category = &cat
	}

	response, err := h.strategyService.ListPlans(c.Context(), filters)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to list strategies"})
	}

	return c.JSON(response)
}

// GetStrategy retrieves a single strategic plan by ID.
// @Summary Get a strategic plan
// @Description Get detailed information about a specific strategic plan
// @Tags Strategy
// @Produce json
// @Param id path string true "Plan ID (UUID)"
// @Success 200 {object} models.PlanResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id} [get]
func (h *StrategyHandler) GetStrategy(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	response, err := h.strategyService.GetPlan(c.Context(), planID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Strategy not found"})
		}
		if strings.Contains(err.Error(), "unauthorized") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get strategy"})
	}

	return c.JSON(response)
}

// ArchiveStrategy archives a strategic plan (soft delete).
// @Summary Archive a strategic plan
// @Description Archive (soft delete) a strategic plan
// @Tags Strategy
// @Param id path string true "Plan ID (UUID)"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id} [delete]
func (h *StrategyHandler) ArchiveStrategy(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	if err := h.strategyService.ArchivePlan(c.Context(), planID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Strategy not found"})
		}
		if strings.Contains(err.Error(), "unauthorized") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to archive strategy"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RegenerateSection regenerates a specific section with fresh AI output.
// @Summary Regenerate a plan section
// @Description Regenerate a specific section of a strategic plan using AI
// @Tags Strategy
// @Accept json
// @Produce json
// @Param id path string true "Plan ID (UUID)"
// @Param type path string true "Section type (executive_brief, strategic_context, recommended_approach, phased_execution, immediate_action)"
// @Param request body models.RegenerateSectionRequest false "Regeneration options"
// @Success 200 {object} models.PlanSection
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id}/sections/{type}/regenerate [post]
func (h *StrategyHandler) RegenerateSection(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	sectionType := models.SectionType(c.Params("type"))
	if _, ok := models.SectionOrder[sectionType]; !ok {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid section type"})
	}

	var req models.RegenerateSectionRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
		}
		if err := middleware.Validate(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		}
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	section, err := h.strategyService.RegenerateSection(c.Context(), planID, sectionType, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: err.Error()})
		}
		if strings.Contains(err.Error(), "unauthorized") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
		}
		if strings.Contains(err.Error(), "limit reached") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to regenerate section"})
	}

	return c.JSON(section)
}

// RefineSection refines a section based on user feedback.
// @Summary Refine a plan section
// @Description Refine a specific section with targeted user feedback
// @Tags Strategy
// @Accept json
// @Produce json
// @Param id path string true "Plan ID (UUID)"
// @Param type path string true "Section type"
// @Param request body models.RefineSectionRequest true "Refinement request"
// @Success 200 {object} models.PlanSection
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id}/sections/{type}/refine [post]
func (h *StrategyHandler) RefineSection(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	sectionType := models.SectionType(c.Params("type"))
	if _, ok := models.SectionOrder[sectionType]; !ok {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid section type"})
	}

	var req models.RefineSectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid request body"})
	}
	if err := middleware.Validate(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	section, err := h.strategyService.RefineSection(c.Context(), planID, sectionType, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: err.Error()})
		}
		if strings.Contains(err.Error(), "unauthorized") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
		}
		if strings.Contains(err.Error(), "limit reached") || strings.Contains(err.Error(), "requires a") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to refine section"})
	}

	return c.JSON(section)
}

// ListVersions lists all versions of a strategic plan.
// @Summary List plan versions
// @Description Get the version history of a strategic plan
// @Tags Strategy
// @Produce json
// @Param id path string true "Plan ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id}/versions [get]
func (h *StrategyHandler) ListVersions(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	// Verify ownership
	plan, err := h.planRepo.GetByID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get plan"})
	}
	if plan == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Strategy not found"})
	}
	if plan.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
	}

	versions, err := h.versionRepo.ListPlanVersions(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to list versions"})
	}

	return c.JSON(fiber.Map{"versions": versions})
}

// GetVersion retrieves a specific version of a strategic plan.
// @Summary Get a plan version
// @Description Get a specific version snapshot of a strategic plan
// @Tags Strategy
// @Produce json
// @Param id path string true "Plan ID (UUID)"
// @Param version path int true "Version number"
// @Success 200 {object} models.PlanVersion
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id}/versions/{version} [get]
func (h *StrategyHandler) GetVersion(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	version, err := strconv.Atoi(c.Params("version"))
	if err != nil || version < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid version number"})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	// Verify ownership
	plan, err := h.planRepo.GetByID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get plan"})
	}
	if plan == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Strategy not found"})
	}
	if plan.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
	}

	pv, err := h.versionRepo.GetPlanVersion(c.Context(), planID, version)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get version"})
	}
	if pv == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Version not found"})
	}

	return c.JSON(pv)
}

// ListSectionVersions lists all versions of a specific section.
// @Summary List section versions
// @Description Get the version history of a specific plan section
// @Tags Strategy
// @Produce json
// @Param id path string true "Plan ID (UUID)"
// @Param type path string true "Section type"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id}/sections/{type}/versions [get]
func (h *StrategyHandler) ListSectionVersions(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	sectionType := models.SectionType(c.Params("type"))
	if _, ok := models.SectionOrder[sectionType]; !ok {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid section type"})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	// Verify ownership
	plan, err := h.planRepo.GetByID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get plan"})
	}
	if plan == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Strategy not found"})
	}
	if plan.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
	}

	// Get the section to find its ID
	section, err := h.sectionRepo.GetByPlanAndType(c.Context(), planID, sectionType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get section"})
	}
	if section == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: fmt.Sprintf("Section %s not found", sectionType)})
	}

	versions, err := h.versionRepo.ListSectionVersions(c.Context(), section.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to list section versions"})
	}

	return c.JSON(fiber.Map{"versions": versions})
}

// GetSimilarStrategies finds strategies similar to a given plan.
// @Summary Get similar strategies
// @Description Find strategies similar to the specified plan using vector similarity
// @Tags Strategy
// @Produce json
// @Param id path string true "Plan ID (UUID)"
// @Param limit query int false "Max results" default(5)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id}/similar [get]
func (h *StrategyHandler) GetSimilarStrategies(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	limit := c.QueryInt("limit", 5)
	if limit > 20 {
		limit = 20
	}
	if limit < 1 {
		limit = 5
	}

	// Get the plan to verify ownership and get its embedding
	plan, err := h.planRepo.GetByID(c.Context(), planID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get plan"})
	}
	if plan == nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Strategy not found"})
	}
	if plan.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
	}

	// Check if plan has an embedding
	if len(plan.ContentEmbedding.Slice()) == 0 {
		return c.JSON(fiber.Map{"similar": []interface{}{}})
	}

	similar, err := h.planRepo.FindSimilar(c.Context(), plan.ContentEmbedding, plan.OrganizationID, &planID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to find similar strategies"})
	}

	return c.JSON(fiber.Map{"similar": similar})
}

// ExportStrategy exports a strategic plan in the requested format.
// @Summary Export a strategic plan
// @Description Export a strategic plan as markdown or JSON
// @Tags Strategy
// @Produce octet-stream
// @Param id path string true "Plan ID (UUID)"
// @Param format query string false "Export format (markdown, json)" default(markdown)
// @Success 200 {file} file
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/strategies/{id}/export [get]
func (h *StrategyHandler) ExportStrategy(c *fiber.Ctx) error {
	planID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid plan ID"})
	}

	format := c.Query("format", "markdown")
	if format != "markdown" && format != "json" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Unsupported export format. Use 'markdown' or 'json'"})
	}

	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	data, contentType, err := h.strategyService.ExportPlan(c.Context(), planID, userID, format)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "Strategy not found"})
		}
		if strings.Contains(err.Error(), "unauthorized") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "Access denied"})
		}
		if strings.Contains(err.Error(), "requires a") {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to export strategy"})
	}

	ext := "md"
	if format == "json" {
		ext = "json"
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"strategy-%s.%s\"", planID.String()[:8], ext))
	return c.Send(data)
}
