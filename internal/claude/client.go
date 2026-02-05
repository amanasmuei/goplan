// Package claude provides Claude API client for AI-powered features.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/goplan/goplan/internal/config"
)

const (
	// BaseURL is the base URL for the Claude API.
	BaseURL = "https://api.anthropic.com/v1"

	// MessagesEndpoint is the endpoint for the messages API.
	MessagesEndpoint = "/messages"

	// AnthropicVersion is the API version header value.
	AnthropicVersion = "2023-06-01"

	// DefaultModel is the default Claude model to use.
	DefaultModel = "claude-sonnet-4-20250514"

	// DefaultMaxTokens is the default maximum tokens for responses.
	DefaultMaxTokens = 4096

	// DefaultTemperature is the default temperature for responses.
	DefaultTemperature = 0.7
)

// Client is a client for the Claude API.
type Client struct {
	httpClient  *http.Client
	apiKey      string
	model       string
	maxTokens   int
	temperature float64

	// Rate limiting
	rateLimiter *rateLimiter
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// NewClient creates a new Claude API client.
func NewClient(cfg config.AIConfig, opts ...ClientOption) *Client {
	model := cfg.ClaudeModel
	if model == "" {
		model = DefaultModel
	}

	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = DefaultMaxTokens
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		apiKey:      cfg.ClaudeAPIKey,
		model:       model,
		maxTokens:   maxTokens,
		temperature: DefaultTemperature,
		rateLimiter: newRateLimiter(60, time.Minute), // Default: 60 RPM
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithTemperature sets the temperature for the client.
func WithTemperature(temp float64) ClientOption {
	return func(c *Client) {
		c.temperature = temp
	}
}

// WithRateLimit sets the rate limit for the client.
func WithRateLimit(requests int, window time.Duration) ClientOption {
	return func(c *Client) {
		c.rateLimiter = newRateLimiter(requests, window)
	}
}

// Message represents a message in the conversation.
type Message struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// ContentPart represents a part of the message content.
type ContentPart struct {
	Type      string     `json:"type"`
	Text      string     `json:"text,omitempty"`
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Input     any        `json:"input,omitempty"`
	ToolUseID string     `json:"tool_use_id,omitempty"`
	Content   string     `json:"content,omitempty"`
	IsError   bool       `json:"is_error,omitempty"`
}

// NewTextMessage creates a new text message.
func NewTextMessage(role, text string) Message {
	return Message{
		Role: role,
		Content: []ContentPart{
			{Type: "text", Text: text},
		},
	}
}

// NewToolResultMessage creates a new tool result message.
func NewToolResultMessage(toolUseID string, content string, isError bool) Message {
	return Message{
		Role: "user",
		Content: []ContentPart{
			{
				Type:      "tool_result",
				ToolUseID: toolUseID,
				Content:   content,
				IsError:   isError,
			},
		},
	}
}

// Tool represents a tool that Claude can use.
type Tool struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	InputSchema JSONSchema `json:"input_schema"`
}

// JSONSchema represents a JSON schema for tool input.
type JSONSchema struct {
	Type       string                `json:"type"`
	Properties map[string]Property   `json:"properties,omitempty"`
	Required   []string              `json:"required,omitempty"`
}

// Property represents a property in the JSON schema.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Items       *Property `json:"items,omitempty"`
}

// MessagesRequest represents a request to the messages API.
type MessagesRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// MessagesResponse represents a response from the messages API.
type MessagesResponse struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Role         string        `json:"role"`
	Content      []ContentPart `json:"content"`
	Model        string        `json:"model"`
	StopReason   string        `json:"stop_reason"`
	StopSequence *string       `json:"stop_sequence,omitempty"`
	Usage        Usage         `json:"usage"`
}

// Usage represents token usage information.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// GetTextContent returns all text content from the response.
func (r *MessagesResponse) GetTextContent() string {
	var result string
	for _, part := range r.Content {
		if part.Type == "text" {
			result += part.Text
		}
	}
	return result
}

// GetToolUses returns all tool use content parts from the response.
func (r *MessagesResponse) GetToolUses() []ContentPart {
	var toolUses []ContentPart
	for _, part := range r.Content {
		if part.Type == "tool_use" {
			toolUses = append(toolUses, part)
		}
	}
	return toolUses
}

// HasToolUse returns true if the response contains tool use.
func (r *MessagesResponse) HasToolUse() bool {
	return r.StopReason == "tool_use"
}

// APIError represents an error from the Claude API.
type APIError struct {
	Type   string      `json:"type"`
	Detail ErrorDetail `json:"error"`
}

// ErrorDetail contains error details.
type ErrorDetail struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("claude API error (%s): %s", e.Detail.Type, e.Detail.Message)
}

// CreateMessage sends a message to Claude and returns the response.
func (c *Client) CreateMessage(ctx context.Context, req *MessagesRequest) (*MessagesResponse, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Set defaults if not provided
	if req.Model == "" {
		req.Model = c.model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.maxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = c.temperature
	}

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+MessagesEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("Anthropic-Version", AnthropicVersion)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return nil, &apiErr
	}

	// Parse response
	var messagesResp MessagesResponse
	if err := json.Unmarshal(respBody, &messagesResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &messagesResp, nil
}

// StreamHandler handles streaming events.
type StreamHandler func(event StreamEvent) error

// StreamEvent represents a streaming event.
type StreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta *Delta `json:"delta,omitempty"`
}

// Delta represents a streaming delta.
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// CreateMessageStream sends a message to Claude and streams the response.
func (c *Client) CreateMessageStream(ctx context.Context, req *MessagesRequest, handler StreamHandler) (*MessagesResponse, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Set defaults
	if req.Model == "" {
		req.Model = c.model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.maxTokens
	}
	req.Stream = true

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+MessagesEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("Anthropic-Version", AnthropicVersion)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return nil, &apiErr
	}

	// Process SSE stream
	return c.processSSEStream(resp.Body, handler)
}

// processSSEStream processes a Server-Sent Events stream.
func (c *Client) processSSEStream(reader io.Reader, handler StreamHandler) (*MessagesResponse, error) {
	var finalResponse MessagesResponse
	var textContent string
	buf := make([]byte, 4096)

	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read stream: %w", err)
		}

		data := string(buf[:n])
		lines := bytes.Split([]byte(data), []byte("\n"))

		for _, line := range lines {
			lineStr := string(line)
			if len(lineStr) == 0 || lineStr == "" {
				continue
			}

			// Parse SSE event
			if len(lineStr) > 6 && lineStr[:6] == "data: " {
				jsonData := lineStr[6:]
				if jsonData == "[DONE]" {
					break
				}

				var event StreamEvent
				if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
					continue
				}

				// Handle content block delta
				if event.Type == "content_block_delta" && event.Delta != nil {
					textContent += event.Delta.Text
					if handler != nil {
						if err := handler(event); err != nil {
							return nil, err
						}
					}
				}

				// Handle message stop
				if event.Type == "message_stop" {
					break
				}
			}
		}
	}

	// Build final response
	finalResponse.Content = []ContentPart{
		{Type: "text", Text: textContent},
	}

	return &finalResponse, nil
}

// rateLimiter implements a simple rate limiter.
type rateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	window    time.Duration
	lastReset time.Time
}

// newRateLimiter creates a new rate limiter.
func newRateLimiter(maxRequests int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:    maxRequests,
		maxTokens: maxRequests,
		window:    window,
		lastReset: time.Now(),
	}
}

// Wait waits for a rate limit token to become available.
func (r *rateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Reset tokens if window has passed
	now := time.Now()
	if now.Sub(r.lastReset) >= r.window {
		r.tokens = r.maxTokens
		r.lastReset = now
	}

	// Check if we have tokens available
	if r.tokens > 0 {
		r.tokens--
		return nil
	}

	// Calculate wait time
	waitTime := r.window - now.Sub(r.lastReset)
	if waitTime <= 0 {
		r.tokens = r.maxTokens - 1
		r.lastReset = now
		return nil
	}

	// Wait for rate limit reset
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		r.tokens = r.maxTokens - 1
		r.lastReset = time.Now()
		return nil
	}
}
