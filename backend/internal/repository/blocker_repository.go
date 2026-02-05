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

type BlockerRepository struct {
	db *pgxpool.Pool
}

func NewBlockerRepository(db *pgxpool.Pool) *BlockerRepository {
	return &BlockerRepository{db: db}
}

func (r *BlockerRepository) Create(ctx context.Context, b *models.TaskBlocker) error {
	b.ID = uuid.New()
	b.CreatedAt = time.Now()

	query := `
		INSERT INTO task_blockers (id, task_id, blocker_type, description, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		b.ID, b.TaskID, b.BlockerType, b.Description, b.CreatedBy, b.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create blocker: %w", err)
	}
	return nil
}

func (r *BlockerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.TaskBlocker, error) {
	query := `
		SELECT id, task_id, blocker_type, description, resolved_at, days_blocked, created_by, created_at
		FROM task_blockers WHERE id = $1`

	b := &models.TaskBlocker{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.TaskID, &b.BlockerType, &b.Description, &b.ResolvedAt,
		&b.DaysBlocked, &b.CreatedBy, &b.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get blocker: %w", err)
	}
	return b, nil
}

func (r *BlockerRepository) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]models.TaskBlocker, error) {
	query := `
		SELECT id, task_id, blocker_type, description, resolved_at, days_blocked, created_by, created_at
		FROM task_blockers WHERE task_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list blockers: %w", err)
	}
	defer rows.Close()

	var blockers []models.TaskBlocker
	for rows.Next() {
		var b models.TaskBlocker
		err := rows.Scan(
			&b.ID, &b.TaskID, &b.BlockerType, &b.Description, &b.ResolvedAt,
			&b.DaysBlocked, &b.CreatedBy, &b.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blocker: %w", err)
		}
		blockers = append(blockers, b)
	}
	return blockers, nil
}

func (r *BlockerRepository) Resolve(ctx context.Context, id uuid.UUID, daysBlocked float64) error {
	now := time.Now()
	query := `UPDATE task_blockers SET resolved_at = $2, days_blocked = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, now, daysBlocked)
	if err != nil {
		return fmt.Errorf("failed to resolve blocker: %w", err)
	}
	return nil
}

func (r *BlockerRepository) GetBlockerPatterns(ctx context.Context, orgID uuid.UUID, limit int) (map[models.BlockerType]int, error) {
	query := `
		SELECT blocker_type, COUNT(*) as count
		FROM task_blockers tb
		JOIN tasks t ON tb.task_id = t.id
		WHERE t.organization_id = $1
		GROUP BY blocker_type
		ORDER BY count DESC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, orgID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get blocker patterns: %w", err)
	}
	defer rows.Close()

	patterns := make(map[models.BlockerType]int)
	for rows.Next() {
		var blockerType models.BlockerType
		var count int
		if err := rows.Scan(&blockerType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan blocker pattern: %w", err)
		}
		patterns[blockerType] = count
	}
	return patterns, nil
}
