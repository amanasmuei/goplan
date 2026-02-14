package workers

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/goplan/backend/internal/services"
)

type EmbeddingWorker struct {
	pool            *pgxpool.Pool
	embeddingClient *services.EmbeddingClient
	interval        time.Duration
	batchSize       int
	stopCh          chan struct{}
	wg              sync.WaitGroup
	processing      atomic.Bool
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewEmbeddingWorker(pool *pgxpool.Pool, embeddingClient *services.EmbeddingClient) *EmbeddingWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &EmbeddingWorker{
		pool:            pool,
		embeddingClient: embeddingClient,
		interval:        5 * time.Minute,
		batchSize:       50,
		stopCh:          make(chan struct{}),
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (w *EmbeddingWorker) Start() {
	w.wg.Add(1)
	go w.run()
	slog.Info("Embedding worker started")
}

func (w *EmbeddingWorker) Stop() {
	w.cancel()
	close(w.stopCh)
	w.wg.Wait()
	slog.Info("Embedding worker stopped")
}

func (w *EmbeddingWorker) run() {
	defer w.wg.Done()

	// Run immediately on start
	w.processTasksWithoutEmbeddings()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processTasksWithoutEmbeddings()
		}
	}
}

func (w *EmbeddingWorker) processTasksWithoutEmbeddings() {
	// Concurrency guard: skip if a previous batch is still processing
	if !w.processing.CompareAndSwap(false, true) {
		slog.Info("Skipping embedding batch: previous batch still processing")
		return
	}
	defer w.processing.Store(false)

	// Use the worker's context as parent so cancellation propagates on Stop()
	ctx, cancel := context.WithTimeout(w.ctx, 10*time.Minute)
	defer cancel()

	// Find tasks without embeddings
	tasks, err := w.findTasksWithoutEmbeddings(ctx, w.batchSize)
	if err != nil {
		slog.Error("Error finding tasks without embeddings", "error", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	slog.Info("Processing tasks without embeddings", "count", len(tasks))

	// Prepare texts for batch embedding
	texts := make([]string, len(tasks))
	for i, task := range tasks {
		texts[i] = task.Title + " " + task.Description
	}

	// Generate batch embeddings
	embeddings, err := w.embeddingClient.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		slog.Error("Error generating batch embeddings", "error", err)
		// Fall back to individual processing
		w.processIndividually(ctx, tasks)
		return
	}

	// Update tasks with embeddings
	for i, task := range tasks {
		if i < len(embeddings) {
			if err := w.updateTaskEmbedding(ctx, task.ID, embeddings[i]); err != nil {
				slog.Error("Error updating embedding for task", "task_id", task.ID, "error", err)
			}
		}
	}

	slog.Info("Successfully updated embeddings", "count", len(tasks))
}

func (w *EmbeddingWorker) processIndividually(ctx context.Context, tasks []taskRow) {
	for _, task := range tasks {
		// Check for context cancellation between iterations
		select {
		case <-ctx.Done():
			slog.Info("Embedding worker: context cancelled during individual processing")
			return
		default:
		}

		embedding, err := w.embeddingClient.GenerateEmbedding(ctx, task.Title+" "+task.Description)
		if err != nil {
			slog.Error("Error generating embedding for task", "task_id", task.ID, "error", err)
			continue
		}
		if err := w.updateTaskEmbedding(ctx, task.ID, embedding); err != nil {
			slog.Error("Error updating embedding for task", "task_id", task.ID, "error", err)
		}
	}
}

type taskRow struct {
	ID          uuid.UUID
	Title       string
	Description string
}

func (w *EmbeddingWorker) findTasksWithoutEmbeddings(ctx context.Context, limit int) ([]taskRow, error) {
	query := `
		SELECT id, title, description
		FROM tasks
		WHERE description_embedding IS NULL
		  AND status != 'cancelled'
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := w.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskRow
	for rows.Next() {
		var task taskRow
		if err := rows.Scan(&task.ID, &task.Title, &task.Description); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (w *EmbeddingWorker) updateTaskEmbedding(ctx context.Context, id uuid.UUID, embedding pgvector.Vector) error {
	query := `UPDATE tasks SET description_embedding = $1, updated_at = NOW() WHERE id = $2`
	_, err := w.pool.Exec(ctx, query, embedding, id)
	return err
}

// ProcessUpdatedTask processes a single task that was just updated
func (w *EmbeddingWorker) ProcessUpdatedTask(ctx context.Context, taskID uuid.UUID, title, description string) error {
	embedding, err := w.embeddingClient.GenerateEmbedding(ctx, title+" "+description)
	if err != nil {
		return err
	}
	return w.updateTaskEmbedding(ctx, taskID, embedding)
}
