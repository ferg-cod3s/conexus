# Phase 6: RAG Retrieval Pipeline + Production Foundations

## Status: 🚧 In Progress
**Start Date**: 2025-01-15  
**Target Completion**: 2025-02-15 (4 weeks)

## Overview
Phase 6 implements the core RAG (Retrieval-Augmented Generation) pipeline with hybrid search capabilities and minimal production foundations. This phase establishes Conexus as a functional MCP server that can index codebases and serve relevant context to LLM agents.

**Approach**: Option C (Hybrid) - Core RAG + minimal production essentials for a working MVP.

## Success Criteria
- ✅ Index `tests/fixtures` directory in <10 seconds
- ✅ Hybrid search returns relevant results (BM25 + vector fusion)
- ✅ MCP server exposes `context.search` and `resources.*` via stdio
- ✅ 80%+ unit test coverage across new packages
- ✅ Deterministic integration tests pass
- ✅ Docker deployment with SQLite persistence works
- ✅ Basic observability (metrics, health checks, structured logging)

## Technical Decisions

### Vector Storage & Search
- **Primary Store**: SQLite with `modernc.org/sqlite` (pure Go, CGO-free)
- **BM25**: SQLite FTS5 extension for sparse retrieval
- **Vector Similarity**: In-memory vector index (MVP); SQLite blob storage
- **Hybrid Fusion**: Reciprocal Rank Fusion (RRF) combining BM25 + dense vectors

### Embeddings
- **Default**: Mock embedder (deterministic, zero-cost)
- **Interface**: `Embedder` with provider hooks for future integrations
- **Deferred**: OpenAI, Voyage, Cohere, local models (post-MVP)

### MCP Protocol
- **Transport**: stdio only (JSON-RPC over stdin/stdout)
- **Deferred**: HTTP/SSE transports
- **Tools**: `context.search`, `context.index` (optional)
- **Resources**: `codebase://`, `docs://` URIs

### Storage & Indexing
- **Hashing**: Merkle tree for incremental indexing
- **Chunking**: Treesitter AST-based (code), sliding window (docs)
- **Metadata**: File path, language, chunk type, git commit hash

## Package Architecture

```
internal/
  indexer/              # File system → chunks → metadata
    walker.go           # Directory traversal with .gitignore support
    merkle.go           # Content hashing for incremental updates
    chunker.go          # Code & doc chunking strategies
    chunker_code.go     # Treesitter AST-based chunking
    chunker_docs.go     # Markdown/text sliding window
    
  embedding/            # Pluggable embedding generation
    embedder.go         # Interface + provider registry
    mock.go             # Deterministic mock embedder (default)
    provider.go         # Provider hooks (OpenAI, Voyage, etc.)
    
  vectorstore/          # Storage abstractions
    store.go            # VectorStore interface
    sqlite/             # SQLite implementation
      store.go          # Schema, CRUD operations
      fts5.go           # BM25 full-text search
      vector.go         # Vector similarity (in-memory MVP)
      
  search/               # Hybrid search & reranking
    retriever.go        # Retriever interface
    hybrid/             # Sparse + dense fusion
      fusion.go         # RRF/weighted combination
    rerank/             # Cross-encoder reranking
      reranker.go       # Interface
      lexical.go        # Simple lexical reranker (MVP)
      
  mcp/                  # MCP protocol implementation
    server.go           # JSON-RPC server over stdio
    handlers.go         # Tool/resource request handlers
    schema.go           # MCP message types
    
cmd/conexus/            # CLI entrypoint
  main.go               # Command router
  cmd_index.go          # Index subcommand
  cmd_search.go         # Search subcommand
  cmd_serve.go          # MCP server subcommand
  cmd_status.go         # Status/stats subcommand
```

## Task Breakdown (19 Tasks)

### High Priority: Core Pipeline (11 tasks)

#### 6.1: Planning & Interfaces
- [x] **6.1.1** Create PHASE6-PLAN.md (this document)
- [ ] **6.1.2** Define core interfaces in `internal/indexer`, `internal/embedding`, `internal/vectorstore`, `internal/search`
- [ ] **6.1.3** Create package READMEs with usage examples

#### 6.2: Indexing Pipeline
- [ ] **6.2.1** Implement `internal/indexer/walker.go` with .gitignore support
- [ ] **6.2.2** Implement `internal/indexer/merkle.go` for content hashing
- [ ] **6.2.3** Implement `internal/indexer/chunker_code.go` (AST-based chunking)
- [ ] **6.2.4** Implement `internal/indexer/chunker_docs.go` (sliding window)
- [ ] **6.2.5** Unit tests for indexer package (80%+ coverage)

#### 6.3: Embedding Layer
- [ ] **6.3.1** Implement `internal/embedding/embedder.go` interface + registry
- [ ] **6.3.2** Implement `internal/embedding/mock.go` (deterministic embedder)
- [ ] **6.3.3** Unit tests for embedding package

#### 6.4: Vector Storage
- [ ] **6.4.1** Implement `internal/vectorstore/sqlite/store.go` (schema, CRUD)
- [ ] **6.4.2** Implement `internal/vectorstore/sqlite/fts5.go` (BM25 search)
- [ ] **6.4.3** Implement `internal/vectorstore/sqlite/vector.go` (in-memory index)
- [ ] **6.4.4** Unit tests for vectorstore package

#### 6.5: Hybrid Search
- [ ] **6.5.1** Implement `internal/search/hybrid/fusion.go` (RRF combiner)
- [ ] **6.5.2** Implement `internal/search/rerank/lexical.go` (simple reranker)
- [ ] **6.5.3** Unit tests for search packages

#### 6.6: MCP Server
- [ ] **6.6.1** Implement `internal/mcp/server.go` (JSON-RPC stdio handler)
- [ ] **6.6.2** Implement `internal/mcp/handlers.go` (context.search, resources.*)
- [ ] **6.6.3** Implement `internal/mcp/schema.go` (MCP message types)
- [ ] **6.6.4** Unit tests for MCP package

#### 6.7: CLI Commands
- [ ] **6.7.1** Implement `cmd/conexus/cmd_index.go` (index subcommand)
- [ ] **6.7.2** Implement `cmd/conexus/cmd_search.go` (search subcommand)
- [ ] **6.7.3** Implement `cmd/conexus/cmd_serve.go` (MCP server mode)
- [ ] **6.7.4** Implement `cmd/conexus/cmd_status.go` (stats/health check)

#### 6.8: Integration Testing
- [ ] **6.8.1** E2E test: index tests/fixtures → verify chunks stored
- [ ] **6.8.2** E2E test: hybrid search returns relevant results
- [ ] **6.8.3** E2E test: MCP stdio protocol handles requests/responses

### Medium Priority: Production Basics (6 tasks)

#### 6.9: Configuration
- [ ] **6.9.1** Implement `internal/config` package (env + YAML config)
- [ ] **6.9.2** Add default config with SQLite paths, embedding settings

#### 6.10: Observability
- [ ] **6.10.1** Add structured logging with `slog` (context-aware)
- [ ] **6.10.2** Add metrics with `expvar` or `promhttp` (indexing/search stats)
- [ ] **6.10.3** Add health check endpoint (for future HTTP transport)

#### 6.11: Deployment
- [ ] **6.11.1** Create Dockerfile (multi-stage build, CGO-free)
- [ ] **6.11.2** Create docker-compose.yml with SQLite volume persistence

#### 6.12: Performance
- [ ] **6.12.1** Add benchmarks for indexing throughput (chunks/sec)
- [ ] **6.12.2** Add benchmarks for search latency (p50, p95, p99)

#### 6.13: Security
- [ ] **6.13.1** Implement path traversal safeguards in file walker
- [ ] **6.13.2** Implement file type allowlist (no binaries, secrets)
- [ ] **6.13.3** Add PII redaction hooks (placeholder for future)

#### 6.14: Documentation
- [ ] **6.14.1** Update README.md with quickstart and architecture overview
- [ ] **6.14.2** Create docs/guides/cli-usage.md
- [ ] **6.14.3** Create docs/guides/mcp-setup.md
- [ ] **6.14.4** Add architecture diagrams (indexing + search flows)

### Low Priority: Release (1 task)

#### 6.15: Release Prep
- [ ] **6.15.1** Bump version to v0.1.0-alpha
- [ ] **6.15.2** Generate CHANGELOG.md from commit history
- [ ] **6.15.3** Build release binaries (Linux, macOS, Windows)

## Dependencies & Blockers

### External Dependencies
- `modernc.org/sqlite` - SQLite pure-Go driver (CGO-free)
- Consider: `github.com/blevesearch/bleve` for FTS if FTS5 insufficient
- Consider: `github.com/chewxy/math32` for float32 vector ops

### Internal Dependencies
- Phase 5 orchestrator can optionally use MCP tools once available
- CLI commands require all pipeline components (indexer → search)

### Known Blockers
- None currently; all dependencies available

## Testing Strategy

### Unit Tests (80%+ coverage target)
- Mock file system for walker tests (`afero.MemMapFs`)
- Deterministic mock embedder for reproducible tests
- In-memory SQLite (`:memory:`) for vectorstore tests
- Table-driven tests for chunking strategies

### Integration Tests
- Use `tests/fixtures` as canonical test corpus
- Verify indexing correctness (chunk count, metadata)
- Verify search relevance (known queries → expected results)
- Verify MCP protocol compliance (stdio request/response)

### Benchmarks
- `BenchmarkIndexing` - 1000 files, measure throughput
- `BenchmarkSearch` - p50/p95/p99 latency under load
- `BenchmarkEmbedding` - mock vs future real providers

## Migration Plan

### From Phase 5 → Phase 6
- No breaking changes to existing orchestrator/agent/validation packages
- New packages (`indexer`, `embedding`, `vectorstore`, `search`, `mcp`) are additive
- CLI updated with new subcommands (backward compatible)

### Database Migrations
- SQLite schema versioning via `PRAGMA user_version`
- Migration scripts in `internal/vectorstore/sqlite/migrations/`
- Auto-migrate on startup (detect version, apply deltas)

## Rollout Strategy

### Week 1: Foundation (Tasks 6.1-6.3)
- Define interfaces, implement indexer + embedding layers
- Unit tests passing for core components

### Week 2: Storage & Search (Tasks 6.4-6.5)
- Implement SQLite vectorstore with BM25 + vector support
- Implement hybrid search fusion + reranking
- Integration tests for retrieval pipeline

### Week 3: MCP + CLI (Tasks 6.6-6.8)
- Implement MCP server with stdio transport
- Add CLI subcommands (index, search, serve, status)
- E2E tests for MCP protocol

### Week 4: Production & Release (Tasks 6.9-6.15)
- Add config, observability, deployment artifacts
- Security pass and documentation updates
- Release v0.1.0-alpha

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| SQLite FTS5 insufficient for BM25 | High | Evaluate `blevesearch/bleve` as fallback |
| Vector similarity slow in-memory | Medium | Acceptable for MVP; defer HNSW/IVF to Phase 7 |
| MCP protocol changes | Medium | Pin to stable MCP spec version; version handlers |
| Treesitter binding complexity | Medium | Start with simple regex-based chunking fallback |
| Embedding provider costs | Low | Mock embedder default; providers opt-in |

## Success Metrics

### Performance Targets
- **Indexing**: 1000+ chunks/second (small Go files)
- **Search Latency**: <100ms p95 for 10K chunk corpus
- **Memory**: <500MB RSS for 50K chunks indexed

### Quality Targets
- **Test Coverage**: 80%+ unit, 100% integration
- **Search Relevance**: Top-5 recall >70% on test queries
- **Stability**: Zero panics, graceful error handling

### Adoption Targets (Post-Release)
- 10+ GitHub stars in first week
- 3+ community feedback issues filed
- 1+ external contributor PR

## Open Questions

1. **Chunking Strategy**: Use Treesitter or simpler heuristics initially?
   - **Decision**: Start with regex/line-based, add Treesitter in 6.2.3 if time permits

2. **Vector Dimensions**: 384 (all-MiniLM) or 1536 (OpenAI)?
   - **Decision**: Configurable; default 384 for efficiency

3. **MCP Tool Naming**: `context.search` or `conexus.search`?
   - **Decision**: `context.search` (generic), `conexus.*` for custom tools

4. **CLI vs Library**: Expose Go API for embedding in other tools?
   - **Decision**: CLI-first for Phase 6; library API in Phase 7

## References
- [Development Roadmap](docs/Development-Roadmap.md)
- [PRD - User Stories](docs/PRD.md)
- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [SQLite FTS5 Extension](https://www.sqlite.org/fts5.html)
- [Reciprocal Rank Fusion Paper](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf)

---

**Last Updated**: 2025-01-15  
**Next Review**: 2025-01-22 (Week 1 checkpoint)
