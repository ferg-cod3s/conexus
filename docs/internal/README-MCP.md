# Conexus MCP Integration Guide

This guide shows how to integrate Conexus with Claude Code and OpenCode development environments using the Model Context Protocol (MCP).

## Prerequisites

- Conexus built and tested (see main README)
- Claude Code CLI installed
- Node.js for OpenCode (if applicable)

## Claude Code Integration

### 1. Create MCP Configuration

Create `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "conexus": {
      "command": "/path/to/conexus/conexus",
      "args": [],
      "env": {
        "CONEXUS_PORT": "8080",
        "CONEXUS_DB_PATH": "/path/to/conexus/data/conexus.db",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  }
}
```

Replace `/path/to/conexus` with your actual Conexus path.

### 2. Start Claude Code

```bash
claude
```

### 3. Test MCP Connection

In Claude Code, try:
```
/mcp conexus tools/list
```

You should see the 4 Conexus tools listed.

### 4. Use Conexus Tools

```bash
# Check index status
/mcp conexus tools/call context.index_control {"action": "status"}

# List connectors
/mcp conexus tools/call context.connector_management {"action": "list"}

# Search code (when indexing is implemented)
/mcp conexus tools/call context.search {"query": "function definitions"}
```

## OpenCode Integration

### 1. MCP Configuration

For OpenCode, create a configuration file or environment variable:

```bash
export MCP_SERVERS='{
  "conexus": {
    "command": "/path/to/conexus/conexus",
    "args": [],
    "env": {
      "CONEXUS_PORT": "8080",
      "CONEXUS_DB_PATH": "/path/to/conexus/data/conexus.db"
    }
  }
}'
```

### 2. Start OpenCode

```bash
opencode
```

### 3. Test Integration

Use the MCP tools through OpenCode's interface.

## Development Workflow

### Code Search & Analysis

1. **Index your codebase** (when implemented):
   ```
   /mcp conexus tools/call context.index_control {"action": "start"}
   ```

2. **Search for code patterns**:
   ```
   /mcp conexus tools/call context.search {"query": "authentication logic"}
   ```

3. **Find related information**:
   ```
   /mcp conexus tools/call context.get_related_info {"file_path": "auth.go"}
   ```

### Debugging & Troubleshooting

- **Check server health**: Visit `http://localhost:8080/health`
- **View MCP logs**: Check Conexus server logs
- **Test tools manually**: Use curl commands from main README

### Performance Monitoring

- **Response times**: Should be <100ms for most operations
- **Error rates**: Should be 0% under normal conditions
- **Index status**: Monitor document count and indexing progress

## Troubleshooting

### Connection Issues

**Problem**: "MCP server not responding"
**Solution**: 
1. Check if Conexus is running: `curl http://localhost:8080/health`
2. Verify MCP configuration paths
3. Check server logs for errors

**Problem**: "Tool not found"
**Solution**:
1. Verify tool name spelling
2. Check MCP server is connected: `/mcp conexus tools/list`

### Performance Issues

**Problem**: Slow responses
**Solution**:
1. Check system resources (CPU, memory)
2. Verify database is not corrupted
3. Check network latency

**Problem**: Search returns no results
**Solution**:
1. Confirm codebase is indexed
2. Check index status: `/mcp conexus tools/call context.index_control {"action": "status"}`

## Advanced Configuration

### Custom Port

```json
{
  "mcpServers": {
    "conexus": {
      "env": {
        "CONEXUS_PORT": "9090"
      }
    }
  }
}
```

### Database Location

```json
{
  "mcpServers": {
    "conexus": {
      "env": {
        "CONEXUS_DB_PATH": "/custom/path/conexus.db"
      }
    }
  }
}
```

### Logging Level

```json
{
  "mcpServers": {
    "conexus": {
      "env": {
        "CONEXUS_LOG_LEVEL": "debug"
      }
    }
  }
}
```

## Next Steps

1. **Implement codebase indexing** for full search functionality
2. **Add more MCP tools** for enhanced development workflow
3. **Integrate with CI/CD** for automated indexing
4. **Add team collaboration features** via shared indexes

---

**Status**: ✅ MCP Integration Ready
**Compatibility**: Claude Code, OpenCode
**Tools Available**: 4 MCP tools functional
