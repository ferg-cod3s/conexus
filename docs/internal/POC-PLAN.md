# Conexus POC Development Plan

**Project Status**: Planning Phase → POC Development
**GitHub Project**: [Conexus POC Development](https://github.com/users/ferg-cod3s/projects/3)
**Total Tasks**: 35+ items across 5 phases
**Last Updated**: 2025-10-13

> **📋 All implementation tasks, progress tracking, and detailed work items are managed in the GitHub Project linked above. This document serves as the comprehensive reference guide.**

---

## Overview

This document references the comprehensive Proof of Concept (POC) development plan for Conexus, a multi-agent AI system for autonomous software development. All detailed tasks, milestones, and progress tracking are maintained in the GitHub Project linked above.

## Research Foundation

The POC plan is based on comprehensive research conducted through the `/research-enhanced` workflow, which analyzed:

- **Codebase Structure**: Current documentation-only state with comprehensive architectural specifications
- **Existing Documentation**: README.md (770 lines), thoughts/ directory with 4 design documents, agent specifications
- **Architecture Patterns**: Multi-agent orchestration, AGENT_OUTPUT_V1 protocol, workflow coordination
- **Technical Requirements**: Evidence-grounded analysis, permission boundaries, escalation protocols

### Key Research Findings

1. **Current State**: Documentation-complete, implementation-absent project
2. **Architecture**: Multi-agent system with centralized orchestration
3. **Core Protocol**: AGENT_OUTPUT_V1 JSON schema for standardized agent outputs
4. **Agent Specialization**: 7+ specialized agents (locator, analyzer, pattern-finder, orchestrator, etc.)
5. **Technology Stack**: Language-agnostic design, recommendation: TypeScript/Node.js for AI ecosystem maturity

---

## Development Phases

> **📊 Detailed tasks, acceptance criteria, and progress tracking for all phases are managed in the [GitHub Project](https://github.com/users/ferg-cod3s/projects/3).**

### Phase 0: Project Initialization
**Focus**: Development environment and project scaffolding
- Initialize TypeScript project with bun
- Setup tooling (ESLint, Prettier, testing framework)
- Create directory structure
- Configure development scripts

### Phase 1: Core Infrastructure Setup
**Focus**: Foundational components and protocols

**Core Tasks** (managed in GitHub Project):
- 1.1: Define AGENT_OUTPUT_V1 TypeScript types
- 1.2: Implement tool execution framework
- 1.3: Build process-based agent isolation
- 1.4: Create agent communication protocol
- 1.5: Security & permission framework
- 1.6: Development environment setup

**Key Deliverables**:
- TypeScript type definitions for AGENT_OUTPUT_V1
- Permissioned tool executor (read, grep, glob, list)
- Agent process spawning/management
- JSON-RPC communication layer
- Security boundaries and audit logging

---

### Phase 2: Essential Agents Implementation
**Focus**: Core codebase analysis capabilities

Tasks managed in GitHub Project:
- 2.1: Implement codebase-locator agent
- 2.2: Implement codebase-analyzer agent
- 2.3: Implement basic orchestrator

**Key Deliverables**:
- Working codebase-locator (file/symbol search)
- Working codebase-analyzer (control flow, data flow analysis)
- Basic orchestrator with sequential workflows

---

### Phase 3: Orchestration & Workflow
**Focus**: Multi-agent coordination and state management

Tasks managed in GitHub Project:
- 3.1: Implement intent parsing logic
- 3.2: Build multi-agent workflow coordination
- 3.3: Implement escalation protocol
- 3.4: Create state management & caching

**Key Deliverables**:
- Intent parser for user request routing
- Sequential and parallel workflow execution
- Agent escalation mechanism
- Conversation state tracking and result caching

---

### Phase 4: Quality & Validation
**Focus**: Validation systems and testing

Tasks managed in GitHub Project:
- 4.1: Build evidence validation system
- 4.2: Implement output schema validation
- 4.3: Create integration testing framework
- 4.4: Add performance profiling

**Key Deliverables**:
- Evidence validator (100% claim backing requirement)
- JSON schema validator for AGENT_OUTPUT_V1
- End-to-end testing framework
- Performance monitoring and metrics

---

### Phase 5: Advanced Features
**Focus**: Enhanced capabilities and additional agents

Tasks managed in GitHub Project:
- 5.1: Implement parallel workflow execution
- 5.2: Add specialized agents
- 5.3: Build pattern recognition capabilities
- 5.4: Implement advanced caching strategies

**Key Deliverables**:
- Concurrent agent execution with result aggregation
- Additional specialized agents (pattern-finder, thoughts-analyzer, etc.)
- Cross-module pattern detection
- Content-hash based caching with intelligent invalidation

---

## Critical Implementation Decisions

### 1. Technology Stack
**Selected**: Go 1.23+
- **Superior performance** for multi-agent process management
- **Native concurrency** with goroutines for parallel agent execution
- **Strong type system** with excellent JSON handling
- **Built-in process spawning** and IPC capabilities
- **Fast compilation** and excellent tooling
- **Single binary deployment** with no runtime dependencies
- **Native testing** with built-in test framework

**Why Go Over TypeScript**:
- Better performance for process-intensive workloads
- Native concurrency primitives (goroutines/channels)
- No Node.js runtime overhead
- Simpler deployment model
- Built-in testing without external frameworks
- Excellent standard library for file I/O and process management

### 2. Agent Isolation Model
**Chosen**: Process-Based Isolation
- True isolation between agents
- OS-level permission enforcement
- Crash containment
- Enables parallel execution
- Higher overhead but better security and scalability

### 3. Communication Protocol
**Chosen**: JSON-RPC over stdio
- Standard protocol with good library support
- Simple request-response pattern
- Easy to debug and monitor
- Compatible with process-based isolation

### 4. State Management
**Chosen**: Immutable Context Propagation
- Append-only agent history
- No mutation of previous outputs
- Cache-friendly (deterministic outputs)
- Easier reasoning about workflow state

---

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│            User Interface Layer                 │
│  (CLI, API, IDE Extension)                      │
└───────────────┬─────────────────────────────────┘
                │
┌───────────────▼─────────────────────────────────┐
│          Orchestrator Service                   │
│  - Request routing                              │
│  - Workflow management                          │
│  - State coordination                           │
└───┬───────────────────────────────────────┬─────┘
    │                                       │
┌───▼──────────────┐              ┌────────▼──────┐
│  Agent Registry  │              │  Tool Executor│
│  - Agent specs   │              │  - Permissions│
│  - Capabilities  │              │  - Sandboxing │
└───┬──────────────┘              └────────┬──────┘
    │                                      │
┌───▼──────────────────────────────────────▼─────┐
│           Agent Process Pool                   │
│  ┌─────────────┐  ┌─────────────┐             │
│  │  Analyzer   │  │  Locator    │  ...        │
│  └─────────────┘  └─────────────┘             │
└────────────────────────────────────────────────┘
```

---

## AGENT_OUTPUT_V1 Protocol

All agents must produce outputs conforming to the AGENT_OUTPUT_V1 JSON schema. Key requirements:

### Core Structure
```json
{
  "version": "AGENT_OUTPUT_V1",
  "component_name": "string",
  "scope_description": "string",
  "overview": "string",
  "entry_points": [...],
  "call_graph": [...],
  "data_flow": {...},
  "state_management": [...],
  "side_effects": [...],
  "error_handling": [...],
  "configuration": [...],
  "patterns": [...],
  "concurrency": [...],
  "external_dependencies": [...],
  "limitations": [...],
  "open_questions": [...],
  "raw_evidence": [...]
}
```

### Critical Requirements
1. **100% Evidence Backing**: Every claim must have corresponding `raw_evidence` entry
2. **Absolute Paths**: All file references use absolute paths from repository root
3. **Line Number Precision**: Minimal, precise line ranges (avoid broad spans)
4. **No Nulls**: Omit unavailable sections or return empty arrays
5. **No Speculation**: Mark inferred information explicitly; use `limitations` for unknowns

---

## Testing Strategy

### Test Pyramid

**Level 1: Unit Tests** (Agent Logic)
- Individual agent behavior
- Tool execution
- Output schema validation
- Evidence validation

**Level 2: Integration Tests** (Agent Communication)
- Agent-to-agent communication
- Workflow coordination
- Escalation protocols
- State management

**Level 3: End-to-End Tests** (Full Workflows)
- Complete user request workflows
- Multi-agent coordination
- Real codebase analysis
- Performance benchmarking

---

## Success Criteria

### POC Completion Criteria

1. **Working Agents**: At minimum, codebase-locator and codebase-analyzer functional
2. **Valid Outputs**: All agent outputs conform to AGENT_OUTPUT_V1 schema
3. **Evidence Backing**: 100% of analysis claims backed by file:line evidence
4. **Sequential Workflows**: Orchestrator can coordinate 2+ agents in sequence
5. **Escalation**: Agents correctly escalate out-of-scope requests
6. **Test Coverage**: >80% coverage of core components
7. **Real-World Test**: Successfully analyze a small TypeScript codebase (e.g., Express.js "Hello World")

### Performance Targets (POC)

- Agent response time: <5 seconds for simple queries
- File analysis: <1 second per 1000 lines of code
- Symbol search: <2 seconds across 10,000 file repository
- Memory usage: <500MB per agent process

---

## Repository Structure (Implemented)

```
/home/f3rg/src/github/conexus/
├── README.md                      # Project overview (existing)
├── POC-PLAN.md                    # This file
├── LICENSE                        # MIT License (existing)
├── .gitignore                     # Git configuration
├── go.mod                         # Go module definition
├── go.sum                         # Go dependency checksums (generated)
├── cmd/
│   └── conexus/                   # Main application entry point
│       └── main.go                # CLI application
├── internal/                      # Private application code
│   ├── agent/                     # Agent implementations
│   │   ├── locator/               # codebase-locator agent
│   │   ├── analyzer/              # codebase-analyzer agent
│   │   └── orchestrator/          # orchestrator agent
│   ├── orchestrator/              # Orchestration logic
│   ├── protocol/                  # Communication protocols (JSON-RPC)
│   ├── tool/                      # Tool execution framework
│   └── process/                   # Process management
├── pkg/                           # Public API packages
│   ├── schema/                    # AGENT_OUTPUT_V1 schema
│   └── types/                     # Shared type definitions
├── tests/                         # Test files (*_test.go)
├── bin/                           # Compiled binaries (gitignored)
├── docs/                          # Documentation
│   ├── agents/                    # Agent specifications (existing)
│   ├── architecture/              # Architecture docs (existing)
│   └── protocols/                 # Protocol specs (existing)
└── examples/                      # Example codebases for testing
```

**Go Project Structure Conventions**:
- `cmd/`: Main applications (one subdirectory per binary)
- `internal/`: Private application code (cannot be imported by other projects)
- `pkg/`: Public libraries (can be imported by external projects)
- Tests live alongside code as `*_test.go` files
- No separate `tests/` hierarchy needed with Go conventions

---

## Task Management

### GitHub Project Organization

The [Conexus POC Development Project](https://github.com/users/ferg-cod3s/projects/3) contains:

#### 📋 **Reference Tasks**
- POC Plan Documentation (this document)
- Research Integration Tasks
- POC Success Demo
- POC Metrics & Evaluation

#### 🏗️ **Implementation Phases** (35+ tasks total)
- **Phase 0**: Project Initialization (1 task)
- **Phase 1**: Core Infrastructure Setup (6 tasks + completion criteria)
- **Phase 2**: Essential Agents Implementation (3 tasks + completion criteria)
- **Phase 3**: Orchestration & Workflow (4 tasks + completion criteria)
- **Phase 4**: Quality & Validation (4 tasks)
- **Phase 5**: Advanced Features (4 tasks)

#### ✅ **Completion Criteria**
- Phase-specific validation criteria
- Success metrics and testing requirements
- Documentation and quality gates

### Next Steps

### Immediate Actions (Start with Phase 0)

1. **✅ Phase 0: Project Initialization COMPLETE** (GitHub Project Task)
   ```bash
   cd /home/f3rg/src/github/conexus
   go mod init github.com/ferg-cod3s/conexus
   mkdir -p cmd/conexus internal/{agent,orchestrator,protocol,tool,process} pkg/{schema,types}
   go build -o bin/conexus ./cmd/conexus
   ```

2. **Follow GitHub Project Task Sequence**
   - Complete Phase 0 initialization
   - Move to Phase 1 core infrastructure
   - Track progress in GitHub Project

3. **Use Completion Criteria**
   - Validate each phase against defined criteria
   - Run tests and performance benchmarks
   - Update documentation

4. **Maintain Project Documentation**
   - Update GitHub Project as tasks are completed
   - Document decisions in `thoughts/` directory
   - Add implementation examples to `docs/`

---

## Resources

### Documentation
- **Project README**: `/home/f3rg/src/github/conexus/README.md`
- **Agent Specifications**: `/home/f3rg/src/github/conexus/docs/agents/`
- **Architecture Docs**: `/home/f3rg/src/github/conexus/docs/architecture/`
- **Protocols**: `/home/f3rg/src/github/conexus/docs/protocols/`
- **Design Thoughts**: `/home/f3rg/src/github/conexus/thoughts/`

### External References
- Go Documentation: https://go.dev/doc/
- Go Standard Library: https://pkg.go.dev/std
- JSON Schema in Go: https://github.com/santhosh-tekuri/jsonschema
- Go Anthropic SDK: https://github.com/anthropics/anthropic-sdk-go
- Go Process Management: https://pkg.go.dev/os/exec

### Related Projects
- Claude Code: https://docs.claude.com/claude-code
- OpenAI Swarm: https://github.com/openai/swarm (multi-agent patterns)
- LangGraph: https://github.com/langchain-ai/langgraph (agent orchestration)

---

## Contact & Contributions

**Project Owner**: ferg-cod3s
**Repository**: https://github.com/ferg-cod3s/conexus
**GitHub Project**: https://github.com/users/ferg-cod3s/projects/3

For questions, issues, or contributions, please:
1. Check existing GitHub Issues
2. Review documentation in `/docs`
3. Consult design thoughts in `/thoughts`
4. Create new issue with detailed description

---

## Appendix: Research Summary

### Research Methodology

The POC plan was developed using the `/research-enhanced` command, which orchestrated multiple specialized research agents:

1. **codebase-locator**: Mapped project structure and identified existing files
2. **thoughts-locator**: Discovered and categorized all documentation
3. **codebase-pattern-finder**: Analyzed documented patterns (no implementation exists yet)
4. **codebase-analyzer**: Extracted technical specifications from documentation
5. **thoughts-analyzer**: Synthesized insights from design documents

### Key Insights from Research

**From README.md Analysis**:
- 770-line comprehensive architecture specification
- Multi-agent system with 7+ specialized agent types
- AGENT_OUTPUT_V1 protocol as core communication standard
- Evidence-grounded analysis as fundamental principle
- Process-based isolation for security and scalability

**From thoughts/ Analysis**:
- Date-stamped design documents (all from 2025-01-09)
- Focus on iterative refinement and self-documenting workflows
- Context-aware conversation threading concepts
- Project foundation decisions documented

**From Pattern Analysis**:
- Modular agent design with standard interfaces
- Event-driven architecture for agent communication
- Declarative workflow definitions
- Configuration-over-convention approach
- Fail-safe design with multiple error handling layers

### Research Confidence Scores

- Documentation Quality: 1.0 (excellent)
- Implementation Readiness: 0.0 (non-existent)
- Architectural Clarity: 1.0 (very clear)
- Feasibility: 0.8 (ambitious but achievable)

---

**Status**: Ready to begin Phase 1 implementation. All tasks tracked in GitHub Project linked above.
