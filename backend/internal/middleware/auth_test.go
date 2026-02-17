package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goplan/backend/internal/config"
	"github.com/goplan/backend/internal/models"
)

const testSecret = "test-secret-key-for-unit-tests-minimum-32"

func testJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		Secret:          testSecret,
		ExpirationHours: 24,
	}
}

// generateValidToken creates a valid JWT token for testing.
func generateValidToken(userID, orgID uuid.UUID, email string, role models.UserRole) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         userID.String(),
		"email":           email,
		"role":            string(role),
		"organization_id": orgID.String(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
		"type":            "access",
		"iss":             "goplan",
		"iat":             time.Now().Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))
	return tokenString
}

// generateExpiredToken creates an expired JWT token.
func generateExpiredToken(userID, orgID uuid.UUID, email string, role models.UserRole) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         userID.String(),
		"email":           email,
		"role":            string(role),
		"organization_id": orgID.String(),
		"exp":             time.Now().Add(-1 * time.Hour).Unix(),
		"type":            "access",
		"iss":             "goplan",
		"iat":             time.Now().Add(-2 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))
	return tokenString
}

// parseJSONResponse parses a JSON response body.
func parseJSONResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

// --- Authenticate middleware tests ---

func TestAuthenticate_MissingAuthorizationHeader(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	result := parseJSONResponse(t, resp)
	assert.Equal(t, "Missing authorization header", result["error"])
}

func TestAuthenticate_InvalidBearerFormat(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "just-a-token"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"single word", "Token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.header)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			result := parseJSONResponse(t, resp)
			assert.Equal(t, "Invalid authorization header format", result["error"])
		})
	}
}

func TestAuthenticate_ValidToken(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	userID := uuid.New()
	orgID := uuid.New()
	email := "test@example.com"
	role := models.UserRoleAdmin

	var capturedUserID uuid.UUID
	var capturedOrgID uuid.UUID
	var capturedEmail string
	var capturedRole models.UserRole

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		capturedUserID = c.Locals("user_id").(uuid.UUID)
		capturedOrgID = c.Locals("organization_id").(uuid.UUID)
		capturedEmail = c.Locals("email").(string)
		capturedRole = c.Locals("role").(models.UserRole)
		return c.SendString("ok")
	})

	token := generateValidToken(userID, orgID, email, role)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, userID, capturedUserID)
	assert.Equal(t, orgID, capturedOrgID)
	assert.Equal(t, email, capturedEmail)
	assert.Equal(t, role, capturedRole)
}

func TestAuthenticate_ExpiredToken(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	userID := uuid.New()
	orgID := uuid.New()
	token := generateExpiredToken(userID, orgID, "test@example.com", models.UserRoleMember)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	result := parseJSONResponse(t, resp)
	assert.Equal(t, "Invalid or expired token", result["error"])
}

func TestAuthenticate_WrongSigningMethod(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Create token with "none" signing method (alg: none attack)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id":         uuid.New().String(),
		"email":           "attacker@example.com",
		"role":            "admin",
		"organization_id": uuid.New().String(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticate_WrongSecret(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Sign with a different secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         uuid.New().String(),
		"email":           "test@example.com",
		"role":            "admin",
		"organization_id": uuid.New().String(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("completely-different-secret-key-1234"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticate_MissingUserIDClaim(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Token without user_id claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":           "test@example.com",
		"role":            "admin",
		"organization_id": uuid.New().String(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticate_InvalidUserIDFormat(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Token with non-UUID user_id
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         "not-a-uuid",
		"email":           "test@example.com",
		"role":            "admin",
		"organization_id": uuid.New().String(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	result := parseJSONResponse(t, resp)
	assert.Equal(t, "Invalid user ID in token", result["error"])
}

func TestAuthenticate_MissingOrganizationIDClaim(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Token without organization_id claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uuid.New().String(),
		"email":   "test@example.com",
		"role":    "admin",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticate_MissingEmailClaim(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Token without email
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         uuid.New().String(),
		"role":            "admin",
		"organization_id": uuid.New().String(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthenticate_MissingRoleClaim(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Token without role
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         uuid.New().String(),
		"email":           "test@example.com",
		"organization_id": uuid.New().String(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// --- RequireRole middleware tests ---

func TestRequireRole_NoRoleInContext(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.RequireRole(models.UserRoleAdmin))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	result := parseJSONResponse(t, resp)
	assert.Equal(t, "Access denied", result["error"])
}

func TestRequireRole_AllowedRoles(t *testing.T) {
	tests := []struct {
		name          string
		userRole      models.UserRole
		allowedRoles  []models.UserRole
		expectAllowed bool
	}{
		{
			name:          "admin allowed for admin route",
			userRole:      models.UserRoleAdmin,
			allowedRoles:  []models.UserRole{models.UserRoleAdmin},
			expectAllowed: true,
		},
		{
			name:          "member denied for admin route",
			userRole:      models.UserRoleMember,
			allowedRoles:  []models.UserRole{models.UserRoleAdmin},
			expectAllowed: false,
		},
		{
			name:          "member allowed for member+admin route",
			userRole:      models.UserRoleMember,
			allowedRoles:  []models.UserRole{models.UserRoleAdmin, models.UserRoleMember},
			expectAllowed: true,
		},
		{
			name:          "team_lead allowed for team_lead route",
			userRole:      models.UserRoleTeamLead,
			allowedRoles:  []models.UserRole{models.UserRoleTeamLead, models.UserRoleAdmin},
			expectAllowed: true,
		},
		{
			name:          "member denied for team_lead only",
			userRole:      models.UserRoleMember,
			allowedRoles:  []models.UserRole{models.UserRoleTeamLead},
			expectAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			m := NewAuthMiddleware(testJWTConfig())

			// Inject role into Locals to simulate post-auth state
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("role", tt.userRole)
				return c.Next()
			})
			app.Use(m.RequireRole(tt.allowedRoles...))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			if tt.expectAllowed {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusForbidden, resp.StatusCode)
				result := parseJSONResponse(t, resp)
				assert.Equal(t, "Insufficient permissions", result["error"])
			}
		})
	}
}

func TestRequireRole_WrongTypeInContext(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	// Set role as string instead of models.UserRole
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // wrong type: string, not models.UserRole
		return c.Next()
	})
	app.Use(m.RequireRole(models.UserRoleAdmin))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// --- GenerateToken tests ---

func TestGenerateToken_ProducesValidToken(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()
	claims := &models.AuthClaims{
		UserID:         userID,
		Email:          "test@example.com",
		Role:           models.UserRoleAdmin,
		OrganizationID: orgID,
	}

	tokenString, err := GenerateToken(claims, testSecret, 24)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Parse and validate the generated token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatal("unexpected signing method")
		}
		return []byte(testSecret), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, userID.String(), mapClaims["user_id"])
	assert.Equal(t, "test@example.com", mapClaims["email"])
	assert.Equal(t, "admin", mapClaims["role"])
	assert.Equal(t, orgID.String(), mapClaims["organization_id"])
	assert.Equal(t, "access", mapClaims["type"])
	assert.Equal(t, "goplan", mapClaims["iss"])
}

func TestGenerateToken_ExpiresAtCorrectTime(t *testing.T) {
	claims := &models.AuthClaims{
		UserID:         uuid.New(),
		Email:          "test@example.com",
		Role:           models.UserRoleMember,
		OrganizationID: uuid.New(),
	}

	expirationHours := 2
	tokenString, err := GenerateToken(claims, testSecret, expirationHours)
	require.NoError(t, err)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	require.NoError(t, err)

	mapClaims := token.Claims.(jwt.MapClaims)
	exp := int64(mapClaims["exp"].(float64))
	iat := int64(mapClaims["iat"].(float64))

	expectedDuration := int64(expirationHours * 3600)
	actualDuration := exp - iat

	// Allow 5 second tolerance for test execution time
	assert.InDelta(t, expectedDuration, actualDuration, 5)
}

// --- Integration: Authenticate + RequireRole chained ---

func TestAuthenticate_ThenRequireRole(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	tests := []struct {
		name           string
		role           models.UserRole
		requiredRoles  []models.UserRole
		expectedStatus int
	}{
		{
			name:           "admin passes admin check",
			role:           models.UserRoleAdmin,
			requiredRoles:  []models.UserRole{models.UserRoleAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "member fails admin check",
			role:           models.UserRoleMember,
			requiredRoles:  []models.UserRole{models.UserRoleAdmin},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "member passes member+admin check",
			role:           models.UserRoleMember,
			requiredRoles:  []models.UserRole{models.UserRoleAdmin, models.UserRoleMember},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			m := NewAuthMiddleware(testJWTConfig())

			app.Use(m.Authenticate())
			app.Use(m.RequireRole(tt.requiredRoles...))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			token := generateValidToken(userID, orgID, "test@example.com", tt.role)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestAuthenticate_NoTokenThenRequireRole(t *testing.T) {
	app := fiber.New()
	m := NewAuthMiddleware(testJWTConfig())

	app.Use(m.Authenticate())
	app.Use(m.RequireRole(models.UserRoleAdmin))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should be stopped at Authenticate, not RequireRole
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
