# Conexus Load Testing Suite

Comprehensive load testing for the Conexus MCP Server using [k6](https://k6.io/).

## Overview

This test suite validates system performance, stability, and resilience under various load conditions:

- **Smoke Test** - Quick validation (1-5 VUs, 1 minute)
- **Load Test** - Target performance (100-150 VUs, 14 minutes)
- **Stress Test** - Breaking point discovery (50-500 VUs, 20 minutes)
- **Soak Test** - Long-term stability (75 VUs, 20+ minutes)
- **Spike Test** - Traffic burst resilience (50-600 VUs, 13 minutes)

## Prerequisites

### Install k6

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
# Debian/Ubuntu
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Docker:**
```bash
docker pull grafana/k6:latest
```

### Start Conexus Server

```bash
# Build and start server
docker-compose up -d

# Or run locally
go build ./cmd/conexus
./conexus
```

Verify server is running:
```bash
curl http://localhost:8080/health
```

### Create Results Directory

```bash
mkdir -p tests/load/results
```

## Running Tests

### Quick Start - Smoke Test

Validate basic functionality (recommended first step):

```bash
k6 run tests/load/smoke-test.js
```

Expected output: All checks pass, p95 < 500ms, no errors.

### Load Test - Target Performance

Test system under realistic load:

```bash
k6 run tests/load/load-test.js
```

**Parameters:**
- Duration: ~14 minutes
- Peak VUs: 150
- Expected requests: 8,000-12,000
- Target: p95 < 1s, p99 < 2s, errors < 1%

**What it tests:**
- 60% context.search operations
- 20% get_related_info operations
- 15% index_control operations
- 5% connector_management operations

### Stress Test - Find Breaking Points

Progressively increase load to find system limits:

```bash
k6 run tests/load/stress-test.js
```

**Parameters:**
- Duration: ~20 minutes
- Load progression: 50 → 100 → 200 → 300 → 400 → 500 VUs
- Expected: System should handle 200+ VUs gracefully

**What to observe:**
- At what VU count does p95 exceed 1s?
- At what VU count does error rate exceed 1%?
- Does system recover when load decreases?
- What resources are exhausted?

### Soak Test - Long-Term Stability

Test for memory leaks and degradation:

```bash
k6 run tests/load/soak-test.js
```

**Parameters:**
- Duration: 20-30 minutes (configurable)
- Sustained VUs: 75
- Target: No degradation over time

**What to monitor:**
- Consistent response times throughout test
- Stable memory usage (no leaks)
- No connection leaks
- No goroutine leaks

**Custom duration:**
```bash
k6 run --duration 30m tests/load/soak-test.js
```

### Spike Test - Traffic Burst Resilience

Test sudden traffic spikes:

```bash
k6 run tests/load/spike-test.js
```

**Parameters:**
- Duration: ~13 minutes
- Spike pattern: 50 → 200 → 50 → 400 → 50 → 600 → 50
- Target: Graceful degradation, fast recovery

**What to observe:**
- Does system return 503 (graceful) vs crash?
- How quickly does system recover?
- Are baseline metrics restored?

## Saving Results

### JSON Output

Save detailed results to JSON:

```bash
k6 run --out json=tests/load/results/load-test.json tests/load/load-test.js
```

### CSV Output

Save metrics to CSV:

```bash
k6 run --out csv=tests/load/results/load-test.csv tests/load/load-test.js
```

### Multiple Outputs

Combine multiple output formats:

```bash
k6 run \
  --out json=tests/load/results/load-test.json \
  --out csv=tests/load/results/load-test.csv \
  tests/load/load-test.js
```

## Monitoring During Tests

### Prometheus Metrics

View real-time metrics:
```bash
curl http://localhost:9090/metrics
```

### Grafana Dashboards

If observability stack is running:
```bash
docker-compose -f docker-compose.observability.yml up -d
```

Open Grafana: http://localhost:3000

### k6 Cloud (Optional)

Stream results to k6 Cloud for advanced analysis:

```bash
k6 login cloud --token YOUR_TOKEN
k6 run --out cloud tests/load/load-test.js
```

## Interpreting Results

### Key Metrics

**Response Time Metrics:**
- `http_req_duration` - Total request duration
- `p(95)` - 95th percentile (target: < 1s)
- `p(99)` - 99th percentile (target: < 2s)
- `min/max` - Range of response times

**Throughput Metrics:**
- `http_reqs` - Total requests per second
- `iterations` - Virtual user iterations per second

**Error Metrics:**
- `http_req_failed` - Percentage of failed HTTP requests
- `error_rate` - Custom error rate metric
- `checks` - Percentage of passed checks

### Success Criteria

**Smoke Test:**
- ✅ All checks pass (100%)
- ✅ p95 < 500ms
- ✅ No errors

**Load Test:**
- ✅ p95 < 1s at 100 VUs
- ✅ p99 < 2s at 100 VUs
- ✅ Error rate < 1%
- ✅ Can sustain 150 VUs

**Stress Test:**
- ✅ Handles 200+ VUs with p95 < 2s
- ✅ Graceful degradation (503s, not crashes)
- ✅ System recovers when load decreases

**Soak Test:**
- ✅ No performance degradation over time
- ✅ Memory usage stable (no leaks)
- ✅ Error rate remains < 1%
- ✅ First 5min vs last 5min metrics similar

**Spike Test:**
- ✅ Survives sudden spikes without crashing
- ✅ Returns to baseline within 1-2 minutes
- ✅ Error rate < 5% during spikes

### Reading k6 Output

Example output:
```
     ✓ search: status 200
     ✓ search: has result

     checks.........................: 99.85% ✓ 15978   ✗ 24
     data_received..................: 45 MB  56 kB/s
     data_sent......................: 8.9 MB 11 kB/s
     http_req_duration..............: avg=245ms min=12ms med=198ms max=2.1s p(95)=612ms p(99)=989ms
     http_reqs......................: 16002  20.0/s
     iterations.....................: 16002  20.0/s
     vus............................: 100    min=1    max=150
```

**Interpretation:**
- ✅ 99.85% checks passed (15,978 successes, 24 failures)
- ✅ p95 = 612ms (under 1s target)
- ✅ p99 = 989ms (under 2s target)
- ✅ 20 req/s throughput with 100-150 VUs
- ⚠️ Max response time 2.1s indicates some slow requests

## Troubleshooting

### Server Not Responding

```bash
# Check if server is running
curl http://localhost:8080/health

# Check server logs
docker-compose logs conexus

# Restart server
docker-compose restart conexus
```

### High Error Rates

**Possible causes:**
1. Server overloaded - reduce VUs
2. Database locked - check SQLite configuration
3. Memory exhausted - check system resources
4. Too many open connections - check connection limits

**Solutions:**
```bash
# Reduce load
k6 run --vus 50 --duration 2m tests/load/load-test.js

# Check system resources
docker stats
top -p $(pgrep conexus)

# Check metrics
curl http://localhost:9090/metrics | grep go_goroutines
```

### Slow Response Times

**Possible causes:**
1. Cold start - embedding model loading
2. Disk I/O - SQLite performance
3. CPU saturation
4. Memory pressure

**Solutions:**
1. Run smoke test first (warm-up)
2. Use SSD for database
3. Tune SQLite pragmas (WAL mode, cache size)
4. Monitor system metrics during test

### k6 Connection Errors

**Error:** `dial tcp: too many open files`

**Solution:** Increase file descriptor limit:
```bash
ulimit -n 10000
```

**Error:** `context deadline exceeded`

**Solution:** Server too slow - reduce load or increase timeout:
```bash
k6 run --http-debug=full tests/load/load-test.js
```

## Advanced Usage

### Custom Configuration

Override defaults via environment variables:

```bash
# Custom server URL
BASE_URL=http://localhost:9000 k6 run tests/load/load-test.js

# Custom results directory
RESULTS_DIR=/tmp/load-results k6 run tests/load/load-test.js

# Custom soak duration
SOAK_DURATION=30m k6 run tests/load/soak-test.js
```

### Running in Docker

```bash
docker run --rm -i \
  --network=host \
  grafana/k6:latest \
  run - < tests/load/load-test.js
```

### Continuous Integration

GitHub Actions example:

```yaml
- name: Run Load Tests
  run: |
    docker-compose up -d
    sleep 10
    k6 run --out json=results.json tests/load/smoke-test.js
    
- name: Validate Results
  run: |
    jq '.metrics.http_req_duration."p(95)" < 1000' results.json
```

## Best Practices

1. **Always run smoke test first** - validates basic functionality
2. **Warm up the server** - run small load before main test
3. **Monitor during tests** - watch metrics, logs, system resources
4. **Run tests multiple times** - ensure consistent results
5. **Test with realistic data** - populate database before testing
6. **Baseline before changes** - compare results before/after changes
7. **Document findings** - record bottlenecks and optimizations

## Performance Optimization Tips

Based on test results, consider:

1. **SQLite Tuning:**
   - Enable WAL mode: `PRAGMA journal_mode=WAL;`
   - Increase cache: `PRAGMA cache_size=10000;`
   - Tune page size: `PRAGMA page_size=4096;`

2. **Connection Pooling:**
   - Set max open connections: `db.SetMaxOpenConns(100)`
   - Set max idle connections: `db.SetMaxIdleConns(10)`
   - Set connection lifetime: `db.SetConnMaxLifetime(time.Hour)`

3. **Caching:**
   - Cache embedding results
   - Cache search results (with TTL)
   - Use Redis for distributed caching

4. **Concurrency:**
   - Tune worker pools
   - Use goroutine limits
   - Implement request queueing

5. **Resource Limits:**
   - Set memory limits
   - Set CPU limits
   - Implement rate limiting

## Next Steps

After completing load tests:

1. ✅ Analyze results and identify bottlenecks
2. ✅ Document performance characteristics
3. ✅ Implement optimizations if needed
4. ✅ Re-run tests to validate improvements
5. ✅ Document system limits and capacity
6. ✅ Create deployment guide (Task 7.6)

## References

- [k6 Documentation](https://k6.io/docs/)
- [k6 Best Practices](https://k6.io/docs/misc/fine-tuning-os/)
- [HTTP Load Testing Guide](https://k6.io/docs/test-types/load-testing/)
- [Prometheus Metrics](https://prometheus.io/docs/concepts/metric_types/)
