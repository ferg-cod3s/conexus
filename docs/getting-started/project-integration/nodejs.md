# Node.js/JavaScript Project Integration

This guide covers integrating Conexus with Node.js, JavaScript, and TypeScript projects.

## Quick Setup

### 1. Install Conexus

```bash
# Clone Conexus repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Build binaries
./scripts/build-binaries.sh

# Or install via npm (if available)
npm install -g @agentic-conexus/mcp
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
    "typescript-pro": {
      "tools": {
        "conexus": true
      }
    },
    "javascript-pro": {
      "tools": {
        "conexus": true
      }
    },
    "frontend-developer": {
      "tools": {
        "conexus": true
      }
    },
    "api-builder-enhanced": {
      "tools": {
        "conexus": true
      }
    }
  }
}
```

**For Claude Desktop** (`claude_desktop_config.json`):

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

Create `.conexus/config.yml` in your project root:

```yaml
project:
  name: "my-nodejs-app"
  description: "Node.js web application"

codebase:
  root: "."
  include_patterns:
    - "**/*.js"
    - "**/*.ts"
    - "**/*.jsx"
    - "**/*.tsx"
    - "**/*.json"
    - "**/*.md"
  exclude_patterns:
    - "**/node_modules/**"
    - "**/dist/**"
    - "**/build/**"
    - "**/coverage/**"
    - "**/.next/**"
    - "**/.nuxt/**"
    - "**/public/**"

indexing:
  auto_reindex: true
  reindex_interval: "30m"
  chunk_size: 500

search:
  max_results: 50
  similarity_threshold: 0.7
```

## Framework-Specific Examples

### Express.js API

**Project Structure:**
```
my-express-api/
├── src/
│   ├── routes/
│   ├── middleware/
│   ├── models/
│   └── controllers/
├── package.json
├── .conexus/
│   └── config.yml
└── .opencode/
    └── opencode.jsonc
```

**Recommended Queries:**
- "Find all API route handlers"
- "Show me the authentication middleware"
- "Search for database query functions"
- "Locate error handling patterns"

### Next.js Application

**Configuration:**
```yaml
# .conexus/config.yml
codebase:
  include_patterns:
    - "**/*.js"
    - "**/*.ts"
    - "**/*.jsx"
    - "**/*.tsx"
    - "pages/**/*.js"
    - "app/**/*.js"
    - "components/**/*.js"
  exclude_patterns:
    - "**/.next/**"
    - "**/out/**"
    - "**/public/**"
```

**Recommended Agents:**
- `nextjs-pro`
- `typescript-pro`
- `frontend-developer`

**Useful Queries:**
- "Find all API routes in the pages/api directory"
- "Show me the React components"
- "Search for server-side rendering logic"
- "Locate the data fetching functions"

### React Application

**Setup:**
```jsonc
// .opencode/opencode.jsonc
{
  "agent": {
    "react-pro": {
      "tools": {
        "conexus": true
      }
    },
    "typescript-pro": {
      "tools": {
        "conexus": true
      }
    }
  }
}
```

**Common Queries:**
- "Find all React components"
- "Show me the state management code"
- "Search for API calls"
- "Locate the routing configuration"

### Vue.js Application

**Configuration:**
```yaml
# .conexus/config.yml
codebase:
  include_patterns:
    - "**/*.js"
    - "**/*.ts"
    - "**/*.vue"
    - "**/*.json"
  exclude_patterns:
    - "**/dist/**"
    - "**/node_modules/**"
```

**Recommended Agents:**
- `vue-pro`
- `javascript-pro`

### NestJS Application

**Project Structure:**
```
nestjs-app/
├── src/
│   ├── modules/
│   ├── controllers/
│   ├── services/
│   ├── entities/
│   └── dto/
├── .conexus/
└── .opencode/
```

**Queries:**
- "Find all controllers and their routes"
- "Show me the service classes"
- "Search for database entities"
- "Locate the dependency injection setup"

## Development Workflow

### Package.json Scripts

Add Conexus to your development workflow:

```json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "conexus:index": "../conexus/bin/conexus-darwin-arm64 index",
    "conexus:search": "../conexus/bin/conexus-darwin-arm64 search",
    "predev": "npm run conexus:index"
  }
}
```

### VS Code Integration

Add to `.vscode/settings.json`:

```json
{
  "mcp.server.conexus": {
    "command": "conexus",
    "args": [],
    "env": {
      "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite",
      "CONEXUS_ROOT_PATH": "${workspaceFolder}"
    }
  }
}
```

### Pre-commit Hooks

Use Conexus in your pre-commit workflow:

```bash
# .husky/pre-commit
#!/bin/sh
. "$(dirname "$0")/_/husky.sh"

# Index codebase before commit
../conexus/bin/conexus-darwin-arm64 index --quiet

# Run tests
npm test
```

## Performance Optimization

### For Large Codebases

```yaml
# .conexus/config.yml
indexing:
  chunk_size: 250
  workers: 2
  memory_limit: "512MB"

search:
  max_results: 25
  cache_enabled: true
  cache_ttl: "1h"
```

### Memory Management

```bash
# Environment variables
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=256MB
export CONEXUS_INDEXING_MEMORY_LIMIT=128MB
export NODE_OPTIONS="--max-old-space-size=4096"
```

## Troubleshooting

### Common Node.js Issues

**"Cannot find module" errors:**
```bash
# Ensure dependencies are installed
npm install

# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

**TypeScript compilation errors:**
```bash
# Check TypeScript configuration
npx tsc --noEmit

# Update type definitions
npm install --save-dev @types/node@latest
```

**MCP connection issues:**
```bash
# Test Conexus directly
../conexus/bin/conexus-darwin-arm64

# Check logs
tail -f .conexus/conexus.log
```

### Framework-Specific Issues

**Next.js:**
- Ensure `.next` is excluded from indexing
- Add `pages/` and `app/` to include patterns

**Create React App:**
- Exclude `build/` and `public/` directories
- Include `src/` and configuration files

**Vue CLI:**
- Exclude `dist/` and `node_modules/`
- Include `src/` and `public/`

## Best Practices

1. **Exclude Generated Files:** Always exclude `node_modules/`, `dist/`, `build/`, and framework-specific build directories

2. **Include Source Files:** Make sure to include `.js`, `.ts`, `.jsx`, `.tsx`, and configuration files

3. **Use TypeScript:** If using TypeScript, ensure `.ts` files are included for better code understanding

4. **Regular Reindexing:** Set up automatic reindexing for active development

5. **Team Consistency:** Share `.conexus/config.yml` in version control for team consistency

## Integration Examples

### With ESLint

```javascript
// .eslintrc.js
module.exports = {
  // ... existing config
  rules: {
    // Custom rules that work with Conexus
  }
}
```

### With Jest

```javascript
// jest.config.js
module.exports = {
  // ... existing config
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  testMatch: [
    '<rootDir>/src/**/__tests__/**/*.(js|jsx|ts|tsx)',
    '<rootDir>/src/**/*.(test|spec).(js|jsx|ts|tsx)'
  ]
}
```

### With Webpack

```javascript
// webpack.config.js
module.exports = {
  // ... existing config
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  }
}
```

This integration allows Conexus to understand your Node.js/JavaScript codebase structure, dependencies, and patterns for more accurate AI assistance.