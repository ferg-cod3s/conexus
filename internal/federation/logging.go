package federation

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// LogFields represents structured log fields for federation operations
type LogFields struct {
	Operation       string        `json:"operation"`
	Source          string        `json:"source,omitempty"`
	Duration        time.Duration `json:"duration,omitempty"`
	ItemCount       int           `json:"item_count,omitempty"`
	Status          string        `json:"status"`
	Error           string        `json:"error,omitempty"`
	Query           string        `json:"query,omitempty"`
	SourceCount     int           `json:"source_count,omitempty"`
	ErrorCount      int           `json:"error_count,omitempty"`
	DuplicateCount  int           `json:"duplicate_count,omitempty"`
	UniqueCount     int           `json:"unique_count,omitempty"`
	RelationshipCount int         `json:"relationship_count,omitempty"`
	QueryID         string        `json:"query_id,omitempty"`
}

// Logger provides structured logging for federation operations
type Logger struct {
	logger *slog.Logger
	ctx    context.Context
}

// NewLogger creates a new federation logger
func NewLogger(ctx context.Context) *Logger {
	return &Logger{
		logger: slog.Default(),
		ctx:    ctx,
	}
}

// LogQueryStart logs the start of a multi-source query
func (l *Logger) LogQueryStart(queryID, query string, sourceCount int) {
	l.logger.Info("Query started",
		slog.String("operation", "query_start"),
		slog.String("query_id", queryID),
		slog.String("query", query),
		slog.Int("source_count", sourceCount),
		slog.String("status", "started"),
	)
}

// LogQueryEnd logs the completion of a multi-source query
func (l *Logger) LogQueryEnd(queryID string, duration time.Duration, totalItems, errorCount, relationshipsFound int, success bool) {
	status := "success"
	if !success {
		status = "failed"
		l.logger.Warn("Query completed",
			slog.String("operation", "query_end"),
			slog.String("query_id", queryID),
			slog.Duration("duration", duration),
			slog.Int("total_items", totalItems),
			slog.Int("error_count", errorCount),
			slog.Int("relationships_found", relationshipsFound),
			slog.String("status", status),
		)
	} else {
		l.logger.Info("Query completed",
			slog.String("operation", "query_end"),
			slog.String("query_id", queryID),
			slog.Duration("duration", duration),
			slog.Int("total_items", totalItems),
			slog.Int("error_count", errorCount),
			slog.Int("relationships_found", relationshipsFound),
			slog.String("status", status),
		)
	}
}

// LogSourceQuery logs a query to a specific source
func (l *Logger) LogSourceQuery(queryID, source string, duration time.Duration, itemCount int, err error) {
	status := "success"
	var errMsg string

	if err != nil {
		status = "failed"
		errMsg = err.Error()
		l.logger.Warn(fmt.Sprintf("Source query: %s", source),
			slog.String("operation", "source_query"),
			slog.String("query_id", queryID),
			slog.String("source", source),
			slog.Duration("duration", duration),
			slog.Int("item_count", itemCount),
			slog.String("status", status),
			slog.String("error", errMsg),
		)
	} else {
		l.logger.Info(fmt.Sprintf("Source query: %s", source),
			slog.String("operation", "source_query"),
			slog.String("query_id", queryID),
			slog.String("source", source),
			slog.Duration("duration", duration),
			slog.Int("item_count", itemCount),
			slog.String("status", status),
		)
	}
}

// LogMergeStart logs the start of a merge operation
func (l *Logger) LogMergeStart(sourceCount int) {
	l.logger.Info("Merge operation started",
		slog.String("operation", "merge_start"),
		slog.Int("source_count", sourceCount),
		slog.String("status", "started"),
	)
}

// LogMergeEnd logs the completion of a merge operation
func (l *Logger) LogMergeEnd(duration time.Duration, stats DeduplicationStats) {
	l.logger.Info("Merge operation completed",
		slog.String("operation", "merge_end"),
		slog.Duration("duration", duration),
		slog.Int("merged_items", stats.MergedResults),
		slog.Int("duplicates_found", stats.DuplicatesFound),
		slog.Int("unique_items", stats.UniqueResults),
		slog.Int("total_results", stats.TotalResults),
		slog.String("status", "success"),
	)
}

// LogRelationshipDetection logs relationship detection results
func (l *Logger) LogRelationshipDetection(relationshipsFound int, duration time.Duration) {
	l.logger.Info("Relationship detection completed",
		slog.String("operation", "relationship_detection"),
		slog.Int("relationships_found", relationshipsFound),
		slog.Duration("duration", duration),
		slog.String("status", "success"),
	)
}

// LogError logs an error with federation context
func (l *Logger) LogError(operation string, err error, context map[string]interface{}) {
	msg := fmt.Sprintf("Operation failed: %s", operation)
	l.logger.Error(msg,
		slog.String("operation", operation),
		slog.String("error", err.Error()),
	)
}
