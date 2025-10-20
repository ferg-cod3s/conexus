#!/bin/bash

# LocalHaven-CMS Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/LocalHaven-CMS"
cd "$PROJECT_DIR"

echo "Setting up Conexus for LocalHaven-CMS..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/LocalHaven-CMS-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# LocalHaven-CMS Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "LocalHaven-CMS configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"