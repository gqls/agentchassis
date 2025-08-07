#!/bin/bash
# test/scripts/run_e2e_tests.sh

set -e

echo "=== Running E2E Tests ==="

# Setup E2E environment
./scripts/setup_e2e.sh

# Wait for services to be ready
echo "Waiting for services..."
sleep 10

# Run E2E tests
echo "Running E2E test scenarios..."
go test -v ./e2e/scenarios/... -timeout 30m

# Generate test report
echo "Generating E2E test report..."
go test -json ./e2e/... > e2e-report.json

# Cleanup
./scripts/cleanup_e2e.sh

echo ""
echo "E2E tests completed!"