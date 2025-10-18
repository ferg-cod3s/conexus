# Task 8.6: Federation Integration & Observability - PLAN

**Status**: 🚀 READY FOR NEXT PHASE  
**Date**: 2025-10-17  
**Branch**: mvp  
**Previous Commit**: eb1e581

## Executive Summary

Federation layer is **fully integrated** with MCP handlers. The `context.search` tool now uses the federation service to search across multiple data sources with intelligent result merging, deduplication, and relationship detection.

**Current State**: ✅ All tests passing (41 federation + 30+ MCP handler tests)  
**Build Status**: ✅ Clean build with no warnings  
**Integration Status**: ✅ MCP handlers calling federation service successfully

## What We Accomplished This Session

### 1. Verified Federation Integration ✅
- Confirmed `internal/mcp/server.go` initializes federation service (line 56)
- Verified `internal/mcp/handlers.go` uses federation in `handleContextSearch` (line 66)
- Confirmed federation response is properly converted to MCP response format (lines 84-108)
- All federation tests passing: 41/41 test cases

### 2. Fixed Build Issues ✅
- Moved debug test files from root to `tests/debug/` directory
- Resolved multiple `main` function redeclaration errors
- Build now completes cleanly without warnings
- Successfully built `cmd/conexus` binary

### 3. Test Coverage Verification ✅
- **Federation tests**: 41 test cases, 100% pass rate
  - Detector tests: 16 cases (relationship detection)
  - Merger tests: 9 cases (deduplication, diversity, scoring)
  - Service tests: 5 cases (parallel execution, pagination, timeout)
- **MCP handler tests**: 30+ test cases covering search operations
- No regressions or failures

## Current Architecture

```
┌─────────────────────────────────────────┐
│         MCP Protocol Layer              │
│      internal/mcp/server.go             │
└──────────────┬──────────────────────────┘
               │
               │ handles "tools/call"
               ▼
┌─────────────────────────────────────────┐
│     MCP Handler Layer                   │
│   internal/mcp/handlers.go              │
│  - handleContextSearch (line 23)        │
│  - handleGetRelatedInfo (line 112)      │
│  - handleIndexControl (line 213)        │
│  - handleConnectorManagement (line 268) │
└──────────────┬──────────────────────────┘
               │
               │ uses
               ▼
┌─────────────────────────────────────────┐
│      Federation Service Layer           │
│  internal/federation/service.go (333)   │
│  - Parallel connector execution         │
│  - Timeout management                   │
│  - Result merging coordination          │
└──────────────┬──────────────────────────┘
               │
      ┌────────┼────────┬────────┐
      │        │        │        │
      ▼        ▼        ▼        ▼
┌──────────┐┌──────────┐┌──────────┐┌──────────┐
│ Merger   ││Detector  ││Connector ││Connector │
│Service   ││Service   ││Manager   ││Results   │
└──────────┘└──────────┘└──────────┘└──────────┘
```

## Key Integration Points

### 1. MCP Server Initialization (server.go:56)
```go
federationSvc := federation.NewService(connectorManager, vectorStore)
```

### 2. Search Handler Integration (handlers.go:66)
```go
federationResponse, err := s.federationSvc.Search(ctx, &req, s.embedder)
```

### 3. Response Conversion (handlers.go:84-108)
```go
return schema.SearchResponse{
    Results:    searchResults,
    TotalCount: federationResponse.TotalCount,
    QueryTime:  federationResponse.QueryTime,
    Offset:     federationResponse.Offset,
    Limit:      federationResponse.Limit,
    HasMore:    federationResponse.HasMore,
}
```

## Next Phase: Task 8.6 (Coming Next)

### Priority 1: Observability & Metrics ⭐
**Goal**: Track federation performance, identify bottlenecks, monitor connector health

**Tasks**:
1. Add federation-specific metrics
   - Parallel execution time per connector
   - Result merge/dedup latency
   - Score normalization impact
   - Connector success/failure rates

2. Add logging context
   - Federation request ID tracing
   - Per-connector execution time
   - Relationship detection stats
   - Cache hit/miss rates

3. Integration with observability
   - Metrics collection in MetricsCollector
   - Error tracking for connector failures
   - Performance dashboards

**Files to Create/Modify**:
- `internal/federation/metrics.go` (new)
- `internal/observability/federation_metrics.go` (new)
- Update `internal/mcp/server.go` to wire metrics

### Priority 2: Federation-Aware Caching ⭐
**Goal**: Cache federation results while maintaining correctness across multiple sources

**Tasks**:
1. Extend SearchCache to be federation-aware
   - Cache key includes all connector IDs
   - TTL based on slowest connector
   - Invalidate on connector config changes

2. Implement smart invalidation
   - Per-connector result caching
   - Merged results caching
   - Cache warming strategies

3. Performance benchmarking
   - Measure cache effectiveness
   - Compare with/without federation caching

**Files to Create/Modify**:
- `internal/search/federation_cache.go` (new)
- Update `internal/federation/service.go` to use cache
- Update `internal/mcp/server.go` to initialize federation cache

### Priority 3: GitHub Connector for Federation
**Goal**: Enable searching GitHub issues/PRs as a federation source

**Tasks**:
1. Create GitHub connector implementation
   - GitHub API integration
   - Token-based authentication
   - Rate limiting handling
   - Context-aware search queries

2. Implement search interface
   - Map GitHub issues to SearchResultItem
   - Include PR relationships
   - Handle complex GitHub queries

3. Add connector lifecycle hooks
   - Validate credentials
   - Monitor API quota
   - Handle connection failures

**Files to Create/Modify**:
- `internal/connectors/github/connector.go` (new)
- `internal/connectors/github/search.go` (new)
- `internal/connectors/github/auth.go` (new)
- `internal/connectors/manager.go` (update)

### Priority 4: Comprehensive E2E Integration Tests
**Goal**: Ensure federation works correctly end-to-end with MCP

**Tasks**:
1. Create integration test suite
   - Multi-source search scenario
   - Relationship detection validation
   - Score normalization verification
   - Pagination correctness

2. Add stress tests
   - High-volume result merging
   - Timeout handling under load
   - Cache effectiveness under load

3. Create test fixtures
   - Multiple mock connectors
   - Realistic result sets
   - Complex relationship scenarios

**Files to Create/Modify**:
- `tests/integration/federation_e2e_test.go` (new)
- `tests/integration/mcp_federation_test.go` (new)
- `tests/fixtures/federation_*.go` (new)

## Implementation Sequence

```
Week 1:
├─ Add federation metrics & logging
├─ Implement federation-aware caching  
└─ Create GitHub connector foundation

Week 2:
├─ Complete GitHub connector implementation
├─ Write E2E integration tests
└─ Performance benchmarking

Week 3:
├─ Optimize based on metrics
├─ Cache warming strategies
└─ Documentation & examples
```

## Success Criteria

✅ **Metrics**: Can track federation operation performance  
✅ **Caching**: Federation results cached with correct invalidation  
✅ **GitHub**: Can search GitHub as federation source  
✅ **Tests**: All E2E tests passing with >80% code coverage  
✅ **Performance**: Federation queries <2s p95 latency  

## Testing Plan

### Unit Tests
- Metrics collection correctness
- Cache invalidation logic
- GitHub connector search mapping
- GitHub rate limit handling

### Integration Tests
- MCP → Federation → GitHub connector flow
- Multi-source federation with GitHub
- Cache invalidation on connector changes
- Error scenarios (GitHub API failures, timeouts)

### Performance Tests
- Federation search latency distribution
- Cache hit rates under realistic load
- GitHub API rate limit impact
- Memory usage with large result sets

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| GitHub rate limiting | High | High | Implement token rotation, backoff strategy |
| Federation latency regression | Medium | High | Add baseline benchmarks, SLO monitoring |
| Cache invalidation bugs | Medium | Medium | Comprehensive test coverage, assertions |
| Integration complexity | Medium | Medium | Incremental integration, feature flags |

## Documentation Needs

1. **Architecture Document**: Federation system design and data flow
2. **Integration Guide**: How to add new connectors to federation
3. **Operations Guide**: Monitoring federation metrics, debugging issues
4. **API Guide**: MCP tool usage with federation examples

## Files to Track

**Modified**:
- `internal/mcp/server.go` - Federation initialization
- `internal/mcp/handlers.go` - Federation integration

**New**:
- `internal/federation/metrics.go`
- `internal/observability/federation_metrics.go`
- `internal/search/federation_cache.go`
- `internal/connectors/github/*.go`
- `tests/integration/federation_e2e_test.go`
- `tests/integration/mcp_federation_test.go`

## Rollout Strategy

1. **Phase 1**: Metrics & caching in feature-flagged mode
2. **Phase 2**: GitHub connector with limited availability
3. **Phase 3**: Full federation roll-out with monitoring
4. **Phase 4**: Optimization and scale testing

## Decision Log

**Decisions Made**:
1. ✅ Keep federation layer separate from connectors for clarity
2. ✅ Use parallel execution with timeout management
3. ✅ Implement score normalization post-ranking
4. ✅ Store pagination state in federation service

**Open Decisions**:
- Cache TTL strategy (per-connector vs global?)
- GitHub connector auth model (tokens vs OAuth?)
- Rate limiting strategy (queue vs backoff?)

---

**Next Session**: Start implementing federation metrics and caching layer
