#!/bin/bash
# test/scripts/test-harness.sh

set -e

echo "=== TEST HARNESS STARTING ==="
echo "Time: $(date)"
echo "Environment:"
env | grep -E "^GO|^TEST" | sort

echo ""
echo "Test suite: ${TEST_SUITE:-e2e}"
echo ""

# Ensure output is not buffered
exec 2>&1

case "${TEST_SUITE:-e2e}" in
    unit)
        echo "Running unit tests..."
        go test -v -count=1 -timeout 30m ./test/unit/...
        ;;
    integration)
        echo "Running integration tests..."
        go test -v -count=1 -timeout 30m ./test/integration/...
        ;;
    e2e)
        echo "Running E2E tests..."
        # Run with explicit output and no caching
        go test -v -count=1 -timeout 30m ./test/e2e/... 2>&1 | tee /tmp/test.log
        EXIT_CODE=${PIPESTATUS[0]}

        # If tests failed, show the log
        if [ $EXIT_CODE -ne 0 ]; then
            echo ""
            echo "=== TEST FAILURES DETECTED ==="
            grep -E "FAIL|ERROR|panic" /tmp/test.log || true
        fi

        exit $EXIT_CODE
        ;;
    *)
        echo "Unknown test suite: ${TEST_SUITE}"
        exit 1
        ;;
esac