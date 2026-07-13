#!/bin/bash
# test/scripts/run_e2e_tests.sh

echo "=== E2E TEST EXECUTION STARTING ==="
echo "Time: $(date)"
echo "Current directory: $(pwd)"
echo "Go version: $(go version)"

# List test files
echo ""
echo "E2E test files found:"
find test/e2e -name "*.go" -type f 2>/dev/null | head -20 || echo "No files found"

echo ""
echo "=== RUNNING E2E TESTS WITH FULL OUTPUT ==="

# Run tests with maximum verbosity and real-time output
# Using script command to force line buffering
script -q -c "go test -v -count=1 -timeout 30s ./test/e2e/... 2>&1" /dev/null | cat

TEST_EXIT_CODE=${PIPESTATUS[0]}

echo ""
echo "=== TEST EXECUTION COMPLETED ==="
echo "Exit code: $TEST_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "✓ e2e tests completed successfully"
else
    echo "✗ e2e tests failed with code: $TEST_EXIT_CODE"
fi

exit $TEST_EXIT_CODE