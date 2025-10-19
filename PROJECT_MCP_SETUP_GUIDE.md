# Project-Specific Conexus MCP Setup Guide

## Overview

Conexus is now configured as a **project-specific MCP server**. Each project gets its own isolated database and configuration, ensuring no conflicts between different codebases.

## Quick Setup for Your Projects

### 1. Set Up a New Project

```bash
# Navigate to your project directory
cd /path/to/your/project

# Run the setup script
/Users/johnferguson/Github/conexus/setup-project-mcp.sh
```

### 2. Update Claude Desktop Configuration

```bash
# Automatically update Claude Desktop with all configured projects
/Users/johnferguson/Github/conexus/update-claude-config.sh
```

### 3. Restart Claude Desktop

After updating the configuration, restart Claude Desktop to load the new MCP servers.

## What Gets Created

For each project, the setup creates:

```
your-project/
├── .conexus/                    # Project-specific Conexus data
│   ├── config.yml              # Project configuration
│   ├── mcp-config.json         # MCP server configuration
│   └── conexus.db              # Project database (created on first run)
└── .gitignore                  # Updated to ignore .conexus/
```

## Current Configuration

Your Claude Desktop is now configured with:

- **conexus-conexus**: For the Conexus project itself
- **codeflow-tools**: Your existing MCP server (preserved)

## Managing Multiple Projects

### List All Configured Projects

```bash
/Users/johnferguson/Github/conexus/list-projects.sh
```

### Add a New Project

```bash
/Users/johnferguson/Github/conexus/setup-project-mcp.sh /path/to/new/project
/Users/johnferguson/Github/conexus/update-claude-config.sh
```

### Remove a Project

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` and remove the project's server configuration.

## Usage in Claude

Once configured, you can use Conexus tools in Claude:

1. **Index your project**: "Please index this project for search"
2. **Search code**: "Search for functions related to authentication"
3. **Get context**: "What files are related to this component?"
4. **Find related info**: "Show me related code for this file"

Each project operates independently with its own:
- **Database**: Isolated vector embeddings
- **Configuration**: Project-specific settings
- **Index**: Separate file indexing

## Environment Variables

The setup automatically configures these environment variables for each project:

- `CONEXUS_CONFIG_FILE`: Path to project-specific config
- `CONEXUS_DB_PATH`: Path to project-specific database
- `CONEXUS_LOG_LEVEL`: Logging level (info)
- `CONEXUS_ROOT_PATH`: Project root directory

## Troubleshooting

### Database Not Created

The database is created automatically when Conexus first runs. If you see "Database: Not created yet" in the project list, just restart Claude Desktop and try using a Conexus tool.

### MCP Server Not Loading

1. Check that the binary path is correct: `/Users/johnferguson/Github/conexus/conexus`
2. Verify the configuration file syntax is valid JSON
3. Restart Claude Desktop

### Permission Issues

```bash
# Fix permissions if needed
chmod +x /Users/johnferguson/Github/conexus/conexus
chmod -R 755 /path/to/project/.conexus
```

## Advanced Configuration

### Custom Database Location

Edit the project's `.conexus/config.yml`:

```yaml
database:
  path: "/custom/path/to/database.db"
```

### Custom Indexing Settings

```yaml
indexer:
  root_path: "/path/to/project"
  chunk_size: 1024        # Larger chunks for more context
  chunk_overlap: 100      # More overlap for better continuity
```

### Enable Observability

```yaml
observability:
  metrics:
    enabled: true
    port: 9091
  tracing:
    enabled: true
    endpoint: "http://localhost:4318"
```

## Best Practices

1. **One Project Per Repository**: Each Git repository gets its own Conexus configuration
2. **Regular Updates**: Re-run the setup script when projects change significantly
3. **Database Backups**: Important project data is in `.conexus/conexus.db`
4. **Git Integration**: The `.conexus/` directory is automatically ignored

## File Locations

- **Conexus Binary**: `/Users/johnferguson/Github/conexus/conexus`
- **Setup Scripts**: `/Users/johnferguson/Github/conexus/setup-project-mcp.sh`
- **Project Lister**: `/Users/johnferguson/Github/conexus/list-projects.sh`
- **Config Updater**: `/Users/johnferguson/Github/conexus/update-claude-config.sh`
- **Claude Config**: `~/Library/Application Support/Claude/claude_desktop_config.json`

## Next Steps

1. **Set up your main projects** using the setup script
2. **Test the integration** by asking Claude to search your code
3. **Customize settings** for each project as needed
4. **Enjoy project-specific AI assistance** with isolated contexts!

Your Conexus MCP server is now ready for project-specific use across all your development projects!