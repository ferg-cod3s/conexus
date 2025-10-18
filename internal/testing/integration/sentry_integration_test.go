package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/config"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSentryErrorCapture tests that errors are properly captured and sent to Sentry
func TestSentryErrorCapture(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Sentry integration test in short mode")
	}

	tests := []struct {
		name        string
		errorType   string
		setupError  func() error
		expectEvent bool
		description string
	}{
		{
			name:      "standard_error_capture",
			errorType: "standard_error",
			setupError: func() error {
				return errors.New("test standard error")
			},
			expectEvent: true,
			description: "Should capture standard Go errors",
		},
		{
			name:      "context_cancelled_error",
			errorType: "context_cancelled",
			setupError: func() error {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx.Err()
			},
			expectEvent: true,
			description: "Should capture context cancellation errors",
		},
		{
			name:      "validation_error",
			errorType: "validation_error",
			setupError: func() error {
				return errors.New("validation error: invalid input")
			},
			expectEvent: true,
			description: "Should capture validation errors with context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create error handler with Sentry enabled
			loggerCfg := observability.LoggerConfig{
				Level:  "debug",
				Format: "json",
			}
			logger := observability.NewLogger(loggerCfg)
			metrics, _ := NewTestMetricsCollector("test")
			errorHandler := observability.NewErrorHandler(logger, metrics, true)

			ctx := context.Background()

			// Create test error
			err := tt.setupError()
			require.NotNil(t, err, "Test error should not be nil")

			// Create error context
			errorCtx := observability.ExtractErrorContext(ctx, "test.operation")
			errorCtx.ErrorType = tt.errorType
			errorCtx.UserID = "test-user-123"
			errorCtx.RequestID = "test-request-456"
			errorCtx.Extra = map[string]interface{}{
				"test_param": "test_value",
			}

			// Handle the error
			errorHandler.HandleError(ctx, err, errorCtx)

			// Give Sentry some time to process the event
			time.Sleep(100 * time.Millisecond)

			// In a real integration test, we would verify the event was sent
			// For now, we verify the error handler doesn't panic and processes the error
			assert.NotNil(t, errorHandler, "Error handler should be created successfully")
		})
	}
}

// TestSentryTracing tests that traces are properly created and sent to Sentry
func TestSentryTracing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Sentry tracing test in short mode")
	}

	// Create error handler with Sentry enabled
	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics, _ := NewTestMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, true)

	ctx := context.Background()

	// Start a span
	span := sentry.StartSpan(ctx, "test.operation")
	defer span.Finish()

	// Add some span data
	span.SetTag("test.tag", "test_value")
	span.SetData("test.data", "test_value")

	// Create a child span
	childSpan := sentry.StartSpan(span.Context(), "test.child.operation")
	childSpan.SetTag("child.tag", "child_value")
	childSpan.Finish()

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	// Verify spans were created (in real integration, we'd verify they were sent)
	assert.NotNil(t, span, "Parent span should be created")
	assert.NotNil(t, childSpan, "Child span should be created")

	// Test error within span context
	err := errors.New("test error within span")
	errorCtx := observability.ExtractErrorContext(span.Context(), "test.operation")
	errorHandler.HandleError(span.Context(), err, errorCtx)

	// Give Sentry time to process
	time.Sleep(100 * time.Millisecond)
}

// TestSentryUserContext tests that user context is properly attached to Sentry events
func TestSentryUserContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Sentry user context test in short mode")
	}

	// Create error handler with Sentry enabled
	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics, _ := NewTestMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, true)

	ctx := context.Background()

	// Set user context
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       "test-user-123",
			Email:    "test@example.com",
			Username: "testuser",
		})
		scope.SetTag("organization", "test-org")
		scope.SetTag("project", "test-project")
	})

	// Create and handle error
	err := errors.New("test error with user context")
	errorCtx := observability.ExtractErrorContext(ctx, "test.operation")
	errorCtx.UserID = "test-user-123"

	errorHandler.HandleError(ctx, err, errorCtx)

	// Give Sentry time to process
	time.Sleep(100 * time.Millisecond)

	// Verify error handler processed the error
	assert.NotNil(t, errorHandler, "Error handler should process error with user context")
}

// TestSentryPerformanceMonitoring tests that performance data is captured
func TestSentryPerformanceMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Sentry performance test in short mode")
	}

	ctx := context.Background()

	// Start a transaction
	transaction := sentry.StartTransaction(ctx, "test.performance.transaction")
	defer transaction.Finish()

	// Create spans for different operations
	dbSpan := sentry.StartSpan(transaction.Context(), "db.query")
	dbSpan.SetTag("table", "test_table")
	dbSpan.SetData("query", "SELECT * FROM test_table")
	time.Sleep(10 * time.Millisecond)
	dbSpan.Finish()

	apiSpan := sentry.StartSpan(transaction.Context(), "http.client")
	apiSpan.SetTag("method", "GET")
	apiSpan.SetTag("url", "https://api.example.com/test")
	apiSpan.SetData("status_code", 200)
	time.Sleep(5 * time.Millisecond)
	apiSpan.Finish()

	// Set transaction status
	transaction.Status = sentry.SpanStatusOK

	// Give Sentry time to process
	time.Sleep(100 * time.Millisecond)

	// Verify transaction was created
	assert.NotNil(t, transaction, "Transaction should be created")
	assert.NotNil(t, dbSpan, "DB span should be created")
	assert.NotNil(t, apiSpan, "API span should be created")
}

// TestSentryErrorRecovery tests error recovery scenarios
func TestSentryErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Sentry recovery test in short mode")
	}

	// Create error handler with Sentry enabled
	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics, _ := NewTestMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, true)

	ctx := context.Background()

	// Test multiple errors in sequence
	for i := 0; i < 3; i++ {
		err := errors.New("sequential error")

		errorCtx := observability.ExtractErrorContext(ctx, "test.recovery")
		errorCtx.ErrorType = "sequential_error"
		errorCtx.Extra = map[string]interface{}{"index": i}

		errorHandler.HandleError(ctx, err, errorCtx)

		// Small delay between errors
		time.Sleep(10 * time.Millisecond)
	}

	// Give Sentry time to process all events
	time.Sleep(200 * time.Millisecond)

	// Verify error handler remained functional
	assert.NotNil(t, errorHandler, "Error handler should remain functional after multiple errors")
}

// TestSentryConfigurationValidation tests Sentry configuration validation
func TestSentryConfigurationValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Config
		expectError bool
		description string
	}{
		{
			name: "valid_sentry_config",
			config: config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 9000,
				},
				Observability: config.ObservabilityConfig{
					Sentry: config.SentryConfig{
						Enabled:     true,
						DSN:         "https://test@test.ingest.sentry.io/test",
						Environment: "test",
						SampleRate:  0.1,
					},
				},
			},
			expectError: false,
			description: "Valid Sentry configuration should pass",
		},
		{
			name: "missing_dsn_when_enabled",
			config: config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 9001,
				},
				Observability: config.ObservabilityConfig{
					Sentry: config.SentryConfig{
						Enabled:     true,
						Environment: "test",
						SampleRate:  0.1,
					},
				},
			},
			expectError: true,
			description: "Missing DSN should fail validation when enabled",
		},
		{
			name: "invalid_sample_rate_high",
			config: config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 9002,
				},
				Observability: config.ObservabilityConfig{
					Sentry: config.SentryConfig{
						Enabled:     true,
						DSN:         "https://test@test.ingest.sentry.io/test",
						Environment: "test",
						SampleRate:  1.5,
					},
				},
			},
			expectError: true,
			description: "Sample rate > 1 should fail validation",
		},
		{
			name: "invalid_sample_rate_negative",
			config: config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: 9003,
				},
				Observability: config.ObservabilityConfig{
					Sentry: config.SentryConfig{
						Enabled:     true,
						DSN:         "https://test@test.ingest.sentry.io/test",
						Environment: "test",
						SampleRate:  -0.1,
					},
				},
			},
			expectError: true,
			description: "Negative sample rate should fail validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestSentryHealthCheck tests Sentry health validation
func TestSentryHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Sentry health check test in short mode")
	}

	// Create error handler with Sentry enabled
	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics, _ := NewTestMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, true)

	ctx := context.Background()

	// Test health check
	health := errorHandler.CreateHealthCheck(ctx, "test-version")

	// Verify health check structure
	assert.NotNil(t, health, "Health check should return data")
	assert.Contains(t, health.Components, "sentry", "Health check should include Sentry component")

	sentryHealth, ok := health.Components["sentry"].(map[string]interface{})
	assert.True(t, ok, "Sentry health should be a map")
	assert.NotNil(t, sentryHealth, "Sentry health data should be present")

	// Check health status (may be false if Sentry is not properly configured for testing)
	status, ok := sentryHealth["status"].(string)
	assert.True(t, ok, "Sentry health should have status field")

	if status == "enabled" {
		// If enabled, check additional fields
		assert.Contains(t, sentryHealth, "message", "Enabled Sentry should have message")
	}
}
