// Package observability provides Prometheus metrics, OpenTelemetry tracing,
// and structured logging for Conexus.
package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector holds all Prometheus metrics for Conexus.
type MetricsCollector struct {
	// MCP request metrics
	MCPRequestsTotal    *prometheus.CounterVec
	MCPRequestDuration  *prometheus.HistogramVec
	MCPRequestsInFlight *prometheus.GaugeVec
	MCPErrors           *prometheus.CounterVec

	// Indexer metrics
	IndexerOperations  *prometheus.CounterVec
	IndexerDuration    *prometheus.HistogramVec
	IndexedFilesTotal  prometheus.Counter
	IndexedChunksTotal prometheus.Counter
	IndexerErrorsTotal *prometheus.CounterVec

	// Embedding metrics
	EmbeddingRequests    *prometheus.CounterVec
	EmbeddingDuration    *prometheus.HistogramVec
	EmbeddingCacheHits   prometheus.Counter
	EmbeddingCacheMisses prometheus.Counter
	EmbeddingErrorsTotal *prometheus.CounterVec

	// Search cache metrics
	SearchCacheHits   prometheus.Counter
	SearchCacheMisses prometheus.Counter

	// Vector store metrics
	VectorSearchRequests *prometheus.CounterVec
	VectorSearchDuration *prometheus.HistogramVec
	VectorSearchResults  *prometheus.HistogramVec
	VectorStoreSize      prometheus.Gauge

	// System metrics
	SystemStartTime prometheus.Gauge
	SystemHealth    *prometheus.GaugeVec
}

// NewMetricsCollector creates and registers all Prometheus metrics.
func NewMetricsCollector(namespace string) *MetricsCollector {
	if namespace == "" {
		namespace = "conexus"
	}

	return &MetricsCollector{
		// MCP request metrics
		MCPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "mcp_requests_total",
				Help:      "Total number of MCP requests by method and status",
			},
			[]string{"method", "status"},
		),
		MCPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "mcp_request_duration_seconds",
				Help:      "MCP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"method"},
		),
		MCPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "mcp_requests_in_flight",
				Help:      "Number of MCP requests currently being handled",
			},
			[]string{"method"},
		),
		MCPErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "mcp_errors_total",
				Help:      "Total number of MCP errors by method and error type",
			},
			[]string{"method", "error_type"},
		),

		// Indexer metrics
		IndexerOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexer_operations_total",
				Help:      "Total number of indexer operations by type and status",
			},
			[]string{"operation", "status"},
		),
		IndexerDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "indexer_operation_duration_seconds",
				Help:      "Indexer operation duration in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60},
			},
			[]string{"operation"},
		),
		IndexedFilesTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexed_files_total",
				Help:      "Total number of files indexed",
			},
		),
		IndexedChunksTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexed_chunks_total",
				Help:      "Total number of chunks indexed",
			},
		),
		IndexerErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexer_errors_total",
				Help:      "Total number of indexer errors by type",
			},
			[]string{"error_type"},
		),

		// Embedding metrics
		EmbeddingRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_requests_total",
				Help:      "Total number of embedding requests by provider and status",
			},
			[]string{"provider", "status"},
		),
		EmbeddingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "embedding_duration_seconds",
				Help:      "Embedding generation duration in seconds",
				Buckets:   []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"provider"},
		),
		EmbeddingCacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_cache_hits_total",
				Help:      "Total number of embedding cache hits",
			},
		),
		EmbeddingCacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_cache_misses_total",
				Help:      "Total number of embedding cache misses",
			},
		),
		SearchCacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "search_cache_hits_total",
				Help:      "Total number of search cache hits",
			},
		),
		SearchCacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "search_cache_misses_total",
				Help:      "Total number of search cache misses",
			},
		),
		EmbeddingErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_errors_total",
				Help:      "Total number of embedding errors by provider and type",
			},
			[]string{"provider", "error_type"},
		),

		// Vector store metrics
		VectorSearchRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "vector_search_requests_total",
				Help:      "Total number of vector search requests by type and status",
			},
			[]string{"search_type", "status"},
		),
		VectorSearchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "vector_search_duration_seconds",
				Help:      "Vector search duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"search_type"},
		),
		VectorSearchResults: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "vector_search_results_count",
				Help:      "Number of results returned by vector search",
				Buckets:   []float64{0, 1, 5, 10, 25, 50, 100, 250, 500},
			},
			[]string{"search_type"},
		),
		VectorStoreSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "vector_store_size_bytes",
				Help:      "Total size of vector store in bytes",
			},
		),

		// System metrics
		SystemStartTime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_start_time_seconds",
				Help:      "Unix timestamp when the system started",
			},
		),
		SystemHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_health_status",
				Help:      "System health status (1 = healthy, 0 = unhealthy)",
			},
			[]string{"component"},
		),
	}
}

// RecordMCPRequest records metrics for an MCP request.
func (m *MetricsCollector) RecordMCPRequest(method, status string, duration time.Duration) {
	m.MCPRequestsTotal.WithLabelValues(method, status).Inc()
	m.MCPRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordMCPError records an MCP error.
func (m *MetricsCollector) RecordMCPError(method, errorType string) {
	m.MCPErrors.WithLabelValues(method, errorType).Inc()
}

// TrackMCPInFlight tracks in-flight MCP requests.
func (m *MetricsCollector) TrackMCPInFlight(method string, delta float64) {
	m.MCPRequestsInFlight.WithLabelValues(method).Add(delta)
}

// RecordIndexerOperation records metrics for an indexer operation.
func (m *MetricsCollector) RecordIndexerOperation(operation, status string, duration time.Duration) {
	m.IndexerOperations.WithLabelValues(operation, status).Inc()
	m.IndexerDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordIndexedFiles increments the indexed files counter.
func (m *MetricsCollector) RecordIndexedFiles(count int) {
	m.IndexedFilesTotal.Add(float64(count))
}

// RecordIndexedChunks increments the indexed chunks counter.
func (m *MetricsCollector) RecordIndexedChunks(count int) {
	m.IndexedChunksTotal.Add(float64(count))
}

// RecordIndexerError records an indexer error.
func (m *MetricsCollector) RecordIndexerError(errorType string) {
	m.IndexerErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordEmbedding records metrics for an embedding request.
func (m *MetricsCollector) RecordEmbedding(provider, status string, duration time.Duration) {
	m.EmbeddingRequests.WithLabelValues(provider, status).Inc()
	m.EmbeddingDuration.WithLabelValues(provider).Observe(duration.Seconds())
}

// RecordEmbeddingCacheHit records a cache hit.
func (m *MetricsCollector) RecordEmbeddingCacheHit() {
	m.EmbeddingCacheHits.Inc()
}

// RecordEmbeddingCacheMiss records a cache miss.
func (m *MetricsCollector) RecordEmbeddingCacheMiss() {
	m.EmbeddingCacheMisses.Inc()
}

// RecordSearchCacheHit records a search cache hit.
func (m *MetricsCollector) RecordSearchCacheHit() {
	m.SearchCacheHits.Inc()
}

// RecordSearchCacheMiss records a search cache miss.
func (m *MetricsCollector) RecordSearchCacheMiss() {
	m.SearchCacheMisses.Inc()
}

// RecordEmbeddingError records an embedding error.
func (m *MetricsCollector) RecordEmbeddingError(provider, errorType string) {
	m.EmbeddingErrorsTotal.WithLabelValues(provider, errorType).Inc()
}

// RecordVectorSearch records metrics for a vector search request.
func (m *MetricsCollector) RecordVectorSearch(searchType, status string, duration time.Duration, resultCount int) {
	m.VectorSearchRequests.WithLabelValues(searchType, status).Inc()
	m.VectorSearchDuration.WithLabelValues(searchType).Observe(duration.Seconds())
	m.VectorSearchResults.WithLabelValues(searchType).Observe(float64(resultCount))
}

// UpdateVectorStoreSize updates the vector store size metric.
func (m *MetricsCollector) UpdateVectorStoreSize(sizeBytes int64) {
	m.VectorStoreSize.Set(float64(sizeBytes))
}

// SetSystemStartTime sets the system start time.
func (m *MetricsCollector) SetSystemStartTime(startTime time.Time) {
	m.SystemStartTime.Set(float64(startTime.Unix()))
}

// SetComponentHealth sets the health status of a component.
func (m *MetricsCollector) SetComponentHealth(component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	m.SystemHealth.WithLabelValues(component).Set(value)
}
