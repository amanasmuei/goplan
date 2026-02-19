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

type VersionRepository struct {
	db *pgxpool.Pool
}

func NewVersionRepository(db *pgxpool.Pool) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) CreateSectionVersion(ctx context.Context, v *models.SectionVersion) error {
	v.ID = uuid.New()
	v.CreatedAt = time.Now()

	query := `
		INSERT INTO section_versions (
			id, section_id, plan_id, version, content,
			refinement_context, generated_by, token_usage, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(ctx, query,
		v.ID, v.SectionID, v.PlanID, v.Version, v.Content,
		v.RefinementContext, v.GeneratedBy, v.TokenUsage, v.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create section version: %w", err)
	}
	return nil
}

func (r *VersionRepository) ListSectionVersions(ctx context.Context, sectionID uuid.UUID) ([]models.SectionVersion, error) {
	query := `
		SELECT id, section_id, plan_id, version, content,
			refinement_context, generated_by, token_usage, created_at
		FROM section_versions WHERE section_id = $1
		ORDER BY version DESC`

	rows, err := r.db.Query(ctx, query, sectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list section versions: %w", err)
	}
	defer rows.Close()

	var versions []models.SectionVersion
	for rows.Next() {
		var v models.SectionVersion
		err := rows.Scan(
			&v.ID, &v.SectionID, &v.PlanID, &v.Version, &v.Content,
			&v.RefinementContext, &v.GeneratedBy, &v.TokenUsage, &v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan section version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

func (r *VersionRepository) GetSectionVersion(ctx context.Context, sectionID uuid.UUID, version int) (*models.SectionVersion, error) {
	query := `
		SELECT id, section_id, plan_id, version, content,
			refinement_context, generated_by, token_usage, created_at
		FROM section_versions WHERE section_id = $1 AND version = $2`

	v := &models.SectionVersion{}
	err := r.db.QueryRow(ctx, query, sectionID, version).Scan(
		&v.ID, &v.SectionID, &v.PlanID, &v.Version, &v.Content,
		&v.RefinementContext, &v.GeneratedBy, &v.TokenUsage, &v.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get section version: %w", err)
	}
	return v, nil
}

func (r *VersionRepository) CreatePlanVersion(ctx context.Context, v *models.PlanVersion) error {
	v.ID = uuid.New()
	v.CreatedAt = time.Now()

	query := `
		INSERT INTO plan_versions (
			id, plan_id, version, snapshot, change_summary, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		v.ID, v.PlanID, v.Version, v.Snapshot, v.ChangeSummary, v.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create plan version: %w", err)
	}
	return nil
}

func (r *VersionRepository) ListPlanVersions(ctx context.Context, planID uuid.UUID) ([]models.PlanVersion, error) {
	query := `
		SELECT id, plan_id, version, snapshot, change_summary, created_at
		FROM plan_versions WHERE plan_id = $1
		ORDER BY version DESC`

	rows, err := r.db.Query(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to list plan versions: %w", err)
	}
	defer rows.Close()

	var versions []models.PlanVersion
	for rows.Next() {
		var v models.PlanVersion
		err := rows.Scan(
			&v.ID, &v.PlanID, &v.Version, &v.Snapshot, &v.ChangeSummary, &v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

func (r *VersionRepository) GetPlanVersion(ctx context.Context, planID uuid.UUID, version int) (*models.PlanVersion, error) {
	query := `
		SELECT id, plan_id, version, snapshot, change_summary, created_at
		FROM plan_versions WHERE plan_id = $1 AND version = $2`

	v := &models.PlanVersion{}
	err := r.db.QueryRow(ctx, query, planID, version).Scan(
		&v.ID, &v.PlanID, &v.Version, &v.Snapshot, &v.ChangeSummary, &v.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get plan version: %w", err)
	}
	return v, nil
}
