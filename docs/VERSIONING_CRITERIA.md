# Versioning Criteria - Conexus

## Overview

This document defines clear criteria for version bumps following [Semantic Versioning](https://semver.org/) principles. All version decisions should reference this document to ensure consistency and proper communication with users.

---

## 🏷️ Version Strategy

### Current Version: `0.1.2-alpha`

**Pre-release Status**: All versions before 1.0.0 are alpha/beta releases for testing and feedback collection.

---

## 📋 Version Bump Criteria

### 🩹 Patch Release (0.1.x → 0.1.2)

**Definition**: Bug fixes and small features that don't break existing functionality.

**When to Use Patch**:
- ✅ **Critical Bug Fixes**: MCP compliance violations, security issues, crashes
- ✅ **Small Features**: 1-2 new MCP tools, minor enhancements
- ✅ **Performance Improvements**: Optimizations that don't change APIs
- ✅ **Documentation**: Major documentation updates
- ✅ **Test Improvements**: Better test coverage, CI/CD fixes

**Examples**:
- Fix MCP tool naming convention (underscore → dot)
- Add `context.explain` and `context.grep` tools
- Fix memory leaks or performance regressions
- Update integration guides

**Process**: 
- Increment patch number: `0.1.1` → `0.1.2`
- Maintain `-alpha` pre-release suffix
- No breaking changes expected

---

### ⬆️ Minor Release (0.1.x → 0.2.0)

**Definition**: Significant new functionality that adds value but maintains backward compatibility.

**When to Use Minor**:

#### 🎯 Core MCP Enhancements
- ✅ **Multi-Agent Architecture**: Specialized agents for complex analysis
- ✅ **Advanced Search**: Code relationships, dependency mapping, semantic chunking
- ✅ **Enhanced Connectors**: GitHub, Jira, Slack integrations with full CRUD
- ✅ **Workflow Orchestration**: Multi-step analysis pipelines
- ✅ **Evidence Validation**: Complete traceability for all results

#### 🔧 Platform Features
- ✅ **Real-time Updates**: Live indexing, file watching
- ✅ **Advanced Filtering**: Date ranges, source types, custom filters
- ✅ **Performance Optimization**: Vector search improvements, caching layers
- ✅ **Configuration Management**: Environment-based config, validation
- ✅ **Monitoring & Observability**: Metrics, tracing, alerting

#### 📚 Developer Experience
- ✅ **SDK Development**: TypeScript/Python client libraries
- ✅ **CLI Enhancements**: Better commands, interactive mode
- ✅ **Plugin System**: Extensible architecture for custom tools
- ✅ **Testing Framework**: Built-in testing tools and validation

**Examples**:
- Implement multi-agent system with specialized analysis agents
- Add GitHub connector with full sync capabilities
- Implement real-time file watching and incremental indexing
- Add comprehensive workflow orchestration engine
- Release TypeScript SDK for client integration

**Process**:
- Increment minor number: `0.1.2` → `0.2.0`
- Reset patch to 0
- Maintain `-alpha` until stable
- Backward compatibility maintained

---

### 🚀 Major Release (0.x.x → 1.0.0)

**Definition**: Breaking changes or production-ready milestone with enterprise features.

**When to Use Major**:

#### 🏢 Production Readiness (v1.0.0)
- ✅ **Enterprise Security**: RBAC, audit logging, compliance (SOC 2, GDPR)
- ✅ **Multi-tenancy**: Team workspaces, isolation, collaboration
- ✅ **Scalability**: Distributed processing, cloud deployment
- ✅ **Reliability**: 99.9% uptime, disaster recovery, backups
- ✅ **Enterprise Integrations**: SSO, LDAP, SAML, enterprise systems

#### 🔧 Breaking Changes
- ✅ **API Changes**: Modify existing MCP tool interfaces
- ✅ **Protocol Changes**: Update MCP version, breaking compatibility
- ✅ **Database Schema**: Major schema changes requiring migration
- ✅ **Configuration**: Changes that break existing configs
- ✅ **Platform Support**: Drop support for older OS/versions

#### 🏗️ Architecture Evolution
- ✅ **Microservices**: Split monolith into distributed services
- ✅ **Cloud Native**: Kubernetes deployment, auto-scaling
- ✅ **Advanced AI**: Multi-model support, custom fine-tuning
- ✅ **Enterprise Features**: Advanced analytics, reporting, governance

**Examples**:
- Full enterprise-ready platform with RBAC and multi-tenancy
- Breaking change to MCP protocol (new version)
- Major database schema redesign
- Cloud-native architecture with Kubernetes deployment
- Advanced AI features with custom model fine-tuning

**Process**:
- Increment major number: `0.2.0` → `1.0.0`
- Reset minor and patch to 0
- Remove `-alpha` suffix for stable release
- Comprehensive migration guide required
- Extended deprecation period for breaking changes

---

## 📊 Version Decision Matrix

| Scenario | Version | Rationale |
|-----------|----------|-----------|
| Bug fix (MCP compliance) | 0.1.2-alpha | Critical fix, no new features |
| Add 2-3 MCP tools | 0.1.3-alpha | Small feature addition |
| Implement multi-agent system | 0.2.0-alpha | Significant new functionality |
| Add GitHub connector | 0.2.0-alpha | Major integration capability |
| Real-time indexing | 0.2.0-alpha | Performance/architecture improvement |
| Enterprise security features | 1.0.0 | Production-ready milestone |
| Breaking API changes | 1.0.0 | Requires major version bump |
| Multi-tenant architecture | 1.0.0 | Enterprise feature set |

---

## 🎯 Current Roadmap Targets

### v0.1.2-alpha (Current)
- ✅ MCP compliance fix
- ✅ Add `context.explain` and `context.grep` tools
- ✅ Test suite improvements

### v0.2.0-alpha (Next Minor)
- 🔄 Multi-agent architecture implementation
- 🔄 Advanced search with code relationships
- 🔄 Enhanced connector management
- 🔄 Real-time indexing capabilities

### v1.0.0 (Production Ready)
- 📋 Enterprise security and compliance
- 📋 Multi-tenant support
- 📋 Cloud deployment capabilities
- 📋 Advanced monitoring and observability

---

## 🔄 Version Process

### Pre-Release (Alpha/Beta)
1. **Development**: Feature implementation on feature branches
2. **Testing**: Comprehensive test suite, integration tests
3. **Documentation**: Update all relevant documentation
4. **Review**: Architecture review, security scan
5. **Release**: Tag with `-alpha` or `-beta` suffix

### Stable Release
1. **Feature Freeze**: No new features in release cycle
2. **Stabilization**: Focus on bugs and performance
3. **Production Testing**: Real-world usage validation
4. **Documentation**: Complete user guides and migration docs
5. **Release**: Remove pre-release suffix

### Release Checklist
- [ ] All tests passing (100% success rate)
- [ ] Security audit completed (0 vulnerabilities)
- [ ] Performance benchmarks meet targets
- [ ] Documentation updated and reviewed
- [ ] Migration guide prepared (for breaking changes)
- [ ] Release notes drafted
- [ ] Version numbers updated consistently
- [ ] Git tag created and pushed
- [ ] npm package published
- [ ] Release announcement prepared

---

## 📝 Version History

### v0.1.2-alpha (Current)
- **Fixed**: MCP tool naming convention compliance
- **Added**: `context.explain` and `context.grep` tools
- **Improved**: Test coverage and integration testing

### v0.1.1-alpha
- **Added**: Basic MCP server functionality
- **Added**: 4 core MCP tools
- **Improved**: Performance optimizations

### v0.1.0-mvp
- **Released**: Production-ready MVP
- **Completed**: Phase 7 deliverables
- **Achieved**: 251/251 tests passing

---

## 🤝 Contributing to Version Decisions

When proposing version changes:

1. **Reference this document** for criteria
2. **Provide clear rationale** with examples
3. **Consider user impact** and migration needs
4. **Document breaking changes** thoroughly
5. **Follow semantic versioning** principles

**Version Decision Template**:
```
Proposed Version: 0.x.x-alpha
Justification: [Reference specific criteria from this document]
Changes: [List of changes included]
Impact: [User impact and migration requirements]
Timeline: [Expected release date]
```

---

## 🔗 Related Documentation

- **[AGENTS.md](../AGENTS.md)** - Development guide for AI agents
- **[Contributing Guide](contributing/contributing-guide.md)** - How to contribute to Conexus
- **[Testing Strategy](contributing/testing-strategy.md)** - Testing requirements and best practices

---

**Last Updated**: 2025-10-26  
**Next Review**: Before each version bump decision  
**Maintainers**: Conexus development team