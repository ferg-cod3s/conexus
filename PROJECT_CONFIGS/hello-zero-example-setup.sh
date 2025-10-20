#!/bin/bash

# hello-zero-example Conexus Setup Script
set -e

PROJECT_DIR="/Users/johnferguson/Github/hello-zero-example"
cd "$PROJECT_DIR"

echo "Setting up Conexus for hello-zero-example..."

# Create .opencode directory
mkdir -p .opencode

# Copy configuration
cp ./PROJECT_CONFIGS/hello-zero-example-opencode.jsonc .opencode/opencode.jsonc

# Create data directory
mkdir -p data

# Create environment file
cat > .env << EOF
# hello-zero-example Environment
CONEXUS_DB_PATH=./data/conexus.db
CONEXUS_PORT=0
CONEXUS_LOG_LEVEL=info
EOF

echo "hello-zero-example configured successfully!"
echo "Database: ./data/conexus.db"
echo "Run 'source .env' to load environment variables"