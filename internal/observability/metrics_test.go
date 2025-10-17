package observability

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

// newTestMetricsCollector creates a MetricsCollector with a custom registry for testing
func newTestMetricsCollector(t *testing.T) (*MetricsCollector, *prometheus.Registry) {
	t.Helper()

	registry := prometheus.NewRegistry()
	namespace := "test"

	collector := &MetricsCollector{
		// MCP request metrics
		MCPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "mcp_requests_total",
				Help:      "Total number of MCP requests by method and status",
			},
			[]string{"method", "status"},
		),
		MCPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "mcp_request_duration_seconds",
				Help:      "MCP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"method"},
		),
		MCPRequestsInFlight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "mcp_requests_in_flight",
				Help:      "Number of MCP requests currently being handled",
			},
			[]string{"method"},
		),
		MCPErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "mcp_errors_total",
				Help:      "Total number of MCP errors by method and error type",
			},
			[]string{"method", "error_type"},
		),

		// Indexer metrics
		IndexerOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexer_operations_total",
				Help:      "Total number of indexer operations by type and status",
			},
			[]string{"operation", "status"},
		),
		IndexerDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "indexer_operation_duration_seconds",
				Help:      "Indexer operation duration in seconds",
				Buckets:   []float64{.01, .05, .1, .5, 1, 5, 10},
			},
			[]string{"operation"},
		),
		IndexedFilesTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexed_files_total",
				Help:      "Total number of files indexed",
			},
		),
		IndexedChunksTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexed_chunks_total",
				Help:      "Total number of code chunks indexed",
			},
		),
		IndexerErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "indexer_errors_total",
				Help:      "Total number of indexer errors by type",
			},
			[]string{"error_type"},
		),

		// Embedding metrics
		EmbeddingRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_requests_total",
				Help:      "Total number of embedding requests by provider and status",
			},
			[]string{"provider", "status"},
		),
		EmbeddingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "embedding_duration_seconds",
				Help:      "Embedding request duration in seconds",
				Buckets:   []float64{.01, .05, .1, .5, 1, 2.5, 5},
			},
			[]string{"provider"},
		),
		EmbeddingCacheHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_cache_hits_total",
				Help:      "Total number of embedding cache hits",
			},
		),
		EmbeddingCacheMisses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_cache_misses_total",
				Help:      "Total number of embedding cache misses",
			},
		),
		SearchCacheHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "search_cache_hits_total",
				Help:      "Total number of search cache hits",
			},
		),
		SearchCacheMisses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "search_cache_misses_total",
				Help:      "Total number of search cache misses",
			},
		),
		EmbeddingErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "embedding_errors_total",
				Help:      "Total number of embedding errors by provider and type",
			},
			[]string{"provider", "error_type"},
		),

		// Vector store metrics
		VectorSearchRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "vector_search_requests_total",
				Help:      "Total number of vector search requests by type and status",
			},
			[]string{"search_type", "status"},
		),
		VectorSearchDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "vector_search_duration_seconds",
				Help:      "Vector search duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .5},
			},
			[]string{"search_type"},
		),
		VectorSearchResults: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "vector_search_results",
				Help:      "Number of results returned by vector search",
				Buckets:   []float64{1, 5, 10, 25, 50, 100},
			},
			[]string{"search_type"},
		),
		VectorStoreSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "vector_store_size_bytes",
				Help:      "Current size of vector store in bytes",
			},
		),

		// System metrics
		SystemStartTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_start_time_seconds",
				Help:      "Unix timestamp of system start time",
			},
		),
		SystemHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "system_health",
				Help:      "Health status of system components (1=healthy, 0=unhealthy)",
			},
			[]string{"component"},
		),
	}

	// Register all metrics with the custom registry
	registry.MustRegister(
		collector.MCPRequestsTotal,
		collector.MCPRequestDuration,
		collector.MCPRequestsInFlight,
		collector.MCPErrors,
		collector.IndexerOperations,
		collector.IndexerDuration,
		collector.IndexedFilesTotal,
		collector.IndexedChunksTotal,
		collector.IndexerErrorsTotal,
		collector.EmbeddingRequests,
		collector.EmbeddingDuration,
		collector.EmbeddingCacheHits,
		collector.EmbeddingCacheMisses,
		collector.EmbeddingErrorsTotal,
		collector.VectorSearchRequests,
		collector.VectorSearchDuration,
		collector.VectorSearchResults,
		collector.VectorStoreSize,
		collector.SystemStartTime,
		collector.SystemHealth,
	)

	return collector, registry
}

func TestRecordMCPRequest(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	tests := []struct {
		name      string
		method    string
		status    string
		duration  time.Duration
		wantCount float64
	}{
		{
			name:      "successful request",
			method:    "tools/list",
			status:    "success",
			duration:  100 * time.Millisecond,
			wantCount: 1,
		},
		{
			name:      "error request",
			method:    "tools/call",
			status:    "error",
			duration:  50 * time.Millisecond,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.RecordMCPRequest(tt.method, tt.status, tt.duration)

			// Verify counter incremented
			count := testutil.ToFloat64(collector.MCPRequestsTotal.WithLabelValues(tt.method, tt.status))
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestRecordMCPError(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	tests := []struct {
		name      string
		method    string
		errorType string
		wantCount float64
	}{
		{
			name:      "validation error",
			method:    "tools/call",
			errorType: "validation",
			wantCount: 1,
		},
		{
			name:      "timeout error",
			method:    "search/code",
			errorType: "timeout",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.RecordMCPError(tt.method, tt.errorType)

			count := testutil.ToFloat64(collector.MCPErrors.WithLabelValues(tt.method, tt.errorType))
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestTrackMCPInFlight(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	method := "tools/list"

	// Start tracking
	collector.TrackMCPInFlight(method, 1.0)
	count := testutil.ToFloat64(collector.MCPRequestsInFlight.WithLabelValues(method))
	assert.Equal(t, float64(1), count)

	// Stop tracking
	collector.TrackMCPInFlight(method, -1.0)
	count = testutil.ToFloat64(collector.MCPRequestsInFlight.WithLabelValues(method))
	assert.Equal(t, float64(0), count)
}

func TestRecordIndexerOperation(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	tests := []struct {
		name      string
		operation string
		status    string
		duration  time.Duration
		wantCount float64
	}{
		{
			name:      "successful index",
			operation: "index",
			status:    "success",
			duration:  500 * time.Millisecond,
			wantCount: 1,
		},
		{
			name:      "failed scan",
			operation: "scan",
			status:    "error",
			duration:  100 * time.Millisecond,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.RecordIndexerOperation(tt.operation, tt.status, tt.duration)

			count := testutil.ToFloat64(collector.IndexerOperations.WithLabelValues(tt.operation, tt.status))
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestRecordIndexedFiles(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	// Record 5 files
	collector.RecordIndexedFiles(5)
	count := testutil.ToFloat64(collector.IndexedFilesTotal)
	assert.Equal(t, float64(5), count)

	// Record 3 more files
	collector.RecordIndexedFiles(3)
	count = testutil.ToFloat64(collector.IndexedFilesTotal)
	assert.Equal(t, float64(8), count)
}

func TestRecordIndexedChunks(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	// Record 100 chunks
	collector.RecordIndexedChunks(100)
	count := testutil.ToFloat64(collector.IndexedChunksTotal)
	assert.Equal(t, float64(100), count)

	// Record 50 more chunks
	collector.RecordIndexedChunks(50)
	count = testutil.ToFloat64(collector.IndexedChunksTotal)
	assert.Equal(t, float64(150), count)
}

func TestRecordIndexerError(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	errorType := "parse_error"
	collector.RecordIndexerError(errorType)

	count := testutil.ToFloat64(collector.IndexerErrorsTotal.WithLabelValues(errorType))
	assert.Equal(t, float64(1), count)
}

func TestRecordEmbedding(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	tests := []struct {
		name      string
		provider  string
		status    string
		duration  time.Duration
		wantCount float64
	}{
		{
			name:      "successful embedding",
			provider:  "openai",
			status:    "success",
			duration:  50 * time.Millisecond,
			wantCount: 1,
		},
		{
			name:      "failed embedding",
			provider:  "cohere",
			status:    "error",
			duration:  20 * time.Millisecond,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.RecordEmbedding(tt.provider, tt.status, tt.duration)

			count := testutil.ToFloat64(collector.EmbeddingRequests.WithLabelValues(tt.provider, tt.status))
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestRecordEmbeddingCache(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	// Record cache hit
	collector.RecordEmbeddingCacheHit()
	hits := testutil.ToFloat64(collector.EmbeddingCacheHits)
	assert.Equal(t, float64(1), hits)

	// Record cache miss
	collector.RecordEmbeddingCacheMiss()
	misses := testutil.ToFloat64(collector.EmbeddingCacheMisses)
	assert.Equal(t, float64(1), misses)
}

func TestRecordSearchCache(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	// Record cache hit
	collector.RecordSearchCacheHit()
	hits := testutil.ToFloat64(collector.SearchCacheHits)
	assert.Equal(t, float64(1), hits)

	// Record cache miss
	collector.RecordSearchCacheMiss()
	misses := testutil.ToFloat64(collector.SearchCacheMisses)
	assert.Equal(t, float64(1), misses)
}

func TestRecordEmbeddingError(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	provider := "openai"
	errorType := "rate_limit"

	collector.RecordEmbeddingError(provider, errorType)

	count := testutil.ToFloat64(collector.EmbeddingErrorsTotal.WithLabelValues(provider, errorType))
	assert.Equal(t, float64(1), count)
}

func TestRecordVectorSearch(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	tests := []struct {
		name        string
		searchType  string
		status      string
		duration    time.Duration
		resultCount int
		wantCount   float64
	}{
		{
			name:        "successful semantic search",
			searchType:  "semantic",
			status:      "success",
			duration:    25 * time.Millisecond,
			resultCount: 10,
			wantCount:   1,
		},
		{
			name:        "successful hybrid search",
			searchType:  "hybrid",
			status:      "success",
			duration:    50 * time.Millisecond,
			resultCount: 25,
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.RecordVectorSearch(tt.searchType, tt.status, tt.duration, tt.resultCount)

			count := testutil.ToFloat64(collector.VectorSearchRequests.WithLabelValues(tt.searchType, tt.status))
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestUpdateVectorStoreSize(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	sizeBytes := int64(1024 * 1024 * 100) // 100 MB
	collector.UpdateVectorStoreSize(sizeBytes)

	size := testutil.ToFloat64(collector.VectorStoreSize)
	assert.Equal(t, float64(sizeBytes), size)
}

func TestSetSystemStartTime(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	startTime := time.Now()
	collector.SetSystemStartTime(startTime)

	value := testutil.ToFloat64(collector.SystemStartTime)
	assert.Equal(t, float64(startTime.Unix()), value)
}

func TestSetComponentHealth(t *testing.T) {
	collector, _ := newTestMetricsCollector(t)

	tests := []struct {
		name      string
		component string
		healthy   bool
		wantValue float64
	}{
		{
			name:      "healthy component",
			component: "indexer",
			healthy:   true,
			wantValue: 1.0,
		},
		{
			name:      "unhealthy component",
			component: "embedding",
			healthy:   false,
			wantValue: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.SetComponentHealth(tt.component, tt.healthy)

			value := testutil.ToFloat64(collector.SystemHealth.WithLabelValues(tt.component))
			assert.Equal(t, tt.wantValue, value)
		})
	}
}
