#!/bin/bash

# f3rg-redis Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/f3rg-redis"
cd "$PROJECT_DIR"

echo "Setting up Conexus for f3rg-redis..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/f3rg-redis-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# f3rg-redis Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "f3rg-redis configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"