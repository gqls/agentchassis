# test/CONTRIBUTING.md
# Contributing to Agent Chassis Tests

## Test Organization

Tests are organized by type:
- `unit/` - Fast, isolated component tests
- `integration/` - Tests with real dependencies (DB, Kafka)
- `e2e/` - Full system tests
- `performance/` - Benchmarks and load tests

## Writing Tests

### Unit Tests
```go
func TestComponentName(t *testing.T) {
    // Use table-driven tests
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

Integration Tests

Always check testing.Short() to skip in unit test runs
Use real dependencies but isolated test data
Clean up after tests

E2E Tests

Use correlation IDs starting with test-e2e-
Verify complete workflows
Test error scenarios

Running Tests
Local Development

# Run unit tests only
make test-unit

# Run specific test
go test -run TestSpawnAgentAction ./unit/actions/

# Run with coverage
go test -cover ./...

In Kubernetes

# Run test job
make test-k8s

# Run specific agent test
make test-agent AGENT=content-creator

Adding New Tests

Choose appropriate test type (unit/integration/e2e)
Create test file in correct directory
Use existing helpers from test/unit/helpers/
Add fixtures if needed to test/fixtures/
Update Makefile if adding new test category

Test Data

Use test/migrations/ for database setup
Correlation IDs should start with test-
Clean up test data after runs

Debugging Tests
View Logs

# Kafka messages
kubectl exec -it kafka-client-test -n kafka -- \
  kafka-console-consumer \
  --bootstrap-server kafka:9092 \
  --topic system.agent.generic.process \
  --from-beginning

# Database state
kubectl exec -it postgres-clients-0 -- \
  psql -U clients_user -d clients_db \
  -c "SELECT * FROM orchestrator_state WHERE correlation_id LIKE 'test-%'"


Test Dashboard

# Run test dashboard
go run test/tools/dashboard/test_dashboard.go

# Access at http://localhost:8090

Best Practices

Isolation: Tests should not depend on external state
Clarity: Test names should describe what they test
Speed: Unit tests should complete in milliseconds
Reliability: Tests should not be flaky
Coverage: Aim for >80% coverage on critical paths

## 15. Final Test Script for CI/CD

```bash
#!/bin/bash
# test/scripts/ci_test_runner.sh

set -e

echo "=== Agent Chassis CI Test Runner ==="
echo "Environment: ${ENVIRONMENT:-development}"
echo "Test Suite: ${TEST_SUITE:-all}"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test results
FAILED=0

# Run test suite
run_tests() {
    local suite=$1
    echo -e "${YELLOW}Running ${suite} tests...${NC}"
    
    if make test-${suite}; then
        echo -e "${GREEN}✓ ${suite} tests passed${NC}"
    else
        echo -e "${RED}✗ ${suite} tests failed${NC}"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

# Setup
echo -e "${YELLOW}Setting up test environment...${NC}"
make test-setup || exit 1

# Run tests based on suite
case ${TEST_SUITE} in
    unit)
        run_tests unit
        ;;
    integration)
        run_tests integration
        ;;
    e2e)
        run_tests e2e
        ;;
    performance)
        run_tests performance
        ;;
    all)
        run_tests unit
        run_tests integration
        run_tests e2e
        ;;
    *)
        echo -e "${RED}Unknown test suite: ${TEST_SUITE}${NC}"
        exit 1
        ;;
esac

# Generate test report
echo -e "${YELLOW}Generating test report...${NC}"
go test -json ./... > test-report.json || true

# Coverage report
echo -e "${YELLOW}Generating coverage report...${NC}"
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Cleanup
echo -e "${YELLOW}Cleaning up...${NC}"
make test-clean

# Summary
echo ""
echo "=== Test Summary ==="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}${FAILED} test suite(s) failed${NC}"
    exit 1
fi