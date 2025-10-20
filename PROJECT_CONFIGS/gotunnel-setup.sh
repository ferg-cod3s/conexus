#!/bin/bash

# gotunnel Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/gotunnel"
cd "$PROJECT_DIR"

echo "Setting up Conexus for gotunnel..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/gotunnel-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# gotunnel Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "gotunnel configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"