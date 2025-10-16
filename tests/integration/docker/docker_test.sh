#!/bin/bash
# Conexus Docker Integration Tests
# Tests Docker Compose deployment, MCP protocol, and data persistence
#
# Usage: ./docker_test.sh [--cleanup]
#   --cleanup: Clean up containers and volumes after tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
MCP_ENDPOINT="http://localhost:8080/mcp"
HEALTH_ENDPOINT="http://localhost:8080/health"
MAX_WAIT=60  # Maximum seconds to wait for service startup
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$TEST_DIR/../../.." && pwd)"
CLEANUP=false

# Parse arguments
for arg in "$@"; do
  case $arg in
    --cleanup)
      CLEANUP=true
      shift
      ;;
  esac
done

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[PASS]${NC} $1"
  ((TESTS_PASSED++))
}

log_error() {
  echo -e "${RED}[FAIL]${NC} $1"
  ((TESTS_FAILED++))
}

log_warning() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

test_start() {
  ((TESTS_RUN++))
  log_info "Test $TESTS_RUN: $1"
}

wait_for_service() {
  local endpoint=$1
  local max_wait=$2
  local elapsed=0
  
  log_info "Waiting for service at $endpoint (max ${max_wait}s)..."
  
  while [ $elapsed -lt $max_wait ]; do
    if curl -sf "$endpoint" > /dev/null 2>&1; then
      log_success "Service ready after ${elapsed}s"
      return 0
    fi
    sleep 1
    ((elapsed++))
  done
  
  log_error "Service did not become ready within ${max_wait}s"
  return 1
}

cleanup() {
  log_info "Cleaning up test environment..."
  cd "$PROJECT_ROOT"
  docker compose down -v 2>/dev/null || true
  docker compose -f docker-compose.yml -f docker-compose.dev.yml down -v 2>/dev/null || true
  rm -rf ./data/test-* 2>/dev/null || true
  log_success "Cleanup complete"
}

# Trap to ensure cleanup on exit if requested
if [ "$CLEANUP" = true ]; then
  trap cleanup EXIT
fi

# Main test execution
main() {
  echo ""
  echo "=========================================="
  echo "  Conexus Docker Integration Tests"
  echo "=========================================="
  echo ""
  
  cd "$PROJECT_ROOT"
  
  # Pre-flight checks
  log_info "Running pre-flight checks..."
  
  if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
  fi
  
  if ! command -v docker compose &> /dev/null; then
    log_error "Docker Compose is not installed or not in PATH"
    exit 1
  fi
  
  if ! command -v curl &> /dev/null; then
    log_error "curl is not installed or not in PATH"
    exit 1
  fi
  
  if ! command -v jq &> /dev/null; then
    log_warning "jq is not installed - JSON validation tests will be skipped"
  fi
  
  log_success "Pre-flight checks passed"
  echo ""
  
  # Ensure clean state
  log_info "Ensuring clean state..."
  cleanup
  echo ""
  
  # ==========================================
  # Test 1: Docker Image Build
  # ==========================================
  test_start "Docker image builds successfully"
  
  if docker compose build; then
    log_success "Docker image built successfully"
    
    # Check image size
    IMAGE_SIZE=$(docker images conexus:latest --format "{{.Size}}")
    log_info "Image size: $IMAGE_SIZE"
    
    # Image should be reasonably small (< 50MB for Alpine-based)
    SIZE_BYTES=$(docker images conexus:latest --format "{{.Size}}" | sed 's/MB//' | awk '{print int($1)}')
    if [ "$SIZE_BYTES" -lt 50 ] 2>/dev/null; then
      log_success "Image size is optimal (< 50MB)"
    else
      log_warning "Image size is larger than expected (target: < 50MB)"
    fi
  else
    log_error "Docker image build failed"
    exit 1
  fi
  echo ""
  
  # ==========================================
  # Test 2: Container Startup (Production)
  # ==========================================
  test_start "Container starts successfully (production mode)"
  
  if docker compose up -d; then
    log_success "Container started"
    
    # Wait for health check
    if wait_for_service "$HEALTH_ENDPOINT" $MAX_WAIT; then
      log_success "Health check passed"
    else
      log_error "Health check failed"
      docker compose logs
      exit 1
    fi
  else
    log_error "Container failed to start"
    docker compose logs
    exit 1
  fi
  echo ""
  
  # ==========================================
  # Test 3: MCP Protocol - tools/list
  # ==========================================
  test_start "MCP protocol: tools/list endpoint"
  
  RESPONSE=$(curl -sf -X POST "$MCP_ENDPOINT" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}')
  
  if [ $? -eq 0 ]; then
    log_success "tools/list endpoint responded"
    
    # Validate JSON-RPC format
    if echo "$RESPONSE" | jq -e '.jsonrpc == "2.0"' > /dev/null 2>&1; then
      log_success "Response is valid JSON-RPC 2.0"
    else
      log_error "Response is not valid JSON-RPC 2.0"
    fi
    
    # Check for expected tools
    if echo "$RESPONSE" | jq -e '.result.tools | length >= 4' > /dev/null 2>&1; then
      log_success "Response contains at least 4 tools"
      
      # List available tools
      if command -v jq &> /dev/null; then
        log_info "Available tools:"
        echo "$RESPONSE" | jq -r '.result.tools[].name' | sed 's/^/  - /'
      fi
    else
      log_error "Response does not contain expected tools"
    fi
  else
    log_error "tools/list endpoint failed"
  fi
  echo ""
  
  # ==========================================
  # Test 4: MCP Protocol - Each Tool
  # ==========================================
  test_start "MCP protocol: Individual tool invocations"
  
  # Test codebase/locate
  RESPONSE=$(curl -sf -X POST "$MCP_ENDPOINT" \
    -H "Content-Type: application/json" \
    -d '{
      "jsonrpc":"2.0",
      "id":2,
      "method":"tools/call",
      "params":{
        "name":"codebase/locate",
        "arguments":{"query":"main.go"}
      }
    }')
  
  if [ $? -eq 0 ] && echo "$RESPONSE" | jq -e '.result' > /dev/null 2>&1; then
    log_success "codebase/locate tool works"
  else
    log_error "codebase/locate tool failed"
  fi
  
  # Test codebase/analyze
  RESPONSE=$(curl -sf -X POST "$MCP_ENDPOINT" \
    -H "Content-Type: application/json" \
    -d '{
      "jsonrpc":"2.0",
      "id":3,
      "method":"tools/call",
      "params":{
        "name":"codebase/analyze",
        "arguments":{"path":"cmd/conexus/main.go"}
      }
    }')
  
  if [ $? -eq 0 ] && echo "$RESPONSE" | jq -e '.result' > /dev/null 2>&1; then
    log_success "codebase/analyze tool works"
  else
    log_error "codebase/analyze tool failed"
  fi
  
  echo ""
  
  # ==========================================
  # Test 5: Configuration Loading
  # ==========================================
  test_start "Environment variable configuration"
  
  # Check logs for config loading
  LOGS=$(docker compose logs conexus 2>&1)
  
  if echo "$LOGS" | grep -q "Server listening"; then
    log_success "Server started with configuration"
    
    # Check if it's using the correct port
    if echo "$LOGS" | grep -q "8080"; then
      log_success "Server listening on configured port 8080"
    else
      log_warning "Server port not found in logs"
    fi
  else
    log_error "Server startup message not found in logs"
  fi
  echo ""
  
  # ==========================================
  # Test 6: Data Persistence
  # ==========================================
  test_start "Data persistence across restarts"
  
  # Create some data by indexing
  log_info "Creating data via indexing..."
  curl -sf -X POST "$MCP_ENDPOINT" \
    -H "Content-Type: application/json" \
    -d '{
      "jsonrpc":"2.0",
      "id":4,
      "method":"tools/call",
      "params":{
        "name":"codebase/locate",
        "arguments":{"query":"function"}
      }
    }' > /dev/null
  
  # Check that database file exists
  if [ -f "./data/conexus.db" ]; then
    log_success "Database file created"
    DB_SIZE_BEFORE=$(stat -f%z "./data/conexus.db" 2>/dev/null || stat -c%s "./data/conexus.db" 2>/dev/null)
    log_info "Database size before restart: $DB_SIZE_BEFORE bytes"
  else
    log_error "Database file not found"
  fi
  
  # Restart container
  log_info "Restarting container..."
  docker compose restart
  
  if wait_for_service "$HEALTH_ENDPOINT" $MAX_WAIT; then
    log_success "Container restarted successfully"
    
    # Check database still exists and has data
    if [ -f "./data/conexus.db" ]; then
      DB_SIZE_AFTER=$(stat -f%z "./data/conexus.db" 2>/dev/null || stat -c%s "./data/conexus.db" 2>/dev/null)
      log_info "Database size after restart: $DB_SIZE_AFTER bytes"
      
      if [ "$DB_SIZE_AFTER" -ge "$DB_SIZE_BEFORE" ]; then
        log_success "Database persisted across restart"
      else
        log_error "Database data may have been lost"
      fi
    else
      log_error "Database file missing after restart"
    fi
  else
    log_error "Container failed to restart"
  fi
  echo ""
  
  # ==========================================
  # Test 7: Graceful Shutdown
  # ==========================================
  test_start "Graceful shutdown"
  
  log_info "Stopping container..."
  if docker compose stop --timeout 10; then
    log_success "Container stopped gracefully"
    
    # Check for error logs
    STOP_LOGS=$(docker compose logs conexus --tail=20 2>&1)
    if ! echo "$STOP_LOGS" | grep -qi "panic\|fatal error"; then
      log_success "No errors during shutdown"
    else
      log_warning "Errors detected in shutdown logs"
    fi
  else
    log_error "Container failed to stop gracefully"
  fi
  
  docker compose down
  echo ""
  
  # ==========================================
  # Test 8: Development Mode
  # ==========================================
  test_start "Development mode configuration"
  
  log_info "Starting in development mode..."
  if docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d; then
    log_success "Dev container started"
    
    if wait_for_service "$HEALTH_ENDPOINT" $MAX_WAIT; then
      log_success "Dev container health check passed"
      
      # Check for debug logging
      DEV_LOGS=$(docker compose logs conexus 2>&1)
      if echo "$DEV_LOGS" | grep -qi "debug\|DBG"; then
        log_success "Debug logging enabled in dev mode"
      else
        log_warning "Debug logging not detected in dev mode"
      fi
    else
      log_error "Dev container health check failed"
    fi
    
    docker compose -f docker-compose.yml -f docker-compose.dev.yml down
  else
    log_error "Dev container failed to start"
  fi
  echo ""
  
  # ==========================================
  # Test Results Summary
  # ==========================================
  echo ""
  echo "=========================================="
  echo "  Test Results Summary"
  echo "=========================================="
  echo "Tests Run:    $TESTS_RUN"
  echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
  echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
  echo "=========================================="
  echo ""
  
  if [ $TESTS_FAILED -eq 0 ]; then
    log_success "All tests passed! ✅"
    return 0
  else
    log_error "Some tests failed ❌"
    return 1
  fi
}

# Run main test suite
main
exit $?
