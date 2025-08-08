# test/Makefile (Updated with build commands)
.PHONY: all unit integration e2e performance setup clean build

# Variables
REGISTRY ?= docker.io/aqls
IMAGE_TAG ?= latest
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

KAFKA_SERVICE ?= personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092

export DOCKER_USERNAME=aqls
export DOCKER_PASSWORD=AaD02432123!
export DOCKER_EMAIL=aaa@designconsultancy.co.uk

# Colors
YELLOW := \033[1;33m
GREEN := \033[1;32m
RED := \033[1;31m
NC := \033[0m

# Default target
all: unit integration

#################################
# Build Commands
#################################

# Build all test components
.PHONY: build
build: build-test-runner build-test-tools build-test-harness build-test-monitor build-mock-agents ## Build all test components

.PHONY: build-test-runner
build-test-runner: ## Build the main test runner binary
	@echo "$(YELLOW)Building test runner...$(NC)"
	@go build -ldflags="-X main.Version=$(IMAGE_TAG) -X main.BuildDate=$(BUILD_DATE)" \
		-o bin/test-runner ./cmd/runner
	@echo "$(GREEN)✓ Test runner built: bin/test-runner$(NC)"

.PHONY: build-test-tools
build-test-tools: ## Build all test utility tools
	@echo "$(YELLOW)Building test tools...$(NC)"
	@mkdir -p bin/tools
	
	# Log analyzer
	@go build -o bin/tools/log-analyzer ./tools/log-analyzer/parse_test_logs.go
	@echo "$(GREEN)✓ Built: log-analyzer$(NC)"
	
	# DB inspector
	@go build -o bin/tools/db-inspector ./tools/db-inspector/check_state.go
	@echo "$(GREEN)✓ Built: db-inspector$(NC)"
	
	# Kafka producer
	@go build -o bin/tools/kafka-producer ./tools/kafka-producer/send_test_message.go
	@echo "$(GREEN)✓ Built: kafka-producer$(NC)"
	
	# Test dashboard
	@go build -o bin/tools/test-dashboard ./tools/dashboard/test_dashboard.go
	@echo "$(GREEN)✓ Built: test-dashboard$(NC)"
	
	# Individual agent tester
	@go build -o bin/tools/test-agent ./scripts/test_individual_agent.go
	@echo "$(GREEN)✓ Built: test-agent$(NC)"

.PHONY: build-test-harness
build-test-harness: ## Build test harness Docker image
	@echo "$(YELLOW)Building test harness Docker image...$(NC)"
	@docker build -f docker/test-harness.dockerfile \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg VERSION=$(IMAGE_TAG) \
		-t $(REGISTRY)/test-harness:$(IMAGE_TAG) \
		-t $(REGISTRY)/test-harness:latest \
		..
	@echo "$(GREEN)✓ Test harness image built: $(REGISTRY)/test-harness:$(IMAGE_TAG)$(NC)"

.PHONY: build-test-monitor
build-test-monitor: ## Build workflow monitor test image
	@echo "$(YELLOW)Building test monitor image...$(NC)"
	@docker build -f docker/test-monitor.dockerfile \
		-t $(REGISTRY)/test-monitor:$(IMAGE_TAG) \
		-t $(REGISTRY)/test-monitor:latest \
		..
	@echo "$(GREEN)✓ Test monitor image built$(NC)"

.PHONY: build-mock-agents
build-mock-agents: ## Build mock agent images for testing
	@echo "$(YELLOW)Building mock agents...$(NC)"
	@for agent in generic reasoning content-creator web-search; do \
		echo "Building mock-$$agent..."; \
		docker build -f docker/mock-$$agent.dockerfile \
			-t $(REGISTRY)/mock-$$agent:$(IMAGE_TAG) \
			..; \
	done
	@echo "$(GREEN)✓ Mock agents built$(NC)"

#################################
# Push Commands
#################################

.PHONY: push-test-images
push-test-images: ## Push all test images to registry
	@echo "$(YELLOW)Pushing test images...$(NC)"
	@docker push $(REGISTRY)/test-harness:$(IMAGE_TAG)
	@docker push $(REGISTRY)/test-harness:latest
	@docker push $(REGISTRY)/test-monitor:$(IMAGE_TAG)
	@docker push $(REGISTRY)/test-monitor:latest
	@echo "$(GREEN)✓ Test images pushed$(NC)"

#################################
# Original Test Commands (Updated)
#################################

# Setup test environment
setup:
	@echo "$(YELLOW)Setting up test database...$(NC)"
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < migrations/001_test_schema.sql
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < migrations/002_test_data.sql
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < migrations/003_human_tasks.sql 2>/dev/null || true
	@echo "$(GREEN)✓ Test setup complete$(NC)"

# Unit tests
unit: build-test-tools
	@echo "$(YELLOW)Running unit tests...$(NC)"
	@gotestsum --format testname -- -v -short -coverprofile=coverage-unit.out ./unit/...

# Integration tests  
integration: setup build-test-tools
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@gotestsum --format testname -- -v -coverprofile=coverage-integration.out ./integration/...

# E2E tests
e2e: setup build-test-tools
	@echo "$(YELLOW)Running E2E tests...$(NC)"
	@./scripts/setup_e2e.sh
	@gotestsum --format testname -- -v -timeout=30m -coverprofile=coverage-e2e.out ./e2e/...
	@./scripts/cleanup_e2e.sh

# Performance tests
performance: build-test-tools
	@echo "$(YELLOW)Running performance tests...$(NC)"
	@go test ./performance/... -bench=. -benchmem -benchtime=10s | tee performance-results.txt

# Run tests in Kubernetes
test-k8s: build-test-harness push-test-images
	@kubectl apply -f k8s/jobs/test-spawning-job.yaml
	@kubectl wait --for=condition=complete job/test-spawning --timeout=300s -n ai-persona-system
	@kubectl logs job/test-spawning -n ai-persona-system

# Verify test results
verify:
	@echo "$(YELLOW)Verifying test results...$(NC)"
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < scripts/verify_results.sql

# Clean up test data
clean:
	@echo "$(YELLOW)Cleaning up test data...$(NC)"
	@kubectl delete job -l app=test -n ai-persona-system --ignore-not-found
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-%';"
	@rm -rf bin/ coverage-*.out performance-results.txt

#################################
# Test Running Tools
#################################

.PHONY: run-log-analyzer
run-log-analyzer: build-test-tools ## Run log analyzer on test results
	@echo "$(YELLOW)Running log analyzer...$(NC)"
	@./bin/tools/log-analyzer -file=test.log -summary

.PHONY: run-db-inspector  
run-db-inspector: build-test-tools ## Run database state inspector
	@echo "$(YELLOW)Running DB inspector...$(NC)"
	@./bin/tools/db-inspector -db="postgres://clients_user:password@localhost:5432/clients_db"

.PHONY: run-test-dashboard
run-test-dashboard: build-test-tools ## Run test monitoring dashboard
	@echo "$(YELLOW)Starting test dashboard on :8090...$(NC)"
	@./bin/tools/test-dashboard

#################################
# Coverage Commands
#################################

.PHONY: coverage
coverage: ## Generate coverage report
	@echo "$(YELLOW)Generating coverage report...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

.PHONY: coverage-merge
coverage-merge: ## Merge coverage reports from different test suites
	@echo "$(YELLOW)Merging coverage reports...$(NC)"
	@gocov convert coverage-unit.out > coverage-unit.json
	@gocov convert coverage-integration.out > coverage-integration.json  
	@gocov convert coverage-e2e.out > coverage-e2e.json
	@echo '{"packages":[]}' | jq -s '.[0].packages = ([.[].packages] | flatten) | .[0]' \
		coverage-unit.json coverage-integration.json coverage-e2e.json > coverage-merged.json
	@gocov-html < coverage-merged.json > coverage-merged.html
	@echo "$(GREEN)✓ Merged coverage report: coverage-merged.html$(NC)"

#################################
# Additional Test Targets
#################################

.PHONY: test-groups
test-groups: setup ## Test agent group functionality
	@echo "$(YELLOW)Testing agent groups...$(NC)"
	@go test -v ./integration/agents/agent_group_test.go

.PHONY: test-human-loop
test-human-loop: setup ## Test human-in-the-loop workflows
	@echo "$(YELLOW)Testing human-in-the-loop...$(NC)"
	@go test -v ./e2e/scenarios/human_in_loop_test.go

.PHONY: test-group-human-integration
test-group-human-integration: setup ## Test groups with human approval
	@echo "$(YELLOW)Testing group + human integration...$(NC)"
	@kubectl apply -f fixtures/workflows/group_human_workflow.yaml
	@go test -v -run TestGroupHumanIntegration ./e2e/

.PHONY: simulate-human-response
simulate-human-response: ## Simulate human task completion
	@echo "$(YELLOW)Simulating human response...$(NC)"
	@kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
		"UPDATE human_tasks SET status = 'COMPLETED', \
		 response_data = '{\"approved\": true, \"feedback\": \"Approved by test\"}' \
		 WHERE status = 'PENDING' AND correlation_id LIKE 'test-%' \
		 LIMIT 1;"

#################################
# Test Harness Commands (Updated)
#################################

# Build and push before deploying
.PHONY: harness-deploy
harness-deploy: build-test-harness push-test-images ## Deploy test harness to Kubernetes
	@echo "$(YELLOW)Checking prerequisites...$(NC)"
	@kubectl get secret docker-hub-creds -n ai-persona-system > /dev/null 2>&1 || \
		(echo "$(RED)Error: Docker Hub secret not found. Run 'make setup-docker-secret' first$(NC)" && exit 1)
	@echo "$(YELLOW)Deploying test harness to Kubernetes...$(NC)"
	@sed 's|$$(REGISTRY)|$(REGISTRY)|g; s|$$(IMAGE_TAG)|$(IMAGE_TAG)|g; s|$$(TEST_SUITE)|all|g; s|$$(TEST_COMMAND)|test|g' \
		k8s/jobs/test-harness-job.yaml | kubectl apply -f -
	@echo "$(GREEN)✓ Test harness deployed$(NC)"

.PHONY: harness-run
harness-run: build-test-harness push-test-images ## Run tests in Kubernetes
	@echo "$(YELLOW)Running tests in Kubernetes...$(NC)"
	@kubectl delete job test-harness -n ai-persona-system --ignore-not-found
	@sed 's|$$(REGISTRY)|$(REGISTRY)|g; s|$$(IMAGE_TAG)|$(IMAGE_TAG)|g; s|$$(TEST_SUITE)|$(TEST_SUITE)|g; s|$$(TEST_COMMAND)|test|g' \
		k8s/jobs/test-harness-job.yaml | kubectl apply -f -
	@echo "Waiting for test job to complete..."
	@kubectl wait --for=condition=complete job/test-harness -n ai-persona-system --timeout=1h || \
		(kubectl logs job/test-harness -n ai-persona-system && exit 1)

.PHONY: harness-shell
harness-shell: build-test-harness push-test-images ## Open shell in test harness pod
	@echo "$(YELLOW)Starting test harness shell...$(NC)"
	@kubectl run test-harness-shell -n ai-persona-system \
		--image=$(REGISTRY)/test-harness:$(IMAGE_TAG) \
		--rm -it --restart=Never \
		--env="DB_HOST=postgres-clients" \
		--env="DB_USER=clients_user" \
		--env="DB_NAME=clients_db" \
		--env="DB_PASSWORD=$${CLIENTS_DB_PASSWORD}" \
		--env="KAFKA_BROKERS=$(KAFKA_SERVICE)" \
		--env="CLIENT_ID=demo_client" \
		-- shell

.PHONY: harness-logs
harness-logs: ## View test harness logs
	@kubectl logs -f job/test-harness -n ai-persona-system

.PHONY: harness-results
harness-results: ## Copy test results from PVC
	@echo "$(YELLOW)Copying test results...$(NC)"
	@mkdir -p ./results
	@kubectl cp ai-persona-system/test-harness:/results ./results
	@echo "$(GREEN)✓ Results copied to ./results$(NC)"

.PHONY: harness-clean
harness-clean: ## Clean up test harness resources
	@kubectl delete job test-harness -n ai-persona-system --ignore-not-found
	@kubectl delete pod -l app=test-harness -n ai-persona-system --ignore-not-found

# Run specific test suites
.PHONY: harness-unit
harness-unit: ## Run unit tests in Kubernetes
	@TEST_SUITE=unit $(MAKE) harness-run

.PHONY: harness-integration  
harness-integration: ## Run integration tests in Kubernetes
	@TEST_SUITE=integration $(MAKE) harness-run

.PHONY: harness-e2e
harness-e2e: ## Run E2E tests in Kubernetes
	@TEST_SUITE=e2e $(MAKE) harness-run

.PHONY: harness-performance
harness-performance: ## Run performance tests in Kubernetes
	@TEST_SUITE=performance $(MAKE) harness-run


#################################
# Agent Spawning Tests
#################################
.PHONY: test-spawning-local
test-spawning-local: ## Run agent spawning tests locally (requires port-forwards)
	@echo "$(YELLOW)Running agent spawning tests locally...$(NC)"
	@echo "$(YELLOW)Setting up port forwards...$(NC)"
	@kubectl port-forward -n $(PROJECT_NAME) svc/postgres-clients 5432:5432 &
	@kubectl port-forward -n $(PROJECT_NAME) svc/kafka-cluster-kafka-bootstrap 9092:9092 &
	@sleep 5
	@echo "$(GREEN)Running tests...$(NC)"
	DB_HOST=localhost \
	DB_NAME=clients_db \
	DB_USER=clients_user \
	DB_PASSWORD=$${CLIENTS_DB_PASSWORD} \
	KAFKA_BROKERS=localhost:9092 \
	CLIENT_ID=demo_client \
	go run cmd/test-spawning/main.go
	@echo "$(GREEN)Tests completed!$(NC)"
	@pkill -f "port-forward.*postgres-clients" || true
	@pkill -f "port-forward.*kafka-cluster" || true

.PHONY: test-spawning-quick
test-spawning-quick: ## Quick spawn test using kubectl run
	@echo "$(YELLOW)Running quick spawn test...$(NC)"
	kubectl run spawn-test-$(shell date +%s) \
		--image=$(REGISTRY)/agent-chassis:$(IMAGE_TAG) \
		--rm -it --restart=Never \
		-n $(PROJECT_NAME) \
		--env="DB_HOST=postgres-clients" \
		--env="DB_NAME=clients_db" \
		--env="DB_USER=clients_user" \
		--env="DB_PASSWORD=$${CLIENTS_DB_PASSWORD}" \
		--env="CLIENT_ID=demo_client" \
		-- /bin/sh -c 'echo "Quick spawn test completed"'

.PHONY: test-spawning-watch
test-spawning-watch: ## Watch spawning test logs
	@echo "$(YELLOW)Watching test logs...$(NC)"
	kubectl logs -f -n $(PROJECT_NAME) -l job-name --tail=100

.PHONY: test-spawning-cleanup
test-spawning-cleanup: ## Clean up test jobs
	@echo "$(YELLOW)Cleaning up test jobs...$(NC)"
	kubectl delete jobs -n $(PROJECT_NAME) -l app=test-spawning --ignore-not-found
	@echo "$(GREEN)Cleanup complete$(NC)"

.PHONY: test-spawning-verify
test-spawning-verify: ## Verify spawning test results in database
	@echo "$(YELLOW)Verifying test results...$(NC)"
	@kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT id, name, config->>'agent_type' as type, is_active, created_at \
		 FROM client_demo_client.agent_instances \
		 WHERE created_at > NOW() - INTERVAL '1 hour' \
		 ORDER BY created_at DESC LIMIT 10;"
	@echo ""
	@kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT correlation_id, status, current_step, \
		 execution_metadata->>'completed_steps' as completed \
		 FROM orchestrator_state \
		 WHERE created_at > NOW() - INTERVAL '1 hour' \
		 ORDER BY created_at DESC LIMIT 10;"
	@echo ""
	@kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT id, name, group_type, usage_count, last_used_at \
		 FROM agent_groups \
		 WHERE last_used_at > NOW() - INTERVAL '1 hour' \
		 ORDER BY last_used_at DESC NULLS LAST;"

.PHONY: test-spawning-setup
test-spawning-setup: ## Set up database for spawning tests
	@echo "$(YELLOW)Setting up test data...$(NC)"
	@kubectl exec -i postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db << 'EOSQL'
	-- Ensure client schema exists
	CREATE SCHEMA IF NOT EXISTS client_demo_client;
	-- Create tables if they don't exist
	CREATE TABLE IF NOT EXISTS client_demo_client.agent_instances (
	     id UUID PRIMARY KEY,
	     template_id UUID NOT NULL,
	     owner_user_id VARCHAR(255) NOT NULL,
	     name VARCHAR(255) NOT NULL,
	     config JSONB NOT NULL,
	     is_active BOOLEAN DEFAULT true,
	     created_at TIMESTAMP DEFAULT NOW(),
	     updated_at TIMESTAMP DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS agent_groups (
	     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	     name VARCHAR(255) NOT NULL,
	     group_type VARCHAR(100) NOT NULL,
	     version VARCHAR(20) DEFAULT '1.0.0',
	     parent_id UUID,
	     agent_configs JSONB NOT NULL,
	     orchestration_workflow JSONB NOT NULL,
	     capabilities JSONB DEFAULT '[]',
	     usage_count INTEGER DEFAULT 0,
	     performance_metrics JSONB DEFAULT '{}',
	     mutation_history JSONB DEFAULT '[]',
	     last_used_at TIMESTAMP,
	     is_active BOOLEAN DEFAULT true,
	     created_at TIMESTAMP DEFAULT NOW(),
	     updated_at TIMESTAMP DEFAULT NOW()
	);
	-- Insert test group
	INSERT INTO agent_groups (name, group_type, agent_configs, orchestration_workflow)
	VALUES (
	     'Website Builder Team',
	     'website-builder',
	     '[
	        {"role": "architect", "agent_type": "site-architect"},
	        {"role": "designer", "agent_type": "visual-designer"},
	        {"role": "developer", "agent_type": "html-developer"}
	     ]'::jsonb,
	     '{
	        "start_step": "plan",
	        "steps": {
	           "plan": {"action": "create_site_plan", "next_step": "design"},
	           "design": {"action": "create_design", "next_step": "develop"},
	           "develop": {"action": "build_html", "next_step": "complete"},
	           "complete": {"action": "complete_workflow"}
	        }
	     }'::jsonb
	) ON CONFLICT DO NOTHING;
	\echo 'Test data setup complete!'
	EOSQL
	@echo "$(GREEN)Test setup complete$(NC)"

.PHONY: test-spawning-full
test-spawning-full: test-spawning-setup test-spawning test-spawning-verify ## Run full spawning test suite
	@echo "$(GREEN)Full test suite completed!$(NC)"

# Build test image if needed
.PHONY: build-test-spawning
build-test-spawning: ## Build test spawning image
	@echo "$(YELLOW)Building test spawning image...$(NC)"
	docker build -t $(REGISTRY)/test-spawning:$(IMAGE_TAG) \
		-f build/docker/test/test-spawning.dockerfile .

#################################
# Testing
#################################
.PHONY: test-all
test-all: test-unit test-integration test-e2e ## Run all tests

.PHONY: test-unit
test-unit: ## Run unit tests
	@echo "$(YELLOW)Running unit tests...$(NC)"
	@cd test && make unit

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@cd test && make integration

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(YELLOW)Running E2E tests...$(NC)"
	@cd test && make e2e

.PHONY: test-spawning
test-spawning: ## Test agent spawning system
	@echo "$(YELLOW)Testing agent spawning...$(NC)"
	@kubectl apply -f test/k8s/jobs/test-spawning-job.yaml
	@kubectl wait --for=condition=complete job/test-agent-spawning --timeout=300s -n $(PROJECT_NAME)
	@kubectl logs job/test-agent-spawning -n $(PROJECT_NAME)

.PHONY: test-agent
test-agent: ## Test specific agent (AGENT=content-creator)
	@echo "$(YELLOW)Testing $(AGENT) agent...$(NC)"
	@cd test && make test-agent AGENT=$(AGENT)

.PHONY: test-verify
test-verify: ## Verify test results in database
	@echo "$(YELLOW)Verifying test results...$(NC)"
	@cd test && make verify

.PHONY: test-clean
test-clean: ## Clean up test resources
	@echo "$(YELLOW)Cleaning up test resources...$(NC)"
	@cd test && make clean

#################################
# CI/CD Commands
#################################

.PHONY: ci-test
ci-test: install-tools build test-all ## Run full CI test suite

.PHONY: test-all
test-all: unit integration e2e ## Run all test suites locally

.PHONY: test-all-k8s
test-all-k8s: harness-deploy harness-run harness-results ## Run all tests in Kubernetes



#################################
# Help
#################################

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-30s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

### claude extras
#################################
# Kubernetes Deployment Commands
#################################

.PHONY: deploy-test-infrastructure
deploy-test-infrastructure: ## Deploy all test infrastructure to Kubernetes
	@echo "$(YELLOW)Deploying test infrastructure to Kubernetes...$(NC)"
	# Deploy test ConfigMaps
	@kubectl apply -f k8s/configmaps/test-configs.yaml
	@kubectl apply -f k8s/configmaps/test-scripts.yaml
	# Deploy test pods for interactive testing
	@kubectl apply -f k8s/pods/kafka-test-pod.yaml
	@kubectl apply -f k8s/pods/agent-test-pod.yaml
	@echo "$(GREEN)✓ Test infrastructure deployed$(NC)"

.PHONY: deploy-mock-agents
deploy-mock-agents: build-mock-agents push-mock-agents ## Deploy mock agents to Kubernetes
	@echo "$(YELLOW)Deploying mock agents...$(NC)"
	@for agent in generic reasoning content-creator web-search; do \
		kubectl run mock-$$agent \
			--image=$(REGISTRY)/mock-$$agent:$(IMAGE_TAG) \
			--env="KAFKA_BROKERS=$(KAFKA_SERVICE)" \
			--env="KAFKA_TOPIC=system.agent.$$agent.process" \
			-n ai-persona-system \
			--restart=Always || true; \
	done
	@echo "$(GREEN)✓ Mock agents deployed$(NC)"

.PHONY: push-mock-agents
push-mock-agents: ## Push mock agent images to registry
	@echo "$(YELLOW)Pushing mock agent images...$(NC)"
	@for agent in generic reasoning content-creator web-search; do \
		docker push $(REGISTRY)/mock-$$agent:$(IMAGE_TAG); \
	done
	@echo "$(GREEN)✓ Mock agents pushed$(NC)"

.PHONY: deploy-all
deploy-all: deploy-test-infrastructure harness-deploy ## Deploy everything needed for testing
	@echo "$(GREEN)✓ All test components deployed to Kubernetes$(NC)"
	@echo ""
	@echo "Next steps:"
	@echo "  make harness-run          # Run all tests"
	@echo "  make harness-shell        # Open interactive test shell"
	@echo "  make test-status          # Check test status"
	@echo "  make harness-logs         # View test logs"

.PHONY: test-status
test-status: ## Check status of all test components
	@echo "$(YELLOW)=== Test Infrastructure Status ===$(NC)"
	@echo ""
	@echo "Test Harness Job:"
	@kubectl get job test-harness -n ai-persona-system --ignore-not-found || echo "  Not deployed"
	@echo ""
	@echo "Test Pods:"
	@kubectl get pods -n ai-persona-system -l app=test
	@echo ""
	@echo "Mock Agents:"
	@kubectl get pods -n ai-persona-system | grep mock- || echo "  No mock agents running"
	@echo ""
	@echo "Recent Test Workflows:"
	@kubectl exec postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -t -c \
		"SELECT correlation_id, status, current_step, \
		 EXTRACT(EPOCH FROM (NOW() - created_at))::int as seconds_ago \
		 FROM orchestrator_state \
		 WHERE correlation_id LIKE 'test-%' \
		 ORDER BY created_at DESC LIMIT 5;" 2>/dev/null || echo "  No recent test workflows"

.PHONY: test-quick-verify
test-quick-verify: ## Quick verification that test infrastructure is working
	@echo "$(YELLOW)Running quick infrastructure test...$(NC)"
	# Test database connectivity
	@kubectl exec postgres-clients-0 -n ai-persona-system -- \
		psql -U clients_user -d clients_db -c "SELECT 1" > /dev/null && \
		echo "$(GREEN)✓ Database accessible$(NC)" || \
		echo "$(RED)✗ Database not accessible$(NC)"
	# Test Kafka connectivity
	@kubectl exec kafka-cluster-kafka-0 -n ai-persona-system -c kafka -- \
		kafka-topics.sh --bootstrap-server localhost:9092 --list > /dev/null 2>&1 && \
		echo "$(GREEN)✓ Kafka accessible$(NC)" || \
		echo "$(RED)✗ Kafka not accessible$(NC)"
	# Check test harness image
	@kubectl run test-harness-check --rm -it --restart=Never \
		--image=$(REGISTRY)/test-harness:$(IMAGE_TAG) \
		-n ai-persona-system \
		-- echo "Test harness image OK" && \
		echo "$(GREEN)✓ Test harness image available$(NC)" || \
		echo "$(RED)✗ Test harness image not available$(NC)"

.PHONY: clean-test-k8s
clean-test-k8s: harness-clean ## Clean up all test resources from Kubernetes
	@echo "$(YELLOW)Cleaning up test resources...$(NC)"
	# Delete test pods
	@kubectl delete pod kafka-test-pod agent-test-pod kafkacat-test-pod -n ai-persona-system --ignore-not-found
	# Delete mock agents
	@kubectl delete pod -l app=mock-agent -n ai-persona-system --ignore-not-found
	@for agent in generic reasoning content-creator web-search; do \
		kubectl delete pod mock-$$agent -n ai-persona-system --ignore-not-found; \
	done
	# Delete test ConfigMaps
	@kubectl delete configmap test-configs test-scripts -n ai-persona-system --ignore-not-found
	# Clean test data from database
	@kubectl exec postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
		"DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-%';" || true
	@echo "$(GREEN)✓ Test resources cleaned$(NC)"

.PHONY: test-e2e-k8s
test-e2e-k8s: deploy-all ## Run E2E tests in Kubernetes with full setup
	@echo "$(YELLOW)Running E2E tests in Kubernetes...$(NC)"
	@kubectl apply -f k8s/jobs/test-e2e-job.yaml
	@kubectl wait --for=condition=complete job/test-e2e-suite -n ai-persona-system --timeout=1h || \
		(kubectl logs job/test-e2e-suite -n ai-persona-system && exit 1)
	@echo "$(GREEN)✓ E2E tests completed$(NC)"

# Convenience target to run a quick test
.PHONY: test-quick
test-quick: deploy-test-infrastructure ## Run a quick test to verify everything works
	@echo "$(YELLOW)Running quick test...$(NC)"
	@kubectl run quick-test-$(shell date +%s) \
		--image=$(REGISTRY)/test-harness:$(IMAGE_TAG) \
		--rm -it --restart=Never \
		-n ai-persona-system \
		--env="DB_HOST=postgres-clients" \
		--env="DB_USER=clients_user" \
		--env="DB_NAME=clients_db" \
		--env="DB_PASSWORD=$${CLIENTS_DB_PASSWORD}" \
		--env="KAFKA_BROKERS=$(KAFKA_SERVICE)" \
		--env="CLIENT_ID=demo_client" \
		--env="TEST_SUITE=unit" \
		--env="TEST_PATTERN=TestValidWorkflow" \
		-- test
	@echo "$(GREEN)✓ Quick test completed$(NC)"


#################################
# Setup Commands
#################################

.PHONY: setup-docker-secret
setup-docker-secret: ## Create Docker Hub pull secret from environment variables
	@if [ -z "$$DOCKER_USERNAME" ] || [ -z "$$DOCKER_PASSWORD" ]; then \
		echo "$(RED)Error: DOCKER_USERNAME and DOCKER_PASSWORD environment variables must be set$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Creating Docker Hub secret from environment variables...$(NC)"
	@kubectl create secret docker-registry docker-hub-creds \
		--docker-server=docker.io \
		--docker-username="$$DOCKER_USERNAME" \
		--docker-password="$$DOCKER_PASSWORD" \
		--docker-email="aaa@designconsultancy.co.uk" \
		-n ai-persona-system \
		--dry-run=client -o yaml | kubectl apply -f -
	@echo "$(GREEN)✓ Docker Hub secret created/updated$(NC)"

.PHONY: setup-all
setup-all: setup setup-docker-secret ## Complete setup including Docker secret
	@echo "$(GREEN)✓ All setup complete$(NC)"

.PHONY: verify-docker-secret
verify-docker-secret: ## Verify Docker secret is configured correctly
	@echo "$(YELLOW)Verifying Docker secret...$(NC)"
	@kubectl get secret docker-hub-creds -n ai-persona-system -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | jq '.auths' | grep -q "docker.io" && \
		echo "$(GREEN)✓ Docker Hub secret exists and contains docker.io credentials$(NC)" || \
		echo "$(RED)✗ Docker Hub secret not found or invalid$(NC)"

