# Developer Onboarding Guide

**Conexus (Agentic Context Engine)**  
**Version:** 1.0  
**Last Updated:** October 12, 2025

## Welcome to the Conexus Team! 🚀

This guide will help you get from zero to productive contributor in under 2 hours. By the end, you'll have a fully functional local development environment and will have made your first contribution.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Setup](#environment-setup)
- [Project Structure](#project-structure)
- [Local Development Workflow](#local-development-workflow)
- [Your First Contribution](#your-first-contribution)
- [Debugging](#debugging)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

---

## Prerequisites

### Required Software

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.21+ | Primary backend language |
| **Docker** | 20.10+ | Local service orchestration |
| **Docker Compose** | 2.0+ | Multi-container management |
| **Git** | 2.30+ | Version control |
| **PostgreSQL** | 15+ | Metadata storage |
| **Make** | 4.0+ | Build automation |

### Installation Commands

**macOS (using Homebrew):**
```bash
brew install go docker docker-compose git postgresql@15 make
brew install --cask docker
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y golang-1.21 docker.io docker-compose git postgresql-15 make
sudo usermod -aG docker $USER  # Add yourself to docker group
newgrp docker  # Activate group without logout
```

**Verify Installations:**
```bash
go version          # Should show go1.21 or higher
docker --version    # Should show 20.10+
docker compose version
git --version
psql --version
make --version
```

### Recommended IDE Setup

**VSCode (Recommended):**
```bash
# Install VSCode
brew install --cask visual-studio-code  # macOS
# or download from https://code.visualstudio.com/

# Install Go extension
code --install-extension golang.go
code --install-extension ms-azuretools.vscode-docker
```

**GoLand (Alternative):**
- Download from https://www.jetbrains.com/go/
- Professional IDE with built-in Go support

### Required Accounts

- **GitHub Account** - For code access and contributions
- **Docker Hub Account** (optional) - For pulling private images

---

## Environment Setup

### Step 1: Clone the Repository

```bash
# Clone the repo
git clone https://github.com/conexus-org/conexus.git
cd conexus

# Verify you're on the main branch
git branch
```

### Step 2: Install Go Dependencies

```bash
# Download all Go modules
go mod download

# Verify dependencies
go mod verify

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install gotest.tools/gotestsum@latest
```

### Step 3: Set Up Local Services

**Start PostgreSQL and Qdrant via Docker Compose:**

```bash
# Start all services in background
docker compose up -d

# Verify services are running
docker compose ps

# Expected output:
# NAME                SERVICE    STATUS
# conexus-postgres-1      postgres   Up
# conexus-qdrant-1        qdrant     Up
# conexus-redis-1         redis      Up
```

**Service URLs:**
- PostgreSQL: `localhost:5432`
- Qdrant (Vector DB): `localhost:6333`
- Redis (Cache): `localhost:6379`

### Step 4: Configure Environment Variables

```bash
# Copy example config
cp .env.example .env

# Edit with your preferred editor
vim .env  # or nano, code, etc.
```

**Minimum Required Configuration:**

```bash
# Database
DATABASE_URL=postgresql://conexus:ace_dev@localhost:5432/ace_dev?sslmode=disable

# Vector Database
QDRANT_URL=http://localhost:6333
QDRANT_API_KEY=  # Leave empty for local dev

# API Configuration
API_PORT=8080
LOG_LEVEL=debug

# MCP Configuration
MCP_SERVER_PORT=9090

# Optional: External Services (can be mocked locally)
# EMBEDDING_API_KEY=your_key_here
# RERANKER_API_KEY=your_key_here
```

### Step 5: Initialize Database

```bash
# Run migrations
make db-migrate

# Seed development data (optional)
make db-seed

# Verify database setup
psql $DATABASE_URL -c "\dt"  # Should show tables
```

### Step 6: Build and Run

```bash
# Build the application
make build

# Run all tests to verify setup
make test

# Start the development server
make dev

# Expected output:
# 2025-10-12T10:00:00Z INFO Starting Conexus server
# 2025-10-12T10:00:00Z INFO HTTP server listening on :8080
# 2025-10-12T10:00:00Z INFO MCP server listening on :9090
```

### Step 7: Verify Installation

Open a new terminal and test the API:

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","version":"0.1.0","timestamp":"2025-10-12T10:00:00Z"}

# Test search endpoint (will be empty initially)
curl http://localhost:8080/api/v1/search?q=test

# Expected response:
# {"results":[],"total":0,"took_ms":5}
```

🎉 **Congratulations!** Your development environment is ready!

---

## Project Structure

```
conexus/
├── cmd/
│   ├── server/          # Main HTTP/MCP server
│   ├── indexer/         # Background indexing service
│   └── cli/             # Command-line tools
├── internal/
│   ├── api/             # REST API handlers
│   │   ├── handlers/    # HTTP request handlers
│   │   ├── middleware/  # Auth, logging, rate limiting
│   │   └── routes/      # Route definitions
│   ├── mcp/             # Model Context Protocol implementation
│   │   ├── server/      # MCP server logic
│   │   └── handlers/    # MCP command handlers
│   ├── retrieval/       # Core context engine
│   │   ├── embeddings/  # Embedding generation
│   │   ├── search/      # Vector + hybrid search
│   │   ├── rerank/      # Result reranking
│   │   └── graph/       # GraphRAG implementation
│   ├── indexing/        # Document ingestion pipeline
│   │   ├── chunking/    # Text chunking strategies
│   │   ├── processors/  # Language-specific processors
│   │   └── connectors/  # Data source connectors
│   ├── storage/         # Data layer
│   │   ├── postgres/    # PostgreSQL client
│   │   ├── qdrant/      # Vector DB client
│   │   └── redis/       # Cache client
│   ├── auth/            # Authentication & authorization
│   ├── config/          # Configuration management
│   └── telemetry/       # Logging, metrics, tracing
├── pkg/                 # Public libraries (can be imported externally)
│   ├── models/          # Shared data structures
│   └── utils/           # Common utilities
├── test/
│   ├── integration/     # Integration tests
│   ├── fixtures/        # Test data
│   └── mocks/           # Mock implementations
├── docs/                # Documentation (you are here!)
├── scripts/             # Development automation
├── deployments/         # Docker, K8s, Terraform configs
└── .github/             # CI/CD workflows

```

### Key Directories Explained

**`cmd/`** - Application entry points. Each subdirectory is a separate binary.

**`internal/`** - Private application code. Cannot be imported by external projects.
- `api/` - REST API implementation (HTTP handlers, middleware)
- `mcp/` - MCP server for IDE integration
- `retrieval/` - **Core context engine** (embeddings, search, reranking)
- `indexing/` - Document processing and ingestion
- `storage/` - Database clients and data access

**`pkg/`** - Public libraries that could be used by external projects.

**`test/`** - Test code and fixtures separate from source.

### Module Organization

Conexus uses Go modules with clear dependency boundaries:

```
github.com/conexus-org/conexus
├── internal/api         (depends on: retrieval, storage, auth)
├── internal/retrieval   (depends on: storage)
├── internal/indexing    (depends on: storage, retrieval)
└── internal/storage     (no internal dependencies)
```

---

## Local Development Workflow

### Daily Development Cycle

```bash
# 1. Start your day: sync with main
git checkout main
git pull origin main

# 2. Create feature branch
git checkout -b feature/your-feature-name

# 3. Start services (if not already running)
docker compose up -d
make dev  # Runs with hot reload

# 4. Make changes, run tests frequently
make test-unit              # Fast unit tests
make test-integration       # Slower integration tests

# 5. Check code quality
make lint                   # Run golangci-lint
make fmt                    # Format code

# 6. Run full validation before commit
make validate               # Runs: fmt, lint, test, build

# 7. Commit and push
git add .
git commit -m "feat: add new feature"
git push origin feature/your-feature-name
```

### Hot Reload Development

We use `air` for hot reload during development:

```bash
# Install air (if not already installed)
go install github.com/cosmtrek/air@latest

# Start with hot reload
make dev
# or directly:
air

# Now any code changes automatically rebuild and restart the server
```

### Running Specific Tests

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run specific package tests
go test ./internal/retrieval/...

# Run single test function
go test -v -run TestSearchHandler ./internal/api/handlers

# Run with coverage
make test-coverage
open coverage.html  # View coverage report
```

### Database Operations

```bash
# Create new migration
make db-migration name=add_user_preferences

# Apply migrations
make db-migrate

# Rollback last migration
make db-rollback

# Reset database (WARNING: destroys all data)
make db-reset

# Access PostgreSQL console
make db-console
# or directly:
psql $DATABASE_URL
```

### Working with Qdrant (Vector DB)

```bash
# Access Qdrant web UI
open http://localhost:6333/dashboard

# View collections via API
curl http://localhost:6333/collections

# Create test collection
curl -X PUT http://localhost:6333/collections/test \
  -H 'Content-Type: application/json' \
  -d '{"vectors":{"size":1536,"distance":"Cosine"}}'
```

### Debugging

**VSCode Launch Configuration** (`.vscode/launch.json`):

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/server",
      "env": {
        "DATABASE_URL": "postgresql://conexus:ace_dev@localhost:5432/ace_dev?sslmode=disable"
      },
      "args": []
    },
    {
      "name": "Debug Current Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}"
    }
  ]
}
```

**GoLand/IntelliJ:**
1. Right-click `cmd/server/main.go`
2. Select "Debug 'go build server'"
3. Set breakpoints by clicking line numbers

### Logging and Monitoring

```bash
# View server logs
make logs

# View specific service logs
docker compose logs postgres
docker compose logs qdrant

# Follow logs in real-time
docker compose logs -f

# View structured logs with filtering (requires jq)
make logs | jq 'select(.level == "error")'
```

---

## Your First Contribution

Let's make a simple but meaningful contribution to get familiar with the codebase.

### Starter Task: Add Repository Stats Endpoint

**Goal:** Add a new API endpoint that returns repository statistics.

**Location:** `internal/api/handlers/stats.go`

**Step 1: Create the handler**

```go
// internal/api/handlers/stats.go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/conexus-org/conexus/internal/storage/postgres"
)

type StatsHandler struct {
	db *postgres.Client
}

func NewStatsHandler(db *postgres.Client) *StatsHandler {
	return &StatsHandler{db: db}
}

type StatsResponse struct {
	TotalDocuments int64     `json:"total_documents"`
	TotalChunks    int64     `json:"total_chunks"`
	IndexedAt      time.Time `json:"indexed_at"`
	Version        string    `json:"version"`
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	// TODO: Implement actual database queries
	stats := StatsResponse{
		TotalDocuments: 0,
		TotalChunks:    0,
		IndexedAt:      time.Now(),
		Version:        "0.1.0",
	}

	c.JSON(http.StatusOK, stats)
}
```

**Step 2: Add the route**

```go
// internal/api/routes/routes.go
// Add to the existing routes

func SetupRoutes(r *gin.Engine, handlers *Handlers) {
	// ... existing routes ...
	
	api := r.Group("/api/v1")
	{
		// Add stats route
		api.GET("/stats", handlers.Stats.GetStats)
	}
}
```

**Step 3: Write a test**

```go
// internal/api/handlers/stats_test.go
package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/conexus-org/conexus/internal/api/handlers"
)

func TestStatsHandler_GetStats(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	handler := handlers.NewStatsHandler(nil) // Mock DB if needed
	router.GET("/api/v1/stats", handler.GetStats)

	// Execute
	req, _ := http.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response handlers.StatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "0.1.0", response.Version)
}
```

**Step 4: Test your changes**

```bash
# Run the test
go test ./internal/api/handlers/stats_test.go

# Start the server
make dev

# Test the endpoint
curl http://localhost:8080/api/v1/stats
```

**Step 5: Submit your PR**

```bash
git add .
git commit -m "feat: add repository stats endpoint"
git push origin feature/add-stats-endpoint

# Create PR on GitHub
gh pr create --title "Add repository stats endpoint" \
  --body "Adds GET /api/v1/stats endpoint for repository statistics"
```

---

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Error:** `bind: address already in use`

**Solution:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use different port in .env
API_PORT=8081
```

#### 2. Database Connection Failed

**Error:** `connection refused` or `authentication failed`

**Solution:**
```bash
# Check PostgreSQL is running
docker compose ps postgres

# Restart PostgreSQL
docker compose restart postgres

# Verify credentials match .env
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL -c "SELECT 1"
```

#### 3. Go Module Errors

**Error:** `missing go.sum entry`

**Solution:**
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

#### 4. Docker Compose Issues

**Error:** `network not found` or `volume errors`

**Solution:**
```bash
# Stop all containers
docker compose down

# Remove volumes (WARNING: deletes data)
docker compose down -v

# Rebuild and restart
docker compose up -d --build
```

#### 5. Test Failures

**Error:** Tests pass locally but fail in CI

**Solution:**
```bash
# Run tests in same environment as CI
docker compose -f docker-compose.test.yml up --abort-on-container-exit

# Check for race conditions
go test -race ./...

# Check for environment-specific issues
make test-ci
```

#### 6. Qdrant Connection Issues

**Error:** `failed to connect to Qdrant`

**Solution:**
```bash
# Check Qdrant status
curl http://localhost:6333/health

# Restart Qdrant
docker compose restart qdrant

# Check logs
docker compose logs qdrant
```

---

## Next Steps

### Learn the Codebase

1. **Read the Architecture Docs**
   - [Technical Architecture](../Technical-Architecture.md)
   - [API Specification](../API-Specification.md)
   - [Security & Compliance](../Security-Compliance.md)

2. **Explore Key Packages**
   - Start with `internal/retrieval` (core engine)
   - Then `internal/indexing` (document processing)
   - Finally `internal/api` and `internal/mcp` (interfaces)

3. **Review Test Examples**
   - `test/integration/search_test.go` - End-to-end search tests
   - `internal/retrieval/search_test.go` - Unit test examples

### Pick Up a Task

Check the [Good First Issues](https://github.com/conexus-org/conexus/labels/good-first-issue) label on GitHub.

Suggested starter tasks:
- Add new data source connector
- Improve test coverage for a package
- Add API endpoint validation
- Enhance logging in a specific module

### Join the Team

- **Slack:** #conexus-dev channel
- **Daily Standup:** 9:30 AM ET (optional for new members)
- **Code Reviews:** Participate in PR reviews to learn
- **Documentation:** Help improve docs as you learn

### Development Best Practices

- **Commit Often:** Small, atomic commits are easier to review
- **Test First:** Write tests for new features (TDD encouraged)
- **Ask Questions:** No question is too small - ask in #conexus-dev
- **Review Code:** Learn by reviewing others' PRs
- **Document:** Add comments for complex logic

---

## Resources

### Internal Documentation

- [Contributing Guide](../contributing/contributing-guide.md)
- [Testing Strategy](../contributing/testing-strategy.md)
- [Code Style Guide](../contributing/code-style.md)

### External Resources

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Qdrant Documentation](https://qdrant.tech/documentation/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

### Team Contacts

- **Tech Lead:** tech-lead@conexus-org.com
- **DevOps:** devops@conexus-org.com
- **Questions:** #conexus-dev on Slack

---

**Welcome aboard! We're excited to have you on the team.** 🎉

If you run into any issues with this guide, please open an issue or PR to improve it for the next developer.
