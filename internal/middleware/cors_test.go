package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	logger := observability.NewLogger(observability.LoggerConfig{
		Level:  "error",
		Format: "json",
	})

	tests := []struct {
		name            string
		config          CORSConfig
		requestOrigin   string
		requestMethod   string
		requestHeaders  string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name: "CORS disabled - no headers added",
			config: CORSConfig{
				Enabled: false,
			},
			requestOrigin:   "http://example.com",
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{},
		},
		{
			name: "CORS enabled - allowed origin",
			config: CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"http://example.com"},
				AllowCredentials: true,
				ExposedHeaders:   []string{"X-Custom-Header"},
			},
			requestOrigin:  "http://example.com",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Expose-Headers":    "X-Custom-Header",
			},
		},
		{
			name: "CORS enabled - disallowed origin",
			config: CORSConfig{
				Enabled:        true,
				AllowedOrigins: []string{"http://allowed.com"},
			},
			requestOrigin:   "http://disallowed.com",
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{}, // No CORS headers for disallowed origins
		},
		{
			name: "CORS preflight - allowed",
			config: CORSConfig{
				Enabled:          true,
				AllowedOrigins:   []string{"http://example.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			requestOrigin:  "http://example.com",
			requestMethod:  "OPTIONS",
			requestHeaders: "content-type,authorization",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "http://example.com",
				"Access-Control-Allow-Methods":     "GET, POST, PUT",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "3600",
			},
		},
		{
			name: "CORS preflight - disallowed origin",
			config: CORSConfig{
				Enabled:        true,
				AllowedOrigins: []string{"http://allowed.com"},
			},
			requestOrigin:   "http://disallowed.com",
			requestMethod:   "OPTIONS",
			expectedStatus:  http.StatusForbidden,
			expectedHeaders: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCORSMiddleware(tt.config, logger)

			handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.requestMethod, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			if tt.requestHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tt.requestHeaders)
			}
			if tt.requestMethod == "OPTIONS" {
				req.Header.Set("Access-Control-Request-Method", "PUT")
			}

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			for header, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(header)
				if header == "Access-Control-Allow-Methods" || header == "Access-Control-Allow-Headers" {
					// For these headers, just check that the expected value is contained
					assert.Contains(t, actualValue, expectedValue, "Header %s should contain expected value", header)
				} else {
					assert.Equal(t, expectedValue, actualValue, "Header %s should match", header)
				}
			}
		})
	}
}

func TestIsOriginAllowed(t *testing.T) {
	cm := &CORSMiddleware{
		config: CORSConfig{
			Enabled:        true,
			AllowedOrigins: []string{"http://example.com", "https://*.domain.com"},
		},
	}

	tests := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"exact match", "http://example.com", true},
		{"wildcard match", "https://sub.domain.com", true},
		{"no match", "http://other.com", false},
		{"empty origin", "", false},
		{"disabled CORS", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "disabled CORS" {
				cm.config.Enabled = false
				defer func() { cm.config.Enabled = true }()
			}
			result := cm.isOriginAllowed(tt.origin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.False(t, config.Enabled)
	assert.Empty(t, config.AllowedOrigins)
	assert.Equal(t, []string{"GET", "POST"}, config.AllowedMethods)
	assert.Equal(t, []string{"Content-Type", "Authorization"}, config.AllowedHeaders)
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
}
