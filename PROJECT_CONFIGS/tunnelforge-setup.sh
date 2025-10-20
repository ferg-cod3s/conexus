#!/bin/bash

# tunnelforge Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/tunnelforge"
cd "$PROJECT_DIR"

echo "Setting up Conexus for tunnelforge..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/tunnelforge-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# tunnelforge Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "tunnelforge configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"