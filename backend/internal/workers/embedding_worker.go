package workers

import (
	"context"
	"log"
	"sync"
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
}

func NewEmbeddingWorker(pool *pgxpool.Pool, embeddingClient *services.EmbeddingClient) *EmbeddingWorker {
	return &EmbeddingWorker{
		pool:            pool,
		embeddingClient: embeddingClient,
		interval:        5 * time.Minute,
		batchSize:       50,
		stopCh:          make(chan struct{}),
	}
}

func (w *EmbeddingWorker) Start() {
	w.wg.Add(1)
	go w.run()
	log.Println("Embedding worker started")
}

func (w *EmbeddingWorker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	log.Println("Embedding worker stopped")
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Find tasks without embeddings
	tasks, err := w.findTasksWithoutEmbeddings(ctx, w.batchSize)
	if err != nil {
		log.Printf("Error finding tasks without embeddings: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	log.Printf("Processing %d tasks without embeddings", len(tasks))

	// Prepare texts for batch embedding
	texts := make([]string, len(tasks))
	for i, task := range tasks {
		texts[i] = task.Title + " " + task.Description
	}

	// Generate batch embeddings
	embeddings, err := w.embeddingClient.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		log.Printf("Error generating batch embeddings: %v", err)
		// Fall back to individual processing
		w.processIndividually(ctx, tasks)
		return
	}

	// Update tasks with embeddings
	for i, task := range tasks {
		if i < len(embeddings) {
			if err := w.updateTaskEmbedding(ctx, task.ID, embeddings[i]); err != nil {
				log.Printf("Error updating embedding for task %s: %v", task.ID, err)
			}
		}
	}

	log.Printf("Successfully updated embeddings for %d tasks", len(tasks))
}

func (w *EmbeddingWorker) processIndividually(ctx context.Context, tasks []taskRow) {
	for _, task := range tasks {
		embedding, err := w.embeddingClient.GenerateEmbedding(ctx, task.Title+" "+task.Description)
		if err != nil {
			log.Printf("Error generating embedding for task %s: %v", task.ID, err)
			continue
		}
		if err := w.updateTaskEmbedding(ctx, task.ID, embedding); err != nil {
			log.Printf("Error updating embedding for task %s: %v", task.ID, err)
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
