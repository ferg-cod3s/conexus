// Package ratelimit provides distributed rate limiting with Redis backend.
// Supports sliding window and token bucket algorithms with configurable limits.
package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Algorithm represents the rate limiting algorithm to use
type Algorithm string

const (
	// SlidingWindow uses a sliding window algorithm for smooth rate limiting
	SlidingWindow Algorithm = "sliding_window"
	// TokenBucket uses a token bucket algorithm allowing burst capacity
	// nosemgrep: go-hardcoded-credentials
	TokenBucket Algorithm = "token_bucket" // Algorithm name, not a credential
)

// LimiterType represents the type of rate limiter (IP-based or token-based)
type LimiterType string

const (
	// IPLimiter limits by client IP address
	IPLimiter LimiterType = "ip"
	// TokenLimiter limits by authentication token/API key
	// nosemgrep: go-hardcoded-credentials
	TokenLimiter LimiterType = "token" // Limiter type name, not a credential
)

// Config holds rate limiting configuration
type Config struct {
	// Enabled determines if rate limiting is active
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Algorithm to use for rate limiting
	Algorithm Algorithm `json:"algorithm" yaml:"algorithm"`

	// Redis configuration for distributed rate limiting
	Redis RedisConfig `json:"redis" yaml:"redis"`

	// Default limits for different endpoint types
	Default LimitConfig `json:"default" yaml:"default"`
	Health  LimitConfig `json:"health" yaml:"health"`
	Webhook LimitConfig `json:"webhook" yaml:"webhook"`
	Auth    LimitConfig `json:"auth" yaml:"auth"`

	// BurstMultiplier allows burst capacity above the rate limit
	BurstMultiplier float64 `json:"burst_multiplier" yaml:"burst_multiplier"`

	// CleanupInterval for cleaning up expired entries (for in-memory fallback)
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Addr      string `json:"addr" yaml:"addr"`
	Password  string `json:"password" yaml:"password"`
	DB        int    `json:"db" yaml:"db"`
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
}

// LimitConfig holds rate limit configuration for a specific endpoint type
type LimitConfig struct {
	Requests int           `json:"requests" yaml:"requests"` // requests per window
	Window   time.Duration `json:"window" yaml:"window"`     // time window
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	config   Config
	redis    *redis.Client
	inMemory *InMemoryLimiter // fallback when Redis is unavailable
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config Config) (*RateLimiter, error) {
	rl := &RateLimiter{
		config:   config,
		inMemory: NewInMemoryLimiter(config.CleanupInterval),
	}

	if config.Redis.Enabled {
		rl.redis = redis.NewClient(&redis.Options{
			Addr:     config.Redis.Addr,
			Password: config.Redis.Password,
			DB:       config.Redis.DB,
		})

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rl.redis.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %w", err)
		}
	}

	return rl, nil
}

// Allow checks if a request should be allowed based on the rate limit
func (rl *RateLimiter) Allow(ctx context.Context, limiterType LimiterType, identifier string, limitConfig LimitConfig) (*Result, error) {
	if !rl.config.Enabled {
		return &Result{Allowed: true}, nil
	}

	key := rl.buildKey(limiterType, identifier)

	switch rl.config.Algorithm {
	case SlidingWindow:
		return rl.allowSlidingWindow(ctx, key, limitConfig)
	case TokenBucket:
		return rl.allowTokenBucket(ctx, key, limitConfig)
	default:
		return rl.allowSlidingWindow(ctx, key, limitConfig) // default to sliding window
	}
}

// allowSlidingWindow implements sliding window rate limiting
func (rl *RateLimiter) allowSlidingWindow(ctx context.Context, key string, limitConfig LimitConfig) (*Result, error) {
	now := time.Now().UnixMilli()
	windowStart := now - limitConfig.Window.Milliseconds()

	if rl.redis != nil {
		return rl.allowSlidingWindowRedis(ctx, key, limitConfig, now, windowStart)
	}

	return rl.inMemory.AllowSlidingWindow(key, limitConfig, now, windowStart)
}

// allowSlidingWindowRedis implements sliding window with Redis
func (rl *RateLimiter) allowSlidingWindowRedis(ctx context.Context, key string, limitConfig LimitConfig, now, windowStart int64) (*Result, error) {
	// Add current request timestamp
	if err := rl.redis.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now}).Err(); err != nil {
		return nil, fmt.Errorf("failed to add request to Redis: %w", err)
	}

	// Remove old entries outside the window
	if err := rl.redis.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("(%d", windowStart)).Err(); err != nil {
		return nil, fmt.Errorf("failed to remove old entries: %w", err)
	}

	// Count requests in current window
	count, err := rl.redis.ZCard(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count requests: %w", err)
	}

	// Set expiration on the key (window duration + buffer)
	if err := rl.redis.Expire(ctx, key, limitConfig.Window*2).Err(); err != nil {
		return nil, fmt.Errorf("failed to set expiration: %w", err)
	}

	allowed := count <= int64(limitConfig.Requests)
	var retryAfter time.Duration

	if !allowed {
		// Calculate retry-after based on oldest request in window
		oldest, err := rl.redis.ZRangeWithScores(ctx, key, 0, 0).Result()
		if err == nil && len(oldest) > 0 {
			oldestTime := int64(oldest[0].Score)
			retryAfter = time.Duration(windowStart-oldestTime) * time.Millisecond
			if retryAfter < 0 {
				retryAfter = limitConfig.Window
			}
		} else {
			retryAfter = limitConfig.Window
		}
	}

	return &Result{
		Allowed:      allowed,
		Remaining:    max(0, int64(limitConfig.Requests)-count),
		RetryAfter:   retryAfter,
		ResetTime:    time.UnixMilli(now + limitConfig.Window.Milliseconds()),
		CurrentCount: count,
		Limit:        int64(limitConfig.Requests),
	}, nil
}

// allowTokenBucket implements token bucket rate limiting
func (rl *RateLimiter) allowTokenBucket(ctx context.Context, key string, limitConfig LimitConfig) (*Result, error) {
	now := time.Now()
	rate := float64(limitConfig.Requests) / limitConfig.Window.Seconds()
	burst := int(float64(limitConfig.Requests) * rl.config.BurstMultiplier)

	if rl.redis != nil {
		return rl.allowTokenBucketRedis(ctx, key, rate, burst, now)
	}

	return rl.inMemory.AllowTokenBucket(key, rate, burst, now)
}

// allowTokenBucketRedis implements token bucket with Redis
func (rl *RateLimiter) allowTokenBucketRedis(ctx context.Context, key string, rate float64, burst int, now time.Time) (*Result, error) {
	// Use Redis Lua script for atomic token bucket operations
	script := `
		local key = KEYS[1]
		local rate = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local data = redis.call('HMGET', key, 'tokens', 'last_update')
		local tokens = tonumber(data[1]) or burst
		local last_update = tonumber(data[2]) or now

		local elapsed = now - last_update
		local new_tokens = math.min(burst, tokens + elapsed * rate)

		local allowed = new_tokens >= 1

		if allowed then
			new_tokens = new_tokens - 1
		end

		redis.call('HMSET', key, 'tokens', new_tokens, 'last_update', now)
		redis.call('EXPIRE', key, math.ceil(burst / rate * 2))

		return {allowed and 1 or 0, new_tokens, math.ceil((1 - new_tokens) / rate)}
	`

	result, err := rl.redis.Eval(ctx, script, []string{key}, rate, burst, now.Unix()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to execute token bucket script: %w", err)
	}

	results := result.([]interface{})
	allowed := results[0].(int64) == 1
	remainingTokens := results[1].(int64)
	retryAfterSeconds := results[2].(int64)

	return &Result{
		Allowed:      allowed,
		Remaining:    remainingTokens,
		RetryAfter:   time.Duration(retryAfterSeconds) * time.Second,
		ResetTime:    now.Add(time.Duration(float64(burst)/rate) * time.Second),
		CurrentCount: int64(burst) - remainingTokens,
		Limit:        int64(burst),
	}, nil
}

// buildKey creates a Redis key for the rate limiter
func (rl *RateLimiter) buildKey(limiterType LimiterType, identifier string) string {
	prefix := "ratelimit"
	if rl.config.Redis.KeyPrefix != "" {
		prefix = rl.config.Redis.KeyPrefix
	}

	// Sanitize identifier to prevent Redis injection
	sanitizedID := strings.ReplaceAll(identifier, ":", "_")
	sanitizedID = strings.ReplaceAll(sanitizedID, " ", "_")

	return fmt.Sprintf("%s:%s:%s", prefix, limiterType, sanitizedID)
}

// Result represents the result of a rate limit check
type Result struct {
	Allowed      bool          `json:"allowed"`
	Remaining    int64         `json:"remaining"`
	RetryAfter   time.Duration `json:"retry_after"`
	ResetTime    time.Time     `json:"reset_time"`
	CurrentCount int64         `json:"current_count"`
	Limit        int64         `json:"limit"`
}

// GetLimitConfig returns the appropriate limit configuration based on the request path
func (rl *RateLimiter) GetLimitConfig(r *http.Request) LimitConfig {
	path := r.URL.Path

	// Health endpoints
	if strings.HasPrefix(path, "/health") {
		return rl.config.Health
	}

	// Webhook endpoints
	if strings.HasPrefix(path, "/webhook") || strings.HasPrefix(path, "/mcp/webhooks") {
		return rl.config.Webhook
	}

	// Authenticated endpoints (check for auth header or token)
	if rl.hasAuthToken(r) {
		return rl.config.Auth
	}

	// Default for all other endpoints
	return rl.config.Default
}

// hasAuthToken checks if the request has authentication credentials
func (rl *RateLimiter) hasAuthToken(r *http.Request) bool {
	// Check Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") || strings.HasPrefix(auth, "Token ") {
			return true
		}
	}

	// Check for API key in query parameters or headers
	if r.URL.Query().Get("api_key") != "" || r.Header.Get("X-API-Key") != "" {
		return true
	}

	return false
}

// Close closes the rate limiter and cleans up resources
func (rl *RateLimiter) Close() error {
	if rl.redis != nil {
		return rl.redis.Close()
	}
	return nil
}

// max returns the maximum of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// DefaultConfig returns a default rate limiting configuration
func DefaultConfig() Config {
	return Config{
		Enabled:   true,
		Algorithm: SlidingWindow,
		Redis: RedisConfig{
			Enabled:   false,
			Addr:      "localhost:6379",
			Password:  "",
			DB:        0,
			KeyPrefix: "conexus_ratelimit",
		},
		Default: LimitConfig{
			Requests: 100,
			Window:   time.Minute,
		},
		Health: LimitConfig{
			Requests: 1000,
			Window:   time.Minute,
		},
		Webhook: LimitConfig{
			Requests: 10000,
			Window:   time.Minute,
		},
		Auth: LimitConfig{
			Requests: 1000,
			Window:   time.Minute,
		},
		BurstMultiplier: 1.2, // 20% burst capacity
		CleanupInterval: time.Minute * 5,
	}
}
