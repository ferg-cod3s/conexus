# Phase 8: MCP Protocol Completeness & Feature Enhancement

**Status**: 📋 PLANNED  
**Start Date**: October 17, 2025  
**Estimated Duration**: 28-38 hours (7-10 days)  
**Theme**: Complete MCP protocol implementation and enhance core functionality

---

## Executive Summary

Phase 8 builds on the v0.1.0-mvp release by completing the MCP protocol implementation and enhancing core features. This phase focuses on filling gaps in MCP tool coverage, improving indexer intelligence, and ensuring production-grade testing and documentation.

### Strategic Goals
1. **MCP Protocol Completeness**: Implement missing MCP tools for full protocol compliance
2. **Intelligent Indexing**: Replace single-chunk files with semantic code-aware chunking
3. **Data Persistence**: Add connector management with SQLite persistence
4. **Quality Assurance**: Comprehensive testing and documentation for all new features

### Success Criteria
- ✅ 3 new MCP tools fully implemented and tested
- ✅ Code-aware chunking improves search relevance by 30%+
- ✅ Connector CRUD operations persist across restarts
- ✅ 100% test coverage on new code
- ✅ Updated documentation for all new features

---

## Task Breakdown

### Priority 1: Core MCP Tools (HIGH - 14-18 hours)

These tasks complete the MCP protocol implementation with essential tools for context retrieval and resource management.

#### **Task 8.1: Implement `context.get_related_info`** (5-7 hours)
**GitHub Issue**: #56  
**Priority**: 🔴 HIGH  
**Blocked By**: None  
**Blocks**: Task 8.7 (testing)

**Objective**: Implement the `context.get_related_info` MCP tool to find information related to user's active file or ticket.

**Current Status**:
- ✅ Handler exists in `internal/mcp/handlers.go`
- ❌ Returns placeholder response: `{"status": "not_implemented"}`
- ❌ No vectorstore integration

**Requirements**:

1. **File Path Flow**:
   - Input: `file_path` (string)
   - Query vectorstore for chunks matching file path
   - Include cross-file references via symbol heuristics
   - Find related tests/docs via filename patterns
   - Optional: Include git commit history

2. **Ticket ID Flow**:
   - Input: `ticket_id` (string)
   - Map ticket_id to branch/commit/PR messages
   - Return linked files from related commits
   - Include PR descriptions and comments

**API Contract**:
```go
type GetRelatedInfoRequest struct {
    FilePath string `json:"file_path,omitempty"`
    TicketID string `json:"ticket_id,omitempty"`
}

type RelatedItem struct {
    FilePath     string  `json:"file_path"`
    Lines        string  `json:"lines"`
    Score        float32 `json:"score"`
    RelationType string  `json:"relation_type"` // "symbol_ref", "test_file", "commit_history"
    Content      string  `json:"content,omitempty"`
}

type GetRelatedInfoResponse struct {
    RelatedItems []RelatedItem `json:"related_items"`
    Status       string        `json:"status"`
    Message      string        `json:"message"`
}
```

**Implementation Steps**:
1. Create `internal/mcp/related_info.go` with core logic
2. Integrate with vectorstore metadata queries
3. Implement file path relation detection (tests, docs, symbols)
4. Add optional git integration for ticket flow
5. Implement path security validation
6. Add comprehensive error handling

**Testing**:
- Unit tests for both file path and ticket ID flows
- Integration tests with real codebase data
- Edge cases: non-existent files, invalid ticket IDs
- Performance: <500ms response time

**Acceptance Criteria**:
- [ ] File path flow returns 5+ relevant related files
- [ ] Ticket ID flow maps to commits and returns content
- [ ] All file paths validated with `security.SafePath()`
- [ ] Response time <500ms for typical queries
- [ ] Error handling for all invalid inputs
- [ ] 90%+ unit test coverage

**Estimated Time**: 5-7 hours

---

#### **Task 8.2: Complete `context.connector_management` CRUD** (4-5 hours)
**GitHub Issue**: #57  
**Priority**: 🔴 HIGH  
**Blocked By**: None  
**Blocks**: Task 8.7 (testing)

**Objective**: Finish the `context.connector_management` tool with full CRUD operations and SQLite persistence.

**Current Status**:
- ✅ 'list' action works (returns hardcoded local-files connector)
- ❌ 'add', 'update', 'remove' actions only return acknowledgments
- ❌ No database persistence

**Requirements**:

1. **Database Schema**:
```sql
CREATE TABLE connectors (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    config JSON NOT NULL,
    enabled BOOLEAN DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

2. **CRUD Operations**:
   - **Add**: Validate schema, insert to DB, optionally verify connection
   - **Update**: Modify config, update timestamp
   - **Remove**: Soft delete or hard delete based on config
   - **List**: Return all connectors with metadata

3. **Connector Types to Support**:
   - `local-files` (existing)
   - `github` (future)
   - `gitlab` (future)
   - `jira` (future)

**Implementation Steps**:
1. Create migration in `internal/connectors/migrations.go`
2. Implement `ConnectorStore` interface in `internal/connectors/store.go`
3. Update handler in `internal/mcp/handlers.go` to use store
4. Add schema validation per connector type
5. Implement optional connection verification
6. Add rollback on error

**Testing**:
- Unit tests for all CRUD operations
- Integration tests with SQLite database
- Edge cases: duplicate IDs, invalid schemas, concurrent updates
- Data persistence tests (restart server, verify connectors remain)

**Acceptance Criteria**:
- [ ] All CRUD operations functional and persisted
- [ ] Data survives server restarts
- [ ] Schema validation prevents invalid configs
- [ ] Proper error handling with rollback
- [ ] 90%+ unit test coverage
- [ ] Integration tests pass

**Estimated Time**: 4-5 hours

---

#### **Task 8.3: Implement MCP `resources/list` and `resources/read`** (5-6 hours)
**GitHub Issue**: #58  
**Priority**: 🔴 HIGH  
**Blocked By**: Task 8.4 (chunking improves resource read quality)  
**Blocks**: Task 8.7 (testing)

**Objective**: Implement MCP resources endpoints backed by indexer/vectorstore for file listing and content retrieval.

**Current Status**:
- ✅ Endpoints defined in MCP server
- ❌ Return placeholder responses
- ❌ No file content retrieval implementation

**Requirements**:

1. **`resources/list`**:
   - Page through indexed files using vectorstore metadata
   - Support filtering by file type, directory
   - Pagination for large file lists (100+ files)
   - Return file metadata (size, modified time, chunks)

2. **`resources/read`**:
   - Retrieve file content with proper chunking
   - Support line range queries (e.g., lines 10-50)
   - Security: path validation, size caps (10MB max)
   - Optional: content redaction for sensitive data

**API Contract**:
```go
// resources/list
type ResourceListRequest struct {
    Filter    string `json:"filter,omitempty"`    // File pattern
    Directory string `json:"directory,omitempty"` // Root directory
    Page      int    `json:"page,omitempty"`      // Page number
    PerPage   int    `json:"per_page,omitempty"`  // Items per page
}

type ResourceListResponse struct {
    Resources  []Resource `json:"resources"`
    TotalCount int        `json:"total_count"`
    Page       int        `json:"page"`
    TotalPages int        `json:"total_pages"`
}

// resources/read
type ResourceReadRequest struct {
    URI        string `json:"uri"`                  // File path or URI
    StartLine  int    `json:"start_line,omitempty"` // Optional line range
    EndLine    int    `json:"end_line,omitempty"`   // Optional line range
}

type ResourceReadResponse struct {
    URI     string `json:"uri"`
    Content string `json:"content"`
    Lines   string `json:"lines"` // "10-50" format
    Size    int64  `json:"size"`
}
```

**Implementation Steps**:
1. Create `internal/mcp/resources.go` for core logic
2. Integrate with vectorstore for file metadata queries
3. Implement pagination logic
4. Add file content retrieval with chunking
5. Implement line range support
6. Add security validation (path safe, size caps)
7. Add error handling for missing files

**Testing**:
- Unit tests for list and read operations
- Integration tests with real file system
- Edge cases: large files, binary files, missing files
- Performance: <200ms for list, <500ms for read

**Acceptance Criteria**:
- [ ] `resources/list` returns paginated file list from indexer
- [ ] `resources/read` retrieves file content with line ranges
- [ ] All paths validated with `security.SafePath()`
- [ ] Size caps enforced (10MB max per file)
- [ ] Pagination works for 1000+ file codebases
- [ ] 90%+ unit test coverage

**Estimated Time**: 5-6 hours

---

### Priority 2: Core Feature Enhancements (MEDIUM - 8-10 hours)

These tasks improve indexing intelligence and add supporting functionality.

#### **Task 8.4: Indexer Code-Aware Chunking** (4-5 hours)
**GitHub Issue**: #59  
**Priority**: 🟡 MEDIUM  
**Blocked By**: None  
**Blocks**: Task 8.3 (improves resource read quality)

**Objective**: Replace single-chunk indexing with semantic code-aware chunking for better search relevance.

**Current Status**:
- ✅ Indexer works with single chunk per file
- ❌ No semantic code boundaries respected
- ❌ `BytesProcessed` always returns 0

**Requirements**:

1. **Code-Aware Chunking Strategy**:
   - **Go files**: Split on function boundaries, struct definitions
   - **Markdown**: Split on heading boundaries (## headers)
   - **JSON/YAML**: Split on top-level keys
   - **Generic**: Fallback to 500-line chunks with overlap

2. **Chunk Metadata**:
   - Function/class/section name
   - Line number range
   - Chunk type (function, struct, section, generic)

3. **Metrics**:
   - Compute accurate `BytesProcessed`
   - Track chunks per file
   - Report chunk size distribution

**Implementation Steps**:
1. Create `internal/indexer/chunker/` package
2. Implement `Chunker` interface with language-specific strategies
3. Update `internal/indexer/indexer.go` to use chunker
4. Add chunk metadata to vectorstore
5. Compute and return accurate `BytesProcessed`
6. Update tests to handle multi-chunk files

**Testing**:
- Unit tests for each chunker type (Go, Markdown, JSON, generic)
- Integration tests with real files
- Verify search relevance improvement (30%+ better recall)
- Edge cases: empty files, single-line files, large files

**Acceptance Criteria**:
- [ ] Go files chunked on function/struct boundaries
- [ ] Markdown files chunked on heading boundaries
- [ ] JSON files chunked on top-level keys
- [ ] Accurate `BytesProcessed` reported
- [ ] Search relevance improved by 30%+ (measured by precision@5)
- [ ] 90%+ unit test coverage

**Estimated Time**: 4-5 hours

---

#### **Task 8.5: Add Additional MCP Tools** (4-5 hours)
**GitHub Issues**: #60 (grep), #61 (agent locator), #62 (process manager)  
**Priority**: 🟡 MEDIUM  
**Blocked By**: None  
**Blocks**: Task 8.7 (testing)

**Objective**: Implement additional MCP tools for grep search, agent location, and process management.

**Sub-Tasks**:

**8.5.1: Implement Grep Tool** (1.5-2 hours)
- Add `context.grep` MCP tool for regex search
- Support file pattern filtering
- Return matched lines with context
- Limit to 1000 matches for performance

**8.5.2: Complete Agent Locator** (1.5-2 hours)
- Implement `agent.locate_files` for finding agent definitions
- Query vectorstore for agent-related files
- Support filtering by agent type
- Return file paths with confidence scores

**8.5.3: Process Manager Path Resolution** (1-1.5 hours)
- Fix path resolution in process manager
- Ensure proper working directory handling
- Add validation for executable paths

**Acceptance Criteria**:
- [ ] Grep tool supports regex patterns
- [ ] Agent locator finds relevant files with >80% accuracy
- [ ] Process manager resolves paths correctly
- [ ] All tools have error handling
- [ ] 85%+ unit test coverage

**Estimated Time**: 4-5 hours

---

### Priority 3: Testing & Documentation (HIGH - 6-10 hours)

These tasks ensure quality and usability of all new features.

#### **Task 8.6: Configurable Runtime Environment** (2-3 hours)
**GitHub Issue**: #63  
**Priority**: 🟡 MEDIUM  
**Blocked By**: None

**Objective**: Make runtime environment configurable for different deployment scenarios.

**Requirements**:
- Add environment variables for key settings
- Support config file overrides
- Validate configuration on startup
- Document all config options

**Acceptance Criteria**:
- [ ] Environment variables override defaults
- [ ] Config validation on startup
- [ ] Documentation updated

**Estimated Time**: 2-3 hours

---

#### **Task 8.7: Comprehensive Testing for New Handlers** (3-4 hours)
**GitHub Issue**: #64  
**Priority**: 🔴 HIGH  
**Blocked By**: Tasks 8.1, 8.2, 8.3, 8.5  
**Blocks**: None

**Objective**: Add comprehensive tests and update documentation for all new MCP handlers.

**Requirements**:

1. **Unit Tests**:
   - Test all new handlers (get_related_info, resources, connector_mgmt)
   - Test error conditions and edge cases
   - Target: 90%+ coverage

2. **Integration Tests**:
   - End-to-end MCP HTTP endpoint tests
   - Multi-tool workflow tests
   - Performance regression tests

3. **Documentation Updates**:
   - Update API reference with new tools
   - Add examples to MCP integration guide
   - Update getting-started guide

**Acceptance Criteria**:
- [ ] 90%+ test coverage for new code
- [ ] Integration test suite passes
- [ ] API documentation updated
- [ ] MCP integration guide examples working

**Estimated Time**: 3-4 hours

---

#### **Task 8.8: GitHub Project Management** (1-3 hours)
**GitHub Issue**: #65  
**Priority**: 🔴 HIGH  
**Blocked By**: None  
**Blocks**: None

**Objective**: Set up GitHub project management for tracking Phase 8 progress.

**Requirements**:
- Create feature branches for each task
- Set up draft PRs with implementation plans
- Track progress and dependencies
- Update project board

**Branch Strategy**:
- `feat/mcp-related-info` (Task 8.1)
- `feat/mcp-connector-mgmt` (Task 8.2)
- `feat/mcp-resources` (Task 8.3)
- `feat/indexer-chunking` (Task 8.4)
- `feat/mcp-additional-tools` (Task 8.5)
- `feat/env-config` (Task 8.6)
- `test/mcp-suite` (Task 8.7)

**Acceptance Criteria**:
- [ ] All feature branches created
- [ ] Draft PRs with scope outlined
- [ ] Dependencies clearly documented
- [ ] Progress tracked in GitHub project

**Estimated Time**: 1-3 hours

---

## Dependency Graph

```
Task 8.8 (Project Setup)
    ↓
Task 8.1 (get_related_info) ────┐
Task 8.2 (connector_mgmt)    ────┤
Task 8.4 (chunking) ─────────────┤
    ↓                            ↓
Task 8.3 (resources)         ────┤
Task 8.5 (additional tools)  ────┤
Task 8.6 (env config)        ────┤
    ↓                            ↓
Task 8.7 (testing & docs) ◄──────┘
```

**Critical Path**: 8.8 → 8.1 → 8.7 (9-14 hours minimum)

---

## Timeline & Milestones

### Week 1 (Days 1-3): Foundation
- **Day 1**: Task 8.8 (project setup) + Task 8.1 start
- **Day 2**: Task 8.1 completion + Task 8.2 start
- **Day 3**: Task 8.2 completion + Task 8.4 start

**Milestone 1**: Core MCP tools in progress (3 days)

### Week 2 (Days 4-7): Implementation
- **Day 4**: Task 8.4 completion + Task 8.3 start
- **Day 5**: Task 8.3 completion + Task 8.5 start
- **Day 6**: Task 8.5 completion + Task 8.6 start
- **Day 7**: Task 8.6 completion + Task 8.7 start

**Milestone 2**: All features implemented (7 days)

### Week 2 (Days 8-10): Testing & Documentation
- **Day 8-9**: Task 8.7 (testing & docs)
- **Day 10**: Final validation, bug fixes, Phase 8 completion

**Milestone 3**: Phase 8 complete, ready for v0.2.0 release (10 days)

---

## Success Metrics

### Functional Metrics
- ✅ 3 new MCP tools fully operational
- ✅ Code-aware chunking active for all indexed files
- ✅ Connector CRUD operations persist across restarts
- ✅ 100% of new code covered by tests

### Performance Metrics
- ✅ `get_related_info` responds in <500ms
- ✅ `resources/list` handles 1000+ files with pagination
- ✅ `resources/read` serves files in <500ms
- ✅ Chunking improves search precision by 30%+

### Quality Metrics
- ✅ 251+ tests passing (100% pass rate maintained)
- ✅ 0 security vulnerabilities (gosec scan clean)
- ✅ 90%+ test coverage on new code
- ✅ Documentation updated for all new features

---

## Risk Assessment

### High Risk
- **Code-aware chunking complexity**: May require language-specific parsers
  - **Mitigation**: Start with simple heuristics, iterate based on results
  
- **Vectorstore query performance**: Complex related info queries may be slow
  - **Mitigation**: Add caching layer, optimize metadata queries

### Medium Risk
- **Connector verification**: Optional verification may fail for some connector types
  - **Mitigation**: Make verification optional, provide clear error messages

- **Test coverage target**: 90% may be ambitious for some handlers
  - **Mitigation**: Prioritize critical paths, document uncovered edge cases

### Low Risk
- **Documentation updates**: May require significant writing
  - **Mitigation**: Reuse existing examples, focus on new tool usage

---

## Phase 8 Completion Criteria

### Must-Have (Blocking v0.2.0 release)
- [ ] Tasks 8.1, 8.2, 8.3 complete (core MCP tools)
- [ ] Task 8.7 complete (testing & documentation)
- [ ] All tests passing (251+ tests)
- [ ] 0 security vulnerabilities
- [ ] Updated documentation

### Should-Have (v0.2.0 or v0.2.1)
- [ ] Task 8.4 complete (code-aware chunking)
- [ ] Task 8.5 complete (additional tools)
- [ ] Task 8.6 complete (configurable env)

### Nice-to-Have (Future phases)
- [ ] Enhanced git integration for ticket flow
- [ ] Advanced connector types (GitHub, GitLab, Jira)
- [ ] Real-time indexing updates

---

## Next Steps

1. **Immediate**: Create GitHub issues for sub-tasks if needed
2. **Day 1**: Start Task 8.8 (project setup) and Task 8.1 (get_related_info)
3. **Week 1**: Complete core MCP tools (Tasks 8.1, 8.2)
4. **Week 2**: Complete remaining tasks and testing
5. **Day 10**: Phase 8 completion review, prepare v0.2.0 release

---

## References

- **Phase 7 Status**: `PHASE7-STATUS.md` (v0.1.0-mvp baseline)
- **GitHub Issues**: #56-65 (MCP completeness)
- **GitHub Project**: "Conexus POC Development"
- **Previous Roadmap**: `TODO.md` (now superseded by this plan)

---

**Last Updated**: October 17, 2025  
**Status**: 📋 PLANNED  
**Next Review**: After Task 8.1 completion
