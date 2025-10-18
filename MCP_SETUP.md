# Conexus MCP Setup Guide

## Overview
Conexus provides 4 MCP tools for intelligent code and document search:
- `context.search` - Comprehensive semantic search
- `context.get_related_info` - Find related information for files/tickets
- `context.index_control` - Control indexing operations
- `context.connector_management` - Manage data connectors

## Quick Setup

### 1. Ensure Binary is Built
```bash
cd /home/f3rg/src/github/conexus
go build -o ./bin/conexus ./cmd/conexus
```

### 2. Configure MCP Client

#### For Claude Desktop:
```bash
# Copy config to Claude's MCP directory
mkdir -p ~/.claude
cp mcp-config-claude.json ~/.claude/mcp.json
```

#### For OpenCode:
```bash
# Copy config to your OpenCode config
cp mcp-config-opencode.json ~/opencode-config.json
# Or merge with existing config
```

### 3. Test Setup
```bash
./test-mcp.sh
```

## MCP Tools Usage

### context.search
Search for code, documents, and discussions using natural language.

```json
{
  "method": "tools/call",
  "params": {
    "name": "context.search",
    "arguments": {
      "query": "authentication logic",
      "workContext": {
        "active_file": "auth.go",
        "git_branch": "feature/auth"
      }
    }
  }
}
```

### context.get_related_info
Find information related to a specific file or ticket.

```json
{
  "method": "tools/call", 
  "params": {
    "name": "context.get_related_info",
    "arguments": {
      "file_path": "internal/auth/service.go"
    }
  }
}
```

### context.index_control
Control indexing operations.

```json
{
  "method": "tools/call",
  "params": {
    "name": "context.index_control", 
    "arguments": {
      "action": "status"
    }
  }
}
```

### context.connector_management
Manage data source connectors.

```json
{
  "method": "tools/call",
  "params": {
    "name": "context.connector_management",
    "arguments": {
      "action": "list"
    }
  }
}
```

## Troubleshooting

### Binary won't start in stdio mode
- Ensure config.yml has `port: 0` or use `CONEXUS_PORT=0`
- Check that port validation allows 0 (modified in this build)

### No search results
- Run `./check_db.go` to verify database has indexed documents
- Check that indexer has processed files: `context.index_control status`

### MCP client connection fails
- Verify binary path in config is correct
- Check that database path is accessible
- Test manually: `echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/conexus`
