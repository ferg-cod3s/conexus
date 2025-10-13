# License Recommendation for Conexus

**Status**: Draft Recommendation  
**Version**: 1.0  
**Last Updated**: 2025-10-12  
**Author**: F3RG

## Executive Summary

**Recommendation**: Adopt **FSL-1.1-MIT (Functional Source License)** with 2-year conversion to MIT.

### Why FSL Over Alternatives

| Criteria | FSL | BSL | SSPL | Elastic 2.0 | AGPL |
|----------|-----|-----|------|-------------|------|
| Protects from cloud competitors | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes | ⚠️ Partial |
| Converts to permissive OSS | ✅ 2 years | ✅ 3-4 years | ❌ Never | ❌ Never | ❌ Never |
| Developer-friendly | ✅ Very | ⚠️ Complex | ❌ Confusing | ❌ Restricted | ⚠️ Viral |
| Simple, consistent terms | ✅ Yes | ❌ Variable | ⚠️ Complex | ⚠️ Complex | ✅ Yes |
| Used by successful companies | ✅ Sentry, Codecov | ⚠️ MariaDB, CockroachDB | ⚠️ MongoDB | ⚠️ Elastic | ⚠️ Many |
| Community acceptance | ✅ Growing | ⚠️ Mixed | ❌ Poor | ❌ Poor | ⚠️ Restrictive |

---

## License Options Deep Dive

### 1. FSL (Functional Source License) ⭐ RECOMMENDED

**Overview**:
- Created by Sentry (Functional Software, Inc.) in 2023
- Designed specifically for SaaS companies valuing user freedom AND developer sustainability
- Converts to Apache 2.0 or MIT after **2 years**
- Simple, consistent terms (no customization allowed)

**Key Features**:
```
✅ Source code visible and modifiable
✅ Can be used for almost any purpose
❌ Cannot compete directly with producer's SaaS offering
⏰ Automatically becomes permissive OSS after 2 years
```

**Protection Mechanism**:
```
You cannot:
- Host Conexus as a commercial service competing with Conexus's official offerings
- Sell Conexus hosting/management as a product
- Offer Conexus as a managed service without authorization

You can:
- Use Conexus internally at your company (any size)
- Modify and contribute back to Conexus
- Self-host Conexus for your organization
- Study, learn, and fork the code
- Use it for commercial software development
```

**Companies Using FSL**:
- **Sentry** (error monitoring) - $3B valuation
- **Codecov** (code coverage) - GitHub competitor
- **GitButler** (version control UI)
- **Convex** (backend platform)
- **PowerSync** (sync engine)

**Pros**:
- ✅ Simple, clear language (no legal ambiguity)
- ✅ Short conversion timeline (2 years)
- ✅ Developer-friendly (use for anything except direct competition)
- ✅ Growing ecosystem and acceptance
- ✅ No "Additional Use Grant" complexity (unlike BSL)
- ✅ Backed by successful SaaS company (Sentry)

**Cons**:
- ⚠️ Relatively new (2023) - less legal precedent
- ⚠️ Cannot customize terms (unlike BSL)
- ⚠️ Not OSI-approved (but becomes MIT/Apache after 2 years)

**Conexus Fit**: ⭐⭐⭐⭐⭐ **Excellent**

---

### 2. BSL (Business Source License)

**Overview**:
- Created by MariaDB in 2013, updated to v1.1 in 2017
- Allows customization via "Additional Use Grant"
- Converts to any license after 3-4 years
- Variable implementation across companies

**Key Features**:
```
✅ Source code visible and modifiable
❌ Commercial use restrictions defined per implementation
⏰ Converts to specified OSS license after X years (typically 3-4)
```

**Protection Mechanism**:
```
Each BSL implementation defines its own "Additional Use Grant"
- Some allow non-commercial use only
- Some allow internal use but not SaaS
- Some allow SaaS below certain revenue thresholds
- Highly variable and potentially confusing
```

**Companies Using BSL**:
- **MariaDB** (database) - original creator
- **CockroachDB** (distributed SQL)
- **HashiCorp** (Terraform, Vault) - controversial switch

**Pros**:
- ✅ More established (since 2013)
- ✅ Customizable restrictions
- ✅ Longer delay before conversion (protects longer)

**Cons**:
- ❌ Each implementation is essentially a new license (complexity)
- ❌ Longer conversion (3-4 years vs FSL's 2 years)
- ❌ Additional Use Grant creates variability
- ❌ Mixed community reception
- ⚠️ HashiCorp's controversial switch damaged BSL reputation

**Conexus Fit**: ⭐⭐⭐ **Good, but FSL is simpler**

---

### 3. SSPL (Server Side Public License)

**Overview**:
- Created by MongoDB in 2018
- Strong copyleft that extends to entire service stack
- Never converts to permissive license

**Key Features**:
```
✅ Source code visible and modifiable
❌ If you offer as a service, must open-source ENTIRE stack
❌ Never becomes permissive OSS
```

**Protection Mechanism**:
```
"Service Source Code" clause:
- If you provide Conexus as a service, you must release:
  - All management software
  - All monitoring software
  - All orchestration software
  - Entire hosting infrastructure code

This makes cloud provider hosting economically unviable.
```

**Companies Using SSPL**:
- **MongoDB** (database)
- Few others (limited adoption)

**Pros**:
- ✅ Very strong protection against cloud providers
- ✅ Precedent with large company (MongoDB)

**Cons**:
- ❌ Never becomes permissive OSS
- ❌ Extremely restrictive and complex
- ❌ Rejected by OSI as "not open source"
- ❌ Poor community acceptance
- ❌ May discourage legitimate use cases

**Conexus Fit**: ⭐ **Poor - too restrictive**

---

### 4. Elastic License 2.0

**Overview**:
- Created by Elastic in 2021
- Source-available but not open-source
- Never converts to permissive license

**Key Features**:
```
✅ Source code visible and modifiable
❌ Cannot provide as managed service
❌ Cannot circumvent license key/payment features
❌ Never becomes OSS
```

**Companies Using Elastic License**:
- **Elastic** (Elasticsearch, Kibana)

**Pros**:
- ✅ Simple terms
- ✅ Clear restriction on managed services

**Cons**:
- ❌ Never becomes permissive OSS
- ❌ Limits legitimate self-hosting use cases
- ❌ Controversial in community

**Conexus Fit**: ⭐⭐ **Fair - but never opens up**

---

### 5. AGPL (Affero General Public License v3)

**Overview**:
- FSF copyleft license (2007)
- Viral copyleft extending to network use
- Already open-source

**Key Features**:
```
✅ OSI-approved open-source
❌ Strong viral copyleft (infects entire codebase)
❌ Network use triggers copyleft obligations
```

**Protection Mechanism**:
```
If you modify Conexus and offer it as a service:
- Must release your modifications
- Does NOT prevent cloud providers from hosting original
- Does NOT protect business model
```

**Companies Using AGPL**:
- Many open-source projects
- Some requiring dual licensing for commercial use

**Pros**:
- ✅ OSI-approved open-source
- ✅ Well-understood legal terms
- ✅ Strong community acceptance

**Cons**:
- ❌ Does not protect against AWS/Google hosting unmodified version
- ❌ Viral copyleft scares enterprise adopters
- ❌ Cannot monetize through licensing
- ❌ Must rely on support/services for revenue

**Conexus Fit**: ⭐ **Poor - doesn't protect business model**

---

## Recommended License Structure for Conexus

### Dual Repository Strategy

**Option A: Single License (Simpler) ⭐ RECOMMENDED**

```
conexus/
├── LICENSE.md              # FSL-1.1-MIT
├── CONTRIBUTING.md         # CLA required
└── [all code under FSL]
```

**License Header**:
```
# Conexus - Advanced Context Engine
# Copyright (C) 2025 Conexus Contributors
# 
# Licensed under the FSL-1.1-MIT (Functional Source License)
# https://fsl.software/
# 
# This software is licensed under FSL-1.1-MIT and converts to MIT License
# two years from the date of publication. See LICENSE.md for details.
```

**What this means**:
- All Conexus code visible and usable
- Cannot compete as managed service
- Community features + Enterprise features both under FSL
- After 2 years, specific version becomes MIT (fully open)
- Simple, consistent, developer-friendly

---

**Option B: Dual License (More Complex)**

```
conexus/
├── LICENSE.md              # FSL for enterprise features
├── LICENSE-COMMUNITY.md    # MIT for community features
├── community/              # MIT licensed
│   ├── core-indexing/
│   ├── basic-rag/
│   └── local-storage/
└── enterprise/             # FSL licensed
    ├── multi-repo/
    ├── cbac/
    ├── sso/
    └── cloud-sync/
```

**Pros**:
- Core features immediately open (community building)
- Enterprise features protected longer

**Cons**:
- Complex to maintain
- Contributor confusion
- License boundary enforcement difficult

**Verdict**: Stick with **Option A (Single FSL)** for simplicity.

---

## CLA (Contributor License Agreement) Strategy

### Required: Contributor License Agreement

**Why CLA**:
```
Without CLA:
❌ Cannot change license later
❌ Cannot offer enterprise exceptions
❌ Cannot defend against IP claims
❌ Cannot build enterprise business
```

**CLA Structure**:
```
1. Grant of Rights
   - Contributors grant Conexus perpetual license to use contributions
   - Contributors retain copyright
   - Conexus can sublicense (for enterprise deals)

2. Representations
   - Contributor owns the code or has permission
   - No known patent claims
   - Complies with employer policies

3. No Assignment
   - Contributor keeps copyright
   - Just grants Conexus necessary rights
```

**Implementation**:
- Use [CLA Assistant](https://github.com/cla-assistant/cla-assistant) for automation
- Sign on first PR contribution
- Store in database, not per-file

---

## License Text for Conexus

### Complete FSL-1.1-MIT License

```markdown
# Functional Source License, Version 1.1, MIT Future License

## Abbreviation

FSL-1.1-MIT

## Notice

Copyright 2025 Conexus Contributors

## Terms and Conditions

### Licensor

The person or entity offering the Software under these Terms and Conditions.

### Software

The software the Licensor makes available under these Terms and Conditions.

### License Grant

Subject to your compliance with this License Grant and the Patents, Redistribution and Trademark clauses below, the Licensor grants you the right to use, copy, modify, create derivative works, redistribute, and make non-production use of the Software.

### Use Limitation

The Software is provided for non-commercial and commercial use, except that you may not use the Software for a Competing Use. A Competing Use means use of the Software in or for a commercial product or service that competes with the Software or any other product or service offered by Licensor using the Software.

### Patents

To the extent your use for a non-Competing Use would necessarily infringe a patent claim the Licensor can license without payment to a third party, the Licensor grants you a license under that patent claim. However, your license for that patent claim terminates if you make any patent claim against the Licensor or the Software.

### Redistribution

The Terms and Conditions apply to all copies, modifications, and derivative works of the Software.

### Disclaimer

The Software is provided "as-is" with no warranties of any kind.

### Trademark

You may not use the marks of the Licensor in any way that suggests your product or service is the Licensor's, or otherwise indicates any relationship between the Licensor and you.

### Future License

On the two-year anniversary of the first publication of this version of the Software, or on the two-year anniversary of a release of this version of the Software, or when the Licensor ceases to support the Software or the Licensor (or any successor of the Licensor), the License Grant automatically becomes a grant of rights under the terms of the MIT License for that version of the Software.
```

---

## Communicating License Choice to Community

### FAQ Additions

**"Why isn't Conexus fully open-source (MIT/Apache) from day one?"**

> Conexus uses the Functional Source License (FSL), which balances user freedom with business sustainability. You can:
> - Use Conexus for any purpose (including commercial software development)
> - Modify and contribute improvements
> - Self-host for your organization
> - Study and learn from the code
> 
> The only restriction: You can't compete directly by offering Conexus as a managed service.
> 
> After 2 years, each version automatically becomes MIT licensed (fully open-source).

**"Can I use Conexus at my company?"**

> Yes! Conexus is free for internal use at companies of any size. You can:
> - Self-host for your development team
> - Customize for your workflows
> - Use for commercial software development
> 
> The only restriction is offering Conexus as a competing hosted service.

**"What about contributions?"**

> We welcome contributions! You'll sign a standard CLA (Contributor License Agreement) that:
> - Lets you keep copyright to your code
> - Grants Conexus rights to include your code
> - Ensures the project can sustainably continue
> 
> Same model as Linux Foundation, Apache, and other major projects.

**"Why not AGPL?"**

> AGPL is OSI-approved open-source but doesn't protect the project's sustainability. Cloud providers could host unmodified Conexus without contributing back, undermining our ability to invest in development.
> 
> FSL strikes a better balance: nearly identical freedoms for users, but ensures the project remains sustainable.

---

## Comparison: Conexus vs Competitors

| Project | License | Converts to Open? | Can Self-Host? | Enterprise Model |
|---------|---------|-------------------|----------------|------------------|
| **Conexus** | **FSL-1.1-MIT** | **✅ Yes (2 years)** | **✅ Yes** | **Open-core with FSL** |
| Sentry | FSL-1.1-ALv2 | ✅ Yes (2 years) | ✅ Yes | Open-core with FSL |
| Sourcegraph | Proprietary | ❌ No | ⚠️ Limited | Proprietary with free tier |
| Continue | Apache 2.0 | ✅ Already open | ✅ Yes | Open-core with commercial tiers |
| LangChain | MIT | ✅ Already open | ✅ Yes | Open frameworks + commercial platforms |
| MongoDB | SSPL | ❌ No | ✅ Yes | SSPL copyleft |
| Elastic | Elastic 2.0 | ❌ No | ✅ Yes | Source-available, restrictive |

---

## Implementation Checklist

### Phase 1: Legal Documents (Week 1)
- [ ] Adopt FSL-1.1-MIT as primary license
- [ ] Create `LICENSE.md` with complete FSL text
- [ ] Add license headers to all source files
- [ ] Draft CLA using standard template
- [ ] Set up CLA Assistant or similar tool

### Phase 2: Documentation (Week 2)
- [ ] Update README with license explanation
- [ ] Create FAQ addressing common questions
- [ ] Add "Why FSL?" section to docs
- [ ] Document contribution process with CLA
- [ ] Create comparison table vs competitors

### Phase 3: Community Communication (Week 3)
- [ ] Blog post announcing license choice
- [ ] Reddit/HN post explaining rationale
- [ ] Update website with license info
- [ ] Add license badge to README
- [ ] Prepare talking points for interviews

### Phase 4: Technical Implementation (Week 4)
- [ ] Add license check to CI/CD
- [ ] Update package.json / pyproject.toml with license field
- [ ] Create LICENSE-THIRD-PARTY.md for dependencies
- [ ] Add SPDX identifiers to files
- [ ] Set up automated CLA checks on PRs

---

## Risks and Mitigation

### Risk 1: Community Pushback

**Risk**: Developers prefer "true" open-source (OSI-approved)

**Mitigation**:
- Emphasize 2-year conversion to MIT
- Highlight that usage freedoms are nearly identical to MIT
- Show successful FSL adopters (Sentry, Codecov)
- Position as "delayed open-source" not "closed-source"

### Risk 2: Enterprise Hesitation

**Risk**: Legal teams may be unfamiliar with FSL

**Mitigation**:
- Provide clear FAQ for legal review
- Offer comparison to well-known licenses
- Highlight enterprise adopters using FSL
- Provide reference legal opinions (if available)

### Risk 3: Contributor Confusion

**Risk**: CLA + FSL may seem complex

**Mitigation**:
- Simple, automated CLA process
- Clear contribution guide
- FAQ addressing common concerns
- Show that major projects use CLAs

---

## Final Recommendation

### License: FSL-1.1-MIT

**Rationale**:
1. **Protects business model** from cloud provider competition
2. **Developer-friendly** with minimal restrictions
3. **Becomes MIT** after 2 years (true open-source)
4. **Simple and consistent** (no custom terms)
5. **Proven by Sentry** and other successful SaaS companies
6. **Growing ecosystem** with increasing adoption

### Next Steps:
1. Adopt FSL-1.1-MIT immediately
2. Implement CLA process
3. Update all documentation
4. Communicate clearly to community
5. Monitor community feedback and adjust messaging

**This positions Conexus as a sustainable, developer-friendly, community-oriented project that will become fully open-source while protecting against harmful free-riding.**

---

## References

- [FSL Official Site](https://fsl.software/)
- [Sentry FSL Announcement](https://blog.sentry.io/introducing-the-functional-source-license-freedom-without-free-riding/)
- [Business Source License](https://mariadb.com/bsl11/)
- [Choosing an Open Source License](https://choosealicense.com/)
- [OSI Approved Licenses](https://opensource.org/licenses/)

---

**Document Status**: Ready for review and implementation  
**Approval Required**: Founder/CTO, Legal counsel (if available)
