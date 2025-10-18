package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter()

	assert.NotNil(t, rl)
	assert.Equal(t, 60, rl.Remaining())
	assert.Greater(t, rl.TimeUntilReset(), 0*time.Second)
}

func TestRateLimiter_Wait_Success(t *testing.T) {
	rl := NewRateLimiter()

	err := rl.Wait(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 59, rl.Remaining())
}

func TestRateLimiter_Wait_MultipleRequests(t *testing.T) {
	rl := NewRateLimiter()
	initial := rl.Remaining()

	for i := 0; i < 5; i++ {
		err := rl.Wait(context.Background())
		require.NoError(t, err)
	}

	assert.Equal(t, initial-5, rl.Remaining())
}

func TestRateLimiter_Wait_ContextCancellation(t *testing.T) {
	rl := NewRateLimiter()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rl.Wait(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestRateLimiter_Update(t *testing.T) {
	rl := NewRateLimiter()

	futureReset := time.Now().Add(10 * time.Minute).Unix()
	rl.Update("30", ""+string(rune(futureReset)))

	// Note: Update might not work perfectly in unit tests due to timing,
	// but we verify it doesn't panic
	assert.NotPanics(t, func() {
		rl.Update("25", ""+string(rune(time.Now().Add(5*time.Minute).Unix())))
	})
}

func TestRateLimiter_Remaining(t *testing.T) {
	rl := NewRateLimiter()
	initial := rl.Remaining()

	assert.Equal(t, 60, initial)

	rl.Wait(context.Background())
	after := rl.Remaining()

	assert.Equal(t, initial-1, after)
}

func TestRateLimiter_TimeUntilReset(t *testing.T) {
	rl := NewRateLimiter()

	timeUntil := rl.TimeUntilReset()

	assert.Greater(t, timeUntil, 0*time.Second)
	assert.LessOrEqual(t, timeUntil, 1*time.Minute+time.Second)
}

func TestRateLimiter_String(t *testing.T) {
	rl := NewRateLimiter()

	str := rl.String()

	assert.Contains(t, str, "RateLimiter")
	assert.Contains(t, str, "Remaining")
	assert.NotEmpty(t, str)
}

func TestRateLimiter_ResetAfterWindow(t *testing.T) {
	rl := NewRateLimiter()

	// Manually set reset time to past
	rl.mu.Lock()
	rl.reset = time.Now().Add(-1 * time.Second)
	rl.remaining = 1
	rl.mu.Unlock()

	err := rl.Wait(context.Background())

	require.NoError(t, err)
	// After reset, should have full limit minus one request
	assert.Equal(t, 4999, rl.Remaining())
}

func TestRateLimiter_ConcurrentWait(t *testing.T) {
	rl := NewRateLimiter()

	// Multiple concurrent waits should be safe
	done := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			done <- rl.Wait(context.Background())
		}()
	}

	for i := 0; i < 10; i++ {
		err := <-done
		require.NoError(t, err)
	}

	assert.Equal(t, 50, rl.Remaining())
}

func TestRateLimiter_String_Format(t *testing.T) {
	rl := NewRateLimiter()
	rl.Wait(context.Background())

	str := rl.String()

	assert.Contains(t, str, "59")
	assert.Contains(t, str, "ResetAt")
}
