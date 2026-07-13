#!/bin/bash
# test/scripts/run_integration_tests.sh

# Don't exit on error immediately - we want to see all failures
set +e

echo "========================================="
echo "Starting integration tests at $(date)"
echo "========================================="
echo ""
echo "Environment Information:"
echo "  Working Directory: $(pwd)"
echo "  Go Version: $(go version)"
echo "  TEST_TIMEOUT: ${TEST_TIMEOUT:-30m}"
echo "  GOPATH: ${GOPATH}"
echo "  Hostname: $(hostname)"
echo ""

# Set test timeout
export TEST_TIMEOUT=${TEST_TIMEOUT:-30m}

# Create temp directory for logs
TEMP_DIR="/tmp/test_logs_$$"
mkdir -p ${TEMP_DIR}
echo "Log directory: ${TEMP_DIR}"
echo ""

# Function to run tests with detailed output
run_test_package() {
    local package=$1
    local log_file="${TEMP_DIR}/test_${package}.log"

    echo "========================================="
    echo ">>> Testing package: integration/${package}"
    echo ">>> Start time: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "-----------------------------------------"

    # List test files in the package
    echo "Test files in package:"
    if [ -d "./test/integration/${package}" ]; then
        ls -la ./test/integration/${package}/*.go 2>/dev/null | head -10 || echo "  No .go files found"
    else
        echo "  WARNING: Directory ./test/integration/${package} does not exist!"
    fi
    echo ""

    # Run tests with maximum verbosity
    echo "Running: go test -v -timeout ${TEST_TIMEOUT} ./test/integration/${package}/..."
    echo "-----------------------------------------"

    # Run test and capture output
    go test -v -timeout ${TEST_TIMEOUT} -count=1 ./test/integration/${package}/... 2>&1 | tee ${log_file}
    local exit_code=${PIPESTATUS[0]}

    echo "-----------------------------------------"
    echo ">>> Package ${package} completed with exit code: ${exit_code}"
    echo ">>> End time: $(date '+%Y-%m-%d %H:%M:%S')"

    # Analyze the output
    echo ""
    echo "Test Analysis for ${package}:"

    # Count pass/fail
    local pass_count=$(grep -c "^--- PASS:" ${log_file} || echo "0")
    local fail_count=$(grep -c "^--- FAIL:" ${log_file} || echo "0")
    local skip_count=$(grep -c "^--- SKIP:" ${log_file} || echo "0")

    echo "  Tests Passed: ${pass_count}"
    echo "  Tests Failed: ${fail_count}"
    echo "  Tests Skipped: ${skip_count}"

    # Show failures if any
    if [ ${fail_count} -gt 0 ]; then
        echo ""
        echo "  Failed tests:"
        grep "^--- FAIL:" ${log_file} | head -10
    fi

    # Check for panics
    if grep -q "panic:" ${log_file}; then
        echo ""
        echo "  WARNING: Panic detected!"
        grep -A 5 "panic:" ${log_file} | head -20
    fi

    # Check for build errors
    if grep -q "cannot find package\|undefined:\|cannot find module" ${log_file}; then
        echo ""
        echo "  Build/Compilation errors detected:"
        grep -E "cannot find package|undefined:|cannot find module" ${log_file} | head -10
    fi

    # Show package summary
    echo ""
    echo "  Package Summary:"
    grep -E "^(ok|FAIL)\s+github.com/gqls/agentchassis/test/integration/${package}" ${log_file} || echo "    No summary found"

    echo "========================================="
    echo ""

    return ${exit_code}
}

# Track results
FAILED_PACKAGES=""
PASSED_PACKAGES=""
SKIPPED_PACKAGES=""
TOTAL_TESTS_RUN=0
TOTAL_TESTS_PASSED=0
TOTAL_TESTS_FAILED=0

# Check which packages exist
echo "Discovering test packages..."
for package in agents database kafka; do
    if [ -d "./test/integration/${package}" ]; then
        echo "  ✓ Found: ${package}"
    else
        echo "  ✗ Missing: ${package}"
        SKIPPED_PACKAGES="${SKIPPED_PACKAGES} ${package}"
    fi
done
echo ""

# Run tests for each package
for package in agents database kafka; do
    if [ ! -d "./test/integration/${package}" ]; then
        echo "Skipping missing package: ${package}"
        continue
    fi

    if run_test_package ${package}; then
        echo "✓ Package ${package} PASSED"
        PASSED_PACKAGES="${PASSED_PACKAGES} ${package}"
    else
        echo "✗ Package ${package} FAILED"
        FAILED_PACKAGES="${FAILED_PACKAGES} ${package}"
    fi

    # Add some space between packages
    echo ""
    echo ""
done

# Final summary
echo "========================================="
echo "FINAL TEST SUMMARY"
echo "========================================="
echo "Execution completed at: $(date)"
echo ""
echo "Package Results:"
echo "  Passed packages: ${PASSED_PACKAGES:-none}"
echo "  Failed packages: ${FAILED_PACKAGES:-none}"
echo "  Skipped packages: ${SKIPPED_PACKAGES:-none}"
echo ""

# Aggregate all test results
echo "Aggregate Test Statistics:"
for package in agents database kafka; do
    if [ -f "${TEMP_DIR}/test_${package}.log" ]; then
        pass=$(grep -c "^--- PASS:" "${TEMP_DIR}/test_${package}.log" || echo "0")
        fail=$(grep -c "^--- FAIL:" "${TEMP_DIR}/test_${package}.log" || echo "0")
        TOTAL_TESTS_PASSED=$((TOTAL_TESTS_PASSED + pass))
        TOTAL_TESTS_FAILED=$((TOTAL_TESTS_FAILED + fail))
    fi
done
TOTAL_TESTS_RUN=$((TOTAL_TESTS_PASSED + TOTAL_TESTS_FAILED))

echo "  Total tests run: ${TOTAL_TESTS_RUN}"
echo "  Total passed: ${TOTAL_TESTS_PASSED}"
echo "  Total failed: ${TOTAL_TESTS_FAILED}"
echo ""

# Check for common issues
echo "Common Issues Check:"

# Check for timeout
if grep -q "panic: test timed out" ${TEMP_DIR}/*.log 2>/dev/null; then
    echo "  ⚠ Timeout detected in tests"
fi

# Check for race conditions
if grep -q "WARNING: DATA RACE" ${TEMP_DIR}/*.log 2>/dev/null; then
    echo "  ⚠ Race condition detected"
fi

# Check for compilation errors
if grep -q "build failed\|cannot find package" ${TEMP_DIR}/*.log 2>/dev/null; then
    echo "  ⚠ Build/compilation errors detected"
fi

echo ""
echo "Log files saved in: ${TEMP_DIR}"
echo "To review a specific package log: cat ${TEMP_DIR}/test_<package>.log"
echo ""

# Determine exit status
if [ -n "${FAILED_PACKAGES}" ]; then
    echo "✗ OVERALL RESULT: FAILED"
    echo "  Some tests failed. Review the output above for details."
    FINAL_EXIT=1
else
    echo "✓ OVERALL RESULT: PASSED"
    echo "  All tests passed successfully!"
    FINAL_EXIT=0
fi

echo "========================================="
echo ""

# Show how to get more details
if [ ${FINAL_EXIT} -ne 0 ]; then
    echo "For debugging, you can:"
    echo "  1. Review full logs: ls -la ${TEMP_DIR}/"
    echo "  2. Search for specific errors: grep -r 'ERROR\\|FAIL' ${TEMP_DIR}/"
    echo "  3. Check specific package: cat ${TEMP_DIR}/test_<package>.log"
    echo ""
fi

# Clean up old temp directories (keep last 5)
ls -dt /tmp/test_logs_* 2>/dev/null | tail -n +6 | xargs rm -rf 2>/dev/null || true

exit ${FINAL_EXIT}