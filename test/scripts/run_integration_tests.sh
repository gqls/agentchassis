#!/bin/bash
# test/scripts/run_integration_tests.sh

set -e

echo "Starting integration tests..."
echo "================================"

# Set test timeout
export TEST_TIMEOUT=${TEST_TIMEOUT:-30m}

# Run tests with verbose output and show which tests are running
echo "Running tests with timeout: $TEST_TIMEOUT"

# Run each package separately to identify failures
FAILED_PACKAGES=""
PASSED_PACKAGES=""

for package in agents database kafka; do
    echo ""
    echo "Testing package: $package"
    echo "------------------------"

    if go test -v -timeout $TEST_TIMEOUT ./test/integration/$package/... 2>&1; then
        echo "✓ Package $package PASSED"
        PASSED_PACKAGES="$PASSED_PACKAGES $package"
    else
        echo "✗ Package $package FAILED"
        FAILED_PACKAGES="$FAILED_PACKAGES $package"
    fi
done

echo ""
echo "================================"
echo "Test Summary:"
echo "Passed packages:$PASSED_PACKAGES"
echo "Failed packages:$FAILED_PACKAGES"

if [ -n "$FAILED_PACKAGES" ]; then
    echo "✗ Some tests failed"
    exit 1
else
    echo "✓ All tests passed"
    exit 0
fi