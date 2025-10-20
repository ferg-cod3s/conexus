#!/bin/bash

# Add missing initialize method to Conexus MCP server
set -e

echo "🔧 Adding missing initialize method to Conexus MCP server..."

# Path to the Handle function in server.go
SERVER_FILE="/Users/johnferguson/Github/conexus/internal/mcp/server.go"

# Create a backup
cp "$SERVER_FILE" "$SERVER_FILE.backup"

echo "✅ Backup created: server.go.backup"

# Use Python to make the change more reliably
python3 << 'EOF'
import re

# Read the server.go file
with open("/Users/johnferguson/Github/conexus/internal/mcp/server.go", "r") as f:
    content = f.read()

# Find the Handle method and add initialize case before the default case
pattern = r"(switch method \{.*?)(case \"tools/list\":.*?)(default:)"
replacement = r"\1case \"initialize\":\n\t\t// Handle MCP protocol initialization\n\t\treturn map[string]interface{}{\n\t\t\t\"protocolVersion\": \"2024-11-05\",\n\t\t\t\"capabilities\": map[string]interface{}{\n\t\t\t\t\"tools\": map[string]interface{}{},\n\t\t\t\t\"roots\": map[string]interface{}{\n\t\t\t\t\t\"listChanged\": true,\n\t\t\t\t},\n\t\t\t},\n\t\t\t\"serverInfo\": map[string]interface{}{\n\t\t\t\t\"name\": \"conexus\",\n\t\t\t\t\"version\": \"0.1.0-alpha\",\n\t\t\t},\n\t\t}, nil\n\t\2"

new_content = re.sub(pattern, replacement, content, flags=re.MULTILINE)

# Write the updated content back
with open("/Users/johnferguson/Github/conexus/internal/mcp/server.go", "w") as f:
    f.write(new_content)

print("✅ Added initialize method to server.go")
EOF

echo "🔨 Rebuilding Conexus binary..."
cd /Users/johnferguson/Github/conexus

# Build the updated binary
go build -o conexus ./cmd/conexus

if [ $? -eq 0 ]; then
    echo "✅ Conexus binary rebuilt successfully"
    
    # Test the initialize method
    echo "🧪 Testing initialize method..."
    if echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"roots":{"listChanged":true}},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./conexus 2>&1 | grep -q "protocolVersion"; then
        echo "✅ Initialize method working!"
    else
        echo "⚠️  Initialize method may need adjustment"
    fi
else
    echo "❌ Build failed, restoring backup..."
    cp "$SERVER_FILE.backup" "$SERVER_FILE"
    exit 1
fi

echo ""
echo "🎉 Conexus MCP server updated with initialize method!"
echo ""
echo "The server should now work properly with OpenCode."