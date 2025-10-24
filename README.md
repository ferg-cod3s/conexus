# Conexus - The Agentic Context Engine

**Version**: 0.1.0-mvp (Phase 7 - Production Ready)  
**Status**: ✅ Production Ready  
**Go Version**: 1.23.4

[![Go Tests](https://img.shields.io/badge/tests-passing-brightgreen)]()
[![Coverage](https://img.shields.io/badge/coverage-85%25-green)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

---

## 🎯 Overview

Conexus is an **agentic context engine** that transforms Large Language Models (LLMs) into expert engineering assistants. It provides a **multi-agent system** for analyzing codebases, with built-in validation, profiling, and workflow orchestration.

### Key Features

- 🤖 **Multi-Agent Architecture**: Specialized agents for locating and analyzing code
- 🔌 **MCP Integration**: First-class Model Context Protocol support for AI assistants
- ✅ **Evidence-Backed Validation**: 100% evidence traceability for all agent outputs
- 📊 **Performance Profiling**: Real-time metrics and bottleneck detection
- 🔄 **Workflow Orchestration**: Complex multi-agent workflows with state management
- 🏗️ **AGENT_OUTPUT_V1**: Standardized JSON schema for agent communication
- 🧪 **Comprehensive Testing**: 53+ integration tests with real-world validation

---

## 🚀 Quick Start

### Prerequisites

- **Node.js 18+** or **Bun** (for npm/bunx installation)
- Git

### Installation

**Option 1: NPM/Bunx (Recommended - Pre-built Binaries)**

```bash
# Install globally with npm
npm install -g @agentic-conexus/mcp

# Or use with bunx (no installation needed)
bunx @agentic-conexus/mcp

# Or use with npx
npx @agentic-conexus/mcp
```

> **Note**: Pre-built binaries are included for:
> - macOS (Intel & Apple Silicon)
> - Linux (amd64 & arm64)
> - Windows (amd64)

**Option 2: From Source (For Development)**

```bash
# Clone the repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Requires Go 1.23.4+ - https://go.dev/dl/

# Build from source
go build -o conexus ./cmd/conexus

# Run tests
go test ./...
```

### Basic Usage

```bash
# Run the MCP server (stdio mode - default)
npx @agentic-conexus/mcp

# Run with environment variables
CONEXUS_DB_PATH=./data/db.sqlite CONEXUS_LOG_LEVEL=debug npx @agentic-conexus/mcp

# Run in HTTP mode (for testing)
CONEXUS_PORT=3000 npx @agentic-conexus/mcp
```

---

---

## 🔌 MCP Integration

Conexus provides first-class support for the [Model Context Protocol (MCP)](https://modelcontextprotocol.io), enabling seamless integration with AI assistants like Claude Desktop and Cursor.

### Why Use Conexus with AI Assistants?

Conexus provides **measurable context retention improvements** over standard LLM interactions:

#### 🔄 **Persistent Context Management**
- **Conversation History**: Full multi-turn conversation tracking [Source: internal/orchestrator/state/manager.go:42-56]
- **Session Persistence**: State preservation across interactions [Source: internal/orchestrator/state/README.md]
- **Context Accumulation**: Build context over multiple agent interactions [Source: internal/orchestrator/orchestrator.go:122-124]

#### 🔍 **Intelligent Context Retrieval**
- **Hybrid Search**: Vector similarity + BM25 keyword search [Source: internal/search/search.go:95-110]
- **Multi-Level Caching**: 98% cache hit rate reducing redundant computations [Source: docs/architecture/context-engine-internals.md:9870-10127]
- **Context-Aware Ranking**: Freshness, authority, and diversity scoring [Source: docs/architecture/context-engine-internals.md:580-617]

#### 📊 **Performance Advantages**
- **26x Faster Context Retrieval**: 1.5ms with caching vs 40ms full retrieval [Source: PERFORMANCE_BASELINE.md]
- **85-92% Recall**: At 20 results vs manual context discovery [Source: PERFORMANCE_BASELINE.md]
- **Sub-Second Assembly**: For typical codebases vs minutes of manual searching [Source: PERFORMANCE_BASELINE.md]

#### 🛠️ **Built-in Tools**
- **6 Powerful MCP Tools**: For comprehensive code understanding [Source: internal/mcp/README.md]
- **Evidence-Backed Results**: 100% traceability for all findings [Source: internal/validation/evidence/]
- **Real-time Indexing**: Keep context fresh with incremental updates [Source: PERFORMANCE_BASELINE.md]

### Quick MCP Setup (<5 minutes)

**Option 1: NPM/Bunx (Recommended for MCP clients)**

```bash
# Install globally with npm
npm install -g @ferg-cod3s/conexus

# Or use with bunx (no installation needed)
bunx @ferg-cod3s/conexus
```

Configure in your MCP client (OpenCode, Claude Desktop, etc.):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["@ferg-cod3s/conexus"],
      "env": {
        "CONEXUS_DB_PATH": "/path/to/your/project/.conexus/db.sqlite"
      }
    }
  }
}
```

**Option 2: Go Install (For development)**

```bash
# Install Conexus
go install github.com/ferg-cod3s/conexus/cmd/conexus@latest

# Start the MCP server (stdio mode by default)
conexus

# Or run in HTTP mode
CONEXUS_PORT=3000 conexus
```

Configure for stdio mode (recommended for MCP):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "conexus",
      "env": {
        "CONEXUS_DB_PATH": "/path/to/your/project/.conexus/db.sqlite"
      }
    }
  }
}
```

**Test the integration:**

In your MCP client (OpenCode, Claude Desktop, etc.):

```
You: "Search for HTTP handler functions in this codebase"

AI Assistant: [Uses context.search tool]
Found 5 HTTP handlers:
- HandleRequest in internal/server/handler.go:42-68
- HandleHealth in internal/server/health.go:15-22
...
```

**Environment Variables:**

- `CONEXUS_DB_PATH`: Path to SQLite database (default: `~/.conexus/db.sqlite`)
- `CONEXUS_LOG_LEVEL`: Log level: debug, info, warn, error (default: `info`)
- `CONEXUS_PORT`: Run in HTTP mode instead of stdio (for development)

### Available MCP Tools

| Tool | Status | Description |
|------|--------|-------------|
| `context.search` | ✅ Fully Implemented | Semantic search with hybrid vector+BM25, work context boosting, and semantic reranking |
| `context.get_related_info` | ✅ Fully Implemented | Get related files, functions, and context for specific files or tickets |
| `context.explain` | ✅ Fully Implemented | Detailed code explanations with examples and complexity assessment |
| `context.grep` | ✅ Fully Implemented | Fast pattern matching using ripgrep with regex support |
| `context.index_control` | ✅ Fully Implemented | Full indexing operations (start, stop, status, reindex, sync) |
| `context.connector_management` | ✅ Fully Implemented | Complete CRUD operations for data source connectors with SQLite persistence |

### Example Queries

**Code Understanding:**
```
"Show me all database query functions"
"Find the authentication middleware implementation"
"What functions handle user registration?"
```

**Bug Investigation:**
```
"Search for error handling in the payment module"
"Find all functions that access the user database"
"Show panic or fatal calls in the codebase"
```

**Feature Development:**
```
"Locate API endpoint handlers"
"Find all struct definitions related to orders"
"Search for configuration loading functions"
```

### Project-Specific Installation

For using Conexus with specific projects, you can configure it to work with your existing codebase structure:

#### 1. Per-Project MCP Server Configuration

Create a project-specific MCP configuration:

```json
{
  "mcpServers": {
    "conexus-myproject": {
      "command": "conexus",
      "args": ["mcp", "--root", "/path/to/your/project"],
      "env": {
        "CONEXUS_LOG_LEVEL": "info",
        "CONEXUS_CONFIG": "/path/to/your/project/conexus.yml"
      }
    }
  }
}
```

#### 2. Project Configuration File

Create a `conexus.yml` file in your project root:

```yaml
# conexus.yml - Project-specific configuration
project:
  name: "my-project"
  description: "Web application backend"

# Codebase settings
codebase:
  root: "."
  include_patterns:
    - "**/*.go"
    - "**/*.js"
    - "**/*.ts"
    - "**/*.py"
  exclude_patterns:
    - "**/node_modules/**"
    - "**/vendor/**"
    - "**/dist/**"
    - "**/.git/**"

# Search configuration
search:
  max_results: 50
  similarity_threshold: 0.7
  enable_fts: true

# Indexing settings
indexing:
  auto_reindex: true
  reindex_interval: "1h"
  chunk_size: 1000
```

#### 3. Docker Integration for Teams

For team environments, use Docker to ensure consistent configuration:

```yaml
# docker-compose.conexus.yml
version: '3.8'
services:
  conexus:
    image: conexus:latest
    container_name: conexus-myproject
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - ./:/workspace:ro
      - ./data:/data
    environment:
      - CONEXUS_ROOT_PATH=/workspace
      - CONEXUS_LOG_LEVEL=info
      - CONEXUS_CONFIG=/workspace/conexus.yml
    working_dir: /workspace
```

```bash
# Start for your project
docker-compose -f docker-compose.conexus.yml up -d

# Test the connection
curl http://localhost:3000/health
```

#### 4. Project Type Examples

**Node.js Project:**
```yaml
codebase:
  include_patterns:
    - "**/*.js"
    - "**/*.ts"
    - "**/*.json"
    - "**/*.md"
  exclude_patterns:
    - "**/node_modules/**"
    - "**/coverage/**"
    - "**/dist/**"
```

**Python Project:**
```yaml
codebase:
  include_patterns:
    - "**/*.py"
    - "**/*.md"
    - "**/requirements*.txt"
    - "**/pyproject.toml"
  exclude_patterns:
    - "**/__pycache__/**"
    - "**/venv/**"
    - "**/env/**"
    - "**/.pytest_cache/**"
```

**Go Project:**
```yaml
codebase:
  include_patterns:
    - "**/*.go"
    - "**/go.mod"
    - "**/go.sum"
    - "**/*.md"
  exclude_patterns:
    - "**/vendor/**"
```

**Monorepo:**
```yaml
codebase:
  include_patterns:
    - "packages/**/*.ts"
    - "packages/**/*.js"
    - "apps/**/*.ts"
    - "apps/**/*.js"
  exclude_patterns:
    - "**/node_modules/**"
    - "**/dist/**"
    - "**/build/**"
```

#### 5. Claude Desktop Project Templates

Create reusable templates for different project types:

```json
{
  "mcpServers": {
    "conexus-nodejs": {
      "command": "conexus",
      "args": ["mcp", "--root", "$PROJECT_ROOT"],
      "env": {
        "CONEXUS_CONFIG": "$PROJECT_ROOT/.conexus/nodejs.yml"
      }
    },
    "conexus-python": {
      "command": "conexus", 
      "args": ["mcp", "--root", "$PROJECT_ROOT"],
      "env": {
        "CONEXUS_CONFIG": "$PROJECT_ROOT/.conexus/python.yml"
      }
    }
  }
}
```

### Advanced Configuration

For production deployments, custom embedding providers, and advanced search optimization, see the **[MCP Integration Guide](docs/getting-started/mcp-integration-guide.md)**.

**Topics covered:**
- Custom embedding providers (OpenAI, Anthropic, Ollama, Cohere)
- Vector store backends (SQLite, PostgreSQL, memory)
- Search optimization strategies
- Security configuration (RBAC, API keys, audit logging)
- Troubleshooting common issues
- Multiple instance support (monorepos)



---

## 📈 Context Retention vs Standard LLM

Conexus provides **significant improvements** over standard LLM context limitations:

### Standard LLM Limitations
- ❌ **Fixed Context Window**: Typically 8K-32K tokens
- ❌ **No Persistent Memory**: Each interaction starts fresh
- ❌ **Manual Context Gathering**: User must find and provide relevant code
- ❌ **No Codebase-Specific Knowledge**: Generic training data only

### Conexus Improvements
- ✅ **Unlimited Context**: Through intelligent retrieval and assembly
- ✅ **Persistent Sessions**: Full conversation history and state management
- ✅ **Automated Context Discovery**: Hybrid search finds relevant code automatically
- ✅ **Codebase-Specific Intelligence**: Indexed knowledge of your actual code

### Measurable Impact

| Metric | Standard LLM | Conexus | Improvement |
|--------|---------------|---------|-------------|
| **Context Window** | 8K-32K tokens | Unlimited | ∞ |
| **Session Memory** | None | Persistent | +100% |
| **Context Retrieval** | Manual search | 11ms automated | 26x faster |
| **Code Discovery** | User-dependent | 85-92% recall | Significantly higher |
| **Memory Efficiency** | Load entire codebase | 58MB for 10K files | 42% under target |

### Real-World Benefits

**For Developers:**
- **Faster Onboarding**: New team members get instant codebase context
- **Reduced Context Switching**: AI maintains conversation state across complex tasks
- **Better Code Reviews**: Automated evidence backing ensures accurate analysis

**For Teams:**
- **Consistent Understanding**: Shared context across all team members
- **Knowledge Preservation**: Critical insights retained in conversation history
- **Scalable Expertise**: AI assistant learns your specific codebase patterns

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

All performance metrics are sourced from comprehensive benchmarks documented in [`PERFORMANCE_BASELINE.md`](PERFORMANCE_BASELINE.md):

#### Context Retrieval Performance
- **Search Latency**: 10.35ms routing + 0.81ms BM25 search = **~11ms total** [Source: PERFORMANCE_BASELINE.md]
- **Cache Hit Rate**: 98% (85% L1 + 10% L2 + 3% L3) [Source: docs/architecture/context-engine-internals.md]
- **Vector Search**: 248ms for 1K documents, 2.18s for 10K documents [Source: PERFORMANCE_BASELINE.md]
- **Hybrid Search**: 1.96s for 10K documents (vector + BM25 fusion) [Source: PERFORMANCE_BASELINE.md]

#### System Performance
- **Single Agent Execution**: <50ms (framework overhead: 67μs) [Source: PERFORMANCE_BASELINE.md]
- **Multi-Agent Workflow**: <100ms (orchestrator overhead: 10.35ms) [Source: PERFORMANCE_BASELINE.md]
- **Real Codebase Analysis**: <100ms per file (indexing: 450 files/sec) [Source: PERFORMANCE_BASELINE.md]

#### Indexing Performance
- **File Walking**: 65,000 files/sec (65x target) [Source: PERFORMANCE_BASELINE.md]
- **Chunking**: 45,000-79,000 files/sec (450-790x target) [Source: PERFORMANCE_BASELINE.md]
- **Full Index with Embeddings**: 450 files/sec (4.5x target) [Source: PERFORMANCE_BASELINE.md]
- **Memory Usage**: 58MB for 10K files (42% under 100MB target) [Source: PERFORMANCE_BASELINE.md]

#### Local Performance Under Load
- **Sustained Processing**: 149 req/s during intensive analysis [Source: docs/operations/production-readiness-checklist.md]
- **P95 Response Time**: 612ms during heavy workloads [Source: tests/load/README.md]
- **P99 Response Time**: 989ms during peak processing [Source: tests/load/README.md]
- **Memory Efficiency**: 58MB for 10K files (42% under 100MB target) [Source: PERFORMANCE_BASELINE.md]

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

## 🔄 Workflow Integration

### Overview

Conexus provides a powerful workflow integration system that combines validation, profiling, and quality gates into coordinated multi-agent workflows.

### Basic Orchestrator Usage

```go
package main

import (
    "context"
    "github.com/ferg-cod3s/conexus/internal/orchestrator"
    "github.com/ferg-cod3s/conexus/internal/process"
    "github.com/ferg-cod3s/conexus/internal/tool"
    "github.com/ferg-cod3s/conexus/internal/validation/evidence"
)

func main() {
    // Create orchestrator with default configuration
    config := orchestrator.OrchestratorConfig{
        ProcessManager:    process.NewManager(),
        ToolExecutor:      tool.NewExecutor(),
        EvidenceValidator: evidence.NewValidator(false), // false = non-strict mode
        QualityGates:      orchestrator.DefaultQualityGates(),
        EnableProfiling:   true,
    }
    orch := orchestrator.NewWithConfig(config)
    
    // Execute a workflow
    ctx := context.Background()
    result, err := orch.HandleRequest(ctx, "find all HTTP handlers", permissions)
    if err != nil {
        log.Fatal(err)
    }
    
    // Access results with profiling data
    fmt.Printf("Completed in %v\n", result.Duration)
    fmt.Printf("Evidence coverage: %.1f%%\n", result.Profile.EvidenceCoverage)
}
```

### Quality Gate Presets

Conexus provides three quality gate configurations:

#### 1. Default Quality Gates (Balanced)
```go
config := orchestrator.OrchestratorConfig{
    QualityGates: orchestrator.DefaultQualityGates(),
}
```
- ✅ 100% evidence backing required
- ✅ 5-minute max workflow time
- ✅ 1-minute max agent execution time
- ✅ Blocks on validation failures

#### 2. Relaxed Quality Gates (Development)
```go
config := orchestrator.OrchestratorConfig{
    QualityGates: orchestrator.RelaxedQualityGates(),
}
```
- ⚠️ 80% evidence coverage minimum
- ⚠️ 10-minute max workflow time
- ⚠️ Allows up to 5 unbacked claims

#### 3. Strict Quality Gates (Production)
```go
config := orchestrator.OrchestratorConfig{
    QualityGates: orchestrator.StrictQualityGates(),
}
```
- 🔒 100% evidence backing enforced
- 🔒 2-minute max workflow time
- 🔒 30-second max agent execution time
- 🔒 Blocks on all failures (validation + performance)

### Custom Quality Gates

```go
config := orchestrator.OrchestratorConfig{
    QualityGates: &orchestrator.QualityGateConfig{
        RequireEvidenceBacking:    true,
        MinEvidenceCoverage:       95.0,
        AllowUnbackedClaims:       2,
        MaxExecutionTime:          3 * time.Minute,
        MaxAgentExecutionTime:     30 * time.Second,
        BlockOnValidationFailure:  true,
        BlockOnPerformanceFailure: false,
    },
}
```

### Profiling Integration

Enable automatic profiling to capture performance metrics:

```go
config := orchestrator.OrchestratorConfig{
    EnableProfiling: true,
}

result, _ := orch.ExecuteWorkflow(ctx, workflow, permissions)

// Access profiling data
profile := result.Profile
fmt.Printf("Total duration: %v\n", profile.TotalDuration)
fmt.Printf("Agent time: %v\n", profile.AgentExecutionTime)
fmt.Printf("Validation time: %v\n", profile.ValidationTime)
fmt.Printf("Profiling overhead: %.2f%%\n", profile.ProfilingOverheadPercent)
```

### Validation Integration

Evidence validation is automatically integrated:

```go
// Strict mode - requires 100% evidence backing
validator := evidence.NewValidator(true)

// Non-strict mode - allows partial evidence
validator := evidence.NewValidator(false)

config := orchestrator.OrchestratorConfig{
    EvidenceValidator: validator,
}
```

### Workflow Reports

Generate comprehensive workflow reports:

```go
result, _ := orch.ExecuteWorkflow(ctx, workflow, permissions)

// Generate workflow report
report := orchestrator.GenerateWorkflowReport(result)

fmt.Println(report.ExecutionSummary)
fmt.Println(report.ValidationReport)
fmt.Println(report.PerformanceReport)
```

**Example report output:**
```
=== Workflow Execution Report ===

Execution Summary:
  Duration: 127ms
  Agents Executed: 2
  Status: ✅ Success

Validation Report:
  Evidence Coverage: 100.0%
  Backed Claims: 15
  Unbacked Claims: 0
  Status: ✅ Passed

Performance Report:
  Agent Execution: 85ms (66.9%)
  Validation: 12ms (9.4%)
  Profiling Overhead: 1.2%
  Status: ✅ Within Limits
```

### Best Practices

1. **Use Default Gates for Most Cases**: Balanced performance and quality
2. **Enable Profiling in Development**: Identify bottlenecks early
3. **Strict Mode for Production**: Maximum confidence in production workflows
4. **Monitor Profiling Overhead**: Keep under 10% for production systems
5. **Review Validation Reports**: Ensure evidence backing meets standards

See **[Testing Strategy](docs/contributing/testing-strategy.md)** for workflow testing patterns.

---

## 🐳 Docker Deployment

### Quick Start with Docker

```bash
# Pull and run the latest image (when available)
docker pull conexus:latest
docker run -d -p 8080:8080 --name conexus conexus:latest

# Or build locally
docker build -t conexus:latest .
docker run -d -p 8080:8080 --name conexus conexus:latest

# Test the service
curl http://localhost:8080/health
```

### Docker Compose (Recommended)

**Production deployment:**

```bash
# Start the service
docker compose up -d

# View logs
docker compose logs -f

# Stop the service
docker compose down

# Rebuild after code changes
docker compose up -d --build
```

**Development deployment:**

```bash
# Use development configuration with debug logging
docker compose -f docker-compose.yml -f docker-compose.dev.yml up

# View debug logs
docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f
```

### Configuration

**Environment Variables:**

```bash
# Server configuration
CONEXUS_HOST=0.0.0.0              # Server bind address
CONEXUS_PORT=8080                  # Server port

# Database configuration
CONEXUS_DB_PATH=/data/conexus.db   # SQLite database path

# Codebase configuration
CONEXUS_ROOT_PATH=/data/codebase   # Path to codebase to index

# Logging configuration
CONEXUS_LOG_LEVEL=info             # Log level (debug|info|warn|error)
CONEXUS_LOG_FORMAT=json            # Log format (json|text)

# Embedding configuration (optional)
CONEXUS_EMBEDDING_PROVIDER=openai  # Embedding provider (mock|openai)
CONEXUS_EMBEDDING_MODEL=text-embedding-3-small
OPENAI_API_KEY=sk-...              # OpenAI API key
```

**Volume Mounts:**

```yaml
volumes:
  # Persistent database storage
  - ./data:/data
  
  # Optional: Mount your codebase for indexing
  - /path/to/your/code:/data/codebase:ro
  
  # Optional: Mount config file
  - ./config.yml:/app/config.yml:ro
```

### Docker Image Details

**Multi-stage build:**
- **Builder**: `golang:1.24-alpine` (CGO enabled for SQLite)
- **Runtime**: `alpine:3.19` (minimal base, ca-certificates + sqlite-libs)

**Image specifications:**
- **Size**: ~19.5MB (optimized with multi-stage build)
- **User**: Non-root `conexus:1000`
- **Port**: 8080 (HTTP + MCP over JSON-RPC 2.0)
- **Health Check**: `GET /health` every 30s

**Security features:**
- Non-root execution (UID 1000)
- Static binary (no dynamic linking)
- Minimal attack surface (Alpine base)
- Read-only config option
- Health check monitoring

### MCP Server Endpoints

Once running, the service exposes:

**HTTP Endpoints:**
```bash
# Health check
curl http://localhost:8080/health
# Response: {"status":"healthy","version":"0.1.0-alpha"}

# Service info
curl http://localhost:8080/
# Response: Service info with MCP endpoint

# MCP JSON-RPC endpoint
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
```

**MCP Tools:**
1. `context.search` - Comprehensive search with filters
2. `context.get_related_info` - File/ticket context retrieval
3. `context.index_control` - Indexing operations
4. `context.connector_management` - Data source management

### Production Deployment

**With Docker Compose:**

```yaml
# docker-compose.prod.yml
services:
  conexus:
    image: conexus:latest
    restart: always
    environment:
      - CONEXUS_LOG_LEVEL=info
      - CONEXUS_LOG_FORMAT=json
    volumes:
      - conexus-data:/data
      - /mnt/codebase:/data/codebase:ro
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

volumes:
  conexus-data:
    driver: local
```

**Deploy:**
```bash
docker compose -f docker-compose.prod.yml up -d
```

### Monitoring

**Check health:**
```bash
# Container status
docker compose ps

# Health check status
docker inspect conexus | jq '.[0].State.Health'

# View logs
docker compose logs -f

# Check metrics
curl http://localhost:8080/health
```

**Troubleshooting:**
```bash
# View container logs
docker compose logs --tail=100

# Execute commands in container
docker compose exec conexus sh

# Check database
docker compose exec conexus ls -la /data/

# Restart service
docker compose restart
```

### Building from Source

```bash
# Build Docker image
docker build -t conexus:custom .

# Build with specific Go version
docker build --build-arg GO_VERSION=1.24 -t conexus:custom .

# Build and tag
docker build -t conexus:v0.1.0 -t conexus:latest .

# Push to registry (configure your registry)
docker tag conexus:latest registry.example.com/conexus:latest
docker push registry.example.com/conexus:latest
```

### Docker Best Practices

1. **Use Docker Compose** for orchestration
2. **Mount volumes** for data persistence
3. **Configure environment variables** for secrets
4. **Enable health checks** for monitoring
5. **Use named volumes** in production
6. **Check logs regularly** with `docker compose logs`
7. **Backup database** in `/data` directory regularly
8. **Limit resources** with Docker resource constraints if needed

---

## 📖 Performance & Sourcing

All performance claims in this README are backed by comprehensive benchmarks and documented sources:

### Primary Sources

| Document | What It Contains | Location |
|----------|------------------|----------|
| **[PERFORMANCE_BASELINE.md](PERFORMANCE_BASELINE.md)** | 71 benchmarks across all components | Root directory |
| **[Context Engine Internals](docs/architecture/context-engine-internals.md)** | Caching and retrieval algorithms | docs/architecture/ |
| **[Load Test Results](tests/load/results/)** | Stress testing and concurrency analysis | tests/load/results/ |
| **[Component Documentation](internal/)** | Implementation details and capabilities | internal/*/README.md |

### Benchmark Methodology

- **Test Environment**: AMD FX-9590, Linux, Go 1.24.9 [Source: PERFORMANCE_BASELINE.md:3-7]
- **Total Benchmarks**: 71 individual tests across vectorstore, indexer, and orchestrator [Source: PERFORMANCE_BASELINE.md:540-549]
- **Pass Rate**: 89% (17/19 targets met) [Source: PERFORMANCE_BASELINE.md:551-559]
- **Test Duration**: ~15 minutes total execution [Source: PERFORMANCE_BASELINE.md:672]

### Verification

To verify these metrics:
```bash
# Run performance benchmarks
cd tests/load
./run_benchmarks.sh

# Check current system performance
go test -bench=. ./...

# View detailed metrics
cat PERFORMANCE_BASELINE.md
```

### Context Retention Evidence

The context retention improvements are demonstrated through:
- **Session Management**: Full conversation history in `internal/orchestrator/state/manager.go`
- **Caching System**: 3-tier architecture in `docs/architecture/context-engine-internals.md:9870-10127`
- **Search Performance**: Hybrid search results in `internal/search/search.go`
- **Load Testing**: Concurrent user validation in `tests/load/results/STRESS_TEST_ANALYSIS.md`

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
- ✅ **Task 5.2**: Documentation Updates (completed with performance sourcing)
- ⏳ **Task 5.3**: Workflow Integration (pending)
- ⏳ **Task 5.4**: Protocol Tests (optional)

### Test Results

```
✅ All 53 integration tests passing
✅ Execution time: <1 second
✅ Evidence validation: 100%
✅ Schema compliance: 100%
✅ Real-world analysis: 5 scenarios validated
✅ Performance benchmarks: 71 tests with documented sources
✅ Context retention: Measurable improvements over standard LLM
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
