# Conexus Universal Setup Script

A comprehensive setup script that automatically configures Conexus for any project by detecting the project type and creating appropriate configurations.

## Features

- **Automatic Project Detection**: Detects project type from common configuration files
- **Smart Agent Configuration**: Includes relevant agent tools based on project type
- **Idempotent**: Can be run multiple times safely
- **Cross-Platform**: Works on macOS, Linux, and Windows (with appropriate binary paths)
- **Sentry Integration**: Pre-configured with Sentry DSN for error tracking
- **Environment Variables**: Creates `.env` file with all necessary variables

## Supported Project Types

| Project Type | Detection File | Included Agents |
|-------------|----------------|-----------------|
| **Node.js** | `package.json` | typescript-pro, javascript-pro, frontend-developer, api-builder-enhanced, astro-pro, nextjs-pro, react-pro, vue-pro, node-js-developer |
| **Go** | `go.mod` | go-expert, golang-developer, api-builder-enhanced, backend-developer |
| **Python** | `pyproject.toml` | python-pro, python-developer, api-builder-enhanced, data-scientist, backend-developer |
| **Rust** | `Cargo.toml` | rust-pro, rust-developer, systems-programmer, backend-developer |
| **PHP** | `composer.json` | php-pro, php-developer, laravel-pro, backend-developer |
| **Ruby** | `Gemfile` | ruby-pro, ruby-on-rails-pro, ruby-developer, backend-developer |
| **Java** | `pom.xml`, `build.gradle*` | java-pro, spring-boot-pro, java-developer, backend-developer |
| **Docker** | `Dockerfile`, `docker/` | devops-engineer, infrastructure-developer, container-specialist |
| **Generic** | Fallback | full-stack-developer, software-engineer |

All configurations include the `conexus-expert` agent for advanced Conexus operations.

## Usage

### Basic Setup

1. Copy the script to your project directory:
   ```bash
   cp PROJECT_CONFIGS/setup-conexus.sh /path/to/your/project/
   cd /path/to/your/project
   ```

2. Make it executable and run:
   ```bash
   chmod +x setup-conexus.sh
   ./setup-conexus.sh
   ```

### What It Creates

The script creates the following files and directories:

```
your-project/
├── .opencode/
│   └── opencode.jsonc    # MCP configuration with project-specific agents
├── data/                 # Directory for Conexus database
│   └── conexus.db        # SQLite database (created later by Conexus)
└── .env                  # Environment variables
```

### Configuration Details

#### MCP Configuration (.opencode/opencode.jsonc)
- Uses local Conexus binary: `../bin/conexus-darwin-arm64`
- Pre-configured with Sentry DSN for error tracking
- Environment variables with fallbacks
- Project-type-specific agent tools enabled

#### Environment Variables (.env)
```bash
# Core Conexus settings
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0                    # Use stdio mode
CONEXUS_LOG_LEVEL=info

# Optional Sentry configuration (commented out by default)
# CONEXUS_SENTRY_ENABLED=true
# CONEXUS_SENTRY_DSN=https://7e54c8bc81fb554a460d4331e5c23fe0@sentry.fergify.work/15
# CONEXUS_SENTRY_ENVIRONMENT=development
```

## Requirements

- Conexus binary available at `../bin/conexus-darwin-arm64` (relative to project root)
- Bash shell
- Standard Unix tools (`mkdir`, `cat`, `chmod`)

## Customization

### Binary Path
If your Conexus binary is in a different location, edit the script:

```bash
# Change this line in the script:
local binary_path="../bin/conexus-darwin-arm64"
# To your actual path:
local binary_path="/absolute/path/to/conexus"
```

### Sentry Configuration
The script includes Sentry configuration for error tracking. To disable:

1. Comment out or remove the Sentry-related environment variables in `.env`
2. Or set `CONEXUS_SENTRY_ENABLED=false` in the MCP configuration

### Adding New Project Types

To add support for a new project type:

1. Add detection logic in the `detect_project_type()` function
2. Add agent configuration in the `get_agent_config()` function
3. Test with a sample project

Example:
```bash
# In detect_project_type()
elif [ -f "some-config-file" ]; then
    project_type="newtype"

# In get_agent_config()
"newtype")
    cat << 'EOF'
"newtype-expert": {
  "tools": {
    "conexus": true
  }
},
EOF
    ;;
```

## Troubleshooting

### Binary Not Found
```
✗ Conexus binary not found at ../bin/conexus-darwin-arm64
```

**Solution**: Ensure the Conexus binary exists at the expected path or update the `binary_path` variable in the script.

### Permission Denied
```
bash: ./setup-conexus.sh: Permission denied
```

**Solution**: Make the script executable:
```bash
chmod +x setup-conexus.sh
```

### Project Not Detected
If your project type isn't detected correctly:

1. Check that the expected configuration file exists
2. Verify the file format matches expectations
3. The script falls back to "generic" type with basic agents

### JSON Parsing Errors
For Node.js/PHP projects, ensure `jq` is available:
```bash
# Install jq if needed
brew install jq  # macOS
apt-get install jq  # Ubuntu/Debian
```

## Examples

### Node.js Project
```bash
$ ./setup-conexus.sh
ℹ Detected project: my-node-app
ℹ Project type: nodejs
✓ Conexus binary found and ready
✓ Setup completed successfully!
```

Creates configuration with TypeScript, JavaScript, and frontend agents.

### Go Project
```bash
$ ./setup-conexus.sh
ℹ Detected project: my-go-service
ℹ Project type: go
✓ Conexus binary found and ready
✓ Setup completed successfully!
```

Creates configuration with Go and backend development agents.

### Python Project
```bash
$ ./setup-conexus.sh
ℹ Detected project: ml-pipeline
ℹ Project type: python
✓ Conexus binary found and ready
✓ Setup completed successfully!
```

Creates configuration with Python, data science, and backend agents.

## Integration with MCP Clients

After setup, use the generated `.opencode/opencode.jsonc` with:

- **Cursor**: Load the configuration file
- **Claude Code**: Reference the `.opencode` directory
- **Other MCP clients**: Import the JSON configuration

## Security Notes

- The script sets up local-only Conexus operation (stdio mode)
- No network calls are made during setup
- Sentry DSN is included but commented out by default
- All data stays local to your machine

## Contributing

To improve the setup script:

1. Test with additional project types
2. Add more agent configurations
3. Improve detection logic
4. Update documentation

## License

Same as Conexus project - MIT License.