# MCP Inspector Integration Tests

Comprehensive test suite for validating Conexus MCP server compliance using the official MCP Inspector tool.

## Quick Start

```bash
# 1. Build Conexus binary
go build -o conexus ./cmd/conexus

# 2. Run automated tests
npm run test:inspector

# 3. Or run inspector interactively
npm run inspector
```

## Test Coverage

This test suite validates:

### Protocol Compliance (2 tests)
- ✅ JSON-RPC 2.0 format validation
- ✅ Request ID tracking

### Tool Availability (1 test)
- ✅ tools/list endpoint returns all 8 MCP tools

### Tool Invocations (6 tests)
- ✅ `context.search` - Semantic search
- ✅ `context.grep` - Pattern matching
- ✅ `context.index_control` - Index management
- ✅ `context.get_related_info` - Related file discovery
- ✅ `context.connector_management` - Connector CRUD
- ✅ Additional tools: `context.explain`, `github.sync_*`

### Error Handling (2 tests)
- ✅ Invalid tool name rejection
- ✅ Invalid parameter handling

## Usage

### Automated Tests

```bash
# Run all tests
node test-inspector.js --conexus-path ./conexus

# Expected output:
# ==========================================
#   MCP Inspector Integration Tests
# ==========================================
#
# [INFO] Starting MCP Inspector...
# [PASS] MCP Inspector started
#
# [INFO] Test 1: tools/list returns available tools
# [PASS] Found 8 tools
# ...
# Tests Run:    10
# Tests Passed: 10
# Tests Failed: 0
# ==========================================
```

### Interactive Inspector

```bash
# Start inspector
npx -y @modelcontextprotocol/inspector ./conexus

# Inspector commands:
> tools                                    # List all tools
> describe context.search                  # Show tool schema
> call context.search {"query": "test"}    # Test a tool
> validate                                 # Check compliance
> exit                                     # Exit inspector
```

### Custom Configuration

```bash
# With custom config file
CONEXUS_CONFIG=./config.test.yml npm run inspector

# With specific binary path
node test-inspector.js --conexus-path /path/to/conexus
```

## Test Scenarios

### 1. Basic Tool Discovery
```javascript
// Tests that all 8 MCP tools are registered
// Expected: context.search, context.get_related_info,
//           context.index_control, context.connector_management,
//           context.explain, context.grep, github.sync_*
```

### 2. Search Tool
```javascript
call context.search {
  "query": "function",
  "top_k": 5
}
// Expected: Returns search results with scores
```

### 3. Grep Tool
```javascript
call context.grep {
  "pattern": "func",
  "include": "*.go"
}
// Expected: Returns matched lines with file paths
```

### 4. Index Status
```javascript
call context.index_control {
  "action": "status"
}
// Expected: Returns indexing status and metrics
```

### 5. Related Info
```javascript
call context.get_related_info {
  "file_path": "main.go"
}
// Expected: Returns related files, PRs, issues
```

### 6. Connector Management
```javascript
call context.connector_management {
  "action": "list"
}
// Expected: Returns list of configured connectors
```

## Troubleshooting

### Inspector not found
```bash
# Install globally
npm install -g @modelcontextprotocol/inspector

# Or use npx (no installation)
npx @modelcontextprotocol/inspector ./conexus
```

### Binary not found
```bash
# Build the binary first
go build -o conexus ./cmd/conexus

# Verify it exists
ls -la ./conexus
```

### Node.js version
```bash
# Check Node.js version (requires >=16)
node --version

# Should be v16.0.0 or higher
```

### Test failures
```bash
# Check Conexus can start
./conexus --help

# Check config is valid
./conexus --config config.test.yml

# Run with verbose output
node test-inspector.js --conexus-path ./conexus 2>&1 | tee inspector-test.log
```

## CI/CD Integration

Add to your GitHub Actions workflow:

```yaml
- name: Run MCP Inspector Tests
  run: |
    go build -o conexus ./cmd/conexus
    npm run test:inspector
```

## Documentation

- [MCP Inspector Official Docs](https://modelcontextprotocol.io/docs/tools/inspector)
- [MCP Testing Guide](../../../docs/getting-started/mcp-testing-guide.md)
- [MCP Integration Guide](../../../docs/getting-started/mcp-integration-guide.md)

## Related Files

- `test-inspector.js` - Main test suite (this directory)
- `../../../scripts/test-stdio.sh` - stdio transport test script
- `../../../docs/getting-started/mcp-testing-guide.md` - Complete testing guide
- `../../../config.test.yml` - Test configuration

## Test Results

Last run: 2025-11-10
- Tests: 10/10 passing ✅
- Protocol compliance: JSON-RPC 2.0 ✅
- Tool count: 8/8 registered ✅
- Error handling: Validated ✅

---

**Status**: ✅ All tests passing
**Last Updated**: 2025-11-10
**MCP Spec Version**: 2024-11-05
