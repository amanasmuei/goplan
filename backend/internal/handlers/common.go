package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func getUserContext(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userIDStr := c.Locals("user_id")
	orgIDStr := c.Locals("organization_id")

	if userIDStr == nil || orgIDStr == nil {
		return uuid.Nil, uuid.Nil, errors.New("unauthorized: missing user context")
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		return uuid.Nil, uuid.Nil, errors.New("invalid user ID")
	}

	orgID, ok := orgIDStr.(uuid.UUID)
	if !ok {
		return uuid.Nil, uuid.Nil, errors.New("invalid organization ID")
	}

	return userID, orgID, nil
}

func getUserID(c *fiber.Ctx) *uuid.UUID {
	userIDVal := c.Locals("user_id")
	if userIDVal == nil {
		return nil
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return nil
	}
	return &userID
}

func getOrganizationID(c *fiber.Ctx) *uuid.UUID {
	orgIDVal := c.Locals("organization_id")
	if orgIDVal == nil {
		return nil
	}
	orgID, ok := orgIDVal.(uuid.UUID)
	if !ok {
		return nil
	}
	return &orgID
}
