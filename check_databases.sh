#!/bin/bash

# Script to check and index conexus databases across projects
set -e

echo "Checking conexus databases across projects..."

# Function to check database status
check_db_status() {
    local project_path="$1"
    local project_name="$2"

    echo ""
    echo "=== Checking $project_name ==="
    echo "Path: $project_path"

    # Check if data directory exists
    if [ ! -d "$project_path/data" ]; then
        echo "❌ No data directory found"
        return 1
    fi

    # Check if database exists
    if [ ! -f "$project_path/data/conexus.db" ]; then
        echo "❌ No conexus.db file found"
        return 1
    fi

    # Get database size
    local db_size=$(stat -f%z "$project_path/data/conexus.db" 2>/dev/null || stat -c%s "$project_path/data/conexus.db" 2>/dev/null || echo "unknown")
    echo "Database size: ${db_size} bytes"

    # Check document count using sqlite3
    local doc_count=$(sqlite3 "$project_path/data/conexus.db" "SELECT COUNT(*) FROM documents;" 2>/dev/null || echo "error")
    if [ "$doc_count" = "error" ]; then
        echo "❌ Cannot access database"
        return 1
    fi

    echo "Documents indexed: $doc_count"

    if [ "$doc_count" -eq 0 ]; then
        echo "⚠️  Database exists but no documents indexed"
        return 2
    else
        echo "✅ Database properly indexed"
        return 0
    fi
}

# Function to trigger indexing for a project
trigger_indexing() {
    local project_path="$1"
    local project_name="$2"

    echo ""
    echo "=== Triggering indexing for $project_name ==="

    # Change to project directory
    cd "$project_path"

    # Set environment variables
    export CONEXUS_DB_PATH="./data/conexus.db"
    export CONEXUS_ROOT_PATH="."
    export CONEXUS_LOG_LEVEL="info"

    # Trigger force reindex using MCP
    echo "Starting force reindex..."
    local result=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"force_reindex"}}}' | timeout 30s ../conexus/bin/conexus-darwin-arm64 2>/dev/null | jq -r '.result.message' 2>/dev/null || echo "timeout")

    if [ "$result" = "Force reindex started" ]; then
        echo "✅ Force reindex started successfully"

        # Wait a bit for indexing to complete
        echo "Waiting for indexing to complete..."
        sleep 10

        # Check status
        local status_result=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | timeout 10s ../conexus/bin/conexus-darwin-arm64 2>/dev/null | jq -r '.result.message' 2>/dev/null || echo "error")

        if [[ "$status_result" == *"documents"* ]]; then
            echo "✅ Indexing completed: $status_result"
        else
            echo "⚠️  Indexing may still be in progress or encountered issues"
        fi
    else
        echo "❌ Failed to start indexing: $result"
    fi

    # Return to conexus directory
    cd - > /dev/null
}

# Projects to check
projects=(
    "/Users/johnferguson/Github/conexus:conexus"
    "/Users/johnferguson/Github/codeflow:codeflow"
    "/Users/johnferguson/Github/opencode:opencode"
    "/Users/johnferguson/Github/opencode/packages/opencode:opencode/packages/opencode"
    "/Users/johnferguson/Github/spring-creek-baptist:spring-creek-baptist"
    "/Users/johnferguson/Github/tunnelforge:tunnelforge"
)

# Check all projects
echo "=== DATABASE STATUS CHECK ==="
for project_info in "${projects[@]}"; do
    IFS=':' read -r project_path project_name <<< "$project_info"
    check_db_status "$project_path" "$project_name"
done

echo ""
echo "=== INDEXING TRIGGER ==="
echo "Note: This script will attempt to trigger indexing for projects that need it."
echo "Indexing may take some time depending on project size."

# Trigger indexing for projects that need it
for project_info in "${projects[@]}"; do
    IFS=':' read -r project_path project_name <<< "$project_info"

    # Check if project needs indexing
    if [ -d "$project_path/data" ] && [ -f "$project_path/data/conexus.db" ]; then
        doc_count=$(sqlite3 "$project_path/data/conexus.db" "SELECT COUNT(*) FROM documents;" 2>/dev/null || echo "0")
        if [ "$doc_count" -eq 0 ]; then
            trigger_indexing "$project_path" "$project_name"
        fi
    fi
done

echo ""
echo "=== SUMMARY ==="
echo "Database check and indexing trigger completed."
echo "Check the output above for detailed status of each project."