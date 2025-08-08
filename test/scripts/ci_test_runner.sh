## 15. Final Test Script for CI/CD

bash
#!/bin/bash
# test/scripts/ci_test_runner.sh

set -e

echo "=== Agent Chassis CI Test Runner ==="
echo "Environment: ${ENVIRONMENT:-production}"
echo "Test Suite: ${TEST_SUITE:-all}"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test results
FAILED=0

# Run test suite
run_tests() {
    local suite=$1
    echo -e "${YELLOW}Running ${suite} tests...${NC}"

    if make test-${suite}; then
        echo -e "${GREEN}✓ ${suite} tests passed${NC}"
    else
        echo -e "${RED}✗ ${suite} tests failed${NC}"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

# Setup
echo -e "${YELLOW}Setting up test environment...${NC}"
make test-setup || exit 1

# Run tests based on suite
case ${TEST_SUITE} in
    unit)
        run_tests unit
        ;;
    integration)
        run_tests integration
        ;;
    e2e)
        run_tests e2e
        ;;
    performance)
        run_tests performance
        ;;
    all)
        run_tests unit
        run_tests integration
        run_tests e2e
        ;;
    *)
        echo -e "${RED}Unknown test suite: ${TEST_SUITE}${NC}"
        exit 1
        ;;
esac

# Generate test report
echo -e "${YELLOW}Generating test report...${NC}"
go test -json ./... > test-report.json || true

# Coverage report
echo -e "${YELLOW}Generating coverage report...${NC}"
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Cleanup
echo -e "${YELLOW}Cleaning up...${NC}"
make test-clean

# Summary
echo ""
echo "=== Test Summary ==="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}${FAILED} test suite(s) failed${NC}"
    exit 1
fi