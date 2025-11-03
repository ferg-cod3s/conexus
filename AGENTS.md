# AGENTS.md - Development Guide for AI Agents

## 📚 Related Documentation

- **[Versioning Criteria](docs/VERSIONING_CRITERIA.md)** - When and how to bump versions
- **[Contributing Guide](docs/contributing/contributing-guide.md)** - How to contribute to Conexus
- **[Testing Strategy](docs/contributing/testing-strategy.md)** - Testing requirements and best practices
- **[Technical Architecture](docs/Technical-Architecture.md)** - System design and architecture
- **[Security & Compliance](docs/Security-Compliance.md)** - Security framework and compliance

---

## 🎯 Project Overview

**Conexus** is a **Model Context Protocol (MCP) server** that provides AI assistants with intelligent context about your codebase. It enables semantic search, code understanding, and project knowledge retrieval through standardized MCP tools.

### Current Status
- **Version**: 0.2.1-alpha
- **Status**: ✅ MCP Server Ready
- **Go Version**: 1.24.0
- **Test Coverage**: 85-90% target
- **License**: MIT

### Key Metrics
- **Go Files**: 156 total files
- **Lines of Code**: ~59,142 lines
- **Test Files**: 69 test files
- **Test Packages**: 41 packages
- **Documentation**: 78+ markdown files
- **README Size**: 1,023 lines

---

## 🛠️ Build/Lint/Test Commands

### Core Commands
```bash
# Build the main binary
go build -o conexus ./cmd/conexus

# Build for all platforms
./scripts/build-binaries.sh

# Run all tests
go test ./...

# Run specific test
go test -run TestSpecificFunction ./path/to/package

# Run tests with coverage
go test -cover ./...

# Run integration tests only
go test ./internal/testing/integration/...

# Run with verbose output
go test -v ./internal/testing/integration

# Run with race detector
go test -race ./...
```

### Performance & Benchmarks
```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkFunction ./path/to/package

# Profile tests
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
```

### Code Quality
```bash
# Run linter (requires golangci-lint)
golangci-lint run

# Format code
gofmt -s -w .

# Check for unused dependencies
go mod tidy

# Verify dependencies
go mod verify
```

### Docker Commands
```bash
# Build Docker image
docker build -t conexus:latest .

# Run with Docker Compose
docker-compose up -d

# Development environment
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Production environment
docker-compose -f docker-compose.prod.yml up

# Observability stack
docker-compose -f docker-compose.observability.yml up
```

---

## 📋 Code Style Guidelines

### Import Organization
```go
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // Third-party packages
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    // Internal packages (relative to project root)
    "github.com/ferg-cod3s/conexus/internal/agent"
    "github.com/ferg-cod3s/conexus/pkg/schema"
)
```

### Naming Conventions
- **Packages**: lowercase, single word when possible (`agent`, `orchestrator`, `vectorstore`)
- **Functions**: CamelCase, exported if public (`NewAgent()`, `executeWorkflow()`)
- **Variables**: camelCase, descriptive names (`userProfile`, `searchResults`)
- **Constants**: UPPER_SNAKE_CASE for exported (`MAX_RETRIES`, `DEFAULT_TIMEOUT`)
- **Interfaces**: usually -er suffix (`Agent`, `Locator`, `Analyzer`)

### Error Handling
```go
// Define specific error types
var (
    ErrInvalidInput = errors.New("invalid input provided")
    ErrNotFound     = errors.New("resource not found")
)

// Use error wrapping for context
func GetUser(ctx context.Context, id int64) (*User, error) {
    user, err := database.GetUserByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}
```

### Testing Patterns
- Use table-driven tests for multiple scenarios
- Follow Arrange-Act-Assert pattern
- Use testify/assert and testify/require
- Mock external dependencies
- Test both success and error paths

### Types & Interfaces
- Use concrete types where possible
- Define interfaces for behavior, not data
- Prefer composition over inheritance
- Use struct tags for JSON/DB serialization

### Documentation
- Package comments should explain purpose and usage
- Exported functions need godoc comments
- Include example usage in documentation
- Document error conditions and edge cases

---

## 🏗️ Project Structure and Architecture

### Directory Structure
```
conexus/
├── cmd/conexus/          # MCP server entry point
├── internal/
│   ├── agent/            # Agent system components
│   ├── mcp/             # MCP server implementation
│   ├── search/          # Search engine
│   ├── indexer/         # File indexing
│   ├── vectorstore/     # Vector database
│   ├── connectors/      # Data source connectors
│   ├── orchestrator/    # Workflow orchestration
│   ├── validation/      # Evidence validation
│   ├── security/        # Security utilities
│   └── tool/           # Tool execution
├── pkg/schema/          # Public schemas
├── tests/               # Test suite
├── docs/                # Documentation
├── scripts/             # Build and deployment scripts
├── observability/       # Monitoring configs
└── PROJECT_CONFIGS/     # Project configurations
```

### Core Components

| Component | Description | Status |
|-----------|-------------|--------|
| **MCP Server** | JSON-RPC 2.0 server with stdio transport | ✅ Complete |
| **Search Engine** | Hybrid vector + BM25 semantic search | ✅ Complete |
| **Index Manager** | File watching and incremental indexing | ✅ Complete |
| **Vector Store** | SQLite-backed vector embeddings | ✅ Complete |
| **File Scanner** | Intelligent code file discovery | ✅ Complete |
| **Orchestrator** | Multi-agent workflow coordination | ✅ Complete |
| **Validation** | Evidence validation and traceability | ✅ Complete |

### Technology Stack
- **Language**: Go 1.24.0
- **Database**: SQLite (with modernc.org/sqlite driver)
- **Vector Storage**: Built-in vector store with cosine similarity
- **Testing**: testify framework
- **Linting**: golangci-lint
- **Documentation**: Markdown with Mermaid diagrams
- **Containerization**: Docker with multi-stage builds
- **CI/CD**: GitHub Actions

---

## 🧪 Testing Patterns and Requirements

### Testing Pyramid
```
    /\
   /  \      10% - E2E Tests (slow, expensive)
  /____\
 /      \    20% - Integration Tests (moderate speed)
/________\
/          \  70% - Unit Tests (fast, cheap)
/____________\
```

### Test Distribution
| Test Type | Percentage | Count Target | Execution Time |
|-----------|------------|--------------|----------------|
| Unit | 70% | ~1000 tests | <2 minutes |
| Integration | 20% | ~200 tests | <3 minutes |
| E2E/Performance | 10% | ~50 tests | <5 minutes |

### Coverage Requirements
| Component | Unit Coverage | Integration Coverage |
|-----------|---------------|----------------------|
| Core retrieval | 90%+ | 80%+ |
| Indexing pipeline | 85%+ | 70%+ |
| API handlers | 80%+ | 90%+ |
| Storage layer | 85%+ | 85%+ |
| Utilities | 80%+ | N/A |
| **Overall Target** | **85-90%** | **75-80%** |

### Test Structure
```go
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange - Set up test data and dependencies
    mockRepo := new(MockRepository)
    mockRepo.On("Search", mock.Anything, "query").Return([]Result{
        {ID: "1", Title: "Test"},
    }, nil)
    
    handler := NewSearchHandler(mockRepo)
    req := httptest.NewRequest("GET", "/search?q=query", nil)
    w := httptest.NewRecorder()
    
    // Act - Execute the code under test
    handler.HandleSearch(w, req)
    
    // Assert - Verify expectations
    assert.Equal(t, http.StatusOK, w.Code)
    mockRepo.AssertExpectations(t)
}
```

### Running Tests
```bash
# Run all tests
make test  # or go test ./...

# Run only unit tests (fast)
make test-unit
go test ./... -short

# Run only integration tests
make test-integration
go test ./internal/testing/integration/...

# Run with coverage
make test-coverage
go test ./... -coverprofile=coverage.out

# Run specific test
go test -v -run TestSearchHandler ./internal/api/handlers

# Run benchmarks
go test -bench=. ./internal/retrieval

# Run with race detector
make test-race
go test -race ./...
```

---

## 📊 Performance Benchmarks and Targets

### Current Performance Status (from PERFORMANCE_BASELINE.md)

| Component | Metric | Target | Actual | Status |
|-----------|--------|--------|--------|--------|
| **Vectorstore** | Query latency | <1s p95 | 2.18s | ❌ FAIL |
| | Indexing throughput | >100 files/sec | 290+ files/sec | ✅ PASS |
| | Memory (10K chunks) | <100MB | 150MB | ⚠️ OVER |
| **Indexer** | File walking | >1K files/sec | 65K files/sec | ✅ EXCELLENT |
| | Chunking | >100 files/sec | 45K-79K files/sec | ✅ EXCELLENT |
| | Full index | >100 files/sec | 450 files/sec | ✅ EXCELLENT |
| | Memory (10K files) | <100MB | 58MB | ✅ EXCELLENT |
| **Orchestrator** | Request routing | <1s p95 | 10.35ms | ✅ EXCELLENT |
| | Agent invocation | <100ms | 67μs | ✅ EXCELLENT |
| | Workflow execution | <1s | 10.35ms | ✅ EXCELLENT |

### Key Performance Insights
- **Overall Pass Rate**: 89% (17/19 targets met)
- **Critical Issue**: Vector search latency for 10K docs (2.18s vs <1s target)
- **Strengths**: Indexer and orchestrator exceed all targets significantly
- **Memory Usage**: Efficient at 58MB for 10K files (42% under target)

### Benchmark Commands
```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific component benchmarks
go test -bench=BenchmarkVectorSearch ./internal/vectorstore
go test -bench=BenchmarkIndexing ./internal/indexer
go test -bench=BenchmarkOrchestration ./internal/orchestrator

# Profile benchmarks
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./...

# Run performance regression tests
cd tests/load
./run_benchmarks.sh
```

---

## 🔒 Security and Compliance Status

### Security Framework
- **Local-First Processing**: All operations performed locally on user's machine
- **No Data Exfiltration**: No network calls with user code/private data
- **Privacy-First Architecture**: Only embeddings with obfuscated filenames in cloud
- **Dependency Security**: Automated vulnerability scanning with Snyk/Dependabot
- **Secure Contribution**: PR reviews with security focus and SAST/DAST integration

### Compliance Status
| Framework | Status | Implementation |
|------------|--------|----------------|
| **GDPR** | 🔄 In Progress | Right to erasure, data portability, consent management |
| **HIPAA** | 📋 Planned | PHI handling, BAAs, technical safeguards |
| **SOC 2** | 📋 Planned | Security, availability, processing integrity, confidentiality |
| **Input Validation** | ✅ Implemented | Comprehensive validation framework |
| **Secrets Management** | ✅ Implemented | Rotation policies, HSM integration |

### Threat Modeling
- **STRIDE Analysis**: Complete with mitigations for all threat categories
- **PASTA Methodology**: Risk-centric view with attack simulation
- **Attack Surface Analysis**: Network, data, code, user, and third-party surfaces
- **Automated Scanning**: SAST, DAST, and dependency scanning in CI/CD

### Security Commands
```bash
# Run security scan
gosec ./...

# Check for known vulnerabilities
govulncheck ./...

# Run SAST (if available)
sonar-scanner

# Audit dependencies
go list -json -m all | nancy sleuth
```

---

## 🔄 Development Workflow and Processes

### Git Workflow
```
main                    # Production-ready code
├── develop            # Integration branch for features
    ├── feature/       # New features (feature/user-authentication)
    ├── bugfix/        # Bug fixes (bugfix/database-connection-leak)
    ├── hotfix/        # Critical fixes (hotfix/security-vulnerability)
    ├── refactor/      # Code refactoring (refactor/context-retrieval-engine)
    └── docs/          # Documentation updates (docs/api-endpoints)
```

### Branch Protection
- **Main Branch**: Required reviews, status checks, no force pushes
- **Dev Branch**: Status checks only, allows force pushes for development
- **Feature Branches**: No protection, created from dev

### Commit Message Format
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Pull Request Process
1. Create feature branch from `dev`
2. Develop and test locally
3. Push and create PR to `dev`
4. Automated checks run (tests, lint, security)
5. Code review and address feedback
6. Merge to `dev`
7. `dev` auto-syncs to `main` daily

### CI/CD Pipeline
- **Triggers**: Push to main/dev, PRs
- **Tests**: Unit tests, integration tests, race detection
- **Quality**: Linting, security scanning, coverage checks
- **Build**: Cross-platform binary compilation
- **Release**: Auto-tagging and GitHub releases on main merge

---

## 📈 Current Version and Roadmap

### Version Information
- **Current Version**: 0.2.1-alpha
- **Version Strategy**: Semantic Versioning (SemVer)
- **Pre-release Status**: Alpha (pre-1.0 releases)
- **Next Version**: 0.3.0-alpha (minor release)

### Version Bump Criteria
| Type | When to Use | Examples |
|------|-------------|----------|
| **Patch (0.1.x)** | Bug fixes, small features, performance improvements | Fix MCP compliance, add new MCP tools |
| **Minor (0.x.0)** | Significant new functionality, maintains backward compatibility | Multi-agent architecture, advanced search, connectors, real embedding providers |
| **Major (x.0.0)** | Breaking changes, production-ready milestone | Enterprise features, API changes, cloud architecture |

### Current Roadmap

#### v0.1.2-alpha (Current)
- ✅ MCP compliance fix
- ✅ Add `context.explain` and `context.grep` tools
- ✅ Test suite improvements

#### v0.2.0-alpha (Next Minor)
- 🔄 Multi-agent architecture implementation
- 🔄 Advanced search with code relationships
- 🔄 Enhanced connector management
- 🔄 Real-time indexing capabilities

#### v1.0.0 (Production Ready)
- 📋 Enterprise security and compliance
- 📋 Multi-tenant support
- 📋 Cloud deployment capabilities
- 📋 Advanced monitoring and observability

---

## 🛠️ All Relevant Commands and Tools

### Development Commands
```bash
# Environment setup
go mod download
go mod tidy

# Build commands
go build -o conexus ./cmd/conexus
./scripts/build-binaries.sh

# Testing commands
go test ./...                           # All tests
go test -v ./...                        # Verbose output
go test -race ./...                     # Race detection
go test -cover ./...                    # With coverage
go test -bench=. ./...                  # Benchmarks
go test -run TestFunction ./path/to/pkg   # Specific test

# Code quality
golangci-lint run                      # Linting
gofmt -s -w .                         # Formatting
go vet ./...                           # Static analysis
go mod verify                           # Verify dependencies
```

### Docker Commands
```bash
# Build and run
docker build -t conexus:latest .
docker run -p 8080:8080 conexus:latest

# Docker Compose environments
docker-compose up -d                    # Production
docker-compose -f docker-compose.dev.yml up  # Development
docker-compose -f docker-compose.observability.yml up  # Monitoring

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t conexus:latest .
```

### MCP Server Commands
```bash
# Run MCP server (stdio mode)
./bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)

# Run with environment variables
CONEXUS_DB_PATH=./data/db.sqlite CONEXUS_LOG_LEVEL=debug ./bin/conexus-*

# Run in HTTP mode (for testing)
CONEXUS_PORT=3000 ./bin/conexus-*

# Test MCP tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus
```

### Performance Commands
```bash
# Run performance benchmarks
cd tests/load
./run_benchmarks.sh

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

### Monitoring and Debugging
```bash
# Health check
curl http://localhost:8080/health

# Metrics endpoint
curl http://localhost:8080/metrics

# Debug endpoints
curl http://localhost:8080/debug/pprof/
curl http://localhost:8080/debug/pprof/heap
curl http://localhost:8080/debug/pprof/goroutine

# Observability stack
docker-compose -f docker-compose.observability.yml up -d
# Grafana: http://localhost:3000
# Prometheus: http://localhost:9090
```

### Configuration Commands
```bash
# Generate configuration
./conexus config init

# Validate configuration
./conexus config validate

# Test database connection
./conexus db test

# Index a project
./conexus index --path /path/to/project

# Search functionality
./conexus search "query terms"
```

---

## 📚 Key Dependencies

### Core Dependencies
```go
// Testing
require github.com/stretchr/testify v1.11.1

// Database
require modernc.org/sqlite v1.39.1
require modernc.org/memory v1.11.0

// Observability
require go.opentelemetry.io/otel v1.38.0
require github.com/getsentry/sentry-go v0.36.0

// HTTP/Networking
require github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2
require golang.org/x/net v0.43.0

// Security
require golang.org/x/crypto v0.41.0
require github.com/golang-jwt/jwt/v5 v5.3.0
```

### Development Tools
- **golangci-lint**: Go linting and static analysis
- **testify**: Testing framework with assertions and mocks
- **gofmt**: Go code formatting
- **go mod**: Go module management
- **docker**: Containerization
- **github-cli**: GitHub operations

---

## 🎯 Best Practices Summary

### DO ✅
- Write tests before implementation (TDD)
- Use table-driven tests for multiple scenarios
- Test error paths, not just happy path
- Use meaningful test names
- Keep tests independent and fast
- Mock external dependencies
- Follow Go idioms and best practices
- Document public APIs and complex logic
- Handle errors gracefully with context
- Use structured logging
- Write clear, descriptive commit messages
- Review security implications of changes

### DON'T ❌
- Test implementation details instead of behavior
- Write flaky or non-deterministic tests
- Use real databases in unit tests
- Ignore test failures
- Hardcode configuration values
- Commit sensitive data or secrets
- Make breaking changes without version bump
- Skip error handling
- Write overly complex code
- Use global variables
- Ignore performance implications
- Forget to update documentation

---

## 🔗 Quick Reference

### Environment Variables
```bash
CONEXUS_DB_PATH=/path/to/database.sqlite     # Database location
CONEXUS_LOG_LEVEL=info                      # debug|info|warn|error
CONEXUS_PORT=8080                           # HTTP port (optional)
CONEXUS_ROOT_PATH=/path/to/project           # Project root to index

# Embedding Configuration
CONEXUS_EMBEDDING_PROVIDER=mock             # Provider: mock|anthropic
CONEXUS_EMBEDDING_MODEL=mock-768           # Model name (provider-specific)
CONEXUS_EMBEDDING_DIMENSIONS=768            # Vector dimensions
ANTHROPIC_API_KEY=your-api-key-here         # Required for Anthropic provider
```

### Configuration Files
- `config.yml` - Main configuration
- `config.example.yml` - Configuration template
- `.golangci.yml` - Linting configuration
- `docker-compose*.yml` - Docker configurations
- `.github/workflows/` - CI/CD workflows

### Important Ports
- `8080` - Main HTTP server
- `3000` - Development server (if configured)
- `9090` - Metrics (Prometheus)
- `6333` - Qdrant vector database (if used)

### Test Status Quick Check
```bash
# Current status: 25/27 packages passing
go test ./... 2>&1 | grep -E "(PASS|FAIL|ok)" | tail -5
```

---

**Last Updated**: 2025-10-26  
**Project Version**: 0.1.2-alpha  
**Document Version**: 2.0

For questions or improvements to this guide, see the [contributing guide](docs/contributing/contributing-guide.md) or open an issue.

---

# Deep Research & Analysis Command

Conducts comprehensive research across your codebase, documentation, and external sources to provide deep understanding and actionable insights.

## How It Works

This command orchestrates multiple specialized agents in a carefully designed workflow:

### Phase 1: Discovery (Parallel)

- 🔍 **codebase-locator** finds relevant files and components
- 📚 **research-locator** discovers existing documentation and notes

### Phase 2: Analysis (Sequential)

- 🧠 **codebase-analyzer** understands implementation details
- 💡 **research-analyzer** extracts insights from documentation

### Phase 3: External Research (Optional)

- 🌐 **web-search-researcher** gathers external context and best practices

## When to Use

**Perfect for:**

- Starting work on unfamiliar parts of the codebase
- Planning new features or major changes
- Understanding complex systems or architectures
- Debugging issues that span multiple components
- Creating onboarding documentation

**Example Research Questions:**

- "How does the user authentication system work?"
- "What's the current state of our API rate limiting?"
- "How should we implement real-time notifications?"
- "What are the performance bottlenecks in our data processing pipeline?"

## What You'll Get

### Research Report Includes:

- **Code Analysis**: File locations, key functions, and implementation patterns
- **Documentation Insights**: Existing docs, decisions, and context
- **Architecture Overview**: How components interact and data flows
- **External Research**: Best practices, alternatives, and recommendations
- **Action Items**: Specific next steps based on findings

### Sample Output Structure:

```
## Research Summary
- Objective: [Your research question]
- Key Findings: [3-5 major insights]
- Confidence Level: [High/Medium/Low]

## Codebase Analysis
- Core Files: [List with explanations]
- Key Functions: [Important methods and their purposes]
- Data Flow: [How information moves through the system]

## Documentation Insights
- Existing Docs: [Relevant documentation found]
- Past Decisions: [Architecture decisions and reasoning]
- Known Issues: [Documented problems or limitations]

## Recommendations
- Immediate Actions: [What to do first]
- Long-term Considerations: [Strategic recommendations]
- Potential Risks: [Things to watch out for]
```

## Platform-Specific Usage

### Claude Code (.claude.ai/code)

Use direct command arguments with native parsing:

```bash
# Basic research with defaults
/research "How does the authentication system work?"

# Advanced research with explicit parameters
/research "Analyze user session management" --scope=codebase --depth=deep

# Research from ticket file
/research --ticket="docs/tickets/auth-ticket.md" --scope=both --depth=medium
```

**Default Values:**

- `scope`: `"codebase"`
- `depth`: `"medium"`

### OpenCode (opencode.ai)

Use YAML frontmatter format for argument specification:

```yaml
---
name: research
mode: command
scope: codebase
depth: deep
model: anthropic/claude-sonnet-4
temperature: 0.1
---
Analyze the authentication system including user models, session handling, middleware, and security patterns.
```

**Default Values:**

- `scope`: `"both"` (codebase + thoughts)
- `depth`: `"medium"`
- `model`: `"anthropic/claude-sonnet-4"`
- `temperature`: `0.1`

### MCP-Compatible Clients (Cursor, VS Code, etc.)

Use JSON parameter format for structured arguments:

```json
{
  "tool": "research",
  "parameters": {
    "query": "How does the authentication system work?",
    "scope": "codebase",
    "depth": "deep",
    "ticket": "docs/tickets/auth-ticket.md"
  }
}
```

**Default Values:**

- Same as Claude Code defaults
- JSON schema validation
- Structured parameter passing

## Pro Tips

1. **Be Specific**: "Research authentication" vs "Research OAuth2 implementation and session management"
2. **Set Context**: Include any constraints, requirements, or specific areas of focus
3. **Follow Up**: Use results to inform `/plan` and `/execute` commands
4. **Iterate**: Research findings often lead to more specific research questions
5. **Platform Awareness**: Use platform-specific syntax (direct args vs YAML vs JSON) for optimal results

## Enhanced Subagent Orchestration

### Advanced Research Workflow

For complex research requiring deep analysis across multiple domains:

#### Phase 1: Comprehensive Discovery (Parallel Execution)

- **codebase-locator**: Maps all relevant files, components, and directory structures
- **research-locator**: Discovers existing documentation, past decisions, and technical notes
- **codebase-pattern-finder**: Identifies recurring implementation patterns and architectural approaches
- **web-search-researcher**: Gathers external best practices and industry standards (when applicable)

#### Phase 2: Deep Analysis (Sequential Processing)

- **codebase-analyzer**: Provides detailed implementation understanding with file:line evidence
- **research-analyzer**: Extracts actionable insights from documentation and historical context
- **system-architect**: Analyzes architectural implications and design patterns
- **performance-engineer**: Evaluates performance characteristics and optimization opportunities

#### Phase 3: Domain-Specific Assessment (Conditional)

- **database-expert**: Analyzes data architecture and persistence patterns
- **api-builder**: Evaluates API design and integration approaches
- **security-scanner**: Assesses security architecture and potential vulnerabilities
- **compliance-expert**: Reviews regulatory compliance requirements
- **infrastructure-builder**: Analyzes deployment and infrastructure implications

#### Phase 4: Synthesis & Validation (Parallel)

- **code-reviewer**: Validates research findings against code quality standards
- **test-generator**: Identifies testing gaps and coverage requirements
- **quality-testing-performance-tester**: Provides performance benchmarking insights

### Orchestration Best Practices

1. **Parallel Discovery**: Always start with multiple locators running simultaneously for comprehensive coverage
2. **Sequential Analysis**: Process analyzers sequentially to build upon locator findings
3. **Domain Escalation**: Engage domain specialists when research reveals specialized concerns
4. **Validation Gates**: Use reviewer agents to validate findings before synthesis
5. **Iterative Refinement**: Re-engage subagents as new questions emerge from initial findings

### Research Quality Indicators

- **Comprehensive Coverage**: Multiple agents provide overlapping validation
- **Evidence-Based**: All findings include specific file:line references
- **Contextual Depth**: Historical decisions and architectural rationale included
- **Actionable Insights**: Clear next steps and implementation guidance provided
- **Risk Assessment**: Potential issues and constraints identified

### Performance Optimization

- **Agent Sequencing**: Optimized order minimizes redundant analysis
- **Context Sharing**: Agents share findings to avoid duplicate work
- **Early Termination**: Stop analysis when sufficient understanding is achieved
- **Caching Strategy**: Leverage cached results for similar research topics

## Integration with Other Commands

- **→ /plan**: Use research findings to create detailed implementation plans
- **→ /execute**: Begin implementation with full context
- **→ /document**: Create documentation based on research insights
- **→ /review**: Validate that implementation matches research findings

---

_Ready to dive deep? Ask me anything about your codebase and I'll provide comprehensive insights to guide your next steps._