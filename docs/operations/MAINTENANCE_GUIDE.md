# Maintenance Guide

This guide documents maintenance procedures, deployment processes, and troubleshooting for the Conexus project.

## Quick Reference

- **Latest Stable**: Commit `d178de6` (main branch)
- **Test Status**: 25/27 packages PASS ✓
- **Build Command**: `go build ./cmd/conexus`
- **Test Command**: `go test ./...`

## Current Status

### Production Ready ✓
- MCP variable shadowing bug **FIXED**
- Federation incomplete code **REMOVED**
- Sentry configuration **FIXED**
- All core tests **PASSING**

### Known Pre-Existing Issues
1. `tests/debug` - Multiple main functions (not blocking)
2. `internal/testing/integration` - Infrastructure issues (not blocking)

These do not affect production deployments.

## Release Management

### New Releases

1. **Create a version tag**:
   ```bash
   git tag -a v1.x.x -m "Release v1.x.x"
   git push origin v1.x.x
   ```

2. **Update CHANGELOG.md**:
   - Document changes under `[Unreleased]` section
   - Create new version section with date

3. **Update RELEASE_NOTES.md**:
   - Add version number and date
   - Document bug fixes and new features
   - Note any breaking changes

4. **Create GitHub Release**:
   - Use `gh release create v1.x.x --generate-notes`

### Hotfixes

For urgent patches:

1. Create branch from main:
   ```bash
   git checkout -b hotfix/issue-description main
   ```

2. Apply fix and test:
   ```bash
   go test ./...
   ```

3. Commit with clear message:
   ```bash
   git commit -m "fix(package): issue description"
   ```

4. Merge to main:
   ```bash
   git checkout main
   git merge --no-ff hotfix/issue-description
   ```

5. Tag and push:
   ```bash
   git tag -a v1.x.x
   git push origin main
   git push origin v1.x.x
   ```

## Deployment Process

### Pre-Deployment Checklist

- [ ] All tests passing: `go test ./...`
- [ ] Code review completed
- [ ] CHANGELOG.md updated
- [ ] No uncommitted changes
- [ ] Remote branch up to date

### Building

```bash
# Standard build
go build ./cmd/conexus

# With version information
go build -ldflags="-X main.Version=v1.x.x" ./cmd/conexus

# Docker build
docker build -t conexus:latest .
```

### Testing Before Deployment

```bash
# Full test suite
go test ./...

# Specific package
go test ./internal/mcp -v

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Deployment Steps

1. **Build application**
   ```bash
   go build -o /path/to/conexus ./cmd/conexus
   ```

2. **Configure environment**
   ```bash
   cp config.example.yml config.yml
   # Edit config.yml with deployment settings
   ```

3. **Run application**
   ```bash
   ./conexus --config config.yml
   ```

4. **Verify functionality**
   - Check logs for errors
   - Verify MCP endpoints responding
   - Test context search functionality

## Monitoring and Logging

### Log Levels

Configure in `config.yml`:
```yaml
logging:
  level: info  # debug, info, warn, error
  format: json # json or text
```

### Key Metrics to Monitor

- MCP request latency
- Context search success rate
- Vector store query performance
- Agent orchestration timing

### Prometheus Integration

Metrics available at `/metrics` endpoint:
```bash
curl http://localhost:8080/metrics
```

See `docs/operations/monitoring-guide.md` for detailed setup.

## Troubleshooting

### MCP Context Search Not Working

**Symptoms**: Search queries returning empty results

**Diagnosis**:
```bash
# Check MCP service logs
grep -i "search\|context" /var/log/conexus/app.log

# Run tests
go test ./internal/mcp -v -run TestContextSearch
```

**Solution**:
- Verify vector store initialized: Check `internal/vectorstore/sqlite`
- Confirm embeddings generating: Check embedding service logs
- Check for variable shadowing issues in handlers (fixed in commit `dd6afda`)

### Test Failures

**For `tests/debug` failures**:
- These are pre-existing infrastructure issues
- Do not affect production
- Use standard Go testing approach instead

**For `internal/testing/integration` failures**:
- Pre-existing integration test issues
- Unit tests are comprehensive
- Integration test suite needs refactoring (future work)

**For other test failures**:
```bash
# Run specific test with verbose output
go test -v -run TestName ./path

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

## Rollback Procedures

If critical issues discovered after deployment:

### Quick Rollback

```bash
# View recent commits
git log --oneline -10

# Revert to previous version
git revert HEAD
git push origin main

# Rebuild and redeploy
go build ./cmd/conexus
```

### Safe Rollback Points

- `d178de6` - Latest (documentation cleanup)
- `5c82e71` - Federation cleanup
- `dd6afda` - MCP fix (verified working)
- `c175161` - Sentry config fix

### Database Rollback

If database schema issues:

```bash
# Backup current database
cp conexus.db conexus.db.backup

# Restore from backup
cp conexus.db.backup conexus.db

# Restart application
```

## Performance Optimization

### Profiling

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof ./internal/orchestrator

# Analyze profile
go tool pprof cpu.prof

# View memory profile
go test -memprofile=mem.prof ./internal/vectorstore
go tool pprof mem.prof
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/indexer

# Compare benchmarks
go test -bench=. -benchmem ./internal/orchestrator > new.txt
benchstat old.txt new.txt
```

### Common Optimizations

1. **Vector Search Performance**
   - Check indexer configuration in `internal/indexer/`
   - Verify vector dimensions match embedding model
   - Monitor query complexity

2. **MCP Response Time**
   - Check context search limits in handlers
   - Profile context gathering overhead
   - Monitor external API calls

3. **Memory Usage**
   - Monitor vector cache size
   - Check for memory leaks in process manager
   - Profile long-running operations

## Documentation

### Keeping Documentation Current

1. **Code Changes**: Update relevant docs in `docs/`
2. **Configuration**: Update `docs/getting-started/`
3. **Architecture**: Update `docs/architecture/`
4. **Release**: Update `docs/operations/CHANGELOG.md`

### Documentation Structure

```
docs/
├── architecture/          # Architecture decisions
├── contributing/          # Contribution guidelines
├── getting-started/       # Onboarding docs
├── operations/            # Operational guides
│   ├── CHANGELOG.md      # Version history
│   ├── RELEASE_NOTES.md  # Current release info
│   └── MAINTENANCE_GUIDE.md  # This file
├── research/             # Research documents
└── README.md             # Main documentation
```

## Support and Contact

For maintenance issues:

1. Check this guide first
2. Review `docs/contributing/contributing-guide.md`
3. Check GitHub issues for similar problems
4. Create detailed issue if problem persists

### Reporting Issues

Include:
- Steps to reproduce
- Expected vs actual behavior
- Error logs and stack traces
- Environment details (OS, Go version, etc.)

## References

- Main README: `docs/README.md`
- Architecture: `docs/architecture/`
- Contributing: `docs/contributing/contributing-guide.md`
- Development: `.claude-mcp/CLAUDE.md`
- Roadmap: `docs/Development-Roadmap.md`

---

**Last Updated**: October 18, 2025  
**Status**: ✅ Current and Maintained
