package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/security/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitMiddleware_Allow(t *testing.T) {
	// Create rate limiter
	rl, err := ratelimit.NewRateLimiter(ratelimit.Config{
		Enabled:   true,
		Algorithm: ratelimit.SlidingWindow,
		Default: ratelimit.LimitConfig{
			Requests: 2,
			Window:   time.Minute,
		},
		BurstMultiplier: 1.0,
		CleanupInterval: time.Minute,
	})
	require.NoError(t, err)
	defer rl.Close()

	// Create metrics collector with test registry
	reg := prometheus.NewRegistry()
	metrics := observability.NewMetricsCollectorWithRegistry("test", reg)

	// Create middleware
	config := RateLimitConfig{
		RateLimiter:      rl,
		MetricsCollector: metrics,
		SkipPaths:        []string{"/skip"},
	}
	logger := observability.NewLogger(observability.LoggerConfig{Level: "info"})
	middleware := NewRateLimitMiddleware(config, logger)

	// Create test handler
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test allowing requests within limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())

		// Check rate limit headers
		remaining := w.Header().Get("X-RateLimit-Remaining")
		assert.NotEmpty(t, remaining)
	}

	// Test rate limit exceeded
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "rate_limit_exceeded")

	// Check rate limit headers
	retryAfter := w.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfter)
}

func TestRateLimitMiddleware_SkipPaths(t *testing.T) {
	// Create rate limiter with very low limit
	rl, err := ratelimit.NewRateLimiter(ratelimit.Config{
		Enabled:   true,
		Algorithm: ratelimit.SlidingWindow,
		Default: ratelimit.LimitConfig{
			Requests: 0, // No requests allowed
			Window:   time.Minute,
		},
		BurstMultiplier: 1.0,
		CleanupInterval: time.Minute,
	})
	require.NoError(t, err)
	defer rl.Close()

	// Create middleware with skip paths
	config := RateLimitConfig{
		RateLimiter:      rl,
		MetricsCollector: observability.NewMetricsCollector("test"),
		SkipPaths:        []string{"/health"},
	}
	logger := observability.NewLogger(observability.LoggerConfig{Level: "info"})
	middleware := NewRateLimitMiddleware(config, logger)

	// Create test handler
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test that skipped path is allowed even with 0 limit
	req := httptest.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestRateLimitMiddleware_AuthToken(t *testing.T) {
	// Create rate limiter
	rl, err := ratelimit.NewRateLimiter(ratelimit.Config{
		Enabled:   true,
		Algorithm: ratelimit.SlidingWindow,
		Default: ratelimit.LimitConfig{
			Requests: 1,
			Window:   time.Minute,
		},
		Auth: ratelimit.LimitConfig{
			Requests: 2,
			Window:   time.Minute,
		},
		BurstMultiplier: 1.0,
		CleanupInterval: time.Minute,
	})
	require.NoError(t, err)
	defer rl.Close()

	// Create middleware
	reg := prometheus.NewRegistry()
	metrics := observability.NewMetricsCollectorWithRegistry("test", reg)
	config := RateLimitConfig{
		RateLimiter:      rl,
		MetricsCollector: metrics,
	}
	logger := observability.NewLogger(observability.LoggerConfig{Level: "info"})
	middleware := NewRateLimitMiddleware(config, logger)

	// Create test handler
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test authenticated request (should use auth limits)
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Check that remaining requests reflect auth limits
	remaining := w.Header().Get("X-RateLimit-Remaining")
	assert.Equal(t, "1", remaining) // 2 requests allowed, 1 remaining
}

func TestRateLimitMiddleware_Disabled(t *testing.T) {
	// Create disabled rate limiter
	rl, err := ratelimit.NewRateLimiter(ratelimit.Config{
		Enabled: false,
	})
	require.NoError(t, err)
	defer rl.Close()

	// Create middleware
	reg := prometheus.NewRegistry()
	metrics := observability.NewMetricsCollectorWithRegistry("test", reg)
	config := RateLimitConfig{
		RateLimiter:      rl,
		MetricsCollector: metrics,
	}
	logger := observability.NewLogger(observability.LoggerConfig{Level: "info"})
	middleware := NewRateLimitMiddleware(config, logger)

	// Create test handler
	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test that all requests are allowed when disabled
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	}
}
