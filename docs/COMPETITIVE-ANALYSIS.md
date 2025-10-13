# Conexus Competitive Analysis

**Status**: Market Research  
**Version**: 1.0  
**Last Updated**: 2025-10-12  
**Analyst**: F3RG

## Executive Summary

The AI code context market is rapidly consolidating around **three distinct layers**:

1. **End-User Tools** (GitHub Copilot, Cursor) - $10-50B market, dominated by proprietary freemium
2. **Infrastructure/Platforms** (LangChain, LlamaIndex, Sourcegraph) - $5-20B market, open-core models winning
3. **Specialized Enterprise** (Tabnine, AWS Q Developer) - $3-10B market, security/compliance focus

**Conexus's Opportunity**: Position as **infrastructure for AI coding tools** (Layer 2), not end-user assistant (Layer 1).

### Market Gap Identified

**None of the major players offer:**
- ✅ Specialized context engine as standalone infrastructure
- ✅ Open-core with true permissive conversion timeline
- ✅ Repository-aware context with semantic understanding
- ✅ Privacy-first architecture with full air-gap support
- ✅ Pluggable into any AI coding tool (not locked to one interface)

**Strategic Positioning**: "The Stripe of AI Code Context" - infrastructure that powers other tools.

---

## Market Segmentation

### Layer 1: End-User AI Coding Assistants

**Market Size**: $10-50B (high growth, competitive)  
**Revenue Model**: Freemium SaaS ($10-39/user/month)  
**Competition**: Intense, dominated by well-funded players

| Player | Funding | Market Position | Strength | Weakness |
|--------|---------|----------------|----------|----------|
| **GitHub Copilot** | Microsoft-backed | Market leader | GitHub integration, scale | Generic context |
| **Cursor** | $900M Series C | Fast growth | AI-first UX, Tab model | Proprietary, costly |
| **Continue** | YC-backed | Open-source leader | Model flexibility | Smaller ecosystem |
| **Cody/Amp** | Sourcegraph-backed | Enterprise focus | Code intelligence | Complex pricing |

**Verdict for Conexus**: ❌ **Do NOT compete directly** - too crowded, requires massive marketing/sales investment.

---

### Layer 2: Infrastructure & Platforms (Conexus'S TARGET) ⭐

**Market Size**: $5-20B (growing rapidly)  
**Revenue Model**: Open-core + managed platforms ($500-5000+/mo for teams)  
**Competition**: Moderate, space for differentiation

| Player | Model | Strength | Weakness | Conexus Advantage |
|--------|-------|----------|----------|---------------|
| **LangChain** | Open-core | Agent orchestration, huge ecosystem | Not code-specific, complex | Conexus is code-specialized |
| **LlamaIndex** | Open-core | Document processing, 500M docs | General-purpose RAG | Conexus understands code semantics |
| **Sourcegraph** | Proprietary freemium | Code search, enterprise traction | Expensive, complex | Conexus is pluggable, simpler |
| **Pinecone/Qdrant** | Vector DB | Infrastructure layer | Generic vectors, no code awareness | Conexus adds semantic understanding |

**Verdict for Conexus**: ✅ **COMPETE HERE** - clear differentiation, less crowded, infrastructure model fits.

---

### Layer 3: Specialized Enterprise Tools

**Market Size**: $3-10B (niche but high-margin)  
**Revenue Model**: Enterprise contracts ($50k-500k/year)  
**Competition**: Specialized by vertical/compliance need

| Player | Focus | Strength | Weakness |
|--------|-------|----------|----------|
| **Tabnine** | Privacy/air-gap | Enterprise trust, compliance | Limited innovation |
| **AWS Q Developer** | AWS ecosystem | Deep AWS integration | Locked to AWS |
| **Replit Agent** | Rapid prototyping | No-code to code | Not for production |

**Verdict for Conexus**: ⚠️ **Partnership potential** - Conexus could power their context layer.

---

## Detailed Competitor Profiles

### 1. GitHub Copilot (Microsoft)

**Business Model**: Proprietary freemium  
**License**: Closed-source  
**Pricing**:
- Free: $0 (50 agent chats/mo, 2k completions)
- Pro: $10/mo (unlimited completions, GPT-5 access)
- Pro+: $39/mo (all models including Claude Opus 4.1, o3)
- Business: $19/user/mo (IP indemnity, policy management)
- Enterprise: Custom (GitHub integration, SAML/SSO)

**Key Features**:
- Code completion, chat, agent mode
- Multi-model support (GPT, Claude, Gemini)
- GitHub integration (PR reviews, security scanning)
- Code review automation
- IP indemnity for enterprise

**Market Position**:
- 🏆 Market leader by volume (millions of users)
- Deep integration with GitHub platform
- Strong brand recognition
- Microsoft backing ensures longevity

**Strengths**:
- ✅ Massive distribution via GitHub
- ✅ Multi-model flexibility
- ✅ Enterprise trust (Microsoft backing)
- ✅ Continuous innovation (agent mode, code review)

**Weaknesses**:
- ❌ Generic context (doesn't deeply understand codebases)
- ❌ Cloud-only (privacy concerns for some enterprises)
- ❌ Lock-in to GitHub ecosystem
- ❌ Can be expensive at scale ($39/user for Pro+)

**How Conexus Competes**:
- 🎯 **Don't compete directly** - Conexus should power Copilot's context layer
- 📊 Conexus offers deeper codebase understanding
- 🔒 Conexus supports air-gapped deployments
- 🔌 Conexus is platform-agnostic (not locked to GitHub)

---

### 2. Cursor (Anysphere)

**Business Model**: Proprietary freemium  
**License**: Closed-source  
**Pricing**:
- Free: Limited features
- Pro: $20/mo (full features, unlimited usage)

**Key Features**:
- AI-first IDE (fork of VS Code)
- Custom Tab model for predictions
- Agent mode for complex tasks
- Codebase understanding
- Chat interface integrated into editor

**Market Position**:
- 🚀 Fastest-growing AI coding assistant
- Millions of professional developers
- $900M Series C (2025) at premium valuation
- Research-driven improvements

**Strengths**:
- ✅ Best-in-class UX for AI coding
- ✅ Deep research team improving models
- ✅ Fast iteration and improvement
- ✅ Strong community and advocacy

**Weaknesses**:
- ❌ Proprietary (no self-hosting)
- ❌ Expensive for teams ($20/user vs Copilot $10)
- ❌ Requires custom IDE (not plugin-based)
- ❌ Context limited to visible files + basic RAG

**How Conexus Competes**:
- 🎯 **Partnership opportunity** - Cursor could integrate Conexus for context
- 📊 Conexus provides enterprise-grade context beyond file-level
- 🔒 Conexus supports air-gapped deployments Cursor can't offer
- 🔌 Conexus as infrastructure allows Cursor to focus on UX

---

### 3. Continue.dev (YC)

**Business Model**: Open-core  
**License**: Apache 2.0 (open-source) + commercial tiers  
**Pricing**:
- Solo: $0 (open-source, unlimited)
- Team: $20/user/mo (centralized config, collaboration)
- Enterprise: Custom (SSO, compliance, SLAs)

**Key Features**:
- Open-source IDE extensions (VS Code, JetBrains)
- Model flexibility (any LLM provider)
- MCP tool integration
- Custom rules and agents
- Self-hosted options

**Market Position**:
- 🔓 Leading open-source AI coding assistant
- YC-backed with growing community
- Developer-first approach
- Strong GitHub presence (15k+ stars)

**Strengths**:
- ✅ True open-source core (not source-available)
- ✅ Model flexibility (no vendor lock-in)
- ✅ Self-hosting for privacy
- ✅ Community-driven development

**Weaknesses**:
- ❌ Smaller team vs Copilot/Cursor
- ❌ Less polished UX than Cursor
- ❌ Monetization still evolving
- ❌ Context capabilities basic (file-level RAG)

**How Conexus Competes**:
- 🎯 **Collaboration opportunity** - Continue could integrate Conexus as context engine
- 📊 Conexus provides enterprise-grade context Continue lacks
- 🤝 Both open-source values aligned (FSL → MIT for Conexus)
- 🔌 Conexus complements Continue's model flexibility

**Verdict**: Continue is a **potential partner**, not direct competitor.

---

### 4. Cody / Amp (Sourcegraph)

**Business Model**: Proprietary freemium (evolving to Amp)  
**License**: Closed-source  
**Pricing**:
- Free: Basic features
- Pro: $9/mo (advanced features)
- Enterprise: Custom (code intelligence, security)

**Key Features**:
- Codebase understanding (Sourcegraph integration)
- Code search + AI chat
- Enterprise-scale deployments
- Deep Search (agentic code search)
- Context from millions of LOC

**Market Position**:
- 🏢 Enterprise leader (4/6 top US banks, 15+ gov agencies)
- Deep codebase intelligence
- Trusted for security/compliance
- Transitioning from Cody to Amp (rebranding)

**Strengths**:
- ✅ Best-in-class code intelligence
- ✅ Enterprise trust and adoption
- ✅ Deep codebase understanding
- ✅ Proven at scale (millions of LOC)

**Weaknesses**:
- ❌ Expensive for mid-market
- ❌ Complex setup and maintenance
- ❌ Proprietary (no self-hosting for code search)
- ❌ Locked to Sourcegraph platform

**How Conexus Competes**:
- 🎯 **Conexus as lightweight alternative** - simpler setup, lower cost
- 📊 Conexus offers pluggable context vs Sourcegraph's full platform
- 🔒 Conexus supports true air-gap (Sourcegraph requires cloud components)
- 💰 Conexus open-core model more accessible for mid-market

**Verdict**: Sourcegraph owns **enterprise code intelligence**, Conexus should target **mid-market and infrastructure layer**.

---

### 5. Tabnine

**Business Model**: Proprietary freemium  
**License**: Closed-source  
**Pricing**:
- Free: Basic completions
- Pro: $12/mo (advanced features)
- Enterprise: Custom (air-gap, compliance)

**Key Features**:
- Privacy-first (air-gapped deployments)
- Enterprise governance
- On-premise hosting
- Compliance (SOC2, HIPAA, GDPR)
- Multi-agent capabilities

**Market Position**:
- 🔒 Privacy/security leader
- Gartner top-ranked AI assistant
- Trusted by enterprises with strict compliance
- Focus on financial services, healthcare, government

**Strengths**:
- ✅ True air-gapped deployments
- ✅ Enterprise trust for compliance
- ✅ Proven security model
- ✅ On-premise options

**Weaknesses**:
- ❌ Less innovation than Copilot/Cursor
- ❌ Smaller model selection
- ❌ Proprietary (no open-source option)
- ❌ Limited context capabilities

**How Conexus Competes**:
- 🎯 **Conexus as open alternative to Tabnine** - better context, open-core
- 🔒 Conexus supports air-gap + open-source transparency
- 📊 Conexus offers advanced context beyond Tabnine's completions
- 💰 Conexus lower cost with self-hosting

**Verdict**: Tabnine vulnerable to **open-core alternative with better context**.

---

### 6. Amazon Q Developer (AWS)

**Business Model**: Proprietary freemium  
**License**: Closed-source  
**Pricing**:
- Free: 50 agentic interactions/mo, 1k lines transformation
- Pro: $19/mo (unlimited, advanced features)
- Enterprise: Custom (AWS integration, security)

**Key Features**:
- AWS ecosystem integration
- Agentic transformations (Java upgrades, .NET porting)
- Security scanning (CodeGuru integration)
- Infrastructure as code support
- Multi-language completions

**Market Position**:
- 🏢 AWS-first developers
- Deep AWS Console integration
- Security/compliance built-in
- Part of broader AWS AI strategy

**Strengths**:
- ✅ Deep AWS integration (Lambda, ECS, etc.)
- ✅ Agentic code transformations
- ✅ Security scanning built-in
- ✅ Free tier attractive

**Weaknesses**:
- ❌ Locked to AWS ecosystem
- ❌ Limited outside AWS context
- ❌ Generic AI (not leading edge)
- ❌ Weaker for non-AWS codebases

**How Conexus Competes**:
- 🎯 **Conexus for non-AWS teams** - platform-agnostic
- 📊 Conexus provides better context for multi-cloud
- 🔌 Conexus pluggable into any tool (not AWS-only)

**Verdict**: AWS Q Developer serves **AWS-centric teams**, Conexus targets **broader market**.

---

### 7. LangChain (Infrastructure)

**Business Model**: Open-core  
**License**: MIT (frameworks) + proprietary (platforms)  
**Pricing**:
- Open-source: Free (LangChain, LangGraph)
- LangSmith: Free tier, Pro $29/mo, Enterprise custom
- LangGraph Platform: Custom pricing

**Key Features**:
- Agent orchestration (LangGraph)
- Evaluation and observability (LangSmith)
- Deployment platform (LangGraph Platform)
- 1M+ integrations and tools
- Production agent management

**Market Position**:
- 🏆 Leading agent framework (1M+ developers)
- Used by Replit, Rakuten, Klarna
- Strong community and ecosystem
- Well-funded ($35M Series A)

**Strengths**:
- ✅ Massive developer adoption
- ✅ Comprehensive tooling (build → deploy → monitor)
- ✅ Open-source core builds trust
- ✅ General-purpose (not just code)

**Weaknesses**:
- ❌ Not specialized for code context
- ❌ Complex for simple use cases
- ❌ Steep learning curve
- ❌ Expensive for production at scale

**How Conexus Competes**:
- 🎯 **Conexus as code-specialized complement** to LangChain
- 📊 Conexus provides code-specific context LangChain doesn't
- 🤝 LangChain apps could use Conexus as context source
- 🔌 Conexus integrates WITH LangChain (not competes)

**Verdict**: LangChain is **complementary**, not competitor. Conexus should integrate with LangChain.

---

### 8. LlamaIndex (Infrastructure)

**Business Model**: Open-core  
**License**: MIT (core) + proprietary (LlamaCloud)  
**Pricing**:
- Open-source: Free (LlamaIndex framework)
- LlamaCloud: Paid tiers for document processing

**Key Features**:
- Document processing (LlamaParse/Extract)
- Agent frameworks
- Enterprise data connectors
- 500M+ documents processed
- 4M+ monthly downloads

**Market Position**:
- 🏆 Leading RAG framework
- 200k+ LlamaCloud users
- Strong enterprise adoption
- Focused on document/knowledge agents

**Strengths**:
- ✅ Best document processing (LlamaParse)
- ✅ Strong enterprise connectors
- ✅ Growing managed platform (LlamaCloud)
- ✅ Open-source core builds trust

**Weaknesses**:
- ❌ General-purpose RAG (not code-specific)
- ❌ Document-focused (not codebase-aware)
- ❌ Requires significant integration work
- ❌ Monetization via processing (limits margins)

**How Conexus Competes**:
- 🎯 **Conexus as code-specialized alternative** to LlamaIndex
- 📊 Conexus understands code semantics, not just documents
- 🤝 LlamaIndex could use Conexus for code context
- 🔌 Conexus integrates WITH LlamaIndex (complementary)

**Verdict**: LlamaIndex is **adjacent**, Conexus specializes where LlamaIndex is general.

---

## Competitive Matrix

### Feature Comparison

| Feature | Conexus | Copilot | Cursor | Continue | Cody | Tabnine | LangChain | LlamaIndex |
|---------|-----|---------|--------|----------|------|---------|-----------|------------|
| **Open-Source Core** | ✅ FSL→MIT | ❌ No | ❌ No | ✅ Apache | ❌ No | ❌ No | ✅ MIT | ✅ MIT |
| **Self-Hosting** | ✅ Yes | ❌ No | ❌ No | ✅ Yes | ⚠️ Limited | ✅ Yes | ✅ Yes | ✅ Yes |
| **Air-Gap Support** | ✅ Yes | ❌ No | ❌ No | ✅ Yes | ✅ Yes | ✅ Yes | ⚠️ DIY | ⚠️ DIY |
| **Code-Specific Context** | ✅ Deep | ⚠️ Basic | ⚠️ Basic | ⚠️ Basic | ✅ Deep | ⚠️ Basic | ❌ No | ❌ No |
| **Multi-Repo Support** | ✅ Native | ⚠️ Limited | ⚠️ Limited | ⚠️ Basic | ✅ Yes | ⚠️ Limited | ⚠️ DIY | ⚠️ DIY |
| **Semantic Understanding** | ✅ Yes | ⚠️ LLM-only | ⚠️ LLM-only | ⚠️ LLM-only | ✅ Yes | ⚠️ Basic | ❌ N/A | ❌ N/A |
| **Pluggable Architecture** | ✅ Yes | ❌ No | ❌ No | ✅ Yes | ⚠️ Limited | ❌ No | ✅ Yes | ✅ Yes |
| **CBAC (Access Control)** | ✅ Yes | ❌ No | ❌ No | ❌ No | ⚠️ Basic | ⚠️ Basic | ❌ N/A | ❌ N/A |
| **Free Tier** | ✅ Unlimited | ✅ Limited | ⚠️ Limited | ✅ Unlimited | ✅ Limited | ✅ Limited | ✅ Unlimited | ✅ Unlimited |
| **Enterprise Compliance** | ✅ SOC2 | ✅ SOC2 | ⚠️ Limited | ⚠️ DIY | ✅ SOC2 | ✅ SOC2+ | ⚠️ DIY | ⚠️ DIY |

### Pricing Comparison (Per User/Month)

| Product | Free | Pro/Individual | Team/Business | Enterprise |
|---------|------|----------------|---------------|------------|
| **Conexus** (projected) | ✅ Unlimited (self-host) | N/A (infra, not SaaS) | $50-100/mo (managed) | Custom |
| GitHub Copilot | Limited (50 chats) | $10-39 | $19 | Custom |
| Cursor | Limited | $20 | N/A | N/A |
| Continue | ✅ Unlimited | N/A | $20 | Custom |
| Cody/Amp | Limited | $9 | N/A | Custom |
| Tabnine | Limited | $12 | N/A | Custom |
| Amazon Q | Limited (50 chats) | $19 | N/A | Custom |
| LangChain | ✅ Unlimited (OSS) | N/A | LangSmith $29+ | Custom |
| LlamaIndex | ✅ Unlimited (OSS) | N/A | LlamaCloud paid | Custom |

---

## Market Positioning for Conexus

### Where Conexus Fits

```
┌─────────────────────────────────────────────────────────┐
│                     End-User Layer                       │
│  (GitHub Copilot, Cursor, Continue, IDEs)               │
│                                                          │
│  ← Conexus integrates here via plugins/APIs ←              │
└─────────────────────────────────────────────────────────┘
                            ↑
                            │
┌─────────────────────────────────────────────────────────┐
│              🎯 Conexus's Infrastructure Layer 🎯             │
│                                                          │
│  • Code-specific context engine                         │
│  • Repository-aware semantic understanding              │
│  • Multi-repo dependency graph                          │
│  • Access control (CBAC)                                │
│  • Pluggable into any AI coding tool                    │
│                                                          │
│  Open-Core Model:                                       │
│  - Community: Self-hosted, single-repo, basic RAG       │
│  - Enterprise: Multi-repo, CBAC, cloud sync, SLAs       │
└─────────────────────────────────────────────────────────┘
                            ↑
                            │
┌─────────────────────────────────────────────────────────┐
│                    Foundation Layer                      │
│  (LangChain, LlamaIndex, Vector DBs, LLM APIs)          │
│                                                          │
│  ← Conexus uses these as building blocks ←                 │
└─────────────────────────────────────────────────────────┘
```

### Conexus's Unique Value Proposition

**"The Stripe of AI Code Context"**

Just as Stripe provides payment infrastructure that powers thousands of applications, **Conexus provides code context infrastructure that powers AI coding tools**.

**Why Conexus Wins**:

1. **Code-Specialized** 🎯
   - Not general-purpose RAG (like LlamaIndex)
   - Understands code semantics, dependencies, architecture
   - Repository-aware, not just file-aware

2. **Pluggable Architecture** 🔌
   - Works with Copilot, Cursor, Continue, custom tools
   - Not locked to one IDE or interface
   - API-first design

3. **Open-Core Trust** 🔓
   - FSL converts to MIT (2 years)
   - Self-hostable core
   - Transparent, auditable code

4. **Enterprise-Ready** 🏢
   - CBAC for multi-tenant security
   - Air-gap support
   - Compliance built-in (SOC2, HIPAA)

5. **Developer-First** 💻
   - Free unlimited self-hosting
   - Excellent docs and DX
   - Community-driven roadmap

---

## Competitive Strategy

### Phase 1: Infrastructure Play (Year 1) ⭐

**Target**: Developers building AI coding tools

**Go-to-Market**:
- Open-source core (FSL-1.1-MIT)
- Excellent documentation and examples
- Integrations with Continue, LangChain
- Community building (GitHub, Discord)

**Revenue Model**:
- Free: Self-hosted, single-repo
- Managed Cloud: $50-100/mo for teams
- Enterprise: Multi-repo, CBAC, SLAs ($5k-50k/year)

**Success Metrics**:
- 10k+ GitHub stars
- 100+ companies self-hosting
- 10-20 enterprise customers

---

### Phase 2: Developer Tool Play (Year 2-3)

**Target**: Individual developers and small teams

**Go-to-Market**:
- VS Code extension powered by Conexus
- "Better context for your AI assistant"
- Freemium SaaS model

**Revenue Model**:
- Free: Local indexing, basic features
- Pro: $10/mo for cloud sync, multi-repo
- Team: $20/user/mo for collaboration

**Success Metrics**:
- 100k+ installations
- 10k+ paying users
- $1-5M ARR

---

### Phase 3: Enterprise Platform Play (Year 3-5)

**Target**: Large enterprises with complex codebases

**Go-to-Market**:
- Enterprise sales team
- SOC2, ISO 27001 certifications
- Customer success team

**Revenue Model**:
- Enterprise: $50k-500k/year contracts
- Success-based pricing (per-repo or per-user)
- Professional services for implementation

**Success Metrics**:
- 50-100 enterprise customers
- $10-50M ARR
- Market leader in code context infrastructure

---

## Threats and Opportunities

### Threats

1. **GitHub Copilot Improves Context** 🔴 High Risk
   - Microsoft has resources to build deep context
   - GitHub platform advantage (all code on GitHub)
   - **Mitigation**: Focus on multi-platform, open-source differentiation

2. **LangChain/LlamaIndex Add Code Specialization** 🟡 Medium Risk
   - General frameworks could add code-specific features
   - Larger communities and ecosystems
   - **Mitigation**: Stay focused, move faster, better UX

3. **Open-Source Clone** 🟡 Medium Risk
   - Someone forks Conexus or builds alternative
   - MIT license after 2 years enables this
   - **Mitigation**: Strong brand, community, managed offering

4. **Enterprise Lock-in to Existing Tools** 🟡 Medium Risk
   - Enterprises may prefer all-in-one (Copilot Enterprise)
   - Switching costs high for established workflows
   - **Mitigation**: Sell as complement, not replacement

### Opportunities

1. **Continue Partnership** 🟢 High Opportunity
   - Leading open-source assistant needs better context
   - Aligned values (open-source, developer-first)
   - **Action**: Reach out to Continue team, propose integration

2. **Cursor Integration** 🟢 High Opportunity
   - Cursor lacks deep codebase context
   - Well-funded, growing fast
   - **Action**: Build Conexus plugin for Cursor, demo to team

3. **Enterprise Tabnine Alternative** 🟢 Medium Opportunity
   - Tabnine expensive, limited innovation
   - Open-core Conexus could displace
   - **Action**: Target Tabnine customers with better context + lower cost

4. **LangChain Ecosystem** 🟢 Medium Opportunity
   - LangChain agents need code context
   - Build Conexus as LangChain integration
   - **Action**: Create LangChain retriever using Conexus

---

## Recommended Next Steps

### Immediate (Next 2 Weeks)

1. **Finalize Positioning** 📋
   - Commit to "Infrastructure Layer" positioning
   - Draft messaging: "The Stripe of AI Code Context"
   - Create positioning doc for website/marketing

2. **Competitive Differentiation** 🎯
   - Update website to highlight vs competitors
   - Create comparison pages (Conexus vs Copilot, vs Continue, etc.)
   - Emphasize code-specialization + open-core

3. **Partnership Outreach** 🤝
   - Email Continue.dev team about integration
   - Demo Conexus to Cursor engineers
   - Join LangChain Discord, propose integration

### Short-Term (Next 1-2 Months)

4. **Build Integrations** 🔌
   - VS Code extension using Conexus
   - LangChain retriever integration
   - Continue.dev context provider

5. **Community Building** 👥
   - Launch Discord server
   - Weekly office hours
   - Contributor onboarding program

6. **Enterprise Validation** 🏢
   - 5-10 design partner interviews
   - Validate CBAC requirements
   - Define enterprise feature roadmap

### Long-Term (Next 6-12 Months)

7. **Product-Market Fit** 📊
   - 100+ self-hosted deployments
   - 10-20 enterprise pilot customers
   - Clear product-market fit signals

8. **Managed Offering** ☁️
   - Launch Conexus Cloud (managed hosting)
   - Freemium tier + paid tiers
   - Enterprise tier with SLAs

9. **Scale Go-to-Market** 📈
   - Hire developer advocates
   - Build enterprise sales team
   - Expand partnerships (Sourcegraph, Replit, etc.)

---

## Conclusion

**Conexus's best path forward**:

1. **Position as Infrastructure** (Layer 2), not end-user tool
2. **Target developers building AI tools** first, then expand to direct users
3. **Open-core model** (FSL-1.1-MIT) to build trust and community
4. **Specialize in code context** - don't compete with general RAG frameworks
5. **Build partnerships** with Continue, Cursor, LangChain early

**Key Success Factors**:
- ✅ Better code context than any competitor
- ✅ Pluggable architecture (works with everything)
- ✅ Open-source trust (FSL → MIT)
- ✅ Enterprise-grade security (CBAC, air-gap)
- ✅ Developer-first experience

**The market is wide open for a specialized, open-core code context infrastructure player. Conexus can own this space.**

---

## Appendix: Competitor Links

- [GitHub Copilot](https://github.com/features/copilot)
- [Cursor](https://cursor.sh/)
- [Continue](https://continue.dev/)
- [Cody / Amp (Sourcegraph)](https://sourcegraph.com/cody)
- [Tabnine](https://tabnine.com/)
- [Amazon Q Developer](https://aws.amazon.com/q/developer/)
- [Replit Agent](https://replit.com/site/ghostwriter)
- [LangChain](https://langchain.com/)
- [LlamaIndex](https://llamaindex.ai/)

---

**Document Status**: Ready for strategic review  
**Next Review**: After 10 enterprise customer interviews  
**Owner**: Product/Strategy Team
