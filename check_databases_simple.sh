#!/bin/bash

# Script to check conexus databases across projects
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
echo "=== SUMMARY ==="
echo "Database check completed."
echo "Projects with 0 documents need indexing."