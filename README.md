# Conexus - The Agentic Context Engine

**Version**: 0.0.5 (Phase 5 - Integration & Documentation)  
**Status**: 🚧 Active Development  
**Go Version**: 1.23.4

[![Go Tests](https://img.shields.io/badge/tests-passing-brightgreen)]()
[![Coverage](https://img.shields.io/badge/coverage-85%25-green)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

---

## 🎯 Overview

Conexus is an **agentic context engine** that transforms Large Language Models (LLMs) into expert engineering assistants. It provides a **multi-agent system** for analyzing codebases, with built-in validation, profiling, and workflow orchestration.

### Key Features

- 🤖 **Multi-Agent Architecture**: Specialized agents for locating and analyzing code
- ✅ **Evidence-Backed Validation**: 100% evidence traceability for all agent outputs
- 📊 **Performance Profiling**: Real-time metrics and bottleneck detection
- 🔄 **Workflow Orchestration**: Complex multi-agent workflows with state management
- 🏗️ **AGENT_OUTPUT_V1**: Standardized JSON schema for agent communication
- 🧪 **Comprehensive Testing**: 53+ integration tests with real-world validation

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.23.4+** ([download](https://go.dev/dl/))
- Git
- Linux/macOS/Windows with WSL

### Installation

```bash
# Clone the repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Install dependencies
go mod download

# Build the project
go build ./cmd/conexus

# Run tests
go test ./...
```

### Basic Usage

```bash
# Run the Conexus agent (development)
./conexus

# Run with verbose logging
./conexus -v

# Run specific agent
./conexus agent locator --pattern "func.*Handler"
```

---

## 📚 Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────┐
│                   Orchestrator                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │   Intent    │  │  Workflow   │  │    State    │ │
│  │   Parser    │  │   Engine    │  │  Manager    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
  ┌─────▼─────┐   ┌────▼─────┐   ┌────▼─────┐
  │  Locator  │   │ Analyzer │   │  Future  │
  │   Agent   │   │  Agent   │   │  Agents  │
  └─────┬─────┘   └────┬─────┘   └──────────┘
        │               │
        └───────┬───────┘
                │
  ┌─────────────▼─────────────┐
  │      Validation Layer     │
  │  ┌──────────┐ ┌─────────┐ │
  │  │ Evidence │ │ Schema  │ │
  │  │Validator │ │Validator│ │
  │  └──────────┘ └─────────┘ │
  └───────────────────────────┘
                │
  ┌─────────────▼─────────────┐
  │     Profiling Layer       │
  │  ┌──────────┐ ┌─────────┐ │
  │  │Collector │ │Reporter │ │
  │  └──────────┘ └─────────┘ │
  └───────────────────────────┘
```

### Core Components

| Component | Description | Status |
|-----------|-------------|--------|
| **Orchestrator** | Workflow engine, intent parsing, state management | ✅ Complete |
| **Locator Agent** | Find files/functions matching patterns | ✅ Complete |
| **Analyzer Agent** | Extract control flow and data dependencies | ✅ Complete |
| **Evidence Validator** | Verify 100% evidence backing | ✅ Complete |
| **Schema Validator** | Validate AGENT_OUTPUT_V1 format | ✅ Complete |
| **Profiler** | Performance metrics and reporting | ✅ Complete |
| **Integration Framework** | End-to-end testing harness | ✅ Complete |

---

## 🧪 Testing

### Test Suite Overview

Conexus has **53 integration tests** covering real-world scenarios:

```bash
# Run all tests
go test ./...

# Run integration tests only
go test ./internal/testing/integration

# Run with verbose output
go test -v ./internal/testing/integration

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestLocatorAnalyzerIntegration ./internal/testing/integration
```

### Test Categories

| Category | Tests | Coverage |
|----------|-------|----------|
| **Framework Tests** | 13 | Core test infrastructure |
| **Duration Tests** | 7 | Performance regression detection |
| **E2E Fixture Tests** | 4 | Workflow execution with test fixtures |
| **Advanced Workflows** | 7 | Complex multi-step scenarios |
| **Coordination Tests** | 5 | Multi-agent communication |
| **Real-World Tests** | 5 | Actual Conexus source code analysis |

### Performance Benchmarks

- **Full Test Suite**: <1 second
- **Single Agent Execution**: <50ms
- **Multi-Agent Workflow**: <100ms
- **Real Codebase Analysis**: <100ms per file

---

## 📖 Documentation

### User Guides

- **[Validation Guide](docs/validation-guide.md)** - Evidence and schema validation
- **[Profiling Guide](docs/profiling-guide.md)** - Performance monitoring and optimization
- **[API Reference](docs/api-reference.md)** - Complete API documentation

### Architecture Documentation

- **[Technical Architecture](docs/Technical-Architecture.md)** - System design overview
- **[Integration Architecture](docs/architecture/integration.md)** - Component integration
- **[Context Engine Internals](docs/architecture/context-engine-internals.md)** - Core algorithms
- **[Data Architecture](docs/architecture/data-architecture.md)** - Data flow and storage

### Development Resources

- **[Developer Onboarding](docs/getting-started/developer-onboarding.md)** - Getting started guide
- **[Contributing Guide](docs/contributing/contributing-guide.md)** - Contribution guidelines
- **[Testing Strategy](docs/contributing/testing-strategy.md)** - Testing best practices
- **[Operations Guide](docs/operations/operations-guide.md)** - Deployment and operations

---

## 🔧 Configuration

### Agent Configuration

Conexus agents use environment variables for configuration:

```bash
# Enable verbose logging
export CONEXUS_LOG_LEVEL=debug

# Set profiling interval (ms)
export CONEXUS_PROFILE_INTERVAL=100

# Enable evidence validation
export CONEXUS_VALIDATE_EVIDENCE=true

# Set cache directory
export CONEXUS_CACHE_DIR=~/.cache/conexus
```

### Validation Configuration

```bash
# Require 100% evidence backing (default: true)
export CONEXUS_REQUIRE_FULL_EVIDENCE=true

# Schema validation mode (strict|lenient)
export CONEXUS_SCHEMA_MODE=strict

# Max validation errors before failing
export CONEXUS_MAX_VALIDATION_ERRORS=10
```

---

## 🎯 AGENT_OUTPUT_V1 Schema

All agents produce standardized output following the **AGENT_OUTPUT_V1** schema:

```json
{
  "schema_version": "AGENT_OUTPUT_V1",
  "task_description": "Locate all HTTP handler functions",
  "result_summary": "Found 5 HTTP handlers in 3 files",
  "confidence_score": 0.95,
  "items": [
    {
      "type": "function",
      "name": "HandleRequest",
      "file_path": "/internal/server/handler.go",
      "line_start": 42,
      "line_end": 68,
      "evidence_file_path": "/internal/server/handler.go",
      "evidence_line_start": 42,
      "evidence_line_end": 68,
      "classification": "primary",
      "explanation": "HTTP handler implementing request processing logic"
    }
  ],
  "files_examined": ["/internal/server/handler.go"],
  "metadata": {
    "agent_name": "locator",
    "execution_time_ms": 45,
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

**Key Requirements**:
- ✅ **100% Evidence Backing**: Every item must have valid file/line references
- ✅ **Schema Compliance**: All required fields must be present
- ✅ **Confidence Score**: Between 0.0 and 1.0
- ✅ **Structured Items**: Typed items with classification

See **[API Reference](docs/api-reference.md)** for complete schema documentation.

---

## 🏗️ Development Workflow

### Project Structure

```
conexus/
├── cmd/conexus/          # Main entry point
├── internal/
│   ├── agent/           # Agent implementations
│   │   ├── locator/     # File/function locator
│   │   └── analyzer/    # Code analyzer
│   ├── orchestrator/    # Workflow orchestration
│   │   ├── intent/      # Intent parsing
│   │   ├── workflow/    # Workflow engine
│   │   ├── state/       # State management
│   │   └── escalation/  # Error handling
│   ├── validation/      # Validation systems
│   │   ├── evidence/    # Evidence validation
│   │   └── schema/      # Schema validation
│   ├── profiling/       # Performance profiling
│   ├── protocol/        # JSON-RPC protocol
│   └── testing/         # Integration testing
├── pkg/schema/          # Public schemas
├── tests/fixtures/      # Test fixtures
└── docs/                # Documentation
```

### Adding a New Agent

1. Create agent directory: `internal/agent/myagent/`
2. Implement agent interface:
   ```go
   type Agent interface {
       Execute(ctx context.Context, req Request) (*schema.AgentOutput, error)
   }
   ```
3. Add tests in `myagent_test.go`
4. Register in orchestrator
5. Add integration tests

See **[Contributing Guide](docs/contributing/contributing-guide.md)** for details.

---

## 📊 Current Status

### Phase 5 Progress (95% Complete)

- ✅ **Task 5.1**: Integration Testing Framework (53 tests passing)
- 🔄 **Task 5.2**: Documentation Updates (in progress)
- ⏳ **Task 5.3**: Workflow Integration (pending)
- ⏳ **Task 5.4**: Protocol Tests (optional)

### Test Results

```
✅ All 53 integration tests passing
✅ Execution time: <1 second
✅ Evidence validation: 100%
✅ Schema compliance: 100%
✅ Real-world analysis: 5 scenarios validated
```

See **[PHASE5-STATUS.md](PHASE5-STATUS.md)** for detailed status.

---

## 🛣️ Roadmap

### Phase 6: Optimization (Planned)

- ⏳ Advanced caching strategies
- ⏳ Parallel agent execution
- ⏳ Performance optimization
- ⏳ Memory usage reduction

### Phase 7: Production Readiness (Planned)

- ⏳ CLI enhancements
- ⏳ Configuration management
- ⏳ Deployment automation
- ⏳ Monitoring dashboards

### Future Agents (Planned)

- ⏳ Pattern recognition agent
- ⏳ Thoughts analyzer agent
- ⏳ Dependency analyzer agent
- ⏳ Security audit agent

---

## 🤝 Contributing

We welcome contributions! Please see:

- **[Contributing Guide](docs/contributing/contributing-guide.md)** - How to contribute
- **[Testing Strategy](docs/contributing/testing-strategy.md)** - Testing requirements
- **[Code Style](CLAUDE.md)** - Coding conventions

### Quick Contribution Checklist

- [ ] Fork the repository
- [ ] Create a feature branch
- [ ] Write tests for new features
- [ ] Ensure all tests pass (`go test ./...`)
- [ ] Follow code style guidelines
- [ ] Update documentation
- [ ] Submit a pull request

---

## 📄 License

This project is licensed under the **MIT License** - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- **[Anthropic](https://anthropic.com)** - MCP protocol and Claude integration
- **[Model Context Protocol](https://modelcontextprotocol.io)** - Standards-based integration
- Go community for excellent tooling

---

## 📞 Support & Contact

- **Issues**: [GitHub Issues](https://github.com/ferg-cod3s/conexus/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ferg-cod3s/conexus/discussions)
- **Email**: support@conexus.dev (coming soon)

---

## 🔗 Related Projects

- **[MCP](https://modelcontextprotocol.io)** - Model Context Protocol specification
- **[Claude Code](https://claude.ai)** - AI-powered development assistant
- **[OpenCode](https://github.com/opencode-ai)** - Open-source AI coding tools

---

**Built with ❤️ by the Conexus team**
