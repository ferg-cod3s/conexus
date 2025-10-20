#!/bin/bash

# opencode-nexus Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/opencode-nexus"
cd "$PROJECT_DIR"

echo "Setting up Conexus for opencode-nexus..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/opencode-nexus-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# opencode-nexus Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "opencode-nexus configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"