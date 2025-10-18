#!/bin/bash
# Test script for Conexus MCP server

echo "Testing Conexus MCP Server..."
echo

# Test tools/list
echo "1. Testing tools/list..."
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/conexus | jq '.result.tools | length'
echo

# Test context.search
echo "2. Testing context.search..."
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"test","workContext":"Testing conexus"}}}' | ./bin/conexus | jq '.result | {results_count: (.results | length), query_time: .queryTime}'
echo

# Test context.index_control status
echo "3. Testing context.index_control status..."
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | ./bin/conexus | jq '.result.message'
echo

# Test context.connector_management list
echo "4. Testing context.connector_management list..."
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"context.connector_management","arguments":{"action":"list"}}}' | ./bin/conexus | jq '.result | {connectors_count: (.connectors | length)}'

echo
echo "MCP tests completed!"
