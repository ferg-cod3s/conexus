#!/bin/bash

# Test Conexus in the OpenCode project
set -e

echo "🧪 Testing Conexus in OpenCode Project"
echo "===================================="

cd /Users/johnferguson/Github/opencode

# Test 1: Check configuration exists
echo "1. Checking project configuration..."
if [ -f ".opencode/opencode.jsonc" ]; then
    echo "✅ Project configuration exists"
else
    echo "❌ Project configuration missing"
    exit 1
fi

# Test 2: Check environment file
echo "2. Checking environment file..."
if [ -f ".env" ]; then
    echo "✅ Environment file exists"
    echo "   Contents:"
    cat .env | sed 's/^/   /'
else
    echo "❌ Environment file missing"
    exit 1
fi

# Test 3: Check data directory
echo "3. Checking data directory..."
if [ -d "data" ]; then
    echo "✅ Data directory exists"
    if [ -f "data/conexus.db" ]; then
        echo "✅ Conexus database exists"
        db_size=$(ls -lh data/conexus.db | awk '{print $5}')
        echo "   Database size: $db_size"
    else
        echo "ℹ️  Database will be created on first use"
    fi
else
    echo "❌ Data directory missing"
    exit 1
fi

# Test 4: Test binary access
echo "4. Testing Conexus binary access..."
if [ -f "/Users/johnferguson/Github/conexus/conexus" ]; then
    echo "✅ Conexus binary accessible"
else
    echo "❌ Conexus binary not found"
    exit 1
fi

echo ""
echo "🎉 OpenCode Project Conexus Setup Complete!"
echo ""
echo "To test in OpenCode:"
echo "1. cd ~/Github/opencode"
echo "2. source .env"
echo "3. opencode"
echo "4. @conexus-expert analyze this OpenCode codebase"
echo ""
echo "Available agents for this project:"
echo "- @conexus-expert (semantic search)"
echo "- @typescript-pro (TypeScript expertise)"
echo "- @frontend-developer (frontend development)"
echo "- @systems-programming-expert (systems programming)"