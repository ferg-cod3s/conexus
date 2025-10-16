# Task 6.6: Deployment & Configuration

**Status**: 🚧 In Progress  
**Start**: 2025-01-15  
**Est. Duration**: 3-4 hours  
**Dependencies**: Tasks 6.2-6.5 ✅

---

## Objectives
1. Implement flexible configuration system
2. Create production-ready Docker deployment
3. Enable environment-based configuration
4. Ensure SQLite data persistence
5. Add graceful shutdown handling

---

## Subtask Breakdown

### 6.6.1: Design Deployment Architecture (~30 min)
- [ ] Define configuration schema
- [ ] Plan environment variable naming convention
- [ ] Design YAML/JSON config format
- [ ] Define validation rules
- [ ] Document configuration options

### 6.6.2: Configuration Management (~1.5 hours)
**Files to Create**:
- `internal/config/config.go`
- `internal/config/config_test.go`

**Requirements**:
- Load from environment variables
- Load from optional config file (YAML/JSON)
- Merge and override logic (env > file > defaults)
- Comprehensive validation
- Default values for all settings
- 80%+ test coverage

### 6.6.3: Dockerfile Creation (~1 hour)
**Files to Create**:
- `Dockerfile`
- `.dockerignore`

**Requirements**:
- Multi-stage build (builder + minimal runtime)
- Non-root user for security
- Health check command
- Layer caching optimization
- Minimal image size
- Clear build instructions

### 6.6.4: Docker Compose Setup (~1 hour)
**Files to Create**:
- `docker-compose.yml`

**Requirements**:
- Service definition for conexus
- SQLite volume mount configuration
- Environment variable examples
- Health checks
- Restart policy
- Port mapping
- Network configuration

### 6.6.5: Integration & Testing (~1 hour)
**Files to Modify**:
- `cmd/conexus/main.go`

**Requirements**:
- Load configuration on startup
- Initialize all components with config
- Graceful shutdown (SIGTERM/SIGINT)
- Signal handling
- Cleanup on exit

**Testing**:
- [ ] Build Docker image
- [ ] Run container with defaults
- [ ] Test env var overrides
- [ ] Test config file loading
- [ ] Verify SQLite persistence
- [ ] Test graceful shutdown
- [ ] Create completion doc

---

## Configuration Schema (Initial Design)

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Indexer  IndexerConfig
    Logging  LoggingConfig
}

type ServerConfig struct {
    Host string // Default: "0.0.0.0"
    Port int    // Default: 8080
}

type DatabaseConfig struct {
    Path string // Default: "./data/conexus.db"
}

type IndexerConfig struct {
    RootPath     string // Default: "."
    ChunkSize    int    // Default: 512
    ChunkOverlap int    // Default: 50
}

type LoggingConfig struct {
    Level  string // Default: "info"
    Format string // Default: "json"
}
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONEXUS_HOST` | Server bind address | `0.0.0.0` |
| `CONEXUS_PORT` | Server port | `8080` |
| `CONEXUS_DB_PATH` | SQLite database path | `./data/conexus.db` |
| `CONEXUS_ROOT_PATH` | Codebase root to index | `.` |
| `CONEXUS_CHUNK_SIZE` | Token chunk size | `512` |
| `CONEXUS_CHUNK_OVERLAP` | Chunk overlap tokens | `50` |
| `CONEXUS_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `CONEXUS_LOG_FORMAT` | Log format (json/text) | `json` |
| `CONEXUS_CONFIG_FILE` | Path to config file | (optional) |

---

## Config File Format (YAML)

```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  path: ./data/conexus.db

indexer:
  root_path: .
  chunk_size: 512
  chunk_overlap: 50

logging:
  level: info
  format: json
```

---

## Success Criteria

### Functionality
- [ ] Config loads from env vars
- [ ] Config loads from file
- [ ] Env vars override file config
- [ ] Defaults work when nothing specified
- [ ] Validation catches invalid configs
- [ ] Docker image builds successfully
- [ ] Container runs with default config
- [ ] SQLite data persists across restarts
- [ ] Graceful shutdown works

### Quality
- [ ] Config tests at 80%+ coverage
- [ ] All tests passing
- [ ] Code follows project conventions
- [ ] Documentation complete
- [ ] Dockerfile optimized (multi-stage)
- [ ] Security: non-root user
- [ ] Health checks functional

### Documentation
- [ ] Config options documented
- [ ] Deployment instructions clear
- [ ] Docker usage examples
- [ ] Environment variable reference
- [ ] Completion document created

---

## Testing Plan

### Unit Tests (internal/config)
1. **Default Values**
   - Test all defaults applied correctly
   - Test nil/empty config handling

2. **Environment Variables**
   - Test each env var loads correctly
   - Test type conversions (string->int, etc.)
   - Test invalid values handled

3. **Config File Loading**
   - Test YAML parsing
   - Test JSON parsing (if supported)
   - Test file not found handling
   - Test invalid file content

4. **Merge Logic**
   - Test env overrides file
   - Test file overrides defaults
   - Test partial configs merge correctly

5. **Validation**
   - Test port range validation
   - Test path validation
   - Test log level validation
   - Test required field validation

### Integration Tests
1. **Docker Build**
   ```bash
   docker build -t conexus:test .
   ```

2. **Run with Defaults**
   ```bash
   docker run -p 8080:8080 conexus:test
   ```

3. **Run with Env Vars**
   ```bash
   docker run -e CONEXUS_PORT=9090 -p 9090:9090 conexus:test
   ```

4. **Run with Config File**
   ```bash
   docker run -v ./config.yml:/etc/conexus/config.yml \
              -e CONEXUS_CONFIG_FILE=/etc/conexus/config.yml \
              conexus:test
   ```

5. **Volume Persistence**
   ```bash
   docker run -v conexus-data:/data \
              -e CONEXUS_DB_PATH=/data/conexus.db \
              conexus:test
   ```

---

## Implementation Order

1. ✅ Create plan document
2. ⏳ Design configuration schema
3. ⏳ Implement config package
4. ⏳ Write config tests (80%+ coverage)
5. ⏳ Create Dockerfile
6. ⏳ Create Docker Compose
7. ⏳ Update main.go
8. ⏳ Test deployment
9. ⏳ Create completion doc

---

**Next Step**: Begin 6.6.2 (Configuration Implementation)
