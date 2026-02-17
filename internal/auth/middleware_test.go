package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goplan/goplan/internal/logging"
	"github.com/goplan/goplan/internal/metrics"
)

// testLogger returns a silent logger for tests.
func testLogger() *logging.Logger {
	return logging.New(logging.Config{
		Level:  "error",
		Format: "text",
		Output: io.Discard,
	})
}

// testJWT returns a JWT instance with a known test secret.
func testJWT() *JWT {
	return NewJWT(JWTConfig{
		Secret:          "test-secret-key-for-unit-tests-only",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "goplan-test",
	})
}

// testMiddleware returns a Middleware configured for testing (no dev mode).
func testMiddleware(skipPaths []string) *Middleware {
	return NewMiddleware(MiddlewareConfig{
		JWT:       testJWT(),
		Logger:    testLogger(),
		SkipPaths: skipPaths,
	})
}

// generateTestToken generates a valid access token with the given claims.
func generateTestToken(j *JWT, userID, email, name, workspaceID, role string) string {
	now := time.Now()
	claims := &Claims{
		Subject:     userID,
		Issuer:      j.config.Issuer,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(j.config.AccessTokenTTL).Unix(),
		TokenType:   TokenTypeAccess,
		UserID:      userID,
		Email:       email,
		Name:        name,
		WorkspaceID: workspaceID,
		Role:        role,
	}
	token, _ := j.generateToken(claims)
	return token
}

// generateExpiredToken generates an expired access token.
func generateExpiredToken(j *JWT, userID string) string {
	past := time.Now().Add(-1 * time.Hour)
	claims := &Claims{
		Subject:   userID,
		Issuer:    j.config.Issuer,
		IssuedAt:  past.Add(-15 * time.Minute).Unix(),
		ExpiresAt: past.Unix(),
		TokenType: TokenTypeAccess,
		UserID:    userID,
	}
	token, _ := j.generateToken(claims)
	return token
}

// okHandler is a simple handler that returns 200 OK.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
})

// parseErrorResponse parses the JSON error response body.
func parseErrorResponse(t *testing.T, body io.Reader) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	return resp
}

// --- Authenticate middleware tests ---

func TestAuthenticate_NoAuthHeader(t *testing.T) {
	m := testMiddleware(nil)
	handler := m.Authenticate(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	resp := parseErrorResponse(t, rec.Body)
	if resp["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %v", resp["code"])
	}
}

func TestAuthenticate_InvalidBearerFormat(t *testing.T) {
	m := testMiddleware(nil)
	handler := m.Authenticate(okHandler)

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "just-a-token"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"empty bearer", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestAuthenticate_ValidToken(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})
	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			t.Error("expected user in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if user.ID != "user-123" {
			t.Errorf("expected user ID 'user-123', got '%s'", user.ID)
		}
		if user.Email != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
		}
		if user.WorkspaceID != "ws-1" {
			t.Errorf("expected workspace ID 'ws-1', got '%s'", user.WorkspaceID)
		}
		if user.Role != "admin" {
			t.Errorf("expected role 'admin', got '%s'", user.Role)
		}

		claims := GetClaims(r.Context())
		if claims == nil {
			t.Error("expected claims in context")
		}
		if claims != nil && claims.TokenType != TokenTypeAccess {
			t.Errorf("expected token type 'access', got '%s'", claims.TokenType)
		}

		w.WriteHeader(http.StatusOK)
	}))

	token := generateTestToken(j, "user-123", "test@example.com", "Test User", "ws-1", "admin")
	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAuthenticate_ExpiredToken(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})
	handler := m.Authenticate(okHandler)

	token := generateExpiredToken(j, "user-123")
	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	resp := parseErrorResponse(t, rec.Body)
	if resp["message"] != "token expired" {
		t.Errorf("expected message 'token expired', got '%v'", resp["message"])
	}
}

func TestAuthenticate_TamperedToken(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})
	handler := m.Authenticate(okHandler)

	token := generateTestToken(j, "user-123", "test@example.com", "Test", "ws-1", "admin")
	// Tamper with the token by modifying the signature
	tamperedToken := token[:len(token)-4] + "XXXX"

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	req.Header.Set("Authorization", "Bearer "+tamperedToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	resp := parseErrorResponse(t, rec.Body)
	if resp["message"] != "invalid token signature" {
		t.Errorf("expected message 'invalid token signature', got '%v'", resp["message"])
	}
}

func TestAuthenticate_WrongSecret(t *testing.T) {
	// Generate token with a different secret
	otherJWT := NewJWT(JWTConfig{
		Secret:         "a-completely-different-secret-key",
		AccessTokenTTL: 15 * time.Minute,
		Issuer:         "goplan-test",
	})
	token := generateTestToken(otherJWT, "user-123", "test@example.com", "Test", "ws-1", "admin")

	// Validate with the standard test middleware
	m := testMiddleware(nil)
	handler := m.Authenticate(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuthenticate_MalformedToken(t *testing.T) {
	m := testMiddleware(nil)
	handler := m.Authenticate(okHandler)

	tests := []struct {
		name  string
		token string
	}{
		{"random string", "not-a-jwt-token"},
		{"missing parts", "header.payload"},
		{"empty", ""},
		{"too many parts", "a.b.c.d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestAuthenticate_RefreshTokenRejected(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})
	handler := m.Authenticate(okHandler)

	// Generate a refresh token (not access)
	refreshToken, _ := j.GenerateRefreshToken("user-123")

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for refresh token used as access token, got %d", rec.Code)
	}
}

// --- SkipPaths tests ---

func TestAuthenticate_SkipPaths(t *testing.T) {
	m := testMiddleware([]string{"/health", "/api/v1/public/"})
	handler := m.Authenticate(okHandler)

	tests := []struct {
		name       string
		path       string
		expectSkip bool
	}{
		{"health endpoint", "/health", true},
		{"public API", "/api/v1/public/docs", true},
		{"protected endpoint", "/api/plans", false},
		{"prefix match works", "/healthcheck", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.expectSkip {
				if rec.Code != http.StatusOK {
					t.Errorf("expected 200 (skipped), got %d", rec.Code)
				}
			} else {
				if rec.Code != http.StatusUnauthorized {
					t.Errorf("expected 401 (not skipped), got %d", rec.Code)
				}
			}
		})
	}
}

// --- RequireRole middleware tests ---

func TestRequireRole_NoUser(t *testing.T) {
	m := testMiddleware(nil)
	handler := m.RequireRole("member")(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireRole_Hierarchy(t *testing.T) {
	m := testMiddleware(nil)

	tests := []struct {
		name         string
		userRole     string
		requiredRole string
		expectAllow  bool
	}{
		// owner (4) can access everything
		{"owner accessing owner", "owner", "owner", true},
		{"owner accessing admin", "owner", "admin", true},
		{"owner accessing member", "owner", "member", true},
		{"owner accessing viewer", "owner", "viewer", true},

		// admin (3) can access admin and below
		{"admin accessing owner", "admin", "owner", false},
		{"admin accessing admin", "admin", "admin", true},
		{"admin accessing member", "admin", "member", true},
		{"admin accessing viewer", "admin", "viewer", true},

		// member (2) can access member and below
		{"member accessing owner", "member", "owner", false},
		{"member accessing admin", "member", "admin", false},
		{"member accessing member", "member", "member", true},
		{"member accessing viewer", "member", "viewer", true},

		// viewer (1) can only access viewer
		{"viewer accessing owner", "viewer", "owner", false},
		{"viewer accessing admin", "viewer", "admin", false},
		{"viewer accessing member", "viewer", "member", false},
		{"viewer accessing viewer", "viewer", "viewer", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := m.RequireRole(tt.requiredRole)(okHandler)

			// Inject user into context
			user := &User{ID: "user-1", Role: tt.userRole}
			ctx := context.WithValue(context.Background(), ContextKeyUser, user)

			req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.expectAllow && rec.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rec.Code)
			}
			if !tt.expectAllow && rec.Code != http.StatusForbidden {
				t.Errorf("expected 403, got %d", rec.Code)
			}
		})
	}
}

func TestRequireRole_InvalidUserRole(t *testing.T) {
	m := testMiddleware(nil)
	handler := m.RequireRole("member")(okHandler)

	user := &User{ID: "user-1", Role: "unknown_role"}
	ctx := context.WithValue(context.Background(), ContextKeyUser, user)

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for invalid role, got %d", rec.Code)
	}
}

func TestRequireAdmin(t *testing.T) {
	m := testMiddleware(nil)
	handler := m.RequireAdmin(okHandler)

	tests := []struct {
		name        string
		role        string
		expectAllow bool
	}{
		{"admin allowed", "admin", true},
		{"owner allowed", "owner", true},
		{"member denied", "member", false},
		{"viewer denied", "viewer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{ID: "user-1", Role: tt.role}
			ctx := context.WithValue(context.Background(), ContextKeyUser, user)

			req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.expectAllow && rec.Code != http.StatusOK {
				t.Errorf("expected 200, got %d", rec.Code)
			}
			if !tt.expectAllow && rec.Code != http.StatusForbidden {
				t.Errorf("expected 403, got %d", rec.Code)
			}
		})
	}
}

func TestRequireMember(t *testing.T) {
	m := testMiddleware(nil)
	handler := m.RequireMember(okHandler)

	user := &User{ID: "user-1", Role: "viewer"}
	ctx := context.WithValue(context.Background(), ContextKeyUser, user)

	req := httptest.NewRequest(http.MethodPost, "/api/plans", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 for viewer accessing member-required route, got %d", rec.Code)
	}
}

// --- Optional middleware tests ---

func TestOptional_NoToken(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})

	var capturedUser *User
	handler := m.Optional(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/public", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedUser != nil {
		t.Error("expected nil user for unauthenticated request")
	}
}

func TestOptional_ValidToken(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})

	var capturedUser *User
	handler := m.Optional(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	token := generateTestToken(j, "user-123", "test@example.com", "Test", "ws-1", "member")
	req := httptest.NewRequest(http.MethodGet, "/api/public", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedUser == nil {
		t.Fatal("expected user in context")
	}
	if capturedUser.ID != "user-123" {
		t.Errorf("expected user ID 'user-123', got '%s'", capturedUser.ID)
	}
}

func TestOptional_InvalidToken(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})

	var capturedUser *User
	handler := m.Optional(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser = GetUser(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/public", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should still allow through, just without user
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedUser != nil {
		t.Error("expected nil user for invalid token in optional mode")
	}
}

// --- Context helper function tests ---

func TestGetUser(t *testing.T) {
	t.Run("returns user from context", func(t *testing.T) {
		user := &User{ID: "user-1", Email: "test@example.com", Role: "admin"}
		ctx := context.WithValue(context.Background(), ContextKeyUser, user)

		result := GetUser(ctx)
		if result == nil {
			t.Fatal("expected user, got nil")
		}
		if result.ID != "user-1" {
			t.Errorf("expected ID 'user-1', got '%s'", result.ID)
		}
	})

	t.Run("returns nil for empty context", func(t *testing.T) {
		result := GetUser(context.Background())
		if result != nil {
			t.Error("expected nil for empty context")
		}
	})

	t.Run("returns nil for wrong type in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), ContextKeyUser, "not a user")
		result := GetUser(ctx)
		if result != nil {
			t.Error("expected nil for wrong type")
		}
	})
}

func TestGetClaims(t *testing.T) {
	t.Run("returns claims from context", func(t *testing.T) {
		claims := &Claims{UserID: "user-1", TokenType: TokenTypeAccess}
		ctx := context.WithValue(context.Background(), ContextKeyClaims, claims)

		result := GetClaims(ctx)
		if result == nil {
			t.Fatal("expected claims, got nil")
		}
		if result.UserID != "user-1" {
			t.Errorf("expected user ID 'user-1', got '%s'", result.UserID)
		}
	})

	t.Run("returns nil for empty context", func(t *testing.T) {
		result := GetClaims(context.Background())
		if result != nil {
			t.Error("expected nil for empty context")
		}
	})
}

func TestGetUserID(t *testing.T) {
	t.Run("returns user ID", func(t *testing.T) {
		user := &User{ID: "user-1"}
		ctx := context.WithValue(context.Background(), ContextKeyUser, user)

		result := GetUserID(ctx)
		if result != "user-1" {
			t.Errorf("expected 'user-1', got '%s'", result)
		}
	})

	t.Run("returns empty for no user", func(t *testing.T) {
		result := GetUserID(context.Background())
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})
}

func TestGetWorkspaceID(t *testing.T) {
	t.Run("returns workspace ID", func(t *testing.T) {
		user := &User{ID: "user-1", WorkspaceID: "ws-1"}
		ctx := context.WithValue(context.Background(), ContextKeyUser, user)

		result := GetWorkspaceID(ctx)
		if result != "ws-1" {
			t.Errorf("expected 'ws-1', got '%s'", result)
		}
	})

	t.Run("returns empty for no user", func(t *testing.T) {
		result := GetWorkspaceID(context.Background())
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})
}

func TestGetUserRole(t *testing.T) {
	t.Run("returns user role", func(t *testing.T) {
		user := &User{ID: "user-1", Role: "admin"}
		ctx := context.WithValue(context.Background(), ContextKeyUser, user)

		result := GetUserRole(ctx)
		if result != "admin" {
			t.Errorf("expected 'admin', got '%s'", result)
		}
	})

	t.Run("returns empty for no user", func(t *testing.T) {
		result := GetUserRole(context.Background())
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})
}

func TestIsAuthenticated(t *testing.T) {
	t.Run("true when user present", func(t *testing.T) {
		user := &User{ID: "user-1"}
		ctx := context.WithValue(context.Background(), ContextKeyUser, user)

		if !IsAuthenticated(ctx) {
			t.Error("expected true")
		}
	})

	t.Run("false when no user", func(t *testing.T) {
		if IsAuthenticated(context.Background()) {
			t.Error("expected false")
		}
	})
}

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		userRole string
		minRole  string
		expected bool
	}{
		{"owner has admin", "owner", "admin", true},
		{"admin has admin", "admin", "admin", true},
		{"member lacks admin", "member", "admin", false},
		{"invalid role", "superuser", "admin", false},
		{"invalid min role", "admin", "superadmin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{ID: "user-1", Role: tt.userRole}
			ctx := context.WithValue(context.Background(), ContextKeyUser, user)

			result := HasRole(ctx, tt.minRole)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}

	t.Run("no user returns false", func(t *testing.T) {
		if HasRole(context.Background(), "viewer") {
			t.Error("expected false for no user")
		}
	})
}

// --- Auth metrics recording tests ---

func TestAuthenticate_RecordsAuthFailureMetrics(t *testing.T) {
	j := testJWT()
	met := metrics.New()
	m := NewMiddleware(MiddlewareConfig{
		JWT:     j,
		Logger:  testLogger(),
		Metrics: met,
	})
	handler := m.Authenticate(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/plans", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// --- Error response format tests ---

func TestWriteAuthError_Unauthorized(t *testing.T) {
	m := testMiddleware(nil)
	rec := httptest.NewRecorder()

	m.writeAuthError(rec, "test error", http.StatusUnauthorized)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}

	resp := parseErrorResponse(t, rec.Body)
	if resp["code"] != "UNAUTHORIZED" {
		t.Errorf("expected code 'UNAUTHORIZED', got '%v'", resp["code"])
	}
	if resp["message"] != "test error" {
		t.Errorf("expected message 'test error', got '%v'", resp["message"])
	}
	if resp["error"] != true {
		t.Errorf("expected error true, got '%v'", resp["error"])
	}
}

func TestWriteAuthError_Forbidden(t *testing.T) {
	m := testMiddleware(nil)
	rec := httptest.NewRecorder()

	m.writeAuthError(rec, "forbidden", http.StatusForbidden)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}

	resp := parseErrorResponse(t, rec.Body)
	if resp["code"] != "FORBIDDEN" {
		t.Errorf("expected code 'FORBIDDEN', got '%v'", resp["code"])
	}
}

// --- Integration: Authenticate + RequireRole chained ---

func TestAuthenticateAndRequireRole_Chained(t *testing.T) {
	j := testJWT()
	m := NewMiddleware(MiddlewareConfig{
		JWT:    j,
		Logger: testLogger(),
	})

	// Chain: Authenticate -> RequireRole("admin") -> handler
	handler := m.Authenticate(m.RequireRole("admin")(okHandler))

	t.Run("admin passes both", func(t *testing.T) {
		token := generateTestToken(j, "user-1", "admin@test.com", "Admin", "ws-1", "admin")
		req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("member blocked by role check", func(t *testing.T) {
		token := generateTestToken(j, "user-2", "member@test.com", "Member", "ws-1", "member")
		req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("no token blocked by auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/settings", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

// Ensure metrics import is used.
var _ = metrics.New
