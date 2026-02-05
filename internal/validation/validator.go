// Package validation provides input validation and sanitization for the GoPlan backend.
package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/goplan/goplan/internal/domain/shared"
)

// Common validation errors.
var (
	ErrRequestTooLarge    = errors.New("request body too large")
	ErrInvalidJSON        = errors.New("invalid JSON")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrSuspiciousInput    = errors.New("suspicious input detected")
)

// Config holds validator configuration.
type Config struct {
	MaxBodySize       int64  // Maximum request body size in bytes
	MaxStringLength   int    // Maximum string field length
	MaxArrayLength    int    // Maximum array length
	StrictMode        bool   // Enable strict validation
	AllowedHTMLTags   []string
	SanitizeHTML      bool
}

// DefaultConfig returns the default validator configuration.
func DefaultConfig() Config {
	return Config{
		MaxBodySize:     1 * 1024 * 1024, // 1MB
		MaxStringLength: 10000,
		MaxArrayLength:  1000,
		StrictMode:      true,
		SanitizeHTML:    true,
		AllowedHTMLTags: []string{},
	}
}

// Validator provides input validation and sanitization.
type Validator struct {
	config Config

	// Common regex patterns
	emailRegex     *regexp.Regexp
	slugRegex      *regexp.Regexp
	uuidRegex      *regexp.Regexp
	urlRegex       *regexp.Regexp
	sqlPatternRegex *regexp.Regexp
	xssPatternRegex *regexp.Regexp
}

// New creates a new validator.
func New(config Config) *Validator {
	return &Validator{
		config:          config,
		emailRegex:      regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
		slugRegex:       regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`),
		uuidRegex:       regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
		urlRegex:        regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`),
		sqlPatternRegex: regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|update\s+.*\s+set|delete\s+from|drop\s+table|';\s*--|;\s*drop|xp_cmdshell|exec\s+master)`),
		xssPatternRegex: regexp.MustCompile(`(?i)(<script|javascript:|on\w+\s*=|<iframe|<object|<embed|expression\s*\(|url\s*\(|@import)`),
	}
}

// ValidationResult holds the result of a validation.
type ValidationResult struct {
	errors []*shared.DomainError
}

// NewValidationResult creates a new validation result.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		errors: make([]*shared.DomainError, 0),
	}
}

// AddError adds a validation error.
func (r *ValidationResult) AddError(field, message string) {
	r.errors = append(r.errors, shared.NewValidationError(field, message))
}

// HasErrors returns true if there are validation errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.errors) > 0
}

// Error returns the validation errors.
func (r *ValidationResult) Error() error {
	if !r.HasErrors() {
		return nil
	}
	return &shared.ValidationErrors{Errors: r.errors}
}

// Errors returns the list of errors.
func (r *ValidationResult) Errors() []*shared.DomainError {
	return r.errors
}

// DecodeAndValidateJSON decodes and validates JSON from a request.
func (v *Validator) DecodeAndValidateJSON(r *http.Request, target interface{}) error {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		return ErrInvalidContentType
	}

	// Limit body size
	if r.Body == nil {
		return shared.NewValidationError("body", "request body is required")
	}
	r.Body = http.MaxBytesReader(nil, r.Body, v.config.MaxBodySize)

	// Decode JSON
	decoder := json.NewDecoder(r.Body)
	if v.config.StrictMode {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return shared.NewValidationError("body", "request body is required")
		}
		if _, ok := err.(*http.MaxBytesError); ok {
			return ErrRequestTooLarge
		}
		return shared.NewValidationError("body", "invalid JSON: "+err.Error())
	}

	return nil
}

// ValidateString validates a string field.
func (v *Validator) ValidateString(field, value string, minLen, maxLen int, required bool) error {
	if value == "" {
		if required {
			return shared.NewValidationError(field, fmt.Sprintf("%s is required", field))
		}
		return nil
	}

	// Check length
	length := utf8.RuneCountInString(value)
	if length < minLen {
		return shared.NewValidationError(field, fmt.Sprintf("%s must be at least %d characters", field, minLen))
	}
	if maxLen > 0 && length > maxLen {
		return shared.NewValidationError(field, fmt.Sprintf("%s must be at most %d characters", field, maxLen))
	}

	// Check for null bytes
	if strings.ContainsRune(value, 0) {
		return shared.NewValidationError(field, "invalid characters in "+field)
	}

	return nil
}

// ValidateEmail validates an email address.
func (v *Validator) ValidateEmail(field, email string, required bool) error {
	if email == "" {
		if required {
			return shared.NewValidationError(field, "email is required")
		}
		return nil
	}

	if !v.emailRegex.MatchString(email) {
		return shared.NewValidationError(field, "invalid email format")
	}

	return nil
}

// ValidateSlug validates a URL slug.
func (v *Validator) ValidateSlug(field, slug string, minLen, maxLen int, required bool) error {
	if err := v.ValidateString(field, slug, minLen, maxLen, required); err != nil {
		return err
	}

	if slug != "" && !v.slugRegex.MatchString(slug) {
		return shared.NewValidationError(field, "slug must contain only lowercase letters, numbers, and hyphens")
	}

	return nil
}

// ValidateUUID validates a UUID.
func (v *Validator) ValidateUUID(field, uuid string, required bool) error {
	if uuid == "" {
		if required {
			return shared.NewValidationError(field, field+" is required")
		}
		return nil
	}

	if !v.uuidRegex.MatchString(uuid) {
		return shared.NewValidationError(field, "invalid "+field+" format")
	}

	return nil
}

// ValidateURL validates a URL.
func (v *Validator) ValidateURL(field, url string, required bool) error {
	if url == "" {
		if required {
			return shared.NewValidationError(field, field+" is required")
		}
		return nil
	}

	if !v.urlRegex.MatchString(url) {
		return shared.NewValidationError(field, "invalid URL format")
	}

	return nil
}

// SanitizeString sanitizes a string by escaping HTML and removing dangerous content.
func (v *Validator) SanitizeString(s string) string {
	if s == "" {
		return s
	}

	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")

	// Escape HTML if configured
	if v.config.SanitizeHTML {
		s = html.EscapeString(s)
	}

	// Trim whitespace
	s = strings.TrimSpace(s)

	return s
}

// SanitizeHTML sanitizes HTML content, allowing only safe tags.
func (v *Validator) SanitizeHTML(s string) string {
	// For now, just escape all HTML
	// In production, you might want to use a proper HTML sanitizer like bluemonday
	return html.EscapeString(s)
}

// CheckSQLInjection checks for potential SQL injection patterns.
func (v *Validator) CheckSQLInjection(value string) bool {
	return v.sqlPatternRegex.MatchString(value)
}

// CheckXSS checks for potential XSS patterns.
func (v *Validator) CheckXSS(value string) bool {
	return v.xssPatternRegex.MatchString(value)
}

// ValidateSafeInput validates that input is free from injection attacks.
func (v *Validator) ValidateSafeInput(field, value string) error {
	if value == "" {
		return nil
	}

	if v.CheckSQLInjection(value) {
		return shared.NewValidationError(field, "potentially unsafe input detected")
	}

	if v.CheckXSS(value) {
		return shared.NewValidationError(field, "potentially unsafe input detected")
	}

	return nil
}

// ValidateAndSanitize validates and sanitizes a string.
func (v *Validator) ValidateAndSanitize(field, value string, minLen, maxLen int, required bool) (string, error) {
	// First validate
	if err := v.ValidateString(field, value, minLen, maxLen, required); err != nil {
		return "", err
	}

	// Check for dangerous content
	if err := v.ValidateSafeInput(field, value); err != nil {
		return "", err
	}

	// Then sanitize
	return v.SanitizeString(value), nil
}

// ValidatePassword validates a password.
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return shared.NewValidationError("password", "password is required")
	}

	if len(password) < 8 {
		return shared.NewValidationError("password", "password must be at least 8 characters")
	}

	if len(password) > 128 {
		return shared.NewValidationError("password", "password must be at most 128 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return shared.NewValidationError("password", "password must contain uppercase, lowercase, and numeric characters")
	}

	// Special character is optional but recommended
	_ = hasSpecial

	return nil
}

// ValidateArray validates an array length.
func (v *Validator) ValidateArray(field string, length, maxLength int) error {
	if maxLength <= 0 {
		maxLength = v.config.MaxArrayLength
	}

	if length > maxLength {
		return shared.NewValidationError(field, fmt.Sprintf("%s cannot contain more than %d items", field, maxLength))
	}

	return nil
}

// ValidateEnum validates that a value is one of the allowed values.
func (v *Validator) ValidateEnum(field, value string, allowed []string, required bool) error {
	if value == "" {
		if required {
			return shared.NewValidationError(field, field+" is required")
		}
		return nil
	}

	for _, a := range allowed {
		if value == a {
			return nil
		}
	}

	return shared.NewValidationError(field, fmt.Sprintf("invalid %s; must be one of: %s", field, strings.Join(allowed, ", ")))
}

// RequestSizeLimitMiddleware returns middleware that limits request body size.
func (v *Validator) RequestSizeLimitMiddleware(maxSize int64) func(http.Handler) http.Handler {
	if maxSize <= 0 {
		maxSize = v.config.MaxBodySize
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}

// Global validator instance.
var defaultValidator *Validator

// SetDefault sets the default validator.
func SetDefault(v *Validator) {
	defaultValidator = v
}

// Default returns the default validator.
func Default() *Validator {
	if defaultValidator == nil {
		defaultValidator = New(DefaultConfig())
	}
	return defaultValidator
}
