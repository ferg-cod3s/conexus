# Complete Conexus Setup Guide

## 🚀 Quick Start

### 1. Global Setup (Run Once)

```bash
# Create global directories
mkdir -p ~/.config/opencode/agent

# Move global configuration
mv ./global-opencode.jsonc ~/.config/opencode/opencode.jsonc

# Move conexus expert agent
mv ./conexus-expert.md ~/.config/opencode/agent/

# Make setup script available globally
mkdir -p ~/.local/bin
mv ./setup-conexus-env.sh ~/.local/bin/conexus-setup
```

### 2. Project Setup

#### High-Priority Projects (Recommended)
```bash
# Run the master setup script
./setup-all-projects.sh
```

#### All Web Projects (Optional)
```bash
# Setup remaining web projects
./setup-web-projects.sh
```

## 📁 Project Configurations Created

### High-Priority Projects
- ✅ **LocalHaven-CMS** - Go web application
- ✅ **gotunnel** - Go networking project
- ✅ **rune** - Go project
- ✅ **tunnelforge** - Go + Tauri desktop app
- ✅ **f3rg-redis** - Rust Redis implementation
- ✅ **advent_of_code** - Rust algorithm solutions
- ✅ **hello-zero-example** - Astro + TypeScript
- ✅ **coolify-mcp-server** - MCP server in TypeScript
- ✅ **opencode-nexus** - Tauri + TypeScript app

### Web Projects (Optional)
- ✅ **jferguson.info** - Personal site
- ✅ **valkyrie-fitness** - Fitness site
- ✅ **spring-creek-baptist** - Church site
- ✅ **mux-otw** - Video project
- ✅ **unFergettable2018** - Legacy site
- ✅ **pie** - Web project
- ✅ **ogdrip** - Web project
- ✅ **zero-docs** - Documentation
- ✅ **sand-and-sagebrush** - Web project

## 🧪 Testing Your Setup

### Test Individual Project
```bash
cd ~/Github/LocalHaven-CMS
source .env
opencode
```

Once in OpenCode, test with:
```
@conexus-expert analyze this codebase and find the main entry points
@conexus-expert search for authentication-related functions
@conexus-expert index the current project structure
```

### Test Global Agent
```bash
cd any-project
opencode
```

Test with:
```
@conexus-expert help me understand this codebase
```

## 🔧 Configuration Details

### Environment Variables
Each project now has a `.env` file with:
```bash
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
```

### Project Structure
```
project/
├── .opencode/
│   └── opencode.jsonc
├── data/
│   └── conexus.db (created on first use)
├── .env
└── [your existing files]
```

### Agent Access
- **@conexus-expert** - Available globally for all projects
- **Language-specific agents** - Available per project (go-expert, rust-pro, typescript-pro, etc.)

## 🎯 Usage Examples

### Code Analysis
```bash
@conexus-expert analyze the architecture of this Go application
@conexus-expert find all database connection patterns
@conexus-expert identify security-related code
```

### Semantic Search
```bash
@conexus-expert search for functions that handle user authentication
@conexus-expert find code related to error handling
@conexus-expert locate API endpoint definitions
```

### Cross-Project Learning
```bash
@conexus-expert how is authentication implemented in other Go projects?
@conexus-expert show me patterns for database connections across my codebases
```

## 🛠️ Troubleshooting

### Database Not Found
```bash
# Create data directory
mkdir -p data

# Initialize database (if needed)
./conexus --init
```

### MCP Server Not Starting
```bash
# Check Conexus binary
ls -la /Users/johnferguson/Github/conexus/cmd/conexus/conexus

# Test manually
go run /Users/johnferguson/Github/conexus/cmd/conexus --help
```

### Environment Variables Not Loading
```bash
# Source the environment file
source .env

# Check variables
echo $CONEXUS_DB_PATH
```

### Agent Not Available
```bash
# Check global configuration
cat ~/.config/opencode/opencode.jsonc

# Restart OpenCode to reload configuration
```

## 📈 Benefits Achieved

1. **Semantic Code Search** - Find code by functionality, not just names
2. **Cross-Project Pattern Recognition** - Learn from all your projects
3. **Automated Code Analysis** - Understand complex codebases quickly
4. **Language-Specific Expertise** - Tailored agents for each tech stack
5. **Project Isolation** - Each project has its own database and context

## 🔄 Ongoing Maintenance

### Adding New Projects
```bash
# Copy the appropriate template
cp ./PROJECT_CONFIGS/web-project-opencode.jsonc ~/Github/new-project/.opencode/opencode.jsonc

# Or run the setup script
cd ~/Github/new-project
~/.local/bin/conexus-setup
```

### Updating Configuration
- Global agents: `~/.config/opencode/agent/`
- Global config: `~/.config/opencode/opencode.jsonc`
- Project configs: `project/.opencode/opencode.jsonc`

## 🎉 You're All Set!

Your GitHub portfolio is now equipped with powerful semantic search and code analysis capabilities. Start exploring your codebases in new ways with the Conexus expert agent!