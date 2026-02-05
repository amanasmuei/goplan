// Package tools provides MCP tool implementations for GoPlan.
package tools

import (
	"fmt"

	"github.com/goplan/goplan/internal/domain/shared"
)

// getRequiredString extracts a required string argument.
func getRequiredString(args map[string]interface{}, key string) (string, error) {
	v, ok := args[key]
	if !ok {
		return "", shared.NewValidationError(key, fmt.Sprintf("%s is required", key))
	}
	s, ok := v.(string)
	if !ok {
		return "", shared.NewValidationError(key, fmt.Sprintf("%s must be a string", key))
	}
	if s == "" {
		return "", shared.NewValidationError(key, fmt.Sprintf("%s cannot be empty", key))
	}
	return s, nil
}

// getOptionalString extracts an optional string argument.
func getOptionalString(args map[string]interface{}, key string) *string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	return &s
}

// getOptionalInt extracts an optional int argument.
func getOptionalInt(args map[string]interface{}, key string) *int {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	switch n := v.(type) {
	case float64:
		i := int(n)
		return &i
	case int:
		return &n
	case int64:
		i := int(n)
		return &i
	}
	return nil
}

// getOptionalFloat extracts an optional float argument.
func getOptionalFloat(args map[string]interface{}, key string) *float64 {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	switch n := v.(type) {
	case float64:
		return &n
	case int:
		f := float64(n)
		return &f
	case int64:
		f := float64(n)
		return &f
	}
	return nil
}

// getInt extracts an int argument with a default value.
func getInt(args map[string]interface{}, key string, defaultVal int) int {
	v, ok := args[key]
	if !ok || v == nil {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return defaultVal
}

// getStringSlice extracts a string slice argument.
func getStringSlice(args map[string]interface{}, key string) []string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	slice, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// getMap extracts a map argument.
func getMap(args map[string]interface{}, key string) map[string]interface{} {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}
	return m
}
