#!/bin/bash

# coolify-mcp-server Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/coolify-mcp-server"
cd "$PROJECT_DIR"

echo "Setting up Conexus for coolify-mcp-server..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/coolify-mcp-server-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# coolify-mcp-server Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "coolify-mcp-server configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"