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

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, tier, max_plans, max_regenerations_per_day,
			can_export, can_version_history, can_refine,
			stripe_subscription_id, current_period_start, current_period_end,
			created_at, updated_at
		FROM subscriptions WHERE user_id = $1`

	sub := &models.Subscription{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.Tier, &sub.MaxPlans, &sub.MaxRegenerationsPerDay,
		&sub.CanExport, &sub.CanVersionHistory, &sub.CanRefine,
		&sub.StripeSubscriptionID, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	sub.ID = uuid.New()
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	query := `
		INSERT INTO subscriptions (
			id, user_id, tier, max_plans, max_regenerations_per_day,
			can_export, can_version_history, can_refine,
			stripe_subscription_id, current_period_start, current_period_end,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(ctx, query,
		sub.ID, sub.UserID, sub.Tier, sub.MaxPlans, sub.MaxRegenerationsPerDay,
		sub.CanExport, sub.CanVersionHistory, sub.CanRefine,
		sub.StripeSubscriptionID, sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
		sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	sub.UpdatedAt = time.Now()

	query := `
		UPDATE subscriptions SET
			tier = $2, max_plans = $3, max_regenerations_per_day = $4,
			can_export = $5, can_version_history = $6, can_refine = $7,
			stripe_subscription_id = $8, current_period_start = $9,
			current_period_end = $10, updated_at = $11
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		sub.ID, sub.Tier, sub.MaxPlans, sub.MaxRegenerationsPerDay,
		sub.CanExport, sub.CanVersionHistory, sub.CanRefine,
		sub.StripeSubscriptionID, sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd, sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) CountRegenerationsToday(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*) FROM generation_log
		WHERE user_id = $1 AND action = 'regenerate' AND created_at >= CURRENT_DATE`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count regenerations: %w", err)
	}
	return count, nil
}
