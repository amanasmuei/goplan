package workers

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
	"github.com/goplan/backend/internal/services"
)

type StrategyEmbeddingWorker struct {
	planRepo        *repository.PlanRepository
	embeddingClient *services.EmbeddingClient
	interval        time.Duration
	batchSize       int
	running         atomic.Bool
}

func NewStrategyEmbeddingWorker(planRepo *repository.PlanRepository, embeddingClient *services.EmbeddingClient, interval time.Duration, batchSize int) *StrategyEmbeddingWorker {
	return &StrategyEmbeddingWorker{
		planRepo:        planRepo,
		embeddingClient: embeddingClient,
		interval:        interval,
		batchSize:       batchSize,
	}
}

func (w *StrategyEmbeddingWorker) Start(ctx context.Context) {
	slog.Info("Strategy embedding worker started", "interval", w.interval, "batch_size", w.batchSize)

	// Run immediately on start
	w.processOnce(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Strategy embedding worker stopped")
			return
		case <-ticker.C:
			w.processOnce(ctx)
		}
	}
}

func (w *StrategyEmbeddingWorker) processOnce(ctx context.Context) {
	if !w.running.CompareAndSwap(false, true) {
		slog.Info("Strategy embedding worker: skipping, previous batch still processing")
		return
	}
	defer w.running.Store(false)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	plans, err := w.planRepo.FindWithoutEmbedding(ctx, w.batchSize)
	if err != nil {
		slog.Error("Strategy embedding worker: failed to find plans", "error", err)
		return
	}

	if len(plans) == 0 {
		return
	}

	slog.Info("Strategy embedding worker: processing plans", "count", len(plans))

	texts := make([]string, len(plans))
	for i, plan := range plans {
		texts[i] = plan.Title + " " + plan.OriginalInput
	}

	embeddings, err := w.embeddingClient.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		slog.Error("Strategy embedding worker: batch embedding failed, falling back to individual", "error", err)
		w.processIndividually(ctx, plans)
		return
	}

	updated := 0
	for i, plan := range plans {
		if i < len(embeddings) {
			if err := w.planRepo.UpdateEmbedding(ctx, plan.ID, embeddings[i]); err != nil {
				slog.Error("Strategy embedding worker: failed to update embedding", "plan_id", plan.ID, "error", err)
				continue
			}
			updated++
		}
	}

	slog.Info("Strategy embedding worker: batch complete", "updated", updated, "total", len(plans))
}

func (w *StrategyEmbeddingWorker) processIndividually(ctx context.Context, plans []models.StrategicPlan) {
	updated := 0
	for _, plan := range plans {
		select {
		case <-ctx.Done():
			slog.Info("Strategy embedding worker: context cancelled during individual processing")
			return
		default:
		}

		text := plan.Title + " " + plan.OriginalInput
		embedding, err := w.embeddingClient.GenerateEmbedding(ctx, text)
		if err != nil {
			slog.Error("Strategy embedding worker: failed to generate embedding", "plan_id", plan.ID, "error", err)
			continue
		}

		if err := w.planRepo.UpdateEmbedding(ctx, plan.ID, embedding); err != nil {
			slog.Error("Strategy embedding worker: failed to update embedding", "plan_id", plan.ID, "error", err)
			continue
		}
		updated++
	}

	slog.Info("Strategy embedding worker: individual processing complete", "updated", updated, "total", len(plans))
}
