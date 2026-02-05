// Package claude provides Claude API integration with MCP tools.
package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/goplan/goplan/internal/mcp"
	"github.com/goplan/goplan/internal/mcp/tools"
)

// Bridge connects MCP tools to Claude API.
type Bridge struct {
	registry     *mcp.ToolRegistry
	safetyChecker *SafetyChecker
}

// NewBridge creates a new Bridge.
func NewBridge(registry *mcp.ToolRegistry, safetyChecker *SafetyChecker) *Bridge {
	return &Bridge{
		registry:      registry,
		safetyChecker: safetyChecker,
	}
}

// ConvertToolsToClaudeFormat converts MCP tool definitions to Claude tool format.
func (b *Bridge) ConvertToolsToClaudeFormat() []Tool {
	definitions := tools.GetToolDefinitions()
	claudeTools := make([]Tool, 0, len(definitions))

	for _, def := range definitions {
		tool := Tool{
			Name:        def.Name,
			Description: def.Description,
			InputSchema: convertArgumentsToSchema(def.Arguments),
		}
		claudeTools = append(claudeTools, tool)
	}

	return claudeTools
}

// convertArgumentsToSchema converts MCP argument definitions to JSON schema.
func convertArgumentsToSchema(args []mcp.ArgumentDef) JSONSchema {
	schema := JSONSchema{
		Type:       "object",
		Properties: make(map[string]Property),
		Required:   []string{},
	}

	for _, arg := range args {
		prop := Property{
			Type:        convertType(arg.Type),
			Description: arg.Description,
		}

		// Handle array types
		if arg.Type == "array" {
			prop.Items = &Property{Type: "string"}
		}

		schema.Properties[arg.Name] = prop

		if arg.Required {
			schema.Required = append(schema.Required, arg.Name)
		}
	}

	return schema
}

// convertType converts MCP types to JSON schema types.
func convertType(mcpType string) string {
	switch mcpType {
	case "string":
		return "string"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string"
	}
}

// ToolCallResult represents the result of a tool call.
type ToolCallResult struct {
	ToolUseID       string
	ToolName        string
	Result          interface{}
	Error           error
	RequiresApproval bool
	ApprovalRequest *ApprovalRequest
}

// ExecuteToolCall executes a tool call from Claude's response.
func (b *Bridge) ExecuteToolCall(ctx context.Context, execCtx mcp.ExecutionContext, toolUse ContentPart) *ToolCallResult {
	result := &ToolCallResult{
		ToolUseID: toolUse.ID,
		ToolName:  toolUse.Name,
	}

	// Parse tool input
	args, ok := toolUse.Input.(map[string]interface{})
	if !ok {
		// Try to convert from JSON
		inputBytes, err := json.Marshal(toolUse.Input)
		if err != nil {
			result.Error = fmt.Errorf("failed to marshal tool input: %w", err)
			return result
		}
		if err := json.Unmarshal(inputBytes, &args); err != nil {
			result.Error = fmt.Errorf("failed to parse tool input: %w", err)
			return result
		}
	}

	// Check if approval is required
	if b.safetyChecker != nil {
		approvalReq, requiresApproval := b.safetyChecker.CheckToolCall(toolUse.Name, args, execCtx)
		if requiresApproval {
			result.RequiresApproval = true
			result.ApprovalRequest = approvalReq
			return result
		}
	}

	// Execute the tool
	toolResult, err := b.registry.ExecuteTool(ctx, toolUse.Name, execCtx, args)
	if err != nil {
		result.Error = err
		log.Printf("Tool execution failed: %s - %v", toolUse.Name, err)
		return result
	}

	result.Result = toolResult
	log.Printf("Tool execution succeeded: %s", toolUse.Name)
	return result
}

// ExecuteToolCalls executes multiple tool calls from Claude's response.
func (b *Bridge) ExecuteToolCalls(ctx context.Context, execCtx mcp.ExecutionContext, toolUses []ContentPart) []*ToolCallResult {
	results := make([]*ToolCallResult, 0, len(toolUses))

	for _, toolUse := range toolUses {
		result := b.ExecuteToolCall(ctx, execCtx, toolUse)
		results = append(results, result)
	}

	return results
}

// FormatToolResults formats tool call results as messages for Claude.
func (b *Bridge) FormatToolResults(results []*ToolCallResult) []Message {
	messages := make([]Message, 0, len(results))

	for _, result := range results {
		var content string
		isError := false

		if result.RequiresApproval {
			content = fmt.Sprintf("This action requires user approval. Request ID: %s. Please inform the user that their approval is needed for: %s",
				result.ApprovalRequest.RequestID, result.ApprovalRequest.Description)
		} else if result.Error != nil {
			content = fmt.Sprintf("Error executing tool: %s", result.Error.Error())
			isError = true
		} else {
			// Marshal result to JSON
			resultBytes, err := json.Marshal(result.Result)
			if err != nil {
				content = fmt.Sprintf("Error formatting result: %s", err.Error())
				isError = true
			} else {
				content = string(resultBytes)
			}
		}

		messages = append(messages, NewToolResultMessage(result.ToolUseID, content, isError))
	}

	return messages
}

// Conversation manages a multi-turn conversation with Claude.
type Conversation struct {
	client   *Client
	bridge   *Bridge
	messages []Message
	tools    []Tool
	system   string
}

// NewConversation creates a new conversation.
func NewConversation(client *Client, bridge *Bridge, systemPrompt string) *Conversation {
	return &Conversation{
		client:   client,
		bridge:   bridge,
		messages: []Message{},
		tools:    bridge.ConvertToolsToClaudeFormat(),
		system:   systemPrompt,
	}
}

// AddUserMessage adds a user message to the conversation.
func (c *Conversation) AddUserMessage(content string) {
	c.messages = append(c.messages, NewTextMessage("user", content))
}

// AddAssistantMessage adds an assistant message to the conversation.
func (c *Conversation) AddAssistantMessage(response *MessagesResponse) {
	c.messages = append(c.messages, Message{
		Role:    "assistant",
		Content: response.Content,
	})
}

// AddToolResults adds tool results to the conversation.
func (c *Conversation) AddToolResults(results []*ToolCallResult) {
	toolResultMessages := c.bridge.FormatToolResults(results)
	c.messages = append(c.messages, toolResultMessages...)
}

// ConversationResponse represents the result of processing a conversation turn.
type ConversationResponse struct {
	TextContent      string
	ToolResults      []*ToolCallResult
	PendingApprovals []*ApprovalRequest
	Response         *MessagesResponse
}

// ProcessTurn sends the current messages to Claude and processes tool calls.
func (c *Conversation) ProcessTurn(ctx context.Context, execCtx mcp.ExecutionContext) (*ConversationResponse, error) {
	// Create request
	req := &MessagesRequest{
		Messages: c.messages,
		System:   c.system,
		Tools:    c.tools,
	}

	// Send to Claude
	response, err := c.client.CreateMessage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Add assistant message to history
	c.AddAssistantMessage(response)

	result := &ConversationResponse{
		TextContent: response.GetTextContent(),
		Response:    response,
	}

	// Process tool calls if any
	if response.HasToolUse() {
		toolUses := response.GetToolUses()
		toolResults := c.bridge.ExecuteToolCalls(ctx, execCtx, toolUses)
		result.ToolResults = toolResults

		// Separate pending approvals
		for _, tr := range toolResults {
			if tr.RequiresApproval {
				result.PendingApprovals = append(result.PendingApprovals, tr.ApprovalRequest)
			}
		}

		// If no pending approvals, add tool results and continue conversation
		if len(result.PendingApprovals) == 0 {
			c.AddToolResults(toolResults)
		}
	}

	return result, nil
}

// ProcessTurnWithContinuation processes a turn and continues until no more tool calls.
func (c *Conversation) ProcessTurnWithContinuation(ctx context.Context, execCtx mcp.ExecutionContext, maxIterations int) (*ConversationResponse, error) {
	var finalResult *ConversationResponse
	iterations := 0

	for iterations < maxIterations {
		iterations++

		result, err := c.ProcessTurn(ctx, execCtx)
		if err != nil {
			return nil, err
		}

		finalResult = result

		// Stop if there are pending approvals
		if len(result.PendingApprovals) > 0 {
			return result, nil
		}

		// Stop if no more tool calls
		if !result.Response.HasToolUse() {
			return result, nil
		}
	}

	return finalResult, nil
}

// ContinueAfterApproval continues the conversation after approvals are handled.
func (c *Conversation) ContinueAfterApproval(ctx context.Context, execCtx mcp.ExecutionContext, approvedResults []*ToolCallResult) (*ConversationResponse, error) {
	// Add the approved tool results
	c.AddToolResults(approvedResults)

	// Continue processing
	return c.ProcessTurnWithContinuation(ctx, execCtx, 10)
}

// GetMessages returns the conversation messages.
func (c *Conversation) GetMessages() []Message {
	return c.messages
}

// Reset clears the conversation history.
func (c *Conversation) Reset() {
	c.messages = []Message{}
}
