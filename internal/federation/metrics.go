package federation

import (
	"sync"
)


// SourceMetrics holds metrics for a specific source
type SourceMetrics struct {
	QueriesTotal    int64
	ErrorsTotal     int64
	TotalDuration   int64 // in nanoseconds
	AverageDuration int64 // in nanoseconds
}

// MetricsSnapshot is a snapshot of current metrics at a point in time
type MetricsSnapshot struct {
	// Query metrics
	QueriesTotal        int64
	QueryErrorsTotal    int64
	AverageQueryDuration int64 // in nanoseconds
	MinQueryDuration    int64  // in nanoseconds
	MaxQueryDuration    int64  // in nanoseconds

	// Per-source metrics
	SourceMetrics       map[string]SourceMetrics
	ActiveSourcesCount  int64

	// Deduplication metrics
	TotalResultsProcessed int64
	DuplicatesFound       int64
	UniqueResultsCreated  int64
	MergedResultsCreated  int64

	// Merge operation metrics
	MergeOperationsTotal int64
	AverageMergeDuration int64 // in nanoseconds
	MinMergeDuration     int64  // in nanoseconds
	MaxMergeDuration     int64  // in nanoseconds

	// Relationship detection metrics
	RelationshipDetectionsTotal int64
	RelationshipsFoundTotal     int64

	// Error tracking by type
	ErrorsByType map[string]int64
}

// MetricsCollector collects observability metrics for the federation service
type MetricsCollector struct {
	mu sync.RWMutex

	// Query metrics
	queriesTotal       int64
	queryErrorsTotal   int64
	totalQueryDuration int64 // in nanoseconds
	queryMinDuration   int64
	queryMaxDuration   int64

	// Per-source metrics
	sourceQueriesTotal map[string]int64
	sourceErrorsTotal  map[string]int64
	sourceTotalDuration map[string]int64 // in nanoseconds

	// Deduplication metrics
	totalResultsProcessed int64
	duplicatesFound       int64
	uniqueResultsCreated  int64
	mergedResultsCreated  int64

	// Merge operation metrics
	mergeOperationsTotal  int64
	totalMergeDuration    int64 // in nanoseconds
	mergeMinDuration      int64
	mergeMaxDuration      int64

	// Relationship detection metrics
	relationshipDetectionsTotal int64
	relationshipsFoundTotal     int64

	// Active sources tracking
	activeSources map[string]bool

	// Error tracking by type
	errorsByType map[string]int64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		sourceQueriesTotal:  make(map[string]int64),
		sourceErrorsTotal:   make(map[string]int64),
		sourceTotalDuration: make(map[string]int64),
		activeSources:       make(map[string]bool),
		errorsByType:        make(map[string]int64),
	}
}

// RecordQueryStart is called at the start of a query operation
func (mc *MetricsCollector) RecordQueryStart() {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.queriesTotal++
}

// RecordQueryEnd is called at the end of a query operation
func (mc *MetricsCollector) RecordQueryEnd(duration int64, sourcesQueried []string, success bool) {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !success {
		mc.queryErrorsTotal++
	}

	mc.totalQueryDuration += duration

	if mc.queryMinDuration == 0 || duration < mc.queryMinDuration {
		mc.queryMinDuration = duration
	}
	if duration > mc.queryMaxDuration {
		mc.queryMaxDuration = duration
	}

	// Record active sources
	for _, source := range sourcesQueried {
		mc.activeSources[source] = true
	}
}

// RecordSourceQuery records a query to a specific source
func (mc *MetricsCollector) RecordSourceQuery(source string, duration int64, success bool, itemCount int) {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.sourceQueriesTotal[source]++
	mc.sourceTotalDuration[source] += duration

	if !success {
		mc.sourceErrorsTotal[source]++
	}

	mc.activeSources[source] = true
}

// RecordDeduplicationStats records deduplication statistics
func (mc *MetricsCollector) RecordDeduplicationStats(stats DeduplicationStats) {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.totalResultsProcessed += int64(stats.TotalResults)
	mc.duplicatesFound += int64(stats.DuplicatesFound)
	mc.uniqueResultsCreated += int64(stats.UniqueResults)
	mc.mergedResultsCreated += int64(stats.MergedResults)
}

// RecordMergeOperation records a merge operation
func (mc *MetricsCollector) RecordMergeOperation(duration int64) {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.mergeOperationsTotal++
	mc.totalMergeDuration += duration

	if mc.mergeMinDuration == 0 || duration < mc.mergeMinDuration {
		mc.mergeMinDuration = duration
	}
	if duration > mc.mergeMaxDuration {
		mc.mergeMaxDuration = duration
	}
}

// RecordRelationshipDetection records a relationship detection operation
func (mc *MetricsCollector) RecordRelationshipDetection(relationshipsFound int) {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.relationshipDetectionsTotal++
	mc.relationshipsFoundTotal += int64(relationshipsFound)
}

// RecordError records an error of a specific type
func (mc *MetricsCollector) RecordError(errorType string) {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.errorsByType[errorType]++
}

// GetMetrics returns a snapshot of current metrics
func (mc *MetricsCollector) GetMetrics() MetricsSnapshot {
	if mc == nil {
		return MetricsSnapshot{
			SourceMetrics: make(map[string]SourceMetrics),
			ErrorsByType:  make(map[string]int64),
		}
	}
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Calculate averages
	var avgQueryDuration, avgMergeDuration int64
	if mc.queriesTotal > 0 {
		avgQueryDuration = mc.totalQueryDuration / mc.queriesTotal
	}
	if mc.mergeOperationsTotal > 0 {
		avgMergeDuration = mc.totalMergeDuration / mc.mergeOperationsTotal
	}

	// Copy source metrics
	sourceMetrics := make(map[string]SourceMetrics)
	for source := range mc.activeSources {
		var avgDuration int64
		if count := mc.sourceQueriesTotal[source]; count > 0 {
			avgDuration = mc.sourceTotalDuration[source] / count
		}

		sourceMetrics[source] = SourceMetrics{
			QueriesTotal:    mc.sourceQueriesTotal[source],
			ErrorsTotal:     mc.sourceErrorsTotal[source],
			TotalDuration:   mc.sourceTotalDuration[source],
			AverageDuration: avgDuration,
		}
	}

	// Copy error metrics
	errorMetrics := make(map[string]int64)
	for errType, count := range mc.errorsByType {
		errorMetrics[errType] = count
	}

	return MetricsSnapshot{
		QueriesTotal:               mc.queriesTotal,
		QueryErrorsTotal:           mc.queryErrorsTotal,
		AverageQueryDuration:       avgQueryDuration,
		MinQueryDuration:           mc.queryMinDuration,
		MaxQueryDuration:           mc.queryMaxDuration,
		SourceMetrics:              sourceMetrics,
		ActiveSourcesCount:         int64(len(mc.activeSources)),
		TotalResultsProcessed:      mc.totalResultsProcessed,
		DuplicatesFound:            mc.duplicatesFound,
		UniqueResultsCreated:       mc.uniqueResultsCreated,
		MergedResultsCreated:       mc.mergedResultsCreated,
		MergeOperationsTotal:       mc.mergeOperationsTotal,
		AverageMergeDuration:       avgMergeDuration,
		MinMergeDuration:           mc.mergeMinDuration,
		MaxMergeDuration:           mc.mergeMaxDuration,
		RelationshipDetectionsTotal: mc.relationshipDetectionsTotal,
		RelationshipsFoundTotal:     mc.relationshipsFoundTotal,
		ErrorsByType:               errorMetrics,
	}
}

// Reset clears all metrics
func (mc *MetricsCollector) Reset() {
	if mc == nil {
		return
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.queriesTotal = 0
	mc.queryErrorsTotal = 0
	mc.totalQueryDuration = 0
	mc.queryMinDuration = 0
	mc.queryMaxDuration = 0
	mc.sourceQueriesTotal = make(map[string]int64)
	mc.sourceErrorsTotal = make(map[string]int64)
	mc.sourceTotalDuration = make(map[string]int64)
	mc.totalResultsProcessed = 0
	mc.duplicatesFound = 0
	mc.uniqueResultsCreated = 0
	mc.mergedResultsCreated = 0
	mc.mergeOperationsTotal = 0
	mc.totalMergeDuration = 0
	mc.mergeMinDuration = 0
	mc.mergeMaxDuration = 0
	mc.relationshipDetectionsTotal = 0
	mc.relationshipsFoundTotal = 0
	mc.activeSources = make(map[string]bool)
	mc.errorsByType = make(map[string]int64)
}
