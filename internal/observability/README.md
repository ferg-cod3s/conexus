# Observability Package

This package provides comprehensive observability capabilities for Conexus, including:

- **Prometheus metrics** - Performance and business metrics
- **OpenTelemetry tracing** - Distributed tracing for request flows
- **Structured logging** - JSON logs with context correlation

## Components

### Metrics (`metrics.go`)

Prometheus metrics for monitoring system health and performance.

**Metric Categories:**
- MCP request metrics (requests, duration, errors, in-flight)
- Indexer metrics (operations, duration, files, chunks, errors)
- Embedding metrics (requests, duration, cache hits/misses, errors)
- Vector store metrics (searches, duration, results, size)
- System metrics (start time, health status)

**Usage:**
```go
metrics := observability.NewMetricsCollector("conexus")

// Record MCP request
start := time.Now()
metrics.TrackMCPInFlight("tools/list", 1)
defer metrics.TrackMCPInFlight("tools/list", -1)

// ... handle request ...

metrics.RecordMCPRequest("tools/list", "success", time.Since(start))
```

### Structured Logging (`logger.go`)

Context-aware structured logging with JSON output.

**Features:**
- Multiple log levels (debug, info, warn, error)
- JSON or text format
- Context extraction (trace ID, request ID, user ID)
- Specialized logging methods for MCP, indexer, embedding, search

**Usage:**
```go
logger := observability.NewLogger(observability.LoggerConfig{
    Level:     "info",
    Format:    "json",
    AddSource: true,
})

// Basic logging
logger.Info("server started", "port", 8080)

// Context-aware logging
ctx = context.WithValue(ctx, observability.TraceIDKey, traceID)
logger.InfoContext(ctx, "processing request", "method", "tools/list")

// Specialized logging
logger.LogMCPRequest(ctx, "tools/list", params, duration)
```

### OpenTelemetry Tracing (`tracing.go`)

Distributed tracing for end-to-end request visibility.

**Features:**
- OTLP exporter for sending traces to collectors
- Configurable sampling rates
- Built-in instrumentation helpers
- Trace and span ID extraction

**Usage:**
```go
tracer, err := observability.NewTracerProvider(observability.TracerConfig{
    ServiceName:    "conexus",
    ServiceVersion: "0.1.0",
    Environment:    "production",
    OTLPEndpoint:   "localhost:4317",
    SamplingRate:   1.0,
    Enabled:        true,
})
if err != nil {
    log.Fatal(err)
}
defer tracer.Shutdown(context.Background())

// Instrument operations
ctx, span := observability.InstrumentMCPRequest(ctx, tracer.Tracer(), "tools/list")
defer span.End()

// Add attributes
observability.SetSpanAttributes(ctx,
    attribute.String("param", "value"),
)

// Record errors
if err != nil {
    observability.SetSpanError(ctx, err)
}
```

## Integration Example

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/ferg-cod3s/conexus/internal/observability"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Initialize observability
    metrics := observability.NewMetricsCollector("conexus")
    logger := observability.NewLogger(observability.DefaultLoggerConfig())
    tracer, _ := observability.NewTracerProvider(observability.DefaultTracerConfig())
    defer tracer.Shutdown(context.Background())

    metrics.SetSystemStartTime(time.Now())
    metrics.SetComponentHealth("server", true)

    // Expose metrics
    http.Handle("/metrics", promhttp.Handler())

    // Example MCP handler
    http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        start := time.Now()
        method := "tools/list"

        // Trace the request
        ctx, span := observability.InstrumentMCPRequest(ctx, tracer.Tracer(), method)
        defer span.End()

        // Track in-flight
        metrics.TrackMCPInFlight(method, 1)
        defer metrics.TrackMCPInFlight(method, -1)

        // Log request
        logger.InfoContext(ctx, "handling MCP request", "method", method)

        // ... handle request ...

        // Record metrics
        duration := time.Since(start)
        metrics.RecordMCPRequest(method, "success", duration)
        logger.LogMCPResponse(ctx, method, true, duration)
    })

    logger.Info("server starting", "port", 8080)
    http.ListenAndServe(":8080", nil)
}
```

## Prometheus Metrics Reference

### MCP Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `conexus_mcp_requests_total` | Counter | method, status | Total MCP requests |
| `conexus_mcp_request_duration_seconds` | Histogram | method | MCP request duration |
| `conexus_mcp_requests_in_flight` | Gauge | method | In-flight MCP requests |
| `conexus_mcp_errors_total` | Counter | method, error_type | MCP errors |

### Indexer Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `conexus_indexer_operations_total` | Counter | operation, status | Indexer operations |
| `conexus_indexer_operation_duration_seconds` | Histogram | operation | Indexer duration |
| `conexus_indexed_files_total` | Counter | - | Files indexed |
| `conexus_indexed_chunks_total` | Counter | - | Chunks indexed |
| `conexus_indexer_errors_total` | Counter | error_type | Indexer errors |

### Embedding Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `conexus_embedding_requests_total` | Counter | provider, status | Embedding requests |
| `conexus_embedding_duration_seconds` | Histogram | provider | Embedding duration |
| `conexus_embedding_cache_hits_total` | Counter | - | Cache hits |
| `conexus_embedding_cache_misses_total` | Counter | - | Cache misses |
| `conexus_embedding_errors_total` | Counter | provider, error_type | Embedding errors |

### Vector Search Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `conexus_vector_search_requests_total` | Counter | search_type, status | Search requests |
| `conexus_vector_search_duration_seconds` | Histogram | search_type | Search duration |
| `conexus_vector_search_results_count` | Histogram | search_type | Result count |
| `conexus_vector_store_size_bytes` | Gauge | - | Store size |

### System Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `conexus_system_start_time_seconds` | Gauge | - | System start time |
| `conexus_system_health_status` | Gauge | component | Health status (1=healthy, 0=unhealthy) |

## Configuration

### Environment Variables

```bash
# Logging
LOG_LEVEL=info           # debug, info, warn, error
LOG_FORMAT=json          # json, text

# Tracing
OTEL_ENABLED=true
OTEL_ENDPOINT=localhost:4317
OTEL_SAMPLING_RATE=1.0   # 0.0 to 1.0

# Metrics
METRICS_NAMESPACE=conexus
```

## Docker Compose Integration

See `docker-compose.observability.yml` for a complete observability stack with:
- Prometheus (metrics collection)
- Jaeger (distributed tracing)
- Grafana (dashboards)

## Performance Impact

- **Metrics**: Minimal overhead (~1-2% CPU, <10MB memory)
- **Logging**: JSON serialization adds ~5-10μs per log entry
- **Tracing**: ~10-20μs per span with 1.0 sampling
  - Recommended: 0.1 sampling rate in production (10% of traces)

## Best Practices

1. **Use context everywhere** - Pass context.Context for correlation
2. **Set appropriate log levels** - Use debug sparingly in production
3. **Sample traces in production** - Start with 10% (0.1) sampling
4. **Monitor metric cardinality** - Avoid high-cardinality labels
5. **Add business metrics** - Not just technical metrics
6. **Use structured logging** - No string concatenation in logs
7. **Close spans explicitly** - Use defer span.End()

## Testing

Run the test suite:
```bash
go test ./internal/observability/...
```

## Dependencies

- `github.com/prometheus/client_golang` - Prometheus metrics
- `go.opentelemetry.io/otel` - OpenTelemetry tracing
- `log/slog` - Structured logging (Go 1.21+)
