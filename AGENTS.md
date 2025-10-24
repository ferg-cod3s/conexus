# AGENTS.md - Development Guide for AI Agents

## Build/Lint/Test Commands

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

## Code Style Guidelines

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

## Project Structure Notes
- Main entry point: `cmd/conexus/`
- Core logic: `internal/` (private packages)
- Public APIs: `pkg/` (public packages)
- Tests: Co-located with source files (`*_test.go`)
- Integration tests: `internal/testing/integration/`

## Key Dependencies
- testify: Testing framework and assertions
- SQLite: Embedded database (via modernc.org/sqlite)
- Standard library: Prefer stdlib over external dependencies