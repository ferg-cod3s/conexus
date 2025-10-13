# Conexus Open-Core Business Strategy

**Status**: Strategic Plan  
**Version**: 1.0  
**Last Updated**: 2025-10-12  
**Owner**: Founder/Product Team

## Executive Summary

Conexus will follow an **infrastructure-focused open-core model**, positioning as "The Stripe of AI Code Context" - essential infrastructure that powers AI coding tools rather than competing as an end-user product.

### Strategic Positioning

**Primary Market**: Infrastructure layer (B2B/B2D)  
**Secondary Market**: Direct to developers (future expansion)  
**License**: FSL-1.1-MIT (converts to MIT after 2 years)  
**Revenue Model**: Self-hosted free + managed cloud + enterprise features

### Go-to-Market Phases

| Phase | Timeline | Target | Strategy | Revenue Goal |
|-------|----------|--------|----------|--------------|
| **Phase 1: Infrastructure** | Year 1 | Tool builders, open-source projects | Open-core + community | $100k-500k ARR |
| **Phase 2: Developer Tools** | Year 2-3 | Individual developers, small teams | Freemium SaaS | $1-5M ARR |
| **Phase 3: Enterprise** | Year 3-5 | Large enterprises | Enterprise contracts | $10-50M ARR |

---

## Open-Core Model Definition

### What is Open-Core?

Open-core is a business model where:
1. **Core product is open-source** (or source-available with permissive timeline)
2. **Advanced features are proprietary** (sold to enterprises)
3. **Community builds on core**, enterprises pay for scale/security/support

**Examples of Successful Open-Core**:
- GitLab: $10B+ valuation, 90% codebase open-source
- Sentry: $3B valuation, FSL core + enterprise features
- Elastic: $10B+ valuation (now Elastic License, formerly Apache)
- HashiCorp: $5B+ IPO (now BSL)
- Databricks: $43B valuation, open Spark core

### Why Open-Core for Conexus?

**Advantages**:
- ✅ Build developer trust and community
- ✅ Faster adoption (free self-hosting)
- ✅ Community contributions improve core
- ✅ Enterprise customers pay for scale/security
- ✅ Defensible moat (community + brand)

**Disadvantages**:
- ⚠️ Competitors can fork (mitigated by FSL, brand, community)
- ⚠️ Feature split requires careful planning
- ⚠️ Need both community AND enterprise strategy

---

## Feature Split: Community vs Enterprise

### Decision Framework

**Core Principle**: "What would a solo developer or small team need?" → Community  
**Enterprise Principle**: "What do large orgs with complex needs require?" → Enterprise

### Community Edition (FSL-1.1-MIT, Free Forever)

**Target Users**: 
- Individual developers
- Small teams (2-10 people)
- Open-source projects
- Students/educators

**Features** (All Free, Self-Hosted):

| Feature | Description | Why Community |
|---------|-------------|---------------|
| **Local Indexing** | Index repositories on local machine | Essential for solo devs |
| **Basic RAG** | Vector-based retrieval | Core functionality |
| **Single Repository** | Context for one repo at a time | Sufficient for small projects |
| **File-Based Storage** | Local file system for vectors | Simple, no infrastructure needed |
| **CLI Interface** | Command-line tool | Developer-friendly, automation |
| **VS Code Extension** | Basic IDE integration | Reach developers where they work |
| **API Server** | REST API for custom integrations | Extensibility |
| **LangChain Integration** | Use Conexus with LangChain apps | Ecosystem growth |
| **Semantic Search** | Code understanding and search | Core value prop |
| **Documentation** | Full docs, guides, examples | Community enablement |

**Infrastructure**:
- Self-hosted (Docker, binary, pip install)
- Local Qdrant or in-memory vector store
- No cloud dependencies
- Air-gap capable

**Support**:
- Community Discord/forums
- GitHub issues
- Public documentation
- No SLAs

---

### Enterprise Edition (Proprietary, Paid)

**Target Users**:
- Large enterprises (100+ developers)
- Companies with compliance requirements
- Organizations with complex codebases
- Regulated industries (finance, healthcare)

**Features** (Paid Subscription):

| Feature | Description | Why Enterprise |
|---------|-------------|----------------|
| **Multi-Repository** | Index and search across 100+ repos | Scale requirement |
| **Dependency Graph** | Cross-repo dependency tracking | Complex codebases |
| **CBAC (Context-Based Access Control)** | Fine-grained permissions | Security requirement |
| **SSO/SAML** | Enterprise authentication | IT requirement |
| **SCIM Provisioning** | User lifecycle management | IT requirement |
| **Audit Logging** | Track all context queries | Compliance requirement |
| **Cloud Sync** | Sync indexes across team | Collaboration at scale |
| **High Availability** | Multi-region, failover | Enterprise SLA |
| **Advanced Analytics** | Usage patterns, performance metrics | Management visibility |
| **Priority Support** | SLA-backed support (24/7 for critical) | Risk mitigation |
| **Custom Deployment** | On-premise, VPC, air-gap options | Enterprise IT requirements |
| **Compliance Certifications** | SOC2, ISO 27001, HIPAA | Regulatory requirements |
| **Team Management** | Role-based access, teams, projects | Organizational structure |
| **Advanced Caching** | Distributed cache for performance | Scale optimization |
| **Webhooks** | Real-time notifications | Integration needs |
| **Custom Models** | Bring your own embeddings | Specialized use cases |

**Infrastructure**:
- Managed cloud (Conexus Cloud)
- Self-hosted enterprise (Kubernetes, Terraform)
- Hybrid deployment options
- Dedicated support engineers

**Support**:
- Dedicated Slack channel
- Technical account manager (TAM)
- SLA: 99.9% uptime
- Response times: 1-4 hours (based on severity)
- Onboarding and training included

---

### Feature Split Rationale

**Philosophy**: "Free for individuals, paid for organizations"

**Gray Areas** (Requiring Decisions):

| Feature | Decision | Rationale |
|---------|----------|-----------|
| **Multi-Repo (2-5 repos)** | 🟢 Community | Small teams need this, not just enterprises |
| **Multi-Repo (100+ repos)** | 🔴 Enterprise | Scale management becomes complex |
| **Basic Access Control** | 🟢 Community | Simple user auth is reasonable |
| **CBAC (Fine-Grained)** | 🔴 Enterprise | Complex permissions require support |
| **Cloud Sync (Small Teams)** | 🟡 Pro Tier | Middle ground for small businesses |
| **Analytics (Basic)** | 🟢 Community | Help users understand usage |
| **Analytics (Advanced)** | 🔴 Enterprise | Deep insights for management |

---

## Pricing Strategy

### Tier Structure

```
┌─────────────────────────────────────────────────────┐
│                  COMMUNITY (FREE)                    │
│  ✅ Self-hosted, unlimited usage                     │
│  ✅ Single-repo or up to 5 small repos              │
│  ✅ Full core features                              │
│  ✅ Community support                               │
│                                                      │
│  Target: Individual developers, OSS projects        │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│               PRO (Future - Phase 2)                 │
│  💰 $10-20/user/month                               │
│  ✅ Cloud-hosted (Conexus Cloud)                        │
│  ✅ Multi-repo (up to 20 repos)                     │
│  ✅ Team collaboration (10-50 users)                │
│  ✅ Email support                                   │
│                                                      │
│  Target: Small teams, startups (10-50 people)       │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│            ENTERPRISE (Phase 1 Priority)             │
│  💰 $50k-500k/year (custom contracts)               │
│  ✅ Unlimited repos and users                       │
│  ✅ CBAC, SSO, SCIM                                 │
│  ✅ Compliance certifications                       │
│  ✅ SLA + dedicated support                         │
│  ✅ Custom deployment (on-prem, VPC, hybrid)        │
│  ✅ Professional services                           │
│                                                      │
│  Target: Enterprises (500+ developers)              │
└─────────────────────────────────────────────────────┘
```

### Pricing Models

**Option 1: Per-Repository Pricing** (Recommended for Phase 1)
```
Community: Free (self-hosted, up to 5 repos)
Pro: $50/repo/month (up to 20 repos)
Enterprise: Custom ($500-5000/repo/month for 100+ repos)
```

**Why this works**:
- ✅ Scales with value (more repos = more value)
- ✅ Simple to understand and calculate
- ✅ Aligns with customer growth
- ✅ Similar to other code tools (Sourcegraph, etc.)

**Option 2: Per-User Pricing** (Alternative)
```
Community: Free (self-hosted)
Pro: $10/user/month (10-50 users)
Enterprise: $20-50/user/month (50+ users)
```

**Why this could work**:
- ✅ Common SaaS model (familiar)
- ✅ Predictable for customers
- ⚠️ Less aligned with infrastructure positioning
- ⚠️ Doesn't scale with codebase complexity

**Recommendation**: Start with **per-repository pricing** for infrastructure positioning, consider per-user for Pro tier later.

---

### Pricing Examples

**Scenario 1: Startup (20 repos, 30 developers)**
- Community: $0 (self-host)
- Pro: $1,000/month (20 repos × $50)
- Enterprise: Not needed yet

**Scenario 2: Mid-Market (100 repos, 200 developers)**
- Community: Not viable (needs CBAC)
- Enterprise: $50k/year (negotiated)
  - Includes: CBAC, SSO, cloud-hosted, support
  - OR self-hosted with support contract

**Scenario 3: Large Enterprise (500 repos, 2000 developers)**
- Enterprise: $200k-500k/year (negotiated)
  - Includes: Unlimited repos, dedicated TAM, custom deployment, 99.9% SLA
  - Professional services: $50k-150k (onboarding, training, custom integrations)

---

## Revenue Model

### Phase 1: Infrastructure (Year 1)

**Primary Revenue**: Enterprise contracts  
**Target**: $100k-500k ARR  
**Customers**: 10-20 enterprise pilot customers

**Revenue Breakdown**:
- Enterprise contracts: $50k-150k each
- Professional services: $5k-50k per engagement
- Managed hosting: $5k-20k/month for pilot customers

**Cost Structure**:
- Infrastructure: $5k-20k/month (cloud hosting, Qdrant, etc.)
- Personnel: 3-5 FTE (engineering + founder)
- Sales/Marketing: Minimal (founder-led sales)

**Gross Margin**: 70-80% (infrastructure business)

---

### Phase 2: Developer Tools (Year 2-3)

**Primary Revenue**: Pro tier subscriptions  
**Target**: $1-5M ARR  
**Customers**: 500-2000 paying Pro users + 20-50 enterprises

**Revenue Breakdown**:
- Pro tier: $10-20/user/month × 1000 users = $120k-240k/year
- Enterprise: $50k-200k × 30 customers = $1.5-6M/year
- Professional services: $200k-500k/year

**Cost Structure**:
- Infrastructure: $50k-150k/year (scaled cloud)
- Personnel: 10-15 FTE (engineering, support, sales)
- Sales/Marketing: $500k-1M/year
- Overhead: $200k-500k/year

**Gross Margin**: 60-70% (mixed infrastructure + SaaS)

---

### Phase 3: Enterprise Platform (Year 3-5)

**Primary Revenue**: Enterprise contracts  
**Target**: $10-50M ARR  
**Customers**: 100-200 enterprise customers

**Revenue Breakdown**:
- Enterprise: $100k-500k × 100 customers = $10-50M/year
- Pro tier: $500k-2M/year (steady state)
- Professional services: $2-5M/year

**Cost Structure**:
- Infrastructure: $500k-2M/year (multi-region, HA)
- Personnel: 50-100 FTE (engineering, sales, support, success)
- Sales/Marketing: $5-10M/year
- Overhead: $2-5M/year

**Gross Margin**: 70-80% (enterprise software)

---

## Go-to-Market Strategy

### Phase 1: Infrastructure Play (Year 1) ⭐ CURRENT PRIORITY

**Objective**: Build community and land enterprise design partners

**Target Audience**:
1. **Tool Builders**: Developers building AI coding assistants (Continue, Cursor, custom tools)
2. **Enterprise Teams**: Large companies needing better code context
3. **Open-Source Projects**: High-profile OSS projects that could showcase Conexus

**Channels**:
- 🔓 **Open-Source**: GitHub, community building
- 📝 **Content**: Blog posts, technical deep-dives, conference talks
- 🤝 **Partnerships**: Integrate with Continue, LangChain, Sourcegraph
- 🏢 **Direct Sales**: Founder-led outreach to 50 target enterprises
- 🎯 **Product Hunt**: High-visibility launch

**Key Metrics**:
- GitHub stars: 10,000+ (12 months)
- Self-hosted deployments: 1,000+ (12 months)
- Enterprise customers: 10-20 (12 months)
- ARR: $100k-500k (12 months)

**Activities** (Month-by-Month):

**Months 1-3: Foundation**
- [ ] Complete open-source core (FSL-1.1-MIT)
- [ ] Launch GitHub repo with excellent docs
- [ ] Build VS Code extension
- [ ] Create Docker deployment
- [ ] Write 10 technical blog posts
- [ ] Launch community Discord

**Months 4-6: Partnerships**
- [ ] Integrate with Continue.dev
- [ ] Build LangChain integration
- [ ] Reach out to 100 potential design partners
- [ ] Land 5 enterprise pilot customers ($0 or discounted)
- [ ] Conference talks (2-3)

**Months 7-9: Enterprise Validation**
- [ ] Build enterprise features (CBAC, SSO, SCIM)
- [ ] Complete SOC2 compliance
- [ ] Close 10 paid enterprise customers
- [ ] Launch managed cloud (beta)
- [ ] Expand team (2-3 hires)

**Months 10-12: Scale**
- [ ] Product Hunt launch
- [ ] Close 20 total enterprise customers
- [ ] $100k-500k ARR
- [ ] 10,000+ GitHub stars
- [ ] Raise seed round (optional: $2-5M)

---

### Phase 2: Developer Tools (Year 2-3)

**Objective**: Scale to thousands of paying Pro users

**Target Audience**:
1. **Individual Developers**: Freelancers, consultants
2. **Small Teams**: Startups, agencies (5-50 people)
3. **Open-Source Maintainers**: OSS projects needing better context

**Channels**:
- 💻 **Product-Led Growth**: Freemium SaaS with self-serve upgrade
- 📱 **Social**: Twitter/X, Reddit, HackerNews
- 🎥 **Video**: YouTube tutorials, demos
- 📧 **Email**: Drip campaigns for trial users
- 🎤 **Community**: Discord, office hours, hackathons

**Key Metrics**:
- Free users: 50,000+ (24 months)
- Pro users: 1,000-5,000 (24 months)
- Enterprise customers: 30-50 (24 months)
- ARR: $1-5M (24 months)
- NRR (Net Revenue Retention): >110%

**Activities**:
- Launch Conexus Cloud (managed SaaS)
- Build self-serve onboarding
- Implement usage-based pricing
- Hire developer advocates (2-3)
- Scale content marketing
- Expand partnerships (Replit, Sourcegraph)

---

### Phase 3: Enterprise Platform (Year 3-5)

**Objective**: Become market leader in code context infrastructure

**Target Audience**:
1. **Fortune 500**: Large enterprises with 1000+ developers
2. **Mid-Market**: Companies with 100-500 developers
3. **Regulated Industries**: Finance, healthcare, government

**Channels**:
- 🏢 **Enterprise Sales**: Dedicated sales team (10-20 reps)
- 🤝 **Partnerships**: Reseller agreements (AWS, Azure Marketplace)
- 🎤 **Events**: Sponsor conferences, host user conference
- 📊 **Analysts**: Engage Gartner, Forrester
- 💼 **Account-Based Marketing**: Target top 500 companies

**Key Metrics**:
- Enterprise customers: 100-200 (48 months)
- ARR: $10-50M (48 months)
- Logo retention: >90%
- NRR: >120%
- Gross margin: >75%

**Activities**:
- Build enterprise sales team
- Launch customer success program
- Achieve ISO 27001, HIPAA certifications
- Expand internationally (EMEA, APAC)
- Consider Series A/B fundraising

---

## Customer Segmentation

### Segment 1: Solo Developers / OSS Projects (Community)

**Profile**:
- 1 developer or small OSS project
- 1-5 repositories
- Price-sensitive
- Technically sophisticated

**Needs**:
- Free, self-hosted solution
- Simple setup (< 30 minutes)
- Good documentation
- Community support

**Acquisition**:
- GitHub discovery
- Reddit, HackerNews
- Technical blog posts

**Monetization**:
- None (community building)
- Potential upgrade to Pro later

---

### Segment 2: Small Teams / Startups (Pro Tier)

**Profile**:
- 5-50 developers
- 5-20 repositories
- Growing fast
- Budget-conscious but willing to pay for value

**Needs**:
- Cloud-hosted (don't want to manage infra)
- Team collaboration
- Good support
- Simple pricing

**Acquisition**:
- Product-led growth (freemium)
- Word-of-mouth
- Content marketing
- Product Hunt

**Monetization**:
- Pro tier: $10-20/user/month
- OR: $50/repo/month
- Target: $500-2000/month per customer

---

### Segment 3: Mid-Market Companies (Enterprise Lite)

**Profile**:
- 50-500 developers
- 20-100 repositories
- Security-conscious
- Need some enterprise features

**Needs**:
- SSO, SCIM
- Priority support
- SLA (99% uptime)
- Self-hosted or cloud options

**Acquisition**:
- Direct sales (inside sales)
- Partnerships
- Referrals from existing customers

**Monetization**:
- Enterprise tier: $20k-100k/year
- Target: $50k average deal

---

### Segment 4: Large Enterprises (Enterprise)

**Profile**:
- 500+ developers
- 100+ repositories
- Complex org structure
- High security/compliance requirements

**Needs**:
- CBAC, audit logging
- Multi-region deployment
- Dedicated support
- Professional services
- Compliance certifications

**Acquisition**:
- Enterprise sales (field sales)
- Analyst relations (Gartner)
- Partnerships (AWS, Azure)
- Executive relationships

**Monetization**:
- Enterprise tier: $100k-500k/year
- Professional services: $50k-200k
- Target: $200k average deal

---

## Competitive Moat & Defensibility

### How Conexus Maintains Competitive Advantage

**1. Community Moat** 🏔️
- Open-source core builds community
- Contributors improve product
- Network effects (more users = more feedback = better product)
- Switching costs (community lock-in)

**2. Technical Moat** 🔬
- Deep code understanding (semantic analysis)
- Repository-aware context (not just files)
- Efficient indexing (faster than competitors)
- Proprietary algorithms (not in community edition)

**3. Data Moat** 📊
- Anonymized usage patterns improve model
- Enterprise feedback drives roadmap
- Scale advantages (more data = better features)

**4. Brand Moat** 🏷️
- "The Stripe of AI Code Context"
- Developer trust and recognition
- Enterprise reputation

**5. Integration Moat** 🔌
- Partnerships with Continue, Cursor, LangChain
- Ecosystem lock-in (switching costs)
- Standard API becomes industry standard

**6. Compliance Moat** ✅
- SOC2, ISO 27001, HIPAA certifications
- Enterprise customers trust certifications
- Competitors need years to catch up

---

## Risk Analysis

### Business Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| **GitHub Copilot improves context** | High | High | Focus on pluggable infra, not end-user tool |
| **LangChain adds code specialization** | Medium | Medium | Move faster, better UX, partnerships |
| **Enterprise hesitant on FSL** | Medium | Medium | Education, legal FAQ, FSL → MIT conversion |
| **Open-source clone emerges** | Medium | Low | Brand, community, managed offering |
| **Market grows slower than expected** | Low | High | Diversify revenue (Pro + Enterprise) |

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| **Scaling challenges (100+ repos)** | Medium | High | Early performance testing, distributed architecture |
| **Accuracy issues (retrieval quality)** | Medium | High | Continuous improvement, user feedback loops |
| **Integration complexity** | Low | Medium | Excellent API design, comprehensive docs |
| **Security vulnerabilities** | Low | High | Security audits, bug bounty program |

### Operational Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| **Inability to hire talent** | Medium | High | Remote-first, competitive comp, mission-driven |
| **Customer support overwhelm** | Medium | Medium | Self-service docs, community support, automation |
| **Compliance delays (SOC2)** | Medium | Medium | Start early, use consultants, automate |
| **Cash flow challenges** | Low | High | Raise capital, maintain runway, focus on ARR |

---

## Success Metrics (OKRs)

### Year 1 Objectives

**Objective 1: Build Community** 🏔️
- KR1: 10,000+ GitHub stars
- KR2: 1,000+ self-hosted deployments
- KR3: 100+ active Discord members
- KR4: 50+ external contributors

**Objective 2: Validate Enterprise Market** 🏢
- KR1: 10-20 enterprise customers
- KR2: $100k-500k ARR
- KR3: <20% churn
- KR4: NPS >40

**Objective 3: Build Product Foundation** 🔨
- KR1: 99% uptime for managed cloud
- KR2: <1s average query latency
- KR3: Support for 100+ repos per customer
- KR4: 5+ integrations (Continue, LangChain, etc.)

---

### Year 2-3 Objectives

**Objective 1: Scale Revenue** 💰
- KR1: $1-5M ARR
- KR2: 30-50 enterprise customers
- KR3: 1,000-5,000 Pro users
- KR4: >110% NRR

**Objective 2: Expand Market** 🌍
- KR1: Launch in EMEA
- KR2: 10+ partnerships (tool builders)
- KR3: Featured in Gartner report
- KR4: 3+ AWS/Azure Marketplace listings

**Objective 3: Operational Excellence** 🎯
- KR1: SOC2 Type II certified
- KR2: 99.9% uptime SLA
- KR3: Customer satisfaction >90%
- KR4: <10% logo churn

---

## Recommended Next Actions

### Immediate (Next 2 Weeks)

1. **Finalize Strategy** 📋
   - [ ] Review and approve open-core strategy
   - [ ] Commit to FSL-1.1-MIT license
   - [ ] Define Community vs Enterprise feature split
   - [ ] Set pricing structure (per-repo recommended)

2. **Update Documentation** 📝
   - [ ] Update PRD with open-core model
   - [ ] Update roadmap with enterprise features
   - [ ] Create FAQ for license and pricing
   - [ ] Update website copy

3. **Technical Preparation** 🔨
   - [ ] Implement FSL license headers
   - [ ] Set up CLA process
   - [ ] Create enterprise feature flags
   - [ ] Plan multi-tenancy architecture

### Short-Term (Next 1-2 Months)

4. **Community Launch** 🚀
   - [ ] Finalize docs and examples
   - [ ] Launch GitHub repo (public)
   - [ ] Product Hunt launch
   - [ ] HackerNews "Show HN" post
   - [ ] Launch Discord community

5. **Enterprise Pipeline** 🏢
   - [ ] Identify 100 target enterprise customers
   - [ ] Outreach to 50 for design partner program
   - [ ] Close 5 pilot customers (discounted)
   - [ ] Build enterprise feature roadmap

6. **Partnership Outreach** 🤝
   - [ ] Email Continue.dev team
   - [ ] Demo to Cursor engineers
   - [ ] Integrate with LangChain
   - [ ] Join relevant Discord/Slack communities

### Long-Term (Next 6-12 Months)

7. **Scale Operations** 📈
   - [ ] Hire 2-3 engineers
   - [ ] Hire 1 DevRel / community manager
   - [ ] Build enterprise sales playbook
   - [ ] Set up customer success processes

8. **Product Expansion** 🔨
   - [ ] Launch managed cloud (Conexus Cloud)
   - [ ] Complete enterprise features (CBAC, SSO, SCIM)
   - [ ] Achieve SOC2 compliance
   - [ ] Build analytics dashboard

9. **Fundraising (Optional)** 💰
   - [ ] Prepare pitch deck
   - [ ] Target $2-5M seed round
   - [ ] Focus on strategic investors (infrastructure VCs)
   - [ ] Consider: a16z, Bessemer, Redpoint, Accel

---

## Conclusion

**Conexus's open-core strategy positions us for sustainable growth**:

1. **Community-first**: Build trust through open-source core
2. **Infrastructure focus**: "The Stripe of AI Code Context"
3. **Enterprise monetization**: Sell scale, security, support
4. **Clear path**: Infrastructure → Developer Tools → Enterprise Platform

**Key Success Factors**:
- ✅ Excellent developer experience (community adoption)
- ✅ Superior code context (technical moat)
- ✅ Open-core trust (FSL → MIT conversion)
- ✅ Enterprise-ready (compliance, security, scale)
- ✅ Partnership ecosystem (integrations with all tools)

**Next Milestone**: Launch open-source core + land 10 enterprise design partners (6 months)

---

## Appendix: Open-Core Case Studies

### GitLab: The Gold Standard

**Model**: Open-core (MIT core, proprietary enterprise)  
**Valuation**: $10B+ at IPO (2021)  
**Timeline**: 2011 (founded) → 2021 (IPO) = 10 years

**Key Learnings**:
- Community Edition drove adoption
- Enterprise Edition monetized at scale
- Clear feature split (basic vs advanced)
- Developer-first approach

**Conexus Parallel**: We can follow similar path but faster (AI tailwind)

---

### Sentry: FSL Pioneer

**Model**: FSL-1.1-ALv2 (new license, 2023)  
**Valuation**: $3B (2021)  
**Timeline**: 2008 (OSS) → 2021 ($3B) = 13 years

**Key Learnings**:
- Switched from BSD to FSL to protect business
- Community embraced FSL (developer-friendly)
- Self-hosted option builds trust
- Error monitoring = mission-critical → high willingness to pay

**Conexus Parallel**: We can adopt FSL from day 1, avoid license change drama

---

### Elastic: Cautionary Tale

**Model**: Apache → SSPL/Elastic License (2018-2021)  
**Valuation**: $10B+ (public)  
**Timeline**: 2012 (founded) → 2021 (license change) = 9 years

**Key Learnings**:
- License change was controversial
- AWS competed directly (OpenSearch fork)
- Community frustrated by late change
- Still successful but damaged trust

**Conexus Parallel**: Start with protective license (FSL), don't change later

---

**Document Status**: Strategic plan ready for execution  
**Next Review**: After first 10 enterprise customers  
**Owner**: Founder/CEO
