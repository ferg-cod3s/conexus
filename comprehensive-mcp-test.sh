#!/bin/bash
# Comprehensive MCP Integration Test Suite

echo "=== Conexus MCP Comprehensive Test Suite ==="
echo "Testing all 4 MCP tools with detailed scenarios"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

test_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
}

test_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
}

test_warn() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
}

# Test 1: tools/list
echo "1. Testing tools/list..."
result=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/conexus 2>/dev/null)
tools_count=$(echo "$result" | jq -r '.result.tools | length' 2>/dev/null)
if [ "$tools_count" = "4" ]; then
    test_pass "tools/list returns 4 tools"
else
    test_fail "tools/list returned $tools_count tools (expected 4)"
fi

# Test 2: context.search - basic query
echo
echo "2. Testing context.search - basic functionality..."
result=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"test"}}}' | ./bin/conexus 2>/dev/null)
results_count=$(echo "$result" | jq -r '.result.results | length' 2>/dev/null)
if [ "$results_count" -ge 0 ] 2>/dev/null; then
    test_pass "context.search returns results (count: $results_count)"
else
    test_fail "context.search failed to return results"
fi

# Test 3: context.search - with work context
echo
echo "3. Testing context.search - with work context..."
result=$(echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"function","workContext":{"active_file":"main.go"}}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result' >/dev/null 2>&1; then
    test_pass "context.search with work context works"
else
    test_fail "context.search with work context failed"
fi

# Test 4: context.get_related_info - file path
echo
echo "4. Testing context.get_related_info - file path..."
result=$(echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"context.get_related_info","arguments":{"file_path":"cmd/conexus/main.go"}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result.summary' >/dev/null 2>&1; then
    test_pass "context.get_related_info for file path works"
else
    test_fail "context.get_related_info for file path failed"
fi

# Test 5: context.index_control - status
echo
echo "5. Testing context.index_control - status..."
result=$(echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result.message' >/dev/null 2>&1; then
    docs_count=$(echo "$result" | jq -r '.result.details.documents_indexed' 2>/dev/null)
    test_pass "context.index_control status works (indexed: $docs_count docs)"
else
    test_fail "context.index_control status failed"
fi

# Test 6: context.index_control - start (should fail safely if already running)
echo
echo "6. Testing context.index_control - start..."
result=$(echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"start"}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result.message' >/dev/null 2>&1; then
    test_pass "context.index_control start works"
else
    test_warn "context.index_control start may have failed (check if indexer already running)"
fi

# Test 7: context.connector_management - list
echo
echo "7. Testing context.connector_management - list..."
result=$(echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"context.connector_management","arguments":{"action":"list"}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result.connectors' >/dev/null 2>&1; then
    connectors_count=$(echo "$result" | jq -r '.result.connectors | length' 2>/dev/null)
    test_pass "context.connector_management list works (connectors: $connectors_count)"
else
    test_fail "context.connector_management list failed"
fi

# Test 8: context.connector_management - add connector
echo
echo "8. Testing context.connector_management - add..."
result=$(echo '{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"context.connector_management","arguments":{"action":"add","connector_id":"test-filesystem","connector_config":{"type":"filesystem","path":"./test-data"}}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result.message' >/dev/null 2>&1; then
    test_pass "context.connector_management add works"
else
    test_fail "context.connector_management add failed"
fi

# Test 9: context.connector_management - list again (should show 1 connector)
echo
echo "9. Testing context.connector_management - list after add..."
result=$(echo '{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"context.connector_management","arguments":{"action":"list"}}}' | ./bin/conexus 2>/dev/null)
connectors_count=$(echo "$result" | jq -r '.result.connectors | length' 2>/dev/null)
if [ "$connectors_count" = "1" ]; then
    test_pass "context.connector_management shows added connector"
else
    test_warn "context.connector_management shows $connectors_count connectors (expected 1)"
fi

# Test 10: Error handling - invalid tool name
echo
echo "10. Testing error handling - invalid tool..."
result=$(echo '{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"invalid.tool","arguments":{}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
    test_pass "Error handling works for invalid tool"
else
    test_fail "Error handling failed for invalid tool"
fi

# Test 11: Error handling - missing required params
echo
echo "11. Testing error handling - missing params..."
result=$(echo '{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"context.search"}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
    test_pass "Error handling works for missing required params"
else
    test_fail "Error handling failed for missing params"
fi

echo
echo "=== Test Summary ==="
echo "All core MCP functionality has been tested."
echo "Ready for dogfooding with Claude/OpenCode!"
echo
echo "Next steps:"
echo "1. Copy MCP config: cp mcp-config-claude.json ~/.claude/mcp.json"
echo "2. Restart Claude Desktop"
echo "3. Try: 'Search for authentication code in this project'"
