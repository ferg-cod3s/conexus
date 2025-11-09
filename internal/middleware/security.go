package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/observability"
)

// SecurityConfig holds configuration for security headers
type SecurityConfig struct {
	// Content Security Policy
	CSP CSPConfig `json:"csp" yaml:"csp"`

	// HTTP Strict Transport Security
	HSTS HSTSConfig `json:"hsts" yaml:"hsts"`

	// X-Frame-Options
	XFrameOptions string `json:"x_frame_options" yaml:"x_frame_options"`

	// X-Content-Type-Options
	XContentTypeOptions string `json:"x_content_type_options" yaml:"x_content_type_options"`

	// Referrer-Policy
	ReferrerPolicy string `json:"referrer_policy" yaml:"referrer_policy"`

	// Permissions-Policy
	PermissionsPolicy string `json:"permissions_policy" yaml:"permissions_policy"`
}

// CSPConfig holds Content Security Policy configuration
type CSPConfig struct {
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Default []string `json:"default" yaml:"default"`
	Script  []string `json:"script" yaml:"script"`
	Style   []string `json:"style" yaml:"style"`
	Image   []string `json:"image" yaml:"image"`
	Font    []string `json:"font" yaml:"font"`
	Connect []string `json:"connect" yaml:"connect"`
	Media   []string `json:"media" yaml:"media"`
	Object  []string `json:"object" yaml:"object"`
	Frame   []string `json:"frame" yaml:"frame"`
	Report  string   `json:"report" yaml:"report"`
}

// HSTSConfig holds HTTP Strict Transport Security configuration
type HSTSConfig struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	MaxAge            int  `json:"max_age" yaml:"max_age"`
	IncludeSubdomains bool `json:"include_subdomains" yaml:"include_subdomains"`
	Preload           bool `json:"preload" yaml:"preload"`
}

// SecurityMiddleware provides security headers middleware
type SecurityMiddleware struct {
	config SecurityConfig
	logger *observability.Logger
}

// NewSecurityMiddleware creates a new security middleware with default restrictive configuration
func NewSecurityMiddleware(config SecurityConfig, logger *observability.Logger) *SecurityMiddleware {
	// Set restrictive defaults if not configured
	if config.XFrameOptions == "" {
		config.XFrameOptions = "DENY"
	}
	if config.XContentTypeOptions == "" {
		config.XContentTypeOptions = "nosniff"
	}
	if config.ReferrerPolicy == "" {
		config.ReferrerPolicy = "strict-origin-when-cross-origin"
	}
	if config.PermissionsPolicy == "" {
		config.PermissionsPolicy = "camera=(), microphone=(), geolocation=(), payment=()"
	}

	// Set CSP defaults if enabled but not configured
	if config.CSP.Enabled {
		if len(config.CSP.Default) == 0 {
			config.CSP.Default = []string{"'none'"}
		}
		if len(config.CSP.Script) == 0 {
			config.CSP.Script = []string{"'self'"}
		}
		if len(config.CSP.Style) == 0 {
			config.CSP.Style = []string{"'self'"}
		}
		if len(config.CSP.Image) == 0 {
			config.CSP.Image = []string{"'self'"}
		}
		if len(config.CSP.Font) == 0 {
			config.CSP.Font = []string{"'self'"}
		}
		if len(config.CSP.Connect) == 0 {
			config.CSP.Connect = []string{"'self'"}
		}
		if len(config.CSP.Media) == 0 {
			config.CSP.Media = []string{"'none'"}
		}
		if len(config.CSP.Object) == 0 {
			config.CSP.Object = []string{"'none'"}
		}
		if len(config.CSP.Frame) == 0 {
			config.CSP.Frame = []string{"'none'"}
		}
	}

	// Set HSTS defaults if enabled but not configured
	if config.HSTS.Enabled && config.HSTS.MaxAge == 0 {
		config.HSTS.MaxAge = 31536000 // 1 year
	}

	return &SecurityMiddleware{
		config: config,
		logger: logger,
	}
}

// Middleware returns an HTTP middleware function that adds security headers
func (sm *SecurityMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add Content Security Policy
		if sm.config.CSP.Enabled {
			csp := sm.buildCSP()
			w.Header().Set("Content-Security-Policy", csp)
		}

		// Add HTTP Strict Transport Security
		if sm.config.HSTS.Enabled {
			hsts := sm.buildHSTS()
			w.Header().Set("Strict-Transport-Security", hsts)
		}

		// Add X-Frame-Options
		if sm.config.XFrameOptions != "" {
			w.Header().Set("X-Frame-Options", sm.config.XFrameOptions)
		}

		// Add X-Content-Type-Options
		if sm.config.XContentTypeOptions != "" {
			w.Header().Set("X-Content-Type-Options", sm.config.XContentTypeOptions)
		}

		// Add Referrer-Policy
		if sm.config.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", sm.config.ReferrerPolicy)
		}

		// Add Permissions-Policy
		if sm.config.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", sm.config.PermissionsPolicy)
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)

		// Log security headers application
		duration := time.Since(start)
		sm.logger.Debug("Applied security headers",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", duration.Milliseconds(),
			"csp_enabled", sm.config.CSP.Enabled,
			"hsts_enabled", sm.config.HSTS.Enabled,
		)
	})
}

// buildCSP constructs the Content Security Policy header value
func (sm *SecurityMiddleware) buildCSP() string {
	var directives []string

	// Default policy
	if len(sm.config.CSP.Default) > 0 {
		directives = append(directives, "default-src "+strings.Join(sm.config.CSP.Default, " "))
	}

	// Script sources
	if len(sm.config.CSP.Script) > 0 {
		directives = append(directives, "script-src "+strings.Join(sm.config.CSP.Script, " "))
	}

	// Style sources
	if len(sm.config.CSP.Style) > 0 {
		directives = append(directives, "style-src "+strings.Join(sm.config.CSP.Style, " "))
	}

	// Image sources
	if len(sm.config.CSP.Image) > 0 {
		directives = append(directives, "img-src "+strings.Join(sm.config.CSP.Image, " "))
	}

	// Font sources
	if len(sm.config.CSP.Font) > 0 {
		directives = append(directives, "font-src "+strings.Join(sm.config.CSP.Font, " "))
	}

	// Connect sources
	if len(sm.config.CSP.Connect) > 0 {
		directives = append(directives, "connect-src "+strings.Join(sm.config.CSP.Connect, " "))
	}

	// Media sources
	if len(sm.config.CSP.Media) > 0 {
		directives = append(directives, "media-src "+strings.Join(sm.config.CSP.Media, " "))
	}

	// Object sources
	if len(sm.config.CSP.Object) > 0 {
		directives = append(directives, "object-src "+strings.Join(sm.config.CSP.Object, " "))
	}

	// Frame sources
	if len(sm.config.CSP.Frame) > 0 {
		directives = append(directives, "frame-src "+strings.Join(sm.config.CSP.Frame, " "))
	}

	// Report URI
	if sm.config.CSP.Report != "" {
		directives = append(directives, "report-uri "+sm.config.CSP.Report)
	}

	return strings.Join(directives, "; ")
}

// buildHSTS constructs the HTTP Strict Transport Security header value
func (sm *SecurityMiddleware) buildHSTS() string {
	hsts := fmt.Sprintf("max-age=%d", sm.config.HSTS.MaxAge)

	if sm.config.HSTS.IncludeSubdomains {
		hsts += "; includeSubDomains"
	}

	if sm.config.HSTS.Preload {
		hsts += "; preload"
	}

	return hsts
}

// DefaultSecurityConfig returns a restrictive default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		CSP: CSPConfig{
			Enabled: true,
			Default: []string{"'none'"},
			Script:  []string{"'self'"},
			Style:   []string{"'self'"},
			Image:   []string{"'self'"},
			Font:    []string{"'self'"},
			Connect: []string{"'self'"},
			Media:   []string{"'none'"},
			Object:  []string{"'none'"},
			Frame:   []string{"'none'"},
		},
		HSTS: HSTSConfig{
			Enabled:           true,
			MaxAge:            31536000, // 1 year
			IncludeSubdomains: true,
			Preload:           false,
		},
		XFrameOptions:       "DENY",
		XContentTypeOptions: "nosniff",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
		PermissionsPolicy:   "camera=(), microphone=(), geolocation=(), payment=()",
	}
}
