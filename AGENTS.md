# Agent Guidelines for Conexus

## Build & Test Commands
- **Run all tests**: `go test ./...`
- **Run single test**: `go test -run TestName ./path/to/package`
- **Run specific package**: `go test ./internal/agent/analyzer`
- **Run with verbose**: `go test -v ./...`
- **Build**: `go build ./cmd/conexus`
- **Note**: This is primarily a Go project (Go 1.23.4). The README mentions Bun/TypeScript, but all code is in Go.

## Code Style & Conventions

### Imports
- Standard library, then blank line, then external packages, then internal packages
- Example: `import "context"` → `import "github.com/stretchr/testify/assert"` → `import "github.com/ferg-cod3s/conexus/pkg/schema"`
- Use `testify/assert` and `testify/require` for testing

### Naming & Types
- Use camelCase for private, PascalCase for exported
- Interface naming: `Agent`, `Tool`, `Validator` (no "I" prefix)
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Always use `context.Context` as first parameter for functions that may block

### Error Handling
- Return errors, don't panic (except in initialization)
- Use structured errors with `schema.AgentError` for agent responses
- Validate inputs early and return descriptive errors

### Testing
- Follow Arrange-Act-Assert pattern
- Table-driven tests for multiple cases
- Target 80-90% coverage
- Tests in `*_test.go` files alongside implementation
