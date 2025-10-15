# Component Integration Architecture

## Overview

Conexus implements a **multi-agent orchestration system** where specialized agents collaborate to analyze codebases through a sophisticated workflow engine. This document describes how the core components integrate to provide reliable, validated, and performant code analysis.

### Core Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Request                              │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   JSON-RPC Protocol   │
                    │   (transport layer)   │
                    └───────────┬───────────┘
                                │
                                ▼
        ┌───────────────────────────────────────────────┐
        │          Orchestrator (Core)                  │
        │  • Request routing (Router)                   │
        │  • Agent registry & factory                   │
        │  • Workflow execution                         │
        │  • Escalation handling                        │
        └───────┬───────────────────────────────────────┘
                │
                ├─────────────┬──────────────┬──────────────┐
                ▼             ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
        │ Analyzer │  │ Locator  │  │ Custom   │  │ Future   │
        │  Agent   │  │  Agent   │  │  Agents  │  │  Agents  │
        └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘
              │             │              │              │
              └─────────────┴──────────────┴──────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   Tool Execution      │
                    │  • File operations    │
                    │  • AST parsing        │
                    │  • Pattern matching   │
                    └───────────┬───────────┘
                                │
                ┌───────────────┴────────────────┐
                ▼                                ▼
        ┌──────────────┐              ┌──────────────────┐
        │  Validation  │              │    Profiling     │
        │  • Evidence  │              │  • Metrics       │
        │  • Schema    │              │  • Bottlenecks   │
        └──────────────┘              └──────────────────┘
                │                                │
                └────────────┬───────────────────┘
                             ▼
                    ┌─────────────────┐
                    │  State Manager  │
                    │  • Sessions     │
                    │  • History      │
                    │  • Caching      │
                    └─────────────────┘
```

## Component Interaction

### 1. Orchestrator as Central Hub

**File**: `internal/orchestrator/orchestrator.go`

The Orchestrator is the central coordination point that:

```go
type Orchestrator struct {
    router    *Router              // Routes requests to appropriate agents
    agents    map[string]Agent     // Registered agent implementations
    factory   *AgentFactory        // Creates agent instances
    state     *state.Manager       // Manages sessions and history
    profiler  *profiling.Profiler  // Performance monitoring
}
```

**Key Responsibilities**:
- **Request Routing**: Keyword-based agent selection
- **Workflow Management**: Sequential, parallel, and conditional execution
- **Dynamic Escalation**: Runtime workflow modification
- **State Coordination**: Session and history management

#### Orchestrator Initialization

```go
func NewOrchestrator(cfg *Config) (*Orchestrator, error) {
    // 1. Create state manager for session tracking
    stateManager := state.NewManager(cfg.StateConfig)
    
    // 2. Initialize profiler for performance monitoring
    profiler := profiling.NewProfiler(cfg.ProfilingConfig)
    
    // 3. Register agents via factory pattern
    factory := NewAgentFactory()
    factory.Register("analyzer", analyzer.New)
    factory.Register("locator", locator.New)
    
    // 4. Build router with agent keyword mappings
    router := NewRouter(factory)
    router.AddRoute([]string{"analyze", "explain"}, "analyzer")
    router.AddRoute([]string{"find", "locate", "where"}, "locator")
    
    return &Orchestrator{
        router:   router,
        factory:  factory,
        state:    stateManager,
        profiler: profiler,
    }, nil
}
```

### 2. Request Processing Flow

```
User Request
     │
     ▼
┌─────────────────────────────────────────────────────────┐
│ 1. Protocol Layer (internal/protocol/jsonrpc.go)       │
│    • Parse JSON-RPC request                             │
│    • Validate method and parameters                     │
│    • Extract user intent and context                    │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 2. Orchestrator Entry (HandleRequest)                  │
│    • Create or retrieve session                         │
│    • Start performance profiling                        │
│    • Route to appropriate agent                         │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 3. Router Analysis (internal/orchestrator/router.go)   │
│    • Extract keywords from request                      │
│    • Match keywords to agent capabilities               │
│    • Select primary agent (analyzer/locator/custom)     │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 4. Workflow Creation                                    │
│    • Build WorkflowStep for selected agent              │
│    • Determine execution mode (sequential/parallel)     │
│    • Prepare agent context and parameters               │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 5. Agent Execution (invokeAgent)                       │
│    • Factory creates agent instance                     │
│    • Agent.Execute() with context and request           │
│    • Tool execution (file ops, AST parsing)             │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 6. Validation & Quality Gates                          │
│    • Schema validation (AgentOutputV1 structure)        │
│    • Evidence validation (coverage, backing)            │
│    • Error detection and handling                       │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 7. Escalation Check                                     │
│    • If escalation.Required == true:                    │
│      → Create new workflow steps                        │
│      → Execute escalated agents                         │
│      → Aggregate results                                │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 8. State Update                                         │
│    • Record history entry (with escalations)            │
│    • Update session state                               │
│    • Cache results if applicable                        │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ 9. Response Formatting                                  │
│    • Aggregate multi-agent results                      │
│    • Include profiling metrics                          │
│    • Format as JSON-RPC response                        │
└──────────────────────────┬──────────────────────────────┘
                           │
                           ▼
                    User Response
```

### 3. Workflow Execution Model

**File**: `internal/orchestrator/workflow/engine.go`

The workflow engine supports three execution modes:

```go
type ExecutionMode string

const (
    Sequential  ExecutionMode = "sequential"  // Steps run one after another
    Parallel    ExecutionMode = "parallel"    // Steps run concurrently
    Conditional ExecutionMode = "conditional" // Steps run based on conditions
)
```

#### Sequential Execution

```go
func (e *Engine) executeSequential(ctx context.Context, steps []WorkflowStep) error {
    for i, step := range steps {
        // Update step status
        step.Status = StatusRunning
        
        // Execute with timeout and cancellation support
        result, err := e.executeStep(ctx, step)
        if err != nil {
            step.Status = StatusFailed
            step.Error = err
            return fmt.Errorf("step %d failed: %w", i, err)
        }
        
        // Store result for next step
        step.Result = result
        step.Status = StatusCompleted
        
        // Handle dynamic escalations
        if result.Escalation != nil && result.Escalation.Required {
            newSteps := e.buildEscalationSteps(result.Escalation)
            steps = append(steps, newSteps...)
        }
    }
    return nil
}
```

#### Parallel Execution

```go
func (e *Engine) executeParallel(ctx context.Context, steps []WorkflowStep) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(steps))
    
    for i := range steps {
        wg.Add(1)
        go func(step *WorkflowStep) {
            defer wg.Done()
            
            step.Status = StatusRunning
            result, err := e.executeStep(ctx, *step)
            if err != nil {
                step.Status = StatusFailed
                errChan <- err
                return
            }
            
            step.Result = result
            step.Status = StatusCompleted
        }(&steps[i])
    }
    
    wg.Wait()
    close(errChan)
    
    // Check for errors
    if len(errChan) > 0 {
        return <-errChan
    }
    return nil
}
```

#### Workflow Step Lifecycle

```
┌──────────────┐
│   Pending    │ ← Initial state
└──────┬───────┘
       │ Engine.Execute()
       ▼
┌──────────────┐
│   Running    │ ← Active execution
└──────┬───────┘
       │
       ├──────────────┬──────────────┬──────────────┐
       ▼              ▼              ▼              ▼
┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐
│ Completed  │ │   Failed   │ │ Cancelled  │ │ Escalated  │
│ (success)  │ │  (error)   │ │ (timeout)  │ │ (dynamic)  │
└────────────┘ └────────────┘ └────────────┘ └─────┬──────┘
                                                     │
                                                     ▼
                                            ┌─────────────────┐
                                            │ New Steps Added │
                                            │ Continue Exec   │
                                            └─────────────────┘
```

### 4. State Management Integration

**File**: `internal/orchestrator/state/manager.go`

The state manager tracks sessions, conversation history, and cached results:

```go
type Manager struct {
    sessions map[string]*Session      // Active sessions by ID
    cache    *Cache                    // Result caching layer
    mu       sync.RWMutex              // Thread-safe access
}

type Session struct {
    ID        string                   // Unique session identifier
    History   []HistoryEntry           // Conversation history
    State     map[string]interface{}   // Key-value state storage
    CreatedAt time.Time
    UpdatedAt time.Time
}

type HistoryEntry struct {
    Request     string                 // User request
    Response    schema.AgentResponse   // Agent response
    Escalations []EscalationRecord     // Dynamic escalations
    Timestamp   time.Time
}
```

#### Session Lifecycle

```go
// 1. Session Creation
session := state.CreateSession(sessionID)

// 2. Add History Entry
entry := HistoryEntry{
    Request:  userRequest,
    Response: agentResponse,
    Escalations: []EscalationRecord{
        {
            Reason:     "Additional context needed",
            TargetAgent: "locator",
            AddedSteps:  2,
        },
    },
    Timestamp: time.Now(),
}
session.AddHistory(entry)

// 3. State Storage
session.SetState("last_analyzed_file", "/path/to/file.go")
session.SetState("analysis_depth", 3)

// 4. Caching
cache.Set(sessionID, "analysis_result", agentOutput, 5*time.Minute)

// 5. Session Cleanup
state.DeleteSession(sessionID) // Manual cleanup
state.CleanupInactiveSessions(24 * time.Hour) // Automatic cleanup
```

### 5. Agent Integration

#### Agent Interface

All agents implement a standard interface:

```go
type Agent interface {
    // Execute processes a request and returns a validated response
    Execute(ctx context.Context, req AgentRequest) (schema.AgentResponse, error)
    
    // Name returns the agent's identifier
    Name() string
    
    // Capabilities returns supported operations
    Capabilities() []string
}
```

#### Agent Factory Pattern

```go
type AgentFactory struct {
    constructors map[string]AgentConstructor
}

type AgentConstructor func(config *Config) (Agent, error)

// Registration
factory.Register("analyzer", func(cfg *Config) (Agent, error) {
    return analyzer.New(cfg), nil
})

// Usage
agent, err := factory.Create("analyzer", config)
```

#### Example: Analyzer Agent Integration

**File**: `internal/agent/analyzer/analyzer.go`

```go
func (a *Analyzer) Execute(ctx context.Context, req AgentRequest) (schema.AgentResponse, error) {
    // 1. Validate input
    if req.FilePath == "" {
        return schema.AgentResponse{}, fmt.Errorf("file path required")
    }
    
    // 2. Read and parse file
    content, err := os.ReadFile(req.FilePath)
    if err != nil {
        return a.errorResponse(err), nil
    }
    
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, req.FilePath, content, parser.ParseComments)
    if err != nil {
        return a.errorResponse(err), nil
    }
    
    // 3. Perform analysis
    output := &schema.AgentOutputV1{
        Reasoning:    "Analysis pipeline execution",
        Hypotheses:   a.buildHypotheses(node),
        Observations: a.detectPatterns(node, fset),
        RawEvidence:  a.extractEvidence(node, fset, content),
    }
    
    // 4. Validate output
    validator := evidence.NewValidator()
    validation := validator.Validate(output)
    if validation.Coverage < 0.8 {
        output.Reasoning += fmt.Sprintf(
            "\nWarning: Low evidence coverage (%.1f%%)",
            validation.Coverage*100,
        )
    }
    
    // 5. Return structured response
    return schema.AgentResponse{
        Success: true,
        Output:  output,
    }, nil
}
```

### 6. Validation Integration

**File**: `internal/validation/evidence/validator.go`

Validation occurs at multiple integration points:

```go
// 1. Schema Validation (at agent output)
schemaValidator := schema.NewValidator()
if err := schemaValidator.Validate(output); err != nil {
    return fmt.Errorf("schema validation failed: %w", err)
}

// 2. Evidence Validation (after agent execution)
evidenceValidator := evidence.NewValidator()
result := evidenceValidator.Validate(output)

if result.Coverage < 0.7 {
    // Low coverage - consider escalation
    escalation := &schema.Escalation{
        Required: true,
        Reason:   fmt.Sprintf("Evidence coverage %.1f%% below threshold", result.Coverage*100),
    }
}

// 3. Validation Result Structure
type ValidationResult struct {
    Valid           bool              // Overall validity
    Coverage        float64           // Evidence coverage (0.0-1.0)
    UnbackedClaims  []string          // Claims without evidence
    MissingFiles    []string          // Referenced but non-existent files
    SectionResults  map[string]bool   // Per-section validation
}
```

#### Validation Flow Integration

```
Agent Output
     │
     ▼
┌─────────────────────────┐
│  Schema Validation      │ ← Structural correctness
│  • Required fields      │
│  • Type checking        │
│  • Format validation    │
└──────────┬──────────────┘
           │ PASS
           ▼
┌─────────────────────────┐
│  Evidence Validation    │ ← Content quality
│  • Build evidence index │
│  • Check section backing│
│  • Calculate coverage   │
└──────────┬──────────────┘
           │
           ├─── Coverage ≥ 80% ───→ ✓ High Quality
           │
           ├─── Coverage 60-80% ──→ ⚠ Warning (Continue)
           │
           └─── Coverage < 60% ───→ ✗ Escalate
                                     └─→ Request more evidence
                                     └─→ Add locator agent
```

### 7. Escalation Mechanism

**File**: `internal/orchestrator/escalation/handler.go`

Dynamic workflow modification based on runtime needs:

```go
type EscalationHandler struct {
    policies  map[string]EscalationPolicy  // Policies by agent type
    history   *EscalationHistory           // Track escalation patterns
}

type EscalationPolicy struct {
    Triggers   []Trigger                   // Conditions requiring escalation
    TargetAgent string                     // Agent to escalate to
    MaxDepth   int                         // Prevent infinite escalation
}

type Trigger struct {
    Type      TriggerType                  // Coverage, Error, Timeout, etc.
    Threshold interface{}                  // Trigger-specific threshold
}
```

#### Escalation Flow

```go
func (h *EscalationHandler) HandleEscalation(
    ctx context.Context,
    response schema.AgentResponse,
    workflow *Workflow,
) error {
    if response.Escalation == nil || !response.Escalation.Required {
        return nil // No escalation needed
    }
    
    // 1. Check escalation depth
    if h.history.GetDepth(workflow.ID) >= MaxEscalationDepth {
        return fmt.Errorf("max escalation depth reached")
    }
    
    // 2. Determine target agent
    targetAgent := response.Escalation.TargetAgent
    if targetAgent == "" {
        targetAgent = h.selectAgentByReason(response.Escalation.Reason)
    }
    
    // 3. Create new workflow steps
    newSteps := []WorkflowStep{
        {
            AgentName: targetAgent,
            Input:     h.prepareEscalationContext(response),
            Status:    StatusPending,
        },
    }
    
    // 4. Inject steps into workflow
    workflow.AddSteps(newSteps)
    
    // 5. Record escalation
    h.history.Record(EscalationRecord{
        WorkflowID:  workflow.ID,
        FromAgent:   response.AgentName,
        ToAgent:     targetAgent,
        Reason:      response.Escalation.Reason,
        Timestamp:   time.Now(),
    })
    
    return nil
}
```

#### Common Escalation Scenarios

| Trigger | From Agent | To Agent | Reason |
|---------|-----------|----------|---------|
| Low evidence coverage (<60%) | Analyzer | Locator | Need more file context |
| File not found | Any | Locator | Resolve file path |
| Complex dependency | Analyzer | Analyzer | Deeper analysis required |
| Validation failure | Any | Review | Manual inspection needed |
| Performance threshold | Any | Optimizer | Optimization needed |

### 8. Error Handling Integration

#### Error Propagation Flow

```
Agent Error
     │
     ▼
┌─────────────────────────────────────┐
│  Agent Error Wrapping               │
│  schema.AgentError{                 │
│    Code: "ANALYSIS_FAILED",         │
│    Message: "Parse error",          │
│    Details: map[string]interface{}  │
│  }                                  │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Workflow Engine Error Handler      │
│  • Log error with context           │
│  • Update step status to Failed     │
│  • Decide: Retry, Escalate, or Stop │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Orchestrator Error Handler         │
│  • Aggregate multi-step errors      │
│  • Create user-friendly message     │
│  • Include debug information        │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Protocol Layer Error Formatting    │
│  JSON-RPC error response:           │
│  {                                  │
│    "error": {                       │
│      "code": -32603,                │
│      "message": "Internal error",   │
│      "data": { ... }                │
│    }                                │
│  }                                  │
└─────────────────────────────────────┘
```

#### Error Recovery Strategies

```go
// 1. Automatic Retry (transient failures)
func (e *Engine) executeWithRetry(ctx context.Context, step WorkflowStep) error {
    for attempt := 1; attempt <= MaxRetries; attempt++ {
        err := e.executeStep(ctx, step)
        if err == nil {
            return nil
        }
        
        if !isTransient(err) {
            return err // Don't retry permanent failures
        }
        
        time.Sleep(backoff(attempt))
    }
    return fmt.Errorf("max retries exceeded")
}

// 2. Graceful Degradation (partial results)
func (o *Orchestrator) handlePartialFailure(results []AgentResponse) AgentResponse {
    successfulResults := filter(results, func(r AgentResponse) bool {
        return r.Success
    })
    
    if len(successfulResults) > 0 {
        return mergeResults(successfulResults, WithWarning("Some agents failed"))
    }
    
    return errorResponse("All agents failed")
}

// 3. Circuit Breaker (prevent cascade failures)
func (o *Orchestrator) executeWithCircuitBreaker(agent Agent) (schema.AgentResponse, error) {
    if o.circuitBreaker.IsOpen(agent.Name()) {
        return schema.AgentResponse{}, fmt.Errorf("circuit breaker open for %s", agent.Name())
    }
    
    resp, err := agent.Execute(ctx, req)
    
    if err != nil {
        o.circuitBreaker.RecordFailure(agent.Name())
    } else {
        o.circuitBreaker.RecordSuccess(agent.Name())
    }
    
    return resp, err
}
```

### 9. Performance Monitoring Integration

**File**: `internal/profiling/profiler.go`

The profiler integrates at key execution points:

```go
type Profiler struct {
    collectors map[string]*Collector    // Per-component collectors
    reporter   *Reporter                // Centralized reporting
}

// Integration Points:
// 1. Request Entry
profile := profiler.Start("request", sessionID)
defer profile.Stop()

// 2. Agent Execution
agentProfile := profiler.Start("agent", agentName)
result, err := agent.Execute(ctx, req)
agentProfile.Stop()

// 3. Tool Execution
toolProfile := profiler.Start("tool", "file_read")
content, err := os.ReadFile(path)
toolProfile.Stop()

// 4. Validation
validationProfile := profiler.Start("validation", "evidence")
result := validator.Validate(output)
validationProfile.Stop()
```

#### Performance Metrics Structure

```go
type Metrics struct {
    Component    string                  // "orchestrator", "analyzer", etc.
    Operation    string                  // "execute", "validate", etc.
    Duration     time.Duration           // Total execution time
    MemoryUsed   uint64                  // Peak memory usage
    Goroutines   int                     // Active goroutines
    Bottlenecks  []Bottleneck            // Identified performance issues
}

type Bottleneck struct {
    Location    string                   // Code location
    Type        BottleneckType           // IO, CPU, Memory, Lock
    Impact      float64                  // Percentage of total time
    Suggestions []string                 // Optimization suggestions
}
```

#### Profiling Report Integration

```
Request Start
     │
     ├─→ [PROFILE] orchestrator.handle_request START
     │
     ├─→ [PROFILE] router.select_agent START
     │   └─→ [PROFILE] router.select_agent STOP (1.2ms)
     │
     ├─→ [PROFILE] analyzer.execute START
     │   │
     │   ├─→ [PROFILE] tool.file_read START
     │   │   └─→ [PROFILE] tool.file_read STOP (15.3ms) ⚠ BOTTLENECK
     │   │
     │   ├─→ [PROFILE] tool.ast_parse START
     │   │   └─→ [PROFILE] tool.ast_parse STOP (42.7ms)
     │   │
     │   └─→ [PROFILE] analyzer.execute STOP (89.4ms)
     │
     ├─→ [PROFILE] validator.validate START
     │   └─→ [PROFILE] validator.validate STOP (3.1ms)
     │
     └─→ [PROFILE] orchestrator.handle_request STOP (112.8ms)

Bottleneck Report:
  • tool.file_read: 13.6% of total time (I/O bound)
    Suggestion: Implement file caching
  • ast_parse: 37.9% of total time (CPU bound)
    Suggestion: Consider parallel parsing for multiple files
```

## Integration Examples

### Example 1: Simple Analysis Request

```go
// User Request
request := JSONRPCRequest{
    Method: "analyze",
    Params: map[string]interface{}{
        "file": "internal/agent/analyzer/analyzer.go",
        "session_id": "session-123",
    },
}

// Orchestrator Processing
session := state.GetOrCreateSession("session-123")
profile := profiler.Start("request", "session-123")

// Route to analyzer agent
agent := router.SelectAgent(request.Method)

// Execute workflow
workflow := &Workflow{
    Steps: []WorkflowStep{
        {
            AgentName: "analyzer",
            Input: AgentRequest{
                FilePath: "internal/agent/analyzer/analyzer.go",
            },
        },
    },
    Mode: Sequential,
}

result, err := engine.Execute(ctx, workflow)

// Validate result
validation := validator.Validate(result.Output)
if validation.Coverage < 0.8 {
    // Escalate for more evidence
    workflow.AddSteps([]WorkflowStep{
        {
            AgentName: "locator",
            Input: AgentRequest{
                Pattern: "analyzer",
            },
        },
    })
    result, _ = engine.Execute(ctx, workflow)
}

// Update session history
session.AddHistory(HistoryEntry{
    Request:  request.Method,
    Response: result,
    Timestamp: time.Now(),
})

profile.Stop()

// Return response
response := JSONRPCResponse{
    Result: result,
    Metadata: map[string]interface{}{
        "duration_ms": profile.Duration().Milliseconds(),
        "validation": validation,
    },
}
```

### Example 2: Multi-Agent Workflow with Escalation

```go
// Complex request requiring multiple agents
request := JSONRPCRequest{
    Method: "comprehensive_analysis",
    Params: map[string]interface{}{
        "pattern": "orchestrator",
        "depth": "full",
    },
}

// Initial workflow: Locate then Analyze
workflow := &Workflow{
    Steps: []WorkflowStep{
        {
            AgentName: "locator",
            Input: AgentRequest{
                Pattern: "orchestrator",
            },
        },
        {
            AgentName: "analyzer",
            Input: AgentRequest{
                // FilePath populated from locator result
            },
        },
    },
    Mode: Sequential,
}

// Execute with escalation handling
result, err := engine.Execute(ctx, workflow)

// Locator completed, analyzer needs the file path
workflow.Steps[1].Input.FilePath = result.Steps[0].Result.Output.Files[0]

// Analyzer executes and returns low coverage
analyzerResult := result.Steps[1].Result
if analyzerResult.Escalation.Required {
    // Dynamic escalation: Add more context
    escalationHandler.HandleEscalation(ctx, analyzerResult, workflow)
    
    // New steps added:
    // - Locator: Find related files
    // - Analyzer: Analyze related files
    // - Aggregator: Merge results
    
    result, _ = engine.Execute(ctx, workflow)
}

// Final result includes all agent outputs
response := aggregateResults(result)
```

### Example 3: Parallel Agent Execution

```go
// Request to analyze multiple files simultaneously
request := JSONRPCRequest{
    Method: "batch_analyze",
    Params: map[string]interface{}{
        "files": []string{
            "internal/orchestrator/orchestrator.go",
            "internal/agent/analyzer/analyzer.go",
            "internal/validation/evidence/validator.go",
        },
    },
}

// Create parallel workflow
workflow := &Workflow{
    Mode: Parallel,
    Steps: []WorkflowStep{},
}

// Add step for each file
for _, file := range request.Params["files"].([]string) {
    workflow.Steps = append(workflow.Steps, WorkflowStep{
        AgentName: "analyzer",
        Input: AgentRequest{
            FilePath: file,
        },
    })
}

// Execute all in parallel
var wg sync.WaitGroup
results := make([]schema.AgentResponse, len(workflow.Steps))

for i, step := range workflow.Steps {
    wg.Add(1)
    go func(idx int, s WorkflowStep) {
        defer wg.Done()
        agent := factory.Create(s.AgentName)
        results[idx], _ = agent.Execute(ctx, s.Input)
    }(i, step)
}

wg.Wait()

// Aggregate parallel results
aggregated := schema.AgentOutputV1{
    Reasoning: "Parallel analysis of multiple files",
    Observations: []schema.Observation{},
}

for _, result := range results {
    aggregated.Observations = append(
        aggregated.Observations,
        result.Output.Observations...,
    )
}
```

## Integration Testing

### Testing Component Interactions

```go
func TestOrchestratorAgentIntegration(t *testing.T) {
    // Setup
    orchestrator := NewOrchestrator(testConfig)
    
    // Test agent registration
    analyzer := analyzer.New()
    orchestrator.RegisterAgent("analyzer", analyzer)
    
    // Test request routing
    request := AgentRequest{
        Method: "analyze",
        FilePath: "test.go",
    }
    
    selectedAgent := orchestrator.router.SelectAgent(request.Method)
    assert.Equal(t, "analyzer", selectedAgent)
    
    // Test execution
    response, err := orchestrator.HandleRequest(context.Background(), request)
    assert.NoError(t, err)
    assert.True(t, response.Success)
    
    // Test validation integration
    validation := validator.Validate(response.Output)
    assert.Greater(t, validation.Coverage, 0.7)
    
    // Test state management
    session := orchestrator.state.GetSession(request.SessionID)
    assert.Len(t, session.History, 1)
}
```

## Troubleshooting Integration Issues

### Common Integration Problems

#### 1. Agent Not Found

**Symptom**: `agent "analyzer" not registered`

**Cause**: Agent not registered in factory

**Solution**:
```go
factory.Register("analyzer", analyzer.New)
```

#### 2. Validation Failure

**Symptom**: Low evidence coverage (<60%)

**Cause**: Agent not providing sufficient evidence

**Solution**:
```go
// Enable escalation
if validation.Coverage < 0.6 {
    escalation := &schema.Escalation{
        Required: true,
        TargetAgent: "locator",
        Reason: "Need additional context files",
    }
}
```

#### 3. Workflow Timeout

**Symptom**: Context deadline exceeded

**Cause**: Long-running agent or workflow

**Solution**:
```go
// Increase timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

// Or enable parallel execution
workflow.Mode = Parallel
```

#### 4. State Inconsistency

**Symptom**: Session history mismatch

**Cause**: Concurrent session updates

**Solution**:
```go
// Use state manager's built-in locking
state.mu.Lock()
session.AddHistory(entry)
state.mu.Unlock()
```

#### 5. Profiler Overhead

**Symptom**: Performance degradation

**Cause**: Too fine-grained profiling

**Solution**:
```go
// Profile only key operations
profiler.SetGranularity(GranularityCoarse)

// Or disable in production
if !cfg.DebugMode {
    profiler.Disable()
}
```

## Performance Considerations

### Integration Overhead

| Component | Overhead | Optimization |
|-----------|----------|--------------|
| Request routing | <1ms | Pre-compile keyword patterns |
| Agent factory | <1ms | Instance pooling |
| Schema validation | 1-5ms | Lazy validation |
| Evidence validation | 5-20ms | Parallel section validation |
| State persistence | 2-10ms | Batch updates |
| Profiling | 1-3% | Sampling mode |

### Optimization Strategies

```go
// 1. Agent Instance Pooling
type AgentPool struct {
    instances map[string][]Agent
    mu        sync.Mutex
}

func (p *AgentPool) Get(name string) Agent {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    if len(p.instances[name]) > 0 {
        agent := p.instances[name][0]
        p.instances[name] = p.instances[name][1:]
        return agent
    }
    
    return factory.Create(name)
}

// 2. Result Caching
cache.Set(cacheKey, result, 5*time.Minute)

// 3. Lazy Validation
validator.SetMode(ValidationModeLazy) // Validate only on access

// 4. Parallel Validation Sections
validator.ValidateParallel(output, []string{
    "entry_points",
    "call_graph",
    "data_flow",
})
```


## Workflow Integration Architecture

### Overview

The workflow integration system provides a unified configuration layer that coordinates quality gates, evidence validation, profiling, and multi-agent execution. This architecture enables consistent quality assurance across all orchestration operations.

### Component Relationships

```
┌─────────────────────────────────────────────────────────────────┐
│                    OrchestratorConfig                            │
│  ┌───────────────┐  ┌────────────────┐  ┌──────────────────┐   │
│  │ ProcessManager│  │  ToolExecutor  │  │ EvidenceValidator│   │
│  └───────┬───────┘  └────────┬───────┘  └────────┬─────────┘   │
│          │                   │                    │             │
│  ┌───────┴──────────────────┴────────────────────┴─────────┐   │
│  │              Orchestrator Core Instance                  │   │
│  └───────┬──────────────────────────────────────────────────┘   │
│          │                                                       │
│  ┌───────┴───────┐    ┌──────────────┐    ┌────────────────┐   │
│  │ QualityGates  │    │EnableProfiling│    │ WorkflowEngine │   │
│  └───────┬───────┘    └──────┬───────┘    └────────┬───────┘   │
└──────────┼────────────────────┼─────────────────────┼───────────┘
           │                    │                     │
           ▼                    ▼                     ▼
    ┌────────────┐      ┌────────────┐       ┌────────────┐
    │Validation  │      │ Profiling  │       │  Workflow  │
    │  Pipeline  │      │  Collector │       │  Executor  │
    └────────────┘      └────────────┘       └────────────┘
```

### OrchestratorConfig Structure

The `OrchestratorConfig` provides centralized control over all workflow integration components:

```go
type OrchestratorConfig struct {
    // Core execution components
    ProcessManager    *process.Manager      // Subprocess management
    ToolExecutor      *tool.Executor        // Tool execution interface
    
    // Quality assurance
    EvidenceValidator *evidence.Validator   // Evidence validation (strict/non-strict)
    QualityGates      QualityGates          // Quality thresholds and requirements
    
    // Performance monitoring
    EnableProfiling   bool                  // Enable/disable profiling
    Profiler          *profiling.Profiler   // Custom profiler instance (optional)
    
    // Workflow control
    MaxConcurrentSteps int                  // Parallel execution limit
    TimeoutDuration    time.Duration        // Global timeout
}

type QualityGates struct {
    MinEvidenceScore    float64        // Minimum evidence coverage (0.0-100.0)
    MaxExecutionTime    time.Duration  // Maximum workflow execution time
    RequireEvidence     bool           // Fail if no evidence provided
    RequireProfiling    bool           // Fail if profiling data missing
    AllowPartialSuccess bool           // Allow partial workflow completion
}
```

### Integration Flow

#### 1. Configuration Phase

```
User Configuration
        │
        ▼
┌─────────────────────────────────────────┐
│  Configure Individual Components       │
│  • Create ProcessManager                │
│  • Create ToolExecutor                  │
│  • Create EvidenceValidator (mode)      │
│  • Select QualityGates preset           │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│  Build OrchestratorConfig               │
│  config := OrchestratorConfig{          │
│      ProcessManager:    pm,             │
│      ToolExecutor:      te,             │
│      EvidenceValidator: ev,             │
│      QualityGates:      gates,          │
│      EnableProfiling:   true,           │
│  }                                      │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│  Initialize Orchestrator                │
│  orch := NewWithConfig(config)          │
│  • Inject all components                │
│  • Wire up integration points           │
│  • Initialize workflow engine           │
└─────────────────────────────────────────┘
```

#### 2. Execution Phase

```
Request Received
        │
        ▼
┌─────────────────────────────────────────┐
│  Pre-Execution Quality Gates            │
│  • Validate request parameters          │
│  • Check resource availability          │
│  • Initialize profiling (if enabled)    │
└─────────────┬───────────────────────────┘
              │ PASS
              ▼
┌─────────────────────────────────────────┐
│  Workflow Execution                     │
│  • Create workflow steps                │
│  • Execute agent(s) via ProcessManager  │
│  • Collect execution metrics            │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│  Evidence Validation                    │
│  • Validate output structure            │
│  • Check evidence coverage              │
│  • Calculate quality score              │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│  Post-Execution Quality Gates           │
│  • Check MinEvidenceScore threshold     │
│  • Verify MaxExecutionTime not exceeded │
│  • Validate profiling data (if required)│
└─────────────┬───────────────────────────┘
              │
              ├── PASS ────────────────┐
              │                        │
              └── FAIL ───┐            │
                          │            │
                          ▼            ▼
                    ┌──────────┐  ┌──────────┐
                    │  Reject  │  │  Accept  │
                    │  Result  │  │  Result  │
                    └──────────┘  └──────────┘
```

### Quality Gate Flow

The quality gate system enforces consistent quality standards across all workflow executions:

```
┌────────────────────────────────────────────────────────────┐
│                   Quality Gate Pipeline                     │
└────────────────────────────────────────────────────────────┘

Input: AgentResponse + QualityGates configuration

    │
    ▼
┌─────────────────────────────────────────┐
│  Gate 1: Evidence Coverage Check        │
│  • Calculate coverage percentage        │
│  • Compare to MinEvidenceScore          │
│  • Verdict: PASS (≥) / FAIL (<)         │
└─────────────┬───────────────────────────┘
              │ PASS
              ▼
┌─────────────────────────────────────────┐
│  Gate 2: Execution Time Check           │
│  • Measure total execution duration     │
│  • Compare to MaxExecutionTime          │
│  • Verdict: PASS (≤) / FAIL (>)         │
└─────────────┬───────────────────────────┘
              │ PASS
              ▼
┌─────────────────────────────────────────┐
│  Gate 3: Evidence Presence Check        │
│  • Count evidence items                 │
│  • If RequireEvidence && count == 0     │
│  • Verdict: PASS (count > 0) / FAIL     │
└─────────────┬───────────────────────────┘
              │ PASS
              ▼
┌─────────────────────────────────────────┐
│  Gate 4: Profiling Data Check           │
│  • Verify profiling data exists         │
│  • If RequireProfiling && data == nil   │
│  • Verdict: PASS (exists) / FAIL        │
└─────────────┬───────────────────────────┘
              │ PASS
              ▼
┌─────────────────────────────────────────┐
│  Gate 5: Partial Success Handling       │
│  • If any previous gate FAILED          │
│  • Check AllowPartialSuccess            │
│  • Override: PASS if allowed            │
└─────────────┬───────────────────────────┘
              │
              ▼
        Final Verdict
```

### Quality Gate Presets

Three predefined configurations for common scenarios:

```go
// DefaultQualityGates: Balanced approach
func DefaultQualityGates() QualityGates {
    return QualityGates{
        MinEvidenceScore:    100.0,            // Perfect coverage
        MaxExecutionTime:    5 * time.Minute,  // Reasonable timeout
        RequireEvidence:     true,             // Evidence mandatory
        RequireProfiling:    false,            // Profiling optional
        AllowPartialSuccess: false,            // No partial results
    }
}

// RelaxedQualityGates: For exploratory/rapid development
func RelaxedQualityGates() QualityGates {
    return QualityGates{
        MinEvidenceScore:    80.0,             // 80% coverage OK
        MaxExecutionTime:    10 * time.Minute, // Longer timeout
        RequireEvidence:     true,             // Still need evidence
        RequireProfiling:    false,            // No profiling required
        AllowPartialSuccess: true,             // Accept partial results
    }
}

// StrictQualityGates: For production/critical workflows
func StrictQualityGates() QualityGates {
    return QualityGates{
        MinEvidenceScore:    100.0,            // Perfect coverage
        MaxExecutionTime:    2 * time.Minute,  // Strict timeout
        RequireEvidence:     true,             // Evidence mandatory
        RequireProfiling:    true,             // Profiling mandatory
        AllowPartialSuccess: false,            // No partial results
    }
}
```

### Evidence Validation Integration

Evidence validation operates in two modes based on configuration:

```
┌──────────────────────────────────────────────────────────────┐
│              EvidenceValidator Configuration                  │
└──────────────────────────────────────────────────────────────┘

evidence.NewValidator(strictMode bool)

    │
    ├─── strictMode = true ─────┐
    │                           │
    │                           ▼
    │              ┌────────────────────────────┐
    │              │  Strict Validation Mode     │
    │              │  • File existence checked   │
    │              │  • Line numbers verified    │
    │              │  • Snippets validated       │
    │              │  • Missing = ERROR          │
    │              └────────────────────────────┘
    │
    └─── strictMode = false ────┐
                                │
                                ▼
                   ┌────────────────────────────┐
                   │  Non-Strict Validation Mode │
                   │  • File existence optional  │
                   │  • Line numbers trusted     │
                   │  • Snippets assumed valid   │
                   │  • Missing = WARNING        │
                   └────────────────────────────┘

Usage Example:

// Development: Non-strict (warnings only)
validator := evidence.NewValidator(false)

// Production: Strict (errors for missing evidence)
validator := evidence.NewValidator(true)
```

### Profiling Integration

Performance monitoring integration with conditional enablement:

```
┌──────────────────────────────────────────────────────────────┐
│                  Profiling Integration                        │
└──────────────────────────────────────────────────────────────┘

OrchestratorConfig.EnableProfiling = true
           │
           ▼
    ┌──────────────────────────┐
    │  Initialize Profiler     │
    │  • Start global timer    │
    │  • Prepare collectors    │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Execution Start         │
    │  profiler.Start("phase") │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Agent Execution         │
    │  • Record start time     │
    │  • Execute agent         │
    │  • Record end time       │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Execution Complete      │
    │  profiler.Stop("phase")  │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Aggregate Metrics       │
    │  • Total duration        │
    │  • Phase breakdown       │
    │  • Agent execution times │
    │  • Tool execution times  │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Attach to Response      │
    │  response.ProfilingData  │
    └──────────────────────────┘

If EnableProfiling = false:
  → No profiler initialization
  → No metrics collection
  → ProfilingData = nil
```

### Report Generation Pipeline

Comprehensive workflow reporting with integrated metrics:

```
┌──────────────────────────────────────────────────────────────┐
│                 Report Generation Flow                        │
└──────────────────────────────────────────────────────────────┘

orchestrator.GenerateReport(result)
           │
           ▼
    ┌──────────────────────────┐
    │  Extract Components      │
    │  • Execution status      │
    │  • Quality gate results  │
    │  • Evidence summary      │
    │  • Profiling data        │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Format Header           │
    │  • Workflow ID           │
    │  • Timestamp             │
    │  • Success/Failure       │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Quality Gates Section   │
    │  • Gate configurations   │
    │  • Pass/Fail status      │
    │  • Evidence score        │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Performance Metrics     │
    │  • Total duration        │
    │  • Agent execution time  │
    │  • Phase breakdown       │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Evidence Summary        │
    │  • Total items           │
    │  • By type (file/line)   │
    │  • Coverage percentage   │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Recommendations         │
    │  • Quality improvements  │
    │  • Performance tips      │
    │  • Evidence suggestions  │
    └──────────┬───────────────┘
               │
               ▼
       Formatted Report String
```

### Multi-Agent Coordination Flow

Workflow integration enables seamless multi-agent coordination:

```
Request: "Analyze orchestrator implementation"
           │
           ▼
    ┌──────────────────────────┐
    │  Workflow Planning       │
    │  • Identify required     │
    │    agents: locator,      │
    │    analyzer              │
    │  • Determine sequence    │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Phase 1: Locate         │
    │  • Agent: locator        │
    │  • Task: Find files      │
    │  • Pattern: "orchestrator│
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Quality Gate: Phase 1   │
    │  • Evidence coverage OK? │
    │  • Execution time OK?    │
    └──────────┬───────────────┘
               │ PASS
               ▼
    ┌──────────────────────────┐
    │  Phase 2: Analyze        │
    │  • Agent: analyzer       │
    │  • Task: Analyze files   │
    │  • Input: Phase 1 output │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Quality Gate: Phase 2   │
    │  • Evidence coverage OK? │
    │  • Execution time OK?    │
    └──────────┬───────────────┘
               │ PASS
               ▼
    ┌──────────────────────────┐
    │  Aggregate Results       │
    │  • Combine outputs       │
    │  • Merge evidence        │
    │  • Total profiling data  │
    └──────────┬───────────────┘
               │
               ▼
    ┌──────────────────────────┐
    │  Final Quality Gate      │
    │  • Overall evidence      │
    │    coverage: 95%         │
    │  • Total time: 1.2s      │
    │  • Verdict: PASS         │
    └──────────┬───────────────┘
               │
               ▼
         Final Response
```

### Configuration Best Practices

#### Development Environment

```go
config := orchestrator.OrchestratorConfig{
    ProcessManager:    process.NewManager(),
    ToolExecutor:      tool.NewExecutor(),
    EvidenceValidator: evidence.NewValidator(false), // Non-strict
    QualityGates:      orchestrator.RelaxedQualityGates(),
    EnableProfiling:   true, // Identify bottlenecks
}
```

#### CI/CD Environment

```go
config := orchestrator.OrchestratorConfig{
    ProcessManager:    process.NewManager(),
    ToolExecutor:      tool.NewExecutor(),
    EvidenceValidator: evidence.NewValidator(true), // Strict
    QualityGates:      orchestrator.DefaultQualityGates(),
    EnableProfiling:   false, // Reduce overhead
}
```

#### Production Environment

```go
config := orchestrator.OrchestratorConfig{
    ProcessManager:    process.NewManager(),
    ToolExecutor:      tool.NewExecutor(),
    EvidenceValidator: evidence.NewValidator(true), // Strict
    QualityGates:      orchestrator.StrictQualityGates(),
    EnableProfiling:   true, // Monitor performance
    MaxConcurrentSteps: 10,  // Limit resource usage
    TimeoutDuration:    30 * time.Second, // Fast fail
}
```

### Integration Metrics

Key metrics tracked across workflow integration:

```go
type WorkflowMetrics struct {
    // Execution metrics
    TotalDuration       time.Duration
    AgentExecutionTime  time.Duration
    ValidationTime      time.Duration
    
    // Quality metrics
    EvidenceCoverage    float64  // 0.0-100.0
    QualityGatesPassed  int
    QualityGatesFailed  int
    
    // Evidence metrics
    TotalEvidenceItems  int
    FileEvidence        int
    LineEvidence        int
    
    // Performance metrics
    PhaseCount          int
    PhaseBreakdown      []PhaseMetric
}

type PhaseMetric struct {
    Name        string
    Duration    time.Duration
    AgentName   string
    Success     bool
}
```

### Example: Complete Workflow Integration

```go
func AnalyzeWithFullIntegration(ctx context.Context, pattern string) error {
    // 1. Configure orchestrator
    config := orchestrator.OrchestratorConfig{
        ProcessManager:    process.NewManager(),
        ToolExecutor:      tool.NewExecutor(),
        EvidenceValidator: evidence.NewValidator(true),
        QualityGates: orchestrator.QualityGates{
            MinEvidenceScore:    90.0,
            MaxExecutionTime:    3 * time.Minute,
            RequireEvidence:     true,
            RequireProfiling:    true,
            AllowPartialSuccess: false,
        },
        EnableProfiling: true,
    }
    
    // 2. Create orchestrator
    orch := orchestrator.NewWithConfig(config)
    
    // 3. Execute workflow
    result, err := orch.Execute(ctx, orchestrator.Request{
        Objective: fmt.Sprintf("Analyze codebase for pattern: %s", pattern),
        Context: map[string]interface{}{
            "pattern": pattern,
        },
    })
    if err != nil {
        return fmt.Errorf("workflow execution failed: %w", err)
    }
    
    // 4. Generate comprehensive report
    report := orch.GenerateReport(result)
    fmt.Println(report)
    
    // 5. Verify quality gates
    if !result.Success {
        return fmt.Errorf("quality gates not met: %s", result.Error)
    }
    
    // 6. Extract and log metrics
    if result.ProfilingData != nil {
        fmt.Printf("Total Duration: %s\n", result.ProfilingData.TotalDuration)
        fmt.Printf("Agent Execution: %s\n", result.ProfilingData.AgentExecutionTime)
        fmt.Printf("Evidence Items: %d\n", len(result.Evidence))
        fmt.Printf("Quality Score: %.1f%%\n", result.QualityScore)
    }
    
    return nil
}
```

### Summary

The workflow integration architecture provides:

✅ **Unified Configuration**: Single config point for all components  
✅ **Quality Enforcement**: Consistent quality gates across workflows  
✅ **Flexible Validation**: Strict and non-strict evidence validation modes  
✅ **Performance Monitoring**: Integrated profiling with conditional enablement  
✅ **Comprehensive Reporting**: Detailed workflow execution reports  
✅ **Multi-Agent Coordination**: Seamless agent-to-agent integration  
✅ **Environment Adaptation**: Presets for development, CI/CD, production

This architecture ensures consistent, high-quality, and performant multi-agent workflows throughout the Conexus system.

## Future Integration Plans

### Phase 6 Enhancements

1. **Plugin System**: Dynamic agent loading
2. **Distributed Execution**: Agent execution across multiple nodes
3. **Event Streaming**: Real-time workflow progress updates
4. **Advanced Caching**: Semantic caching with similarity matching
5. **ML Integration**: Predictive agent selection and escalation

## Related Documentation

- [Validation Guide](../validation-guide.md) - Evidence validation system
- [Profiling Guide](../profiling-guide.md) - Performance monitoring
- [Testing Strategy](../contributing/testing-strategy.md) - Integration testing
- [Orchestrator README](../../internal/orchestrator/README.md) - Core orchestration

## Summary

Conexus's component integration architecture provides:

✅ **Centralized Orchestration**: Single coordination point for all agents
✅ **Flexible Workflows**: Sequential, parallel, and conditional execution
✅ **Dynamic Escalation**: Runtime workflow modification
✅ **Comprehensive Validation**: Multi-layer quality gates
✅ **Performance Monitoring**: Fine-grained profiling
✅ **Stateful Sessions**: Full conversation history tracking
✅ **Error Resilience**: Retry, degradation, and circuit breaker patterns

The integration layer ensures reliable, performant, and maintainable multi-agent collaboration.
