#!/bin/bash

# Add missing initialize method to Conexus MCP server
set -e

echo "🔧 Adding missing initialize method to Conexus MCP server..."

# Path to the server.go file
SERVER_FILE="/Users/johnferguson/Github/conexus/internal/mcp/server.go"

# Create a backup
cp "$SERVER_FILE" "$SERVER_FILE.backup"

echo "✅ Backup created: server.go.backup"

# Add the initialize method to the Handle function
# We need to add a case for "initialize" method before the default case

# The initialize method should return server capabilities
cat > /tmp/initialize_patch.txt << 'EOF'
	case "initialize":
		// Handle MCP protocol initialization
		return map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name": "conexus",
				"version": "0.1.0-alpha",
			},
		}, nil
	default:
EOF

# Insert the initialize case before the default case
sed -i '/default:/i\
	case "initialize":\
		// Handle MCP protocol initialization\
		return map[string]interface{}{\
			"protocolVersion": "2024-11-05",\
			"capabilities": map[string]interface{}{\
				"tools": map[string]interface{}{},\
				"roots": map[string]interface{}{\
					"listChanged": true,\
				},\
			},\
			"serverInfo": map[string]interface{}{\
				"name": "conexus",\
				"version": "0.1.0-alpha",\
			},\
		}, nil\
' "$SERVER_FILE"

echo "✅ Added initialize method to server.go"

# Now rebuild the Conexus binary
echo "🔨 Rebuilding Conexus binary..."
cd /Users/johnferguson/Github/conexus

# Build the updated binary
go build -o conexus ./cmd/conexus

if [ $? -eq 0 ]; then
    echo "✅ Conexus binary rebuilt successfully"
    
    # Test the initialize method
    echo "🧪 Testing initialize method..."
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"roots":{"listChanged":true}},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./conexus 2>&1 | head -5
    
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