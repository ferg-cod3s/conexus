# Operations Documentation

Welcome to the Conexus Operations section. This directory contains all documentation needed to maintain, deploy, and troubleshoot the Conexus project.

## Quick Links

### 📋 Essential Documents

- **[CHANGELOG.md](./CHANGELOG.md)** - Version history and all changes
  - Breaking changes
  - New features
  - Bug fixes
  - Security updates

- **[RELEASE_NOTES.md](./RELEASE_NOTES.md)** - Current release information
  - Latest fixes and improvements
  - Known limitations
  - Deployment instructions
  - Rollback procedures

- **[MAINTENANCE_GUIDE.md](./MAINTENANCE_GUIDE.md)** - Day-to-day operations
  - Build and deploy procedures
  - Troubleshooting guide
  - Performance optimization
  - Monitoring setup

### 📊 Project Status

**Latest Release**: Commit `c6aba02`  
**Test Coverage**: 25/27 packages PASS ✓  
**Status**: ✅ Production Ready

### 🔧 Quick Commands

```bash
# Build
go build ./cmd/conexus

# Test
go test ./...

# Deploy
go build -o conexus ./cmd/conexus
./conexus --config config.yml

# Troubleshoot
go test ./internal/mcp -v
```

## Documentation Structure

```
docs/
├── operations/              ← You are here
│   ├── README.md           # This file
│   ├── CHANGELOG.md        # Version history
│   ├── RELEASE_NOTES.md    # Current release info
│   └── MAINTENANCE_GUIDE.md # Operations procedures
├── architecture/           # Technical architecture
├── contributing/           # Contribution guidelines
├── getting-started/        # Developer onboarding
├── research/              # Research documents
└── README.md              # Main documentation
```

## Key Information

### Current Release Status

✅ **All Core Systems Operational**
- MCP context search: Fixed ✓
- Federation code: Cleaned up ✓
- Sentry integration: Fixed ✓
- Test suite: 25/27 packages passing ✓

### Known Issues

| Issue | Severity | Status | Impact |
|-------|----------|--------|--------|
| `tests/debug` multiple main functions | Low | Pre-existing | Test infrastructure only |
| `internal/testing/integration` failures | Low | Pre-existing | Unit tests comprehensive |

### Recent Changes

1. **MCP Variable Shadowing Fix** (commit `dd6afda`)
   - Fixed context search result handling
   - All 57 MCP tests now passing

2. **Federation Cleanup** (commit `5c82e71`)
   - Removed incomplete GitHub connector code
   - Test suite now clean

3. **Sentry Configuration Fix** (commit `c175161`)
   - Complete test configuration
   - Proper test isolation

## Getting Help

### For Developers
1. See `docs/getting-started/developer-onboarding.md`
2. Review `.claude-mcp/CLAUDE.md` for code style
3. Check `docs/contributing/contributing-guide.md`

### For Operations
1. Start with [MAINTENANCE_GUIDE.md](./MAINTENANCE_GUIDE.md)
2. Reference [RELEASE_NOTES.md](./RELEASE_NOTES.md)
3. Consult [CHANGELOG.md](./CHANGELOG.md) for version info

### For Troubleshooting
1. Check [MAINTENANCE_GUIDE.md - Troubleshooting](./MAINTENANCE_GUIDE.md#troubleshooting)
2. Review relevant service logs
3. Run diagnostic tests: `go test -v ./...`

## Deployment Checklist

Before deploying to production:

- [ ] All tests passing: `go test ./...`
- [ ] CHANGELOG.md updated
- [ ] Version tagged: `git tag v1.x.x`
- [ ] Release notes written
- [ ] Configuration reviewed
- [ ] Database backups ready
- [ ] Rollback plan documented

See [MAINTENANCE_GUIDE.md - Deployment](./MAINTENANCE_GUIDE.md#deployment-process) for detailed steps.

## Monitoring and Alerting

### Key Metrics
- MCP request latency
- Context search success rate
- Vector store performance
- Agent orchestration timing

### Log Files
- Application logs: Check configured log path
- System logs: `/var/log/`
- Prometheus metrics: `/metrics` endpoint

See [MAINTENANCE_GUIDE.md - Monitoring](./MAINTENANCE_GUIDE.md#monitoring-and-logging) for setup.

## Version Information

### Current Version
- **Latest Commit**: `c6aba02`
- **Branch**: main
- **Status**: Up to date with origin/main

### Stable Versions
- `dd6afda` - MCP fix (verified working) ✓
- `c175161` - Sentry config fix ✓
- `7cb0da1` - Earlier stable point ✓

## Release Process

### Creating a Release
1. Update [CHANGELOG.md](./CHANGELOG.md)
2. Update [RELEASE_NOTES.md](./RELEASE_NOTES.md)
3. Tag commit: `git tag -a v1.x.x`
4. Push tags: `git push origin v1.x.x`

See [MAINTENANCE_GUIDE.md - Release Management](./MAINTENANCE_GUIDE.md#release-management) for detailed instructions.

## Contact and Support

**For Issues**:
1. Check this documentation first
2. Review GitHub issues
3. Create detailed issue report if needed

**Include in Reports**:
- Steps to reproduce
- Expected vs actual behavior
- Error logs and stack traces
- Environment details

## References

- **Main README**: `docs/README.md`
- **Architecture**: `docs/architecture/`
- **Contributing**: `docs/contributing/contributing-guide.md`
- **Development**: `.claude-mcp/CLAUDE.md`
- **Roadmap**: `docs/Development-Roadmap.md`

---

**Last Updated**: October 18, 2025  
**Maintained By**: Conexus Team  
**Status**: ✅ Current and Complete
