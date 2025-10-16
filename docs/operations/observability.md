# Observability Guide

This guide covers metrics collection, distributed tracing, and monitoring for Conexus.

## Table of Contents

- [Overview](#overview)
- [Configuration](#configuration)
- [Metrics](#metrics)
- [Tracing](#tracing)
- [Local Development](#local-development)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Overview

Conexus includes built-in observability features:

- **Metrics**: Prometheus-compatible metrics for monitoring system health and performance
- **Tracing**: OpenTelemetry distributed tracing for request flow analysis
- **Logging**: Structured JSON logging with configurable levels

### Architecture

```
┌─────────────┐
│   Conexus   │
│   Server    │
└──────┬──────┘
       │
       ├─────────► Prometheus (Metrics) :9091/metrics
       │
       ├─────────► OTLP Collector (Traces) :4318
       │
       └─────────► Logs (stdout/file)
```

## Configuration

### Environment Variables

```bash
# Metrics Configuration
CONEXUS_METRICS_ENABLED=true
CONEXUS_METRICS_PORT=9091
CONEXUS_METRICS_PATH=/metrics

# Tracing Configuration
CONEXUS_TRACING_ENABLED=true
CONEXUS_TRACING_ENDPOINT=http://localhost:4318
CONEXUS_TRACING_SAMPLE_RATE=0.1  # 10% sampling

# Logging Configuration
CONEXUS_LOG_LEVEL=info
CONEXUS_LOG_FORMAT=json
```

### Configuration File

Create `config.yaml`:

```yaml
server:
  host: localhost
  port: 8080

observability:
  metrics:
    enabled: true
    port: 9091
    path: /metrics
  
  tracing:
    enabled: true
    endpoint: http://localhost:4318
    sample_rate: 0.1

logging:
  level: info
  format: json
```

Load with:
```bash
conexus start --config config.yaml
```

### Default Values

| Setting | Default | Description |
|---------|---------|-------------|
| `metrics.enabled` | `false` | Enable Prometheus metrics |
| `metrics.port` | `9091` | Metrics HTTP server port |
| `metrics.path` | `/metrics` | Metrics endpoint path |
| `tracing.enabled` | `false` | Enable OpenTelemetry tracing |
| `tracing.endpoint` | `http://localhost:4318` | OTLP HTTP endpoint |
| `tracing.sample_rate` | `0.1` | Trace sampling rate (0.0-1.0) |

## Metrics

### Available Metrics

#### Request Metrics
```
# Total requests processed
conexus_requests_total{method="POST",status="200"} 1234

# Request duration histogram (seconds)
conexus_request_duration_seconds_bucket{method="POST",le="0.1"} 980
conexus_request_duration_seconds_sum{method="POST"} 123.45
conexus_request_duration_seconds_count{method="POST"} 1234
```

#### Indexing Metrics
```
# Files indexed
conexus_files_indexed_total 5678

# Indexing duration
conexus_indexing_duration_seconds 45.6

# Indexing errors
conexus_indexing_errors_total 2
```

#### Vector Store Metrics
```
# Vector operations
conexus_vector_operations_total{operation="insert"} 3456
conexus_vector_operations_total{operation="search"} 7890

# Vector search latency
conexus_vector_search_duration_seconds_bucket{le="0.05"} 7500
conexus_vector_search_duration_seconds_sum 234.56
conexus_vector_search_duration_seconds_count 7890
```

#### Agent Metrics
```
# Agent invocations
conexus_agent_invocations_total{agent="analyzer",status="success"} 456

# Agent execution duration
conexus_agent_duration_seconds_bucket{agent="analyzer",le="1.0"} 400
conexus_agent_duration_seconds_sum{agent="analyzer"} 234.5
conexus_agent_duration_seconds_count{agent="analyzer"} 456
```

### Querying Metrics

#### Using curl
```bash
curl http://localhost:9091/metrics
```

#### Using Prometheus
```promql
# Request rate (requests per second)
rate(conexus_requests_total[5m])

# 95th percentile request latency
histogram_quantile(0.95, rate(conexus_request_duration_seconds_bucket[5m]))

# Error rate
rate(conexus_requests_total{status=~"5.."}[5m]) / rate(conexus_requests_total[5m])

# Agent success rate
sum(rate(conexus_agent_invocations_total{status="success"}[5m])) / 
sum(rate(conexus_agent_invocations_total[5m]))
```

### Alerting Rules

Example Prometheus alerting rules (`alerts.yml`):

```yaml
groups:
  - name: conexus
    rules:
      - alert: HighErrorRate
        expr: |
          rate(conexus_requests_total{status=~"5.."}[5m]) / 
          rate(conexus_requests_total[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }}"

      - alert: SlowRequests
        expr: |
          histogram_quantile(0.95, 
            rate(conexus_request_duration_seconds_bucket[5m])
          ) > 2.0
        for: 5m
        annotations:
          summary: "Slow requests detected"
          description: "95th percentile latency is {{ $value }}s"

      - alert: IndexingFailures
        expr: rate(conexus_indexing_errors_total[5m]) > 0
        for: 5m
        annotations:
          summary: "Indexing failures detected"
          description: "{{ $value }} indexing errors per second"
```

## Tracing

### Span Types

Conexus creates spans for:

1. **HTTP Requests**: `/mcp/*` endpoints
2. **Agent Operations**: Analyzer, Locator invocations
3. **Indexing**: File walking, embedding generation, vector storage
4. **Search**: Vector similarity search, result ranking
5. **Database Operations**: SQLite queries, vector operations

### Span Attributes

Common attributes attached to spans:

```
# Request spans
http.method = "POST"
http.route = "/mcp/analyze"
http.status_code = 200

# Agent spans
agent.type = "analyzer"
agent.input_size = 1024
agent.output_size = 2048

# Search spans
search.query = "authentication logic"
search.limit = 10
search.results_count = 8

# Indexing spans
indexer.path = "/path/to/codebase"
indexer.files_processed = 234
indexer.duration_ms = 5678
```

### Trace Sampling

Configure sampling rate based on environment:

- **Development**: `1.0` (100%) - trace everything
- **Staging**: `0.5` (50%) - trace half of requests
- **Production**: `0.1` (10%) - trace 10% to reduce overhead

```bash
# Development
CONEXUS_TRACING_SAMPLE_RATE=1.0

# Production
CONEXUS_TRACING_SAMPLE_RATE=0.1
```

### Viewing Traces

Using Jaeger UI (see [Local Development](#local-development)):

1. Open http://localhost:16686
2. Select "conexus" service
3. Filter by operation, tags, or duration
4. Click trace to see detailed span timeline

## Local Development

### Quick Start with Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  # Prometheus - Metrics collection
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  # Grafana - Visualization
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana-dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana-datasources.yml:/etc/grafana/provisioning/datasources/datasources.yml

  # Jaeger - Distributed tracing
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # UI
      - "4318:4318"    # OTLP HTTP receiver
    environment:
      - COLLECTOR_OTLP_ENABLED=true

volumes:
  prometheus-data:
  grafana-data:
```

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'conexus'
    static_configs:
      - targets: ['host.docker.internal:9091']
```

Create `grafana-datasources.yml`:

```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

### Start Observability Stack

```bash
# Start Prometheus, Grafana, and Jaeger
docker-compose up -d

# Start Conexus with observability enabled
CONEXUS_METRICS_ENABLED=true \
CONEXUS_TRACING_ENABLED=true \
CONEXUS_TRACING_SAMPLE_RATE=1.0 \
conexus start

# Access dashboards
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000
# Jaeger: http://localhost:16686
```

### Example Grafana Dashboard

Create `grafana-dashboards/conexus.json`:

```json
{
  "dashboard": {
    "title": "Conexus Overview",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(conexus_requests_total[5m])"
          }
        ]
      },
      {
        "title": "P95 Latency",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(conexus_request_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(conexus_requests_total{status=~\"5..\"}[5m])"
          }
        ]
      }
    ]
  }
}
```

## Production Deployment

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    environment: production
    cluster: us-west-2

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

rule_files:
  - "alerts.yml"

scrape_configs:
  - job_name: 'conexus'
    static_configs:
      - targets: ['conexus-1:9091', 'conexus-2:9091', 'conexus-3:9091']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
```

### OpenTelemetry Collector

For production, use the OpenTelemetry Collector as a proxy:

```yaml
# otel-collector-config.yml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
  memory_limiter:
    check_interval: 1s
    limit_mib: 512

exporters:
  jaeger:
    endpoint: jaeger-collector:14250
  logging:
    loglevel: info

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [jaeger, logging]
```

Point Conexus to the collector:

```bash
CONEXUS_TRACING_ENDPOINT=http://otel-collector:4318
```

### Kubernetes Deployment

Example pod annotations for Prometheus scraping:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: conexus
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9091"
    prometheus.io/path: "/metrics"
spec:
  containers:
    - name: conexus
      image: conexus:latest
      env:
        - name: CONEXUS_METRICS_ENABLED
          value: "true"
        - name: CONEXUS_TRACING_ENABLED
          value: "true"
        - name: CONEXUS_TRACING_ENDPOINT
          value: "http://otel-collector:4318"
      ports:
        - containerPort: 8080
          name: http
        - containerPort: 9091
          name: metrics
```

### Security Considerations

1. **Metrics Endpoint**: Restrict access to internal network
   ```bash
   # Only bind to localhost if Prometheus runs on same host
   CONEXUS_METRICS_PORT=9091
   ```

2. **Trace Data**: May contain sensitive information
   - Use OpenTelemetry Collector to filter attributes
   - Configure retention policies in Jaeger
   - Consider trace anonymization

3. **Network Policies**: Limit who can scrape metrics
   ```yaml
   # Kubernetes NetworkPolicy
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: conexus-metrics
   spec:
     podSelector:
       matchLabels:
         app: conexus
     ingress:
       - from:
           - namespaceSelector:
               matchLabels:
                 name: monitoring
         ports:
           - protocol: TCP
             port: 9091
   ```

## Troubleshooting

### Metrics Not Appearing

**Problem**: Prometheus shows "down" or no data

**Solutions**:
1. Check metrics are enabled:
   ```bash
   curl http://localhost:9091/metrics
   # Should return Prometheus-format metrics
   ```

2. Verify port is not blocked:
   ```bash
   netstat -tlnp | grep 9091
   ```

3. Check Prometheus scrape config:
   ```bash
   # In Prometheus UI: Status → Targets
   # Should show conexus target as "UP"
   ```

4. Check Conexus logs:
   ```bash
   # Look for: "Starting metrics server on :9091"
   journalctl -u conexus -f
   ```

### Traces Not Appearing

**Problem**: Jaeger shows no traces

**Solutions**:
1. Verify tracing is enabled and endpoint is reachable:
   ```bash
   CONEXUS_TRACING_ENABLED=true
   CONEXUS_TRACING_ENDPOINT=http://localhost:4318
   curl http://localhost:4318  # Should not refuse connection
   ```

2. Check sampling rate:
   ```bash
   # Set to 100% for testing
   CONEXUS_TRACING_SAMPLE_RATE=1.0
   ```

3. Verify OTLP receiver is running:
   ```bash
   docker-compose logs jaeger
   # Look for: "Listening for HTTP traffic on :4318"
   ```

4. Check for trace export errors:
   ```bash
   # Conexus logs will show export failures
   grep "trace export" /var/log/conexus.log
   ```

### High Memory Usage

**Problem**: Prometheus or Jaeger consuming too much memory

**Solutions**:
1. Reduce metric retention:
   ```yaml
   # prometheus.yml
   storage:
     tsdb:
       retention.time: 7d  # Down from 15d default
   ```

2. Lower trace sampling rate:
   ```bash
   CONEXUS_TRACING_SAMPLE_RATE=0.01  # 1% instead of 10%
   ```

3. Configure resource limits:
   ```yaml
   # docker-compose.yml
   services:
     prometheus:
       deploy:
         resources:
           limits:
             memory: 2G
     jaeger:
       deploy:
         resources:
           limits:
             memory: 1G
   ```

### Missing Metrics

**Problem**: Some expected metrics are not present

**Possible causes**:
1. Feature not used yet (e.g., `conexus_agent_invocations_total` if no agents called)
2. Metrics disabled in code (check feature flags)
3. Scrape interval too long (increase frequency in Prometheus)

**Verify instrumentation**:
```bash
# Check what metrics are actually exposed
curl http://localhost:9091/metrics | grep conexus_
```

### Trace Export Failures

**Problem**: Logs show "failed to export traces"

**Solutions**:
1. Check network connectivity:
   ```bash
   telnet localhost 4318
   ```

2. Verify OTLP endpoint accepts HTTP:
   ```bash
   curl -X POST http://localhost:4318/v1/traces \
     -H "Content-Type: application/json" \
     -d '{"resourceSpans":[]}'
   # Should return 200 or 4xx, not connection refused
   ```

3. Check for TLS issues:
   ```bash
   # If using HTTPS endpoint
   CONEXUS_TRACING_ENDPOINT=https://collector:4318
   # Ensure certificates are valid
   ```

## Performance Impact

### Metrics Overhead

- **Memory**: ~50-100 MB for metrics storage
- **CPU**: <1% for metric collection and exposition
- **Latency**: <1ms added per request

### Tracing Overhead

Depends on sampling rate:

| Sample Rate | CPU Overhead | Network Impact |
|-------------|--------------|----------------|
| 0.01 (1%)   | <1%          | ~100 KB/s      |
| 0.1 (10%)   | ~2-3%        | ~1 MB/s        |
| 1.0 (100%)  | ~10-15%      | ~10 MB/s       |

**Recommendations**:
- **Development**: 100% sampling for full visibility
- **Production**: 1-10% sampling to balance overhead and insight

## Best Practices

1. **Start with Defaults**: Enable observability in dev first
2. **Monitor the Monitors**: Set alerts on Prometheus/Jaeger health
3. **Tune Sampling**: Adjust based on traffic volume and needs
4. **Secure Endpoints**: Never expose metrics/traces to public internet
5. **Correlate Data**: Use trace IDs in logs for cross-reference
6. **Set SLOs**: Define target latency/error rate, alert on violations
7. **Regular Review**: Check dashboards weekly, adjust alerts as needed

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/languages/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/)

---

**Next Steps**: Deploy observability stack and create custom dashboards for your metrics.
