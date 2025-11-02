#!/bin/bash

echo "Testing MCP server indexing..."

# Start force reindex
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context/index","arguments":{"action":"force_reindex"}}}' | ./bin/conexus-darwin-arm64 > mcp_output.log 2>&1 &

MCP_PID=$!
echo "Started MCP server with PID: $MCP_PID"

# Wait a bit for indexing to start
sleep 2

# Check status
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context/index","arguments":{"action":"status"}}}' | ./bin/conexus-darwin-arm64 >> mcp_output.log 2>&1

# Wait more for indexing
sleep 3

# Check status again
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"context/index","arguments":{"action":"status"}}}' | ./bin/conexus-darwin-arm64 >> mcp_output.log 2>&1

# Kill background process if still running
kill $MCP_PID 2>/dev/null

echo "=== MCP Output ==="
cat mcp_output.log