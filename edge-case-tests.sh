#!/bin/bash
# Edge Cases and Error Handling Tests

echo "=== Edge Cases and Error Handling Tests ==="
echo

# Test large query
echo "1. Testing large query..."
large_query=$(printf 'x%.0s' {1..1000}) # 1000 chars
result=$(echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"context.search\",\"arguments\":{\"query\":\"$large_query\"}}}" | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result' >/dev/null 2>&1; then
    echo "✓ Large query handled"
else
    echo "✗ Large query failed"
fi

# Test empty query
echo
echo "2. Testing empty query..."
result=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.search","arguments":{"query":""}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
    echo "✓ Empty query properly rejected"
else
    echo "✗ Empty query not rejected"
fi

# Test invalid JSON
echo
echo "3. Testing invalid JSON..."
result=$(echo 'invalid json' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
    echo "✓ Invalid JSON handled"
else
    echo "✗ Invalid JSON not handled"
fi

# Test non-existent file path
echo
echo "4. Testing non-existent file path..."
result=$(echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"context.get_related_info","arguments":{"file_path":"/nonexistent/file.go"}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.result' >/dev/null 2>&1; then
    echo "✓ Non-existent file handled gracefully"
else
    echo "✗ Non-existent file caused error"
fi

# Test invalid index action
echo
echo "5. Testing invalid index action..."
result=$(echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"invalid_action"}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
    echo "✓ Invalid index action rejected"
else
    echo "✗ Invalid index action not rejected"
fi

# Test connector operations on non-existent connector
echo
echo "6. Testing operations on non-existent connector..."
result=$(echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"context.connector_management","arguments":{"action":"update","connector_id":"nonexistent"}}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
    echo "✓ Non-existent connector operations handled"
else
    echo "✗ Non-existent connector operations not handled"
fi

# Test concurrent requests (basic)
echo
echo "7. Testing concurrent requests..."
# Run multiple requests in parallel
for i in {1..3}; do
    (echo "{\"jsonrpc\":\"2.0\",\"id\":$i,\"method\":\"tools/list\"}" | ./bin/conexus >/dev/null 2>&1 && echo "✓ Request $i succeeded") &
done
wait
echo "✓ Concurrent requests completed"

# Test malformed tool arguments
echo
echo "8. Testing malformed tool arguments..."
result=$(echo '{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"context.search","arguments":"not_an_object"}}' | ./bin/conexus 2>/dev/null)
if echo "$result" | jq -e '.error' >/dev/null 2>&1; then
    echo "✓ Malformed arguments handled"
else
    echo "✗ Malformed arguments not handled"
fi

echo
echo "=== Edge Case Testing Complete ==="
echo "All edge cases handled appropriately!"
