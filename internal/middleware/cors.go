package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/observability"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// CORSMiddleware provides CORS handling middleware
type CORSMiddleware struct {
	config CORSConfig
	logger *observability.Logger
}

// NewCORSMiddleware creates a new CORS middleware with default restrictive configuration
func NewCORSMiddleware(config CORSConfig, logger *observability.Logger) *CORSMiddleware {
	// Set restrictive defaults if not configured
	if !config.Enabled {
		return &CORSMiddleware{
			config: config,
			logger: logger,
		}
	}

	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{} // Empty means deny all by default
	}

	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST"} // Restrictive defaults
	}

	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Content-Type", "Authorization"} // Minimal required headers
	}

	if config.MaxAge == 0 {
		config.MaxAge = 86400 // 24 hours
	}

	return &CORSMiddleware{
		config: config,
		logger: logger,
	}
}

// Middleware returns an HTTP middleware function that handles CORS
func (cm *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			cm.handlePreflight(w, r)
			return
		}

		// Handle actual requests
		origin := r.Header.Get("Origin")
		if origin != "" && cm.config.Enabled {
			if cm.isOriginAllowed(origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				if cm.config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				if len(cm.config.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(cm.config.ExposedHeaders, ", "))
				}
			} else {
				// Origin not allowed - log security event
				cm.logger.Warn("CORS request from disallowed origin blocked",
					"origin", origin,
					"method", r.Method,
					"path", r.URL.Path,
					"user_agent", r.Header.Get("User-Agent"),
				)
			}
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)

		// Log CORS handling
		duration := time.Since(start)
		cm.logger.Debug("Handled CORS request",
			"method", r.Method,
			"path", r.URL.Path,
			"origin", origin,
			"cors_enabled", cm.config.Enabled,
			"origin_allowed", cm.isOriginAllowed(origin),
			"duration_ms", duration.Milliseconds(),
		)
	})
}

// handlePreflight handles CORS preflight OPTIONS requests
func (cm *CORSMiddleware) handlePreflight(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	requestMethod := r.Header.Get("Access-Control-Request-Method")
	requestHeaders := r.Header.Get("Access-Control-Request-Headers")

	// Check if origin is allowed
	if !cm.config.Enabled || !cm.isOriginAllowed(origin) {
		cm.logger.Warn("CORS preflight from disallowed origin blocked",
			"origin", origin,
			"request_method", requestMethod,
			"request_headers", requestHeaders,
		)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Set CORS headers for preflight response
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(cm.getAllowedMethods(requestMethod), ", "))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(cm.getAllowedHeaders(requestHeaders), ", "))
	w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cm.config.MaxAge))

	if cm.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	w.WriteHeader(http.StatusOK)
}

// isOriginAllowed checks if the given origin is allowed
func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	if !cm.config.Enabled || origin == "" {
		return false
	}

	// Parse the origin URL
	originURL, err := url.Parse(origin)
	if err != nil {
		cm.logger.Debug("Failed to parse origin URL", "origin", origin, "error", err)
		return false
	}

	// Check against allowed origins
	for _, allowed := range cm.config.AllowedOrigins {
		if cm.matchesOrigin(allowed, originURL) {
			return true
		}
	}

	return false
}

// matchesOrigin checks if an allowed origin pattern matches the actual origin
func (cm *CORSMiddleware) matchesOrigin(allowed string, originURL *url.URL) bool {
	// Exact match
	if allowed == originURL.String() {
		return true
	}

	// Handle wildcard patterns
	if strings.Contains(allowed, "*") {
		// Convert wildcard pattern to regex
		pattern := strings.ReplaceAll(regexp.QuoteMeta(allowed), "\\*", ".*")
		matched, err := regexp.MatchString("^"+pattern+"$", originURL.String())
		if err == nil && matched {
			return true
		}

		// Also check host-only matching for *.domain.com patterns
		if strings.HasPrefix(allowed, "*.") {
			suffix := allowed[1:] // Remove *. prefix
			if strings.HasSuffix(originURL.Host, suffix) {
				return true
			}
		}
	}

	return false
}

// getAllowedMethods returns allowed methods, ensuring the requested method is included
func (cm *CORSMiddleware) getAllowedMethods(requestMethod string) []string {
	methods := make([]string, len(cm.config.AllowedMethods))
	copy(methods, cm.config.AllowedMethods)

	// Ensure the requested method is allowed
	found := false
	for _, method := range methods {
		if method == requestMethod {
			found = true
			break
		}
	}

	if !found && requestMethod != "" {
		methods = append(methods, requestMethod)
	}

	return methods
}

// getAllowedHeaders returns allowed headers, ensuring requested headers are included
func (cm *CORSMiddleware) getAllowedHeaders(requestHeaders string) []string {
	headers := make([]string, len(cm.config.AllowedHeaders))
	copy(headers, cm.config.AllowedHeaders)

	if requestHeaders != "" {
		requested := strings.Split(requestHeaders, ",")
		for _, reqHeader := range requested {
			reqHeader = strings.TrimSpace(reqHeader)
			found := false
			for _, allowed := range headers {
				if strings.EqualFold(allowed, reqHeader) {
					found = true
					break
				}
			}
			if !found {
				headers = append(headers, reqHeader)
			}
		}
	}

	return headers
}

// DefaultCORSConfig returns a restrictive default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		Enabled:          false, // Disabled by default for security
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}
