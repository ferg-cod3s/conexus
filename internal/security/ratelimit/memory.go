package ratelimit

import (
	"sync"
	"time"
)

// InMemoryLimiter provides in-memory rate limiting as a fallback when Redis is unavailable
type InMemoryLimiter struct {
	mu             sync.RWMutex
	requests       map[string][]int64 // key -> timestamps
	cleanupTicker  *time.Ticker
	cleanupStop    chan struct{}
	cleanupRunning bool
}

// NewInMemoryLimiter creates a new in-memory rate limiter
func NewInMemoryLimiter(cleanupInterval time.Duration) *InMemoryLimiter {
	limiter := &InMemoryLimiter{
		requests:    make(map[string][]int64),
		cleanupStop: make(chan struct{}),
	}

	if cleanupInterval > 0 {
		limiter.startCleanup(cleanupInterval)
	}

	return limiter
}

// AllowSlidingWindow checks if a request should be allowed using sliding window algorithm
func (iml *InMemoryLimiter) AllowSlidingWindow(key string, limitConfig LimitConfig, now, windowStart int64) (*Result, error) {
	iml.mu.Lock()
	defer iml.mu.Unlock()

	// Get existing timestamps for this key
	timestamps := iml.requests[key]
	if timestamps == nil {
		timestamps = make([]int64, 0)
	}

	// Remove old timestamps outside the window
	validTimestamps := make([]int64, 0, len(timestamps))
	for _, ts := range timestamps {
		if ts > windowStart {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	// Check if we can allow this request
	allowed := len(validTimestamps) < limitConfig.Requests

	if allowed {
		// Add current timestamp
		validTimestamps = append(validTimestamps, now)
	}

	// Update the map
	if len(validTimestamps) > 0 {
		iml.requests[key] = validTimestamps
	} else {
		delete(iml.requests, key)
	}

	var retryAfter time.Duration
	if !allowed {
		if len(validTimestamps) > 0 {
			// Calculate retry-after based on oldest timestamp
			oldest := validTimestamps[0]
			retryAfter = time.Duration(windowStart-oldest) * time.Millisecond
			if retryAfter < 0 {
				retryAfter = limitConfig.Window
			}
		} else {
			retryAfter = limitConfig.Window
		}
	}

	return &Result{
		Allowed:      allowed,
		Remaining:    max(0, int64(limitConfig.Requests)-int64(len(validTimestamps))),
		RetryAfter:   retryAfter,
		ResetTime:    time.UnixMilli(now + limitConfig.Window.Milliseconds()),
		CurrentCount: int64(len(validTimestamps)),
		Limit:        int64(limitConfig.Requests),
	}, nil
}

// AllowTokenBucket checks if a request should be allowed using token bucket algorithm
func (iml *InMemoryLimiter) AllowTokenBucket(key string, rate float64, burst int, now time.Time) (*Result, error) {
	iml.mu.Lock()
	defer iml.mu.Unlock()

	// Get existing bucket state
	bucket, exists := iml.requests[key]
	var tokens float64
	var lastUpdate time.Time

	if exists && len(bucket) >= 2 {
		// bucket[0] = tokens (stored as milliseconds for simplicity)
		// bucket[1] = last_update timestamp
		tokens = float64(bucket[0]) / 1000.0 // convert back from milliunits
		lastUpdate = time.UnixMilli(bucket[1])
	} else {
		// Initialize new bucket
		tokens = float64(burst)
		lastUpdate = now
	}

	// Calculate tokens to add since last update
	elapsed := now.Sub(lastUpdate)
	newTokens := elapsed.Seconds() * rate
	tokens = min(float64(burst), tokens+newTokens)

	// Check if we can allow this request
	allowed := tokens >= 1.0

	if allowed {
		tokens -= 1.0
	}

	// Update bucket state
	if allowed || tokens < float64(burst) {
		// Store tokens as milliunits to avoid floating point precision issues
		tokenMilliunits := int64(tokens * 1000)
		iml.requests[key] = []int64{tokenMilliunits, now.UnixMilli()}
	}

	retryAfterSeconds := 0.0
	if !allowed {
		retryAfterSeconds = (1.0 - tokens) / rate
	}

	return &Result{
		Allowed:      allowed,
		Remaining:    int64(tokens),
		RetryAfter:   time.Duration(retryAfterSeconds*1000) * time.Millisecond,
		ResetTime:    now.Add(time.Duration(float64(burst)/rate) * time.Second),
		CurrentCount: int64(float64(burst) - tokens),
		Limit:        int64(burst),
	}, nil
}

// startCleanup starts the background cleanup goroutine
func (iml *InMemoryLimiter) startCleanup(interval time.Duration) {
	iml.cleanupRunning = true
	iml.cleanupTicker = time.NewTicker(interval)

	go func() {
		defer func() {
			iml.cleanupRunning = false
		}()

		for {
			select {
			case <-iml.cleanupTicker.C:
				iml.cleanup()
			case <-iml.cleanupStop:
				return
			}
		}
	}()
}

// cleanup removes expired entries from memory
func (iml *InMemoryLimiter) cleanup() {
	iml.mu.Lock()
	defer iml.mu.Unlock()

	now := time.Now().UnixMilli()
	expiredKeys := make([]string, 0)

	// Find keys with only expired timestamps (older than 1 hour)
	for key, timestamps := range iml.requests {
		if len(timestamps) == 0 {
			expiredKeys = append(expiredKeys, key)
			continue
		}

		// For sliding window: check if all timestamps are expired
		allExpired := true
		for _, ts := range timestamps {
			if ts > now-(time.Hour.Milliseconds()) {
				allExpired = false
				break
			}
		}

		if allExpired {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired keys
	for _, key := range expiredKeys {
		delete(iml.requests, key)
	}
}

// Stop stops the cleanup goroutine
func (iml *InMemoryLimiter) Stop() {
	if iml.cleanupRunning {
		iml.cleanupStop <- struct{}{}
		if iml.cleanupTicker != nil {
			iml.cleanupTicker.Stop()
		}
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
