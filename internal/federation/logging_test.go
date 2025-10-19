package federation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogger_NewLogger(t *testing.T) {
	ctx := context.Background()
	logger := NewLogger(ctx)

	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
	assert.Equal(t, ctx, logger.ctx)
}

func TestLogger_LogQueryStart(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic
	logger.LogQueryStart("query_123", "test query", 5)
}

func TestLogger_LogQueryEnd_Success(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic
	duration := 100 * time.Millisecond
	logger.LogQueryEnd("query_123", duration, 50, 0, 10, true)
}

func TestLogger_LogQueryEnd_Failure(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic - logs at warn level
	duration := 100 * time.Millisecond
	logger.LogQueryEnd("query_123", duration, 25, 2, 5, false)
}

func TestLogger_LogSourceQuery_Success(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic
	duration := 50 * time.Millisecond
	logger.LogSourceQuery("query_123", "source1", duration, 20, nil)
}

func TestLogger_LogSourceQuery_Error(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic - logs at warn level with error
	duration := 50 * time.Millisecond
	err := fmt.Errorf("connection timeout")
	logger.LogSourceQuery("query_123", "source1", duration, 0, err)
}

func TestLogger_LogMergeStart(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic
	logger.LogMergeStart(3)
}

func TestLogger_LogMergeEnd(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic
	duration := 75 * time.Millisecond
	stats := DeduplicationStats{
		TotalResults:    200,
		DuplicatesFound: 50,
		UniqueResults:   150,
		MergedResults:   160,
	}
	logger.LogMergeEnd(duration, stats)
}

func TestLogger_LogRelationshipDetection(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic
	duration := 25 * time.Millisecond
	logger.LogRelationshipDetection(15, duration)
}

func TestLogger_LogError(t *testing.T) {
	logger := NewLogger(context.Background())

	// Should not panic
	err := fmt.Errorf("unexpected error occurred")
	context := map[string]interface{}{
		"query": "test",
		"source": "source1",
	}
	logger.LogError("query_operation", err, context)
}

func TestLogFields_Structure(t *testing.T) {
	fields := LogFields{
		Operation:       "test_op",
		Source:          "test_source",
		Duration:        100 * time.Millisecond,
		ItemCount:       42,
		Status:          "success",
		Error:           "",
		Query:           "test query",
		SourceCount:     3,
		ErrorCount:      0,
		DuplicateCount:  5,
		UniqueCount:     95,
		RelationshipCount: 12,
		QueryID:         "query_456",
	}

	assert.Equal(t, "test_op", fields.Operation)
	assert.Equal(t, "test_source", fields.Source)
	assert.Equal(t, 100*time.Millisecond, fields.Duration)
	assert.Equal(t, 42, fields.ItemCount)
	assert.Equal(t, "success", fields.Status)
	assert.Equal(t, "test query", fields.Query)
	assert.Equal(t, 3, fields.SourceCount)
	assert.Equal(t, 12, fields.RelationshipCount)
}
