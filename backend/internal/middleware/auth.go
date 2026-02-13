package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/goplan/backend/internal/config"
	"github.com/goplan/backend/internal/models"
)

type AuthMiddleware struct {
	config *config.JWTConfig
}

func NewAuthMiddleware(cfg *config.JWTConfig) *AuthMiddleware {
	return &AuthMiddleware{config: cfg}
}

func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(m.config.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		userID, err := uuid.Parse(claims["user_id"].(string))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid user ID in token",
			})
		}

		orgID, err := uuid.Parse(claims["organization_id"].(string))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid organization ID in token",
			})
		}

		c.Locals("user_id", userID)
		c.Locals("organization_id", orgID)
		c.Locals("email", claims["email"])
		c.Locals("role", models.UserRole(claims["role"].(string)))

		return c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")
		if userRole == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		role := userRole.(models.UserRole)
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}

// GenerateToken creates a JWT token for the given claims
func GenerateToken(claims *models.AuthClaims, secret string, expirationHours int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         claims.UserID.String(),
		"email":           claims.Email,
		"role":            string(claims.Role),
		"organization_id": claims.OrganizationID.String(),
		"exp":             time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}
