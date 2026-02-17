// Package auth provides JWT authentication for the GoPlan backend.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Standard JWT errors.
var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenNotYetValid = errors.New("token not yet valid")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrMissingClaims    = errors.New("missing required claims")
)

// TokenType represents the type of JWT token.
type TokenType string

const (
	// TokenTypeAccess represents an access token.
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh represents a refresh token.
	TokenTypeRefresh TokenType = "refresh"
)

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// DefaultJWTConfig returns the default JWT configuration.
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "goplan",
	}
}

// Claims represents JWT claims.
type Claims struct {
	// Standard claims
	Subject   string    `json:"sub"`
	Issuer    string    `json:"iss"`
	IssuedAt  int64     `json:"iat"`
	ExpiresAt int64     `json:"exp"`
	NotBefore int64     `json:"nbf,omitempty"`
	TokenID   string    `json:"jti,omitempty"`

	// Custom claims
	TokenType   TokenType `json:"type"`
	UserID      string    `json:"user_id"`
	Email       string    `json:"email,omitempty"`
	Name        string    `json:"name,omitempty"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	Role        string    `json:"role,omitempty"`
}

// Valid checks if the claims are valid.
func (c *Claims) Valid() error {
	now := time.Now().Unix()

	if c.ExpiresAt != 0 && now > c.ExpiresAt {
		return ErrTokenExpired
	}

	if c.NotBefore != 0 && now < c.NotBefore {
		return ErrTokenNotYetValid
	}

	if c.UserID == "" {
		return ErrMissingClaims
	}

	return nil
}

// JWT provides JWT token generation and validation.
type JWT struct {
	config JWTConfig
}

// NewJWT creates a new JWT instance.
func NewJWT(config JWTConfig) *JWT {
	return &JWT{config: config}
}

// GenerateAccessToken generates a new access token for a user.
func (j *JWT) GenerateAccessToken(userID, email, name string) (string, error) {
	now := time.Now()
	claims := &Claims{
		Subject:   userID,
		Issuer:    j.config.Issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(j.config.AccessTokenTTL).Unix(),
		TokenID:   uuid.New().String(),
		TokenType: TokenTypeAccess,
		UserID:    userID,
		Email:     email,
		Name:      name,
	}
	return j.generateToken(claims)
}

// GenerateRefreshToken generates a new refresh token for a user.
func (j *JWT) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := &Claims{
		Subject:   userID,
		Issuer:    j.config.Issuer,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(j.config.RefreshTokenTTL).Unix(),
		TokenID:   uuid.New().String(),
		TokenType: TokenTypeRefresh,
		UserID:    userID,
	}
	return j.generateToken(claims)
}

// GenerateTokenPair generates both access and refresh tokens.
func (j *JWT) GenerateTokenPair(userID, email, name string) (accessToken, refreshToken string, err error) {
	accessToken, err = j.GenerateAccessToken(userID, email, name)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = j.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken validates a token and returns its claims.
func (j *JWT) ValidateToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Decode header (we don't use it but validate structure)
	_, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Verify signature
	expectedSig := j.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, ErrInvalidSignature
	}

	// Parse claims
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Validate claims
	if err := claims.Valid(); err != nil {
		return nil, err
	}

	return &claims, nil
}

// ValidateAccessToken validates an access token.
func (j *JWT) ValidateAccessToken(token string) (*Claims, error) {
	claims, err := j.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeAccess {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token.
func (j *JWT) ValidateRefreshToken(token string) (*Claims, error) {
	claims, err := j.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshTokens refreshes the access token using a valid refresh token.
func (j *JWT) RefreshTokens(refreshToken string) (newAccessToken, newRefreshToken string, err error) {
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Email, claims.Name)
}

// GetAccessTokenTTL returns the access token TTL.
func (j *JWT) GetAccessTokenTTL() time.Duration {
	return j.config.AccessTokenTTL
}

// GetRefreshTokenTTL returns the refresh token TTL.
func (j *JWT) GetRefreshTokenTTL() time.Duration {
	return j.config.RefreshTokenTTL
}

// generateToken generates a JWT token from claims.
func (j *JWT) generateToken(claims *Claims) (string, error) {
	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Create payload
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Create signature
	message := headerEncoded + "." + payloadEncoded
	signature := j.sign(message)

	return message + "." + signature, nil
}

// sign creates an HMAC-SHA256 signature.
func (j *JWT) sign(message string) string {
	h := hmac.New(sha256.New, []byte(j.config.Secret))
	h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// ExtractBearerToken extracts the token from a Bearer authorization header.
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// TokenResponse represents a token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// NewTokenResponse creates a new token response.
func NewTokenResponse(accessToken, refreshToken string, expiresIn time.Duration) *TokenResponse {
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(expiresIn.Seconds()),
	}
}
