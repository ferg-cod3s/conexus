#!/bin/bash
# Monitor stress test progress

STRESS_PID=132195
LOG_FILE="tests/load/results/stress-test.log"
JSON_FILE="tests/load/results/stress-test.json"

echo "=== Stress Test Monitor ==="
echo "Time: $(date '+%H:%M:%S')"
echo ""

# Check if process is running
if ps -p $STRESS_PID > /dev/null 2>&1; then
    ELAPSED=$(ps -p $STRESS_PID -o etime= | xargs)
    echo "✅ Status: RUNNING (PID $STRESS_PID)"
    echo "⏱️  Elapsed: $ELAPSED"
else
    echo "❌ Status: COMPLETED or FAILED"
    if [ -f "$JSON_FILE" ]; then
        echo "📊 Results file exists"
    fi
    exit 0
fi

echo ""

# Get latest status from log
if [ -f "$LOG_FILE" ]; then
    echo "📈 Latest Progress:"
    grep "running (" "$LOG_FILE" | tail -1
    echo ""
fi

# File sizes
if [ -f "$JSON_FILE" ]; then
    SIZE=$(ls -lh "$JSON_FILE" | awk '{print $5}')
    echo "📁 Results file: $SIZE"
fi

echo ""

# Container health
echo "🏥 Container Status:"
docker ps --filter name=conexus --format "  {{.Status}}" 2>/dev/null || echo "  Not running"

# Quick metric check from JSON (last few VU values)
if [ -f "$JSON_FILE" ] && command -v jq > /dev/null 2>&1; then
    echo ""
    echo "🔢 Current VUs:"
    tail -100 "$JSON_FILE" | jq -s 'map(select(.metric == "vus")) | .[-1] | .data.value' 2>/dev/null || echo "  Parsing..."
fi
