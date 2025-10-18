package github

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// RateLimiter manages GitHub API rate limiting
type RateLimiter struct {
	mu              sync.Mutex
	remaining       int
	reset           time.Time
	requestInterval time.Duration
}

// NewRateLimiter creates a new rate limiter
// GitHub's public API allows 60 requests per minute
// Authenticated requests allow 5000 per hour
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		remaining:       60,
		reset:           time.Now().Add(time.Minute),
		requestInterval: time.Second, // Conservative: 1 request per second
	}
}

// Wait blocks until a request can be made
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if rate limit has been reset
	if time.Now().After(r.reset) {
		r.remaining = 5000 // Reset to authenticated limit
		r.reset = time.Now().Add(time.Hour)
	}

	// If no requests remaining, wait until reset
	if r.remaining <= 1 {
		waitDuration := time.Until(r.reset)
		if waitDuration > 0 {
			select {
			case <-time.After(waitDuration):
				r.remaining = 5000
				r.reset = time.Now().Add(time.Hour)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// Decrement and allow minimal interval between requests
	r.remaining--

	// Apply request interval
	select {
	case <-time.After(r.requestInterval):
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// Update updates the rate limiter based on GitHub API response headers
func (r *RateLimiter) Update(remainingStr, resetStr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if remaining, err := strconv.Atoi(remainingStr); err == nil {
		r.remaining = remaining
	}

	if resetUnix, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
		r.reset = time.Unix(resetUnix, 0)
	}
}

// Remaining returns the current remaining request count
func (r *RateLimiter) Remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.remaining
}

// TimeUntilReset returns the time until the rate limit resets
func (r *RateLimiter) TimeUntilReset() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	resetTime := time.Until(r.reset)
	if resetTime < 0 {
		return 0
	}
	return resetTime
}

// String returns a string representation of the rate limiter status
func (r *RateLimiter) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return fmt.Sprintf("RateLimiter{Remaining: %d, ResetAt: %s}", r.remaining, r.reset.Format(time.RFC3339))
}
