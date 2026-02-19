package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/goplan/backend/internal/claude"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

// StrategyService handles AI-powered strategic plan generation and management.
type StrategyService struct {
	claudeClient    *claude.Client
	planRepo        *repository.PlanRepository
	sectionRepo     *repository.SectionRepository
	versionRepo     *repository.VersionRepository
	genLogRepo      *repository.GenerationLogRepository
	subRepo         *repository.SubscriptionRepository
	embeddingClient *EmbeddingClient
}

// NewStrategyService creates a new StrategyService with all dependencies.
func NewStrategyService(
	claudeClient *claude.Client,
	planRepo *repository.PlanRepository,
	sectionRepo *repository.SectionRepository,
	versionRepo *repository.VersionRepository,
	genLogRepo *repository.GenerationLogRepository,
	subRepo *repository.SubscriptionRepository,
	embeddingClient *EmbeddingClient,
) *StrategyService {
	return &StrategyService{
		claudeClient:    claudeClient,
		planRepo:        planRepo,
		sectionRepo:     sectionRepo,
		versionRepo:     versionRepo,
		genLogRepo:      genLogRepo,
		subRepo:         subRepo,
		embeddingClient: embeddingClient,
	}
}

// strategyResponse is the expected JSON structure from Claude for full plan generation.
type strategyResponse struct {
	Classification struct {
		Category    string `json:"category"`
		SubCategory string `json:"sub_category"`
		Complexity  string `json:"complexity"`
	} `json:"classification"`
	Title    string `json:"title"`
	Sections struct {
		ExecutiveBrief      json.RawMessage `json:"executive_brief"`
		StrategicContext    json.RawMessage `json:"strategic_context"`
		RecommendedApproach json.RawMessage `json:"recommended_approach"`
		PhasedExecution     json.RawMessage `json:"phased_execution"`
		ImmediateAction     json.RawMessage `json:"immediate_action"`
	} `json:"sections"`
}

// GenerateStrategy creates a new strategic plan from user input using Claude AI.
func (s *StrategyService) GenerateStrategy(ctx context.Context, userID, orgID uuid.UUID, req models.CreatePlanRequest) (*models.PlanResponse, error) {
	// Check subscription plan limit
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription: %w", err)
	}
	if sub == nil {
		defaults := models.TierLimits(models.SubscriptionTierFree)
		sub = &defaults
		sub.UserID = userID
	}

	planCount, err := s.planRepo.CountByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count plans: %w", err)
	}
	if planCount >= sub.MaxPlans {
		return nil, fmt.Errorf("plan limit reached (%d/%d). Upgrade your subscription for more plans", planCount, sub.MaxPlans)
	}

	// Determine category for prompt building
	category := models.PlanCategoryGeneric
	if req.Category != nil {
		category = *req.Category
	}

	// Create plan in draft status
	plan := &models.StrategicPlan{
		UserID:         userID,
		OrganizationID: orgID,
		Title:          "Generating...",
		OriginalInput:  req.Input,
		Category:       category,
		Status:         models.PlanStatusDraft,
	}
	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	// Update to generating status
	plan.Status = models.PlanStatusGenerating
	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update plan status: %w", err)
	}

	// Build prompts
	depth := "standard"
	systemPrompt, userPrompt := BuildStrategyPrompt(req.Input, category, depth)

	// Call Claude
	startTime := time.Now()
	messages := []claude.Message{
		claude.NewTextMessage("user", userPrompt),
	}
	claudeReq := &claude.MessagesRequest{
		Model:       "",
		MaxTokens:   8192,
		System:      systemPrompt,
		Messages:    messages,
		Temperature: 0.7,
	}

	resp, err := s.claudeClient.CreateMessage(ctx, claudeReq)
	if err != nil {
		s.logGenerationError(ctx, &plan.ID, userID, "generate", nil, err, startTime)
		plan.Status = models.PlanStatusDraft
		_ = s.planRepo.Update(ctx, plan)
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}
	duration := time.Since(startTime)

	// Extract JSON from response
	rawJSON, err := claude.ExtractJSON(resp.GetTextContent())
	if err != nil {
		s.logGenerationError(ctx, &plan.ID, userID, "generate", nil, err, startTime)
		plan.Status = models.PlanStatusDraft
		_ = s.planRepo.Update(ctx, plan)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Parse the structured response
	var stratResp strategyResponse
	if err := json.Unmarshal([]byte(rawJSON), &stratResp); err != nil {
		s.logGenerationError(ctx, &plan.ID, userID, "generate", nil, err, startTime)
		plan.Status = models.PlanStatusDraft
		_ = s.planRepo.Update(ctx, plan)
		return nil, fmt.Errorf("failed to parse strategy response: %w", err)
	}

	// Update plan with classification results
	plan.Title = stratResp.Title
	if stratResp.Classification.Category != "" {
		cat := models.PlanCategory(stratResp.Classification.Category)
		plan.Category = cat
	}
	if stratResp.Classification.SubCategory != "" {
		plan.SubCategory = &stratResp.Classification.SubCategory
	}
	if stratResp.Classification.Complexity != "" {
		plan.Complexity = &stratResp.Classification.Complexity
	}
	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to update plan classification: %w", err)
	}

	// Create the 5 sections
	sections := []models.PlanSection{
		{
			PlanID:       plan.ID,
			SectionType:  models.SectionTypeExecutiveBrief,
			SectionOrder: models.SectionOrder[models.SectionTypeExecutiveBrief],
			Title:        models.DefaultSectionTitles[models.SectionTypeExecutiveBrief],
			Content:      stratResp.Sections.ExecutiveBrief,
			Version:      1,
		},
		{
			PlanID:       plan.ID,
			SectionType:  models.SectionTypeStrategicContext,
			SectionOrder: models.SectionOrder[models.SectionTypeStrategicContext],
			Title:        models.DefaultSectionTitles[models.SectionTypeStrategicContext],
			Content:      stratResp.Sections.StrategicContext,
			Version:      1,
		},
		{
			PlanID:       plan.ID,
			SectionType:  models.SectionTypeRecommendedApproach,
			SectionOrder: models.SectionOrder[models.SectionTypeRecommendedApproach],
			Title:        models.DefaultSectionTitles[models.SectionTypeRecommendedApproach],
			Content:      stratResp.Sections.RecommendedApproach,
			Version:      1,
		},
		{
			PlanID:       plan.ID,
			SectionType:  models.SectionTypePhasedExecution,
			SectionOrder: models.SectionOrder[models.SectionTypePhasedExecution],
			Title:        models.DefaultSectionTitles[models.SectionTypePhasedExecution],
			Content:      stratResp.Sections.PhasedExecution,
			Version:      1,
		},
		{
			PlanID:       plan.ID,
			SectionType:  models.SectionTypeImmediateAction,
			SectionOrder: models.SectionOrder[models.SectionTypeImmediateAction],
			Title:        models.DefaultSectionTitles[models.SectionTypeImmediateAction],
			Content:      stratResp.Sections.ImmediateAction,
			Version:      1,
		},
	}

	if err := s.sectionRepo.CreateBatch(ctx, sections); err != nil {
		return nil, fmt.Errorf("failed to create sections: %w", err)
	}

	// Create initial SectionVersion for each section
	for i := range sections {
		sv := &models.SectionVersion{
			SectionID:   sections[i].ID,
			PlanID:      plan.ID,
			Version:     1,
			Content:     sections[i].Content,
			GeneratedBy: "claude",
			TokenUsage: map[string]int{
				"input_tokens":  resp.Usage.InputTokens,
				"output_tokens": resp.Usage.OutputTokens,
			},
		}
		if err := s.versionRepo.CreateSectionVersion(ctx, sv); err != nil {
			slog.Error("failed to create section version", "section_id", sections[i].ID, "error", err)
		}
	}

	// Create initial PlanVersion snapshot
	snapshot := map[string]any{
		"plan":     plan,
		"sections": sections,
	}
	changeSummary := "Initial plan generation"
	pv := &models.PlanVersion{
		PlanID:        plan.ID,
		Version:       1,
		Snapshot:      snapshot,
		ChangeSummary: &changeSummary,
	}
	if err := s.versionRepo.CreatePlanVersion(ctx, pv); err != nil {
		slog.Error("failed to create plan version", "plan_id", plan.ID, "error", err)
	}

	// Update plan status to complete
	plan.Status = models.PlanStatusComplete
	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to finalize plan: %w", err)
	}

	// Log successful generation
	durationMs := int(duration.Milliseconds())
	model := resp.Model
	genLog := &models.GenerationLog{
		PlanID:           &plan.ID,
		UserID:           userID,
		Action:           "generate",
		Status:           "success",
		PromptTokens:     &resp.Usage.InputTokens,
		CompletionTokens: &resp.Usage.OutputTokens,
		Model:            &model,
		DurationMs:       &durationMs,
	}
	if err := s.genLogRepo.Create(ctx, genLog); err != nil {
		slog.Error("failed to log generation", "plan_id", plan.ID, "error", err)
	}

	// Async: generate embedding and update plan
	planID := plan.ID
	planInput := plan.OriginalInput
	planTitle := plan.Title
	go func() {
		if s.embeddingClient == nil {
			return
		}
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		embedding, err := s.embeddingClient.GenerateEmbedding(bgCtx, planTitle+" "+planInput)
		if err != nil {
			slog.Error("failed to generate plan embedding", "plan_id", planID, "error", err)
			return
		}
		if err := s.planRepo.UpdateEmbedding(bgCtx, planID, embedding); err != nil {
			slog.Error("failed to update plan embedding", "plan_id", planID, "error", err)
		}
	}()

	return &models.PlanResponse{
		Plan:     *plan,
		Sections: sections,
	}, nil
}

// RegenerateSection regenerates a single section of a plan with fresh AI output.
func (s *StrategyService) RegenerateSection(ctx context.Context, planID uuid.UUID, sectionType models.SectionType, userID uuid.UUID, req models.RegenerateSectionRequest) (*models.PlanSection, error) {
	// Get plan and verify ownership
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("plan not found")
	}
	if plan.UserID != userID {
		return nil, fmt.Errorf("unauthorized: plan belongs to another user")
	}

	// Check subscription allows regeneration
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription: %w", err)
	}
	if sub == nil {
		defaults := models.TierLimits(models.SubscriptionTierFree)
		sub = &defaults
	}

	// Check daily regeneration limit
	todayCount, err := s.subRepo.CountRegenerationsToday(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check regeneration limit: %w", err)
	}
	if todayCount >= sub.MaxRegenerationsPerDay {
		return nil, fmt.Errorf("daily regeneration limit reached (%d/%d)", todayCount, sub.MaxRegenerationsPerDay)
	}

	// Get all sections for context
	allSections, err := s.sectionRepo.GetByPlanID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	// Find the target section
	var targetSection *models.PlanSection
	for i := range allSections {
		if allSections[i].SectionType == sectionType {
			targetSection = &allSections[i]
			break
		}
	}
	if targetSection == nil {
		return nil, fmt.Errorf("section %s not found in plan", sectionType)
	}

	// Build regeneration prompt
	additionalContext := ""
	if req.AdditionalContext != nil {
		additionalContext = *req.AdditionalContext
	}
	systemPrompt, userPrompt := BuildSectionRegenerationPrompt(
		plan.OriginalInput,
		plan.Category,
		sectionType,
		targetSection.Content,
		additionalContext,
		allSections,
	)

	// Call Claude
	startTime := time.Now()
	messages := []claude.Message{
		claude.NewTextMessage("user", userPrompt),
	}
	claudeReq := &claude.MessagesRequest{
		MaxTokens:   4096,
		System:      systemPrompt,
		Messages:    messages,
		Temperature: 0.7,
	}

	resp, err := s.claudeClient.CreateMessage(ctx, claudeReq)
	if err != nil {
		s.logGenerationError(ctx, &planID, userID, "regenerate", &sectionType, err, startTime)
		return nil, fmt.Errorf("AI regeneration failed: %w", err)
	}
	duration := time.Since(startTime)

	// Parse response
	rawJSON, err := claude.ExtractJSON(resp.GetTextContent())
	if err != nil {
		s.logGenerationError(ctx, &planID, userID, "regenerate", &sectionType, err, startTime)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	var newContent json.RawMessage
	if err := json.Unmarshal([]byte(rawJSON), &newContent); err != nil {
		s.logGenerationError(ctx, &planID, userID, "regenerate", &sectionType, err, startTime)
		return nil, fmt.Errorf("failed to parse section content: %w", err)
	}

	// Increment section version
	targetSection.Version++
	targetSection.Content = newContent

	// Save new version in section_versions
	refinementCtx := "regeneration"
	if additionalContext != "" {
		refinementCtx = "regeneration: " + additionalContext
	}
	sv := &models.SectionVersion{
		SectionID:        targetSection.ID,
		PlanID:           planID,
		Version:          targetSection.Version,
		Content:          newContent,
		RefinementContext: &refinementCtx,
		GeneratedBy:      "claude",
		TokenUsage: map[string]int{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
		},
	}
	if err := s.versionRepo.CreateSectionVersion(ctx, sv); err != nil {
		return nil, fmt.Errorf("failed to save section version: %w", err)
	}

	// Update section content
	if err := s.sectionRepo.Update(ctx, targetSection); err != nil {
		return nil, fmt.Errorf("failed to update section: %w", err)
	}

	// Increment plan version and save snapshot
	plan.CurrentVersion++
	if err := s.planRepo.Update(ctx, plan); err != nil {
		slog.Error("failed to update plan version", "plan_id", planID, "error", err)
	}

	// Refresh sections for snapshot
	updatedSections, _ := s.sectionRepo.GetByPlanID(ctx, planID)
	changeSummary := fmt.Sprintf("Regenerated %s section (v%d)", models.DefaultSectionTitles[sectionType], targetSection.Version)
	pv := &models.PlanVersion{
		PlanID:        planID,
		Version:       plan.CurrentVersion,
		Snapshot:      map[string]any{"plan": plan, "sections": updatedSections},
		ChangeSummary: &changeSummary,
	}
	if err := s.versionRepo.CreatePlanVersion(ctx, pv); err != nil {
		slog.Error("failed to create plan version", "plan_id", planID, "error", err)
	}

	// Log generation
	durationMs := int(duration.Milliseconds())
	model := resp.Model
	genLog := &models.GenerationLog{
		PlanID:           &planID,
		UserID:           userID,
		Action:           "regenerate",
		SectionType:      &sectionType,
		Status:           "success",
		PromptTokens:     &resp.Usage.InputTokens,
		CompletionTokens: &resp.Usage.OutputTokens,
		Model:            &model,
		DurationMs:       &durationMs,
	}
	if err := s.genLogRepo.Create(ctx, genLog); err != nil {
		slog.Error("failed to log regeneration", "plan_id", planID, "error", err)
	}

	return targetSection, nil
}

// RefineSection refines a section based on specific user feedback.
func (s *StrategyService) RefineSection(ctx context.Context, planID uuid.UUID, sectionType models.SectionType, userID uuid.UUID, req models.RefineSectionRequest) (*models.PlanSection, error) {
	// Get plan and verify ownership
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("plan not found")
	}
	if plan.UserID != userID {
		return nil, fmt.Errorf("unauthorized: plan belongs to another user")
	}

	// Check subscription allows refinement
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check subscription: %w", err)
	}
	if sub == nil {
		defaults := models.TierLimits(models.SubscriptionTierFree)
		sub = &defaults
	}
	if !sub.CanRefine {
		return nil, fmt.Errorf("refinement requires a Pro subscription or higher")
	}

	// Check daily regeneration limit (refinements count toward the same limit)
	todayCount, err := s.subRepo.CountRegenerationsToday(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check regeneration limit: %w", err)
	}
	if todayCount >= sub.MaxRegenerationsPerDay {
		return nil, fmt.Errorf("daily regeneration limit reached (%d/%d)", todayCount, sub.MaxRegenerationsPerDay)
	}

	// Get the target section
	targetSection, err := s.sectionRepo.GetByPlanAndType(ctx, planID, sectionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get section: %w", err)
	}
	if targetSection == nil {
		return nil, fmt.Errorf("section %s not found in plan", sectionType)
	}

	// Build plan context from original input
	planContext := fmt.Sprintf("Original request: %s\nPlan category: %s", plan.OriginalInput, plan.Category)

	// Build refinement prompt
	systemPrompt, userPrompt := BuildRefinementPrompt(
		sectionType,
		targetSection.Content,
		req.RefinementPrompt,
		planContext,
	)

	// Call Claude
	startTime := time.Now()
	messages := []claude.Message{
		claude.NewTextMessage("user", userPrompt),
	}
	claudeReq := &claude.MessagesRequest{
		MaxTokens:   4096,
		System:      systemPrompt,
		Messages:    messages,
		Temperature: 0.7,
	}

	resp, err := s.claudeClient.CreateMessage(ctx, claudeReq)
	if err != nil {
		s.logGenerationError(ctx, &planID, userID, "refine", &sectionType, err, startTime)
		return nil, fmt.Errorf("AI refinement failed: %w", err)
	}
	duration := time.Since(startTime)

	// Parse response
	rawJSON, err := claude.ExtractJSON(resp.GetTextContent())
	if err != nil {
		s.logGenerationError(ctx, &planID, userID, "refine", &sectionType, err, startTime)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	var newContent json.RawMessage
	if err := json.Unmarshal([]byte(rawJSON), &newContent); err != nil {
		s.logGenerationError(ctx, &planID, userID, "refine", &sectionType, err, startTime)
		return nil, fmt.Errorf("failed to parse section content: %w", err)
	}

	// Increment section version
	targetSection.Version++
	targetSection.Content = newContent

	// Save new version
	refinementCtx := "refinement: " + req.RefinementPrompt
	sv := &models.SectionVersion{
		SectionID:        targetSection.ID,
		PlanID:           planID,
		Version:          targetSection.Version,
		Content:          newContent,
		RefinementContext: &refinementCtx,
		GeneratedBy:      "claude",
		TokenUsage: map[string]int{
			"input_tokens":  resp.Usage.InputTokens,
			"output_tokens": resp.Usage.OutputTokens,
		},
	}
	if err := s.versionRepo.CreateSectionVersion(ctx, sv); err != nil {
		return nil, fmt.Errorf("failed to save section version: %w", err)
	}

	// Update section content
	if err := s.sectionRepo.Update(ctx, targetSection); err != nil {
		return nil, fmt.Errorf("failed to update section: %w", err)
	}

	// Increment plan version
	plan.CurrentVersion++
	if err := s.planRepo.Update(ctx, plan); err != nil {
		slog.Error("failed to update plan version", "plan_id", planID, "error", err)
	}

	// Save plan version snapshot
	updatedSections, _ := s.sectionRepo.GetByPlanID(ctx, planID)
	changeSummary := fmt.Sprintf("Refined %s section (v%d): %s", models.DefaultSectionTitles[sectionType], targetSection.Version, truncateStr(req.RefinementPrompt, 100))
	pv := &models.PlanVersion{
		PlanID:        planID,
		Version:       plan.CurrentVersion,
		Snapshot:      map[string]any{"plan": plan, "sections": updatedSections},
		ChangeSummary: &changeSummary,
	}
	if err := s.versionRepo.CreatePlanVersion(ctx, pv); err != nil {
		slog.Error("failed to create plan version", "plan_id", planID, "error", err)
	}

	// Log generation
	durationMs := int(duration.Milliseconds())
	model := resp.Model
	genLog := &models.GenerationLog{
		PlanID:           &planID,
		UserID:           userID,
		Action:           "refine",
		SectionType:      &sectionType,
		Status:           "success",
		PromptTokens:     &resp.Usage.InputTokens,
		CompletionTokens: &resp.Usage.OutputTokens,
		Model:            &model,
		DurationMs:       &durationMs,
	}
	if err := s.genLogRepo.Create(ctx, genLog); err != nil {
		slog.Error("failed to log refinement", "plan_id", planID, "error", err)
	}

	return targetSection, nil
}

// GetPlan retrieves a plan with all sections, verifying ownership.
func (s *StrategyService) GetPlan(ctx context.Context, planID, userID uuid.UUID) (*models.PlanResponse, error) {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("plan not found")
	}
	if plan.UserID != userID {
		return nil, fmt.Errorf("unauthorized: plan belongs to another user")
	}

	sections, err := s.sectionRepo.GetByPlanID(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sections: %w", err)
	}

	return &models.PlanResponse{
		Plan:     *plan,
		Sections: sections,
	}, nil
}

// ListPlans returns a paginated list of plans matching the given filters.
func (s *StrategyService) ListPlans(ctx context.Context, filters models.PlanFilters) (*models.PlanListResponse, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	plans, total, err := s.planRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.PageSize)))

	return &models.PlanListResponse{
		Plans:      plans,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ArchivePlan sets a plan's status to archived.
func (s *StrategyService) ArchivePlan(ctx context.Context, planID, userID uuid.UUID) error {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return fmt.Errorf("plan not found")
	}
	if plan.UserID != userID {
		return fmt.Errorf("unauthorized: plan belongs to another user")
	}

	return s.planRepo.Delete(ctx, planID)
}

// ExportPlan exports a plan in the requested format.
func (s *StrategyService) ExportPlan(ctx context.Context, planID, userID uuid.UUID, format string) ([]byte, string, error) {
	plan, err := s.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get plan: %w", err)
	}
	if plan == nil {
		return nil, "", fmt.Errorf("plan not found")
	}
	if plan.UserID != userID {
		return nil, "", fmt.Errorf("unauthorized: plan belongs to another user")
	}

	// Check subscription allows export
	sub, err := s.subRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to check subscription: %w", err)
	}
	if sub == nil {
		defaults := models.TierLimits(models.SubscriptionTierFree)
		sub = &defaults
	}
	if !sub.CanExport {
		return nil, "", fmt.Errorf("export requires a Pro subscription or higher")
	}

	sections, err := s.sectionRepo.GetByPlanID(ctx, planID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get sections: %w", err)
	}

	switch format {
	case "markdown":
		md := buildMarkdownExport(plan, sections)
		return []byte(md), "text/markdown", nil
	case "json":
		data, err := json.MarshalIndent(models.PlanResponse{Plan: *plan, Sections: sections}, "", "  ")
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal plan: %w", err)
		}
		return data, "application/json", nil
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// logGenerationError logs a failed generation attempt.
func (s *StrategyService) logGenerationError(ctx context.Context, planID *uuid.UUID, userID uuid.UUID, action string, sectionType *models.SectionType, genErr error, startTime time.Time) {
	durationMs := int(time.Since(startTime).Milliseconds())
	errMsg := genErr.Error()
	genLog := &models.GenerationLog{
		PlanID:       planID,
		UserID:       userID,
		Action:       action,
		SectionType:  sectionType,
		Status:       "error",
		DurationMs:   &durationMs,
		ErrorMessage: &errMsg,
	}
	if err := s.genLogRepo.Create(ctx, genLog); err != nil {
		slog.Error("failed to log generation error", "error", err)
	}
}

// buildMarkdownExport builds a markdown representation of a plan.
func buildMarkdownExport(plan *models.StrategicPlan, sections []models.PlanSection) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", plan.Title))
	sb.WriteString(fmt.Sprintf("**Category:** %s\n", plan.Category))
	if plan.SubCategory != nil {
		sb.WriteString(fmt.Sprintf("**Sub-category:** %s\n", *plan.SubCategory))
	}
	if plan.Complexity != nil {
		sb.WriteString(fmt.Sprintf("**Complexity:** %s\n", *plan.Complexity))
	}
	sb.WriteString(fmt.Sprintf("**Created:** %s\n\n", plan.CreatedAt.Format("January 2, 2006")))
	sb.WriteString("---\n\n")

	for _, section := range sections {
		sb.WriteString(fmt.Sprintf("## %s\n\n", section.Title))

		// Marshal section content to pretty JSON as fallback
		contentBytes, err := json.MarshalIndent(section.Content, "", "  ")
		if err != nil {
			sb.WriteString("*Content unavailable*\n\n")
			continue
		}

		// Try to render structured content based on section type
		rendered := renderSectionContent(section.SectionType, contentBytes)
		sb.WriteString(rendered)
		sb.WriteString("\n---\n\n")
	}

	return sb.String()
}

// renderSectionContent renders section content as markdown based on type.
func renderSectionContent(sectionType models.SectionType, contentBytes []byte) string {
	var sb strings.Builder

	switch sectionType {
	case models.SectionTypeExecutiveBrief:
		var c models.ExecutiveBriefContent
		if err := json.Unmarshal(contentBytes, &c); err == nil {
			sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", c.Summary))
			sb.WriteString(fmt.Sprintf("**Objective:** %s\n\n", c.Objective))
			sb.WriteString(fmt.Sprintf("**Scope:** %s\n\n", c.Scope))
			sb.WriteString(fmt.Sprintf("**Expected Outcome:** %s\n\n", c.ExpectedOutcome))
			sb.WriteString("**Key Stakeholders:**\n")
			for _, s := range c.KeyStakeholders {
				sb.WriteString(fmt.Sprintf("- %s\n", s))
			}
			sb.WriteString(fmt.Sprintf("\n**Timeline Overview:** %s\n\n", c.TimelineOverview))
			return sb.String()
		}

	case models.SectionTypeStrategicContext:
		var c models.StrategicContextContent
		if err := json.Unmarshal(contentBytes, &c); err == nil {
			sb.WriteString(fmt.Sprintf("**Industry Analysis:** %s\n\n", c.IndustryAnalysis))
			sb.WriteString(fmt.Sprintf("**Market Conditions:** %s\n\n", c.MarketConditions))
			sb.WriteString(fmt.Sprintf("**Competitive Landscape:** %s\n\n", c.CompetitiveLandscape))
			sb.WriteString("**Opportunities:**\n")
			for _, o := range c.Opportunities {
				sb.WriteString(fmt.Sprintf("- **%s** (%s impact): %s\n", o.Title, o.Impact, o.Description))
			}
			sb.WriteString("\n**Threats:**\n")
			for _, t := range c.Threats {
				sb.WriteString(fmt.Sprintf("- **%s** (%s impact): %s\n", t.Title, t.Impact, t.Description))
			}
			sb.WriteString("\n**Assumptions:**\n")
			for _, a := range c.Assumptions {
				sb.WriteString(fmt.Sprintf("- %s\n", a))
			}
			sb.WriteString("\n")
			return sb.String()
		}

	case models.SectionTypeRecommendedApproach:
		var c models.RecommendedApproachContent
		if err := json.Unmarshal(contentBytes, &c); err == nil {
			sb.WriteString(fmt.Sprintf("**Core Strategy:** %s\n\n", c.CoreStrategy))
			sb.WriteString(fmt.Sprintf("**Rationale:** %s\n\n", c.Rationale))
			sb.WriteString("**Key Pillars:**\n")
			for _, p := range c.KeyPillars {
				sb.WriteString(fmt.Sprintf("- **%s:** %s\n", p.Name, p.Description))
				for _, kpi := range p.KPIs {
					sb.WriteString(fmt.Sprintf("  - KPI: %s\n", kpi))
				}
			}
			sb.WriteString("\n**Alternatives Considered:**\n")
			for _, a := range c.AlternativesConsidered {
				sb.WriteString(fmt.Sprintf("- **%s** — Pros: %s | Cons: %s | Why not chosen: %s\n", a.Name, a.Pros, a.Cons, a.WhyNotChosen))
			}
			sb.WriteString("\n**Risk Mitigation:**\n")
			for _, r := range c.RiskMitigation {
				sb.WriteString(fmt.Sprintf("- **%s** (likelihood: %s, impact: %s): %s\n", r.Risk, r.Likelihood, r.Impact, r.Mitigation))
			}
			sb.WriteString("\n")
			return sb.String()
		}

	case models.SectionTypePhasedExecution:
		var c models.PhasedExecutionContent
		if err := json.Unmarshal(contentBytes, &c); err == nil {
			for i, phase := range c.Phases {
				sb.WriteString(fmt.Sprintf("### Phase %d: %s (%s)\n\n", i+1, phase.Name, phase.Duration))
				sb.WriteString(fmt.Sprintf("**Objective:** %s\n\n", phase.Objective))
				sb.WriteString("**Milestones:**\n")
				for _, m := range phase.Milestones {
					sb.WriteString(fmt.Sprintf("- %s — Target: %s, Metric: %s\n", m.Name, m.Target, m.Metric))
				}
				sb.WriteString("\n**Actions:**\n")
				for _, a := range phase.Actions {
					sb.WriteString(fmt.Sprintf("- [%s] %s — Owner: %s, Timeline: %s\n", a.Priority, a.Title, a.Owner, a.Timeline))
				}
				sb.WriteString(fmt.Sprintf("\n**Dependencies:** %s\n", phase.Dependencies))
				sb.WriteString(fmt.Sprintf("**Success Criteria:** %s\n\n", phase.SuccessCriteria))
			}
			return sb.String()
		}

	case models.SectionTypeImmediateAction:
		var c models.ImmediateActionContent
		if err := json.Unmarshal(contentBytes, &c); err == nil {
			sb.WriteString(fmt.Sprintf("**Time Horizon:** %s\n\n", c.TimeHorizon))
			sb.WriteString("**Quick Wins:**\n")
			for _, q := range c.QuickWins {
				sb.WriteString(fmt.Sprintf("- [%s] **%s** — %s (Owner: %s, Deadline: %s)\n", q.Priority, q.Title, q.Description, q.Owner, q.Deadline))
			}
			sb.WriteString("\n**Critical Path:**\n")
			for _, cp := range c.CriticalPath {
				sb.WriteString(fmt.Sprintf("- [%s] **%s** — %s (Owner: %s, Deadline: %s)\n", cp.Priority, cp.Title, cp.Description, cp.Owner, cp.Deadline))
			}
			sb.WriteString("\n**Resource Needs:**\n")
			for _, r := range c.ResourceNeeds {
				sb.WriteString(fmt.Sprintf("- %s — Quantity: %s, Urgency: %s\n", r.Resource, r.Quantity, r.Urgency))
			}
			sb.WriteString("\n**Week-by-Week Plan:**\n")
			for _, w := range c.WeekByWeek {
				sb.WriteString(fmt.Sprintf("\n*Week %d: %s*\n", w.Week, w.Focus))
				for _, t := range w.Tasks {
					sb.WriteString(fmt.Sprintf("- [%s] %s — %s (Owner: %s)\n", t.Priority, t.Title, t.Description, t.Owner))
				}
			}
			sb.WriteString("\n")
			return sb.String()
		}
	}

	// Fallback: pretty-print JSON
	return fmt.Sprintf("```json\n%s\n```\n\n", string(contentBytes))
}

// truncateStr truncates a string to the given max length, appending "..." if truncated.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
