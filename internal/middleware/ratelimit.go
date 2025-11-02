package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/security/ratelimit"
)

// RateLimitConfig holds configuration for the rate limiting middleware
type RateLimitConfig struct {
	// RateLimiter is the rate limiter instance to use
	RateLimiter *ratelimit.RateLimiter

	// MetricsCollector for recording rate limiting metrics
	MetricsCollector *observability.MetricsCollector

	// SkipPaths are URL paths that should skip rate limiting
	SkipPaths []string

	// SkipIPs are IP addresses that should skip rate limiting
	SkipIPs []string

	// TrustedProxies contains CIDR ranges for trusted proxies
	TrustedProxies []string
}

// RateLimitMiddleware provides HTTP middleware for rate limiting
type RateLimitMiddleware struct {
	config RateLimitConfig
	logger *observability.Logger
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config RateLimitConfig, logger *observability.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		config: config,
		logger: logger,
	}
}

// Middleware returns an HTTP middleware function that enforces rate limits
func (rlm *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Skip rate limiting for configured paths
		if rlm.shouldSkipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP, considering trusted proxies
		clientIP := rlm.getClientIP(r)
		if rlm.shouldSkipIP(clientIP) {
			next.ServeHTTP(w, r)
			return
		}

		// Determine limiter type and identifier
		limiterType, identifier := rlm.getLimiterInfo(r, clientIP)

		// Get appropriate limit configuration for this request
		limitConfig := rlm.config.RateLimiter.GetLimitConfig(r)

		// Check rate limit
		ctx := r.Context()
		result, err := rlm.config.RateLimiter.Allow(ctx, limiterType, identifier, limitConfig)
		if err != nil {
			rlm.logger.Error("Rate limit check failed",
				"error", err,
				"limiter_type", limiterType,
				"identifier", identifier,
				"path", r.URL.Path,
				"method", r.Method,
			)
			// On error, allow the request to proceed (fail open)
			next.ServeHTTP(w, r)
			return
		}

		// Record metrics
		duration := time.Since(start)
		rlm.recordMetrics(r, result, duration, limiterType)

		// Set rate limit headers
		rlm.setRateLimitHeaders(w, result)

		if !result.Allowed {
			// Rate limit exceeded
			rlm.logger.Warn("Rate limit exceeded",
				"limiter_type", limiterType,
				"identifier", identifier,
				"path", r.URL.Path,
				"method", r.Method,
				"current_count", result.CurrentCount,
				"limit", result.Limit,
				"retry_after", result.RetryAfter,
			)

			// Return 429 Too Many Requests
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			// Simple JSON response
			response := `{"error":"rate_limit_exceeded","message":"Too many requests","retry_after":` +
				`"` + result.RetryAfter.String() + `"}`
			w.Write([]byte(response))
			return
		}

		// Request allowed, continue
		next.ServeHTTP(w, r)
	})
}

// shouldSkipPath checks if the request path should skip rate limiting
func (rlm *RateLimitMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range rlm.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// shouldSkipIP checks if the client IP should skip rate limiting
func (rlm *RateLimitMiddleware) shouldSkipIP(clientIP string) bool {
	for _, skipIP := range rlm.config.SkipIPs {
		if clientIP == skipIP {
			return true
		}
	}
	return false
}

// getClientIP extracts the real client IP, considering trusted proxies
func (rlm *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])

		// Validate the IP and check if it's from a trusted proxy
		if net.ParseIP(clientIP) != nil && rlm.isTrustedProxy(r.RemoteAddr) {
			return clientIP
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // fallback if parsing fails
	}
	return host
}

// isTrustedProxy checks if the given address is from a trusted proxy
func (rlm *RateLimitMiddleware) isTrustedProxy(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	for _, trustedCIDR := range rlm.config.TrustedProxies {
		_, network, err := net.ParseCIDR(trustedCIDR)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// getLimiterInfo determines the limiter type and identifier for the request
func (rlm *RateLimitMiddleware) getLimiterInfo(r *http.Request, clientIP string) (ratelimit.LimiterType, string) {
	// Check for authentication token first
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			// Use a hash of the token for privacy (first 8 chars should be sufficient for rate limiting)
			if len(token) > 8 {
				return ratelimit.TokenLimiter, token[:8]
			}
			return ratelimit.TokenLimiter, token
		}
	}

	// Check for API key in query parameters
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return ratelimit.TokenLimiter, apiKey[:min(8, len(apiKey))]
	}

	// Check for API key in headers
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return ratelimit.TokenLimiter, apiKey[:min(8, len(apiKey))]
	}

	// Default to IP-based limiting
	return ratelimit.IPLimiter, clientIP
}

// setRateLimitHeaders sets the standard rate limit headers on the response
func (rlm *RateLimitMiddleware) setRateLimitHeaders(w http.ResponseWriter, result *ratelimit.Result) {
	w.Header().Set("X-RateLimit-Limit", int64ToString(result.Limit))
	w.Header().Set("X-RateLimit-Remaining", int64ToString(result.Remaining))
	w.Header().Set("X-RateLimit-Reset", int64ToString(result.ResetTime.Unix()))

	if !result.Allowed {
		w.Header().Set("Retry-After", int64ToString(int64(result.RetryAfter.Seconds())))
	}
}

// recordMetrics records rate limiting metrics
func (rlm *RateLimitMiddleware) recordMetrics(r *http.Request, result *ratelimit.Result, duration time.Duration, limiterType ratelimit.LimiterType) {
	if rlm.config.MetricsCollector == nil {
		return
	}

	// Record rate limit check
	resultStr := "allowed"
	if !result.Allowed {
		resultStr = "hit"
	}
	rlm.config.MetricsCollector.RecordRateLimit(string(limiterType), resultStr, duration)

	// Update remaining requests gauge
	if result.Remaining >= 0 {
		// Use a sanitized identifier for metrics (avoid exposing sensitive info)
		identifier := "unknown"
		if limiterType == ratelimit.IPLimiter {
			// For IP limiter, use a hash or just "ip" to avoid exposing IPs in metrics
			identifier = "ip"
		} else if limiterType == ratelimit.TokenLimiter {
			// For token limiter, use a generic label
			identifier = "token"
		}
		rlm.config.MetricsCollector.UpdateRateLimitRemaining(string(limiterType), identifier, result.Remaining)
	}

	// Log significant events
	if !result.Allowed {
		rlm.logger.Info("Rate limit hit",
			"limiter_type", limiterType,
			"path", r.URL.Path,
			"method", r.Method,
			"limit", result.Limit,
			"remaining", result.Remaining,
			"retry_after", result.RetryAfter,
		)
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// int64ToString converts int64 to string
func int64ToString(n int64) string {
	return fmt.Sprintf("%d", n)
}
