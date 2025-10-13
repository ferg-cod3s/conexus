# Security & Compliance Framework

## 1. Open-Source Core Security Model (Trust through Transparency)

   **Local-First Processing:** All indexing, storage, and retrieval operations for the open-source version are performed **locally on the user's machine**.
   **No Data Exfiltration:** The core engine will not make any network calls to a third-party server with user code or private data. All operations are self-contained. The only external calls are to the LLM provider, initiated by the client.
   **Privacy-First Architecture:** We will follow Cursor's model: if and when a cloud component is introduced, it will store only embeddings with obfuscated filenames. The actual source code snippets will be requested from the local client on-demand and never persisted on servers.
   **Dependency Security:** All dependencies will be scanned for vulnerabilities using tools like Snyk or Dependabot as part of the CI/CD pipeline.
   **Secure Contribution:** All PRs will require a review from a core maintainer, with a focus on security implications.

## 2. Commercial Enterprise Edition Security Features

This layer provides the robust security and compliance features required by large organizations.

   **Context-Based Access Control (CBAC):**    
   *   **Implementation:** During ingestion, connectors will fetch permission metadata (e.g., from GitHub teams or Confluence page restrictions). This metadata will be stored alongside the vectors.    
   *   **Enforcement:** At query time, the engine will authenticate the user (via SSO) and filter all retrieval results against their permissions *before* the reranking stage. This ensures that no unauthorized information ever enters the context window.
   **Authentication:** Integration with enterprise identity providers via **SAML and OpenID Connect (OIDC)** for Single Sign-On (SSO).
   **Multi-Tenant Isolation:** In the managed cloud version, tenants will be strictly isolated using separate database schemas/namespaces and tenant-specific encryption keys for data at rest.
   **Compliance & Auditing:**    
   *   The service will be designed to be **SOC 2 Type II** compliant.    
   *   A comprehensive **audit log** will be maintained, recording every query and all context data accessed by each user for security monitoring and compliance.
