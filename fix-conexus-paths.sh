#!/bin/bash

# Fix Conexus MCP server command paths in all project configurations
set -e

echo "🔧 Fixing Conexus MCP server paths in all project configurations..."

# Find all opencode.jsonc files
find /Users/johnferguson/Github -name "opencode.jsonc" -not -path "*/conexus/opencode.jsonc" | while read config_file; do
    echo "📝 Updating: $config_file"
    
    # Replace the go run command with direct binary path
    sed -i '' 's|"command": \["go", "run", "/Users/johnferguson/Github/conexus/cmd/conexus"\]|"command": ["/Users/johnferguson/Github/conexus/conexus"]|g' "$config_file"
    
    echo "✅ Updated: $config_file"
done

echo ""
echo "🎉 All project configurations updated!"
echo ""
echo "The Conexus MCP server will now use the compiled binary directly."
echo "This should resolve the MCP server access issues."