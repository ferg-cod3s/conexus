// Package observability provides Prometheus metrics, OpenTelemetry tracing,
// and structured logging for Conexus.
package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// FederationMetrics holds all Prometheus metrics for federation operations.
type FederationMetrics struct {
	// Federation search metrics
	FederationSearchesTotal       *prometheus.CounterVec
	FederationSearchDuration      *prometheus.HistogramVec
	FederationSearchResults       *prometheus.HistogramVec
	FederationMergedResultsCount  *prometheus.HistogramVec
	FederationDeduplicationRatio  *prometheus.HistogramVec

	// Per-connector metrics
	ConnectorSearchesTotal        *prometheus.CounterVec
	ConnectorSearchDuration       *prometheus.HistogramVec
	ConnectorSearchResults        *prometheus.HistogramVec
	ConnectorErrorsTotal          *prometheus.CounterVec
	ConnectorTimeouts             *prometheus.CounterVec
	ConnectorSuccessRate          *prometheus.GaugeVec

	// Result processing metrics
	ResultMergeDuration           *prometheus.HistogramVec
	ResultDeduplicationDuration   *prometheus.HistogramVec
	ScoreNormalizationDuration    *prometheus.HistogramVec
	PaginationOperations          *prometheus.CounterVec

	// Federation pool metrics
	ActiveConnectors              prometheus.Gauge
	ConnectorExecutionTime        *prometheus.HistogramVec
	ParallelExecutionEfficiency   prometheus.Gauge
}

// NewFederationMetrics creates and registers all federation-related metrics.
func NewFederationMetrics(namespace string) *FederationMetrics {
	if namespace == "" {
		namespace = "conexus"
	}

	return &FederationMetrics{
		// Federation search metrics
		FederationSearchesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "federation_searches_total",
				Help:      "Total number of federation searches by status",
			},
			[]string{"status"},
		),
		FederationSearchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "federation_search_duration_seconds",
				Help:      "Federation search total duration in seconds",
				Buckets:   []float64{.05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"phase"},
		),
		FederationSearchResults: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "federation_search_results_count",
				Help:      "Number of results returned by federation search",
				Buckets:   []float64{0, 1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{},
		),
		FederationMergedResultsCount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "federation_merged_results_count",
				Help:      "Number of results before and after merging",
				Buckets:   []float64{0, 1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"stage"},
		),
		FederationDeduplicationRatio: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "federation_deduplication_ratio",
				Help:      "Ratio of duplicate results removed during deduplication (0-1)",
				Buckets:   []float64{0, 0.05, 0.1, 0.15, 0.2, 0.3, 0.5, 0.75, 1.0},
			},
			[]string{},
		),

		// Per-connector metrics
		ConnectorSearchesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "connector_searches_total",
				Help:      "Total number of searches by connector and status",
			},
			[]string{"connector_id", "connector_type", "status"},
		),
		ConnectorSearchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "connector_search_duration_seconds",
				Help:      "Connector search duration in seconds",
				Buckets:   []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"connector_id", "connector_type"},
		),
		ConnectorSearchResults: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "connector_search_results_count",
				Help:      "Number of results returned by connector",
				Buckets:   []float64{0, 1, 5, 10, 25, 50, 100, 250, 500},
			},
			[]string{"connector_id", "connector_type"},
		),
		ConnectorErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "connector_errors_total",
				Help:      "Total number of connector errors by type and error class",
			},
			[]string{"connector_id", "connector_type", "error_type"},
		),
		ConnectorTimeouts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "connector_timeouts_total",
				Help:      "Total number of connector timeouts",
			},
			[]string{"connector_id", "connector_type"},
		),
		ConnectorSuccessRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "connector_success_rate",
				Help:      "Success rate of connector searches (0-1)",
			},
			[]string{"connector_id", "connector_type"},
		),

		// Result processing metrics
		ResultMergeDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "result_merge_duration_seconds",
				Help:      "Duration of result merging operation in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5},
			},
			[]string{},
		),
		ResultDeduplicationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "result_deduplication_duration_seconds",
				Help:      "Duration of result deduplication in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5},
			},
			[]string{},
		),
		ScoreNormalizationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "score_normalization_duration_seconds",
				Help:      "Duration of score normalization in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1},
			},
			[]string{},
		),
		PaginationOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pagination_operations_total",
				Help:      "Total number of pagination operations",
			},
			[]string{"page_size"},
		),

		// Federation pool metrics
		ActiveConnectors: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_connectors",
				Help:      "Number of active searchable connectors",
			},
		),
		ConnectorExecutionTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "connector_execution_time_seconds",
				Help:      "Execution time for individual connector searches",
				Buckets:   []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"connector_id"},
		),
		ParallelExecutionEfficiency: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "parallel_execution_efficiency",
				Help:      "Efficiency of parallel execution (0-1, where 1 is perfectly parallel)",
			},
		),
	}
}

// RecordFederationSearch records a federation search operation.
func (f *FederationMetrics) RecordFederationSearch(status string, duration time.Duration, resultCount int) {
	f.FederationSearchesTotal.WithLabelValues(status).Inc()
	f.FederationSearchDuration.WithLabelValues("total").Observe(duration.Seconds())
	f.FederationSearchResults.WithLabelValues().Observe(float64(resultCount))
}

// RecordConnectorSearch records a connector search operation.
func (f *FederationMetrics) RecordConnectorSearch(connectorID, connectorType, status string, duration time.Duration, resultCount int) {
	f.ConnectorSearchesTotal.WithLabelValues(connectorID, connectorType, status).Inc()
	f.ConnectorSearchDuration.WithLabelValues(connectorID, connectorType).Observe(duration.Seconds())
	f.ConnectorSearchResults.WithLabelValues(connectorID, connectorType).Observe(float64(resultCount))
	f.ConnectorExecutionTime.WithLabelValues(connectorID).Observe(duration.Seconds())
}

// RecordConnectorError records a connector error.
func (f *FederationMetrics) RecordConnectorError(connectorID, connectorType, errorType string) {
	f.ConnectorErrorsTotal.WithLabelValues(connectorID, connectorType, errorType).Inc()
}

// RecordConnectorTimeout records a connector timeout.
func (f *FederationMetrics) RecordConnectorTimeout(connectorID, connectorType string) {
	f.ConnectorTimeouts.WithLabelValues(connectorID, connectorType).Inc()
}

// UpdateConnectorSuccessRate updates the success rate for a connector.
func (f *FederationMetrics) UpdateConnectorSuccessRate(connectorID, connectorType string, successRate float64) {
	if successRate < 0 {
		successRate = 0
	}
	if successRate > 1 {
		successRate = 1
	}
	f.ConnectorSuccessRate.WithLabelValues(connectorID, connectorType).Set(successRate)
}

// RecordMergeDuration records the time taken to merge results.
func (f *FederationMetrics) RecordMergeDuration(duration time.Duration) {
	f.ResultMergeDuration.WithLabelValues().Observe(duration.Seconds())
}

// RecordDeduplicationDuration records the time taken to deduplicate results.
func (f *FederationMetrics) RecordDeduplicationDuration(duration time.Duration) {
	f.ResultDeduplicationDuration.WithLabelValues().Observe(duration.Seconds())
}

// RecordScoreNormalizationDuration records the time taken to normalize scores.
func (f *FederationMetrics) RecordScoreNormalizationDuration(duration time.Duration) {
	f.ScoreNormalizationDuration.WithLabelValues().Observe(duration.Seconds())
}

// RecordMergedResults records the number of results before and after merging.
func (f *FederationMetrics) RecordMergedResults(stage string, count int) {
	f.FederationMergedResultsCount.WithLabelValues(stage).Observe(float64(count))
}

// RecordDeduplicationRatio records the ratio of duplicate results removed.
func (f *FederationMetrics) RecordDeduplicationRatio(ratio float64) {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	f.FederationDeduplicationRatio.WithLabelValues().Observe(ratio)
}

// RecordPaginationOperation records a pagination operation.
func (f *FederationMetrics) RecordPaginationOperation(pageSize string) {
	f.PaginationOperations.WithLabelValues(pageSize).Inc()
}

// UpdateActiveConnectors updates the count of active connectors.
func (f *FederationMetrics) UpdateActiveConnectors(count int) {
	f.ActiveConnectors.Set(float64(count))
}

// UpdateParallelExecutionEfficiency updates the parallel execution efficiency metric.
func (f *FederationMetrics) UpdateParallelExecutionEfficiency(efficiency float64) {
	if efficiency < 0 {
		efficiency = 0
	}
	if efficiency > 1 {
		efficiency = 1
	}
	f.ParallelExecutionEfficiency.Set(efficiency)
}
