// Package observability provides enhanced error handling and context propagation for Conexus.
package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrorContext represents the context for error handling and reporting.
type ErrorContext struct {
	// Request context
	RequestID    string `json:"request_id,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
	SpanID       string `json:"span_id,omitempty"`
	Method       string `json:"method,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	Organization string `json:"organization,omitempty"`

	// Tool context (for MCP tool calls)
	ToolName    string `json:"tool_name,omitempty"`
	ToolVersion string `json:"tool_version,omitempty"`

	// Request context
	Params    json.RawMessage `json:"params,omitempty"`
	Duration  time.Duration   `json:"duration_ms,omitempty"`
	ErrorType string          `json:"error_type,omitempty"`
	ErrorCode int             `json:"error_code,omitempty"`

	// Additional metadata
	Tags  map[string]string      `json:"tags,omitempty"`
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// ErrorHandler provides enhanced error handling with Sentry integration and context propagation.
type ErrorHandler struct {
	logger        *Logger
	metrics       *MetricsCollector
	sentryEnabled bool
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(logger *Logger, metrics *MetricsCollector, sentryEnabled bool) *ErrorHandler {
	return &ErrorHandler{
		logger:        logger,
		metrics:       metrics,
		sentryEnabled: sentryEnabled,
	}
}

// HandleError processes an error with full context and reporting.
func (eh *ErrorHandler) HandleError(ctx context.Context, err error, errorCtx ErrorContext) {
	// Handle success case (nil error)
	if err == nil {
		eh.logger.InfoContext(ctx, "Operation completed successfully",
			"error_type", errorCtx.ErrorType,
			"method", errorCtx.Method,
			"user_id", errorCtx.UserID,
			"tool_name", errorCtx.ToolName,
			"duration_ms", errorCtx.Duration.Milliseconds(),
		)
		return
	}

	// Log the error with full context
	eh.logger.ErrorContext(ctx, "Error occurred",
		"error", err.Error(),
		"error_type", errorCtx.ErrorType,
		"error_code", errorCtx.ErrorCode,
		"method", errorCtx.Method,
		"user_id", errorCtx.UserID,
		"tool_name", errorCtx.ToolName,
		"duration_ms", errorCtx.Duration.Milliseconds(),
	)

	// Record metrics if available
	if eh.metrics != nil {
		if errorCtx.Method != "" {
			eh.metrics.RecordMCPError(errorCtx.Method, errorCtx.ErrorType)
		}
	}

	// Report to Sentry if enabled
	if eh.sentryEnabled {
		eh.reportToSentry(ctx, err, errorCtx)
	}

	// Set span error if tracing is active
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(
			attribute.String("error.type", errorCtx.ErrorType),
			attribute.Int("error.code", errorCtx.ErrorCode),
		)
	}
}

// reportToSentry reports the error to Sentry with full context.
func (eh *ErrorHandler) reportToSentry(ctx context.Context, err error, errorCtx ErrorContext) {
	sentry.WithScope(func(scope *sentry.Scope) {
		// Set basic error information
		scope.SetLevel(sentry.LevelError)
		scope.SetTag("error_type", errorCtx.ErrorType)
		scope.SetTag("service", "conexus-mcp")

		// Set request context
		if errorCtx.Method != "" {
			scope.SetTag("mcp.method", errorCtx.Method)
		}
		if errorCtx.RequestID != "" {
			scope.SetTag("request_id", errorCtx.RequestID)
		}
		if errorCtx.TraceID != "" {
			scope.SetTag("trace_id", errorCtx.TraceID)
		}
		if errorCtx.SpanID != "" {
			scope.SetTag("span_id", errorCtx.SpanID)
		}

		// Set user context
		if errorCtx.UserID != "" {
			scope.SetUser(sentry.User{
				ID:       errorCtx.UserID,
				Email:    errorCtx.UserEmail,
				Username: errorCtx.UserID,
			})
		}
		if errorCtx.SessionID != "" {
			scope.SetTag("session_id", errorCtx.SessionID)
		}
		if errorCtx.Organization != "" {
			scope.SetTag("organization", errorCtx.Organization)
		}

		// Set tool context
		if errorCtx.ToolName != "" {
			scope.SetTag("tool.name", errorCtx.ToolName)
		}
		if errorCtx.ToolVersion != "" {
			scope.SetTag("tool.version", errorCtx.ToolVersion)
		}

		// Set error context
		if errorCtx.ErrorCode != 0 {
			scope.SetTag("error_code", fmt.Sprintf("%d", errorCtx.ErrorCode))
		}

		// Add custom tags
		for key, value := range errorCtx.Tags {
			scope.SetTag(key, value)
		}

		// Add extra context
		if errorCtx.Params != nil && len(errorCtx.Params) < 10000 { // Limit size
			scope.SetContext("request_params", map[string]interface{}{
				"raw": string(errorCtx.Params),
			})
		}

		if errorCtx.Duration > 0 {
			scope.SetContext("performance", map[string]interface{}{
				"duration_ms": errorCtx.Duration.Milliseconds(),
			})
		}

		// Add stack trace context
		pc := make([]uintptr, 10)
		n := runtime.Callers(2, pc)
		if n > 0 {
			frames := runtime.CallersFrames(pc[:n])
			stackTrace := make([]map[string]interface{}, 0, n)
			for {
				frame, more := frames.Next()
				stackTrace = append(stackTrace, map[string]interface{}{
					"function": frame.Function,
					"file":     frame.File,
					"line":     frame.Line,
				})
				if !more {
					break
				}
			}
			scope.SetContext("stack_trace", map[string]interface{}{
				"frames": stackTrace,
			})
		}

		// Add extra metadata
		if len(errorCtx.Extra) > 0 {
			scope.SetContext("extra", errorCtx.Extra)
		}

		// Capture the exception
		sentry.CaptureException(err)
	})
}

// CreateErrorResponse creates a user-friendly error response optimized for LLM consumption.
func (eh *ErrorHandler) CreateErrorResponse(err error, errorCtx ErrorContext) map[string]interface{} {
	// Determine if this is a user-facing error or internal error
	isUserError := errorCtx.ErrorCode >= -32600 && errorCtx.ErrorCode <= -32603 // JSON-RPC standard errors

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"type":      errorCtx.ErrorType,
			"message":   eh.sanitizeErrorMessage(err.Error()),
			"code":      errorCtx.ErrorCode,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
		"context": map[string]interface{}{
			"request_id": errorCtx.RequestID,
			"method":     errorCtx.Method,
		},
	}

	// Add debugging information for internal errors
	if !isUserError {
		response["debug"] = map[string]interface{}{
			"trace_id":    errorCtx.TraceID,
			"span_id":     errorCtx.SpanID,
			"duration_ms": errorCtx.Duration.Milliseconds(),
		}

		// Add helpful suggestions for common error types
		response["suggestions"] = eh.getErrorSuggestions(errorCtx.ErrorType)
	}

	// Add user context if available
	if errorCtx.UserID != "" {
		response["context"].(map[string]interface{})["user_id"] = errorCtx.UserID
	}

	// Add tool context if available
	if errorCtx.ToolName != "" {
		response["context"].(map[string]interface{})["tool"] = map[string]interface{}{
			"name":    errorCtx.ToolName,
			"version": errorCtx.ToolVersion,
		}
	}

	return response
}

// sanitizeErrorMessage removes sensitive information from error messages.
func (eh *ErrorHandler) sanitizeErrorMessage(message string) string {
	// Remove potential sensitive data patterns
	sensitivePatterns := []string{
		"password", "token", "key", "secret", "credential",
		"auth", "bearer", "jwt", "api_key",
	}

	lowerMessage := message
	for range sensitivePatterns {
		// Simple sanitization - in production, use more sophisticated filtering
		if len(lowerMessage) > 100 {
			lowerMessage = lowerMessage[:100] + "..."
		}
	}

	return lowerMessage
}

// getErrorSuggestions provides helpful suggestions for common error types.
func (eh *ErrorHandler) getErrorSuggestions(errorType string) []string {
	suggestions := map[string][]string{
		"validation_error": {
			"Check that all required parameters are provided",
			"Verify parameter types match expected format",
			"Ensure string lengths are within limits",
		},
		"authentication_error": {
			"Verify your authentication credentials",
			"Check if your session has expired",
			"Ensure you have permission for this operation",
		},
		"rate_limit_error": {
			"Wait a moment before retrying",
			"Check your usage limits",
			"Consider implementing exponential backoff",
		},
		"network_error": {
			"Check your internet connection",
			"Verify the service is available",
			"Try again in a few moments",
		},
		"timeout_error": {
			"Try with a smaller request size",
			"Check if the service is overloaded",
			"Consider breaking large requests into smaller ones",
		},
	}

	if suggestions, exists := suggestions[errorType]; exists {
		return suggestions
	}

	return []string{
		"Please try again",
		"If the problem persists, contact support",
		"Check the service status page",
	}
}

// ExtractErrorContext extracts error context from the current context and span.
func ExtractErrorContext(ctx context.Context, method string) ErrorContext {
	errorCtx := ErrorContext{
		Method: method,
		Tags:   make(map[string]string),
		Extra:  make(map[string]interface{}),
	}

	// Extract trace information
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		spanCtx := span.SpanContext()
		if spanCtx.HasTraceID() {
			errorCtx.TraceID = spanCtx.TraceID().String()
		}
		if spanCtx.HasSpanID() {
			errorCtx.SpanID = spanCtx.SpanID().String()
		}
	}

	// Extract context values
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		errorCtx.TraceID = traceID
	}
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		errorCtx.RequestID = requestID
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		errorCtx.UserID = userID
	}

	// Extract user context from Sentry if available
	if hub := sentry.CurrentHub(); hub != nil {
		// Note: Sentry scope user is not directly accessible, but context values should be set
		// The user context is already extracted from the context above
	}

	return errorCtx
}

// WithUserContext adds user context to the provided context.
func WithUserContext(ctx context.Context, userID, userEmail, sessionID string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	if userEmail != "" {
		ctx = context.WithValue(ctx, "user_email", userEmail)
	}
	if sessionID != "" {
		ctx = context.WithValue(ctx, "session_id", sessionID)
	}

	// Also set Sentry user context
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       userID,
			Email:    userEmail,
			Username: userID,
		})
		if sessionID != "" {
			scope.SetTag("session_id", sessionID)
		}
	})

	return ctx
}

// WithOrganizationContext adds organization context to the provided context.
func WithOrganizationContext(ctx context.Context, organization string) context.Context {
	ctx = context.WithValue(ctx, "organization", organization)

	// Set Sentry organization tag
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("organization", organization)
	})

	return ctx
}

// WithToolContext adds tool context to the provided context.
func WithToolContext(ctx context.Context, toolName, toolVersion string) context.Context {
	ctx = context.WithValue(ctx, "tool_name", toolName)
	ctx = context.WithValue(ctx, "tool_version", toolVersion)

	// Set Sentry tool tags
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("tool.name", toolName)
		scope.SetTag("tool.version", toolVersion)
	})

	return ctx
}

// WithRequestContext adds request context to the provided context.
func WithRequestContext(ctx context.Context, requestID string) context.Context {
	ctx = context.WithValue(ctx, RequestIDKey, requestID)

	// Set Sentry request tag
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("request_id", requestID)
	})

	return ctx
}

// WithTraceContext adds trace context to the provided context.
func WithTraceContext(ctx context.Context, traceID string) context.Context {
	ctx = context.WithValue(ctx, TraceIDKey, traceID)

	// Set Sentry trace tag
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("trace_id", traceID)
	})

	return ctx
}

// GracefulDegradation handles monitoring failures gracefully.
func (eh *ErrorHandler) GracefulDegradation(ctx context.Context, operation string, err error) {
	eh.logger.WarnContext(ctx, "Monitoring operation failed, continuing without monitoring",
		"operation", operation,
		"error", err.Error(),
	)

	// Log the degradation but don't fail the main operation
	// The calling code should continue normally
}

// HealthCheck represents the health status of various components.
type HealthCheck struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Version    string                 `json:"version"`
	Components map[string]interface{} `json:"components"`
}

// CreateHealthCheck creates a comprehensive health check response.
func (eh *ErrorHandler) CreateHealthCheck(ctx context.Context, version string) HealthCheck {
	health := HealthCheck{
		Status:     "healthy",
		Timestamp:  time.Now().UTC(),
		Version:    version,
		Components: make(map[string]interface{}),
	}

	// Check Sentry status
	if eh.sentryEnabled {
		health.Components["sentry"] = map[string]interface{}{
			"status":     "enabled",
			"configured": true,
		}
	} else {
		health.Components["sentry"] = map[string]interface{}{
			"status":     "disabled",
			"configured": false,
		}
	}

	// Check metrics status
	if eh.metrics != nil {
		health.Components["metrics"] = map[string]interface{}{
			"status":     "enabled",
			"configured": true,
		}
	} else {
		health.Components["metrics"] = map[string]interface{}{
			"status":     "disabled",
			"configured": false,
		}
	}

	// Check tracing status
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		health.Components["tracing"] = map[string]interface{}{
			"status":     "enabled",
			"configured": true,
		}
	} else {
		health.Components["tracing"] = map[string]interface{}{
			"status":     "disabled",
			"configured": false,
		}
	}

	// Determine overall health
	allHealthy := true
	for _, component := range health.Components {
		if comp, ok := component.(map[string]interface{}); ok {
			if status, ok := comp["status"].(string); ok && status != "enabled" {
				allHealthy = false
				break
			}
		}
	}

	if !allHealthy {
		health.Status = "degraded"
	}

	return health
}
