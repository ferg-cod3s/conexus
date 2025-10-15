# MCP Package

## Overview
Implements the Model Context Protocol server for exposing Conexus to LLM agents.

## Key Interfaces

### `Server`
MCP protocol server:
- `Serve()` - Start server (blocking)
- `RegisterTool()` - Add tool handlers
- `RegisterResource()` - Add resource handlers

### `Tool`
MCP tool (function):
- `Name()` - Tool identifier (e.g., `context.search`)
- `Schema()` - JSON schema for parameters
- `Execute()` - Run the tool

### `Resource`
MCP resource (read-only data):
- `URI()` - URI pattern (e.g., `codebase://{path}`)
- `Read()` - Retrieve content

### `Transport`
Message transport abstraction (stdio, HTTP, SSE).

## Tools

### `context.search`
Search the indexed codebase:
```json
{
  "query": "authentication middleware",
  "limit": 10,
  "filters": {"language": "go"}
}
```

### `context.index` (optional)
Trigger indexing:
```json
{
  "path": "/path/to/repo",
  "incremental": true
}
```

## Resources

### `codebase://{path}`
Retrieve file content by path:
- `codebase://main.go` → file content
- `codebase://internal/` → directory listing

### `docs://{path}`
Documentation resources.

## Usage Example

```go
import "github.com/ferg-cod3s/conexus/internal/mcp"

transport := &mcp.StdioTransport{
    In:  os.Stdin,
    Out: os.Stdout,
}

server := mcp.NewServer(transport)

// Register tools
server.RegisterTool(&SearchTool{pipeline: searchPipeline})

// Start server
err := server.Serve(ctx)
```

## Protocol

### Request
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "context.search",
    "arguments": {"query": "auth", "limit": 5}
  }
}
```

### Response
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {"type": "text", "text": "...search results..."}
    ]
  }
}
```

## Implementation Status
- [ ] JSON-RPC 2.0 handler
- [ ] Stdio transport
- [ ] Search tool
- [ ] Resource handlers
- [ ] Unit tests
