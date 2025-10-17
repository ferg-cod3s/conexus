#!/bin/bash
# Check stress test at milestone

MILESTONE=$1
EXPECTED_VUS=$2
LOG_FILE="tests/load/results/stress-test.log"
JSON_FILE="tests/load/results/stress-test.json"

echo "=== Milestone Check: $MILESTONE ==="
echo "Time: $(date '+%H:%M:%S')"
echo "Expected VUs: ~$EXPECTED_VUS"
echo ""

# Check if process is still running
if ! ps -p 132195 > /dev/null 2>&1; then
    echo "⚠️  Process completed/stopped early"
    exit 1
fi

# Get latest status
LATEST=$(grep "running (" "$LOG_FILE" | tail -1)
echo "📈 Current Status:"
echo "  $LATEST"
echo ""

# Extract VUs and time
CURRENT_VUS=$(echo "$LATEST" | grep -oP '\d+/500 VUs' | cut -d'/' -f1)
ELAPSED=$(echo "$LATEST" | grep -oP '\d+m\d+\.\d+s' | head -1)

echo "🎯 Milestone Progress:"
echo "  Current VUs: $CURRENT_VUS"
echo "  Elapsed: $ELAPSED"
echo ""

# File size
SIZE=$(ls -lh "$JSON_FILE" | awk '{print $5}')
echo "📁 Results: $SIZE"

# Quick health check
echo ""
echo "🏥 System Health:"
docker ps --filter name=conexus --format "  Status: {{.Status}}" 2>/dev/null
curl -s localhost:8080/health | jq -r '  "  Endpoint: \(.status)"' 2>/dev/null || echo "  Endpoint: checking..."

echo ""
echo "✅ Milestone check complete"
