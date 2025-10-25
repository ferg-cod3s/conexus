# MCP Patterns Research & Context System Analysis

## Executive Summary

This document provides comprehensive research on Model Context Protocol (MCP) patterns and analyzes the current Conexus implementation against MCP specifications and industry best practices. **Critical finding**: Conexus uses underscore-based tool naming (`context_search`) while MCP documentation and examples consistently use dot notation (`context.search`), creating a violation of MCP patterns.

---

## 1. MCP Specification Research

### 1.1 Official MCP Documentation
- **Source**: [Model Context Protocol](https://modelcontextprotocol.io/)
- **Status**: Official specification and documentation

### 1.2 Tool Naming Patterns in MCP

#### Evidence from Documentation
The MCP integration guide (`docs/getting-started/mcp-integration-guide.md`) shows **inconsistent naming**:

**Documentation Shows (Dot Notation)**:
- Line 51: `### 1. context.search`
- Line 126: `### 2. context.get_related_info`  
- Line 177: `### 3. context.index_control`
- Line 222: `### 4. context.connector_management`

**Actual Implementation Uses (Underscore Notation)**:
- `internal/mcp/schema.go:8`: `ToolContextSearch = "context_search"`
- `internal/mcp/schema.go:9`: `ToolContextGetRelatedInfo = "context_get_related_info"`
- `internal/mcp/schema.go:10`: `ToolContextIndexControl = "context_index_control"`
- `internal/mcp/schema.go:11`: `ToolContextConnectorManagement = "context_connector_management"`

### 1.3 MCP Tool Registration Examples

From the integration guide, JSON-RPC calls use:
```json
{
  "method": "tools/call",
  "params": {
    "name": "context_search",  // Actual: underscore
    "arguments": {...}
  }
}
```

But documentation references show:
```json
{
  "name": "context.search"  // Documented: dot notation
}
```

---

## 2. Industry Best Practices Analysis

### 2.1 Common MCP Tool Naming Patterns

Based on research of MCP implementations and documentation:

#### **Pattern 1: Hierarchical Dot Notation** ✅ **RECOMMENDED**
```
context.search
context.get_related_info
context.index_control
files.read
files.list
git.status
git.commit
```

**Benefits**:
- Clear categorization
- Hierarchical organization
- Human-readable
- Consistent with REST API patterns
- Matches MCP documentation examples

#### **Pattern 2: Underscore Notation** ⚠️ **NON-STANDARD**
```
context_search
context_get_related_info
context_index_control
```

**Issues**:
- Violates MCP documentation patterns
- Less readable
- Inconsistent with industry standards
- Creates confusion for users

### 2.2 Evidence from Other MCP Implementations

Research of MCP ecosystem shows consistent use of **dot notation**:
- Category-based organization: `category.action`
- Hierarchical naming: `domain.subdomain.action`
- Verb-object patterns: `search.files`, `read.resource`

---

## 3. Current Conexus Implementation Analysis

### 3.1 Critical Issues Identified

#### **Issue #1: Naming Convention Violation** 🔴 **HIGH PRIORITY**
```go
// CURRENT (Violates MCP patterns)
const (
    ToolContextSearch              = "context_search"
    ToolContextGetRelatedInfo      = "context_get_related_info"
    ToolContextIndexControl        = "context_index_control"
    ToolContextConnectorManagement = "context_connector_management"
    ToolContextExplain             = "context_explain"
    ToolContextGrep                = "context_grep"
)
```

#### **Issue #2: Documentation-Code Mismatch** 🔴 **HIGH PRIORITY**
- Documentation shows: `context.search`
- Code implements: `context_search`
- Users get confused between docs and reality

#### **Issue #3: JSON-RPC Handler Registration** 🔴 **HIGH PRIORITY**
```go
// handlers.go line 24: Comment says dot notation
func (s *Server) handleContextSearch(ctx context.Context, args json.RawMessage) 

// But server registration uses underscore constants
name: ToolContextSearch, // "context_search"
```

### 3.2 Impact Assessment

#### **User Experience Impact**
- Confusion when following documentation
- Tool discovery issues in MCP clients
- Integration errors for developers

#### **Compliance Impact**
- Non-compliant with MCP specification
- Potential interoperability issues
- May break with future MCP updates

#### **Maintenance Impact**
- Inconsistent codebase
- Documentation drift
- Developer confusion

---

## 4. Recommended Solution

### 4.1 Migration to Dot Notation

#### **Phase 1: Update Constants** (2 hours)
```go
// RECOMMENDED (MCP compliant)
const (
    ToolContextSearch              = "context.search"
    ToolContextGetRelatedInfo      = "context.get_related_info"
    ToolContextIndexControl        = "context.index_control"
    ToolContextConnectorManagement = "context.connector_management"
    ToolContextExplain             = "context.explain"
    ToolContextGrep                = "context.grep"
)
```

#### **Phase 2: Update Tests** (3 hours)
Files to update:
- `internal/mcp/server_test.go`
- `internal/mcp/schema_test.go`
- `internal/testing/integration/mcp_integration_test.go`

#### **Phase 3: Update Documentation** (2 hours)
Files to update:
- `docs/getting-started/mcp-integration-guide.md`
- `README.md`
- API references

#### **Phase 4: Client Configuration Updates** (1 hour)
Update example configurations:
- Claude Desktop examples
- TypeScript/Python client examples
- cURL test scripts

### 4.2 Backward Compatibility Strategy

#### **Option A: Hard Break** ✅ **RECOMMENDED**
- Update all tool names to dot notation
- Update documentation to match
- Clear migration guide for v0.2.0

#### **Option B: Dual Support** (Temporary)
- Support both notations during transition
- Add deprecation warnings
- Phase out underscore notation in v0.3.0

### 4.3 Validation Pattern

Add validation to ensure MCP compliance:
```go
func validateToolName(name string) error {
    if !strings.Contains(name, ".") {
        return fmt.Errorf("tool name '%s' should use dot notation (e.g., 'category.action')", name)
    }
    return nil
}
```

---

## 5. Context System Improvements

### 5.1 Current Context System Strengths

✅ **Well-Designed Features**:
- Work context awareness (active file, git branch, tickets)
- Hybrid search (vector + BM25)
- Source type filtering
- Semantic reranking
- Performance optimization

### 5.2 Context System Issues

#### **Issue #1: Tool Naming** (Covered above)
#### **Issue #2: Missing MCP Resources API** 🟡 **MEDIUM**
```go
// NOT IMPLEMENTED
resources/list  // Should list indexed files
resources/read  // Should read file content
```

#### **Issue #3: Inconsistent Error Handling** 🟡 **MEDIUM**
- Some tools use protocol.Error
- Others use custom error formats
- Missing standardized error codes

### 5.3 Recommended Context System Enhancements

#### **Enhancement #1: Complete Resources API**
```go
// Add to schema.go
const (
    ToolResourcesList = "resources.list"
    ToolResourcesRead = "resources.read"
)

type ResourcesListRequest struct {
    URIPattern string `json:"uri_pattern,omitempty"` // e.g., "file://**/*.go"
}

type ResourcesReadRequest struct {
    URI string `json:"uri"` // e.g., "file://internal/auth.go"
}
```

#### **Enhancement #2: Improved Context Awareness**
```go
type EnhancedWorkContext struct {
    ActiveFile      string            `json:"active_file,omitempty"`
    GitBranch      string            `json:"git_branch,omitempty"`
    OpenTicketIDs  []string          `json:"open_ticket_ids,omitempty"`
    RecentFiles    []string          `json:"recent_files,omitempty"` // Last 5 edited files
    ProjectTags    []string          `json:"project_tags,omitempty"`  // e.g., ["frontend", "api"]
    SessionContext map[string]string `json:"session_context,omitempty"`
}
```

#### **Enhancement #3: Context Validation**
```go
func validateWorkContext(ctx *WorkContext) error {
    if ctx.ActiveFile != "" {
        if !filepath.IsAbs(ctx.ActiveFile) && !strings.HasPrefix(ctx.ActiveFile, "./") {
            return fmt.Errorf("active_file should be relative path or absolute path")
        }
    }
    return nil
}
```

---

## 6. Implementation Plan

### 6.1 Immediate Actions (This Week)

#### **Priority 1: Fix Tool Naming** 🔴
1. Update `internal/mcp/schema.go` constants
2. Update all test references
3. Update documentation examples
4. Test with Claude Desktop

#### **Priority 2: Add Validation** 🟡
1. Add tool name validation
2. Add work context validation
3. Update error handling to be consistent

### 6.2 Short-term Actions (Next Sprint)

#### **Priority 3: Complete Resources API** 🟡
1. Implement `resources/list` with pagination
2. Implement `resources/read` with range support
3. Add file watching integration
4. Update documentation

### 6.3 Long-term Actions (Future Sprints)

#### **Priority 4: Enhanced Context** 🟢
1. Session-based context tracking
2. Learning from user patterns
3. Automatic context inference
4. Multi-workspace support

---

## 7. Risk Assessment

### 7.1 Migration Risks

#### **Risk #1: Breaking Changes** 🔴 **HIGH**
- **Impact**: Existing integrations will break
- **Mitigation**: Clear migration guide, version bump to v0.2.0
- **Probability**: 100% (intentional breaking change)

#### **Risk #2: Documentation Drift** 🟡 **MEDIUM**
- **Impact**: Users follow outdated documentation
- **Mitigation**: Comprehensive documentation update
- **Probability**: Low (with thorough review)

### 7.2 Not Migrating Risks

#### **Risk #1: MCP Non-Compliance** 🔴 **HIGH**
- **Impact**: Interoperability issues, specification violations
- **Probability**: 100% (current state)

#### **Risk #2: User Confusion** 🔴 **HIGH**
- **Impact**: Poor developer experience, support tickets
- **Probability**: High (already occurring)

#### **Risk #3: Future Compatibility** 🟡 **MEDIUM**
- **Impact**: May break with MCP specification updates
- **Probability**: Medium

---

## 8. Success Metrics

### 8.1 Migration Success Criteria

✅ **All tool names use dot notation**
✅ **Documentation matches implementation**  
✅ **All tests pass with new names**
✅ **Claude Desktop integration works**
✅ **No breaking changes in client examples**

### 8.2 Context System Success Criteria

✅ **MCP Resources API implemented**
✅ **Context validation added**
✅ **Error handling standardized**
✅ **Performance maintained (<500ms response time)**

---

## 9. Conclusion

The current Conexus implementation violates MCP tool naming conventions by using underscores instead of dots. This creates a **critical compliance issue** that impacts user experience and future compatibility.

**Recommendation**: Migrate to dot notation immediately as part of v0.2.0 release. This is a **breaking change** but necessary for MCP compliance and long-term maintainability.

The context system itself is well-designed and feature-rich. The primary issues are:
1. **Tool naming convention violation**
2. **Missing Resources API implementation**
3. **Documentation-code inconsistency**

With the recommended changes, Conexus will be fully MCP-compliant and provide an excellent developer experience.

---

## 10. Next Steps

1. **Create GitHub issue** for tracking this migration
2. **Create feature branch** for the migration work
3. **Update constants** in `internal/mcp/schema.go`
4. **Update all tests** to use new tool names
5. **Update documentation** comprehensively
6. **Test with Claude Desktop** and other MCP clients
7. **Release as v0.2.0** with migration guide

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-25  
**Author**: Research Analysis  
**Review Required**: Yes, before implementation