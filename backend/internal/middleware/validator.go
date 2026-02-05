package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ValidationMiddleware struct {
	validate *validator.Validate
}

func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validate: validator.New(),
	}
}

func (m *ValidationMiddleware) ValidateBody(payload interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.BodyParser(payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if err := m.validate.Struct(payload); err != nil {
			validationErrors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				validationErrors[err.Field()] = getValidationMessage(err)
			}
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": validationErrors,
			})
		}

		c.Locals("validated_body", payload)
		return c.Next()
	}
}

func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value is too short (minimum: " + err.Param() + ")"
	case "max":
		return "Value is too long (maximum: " + err.Param() + ")"
	case "email":
		return "Invalid email format"
	case "uuid":
		return "Invalid UUID format"
	case "oneof":
		return "Value must be one of: " + err.Param()
	case "gt":
		return "Value must be greater than " + err.Param()
	default:
		return "Invalid value"
	}
}

// Validate validates a struct
func Validate(payload interface{}) error {
	validate := validator.New()
	return validate.Struct(payload)
}
