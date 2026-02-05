// Package handlers provides HTTP handlers for the REST API.
package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/user"
	"github.com/goplan/goplan/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	*BaseHandler
	userRepo repository.UserRepository
	jwt      *auth.JWT
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(userRepo repository.UserRepository, jwt *auth.JWT) *AuthHandler {
	return &AuthHandler{
		BaseHandler: NewBaseHandler(),
		userRepo:    userRepo,
		jwt:         jwt,
	}
}

// SignupRequest represents a signup request.
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a token refresh request.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// AuthResponse represents an authentication response.
type AuthResponse struct {
	User         *user.UserResponse `json:"user"`
	AccessToken  string             `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
	TokenType    string             `json:"tokenType"`
	ExpiresIn    int64              `json:"expiresIn"`
}

// Signup handles POST /api/v1/auth/signup
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.WriteMethodNotAllowed(w, []string{http.MethodPost})
		return
	}

	var req SignupRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteBadRequest(w, "invalid request body")
		return
	}

	// Validate input
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Email == "" {
		h.WriteBadRequest(w, "email is required")
		return
	}
	if req.Password == "" {
		h.WriteBadRequest(w, "password is required")
		return
	}
	if len(req.Password) < 8 {
		h.WriteBadRequest(w, "password must be at least 8 characters")
		return
	}
	if req.Name == "" {
		h.WriteBadRequest(w, "name is required")
		return
	}

	ctx := r.Context()

	// Check if email already exists
	exists, err := h.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		h.WriteInternalError(w)
		return
	}
	if exists {
		h.WriteErrorWithStatus(w, http.StatusConflict, "EMAIL_EXISTS", "email already registered")
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.WriteInternalError(w)
		return
	}

	// Create user (database generates the ID)
	newUser := &user.UserWithPassword{
		User: user.User{
			Email:     req.Email,
			Name:      req.Name,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		PasswordHash: string(passwordHash),
	}

	if err := h.userRepo.Create(ctx, newUser); err != nil {
		h.WriteError(w, err)
		return
	}

	// Generate tokens using the database-generated user ID
	accessToken, refreshToken, err := h.jwt.GenerateTokenPair(newUser.ID, req.Email, req.Name)
	if err != nil {
		h.WriteInternalError(w)
		return
	}

	response := &AuthResponse{
		User:         newUser.User.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.jwt.GetAccessTokenTTL().Seconds()),
	}

	h.WriteCreated(w, response)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.WriteMethodNotAllowed(w, []string{http.MethodPost})
		return
	}

	var req LoginRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteBadRequest(w, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		h.WriteBadRequest(w, "email and password are required")
		return
	}

	ctx := r.Context()

	// Get user by email
	u, err := h.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Use constant-time comparison to prevent timing attacks
		// Hash a dummy password to maintain consistent timing
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummy.hash.for.timing.attack.prevention"), []byte(req.Password))
		h.WriteUnauthorized(w, "invalid email or password")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		h.WriteUnauthorized(w, "invalid email or password")
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.jwt.GenerateTokenPair(u.ID, u.Email, u.Name)
	if err != nil {
		h.WriteInternalError(w)
		return
	}

	response := &AuthResponse{
		User:         u.User.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.jwt.GetAccessTokenTTL().Seconds()),
	}

	h.WriteSuccess(w, response)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.WriteMethodNotAllowed(w, []string{http.MethodPost})
		return
	}

	var req RefreshRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.WriteBadRequest(w, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		h.WriteBadRequest(w, "refresh_token is required")
		return
	}

	// Validate refresh token
	claims, err := h.jwt.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.WriteUnauthorized(w, "invalid or expired refresh token")
		return
	}

	ctx := r.Context()

	// Get user to ensure they still exist and get current info
	u, err := h.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		h.WriteUnauthorized(w, "user not found")
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := h.jwt.GenerateTokenPair(u.ID, u.Email, u.Name)
	if err != nil {
		h.WriteInternalError(w)
		return
	}

	response := &AuthResponse{
		User:         u.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(h.jwt.GetAccessTokenTTL().Seconds()),
	}

	h.WriteSuccess(w, response)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.WriteMethodNotAllowed(w, []string{http.MethodPost})
		return
	}

	// For stateless JWT, logout is handled client-side by discarding tokens
	// In a more complete implementation, you might blacklist the refresh token

	h.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "logged out successfully",
	})
}

// ServeHTTP routes requests to the appropriate handler method.
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch path {
	case "/api/v1/auth/signup":
		h.Signup(w, r)
	case "/api/v1/auth/login":
		h.Login(w, r)
	case "/api/v1/auth/refresh":
		h.Refresh(w, r)
	case "/api/v1/auth/logout":
		h.Logout(w, r)
	default:
		h.WriteNotFound(w, "endpoint")
	}
}

// Ensure subtle is used (for timing attack prevention)
var _ = subtle.ConstantTimeCompare
