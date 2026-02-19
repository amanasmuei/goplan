package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionTier string

const (
	SubscriptionTierFree    SubscriptionTier = "free"
	SubscriptionTierPro     SubscriptionTier = "pro"
	SubscriptionTierProPlus SubscriptionTier = "pro_plus"
)

type Subscription struct {
	ID                      uuid.UUID        `json:"id" db:"id"`
	UserID                  uuid.UUID        `json:"user_id" db:"user_id"`
	Tier                    SubscriptionTier `json:"tier" db:"tier"`
	MaxPlans                int              `json:"max_plans" db:"max_plans"`
	MaxRegenerationsPerDay  int              `json:"max_regenerations_per_day" db:"max_regenerations_per_day"`
	CanExport               bool             `json:"can_export" db:"can_export"`
	CanVersionHistory       bool             `json:"can_version_history" db:"can_version_history"`
	CanRefine               bool             `json:"can_refine" db:"can_refine"`
	StripeSubscriptionID    *string          `json:"-" db:"stripe_subscription_id"`
	CurrentPeriodStart      *time.Time       `json:"current_period_start,omitempty" db:"current_period_start"`
	CurrentPeriodEnd        *time.Time       `json:"current_period_end,omitempty" db:"current_period_end"`
	CreatedAt               time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at" db:"updated_at"`
}

// TierLimits returns the default Subscription limits for a given tier.
func TierLimits(tier SubscriptionTier) Subscription {
	switch tier {
	case SubscriptionTierPro:
		return Subscription{
			Tier:                   SubscriptionTierPro,
			MaxPlans:               25,
			MaxRegenerationsPerDay: 20,
			CanExport:              true,
			CanVersionHistory:      true,
			CanRefine:              true,
		}
	case SubscriptionTierProPlus:
		return Subscription{
			Tier:                   SubscriptionTierProPlus,
			MaxPlans:               100,
			MaxRegenerationsPerDay: 50,
			CanExport:              true,
			CanVersionHistory:      true,
			CanRefine:              true,
		}
	default:
		return Subscription{
			Tier:                   SubscriptionTierFree,
			MaxPlans:               3,
			MaxRegenerationsPerDay: 5,
			CanExport:              false,
			CanVersionHistory:      false,
			CanRefine:              false,
		}
	}
}
