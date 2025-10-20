#!/bin/bash

# advent_of_code Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/advent_of_code"
cd "$PROJECT_DIR"

echo "Setting up Conexus for advent_of_code..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/advent_of_code-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# advent_of_code Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "advent_of_code configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"