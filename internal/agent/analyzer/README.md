# Codebase Analyzer Agent

## Overview

The `codebase-analyzer` agent performs deep analysis of source code to understand control flow, data flow, state management, side effects, and design patterns. It produces comprehensive `AGENT_OUTPUT_V1` formatted outputs with 100% evidence backing.

## Capabilities

### Entry Point Identification
- Exported functions (starting with capital letter)
- Methods on structs
- Function and method signatures

### Call Graph Construction
- Function-to-function relationships
- Method invocations
- Line-level precision for call sites

### Data Flow Analysis
- Input identification (parameters)
- Data transformations (assignments, operations)
- Output tracking (return statements)

### State Management Detection
- Struct field assignments
- Variable mutations
- Memory operations

### Side Effect Tracking
- Logging (fmt.Print*, log.Print*)
- File I/O (os.ReadFile, os.WriteFile)
- HTTP operations
- External interactions

### Error Handling Analysis
- Error return patterns
- Error checks (if err != nil)
- Guard clauses
- Error propagation

### Pattern Detection
- Factory pattern (New* functions)
- Method receivers
- Common Go idioms

### Concurrency Analysis
- Goroutine launches
- Channel usage
- Mutex/RWMutex synchronization
- WaitGroup coordination

## Usage

```go
import (
    "context"
    "github.com/ferg-cod3s/conexus/internal/agent/analyzer"
    "github.com/ferg-cod3s/conexus/internal/tool"
    "github.com/ferg-cod3s/conexus/pkg/schema"
)

// Create executor and agent
executor := tool.NewExecutor()
agent := analyzer.New(executor)

// Create request
req := schema.AgentRequest{
    RequestID: "req-001",
    AgentID:   "codebase-analyzer",
    Task: schema.AgentTask{
        Files:              []string{"/path/to/file.go"},
        AllowedDirectories: []string{"/path/to/project"},
        SpecificRequest:    "analyze this file",
    },
    Permissions: schema.Permissions{
        AllowedDirectories: []string{"/path/to/project"},
        ReadOnly:           true,
        MaxFileSize:        1024 * 1024,
    },
}

// Execute
ctx := context.Background()
response, err := agent.Execute(ctx, req)
```

## Request Format

```json
{
  "request_id": "req-001",
  "agent_id": "codebase-analyzer",
  "task": {
    "files": ["/path/to/file.go"],
    "allowed_directories": ["/path/to/project"],
    "specific_request": "analyze this file"
  },
  "permissions": {
    "allowed_directories": ["/path/to/project"],
    "read_only": true,
    "max_file_size": 1048576
  }
}
```

## Output Format

The agent returns comprehensive `AGENT_OUTPUT_V1` structured data:

```json
{
  "version": "AGENT_OUTPUT_V1",
  "component_name": "file.go",
  "scope_description": "Analysis of 1 file(s): [file.go]",
  "overview": "Analyzed code contains 3 entry point(s), 5 function call(s)...",
  "entry_points": [
    {
      "file": "/path/to/file.go",
      "lines": "5-5",
      "symbol": "Add",
      "role": "function"
    }
  ],
  "call_graph": [
    {
      "from": "/path/to/file.go:Calculate",
      "to": "/path/to/file.go:Process",
      "via_line": 12
    }
  ],
  "data_flow": {
    "inputs": [...],
    "transformations": [...],
    "outputs": [...]
  },
  "state_management": [...],
  "side_effects": [...],
  "error_handling": [...],
  "patterns": [...],
  "concurrency": [...],
  "raw_evidence": [
    {
      "claim": "Entry point: Add",
      "file": "/path/to/file.go",
      "lines": "5-5"
    }
  ]
}
```

## Analysis Methodology

### Text-Based Analysis (Current)

The current implementation uses regex-based pattern matching:

**Advantages:**
- Fast and lightweight
- No external dependencies
- Works with any Go code

**Limitations:**
- May miss complex patterns
- Limited type awareness
- Heuristic-based call graph

### Regex Patterns Used

- **Functions**: `^func\s+(\w+)\(`
- **Methods**: `^func\s+\(.*\)\s+(\w+)\(`
- **Function Calls**: `\b([A-Z]\w+)\(`
- **Assignments**: `^\s*(\w+)\s*:?=`
- **Returns**: `^\s*return\s+(.+)`
- **Error Checks**: `if\s+err\s*!=\s*nil`
- **Logging**: `fmt\.Print|log\.Print`
- **File I/O**: `os\.WriteFile|os\.ReadFile`

### Evidence Generation

Every claim in the output has corresponding evidence:

```go
// For each entry point found
evidence = Evidence{
    Claim: "Entry point: FunctionName",
    File:  "/absolute/path/to/file.go",
    Lines: "12-12",
}
```

## Testing

Test fixtures in `tests/fixtures/`:
- `simple_function.go` - Basic functions
- `multiple_functions.go` - Call graphs
- `struct_methods.go` - Methods and state
- `error_handling.go` - Error patterns
- `side_effects.go` - I/O and logging

Run tests:
```bash
go test ./internal/agent/analyzer -v
```

## Performance

Target metrics:
- Analysis speed: <5s for 500 LOC file
- Memory usage: <200MB per analysis
- Evidence generation: <100ms

Current performance (text-based):
- Very fast: <1s for most files
- Low memory: <50MB typical
- Scales linearly with file size

## Limitations

1. **Text-Based Analysis**: Not AST-aware, may miss complex patterns
2. **Line Numbers**: Approximate for multi-line constructs
3. **Type System**: No type information available
4. **Cross-File Analysis**: Limited to files in request
5. **Language Support**: Go only currently

## Future Enhancements

### Phase 5: AST-Based Analysis

- [ ] Go AST parsing via `go/parser`
- [ ] Precise type information
- [ ] Accurate control flow graphs
- [ ] Cross-package analysis
- [ ] Import graph construction

### Additional Features

- [ ] Interface implementation detection
- [ ] Dependency analysis
- [ ] Cyclomatic complexity calculation
- [ ] Test coverage mapping
- [ ] Performance hotspot identification
- [ ] Security vulnerability detection

### Multi-Language Support

- [ ] TypeScript/JavaScript analysis
- [ ] Python analysis
- [ ] Rust analysis
- [ ] Language-agnostic patterns

## Error Handling

The agent handles errors gracefully:

- **File Read Errors**: Skip file, continue with others
- **Parse Errors**: Mark as limitation, best-effort analysis
- **Permission Errors**: Return error with clear message
- **Empty Files**: Valid empty output with no entry points

## Security

- All file access permission-validated
- Read-only operations only
- No code execution
- Sandboxed tool execution

## Dependencies

- `internal/tool` - Tool execution framework (read, grep)
- `pkg/schema` - AGENT_OUTPUT_V1 types
- `regexp` - Pattern matching (standard library)
- `strings` - Text processing (standard library)

## Integration

The analyzer integrates with:

- **Orchestrator**: For workflow coordination
- **Locator**: To analyze files discovered by locator
- **Tool Executor**: For file reading operations
- **Process Manager**: For future subprocess isolation

## Examples

### Example 1: Analyze Simple Function

**Input:**
```go
func Add(a, b int) int {
    return a + b
}
```

**Output:**
- Entry point: Add (line 1)
- Data input: parameters (a, b int)
- Data output: return value (line 2)
- Evidence: 2 entries

### Example 2: Analyze Call Chain

**Input:**
```go
func Calculate(x int) int {
    result := Process(x)
    return Finalize(result)
}

func Process(v int) int { return v * 2 }
func Finalize(v int) int { return v + 10 }
```

**Output:**
- Entry points: Calculate, Process, Finalize
- Call graph: Calculate→Process, Calculate→Finalize
- Evidence: 5 entries

### Example 3: Analyze Error Handling

**Input:**
```go
func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

**Output:**
- Entry point: Divide
- Error handling: guard clause (line 2-4)
- Error handling: error return (line 3)
- Evidence: 3 entries

## Support

For issues, questions, or contributions:
- Check test files for usage examples
- Review PHASE2-PLAN.md for context
- See pkg/schema/agent_output_v1.go for output format
