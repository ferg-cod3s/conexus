# Conexus - MCP Server for Context-Aware AI Assistants

**Version**: 0.1.2-alpha  
**Status**: ✅ MCP Server Ready  
**Go Version**: 1.23.4

[![Go Tests](https://img.shields.io/badge/tests-passing-brightgreen)]()
[![Coverage](https://img.shields.io/badge/coverage-85%25-green)]()
[![License](https://img.shields.io/badge/license-MIT-blue)]()

---

## 🎯 Overview

Conexus is a **Model Context Protocol (MCP) server** that provides AI assistants with intelligent context about your codebase. It enables semantic search, code understanding, and project knowledge retrieval through standardized MCP tools.

### Key Features

- 🔌 **MCP Server**: First-class Model Context Protocol server for AI assistants
- 🔍 **Semantic Search**: Hybrid vector + keyword search across your codebase
- 📁 **File Context**: Intelligent file relationships and project structure understanding
- ⚡ **Fast Performance**: Sub-second context retrieval with intelligent caching
- 🛡️ **Security First**: Rate limiting, security headers, and input validation
- 🛠️ **Easy Integration**: Works with Claude Desktop, Cursor, and other MCP clients
- 🧪 **Well Tested**: Comprehensive test suite with real-world validation

---

## 🚀 Quick Start

### Prerequisites

- **Node.js 18+** or **Bun** (for npm/bunx installation)
- Git

### Installation

**Option 1: Local Installation (Recommended)**

```bash
# Clone the repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Build the binaries
./scripts/build-binaries.sh

# Run directly
./bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)

# Or install locally
npm install
npm run build:all
npm link
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
./bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)

# Run with environment variables
CONEXUS_DB_PATH=./data/db.sqlite CONEXUS_LOG_LEVEL=debug ./bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)

# Run in HTTP mode (for testing)
CONEXUS_PORT=3000 ./bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)
```

---

## 🔌 MCP Integration

Conexus is a dedicated **MCP server** that provides AI assistants with intelligent context about your codebase through the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). It integrates seamlessly with Claude Desktop, Cursor, and other MCP-compatible clients.

### Why Use Conexus as an MCP Server?

Conexus provides AI assistants with **intelligent codebase context** that goes beyond simple file search:

#### 🔍 **Smart Code Discovery**
- **Semantic Search**: Find code by meaning, not just keywords
- **Hybrid Search**: Combines vector similarity with BM25 keyword matching
- **File Relationships**: Understand how files and functions connect
- **Project Structure**: Intelligent awareness of codebase organization

#### ⚡ **Fast Performance**
- **Sub-Second Retrieval**: Get relevant context in under 1 second
- **Intelligent Caching**: 98% cache hit rate for repeated queries
- **Efficient Indexing**: Quickly processes large codebases

#### 🛠️ **MCP Tools**
- **context.search**: Semantic search across your entire codebase
- **context.get_related_info**: Find files and discussions related to specific code
- **context.index_control**: Manage indexing operations
- **context.connector_management**: Configure data sources

### Quick MCP Setup (<5 minutes)

**Option 1: Local Binary (Recommended for MCP clients)**

```bash
# Clone and build
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus
./scripts/build-binaries.sh

# Use the local binary
./bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)
```

Configure in your MCP client (Claude Desktop, Cursor, etc.):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["@agentic-conexus/mcp"],
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

## 🏗️ Architecture

### MCP Server Architecture

```
┌─────────────────────────────────────────────────────┐
│                  MCP Server                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐ │
│  │   Search    │  │   Index     │  │  Connectors │ │
│  │   Engine    │  │  Manager    │  │  Manager    │ │
│  └─────────────┘  └─────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
   ┌─────▼─────┐   ┌────▼─────┐   ┌────▼─────┐
   │   Vector  │   │  SQLite   │  │  File     │
   │  Search   │  │ Database  │  │ System    │
   │   Store   │   │   Store   │  │ Scanner   │
   └───────────┘   └───────────┘   └───────────┘
```

### Core Components

| Component | Description | Status |
|-----------|-------------|--------|
| **MCP Server** | JSON-RPC 2.0 server with stdio transport | ✅ Complete |
| **Search Engine** | Hybrid vector + BM25 semantic search | ✅ Complete |
| **Index Manager** | File watching and incremental indexing | ✅ Complete |
| **Vector Store** | SQLite-backed vector embeddings | ✅ Complete |
| **File Scanner** | Intelligent code file discovery | ✅ Complete |

---

## 🧪 Testing

### Test Suite Overview

Conexus has comprehensive tests covering the MCP server functionality:

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
go test -run TestMCPServerIntegration ./internal/testing/integration
```

### Performance Benchmarks

Key performance metrics for the MCP server:

#### Search Performance
- **Search Latency**: ~11ms total (routing + BM25 search)
- **Cache Hit Rate**: 98% for repeated queries
- **Vector Search**: 248ms for 1K documents
- **Hybrid Search**: Combines vector + keyword matching

#### Indexing Performance  
- **File Processing**: 65,000 files/sec discovery
- **Indexing Speed**: 450 files/sec with embeddings
- **Memory Usage**: 58MB for 10K files
- **Update Speed**: Incremental updates in <1 second

#### MCP Server Performance
- **Tool Response**: <100ms for most operations
- **Concurrent Requests**: Handles multiple AI assistant queries
- **Memory Efficiency**: Optimized for long-running server processes

For detailed benchmarks, see [`PERFORMANCE_BASELINE.md`](PERFORMANCE_BASELINE.md).

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

### Environment Variables

Configure the MCP server with environment variables:

```bash
# Database location
export CONEXUS_DB_PATH=/path/to/project/.conexus/db.sqlite

# Logging level
export CONEXUS_LOG_LEVEL=info  # debug|info|warn|error

# Run in HTTP mode instead of stdio (for development)
export CONEXUS_PORT=3000

# Project root to index
export CONEXUS_ROOT_PATH=/path/to/project

# Rate Limiting Configuration
export CONEXUS_RATE_LIMIT_ENABLED=true
export CONEXUS_RATE_LIMIT_ALGORITHM=sliding_window  # sliding_window|token_bucket
export CONEXUS_RATE_LIMIT_DEFAULT_REQUESTS=100      # requests per window
export CONEXUS_RATE_LIMIT_DEFAULT_WINDOW=1m         # time window
export CONEXUS_RATE_LIMIT_HEALTH_REQUESTS=1000      # health endpoint limit
export CONEXUS_RATE_LIMIT_WEBHOOK_REQUESTS=10000    # webhook endpoint limit
export CONEXUS_RATE_LIMIT_AUTH_REQUESTS=1000        # authenticated requests limit
# Redis support for distributed rate limiting
export CONEXUS_RATE_LIMIT_REDIS_ENABLED=true
export CONEXUS_RATE_LIMIT_REDIS_ADDR=localhost:6379
export CONEXUS_RATE_LIMIT_REDIS_PASSWORD=your-password

# HTTPS/TLS Configuration (for HTTP mode)
export CONEXUS_TLS_ENABLED=true
export CONEXUS_TLS_CERT_FILE=/path/to/cert.pem
export CONEXUS_TLS_KEY_FILE=/path/to/key.pem
# Or for Let's Encrypt auto-cert:
export CONEXUS_TLS_AUTO_CERT=true
export CONEXUS_TLS_AUTO_CERT_DOMAINS="yourdomain.com,www.yourdomain.com"
export CONEXUS_TLS_AUTO_CERT_EMAIL="admin@yourdomain.com"
```

### HTTPS/TLS Security

Conexus supports HTTPS with automatic TLS certificate management:

#### Development (Self-Signed Certificates)
```bash
# Generate self-signed certificates for development
./scripts/generate-dev-certs.sh localhost ./data/tls

# Configure environment
export CONEXUS_TLS_ENABLED=true
export CONEXUS_TLS_CERT_FILE=./data/tls/cert.pem
export CONEXUS_TLS_KEY_FILE=./data/tls/key.pem
```

#### Production (Let's Encrypt)
```bash
export CONEXUS_TLS_AUTO_CERT=true
export CONEXUS_TLS_AUTO_CERT_DOMAINS="yourdomain.com,api.yourdomain.com"
export CONEXUS_TLS_AUTO_CERT_EMAIL="admin@yourdomain.com"
```

#### Manual Certificates
```bash
export CONEXUS_TLS_CERT_FILE=/etc/ssl/certs/yourdomain.crt
export CONEXUS_TLS_KEY_FILE=/etc/ssl/private/yourdomain.key
```

**Security Features:**
- TLS 1.2+ only (configurable)
- Secure cipher suites by default
- HTTP to HTTPS automatic redirection
- HSTS headers for enhanced security

### MCP Client Configuration

Most configuration is done through your MCP client (Claude Desktop, Cursor, etc.):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "/path/to/your/project/.conexus/db.sqlite",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  }
}
```

---

## 🛠️ Available MCP Tools

Conexus provides these MCP tools for AI assistants:

### `context.search`
Semantic search across your codebase with hybrid vector + keyword matching.

**Usage:**
```
"Search for authentication middleware functions"
"Find database query implementations"
"Show error handling patterns"
```

### `context.get_related_info`
Find files, discussions, and context related to specific files or tickets.

**Usage:**
```
"What's the history of this file?"
"Show PRs related to this issue"
"Find discussions about this component"
```

### `context.index_control`
Manage indexing operations (start, stop, status, reindex).

**Usage:**
```
"Check indexing status"
"Reindex the codebase"
"Start automatic indexing"
```

### `context.connector_management`
Configure data source connectors (GitHub, Slack, Jira, etc.).

**Usage:**
```
"List available connectors"
"Add GitHub connector"
"Configure Slack integration"
```

For detailed API documentation, see **[MCP Integration Guide](docs/getting-started/mcp-integration-guide.md)**.

---

## 🚀 Future Enhancements

While Conexus currently focuses on being a robust MCP server, we have plans for additional capabilities:

### Planned Features

- **🤖 Multi-Agent Architecture**: Specialized agents for complex code analysis tasks
- **✅ Evidence Validation**: Complete traceability for all code analysis results  
- **📊 Advanced Profiling**: Performance metrics and optimization recommendations
- **🔄 Workflow Orchestration**: Complex multi-step analysis workflows
- **🔐 Enterprise Features**: Authentication, authorization, and team management
- **🌐 Enhanced Connectors**: GitHub, Jira, Slack, and other data source integrations

### Enterprise Roadmap

For teams requiring advanced capabilities, we're planning:

- **Multi-tenant Support**: Isolated workspaces and team collaboration
- **Advanced Security**: RBAC, audit logging, and compliance features
- **Scalable Architecture**: Distributed processing and cloud deployment
- **Custom Integrations**: API for building custom data source connectors

These features are being designed based on user feedback and will be released in future versions. The current focus remains on providing the best MCP server experience for individual developers and teams.

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
- **Security Headers**: CSP, HSTS, X-Frame-Options, X-Content-Type-Options
- **CORS Protection**: Configurable cross-origin request handling
- **Rate Limiting**: Configurable request throttling with Redis support
- **Input Validation**: Comprehensive request sanitization

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

## 🏗️ Development

### Project Structure

```
conexus/
├── cmd/conexus/          # MCP server entry point
├── internal/
│   ├── mcp/             # MCP server implementation
│   │   ├── server.go    # Main MCP server
│   │   ├── handlers.go  # Tool handlers
│   │   └── schema.go    # MCP types
│   ├── search/          # Search engine
│   ├── indexer/         # File indexing
│   ├── vectorstore/     # Vector database
│   └── connectors/      # Data source connectors
├── pkg/schema/          # Public schemas
├── tests/               # Test suite
└── docs/                # Documentation
```

### Contributing

We welcome contributions! Please see:
- **[Contributing Guide](docs/contributing/contributing-guide.md)** - How to contribute
- **[Testing Strategy](docs/contributing/testing-strategy.md)** - Testing requirements
- **[Versioning Criteria](docs/VERSIONING_CRITERIA.md)** - When and how to bump versions
- **[Development Guide](AGENTS.md)** - Build, test, and development commands
- **[AI Assistant Guide](CLAUDE.md)** - Guidelines for AI development assistants

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
