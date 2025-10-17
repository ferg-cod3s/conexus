#!/bin/bash
# Analyze stress test results

JSON_FILE="tests/load/results/stress-test.json"
ANALYSIS_FILE="tests/load/results/stress-analysis.json"
REPORT_FILE="tests/load/results/STRESS_TEST_ANALYSIS.md"

echo "Analyzing stress test results..."

if [ ! -f "$JSON_FILE" ]; then
    echo "❌ Error: Results file not found: $JSON_FILE"
    exit 1
fi

echo "📊 Extracting metrics (this may take a minute for large files)..."

# Count total requests
TOTAL_REQS=$(grep '"metric":"http_reqs"' "$JSON_FILE" | wc -l)

# Count errors
TOTAL_ERRORS=$(grep '"metric":"http_req_failed"' "$JSON_FILE" | grep '"value":1' | wc -l)

# Get durations and calculate percentiles
echo "Extracting duration metrics..."
grep '"metric":"http_req_duration"' "$JSON_FILE" | \
    jq -r '.data.value' | \
    sort -n > /tmp/durations.txt

TOTAL_SAMPLES=$(wc -l < /tmp/durations.txt)
P50_LINE=$((TOTAL_SAMPLES / 2))
P95_LINE=$((TOTAL_SAMPLES * 95 / 100))
P99_LINE=$((TOTAL_SAMPLES * 99 / 100))

P50=$(sed -n "${P50_LINE}p" /tmp/durations.txt)
P95=$(sed -n "${P95_LINE}p" /tmp/durations.txt)
P99=$(sed -n "${P99_LINE}p" /tmp/durations.txt)
MAX=$(tail -1 /tmp/durations.txt)

# Get test duration from first and last timestamps
FIRST_TIME=$(head -100 "$JSON_FILE" | grep '"time"' | head -1 | grep -o '"time":[0-9]*' | cut -d: -f2)
LAST_TIME=$(tail -100 "$JSON_FILE" | grep '"time"' | tail -1 | grep -o '"time":[0-9]*' | cut -d: -f2)
DURATION_NS=$((LAST_TIME - FIRST_TIME))
DURATION_SEC=$((DURATION_NS / 1000000000))

# Extract performance by operation
echo "Analyzing by operation..."
for OP in "analyze" "query-knowledge" "search-files"; do
    grep "\"name\":\"$OP\"" "$JSON_FILE" | \
        jq -r '.data.value' | \
        sort -n > "/tmp/${OP}_durations.txt"
    
    OP_COUNT=$(wc -l < "/tmp/${OP}_durations.txt")
    if [ $OP_COUNT -gt 0 ]; then
        OP_P95_LINE=$((OP_COUNT * 95 / 100))
        OP_P99_LINE=$((OP_COUNT * 99 / 100))
        OP_P95=$(sed -n "${OP_P95_LINE}p" "/tmp/${OP}_durations.txt")
        OP_P99=$(sed -n "${OP_P99_LINE}p" "/tmp/${OP}_durations.txt")
        echo "  $OP: count=$OP_COUNT, p95=${OP_P95}ms, p99=${OP_P99}ms"
    fi
done

# Create analysis JSON
cat > "$ANALYSIS_FILE" << EOJSON
{
  "summary": {
    "total_requests": $TOTAL_REQS,
    "total_errors": $TOTAL_ERRORS,
    "error_rate": $(echo "scale=4; $TOTAL_ERRORS * 100 / $TOTAL_REQS" | bc),
    "test_duration_seconds": $DURATION_SEC,
    "total_samples": $TOTAL_SAMPLES
  },
  "overall_performance": {
    "p50_ms": $P50,
    "p95_ms": $P95,
    "p99_ms": $P99,
    "max_ms": $MAX
  }
}
EOJSON

echo "✅ Analysis complete: $ANALYSIS_FILE"
echo ""
echo "📋 Quick Summary:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Total Requests:  $TOTAL_REQS"
echo "Total Errors:    $TOTAL_ERRORS ($(echo "scale=2; $TOTAL_ERRORS * 100 / $TOTAL_REQS" | bc)%)"
echo "Test Duration:   ${DURATION_SEC}s"
echo ""
echo "Overall Performance:"
echo "  p50: ${P50}ms"
echo "  p95: ${P95}ms"
echo "  p99: ${P99}ms"
echo "  max: ${MAX}ms"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

rm -f /tmp/durations.txt /tmp/*_durations.txt
