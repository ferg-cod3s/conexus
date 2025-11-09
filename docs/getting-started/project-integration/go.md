# Go Project Integration

This guide covers integrating Conexus with Go projects, including standard library applications, web frameworks, and microservices.

## Quick Setup

### 1. Install Conexus

```bash
# Clone Conexus repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Build binaries
./scripts/build-binaries.sh

# Or install via go install
go install github.com/ferg-cod3s/conexus/cmd/conexus@latest
```

### 2. Configure MCP Client

**For OpenCode** (`.opencode/opencode.jsonc`):

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["bunx", "-y", "@agentic-conexus/mcp"],
      "environment": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "."
      },
      "enabled": true
    }
  },
  "agent": {
    "go-expert": {
      "tools": {
        "conexus": true
      }
    },
    "golang-developer": {
      "tools": {
        "conexus": true
      }
    },
    "api-builder-enhanced": {
      "tools": {
        "conexus": true
      }
    }
  }
}
```

**For Claude Desktop:**

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "/path/to/project/.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "/path/to/project"
      }
    }
  }
}
```

**For Claude Code** (`~/.claude/mcp.json`):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "."
      }
    }
  }
}
```

### 3. Project Configuration

Create `.conexus/config.yml`:

```yaml
project:
  name: "my-go-app"
  description: "Go web application"

codebase:
  root: "."
  include_patterns:
    - "**/*.go"
    - "**/go.mod"
    - "**/go.sum"
    - "**/*.md"
    - "**/Makefile"
    - "**/Dockerfile"
  exclude_patterns:
    - "**/vendor/**"
    - "**/.git/**"
    - "**/bin/**"
    - "**/dist/**"
    - "**/tmp/**"
    - "**/coverage.out"
    - "**/debug"
    - "**/*.test"

indexing:
  auto_reindex: true
  reindex_interval: "30m"
  chunk_size: 600

search:
  max_results: 50
  similarity_threshold: 0.7
```

## Framework-Specific Examples

### Standard Library HTTP Server

**Project Structure:**
```
go-web-app/
├── main.go
├── handlers/
│   ├── user.go
│   ├── auth.go
│   └── api.go
├── models/
│   ├── user.go
│   └── db.go
├── go.mod
├── .conexus/
└── .opencode/
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.go"
    - "**/go.mod"
    - "**/go.sum"
  exclude_patterns:
    - "**/vendor/**"
    - "**/bin/**"
```

**Queries:**
- "Find all HTTP handlers"
- "Show me the routing setup"
- "Search for database operations"
- "Locate middleware functions"

### Gin Web Framework

**Setup:**
```go
// main.go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    r.Run()
}
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.go"
    - "**/go.mod"
    - "**/go.sum"
  exclude_patterns:
    - "**/vendor/**"
```

**Recommended Agents:**
- `gin-pro`
- `go-expert`

**Queries:**
- "Find all Gin routes"
- "Show me the middleware"
- "Search for handlers"
- "Locate the router setup"

### Fiber Web Framework

**Project Structure:**
```
fiber-app/
├── main.go
├── routes/
├── middleware/
├── models/
├── go.mod
└── .conexus/
```

**Queries:**
- "Find all route definitions"
- "Show me the middleware stack"
- "Search for API endpoints"
- "Locate error handlers"

### Echo Web Framework

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.go"
    - "**/go.mod"
  exclude_patterns:
    - "**/vendor/**"
```

**Recommended Agents:**
- `echo-pro`
- `go-expert`

### Gorilla Mux

**Setup:**
```go
// main.go
package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/products", ProductsHandler).Methods("GET")
    http.ListenAndServe(":8080", r)
}
```

**Queries:**
- "Find all route handlers"
- "Show me the mux router setup"
- "Search for middleware"

### CLI Applications

**Project Structure:**
```
go-cli/
├── cmd/
│   ├── root.go
│   └── server.go
├── internal/
│   ├── config/
│   └── commands/
├── main.go
├── go.mod
└── .conexus/
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.go"
    - "**/go.mod"
    - "**/README.md"
  exclude_patterns:
    - "**/vendor/**"
    - "**/bin/**"
```

**Recommended Agents:**
- `cli-developer`
- `go-expert`

**Queries:**
- "Find all CLI commands"
- "Show me the command structure"
- "Search for flag definitions"
- "Locate the main function"

## Development Workflow

### Go Module Setup

```bash
# Initialize module
go mod init github.com/username/my-go-app

# Add dependencies
go get github.com/gin-gonic/gin
go get github.com/stretchr/testify

# Tidy dependencies
go mod tidy

# Download dependencies
go mod download
```

### Makefile Integration

```makefile
# Makefile
.PHONY: build test clean run index

build:
    go build -o bin/app ./cmd

test:
    go test ./...

test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

clean:
    go clean
    rm -rf bin/

run:
    go run ./cmd

index:
    ../conexus/bin/conexus-darwin-arm64 index

dev: index
    air  # If using air for hot reload
```

### Pre-commit Hooks

```bash
# .githooks/pre-commit
#!/bin/bash

# Index codebase
../conexus/bin/conexus-darwin-arm64 index --quiet

# Run tests
go test ./...

# Run linter
golangci-lint run

# Format code
gofmt -s -w .
goimports -w .
```

### VS Code Integration

```json
// .vscode/settings.json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "mcp.server.conexus": {
    "command": "npx",
    "args": ["-y", "@agentic-conexus/mcp"],
    "env": {
      "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite",
      "CONEXUS_ROOT_PATH": "${workspaceFolder}"
    }
  }
}

// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd",
      "env": {
        "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite"
      }
    }
  ]
}
```

## Testing Integration

### Test Structure

```go
// user_test.go
package models

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserCreate(t *testing.T) {
    // Test implementation
}

func TestUserGetByID(t *testing.T) {
    // Test implementation
}
```

**Conexus can help with:**
- "Find all test functions"
- "Show me the test setup"
- "Search for mocked dependencies"
- "Locate benchmark tests"

### Benchmarking

```go
// user_benchmark_test.go
package models

import (
    "testing"
)

func BenchmarkUserCreate(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Benchmark implementation
    }
}
```

## Performance Optimization

### For Large Go Codebases

```yaml
# .conexus/config.yml
indexing:
  chunk_size: 500
  workers: 4
  memory_limit: "1GB"

search:
  max_results: 40
  cache_enabled: true
  cache_ttl: "1h"

codebase:
  exclude_patterns:
    - "**/vendor/**"
    - "**/bin/**"
    - "**/tmp/**"
    - "**/dist/**"
```

### Memory Management

```bash
# Environment variables
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=1GB
export CONEXUS_INDEXING_MEMORY_LIMIT=512MB

# Go build flags
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
```

## Troubleshooting

### Common Go Issues

**Module issues:**
```bash
# Clean module cache
go clean -modcache

# Reinitialize modules
rm go.sum
go mod tidy

# Check module status
go mod verify
```

**Build errors:**
```bash
# Check Go version
go version

# Clean build cache
go clean -cache

# Rebuild
go build ./...
```

**Test failures:**
```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestUserCreate ./models

# Run with race detection
go test -race ./...
```

### Framework-Specific Issues

**Gin:**
- Include all route definitions
- Check for middleware setup
- Verify handler signatures

**Gorilla Mux:**
- Include router setup code
- Check for subrouters
- Verify middleware chains

**CLI apps:**
- Include cobra/viper setup
- Check command structure
- Verify flag definitions

## Best Practices

1. **Go Modules:** Always include `go.mod` and `go.sum` for dependency understanding

2. **Vendor Directory:** Exclude `vendor/` but include `go.mod` for dependency analysis

3. **Test Files:** Include `*_test.go` files for comprehensive code understanding

4. **Documentation:** Include `README.md`, godoc comments, and example files

5. **Build Files:** Include `Makefile`, `Dockerfile`, and CI configuration

## Integration Examples

### With golangci-lint

```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - gofmt
    - goimports
    - golint
    - govet
    - ineffassign
    - staticcheck
    - unused

issues:
  exclude-use-default: false
  exclude:
    - "G404" # Allow insecure random
```

### With Air (Hot Reload)

```toml
# .air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unnamed = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

### With Docker

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.conexus ./.conexus

EXPOSE 8080
CMD ["./main"]
```

### With GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Cache dependencies
        uses: actions/cache@v3
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Index with Conexus
        run: |
          curl -fsSL https://raw.githubusercontent.com/ferg-cod3s/conexus/main/setup-conexus.sh | bash
          ./conexus index --quiet

      - name: Run tests
        run: go test -v ./...

      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
```

This integration allows Conexus to understand Go codebases, standard library usage, framework patterns, and Go-specific development practices for enhanced AI assistance.