#!/bin/bash

# Test script for Conexus MCP tools

echo "Testing Conexus MCP tools..."

# Test 1: List tools
echo "1. Listing available tools:"
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus | jq '.result.tools[].name'

echo ""

# Test 2: Start indexing
echo "2. Starting indexing:"
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"start","paths":["."]}}}' | ./conexus | jq .

echo ""

# Test 3: Wait and check status
echo "3. Checking indexing status after 5 seconds:"
sleep 5
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | ./conexus | jq .

echo ""

# Test 4: Try a search
echo "4. Trying a search:"
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"function main","work_context":{"active_file":"cmd/conexus/main.go"}}}}' | ./conexus | jq .

echo ""

# Test 5: Try grep
echo "5. Trying grep:"
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"context.grep","arguments":{"pattern":"func main","include":"*.go"}}}' | ./conexus | jq .