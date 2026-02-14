package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/goplan/backend/internal/config"
	"github.com/goplan/backend/internal/models"
	"github.com/goplan/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtConfig *config.JWTConfig
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtConfig *config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=255"`
}

type AuthResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

type UserResponse struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	OrganizationID string `json:"organization_id"`
}

// Login authenticates a user and returns a JWT token
// @Summary User login
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Get user by email
	user, passwordHash, err := h.userRepo.GetByEmailWithPassword(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to authenticate",
		})
	}
	if user == nil {
		// Prevent timing attacks - perform dummy hash comparison
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"), []byte(req.Password))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Verify password
	if passwordHash == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Account not set up for password authentication",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Generate JWT token
	now := time.Now()
	expiresAt := now.Add(time.Duration(h.jwtConfig.ExpirationHours) * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"role":            string(user.Role),
		"organization_id": user.OrganizationID.String(),
		"exp":             expiresAt.Unix(),
		"type":            "access",
		"iss":             "goplan",
		"iat":             now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtConfig.Secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(AuthResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User: UserResponse{
			ID:             user.ID.String(),
			Email:          user.Email,
			Name:           user.Name,
			Role:           string(user.Role),
			OrganizationID: user.OrganizationID.String(),
		},
	})
}

// Register creates a new user account
// @Summary User registration
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, password, and name are required",
		})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 8 characters",
		})
	}

	if len(req.Password) > 72 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must not exceed 72 characters",
		})
	}

	// Check if email already exists
	exists, err := h.userRepo.EmailExists(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check email",
		})
	}
	if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already registered",
		})
	}

	// Get default organization
	orgID, err := h.userRepo.GetDefaultOrganization(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No organization available",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process password",
		})
	}

	// Create user
	user := &models.User{
		Email:          req.Email,
		Name:           req.Name,
		Role:           models.UserRoleMember,
		OrganizationID: orgID,
	}

	if err := h.userRepo.Create(c.Context(), user, string(hashedPassword)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	// Generate JWT token
	now := time.Now()
	expiresAt := now.Add(time.Duration(h.jwtConfig.ExpirationHours) * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":         user.ID.String(),
		"email":           user.Email,
		"role":            string(user.Role),
		"organization_id": user.OrganizationID.String(),
		"exp":             expiresAt.Unix(),
		"type":            "access",
		"iss":             "goplan",
		"iat":             now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtConfig.Secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(AuthResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User: UserResponse{
			ID:             user.ID.String(),
			Email:          user.Email,
			Name:           user.Name,
			Role:           string(user.Role),
			OrganizationID: user.OrganizationID.String(),
		},
	})
}

// GetMe returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	userID := getUserID(c)
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	user, err := h.userRepo.GetByID(c.Context(), *userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(UserResponse{
		ID:             user.ID.String(),
		Email:          user.Email,
		Name:           user.Name,
		Role:           string(user.Role),
		OrganizationID: user.OrganizationID.String(),
	})
}

// ListUsers returns all users in the organization
// @Summary List users
// @Description Get all users in the organization
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {array} UserResponse
// @Failure 401 {object} map[string]string
// @Router /users [get]
func (h *AuthHandler) ListUsers(c *fiber.Ctx) error {
	orgID := getOrganizationID(c)
	if orgID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)
	users, err := h.userRepo.ListByOrganization(c.Context(), *orgID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list users",
		})
	}

	var response []UserResponse
	for _, user := range users {
		response = append(response, UserResponse{
			ID:             user.ID.String(),
			Email:          user.Email,
			Name:           user.Name,
			Role:           string(user.Role),
			OrganizationID: user.OrganizationID.String(),
		})
	}

	return c.JSON(fiber.Map{
		"users": response,
		"total": len(response),
	})
}
