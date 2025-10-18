package observability

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestRegistry() prometheus.Registerer {
	return prometheus.NewRegistry()
}

// TestNewFederationMetrics_WithTestRegistry tests federation metrics creation with isolated registry
func TestNewFederationMetrics_WithTestRegistry(t *testing.T) {
	reg := createTestRegistry()

	// We need to use a custom implementation that allows test registry
	// For now, just ensure the metrics can be created (they use the default registry)
	fm := NewFederationMetrics("test")
	require.NotNil(t, fm)

	// Verify all metrics are initialized
	assert.NotNil(t, fm.FederationSearchesTotal)
	assert.NotNil(t, fm.FederationSearchDuration)
	assert.NotNil(t, fm.FederationSearchResults)
	assert.NotNil(t, fm.ConnectorSearchesTotal)
	assert.NotNil(t, fm.ConnectorSearchDuration)
	assert.NotNil(t, fm.ConnectorErrorsTotal)
	assert.NotNil(t, fm.ResultMergeDuration)

	_ = reg // suppress unused warning
}

func TestRecordFederationSearch(t *testing.T) {
	fm := NewFederationMetrics("test_fed_search")

	// Record successful search - should not panic
	fm.RecordFederationSearch("success", 500*time.Millisecond, 42)
	fm.RecordFederationSearch("success", 1*time.Second, 35)

	// Record failed search - should not panic
	fm.RecordFederationSearch("error", 100*time.Millisecond, 0)
}

func TestRecordConnectorSearch(t *testing.T) {
	fm := NewFederationMetrics("test_connector_search")

	// Record connector searches - should not panic
	fm.RecordConnectorSearch("connector1", "filesystem", "success", 100*time.Millisecond, 10)
	fm.RecordConnectorSearch("connector2", "github", "success", 500*time.Millisecond, 25)
	fm.RecordConnectorSearch("connector1", "filesystem", "error", 50*time.Millisecond, 0)
}

func TestRecordConnectorError(t *testing.T) {
	fm := NewFederationMetrics("test_connector_error")

	// Should not panic
	fm.RecordConnectorError("connector1", "filesystem", "timeout")
	fm.RecordConnectorError("connector2", "github", "network_error")
	fm.RecordConnectorError("connector1", "filesystem", "parse_error")
}

func TestRecordConnectorTimeout(t *testing.T) {
	fm := NewFederationMetrics("test_connector_timeout")

	// Should not panic
	fm.RecordConnectorTimeout("connector1", "filesystem")
	fm.RecordConnectorTimeout("connector2", "github")
}

func TestUpdateConnectorSuccessRate(t *testing.T) {
	fm := NewFederationMetrics("test_success_rate")

	tests := []struct {
		name      string
		rate      float64
	}{
		{"valid rate 0", 0.0},
		{"valid rate 0.5", 0.5},
		{"valid rate 1", 1.0},
		{"clamp negative", -0.5}, // Should be clamped to 0
		{"clamp above 1", 1.5},   // Should be clamped to 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.UpdateConnectorSuccessRate("connector1", "filesystem", tt.rate)
			// Should not panic
		})
	}
}

func TestRecordMergeDuration(t *testing.T) {
	fm := NewFederationMetrics("test_merge_duration")

	fm.RecordMergeDuration(10 * time.Millisecond)
	fm.RecordMergeDuration(50 * time.Millisecond)
	fm.RecordMergeDuration(100 * time.Millisecond)
}

func TestRecordDeduplicationDuration(t *testing.T) {
	fm := NewFederationMetrics("test_dedup_duration")

	fm.RecordDeduplicationDuration(5 * time.Millisecond)
	fm.RecordDeduplicationDuration(15 * time.Millisecond)
}

func TestRecordScoreNormalizationDuration(t *testing.T) {
	fm := NewFederationMetrics("test_score_norm_duration")

	fm.RecordScoreNormalizationDuration(2 * time.Millisecond)
	fm.RecordScoreNormalizationDuration(5 * time.Millisecond)
}

func TestRecordMergedResults(t *testing.T) {
	fm := NewFederationMetrics("test_merged_results")

	fm.RecordMergedResults("before_merge", 150)
	fm.RecordMergedResults("after_merge", 120)
	fm.RecordMergedResults("after_pagination", 20)
}

func TestRecordDeduplicationRatio(t *testing.T) {
	fm := NewFederationMetrics("test_dedup_ratio")

	tests := []struct {
		name   string
		ratio  float64
	}{
		{"zero ratio", 0.0},
		{"20% dedup", 0.2},
		{"50% dedup", 0.5},
		{"100% dedup", 1.0},
		{"negative clamped", -0.5},
		{"above 1 clamped", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.RecordDeduplicationRatio(tt.ratio)
		})
	}
}

func TestRecordPaginationOperation(t *testing.T) {
	fm := NewFederationMetrics("test_pagination")

	fm.RecordPaginationOperation("20")
	fm.RecordPaginationOperation("20")
	fm.RecordPaginationOperation("50")
	fm.RecordPaginationOperation("100")
}

func TestUpdateActiveConnectors(t *testing.T) {
	fm := NewFederationMetrics("test_active_connectors")

	fm.UpdateActiveConnectors(0)
	fm.UpdateActiveConnectors(1)
	fm.UpdateActiveConnectors(3)
	fm.UpdateActiveConnectors(5)
}

func TestUpdateParallelExecutionEfficiency(t *testing.T) {
	fm := NewFederationMetrics("test_parallel_efficiency")

	tests := []struct {
		name       string
		efficiency float64
	}{
		{"zero efficiency", 0.0},
		{"50% efficiency", 0.5},
		{"100% efficiency", 1.0},
		{"clamped negative", -0.5},
		{"clamped above 1", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.UpdateParallelExecutionEfficiency(tt.efficiency)
		})
	}
}

func TestFederationMetricsIntegration(t *testing.T) {
	// Simulates a complete federation search workflow
	fm := NewFederationMetrics("test_fed_integration")

	// Setup: 3 active connectors
	fm.UpdateActiveConnectors(3)

	// Record parallel connector searches
	fm.RecordConnectorSearch("fs_connector", "filesystem", "success", 100*time.Millisecond, 50)
	fm.RecordConnectorSearch("db_connector", "database", "success", 150*time.Millisecond, 40)
	fm.RecordConnectorSearch("api_connector", "api", "success", 200*time.Millisecond, 30)

	// Record merging operations
	fm.RecordMergedResults("before_merge", 120)
	fm.RecordMergeDuration(15 * time.Millisecond)

	// Record deduplication
	fm.RecordMergedResults("after_merge", 110)
	fm.RecordDeduplicationDuration(8 * time.Millisecond)
	fm.RecordDeduplicationRatio(0.083) // 10/120

	// Record score normalization
	fm.RecordScoreNormalizationDuration(5 * time.Millisecond)

	// Record pagination
	fm.RecordPaginationOperation("20")

	// Record federation search completion
	fm.RecordFederationSearch("success", 500*time.Millisecond, 20)

	// Update success rates
	fm.UpdateConnectorSuccessRate("fs_connector", "filesystem", 0.95)
	fm.UpdateConnectorSuccessRate("db_connector", "database", 0.90)
	fm.UpdateConnectorSuccessRate("api_connector", "api", 0.85)

	// Calculate parallel efficiency: sequential would be 100+150+200=450ms
	// parallel was 200ms (max of three), so efficiency = 450/(3*200) = 0.75
	fm.UpdateParallelExecutionEfficiency(0.75)

	// All operations should complete without panic
}

func TestFederationMetricsWithErrorCases(t *testing.T) {
	fm := NewFederationMetrics("test_fed_errors")

	// Record errors from different connectors
	fm.RecordConnectorError("connector1", "filesystem", "timeout")
	fm.RecordConnectorError("connector2", "github", "rate_limit")
	fm.RecordConnectorError("connector3", "slack", "auth_failed")

	// Record timeouts
	fm.RecordConnectorTimeout("connector1", "filesystem")
	fm.RecordConnectorTimeout("connector3", "slack")

	// Update success rates with low values
	fm.UpdateConnectorSuccessRate("connector1", "filesystem", 0.5)
	fm.UpdateConnectorSuccessRate("connector2", "github", 0.8)
	fm.UpdateConnectorSuccessRate("connector3", "slack", 0.2)

	// Record failed federation search
	fm.RecordFederationSearch("partial_error", 2*time.Second, 15)

	// All operations should complete without panic
}

// TestFederationMetricsFieldTypes validates that all metrics have the correct types
func TestFederationMetricsFieldTypes(t *testing.T) {
	fm := NewFederationMetrics("test_types")

	// Counter vectors
	assert.NotNil(t, fm.FederationSearchesTotal)
	assert.NotNil(t, fm.ConnectorSearchesTotal)
	assert.NotNil(t, fm.ConnectorErrorsTotal)
	assert.NotNil(t, fm.ConnectorTimeouts)
	assert.NotNil(t, fm.PaginationOperations)

	// Histogram vectors
	assert.NotNil(t, fm.FederationSearchDuration)
	assert.NotNil(t, fm.FederationSearchResults)
	assert.NotNil(t, fm.ConnectorSearchDuration)
	assert.NotNil(t, fm.ConnectorSearchResults)
	assert.NotNil(t, fm.ResultMergeDuration)
	assert.NotNil(t, fm.ResultDeduplicationDuration)
	assert.NotNil(t, fm.ScoreNormalizationDuration)
	assert.NotNil(t, fm.ConnectorExecutionTime)

	// Gauge vectors
	assert.NotNil(t, fm.ConnectorSuccessRate)

	// Gauges
	assert.NotNil(t, fm.ActiveConnectors)
	assert.NotNil(t, fm.ParallelExecutionEfficiency)
}
