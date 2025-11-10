# CLAUDE.md - AI Assistant Development Guide

## 🤖 Welcome AI Assistants

This guide provides context for AI assistants (like Claude, GPT, etc.) working with the Conexus codebase. It serves as the central hub for development guidelines and decision-making frameworks.

---

## 📚 Essential Documentation

### 🔧 Development Guidelines
- **[@AGENTS.md](./AGENTS.md)** - Complete development guide for AI agents
  - Build/lint/test commands
  - Code style guidelines
  - Project structure notes
  - Testing patterns
  - **[@AGENTS.md#task-management](./AGENTS.md#-task-management)** - Task management and TODO tracking

### 📋 Task Management
- **[@TODO.md](./TODO.md)** - Current project TODO list and phase tracking
  - **CRITICAL**: Update immediately when completing tasks
  - Mark tasks as ✅ COMPLETE with completion date
  - Update metrics and phase status
  - See [@AGENTS.md#task-management](./AGENTS.md#-task-management) for detailed workflow

### 📋 Version Management
- **[Versioning Criteria](./docs/VERSIONING_CRITERIA.md)** - When and how to bump versions
  - Patch vs Minor vs Major release criteria
  - Decision matrix with examples
  - Release process and checklists
  - Current roadmap targets

### 🤝 Contributing
- **[Contributing Guide](./docs/contributing/contributing-guide.md)** - How to contribute
  - Development workflow
  - Code review process
  - Testing requirements

---

## 🎯 Quick Reference for AI Assistants

### When Making Changes
1. **Check Version Impact**: Reference [@docs/VERSIONING_CRITERIA.md](docs/VERSIONING_CRITERIA.md)
2. **Follow Code Style**: Use [@AGENTS.md](./AGENTS.md) guidelines
3. **Test Thoroughly**: Run `go test ./...` before committing
4. **Update TODO**: Mark completed tasks in [@TODO.md](./TODO.md) immediately
5. **Document Changes**: Update relevant documentation

### Common Tasks
```bash
# Build and test
go build -o conexus ./cmd/conexus
go test ./...

# Check code style
golangci-lint run

# Run specific tests
go test -v ./internal/testing/integration
```

### Version Decision Flow
```
Is this a bug fix? → Patch (0.1.x)
Is it a significant feature? → Minor (0.2.0)  
Is it breaking/enterprise? → Major (1.0.0)
```

---

## 🔄 Current Development Status

### Active Version: `0.2.1-alpha`

**Recent Changes**:
- ✅ Fixed MCP compliance (dot notation)
- ✅ Added `context.explain` and `context.grep` tools
- ✅ Multi-agent architecture implementation
- ✅ Advanced search with code relationships
- ✅ Enhanced connector management
- ✅ Improved test coverage

**Next Target**: `0.3.0-alpha`
- Enterprise security and compliance features
- Multi-tenant support
- Cloud deployment capabilities

---

## 🎨 Development Philosophy

### Principles
1. **Semantic Versioning**: Follow strict semver principles
2. **Documentation-First**: Document decisions before implementing
3. **Test-Driven**: Maintain 100% test pass rate
4. **Security-Conscious**: Zero tolerance for security issues
5. **Performance-Aware**: Monitor and optimize continuously

### AI Assistant Guidelines
- **Evidence-Based**: Back all claims with code/tests/docs
- **Conservative Versioning**: When in doubt, use patch
- **Clear Communication**: Explain reasoning for all changes
- **Cross-Reference**: Link to relevant documentation
- **Task Tracking**: Update [@TODO.md](./TODO.md) immediately when tasks complete
  - See [@AGENTS.md#task-management](./AGENTS.md#-task-management) for workflow

---

## 🔗 Quick Links

| Topic | Location | Purpose |
|--------|----------|---------|
| Development Commands | [@AGENTS.md](./AGENTS.md) | Build, test, lint |
| Task Management | [@AGENTS.md#task-management](./AGENTS.md#-task-management) | TODO tracking workflow |
| Current Tasks | [@TODO.md](./TODO.md) | Active TODO list |
| Version Decisions | [@docs/VERSIONING_CRITERIA.md](./docs/VERSIONING_CRITERIA.md) | When to bump versions |
| Contributing | [@docs/contributing/contributing-guide.md](./docs/contributing/contributing-guide.md) | How to contribute |
| Testing | [@docs/contributing/testing-strategy.md](./docs/contributing/testing-strategy.md) | Testing requirements |
| Architecture | [@docs/Technical-Architecture.md](./docs/Technical-Architecture.md) | System design |
| API Reference | [@docs/api-reference.md](./docs/api-reference.md) | Complete API docs |

---

## 💡 Tips for AI Assistants

### Before Making Changes
1. **Check TODO**: Review [@TODO.md](./TODO.md) for current tasks and priorities
2. **Search First**: Use `git grep` and existing tools
3. **Read Context**: Check related files and documentation
4. **Understand Impact**: Consider version implications
5. **Test Locally**: Verify changes work

### After Completing Tasks
1. **Mark Complete**: Update [@TODO.md](./TODO.md) with ✅ COMPLETE status
2. **Add Completion Date**: Include the date task was finished
3. **Update Metrics**: Refresh test counts, coverage percentages
4. **Commit TODO**: Always commit TODO.md updates with your changes
5. **See Full Workflow**: [@AGENTS.md#task-management](./AGENTS.md#-task-management)

### When Proposing Versions
- **Reference Criteria**: Link to specific versioning rules
- **Provide Evidence**: Show what changed and why
- **Consider Users**: Think about migration impact
- **Document Everything**: Update all relevant docs

### Code Review Best Practices
- **Check Version Impact**: Does this warrant a version bump?
- **Verify Tests**: All tests passing?
- **Review Documentation**: Is everything updated?
- **Security Check**: Any vulnerabilities introduced?

---

## 📞 Getting Help

### For AI Assistants
- **Context**: Always reference this document
- **Questions**: Ask for clarification on version decisions
- **Guidance**: Use the criteria matrix for decisions

### For Human Developers
- **Issues**: [GitHub Issues](https://github.com/ferg-cod3s/conexus/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ferg-cod3s/conexus/discussions)
- **Documentation**: Check all linked docs above

---

**Last Updated**: 2025-10-26  
**Maintained By**: Conexus development team and AI assistants  
**Purpose**: Central hub for AI development guidance