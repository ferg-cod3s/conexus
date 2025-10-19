# Task 8.6: Federation Performance Baseline & Observability

## Performance Baseline

### Benchmark Suite Created

Created comprehensive benchmark suite in `internal/federation/benchmark_test.go` covering:

#### Relationship Detection
- `BenchmarkDetectRelationships` - Tests relationship detection across 100, 500, and 1000 results
  - Simulates querying multiple sources with deduplication
  - Measures performance of relationship graph building

#### Graph Operations
- `BenchmarkBuildRelationshipGraph` - Tests graph construction from relationships
  - Validates efficiency of cross-source link establishment
  - Supports scalability analysis

#### Result Merging
- `BenchmarkMergerAddResults` - Tests adding results from multiple sources
  - Benchmark across 100, 500, 1000+ results
  - Validates multi-source aggregation performance

- `BenchmarkMergerMergeAndDeduplicate` - Tests end-to-end deduplication
  - Includes full merge + deduplication workflow
  - Benchmarks across varying dataset sizes

#### Utility Operations
- `BenchmarkMergerHashContent` - Content hashing for deduplication
- `BenchmarkMergerItemToString` - Item-to-string conversion performance
- `BenchmarkDetectorIsRelated` - Relationship detection between item pairs
- `BenchmarkDetectorHasSimilarContent` - String similarity calculation

### Benchmark Helper Functions

- `generateTestQueryResults()` - Creates realistic multi-source query results
- `generateTestItems()` - Generates test data with titles, content, metadata

## Expected Performance Targets

Based on federation package architecture:

| Operation | Dataset | Target | Notes |
|-----------|---------|--------|-------|
| Detect Relationships | 100 results | <10ms | Per 3 sources |
| Detect Relationships | 1000 results | <50ms | Linear complexity expected |
| Build Graph | Simple relationships | <5ms | Adjacency list construction |
| Add Results | 100 items/source | <2ms | Per source |
| Merge & Deduplicate | 300 total items | <20ms | 3 sources × 100 items |
| Hash Content | Single string | <100µs | SHA256 per item |
| Item to String | Single item | <50µs | Marshaling overhead |
| String Similarity | Mid-length strings | <10µs | Levenshtein distance |

## Observability Integration (In Progress)

### Phase 1: Structured Logging

Add structured logging to federation service:

```go
// In Service.QueryMultipleSources
log.WithFields(map[string]interface{}{
    "operation": "query_multiple_sources",
    "query": query,
    "active_connectors": len(conns),
    "timestamp": time.Now(),
}).Info("Starting federated query")
```

### Phase 2: Prometheus Metrics

Add the following metrics to federation package:

- **federation_queries_total** (counter)
  - Labels: source, status (success/error)
  - Tracks total federated queries

- **federation_query_duration_seconds** (histogram)
  - Labels: source, operation
  - Buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0]
  - Measures query latency

- **federation_duplicate_ratio** (gauge)
  - Labels: source_pair
  - Tracks deduplication effectiveness

- **federation_results_count** (gauge)
  - Labels: source
  - Tracks result cardinality

- **federation_merge_duration_seconds** (histogram)
  - Measures merge operation duration

- **federation_active_sources** (gauge)
  - Tracks number of active data sources

### Phase 3: Distributed Tracing (OpenTelemetry)

If OpenTelemetry is available:

- Create span for each multi-source query
- Add child spans for per-source queries
- Track relationships span
- Record merge operation span

### Implementation Strategy

1. **Integration Points**
   - `Service.QueryMultipleSources()` - Main query operation
   - `Detector.DetectRelationships()` - Relationship detection
   - `Merger.MergeAndDeduplicate()` - Deduplication
   - Individual connector queries in goroutines

2. **Context Propagation**
   - Use context.Context for tracing throughout
   - Ensure trace IDs propagate across async operations

3. **Error Tracking**
   - Log query failures with source and error details
   - Track partial failures vs complete failures
   - Record error rates by source

4. **Performance Monitoring**
   - Track query latency distribution
   - Monitor deduplication ratio for anomalies
   - Alert on slow queries (>1 second)

## Success Criteria for Task 8.6

### Performance
- [ ] All benchmarks compile and run successfully
- [ ] Benchmark results available in docs
- [ ] Performance within expected targets
- [ ] No regressions from Task 8.5

### Observability
- [ ] Structured logging integrated
- [ ] Prometheus metrics exported
- [ ] Trace integration complete
- [ ] Documentation updated with observability guide

### Quality
- [ ] All existing tests pass (100% pass rate)
- [ ] No new issues or regressions
- [ ] Code coverage maintained at 90%+
- [ ] Performance validated on realistic datasets

## Deliverables

### By End of Task 8.6

1. **Performance Report**
   - Benchmark results and analysis
   - Performance vs targets assessment
   - Optimization recommendations

2. **Observability Integration**
   - Structured logging implementation
   - Prometheus metrics setup
   - Optional: OpenTelemetry tracing

3. **Documentation**
   - Observability configuration guide
   - Metrics interpretation guide
   - Troubleshooting guide

4. **Task Completion Report**
   - Task 8.6 final summary
   - Sign-off on all deliverables
   - Metrics and KPIs

## Current Status

- ✅ Benchmarks created: 8 benchmark functions
- ✅ Benchmarks committed: feat(federation): add performance benchmarks
- ⏳ Run benchmarks and collect baseline data
- ⏳ Implement observability
- ⏳ Create performance report
- ⏳ Sign-off Task 8.6

