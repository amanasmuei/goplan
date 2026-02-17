// Package handlers provides HTTP handlers for the REST API.
// This file contains comprehensive UAT (User Acceptance Testing) tests.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/goplan/goplan/internal/auth"
	"github.com/goplan/goplan/internal/domain/plan"
	"github.com/goplan/goplan/internal/domain/task"
	"github.com/goplan/goplan/internal/domain/user"
	"github.com/goplan/goplan/internal/domain/workspace"
	"github.com/goplan/goplan/internal/repository"
)

// ====================================================================
// Mock Repositories for UAT Testing
// ====================================================================

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	users    map[string]*user.UserWithPassword
	byEmail  map[string]*user.UserWithPassword
	idSeq    int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[string]*user.UserWithPassword),
		byEmail: make(map[string]*user.UserWithPassword),
	}
}

func (r *MockUserRepository) Create(ctx context.Context, u *user.UserWithPassword) error {
	r.idSeq++
	u.ID = fmt.Sprintf("user-%d", r.idSeq)
	r.users[u.ID] = u
	r.byEmail[u.Email] = u
	return nil
}

func (r *MockUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	if u, ok := r.users[id]; ok {
		return &u.User, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.UserWithPassword, error) {
	if u, ok := r.byEmail[email]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (r *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := r.byEmail[email]
	return ok, nil
}

func (r *MockUserRepository) Update(ctx context.Context, id string, input *user.UpdateUserInput) (*user.User, error) {
	if u, ok := r.users[id]; ok {
		if input.Name != nil {
			u.Name = *input.Name
		}
		u.UpdatedAt = time.Now()
		return &u.User, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (r *MockUserRepository) Delete(ctx context.Context, id string) error {
	delete(r.users, id)
	return nil
}

func (r *MockUserRepository) List(ctx context.Context, pagination repository.Pagination) (*repository.PaginatedResult[user.User], error) {
	var users []user.User
	for _, u := range r.users {
		users = append(users, u.User)
	}
	return &repository.PaginatedResult[user.User]{
		Items:      users,
		TotalCount: int64(len(users)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: 1,
	}, nil
}

func (r *MockUserRepository) GetByIDs(ctx context.Context, ids []string) ([]*user.User, error) {
	var result []*user.User
	for _, id := range ids {
		if u, ok := r.users[id]; ok {
			result = append(result, &u.User)
		}
	}
	return result, nil
}

// MockWorkspaceRepository is a mock implementation of WorkspaceRepository.
type MockWorkspaceRepository struct {
	workspaces map[string]*workspace.Workspace
	members    map[string][]*workspace.WorkspaceMember
	bySlug     map[string]*workspace.Workspace
	idSeq      int
}

func NewMockWorkspaceRepository() *MockWorkspaceRepository {
	return &MockWorkspaceRepository{
		workspaces: make(map[string]*workspace.Workspace),
		members:    make(map[string][]*workspace.WorkspaceMember),
		bySlug:     make(map[string]*workspace.Workspace),
	}
}

func (r *MockWorkspaceRepository) Create(ctx context.Context, ws *workspace.Workspace) error {
	r.workspaces[ws.ID] = ws
	r.bySlug[ws.Slug] = ws
	return nil
}

func (r *MockWorkspaceRepository) GetByID(ctx context.Context, id string) (*workspace.Workspace, error) {
	if ws, ok := r.workspaces[id]; ok {
		return ws, nil
	}
	return nil, fmt.Errorf("workspace not found")
}

func (r *MockWorkspaceRepository) GetBySlug(ctx context.Context, slug string) (*workspace.Workspace, error) {
	if ws, ok := r.bySlug[slug]; ok {
		return ws, nil
	}
	return nil, fmt.Errorf("workspace not found")
}

func (r *MockWorkspaceRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	_, ok := r.bySlug[slug]
	return ok, nil
}

func (r *MockWorkspaceRepository) GetUserWorkspaces(ctx context.Context, userID string) ([]*workspace.Workspace, error) {
	var result []*workspace.Workspace
	for wsID, members := range r.members {
		for _, m := range members {
			if m.UserID == userID {
				if ws, ok := r.workspaces[wsID]; ok {
					result = append(result, ws)
				}
			}
		}
	}
	return result, nil
}

func (r *MockWorkspaceRepository) Update(ctx context.Context, id string, input *workspace.UpdateWorkspaceInput) (*workspace.Workspace, error) {
	if ws, ok := r.workspaces[id]; ok {
		if input.Name != nil {
			ws.Name = *input.Name
		}
		ws.UpdatedAt = time.Now()
		return ws, nil
	}
	return nil, fmt.Errorf("workspace not found")
}

func (r *MockWorkspaceRepository) Delete(ctx context.Context, id string) error {
	if ws, ok := r.workspaces[id]; ok {
		delete(r.bySlug, ws.Slug)
	}
	delete(r.workspaces, id)
	return nil
}

func (r *MockWorkspaceRepository) AddMember(ctx context.Context, member *workspace.WorkspaceMember) error {
	r.members[member.WorkspaceID] = append(r.members[member.WorkspaceID], member)
	return nil
}

func (r *MockWorkspaceRepository) GetMember(ctx context.Context, workspaceID, userID string) (*workspace.WorkspaceMember, error) {
	for _, m := range r.members[workspaceID] {
		if m.UserID == userID {
			return m, nil
		}
	}
	return nil, nil
}

func (r *MockWorkspaceRepository) ListMembers(ctx context.Context, workspaceID string) ([]*workspace.WorkspaceMember, error) {
	return r.members[workspaceID], nil
}

func (r *MockWorkspaceRepository) UpdateMember(ctx context.Context, workspaceID, userID string, role string) error {
	for _, m := range r.members[workspaceID] {
		if m.UserID == userID {
			m.Role = role
			return nil
		}
	}
	return fmt.Errorf("member not found")
}

func (r *MockWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	members := r.members[workspaceID]
	for i, m := range members {
		if m.UserID == userID {
			r.members[workspaceID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return nil
}

// MockPlanRepository is a mock implementation of PlanRepository.
type MockPlanRepository struct {
	plans       map[string]*plan.Plan
	byWorkspace map[string][]*plan.Plan
	idSeq       int
}

func NewMockPlanRepository() *MockPlanRepository {
	return &MockPlanRepository{
		plans:       make(map[string]*plan.Plan),
		byWorkspace: make(map[string][]*plan.Plan),
	}
}

func (r *MockPlanRepository) Create(ctx context.Context, p *plan.Plan) error {
	r.plans[p.ID] = p
	r.byWorkspace[p.WorkspaceID] = append(r.byWorkspace[p.WorkspaceID], p)
	return nil
}

func (r *MockPlanRepository) GetByID(ctx context.Context, id string) (*plan.Plan, error) {
	if p, ok := r.plans[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("plan not found")
}

func (r *MockPlanRepository) GetByWorkspace(ctx context.Context, workspaceID string, pagination repository.Pagination) (*repository.PaginatedResult[*plan.Plan], error) {
	plans := r.byWorkspace[workspaceID]
	return &repository.PaginatedResult[*plan.Plan]{
		Items:      plans,
		TotalCount: int64(len(plans)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: 1,
	}, nil
}

func (r *MockPlanRepository) Update(ctx context.Context, id string, input *plan.UpdatePlanInput) (*plan.Plan, error) {
	if p, ok := r.plans[id]; ok {
		if input.Name != nil {
			p.Name = *input.Name
		}
		if input.Description != nil {
			p.Description = input.Description
		}
		if input.Status != nil {
			p.Status = *input.Status
		}
		p.UpdatedAt = time.Now()
		return p, nil
	}
	return nil, fmt.Errorf("plan not found")
}

func (r *MockPlanRepository) Delete(ctx context.Context, id string) error {
	if p, ok := r.plans[id]; ok {
		plans := r.byWorkspace[p.WorkspaceID]
		for i, plan := range plans {
			if plan.ID == id {
				r.byWorkspace[p.WorkspaceID] = append(plans[:i], plans[i+1:]...)
				break
			}
		}
	}
	delete(r.plans, id)
	return nil
}

func (r *MockPlanRepository) List(ctx context.Context, filter repository.PlanFilterOptions, pagination repository.Pagination) (*repository.PaginatedResult[*plan.Plan], error) {
	var plans []*plan.Plan
	for _, p := range r.plans {
		plans = append(plans, p)
	}
	return &repository.PaginatedResult[*plan.Plan]{
		Items:      plans,
		TotalCount: int64(len(plans)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: 1,
	}, nil
}

// MockTaskRepository is a mock implementation of TaskRepository.
type MockTaskRepository struct {
	tasks  map[string]*task.Task
	byPlan map[string][]*task.Task
	idSeq  int
}

func NewMockTaskRepository() *MockTaskRepository {
	return &MockTaskRepository{
		tasks:  make(map[string]*task.Task),
		byPlan: make(map[string][]*task.Task),
	}
}

func (r *MockTaskRepository) Create(ctx context.Context, t *task.Task) error {
	r.tasks[t.ID] = t
	r.byPlan[t.PlanID] = append(r.byPlan[t.PlanID], t)
	return nil
}

func (r *MockTaskRepository) GetByID(ctx context.Context, id string) (*task.Task, error) {
	if t, ok := r.tasks[id]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("task not found")
}

func (r *MockTaskRepository) GetByIDWithDetails(ctx context.Context, id string) (*task.Task, error) {
	return r.GetByID(ctx, id)
}

func (r *MockTaskRepository) ListByPlan(ctx context.Context, planID string, pagination repository.Pagination) (*repository.PaginatedResult[*task.Task], error) {
	tasks := r.byPlan[planID]
	return &repository.PaginatedResult[*task.Task]{
		Items:      tasks,
		TotalCount: int64(len(tasks)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: 1,
	}, nil
}

func (r *MockTaskRepository) List(ctx context.Context, filter repository.TaskFilterOptions, sort repository.TaskSortOptions, pagination repository.Pagination) (*repository.PaginatedResult[*task.Task], error) {
	var tasks []*task.Task
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	return &repository.PaginatedResult[*task.Task]{
		Items:      tasks,
		TotalCount: int64(len(tasks)),
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: 1,
	}, nil
}

func (r *MockTaskRepository) Update(ctx context.Context, id string, input *task.UpdateTaskInput) (*task.Task, error) {
	if t, ok := r.tasks[id]; ok {
		if input.Title != nil {
			t.Title = *input.Title
		}
		if input.Description != nil {
			t.Description = input.Description
		}
		if input.Status != nil {
			t.Status = *input.Status
		}
		if input.Priority != nil {
			t.Priority = *input.Priority
		}
		t.UpdatedAt = time.Now()
		return t, nil
	}
	return nil, fmt.Errorf("task not found")
}

func (r *MockTaskRepository) Delete(ctx context.Context, id string) error {
	if t, ok := r.tasks[id]; ok {
		tasks := r.byPlan[t.PlanID]
		for i, task := range tasks {
			if task.ID == id {
				r.byPlan[t.PlanID] = append(tasks[:i], tasks[i+1:]...)
				break
			}
		}
	}
	delete(r.tasks, id)
	return nil
}

func (r *MockTaskRepository) Move(ctx context.Context, id string, status string, position int) error {
	if t, ok := r.tasks[id]; ok {
		t.Status = status
		t.Position = position
		return nil
	}
	return fmt.Errorf("task not found")
}

func (r *MockTaskRepository) Count(ctx context.Context, filter repository.TaskFilterOptions) (int64, error) {
	return int64(len(r.tasks)), nil
}

// MockCommentRepository is a mock implementation of CommentRepository.
type MockCommentRepository struct {
	comments map[string]interface{}
}

func NewMockCommentRepository() *MockCommentRepository {
	return &MockCommentRepository{
		comments: make(map[string]interface{}),
	}
}

// ====================================================================
// Test Helpers
// ====================================================================

type uatTestContext struct {
	t           *testing.T
	authHandler *AuthHandler
	userRepo    *MockUserRepository
	jwt         *auth.JWT
	// Created resources for chained tests
	accessToken  string
	refreshToken string
	userID       string
	workspaceID  string
	planID       string
	taskID       string
}

func newUATTestContext(t *testing.T) *uatTestContext {
	userRepo := NewMockUserRepository()
	jwt := auth.NewJWT(auth.JWTConfig{
		Secret:          "test-secret-key-for-uat-testing-12345",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "goplan-test",
	})

	return &uatTestContext{
		t:           t,
		authHandler: NewAuthHandler(userRepo, jwt, nil),
		userRepo:    userRepo,
		jwt:         jwt,
	}
}

func (c *uatTestContext) makeRequest(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	return w
}

func parseResponse(w *httptest.ResponseRecorder, target interface{}) error {
	return json.NewDecoder(w.Body).Decode(target)
}

// wrappedAuthResponse represents the API response structure for auth endpoints
type wrappedAuthResponse struct {
	Success bool          `json:"success"`
	Data    *AuthResponse `json:"data"`
}

// wrappedErrorResponse represents the API error response structure
type wrappedErrorResponse struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ====================================================================
// UAT Test: Authentication Flow
// ====================================================================

func TestUAT_AuthenticationFlow(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	ctx := newUATTestContext(t)

	t.Run("1. Signup with valid data creates user and returns tokens", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "SecurePass123!",
			"name":     "Test User",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
			return
		}

		var wrapped wrappedAuthResponse
		if err := parseResponse(w, &wrapped); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if !wrapped.Success {
			t.Error("Expected success to be true")
		}

		resp := wrapped.Data
		if resp == nil {
			t.Fatal("Expected data in response")
		}

		if resp.AccessToken == "" {
			t.Error("Expected access token in response")
		}
		if resp.RefreshToken == "" {
			t.Error("Expected refresh token in response")
		}
		if resp.TokenType != "Bearer" {
			t.Errorf("Expected token type 'Bearer', got '%s'", resp.TokenType)
		}
		if resp.User == nil {
			t.Error("Expected user in response")
		} else if resp.User.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", resp.User.Email)
		}

		ctx.accessToken = resp.AccessToken
		ctx.refreshToken = resp.RefreshToken
		if resp.User != nil {
			ctx.userID = resp.User.ID
		}
	})

	t.Run("2. Signup with existing email returns 409 conflict", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "AnotherPass123!",
			"name":     "Another User",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("Expected status 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("3. Signup with weak password returns 400", func(t *testing.T) {
		body := map[string]string{
			"email":    "weak@example.com",
			"password": "weak",
			"name":     "Weak User",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("4. Signup with missing email returns 400", func(t *testing.T) {
		body := map[string]string{
			"password": "SecurePass123!",
			"name":     "No Email User",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("5. Login with correct credentials returns tokens", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "SecurePass123!",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Login(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
			return
		}

		var wrapped wrappedAuthResponse
		if err := parseResponse(w, &wrapped); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if wrapped.Data == nil || wrapped.Data.AccessToken == "" {
			t.Error("Expected access token in response")
		}
	})

	t.Run("6. Login with wrong password returns 401", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "WrongPassword123!",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Login(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("7. Login with non-existent email returns 401", func(t *testing.T) {
		body := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "AnyPassword123!",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Login(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("8. Refresh with valid token returns new tokens", func(t *testing.T) {
		if ctx.refreshToken == "" {
			t.Skip("No refresh token from signup")
		}

		body := map[string]string{
			"refreshToken": ctx.refreshToken,
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Refresh(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
			return
		}

		var wrapped wrappedAuthResponse
		if err := parseResponse(w, &wrapped); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if wrapped.Data == nil || wrapped.Data.AccessToken == "" {
			t.Error("Expected new access token")
		}
		if wrapped.Data == nil || wrapped.Data.RefreshToken == "" {
			t.Error("Expected new refresh token")
		}
	})

	t.Run("9. Refresh with invalid token returns 401", func(t *testing.T) {
		body := map[string]string{
			"refreshToken": "invalid-token",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Refresh(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("10. Logout returns success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Logout(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// ====================================================================
// UAT Test: JWT Token Validation
// ====================================================================

func TestUAT_JWTTokenValidation(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	jwt := auth.NewJWT(auth.JWTConfig{
		Secret:          "test-secret-key-for-uat-testing-12345",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		Issuer:          "goplan-test",
	})

	t.Run("Valid token contains correct claims", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		name := "Test User"

		accessToken, _, err := jwt.GenerateTokenPair(userID, email, name)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		claims, err := jwt.ValidateAccessToken(accessToken)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("Expected userID '%s', got '%s'", userID, claims.UserID)
		}
		if claims.Email != email {
			t.Errorf("Expected email '%s', got '%s'", email, claims.Email)
		}
	})

	t.Run("Expired token returns error", func(t *testing.T) {
		// Create JWT with very short expiry (1 second for reliable testing)
		shortJWT := auth.NewJWT(auth.JWTConfig{
			Secret:          "test-secret-key-for-uat-testing-12345",
			AccessTokenTTL:  -1 * time.Second, // Already expired
			RefreshTokenTTL: -1 * time.Second,
			Issuer:          "goplan-test",
		})

		accessToken, _, err := shortJWT.GenerateTokenPair("user-123", "test@example.com", "Test")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Token should already be expired (negative TTL)
		_, err = shortJWT.ValidateAccessToken(accessToken)
		if err == nil {
			t.Error("Expected error for expired token")
		}
	})

	t.Run("Tampered token returns error", func(t *testing.T) {
		accessToken, _, err := jwt.GenerateTokenPair("user-123", "test@example.com", "Test")
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Tamper with the token
		tamperedToken := accessToken + "tampered"

		_, err = jwt.ValidateAccessToken(tamperedToken)
		if err == nil {
			t.Error("Expected error for tampered token")
		}
	})
}

// ====================================================================
// UAT Test: Input Validation
// ====================================================================

func TestUAT_InputValidation(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	ctx := newUATTestContext(t)

	tests := []struct {
		name         string
		body         map[string]interface{}
		expectedCode int
	}{
		{
			name:         "Missing email",
			body:         map[string]interface{}{"password": "SecurePass123!", "name": "Test"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing password",
			body:         map[string]interface{}{"email": "test@example.com", "name": "Test"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Missing name",
			body:         map[string]interface{}{"email": "test@example.com", "password": "SecurePass123!"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty email",
			body:         map[string]interface{}{"email": "", "password": "SecurePass123!", "name": "Test"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Password too short",
			body:         map[string]interface{}{"email": "test@example.com", "password": "short", "name": "Test"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ctx.authHandler.Signup(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("Expected status %d, got %d: %s", tc.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}

// ====================================================================
// UAT Test: HTTP Method Validation
// ====================================================================

func TestUAT_HTTPMethodValidation(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	ctx := newUATTestContext(t)

	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
	}{
		{
			name:         "GET on signup endpoint",
			method:       http.MethodGet,
			path:         "/api/v1/auth/signup",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "PUT on signup endpoint",
			method:       http.MethodPut,
			path:         "/api/v1/auth/signup",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "DELETE on signup endpoint",
			method:       http.MethodDelete,
			path:         "/api/v1/auth/signup",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "GET on login endpoint",
			method:       http.MethodGet,
			path:         "/api/v1/auth/login",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			ctx.authHandler.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("Expected status %d, got %d: %s", tc.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}

// ====================================================================
// UAT Test: Response Format
// ====================================================================

func TestUAT_ResponseFormat(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	ctx := newUATTestContext(t)

	t.Run("Success response has correct structure", func(t *testing.T) {
		body := map[string]string{
			"email":    "format@example.com",
			"password": "SecurePass123!",
			"name":     "Format Test",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		// Check content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Parse and validate structure
		var resp map[string]any
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to parse response as JSON: %v", err)
		}

		// Check wrapper fields exist
		if _, ok := resp["success"]; !ok {
			t.Error("Missing 'success' field in response")
		}
		if _, ok := resp["data"]; !ok {
			t.Error("Missing 'data' field in response")
		}

		// Check data contains required fields
		data, ok := resp["data"].(map[string]any)
		if !ok {
			t.Fatal("'data' field is not an object")
		}
		requiredFields := []string{"user", "accessToken", "refreshToken", "tokenType", "expiresIn"}
		for _, field := range requiredFields {
			if _, ok := data[field]; !ok {
				t.Errorf("Missing required field '%s' in data", field)
			}
		}
	})

	t.Run("Error response has correct structure", func(t *testing.T) {
		body := map[string]string{
			"email": "incomplete@example.com",
			// Missing password and name
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		// Check content type
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Parse and validate structure
		var resp map[string]any
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("Failed to parse response as JSON: %v", err)
		}

		// Error responses should have 'error', 'code', and 'message'
		if _, ok := resp["error"]; !ok {
			t.Error("Missing 'error' field in error response")
		}
		if _, ok := resp["message"]; !ok {
			t.Error("Missing 'message' field in error response")
		}
	})
}

// ====================================================================
// UAT Test: Concurrent Requests
// ====================================================================

func TestUAT_ConcurrentRequests(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	ctx := newUATTestContext(t)

	t.Run("Handle multiple concurrent signups", func(t *testing.T) {
		const numRequests = 10
		done := make(chan bool, numRequests)
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(index int) {
				body := map[string]string{
					"email":    fmt.Sprintf("concurrent%d@example.com", index),
					"password": "SecurePass123!",
					"name":     fmt.Sprintf("Concurrent User %d", index),
				}

				jsonBody, _ := json.Marshal(body)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				ctx.authHandler.Signup(w, req)

				results <- w.Code
				done <- true
			}(i)
		}

		// Wait for all requests
		successCount := 0
		for i := 0; i < numRequests; i++ {
			<-done
			code := <-results
			if code == http.StatusCreated {
				successCount++
			}
		}

		if successCount != numRequests {
			t.Errorf("Expected %d successful signups, got %d", numRequests, successCount)
		}
	})
}

// ====================================================================
// UAT Test: Email Normalization
// ====================================================================

func TestUAT_EmailNormalization(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	ctx := newUATTestContext(t)

	t.Run("Email is normalized to lowercase", func(t *testing.T) {
		body := map[string]string{
			"email":    "TEST@EXAMPLE.COM",
			"password": "SecurePass123!",
			"name":     "Test User",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var wrapped wrappedAuthResponse
		if err := parseResponse(w, &wrapped); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if wrapped.Data == nil || wrapped.Data.User == nil {
			t.Fatal("Expected user in response")
		}

		if wrapped.Data.User.Email != "test@example.com" {
			t.Errorf("Expected normalized email 'test@example.com', got '%s'", wrapped.Data.User.Email)
		}
	})

	t.Run("Email with whitespace is trimmed", func(t *testing.T) {
		body := map[string]string{
			"email":    "  whitespace@example.com  ",
			"password": "SecurePass123!",
			"name":     "Whitespace User",
		}

		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx.authHandler.Signup(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		var wrapped wrappedAuthResponse
		if err := parseResponse(w, &wrapped); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		if wrapped.Data == nil || wrapped.Data.User == nil {
			t.Fatal("Expected user in response")
		}

		if wrapped.Data.User.Email != "whitespace@example.com" {
			t.Errorf("Expected trimmed email 'whitespace@example.com', got '%s'", wrapped.Data.User.Email)
		}
	})
}

// ====================================================================
// UAT Test: Invalid JSON Handling
// ====================================================================

func TestUAT_InvalidJSONHandling(t *testing.T) {
	if os.Getenv("UAT_TEST") == "" && testing.Short() {
		t.Skip("Skipping UAT test (set UAT_TEST=1 to run)")
	}

	ctx := newUATTestContext(t)

	tests := []struct {
		name         string
		body         string
		expectedCode int
	}{
		{
			name:         "Malformed JSON",
			body:         `{"email": "test@example.com", "password": }`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty body",
			body:         ``,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Array instead of object",
			body:         `["email", "password"]`,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ctx.authHandler.Signup(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("Expected status %d, got %d: %s", tc.expectedCode, w.Code, w.Body.String())
			}
		})
	}
}
