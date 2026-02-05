package claude

import (
	"fmt"
	"testing"

	"github.com/goplan/goplan/internal/mcp"
)

func TestBridge_ConvertToolsToClaudeFormat(t *testing.T) {
	registry := mcp.NewToolRegistry()
	checker := NewSafetyChecker()
	bridge := NewBridge(registry, checker)

	tools := bridge.ConvertToolsToClaudeFormat()

	if len(tools) == 0 {
		t.Error("ConvertToolsToClaudeFormat() returned no tools")
	}

	// Check that tools have required fields
	for _, tool := range tools {
		if tool.Name == "" {
			t.Error("Tool has empty name")
		}
		if tool.Description == "" {
			t.Errorf("Tool %s has empty description", tool.Name)
		}
		if tool.InputSchema.Type != "object" {
			t.Errorf("Tool %s has invalid schema type: %s", tool.Name, tool.InputSchema.Type)
		}
	}

	// Check for specific tools we expect
	expectedTools := []string{
		"workspace.list",
		"plan.create",
		"task.create",
		"task.list",
		"milestone.create",
		"activity.get",
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("Expected tool %s not found", expected)
		}
	}
}

func TestConvertArgumentsToSchema(t *testing.T) {
	args := []mcp.ArgumentDef{
		{Name: "planId", Type: "string", Required: true, Description: "Plan ID"},
		{Name: "title", Type: "string", Required: true, Description: "Task title"},
		{Name: "priority", Type: "string", Required: false, Description: "Task priority"},
		{Name: "tags", Type: "array", Required: false, Description: "Task tags"},
		{Name: "estimatedHours", Type: "number", Required: false, Description: "Hours estimate"},
	}

	schema := convertArgumentsToSchema(args)

	// Check type is object
	if schema.Type != "object" {
		t.Errorf("Schema type = %s, want object", schema.Type)
	}

	// Check properties
	if len(schema.Properties) != 5 {
		t.Errorf("Schema has %d properties, want 5", len(schema.Properties))
	}

	// Check required fields
	if len(schema.Required) != 2 {
		t.Errorf("Schema has %d required fields, want 2", len(schema.Required))
	}

	// Check specific properties
	if prop, ok := schema.Properties["planId"]; !ok {
		t.Error("planId property not found")
	} else if prop.Type != "string" {
		t.Errorf("planId type = %s, want string", prop.Type)
	}

	if prop, ok := schema.Properties["tags"]; !ok {
		t.Error("tags property not found")
	} else if prop.Type != "array" {
		t.Errorf("tags type = %s, want array", prop.Type)
	} else if prop.Items == nil {
		t.Error("tags items should not be nil")
	}

	if prop, ok := schema.Properties["estimatedHours"]; !ok {
		t.Error("estimatedHours property not found")
	} else if prop.Type != "number" {
		t.Errorf("estimatedHours type = %s, want number", prop.Type)
	}
}

func TestConvertType(t *testing.T) {
	tests := []struct {
		mcpType string
		want    string
	}{
		{"string", "string"},
		{"number", "number"},
		{"boolean", "boolean"},
		{"array", "array"},
		{"object", "object"},
		{"unknown", "string"}, // Default to string
	}

	for _, tt := range tests {
		t.Run(tt.mcpType, func(t *testing.T) {
			got := convertType(tt.mcpType)
			if got != tt.want {
				t.Errorf("convertType(%q) = %q, want %q", tt.mcpType, got, tt.want)
			}
		})
	}
}

func TestToolCallResult_Formatting(t *testing.T) {
	registry := mcp.NewToolRegistry()
	checker := NewSafetyChecker()
	bridge := NewBridge(registry, checker)

	// Test successful result
	successResult := &ToolCallResult{
		ToolUseID: "tool-1",
		ToolName:  "task.list",
		Result:    map[string]interface{}{"tasks": []string{"task-1", "task-2"}},
	}

	messages := bridge.FormatToolResults([]*ToolCallResult{successResult})
	if len(messages) != 1 {
		t.Errorf("FormatToolResults() returned %d messages, want 1", len(messages))
	}

	// Test error result
	errorResult := &ToolCallResult{
		ToolUseID: "tool-2",
		ToolName:  "task.get",
		Error:     fmt.Errorf("test error"),
	}

	messages = bridge.FormatToolResults([]*ToolCallResult{errorResult})
	if len(messages) != 1 {
		t.Errorf("FormatToolResults() returned %d messages, want 1", len(messages))
	}
}

func TestConversation_MessageManagement(t *testing.T) {
	// Create a conversation without client for basic testing
	conv := &Conversation{
		messages: []Message{},
		system:   "Test system prompt",
	}

	// Test adding user message
	conv.AddUserMessage("Hello")

	if len(conv.messages) != 1 {
		t.Errorf("AddUserMessage() resulted in %d messages, want 1", len(conv.messages))
	}

	if conv.messages[0].Role != "user" {
		t.Errorf("Message role = %s, want user", conv.messages[0].Role)
	}

	// Test reset
	conv.Reset()

	if len(conv.messages) != 0 {
		t.Errorf("Reset() resulted in %d messages, want 0", len(conv.messages))
	}
}
