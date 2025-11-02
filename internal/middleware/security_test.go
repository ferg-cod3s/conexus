package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddleware(t *testing.T) {
	logger := observability.NewLogger(observability.LoggerConfig{
		Level:  "error",
		Format: "json",
	})

	tests := []struct {
		name           string
		config         SecurityConfig
		expectedHeader map[string]string
	}{
		{
			name: "default security headers",
			config: SecurityConfig{
				XFrameOptions:       "DENY",
				XContentTypeOptions: "nosniff",
				ReferrerPolicy:      "strict-origin-when-cross-origin",
				PermissionsPolicy:   "camera=(), microphone=()",
			},
			expectedHeader: map[string]string{
				"X-Frame-Options":        "DENY",
				"X-Content-Type-Options": "nosniff",
				"Referrer-Policy":        "strict-origin-when-cross-origin",
				"Permissions-Policy":     "camera=(), microphone=()",
			},
		},
		{
			name: "CSP enabled",
			config: SecurityConfig{
				CSP: CSPConfig{
					Enabled: true,
					Default: []string{"'none'"},
					Script:  []string{"'self'"},
				},
			},
			expectedHeader: map[string]string{
				"Content-Security-Policy": "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'; media-src 'none'; object-src 'none'; frame-src 'none'",
			},
		},
		{
			name: "HSTS enabled",
			config: SecurityConfig{
				HSTS: HSTSConfig{
					Enabled:           true,
					MaxAge:            31536000,
					IncludeSubdomains: true,
					Preload:           false,
				},
			},
			expectedHeader: map[string]string{
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewSecurityMiddleware(tt.config, logger)

			handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			for header, expectedValue := range tt.expectedHeader {
				actualValue := w.Header().Get(header)
				assert.Equal(t, expectedValue, actualValue, "Header %s should match", header)
			}
		})
	}
}

func TestBuildCSP(t *testing.T) {
	sm := &SecurityMiddleware{}

	tests := []struct {
		name     string
		config   CSPConfig
		expected string
	}{
		{
			name: "basic CSP",
			config: CSPConfig{
				Default: []string{"'none'"},
				Script:  []string{"'self'"},
				Style:   []string{"'self'"},
			},
			expected: "default-src 'none'; script-src 'self'; style-src 'self'",
		},
		{
			name: "CSP with report URI",
			config: CSPConfig{
				Default: []string{"'self'"},
				Report:  "https://example.com/report",
			},
			expected: "default-src 'self'; report-uri https://example.com/report",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm.config.CSP = tt.config
			result := sm.buildCSP()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildHSTS(t *testing.T) {
	sm := &SecurityMiddleware{}

	tests := []struct {
		name     string
		config   HSTSConfig
		expected string
	}{
		{
			name: "HSTS with subdomains",
			config: HSTSConfig{
				MaxAge:            31536000,
				IncludeSubdomains: true,
				Preload:           false,
			},
			expected: "max-age=31536000; includeSubDomains",
		},
		{
			name: "HSTS with preload",
			config: HSTSConfig{
				MaxAge:            63072000,
				IncludeSubdomains: true,
				Preload:           true,
			},
			expected: "max-age=63072000; includeSubDomains; preload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm.config.HSTS = tt.config
			result := sm.buildHSTS()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	assert.True(t, config.CSP.Enabled)
	assert.Equal(t, []string{"'none'"}, config.CSP.Default)
	assert.Equal(t, []string{"'self'"}, config.CSP.Script)
	assert.Equal(t, "DENY", config.XFrameOptions)
	assert.Equal(t, "nosniff", config.XContentTypeOptions)
}
