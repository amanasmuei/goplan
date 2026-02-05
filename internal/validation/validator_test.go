package validation

import (
	"testing"
)

func TestValidator_ValidateString(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name     string
		field    string
		value    string
		minLen   int
		maxLen   int
		required bool
		wantErr  bool
	}{
		{
			name:     "Valid string",
			field:    "title",
			value:    "Hello World",
			minLen:   1,
			maxLen:   100,
			required: true,
			wantErr:  false,
		},
		{
			name:     "Empty required string",
			field:    "title",
			value:    "",
			minLen:   1,
			maxLen:   100,
			required: true,
			wantErr:  true,
		},
		{
			name:     "Empty optional string",
			field:    "title",
			value:    "",
			minLen:   1,
			maxLen:   100,
			required: false,
			wantErr:  false,
		},
		{
			name:     "String too short",
			field:    "title",
			value:    "Hi",
			minLen:   5,
			maxLen:   100,
			required: true,
			wantErr:  true,
		},
		{
			name:     "String too long",
			field:    "title",
			value:    "This is a very long string that exceeds the maximum length",
			minLen:   1,
			maxLen:   10,
			required: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateString(tt.field, tt.value, tt.minLen, tt.maxLen, tt.required)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name     string
		email    string
		required bool
		wantErr  bool
	}{
		{"Valid email", "test@example.com", true, false},
		{"Valid email with subdomain", "test@sub.example.com", true, false},
		{"Valid email with plus", "test+tag@example.com", true, false},
		{"Empty required email", "", true, true},
		{"Empty optional email", "", false, false},
		{"Invalid email no @", "testexample.com", true, true},
		{"Invalid email no domain", "test@", true, true},
		{"Invalid email no local", "@example.com", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEmail("email", tt.email, tt.required)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateSlug(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		{"Valid slug", "my-project", false},
		{"Valid slug with numbers", "project-123", false},
		{"Valid single word", "project", false},
		{"Invalid with uppercase", "My-Project", true},
		{"Invalid with spaces", "my project", true},
		{"Invalid with underscore", "my_project", true},
		{"Invalid starting with hyphen", "-project", true},
		{"Invalid ending with hyphen", "project-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateSlug("slug", tt.slug, 1, 50, true)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateUUID(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"Valid UUID", "123e4567-e89b-12d3-a456-426614174000", false},
		{"Valid UUID uppercase", "123E4567-E89B-12D3-A456-426614174000", false},
		{"Invalid UUID", "not-a-uuid", true},
		{"Invalid UUID short", "123e4567-e89b-12d3-a456", true},
		{"Invalid UUID no hyphens", "123e4567e89b12d3a456426614174000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateUUID("id", tt.uuid, true)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_CheckSQLInjection(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name       string
		input      string
		wantUnsafe bool
	}{
		{"Normal text", "Hello World", false},
		{"SQL injection UNION", "'; UNION SELECT * FROM users --", true},
		{"SQL injection DROP", "'; DROP TABLE users; --", true},
		{"SQL injection INSERT", "INSERT INTO users VALUES ('test')", true},
		{"Normal text with quotes", "It's a nice day", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.CheckSQLInjection(tt.input)
			if result != tt.wantUnsafe {
				t.Errorf("Expected %v, got %v", tt.wantUnsafe, result)
			}
		})
	}
}

func TestValidator_CheckXSS(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name       string
		input      string
		wantUnsafe bool
	}{
		{"Normal text", "Hello World", false},
		{"Script tag", "<script>alert('xss')</script>", true},
		{"Event handler", "<img onerror='alert(1)'>", true},
		{"Javascript URL", "javascript:alert('xss')", true},
		{"Normal HTML", "<p>Hello</p>", false},
		{"Iframe tag", "<iframe src='evil.com'>", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.CheckXSS(tt.input)
			if result != tt.wantUnsafe {
				t.Errorf("Expected %v, got %v", tt.wantUnsafe, result)
			}
		})
	}
}

func TestValidator_SanitizeString(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Normal text", "Hello World", "Hello World"},
		{"HTML tags", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"Null bytes", "Hello\x00World", "HelloWorld"},
		{"Whitespace", "  Hello World  ", "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.SanitizeString(tt.input)
			if result != tt.want {
				t.Errorf("Expected %q, got %q", tt.want, result)
			}
		})
	}
}

func TestValidator_ValidatePassword(t *testing.T) {
	v := New(DefaultConfig())

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", "Password123", false},
		{"Valid complex password", "MyP@ssword123!", false},
		{"Too short", "Pass1", true},
		{"No uppercase", "password123", true},
		{"No lowercase", "PASSWORD123", true},
		{"No number", "PasswordABC", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePassword(tt.password)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidator_ValidateEnum(t *testing.T) {
	v := New(DefaultConfig())
	allowed := []string{"low", "medium", "high", "critical"}

	tests := []struct {
		name     string
		value    string
		required bool
		wantErr  bool
	}{
		{"Valid value", "low", true, false},
		{"Valid value medium", "medium", true, false},
		{"Invalid value", "urgent", true, true},
		{"Empty required", "", true, true},
		{"Empty optional", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEnum("priority", tt.value, allowed, tt.required)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
