# Orchestrator

## Overview

The orchestrator is the central coordination component of Conexus that routes user requests to appropriate agents, manages workflow execution, and handles agent communication.

## Responsibilities

### Request Routing
- Parse user requests to determine intent
- Select appropriate agent(s) to handle the request
- Extract parameters from natural language

### Agent Invocation
- Create agent instances via factory registry
- Build `AgentRequest` with proper context
- Execute agents with permission enforcement

### Workflow Management
- Execute agents sequentially
- Pass context between agent invocations
- Handle agent escalations
- Aggregate results from multiple agents

### Error Handling
- Catch agent execution errors
- Provide user-friendly error messages
- Implement graceful degradation
- Support error recovery strategies

## Architecture

```
┌─────────────────────────────────────────┐
│           Orchestrator                  │
│                                         │
│  ┌────────────┐      ┌──────────────┐  │
│  │   Router   │──────│ Agent        │  │
│  │            │      │ Registry     │  │
│  └────────────┘      └──────────────┘  │
│         │                    │          │
│         │                    │          │
│  ┌──────▼────────────────────▼───────┐ │
│  │      Workflow Engine              │ │
│  │  - Sequential execution           │ │
│  │  - Context propagation            │ │
│  │  - Escalation handling            │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
         │                    │
         ▼                    ▼
   ┌──────────┐        ┌──────────┐
   │ Locator  │        │ Analyzer │
   │  Agent   │        │  Agent   │
   └──────────┘        └──────────┘
```

## Usage

### Basic Setup

```go
import (
    "context"
    "github.com/ferg-cod3s/conexus/internal/orchestrator"
    "github.com/ferg-cod3s/conexus/internal/process"
    "github.com/ferg-cod3s/conexus/internal/tool"
    "github.com/ferg-cod3s/conexus/internal/agent/locator"
    "github.com/ferg-cod3s/conexus/internal/agent/analyzer"
)

// Create dependencies
pm := process.NewManager()
te := tool.NewExecutor()

// Create orchestrator
orch := orchestrator.New(pm, te)

// Register agents
orch.RegisterAgent("codebase-locator", func(executor *tool.Executor) orchestrator.Agent {
    return locator.New(executor)
})

orch.RegisterAgent("codebase-analyzer", func(executor *tool.Executor) orchestrator.Agent {
    return analyzer.New(executor)
})
```

### Handle User Request

```go
ctx := context.Background()
userRequest := "find all .go files"

permissions := schema.Permissions{
    AllowedDirectories: []string{"/path/to/project"},
    ReadOnly:           true,
    MaxFileSize:        1024 * 1024,
    MaxExecutionTime:   30,
}

result, err := orch.HandleRequest(ctx, userRequest, permissions)
if err != nil {
    log.Fatalf("Request failed: %v", err)
}

if result.Success {
    for _, response := range result.Responses {
        fmt.Printf("Agent %s completed in %v\n",
            response.AgentID, response.Duration)
    }
}
```

### Execute Custom Workflow

```go
workflow := &orchestrator.Workflow{
    Steps: []orchestrator.WorkflowStep{
        {
            AgentID: "codebase-locator",
            Request: "find all .go files",
        },
        {
            AgentID: "codebase-analyzer",
            Request: "analyze main.go",
            Files:   []string{"/path/to/main.go"},
        },
    },
}

result, err := orch.ExecuteWorkflow(ctx, workflow, permissions)
```

## Request Routing

### Router Rules

The router uses keyword-based pattern matching:

```go
type RoutingRule struct {
    Keywords []string
    AgentID  string
    Priority int
}
```

### Default Rules

**Locator Agent:**
- Keywords: find, locate, search, files, where
- Priority: 10

**Analyzer Agent:**
- Keywords: analyze, how, works, flow, calls, understand
- Priority: 10

### Custom Rules

```go
router := orch.router
router.AddRule(orchestrator.RoutingRule{
    Keywords: []string{"test", "coverage"},
    AgentID:  "test-analyzer",
    Priority: 15,
})
```

### Routing Examples

| User Request | Routed To |
|--------------|-----------|
| "find all .go files" | codebase-locator |
| "where is main function" | codebase-locator |
| "analyze this code" | codebase-analyzer |
| "how does this work" | codebase-analyzer |
| "understand the flow" | codebase-analyzer |

## Workflow Execution

### Sequential Workflow

Agents execute one after another:

```
User Request
     │
     ▼
┌────────┐     ┌──────────┐     ┌──────────┐
│ Router │────▶│ Agent 1  │────▶│ Agent 2  │
└────────┘     └──────────┘     └──────────┘
                    │                 │
                    ▼                 ▼
                Response 1        Response 2
```

### Context Propagation

Each agent receives context from previous agents:

```go
type ConversationContext struct {
    UserRequest        string
    PreviousAgents     []string
    AccumulatedContext map[string]interface{}
}
```

### Agent Escalation

Agents can request additional help:

```go
response := schema.AgentResponse{
    Status: schema.StatusEscalationRequired,
    Escalation: &schema.Escalation{
        Required:     true,
        TargetAgent:  "codebase-analyzer",
        Reason:       "Need detailed analysis",
        RequiredInfo: "Analyze function behavior",
    },
}
```

The orchestrator automatically adds the escalated agent to the workflow.

## Result Format

```go
type Result struct {
    Success   bool
    Responses []schema.AgentResponse
    Error     string
    Duration  time.Duration
}
```

### Success Result

```json
{
  "success": true,
  "responses": [
    {
      "request_id": "req-12345",
      "agent_id": "codebase-locator",
      "status": "complete",
      "output": {...},
      "duration": 1250000000
    }
  ],
  "duration": 1300000000
}
```

### Error Result

```json
{
  "success": false,
  "responses": [...],
  "error": "Agent execution failed: permission denied",
  "duration": 500000000
}
```

## Agent Registry

### Registration

```go
type AgentFactory func(executor *tool.Executor) Agent

orch.RegisterAgent("my-agent", func(executor *tool.Executor) Agent {
    return myagent.New(executor)
})
```

### Agent Interface

```go
type Agent interface {
    Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error)
}
```

All agents must implement this interface.

## Error Handling

### Error Types

1. **Routing Errors**: Unable to determine appropriate agent
2. **Agent Not Found**: Requested agent not registered
3. **Agent Execution Errors**: Agent failed during execution
4. **Permission Errors**: Access denied to resources

### Error Recovery

```go
// Agent returns error
if response.Status == schema.StatusError {
    if response.Error.Recoverable {
        // Retry with different parameters
    } else {
        // Fail workflow
    }
}
```

## Performance

### Metrics

- **Routing Overhead**: <10ms per request
- **Context Propagation**: <5ms between agents
- **Total Overhead**: <100ms for full workflow

### Optimization

- Agent factories cached after first creation
- Minimal context copying
- Efficient parameter extraction

## Testing

### Unit Tests

```bash
go test ./internal/orchestrator -v
```

### Test Coverage

- Request routing (various user inputs)
- Sequential workflow execution
- Error handling and recovery
- Agent escalation
- Context propagation

### Mock Agents

Tests use `MockAgent` for controlled testing:

```go
type MockAgent struct {
    ExecuteFunc func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error)
}
```

## Configuration

### Default Configuration

```go
orchestrator.New(processManager, toolExecutor)
```

### Custom Router

```go
orch := orchestrator.New(pm, te)
orch.router = orchestrator.NewRouter() // Custom router
```

## Integration

### With Process Manager

Future: Agents run as isolated processes

```go
// Spawn agent process
proc, err := pm.Spawn(ctx, agentBinary, agentConfig)

// Communicate via JSON-RPC
request := buildJSONRPCRequest(agentReq)
proc.Stdin.Write(request)
```

### With Tool Executor

Agents receive shared tool executor:

```go
factory := func(executor *tool.Executor) Agent {
    return agent.New(executor) // Shared executor
}
```

## Limitations

### Current (Phase 2)

- Sequential execution only (no parallelization)
- Simple keyword-based routing (no NLP)
- In-process agent invocation (not isolated)
- Limited context accumulation
- No workflow persistence

### Future Enhancements (Phase 3+)

- [ ] Parallel agent execution
- [ ] Advanced intent parsing (NLU/LLM)
- [ ] Process-based agent isolation
- [ ] Workflow state persistence
- [ ] Dynamic workflow generation
- [ ] Conditional branching
- [ ] Loop detection and prevention

## Examples

### Example 1: Simple File Search

```go
ctx := context.Background()
perms := schema.Permissions{
    AllowedDirectories: []string{"/project"},
    ReadOnly:           true,
}

result, _ := orch.HandleRequest(ctx, "find all *.go files", perms)
// Routes to: codebase-locator
// Returns: List of Go files
```

### Example 2: Code Analysis

```go
result, _ := orch.HandleRequest(ctx, "analyze how Calculate works", perms)
// Routes to: codebase-analyzer
// Returns: AGENT_OUTPUT_V1 with call graph, data flow, etc.
```

### Example 3: Multi-Step Workflow

```go
workflow := &Workflow{
    Steps: []WorkflowStep{
        {AgentID: "codebase-locator", Request: "find main.go"},
        {AgentID: "codebase-analyzer", Request: "analyze main.go"},
    },
}

result, _ := orch.ExecuteWorkflow(ctx, workflow, perms)
// Executes both agents sequentially
// Returns: Combined results
```

## Best Practices

### 1. Register All Agents Before Use

```go
orch.RegisterAgent("agent-id", factory)
// Then use orchestrator
```

### 2. Handle Errors Gracefully

```go
result, err := orch.HandleRequest(ctx, request, perms)
if err != nil {
    log.Printf("Orchestration error: %v", err)
    return
}
if !result.Success {
    log.Printf("Workflow failed: %s", result.Error)
}
```

### 3. Set Appropriate Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := orch.HandleRequest(ctx, request, perms)
```

### 4. Validate Permissions

```go
perms := schema.Permissions{
    AllowedDirectories: validateDirectories(userInput),
    ReadOnly:           true,
    MaxFileSize:        10 * 1024 * 1024, // 10MB
    MaxExecutionTime:   60, // 60 seconds
}
```

## Debugging

### Enable Debug Logging

```go
// Add logging to agent execution
fmt.Printf("Executing agent: %s\n", step.AgentID)
fmt.Printf("Request: %s\n", step.Request)
```

### Trace Workflow Execution

```go
for i, response := range result.Responses {
    fmt.Printf("Step %d: %s (%s) - %v\n",
        i+1, response.AgentID, response.Status, response.Duration)
}
```

## Dependencies

- `internal/process` - Process management (future)
- `internal/tool` - Tool execution
- `pkg/schema` - Request/Response types
- Go standard library

## Support

For issues or questions:
- Review test files for usage patterns
- Check PHASE2-PLAN.md for architecture context
- See agent implementations for integration examples
