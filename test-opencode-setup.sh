#!/bin/bash
# Test OpenCode MCP setup

echo "Testing OpenCode MCP setup..."
echo

# Test that conexus binary works
echo "1. Testing conexus binary..."
if [ -f "./bin/conexus" ]; then
    echo "✓ Binary exists"
else
    echo "✗ Binary not found"
    exit 1
fi

# Test MCP tools/list
echo "2. Testing MCP tools/list..."
result=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/conexus 2>/dev/null | grep -o '"name":"[^"]*"' | wc -l)
if [ "$result" -eq 4 ]; then
    echo "✓ All 4 MCP tools available"
else
    echo "✗ Expected 4 tools, got $result"
fi

# Test database connectivity
echo "3. Testing database..."
if [ -f "data/conexus.db" ]; then
    docs=$(go run check_db.go 2>/dev/null | grep -o "Total documents: [0-9]*" | grep -o "[0-9]*")
    echo "✓ Database ready ($docs documents indexed)"
else
    echo "✗ Database not found"
fi

echo
echo "Setup verification complete!"
echo
echo "To test in OpenCode:"
echo "1. Restart OpenCode if running"
echo "2. Try: 'Search for Go functions in this project'"
echo "3. Or: 'What is the current indexing status?'"
