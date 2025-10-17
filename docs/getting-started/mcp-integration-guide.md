# MCP Integration Guide

## Overview

Conexus implements the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) to enable AI assistants like Claude to access your codebase context, discussions, and project knowledge. This guide covers how to integrate Conexus with MCP-compatible clients.

## Quick Start

### 1. Start the Conexus Server

```bash
# Build the server
go build ./cmd/conexus

# Run with MCP protocol (stdio mode)
./conexus
```

The server communicates via JSON-RPC 2.0 over stdin/stdout.

### 2. Configure Claude Desktop

Add Conexus to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "conexus": {
      "command": "/path/to/conexus",
      "args": [],
      "env": {
        "CONEXUS_CONFIG": "/path/to/config.yml"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

Claude will automatically connect to Conexus and make the tools available.

## Available Tools

Conexus exposes 4 MCP tools for context retrieval and management:

### 1. `context.search`

Performs semantic search across your codebase, discussions, and documents.

**Features**:
- Hybrid search (vector similarity + BM25 keyword matching)
- Work context awareness (active file, git branch, open tickets)
- Source type filtering (files, Slack, GitHub, Jira)
- Date range filtering

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "context.search",
    "arguments": {
      "query": "how does authentication work",
      "work_context": {
        "active_file": "internal/auth/handler.go",
        "git_branch": "feature/oauth",
        "open_ticket_ids": ["PROJ-123"]
      },
      "top_k": 10,
      "filters": {
        "source_types": ["file", "github"],
        "date_range": {
          "from": "2025-01-01T00:00:00Z",
          "to": "2025-10-16T23:59:59Z"
        }
      }
    }
  }
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"results\":[{\"id\":\"chunk-123\",\"content\":\"// Auth middleware validates JWT tokens...\",\"score\":0.95,\"source_type\":\"file\",\"metadata\":{\"file_path\":\"internal/auth/middleware.go\",\"line_number\":42}}],\"total_count\":10,\"query_time_ms\":23.5}"
      }
    ]
  }
}
```

**Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | - | Natural language search query |
| `work_context` | object | No | null | Current working context |
| `work_context.active_file` | string | No | - | File currently being edited |
| `work_context.git_branch` | string | No | - | Current git branch |
| `work_context.open_ticket_ids` | array | No | - | Open ticket/issue IDs |
| `top_k` | integer | No | 20 | Max results to return (max: 100) |
| `filters` | object | No | null | Search filters |
| `filters.source_types` | array | No | all | Filter by source: `["file", "slack", "github", "jira"]` |
| `filters.date_range` | object | No | null | ISO 8601 date range |

**Use Cases**:
- "Find all authentication-related code"
- "Show me discussions about database migration"
- "Search for error handling patterns in our codebase"

---

### 2. `context.get_related_info`

Retrieves information directly related to a file or ticket.

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "context.get_related_info",
    "arguments": {
      "file_path": "internal/mcp/server.go"
    }
  }
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"summary\":\"MCP server implementation with 4 tools\",\"related_prs\":[\"#42\",\"#38\"],\"related_issues\":[\"#35\"],\"discussions\":[{\"channel\":\"#dev\",\"timestamp\":\"2025-01-15T10:30:00Z\",\"summary\":\"Discussed MCP protocol implementation\"}]}"
      }
    ]
  }
}
```

**Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_path` | string | No* | Path to file (relative to repo root) |
| `ticket_id` | string | No* | Ticket/issue ID (e.g., "PROJ-123") |

\* *At least one parameter required*

**Use Cases**:
- "What's the history of this file?"
- "Show me PRs related to this issue"
- "Find discussions about this component"

---

### 3. `context.index_control`

Controls indexing operations.

**⚠️ Status**: Partially implemented (only `status` action works)

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "context.index_control",
    "arguments": {
      "action": "status"
    }
  }
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"status\":\"running\",\"message\":\"Indexer is running\",\"details\":{\"documents_indexed\":\"1234\"}}"
      }
    ]
  }
}
```

**Supported Actions**:
- ✅ `status` - Get current indexing status
- ⚠️ `start` - Not implemented
- ⚠️ `stop` - Not implemented
- ⚠️ `force_reindex` - Not implemented

---

### 4. `context.connector_management`

Manages data source connectors.

**⚠️ Status**: Placeholder implementation (returns hardcoded data)

**Request**:
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "context.connector_management",
    "arguments": {
      "action": "list"
    }
  }
}
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"connectors\":[{\"id\":\"local-files\",\"type\":\"filesystem\",\"name\":\"Local Files\",\"status\":\"active\",\"config\":{}}],\"status\":\"success\"}"
      }
    ]
  }
}
```

**Supported Actions**:
- ⚠️ `list` - Returns hardcoded "local-files" connector
- ⚠️ `add` - Not implemented
- ⚠️ `update` - Not implemented
- ⚠️ `remove` - Not implemented

---

## Protocol Details

### Transport

Conexus uses **JSON-RPC 2.0** over **stdin/stdout** (stdio transport).

### Authentication

**Current**: No authentication required.

**Future**: Will support:
- API key authentication
- OAuth 2.0 for team workspaces
- Per-user access control

### Error Handling

Errors follow JSON-RPC 2.0 specification:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32600,
    "message": "Invalid request",
    "data": {
      "details": "Missing required parameter: query"
    }
  }
}
```

**Error Codes**:

| Code | Meaning | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON |
| -32600 | Invalid request | Missing required fields |
| -32601 | Method not found | Unknown RPC method |
| -32602 | Invalid params | Invalid parameter types |
| -32603 | Internal error | Server-side error |

### Resources API

**⚠️ Status**: Not fully implemented

The MCP spec includes a resources API for listing and reading indexed content:

- `resources/list` - Returns empty list
- `resources/read` - Not implemented

These will be implemented in a future release.

---

## Integration Examples

### TypeScript/Node.js Client

```typescript
import { spawn } from 'child_process';
import { stdin, stdout } from 'process';

// Spawn Conexus server
const conexus = spawn('/path/to/conexus', [], {
  env: { CONEXUS_CONFIG: '/path/to/config.yml' }
});

// Send JSON-RPC request
const request = {
  jsonrpc: '2.0',
  id: 1,
  method: 'tools/call',
  params: {
    name: 'context.search',
    arguments: {
      query: 'authentication implementation',
      top_k: 5
    }
  }
};

conexus.stdin.write(JSON.stringify(request) + '\n');

// Read response
conexus.stdout.on('data', (data) => {
  const response = JSON.parse(data.toString());
  console.log(response.result);
});
```

### Python Client

```python
import subprocess
import json

# Start Conexus
proc = subprocess.Popen(
    ['/path/to/conexus'],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    env={'CONEXUS_CONFIG': '/path/to/config.yml'}
)

# Send request
request = {
    'jsonrpc': '2.0',
    'id': 1,
    'method': 'tools/call',
    'params': {
        'name': 'context.search',
        'arguments': {
            'query': 'error handling patterns',
            'top_k': 10
        }
    }
}

proc.stdin.write(json.dumps(request).encode() + b'\n')
proc.stdin.flush()

# Read response
response = json.loads(proc.stdout.readline())
print(response['result'])
```

### cURL (for testing)

MCP uses stdio, so direct cURL testing requires a wrapper. Use this test script:

```bash
#!/bin/bash
# test-mcp.sh

REQUEST='{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
echo "$REQUEST" | ./conexus
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONEXUS_CONFIG` | Path to config file | `config.yml` |
| `CONEXUS_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `CONEXUS_METRICS_PORT` | Prometheus metrics port | `9090` |

### Config File

```yaml
# config.yml
embedding:
  provider: "openai"
  model: "text-embedding-3-small"
  api_key: "${OPENAI_API_KEY}"

vectorstore:
  type: "sqlite"
  path: "./data/conexus.db"

indexer:
  watch_paths:
    - "./internal"
    - "./pkg"
    - "./cmd"
  ignore_patterns:
    - "*.test.go"
    - "*_test.go"
    - "vendor/*"

observability:
  log_level: "info"
  metrics:
    enabled: true
    port: 9090
  tracing:
    enabled: false
```

---

## Best Practices

### 1. Provide Work Context

Always include work context when available for better results:

```json
{
  "query": "how does caching work",
  "work_context": {
    "active_file": "internal/cache/redis.go",
    "git_branch": "feature/cache-optimization"
  }
}
```

### 2. Use Appropriate top_k

- **Broad exploration**: `top_k: 20-50`
- **Focused search**: `top_k: 5-10`
- **Deep dive**: `top_k: 50-100`

### 3. Filter by Source Type

Narrow results to relevant sources:

```json
{
  "query": "deployment process",
  "filters": {
    "source_types": ["github", "slack"]
  }
}
```

### 4. Handle Errors Gracefully

Always check for JSON-RPC errors:

```typescript
if (response.error) {
  console.error(`Error ${response.error.code}: ${response.error.message}`);
  return;
}

const results = JSON.parse(response.result.content[0].text);
```

---

## Troubleshooting

### Claude Desktop Not Connecting

1. **Check config path**: Verify `claude_desktop_config.json` location
2. **Check Conexus path**: Ensure absolute path to binary
3. **Check logs**: 
   - macOS: `~/Library/Logs/Claude/mcp*.log`
   - Windows: `%APPDATA%\Claude\Logs\mcp*.log`
   - Linux: `~/.local/share/Claude/logs/mcp*.log`

### No Search Results

1. **Check indexing**: Use `context.index_control` with action `status`
2. **Verify config**: Ensure `watch_paths` includes target directories
3. **Check embeddings**: Verify OpenAI API key is valid

### Slow Search Performance

1. **Reduce top_k**: Lower number of results
2. **Add filters**: Narrow search by source type or date
3. **Check metrics**: View `/metrics` endpoint for performance data

### Connection Refused

1. **Check process**: Ensure Conexus is running
2. **Check stdio**: MCP requires stdin/stdout (not HTTP)
3. **Check permissions**: Verify binary is executable

---

## Limitations & Roadmap

### Current Limitations

- ❌ No authentication/authorization
- ❌ Resources API not implemented
- ❌ Index control limited to status checks
- ❌ Connector management is placeholder
- ❌ No multi-tenant support

### Planned Features

- ✅ Authentication (API keys, OAuth)
- ✅ Full index control (start/stop/reindex)
- ✅ Dynamic connector management
- ✅ Resources API implementation
- ✅ Multi-workspace support
- ✅ Real-time indexing updates
- ✅ Query result caching

---

## API Reference

For complete protocol details, see:
- [MCP Specification](https://modelcontextprotocol.io/docs/spec)
- [JSON-RPC 2.0 Spec](https://www.jsonrpc.org/specification)
- [Conexus API Specification](../API-Specification.md)

For implementation details, see:
- `internal/mcp/server.go` - MCP server
- `internal/mcp/handlers.go` - Tool implementations
- `internal/mcp/schema.go` - Type definitions
- `internal/protocol/jsonrpc.go` - Protocol layer

---

## Support

- **Issues**: [GitHub Issues](https://github.com/ferg-cod3s/conexus/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ferg-cod3s/conexus/discussions)
- **Documentation**: [docs/README.md](../README.md)
