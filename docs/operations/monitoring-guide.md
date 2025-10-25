# Monitoring & Observability Guide

## Overview

Conexus provides comprehensive observability through the **four pillars**:
- **Metrics** - Prometheus metrics for performance monitoring
- **Logs** - Structured JSON logging
- **Traces** - Distributed tracing with Jaeger
- **Errors** - Sentry error monitoring and crash reporting

This guide covers deployment, configuration, and usage of the monitoring stack.

---

## Quick Start

### 1. Deploy Observability Stack

```bash
# Start Prometheus, Grafana, Jaeger, and Conexus
docker-compose -f docker-compose.observability.yml up -d

# Verify services are running
docker-compose -f docker-compose.observability.yml ps
```

### 2. Access Dashboards

| Service | URL | Default Credentials |
|---------|-----|---------------------|
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | None |
| Jaeger | http://localhost:16686 | None |
| Conexus Metrics | http://localhost:9091/metrics | None |
| Conexus Health | http://localhost:8080/health | None |

### 3. View Metrics

1. Open Grafana at http://localhost:3000
2. Login with `admin` / `admin`
3. Navigate to **Dashboards** → **Conexus Overview**
4. View real-time metrics

---

## Architecture

```
┌─────────────┐
│   Conexus   │
│   Server    │────┐
└─────────────┘    │
                   │ Metrics (/metrics:9091)
                   ▼
              ┌──────────┐
              │Prometheus│
              │  :9090   │
              └────┬─────┘
                   │ Query
                   ▼
              ┌──────────┐
              │ Grafana  │
              │  :3000   │
              └──────────┘

┌─────────────┐
│   Conexus   │
│   Server    │────┐
└─────────────┘    │ OTLP Traces (4318)
                   ▼
              ┌──────────┐
              │  Jaeger  │
              │  :16686  │
              └──────────┘

┌─────────────┐
│   Conexus   │
│   Server    │────┐
└─────────────┘    │ Errors & Traces (HTTPS)
                   ▼
              ┌──────────┐
              │  Sentry  │
              │ sentry.io │
              └──────────┘
```

---

## Metrics

### Available Metrics

Conexus exposes **25+ Prometheus metrics** at `http://localhost:9091/metrics`:

#### MCP Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `conexus_mcp_requests_total` | Counter | Total MCP requests by method |
| `conexus_mcp_request_duration_seconds` | Histogram | MCP request latency (p50, p95, p99) |
| `conexus_mcp_errors_total` | Counter | Total MCP errors by type |

**Labels**: `method`, `error_type`

#### Indexer Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `conexus_indexer_files_indexed_total` | Counter | Files successfully indexed |
| `conexus_indexer_files_failed_total` | Counter | Files that failed indexing |
| `conexus_indexer_chunks_created_total` | Counter | Total chunks created |
| `conexus_indexer_indexing_duration_seconds` | Histogram | Time to index a file |

#### Embedding Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `conexus_embedding_cache_hits_total` | Counter | Embedding cache hits |
| `conexus_embedding_cache_misses_total` | Counter | Embedding cache misses |
| `conexus_embedding_generation_duration_seconds` | Histogram | Time to generate embeddings |
| `conexus_embedding_batch_size` | Gauge | Current batch size for embeddings |

#### Vector Store Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `conexus_vectorstore_search_duration_seconds` | Histogram | Vector search query time |
| `conexus_vectorstore_insert_duration_seconds` | Histogram | Time to insert vectors |
| `conexus_vectorstore_documents_total` | Gauge | Total documents in vector store |

#### System Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `go_goroutines` | Gauge | Number of active goroutines |
| `go_memstats_alloc_bytes` | Gauge | Allocated memory |
| `go_memstats_heap_inuse_bytes` | Gauge | Heap memory in use |
| `go_gc_duration_seconds` | Summary | GC pause duration |
| `process_cpu_seconds_total` | Counter | Total CPU time |
| `process_resident_memory_bytes` | Gauge | Resident memory size |

### Querying Metrics

#### Prometheus Queries

```promql
# MCP request rate (requests/sec)
rate(conexus_mcp_requests_total[5m])

# MCP p95 latency
histogram_quantile(0.95, rate(conexus_mcp_request_duration_seconds_bucket[5m]))

# MCP error rate
rate(conexus_mcp_errors_total[5m]) / rate(conexus_mcp_requests_total[5m])

# Files indexed per minute
rate(conexus_indexer_files_indexed_total[1m]) * 60

# Embedding cache hit rate
rate(conexus_embedding_cache_hits_total[5m]) / 
  (rate(conexus_embedding_cache_hits_total[5m]) + rate(conexus_embedding_cache_misses_total[5m]))

# Vector search p99 latency
histogram_quantile(0.99, rate(conexus_vectorstore_search_duration_seconds_bucket[5m]))

# Memory usage trend
go_memstats_alloc_bytes

# Active goroutines
go_goroutines
```

---

## Grafana Dashboards

### Conexus Overview Dashboard

The pre-configured dashboard includes **11 panels**:

#### 1. MCP Request Rate
- **Type**: Graph
- **Query**: `rate(conexus_mcp_requests_total[5m])`
- **Description**: Requests per second by method

#### 2. MCP Request Latency (p95)
- **Type**: Graph
- **Query**: `histogram_quantile(0.95, rate(conexus_mcp_request_duration_seconds_bucket[5m]))`
- **Description**: 95th percentile response time

#### 3. MCP Error Rate
- **Type**: Graph
- **Query**: `rate(conexus_mcp_errors_total[5m]) / rate(conexus_mcp_requests_total[5m])`
- **Description**: Error percentage

#### 4. Files Indexed per Minute
- **Type**: Graph
- **Query**: `rate(conexus_indexer_files_indexed_total[1m]) * 60`
- **Description**: Indexing throughput

#### 5. Total Chunks Created
- **Type**: Counter
- **Query**: `conexus_indexer_chunks_created_total`
- **Description**: Cumulative chunks

#### 6. Embedding Cache Hit Rate
- **Type**: Gauge
- **Query**: Cache hit ratio calculation
- **Description**: Percentage of cache hits

#### 7. Embedding Generation Latency
- **Type**: Graph
- **Query**: `rate(conexus_embedding_generation_duration_seconds_sum[5m]) / rate(conexus_embedding_generation_duration_seconds_count[5m])`
- **Description**: Average embedding time

#### 8. Vector Search Latency (p95)
- **Type**: Graph
- **Query**: `histogram_quantile(0.95, rate(conexus_vectorstore_search_duration_seconds_bucket[5m]))`
- **Description**: Search performance

#### 9. Active Goroutines
- **Type**: Graph
- **Query**: `go_goroutines`
- **Description**: Concurrency level

#### 10. Memory Usage
- **Type**: Graph
- **Query**: `go_memstats_alloc_bytes`, `go_memstats_heap_inuse_bytes`
- **Description**: Memory allocation

#### 11. GC Pause Duration
- **Type**: Graph
- **Query**: `rate(go_gc_duration_seconds_sum[5m])`
- **Description**: Garbage collection impact

### Creating Custom Dashboards

1. Navigate to **Dashboards** → **New Dashboard**
2. Add panel → Select visualization type
3. Enter Prometheus query
4. Configure axes, legends, thresholds
5. Save dashboard

**Example Custom Panel** (Error Rate):
```json
{
  "targets": [
    {
      "expr": "rate(conexus_mcp_errors_total[5m])",
      "legendFormat": "{{error_type}}"
    }
  ],
  "title": "MCP Errors by Type",
  "type": "graph"
}
```

---

## Distributed Tracing

### Viewing Traces

1. Open Jaeger UI at http://localhost:16686
2. Select **Service**: `conexus`
3. Choose **Operation**: e.g., `mcp.search`
4. Click **Find Traces**

### Trace Structure

Each MCP request generates a trace with spans:

```
conexus.mcp.search (100ms)
├── embedding.generate (30ms)
├── vectorstore.search (50ms)
│   ├── sqlite.query (20ms)
│   └── bm25.search (15ms)
└── result.format (5ms)
```

### Trace Metadata

Each span includes:
- **Operation name**: `mcp.search`, `indexer.process_file`, etc.
- **Duration**: Time taken
- **Tags**: `method`, `query`, `top_k`, `error`
- **Logs**: Events during span execution

### Sampling Configuration

Control trace sampling in `config.yml`:

```yaml
observability:
  tracing:
    enabled: true
    endpoint: "http://localhost:4318"
    sample_rate: 1.0  # 1.0 = 100%, 0.1 = 10%
```

**Production recommendation**: `sample_rate: 0.1` (10%)

---

## Error Monitoring with Sentry

### Overview

Conexus integrates **Sentry** for comprehensive error monitoring, crash reporting, and performance issue tracking. Sentry provides:

- **Real-time error tracking** with stack traces
- **Distributed tracing** integration
- **User context** and session data
- **Release tracking** and regression detection
- **Performance monitoring** with transaction traces

### Configuration

Enable Sentry in `config.yml`:

```yaml
observability:
  sentry:
    enabled: true
    dsn: "https://your-dsn@sentry.io/project-id"
    environment: "production"
    sample_rate: 1.0  # 1.0 = 100% of errors captured
    release: "v1.0.0"
```

**Environment Variables**:
```bash
export CONEXUS_SENTRY_ENABLED=true
export CONEXUS_SENTRY_DSN="https://your-dsn@sentry.io/project-id"
export CONEXUS_SENTRY_ENVIRONMENT="production"
export CONEXUS_SENTRY_SAMPLE_RATE=1.0
export CONEXUS_SENTRY_RELEASE="v1.0.0"
```

### What Gets Monitored

#### MCP Request Errors
- Invalid JSON-RPC requests
- Protocol errors (method not found, invalid params)
- Handler execution failures
- Tool call errors

#### Tool Execution Tracing
- Each tool call creates a Sentry transaction
- Tool parameters captured as context
- Execution errors automatically captured
- Performance data for slow tool calls

#### Context and Tags
- `mcp.method`: The MCP method called
- `mcp.request_id`: JSON-RPC request ID
- `tool.name`: Name of the tool being executed
- `service`: Always set to "conexus-mcp"

### Viewing Errors in Sentry

1. **Access Sentry Dashboard**: Login to your Sentry account
2. **Navigate to Issues**: View real-time error reports
3. **Filter by Environment**: Use `environment:production`
4. **View Traces**: Click on issues to see full transaction traces

### Error Context

Each error includes:
- **Stack trace** with file names and line numbers
- **Request context** (method, parameters)
- **User information** (if available)
- **Environment data** (Go version, OS, etc.)
- **Breadcrumbs** showing execution flow

### Performance Monitoring

Sentry automatically captures:
- **Transaction traces** for tool calls
- **Database query performance**
- **HTTP request timing**
- **Memory usage spikes**

### Alerting Integration

Configure Sentry alerts for:
- **New error patterns**
- **Error rate spikes**
- **Performance regressions**
- **Release health monitoring**

### Production Best Practices

1. **Sample Rate Tuning**:
   ```yaml
   sample_rate: 0.1  # 10% sampling in production
   ```

2. **Environment Separation**:
   ```yaml
   environment: "production"  # Separate dev/staging/prod
   ```

3. **Release Tracking**:
   ```yaml
   release: "v1.2.3"  # Tag errors by version
   ```

4. **PII Scrubbing**: Sensitive data is automatically filtered

### Troubleshooting Sentry

#### Errors Not Appearing
1. **Check DSN**: Verify DSN is correct and accessible
2. **Check Network**: Ensure outbound HTTPS to Sentry
3. **Check Configuration**: Verify `enabled: true`

#### High Error Volume
1. **Adjust Sample Rate**: Reduce `sample_rate` to 0.1
2. **Filter Known Errors**: Use Sentry's issue filtering
3. **Rate Limiting**: Configure Sentry's rate limits

#### Missing Context
1. **Check Tags**: Ensure proper tagging in code
2. **User Context**: Verify user identification logic
3. **Custom Context**: Add relevant business context

---

## Logs

### Log Format

Conexus uses **structured JSON logging**:

```json
{
  "timestamp": "2025-10-16T10:30:45Z",
  "level": "info",
  "message": "MCP request completed",
  "method": "context_search",
  "duration_ms": 42.5,
  "results": 10,
  "trace_id": "abc123..."
}
```

### Log Levels

| Level | Use Case |
|-------|----------|
| `debug` | Detailed debugging (high volume) |
| `info` | Normal operations |
| `warn` | Potential issues |
| `error` | Errors requiring attention |

### Configuration

```yaml
observability:
  log_level: "info"  # debug, info, warn, error
  log_format: "json" # json or text
```

Environment variable override:
```bash
export CONEXUS_LOG_LEVEL=debug
```

### Viewing Logs

```bash
# Docker logs
docker-compose -f docker-compose.observability.yml logs -f conexus

# Filter by level
docker logs conexus 2>&1 | jq 'select(.level == "error")'

# Search for specific trace
docker logs conexus 2>&1 | jq 'select(.trace_id == "abc123")'
```

---

## Alerting

### Creating Alert Rules

Create `observability/alerts.yml`:

```yaml
groups:
  - name: conexus_alerts
    interval: 30s
    rules:
      # High error rate
      - alert: HighMCPErrorRate
        expr: rate(conexus_mcp_errors_total[5m]) / rate(conexus_mcp_requests_total[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High MCP error rate"
          description: "Error rate is {{ $value | humanizePercentage }}"

      # Slow requests
      - alert: SlowMCPRequests
        expr: histogram_quantile(0.95, rate(conexus_mcp_request_duration_seconds_bucket[5m])) > 1.0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "MCP requests are slow"
          description: "p95 latency is {{ $value }}s"

      # High memory usage
      - alert: HighMemoryUsage
        expr: go_memstats_heap_inuse_bytes > 1e9  # 1GB
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
          description: "Heap usage is {{ $value | humanize1024 }}B"

      # Indexer stalled
      - alert: IndexerStalled
        expr: rate(conexus_indexer_files_indexed_total[10m]) == 0
        for: 15m
        labels:
          severity: critical
        annotations:
          summary: "Indexer has stalled"
          description: "No files indexed in 15 minutes"

      # Low cache hit rate
      - alert: LowCacheHitRate
        expr: |
          rate(conexus_embedding_cache_hits_total[5m]) / 
          (rate(conexus_embedding_cache_hits_total[5m]) + rate(conexus_embedding_cache_misses_total[5m])) < 0.5
        for: 10m
        labels:
          severity: info
        annotations:
          summary: "Low embedding cache hit rate"
          description: "Cache hit rate is {{ $value | humanizePercentage }}"
```

### Configure Prometheus

Add to `observability/prometheus.yml`:

```yaml
rule_files:
  - "/etc/prometheus/alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']
```

Update `docker-compose.observability.yml`:

```yaml
prometheus:
  volumes:
    - ./observability/alerts.yml:/etc/prometheus/alerts.yml:ro
```

### Notification Channels

Configure Alertmanager in `observability/alertmanager.yml`:

```yaml
global:
  slack_api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'

route:
  receiver: 'slack-notifications'
  group_by: ['alertname', 'severity']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 12h

receivers:
  - name: 'slack-notifications'
    slack_configs:
      - channel: '#alerts'
        title: 'Conexus Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

---

## Performance Tuning

### Identifying Bottlenecks

#### 1. Slow MCP Requests
```promql
# Find slowest operations
topk(5, histogram_quantile(0.95, rate(conexus_mcp_request_duration_seconds_bucket[5m])))
```

**Solutions**:
- Reduce `top_k` in search queries
- Add source type filters
- Enable embedding cache
- Optimize vector index

#### 2. High Memory Usage
```promql
# Memory trend
rate(go_memstats_alloc_bytes[5m])
```

**Solutions**:
- Reduce batch sizes
- Tune cache sizes
- Enable memory profiling (see [profiling-guide.md](../profiling-guide.md))

#### 3. Low Indexing Throughput
```promql
# Indexing rate
rate(conexus_indexer_files_indexed_total[5m])
```

**Solutions**:
- Increase worker count
- Optimize file patterns
- Add more ignore patterns
- Profile indexing (see [profiling-guide.md](../profiling-guide.md))

#### 4. Poor Cache Hit Rate
```promql
# Cache efficiency
rate(conexus_embedding_cache_hits_total[5m]) / 
  (rate(conexus_embedding_cache_hits_total[5m]) + rate(conexus_embedding_cache_misses_total[5m]))
```

**Solutions**:
- Increase cache size
- Use consistent queries
- Enable persistent cache

---

## Operational Runbooks

### Runbook: High MCP Error Rate

**Symptoms**: `HighMCPErrorRate` alert firing

**Investigation**:
1. Check error types:
   ```promql
   rate(conexus_mcp_errors_total[5m])
   ```
2. View error logs:
   ```bash
   docker logs conexus 2>&1 | jq 'select(.level == "error")'
   ```
3. Check traces in Jaeger for failed requests

**Common Causes**:
- OpenAI API rate limiting → Check API key, increase timeout
- Invalid search queries → Validate input schema
- Vector store corruption → Reindex database

### Runbook: Slow Performance

**Symptoms**: `SlowMCPRequests` alert firing

**Investigation**:
1. Check p95 latency by operation:
   ```promql
   histogram_quantile(0.95, rate(conexus_mcp_request_duration_seconds_bucket[5m])) by (method)
   ```
2. View slow traces in Jaeger
3. Check system resources:
   ```bash
   docker stats conexus
   ```

**Common Causes**:
- Large result sets → Reduce `top_k`
- Cold embeddings cache → Warm up cache
- High CPU/memory → Scale resources

### Runbook: Indexer Stalled

**Symptoms**: `IndexerStalled` alert firing

**Investigation**:
1. Check indexer status:
   ```bash
   curl http://localhost:8080/health
   ```
2. View indexer logs:
   ```bash
   docker logs conexus 2>&1 | jq 'select(.message | contains("indexer"))'
   ```
3. Check file system permissions

**Common Causes**:
- File permission errors → Fix permissions
- Out of disk space → Clean up data
- Panic in indexer → Check logs, restart

---

## Adding Custom Metrics

### 1. Define Metric in Code

```go
// internal/observability/metrics.go
type MetricsCollector struct {
    myCustomMetric prometheus.Counter
}

func NewMetricsCollector(namespace string) *MetricsCollector {
    mc := &MetricsCollector{
        myCustomMetric: prometheus.NewCounter(prometheus.CounterOpts{
            Namespace: namespace,
            Name:      "my_custom_metric_total",
            Help:      "Description of my custom metric",
        }),
    }
    
    prometheus.MustRegister(mc.myCustomMetric)
    return mc
}

func (mc *MetricsCollector) IncrementCustomMetric() {
    mc.myCustomMetric.Inc()
}
```

### 2. Use Metric

```go
// Increment metric
metrics.IncrementCustomMetric()
```

### 3. Add to Dashboard

Create new panel in Grafana:
- Query: `conexus_my_custom_metric_total`
- Visualization: Counter/Graph
- Save

---

## Troubleshooting

### Metrics Not Appearing

1. **Check endpoint**:
   ```bash
   curl http://localhost:9091/metrics | grep conexus
   ```
2. **Check Prometheus targets**:
   - Open http://localhost:9090/targets
   - Verify `conexus:9091` is UP
3. **Check configuration**:
   ```yaml
   observability:
     metrics:
       enabled: true
       port: 9091
   ```

### Grafana Dashboard Not Loading

1. **Check data source**:
   - Settings → Data Sources → Prometheus
   - Click "Test" - should show "Data source is working"
2. **Check provisioning**:
   ```bash
   docker exec conexus-grafana ls /etc/grafana/provisioning/dashboards
   ```
3. **Re-import dashboard**:
   - Upload `observability/dashboards/conexus-overview.json`

### Traces Not Appearing

1. **Check tracing enabled**:
   ```yaml
   observability:
     tracing:
       enabled: true
       endpoint: "http://jaeger:4318"
   ```
2. **Check Jaeger connection**:
   ```bash
   docker logs conexus 2>&1 | grep -i "trace"
   ```
3. **Check sample rate**: Increase to 1.0 for testing

### High Cardinality Warnings

**Symptom**: Prometheus warnings about high cardinality

**Solution**:
- Remove dynamic labels (user IDs, query strings)
- Use `trace_id` in logs instead of metrics
- Aggregate labels (e.g., `status_code` → `status_class`)

---

## Production Checklist

- [ ] Metrics enabled: `observability.metrics.enabled: true`
- [ ] Tracing sample rate tuned: `sample_rate: 0.1` (10%)
- [ ] Sentry error monitoring enabled: `observability.sentry.enabled: true`
- [ ] Sentry DSN configured securely
- [ ] Sentry sample rate tuned: `sample_rate: 0.1` (10%)
- [ ] Log level set to `info` or `warn`
- [ ] Alerting rules configured
- [ ] Notification channels tested
- [ ] Dashboards imported and tested
- [ ] Prometheus retention configured (default: 15 days)
- [ ] Grafana data backed up
- [ ] Health checks configured
- [ ] Resource limits set in Docker/K8s
- [ ] Metrics endpoint not publicly exposed
- [ ] Grafana password changed from default

---

## Next Steps

- **Add alerting**: Configure Alertmanager and notification channels
- **Custom dashboards**: Create team-specific views
- **SLO tracking**: Define and monitor Service Level Objectives
- **Cost optimization**: Reduce trace sampling in production
- **Advanced queries**: Learn PromQL for complex metrics

---

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Conexus Profiling Guide](../profiling-guide.md)

---

## Support

For monitoring issues:
- Check logs: `docker-compose logs -f conexus`
- View metrics: http://localhost:9091/metrics
- GitHub Issues: [Report monitoring issues](https://github.com/ferg-cod3s/conexus/issues)
