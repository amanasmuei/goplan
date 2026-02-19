package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
)

// SubscriptionHandler handles HTTP requests for subscription operations.
type SubscriptionHandler struct {
	subRepo *repository.SubscriptionRepository
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(subRepo *repository.SubscriptionRepository) *SubscriptionHandler {
	return &SubscriptionHandler{subRepo: subRepo}
}

// GetSubscription returns the current user's subscription details.
// @Summary Get subscription
// @Description Get the current user's subscription tier and limits
// @Tags Subscription
// @Produce json
// @Success 200 {object} models.Subscription
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/subscription [get]
func (h *SubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	userID, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	sub, err := h.subRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "Failed to get subscription"})
	}

	if sub == nil {
		defaults := models.TierLimits(models.SubscriptionTierFree)
		defaults.UserID = userID
		return c.JSON(defaults)
	}

	return c.JSON(sub)
}

// UpgradeSubscription is a stub for future Stripe integration.
// @Summary Upgrade subscription
// @Description Initiate a subscription upgrade (Stripe integration coming soon)
// @Tags Subscription
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/subscription/upgrade [post]
func (h *SubscriptionHandler) UpgradeSubscription(c *fiber.Ctx) error {
	_, _, err := getUserContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Stripe integration coming soon",
		"tiers": fiber.Map{
			"pro": fiber.Map{
				"price":                    "$19/month",
				"max_plans":                25,
				"max_regenerations_per_day": 20,
				"can_export":               true,
				"can_version_history":       true,
				"can_refine":               true,
			},
			"pro_plus": fiber.Map{
				"price":                    "$49/month",
				"max_plans":                100,
				"max_regenerations_per_day": 50,
				"can_export":               true,
				"can_version_history":       true,
				"can_refine":               true,
			},
		},
	})
}
