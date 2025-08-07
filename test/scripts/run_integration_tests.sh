#!/bin/bash
# test/scripts/run_integration_tests.sh

set -e

echo "=== Running Integration Tests ==="

# Check prerequisites
echo "Checking prerequisites..."
if ! nc -z localhost 5432; then
    echo "PostgreSQL is not running on localhost:5432"
    exit 1
fi

if ! nc -z localhost 9092; then
    echo "Kafka is not running on localhost:9092"
    exit 1
fi

# Set up test database
echo "Setting up test database..."
psql -h localhost -U clients_user -d clients_db < migrations/001_test_schema.sql
psql -h localhost -U clients_user -d clients_db < migrations/002_test_data.sql

# Run integration tests
echo "Running integration tests..."
go test -v ./integration/... -count=1

# Clean up
echo "Cleaning up test data..."
psql -h localhost -U clients_user -d clients_db -c "DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-%';"

echo ""
echo "Integration tests completed!"