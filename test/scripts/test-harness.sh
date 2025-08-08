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
                
                # Run tests and capture exit code
                set +e
                gotestsum \
                    --format=testname \
                    --jsonfile=/tmp/test-results.json \
                    --junitfile=/tmp/junit.xml \
                    -- -v -short -count=1 \
                    -coverprofile=/tmp/coverage.out \
                    ./test/unit/...
                TEST_EXIT_CODE=$?
                set -e
                
                # Show coverage summary if coverage file exists
                if [ -f /tmp/coverage.out ]; then
                    echo ""
                    echo -e "${CYAN}=== Coverage Summary ===${NC}"
                    go tool cover -func=/tmp/coverage.out | tail -5 || echo "No coverage data"
                fi
                
                # Parse and display results from JUnit XML (more reliable)
                if [ -f /tmp/junit.xml ]; then
                    echo ""
                    echo -e "${CYAN}=== Test Summary ===${NC}"
                    
                    # Parse JUnit XML for counts
                    TOTAL=$(grep -o 'tests="[0-9]*"' /tmp/junit.xml | head -1 | grep -o '[0-9]*' || echo "0")
                    FAILURES=$(grep -o 'failures="[0-9]*"' /tmp/junit.xml | head -1 | grep -o '[0-9]*' || echo "0")
                    ERRORS=$(grep -o 'errors="[0-9]*"' /tmp/junit.xml | head -1 | grep -o '[0-9]*' || echo "0")
                    SKIPPED=$(grep -o 'skipped="[0-9]*"' /tmp/junit.xml | head -1 | grep -o '[0-9]*' || echo "0")
                    
                    # Calculate passed
                    FAILED=$((FAILURES + ERRORS))
                    PASSED=$((TOTAL - FAILED - SKIPPED))
                    
                    echo -e "Total Tests: ${BLUE}${TOTAL}${NC}"
                    echo -e "Passed: ${GREEN}${PASSED}${NC}"
                    echo -e "Failed: ${RED}${FAILED}${NC}"
                    echo -e "Skipped: ${YELLOW}${SKIPPED}${NC}"
                    
                    # Show failed tests if any
                    if [ "$FAILED" -gt 0 ]; then
                        echo ""
                        echo -e "${RED}=== Failed Tests ===${NC}"
                        # Extract failed test names from JUnit XML
                        grep '<failure\|<error' /tmp/junit.xml -B 2 | grep 'name=' | \
                            sed 's/.*name="\([^"]*\)".*/\1/' | \
                            while read test; do
                                echo -e "  ${RED}✗ $test${NC}"
                            done
                    fi
                    
                    # Exit with test exit code
                    if [ $TEST_EXIT_CODE -ne 0 ]; then
                        echo ""
                        echo -e "${RED}Tests failed with exit code: $TEST_EXIT_CODE${NC}"
                        exit $TEST_EXIT_CODE
                    fi
                elif [ -f /tmp/test-results.json ]; then
                    # Fallback to JSON parsing if JUnit not available
                    echo ""
                    echo -e "${CYAN}=== Test Summary (from JSON) ===${NC}"
                    
                    # Parse JSON for counts (one-liner to avoid multiline issues)
                    TOTAL=$(grep -c '"Test":' /tmp/test-results.json 2>/dev/null | head -1 || echo "0")
                    PASSED=$(grep '"Pass":true' /tmp/test-results.json 2>/dev/null | grep -c '"Test":' || echo "0")
                    FAILED=$(grep '"Pass":false' /tmp/test-results.json 2>/dev/null | grep -c '"Test":' || echo "0")
                    SKIPPED=$(grep '"Skip":true' /tmp/test-results.json 2>/dev/null | grep -c '"Test":' || echo "0")
                    
                    echo -e "Total Tests: ${BLUE}${TOTAL}${NC}"
                    echo -e "Passed: ${GREEN}${PASSED}${NC}"
                    echo -e "Failed: ${RED}${FAILED}${NC}"
                    echo -e "Skipped: ${YELLOW}${SKIPPED}${NC}"
                    
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
                
            else
                # Fallback to standard go test with better formatting
                echo -e "${CYAN}Using standard go test...${NC}"
                
                # Run tests with exit code capture
                set +e
                go test -v -short -count=1 \
                    -coverprofile=/tmp/coverage.out \
                    ./test/unit/... 2>&1 | tee /tmp/test-output.log | while IFS= read -r line; do
                    
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
                TEST_EXIT_CODE=$?
                set -e
                
                # Parse output for summary
                echo ""
                echo -e "${CYAN}=== Test Summary ===${NC}"
                
                PASSED=$(grep -c "^--- PASS:" /tmp/test-output.log 2>/dev/null || echo "0")
                FAILED=$(grep -c "^--- FAIL:" /tmp/test-output.log 2>/dev/null || echo "0")
                TOTAL=$((PASSED + FAILED))
                
                echo -e "Total Tests: ${BLUE}${TOTAL}${NC}"
                echo -e "Passed: ${GREEN}${PASSED}${NC}"
                echo -e "Failed: ${RED}${FAILED}${NC}"
                
                if [ "$FAILED" -gt 0 ]; then
                    echo ""
                    echo -e "${RED}=== Failed Tests ===${NC}"
                    grep "^--- FAIL:" /tmp/test-output.log | while read line; do
                        echo -e "${RED}$line${NC}"
                    done
                fi
                
                # Show coverage
                if [ -f /tmp/coverage.out ]; then
                    echo ""
                    echo -e "${CYAN}=== Coverage Summary ===${NC}"
                    go tool cover -func=/tmp/coverage.out 2>/dev/null | tail -5 || echo "No coverage data"
                fi
                
                if [ $TEST_EXIT_CODE -ne 0 ]; then
                    exit $TEST_EXIT_CODE
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