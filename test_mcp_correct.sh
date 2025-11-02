#!/bin/bash

echo "Testing MCP server with correct tool name..."

# Test with correct tool name: context.index_control
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | ./bin/conexus-darwin-arm64

echo ""
echo "=== Starting force reindex ==="

# Start force reindex with correct tool name
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"force_reindex"}}}' | ./bin/conexus-darwin-arm64