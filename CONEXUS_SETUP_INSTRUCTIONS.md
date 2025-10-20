# Conexus Global Agent Setup Instructions

## Overview
This setup creates a global Conexus expert agent that can work with any project's local Conexus database.

## Files Created

### 1. Global Agent Configuration
- **File**: `~/.config/opencode/agent/conexus-expert.md`
- **Purpose**: Defines the Conexus expert subagent with read-only permissions
- **Features**: Semantic search, code analysis, indexing capabilities

### 2. Global OpenCode Configuration
- **File**: `~/.config/opencode/opencode.jsonc`
- **Purpose**: Main configuration with global MCP servers and agents
- **Includes**: Context7, Grep by Vercel, and Conexus expert agent

### 3. Project-Specific Configuration
- **File**: `.opencode/opencode.jsonc` (per project)
- **Purpose**: Enables Conexus MCP server with project-specific database path
- **Environment**: Uses `{env:CONEXUS_DB_PATH:./data/conexus.db}` for flexible database location

## Installation Steps

### Step 1: Move Global Configuration
```bash
# Move the global config to the correct location
mv ./global-opencode.jsonc ~/.config/opencode/opencode.jsonc

# Move the agent to the correct location  
mv ./conexus-expert.md ~/.config/opencode/agent/
```

### Step 2: Configure Each Project
For each GitHub project that will use Conexus:

```bash
# Navigate to the project directory
cd /path/to/your/project

# Run the setup script
./setup-conexus-env.sh

# Or manually set the environment variable
export CONEXUS_DB_PATH="$(pwd)/data/conexus.db"
```

### Step 3: Create Project-Specific Config
Create `.opencode/opencode.jsonc` in each project:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["./conexus"],
      "enabled": true,
      "environment": {
        "CONEXUS_PORT": "0",
        "CONEXUS_DB_PATH": "{env:CONEXUS_DB_PATH:./data/conexus.db}",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  },
  "agent": {
    "conexus-expert": {
      "tools": {
        "conexus": true
      }
    }
  }
}
```

## Usage

### Using the Conexus Expert Agent
```bash
# In any OpenCode session with Conexus configured:
@conexus-expert analyze this codebase and find similar patterns
@conexus-expert search for functions that handle authentication
@conexus-expert index the current project structure
```

### Environment Variables
- `CONEXUS_DB_PATH`: Path to the SQLite database (project-specific)
- `CONEXUS_PORT`: Server port (0 for auto-assignment)
- `CONEXUS_LOG_LEVEL`: Logging level (info, debug, error)

### Project Structure
```
your-project/
├── .opencode/
│   └── opencode.jsonc
├── data/
│   └── conexus.db
├── .env
└── setup-conexus-env.sh
```

## Benefits

1. **Global Agent**: Use the same Conexus expertise across all projects
2. **Project Isolation**: Each project has its own database
3. **Flexible Paths**: Environment variables allow custom database locations
4. **Read-Only Safety**: Agent has restricted permissions for safe analysis
5. **Semantic Search**: Leverage vector-based code understanding

## Troubleshooting

### Database Not Found
```bash
# Create the data directory and initialize database
mkdir -p data
./conexus --init
```

### MCP Server Not Starting
```bash
# Check the Conexus binary path and permissions
ls -la ./conexus
./conexus --help
```

### Environment Variables Not Loading
```bash
# Source the .env file
source .env

# Or export manually
export CONEXUS_DB_PATH="$(pwd)/data/conexus.db"
```

## Next Steps

1. Test the setup by running `@conexus-expert` in an OpenCode session
2. Index your first codebase using the Conexus tools
3. Explore semantic search capabilities
4. Customize the agent prompt for your specific needs