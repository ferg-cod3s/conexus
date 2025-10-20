#!/bin/bash

# Setup script for Conexus environment variables
# This script configures the CONEXUS_DB_PATH environment variable
# for the current project directory

set -e

# Get the current project directory
PROJECT_DIR="$(pwd)"
PROJECT_NAME="$(basename "$PROJECT_DIR")"

# Default database path
DEFAULT_DB_PATH="$PROJECT_DIR/data/conexus.db"

# Check if custom path provided
if [ "$1" ]; then
    DB_PATH="$1"
else
    DB_PATH="$DEFAULT_DB_PATH"
fi

# Create data directory if it doesn't exist
mkdir -p "$(dirname "$DB_PATH")"

# Export environment variable
export CONEXUS_DB_PATH="$DB_PATH"
export CONEXUS_PORT="0"
export CONEXUS_LOG_LEVEL="info"

echo "Conexus environment configured:"
echo "  Project: $PROJECT_NAME"
echo "  Database: $DB_PATH"
echo "  Port: $CONEXUS_PORT"
echo "  Log Level: $CONEXUS_LOG_LEVEL"
echo ""
echo "To make this permanent, add to your shell profile:"
echo "export CONEXUS_DB_PATH=\"$DB_PATH\""
echo "export CONEXUS_PORT=\"$CONEXUS_PORT\""
echo "export CONEXUS_LOG_LEVEL=\"$CONEXUS_LOG_LEVEL\""

# Create .env file for the project
cat > .env << EOF
# Conexus Environment Variables
CONEXUS_DB_PATH=$DB_PATH
CONEXUS_PORT=$CONEXUS_PORT
CONEXUS_LOG_LEVEL=$CONEXUS_LOG_LEVEL
EOF

echo ""
echo "Created .env file in project directory"
echo "Run 'source .env' to load environment variables"