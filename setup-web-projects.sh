#!/bin/bash

# Setup script for remaining web projects
set -e

echo "🌐 Setting up Conexus for remaining web projects..."
echo "=================================================="

# List of remaining web projects
WEB_PROJECTS=(
    "jferguson.info"
    "valkyrie-fitness"
    "spring-creek-baptist"
    "mux-otw"
    "unFergettable2018"
    "pie"
    "ogdrip"
    "zero-docs"
    "sand-and-sagebrush"
)

# Function to setup web project
setup_web_project() {
    local project_name="$1"
    local project_path="/Users/johnferguson/Github/$project_name"
    
    if [ -d "$project_path" ]; then
        echo "📁 Setting up $project_name..."
        cd "$project_path"
        
        # Create .opencode directory
        mkdir -p .opencode
        
        # Copy generic web configuration
        cp "/Users/johnferguson/Github/conexus/PROJECT_CONFIGS/web-project-opencode.jsonc" .opencode/opencode.jsonc
        
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

# Setup all web projects
for project in "${WEB_PROJECTS[@]}"; do
    setup_web_project "$project"
done

echo ""
echo "🎉 Web projects configuration complete!"
echo ""
echo "These projects are now ready for Conexus semantic search:"