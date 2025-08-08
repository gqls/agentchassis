#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}=== Agent Chassis Test Harness ===${NC}"
echo -e "${BLUE}Running in Kubernetes${NC}"
echo "Namespace: ${NAMESPACE:-default}"
echo "Test Suite: ${TEST_SUITE:-all}"
echo ""

# Function to run test suite
run_test_suite() {
    local suite=$1
    echo -e "${YELLOW}Running ${suite} tests...${NC}"

    # Make sure we're in the right directory
    cd /workspace

    # Verify module is set up correctly
    echo "Checking Go module setup..."
    go list -m || (echo "ERROR: Go module not found!" && exit 1)

    case $suite in
        unit)
            echo "Running unit tests from /workspace..."
            if command -v gotestsum >/dev/null 2>&1; then
                gotestsum --format=testname -- -v -short ./test/unit/...
            else
                go test -v -short ./test/unit/...
            fi
            ;;
        integration)
            go test -v ./test/integration/...
            ;;
        e2e)
            go test -v -timeout=30m ./test/e2e/...
            ;;
        all)
            go test -v ./test/...
            ;;
        *)
            echo -e "${RED}Unknown test suite: ${suite}${NC}"
            exit 1
            ;;
    esac

    echo -e "${GREEN}✓ ${suite} tests completed${NC}"
}

# Main execution
case ${1:-test} in
    test)
        run_test_suite "${TEST_SUITE:-all}"
        ;;
    shell)
        echo -e "${GREEN}Entering test harness shell...${NC)"
        cd /workspace
        exec /bin/bash
        ;;
    validate)
        echo -e "${YELLOW}Validating test environment...${NC}"
        cd /workspace
        echo "Workspace contents:"
        ls -la
        echo ""
        echo "Go module:"
        go list -m
        echo ""
        echo "Available packages:"
        go list ./... | head -20
        echo ""
        echo "Test directory:"
        ls -la test/
        ;;
    *)
        exec "$@"
        ;;
esac