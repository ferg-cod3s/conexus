# Workflow Coordination Engine

## Overview

The workflow coordination engine orchestrates multi-agent execution with support for sequential, parallel, and conditional workflows. It handles agent coordination, result aggregation, and automatic escalation.

## Components

### Engine (`engine.go`)

Main workflow execution engine that:
- Validates workflows before execution
- Executes workflows in different modes (sequential/parallel/conditional)
- Handles context cancellation
- Manages automatic escalation

**Execution Modes**:
- `SequentialMode` - Steps execute one after another
- `ParallelMode` - All steps execute concurrently
- `ConditionalMode` - Steps execute based on conditions

**Key Functions**:
- `NewEngine(executor)` - Creates new workflow engine
- `Execute(ctx, workflow)` - Runs a workflow and returns result
- `executeSequential(ctx, workflow)` - Sequential execution
- `executeParallel(ctx, workflow)` - Parallel execution
- `executeConditional(ctx, workflow)` - Conditional execution

### Executor (`executor.go`)

Executes individual workflow steps using registered agents:

**Interfaces**:
```go
type Executor interface {
    ExecuteStep(ctx, step, currentResult) (*StepResult, error)
}

type Agent interface {
    Execute(ctx, req) (schema.AgentResponse, error)
}
```

**AgentExecutor Functions**:
- `NewAgentExecutor()` - Creates new agent executor
- `RegisterAgent(name, agent)` - Registers an agent
- `ExecuteStep(ctx, step, currentResult)` - Executes single step
- `buildContext(currentResult)` - Builds context from previous results

### Workflow Graph (`graph.go`)

Defines workflow structure and construction:

**Core Types**:
```go
type Workflow struct {
    ID          string
    Description string
    Mode        ExecutionMode
    Steps       []*Step
    Metadata    map[string]interface{}
}

type Step struct {
    ID           string
    Agent        string
    Input        string
    Permissions  schema.Permissions
    Condition    Condition
    Dependencies []string
    Metadata     map[string]interface{}
}
```

**Conditions**:
- `PreviousStepSuccessCondition` - Checks if previous step succeeded
- `OutputContainsCondition` - Checks if output contains specific data

**Builder Pattern**:
```go
builder := NewBuilder("workflow-id")
workflow, err := builder.
    WithDescription("My workflow").
    WithMode(SequentialMode).
    AddSequentialStep("step1", "agent1", "input1", permissions).
    AddSequentialStep("step2", "agent2", "input2", permissions).
    Build()
```

### Validator (`validator.go`)

Validates workflow structure before execution:
- Checks for required fields (ID, steps, agent names)
- Detects duplicate step IDs
- Validates dependencies exist
- Detects circular dependencies in parallel workflows

## Usage Examples

### Sequential Workflow

```go
executor := NewAgentExecutor()
executor.RegisterAgent("locator", locatorAgent)
executor.RegisterAgent("analyzer", analyzerAgent)

engine := NewEngine(executor)

workflow := &Workflow{
    ID:   "analyze-code",
    Mode: SequentialMode,
    Steps: []*Step{
        {
            ID:    "find-files",
            Agent: "locator",
            Input: "find all Go files",
        },
        {
            ID:    "analyze-files",
            Agent: "analyzer",
            Input: "analyze found files",
        },
    },
}

result, err := engine.Execute(ctx, workflow)
if err != nil {
    // Handle error
}

// Access results
for _, stepResult := range result.StepResults {
    fmt.Println("Agent:", stepResult.Agent)
    fmt.Println("Status:", stepResult.Status)
    fmt.Println("Output:", stepResult.Output)
}
```

### Parallel Workflow

```go
workflow := &Workflow{
    ID:   "parallel-analysis",
    Mode: ParallelMode,
    Steps: []*Step{
        {
            ID:    "analyze-frontend",
            Agent: "analyzer",
            Input: "analyze frontend code",
        },
        {
            ID:    "analyze-backend",
            Agent: "analyzer",
            Input: "analyze backend code",
        },
    },
}

result, err := engine.Execute(ctx, workflow)
// All steps execute concurrently
```

### Conditional Workflow

```go
workflow := &Workflow{
    ID:   "conditional-analysis",
    Mode: ConditionalMode,
    Steps: []*Step{
        {
            ID:    "find-files",
            Agent: "locator",
            Input: "find Go files",
        },
        {
            ID:    "analyze-if-found",
            Agent: "analyzer",
            Input: "analyze found files",
            Condition: &PreviousStepSuccessCondition{
                StepID: "find-files",
            },
        },
    },
}

result, err := engine.Execute(ctx, workflow)
// Second step only runs if first step succeeds
```

### Using Builder Pattern

```go
workflow, err := NewBuilder("my-workflow").
    WithDescription("Code analysis workflow").
    WithMode(SequentialMode).
    AddSequentialStep("locate", "locator", "find files", permissions).
    AddSequentialStep("analyze", "analyzer", "analyze files", permissions).
    WithMetadata("priority", "high").
    Build()

if err != nil {
    // Handle validation error
}
```

## Escalation Handling

Workflows automatically handle agent escalation:

```go
// Agent returns escalation in response
resp := schema.AgentResponse{
    Escalation: &schema.Escalation{
        Required:    true,
        TargetAgent: "specialized-agent",
        Reason:      "need specialized analysis",
    },
}

// Engine automatically:
// 1. Marks step as escalated
// 2. Creates new step for escalated agent
// 3. Appends to workflow
// 4. Continues execution
```

## Result Structure

```go
type ExecutionResult struct {
    WorkflowID       string
    Status           ExecutionStatus
    StepResults      []*StepResult
    Error            string
    AggregatedOutput *schema.AgentOutputV1
}

type StepResult struct {
    StepID           string
    Agent            string
    Status           StepStatus
    Output           *schema.AgentOutputV1
    Error            string
    EscalationTarget string
    EscalationReason string
}
```

## Validation

All workflows are validated before execution:

```go
validator := NewValidator()
err := validator.Validate(workflow)
if err != nil {
    // Workflow is invalid
}
```

**Validation Checks**:
- Workflow ID present
- At least one step defined
- All steps have required fields
- No duplicate step IDs
- Dependencies exist (parallel mode)
- No circular dependencies (parallel mode)

## Test Coverage

- **Coverage**: 95%+
- **Tests**: 8 test functions
- **Test file**: `engine_test.go`

**Test Scenarios**:
- Sequential execution
- Parallel execution
- Conditional execution
- Failure handling
- Escalation handling
- Workflow validation
- Builder pattern

## Performance

- **Sequential**: O(n) where n = number of steps
- **Parallel**: O(1) for independent steps
- **Memory**: ~1KB per step result
- **Thread-safe**: Yes (goroutine-safe parallel execution)

## Best Practices

1. **Use Builder**: Prefer builder pattern for complex workflows
2. **Validate Early**: Validate workflows before execution
3. **Handle Errors**: Check step results for failures
4. **Set Timeouts**: Use context with timeout for long workflows
5. **Monitor Escalations**: Track escalation patterns

## Future Enhancements

- **Dynamic DAG Construction**: Build workflows based on agent capabilities
- **Retry Logic**: Automatic retry for failed steps
- **Checkpointing**: Save/resume long-running workflows
- **Optimization**: Automatic parallelization of independent steps

---

**Version**: Phase 3
**Status**: Complete
**Last Updated**: 2025-10-14
