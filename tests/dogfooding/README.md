# Conexus Dogfooding Test Suite

This directory contains a comprehensive test suite that demonstrates Conexus's superiority over standard file search approaches using realistic developer scenarios.

## Overview

The dogfooding tests compare Conexus context search capabilities against traditional file search methods across four key areas:

1. **Code Search** - Finding function implementations, configurations, and patterns
2. **Relationship Discovery** - Understanding dependencies and cross-references  
3. **Context-Aware Search** - Leveraging work context (active files, branches, issues)
4. **Cross-Reference Queries** - Linking tickets, documentation, and code

## Directory Structure

```
tests/dogfooding/
├── scenarios/           # Test scenario definitions
│   ├── code-search-scenarios.js
│   ├── relationship-discovery-scenarios.js
│   ├── context-aware-scenarios.js
│   └── cross-reference-scenarios.js
├── benchmarks/          # Benchmark execution scripts
│   ├── dogfooding-benchmark.js     # Quality comparison benchmark
│   ├── performance-benchmark.js    # Performance measurement
│   └── analyze-results.js          # Results analysis and reporting
└── results/            # Benchmark outputs and reports
    ├── dogfooding-results.json
    ├── performance-results.json
    └── dogfooding-report.md
```

## Prerequisites

- Conexus server running on `http://localhost:8080` (or set `BASE_URL` environment variable)
- Bun runtime installed
- Server health check passes

## Running the Tests

### 1. Start Conexus Server

```bash
go build ./cmd/conexus
./conexus
```

### 2. Run Quality Benchmark

Compares Conexus vs standard search relevance and success rates:

```bash
cd tests/dogfooding/benchmarks
bun run dogfooding-benchmark.js
```

This will:
- Execute all test scenarios
- Compare results quality between Conexus and standard search
- Generate `../results/dogfooding-results.json`

### 3. Run Performance Benchmark

Measures response times and throughput:

```bash
bun run performance-benchmark.js
```

This will:
- Test individual scenario performance
- Run concurrent load test (10 concurrent requests for 30 seconds)
- Generate `../results/performance-results.json`

### 4. Analyze Results

Generate comprehensive reports:

```bash
bun run analyze-results.js
```

This will:
- Analyze benchmark results
- Generate `../results/dogfooding-report.md` with detailed findings

## Expected Results

The benchmarks should demonstrate:

- **50-80% higher relevance scores** for Conexus vs standard search
- **Sub-1000ms P95 response times** for typical queries
- **10+ req/s throughput** under concurrent load
- **<1% error rates** for healthy scenarios

## Scenario Categories

### Code Search Scenarios
Realistic queries developers make when exploring codebases:
- "function implementations for error handling"
- "database connection setup and configuration"
- "API endpoint definitions and handlers"

### Relationship Discovery Scenarios
Queries that require understanding code relationships:
- "dependencies and imports for the MCP handlers"
- "cross-references for the Agent interface"
- "error handling chain from user input to logging"

### Context-Aware Scenarios
Queries that leverage development context:
- "related code for the current file being edited"
- "changes made in the current feature branch"
- "code related to open issue about error handling"

### Cross-Reference Scenarios
Queries linking different domains:
- "code changes for ticket PERF-456"
- "implementation of user authentication requirement"
- "code examples from the API documentation"

## Customization

### Adding New Scenarios

Edit the scenario files in `scenarios/` to add new test cases:

```javascript
export const myScenarios = [
  {
    id: 'my_scenario',
    query: 'my search query',
    description: 'What this scenario tests',
    expectedTypes: ['function', 'type'],
    expectedFiles: ['*.go', '*.js'],
    context: 'When this would be used'
  }
];
```

### Modifying Benchmark Parameters

Edit the benchmark scripts to change:
- `BASE_URL`: Server endpoint
- `CONCURRENT_REQUESTS`: Load test concurrency
- `TOTAL_ITERATIONS`: Number of iterations

## Troubleshooting

### Server Connection Issues
- Ensure Conexus is running on the expected port
- Check server logs for errors
- Verify health endpoint: `curl http://localhost:8080/health`

### Low Relevance Scores
- Check that the codebase has been properly indexed
- Verify MCP tools are functioning correctly
- Review scenario expectations vs actual codebase content

### Performance Issues
- Monitor server resource usage during benchmarks
- Check for database connection issues
- Review Conexus configuration for optimization opportunities

## Integration with CI/CD

Add to your CI pipeline:

```yaml
- name: Run Dogfooding Tests
  run: |
    go build ./cmd/conexus
    ./conexus &
    sleep 5
    cd tests/dogfooding/benchmarks
    bun run dogfooding-benchmark.js
    bun run analyze-results.js
```

## Contributing

When adding new scenarios:
1. Ensure queries are realistic developer use cases
2. Include clear expected outcomes
3. Test scenarios individually before adding to suite
4. Update this README with new scenario categories if needed
