package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/goplan/backend/internal/models"
)

type TaskLinkRepository struct {
	db *pgxpool.Pool
}

func NewTaskLinkRepository(db *pgxpool.Pool) *TaskLinkRepository {
	return &TaskLinkRepository{db: db}
}

func (r *TaskLinkRepository) Create(ctx context.Context, link *models.TaskLink) error {
	link.ID = uuid.New()
	link.CreatedAt = time.Now()

	query := `
		INSERT INTO task_links (id, source_task_id, target_task_id, link_type, created_by, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(ctx, query,
		link.ID, link.SourceTaskID, link.TargetTaskID, link.LinkType,
		link.CreatedBy, link.Notes, link.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task link: %w", err)
	}
	return nil
}

func (r *TaskLinkRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TaskLink, error) {
	query := `
		SELECT id, source_task_id, target_task_id, link_type, created_by, notes, created_at
		FROM task_links WHERE id = $1`

	link := &models.TaskLink{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&link.ID, &link.SourceTaskID, &link.TargetTaskID, &link.LinkType,
		&link.CreatedBy, &link.Notes, &link.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task link: %w", err)
	}
	return link, nil
}

func (r *TaskLinkRepository) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]models.TaskLink, error) {
	query := `
		SELECT id, source_task_id, target_task_id, link_type, created_by, notes, created_at
		FROM task_links WHERE source_task_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list task links: %w", err)
	}
	defer rows.Close()

	var links []models.TaskLink
	for rows.Next() {
		var link models.TaskLink
		err := rows.Scan(
			&link.ID, &link.SourceTaskID, &link.TargetTaskID, &link.LinkType,
			&link.CreatedBy, &link.Notes, &link.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task link: %w", err)
		}
		links = append(links, link)
	}
	return links, nil
}

func (r *TaskLinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM task_links WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task link: %w", err)
	}
	return nil
}

func (r *TaskLinkRepository) Exists(ctx context.Context, sourceID, targetID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM task_links WHERE source_task_id = $1 AND target_task_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, sourceID, targetID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check link existence: %w", err)
	}
	return exists, nil
}

func (r *TaskLinkRepository) CountByTaskID(ctx context.Context, taskID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM task_links WHERE source_task_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, taskID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count task links: %w", err)
	}
	return count, nil
}

// HasCircularDependency checks if adding a dependency from sourceID to targetID would create a cycle.
// It traverses the dependency graph starting from targetID to see if we can reach sourceID.
// If sourceID is reachable from targetID, adding this link would create a circular dependency.
func (r *TaskLinkRepository) HasCircularDependency(ctx context.Context, sourceID, targetID uuid.UUID) (bool, error) {
	// Use recursive CTE to find all tasks that targetID depends on (transitively)
	// If sourceID is in that set, adding this link would create a cycle
	query := `
		WITH RECURSIVE dependency_chain AS (
			-- Base case: direct dependencies of the target task
			SELECT target_task_id
			FROM task_links
			WHERE source_task_id = $1 AND link_type = 'dependent'

			UNION

			-- Recursive case: dependencies of dependencies
			SELECT tl.target_task_id
			FROM task_links tl
			INNER JOIN dependency_chain dc ON tl.source_task_id = dc.target_task_id
			WHERE tl.link_type = 'dependent'
		)
		SELECT EXISTS(SELECT 1 FROM dependency_chain WHERE target_task_id = $2)`

	var hasCircular bool
	err := r.db.QueryRow(ctx, query, targetID, sourceID).Scan(&hasCircular)
	if err != nil {
		return false, fmt.Errorf("failed to check circular dependency: %w", err)
	}
	return hasCircular, nil
}

// GetDependencyChain returns all task IDs in the dependency chain for a given task.
// This includes both direct dependencies and transitive dependencies.
func (r *TaskLinkRepository) GetDependencyChain(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		WITH RECURSIVE dependency_chain AS (
			-- Base case: direct dependencies
			SELECT target_task_id, 1 as depth
			FROM task_links
			WHERE source_task_id = $1 AND link_type = 'dependent'

			UNION

			-- Recursive case: dependencies of dependencies
			SELECT tl.target_task_id, dc.depth + 1
			FROM task_links tl
			INNER JOIN dependency_chain dc ON tl.source_task_id = dc.target_task_id
			WHERE tl.link_type = 'dependent' AND dc.depth < 50  -- Prevent infinite loops
		)
		SELECT DISTINCT target_task_id FROM dependency_chain`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependency chain: %w", err)
	}
	defer rows.Close()

	var chain []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan dependency: %w", err)
		}
		chain = append(chain, id)
	}
	return chain, nil
}

// GetBlockingTasks returns all tasks that are blocking a given task (tasks it depends on that aren't complete).
func (r *TaskLinkRepository) GetBlockingTasks(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT tl.target_task_id
		FROM task_links tl
		JOIN tasks t ON tl.target_task_id = t.id
		WHERE tl.source_task_id = $1
		  AND tl.link_type = 'dependent'
		  AND t.status NOT IN ('completed', 'cancelled')`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocking tasks: %w", err)
	}
	defer rows.Close()

	var blocking []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan blocking task: %w", err)
		}
		blocking = append(blocking, id)
	}
	return blocking, nil
}

// GetDependentTasks returns all tasks that depend on a given task.
func (r *TaskLinkRepository) GetDependentTasks(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT source_task_id
		FROM task_links
		WHERE target_task_id = $1 AND link_type = 'dependent'`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependent tasks: %w", err)
	}
	defer rows.Close()

	var dependents []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan dependent task: %w", err)
		}
		dependents = append(dependents, id)
	}
	return dependents, nil
}
