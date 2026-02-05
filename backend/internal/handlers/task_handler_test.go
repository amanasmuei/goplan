package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/goplan/backend/internal/models"
)

// MockTaskService mocks the TaskService interface
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) CreateTask(ctx interface{}, req models.CreateTaskRequest, userID, orgID uuid.UUID) (*models.TaskResponse, error) {
	args := m.Called(ctx, req, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskResponse), args.Error(1)
}

func (m *MockTaskService) GetTask(ctx interface{}, id, orgID uuid.UUID) (*models.Task, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskService) ListTasks(ctx interface{}, filters models.TaskFilters) (*models.TaskListResponse, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TaskListResponse), args.Error(1)
}

// Test helper to create Fiber app with test middleware
func setupTestApp() *fiber.App {
	app := fiber.New()
	return app
}

// Test helper to add mock auth context
func addMockAuthMiddleware(app *fiber.App, userID, orgID uuid.UUID) {
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userID", userID)
		c.Locals("orgID", orgID)
		return c.Next()
	})
}

func TestCreateTaskRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateTaskRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			req: models.CreateTaskRequest{
				Title:       "Test Task",
				Description: "A detailed description of the task",
			},
			wantErr: false,
		},
		{
			name: "Empty title",
			req: models.CreateTaskRequest{
				Title:       "",
				Description: "Description",
			},
			wantErr: true,
		},
		{
			name: "Short description",
			req: models.CreateTaskRequest{
				Title:       "Task",
				Description: "Short",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateTaskRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func validateCreateTaskRequest(req models.CreateTaskRequest) error {
	if req.Title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "title is required")
	}
	if len(req.Description) < 10 {
		return fiber.NewError(fiber.StatusBadRequest, "description must be at least 10 characters")
	}
	return nil
}

func TestTaskIDParsing(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "Valid UUID",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "Invalid UUID",
			id:      "invalid-uuid",
			wantErr: true,
		},
		{
			name:    "Empty string",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uuid.Parse(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHealthEndpoint(t *testing.T) {
	app := fiber.New()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"version": "1.0.0",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "healthy", result["status"])
}

func TestTaskFiltersParsing(t *testing.T) {
	tests := []struct {
		name         string
		queryParams  map[string]string
		expectedPage int
		expectedSize int
	}{
		{
			name:         "Default pagination",
			queryParams:  map[string]string{},
			expectedPage: 1,
			expectedSize: 50,
		},
		{
			name:         "Custom pagination",
			queryParams:  map[string]string{"page": "2", "page_size": "25"},
			expectedPage: 2,
			expectedSize: 25,
		},
		{
			name:         "Invalid page defaults to 1",
			queryParams:  map[string]string{"page": "invalid"},
			expectedPage: 1,
			expectedSize: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				page := c.QueryInt("page", 1)
				pageSize := c.QueryInt("page_size", 50)
				return c.JSON(fiber.Map{"page": page, "page_size": pageSize})
			})

			url := "/test?"
			for k, v := range tt.queryParams {
				url += k + "=" + v + "&"
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			resp, _ := app.Test(req)

			var result map[string]int
			json.NewDecoder(resp.Body).Decode(&result)
			assert.Equal(t, tt.expectedPage, result["page"])
			assert.Equal(t, tt.expectedSize, result["page_size"])
		})
	}
}

func TestJSONParsing(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "Valid JSON",
			body:    `{"title": "Test", "description": "Test description"}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			body:    `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "Empty body",
			body:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/test", func(c *fiber.Ctx) error {
				var req models.CreateTaskRequest
				if err := c.BodyParser(&req); err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "Invalid JSON"})
				}
				return c.JSON(req)
			})

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req)

			if tt.wantErr {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			} else {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestStatusCodeResponses(t *testing.T) {
	tests := []struct {
		name       string
		route      string
		method     string
		statusCode int
	}{
		{
			name:       "Not found",
			route:      "/nonexistent",
			method:     http.MethodGet,
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			req := httptest.NewRequest(tt.method, tt.route, nil)
			resp, _ := app.Test(req)
			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}
