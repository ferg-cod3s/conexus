# Task 7.6.2: MCP Tool Validation with Real-World Data

**Estimated Time**: 1.5-2 hours  
**Status**: IN PROGRESS  
**Started**: 2025-10-16

## Objectives

1. **Real-World Data Testing**: Test MCP tools with actual codebase (not mocks)
2. **Multi-Step Workflows**: Validate tool chaining and realistic usage patterns
3. **Edge Cases**: Test large result sets, complex queries, error handling
4. **Documentation Verification**: Ensure guides match actual behavior

## Approach

### Phase 1: Setup Test Environment (15 min)
- Use Conexus codebase itself as test corpus
- Index real Go files (~50 files)
- Verify indexing completes successfully
- Validate index statistics

### Phase 2: Tool-by-Tool Validation (45 min)

#### 2.1 `search_codebase` Tool (15 min)
- Test BM25 search with realistic queries
- Validate ranking and relevance
- Test edge cases (empty results, large result sets)
- Verify output format matches documentation

#### 2.2 `locate_relevant_files` Tool (10 min)
- Test file discovery with real paths
- Validate relevance scoring
- Test with different query types
- Verify file metadata accuracy

#### 2.3 `analyze_implementation` Tool (15 min)
- Test evidence extraction on real functions
- Validate control flow analysis
- Test with different file types
- Verify AGENT_OUTPUT_V1 format

#### 2.4 `index_control` Tool (5 min)
- Test status endpoint with real index
- Verify metrics accuracy
- Test error handling

### Phase 3: Multi-Step Workflows (20 min)

#### Workflow 1: Search → Analyze
1. Search for "Handle" in codebase
2. Select top result
3. Analyze implementation
4. Verify evidence chain

#### Workflow 2: Locate → Analyze
1. Locate files related to "protocol"
2. Select relevant file
3. Analyze key function
4. Verify results

#### Workflow 3: Index → Search → Locate
1. Check index status
2. Search for specific term
3. Locate related files
4. Verify consistency

### Phase 4: Documentation Verification (10 min)
- Compare actual behavior with MCP integration guide
- Identify any discrepancies
- Update documentation if needed

## Success Criteria

- ✅ All 4 tools work with real codebase data
- ✅ Search returns relevant results
- ✅ Analysis extracts correct evidence
- ✅ Multi-step workflows complete successfully
- ✅ Edge cases handled gracefully
- ✅ Documentation matches implementation

## Expected Deliverables

1. Validation test results
2. Any bug fixes needed
3. Documentation updates (if needed)
4. Completion report

## Commands

```bash
# Build and run server
go build ./cmd/conexus
./conexus &
SERVER_PID=$!

# Run validation tests
go test -v ./internal/testing/integration -run RealWorld

# Cleanup
kill $SERVER_PID
```

Let's begin!
