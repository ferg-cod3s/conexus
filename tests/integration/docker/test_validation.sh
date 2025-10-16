#!/bin/bash
# Test Validation Script - Validates test structure without requiring Docker
# This script checks that all test files are properly structured

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[PASS]${NC} $1"; }
log_error() { echo -e "${RED}[FAIL]${NC} $1"; }

echo "=========================================="
echo "  Test Structure Validation"
echo "=========================================="
echo ""

VALIDATION_PASSED=0
VALIDATION_FAILED=0

# Check test script exists
log_info "Checking test script..."
if [ -f "tests/integration/docker/docker_test.sh" ]; then
  log_success "docker_test.sh exists"
  ((VALIDATION_PASSED++))
  
  # Check executable
  if [ -x "tests/integration/docker/docker_test.sh" ]; then
    log_success "docker_test.sh is executable"
    ((VALIDATION_PASSED++))
  else
    log_error "docker_test.sh is not executable"
    ((VALIDATION_FAILED++))
  fi
  
  # Check shebang
  if head -1 tests/integration/docker/docker_test.sh | grep -q "^#!/bin/bash"; then
    log_success "docker_test.sh has correct shebang"
    ((VALIDATION_PASSED++))
  else
    log_error "docker_test.sh missing or incorrect shebang"
    ((VALIDATION_FAILED++))
  fi
  
  # Count test cases
  TEST_COUNT=$(grep -c "^test_start" tests/integration/docker/docker_test.sh || echo "0")
  log_info "Found $TEST_COUNT test cases defined"
  if [ "$TEST_COUNT" -ge 8 ]; then
    log_success "Adequate test coverage (>= 8 tests)"
    ((VALIDATION_PASSED++))
  else
    log_error "Insufficient test coverage (< 8 tests)"
    ((VALIDATION_FAILED++))
  fi
else
  log_error "docker_test.sh not found"
  ((VALIDATION_FAILED++))
fi
echo ""

# Check README exists
log_info "Checking documentation..."
if [ -f "tests/integration/docker/README.md" ]; then
  log_success "README.md exists"
  ((VALIDATION_PASSED++))
  
  # Check for required sections
  for section in "Overview" "Prerequisites" "Running Tests" "Troubleshooting"; do
    if grep -q "## $section" tests/integration/docker/README.md; then
      log_success "README contains '$section' section"
      ((VALIDATION_PASSED++))
    else
      log_error "README missing '$section' section"
      ((VALIDATION_FAILED++))
    fi
  done
else
  log_error "README.md not found"
  ((VALIDATION_FAILED++))
fi
echo ""

# Check Docker Compose files
log_info "Checking Docker Compose configuration..."
if [ -f "docker-compose.yml" ]; then
  log_success "docker-compose.yml exists"
  ((VALIDATION_PASSED++))
else
  log_error "docker-compose.yml not found"
  ((VALIDATION_FAILED++))
fi

if [ -f "docker-compose.dev.yml" ]; then
  log_success "docker-compose.dev.yml exists"
  ((VALIDATION_PASSED++))
else
  log_error "docker-compose.dev.yml not found"
  ((VALIDATION_FAILED++))
fi

if [ -f "Dockerfile" ]; then
  log_success "Dockerfile exists"
  ((VALIDATION_PASSED++))
else
  log_error "Dockerfile not found"
  ((VALIDATION_FAILED++))
fi
echo ""

# Summary
echo "=========================================="
echo "  Validation Summary"
echo "=========================================="
echo "Checks Passed: $VALIDATION_PASSED"
echo "Checks Failed: $VALIDATION_FAILED"
echo "=========================================="
echo ""

if [ $VALIDATION_FAILED -eq 0 ]; then
  log_success "All validations passed! ✅"
  log_info "Test suite is ready for execution"
  exit 0
else
  log_error "Some validations failed ❌"
  exit 1
fi
