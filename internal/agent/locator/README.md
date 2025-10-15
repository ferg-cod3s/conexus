# Codebase Locator Agent

## Overview

The `codebase-locator` agent is responsible for discovering files and symbols within a codebase. It provides fast, pattern-based search capabilities using the underlying tool execution framework.

## Capabilities

### File Discovery
- Pattern-based file matching (glob syntax)
- Multi-directory search
- Extension filtering
- Permission-enforced access

### Symbol Search
- Function declarations
- Type definitions (future)
- Interface declarations (future)
- Export statements (future)

## Usage

```go
import (
    "context"
    "github.com/ferg-cod3s/conexus/internal/agent/locator"
    "github.com/ferg-cod3s/conexus/internal/tool"
    "github.com/ferg-cod3s/conexus/pkg/schema"
)

// Create executor and agent
executor := tool.NewExecutor()
agent := locator.New(executor)

// Create request
req := schema.AgentRequest{
    RequestID: "req-001",
    AgentID:   "codebase-locator",
    Task: schema.AgentTask{
        SpecificRequest:    "find all .go files",
        AllowedDirectories: []string{"/path/to/project"},
    },
    Permissions: schema.Permissions{
        AllowedDirectories: []string{"/path/to/project"},
        ReadOnly:           true,
    },
}

// Execute
ctx := context.Background()
response, err := agent.Execute(ctx, req)
```

## Request Format

### File Search Request
```json
{
  "request_id": "req-001",
  "agent_id": "codebase-locator",
  "task": {
    "specific_request": "find all *.go files",
    "allowed_directories": ["/path/to/project"]
  },
  "permissions": {
    "allowed_directories": ["/path/to/project"],
    "read_only": true
  }
}
```

### Symbol Search Request
```json
{
  "request_id": "req-002",
  "agent_id": "codebase-locator",
  "task": {
    "specific_request": "find function Calculate",
    "files": ["/path/to/file.go"],
    "allowed_directories": ["/path/to/project"]
  },
  "permissions": {
    "allowed_directories": ["/path/to/project"],
    "read_only": true
  }
}
```

## Output Format

The agent returns `AGENT_OUTPUT_V1` formatted responses:

```json
{
  "version": "AGENT_OUTPUT_V1",
  "component_name": "File Discovery",
  "scope_description": "File search in directories: [/path]",
  "overview": "Locates files matching specified patterns",
  "entry_points": [
    {
      "file": "/path/to/file.go",
      "lines": "1-1",
      "symbol": "file.go",
      "role": "file"
    }
  ],
  "raw_evidence": [
    {
      "claim": "File found: file.go",
      "file": "/path/to/file.go",
      "lines": "1-1"
    }
  ],
  "limitations": [
    "Pattern matching uses glob syntax only"
  ]
}
```

## Implementation Details

### Search Strategies

1. **Glob Strategy** (Current)
   - Uses `glob` tool for pattern matching
   - Fast file system traversal
   - Wildcard support (*, **, ?)

2. **Grep Strategy** (Partial)
   - Text-based content search
   - Function declaration matching
   - Limited by grep tool implementation

3. **AST Strategy** (Future - Phase 5)
   - Go AST parsing
   - Precise symbol extraction
   - Type-aware searching

### Limitations

- Grep tool is currently a placeholder (needs full implementation)
- No line number extraction from grep results yet
- Symbol search is basic text matching only
- No support for complex symbol queries
- AST-based analysis deferred to Phase 5

### Future Enhancements

- [ ] Complete grep tool implementation with line numbers
- [ ] Add support for interface and type searches
- [ ] Implement caching for repeated searches
- [ ] Add fuzzy matching for symbol names
- [ ] AST-based Go code analysis
- [ ] Support for multiple programming languages
- [ ] Semantic search capabilities

## Testing

Test fixtures are available in `tests/fixtures/`:
- `simple_function.go` - Basic function declarations
- `multiple_functions.go` - Function call chains
- `struct_methods.go` - Method declarations
- `error_handling.go` - Error handling patterns
- `side_effects.go` - Functions with side effects

Run tests:
```bash
go test ./internal/agent/locator -v
```

## Dependencies

- `internal/tool` - Tool execution framework
- `pkg/schema` - AGENT_OUTPUT_V1 types

## Performance

Target metrics:
- File search: <2s for 1000 file repository
- Symbol search: <3s for 100 files
- Memory usage: <100MB per search

## Error Handling

The agent handles errors gracefully:
- Invalid directories are skipped
- Permission errors return appropriate status
- Empty results are valid (no matches found)
- Tool execution errors are caught and logged

## Security

- All file access is permission-validated
- Only allowed directories are searched
- Read-only operations enforced
- No arbitrary code execution
