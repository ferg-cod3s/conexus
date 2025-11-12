package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name   string
		config LoggerConfig
	}{
		{
			name: "json format with debug level",
			config: LoggerConfig{
				Level:     "debug",
				Format:    "json",
				AddSource: true,
			},
		},
		{
			name: "text format with info level",
			config: LoggerConfig{
				Level:     "info",
				Format:    "text",
				AddSource: false,
			},
		},
		{
			name: "default values",
			config: LoggerConfig{
				Level:  "info",
				Format: "text",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.config.Output = &buf

			logger := NewLogger(tt.config)
			require.NotNil(t, logger)
			assert.NotNil(t, logger.logger)
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name      string
		logFunc   func(*Logger, string)
		logLevel  string
		message   string
		shouldLog bool
	}{
		{
			name:      "debug message at debug level",
			logFunc:   func(l *Logger, msg string) { l.Debug(msg) },
			logLevel:  "debug",
			message:   "debug message",
			shouldLog: true,
		},
		{
			name:      "debug message at info level",
			logFunc:   func(l *Logger, msg string) { l.Debug(msg) },
			logLevel:  "info",
			message:   "debug message",
			shouldLog: false,
		},
		{
			name:      "info message at info level",
			logFunc:   func(l *Logger, msg string) { l.Info(msg) },
			logLevel:  "info",
			message:   "info message",
			shouldLog: true,
		},
		{
			name:      "warn message at error level",
			logFunc:   func(l *Logger, msg string) { l.Warn(msg) },
			logLevel:  "error",
			message:   "warn message",
			shouldLog: false,
		},
		{
			name:      "error message at error level",
			logFunc:   func(l *Logger, msg string) { l.Error(msg) },
			logLevel:  "error",
			message:   "error message",
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(LoggerConfig{
				Level:  tt.logLevel,
				Format: "json",
				Output: &buf,
			})

			tt.logFunc(logger, tt.message)

			output := buf.String()
			if tt.shouldLog {
				assert.Contains(t, output, tt.message)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	logger.Info("test message",
		"field1", "value1",
		"field2", 42,
		"field3", true,
	)

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "field1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "field2")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "field3")
	assert.Contains(t, output, "true")
}

func TestLoggerWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, TraceIDKey, "trace-123")
	ctx = context.WithValue(ctx, RequestIDKey, "req-456")
	ctx = context.WithValue(ctx, UserIDKey, "user-789")

	logger.InfoContext(ctx, "context test")

	output := buf.String()
	assert.Contains(t, output, "context test")
	assert.Contains(t, output, "trace-123")
	assert.Contains(t, output, "req-456")
	assert.Contains(t, output, "user-789")
}

func TestLogMCPRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	logger.LogMCPRequest(context.Background(), "tools/list", map[string]interface{}{
		"limit": 10,
	}, 100*time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "mcp_request", logEntry["msg"])
	assert.Equal(t, "tools/list", logEntry["method"])
	assert.NotNil(t, logEntry["params"])
}

func TestLogMCPResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	ctx := context.Background()
	logger.LogMCPResponse(ctx, "tools/list", true, 50*time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "mcp_response", logEntry["msg"])
	assert.Equal(t, "tools/list", logEntry["method"])
	assert.Equal(t, true, logEntry["success"])
	assert.NotNil(t, logEntry["duration_ms"])
}

func TestLogMCPError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "error",
		Format: "json",
		Output: &buf,
	})

	ctx := context.Background()
	testErr := assert.AnError
	logger.LogMCPError(ctx, "tools/call", testErr, 100*time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "mcp_error", logEntry["msg"])
	assert.Equal(t, "tools/call", logEntry["method"])
	assert.NotNil(t, logEntry["error"])
	assert.Contains(t, output, testErr.Error())
}

func TestLogIndexerOperation(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	ctx := context.Background()
	logger.LogIndexerOperation(ctx, "index", "/path/to/repo", 5*time.Second)

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "indexer_operation", logEntry["msg"])
	assert.Equal(t, "index", logEntry["operation"])
	assert.Equal(t, "/path/to/repo", logEntry["path"])
	assert.NotNil(t, logEntry["duration_ms"])
}

func TestLogEmbedding(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "debug",
		Format: "json",
		Output: &buf,
	})

	ctx := context.Background()
	logger.LogEmbedding(ctx, "text-embedding-3-small", 512, 30*time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "embedding_request", logEntry["msg"])
	assert.Equal(t, "text-embedding-3-small", logEntry["provider"])
	assert.Equal(t, float64(512), logEntry["text_length"])
}

func TestLogVectorSearch(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	ctx := context.Background()
	logger.LogVectorSearch(ctx, "semantic", 10, 25*time.Millisecond)

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "vector_search", logEntry["msg"])
	assert.Equal(t, "semantic", logEntry["search_type"])
	assert.Equal(t, float64(10), logEntry["result_count"])
}

func TestLoggerTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "text",
		Output: &buf,
	})

	logger.Info("text format test", "key", "value")

	output := buf.String()
	assert.Contains(t, output, "text format test")
	assert.Contains(t, output, "key=value")
}

func TestLoggerInvalidLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "invalid",
		Format: "json",
		Output: &buf,
	})

	// Should default to INFO level
	logger.Debug("debug message")
	assert.Empty(t, buf.String())

	buf.Reset()
	logger.Info("info message")
	assert.NotEmpty(t, buf.String())
}

func TestLoggerConcurrency(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	done := make(chan bool)
	iterations := 100

	// Concurrent logging
	for i := 0; i < 3; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				logger.Info("concurrent log", "goroutine", id, "iteration", j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify logs were written
	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 3*iterations, len(lines))
}

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	childLogger := logger.With("service", "test-service", "version", "1.0.0")
	childLogger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test-service")
	assert.Contains(t, output, "1.0.0")
	assert.Contains(t, output, "test message")
}

func TestLoggerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(LoggerConfig{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	groupLogger := logger.WithGroup("request")
	groupLogger.Info("request received", "method", "GET", "path", "/api/v1")

	output := buf.String()
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	require.NoError(t, err)

	// In slog, WithGroup creates nested structure
	assert.NotNil(t, logEntry["request"])
}
