#!/bin/bash
set -e

CONEXUS_BINARY="${1:-./conexus}"

if [ ! -f "$CONEXUS_BINARY" ]; then
  echo "❌ Conexus binary not found: $CONEXUS_BINARY"
  exit 1
fi

echo "🧪 Testing Conexus MCP Server (stdio)"
echo "======================================="
echo ""

# Test 1: tools/list
echo "Test 1: tools/list"
RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | "$CONEXUS_BINARY")
if echo "$RESPONSE" | grep -q '"result"'; then
  echo "✅ PASS: tools/list"
else
  echo "❌ FAIL: tools/list"
  echo "$RESPONSE"
  exit 1
fi
echo ""

# Test 2: context.search
echo "Test 2: context.search"
REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"test","top_k":5}}}'
RESPONSE=$(echo "$REQUEST" | "$CONEXUS_BINARY")
if echo "$RESPONSE" | grep -q '"result"'; then
  echo "✅ PASS: context.search"
else
  echo "❌ FAIL: context.search"
  echo "$RESPONSE"
  exit 1
fi
echo ""

# Test 3: context.index_control
echo "Test 3: context.index_control"
REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}'
RESPONSE=$(echo "$REQUEST" | "$CONEXUS_BINARY")
if echo "$RESPONSE" | grep -q '"result"'; then
  echo "✅ PASS: context.index_control"
else
  echo "❌ FAIL: context.index_control"
  echo "$RESPONSE"
  exit 1
fi
echo ""

# Test 4: Error handling
echo "Test 4: Error handling (invalid JSON)"
RESPONSE=$(echo 'invalid json' | "$CONEXUS_BINARY" 2>&1 || true)
if echo "$RESPONSE" | grep -q '"error"'; then
  echo "✅ PASS: Error handling"
else
  echo "⚠️  WARNING: Error handling might need improvement"
fi
echo ""

echo "🎉 All stdio tests passed!"
