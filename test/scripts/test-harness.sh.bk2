#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

    cd /workspace

    echo "Checking Go module setup..."
    go list -m || (echo "ERROR: Go module not found!" && exit 1)

    case $suite in
        unit)
            echo -e "${CYAN}Running unit tests from /workspace...${NC}"
            echo ""

            # Use go test with more verbose output and json for parsing
            if command -v gotestsum >/dev/null 2>&1; then
                echo -e "${CYAN}Using gotestsum for better output...${NC}"
                gotestsum \
                    --format=testname \
                    --jsonfile=/tmp/test-results.json \
                    -- -v -short -count=1 \
                    -coverprofile=/tmp/coverage.out \
                    ./test/unit/...

                # Show coverage summary
                echo ""
                echo -e "${CYAN}=== Coverage Summary ===${NC}"
                go tool cover -func=/tmp/coverage.out | tail -5

                # Show test summary
                echo ""
                echo -e "${CYAN}=== Test Summary ===${NC}"
                gotestsum tool slowest --jsonfile /tmp/test-results.json --num 10

            else
                # Fallback to standard go test with better formatting
                go test -v -short -count=1 \
                    -coverprofile=/tmp/coverage.out \
                    ./test/unit/... 2>&1 | while IFS= read -r line; do

                    # Color code the output
                    if [[ "$line" == "=== RUN"* ]]; then
                        echo -e "${BLUE}$line${NC}"
                    elif [[ "$line" == "--- PASS:"* ]]; then
                        echo -e "${GREEN}$line${NC}"
                    elif [[ "$line" == "--- FAIL:"* ]]; then
                        echo -e "${RED}$line${NC}"
                    elif [[ "$line" == "PASS" ]]; then
                        echo -e "${GREEN}✓ $line${NC}"
                    elif [[ "$line" == "FAIL" ]]; then
                        echo -e "${RED}✗ $line${NC}"
                    elif [[ "$line" == "ok"* ]]; then
                        echo -e "${GREEN}$line${NC}"
                    elif [[ "$line" == "?"* ]]; then
                        echo -e "${YELLOW}$line${NC}"
                    else
                        echo "$line"
                    fi
                done

                # Show coverage
                echo ""
                echo -e "${CYAN}=== Coverage Summary ===${NC}"
                go tool cover -func=/tmp/coverage.out 2>/dev/null | tail -5 || echo "No coverage data"
            fi

            # Count results
            echo ""
            echo -e "${CYAN}=== Final Results ===${NC}"

            # Parse test results if json file exists
            if [ -f /tmp/test-results.json ]; then
                TOTAL=$(grep -c '"Test":' /tmp/test-results.json || echo "0")
                PASSED=$(grep -c '"Pass":true' /tmp/test-results.json || echo "0")
                FAILED=$(grep -c '"Pass":false' /tmp/test-results.json || echo "0")

                echo -e "Total Tests: ${BLUE}$TOTAL${NC}"
                echo -e "Passed: ${GREEN}$PASSED${NC}"
                echo -e "Failed: ${RED}$FAILED${NC}"

                if [ "$FAILED" -gt 0 ]; then
                    echo ""
                    echo -e "${RED}=== Failed Tests ===${NC}"
                    grep '"Pass":false' /tmp/test-results.json | \
                        sed -n 's/.*"Test":"\([^"]*\)".*/\1/p' | \
                        while read test; do
                            echo -e "  ${RED}✗ $test${NC}"
                        done
                fi
            fi
            ;;

        integration)
            go test -v -count=1 ./test/integration/... 2>&1 | while IFS= read -r line; do
                if [[ "$line" == "--- PASS:"* ]]; then
                    echo -e "${GREEN}$line${NC}"
                elif [[ "$line" == "--- FAIL:"* ]]; then
                    echo -e "${RED}$line${NC}"
                else
                    echo "$line"
                fi
            done
            ;;

        e2e)
            go test -v -timeout=30m -count=1 ./test/e2e/... 2>&1 | while IFS= read -r line; do
                if [[ "$line" == "--- PASS:"* ]]; then
                    echo -e "${GREEN}$line${NC}"
                elif [[ "$line" == "--- FAIL:"* ]]; then
                    echo -e "${RED}$line${NC}"
                else
                    echo "$line"
                fi
            done
            ;;

        all)
            go test -v -count=1 ./test/...
            ;;

        *)
            echo -e "${RED}Unknown test suite: ${suite}${NC}"
            exit 1
            ;;
    esac

    echo ""
    echo -e "${GREEN}✓ ${suite} tests completed${NC}"
}

# Main execution
case ${1:-test} in
    test)
        run_test_suite "${TEST_SUITE:-all}"
        ;;
    shell)
        echo -e "${GREEN}Entering test harness shell...${NC}"
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