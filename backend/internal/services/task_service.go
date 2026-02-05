package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

type TaskService struct {
	taskRepo          *repository.TaskRepository
	linkRepo          *repository.TaskLinkRepository
	justificationRepo *repository.JustificationRepository
	blockerRepo       *repository.BlockerRepository
	reviewRepo        *repository.ReviewRepository
	ackRepo           *repository.AcknowledgmentRepository
	embeddingService  EmbeddingService
}

type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error)
}

func NewTaskService(
	taskRepo *repository.TaskRepository,
	linkRepo *repository.TaskLinkRepository,
	justificationRepo *repository.JustificationRepository,
	blockerRepo *repository.BlockerRepository,
	reviewRepo *repository.ReviewRepository,
	ackRepo *repository.AcknowledgmentRepository,
	embeddingService EmbeddingService,
) *TaskService {
	return &TaskService{
		taskRepo:          taskRepo,
		linkRepo:          linkRepo,
		justificationRepo: justificationRepo,
		blockerRepo:       blockerRepo,
		reviewRepo:        reviewRepo,
		ackRepo:           ackRepo,
		embeddingService:  embeddingService,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req models.CreateTaskRequest, userID, orgID uuid.UUID) (*models.TaskResponse, error) {
	task := &models.Task{
		Title:          req.Title,
		Description:    req.Description,
		ProjectID:      req.ProjectID,
		OrganizationID: orgID,
		CreatedBy:      userID,
		AssignedTo:     req.AssignedTo,
		EstimatedDays:  req.EstimatedDays,
		Tags:           req.Tags,
		Status:         models.TaskStatusDraft,
	}

	// Generate embedding for similarity search
	if s.embeddingService != nil {
		embedding, err := s.embeddingService.GenerateEmbedding(ctx, task.Title+" "+task.Description)
		if err == nil {
			task.DescriptionEmbedding = embedding
		}
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	response := &models.TaskResponse{Task: task}

	// Find similar tasks
	if task.DescriptionEmbedding.Slice() != nil && len(task.DescriptionEmbedding.Slice()) > 0 {
		similar, err := s.taskRepo.FindSimilar(ctx, task.DescriptionEmbedding, orgID, &task.ID, 5)
		if err == nil {
			response.SimilarTasks = similar
		}

		// Generate predictions if we have similar tasks
		if len(similar) > 0 {
			predictions := s.generatePredictions(ctx, similar, orgID)
			response.Predictions = predictions

			// Update task with predictions
			if predictions != nil {
				s.taskRepo.UpdatePredictions(ctx, task.ID, predictions.PredictedDaysLow, predictions.PredictedDaysHigh, predictions.Confidence)
				task.PredictedDaysLow = &predictions.PredictedDaysLow
				task.PredictedDaysHigh = &predictions.PredictedDaysHigh
				task.PredictionConfidence = &predictions.Confidence
			}
		}
	}

	// Calculate planning quality assessment
	assessment := s.assessPlanningQuality(task, response.Predictions)
	response.PlanningAssessment = assessment
	if assessment != nil {
		s.taskRepo.UpdatePlanningScore(ctx, task.ID, assessment.Score)
		task.PlanningQualityScore = &assessment.Score
	}

	// Update status to pending acknowledgment if predictions exist
	if response.Predictions != nil || len(response.SimilarTasks) > 0 {
		task.Status = models.TaskStatusPendingAcknowledgment
		s.taskRepo.Update(ctx, task)
	}

	return response, nil
}

func (s *TaskService) GetTask(ctx context.Context, id, orgID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil || task.OrganizationID != orgID {
		return nil, nil
	}
	return task, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, req models.UpdateTaskRequest, orgID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return nil, fmt.Errorf("task not found")
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
		// Regenerate embedding if description changes
		if s.embeddingService != nil {
			embedding, err := s.embeddingService.GenerateEmbedding(ctx, task.Title+" "+task.Description)
			if err == nil {
				task.DescriptionEmbedding = embedding
			}
		}
	}
	if req.EstimatedDays != nil {
		task.EstimatedDays = req.EstimatedDays
	}
	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}
	if req.Tags != nil {
		task.Tags = req.Tags
	}
	if req.Status != nil {
		task.Status = *req.Status
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id, orgID uuid.UUID) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return fmt.Errorf("task not found")
	}
	return s.taskRepo.Delete(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context, filters models.TaskFilters) (*models.TaskListResponse, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 50
	}

	tasks, total, err := s.taskRepo.List(ctx, filters)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filters.PageSize)))

	return &models.TaskListResponse{
		Tasks:      tasks,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *TaskService) GetSimilarTasks(ctx context.Context, id, orgID uuid.UUID) ([]models.SimilarTask, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return nil, fmt.Errorf("task not found")
	}

	if task.DescriptionEmbedding.Slice() == nil || len(task.DescriptionEmbedding.Slice()) == 0 {
		return []models.SimilarTask{}, nil
	}

	return s.taskRepo.FindSimilar(ctx, task.DescriptionEmbedding, orgID, &task.ID, 10)
}

func (s *TaskService) AcknowledgeTask(ctx context.Context, id, userID, orgID uuid.UUID, req *models.AcknowledgmentRequest) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return nil, fmt.Errorf("task not found")
	}

	// Use state machine to validate transition
	if err := models.ValidateTransition(task.Status, models.TaskStatusAcknowledged); err != nil {
		return nil, fmt.Errorf("task cannot be acknowledged in current state: %w", err)
	}

	// Check if task has links or justification
	linkCount, err := s.linkRepo.CountByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	hasJustification, err := s.justificationRepo.Exists(ctx, id)
	if err != nil {
		return nil, err
	}

	if linkCount == 0 && !hasJustification {
		return nil, fmt.Errorf("task must have at least one link or a justification before acknowledgment")
	}

	// Validate acknowledgment request
	if req != nil {
		if err := req.Validate(); err != nil {
			return nil, err
		}
	}

	// Record acknowledgment in history
	ack := &models.TaskAcknowledgment{
		TaskID:         id,
		AcknowledgedBy: userID,
		PredictedLow:   task.PredictedDaysLow,
		PredictedHigh:  task.PredictedDaysHigh,
	}

	if req != nil {
		ack.Action = req.Action
		ack.OriginalEstimate = task.EstimatedDays
		ack.DisagreementNotes = req.DisagreementNotes

		// Handle different acknowledgment actions
		switch req.Action {
		case models.AcknowledgmentAccept:
			// User accepts the prediction as-is
		case models.AcknowledgmentModify:
			// User provides their own estimate
			ack.ModifiedEstimate = req.ModifiedEstimate
			task.EstimatedDays = req.ModifiedEstimate
		case models.AcknowledgmentDisagree:
			// User disagrees with prediction but proceeds
		}
	} else {
		// Default to accept if no request provided
		ack.Action = models.AcknowledgmentAccept
		ack.OriginalEstimate = task.EstimatedDays
	}

	// Save acknowledgment record
	if err := s.ackRepo.Create(ctx, ack); err != nil {
		return nil, fmt.Errorf("failed to record acknowledgment: %w", err)
	}

	now := time.Now()
	task.AcknowledgedAt = &now
	task.Status = models.TaskStatusAcknowledged

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) StartTask(ctx context.Context, id, userID, orgID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return nil, fmt.Errorf("task not found")
	}

	// Use state machine to validate transition
	if err := models.ValidateTransition(task.Status, models.TaskStatusActive); err != nil {
		return nil, fmt.Errorf("task must be acknowledged before starting: %w", err)
	}

	// Check for blocking dependencies
	blockingTasks, err := s.linkRepo.GetBlockingTasks(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check dependencies: %w", err)
	}
	if len(blockingTasks) > 0 {
		task.Status = models.TaskStatusBlocked
		if err := s.taskRepo.Update(ctx, task); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("task is blocked by %d incomplete dependencies", len(blockingTasks))
	}

	now := time.Now()
	task.StartedAt = &now
	task.Status = models.TaskStatusActive

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) CompleteTask(ctx context.Context, id, userID, orgID uuid.UUID) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return nil, fmt.Errorf("task not found")
	}

	// Use state machine to validate transition
	if err := models.ValidateTransition(task.Status, models.TaskStatusPendingReview); err != nil {
		return nil, fmt.Errorf("task must be active to complete: %w", err)
	}

	task.Status = models.TaskStatusPendingReview
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Unblock any tasks that were waiting on this task
	s.tryUnblockDependentTasks(ctx, id)

	return task, nil
}

// tryUnblockDependentTasks checks tasks that depend on the given task and unblocks them if possible
func (s *TaskService) tryUnblockDependentTasks(ctx context.Context, completedTaskID uuid.UUID) {
	// Get all tasks that depend on the completed task
	dependentTasks, err := s.linkRepo.GetDependentTasks(ctx, completedTaskID)
	if err != nil {
		return
	}

	for _, taskID := range dependentTasks {
		s.tryUnblockTask(ctx, taskID)
	}
}

// tryUnblockTask attempts to unblock a task if all its dependencies are complete
func (s *TaskService) tryUnblockTask(ctx context.Context, taskID uuid.UUID) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil {
		return
	}

	// Only unblock tasks that are currently blocked
	if task.Status != models.TaskStatusBlocked {
		return
	}

	// Check if there are still blocking dependencies
	blockingTasks, err := s.linkRepo.GetBlockingTasks(ctx, taskID)
	if err != nil {
		return
	}

	// If no more blockers, unblock the task (move back to acknowledged state)
	if len(blockingTasks) == 0 {
		task.Status = models.TaskStatusAcknowledged
		s.taskRepo.Update(ctx, task)
	}
}

// CheckAndUpdateBlockedStatus checks if a task should be blocked and updates its status
func (s *TaskService) CheckAndUpdateBlockedStatus(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	// Only check for tasks that can be blocked
	if task.Status == models.TaskStatusCompleted || task.Status == models.TaskStatusCancelled {
		return nil
	}

	blockingTasks, err := s.linkRepo.GetBlockingTasks(ctx, taskID)
	if err != nil {
		return err
	}

	if len(blockingTasks) > 0 && task.Status != models.TaskStatusBlocked {
		task.Status = models.TaskStatusBlocked
		return s.taskRepo.Update(ctx, task)
	} else if len(blockingTasks) == 0 && task.Status == models.TaskStatusBlocked {
		task.Status = models.TaskStatusAcknowledged
		return s.taskRepo.Update(ctx, task)
	}

	return nil
}

// GetBlockingInfo returns information about what's blocking a task
func (s *TaskService) GetBlockingInfo(ctx context.Context, taskID, orgID uuid.UUID) ([]models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task == nil || task.OrganizationID != orgID {
		return nil, fmt.Errorf("task not found")
	}

	blockingIDs, err := s.linkRepo.GetBlockingTasks(ctx, taskID)
	if err != nil {
		return nil, err
	}

	var blockingTasks []models.Task
	for _, id := range blockingIDs {
		t, err := s.taskRepo.GetByID(ctx, id)
		if err == nil && t != nil {
			blockingTasks = append(blockingTasks, *t)
		}
	}

	return blockingTasks, nil
}

func (s *TaskService) generatePredictions(ctx context.Context, similar []models.SimilarTask, orgID uuid.UUID) *models.Predictions {
	if len(similar) == 0 {
		return nil
	}

	var totalWeight, weightedSum float64
	var durations []float64

	for _, task := range similar {
		if task.ActualDays != nil {
			weight := task.SimilarityScore
			weightedSum += *task.ActualDays * weight
			totalWeight += weight
			durations = append(durations, *task.ActualDays)
		}
	}

	if totalWeight == 0 {
		return nil
	}

	mean := weightedSum / totalWeight

	// Calculate standard deviation
	var variance float64
	for _, d := range durations {
		variance += (d - mean) * (d - mean)
	}
	stdDev := math.Sqrt(variance / float64(len(durations)))

	// Calculate confidence based on sample size and variance
	confidence := 0.5
	if len(durations) >= 5 {
		confidence += 0.2
	}
	if stdDev/mean < 0.3 { // Low variance
		confidence += 0.2
	}
	confidence = math.Min(confidence, 0.95)

	// Get blocker risks
	blockerRisks := s.calculateBlockerRisks(ctx, similar, orgID)

	return &models.Predictions{
		PredictedDaysLow:  math.Max(1, mean-stdDev),
		PredictedDaysHigh: mean + stdDev,
		Confidence:        confidence,
		BlockerRisks:      blockerRisks,
	}
}

func (s *TaskService) calculateBlockerRisks(ctx context.Context, similar []models.SimilarTask, orgID uuid.UUID) []models.BlockerRisk {
	patterns, err := s.blockerRepo.GetBlockerPatterns(ctx, orgID, 10)
	if err != nil {
		return nil
	}

	var totalTasks = len(similar)
	if totalTasks == 0 {
		return nil
	}

	// Get blocker examples for similar tasks
	blockerExamples := make(map[string][]string)
	for _, task := range similar {
		blockers, err := s.blockerRepo.ListByTaskID(ctx, task.ID)
		if err != nil {
			continue
		}
		for _, b := range blockers {
			key := string(b.BlockerType)
			if len(blockerExamples[key]) < 3 {
				example := fmt.Sprintf("%s: %s", task.Title, truncateString(b.Description, 100))
				blockerExamples[key] = append(blockerExamples[key], example)
			}
		}
	}

	var risks []models.BlockerRisk
	for blockerType, count := range patterns {
		probability := float64(count) / float64(totalTasks*2) // Normalize
		if probability > 0.2 {
			risk := models.BlockerRisk{
				Type:        string(blockerType),
				Probability: math.Min(probability, 1.0),
				Examples:    blockerExamples[string(blockerType)],
			}
			risks = append(risks, risk)
		}
	}

	return risks
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (s *TaskService) assessPlanningQuality(task *models.Task, predictions *models.Predictions) *models.Assessment {
	score := 0.0
	breakdown := make(map[string]int)
	var suggestions []string

	// Clear objective (20 points)
	if len(task.Description) >= 100 {
		breakdown["clear_objective"] = 20
		score += 20
	} else {
		breakdown["clear_objective"] = 10
		score += 10
		suggestions = append(suggestions, "Provide a more detailed description with measurable outcomes")
	}

	// Dependencies identified (20 points) - check for keywords
	depKeywords := []string{"depend", "require", "need", "block", "team", "external"}
	hasDepKeywords := false
	for _, kw := range depKeywords {
		if containsIgnoreCase(task.Description, kw) {
			hasDepKeywords = true
			break
		}
	}
	if hasDepKeywords {
		breakdown["dependencies"] = 20
		score += 20
	} else {
		breakdown["dependencies"] = 0
		suggestions = append(suggestions, "Document any dependencies on external teams or systems")
	}

	// Risks documented (20 points)
	riskKeywords := []string{"risk", "might", "could", "potential", "concern", "issue"}
	hasRiskKeywords := false
	for _, kw := range riskKeywords {
		if containsIgnoreCase(task.Description, kw) {
			hasRiskKeywords = true
			break
		}
	}
	if hasRiskKeywords {
		breakdown["risks"] = 20
		score += 20
	} else {
		breakdown["risks"] = 0
		suggestions = append(suggestions, "Document potential risks or blockers")
	}

	// Acceptance criteria (20 points)
	acKeywords := []string{"done when", "complete when", "success", "accept", "criteria", "verify"}
	hasACKeywords := false
	for _, kw := range acKeywords {
		if containsIgnoreCase(task.Description, kw) {
			hasACKeywords = true
			break
		}
	}
	if hasACKeywords {
		breakdown["acceptance_criteria"] = 20
		score += 20
	} else {
		breakdown["acceptance_criteria"] = 0
		suggestions = append(suggestions, "Add clear acceptance criteria or definition of done")
	}

	// Realistic estimate (20 points)
	if predictions != nil && task.EstimatedDays != nil {
		est := *task.EstimatedDays
		if est >= predictions.PredictedDaysLow && est <= predictions.PredictedDaysHigh*1.5 {
			breakdown["realistic_estimate"] = 20
			score += 20
		} else if est < predictions.PredictedDaysLow {
			breakdown["realistic_estimate"] = 5
			score += 5
			suggestions = append(suggestions, fmt.Sprintf("Your estimate (%.1f days) is below the predicted range (%.1f-%.1f days)", est, predictions.PredictedDaysLow, predictions.PredictedDaysHigh))
		} else {
			breakdown["realistic_estimate"] = 10
			score += 10
		}
	} else {
		breakdown["realistic_estimate"] = 10
		score += 10
		if task.EstimatedDays == nil {
			suggestions = append(suggestions, "Provide a time estimate for this task")
		}
	}

	return &models.Assessment{
		Score:       score,
		Breakdown:   breakdown,
		Suggestions: suggestions,
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsIgnoreCaseHelper(s, substr))
}

func containsIgnoreCaseHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldAt(s, i, substr) {
			return true
		}
	}
	return false
}

func equalFoldAt(s string, start int, substr string) bool {
	for j := 0; j < len(substr); j++ {
		c1 := s[start+j]
		c2 := substr[j]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}
