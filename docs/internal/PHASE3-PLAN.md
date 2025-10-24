# Phase 3 Implementation Plan: Orchestration & Workflow

**Status**: Ready to Start
**Prerequisites**: ✅ Phase 2 Complete (Essential Agents)
**Date**: 2025-10-14

---

## Overview

Phase 3 enhances the orchestrator with advanced coordination capabilities, intelligent routing, state management, and multi-agent workflows. This phase transforms the basic orchestrator into a sophisticated workflow engine.

---

## Current State (Post-Phase 2)

### ✅ What We Have

**Orchestrator Capabilities**:
- Basic keyword-based routing
- Sequential workflow execution
- Agent registry and factories
- Context propagation between agents
- Simple escalation handling
- Permission enforcement

**Working Agents**:
- codebase-locator (file/symbol discovery)
- codebase-analyzer (code analysis)

**Test Coverage**: 82.9% orchestrator, 88% overall Phase 2

### 🔧 What's Missing

**Routing Limitations**:
- Only keyword matching (no NLU)
- No confidence scoring
- No fallback strategies
- Limited parameter extraction

**Workflow Limitations**:
- Sequential only (no parallelization)
- No conditional branching
- No loop detection
- No workflow persistence
- No timeout handling
- No partial result recovery

**State Management**:
- No caching of results
- No session persistence
- Context not retained across requests
- No conversation history

**Escalation**:
- Basic implementation only
- No intelligent agent selection
- No escalation history
- No circular escalation prevention

---

## Phase 3 Tasks Breakdown

### Task 3.1: Implement Intent Parsing Logic

**Purpose**: Transform natural language requests into structured agent invocations

**Location**: `internal/orchestrator/intent/`

**Files to Create**:
```
internal/orchestrator/intent/
├── parser.go            # Main intent parser
├── parser_test.go       # Parser tests
├── patterns.go          # Intent patterns and rules
├── confidence.go        # Confidence scoring
└── README.md            # Intent parsing documentation
```

**Core Responsibilities**:

1. **Natural Language Understanding**:
   - Parse user requests into structured intents
   - Extract entities (file names, symbols, operations)
   - Identify action verbs and targets
   - Handle ambiguous requests

2. **Intent Classification**:
   ```go
   type Intent struct {
       Action     string                 // "find", "analyze", "search"
       Target     string                 // "files", "function", "code"
       Entities   map[string]string      // {"pattern": "*.go", "symbol": "Add"}
       Confidence float64                // 0.0 - 1.0
       AgentID    string                 // Selected agent
       Fallbacks  []string               // Alternative agents
   }
   ```

3. **Entity Extraction**:
   - File patterns (*.go, **/*.ts)
   - Symbol names (function names, types)
   - Directories
   - Code locations (file:line)

4. **Confidence Scoring**:
   - Score routing decisions 0.0-1.0
   - Threshold for fallback selection
   - Multi-agent suggestions for ambiguous requests

**Implementation Steps**:

1. **Create Parser Structure**:
   ```go
   type IntentParser struct {
       patterns   []IntentPattern
       extractors map[string]EntityExtractor
   }

   type IntentPattern struct {
       Name       string
       Regex      *regexp.Regexp
       AgentID    string
       Priority   int
       Extractor  func(string) map[string]string
   }
   ```

2. **Implement Pattern Matching**:
   ```go
   patterns := []IntentPattern{
       {
           Name:     "find_files",
           Regex:    regexp.MustCompile(`find.*files?.*\.(go|ts|js)`),
           AgentID:  "codebase-locator",
           Priority: 10,
       },
       {
           Name:     "analyze_function",
           Regex:    regexp.MustCompile(`(analyze|how|understand).*function\s+(\w+)`),
           AgentID:  "codebase-analyzer",
           Priority: 10,
       },
   }
   ```

3. **Build Confidence Scorer**:
   - Keyword density
   - Pattern strength
   - Entity presence
   - Historical success rate

4. **Implement Entity Extractors**:
   - File pattern extractor
   - Symbol name extractor
   - Directory extractor
   - Code location extractor

5. **Create Fallback Logic**:
   - Confidence threshold (0.6)
   - Multiple agent suggestions
   - Clarification questions

**Acceptance Criteria**:
- ✅ Parse 20+ common request patterns
- ✅ Extract entities with >90% accuracy
- ✅ Confidence scores accurate within 10%
- ✅ Suggest fallbacks for ambiguous requests
- ✅ Tests covering all patterns
- ✅ >85% test coverage

**Estimated Complexity**: Medium-High (250-300 LOC + tests)

---

### Task 3.2: Build Multi-Agent Workflow Coordination

**Purpose**: Enable parallel execution, conditional logic, and advanced workflows

**Location**: `internal/orchestrator/workflow/`

**Files to Create**:
```
internal/orchestrator/workflow/
├── engine.go            # Workflow execution engine
├── engine_test.go       # Engine tests
├── executor.go          # Parallel/sequential executor
├── graph.go             # Workflow DAG structure
├── validator.go         # Workflow validation
└── README.md            # Workflow documentation
```

**Core Responsibilities**:

1. **Workflow Graph**:
   ```go
   type WorkflowGraph struct {
       Nodes      map[string]*WorkflowNode
       Edges      []WorkflowEdge
       StartNode  string
       EndNode    string
   }

   type WorkflowNode struct {
       ID          string
       AgentID     string
       Request     string
       Parallel    bool          // Can run in parallel
       Condition   *Condition    // Conditional execution
       Timeout     time.Duration
   }

   type WorkflowEdge struct {
       From        string
       To          string
       Condition   *Condition
   }
   ```

2. **Execution Strategies**:
   - **Sequential**: One agent at a time (current)
   - **Parallel**: Multiple agents concurrently
   - **Conditional**: Branch based on results
   - **Loop**: Iterate until condition met

3. **Parallel Execution**:
   ```go
   type ParallelExecutor struct {
       maxConcurrent int
       workerpool    chan struct{}
   }

   func (e *ParallelExecutor) ExecuteParallel(
       ctx context.Context,
       nodes []*WorkflowNode,
   ) ([]schema.AgentResponse, error)
   ```

4. **Conditional Branching**:
   ```go
   type Condition struct {
       Type      string // "status", "field_exists", "field_value"
       Field     string
       Operator  string // "==", "!=", ">", "<", "contains"
       Value     interface{}
   }
   ```

5. **Timeout Handling**:
   - Per-agent timeouts
   - Workflow-level timeout
   - Graceful cancellation
   - Partial result recovery

6. **Error Recovery**:
   - Continue on error (with flag)
   - Retry failed agents
   - Fallback to alternative agents
   - Collect partial results

**Implementation Steps**:

1. **Create Workflow Graph Structure**:
   - Node and edge definitions
   - DAG validation (no cycles)
   - Topological sorting

2. **Implement Sequential Executor** (enhance existing):
   - Add timeout support
   - Add error recovery
   - Add condition evaluation

3. **Build Parallel Executor**:
   - Goroutine pool
   - WaitGroup coordination
   - Error collection
   - Result aggregation

4. **Add Conditional Logic**:
   - Condition evaluator
   - Branch selection
   - Skip nodes based on conditions

5. **Create Workflow Validator**:
   - Check DAG structure
   - Validate agent references
   - Check condition syntax
   - Detect infinite loops

6. **Implement Timeout Management**:
   - Context-based cancellation
   - Partial result collection
   - Timeout error reporting

**Acceptance Criteria**:
- ✅ Execute 2+ agents in parallel
- ✅ Handle conditional branching
- ✅ Respect timeouts (agent & workflow)
- ✅ Recover from partial failures
- ✅ Validate workflow graphs
- ✅ Prevent infinite loops
- ✅ Tests covering all execution modes
- ✅ >85% test coverage

**Estimated Complexity**: High (400-500 LOC + tests)

---

### Task 3.3: Implement Escalation Protocol

**Purpose**: Enhanced agent collaboration and intelligent escalation

**Location**: `internal/orchestrator/escalation/`

**Files to Create**:
```
internal/orchestrator/escalation/
├── handler.go           # Escalation handler
├── handler_test.go      # Handler tests
├── policy.go            # Escalation policies
├── history.go           # Escalation tracking
└── README.md            # Escalation documentation
```

**Core Responsibilities**:

1. **Escalation Policy**:
   ```go
   type EscalationPolicy struct {
       MaxDepth          int                    // Max escalation chain depth
       CircularPrevention bool                  // Prevent A→B→A
       TimeoutPerLevel   time.Duration
       FallbackStrategy  string                 // "fail", "best_effort", "retry"
   }
   ```

2. **Escalation Handler**:
   ```go
   type EscalationHandler struct {
       policy  *EscalationPolicy
       history *EscalationHistory
   }

   func (h *EscalationHandler) HandleEscalation(
       ctx context.Context,
       response schema.AgentResponse,
   ) (*WorkflowNode, error)
   ```

3. **Escalation History**:
   - Track escalation chains
   - Detect circular escalations
   - Record success/failure rates
   - Learn from patterns

4. **Intelligent Agent Selection**:
   - Analyze escalation reason
   - Match capabilities to needs
   - Consider agent availability
   - Check historical success

5. **Escalation Context Enrichment**:
   - Pass partial results
   - Include failure reasons
   - Add context breadcrumbs
   - Maintain evidence chain

**Implementation Steps**:

1. **Create Escalation Handler**:
   - Detect escalation requests
   - Select target agent
   - Build escalation context
   - Add to workflow

2. **Implement Policy Engine**:
   - Max depth checking
   - Circular detection
   - Timeout enforcement
   - Fallback selection

3. **Build History Tracker**:
   - Record escalations
   - Track success rates
   - Identify patterns
   - Provide analytics

4. **Add Agent Selector**:
   - Capability matching
   - Success rate weighting
   - Availability checking
   - Load balancing

5. **Create Context Enricher**:
   - Merge agent outputs
   - Preserve evidence
   - Add escalation metadata
   - Maintain traceability

**Acceptance Criteria**:
- ✅ Handle escalation chains 3+ levels deep
- ✅ Prevent circular escalations
- ✅ Intelligent agent selection
- ✅ Track escalation success rates
- ✅ Enrich escalation context
- ✅ Enforce policies consistently
- ✅ Tests covering all scenarios
- ✅ >85% test coverage

**Estimated Complexity**: Medium (200-250 LOC + tests)

---

### Task 3.4: Create State Management & Caching

**Purpose**: Persist workflow state, cache results, manage sessions

**Location**: `internal/orchestrator/state/`

**Files to Create**:
```
internal/orchestrator/state/
├── manager.go           # State manager
├── manager_test.go      # Manager tests
├── cache.go             # Result caching
├── session.go           # Session management
├── persistence.go       # State persistence
└── README.md            # State documentation
```

**Core Responsibilities**:

1. **Session Management**:
   ```go
   type Session struct {
       ID              string
       UserID          string
       StartTime       time.Time
       LastActivity    time.Time
       ConversationLog []ConversationEntry
       Context         map[string]interface{}
       State           SessionState
   }

   type ConversationEntry struct {
       Timestamp    time.Time
       UserRequest  string
       AgentResults []schema.AgentResponse
   }
   ```

2. **Result Caching**:
   ```go
   type Cache struct {
       store      map[string]*CacheEntry
       ttl        time.Duration
       maxSize    int
   }

   type CacheEntry struct {
       Key        string
       Value      *schema.AgentOutputV1
       CreatedAt  time.Time
       AccessedAt time.Time
       HitCount   int
   }
   ```

3. **Cache Key Generation**:
   - Content-based hashing
   - Parameter normalization
   - Version tagging
   - Invalidation rules

4. **State Persistence**:
   - In-memory (default)
   - File-based (JSON)
   - Database (future)
   - State recovery

5. **Context Accumulation**:
   - Merge agent outputs
   - Build knowledge graph
   - Track dependencies
   - Maintain history

6. **Cache Invalidation**:
   - Time-based (TTL)
   - File change detection
   - Manual invalidation
   - LRU eviction

**Implementation Steps**:

1. **Create Session Manager**:
   - Session creation
   - Session lookup
   - Session expiration
   - Cleanup

2. **Implement Cache**:
   - In-memory cache
   - Key generation
   - Get/Set operations
   - Eviction policy (LRU)

3. **Build Persistence Layer**:
   - JSON serialization
   - File-based storage
   - Load/Save operations
   - Atomic writes

4. **Add Context Accumulator**:
   - Merge outputs
   - Deduplicate information
   - Build relationships
   - Query interface

5. **Create Invalidation Logic**:
   - TTL expiration
   - File watcher integration
   - Manual purge API
   - Selective invalidation

6. **Implement Statistics**:
   - Cache hit rate
   - Session duration
   - Context growth
   - Performance metrics

**Acceptance Criteria**:
- ✅ Session management working
- ✅ Cache hit rate >70% for repeated queries
- ✅ State persists across restarts
- ✅ Context accumulation functional
- ✅ Cache invalidation correct
- ✅ Memory-efficient (bounded size)
- ✅ Tests covering all operations
- ✅ >85% test coverage

**Estimated Complexity**: Medium-High (350-400 LOC + tests)

---

## Implementation Order

### Recommended Sequence

**Week 1: Task 3.1 (Intent Parsing)**
- Day 1-2: Pattern matching and entity extraction
- Day 3-4: Confidence scoring and fallback logic
- Day 5: Integration with orchestrator
- Day 6-7: Tests and documentation

**Week 2: Task 3.4 (State Management)**
- Day 1-2: Session management and cache
- Day 3-4: Persistence and context accumulation
- Day 5-6: Cache invalidation and statistics
- Day 7: Tests and documentation

**Week 3: Task 3.2 (Workflow Coordination)**
- Day 1-3: Parallel execution and workflow graph
- Day 4-5: Conditional logic and timeout handling
- Day 6: Workflow validation
- Day 7: Tests and documentation

**Week 4: Task 3.3 (Escalation Protocol)**
- Day 1-2: Escalation handler and policy
- Day 3-4: History tracking and agent selection
- Day 5-6: Context enrichment
- Day 7: Tests and integration testing

**Total Estimated Time**: 4 weeks (part-time development)

**Rationale**:
- Intent parsing first: Improves orchestrator immediately
- State management second: Needed for advanced workflows
- Workflow coordination third: Builds on state management
- Escalation last: Integrates all previous components

---

## Integration Points

### Phase 3 → Phase 2 Integration

1. **Intent Parser → Orchestrator Router**:
   - Replace keyword-based routing
   - Use confidence-scored intents
   - Fallback to existing router if needed

2. **Workflow Engine → Orchestrator**:
   - Replace sequential execution
   - Support both modes (sequential/parallel)
   - Maintain backward compatibility

3. **State Manager → Orchestrator**:
   - Cache agent results
   - Persist workflow state
   - Maintain session context

4. **Escalation Handler → Orchestrator**:
   - Enhance existing escalation
   - Add policy enforcement
   - Track escalation history

### New Orchestrator Architecture

```
┌─────────────────────────────────────────────┐
│              Orchestrator                   │
│                                             │
│  ┌─────────────┐       ┌─────────────┐    │
│  │   Intent    │       │   State     │    │
│  │   Parser    │       │   Manager   │    │
│  └──────┬──────┘       └──────┬──────┘    │
│         │                     │            │
│  ┌──────▼─────────────────────▼─────┐     │
│  │      Workflow Engine              │     │
│  │  - Sequential / Parallel          │     │
│  │  - Conditional / Timeout          │     │
│  │  - Error Recovery                 │     │
│  └──────┬────────────────────┬───────┘     │
│         │                    │             │
│  ┌──────▼──────┐      ┌─────▼────────┐    │
│  │ Escalation  │      │    Cache     │    │
│  │  Handler    │      │   Manager    │    │
│  └─────────────┘      └──────────────┘    │
└─────────────────────────────────────────────┘
```

---

## Testing Strategy

### Unit Tests

**Per Component**:
- Intent parser patterns
- Workflow graph validation
- Parallel execution
- Cache operations
- Session management
- Escalation policies

### Integration Tests

**Cross-Component**:
- Intent → Workflow execution
- Cache → State persistence
- Escalation → Workflow addition
- Parallel agents → Result aggregation

### End-to-End Tests

**Complete Workflows**:
- "Find files, then analyze them" (sequential)
- "Analyze multiple files" (parallel)
- "Search with fallback" (conditional)
- "Deep analysis chain" (escalation)

### Performance Tests

**Metrics**:
- Intent parsing: <50ms
- Cache lookup: <1ms
- Parallel execution: 2x speedup for 2 agents
- Session overhead: <10ms

---

## Success Criteria (Phase 3 Complete)

### Functional Requirements

- ✅ Intent parser handles 20+ request patterns
- ✅ Parallel execution of 2+ agents working
- ✅ Conditional workflows functioning
- ✅ Cache hit rate >70% for repeated queries
- ✅ Session state persists across requests
- ✅ Escalation chains work 3+ levels deep
- ✅ Timeout handling prevents hangs
- ✅ Error recovery provides partial results

### Quality Requirements

- ✅ Test coverage >85% for all Phase 3 components
- ✅ No memory leaks in cache or sessions
- ✅ Workflow validation catches invalid graphs
- ✅ Confidence scores accurate within 10%
- ✅ All components fully documented

### Performance Requirements

- ✅ Intent parsing: <50ms
- ✅ Cache lookup: <1ms
- ✅ Parallel speedup: 2x for 2 agents
- ✅ Session overhead: <10ms
- ✅ Memory usage: <1GB for 100 sessions

### Documentation Requirements

- ✅ README.md for each new component
- ✅ Updated orchestrator README
- ✅ PHASE3-STATUS.md created
- ✅ API documentation for public interfaces

---

## Risk Mitigation

### Risk 1: Complexity Explosion
**Mitigation**:
- Start simple (sequential, then parallel)
- Incremental feature addition
- Comprehensive testing at each step

### Risk 2: Race Conditions (Parallel)
**Mitigation**:
- Use Go's concurrency primitives correctly
- Thorough testing with race detector
- Clear ownership of shared state

### Risk 3: Cache Invalidation
**Mitigation**:
- Start with simple TTL
- Add file watching later
- Manual invalidation as backup

### Risk 4: Intent Ambiguity
**Mitigation**:
- Confidence thresholds
- Multi-agent suggestions
- Fallback to keyword router

### Risk 5: State Persistence
**Mitigation**:
- Start in-memory
- JSON file persistence (simple)
- Database later (Phase 5)

---

## Dependencies

### External Libraries (Consider)

**Intent Parsing**:
- No external dependencies initially (use regex)
- Future: NLP libraries (optional)

**Caching**:
- Standard library only
- Future: Redis integration (optional)

**Persistence**:
- JSON (encoding/json)
- Future: Database drivers (optional)

**Workflow**:
- sync, context (stdlib)
- No external dependencies

### Internal Dependencies

- Phase 1: Tool executor, process manager, protocol
- Phase 2: All agents, orchestrator base
- Go stdlib: sync, context, encoding/json, regexp

---

## Phase 3 Deliverables

### Code Deliverables

1. **Intent Parsing System** (4 files, ~300 LOC)
2. **Workflow Engine** (5 files, ~500 LOC)
3. **Escalation Handler** (4 files, ~250 LOC)
4. **State Management** (5 files, ~400 LOC)

**Total**: ~1,450 LOC implementation
**Tests**: ~1,200 LOC (similar to Phase 2 ratio)

### Documentation Deliverables

1. **4 Component READMEs** (~600 lines total)
2. **Updated Orchestrator README**
3. **PHASE3-STATUS.md** (comprehensive status)
4. **API Documentation** (GoDoc comments)

### Test Deliverables

1. **Unit Tests**: All components >85% coverage
2. **Integration Tests**: Cross-component workflows
3. **E2E Tests**: Complete workflow scenarios
4. **Performance Tests**: Benchmarks and profiling

---

## Phase 3 vs Phase 2 Comparison

| Aspect | Phase 2 | Phase 3 |
|--------|---------|---------|
| Complexity | Medium | High |
| LOC (impl) | 1,027 | ~1,450 |
| LOC (tests) | 926 | ~1,200 |
| Components | 3 | 4 |
| Dependencies | 0 external | 0 external |
| Duration | 1 session | 4 weeks |
| Integration | Sequential | Sequential + Parallel |
| State | Stateless | Stateful |
| Routing | Keywords | Intent-based |

---

## Post-Phase 3 Capabilities

Once Phase 3 is complete, Conexus will support:

**Example 1: Parallel Analysis**
```
User: "Analyze all Go files in src/"
→ Locator finds files (parallel)
→ Analyzer analyzes each file (parallel)
→ Results aggregated and cached
```

**Example 2: Conditional Workflow**
```
User: "Find main function, if found analyze it"
→ Locator searches for main
→ If found: Analyzer analyzes
→ If not found: Report not found
```

**Example 3: Smart Escalation**
```
User: "Understand how authentication works"
→ Locator finds auth files
→ Analyzer attempts analysis
→ Escalates to security-scanner (future)
→ Results merged with context
```

**Example 4: Cached Results**
```
User: "Analyze utils.go"
→ Check cache
→ If cached: Return immediately
→ If not: Analyze and cache
```

---

## Questions to Resolve

1. **Intent Parser Sophistication**: Start with regex or integrate LLM?
   - **Recommendation**: Regex first, LLM in Phase 5

2. **Parallelism Model**: Goroutines or process pool?
   - **Recommendation**: Goroutines (simpler), process pool later

3. **Cache Backend**: In-memory or persistent?
   - **Recommendation**: In-memory with JSON backup

4. **Session Storage**: Memory, file, or database?
   - **Recommendation**: Memory + JSON file for persistence

5. **Workflow Language**: Code or declarative?
   - **Recommendation**: Code-based (Go structs)

---

## Next Steps (Immediate)

1. **Set up Phase 3 GitHub Project items**
2. **Create branch: `phase-3-development`**
3. **Start with Task 3.1: Intent Parsing**
4. **Review and refine plan as needed**

---

## Resources

### Documentation
- [PHASE2-STATUS.md](./PHASE2-STATUS.md) - What we built
- [POC-PLAN.md](./POC-PLAN.md) - Overall vision
- [orchestrator/README.md](./internal/orchestrator/README.md) - Current state

### Code References
- `internal/orchestrator/orchestrator.go` - Base to enhance
- `internal/agent/locator/` - Agent pattern
- `internal/agent/analyzer/` - Agent pattern

### External References
- Go Concurrency: https://go.dev/tour/concurrency
- Workflow Patterns: https://www.workflowpatterns.com/
- Intent Recognition: NLU best practices

---

**Ready to begin Phase 3 implementation!**

Use this plan as a guide throughout Phase 3 development. Update as decisions are made and requirements evolve.
