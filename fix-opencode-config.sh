#!/bin/bash
# Fix OpenCode MCP Configuration
# This script fixes the environment variable issues in OpenCode configuration

echo "🔧 Fixing OpenCode MCP Configuration..."
echo ""

# Backup original config
cp ~/.config/opencode/opencode.jsonc ~/.config/opencode/opencode.jsonc.backup

# Copy the fixed configuration
cp opencode-fixed.jsonc ~/.config/opencode/opencode.jsonc

echo "✅ Configuration updated!"
echo ""
echo "📋 What was fixed:"
echo "  - Replaced \${SENTRY_ACCESS_TOKEN} with actual token value"
echo "  - Replaced \${HOST} with actual host value"
echo "  - Replaced \${ORGANIZATION} with actual organization value"
echo "  - Fixed GitHub token environment variable passing"
echo ""
echo "🚀 Next steps:"
echo "  1. Restart OpenCode: pkill -f 'opencode run'"
echo "  2. Start OpenCode with environment: source ./opencode-env.sh && opencode run [your command]"
echo ""
echo "📝 Environment variables are also available in:"
echo "  - ~/.zshrc (for persistent sessions)"
echo "  - ./opencode-env.sh (for current session)"