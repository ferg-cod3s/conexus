# Conexus Project Documentation

This directory contains comprehensive documentation for the **Conexus (Agentic Context Engine)** project - an open-source RAG system for AI coding assistants.

[![License: FSL-1.1-MIT](https://img.shields.io/badge/License-FSL--1.1--MIT-blue.svg)](https://fsl.software/)

**License**: [FSL-1.1-MIT](https://fsl.software/) - Converts to MIT after 2 years. Prevents direct competition while enabling broad usage.

## 📚 Core Documentation

### Strategic Documents
- **[PRD.md](./PRD.md)** - Product Requirements Document
  - Vision, guiding principles, and product strategy
  - Problem definition and solution architecture
  - Success metrics and non-goals
  - Branding options

### Technical Documentation
- **[Technical-Architecture.md](./Technical-Architecture.md)** - System Architecture
  - Component architecture and data flow
  - Security, observability, and error handling
  - Performance benchmarks and CI/CD strategy
  
- **[API-Specification.md](./API-Specification.md)** - API Reference
  - Complete REST endpoint documentation
  - OpenAPI 3.0 specification
  - Authentication and rate limiting
  - Error codes and response schemas

- **[Security-Compliance.md](./Security-Compliance.md)** - Security Framework
  - STRIDE threat modeling
  - PASTA risk analysis
  - Regulatory compliance (GDPR, HIPAA, SOC 2)
  - Input validation and secrets management

### Business Documentation
- **[Go-to-Market-Strategy.md](./Go-to-Market-Strategy.md)** - GTM Strategy
  - Market sizing and competitive analysis
  - Phase-specific KPIs and success criteria
  - Risk register and resource requirements
  - 3-year plan targeting $300M SOM

- **[Development-Roadmap.md](./Development-Roadmap.md)** - Implementation Plan
  - 18-20 month development timeline
  - Team sizing and resource allocation
  - Epic breakdowns with acceptance criteria
  - Risk tracking and go/no-go gates

## 🚀 Getting Started

### For Developers
- **[Developer Onboarding](./getting-started/developer-onboarding.md)** - Setup guide for new developers
  - Environment setup and prerequisites
  - Local development workflow
  - Testing and debugging
  
- **[MCP Integration Guide](./getting-started/mcp-integration-guide.md)** - Claude Desktop Integration ⭐ NEW
  - Complete MCP tool reference (4 tools)
  - Claude Desktop configuration
  - Protocol specification and examples
  - Integration with TypeScript/Python/cURL

### For Operations
- **[Operations Guide](./operations/operations-guide.md)** - Deployment and operations
  - Docker deployment
  - Configuration management
  - Backup and recovery

- **[Monitoring Guide](./operations/monitoring-guide.md)** - Observability Stack ⭐ NEW
  - Metrics catalog (25+ metrics)
  - Grafana dashboards and alerting
  - Distributed tracing with Jaeger
  - Performance tuning and runbooks

- **[Observability](./operations/observability.md)** - Detailed observability setup
  - Prometheus configuration
  - Logging best practices
  - Troubleshooting guide

## 📊 Documentation Statistics

| Document | Size | Lines | Status |
|----------|------|-------|--------|
| **Strategic** |
| PRD | 12KB | 186 | ✅ Complete |
| Go-to-Market Strategy | 16KB | 230 | ✅ Enhanced |
| Development Roadmap | 12KB | 319 | ✅ Enhanced |
| **Technical** |
| Technical Architecture | 16KB | 318 | ✅ Enhanced |
| API Specification | 24KB | 881 | ✅ Enhanced |
| Security & Compliance | 20KB | 227 | ✅ Enhanced |
| **Guides - Getting Started** |
| Developer Onboarding | 8KB | 150 | ✅ Complete |
| MCP Integration Guide | 28KB | 575 | ✅ NEW |
| **Guides - Operations** |
| Operations Guide | 12KB | 220 | ✅ Complete |
| Monitoring Guide | 35KB | 721 | ✅ NEW |
| Observability | 10KB | 180 | ✅ Complete |
| **Total** | **193KB** | **4,007** | **100%** |

## 🎯 Quick Navigation

**For Developers:**
- Start with [PRD.md](./PRD.md) for product vision
- Review [Technical-Architecture.md](./Technical-Architecture.md) for system design
- Follow [Developer Onboarding](./getting-started/developer-onboarding.md) for setup
- Integrate with [MCP Integration Guide](./getting-started/mcp-integration-guide.md)
- Reference [API-Specification.md](./API-Specification.md) for implementation

**For Business Stakeholders:**
- Review [Go-to-Market-Strategy.md](./Go-to-Market-Strategy.md) for market strategy
- Check [Development-Roadmap.md](./Development-Roadmap.md) for timeline
- Examine [Security-Compliance.md](./Security-Compliance.md) for compliance

**For Security/Compliance:**
- Focus on [Security-Compliance.md](./Security-Compliance.md)
- Review security sections in [Technical-Architecture.md](./Technical-Architecture.md)
- Check authentication in [API-Specification.md](./API-Specification.md)

**For DevOps/SRE:**
- Deploy using [Operations Guide](./operations/operations-guide.md)
- Monitor with [Monitoring Guide](./operations/monitoring-guide.md)
- Configure observability with [Observability](./operations/observability.md)

## 🔄 Document Relationships

```
PRD (What & Why)
  ↓
Technical Architecture (How to Build)
  ↓
Developer Onboarding (How to Start) ← NEW
  ↓
MCP Integration Guide (How to Integrate) ← NEW
  ↓
API Specification (How to Use)
  ↓
Security & Compliance (How to Secure)
  ↓
Operations Guide (How to Deploy)
  ↓
Monitoring Guide (How to Monitor) ← NEW
  ↓
Development Roadmap (When to Build)
  ↓
Go-to-Market Strategy (How to Sell)
```

## 📋 Key Highlights

### Technical Specifications
- **Performance:** <1s query latency, 100+ concurrent users
- **Stack:** Go, PostgreSQL/SQLite, Qdrant, hybrid RAG
- **APIs:** MCP + REST with OpenAPI 3.0
- **Test Coverage:** 80-90% unit tests (218/218 passing)
- **Security:** SOC 2 compliant, STRIDE threat modeled, 0 security issues

### Business Metrics
- **TAM:** $15B (AI coding assistance market)
- **SAM:** $3B (RAG-based solutions)
- **SOM:** $300M in 3 years
- **Timeline:** 18-20 months to GA
- **Funding:** ~$50M over 3 years
- **Team:** 6-8 engineers initially → 66 headcount at scale

### Go-to-Market Phases
1. **Phase 1 (0-6m):** Community building - 750+ GitHub stars
2. **Phase 2 (6-12m):** Enterprise seeding - $250K ARR, 8+ customers
3. **Phase 3 (12-36m):** Commercial scale - $20M+ ARR, 400+ customers

## 🚀 Quick Start Path

### For First-Time Users:
1. **Understand the Vision:** Read [PRD.md](./PRD.md)
2. **Setup Development:** Follow [Developer Onboarding](./getting-started/developer-onboarding.md)
3. **Integrate with Claude:** Configure [MCP Integration](./getting-started/mcp-integration-guide.md)
4. **Deploy & Monitor:** Use [Operations](./operations/operations-guide.md) + [Monitoring](./operations/monitoring-guide.md)

### For Existing Users:
- **MCP Tools:** See [MCP Integration Guide](./getting-started/mcp-integration-guide.md) for 4 available tools
- **Monitoring:** Check [Monitoring Guide](./operations/monitoring-guide.md) for 25+ metrics and alerting
- **API Reference:** Browse [API Specification](./API-Specification.md) for endpoints

## 📝 Documentation Standards

All documents follow these standards:
- ✅ Quantitative metrics (no vague goals)
- ✅ Actionable implementation details
- ✅ Risk mitigation strategies
- ✅ Clear ownership and accountability
- ✅ Consistent terminology
- ✅ Professional formatting

## 🔐 Security Notice

This documentation includes security architecture details. While public, ensure proper access controls for any derived implementation artifacts, credentials, or deployment configurations.

## 📞 Contact & Contribution

For questions, improvements, or contributions to this documentation:
- Open an issue in the main repository
- Follow contribution guidelines in the root README
- Ensure consistency with existing document structure

---

**Last Updated:** October 16, 2025  
**Status:** Documentation Complete ✅  
**Version:** 1.1 (Enterprise-grade + Operations Guides)

