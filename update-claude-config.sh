#!/bin/bash

# Update Claude Desktop configuration with Conexus MCP servers

set -e

CLAUDE_CONFIG_DIR="$HOME/Library/Application Support/Claude"
CLAUDE_CONFIG_FILE="$CLAUDE_CONFIG_DIR/claude_desktop_config.json"
BACKUP_FILE="$CLAUDE_CONFIG_FILE.backup.$(date +%Y%m%d_%H%M%S)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

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

print_info "Claude Desktop Configuration Updater for Conexus"
print_info "================================================="

# Check if Claude config directory exists
if [ ! -d "$CLAUDE_CONFIG_DIR" ]; then
    print_error "Claude Desktop config directory not found: $CLAUDE_CONFIG_DIR"
    print_info "Please make sure Claude Desktop is installed"
    exit 1
fi

# Create backup if config exists
if [ -f "$CLAUDE_CONFIG_FILE" ]; then
    print_info "Creating backup of existing configuration..."
    cp "$CLAUDE_CONFIG_FILE" "$BACKUP_FILE"
    print_success "Backup created: $BACKUP_FILE"
fi

# Start with empty config or existing config
if [ -f "$CLAUDE_CONFIG_FILE" ]; then
    # Read existing config
    config_content=$(cat "$CLAUDE_CONFIG_FILE")
else
    # Start with empty object
    config_content="{}"
fi

# Function to add/update MCP server in config
add_mcp_server() {
    local project_name="$1"
    local project_path="$2"
    local conexus_binary="/Users/johnferguson/Github/conexus/conexus"
    local config_dir="$project_path/.conexus"
    local db_path="$config_dir/conexus.db"
    
    # Check if project is configured
    if [ ! -d "$config_dir" ]; then
        print_warning "Project $project_name is not configured with Conexus (missing .conexus directory)"
        return 1
    fi
    
    # Create server configuration
    local server_config="\"conexus-$project_name\": {
      \"command\": \"$conexus_binary\",
      \"args\": [],
      \"env\": {
        \"CONEXUS_CONFIG_FILE\": \"$config_dir/config.yml\",
        \"CONEXUS_DB_PATH\": \"$db_path\",
        \"CONEXUS_LOG_LEVEL\": \"info\",
        \"CONEXUS_ROOT_PATH\": \"$project_path\"
      }
    }"
    
    # Use jq to merge configurations if available, otherwise simple approach
    if command -v jq >/dev/null 2>&1; then
        # Create temporary file with new server config
        echo "$server_config" > /tmp/new_server.json
        
        # Extract just the server part
        new_server=$(jq '.' /tmp/new_server.json)
        
        # Merge with existing config
        if echo "$config_content" | jq '.mcpServers' >/dev/null 2>&1; then
            # mcpServers exists, add to it
            config_content=$(echo "$config_content" | jq --argjson server "$new_server" '.mcpServers += $server')
        else
            # mcpServers doesn't exist, create it
            config_content=$(echo "$config_content" | jq --argjson server "$new_server" '. + {"mcpServers": $server}')
        fi
        
        rm -f /tmp/new_server.json
    else
        print_warning "jq not found, using simple configuration merge"
        # Simple approach - this is basic and may not handle all cases
        print_info "Consider installing jq for better configuration handling: brew install jq"
        
        # For now, just show what needs to be added manually
        echo ""
        echo "Please manually add the following to your Claude Desktop configuration:"
        echo "$server_config"
        echo ""
        return 0
    fi
    
    print_success "Added conexus-$project_name to configuration"
    return 0
}

# Find all configured projects
projects_found=0

# Check current directory first
current_dir=$(pwd)
current_project=$(basename "$current_dir")
if [ -d "$current_dir/.conexus" ]; then
    print_info "Found configured project: $current_project (current)"
    if add_mcp_server "$current_project" "$current_dir"; then
        ((projects_found++))
    fi
fi

# Search for other projects
for search_dir in "$HOME/Github" "$HOME/projects" "$HOME/Projects" "$HOME/code" "$HOME/workspace"; do
    if [ -d "$search_dir" ]; then
        for project_dir in "$search_dir"/*; do
            if [ -d "$project_dir" ] && [ "$project_dir" != "$current_dir" ] && [ -d "$project_dir/.conexus" ]; then
                project_name=$(basename "$project_dir")
                print_info "Found configured project: $project_name"
                if add_mcp_server "$project_name" "$project_dir"; then
                    ((projects_found++))
                fi
            fi
        done
    fi
done

if [ $projects_found -eq 0 ]; then
    print_warning "No configured Conexus projects found"
    print_info "Run setup-project-mcp.sh to configure projects first"
    exit 0
fi

# Write updated configuration
if command -v jq >/dev/null 2>&1; then
    echo "$config_content" | jq '.' > "$CLAUDE_CONFIG_FILE"
    print_success "Updated Claude Desktop configuration"
else
    print_info "Please manually update your configuration with the server details shown above"
fi

print_info ""
print_info "=== Next Steps ==="
print_info "1. Restart Claude Desktop"
print_info "2. The Conexus tools will be available for your configured projects"
print_info "3. Each project will have its own isolated database"
print_info ""

if [ -f "$BACKUP_FILE" ]; then
    print_info "Your original configuration was backed up to: $BACKUP_FILE"
fi