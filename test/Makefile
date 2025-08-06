
# test/Makefile
.PHONY: all unit integration e2e performance setup clean

# Default target
all: unit integration

# Setup test environment
setup:
	@echo "Setting up test database..."
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < migrations/001_test_schema.sql
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < migrations/002_test_data.sql
	@echo "Test setup complete"

# Unit tests
unit:
	@echo "Running unit tests..."
	@go test ./unit/... -v -short

# Integration tests
integration: setup
	@echo "Running integration tests..."
	@go test ./integration/... -v

# E2E tests
e2e: setup
	@echo "Running E2E tests..."
	@./scripts/setup_e2e.sh
	@go test ./e2e/... -v
	@./scripts/cleanup_e2e.sh

# Performance tests
performance:
	@echo "Running performance tests..."
	@go test ./performance/... -bench=. -benchtime=10s

# Test specific agent
test-agent:
	@echo "Testing agent: $(AGENT)"
	@go run scripts/test_individual_agent.go -agent=$(AGENT)

# Run tests in Kubernetes
test-k8s:
	@kubectl apply -f k8s/jobs/test-spawning-job.yaml
	@kubectl wait --for=condition=complete job/test-spawning --timeout=300s
	@kubectl logs job/test-spawning

# Verify test results
verify:
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < scripts/verify_results.sql

# Clean up test data
clean:
	@echo "Cleaning up test data..."
	@kubectl delete job -l app=test --ignore-not-found
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-%';"


