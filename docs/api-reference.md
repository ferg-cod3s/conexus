# Conexus API Reference

**Version**: AGENT_OUTPUT_V1  
**Last Updated**: Phase 5 - Documentation Complete  
**Status**: Production Ready

---

## Table of Contents

1. [Overview](#overview)
2. [AGENT_OUTPUT_V1 Schema](#agent_output_v1-schema)
3. [Agent Request/Response API](#agent-requestresponse-api)
4. [Evidence Validation API](#evidence-validation-api)
5. [Schema Validation API](#schema-validation-api)
6. [Profiling API](#profiling-api)
7. [Error Codes Reference](#error-codes-reference)
8. [API Conventions](#api-conventions)
9. [Usage Examples](#usage-examples)

---

## Overview

### API Organization

Conexus provides four primary API layers:

1. **Schema API** (`pkg/schema`) - Core data structures for agent communication
2. **Validation API** (`internal/validation`) - Evidence and schema validation
3. **Profiling API** (`internal/profiling`) - Performance monitoring and metrics
4. **Orchestration API** (`internal/orchestrator`) - Workflow execution and coordination

### Versioning

- **Current Schema Version**: `AGENT_OUTPUT_V1`
- **API Stability**: Stable (Production)
- **Breaking Change Policy**: Major version increment required
- **Deprecation Notice Period**: 90 days minimum

### Import Paths

```go
import (
    "github.com/ferg-cod3s/conexus/pkg/schema"                  // Core schema types
    "github.com/ferg-cod3s/conexus/internal/validation/evidence" // Evidence validation
    "github.com/ferg-cod3s/conexus/internal/validation/schema"  // Schema validation
    "github.com/ferg-cod3s/conexus/internal/profiling"          // Performance profiling
)
```

---

## AGENT_OUTPUT_V1 Schema

### Primary Structure

The `AgentOutputV1` structure is the standardized format for all Conexus agent outputs.

```go
type AgentOutputV1 struct {
    Version              string                   // Must be "AGENT_OUTPUT_V1"
    ComponentName        string                   // Component identifier
    ScopeDescription     string                   // Concise scope definition
    Overview             string                   // 2-4 sentence summary
    EntryPoints          []EntryPoint             // Entry point identification
    CallGraph            []CallGraphEdge          // Execution flow graph
    DataFlow             DataFlow                 // Data transformation pipeline
    StateManagement      []StateOperation         // Persistence operations
    SideEffects          []SideEffect             // External interactions
    ErrorHandling        []ErrorHandler           // Error handling paths
    Configuration        []ConfigInfluence        // Configuration influence
    Patterns             []Pattern                // Design patterns (descriptive)
    Concurrency          []ConcurrencyMechanism   // Concurrency mechanisms
    ExternalDependencies []ExternalDependency     // External dependencies
    Limitations          []string                 // Transparency requirements
    OpenQuestions        []string                 // Areas needing clarification
    RawEvidence          []Evidence               // Evidence traceability (MANDATORY)
}
```

#### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | `string` | ✅ Yes | Must be exactly `"AGENT_OUTPUT_V1"` |
| `component_name` | `string` | ✅ Yes | User-supplied or inferred component label |
| `scope_description` | `string` | ⚠️ Recommended | Concise scope definition |
| `overview` | `string` | ⚠️ Recommended | 2-4 sentence HOW summary |
| `raw_evidence` | `[]Evidence` | ✅ Yes | Mandatory evidence backing for all claims |

### Nested Types

#### EntryPoint

Represents a function or symbol serving as an entry into the component.

```go
type EntryPoint struct {
    File   string // Absolute path (required)
    Lines  string // Line range e.g., "24-31" (required)
    Symbol string // Function or export name (required)
    Role   string // handler|service|utility|etc (optional)
}
```

**Example:**
```json
{
  "file": "/home/user/project/internal/agent/analyzer/analyzer.go",
  "lines": "45-52",
  "symbol": "Analyze",
  "role": "service"
}
```

#### CallGraphEdge

Represents a function call relationship in the execution flow.

```go
type CallGraphEdge struct {
    From    string // Source location (file.go:funcA)
    To      string // Target location (other.go:funcB)
    ViaLine int    // Line number of call
}
```

**Example:**
```json
{
  "from": "analyzer.go:Analyze",
  "to": "parser.go:ParseFile",
  "via_line": 67
}
```

#### DataFlow

Describes how data transforms through the system.

```go
type DataFlow struct {
    Inputs          []DataPoint      // Input data points
    Transformations []Transformation // Transformation operations
    Outputs         []DataPoint      // Output data points
}
```

**DataPoint:**
```go
type DataPoint struct {
    Source      string // file.go:line (required)
    Name        string // Variable name (required)
    Type        string // Inferred/simple type (optional)
    Description string // Purpose (optional)
}
```

**Transformation:**
```go
type Transformation struct {
    File        string // File path (required)
    Lines       string // Line range (required)
    Operation   string // parse|validate|map|filter|aggregate|serialize
    Description string // What changes (optional)
    BeforeShape string // Optional data shape before
    AfterShape  string // Optional data shape after
}
```

**Example:**
```json
{
  "inputs": [
    {
      "source": "analyzer.go:45",
      "name": "sourceFiles",
      "type": "[]string",
      "description": "List of Go source files to analyze"
    }
  ],
  "transformations": [
    {
      "file": "/home/user/project/internal/agent/analyzer/parser.go",
      "lines": "23-45",
      "operation": "parse",
      "description": "Parse Go AST from source files"
    }
  ],
  "outputs": [
    {
      "source": "analyzer.go:120",
      "name": "analysisResult",
      "type": "schema.AgentOutputV1",
      "description": "Complete analysis with evidence"
    }
  ]
}
```

#### StateOperation

Represents a persistence or state management operation.

```go
type StateOperation struct {
    File        string // File path (required)
    Lines       string // Line range (required)
    Kind        string // db|cache|memory|fs
    Operation   string // read|write|update|delete
    Entity      string // table|collection|key
    Description string // Operation description (optional)
}
```

**Example:**
```json
{
  "file": "/home/user/project/internal/orchestrator/state/manager.go",
  "lines": "56-78",
  "kind": "cache",
  "operation": "write",
  "entity": "agent_results",
  "description": "Cache validated agent output for reuse"
}
```

#### SideEffect

Represents an external interaction or observable side effect.

```go
type SideEffect struct {
    File        string // File path (required)
    Line        int    // Line number (required)
    Type        string // log|metric|emit|publish|http|fs
    Description string // Effect description (optional)
}
```

**Example:**
```json
{
  "file": "/home/user/project/internal/profiling/profiler.go",
  "line": 143,
  "type": "log",
  "description": "Log execution metrics to stdout"
}
```

#### ErrorHandler

Represents error handling logic in the code.

```go
type ErrorHandler struct {
    File      string // File path (required)
    Lines     string // Line range (required)
    Type      string // throw|catch|guard|retry
    Condition string // Expression or pattern (optional)
    Effect    string // propagate|fallback|retry
}
```

**Example:**
```json
{
  "file": "/home/user/project/internal/orchestrator/workflow/engine.go",
  "lines": "134-142",
  "type": "catch",
  "condition": "err != nil",
  "effect": "propagate"
}
```

#### ConfigInfluence

Represents how configuration affects system behavior.

```go
type ConfigInfluence struct {
    File      string // File path (required)
    Line      int    // Line number (required)
    Kind      string // env|flag|configObject
    Name      string // CONFIG_NAME
    Influence string // Description of impact
}
```

**Example:**
```json
{
  "file": "/home/user/project/internal/profiling/profiler.go",
  "line": 35,
  "kind": "flag",
  "name": "PROFILING_ENABLED",
  "influence": "Enables/disables performance profiling globally"
}
```

#### Pattern

Represents a design pattern usage in the code.

```go
type Pattern struct {
    Name        string // Pattern name (Factory, Observer, etc.)
    File        string // File path (required)
    Lines       string // Line range (required)
    Description string // Existing usage description
}
```

**Example:**
```json
{
  "name": "Factory",
  "file": "/home/user/project/internal/agent/factory.go",
  "lines": "23-67",
  "description": "Factory pattern for creating specialized agents based on type"
}
```

#### ConcurrencyMechanism

Represents concurrent execution patterns in the code.

```go
type ConcurrencyMechanism struct {
    File        string // File path (required)
    Lines       string // Line range (required)
    Mechanism   string // goroutine|channel|mutex|waitgroup
    Description string // Concurrency description (optional)
}
```

**Example:**
```json
{
  "file": "/home/user/project/internal/orchestrator/workflow/executor.go",
  "lines": "89-112",
  "mechanism": "goroutine",
  "description": "Parallel execution of independent workflow nodes"
}
```

#### ExternalDependency

Represents external module or package usage.

```go
type ExternalDependency struct {
    File    string // File path (required)
    Line    int    // Line number (required)
    Module  string // Package or internal boundary
    Purpose string // Why this dependency is needed
}
```

**Example:**
```json
{
  "file": "/home/user/project/internal/protocol/jsonrpc.go",
  "line": 12,
  "module": "encoding/json",
  "purpose": "JSON-RPC protocol encoding/decoding"
}
```

#### Evidence (MANDATORY)

Provides file:line backing for all claims in the output.

```go
type Evidence struct {
    Claim string // The claim being evidenced (required)
    File  string // Absolute path (required)
    Lines string // Line range (required)
}
```

**Evidence Format Requirements:**
- **Claim**: Clear, specific statement being backed
- **File**: Must be absolute path (e.g., `/home/user/project/file.go`)
- **Lines**: Single line (`"42"`) or range (`"42-56"`)

**Example:**
```json
{
  "claim": "Analyzer.Analyze() parses Go source files using go/parser",
  "file": "/home/user/project/internal/agent/analyzer/analyzer.go",
  "lines": "67-89"
}
```

**Evidence Coverage Requirement:**
> **100% of structural claims (EntryPoints, CallGraph, DataFlow, StateManagement, SideEffects, ErrorHandling, Configuration, Patterns, Concurrency) must be backed by evidence entries.**

---

## Agent Request/Response API

### AgentRequest

Defines the input format for agent invocation.

```go
type AgentRequest struct {
    RequestID   string              // Unique request identifier
    AgentID     string              // Target agent identifier
    Task        AgentTask           // Task specification
    Context     ConversationContext // Conversation history
    Permissions Permissions         // Execution permissions
    Timestamp   time.Time           // Request timestamp
}
```

#### AgentTask

```go
type AgentTask struct {
    TargetAgent        string   // Agent to invoke
    Files              []string // Files to analyze
    EntrySymbols       []string // Entry point symbols (optional)
    AllowedDirectories []string // Permitted directories
    SpecificRequest    string   // Detailed task description
}
```

#### ConversationContext

```go
type ConversationContext struct {
    UserRequest        string                 // Original user request
    PreviousAgents     []string               // Previously invoked agents
    AccumulatedContext map[string]interface{} // Accumulated context data
}
```

#### Permissions

```go
type Permissions struct {
    AllowedDirectories []string // Permitted file system paths
    ReadOnly           bool     // Read-only restriction
    MaxFileSize        int64    // Maximum file size (bytes)
    MaxExecutionTime   int      // Maximum execution time (seconds)
}
```

### AgentResponse

Defines the output format from agent execution.

```go
type AgentResponse struct {
    RequestID  string          // Matching request identifier
    AgentID    string          // Responding agent identifier
    Status     ResponseStatus  // Execution status
    Output     *AgentOutputV1  // Agent output (if complete)
    Escalation *Escalation     // Escalation info (if needed)
    Error      *AgentError     // Error details (if error)
    Duration   time.Duration   // Execution duration
    Timestamp  time.Time       // Response timestamp
}
```

#### ResponseStatus

```go
type ResponseStatus string

const (
    StatusComplete           ResponseStatus = "complete"
    StatusPartial            ResponseStatus = "partial"
    StatusEscalationRequired ResponseStatus = "escalation_required"
    StatusError              ResponseStatus = "error"
)
```

**Status Meanings:**

| Status | Meaning | Output Expected | Escalation Expected |
|--------|---------|----------------|---------------------|
| `complete` | Task completed successfully | ✅ Yes | ❌ No |
| `partial` | Task partially completed | ✅ Partial | ⚠️ Optional |
| `escalation_required` | Need assistance from another agent | ⚠️ Optional | ✅ Yes |
| `error` | Task failed with error | ❌ No | ❌ No |

#### Escalation

```go
type Escalation struct {
    Required     bool   // Whether escalation is required
    TargetAgent  string // Agent to escalate to
    Reason       string // Why escalation is needed
    RequiredInfo string // What information is needed
}
```

**Example:**
```json
{
  "required": true,
  "target_agent": "codebase-locator",
  "reason": "Need to locate related test files for complete analysis",
  "required_info": "Test files for internal/agent/analyzer package"
}
```

#### AgentError

```go
type AgentError struct {
    Code        string // Error code (see Error Codes Reference)
    Message     string // Human-readable error message
    Recoverable bool   // Whether error is recoverable
    Details     string // Additional error details (optional)
}
```

**Example:**
```json
{
  "code": "FILE_NOT_FOUND",
  "message": "Source file does not exist: /path/to/missing.go",
  "recoverable": true,
  "details": "Agent can retry with correct file path"
}
```

---

## Evidence Validation API

### Validator

The evidence validator enforces 100% evidence backing for agent outputs.

```go
package evidence

type Validator struct {
    // Internal fields
}

// NewValidator creates a new evidence validator
func NewValidator(strictMode bool) *Validator
```

**Parameters:**
- `strictMode` (bool): If true, validates that evidence files exist on disk

### Validate Method

```go
func (v *Validator) Validate(output *schema.AgentOutputV1) (*ValidationResult, error)
```

**Parameters:**
- `output`: The agent output to validate

**Returns:**
- `*ValidationResult`: Validation results with coverage metrics
- `error`: Error if validation cannot be performed (e.g., nil output)

### ValidationResult

```go
type ValidationResult struct {
    Valid              bool              // Overall validity
    TotalClaims        int               // Total structural claims
    BackedClaims       int               // Claims with evidence
    UnbackedClaims     []UnbackedClaim   // Claims without evidence
    InvalidEvidence    []InvalidEvidence // Invalid evidence entries
    CoveragePercentage float64           // Evidence coverage (0-100)
}
```

#### UnbackedClaim

```go
type UnbackedClaim struct {
    Section     string // Section name (e.g., "entry_points")
    Description string // Claim description
    Index       int    // Index in array
}
```

#### InvalidEvidence

```go
type InvalidEvidence struct {
    File   string // Evidence file path
    Lines  string // Evidence line range
    Reason string // Why evidence is invalid
}
```

### Validation Coverage

**Validated Sections:**
1. ✅ EntryPoints
2. ✅ CallGraph
3. ✅ DataFlow (Inputs, Transformations, Outputs)
4. ✅ StateManagement
5. ✅ SideEffects
6. ✅ ErrorHandling
7. ✅ Configuration
8. ✅ Patterns
9. ✅ Concurrency

**Validation Rules:**
- Each structural claim must have matching evidence entry
- Evidence format: `file:lines` (e.g., `"analyzer.go:45-67"`)
- In strict mode: Files must exist on disk
- In strict mode: Line ranges must be valid (positive, start ≤ end)

### Usage Example

```go
import (
    "github.com/ferg-cod3s/conexus/internal/validation/evidence"
    "github.com/ferg-cod3s/conexus/pkg/schema"
)

// Create validator (strict mode = files must exist)
validator := evidence.NewValidator(true)

// Validate agent output
result, err := validator.Validate(agentOutput)
if err != nil {
    // Handle validation error
    return fmt.Errorf("validation failed: %w", err)
}

// Check results
if !result.Valid {
    fmt.Printf("Validation failed: %.1f%% coverage\n", result.CoveragePercentage)
    fmt.Printf("Unbacked claims: %d\n", len(result.UnbackedClaims))
    
    // Print unbacked claims
    for _, claim := range result.UnbackedClaims {
        fmt.Printf("  %s[%d]: %s\n", claim.Section, claim.Index, claim.Description)
    }
}

// Require 100% coverage
if result.CoveragePercentage < 100.0 {
    return fmt.Errorf("insufficient evidence coverage: %.1f%%", result.CoveragePercentage)
}
```

---

## Schema Validation API

### Validator

The schema validator ensures agent outputs conform to AGENT_OUTPUT_V1 specification.

```go
package schema

type Validator struct {
    // Internal fields
}

// NewValidator creates a new schema validator
func NewValidator(allowPartial bool) *Validator
```

**Parameters:**
- `allowPartial` (bool): If true, allows missing optional fields without error

### Validate Method

```go
func (v *Validator) Validate(output *schema.AgentOutputV1) (*ValidationResult, error)
```

**Parameters:**
- `output`: The agent output to validate

**Returns:**
- `*ValidationResult`: Schema validation results
- `error`: Error if validation cannot be performed

### ValidationResult

```go
type ValidationResult struct {
    Valid         bool               // Overall schema validity
    Errors        []ValidationError  // Critical schema errors
    Warnings      []ValidationWarning // Non-critical warnings
    MissingFields []string           // Missing optional fields
    InvalidFields []InvalidField     // Fields with wrong types
}
```

#### ValidationError

```go
type ValidationError struct {
    Field   string      // Field path (e.g., "entry_points[0].file")
    Message string      // Error message
    Value   interface{} // Invalid value (optional)
}
```

#### ValidationWarning

```go
type ValidationWarning struct {
    Field   string // Field path
    Message string // Warning message
}
```

#### InvalidField

```go
type InvalidField struct {
    Field        string      // Field path
    ExpectedType string      // Expected type
    ActualType   string      // Actual type
    Value        interface{} // Field value
}
```

### Validation Rules

**Critical Errors (result.Valid = false):**
- Missing `version` field
- Wrong `version` value (must be "AGENT_OUTPUT_V1")
- Missing `component_name` field
- Invalid evidence file paths (must be absolute)
- Invalid line ranges (must be positive, start ≤ end)
- Missing required nested fields (e.g., EntryPoint.File, CallGraphEdge.From)

**Warnings (result.Valid may be true):**
- Empty `scope_description`
- Empty or very short `overview` (< 20 characters)
- No evidence provided (empty `raw_evidence` array)

### Usage Example

```go
import (
    "github.com/ferg-cod3s/conexus/internal/validation/schema"
    conexusSchema "github.com/ferg-cod3s/conexus/pkg/schema"
)

// Create validator (partial outputs not allowed)
validator := schema.NewValidator(false)

// Validate schema
result, err := validator.Validate(agentOutput)
if err != nil {
    return fmt.Errorf("schema validation failed: %w", err)
}

// Check for errors
if !result.Valid {
    fmt.Printf("Schema validation failed with %d errors:\n", len(result.Errors))
    for _, err := range result.Errors {
        fmt.Printf("  %s: %s\n", err.Field, err.Message)
    }
    return fmt.Errorf("invalid schema")
}

// Check warnings
if len(result.Warnings) > 0 {
    fmt.Printf("Schema warnings (%d):\n", len(result.Warnings))
    for _, warn := range result.Warnings {
        fmt.Printf("  %s: %s\n", warn.Field, warn.Message)
    }
}
```

---

## Profiling API

### Profiler

The profiler tracks performance metrics for agent executions.

```go
package profiling

type Profiler struct {
    // Internal fields
}

// NewProfiler creates a new performance profiler
func NewProfiler(enabled bool) *Profiler
```

**Parameters:**
- `enabled` (bool): If true, profiling is active; if false, profiling is no-op

### Core Methods

#### StartExecution

```go
func (p *Profiler) StartExecution(ctx context.Context, agent string, request string) *ExecutionContext
```

Begins profiling an agent execution.

**Parameters:**
- `ctx`: Context for execution
- `agent`: Agent identifier
- `request`: Request description

**Returns:**
- `*ExecutionContext`: Execution tracking context

**Example:**
```go
profiler := profiling.NewProfiler(true)
execCtx := profiler.StartExecution(ctx, "analyzer", "Analyze internal/agent/analyzer")
defer execCtx.End(output, err)
```

#### GetAgentMetrics

```go
func (p *Profiler) GetAgentMetrics(agent string) (*AggregateMetrics, bool)
```

Retrieves aggregate metrics for a specific agent.

**Returns:**
- `*AggregateMetrics`: Metrics for the agent
- `bool`: True if metrics exist for agent

#### GetAllMetrics

```go
func (p *Profiler) GetAllMetrics() map[string]*AggregateMetrics
```

Retrieves all aggregate metrics across all agents.

#### GetBottlenecks

```go
func (p *Profiler) GetBottlenecks(threshold time.Duration) []Bottleneck
```

Identifies performance bottlenecks exceeding the threshold.

**Parameters:**
- `threshold`: Duration threshold for bottleneck detection

**Returns:**
- `[]Bottleneck`: List of detected bottlenecks

#### GetReport

```go
func (p *Profiler) GetReport() *PerformanceReport
```

Generates a comprehensive performance report.

**Returns:**
- `*PerformanceReport`: Complete performance report

#### Control Methods

```go
func (p *Profiler) Enable()              // Enable profiling
func (p *Profiler) Disable()             // Disable profiling
func (p *Profiler) IsEnabled() bool      // Check if enabled
func (p *Profiler) Clear()               // Clear all metrics
```

### ExecutionContext

Tracks a single agent execution being profiled.

```go
type ExecutionContext struct {
    // Internal fields
}
```

#### Methods

```go
// StartPhase begins profiling a specific execution phase
func (ec *ExecutionContext) StartPhase(name string)

// EndPhase ends the current execution phase
func (ec *ExecutionContext) EndPhase()

// End completes the execution profiling
func (ec *ExecutionContext) End(output *schema.AgentOutputV1, err error)
```

**Phase Profiling Example:**
```go
execCtx := profiler.StartExecution(ctx, "analyzer", "Analyze component")
defer execCtx.End(output, err)

execCtx.StartPhase("parse")
// ... parsing work ...
execCtx.EndPhase()

execCtx.StartPhase("analyze")
// ... analysis work ...
execCtx.EndPhase()

execCtx.StartPhase("evidence")
// ... evidence collection ...
execCtx.EndPhase()
```

### Data Structures

#### ExecutionProfile

```go
type ExecutionProfile struct {
    ID              string        // Unique execution ID
    Agent           string        // Agent identifier
    Request         string        // Request description
    StartTime       time.Time     // Execution start time
    EndTime         time.Time     // Execution end time
    Duration        time.Duration // Total duration
    MemoryAllocated uint64        // Memory allocated (bytes)
    MemoryFreed     uint64        // Memory freed (bytes)
    GoroutineCount  int           // Goroutine count at start
    Success         bool          // Whether execution succeeded
    Error           error         // Error if failed
    Phases          []PhaseProfile // Phase-level metrics
}
```

#### PhaseProfile

```go
type PhaseProfile struct {
    Name        string        // Phase name
    StartTime   time.Time     // Phase start time
    Duration    time.Duration // Phase duration
    MemoryDelta int64         // Memory change during phase
}
```

#### AggregateMetrics

```go
type AggregateMetrics struct {
    Agent           string        // Agent identifier
    TotalExecutions int           // Total execution count
    SuccessCount    int           // Successful executions
    FailureCount    int           // Failed executions
    TotalDuration   time.Duration // Cumulative duration
    MinDuration     time.Duration // Minimum duration
    MaxDuration     time.Duration // Maximum duration
    AvgDuration     time.Duration // Average duration
    TotalMemory     uint64        // Cumulative memory (bytes)
    AvgMemory       uint64        // Average memory (bytes)
    Percentiles     *Percentiles  // Duration percentiles
}
```

#### Percentiles

```go
type Percentiles struct {
    P50 time.Duration // Median duration
    P90 time.Duration // 90th percentile
    P95 time.Duration // 95th percentile
    P99 time.Duration // 99th percentile
}
```

#### Bottleneck

```go
type Bottleneck struct {
    Agent       string        // Agent with bottleneck
    Type        string        // slow_execution | high_variance
    AvgDuration time.Duration // Average duration
    Threshold   time.Duration // Threshold exceeded
    Severity    string        // low | medium | high | critical
}
```

**Severity Calculation:**
- `critical`: duration > 3.0 × threshold
- `high`: duration > 2.0 × threshold
- `medium`: duration > 1.5 × threshold
- `low`: duration > 1.0 × threshold

#### PerformanceReport

```go
type PerformanceReport struct {
    GeneratedAt          time.Time                    // Report timestamp
    TotalExecutions      int                          // Total executions tracked
    OverallAvgDuration   time.Duration                // Overall avg duration
    OverallAvgMemory     uint64                       // Overall avg memory
    OverallSuccessRate   float64                      // Overall success rate (%)
    AgentMetrics         map[string]*AggregateMetrics // Per-agent metrics
    Bottlenecks          []Bottleneck                 // Detected bottlenecks
}
```

### Complete Usage Example

```go
import (
    "context"
    "fmt"
    "time"
    
    "github.com/ferg-cod3s/conexus/internal/profiling"
    "github.com/ferg-cod3s/conexus/pkg/schema"
)

func executeWithProfiling() error {
    // Create profiler
    profiler := profiling.NewProfiler(true)
    
    // Start execution tracking
    execCtx := profiler.StartExecution(
        context.Background(),
        "analyzer",
        "Analyze internal/agent/analyzer",
    )
    
    var output *schema.AgentOutputV1
    var err error
    
    // Track phases
    execCtx.StartPhase("parse")
    // ... parsing work ...
    execCtx.EndPhase()
    
    execCtx.StartPhase("analyze")
    // ... analysis work ...
    execCtx.EndPhase()
    
    execCtx.StartPhase("validate")
    // ... validation work ...
    execCtx.EndPhase()
    
    // Complete execution
    execCtx.End(output, err)
    
    // Get metrics
    metrics, _ := profiler.GetAgentMetrics("analyzer")
    fmt.Printf("Analyzer Stats:\n")
    fmt.Printf("  Executions: %d\n", metrics.TotalExecutions)
    fmt.Printf("  Success Rate: %.1f%%\n", 
        float64(metrics.SuccessCount)/float64(metrics.TotalExecutions)*100)
    fmt.Printf("  Avg Duration: %v\n", metrics.AvgDuration)
    fmt.Printf("  P95 Duration: %v\n", metrics.Percentiles.P95)
    
    // Check for bottlenecks
    bottlenecks := profiler.GetBottlenecks(100 * time.Millisecond)
    if len(bottlenecks) > 0 {
        fmt.Printf("Performance Bottlenecks:\n")
        for _, b := range bottlenecks {
            fmt.Printf("  %s (%s): %v (severity: %s)\n",
                b.Agent, b.Type, b.AvgDuration, b.Severity)
        }
    }
    
    // Generate full report
    report := profiler.GetReport()
    fmt.Printf("\nPerformance Report:\n")
    fmt.Printf("  Total Executions: %d\n", report.TotalExecutions)
    fmt.Printf("  Overall Success Rate: %.1f%%\n", report.OverallSuccessRate)
    fmt.Printf("  Overall Avg Duration: %v\n", report.OverallAvgDuration)
    
    return nil
}
```

---

## Error Codes Reference

### Standard Error Codes

| Code | Category | Recoverable | Description |
|------|----------|-------------|-------------|
| `FILE_NOT_FOUND` | Input | ✅ Yes | Requested file does not exist |
| `FILE_READ_ERROR` | Input | ✅ Yes | Cannot read file contents |
| `PERMISSION_DENIED` | Security | ❌ No | Operation not permitted by permissions |
| `INVALID_INPUT` | Validation | ✅ Yes | Input parameters are invalid |
| `PARSE_ERROR` | Processing | ⚠️ Maybe | Cannot parse file or data |
| `ANALYSIS_FAILED` | Processing | ⚠️ Maybe | Analysis operation failed |
| `VALIDATION_FAILED` | Validation | ✅ Yes | Output validation failed |
| `EVIDENCE_INSUFFICIENT` | Validation | ✅ Yes | Evidence coverage < 100% |
| `SCHEMA_INVALID` | Validation | ✅ Yes | Schema validation failed |
| `TIMEOUT` | Execution | ⚠️ Maybe | Execution exceeded time limit |
| `OUT_OF_MEMORY` | Resource | ❌ No | Insufficient memory |
| `ESCALATION_REQUIRED` | Coordination | ✅ Yes | Need assistance from another agent |
| `INTERNAL_ERROR` | System | ❌ No | Unexpected internal error |

### Error Handling Guidelines

**For Recoverable Errors:**
1. Set `Recoverable: true` in `AgentError`
2. Provide clear recovery instructions in `Details`
3. Use `StatusPartial` if partial results available
4. Suggest escalation path if applicable

**For Non-Recoverable Errors:**
1. Set `Recoverable: false` in `AgentError`
2. Use `StatusError` response status
3. Log full error context for debugging
4. Return nil `Output` field

**Example Error Response:**
```go
&schema.AgentResponse{
    RequestID: req.RequestID,
    AgentID:   "analyzer",
    Status:    schema.StatusError,
    Error: &schema.AgentError{
        Code:        "FILE_NOT_FOUND",
        Message:     "Source file does not exist: /path/to/missing.go",
        Recoverable: true,
        Details:     "Please verify the file path and ensure the file exists. Agent can retry with correct path.",
    },
    Duration:  100 * time.Millisecond,
    Timestamp: time.Now(),
}
```

---

## API Conventions

### Naming Conventions

#### Package Names
- Single word, lowercase (e.g., `schema`, `profiling`, `evidence`)
- Avoid underscores or mixed case

#### Type Names
- PascalCase for exported types (e.g., `AgentOutputV1`, `ValidationResult`)
- camelCase for unexported types (e.g., `executionContext`)

#### Field Names
- PascalCase for exported fields (e.g., `ComponentName`, `CallGraph`)
- JSON tags use snake_case (e.g., `component_name`, `call_graph`)

#### Method Names
- PascalCase for exported methods (e.g., `Validate`, `StartExecution`)
- Verb-first naming (e.g., `GetMetrics`, `BuildEvidence`, `ValidateSchema`)

### Error Handling Patterns

#### Error Wrapping
```go
if err != nil {
    return fmt.Errorf("failed to parse file: %w", err)
}
```

#### Error Checking
```go
result, err := validator.Validate(output)
if err != nil {
    return nil, err  // Propagate validation error
}

if !result.Valid {
    return nil, fmt.Errorf("validation failed: %d errors", len(result.Errors))
}
```

### Context Usage

Always pass `context.Context` as first parameter for operations that may block:

```go
func (p *Profiler) StartExecution(ctx context.Context, agent string, request string) *ExecutionContext

func (o *Orchestrator) Execute(ctx context.Context, req *schema.AgentRequest) (*schema.AgentResponse, error)
```

### Nil Safety

All public API methods handle nil inputs gracefully:

```go
func (v *Validator) Validate(output *schema.AgentOutputV1) (*ValidationResult, error) {
    if output == nil {
        return nil, fmt.Errorf("output is nil")
    }
    // ... validation logic
}
```

### Thread Safety

- All validators are stateless and thread-safe
- Profiler uses internal mutexes for concurrent access
- No shared mutable state in public APIs

### JSON Serialization

All schema types support JSON marshaling/unmarshaling:

```go
// Serialize
data, err := json.Marshal(agentOutput)

// Deserialize
var output schema.AgentOutputV1
err := json.Unmarshal(data, &output)
```

### Absolute Path Requirements

File paths in schema structures must be absolute:

```go
// ✅ Correct
EntryPoint{
    File: "/home/user/project/internal/agent/analyzer.go",
    Lines: "45-52",
}

// ❌ Wrong
EntryPoint{
    File: "internal/agent/analyzer.go",  // Relative path
    Lines: "45-52",
}
```

### Line Range Format

Line ranges support two formats:

1. **Single line**: `"42"`
2. **Range**: `"42-56"`

Ranges must satisfy:
- Start line > 0
- End line > 0
- Start line ≤ End line

---

## Usage Examples

### Complete Agent Implementation

```go
package analyzer

import (
    "context"
    "fmt"
    "time"

    "github.com/ferg-cod3s/conexus/pkg/schema"
    "github.com/ferg-cod3s/conexus/internal/validation/evidence"
    schemaValidator "github.com/ferg-cod3s/conexus/internal/validation/schema"
    "github.com/ferg-cod3s/conexus/internal/profiling"
)

type Analyzer struct {
    evidenceValidator *evidence.Validator
    schemaValidator   *schemaValidator.Validator
    profiler          *profiling.Profiler
}

func NewAnalyzer() *Analyzer {
    return &Analyzer{
        evidenceValidator: evidence.NewValidator(true),  // Strict mode
        schemaValidator:   schemaValidator.NewValidator(false), // No partial
        profiler:          profiling.NewProfiler(true),  // Profiling enabled
    }
}

func (a *Analyzer) Execute(ctx context.Context, req *schema.AgentRequest) (*schema.AgentResponse, error) {
    // Start profiling
    execCtx := a.profiler.StartExecution(ctx, "analyzer", req.Task.SpecificRequest)
    
    var output *schema.AgentOutputV1
    var executeErr error
    defer execCtx.End(output, executeErr)
    
    // Phase 1: Parse
    execCtx.StartPhase("parse")
    parseResult, err := a.parseFiles(req.Task.Files)
    execCtx.EndPhase()
    
    if err != nil {
        executeErr = err
        return &schema.AgentResponse{
            RequestID: req.RequestID,
            AgentID:   "analyzer",
            Status:    schema.StatusError,
            Error: &schema.AgentError{
                Code:        "PARSE_ERROR",
                Message:     fmt.Sprintf("Failed to parse files: %v", err),
                Recoverable: true,
                Details:     "Verify files are valid Go source code",
            },
            Duration:  time.Since(req.Timestamp),
            Timestamp: time.Now(),
        }, nil
    }
    
    // Phase 2: Analyze
    execCtx.StartPhase("analyze")
    output = a.analyze(parseResult)
    execCtx.EndPhase()
    
    // Phase 3: Validate schema
    execCtx.StartPhase("validate_schema")
    schemaResult, err := a.schemaValidator.Validate(output)
    execCtx.EndPhase()
    
    if err != nil || !schemaResult.Valid {
        executeErr = fmt.Errorf("schema validation failed")
        return &schema.AgentResponse{
            RequestID: req.RequestID,
            AgentID:   "analyzer",
            Status:    schema.StatusError,
            Error: &schema.AgentError{
                Code:        "SCHEMA_INVALID",
                Message:     "Output does not conform to AGENT_OUTPUT_V1 schema",
                Recoverable: true,
                Details:     fmt.Sprintf("Errors: %d", len(schemaResult.Errors)),
            },
            Duration:  time.Since(req.Timestamp),
            Timestamp: time.Now(),
        }, nil
    }
    
    // Phase 4: Validate evidence
    execCtx.StartPhase("validate_evidence")
    evidenceResult, err := a.evidenceValidator.Validate(output)
    execCtx.EndPhase()
    
    if err != nil || !evidenceResult.Valid {
        executeErr = fmt.Errorf("evidence validation failed")
        return &schema.AgentResponse{
            RequestID: req.RequestID,
            AgentID:   "analyzer",
            Status:    schema.StatusError,
            Error: &schema.AgentError{
                Code:        "EVIDENCE_INSUFFICIENT",
                Message:     fmt.Sprintf("Evidence coverage: %.1f%% (100%% required)", evidenceResult.CoveragePercentage),
                Recoverable: true,
                Details:     fmt.Sprintf("Unbacked claims: %d", len(evidenceResult.UnbackedClaims)),
            },
            Duration:  time.Since(req.Timestamp),
            Timestamp: time.Now(),
        }, nil
    }
    
    // Success
    return &schema.AgentResponse{
        RequestID: req.RequestID,
        AgentID:   "analyzer",
        Status:    schema.StatusComplete,
        Output:    output,
        Duration:  time.Since(req.Timestamp),
        Timestamp: time.Now(),
    }, nil
}

func (a *Analyzer) parseFiles(files []string) (interface{}, error) {
    // Implementation
    return nil, nil
}

func (a *Analyzer) analyze(parseResult interface{}) *schema.AgentOutputV1 {
    // Implementation
    return &schema.AgentOutputV1{
        Version:       "AGENT_OUTPUT_V1",
        ComponentName: "analyzer",
        // ... other fields
    }
}
```

### Evidence Collection Pattern

```go
func collectEvidence(output *schema.AgentOutputV1) {
    // Start with empty evidence
    output.RawEvidence = make([]schema.Evidence, 0)
    
    // Evidence for entry points
    for _, ep := range output.EntryPoints {
        output.RawEvidence = append(output.RawEvidence, schema.Evidence{
            Claim: fmt.Sprintf("Entry point %s at %s:%s", ep.Symbol, ep.File, ep.Lines),
            File:  ep.File,
            Lines: ep.Lines,
        })
    }
    
    // Evidence for call graph
    for _, edge := range output.CallGraph {
        fileParts := strings.Split(edge.From, ":")
        if len(fileParts) >= 1 {
            output.RawEvidence = append(output.RawEvidence, schema.Evidence{
                Claim: fmt.Sprintf("Call from %s to %s", edge.From, edge.To),
                File:  fileParts[0],
                Lines: fmt.Sprintf("%d", edge.ViaLine),
            })
        }
    }
    
    // Evidence for transformations
    for _, transform := range output.DataFlow.Transformations {
        output.RawEvidence = append(output.RawEvidence, schema.Evidence{
            Claim: fmt.Sprintf("Transformation: %s", transform.Description),
            File:  transform.File,
            Lines: transform.Lines,
        })
    }
    
    // Continue for all structural sections...
}
```

### Performance Monitoring Pattern

```go
func monitorPerformance(profiler *profiling.Profiler) {
    // Run monitoring loop
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // Get all metrics
        allMetrics := profiler.GetAllMetrics()
        
        // Check each agent
        for agent, metrics := range allMetrics {
            // Log performance
            fmt.Printf("Agent %s: %d executions, %.1f%% success, avg %v\n",
                agent,
                metrics.TotalExecutions,
                float64(metrics.SuccessCount)/float64(metrics.TotalExecutions)*100,
                metrics.AvgDuration,
            )
            
            // Alert on high P95
            if metrics.Percentiles != nil && metrics.Percentiles.P95 > 500*time.Millisecond {
                fmt.Printf("  WARNING: P95 duration high: %v\n", metrics.Percentiles.P95)
            }
        }
        
        // Check bottlenecks
        bottlenecks := profiler.GetBottlenecks(100 * time.Millisecond)
        if len(bottlenecks) > 0 {
            fmt.Printf("\nBottlenecks detected:\n")
            for _, b := range bottlenecks {
                fmt.Printf("  %s: %s (%v, severity: %s)\n",
                    b.Agent, b.Type, b.AvgDuration, b.Severity)
            }
        }
    }
}
```

### Validation Pipeline Pattern

```go
func validateOutput(
    output *schema.AgentOutputV1,
    schemaVal *schemaValidator.Validator,
    evidenceVal *evidence.Validator,
) error {
    // Step 1: Schema validation
    schemaResult, err := schemaVal.Validate(output)
    if err != nil {
        return fmt.Errorf("schema validation error: %w", err)
    }
    
    if !schemaResult.Valid {
        return fmt.Errorf("schema validation failed: %d errors, %d warnings",
            len(schemaResult.Errors), len(schemaResult.Warnings))
    }
    
    // Step 2: Evidence validation
    evidenceResult, err := evidenceVal.Validate(output)
    if err != nil {
        return fmt.Errorf("evidence validation error: %w", err)
    }
    
    if !evidenceResult.Valid {
        return fmt.Errorf("evidence validation failed: %.1f%% coverage (%d unbacked claims)",
            evidenceResult.CoveragePercentage, len(evidenceResult.UnbackedClaims))
    }
    
    // Require 100% evidence coverage
    if evidenceResult.CoveragePercentage < 100.0 {
        return fmt.Errorf("insufficient evidence: %.1f%% (100%% required)",
            evidenceResult.CoveragePercentage)
    }
    
    return nil
}
```

---

## Migration Guide

### Future Schema Versions

When a new schema version is introduced (e.g., `AGENT_OUTPUT_V2`):

1. **Version Detection**:
   ```go
   switch output.Version {
   case "AGENT_OUTPUT_V1":
       // Use V1 validators
   case "AGENT_OUTPUT_V2":
       // Use V2 validators
   default:
       return fmt.Errorf("unsupported schema version: %s", output.Version)
   }
   ```

2. **Backward Compatibility**:
   - V1 validators continue to work with V1 outputs
   - New fields are additive (old code ignores them)
   - Breaking changes require major version bump

3. **Deprecation Process**:
   - 90-day deprecation notice
   - Parallel support for old and new versions
   - Migration tooling provided

### API Stability Promise

- **Stable APIs**: `pkg/schema` types and public validator methods
- **Internal APIs**: May change without notice (internal packages)
- **Experimental**: Marked in comments, subject to change

---

## API Support

### Documentation
- **API Reference**: This document
- **User Guide**: `docs/README.md`
- **Architecture**: `docs/architecture/integration.md`
- **Validation Guide**: `docs/validation-guide.md`
- **Profiling Guide**: `docs/profiling-guide.md`

### Issue Reporting

For API issues:
1. Check this reference documentation
2. Review architecture documentation
3. Check test files for usage examples
4. Open GitHub issue with:
   - API method/type name
   - Expected behavior
   - Actual behavior
   - Minimal reproduction code

### Version Information

```bash
# Check schema version
grep -r "AGENT_OUTPUT_V" pkg/schema/agent_output_v1.go

# Run API tests
go test ./pkg/schema/...
go test ./internal/validation/...
go test ./internal/profiling/...
```

---

**Document Version**: 1.0.0  
**Schema Version**: AGENT_OUTPUT_V1  
**Last Updated**: Phase 5 Documentation Complete  
**Maintained By**: Conexus Development Team
