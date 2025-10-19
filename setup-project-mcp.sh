#!/bin/bash

# Conexus Project-Specific MCP Setup Script
# This script helps configure Conexus as a project-specific MCP server

set -e

CONEXUS_BINARY="/Users/johnferguson/Github/conexus/conexus"
PROJECT_DIR=""
PROJECT_NAME=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_usage() {
    echo "Usage: $0 [project-directory]"
    echo ""
    echo "Examples:"
    echo "  $0 /Users/johnferguson/Github/my-project"
    echo "  $0 .                    # Use current directory"
    echo ""
    echo "This will:"
    echo "1. Create project-specific MCP configuration"
    echo "2. Set up database path for the project"
    echo "3. Configure Claude Desktop integration"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check if conexus binary exists
if [ ! -f "$CONEXUS_BINARY" ]; then
    print_error "Conexus binary not found at $CONEXUS_BINARY"
    print_info "Please build it first with: cd /Users/johnferguson/Github/conexus && go build ./cmd/conexus"
    exit 1
fi

# Get project directory
if [ $# -eq 0 ]; then
    PROJECT_DIR=$(pwd)
else
    PROJECT_DIR="$1"
fi

# Resolve to absolute path
PROJECT_DIR=$(cd "$PROJECT_DIR" && pwd)
PROJECT_NAME=$(basename "$PROJECT_DIR")

print_info "Setting up Conexus MCP for project: $PROJECT_NAME"
print_info "Project directory: $PROJECT_DIR"

# Check if project directory exists
if [ ! -d "$PROJECT_DIR" ]; then
    print_error "Project directory does not exist: $PROJECT_DIR"
    exit 1
fi

# Create project-specific config directory
CONFIG_DIR="$PROJECT_DIR/.conexus"
DB_PATH="$PROJECT_DIR/.conexus/conexus.db"

print_info "Creating Conexus configuration directory..."
mkdir -p "$CONFIG_DIR"

# Create project-specific MCP config
MCP_CONFIG="$CONFIG_DIR/mcp-config.json"
cat > "$MCP_CONFIG" << EOF
{
  "mcpServers": {
    "conexus-$PROJECT_NAME": {
      "command": "$CONEXUS_BINARY",
      "args": [],
      "env": {
        "CONEXUS_CONFIG_FILE": "$CONFIG_DIR/config.yml",
        "CONEXUS_DB_PATH": "$DB_PATH",
        "CONEXUS_LOG_LEVEL": "info",
        "CONEXUS_ROOT_PATH": "$PROJECT_DIR"
      }
    }
  }
}
EOF

print_success "Created MCP configuration: $MCP_CONFIG"

# Create project-specific Conexus config
CONEXUS_CONFIG="$CONFIG_DIR/config.yml"
cat > "$CONEXUS_CONFIG" << EOF
# Conexus Configuration for $PROJECT_NAME

server:
  host: "0.0.0.0"
  port: 0  # stdio mode for MCP

database:
  path: "$DB_PATH"

indexer:
  root_path: "$PROJECT_DIR"
  chunk_size: 512
  chunk_overlap: 50

logging:
  level: "info"
  format: "json"

observability:
  metrics:
    enabled: false
    port: 9091
    path: "/metrics"
  tracing:
    enabled: false
    endpoint: "http://localhost:4318"
    sample_rate: 0.1
  sentry:
    enabled: false
    dsn: ""
EOF

print_success "Created Conexus configuration: $CONEXUS_CONFIG"

# Create .gitignore entry for .conexus directory
GITIGNORE="$PROJECT_DIR/.gitignore"
if [ -f "$GITIGNORE" ]; then
    if ! grep -q "^\.conexus/" "$GITIGNORE"; then
        echo "" >> "$GITIGNORE"
        echo "# Conexus project-specific data" >> "$GITIGNORE"
        echo ".conexus/" >> "$GITIGNORE"
        print_success "Added .conexus/ to .gitignore"
    else
        print_info ".conexus/ already in .gitignore"
    fi
else
    echo "# Conexus project-specific data" > "$GITIGNORE"
    echo ".conexus/" >> "$GITIGNORE"
    print_success "Created .gitignore with .conexus/ entry"
fi

# Test the configuration
print_info "Testing Conexus configuration..."
cd "$PROJECT_DIR"
if timeout 5s "$CONEXUS_BINARY" --help > /dev/null 2>&1; then
    print_success "Conexus binary is working"
else
    print_warning "Conexus test failed (this may be normal if it expects stdio input)"
fi

# Instructions for Claude Desktop
print_info ""
print_info "=== Claude Desktop Setup Instructions ==="
print_info ""
print_info "To use this project with Claude Desktop, add the following to your Claude Desktop configuration:"
print_info ""
echo -e "${YELLOW}# Add this to ~/Library/Application Support/Claude/claude_desktop_config.json${NC}"
echo "{"
echo "  \"mcpServers\": {"
echo "    \"conexus-$PROJECT_NAME\": {"
echo "      \"command\": \"$CONEXUS_BINARY\","
echo "      \"args\": [],"
echo "      \"env\": {"
echo "        \"CONEXUS_CONFIG_FILE\": \"$CONFIG_DIR/config.yml\","
echo "        \"CONEXUS_DB_PATH\": \"$DB_PATH\","
echo "        \"CONEXUS_LOG_LEVEL\": \"info\","
echo "        \"CONEXUS_ROOT_PATH\": \"$PROJECT_DIR\""
echo "      }"
echo "    }"
echo "  }"
echo "}"
print_info ""
print_info "Or simply copy the contents of: $MCP_CONFIG"
print_info ""

# Instructions for usage
print_info "=== Usage Instructions ==="
print_info ""
print_info "1. Restart Claude Desktop after updating the configuration"
print_info "2. Claude will now have access to this specific project's codebase"
print_info "3. The database will be created automatically at: $DB_PATH"
print_info "4. Use the MCP tools to index and search your project"
print_info ""

print_success "Setup complete for project: $PROJECT_NAME"
print_info "Database will be created at: $DB_PATH"
print_info "Configuration files are in: $CONFIG_DIR"