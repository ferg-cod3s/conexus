# Changelog

All notable changes to Conexus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- **MCP**: Resolved variable shadowing bug in `handleContextSearch` that was preventing proper context search results
  - Location: `internal/mcp/handlers.go`, line 162
  - Changed from `:=` (new local variable) to separate `var` declaration + assignment (uses outer scope)
  - Impact: MCP context search functionality now working correctly
  - Tests: All 57 MCP tests passing ✓

### Removed
- **Federation**: Removed incomplete GitHub connector support code (`internal/federation/` directory)
  - Reason: Code was incomplete, had unresolved import dependencies, and was causing test failures
  - Cleanup: Fully reverted via `git revert 5f051dc`
  - Impact: Test suite now clean (25/27 packages PASS)

### Security
- **Sentry**: Updated test configuration to include all required Server config fields
  - Ensures proper test isolation and prevents configuration errors

## Previous Releases

### Documentation
For detailed historical session documentation, see:
- `docs/operations/` - Operational guides and procedures
- `docs/architecture/` - Architecture decisions and diagrams
- `docs/contributing/` - Contribution guidelines

## Known Issues

### Pre-Existing Test Failures (Not Blocking)
- `tests/debug` - Multiple main functions issue (test setup limitation)
- `internal/testing/integration` - Pre-existing integration test infrastructure issues

These failures are unrelated to current development and do not affect production functionality.

## Test Suite Status

### Current: 25/27 Packages PASS ✓

**Passing Packages:**
- Core agent system: `internal/agent/analyzer`, `internal/agent/locator`
- Context and orchestration: `internal/mcp` (57 tests), `internal/orchestrator`
- Data layer: `internal/indexer`, `internal/vectorstore`, `internal/embedding`
- Infrastructure: `internal/config`, `internal/connectors`, `internal/observability`
- Validation and security: `internal/validation`, `internal/security`, `internal/tool`
- Search, protocol, and process management
- Schema definitions: `pkg/schema`

**Failing Packages (Pre-Existing):**
- `tests/debug` - setup failed
- `internal/testing/integration` - infrastructure issues

**No Test Files:**
- `cmd/conexus` - main binary
- `tests/fixtures` - test utilities

## Development Guidelines

### Code Style
- See `.claude-mcp/CLAUDE.md` for Go coding conventions
- Use testify assertions for testing
- Maintain 80-90% test coverage

### Testing
```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/agent/analyzer

# With verbose output
go test -v ./...
```

### Git Workflow
1. Create feature branch from main
2. Make commits with clear, descriptive messages
3. Ensure all tests pass: `go test ./...`
4. Push to remote and create pull request
5. Merge after review and CI checks pass

## Contributing

See `docs/contributing/contributing-guide.md` for detailed contribution guidelines.

## License

See LICENSE file for license information.
