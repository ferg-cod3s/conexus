package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// TraceIDKey is the context key for trace IDs.
	TraceIDKey ContextKey = "trace_id"
	// RequestIDKey is the context key for request IDs.
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user IDs.
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user emails.
	UserEmailKey ContextKey = "user_email"
	// SessionIDKey is the context key for session IDs.
	SessionIDKey ContextKey = "session_id"
	// OrganizationKey is the context key for organization.
	OrganizationKey ContextKey = "organization"
	// ToolNameKey is the context key for tool names.
	ToolNameKey ContextKey = "tool_name"
	// ToolVersionKey is the context key for tool versions.
	ToolVersionKey ContextKey = "tool_version"
)

// Logger wraps slog.Logger with additional context-aware methods.
type Logger struct {
	logger *slog.Logger
}

// LoggerConfig configures the structured logger.
type LoggerConfig struct {
	// Level is the minimum log level (debug, info, warn, error)
	Level string
	// Format is the log format (json, text)
	Format string
	// Output is the output destination (defaults to os.Stdout)
	Output io.Writer
	// AddSource adds source file/line to log entries
	AddSource bool
	// SentryEnabled enables Sentry integration for logs
	SentryEnabled bool
}

// DefaultLoggerConfig returns a default logger configuration.
func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:         "info",
		Format:        "json",
		Output:        os.Stdout,
		AddSource:     true,
		SentryEnabled: false,
	}
}

// sentryHandler is a slog.Handler that sends logs to Sentry.
type sentryHandler struct {
	next slog.Handler
}

func (h *sentryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *sentryHandler) Handle(ctx context.Context, r slog.Record) error {
	// Send to Sentry for error and warn levels
	if r.Level >= slog.LevelWarn {
		var attrs []slog.Attr
		r.Attrs(func(attr slog.Attr) bool {
			attrs = append(attrs, attr)
			return true
		})

		// Convert slog attributes to Sentry context
		sentryCtx := make(map[string]interface{})
		for _, attr := range attrs {
			sentryCtx[attr.Key] = attr.Value.Any()
		}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetContext("log", sentryCtx)
			scope.SetTag("logger", "slog")
			scope.SetTag("level", r.Level.String())

			// Capture as message with context for error and warn logs
			sentry.CaptureMessage(r.Message)
		})
	}

	return h.next.Handle(ctx, r)
}

func (h *sentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &sentryHandler{next: h.next.WithAttrs(attrs)}
}

func (h *sentryHandler) WithGroup(name string) slog.Handler {
	return &sentryHandler{next: h.next.WithGroup(name)}
}

// NewLogger creates a new structured logger.
func NewLogger(cfg LoggerConfig) *Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	if cfg.Format == "text" {
		handler = slog.NewTextHandler(cfg.Output, handlerOpts)
	} else {
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	}

	// Wrap with Sentry handler if enabled
	if cfg.SentryEnabled {
		handler = &sentryHandler{next: handler}
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// WithContext extracts context values and adds them to the logger.
func (l *Logger) WithContext(ctx context.Context) *slog.Logger {
	logger := l.logger

	// Add trace ID if present
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		logger = logger.With("trace_id", traceID)
	}

	// Add request ID if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		logger = logger.With("request_id", requestID)
	}

	// Add user ID if present
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		logger = logger.With("user_id", userID)
	}

	// Add user email if present
	if userEmail, ok := ctx.Value(UserEmailKey).(string); ok && userEmail != "" {
		logger = logger.With("user_email", userEmail)
	}

	// Add session ID if present
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok && sessionID != "" {
		logger = logger.With("session_id", sessionID)
	}

	// Add organization if present
	if organization, ok := ctx.Value(OrganizationKey).(string); ok && organization != "" {
		logger = logger.With("organization", organization)
	}

	// Add tool name if present
	if toolName, ok := ctx.Value(ToolNameKey).(string); ok && toolName != "" {
		logger = logger.With("tool_name", toolName)
	}

	// Add tool version if present
	if toolVersion, ok := ctx.Value(ToolVersionKey).(string); ok && toolVersion != "" {
		logger = logger.With("tool_version", toolVersion)
	}

	return logger
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// DebugContext logs a debug message with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Debug(msg, args...)
}

// InfoContext logs an info message with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Info(msg, args...)
}

// WarnContext logs a warning message with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Warn(msg, args...)
}

// ErrorContext logs an error message with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.WithContext(ctx).Error(msg, args...)
}

// With returns a logger with additional attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// WithGroup returns a logger with a named group.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		logger: l.logger.WithGroup(name),
	}
}

// LogMCPRequest logs an MCP request with standard fields.
func (l *Logger) LogMCPRequest(ctx context.Context, method string, params any, duration time.Duration) {
	l.WithContext(ctx).Info("mcp_request",
		"method", method,
		"duration_ms", duration.Milliseconds(),
		"params", params,
	)
}

// LogMCPResponse logs an MCP response with standard fields.
func (l *Logger) LogMCPResponse(ctx context.Context, method string, success bool, duration time.Duration) {
	l.WithContext(ctx).Info("mcp_response",
		"method", method,
		"success", success,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogMCPError logs an MCP error with standard fields.
func (l *Logger) LogMCPError(ctx context.Context, method string, err error, duration time.Duration) {
	l.WithContext(ctx).Error("mcp_error",
		"method", method,
		"error", err.Error(),
		"duration_ms", duration.Milliseconds(),
	)
}

// LogIndexerOperation logs an indexer operation with standard fields.
func (l *Logger) LogIndexerOperation(ctx context.Context, operation string, path string, duration time.Duration) {
	l.WithContext(ctx).Info("indexer_operation",
		"operation", operation,
		"path", path,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogEmbedding logs an embedding operation with standard fields.
func (l *Logger) LogEmbedding(ctx context.Context, provider string, textLength int, duration time.Duration) {
	l.WithContext(ctx).Info("embedding_request",
		"provider", provider,
		"text_length", textLength,
		"duration_ms", duration.Milliseconds(),
	)
}

// LogVectorSearch logs a vector search operation with standard fields.
func (l *Logger) LogVectorSearch(ctx context.Context, searchType string, resultCount int, duration time.Duration) {
	l.WithContext(ctx).Info("vector_search",
		"search_type", searchType,
		"result_count", resultCount,
		"duration_ms", duration.Milliseconds(),
	)
}

// Underlying returns the underlying slog.Logger.
func (l *Logger) Underlying() *slog.Logger {
	return l.logger
}
