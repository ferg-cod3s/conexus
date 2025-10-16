# Task 7.3: API Documentation - Detailed Plan

## Assessment Summary

### Current State Analysis

**✅ What's Complete:**
1. **Internal Go API Documentation** (`docs/api-reference.md`)
   - 1,595 lines of comprehensive documentation
   - AGENT_OUTPUT_V1 schema fully documented
   - Validation framework API complete
   - Profiling API documented with examples
   - **Audience:** Go developers building agents
   - **Status:** Production-ready ✅

2. **API Specification Document** (`docs/API-Specification.md`)
   - 882 lines covering both REST and MCP interfaces
   - Detailed REST API spec (authentication, rate limiting, endpoints)
   - MCP tool schemas (all 4 tools defined)
   - OpenAPI 3.0.3 specification included
   - **Status:** Specification complete ✅

3. **MCP Implementation** (`internal/mcp/`)
   - 4 tools fully implemented:
     - `context.search` - Hybrid search with filters
     - `context.get_related_info` - File/ticket history
     - `context.index_control` - Index operations
     - `context.connector_management` - Connector CRUD
   - Resources defined (engine://files/ scheme)
   - JSON-RPC 2.0 protocol handlers
   - **Status:** Implementation complete ✅

**❌ What's Missing:**

1. **MCP Integration Guide** (High Priority)
   - No quick-start guide for external users
   - Missing Claude Code integration instructions
   - No IDE integration examples
   - No transport setup guide (stdio vs HTTP/SSE)

2. **MCP Tool Documentation Discrepancy** (High Priority)
   - `internal/mcp/README.md` lists only 2 tools (outdated)
   - Actual implementation has 4 tools
   - Missing tool usage examples
   - No error handling patterns documented

3. **Practical Examples** (Medium Priority)
   - No end-to-end usage scenarios
   - Missing common workflow examples
   - No troubleshooting guide
   - No best practices documented

4. **REST API Implementation Status** (Low Priority)
   - Specification exists, but implementation unclear
   - No REST endpoints found in codebase
   - Unclear if REST is planned or deferred

### Key Findings

1. **MCP is the Primary Interface** ✅
   - REST API is marked as "Secondary" in spec
   - All MCP tools are implemented and working
   - Focus should be on MCP documentation

2. **Implementation vs Documentation Gap**
   - Code is complete and tested
   - External documentation is minimal
   - Internal docs are excellent but not user-facing

3. **Target Audience Split**
   - Internal Go API docs → Go developers building on Conexus
   - External MCP docs → Users integrating Conexus into their workflow

## Detailed Execution Plan

### Phase 1: Update MCP Package Documentation (2 hours)
**Goal:** Bring `internal/mcp/README.md` up to date with implementation

#### Task 7.3.1: Update MCP README ✅
- [ ] Document all 4 tools (not just 2)
- [ ] Add complete JSON schema for each tool
- [ ] Include request/response examples
- [ ] Document error codes and handling
- [ ] Add resource URI scheme documentation
- [ ] Update implementation status checklist

**Files to Update:**
- `internal/mcp/README.md`

**Success Criteria:**
- All 4 tools documented with schemas
- At least 1 example per tool
- Error handling documented
- Implementation checklist complete

---

### Phase 2: Create MCP Integration Guide (2 hours)
**Goal:** Enable external users to integrate Conexus in <5 minutes

#### Task 7.3.2: Create Quick-Start Guide ✅
Create `docs/getting-started/mcp-integration-guide.md`

**Contents:**
1. **Prerequisites**
   - Go 1.23.4+
   - Supported platforms (Linux, macOS, Windows)
   - Optional: Vector DB setup

2. **Installation**
   ```bash
   git clone https://github.com/ferg-cod3s/conexus
   cd conexus
   go build ./cmd/conexus
   ```

3. **Configuration**
   - Copy `config.example.yml` to `config.yml`
   - Configure embedding provider
   - Set up vector store

4. **Running the MCP Server**
   ```bash
   ./conexus --mode=mcp --config=config.yml
   ```

5. **Claude Code Integration**
   - Add to `claude_desktop_config.json`:
   ```json
   {
     "mcpServers": {
       "conexus": {
         "command": "/path/to/conexus",
         "args": ["--mode=mcp", "--config=/path/to/config.yml"],
         "transport": "stdio"
       }
     }
   }
   ```

6. **Verification**
   - Test with simple search query
   - Verify results returned

**Files to Create:**
- `docs/getting-started/mcp-integration-guide.md`

---

#### Task 7.3.3: Create Tool Usage Examples ✅
Create `docs/mcp-tool-examples.md`

**Contents:**

1. **Tool: context.search**
   - Basic search
   - Search with work context (active file, branch, tickets)
   - Filtered search (by source type, date range)
   - Pagination example

2. **Tool: context.get_related_info**
   - Get file history
   - Get ticket-related discussions
   - Combined file + ticket query

3. **Tool: context.index_control**
   - Check index status
   - Start indexing
   - Stop indexing
   - Force reindex

4. **Tool: context.connector_management**
   - List connectors
   - Add GitHub connector
   - Update connector config
   - Remove connector

**Each Example Includes:**
- Request JSON
- Expected response
- Common errors and fixes
- Best practices

**Files to Create:**
- `docs/mcp-tool-examples.md`

---

### Phase 3: Advanced Integration Documentation (1 hour)

#### Task 7.3.4: Create IDE Integration Guide ✅
Create `docs/getting-started/ide-integration.md`

**Contents:**
1. **Claude Code Integration** (detailed)
2. **VS Code Integration** (if MCP extension available)
3. **Cursor Integration** (if supported)
4. **Custom MCP Client** (for other editors)

**Files to Create:**
- `docs/getting-started/ide-integration.md`

---

#### Task 7.3.5: Create Error Handling Guide ✅
Create `docs/mcp-error-handling.md`

**Contents:**
1. **Error Response Format**
   ```json
   {
     "jsonrpc": "2.0",
     "error": {
       "code": -32602,
       "message": "Invalid params",
       "data": "query is required"
     },
     "id": 1
   }
   ```

2. **Common Error Codes**
   - `-32700`: Parse error
   - `-32600`: Invalid request
   - `-32601`: Method not found
   - `-32602`: Invalid params
   - `-32603`: Internal error

3. **Tool-Specific Errors**
   - Search errors (embedding failures, index unavailable)
   - Index control errors (already running, permission denied)
   - Connector errors (config invalid, auth failed)

4. **Retry Strategies**
   - Transient vs permanent errors
   - Exponential backoff example
   - Circuit breaker pattern

**Files to Create:**
- `docs/mcp-error-handling.md`

---

### Phase 4: Documentation Organization & Polish (30 minutes)

#### Task 7.3.6: Update Main Documentation Index ✅
Update `docs/README.md` to organize all documentation

**New Structure:**
```
docs/
├── README.md (updated with clear navigation)
├── getting-started/
│   ├── developer-onboarding.md (existing)
│   ├── mcp-integration-guide.md (NEW)
│   └── ide-integration.md (NEW)
├── api/
│   ├── api-reference.md (existing - rename from root)
│   ├── API-Specification.md (existing - keep in root for now)
│   ├── mcp-tool-examples.md (NEW)
│   └── mcp-error-handling.md (NEW)
├── architecture/ (existing)
├── operations/ (existing)
├── contributing/ (existing)
└── research/ (existing)
```

**Files to Update:**
- `docs/README.md` (create or update)
- Update cross-references in existing docs

---

#### Task 7.3.7: Document REST API Status ✅
Add clarification to `docs/API-Specification.md`

**Add Section:**
```markdown
## Implementation Status

### MCP Server (v1) ✅
- **Status:** Fully implemented and tested
- **Transport:** stdio (primary), HTTP/SSE (planned)
- **Tools:** All 4 tools operational
- **Resources:** File system access via engine:// scheme

### REST API (v1) 📋
- **Status:** Specification complete, implementation deferred
- **Reason:** MCP is primary interface for AI assistants
- **Timeline:** REST API planned for v2.0 (web dashboard, CLI)
- **Current Workaround:** Use MCP over HTTP/SSE for web clients
```

**Files to Update:**
- `docs/API-Specification.md`

---

### Phase 5: Testing & Validation (30 minutes)

#### Task 7.3.8: Validate All Documentation ✅
- [ ] Test all code examples
- [ ] Verify all links work
- [ ] Ensure consistent terminology
- [ ] Check formatting (markdown, code blocks)
- [ ] Spell check all new docs

#### Task 7.3.9: Create Documentation Verification Checklist ✅
Create `docs/DOCUMENTATION_CHECKLIST.md`

**Contents:**
- [ ] All MCP tools documented
- [ ] Quick-start guide tested (<5 min to run)
- [ ] At least 3 examples per tool
- [ ] Error handling documented
- [ ] Claude Code integration verified
- [ ] All links resolve
- [ ] No broken code examples
- [ ] Consistent terminology throughout

---

## Success Criteria

### Must Have (Priority 1)
✅ All 4 MCP tools fully documented with examples
✅ Quick-start guide: user can run Conexus in <5 minutes
✅ Claude Code integration documented and tested
✅ Error handling guide with common issues
✅ `internal/mcp/README.md` updated to match implementation

### Should Have (Priority 2)
✅ Minimum 3 practical examples per tool
✅ IDE integration guide (Claude Code + at least 1 other)
✅ REST API status clarified (implemented or not)
✅ Documentation organization improved

### Nice to Have (Priority 3)
⏸️ Advanced integration patterns (custom transports)
⏸️ Performance tuning guide for large codebases
⏸️ Multi-language client examples

---

## Time Estimate

| Phase | Time | Status |
|-------|------|--------|
| 1. Update MCP README | 2h | ⏳ TODO |
| 2. Create Integration Guides | 2h | ⏳ TODO |
| 3. Advanced Documentation | 1h | ⏳ TODO |
| 4. Organization & Polish | 0.5h | ⏳ TODO |
| 5. Testing & Validation | 0.5h | ⏳ TODO |
| **Total** | **6h** | **0% complete** |

---

## Files to Create (9 new files)

1. ✅ `TASK_7.3_PLAN.md` (this file)
2. ⏳ `docs/getting-started/mcp-integration-guide.md`
3. ⏳ `docs/mcp-tool-examples.md`
4. ⏳ `docs/getting-started/ide-integration.md`
5. ⏳ `docs/mcp-error-handling.md`
6. ⏳ `docs/README.md` (new navigation index)
7. ⏳ `docs/DOCUMENTATION_CHECKLIST.md`
8. ⏳ Update `internal/mcp/README.md`
9. ⏳ Update `docs/API-Specification.md` (add status section)

---

## Dependencies

- ✅ Phase 7 Tasks 7.1 & 7.2 complete
- ✅ MCP implementation complete and tested
- ✅ All 4 tools verified working
- ⏸️ Access to Claude Code for testing (user has)
- ⏸️ Example codebase for indexing demos

---

## Next Steps (Immediate)

1. **Create Task 7.3.1 Branch**
   ```bash
   git checkout -b task/7.3.1-update-mcp-readme
   ```

2. **Start with MCP README Update** (Task 7.3.1)
   - Document all 4 tools
   - Add schemas and examples
   - Update implementation checklist

3. **After 7.3.1, move to Quick-Start** (Task 7.3.2)
   - Create integration guide
   - Test with actual Conexus build

---

## Notes

- **REST API**: Specification exists but not implemented. This is intentional - MCP is primary interface.
- **Documentation Philosophy**: Practical examples > theoretical explanations
- **Target Time**: All documentation should enable user to succeed in <5 minutes
- **Next Major Task**: Performance optimization (Task 7.4) after documentation complete

---

**Status:** Ready to begin Task 7.3.1 (Update MCP README)
**Estimated Completion:** Task 7.3 complete in ~6 hours
**Phase 7 Progress:** ~40% → 60% after Task 7.3
