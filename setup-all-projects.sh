#!/bin/bash

# Master setup script for all Conexus project configurations
set -e

echo "🚀 Setting up Conexus for all GitHub projects..."
echo "=================================================="

# Function to run setup if project directory exists
setup_project() {
    local project_name="$1"
    local project_path="/Users/johnferguson/Github/$project_name"
    
    if [ -d "$project_path" ]; then
        echo "📁 Setting up $project_name..."
        cd "$project_path"
        
        # Create .opencode directory
        mkdir -p .opencode
        
        # Copy configuration if it exists
        if [ -f "/Users/johnferguson/Github/conexus/PROJECT_CONFIGS/${project_name}-opencode.jsonc" ]; then
            cp "/Users/johnferguson/Github/conexus/PROJECT_CONFIGS/${project_name}-opencode.jsonc" .opencode/opencode.jsonc
        fi
        
        # Create data directory
        mkdir -p data
        
        # Create environment file if it doesn't exist
        if [ ! -f .env ]; then
            cat > .env << EOF
# $project_name Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF
        fi
        
        echo "✅ $project_name configured successfully!"
    else
        echo "⚠️  $project_name not found, skipping..."
    fi
}

# High-priority Go projects
echo "🔧 Configuring Go projects..."
setup_project "LocalHaven-CMS"
setup_project "gotunnel"
setup_project "rune"
setup_project "tunnelforge"

# Rust projects
echo "🦀 Configuring Rust projects..."
setup_project "f3rg-redis"
setup_project "advent_of_code"

# TypeScript projects
echo "📜 Configuring TypeScript projects..."
setup_project "hello-zero-example"
setup_project "coolify-mcp-server"
setup_project "opencode-nexus"

echo ""
echo "🎉 Configuration complete!"
echo ""
echo "Next steps:"
echo "1. Move global configuration files (see GLOBAL_SETUP_INSTRUCTIONS.md)"
echo "2. Test each project by running 'source .env' in project directories"
echo "3. Use '@conexus-expert' in OpenCode sessions to test functionality"
echo ""
echo "Example usage:"
echo "  cd ~/Github/LocalHaven-CMS"
echo "  source .env"
echo "  opencode"
echo "  @conexus-expert analyze this codebase"