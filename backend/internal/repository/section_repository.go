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

type SectionRepository struct {
	db *pgxpool.Pool
}

func NewSectionRepository(db *pgxpool.Pool) *SectionRepository {
	return &SectionRepository{db: db}
}

func (r *SectionRepository) CreateBatch(ctx context.Context, sections []models.PlanSection) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO plan_sections (
			id, plan_id, section_type, section_order, title,
			content, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	now := time.Now()
	for i := range sections {
		sections[i].ID = uuid.New()
		sections[i].CreatedAt = now
		sections[i].UpdatedAt = now
		if sections[i].Version == 0 {
			sections[i].Version = 1
		}

		_, err := tx.Exec(ctx, query,
			sections[i].ID, sections[i].PlanID, sections[i].SectionType,
			sections[i].SectionOrder, sections[i].Title,
			sections[i].Content, sections[i].Version,
			sections[i].CreatedAt, sections[i].UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert section %d: %w", i, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit sections batch: %w", err)
	}
	return nil
}

func (r *SectionRepository) GetByPlanID(ctx context.Context, planID uuid.UUID) ([]models.PlanSection, error) {
	query := `
		SELECT id, plan_id, section_type, section_order, title,
			content, version, created_at, updated_at
		FROM plan_sections WHERE plan_id = $1
		ORDER BY section_order`

	rows, err := r.db.Query(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sections by plan: %w", err)
	}
	defer rows.Close()

	var sections []models.PlanSection
	for rows.Next() {
		var s models.PlanSection
		err := rows.Scan(
			&s.ID, &s.PlanID, &s.SectionType, &s.SectionOrder, &s.Title,
			&s.Content, &s.Version, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan section: %w", err)
		}
		sections = append(sections, s)
	}

	return sections, nil
}

func (r *SectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.PlanSection, error) {
	query := `
		SELECT id, plan_id, section_type, section_order, title,
			content, version, created_at, updated_at
		FROM plan_sections WHERE id = $1`

	s := &models.PlanSection{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.PlanID, &s.SectionType, &s.SectionOrder, &s.Title,
		&s.Content, &s.Version, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get section: %w", err)
	}
	return s, nil
}

func (r *SectionRepository) Update(ctx context.Context, section *models.PlanSection) error {
	section.UpdatedAt = time.Now()

	query := `
		UPDATE plan_sections SET
			title = $2, content = $3, version = $4, updated_at = $5
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		section.ID, section.Title, section.Content,
		section.Version, section.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}
	return nil
}

func (r *SectionRepository) GetByPlanAndType(ctx context.Context, planID uuid.UUID, sectionType models.SectionType) (*models.PlanSection, error) {
	query := `
		SELECT id, plan_id, section_type, section_order, title,
			content, version, created_at, updated_at
		FROM plan_sections WHERE plan_id = $1 AND section_type = $2`

	s := &models.PlanSection{}
	err := r.db.QueryRow(ctx, query, planID, sectionType).Scan(
		&s.ID, &s.PlanID, &s.SectionType, &s.SectionOrder, &s.Title,
		&s.Content, &s.Version, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get section by plan and type: %w", err)
	}
	return s, nil
}
