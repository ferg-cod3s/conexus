# Publishing Guide

## Pre-built Binaries Setup ✅

Conexus now ships with **pre-built binaries** for all major platforms. No Go installation required for end users!

### Supported Platforms

- macOS Intel (darwin-amd64)
- macOS Apple Silicon (darwin-arm64)
- Linux x64 (linux-amd64)
- Linux ARM64 (linux-arm64)
- Windows x64 (windows-amd64)

### Package Size

- **Compressed**: ~39.6 MB
- **Uncompressed**: ~100 MB
- **Per binary**: ~19-20 MB (stripped and optimized)

## Publishing to npm

### First Time Setup

```bash
# Login to npm
npm login

# Create organization if needed (or use existing)
# Organization: @agentic-conexus
```

### Build & Publish

```bash
# 1. Build all platform binaries (automatic on prepublishOnly)
npm run build:all

# 2. Verify package contents
npm pack --dry-run

# 3. Publish
npm publish --access public
```

### Version Bumping

```bash
# Patch version (0.1.0 -> 0.1.1)
npm version patch

# Minor version (0.1.0 -> 0.2.0)
npm version minor

# Major version (0.1.0 -> 1.0.0)
npm version major

# Then publish
npm publish --access public
```

## How It Works

### 1. Build Script (`scripts/build-binaries.sh`)

Builds optimized binaries for all platforms:
- Uses `-ldflags="-s -w"` to strip debug info (saves ~30% size)
- Uses `-trimpath` for reproducible builds
- Cross-compiles for all platforms

### 2. Launcher (`bin/conexus.js`)

Smart launcher that:
- Detects user's platform and architecture
- Selects the correct pre-built binary
- Falls back gracefully if binary not found
- Provides helpful error messages

### 3. Package Configuration (`package.json`)

```json
{
  "bin": {
    "conexus": "./bin/conexus.js"
  },
  "scripts": {
    "build:all": "bash scripts/build-binaries.sh",
    "prepublishOnly": "npm run build:all"
  },
  "files": [
    "bin/",
    "README.md",
    "LICENSE",
    "config.example.yml"
  ]
}
```

### 4. NPM Ignore (`.npmignore`)

Excludes source code and includes only:
- Pre-built binaries in `bin/`
- README, LICENSE
- Example config

## User Installation

```bash
# Global install
npm install -g @agentic-conexus/mcp

# Run immediately
conexus

# Or use without install
npx @agentic-conexus/mcp
bunx @agentic-conexus/mcp
```

## Benefits

✅ **No Go installation required** for end users  
✅ **Fast installation** - no compilation step  
✅ **Cross-platform** - works everywhere  
✅ **Simple MCP integration** - just install and run  
✅ **Smaller than bundling Go** - optimized binaries only  

## Troubleshooting

### Binary too large?

Current size (~20MB per binary) is acceptable for Go applications with embedded dependencies. Further optimization possible with:
- UPX compression (not recommended - may trigger AV)
- Removing unused dependencies
- Using CGO for smaller builds (adds compilation complexity)

### Missing platform?

Add to `scripts/build-binaries.sh`:
```bash
PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "freebsd/amd64"  # Add here
)
```

### Testing locally before publish

```bash
# Pack and install locally
npm pack
npm install -g ./agentic-conexus-mcp-0.1.0-alpha.tgz

# Test it works
conexus --help
```
