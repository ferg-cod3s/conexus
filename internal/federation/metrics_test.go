package federation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsCollector_RecordQueryStart(t *testing.T) {
	mc := NewMetricsCollector()
	assert.Equal(t, int64(0), mc.GetMetrics().QueriesTotal)

	mc.RecordQueryStart()
	assert.Equal(t, int64(1), mc.GetMetrics().QueriesTotal)

	mc.RecordQueryStart()
	assert.Equal(t, int64(2), mc.GetMetrics().QueriesTotal)
}

func TestMetricsCollector_RecordQueryEnd(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordQueryStart()
	duration := int64(100 * time.Millisecond)
	mc.RecordQueryEnd(duration, []string{"source1", "source2"}, true)

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(1), metrics.QueriesTotal)
	assert.Equal(t, int64(0), metrics.QueryErrorsTotal)
	assert.Equal(t, int64(2), metrics.ActiveSourcesCount)
	assert.Equal(t, duration, metrics.AverageQueryDuration)
	assert.Equal(t, duration, metrics.MinQueryDuration)
	assert.Equal(t, duration, metrics.MaxQueryDuration)
}

func TestMetricsCollector_RecordQueryEnd_WithError(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordQueryStart()
	mc.RecordQueryEnd(int64(100*time.Millisecond), []string{"source1"}, false)

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(1), metrics.QueriesTotal)
	assert.Equal(t, int64(1), metrics.QueryErrorsTotal)
}

func TestMetricsCollector_RecordSourceQuery(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordSourceQuery("source1", int64(50*time.Millisecond), true, 10)
	mc.RecordSourceQuery("source2", int64(100*time.Millisecond), true, 20)

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(2), metrics.ActiveSourcesCount)

	source1Metrics, ok := metrics.SourceMetrics["source1"]
	require.True(t, ok)
	assert.Equal(t, int64(1), source1Metrics.QueriesTotal)
	assert.Equal(t, int64(0), source1Metrics.ErrorsTotal)
	assert.Equal(t, int64(50*time.Millisecond), source1Metrics.AverageDuration)

	source2Metrics, ok := metrics.SourceMetrics["source2"]
	require.True(t, ok)
	assert.Equal(t, int64(1), source2Metrics.QueriesTotal)
	assert.Equal(t, int64(100*time.Millisecond), source2Metrics.AverageDuration)
}

func TestMetricsCollector_RecordDeduplicationStats(t *testing.T) {
	mc := NewMetricsCollector()

	stats := DeduplicationStats{
		TotalResults:    100,
		DuplicatesFound: 25,
		UniqueResults:   75,
		MergedResults:   80,
	}

	mc.RecordDeduplicationStats(stats)

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(100), metrics.TotalResultsProcessed)
	assert.Equal(t, int64(25), metrics.DuplicatesFound)
	assert.Equal(t, int64(75), metrics.UniqueResultsCreated)
	assert.Equal(t, int64(80), metrics.MergedResultsCreated)
}

func TestMetricsCollector_RecordMergeOperation(t *testing.T) {
	mc := NewMetricsCollector()

	duration1 := int64(50 * time.Millisecond)
	duration2 := int64(150 * time.Millisecond)

	mc.RecordMergeOperation(duration1)
	mc.RecordMergeOperation(duration2)

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(2), metrics.MergeOperationsTotal)
	assert.Equal(t, int64(100*time.Millisecond), metrics.AverageMergeDuration)
	assert.Equal(t, duration1, metrics.MinMergeDuration)
	assert.Equal(t, duration2, metrics.MaxMergeDuration)
}

func TestMetricsCollector_RecordRelationshipDetection(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordRelationshipDetection(5)
	mc.RecordRelationshipDetection(3)

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(2), metrics.RelationshipDetectionsTotal)
	assert.Equal(t, int64(8), metrics.RelationshipsFoundTotal)
}

func TestMetricsCollector_RecordError(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordError("query_error")
	mc.RecordError("merge_error")
	mc.RecordError("query_error")

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(2), metrics.ErrorsByType["query_error"])
	assert.Equal(t, int64(1), metrics.ErrorsByType["merge_error"])
}

func TestMetricsCollector_Reset(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordQueryStart()
	mc.RecordSourceQuery("source1", int64(100*time.Millisecond), true, 10)
	mc.RecordError("test_error")

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(1), metrics.QueriesTotal)
	assert.Equal(t, int64(1), metrics.ActiveSourcesCount)
	assert.Equal(t, int64(1), metrics.ErrorsByType["test_error"])

	mc.Reset()

	metrics = mc.GetMetrics()
	assert.Equal(t, int64(0), metrics.QueriesTotal)
	assert.Equal(t, int64(0), metrics.ActiveSourcesCount)
	assert.Equal(t, 0, len(metrics.ErrorsByType))
}

func TestMetricsCollector_MultipleOperations(t *testing.T) {
	mc := NewMetricsCollector()

	// Simulate multiple queries
	for i := 0; i < 3; i++ {
		mc.RecordQueryStart()
		mc.RecordQueryEnd(int64(100*time.Millisecond), []string{"source1", "source2"}, i < 2)
		if i == 2 {
			mc.RecordSourceQuery("source1", int64(80*time.Millisecond), false, 0)
		}
	}

	// Record deduplication
	mc.RecordDeduplicationStats(DeduplicationStats{
		TotalResults:    300,
		DuplicatesFound: 75,
		UniqueResults:   225,
		MergedResults:   250,
	})

	// Record merge operation
	mc.RecordMergeOperation(int64(50 * time.Millisecond))

	// Record relationship detection
	mc.RecordRelationshipDetection(10)

	metrics := mc.GetMetrics()
	assert.Equal(t, int64(3), metrics.QueriesTotal)
	assert.Equal(t, int64(1), metrics.QueryErrorsTotal)
	assert.Equal(t, int64(300), metrics.TotalResultsProcessed)
	assert.Equal(t, int64(75), metrics.DuplicatesFound)
	assert.Equal(t, int64(1), metrics.MergeOperationsTotal)
	assert.Equal(t, int64(1), metrics.RelationshipDetectionsTotal)
	assert.Equal(t, int64(10), metrics.RelationshipsFoundTotal)
}
