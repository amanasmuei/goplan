// Package middleware provides HTTP middleware for the REST API.
package middleware

import (
	"net/http"
	"strings"
)

// SecurityConfig holds security headers configuration.
type SecurityConfig struct {
	// Content-Security-Policy settings
	CSPDefaultSrc    string
	CSPScriptSrc     string
	CSPStyleSrc      string
	CSPImgSrc        string
	CSPConnectSrc    string
	CSPFontSrc       string
	CSPObjectSrc     string
	CSPMediaSrc      string
	CSPFrameSrc      string
	CSPFrameAncestors string
	CSPReportURI     string

	// HSTS settings
	HSTSMaxAge            int
	HSTSIncludeSubDomains bool
	HSTSPreload           bool

	// Other security settings
	XContentTypeOptions    bool
	XFrameOptions          string // DENY, SAMEORIGIN, or empty to disable
	XXSSProtection         bool
	ReferrerPolicy         string
	PermissionsPolicy      string
	CrossOriginOpenerPolicy string
	CrossOriginEmbedderPolicy string
	CrossOriginResourcePolicy string

	// Enable/disable specific headers
	Enabled bool
}

// DefaultSecurityConfig returns the default security configuration.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		Enabled:               true,
		CSPDefaultSrc:         "'self'",
		CSPScriptSrc:          "'self'",
		CSPStyleSrc:           "'self' 'unsafe-inline'",
		CSPImgSrc:             "'self' data: https:",
		CSPConnectSrc:         "'self'",
		CSPFontSrc:            "'self'",
		CSPObjectSrc:          "'none'",
		CSPMediaSrc:           "'self'",
		CSPFrameSrc:           "'none'",
		CSPFrameAncestors:     "'self'",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubDomains: true,
		HSTSPreload:           false,
		XContentTypeOptions:   true,
		XFrameOptions:         "DENY",
		XXSSProtection:        true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginEmbedderPolicy: "",
		CrossOriginResourcePolicy: "same-origin",
	}
}

// APISecurityConfig returns a security configuration suitable for APIs.
func APISecurityConfig() SecurityConfig {
	return SecurityConfig{
		Enabled:              true,
		XContentTypeOptions:  true,
		XFrameOptions:        "DENY",
		ReferrerPolicy:       "no-referrer",
		HSTSMaxAge:           31536000,
		HSTSIncludeSubDomains: true,
		CrossOriginResourcePolicy: "same-origin",
	}
}

// SecurityHeaders is middleware that adds security headers to responses.
type SecurityHeaders struct {
	config SecurityConfig
}

// NewSecurityHeaders creates a new security headers middleware.
func NewSecurityHeaders(config SecurityConfig) *SecurityHeaders {
	return &SecurityHeaders{config: config}
}

// Handler returns the security headers middleware handler.
func (s *SecurityHeaders) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Content-Security-Policy
		if csp := s.buildCSP(); csp != "" {
			w.Header().Set("Content-Security-Policy", csp)
		}

		// Strict-Transport-Security (HSTS)
		if s.config.HSTSMaxAge > 0 {
			hsts := s.buildHSTS()
			w.Header().Set("Strict-Transport-Security", hsts)
		}

		// X-Content-Type-Options
		if s.config.XContentTypeOptions {
			w.Header().Set("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if s.config.XFrameOptions != "" {
			w.Header().Set("X-Frame-Options", s.config.XFrameOptions)
		}

		// X-XSS-Protection (legacy, but still useful for older browsers)
		if s.config.XXSSProtection {
			w.Header().Set("X-XSS-Protection", "1; mode=block")
		}

		// Referrer-Policy
		if s.config.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", s.config.ReferrerPolicy)
		}

		// Permissions-Policy (formerly Feature-Policy)
		if s.config.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", s.config.PermissionsPolicy)
		}

		// Cross-Origin-Opener-Policy
		if s.config.CrossOriginOpenerPolicy != "" {
			w.Header().Set("Cross-Origin-Opener-Policy", s.config.CrossOriginOpenerPolicy)
		}

		// Cross-Origin-Embedder-Policy
		if s.config.CrossOriginEmbedderPolicy != "" {
			w.Header().Set("Cross-Origin-Embedder-Policy", s.config.CrossOriginEmbedderPolicy)
		}

		// Cross-Origin-Resource-Policy
		if s.config.CrossOriginResourcePolicy != "" {
			w.Header().Set("Cross-Origin-Resource-Policy", s.config.CrossOriginResourcePolicy)
		}

		next.ServeHTTP(w, r)
	})
}

// buildCSP builds the Content-Security-Policy header value.
func (s *SecurityHeaders) buildCSP() string {
	var directives []string

	if s.config.CSPDefaultSrc != "" {
		directives = append(directives, "default-src "+s.config.CSPDefaultSrc)
	}
	if s.config.CSPScriptSrc != "" {
		directives = append(directives, "script-src "+s.config.CSPScriptSrc)
	}
	if s.config.CSPStyleSrc != "" {
		directives = append(directives, "style-src "+s.config.CSPStyleSrc)
	}
	if s.config.CSPImgSrc != "" {
		directives = append(directives, "img-src "+s.config.CSPImgSrc)
	}
	if s.config.CSPConnectSrc != "" {
		directives = append(directives, "connect-src "+s.config.CSPConnectSrc)
	}
	if s.config.CSPFontSrc != "" {
		directives = append(directives, "font-src "+s.config.CSPFontSrc)
	}
	if s.config.CSPObjectSrc != "" {
		directives = append(directives, "object-src "+s.config.CSPObjectSrc)
	}
	if s.config.CSPMediaSrc != "" {
		directives = append(directives, "media-src "+s.config.CSPMediaSrc)
	}
	if s.config.CSPFrameSrc != "" {
		directives = append(directives, "frame-src "+s.config.CSPFrameSrc)
	}
	if s.config.CSPFrameAncestors != "" {
		directives = append(directives, "frame-ancestors "+s.config.CSPFrameAncestors)
	}
	if s.config.CSPReportURI != "" {
		directives = append(directives, "report-uri "+s.config.CSPReportURI)
	}

	return strings.Join(directives, "; ")
}

// buildHSTS builds the Strict-Transport-Security header value.
func (s *SecurityHeaders) buildHSTS() string {
	// Use proper string formatting
	parts := []string{}
	parts = append(parts, "max-age="+formatInt(s.config.HSTSMaxAge))

	if s.config.HSTSIncludeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	if s.config.HSTSPreload {
		parts = append(parts, "preload")
	}

	return strings.Join(parts, "; ")
}

// formatInt formats an integer as a string.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	neg := n < 0
	if neg {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte(n%10) + '0'}, digits...)
		n /= 10
	}

	if neg {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// Security returns the default security headers middleware.
func Security() func(http.Handler) http.Handler {
	sh := NewSecurityHeaders(APISecurityConfig())
	return sh.Handler
}

// SecurityWithConfig returns security headers middleware with custom config.
func SecurityWithConfig(config SecurityConfig) func(http.Handler) http.Handler {
	sh := NewSecurityHeaders(config)
	return sh.Handler
}
