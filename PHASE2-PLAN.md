# Phase 2 Implementation Plan: Essential Agents

**Status**: Ready to Start
**Prerequisites**: ✅ Phase 1 Complete (Core Infrastructure)
**Date**: 2025-10-14

---

## Overview

Phase 2 implements the three core agents that form the foundation of Conexus:
1. **codebase-locator**: File and symbol discovery
2. **codebase-analyzer**: Code analysis with evidence-grounded output
3. **orchestrator**: Request routing and agent coordination

---

## Current State

### Completed (Phase 1)
- ✅ AGENT_OUTPUT_V1 schema (`pkg/schema/`)
- ✅ Tool execution framework (`internal/tool/`)
- ✅ Process management (`internal/process/`)
- ✅ JSON-RPC protocol (`internal/protocol/`)
- ✅ GitHub project updated

### Directory Structure Ready
```
internal/
├── agent/          # Empty - Phase 2 agents go here
├── orchestrator/   # Empty - Orchestrator implementation
├── tool/           # ✅ Complete
├── process/        # ✅ Complete
└── protocol/       # ✅ Complete
```

---

## Phase 2 Tasks Breakdown

### Task 2.1: Implement codebase-locator Agent

**Purpose**: Find files, symbols, and code patterns in the target codebase

**Location**: `internal/agent/locator/`

**Files to Create**:
```
internal/agent/locator/
├── locator.go           # Main agent implementation
├── locator_test.go      # Unit tests
├── strategies.go        # Search strategies
└── README.md            # Agent documentation
```

**Core Responsibilities**:
1. **File Discovery**:
   - Pattern-based file search (glob)
   - Extension filtering
   - Path filtering by allowed directories

2. **Symbol Search**:
   - Function/struct/interface definitions
   - Export declarations
   - Import statements
   - Type definitions

3. **Multi-Strategy Search**:
   - Grep-based text search
   - AST-based symbol extraction (future)
   - Heuristic matching (future)

4. **Output Format**:
   - Must conform to AGENT_OUTPUT_V1
   - Entry points for found symbols
   - Evidence backing for all discoveries

**Implementation Steps**:

1. **Create Agent Structure** (locator.go):
   ```go
   type LocatorAgent struct {
       executor *tool.Executor
       perms    schema.Permissions
   }

   func New(executor *tool.Executor, perms schema.Permissions) *LocatorAgent
   func (a *LocatorAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error)
   ```

2. **Implement Search Strategies** (strategies.go):
   ```go
   type SearchStrategy interface {
       Search(ctx context.Context, query string) ([]SearchResult, error)
   }

   type GlobStrategy struct {}        // Pattern-based file search
   type GrepStrategy struct {}        // Text-based content search
   type SymbolStrategy struct {}      // Symbol extraction (future)
   ```

3. **Build Output Generator**:
   - Convert search results to AGENT_OUTPUT_V1
   - Generate evidence entries for each result
   - Include file paths, line numbers, symbol names

4. **Write Tests** (locator_test.go):
   - Test file discovery with various patterns
   - Test symbol search (basic text-based initially)
   - Test permission boundaries
   - Test AGENT_OUTPUT_V1 compliance

**Acceptance Criteria**:
- ✅ Find files by pattern (e.g., "*.go", "**/*_test.go")
- ✅ Find simple symbols (functions, types) via grep
- ✅ Respect permission boundaries
- ✅ Output valid AGENT_OUTPUT_V1 JSON
- ✅ All claims have evidence backing
- ✅ Tests passing with >80% coverage

**Estimated Complexity**: Medium (150-200 LOC + tests)

---

### Task 2.2: Implement codebase-analyzer Agent

**Purpose**: Analyze code to understand control flow, data flow, and patterns

**Location**: `internal/agent/analyzer/`

**Files to Create**:
```
internal/agent/analyzer/
├── analyzer.go          # Main agent implementation
├── analyzer_test.go     # Unit tests
├── parser.go            # Code parsing utilities
├── flow.go              # Control/data flow analysis
└── README.md            # Agent documentation
```

**Core Responsibilities**:
1. **Entry Point Identification**:
   - Exported functions
   - HTTP handlers
   - Main functions
   - Public interfaces

2. **Call Graph Construction**:
   - Function → function relationships
   - Method invocations
   - Line-level precision

3. **Data Flow Analysis**:
   - Input identification (parameters, reads)
   - Transformations (operations on data)
   - Outputs (returns, writes)

4. **State Management Detection**:
   - Variable assignments
   - Struct field access
   - Database operations (if detectable)

5. **Side Effect Tracking**:
   - Log statements
   - File I/O
   - External calls

6. **Error Handling Analysis**:
   - Error returns
   - Panic/recover
   - Guard clauses

**Implementation Steps**:

1. **Create Agent Structure** (analyzer.go):
   ```go
   type AnalyzerAgent struct {
       executor *tool.Executor
       perms    schema.Permissions
   }

   func New(executor *tool.Executor, perms schema.Permissions) *AnalyzerAgent
   func (a *AnalyzerAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error)
   ```

2. **Implement Basic Parser** (parser.go):
   - Read file content via tool executor
   - Identify function boundaries (basic regex initially)
   - Extract function signatures
   - Find import statements

3. **Build Flow Analyzers** (flow.go):
   ```go
   type FlowAnalyzer interface {
       Analyze(ctx context.Context, code string) (FlowResult, error)
   }

   type CallGraphAnalyzer struct {}   // Finds function calls
   type DataFlowAnalyzer struct {}    // Tracks data transformations
   type StateAnalyzer struct {}       // Finds state operations
   ```

4. **Generate AGENT_OUTPUT_V1**:
   - Populate all relevant sections
   - Create evidence entries for every claim
   - Use file:line references

5. **Write Tests** (analyzer_test.go):
   - Test with sample Go code
   - Verify entry point detection
   - Verify call graph construction
   - Test AGENT_OUTPUT_V1 compliance
   - Test evidence completeness

**Acceptance Criteria**:
- ✅ Identify entry points (exported functions)
- ✅ Build basic call graph (function → function)
- ✅ Track data inputs/outputs
- ✅ Detect side effects (logs, I/O)
- ✅ Output valid AGENT_OUTPUT_V1 JSON
- ✅ 100% evidence backing for all claims
- ✅ Tests passing with >80% coverage

**Estimated Complexity**: High (300-400 LOC + tests)

**Note**: Initial implementation will use text analysis (grep, regex). AST parsing can be added in Phase 5 for more sophisticated analysis.

---

### Task 2.3: Implement Basic Orchestrator

**Purpose**: Route user requests to appropriate agents and coordinate workflows

**Location**: `internal/orchestrator/`

**Files to Create**:
```
internal/orchestrator/
├── orchestrator.go      # Main orchestrator implementation
├── orchestrator_test.go # Unit tests
├── router.go            # Request routing logic
├── workflow.go          # Sequential workflow execution
└── README.md            # Orchestrator documentation
```

**Core Responsibilities**:
1. **Request Routing**:
   - Parse user intent
   - Select appropriate agent(s)
   - Map request to agent task

2. **Agent Invocation**:
   - Spawn agent processes
   - Send AgentRequest via JSON-RPC
   - Collect AgentResponse

3. **Sequential Workflow**:
   - Execute agents in sequence
   - Pass context between agents
   - Aggregate results

4. **Error Handling**:
   - Handle agent failures
   - Process escalation requests
   - Provide user-friendly errors

**Implementation Steps**:

1. **Create Orchestrator Structure** (orchestrator.go):
   ```go
   type Orchestrator struct {
       processManager *process.Manager
       toolExecutor   *tool.Executor
       agentRegistry  map[string]AgentFactory
   }

   func New(pm *process.Manager, te *tool.Executor) *Orchestrator
   func (o *Orchestrator) HandleRequest(ctx context.Context, userRequest string) (Result, error)
   ```

2. **Implement Request Router** (router.go):
   ```go
   type Router struct {}

   func (r *Router) Route(userRequest string) (AgentSelection, error)

   type AgentSelection struct {
       PrimaryAgent   string
       FallbackAgents []string
       Parameters     map[string]interface{}
   }
   ```

3. **Build Workflow Engine** (workflow.go):
   ```go
   type WorkflowEngine struct {
       orchestrator *Orchestrator
   }

   func (w *WorkflowEngine) ExecuteSequential(ctx context.Context, agents []string) ([]schema.AgentResponse, error)
   ```

4. **Integrate Agent Registry**:
   - Register locator agent
   - Register analyzer agent
   - Factory functions for agent creation

5. **Write Tests** (orchestrator_test.go):
   - Test routing logic (user request → agent selection)
   - Test agent invocation (mocked agents)
   - Test sequential workflows
   - Test error handling and escalation

**Acceptance Criteria**:
- ✅ Route simple requests to correct agent
  - "Find all .go files" → locator
  - "Analyze main.go" → analyzer
- ✅ Invoke agents via process manager
- ✅ Execute 2+ agents sequentially
- ✅ Handle agent errors gracefully
- ✅ Process escalation requests
- ✅ Tests passing with >80% coverage

**Estimated Complexity**: High (250-350 LOC + tests)

---

## Implementation Order

**Recommended sequence**:

1. **Week 1: Task 2.1 (codebase-locator)**
   - Day 1-2: Core structure + file discovery
   - Day 3-4: Symbol search (grep-based)
   - Day 5: AGENT_OUTPUT_V1 generation
   - Day 6-7: Tests + documentation

2. **Week 2: Task 2.2 (codebase-analyzer)**
   - Day 1-3: Core structure + basic parsing
   - Day 4-5: Flow analysis (call graph, data flow)
   - Day 6: AGENT_OUTPUT_V1 generation
   - Day 7: Tests + documentation

3. **Week 3: Task 2.3 (orchestrator)**
   - Day 1-2: Core structure + routing
   - Day 3-4: Agent invocation + workflow
   - Day 5-6: Integration + error handling
   - Day 7: Tests + end-to-end validation

**Total Estimated Time**: 3 weeks (assuming 1 developer, part-time)

---

## Testing Strategy

### Unit Tests
- Each agent in isolation
- Mocked tool executor
- Fixed test fixtures (sample Go files)

### Integration Tests
- Agent → tool executor → actual files
- Agent → process manager → JSON-RPC
- Orchestrator → agents (end-to-end)

### Test Fixtures
Create `tests/fixtures/` directory with sample code:
```
tests/fixtures/
├── simple_function.go      # Single function
├── multiple_functions.go   # Multiple functions with calls
├── struct_methods.go       # Struct with methods
├── error_handling.go       # Error patterns
└── side_effects.go         # Logs, I/O operations
```

---

## Success Criteria (Phase 2 Complete)

### Functional Requirements
- ✅ codebase-locator can find files and basic symbols
- ✅ codebase-analyzer produces valid AGENT_OUTPUT_V1
- ✅ Orchestrator routes requests to correct agents
- ✅ Sequential workflows execute successfully
- ✅ All agents respect permission boundaries

### Quality Requirements
- ✅ Test coverage >80% for all components
- ✅ 100% evidence backing in analyzer output
- ✅ All AGENT_OUTPUT_V1 outputs validate
- ✅ No critical bugs in core workflows

### Performance Requirements
- ✅ Locator response: <2s for 1000 file repository
- ✅ Analyzer response: <5s for 500 LOC file
- ✅ Orchestrator overhead: <100ms

### Documentation Requirements
- ✅ README.md for each agent
- ✅ Code comments for public interfaces
- ✅ Test documentation
- ✅ Updated PHASE2-STATUS.md

---

## Dependencies

### External Libraries
- **Go standard library only** (no external dependencies initially)
- Consider adding later:
  - `go/parser` - Go AST parsing (Phase 5)
  - `go/token` - Source position tracking
  - `regexp` - Pattern matching (if not using built-in)

### Internal Dependencies
- `pkg/schema` - AGENT_OUTPUT_V1 types
- `internal/tool` - Tool execution
- `internal/process` - Process management
- `internal/protocol` - JSON-RPC communication

---

## Risk Mitigation

### Risk 1: AST Parsing Complexity
**Mitigation**: Start with text-based analysis (grep, regex). AST parsing is Phase 5 enhancement.

### Risk 2: Evidence Generation Overhead
**Mitigation**: Keep line ranges minimal. Cache file content during analysis session.

### Risk 3: Orchestrator Complexity
**Mitigation**: Start with simple routing rules. Advanced intent parsing is Phase 3.

### Risk 4: Test Coverage
**Mitigation**: Write tests alongside implementation (TDD approach). Use fixtures for consistent testing.

---

## Next Steps (Immediate)

1. **Create test fixtures** (`tests/fixtures/*.go`)
2. **Start Task 2.1**: Implement codebase-locator
   - Create `internal/agent/locator/locator.go`
   - Implement basic file discovery
   - Write first tests

3. **Update GitHub Project**:
   - Mark Phase 2 tasks as "In Progress"
   - Add specific implementation subtasks if needed

4. **Set up continuous validation**:
   - Run `go test ./...` frequently
   - Validate AGENT_OUTPUT_V1 schema compliance

---

## Questions to Resolve

1. **Agent Communication**: Should agents be compiled as separate binaries or invoked as library functions?
   - **Recommendation**: Start as library functions, migrate to processes in Phase 3

2. **Caching Strategy**: When should analysis results be cached?
   - **Recommendation**: Phase 3 concern, skip for now

3. **Parallelization**: Should locator search strategies run in parallel?
   - **Recommendation**: Sequential initially, parallel in Phase 5

4. **Error Recovery**: How should orchestrator handle partial failures?
   - **Recommendation**: Fail fast initially, graceful degradation in Phase 3

---

## Resources

### Documentation
- [POC-PLAN.md](./POC-PLAN.md) - Overall POC plan
- [PHASE1-STATUS.md](./PHASE1-STATUS.md) - Phase 1 completion status
- [README.md](./README.md) - Project overview

### Code References
- `pkg/schema/agent_output_v1.go` - Output schema
- `internal/tool/tool.go` - Tool execution examples
- `internal/process/manager.go` - Process management patterns

### External References
- Go AST: https://pkg.go.dev/go/ast
- JSON-RPC 2.0: https://www.jsonrpc.org/specification
- MCP Protocol: https://modelcontextprotocol.io/

---

**Ready to begin implementation!**

Use this plan as a reference throughout Phase 2 development. Update this document as decisions are made and requirements evolve.
