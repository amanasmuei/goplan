package auth

import (
	"testing"
	"time"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	config := JWTConfig{
		Secret:          "test-secret-key-minimum-32-chars!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	}

	jwt := NewJWT(config)

	t.Run("Generate and validate access token", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		name := "Test User"

		token, err := jwt.GenerateAccessToken(userID, email, name)
		if err != nil {
			t.Fatalf("Failed to generate access token: %v", err)
		}

		if token == "" {
			t.Fatal("Token should not be empty")
		}

		claims, err := jwt.ValidateAccessToken(token)
		if err != nil {
			t.Fatalf("Failed to validate access token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
		}
		if claims.Email != email {
			t.Errorf("Expected email %s, got %s", email, claims.Email)
		}
		if claims.Name != name {
			t.Errorf("Expected name %s, got %s", name, claims.Name)
		}
		if claims.TokenType != TokenTypeAccess {
			t.Errorf("Expected token type %s, got %s", TokenTypeAccess, claims.TokenType)
		}
	})

	t.Run("Generate and validate refresh token", func(t *testing.T) {
		userID := "user-456"

		token, err := jwt.GenerateRefreshToken(userID)
		if err != nil {
			t.Fatalf("Failed to generate refresh token: %v", err)
		}

		claims, err := jwt.ValidateRefreshToken(token)
		if err != nil {
			t.Fatalf("Failed to validate refresh token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
		}
		if claims.TokenType != TokenTypeRefresh {
			t.Errorf("Expected token type %s, got %s", TokenTypeRefresh, claims.TokenType)
		}
	})

	t.Run("Generate token pair", func(t *testing.T) {
		userID := "user-789"
		email := "pair@example.com"
		name := "Pair User"

		accessToken, refreshToken, err := jwt.GenerateTokenPair(userID, email, name)
		if err != nil {
			t.Fatalf("Failed to generate token pair: %v", err)
		}

		if accessToken == "" {
			t.Error("Access token should not be empty")
		}
		if refreshToken == "" {
			t.Error("Refresh token should not be empty")
		}

		// Validate both tokens
		accessClaims, err := jwt.ValidateAccessToken(accessToken)
		if err != nil {
			t.Fatalf("Failed to validate access token: %v", err)
		}
		if accessClaims.UserID != userID {
			t.Errorf("Expected user ID %s in access token, got %s", userID, accessClaims.UserID)
		}

		refreshClaims, err := jwt.ValidateRefreshToken(refreshToken)
		if err != nil {
			t.Fatalf("Failed to validate refresh token: %v", err)
		}
		if refreshClaims.UserID != userID {
			t.Errorf("Expected user ID %s in refresh token, got %s", userID, refreshClaims.UserID)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := jwt.ValidateToken("invalid.token.here")
		if err == nil {
			t.Error("Expected error for invalid token")
		}
	})

	t.Run("Wrong token type", func(t *testing.T) {
		// Generate refresh token but validate as access token
		token, _ := jwt.GenerateRefreshToken("user-wrong")
		_, err := jwt.ValidateAccessToken(token)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		// Create JWT with very short expiry
		shortConfig := JWTConfig{
			Secret:         "test-secret-key-minimum-32-chars!",
			AccessTokenTTL: -1 * time.Hour, // Already expired
			Issuer:         "test",
		}
		shortJWT := NewJWT(shortConfig)

		token, _ := shortJWT.GenerateAccessToken("user-expired", "", "")
		_, err := shortJWT.ValidateToken(token)
		if err != ErrTokenExpired {
			t.Errorf("Expected ErrTokenExpired, got %v", err)
		}
	})
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		wantToken   string
		wantErr     bool
	}{
		{
			name:      "Valid bearer token",
			header:    "Bearer token123",
			wantToken: "token123",
			wantErr:   false,
		},
		{
			name:      "Valid bearer token lowercase",
			header:    "bearer token123",
			wantToken: "token123",
			wantErr:   false,
		},
		{
			name:    "Empty header",
			header:  "",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			header:  "Basic token123",
			wantErr: true,
		},
		{
			name:    "No token value",
			header:  "Bearer",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractBearerToken(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if token != tt.wantToken {
					t.Errorf("Expected token %s, got %s", tt.wantToken, token)
				}
			}
		})
	}
}
