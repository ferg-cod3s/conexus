package ratelimit

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_SlidingWindow(t *testing.T) {
	config := Config{
		Enabled:   true,
		Algorithm: SlidingWindow,
		Default: LimitConfig{
			Requests: 5,
			Window:   time.Minute,
		},
		BurstMultiplier: 1.0,
		CleanupInterval: time.Minute,
	}

	rl, err := NewRateLimiter(config)
	require.NoError(t, err)
	defer rl.Close()

	ctx := context.Background()

	// Test allowing requests within limit
	for i := 0; i < 5; i++ {
		result, err := rl.Allow(ctx, IPLimiter, "127.0.0.1", config.Default)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.Equal(t, int64(5-i-1), result.Remaining)
		assert.Equal(t, int64(5), result.Limit)
	}

	// Test rate limit exceeded
	result, err := rl.Allow(ctx, IPLimiter, "127.0.0.1", config.Default)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
	assert.Equal(t, int64(5), result.Limit)
	assert.True(t, result.RetryAfter > 0)
}

func TestRateLimiter_TokenBucket(t *testing.T) {
	config := Config{
		Enabled:   true,
		Algorithm: TokenBucket,
		Default: LimitConfig{
			Requests: 10, // 10 requests per minute = ~0.167 requests per second
			Window:   time.Minute,
		},
		BurstMultiplier: 2.0, // Allow burst up to 20 tokens
		CleanupInterval: time.Minute,
	}

	rl, err := NewRateLimiter(config)
	require.NoError(t, err)
	defer rl.Close()

	ctx := context.Background()

	// Test allowing burst requests
	for i := 0; i < 20; i++ {
		result, err := rl.Allow(ctx, IPLimiter, "127.0.0.2", config.Default)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}

	// Test rate limit exceeded after burst
	result, err := rl.Allow(ctx, IPLimiter, "127.0.0.2", config.Default)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.True(t, result.RetryAfter > 0)
}

func TestRateLimiter_DifferentLimiters(t *testing.T) {
	config := Config{
		Enabled:   true,
		Algorithm: SlidingWindow,
		Default: LimitConfig{
			Requests: 3,
			Window:   time.Minute,
		},
		BurstMultiplier: 1.0,
		CleanupInterval: time.Minute,
	}

	rl, err := NewRateLimiter(config)
	require.NoError(t, err)
	defer rl.Close()

	ctx := context.Background()

	// Test IP-based limiting
	for i := 0; i < 3; i++ {
		result, err := rl.Allow(ctx, IPLimiter, "192.168.1.1", config.Default)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}

	result, err := rl.Allow(ctx, IPLimiter, "192.168.1.1", config.Default)
	require.NoError(t, err)
	assert.False(t, result.Allowed)

	// Test token-based limiting (different limiter, should not be affected)
	result, err = rl.Allow(ctx, TokenLimiter, "token123", config.Default)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestRateLimiter_Disabled(t *testing.T) {
	config := Config{
		Enabled:   false,
		Algorithm: SlidingWindow,
		Default: LimitConfig{
			Requests: 1,
			Window:   time.Minute,
		},
	}

	rl, err := NewRateLimiter(config)
	require.NoError(t, err)
	defer rl.Close()

	ctx := context.Background()

	// Should always allow when disabled
	for i := 0; i < 10; i++ {
		result, err := rl.Allow(ctx, IPLimiter, "127.0.0.1", config.Default)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
	}
}

func TestRateLimiter_GetLimitConfig(t *testing.T) {
	rl, err := NewRateLimiter(DefaultConfig())
	require.NoError(t, err)
	defer rl.Close()

	tests := []struct {
		path     string
		expected LimitConfig
	}{
		{"/health", DefaultConfig().Health},
		{"/health/check", DefaultConfig().Health},
		{"/webhook/test", DefaultConfig().Webhook},
		{"/mcp/webhooks/test", DefaultConfig().Webhook},
		{"/api/test", DefaultConfig().Default},
		{"/", DefaultConfig().Default},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			require.NoError(t, err)

			result := rl.GetLimitConfig(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRateLimiter_HasAuthToken(t *testing.T) {
	rl, err := NewRateLimiter(DefaultConfig())
	require.NoError(t, err)
	defer rl.Close()

	tests := []struct {
		name     string
		headers  map[string]string
		query    string
		expected bool
	}{
		{
			name:     "bearer token",
			headers:  map[string]string{"Authorization": "Bearer abc123"},
			expected: true,
		},
		{
			name:     "token auth",
			headers:  map[string]string{"Authorization": "Token xyz789"},
			expected: true,
		},
		{
			name:     "api key header",
			headers:  map[string]string{"X-API-Key": "key123"},
			expected: true,
		},
		{
			name:     "api key query",
			query:    "api_key=query123",
			expected: true,
		},
		{
			name:     "no auth",
			headers:  map[string]string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example.com/test?"+tt.query, nil)
			require.NoError(t, err)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := rl.hasAuthToken(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, SlidingWindow, config.Algorithm)
	assert.Equal(t, 100, config.Default.Requests)
	assert.Equal(t, time.Minute, config.Default.Window)
	assert.Equal(t, 1000, config.Health.Requests)
	assert.Equal(t, 10000, config.Webhook.Requests)
	assert.Equal(t, 1000, config.Auth.Requests)
	assert.Equal(t, 1.2, config.BurstMultiplier)
	assert.Equal(t, 5*time.Minute, config.CleanupInterval)
}
