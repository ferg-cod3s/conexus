#!/bin/bash

# Update global OpenCode configuration to include Conexus MCP server
set -e

echo "🔧 Updating global OpenCode configuration..."

# Create the updated global config
cat > ~/.config/opencode/opencode.jsonc << 'EOF'
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "context7": {
      "type": "remote",
      "url": "https://mcp.context7.com/mcp",
      "enabled": true
    },
    "gh_grep": {
      "type": "remote", 
      "url": "https://mcp.grep.app",
      "enabled": true
    },
    "conexus": {
      "type": "local",
      "command": ["/Users/johnferguson/Github/conexus/conexus"],
      "enabled": true,
      "environment": {
        "CONEXUS_PORT": "0",
        "CONEXUS_DB_PATH": "{env:CONEXUS_DB_PATH:./data/conexus.db}",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  },
  "agent": {
    "build": {
      "mode": "primary",
      "model": "anthropic/claude-sonnet-4-20250514",
      "description": "Full development agent with all tools enabled",
      "temperature": 0.3,
      "tools": {
        "write": true,
        "edit": true,
        "bash": true,
        "webfetch": true
      }
    },
    "plan": {
      "mode": "primary", 
      "model": "anthropic/claude-haiku-4-20250514",
      "description": "Planning and analysis agent with restricted permissions",
      "temperature": 0.1,
      "permission": {
        "write": "ask",
        "edit": "ask", 
        "bash": "ask"
      },
      "tools": {
        "write": true,
        "edit": true,
        "bash": true,
        "webfetch": true
      }
    },
    "conexus-expert": {
      "mode": "subagent",
      "model": "anthropic/claude-sonnet-4-20250514",
      "description": "Expert Conexus agent for code analysis, indexing, and semantic search operations",
      "temperature": 0.2,
      "tools": {
        "write": false,
        "edit": false,
        "bash": false,
        "webfetch": true
      },
      "permission": {
        "bash": "deny",
        "edit": "deny", 
        "write": "deny"
      }
    }
  },
  "tools": {
    "context7": true,
    "gh_grep": true,
    "conexus": true
  }
}
EOF

echo "✅ Global configuration updated!"
echo ""
echo "The Conexus MCP server is now available globally."
echo "You can now use @conexus-expert in any OpenCode session."