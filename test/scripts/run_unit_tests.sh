#!/bin/bash
# test/scripts/run_unit_tests.sh

set -e

echo "=== Running Unit Tests ==="

# Set test environment
export GO_ENV=test
export DB_HOST=localhost
export DB_NAME=test_db

# Run unit tests with coverage
echo "Running unit tests with coverage..."
go test -v -short -coverprofile=coverage.out ./unit/...

# Generate coverage report
echo "Generating coverage report..."
go tool cover -func=coverage.out | grep total

# Run specific test suites
echo ""
echo "Test Results by Package:"
go test -v -short ./unit/actions/
go test -v -short ./unit/orchestration/
go test -v -short ./unit/helpers/

echo ""
echo "Unit tests completed!"