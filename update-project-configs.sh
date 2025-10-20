#!/bin/bash

# Update all project configs to use the conexus binary instead of go run
set -e

echo "🔧 Updating project configurations to use conexus binary..."

CONFIGS=(
  "advent_of_code"
  "coolify-mcp-server"
  "f3rg-redis"
  "gotunnel"
  "hello-zero-example"
  "LocalHaven-CMS"
  "opencode-nexus"
  "rune"
  "tunnelforge"
  "web-project"
)

for config in "${CONFIGS[@]}"; do
  config_file="PROJECT_CONFIGS/${config}-opencode.jsonc"
  
  if [ -f "$config_file" ]; then
    echo "  📝 Updating $config_file..."
    
    # Replace the go run command with binary path
    sed -i '' 's|"command": \["go", "run", "/Users/johnferguson/Github/conexus/cmd/conexus"\]|"command": ["/Users/johnferguson/Github/conexus/conexus"]|g' "$config_file"
    
    echo "  ✅ Updated $config"
  else
    echo "  ⚠️  Skipping $config - file not found"
  fi
done

echo ""
echo "🎉 All project configurations updated!"
echo ""
echo "The configs now use the compiled conexus binary with the initialize method."
