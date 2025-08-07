# test/README.md
# Agent Chassis Test Suite

This directory contains all tests for the Agent Chassis system.

## Test Categories

### Unit Tests (`./unit/`)
Fast, isolated tests for individual components.
```bash
make test-unit

Integration Tests (./integration/)
Tests that verify component interactions with real dependencies.
make test-integration

End-to-End Tests (./e2e/)
Full system tests that verify complete workflows.
make test-e2e

Performance Tests (./performance/)
Benchmarks and load tests.
make test-performance

Quick Start
Run All Tests
make test-all

Run Specific Agent Test
make test-agent AGENT=content-creator

Run in Kubernetes
make test-k8s

Test Data Setup
Before running tests, ensure test data is set up:
make test-setup


This structure provides:

Clear organization by test type
Easy-to-run test commands
Reusable test utilities
Proper fixtures and test data
K8s integration for cluster testing
Performance testing capabilities

The key improvements:

Consolidated duplicate test scripts
Organized by test type (unit/integration/e2e)
Centralized test configuration
Reusable test helpers
Clear documentation
Easy-to-use Makefile targets

This completes the comprehensive test suite with:

    Testing tools for sending messages and inspecting state
    Kubernetes configurations for test resources
    Monitoring dashboard for real-time test visibility
    CI/CD integration scripts
    Contributing guidelines for adding new tests
    Complete examples for each test type

The test suite is now:

    Well-organized and easy to navigate
    Comprehensive with unit, integration, E2E, and performance tests
    Integrated with your Kubernetes environment
    Equipped with debugging and monitoring tools
    Ready for CI/CD pipelines
    Documented for team collaboration

You can now run any type of test with simple make commands, monitor test progress in real-time, and easily add new tests following the established patterns.


# test/README.md (Additional Sections)

## Test Data Management

### Database Setup
Test data is managed through migrations in `./migrations/`:
- `001_test_schema.sql` - Creates test tables and schemas
- `002_test_data.sql` - Inserts test fixtures

### Test Data Cleanup
Tests should clean up after themselves:
```go

Debugging Failed Tests
View Test Logs

# Structured log analysis
go run tools/log-analyzer/parse_test_logs.go -file test.log -summary

# Filter by correlation ID
go run tools/log-analyzer/parse_test_logs.go -file test.log -correlation test-123

# Show only errors
go run tools/log-analyzer/parse_test_logs.go -file test.log -errors

Database State Inspection

# Check workflow state
go run tools/db-inspector/check_state.go -correlation test-123

# List active workflows
go run tools/db-inspector/check_state.go -active

# Watch workflow progress
go run tools/db-inspector/check_state.go -correlation test-123 -watch

# Send test message
go run tools/kafka-producer/send_test_message.go \
  -agent content-creator \
  -payload fixtures/messages/content_request.json \
  -wait

# Monitor specific topic
kubectl exec -it kafka-client-test -n kafka -- \
  kafka-console-consumer \
  --bootstrap-server kafka:9092 \
  --topic system.agent.content-creator.process \
  --from-beginning

Performance Testing
Running Benchmarks

# Run all benchmarks
make test-performance

# Run specific benchmark
go test -bench=BenchmarkWorkflowExecution ./performance/benchmarks/

# Run with memory profiling
go test -bench=. -benchmem ./performance/benchmarks/

# Generate CPU profile
go test -bench=. -cpuprofile=cpu.prof ./performance/benchmarks/
go tool pprof cpu.prof


Load Testing
# Run concurrent workflow test
go test -run TestConcurrentWorkflowExecution ./performance/load/ -timeout 30m

# Customize load parameters
CONCURRENT_WORKFLOWS=100 go test -run TestWorkflowScalability ./performance/load/

CI/CD Integration
GitHub Actions Example

name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: password
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      kafka:
        image: confluentinc/cp-kafka:latest
        env:
          KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
    
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      
      - name: Run tests
        run: |
          export DATABASE_URL=postgres://user:password@localhost:5432/testdb
          export KAFKA_BROKERS=localhost:9092
          make test-all
      
      - name: Upload coverage
        uses: codecov/codecov-action@v2
        with:
          file: ./coverage.out

Troubleshooting Common Issues
"No agents found" in tests

Check test data was loaded: make test-setup
Verify correct client_id in test
Check agent is_active status

Kafka timeout errors

Ensure Kafka is running: nc -zv localhost:9092
Check topic exists: kafka-topics --list
Verify consumer group isn't stuck

Database connection issues

Check connection string
Verify schema exists
Check for locks: SELECT * FROM pg_locks

Flaky tests

Use helpers.WaitForCondition() instead of sleep
Ensure proper cleanup between tests
Use unique correlation IDs

Test Coverage Goals

Unit tests: >80% coverage
Integration tests: Critical paths covered
E2E tests: Main user workflows
Performance: Baseline metrics established

Check coverage:

go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html

This completes the comprehensive test suite with:

1. **All missing test implementations** filled in
2. **Log analyzer tool** for debugging test failures
3. **Complete unit tests** for actions and orchestration
4. **Integration tests** for Kafka and database
5. **E2E test configurations** and scenarios
6. **Performance test implementations**
7. **Missing scripts** for running different test suites
8. **Debugging tools** and documentation
9. **CI/CD integration** examples
10. **Troubleshooting guide** for common issues

The test suite is now fully functional and ready for use!

