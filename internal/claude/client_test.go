package claude

import (
	"context"
	"testing"
	"time"

	"github.com/goplan/goplan/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := config.AIConfig{
		ClaudeAPIKey: "test-api-key",
		ClaudeModel:  "claude-sonnet-4-20250514",
		MaxTokens:    4096,
		Timeout:      30 * time.Second,
	}

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.apiKey != "test-api-key" {
		t.Errorf("client.apiKey = %q, want %q", client.apiKey, "test-api-key")
	}

	if client.model != "claude-sonnet-4-20250514" {
		t.Errorf("client.model = %q, want %q", client.model, "claude-sonnet-4-20250514")
	}

	if client.maxTokens != 4096 {
		t.Errorf("client.maxTokens = %d, want %d", client.maxTokens, 4096)
	}
}

func TestNewClient_WithDefaults(t *testing.T) {
	cfg := config.AIConfig{
		ClaudeAPIKey: "test-api-key",
	}

	client := NewClient(cfg)

	if client.model != DefaultModel {
		t.Errorf("client.model = %q, want %q", client.model, DefaultModel)
	}

	if client.maxTokens != DefaultMaxTokens {
		t.Errorf("client.maxTokens = %d, want %d", client.maxTokens, DefaultMaxTokens)
	}

	if client.temperature != DefaultTemperature {
		t.Errorf("client.temperature = %f, want %f", client.temperature, DefaultTemperature)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	cfg := config.AIConfig{
		ClaudeAPIKey: "test-api-key",
	}

	client := NewClient(cfg,
		WithTemperature(0.5),
		WithRateLimit(10, time.Second),
	)

	if client.temperature != 0.5 {
		t.Errorf("client.temperature = %f, want %f", client.temperature, 0.5)
	}

	if client.rateLimiter.maxTokens != 10 {
		t.Errorf("rateLimiter.maxTokens = %d, want %d", client.rateLimiter.maxTokens, 10)
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := newRateLimiter(3, 100*time.Millisecond)
	ctx := context.Background()

	// First 3 requests should succeed immediately
	for i := 0; i < 3; i++ {
		start := time.Now()
		err := limiter.Wait(ctx)
		if err != nil {
			t.Errorf("Wait() error = %v", err)
		}
		if time.Since(start) > 10*time.Millisecond {
			t.Errorf("Wait() took too long for request %d", i)
		}
	}

	// Fourth request should wait
	start := time.Now()
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Wait() error = %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 50*time.Millisecond {
		t.Errorf("Wait() should have waited, but only took %v", elapsed)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	limiter := newRateLimiter(1, time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Use up the one token
	_ = limiter.Wait(ctx)

	// Second request should fail due to context timeout
	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("Wait() should have returned error due to context cancellation")
	}
}

func TestNewTextMessage(t *testing.T) {
	msg := NewTextMessage("user", "Hello, Claude!")

	if msg.Role != "user" {
		t.Errorf("Message role = %q, want %q", msg.Role, "user")
	}

	if len(msg.Content) != 1 {
		t.Errorf("Message has %d content parts, want 1", len(msg.Content))
	}

	if msg.Content[0].Type != "text" {
		t.Errorf("Content type = %q, want %q", msg.Content[0].Type, "text")
	}

	if msg.Content[0].Text != "Hello, Claude!" {
		t.Errorf("Content text = %q, want %q", msg.Content[0].Text, "Hello, Claude!")
	}
}

func TestNewToolResultMessage(t *testing.T) {
	msg := NewToolResultMessage("tool-123", `{"result": "success"}`, false)

	if msg.Role != "user" {
		t.Errorf("Message role = %q, want %q", msg.Role, "user")
	}

	if len(msg.Content) != 1 {
		t.Errorf("Message has %d content parts, want 1", len(msg.Content))
	}

	content := msg.Content[0]
	if content.Type != "tool_result" {
		t.Errorf("Content type = %q, want %q", content.Type, "tool_result")
	}

	if content.ToolUseID != "tool-123" {
		t.Errorf("Content ToolUseID = %q, want %q", content.ToolUseID, "tool-123")
	}

	if content.IsError {
		t.Error("Content IsError should be false")
	}
}

func TestNewToolResultMessage_WithError(t *testing.T) {
	msg := NewToolResultMessage("tool-123", "Error: not found", true)

	content := msg.Content[0]
	if !content.IsError {
		t.Error("Content IsError should be true")
	}
}

func TestMessagesResponse_GetTextContent(t *testing.T) {
	resp := &MessagesResponse{
		Content: []ContentPart{
			{Type: "text", Text: "Hello "},
			{Type: "text", Text: "World!"},
			{Type: "tool_use", ID: "tool-1", Name: "test"},
		},
	}

	text := resp.GetTextContent()
	if text != "Hello World!" {
		t.Errorf("GetTextContent() = %q, want %q", text, "Hello World!")
	}
}

func TestMessagesResponse_GetToolUses(t *testing.T) {
	resp := &MessagesResponse{
		Content: []ContentPart{
			{Type: "text", Text: "Let me help with that."},
			{Type: "tool_use", ID: "tool-1", Name: "task.create"},
			{Type: "tool_use", ID: "tool-2", Name: "task.list"},
		},
	}

	toolUses := resp.GetToolUses()
	if len(toolUses) != 2 {
		t.Errorf("GetToolUses() returned %d tools, want 2", len(toolUses))
	}

	if toolUses[0].ID != "tool-1" {
		t.Errorf("First tool ID = %q, want %q", toolUses[0].ID, "tool-1")
	}
}

func TestMessagesResponse_HasToolUse(t *testing.T) {
	tests := []struct {
		name       string
		stopReason string
		want       bool
	}{
		{"with tool_use", "tool_use", true},
		{"with end_turn", "end_turn", false},
		{"with max_tokens", "max_tokens", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &MessagesResponse{StopReason: tt.stopReason}
			if got := resp.HasToolUse(); got != tt.want {
				t.Errorf("HasToolUse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Type: "error",
		Detail: ErrorDetail{
			Type:    "invalid_request_error",
			Message: "Invalid API key",
		},
	}

	errorMsg := err.Error()
	if errorMsg != "claude API error (invalid_request_error): Invalid API key" {
		t.Errorf("Error() = %q, unexpected format", errorMsg)
	}
}
