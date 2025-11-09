package integration

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/config"
	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/middleware"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/security/ratelimit"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitIntegration(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary database for testing
	tempDB := t.TempDir() + "/test.db"

	// Load minimal config for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 0, // Use random port
		},
		Database: config.DatabaseConfig{
			Path: tempDB,
		},
		Embedding: config.EmbeddingConfig{
			Provider:   "mock",
			Model:      "mock-768",
			Dimensions: 768,
		},
		RateLimit: config.RateLimitConfig{
			Enabled:         true,
			Algorithm:       "sliding_window",
			BurstMultiplier: 1.0, // Allow bursting up to the full rate limit
			Default: config.RateLimitRuleConfig{
				Requests: 10,          // Even higher limit for testing
				Window:   time.Minute, // 1 minute window
			},
			Health: config.RateLimitRuleConfig{
				Requests: 10,
				Window:   time.Minute,
			},
		},
		Security: config.SecurityConfig{
			CSP: config.CSPConfig{
				Enabled: false, // Disable for testing
			},
			HSTS: config.HSTSConfig{
				Enabled: false, // Disable for testing
			},
		},
		CORS: config.CORSConfig{
			Enabled: false, // Disable for testing
		},
	}

	// Initialize components
	logger := observability.NewLogger(observability.LoggerConfig{
		Level:  "error", // Reduce log noise
		Format: "json",
	})

	// Initialize vector store
	vectorStore, err := sqlite.NewStore(cfg.Database.Path)
	require.NoError(t, err)
	defer vectorStore.Close()

	// Initialize connector store
	connectorStore, err := connectors.NewStore(cfg.Database.Path)
	require.NoError(t, err)
	defer connectorStore.Close()

	// Initialize embedder (needed for rate limiter config conversion)
	provider, err := embedding.Get(cfg.Embedding.Provider)
	require.NoError(t, err)

	providerConfig := make(map[string]interface{})
	providerConfig["model"] = cfg.Embedding.Model
	providerConfig["dimensions"] = cfg.Embedding.Dimensions

	_, err = provider.Create(providerConfig)
	require.NoError(t, err)

	// Initialize rate limiter
	rateLimitConfig := ratelimit.Config{
		Enabled: cfg.RateLimit.Enabled,
		Algorithm: func() ratelimit.Algorithm {
			switch cfg.RateLimit.Algorithm {
			case "token_bucket":
				return ratelimit.TokenBucket
			case "sliding_window":
				return ratelimit.SlidingWindow
			default:
				return ratelimit.SlidingWindow
			}
		}(),
		Default: ratelimit.LimitConfig{
			Requests: cfg.RateLimit.Default.Requests,
			Window:   cfg.RateLimit.Default.Window,
		},
		Health: ratelimit.LimitConfig{
			Requests: cfg.RateLimit.Health.Requests,
			Window:   cfg.RateLimit.Health.Window,
		},
	}

	rateLimiter, err := ratelimit.NewRateLimiter(rateLimitConfig)
	require.NoError(t, err)

	// Initialize middleware
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(middleware.RateLimitConfig{
		RateLimiter:    rateLimiter,
		SkipPaths:      cfg.RateLimit.SkipPaths,
		SkipIPs:        cfg.RateLimit.SkipIPs,
		TrustedProxies: cfg.RateLimit.TrustedProxies,
	}, logger)

	securityMiddleware := middleware.NewSecurityMiddleware(middleware.SecurityConfig{
		CSP:  middleware.CSPConfig{Enabled: false},
		HSTS: middleware.HSTSConfig{Enabled: false},
	}, logger)

	corsMiddleware := middleware.NewCORSMiddleware(middleware.CORSConfig{Enabled: false}, logger)

	// Setup HTTP server with middleware
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","version":"test"}`)
	})

	// MCP endpoint
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}`)
	})

	// Apply middleware
	var handler http.Handler = mux
	handler = rateLimitMiddleware.Middleware(handler)
	handler = corsMiddleware.Middleware(handler)
	handler = securityMiddleware.Middleware(handler)

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Test 1: Health endpoint should allow more requests (10 per minute)
	t.Run("HealthEndpoint", func(t *testing.T) {
		for i := 0; i < 3; i++ { // Try 3 requests, should allow all
			req, err := http.NewRequest("GET", server.URL+"/health", nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if i < 2 { // First 2 should succeed
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Contains(t, string(body), "healthy")
			} else { // 3rd might be rate limited depending on timing
				if resp.StatusCode == http.StatusTooManyRequests {
					assert.Contains(t, string(body), "rate_limit_exceeded")
				} else {
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				}
			}
		}
	})

	// Test 2: MCP endpoint should be rate limited
	t.Run("MCPEndpoint", func(t *testing.T) {
		allowedCount := 0
		blockedCount := 0

		for i := 0; i < 15; i++ { // Try 15 requests to ensure we hit the limit
			req, err := http.NewRequest("POST", server.URL+"/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if resp.StatusCode == http.StatusOK {
				allowedCount++
				assert.Contains(t, string(body), "tools", "Allowed request should return tools")
			} else if resp.StatusCode == http.StatusTooManyRequests {
				blockedCount++
				assert.Contains(t, string(body), "rate_limit_exceeded", "Blocked request should return rate limit error")

				// Check rate limit headers on blocked requests
				assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"), "Blocked request should have limit header")
				assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"), "Blocked request should have remaining header")
				assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"), "Blocked request should have reset header")
				assert.NotEmpty(t, resp.Header.Get("Retry-After"), "Blocked request should have retry-after header")
			} else {
				t.Errorf("Unexpected status code: %d", resp.StatusCode)
			}

			// Small delay to ensure requests are processed
			time.Sleep(10 * time.Millisecond)
		}

		// Verify that some requests were allowed and some were blocked
		assert.Greater(t, allowedCount, 0, "At least some requests should be allowed")
		assert.Greater(t, blockedCount, 0, "At least some requests should be blocked")
		t.Logf("Rate limiting test: %d requests allowed, %d requests blocked", allowedCount, blockedCount)
	})

	// Test 3: Verify rate limit headers are present on allowed requests
	t.Run("RateLimitHeaders", func(t *testing.T) {
		// Wait a bit to reset rate limits if needed
		time.Sleep(100 * time.Millisecond)

		req, err := http.NewRequest("GET", server.URL+"/health", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should have rate limit headers
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"))
	})
}
