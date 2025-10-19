#!/bin/bash

# List all projects configured with Conexus MCP

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}=== Conexus MCP Projects ===${NC}"
    echo ""
}

print_project() {
    local project_dir="$1"
    local project_name="$2"
    local config_dir="$project_dir/.conexus"
    
    if [ -d "$config_dir" ]; then
        echo -e "${GREEN}✓${NC} $project_name"
        echo "   Path: $project_dir"
        echo "   Config: $config_dir"
        
        if [ -f "$config_dir/conexus.db" ]; then
            local db_size=$(du -h "$config_dir/conexus.db" 2>/dev/null | cut -f1)
            echo "   Database: $db_size"
        else
            echo "   Database: Not created yet"
        fi
        echo ""
    fi
}

print_header

# Check current directory
current_dir=$(pwd)
current_project=$(basename "$current_dir")
print_project "$current_dir" "$current_project (current)"

# Find other projects with .conexus directories
echo -e "${YELLOW}Other configured projects:${NC}"
echo ""

# Search common project directories
for search_dir in "$HOME/Github" "$HOME/projects" "$HOME/Projects" "$HOME/code" "$HOME/workspace"; do
    if [ -d "$search_dir" ]; then
        for project_dir in "$search_dir"/*; do
            if [ -d "$project_dir" ] && [ "$project_dir" != "$current_dir" ]; then
                project_name=$(basename "$project_dir")
                if [ -d "$project_dir/.conexus" ]; then
                    print_project "$project_dir" "$project_name"
                fi
            fi
        done
    fi
done

echo -e "${BLUE}=== Setup New Project ===${NC}"
echo "To set up a new project, run:"
echo "  /Users/johnferguson/Github/conexus/setup-project-mcp.sh /path/to/your/project"