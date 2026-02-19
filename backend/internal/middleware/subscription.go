package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

// SubscriptionMiddleware provides middleware for checking subscription-based feature access.
type SubscriptionMiddleware struct {
	subRepo  *repository.SubscriptionRepository
	planRepo *repository.PlanRepository
}

// NewSubscriptionMiddleware creates a new SubscriptionMiddleware.
func NewSubscriptionMiddleware(subRepo *repository.SubscriptionRepository, planRepo *repository.PlanRepository) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		subRepo:  subRepo,
		planRepo: planRepo,
	}
}

// RequireSubscription checks that the user's subscription tier allows the given feature.
// Supported features: "regeneration", "refine", "export", "version_history".
func (m *SubscriptionMiddleware) RequireSubscription(feature string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDVal := c.Locals("user_id")
		if userIDVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user context",
			})
		}

		sub, err := m.subRepo.GetByUserID(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check subscription",
			})
		}

		if sub == nil {
			defaults := models.TierLimits(models.SubscriptionTierFree)
			sub = &defaults
		}

		allowed := false
		switch feature {
		case "regeneration":
			// Regeneration is allowed on all tiers, but with daily limits.
			// The service layer checks daily limits, so we just pass through.
			allowed = true
		case "refine":
			allowed = sub.CanRefine
		case "export":
			allowed = sub.CanExport
		case "version_history":
			allowed = sub.CanVersionHistory
		default:
			allowed = false
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "This feature requires a Pro subscription or higher",
				"feature": feature,
				"upgrade": "/api/v1/subscription/upgrade",
			})
		}

		return c.Next()
	}
}

// CheckPlanLimit checks if the user can create more plans based on their subscription.
func (m *SubscriptionMiddleware) CheckPlanLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDVal := c.Locals("user_id")
		if userIDVal == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user context",
			})
		}

		sub, err := m.subRepo.GetByUserID(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check subscription",
			})
		}

		if sub == nil {
			defaults := models.TierLimits(models.SubscriptionTierFree)
			sub = &defaults
		}

		// -1 means unlimited
		if sub.MaxPlans == -1 {
			return c.Next()
		}

		count, err := m.planRepo.CountByUser(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check plan count",
			})
		}

		if count >= sub.MaxPlans {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Plan limit reached. Upgrade your subscription for more plans.",
				"current": count,
				"limit":   sub.MaxPlans,
				"upgrade": "/api/v1/subscription/upgrade",
			})
		}

		return c.Next()
	}
}
