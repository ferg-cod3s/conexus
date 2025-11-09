# Rust Project Integration

This guide covers integrating Conexus with Rust projects, including command-line applications, web services, and systems programming.

## Quick Setup

### 1. Install Conexus

```bash
# Clone Conexus repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Build binaries
./scripts/build-binaries.sh

# Or install via cargo (if available)
cargo install conexus-mcp
```

### 2. Configure MCP Client

**For OpenCode** (`.opencode/opencode.jsonc`):

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["bunx", "-y", "@agentic-conexus/mcp"],
      "environment": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "."
      },
      "enabled": true
    }
  },
  "agent": {
    "rust-pro": {
      "tools": {
        "conexus": true
      }
    },
    "systems-programmer": {
      "tools": {
        "conexus": true
      }
    }
  }
}
```

**For Claude Desktop:**

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "/path/to/project/.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "/path/to/project"
      }
    }
  }
}
```

**For Claude Code** (`~/.claude/mcp.json`):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_ROOT_PATH": "."
      }
    }
  }
}
```

### 3. Project Configuration

Create `.conexus/config.yml`:

```yaml
project:
  name: "my-rust-app"
  description: "Rust application"

codebase:
  root: "."
  include_patterns:
    - "**/*.rs"
    - "**/Cargo.toml"
    - "**/Cargo.lock"
    - "**/*.md"
    - "**/Makefile"
    - "**/Dockerfile"
  exclude_patterns:
    - "**/target/**"
    - "**/Cargo.lock"
    - "**/.git/**"
    - "**/docs/**"

indexing:
  auto_reindex: true
  reindex_interval: "40m"
  chunk_size: 550

search:
  max_results: 50
  similarity_threshold: 0.7
```

## Framework-Specific Examples

### Actix Web Application

**Project Structure:**
```
actix-web-app/
├── Cargo.toml
├── src/
│   ├── main.rs
│   ├── handlers/
│   │   ├── mod.rs
│   │   └── user.rs
│   ├── models/
│   │   ├── mod.rs
│   │   └── user.rs
│   └── lib.rs
├── .conexus/
└── .opencode/
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.rs"
    - "**/Cargo.toml"
  exclude_patterns:
    - "**/target/**"
    - "**/Cargo.lock"
```

**Recommended Agents:**
- `actix-pro`
- `rust-pro`

**Queries:**
- "Find all HTTP handlers"
- "Show me the route definitions"
- "Search for middleware"
- "Locate the application setup"

### Rocket Web Framework

**Setup:**
```rust
// main.rs
#[macro_use] extern crate rocket;

#[get("/")]
fn index() -> &'static str {
    "Hello, world!"
}

#[launch]
fn rocket() -> _ {
    rocket::build().mount("/", routes![index])
}
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.rs"
    - "**/Cargo.toml"
  exclude_patterns:
    - "**/target/**"
```

**Queries:**
- "Find all Rocket routes"
- "Show me the launch configuration"
- "Search for request guards"
- "Locate fairing implementations"

### Axum Web Framework

**Project Structure:**
```
axum-app/
├── Cargo.toml
├── src/
│   ├── main.rs
│   ├── routes/
│   ├── handlers/
│   └── models/
└── .conexus/
```

**Recommended Agents:**
- `axum-pro`
- `rust-pro`

**Queries:**
- "Find all route handlers"
- "Show me the router setup"
- "Search for middleware layers"
- "Locate extractor implementations"

### Tokio Async Runtime

**Setup:**
```rust
// main.rs
use tokio::net::TcpListener;
use tokio::io::{AsyncReadExt, AsyncWriteExt};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let listener = TcpListener::bind("127.0.0.1:8080").await?;
    loop {
        let (mut socket, _) = listener.accept().await?;
        tokio::spawn(async move {
            // Handle connection
        });
    }
}
```

**Queries:**
- "Find all async functions"
- "Show me tokio spawn usage"
- "Search for channel implementations"
- "Locate error handling"

### Command-Line Applications

**Using Clap:**
```rust
// main.rs
use clap::{Arg, Command};

fn main() {
    let matches = Command::new("myapp")
        .version("1.0")
        .author("Author")
        .about("Does awesome things")
        .arg(Arg::new("input")
            .short('i')
            .long("input")
            .value_name("FILE")
            .help("Sets the input file"))
        .get_matches();
}
```

**Configuration:**
```yaml
codebase:
  include_patterns:
    - "**/*.rs"
    - "**/Cargo.toml"
    - "**/README.md"
  exclude_patterns:
    - "**/target/**"
```

**Recommended Agents:**
- `cli-developer`
- `rust-pro`

**Queries:**
- "Find all clap command definitions"
- "Show me the argument parsing"
- "Search for subcommands"
- "Locate help text"

## Development Workflow

### Cargo Setup

```bash
# Create new project
cargo new my-rust-app
cd my-rust-app

# Add dependencies
cargo add actix-web
cargo add tokio
cargo add serde --features json

# Build project
cargo build

# Run tests
cargo test

# Run application
cargo run
```

### Cargo.toml Configuration

```toml
# Cargo.toml
[package]
name = "my-rust-app"
version = "0.1.0"
edition = "2021"

[dependencies]
actix-web = "4.0"
tokio = { version = "1.0", features = ["full"] }
serde = { version = "1.0", features = ["derive"] }
anyhow = "1.0"

[dev-dependencies]
tokio-test = "0.4"
```

### Pre-commit Hooks

```bash
# .githooks/pre-commit
#!/bin/bash

# Index codebase
../conexus/bin/conexus-darwin-arm64 index --quiet

# Run tests
cargo test

# Run clippy
cargo clippy -- -D warnings

# Format code
cargo fmt
```

### VS Code Integration

```json
// .vscode/settings.json
{
  "rust-analyzer.checkOnSave.command": "clippy",
  "rust-analyzer.cargo.features": "all",
  "mcp.server.conexus": {
    "command": "npx",
    "args": ["-y", "@agentic-conexus/mcp"],
    "env": {
      "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite",
      "CONEXUS_ROOT_PATH": "${workspaceFolder}"
    }
  }
}

// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug",
      "type": "lldb",
      "request": "launch",
      "program": "${workspaceFolder}/target/debug/${workspaceFolderBasename}",
      "args": [],
      "cwd": "${workspaceFolder}",
      "env": {
        "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite"
      }
    }
  ]
}
```

## Testing Integration

### Test Structure

```rust
// lib.rs or main.rs
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_add() {
        assert_eq!(add(1, 2), 3);
    }

    #[tokio::test]
    async fn test_async_function() {
        // Async test
    }
}
```

**Conexus can help with:**
- "Find all test functions"
- "Show me the test modules"
- "Search for mocked dependencies"
- "Locate integration tests"

### Benchmarking

```rust
// benches/my_benchmark.rs
use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn fibonacci(n: u64) -> u64 {
    match n {
        0 => 1,
        1 => 1,
        n => fibonacci(n-1) + fibonacci(n-2),
    }
}

fn criterion_benchmark(c: &mut Criterion) {
    c.bench_function("fib 20", |b| b.iter(|| fibonacci(black_box(20))));
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);
```

## Performance Optimization

### For Large Rust Codebases

```yaml
# .conexus/config.yml
indexing:
  chunk_size: 450
  workers: 3
  memory_limit: "1GB"

search:
  max_results: 35
  cache_enabled: true
  cache_ttl: "1h"

codebase:
  exclude_patterns:
    - "**/target/**"
    - "**/docs/**"
    - "**/Cargo.lock"
```

### Memory Management

```bash
# Environment variables
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=1GB
export CONEXUS_INDEXING_MEMORY_LIMIT=512MB

# Cargo build optimizations
export RUSTFLAGS="-C target-cpu=native"
```

## Troubleshooting

### Common Rust Issues

**Cargo issues:**
```bash
# Clean build artifacts
cargo clean

# Update dependencies
cargo update

# Check for outdated dependencies
cargo outdated

# Fix dependency issues
cargo tree
```

**Compilation errors:**
```bash
# Check Rust version
rustc --version
cargo --version

# Update Rust
rustup update

# Check for missing dependencies
cargo check
```

**Test failures:**
```bash
# Run tests with verbose output
cargo test -- --nocapture

# Run specific test
cargo test test_add

# Run with backtrace
RUST_BACKTRACE=1 cargo test
```

### Framework-Specific Issues

**Actix Web:**
- Include all handler functions
- Check for actor implementations
- Verify middleware setup

**Rocket:**
- Include route attributes
- Check for fairing implementations
- Verify request guards

**Axum:**
- Include all route handlers
- Check for extractor implementations
- Verify middleware layers

## Best Practices

1. **Cargo.toml:** Always include `Cargo.toml` for dependency understanding

2. **Target Directory:** Exclude `target/` but include source files

3. **Test Files:** Include test modules and integration tests

4. **Documentation:** Include `README.md`, rustdoc comments, and examples

5. **Build Files:** Include `Makefile`, `Dockerfile`, and CI configuration

## Integration Examples

### With Cargo Watch

```bash
# Install cargo watch
cargo install cargo-watch

# Watch and run tests
cargo watch -x test

# Watch and run application
cargo watch -x run
```

### With Clippy

```toml
# Cargo.toml
[lints.clippy]
pedantic = "warn"
nursery = "warn"
```

### With Rustfmt

```toml
# rustfmt.toml
edition = "2021"
max_width = 100
use_small_heuristics = "Max"
```

### With Docker

```dockerfile
# Dockerfile
FROM rust:1.70-slim as builder

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release

COPY src ./src
RUN touch src/main.rs
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=builder /app/target/release/myapp /usr/local/bin/myapp
COPY --from=builder /app/.conexus ./.conexus

CMD ["myapp"]
```

### With GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Rust
        uses: actions-rust-lang/setup-rust-toolchain@v1
        with:
          components: rustfmt, clippy

      - name: Cache dependencies
        uses: actions/cache@v3
        with:
          path: |
            ~/.cargo/registry
            ~/.cargo/git
            target
          key: ${{ runner.os }}-cargo-${{ hashFiles('**/Cargo.lock') }}

      - name: Index with Conexus
        run: |
          curl -fsSL https://raw.githubusercontent.com/ferg-cod3s/conexus/main/setup-conexus.sh | bash
          ./conexus index --quiet

      - name: Check formatting
        run: cargo fmt --check

      - name: Run clippy
        run: cargo clippy -- -D warnings

      - name: Run tests
        run: cargo test

      - name: Run benchmarks
        run: cargo bench
```

This integration allows Conexus to understand Rust codebases, Cargo workspace structures, async patterns, and Rust-specific development practices for enhanced AI assistance.