# test/Makefile (Updated with build commands)
.PHONY: all unit integration e2e performance setup clean build

# Variables
REGISTRY ?= docker.io/aqls
IMAGE_TAG ?= latest
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

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
build: build-test-runner build-test-tools build-test-harness ## Build all test components

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
# Install Commands
#################################

.PHONY: install-tools
install-tools: ## Install required testing tools
	@echo "$(YELLOW)Installing test tools...$(NC)"
	@go install gotest.tools/gotestsum@latest
	@go install github.com/rakyll/gotest@latest
	@go install github.com/axw/gocov/gocov@latest
	@go install github.com/AlekSi/gocov-xml@latest
	@go install github.com/matm/gocov-html/cmd/gocov-html@latest
	@go install github.com/jstemmer/go-junit-report/v2@latest
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest
	@go install github.com/vektra/mockery/v2@latest
	@echo "$(GREEN)✓ Test tools installed$(NC)"

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

# Test specific agent
test-agent: build-test-tools
	@echo "$(YELLOW)Testing agent: $(AGENT)$(NC)"
	@./bin/tools/test-agent -agent=$(AGENT)

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
	@echo "$(YELLOW)Deploying test harness to Kubernetes...$(NC)"
	@sed 's|$$(REGISTRY)|$(REGISTRY)|g; s|$$(IMAGE_TAG)|$(IMAGE_TAG)|g; s|$$(TEST_SUITE)|all|g; s|$$(TEST_COMMAND)|test|g' \
		k8s/jobs/test-harness-job.yaml | kubectl apply -f -

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
		--env="KAFKA_BROKERS=kafka-cluster-kafka-bootstrap:9092" \
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
.PHONY: test-spawning
test-spawning: ## Run agent spawning tests in Kubernetes
	@echo "$(YELLOW)Active workflows:$(NC)"
	@echo "$(YELLOW)Running agent spawning tests...$(NC)"
	@echo "$(GREEN)Building test image...$(NC)"
	@docker build -f build/docker/test/test-spawning.dockerfile -t $(REGISTRY)/test-spawning:$(IMAGE_TAG) .
	@docker push $(REGISTRY)/test-spawning:$(IMAGE_TAG)
	@echo "$(GREEN)Running test job...$(NC)"
	@JOB_NAME=test-agent-spawning-$$(date +%s) && \
	echo "apiVersion: batch/v1" > /tmp/test-job.yaml && \
	echo "kind: Job" >> /tmp/test-job.yaml && \
	echo "metadata:" >> /tmp/test-job.yaml && \
	echo "  name: $$JOB_NAME" >> /tmp/test-job.yaml && \
	echo "  namespace: $(PROJECT_NAME)" >> /tmp/test-job.yaml && \
	echo "spec:" >> /tmp/test-job.yaml && \
	echo "  template:" >> /tmp/test-job.yaml && \
	echo "    spec:" >> /tmp/test-job.yaml && \
	echo "      restartPolicy: Never" >> /tmp/test-job.yaml && \
	echo "      imagePullSecrets:" >> /tmp/test-job.yaml && \
	echo "      - name: docker-hub-creds" >> /tmp/test-job.yaml && \
	echo "      containers:" >> /tmp/test-job.yaml && \
	echo "      - name: test-runner" >> /tmp/test-job.yaml && \
	echo "        image: $(REGISTRY)/test-spawning:$(IMAGE_TAG)" >> /tmp/test-job.yaml && \
	echo "        env:" >> /tmp/test-job.yaml && \
	echo "        - name: DB_HOST" >> /tmp/test-job.yaml && \
	echo "          value: postgres-clients" >> /tmp/test-job.yaml && \
	echo "        - name: DB_NAME" >> /tmp/test-job.yaml && \
	echo "          value: clients_db" >> /tmp/test-job.yaml && \
	echo "        - name: DB_USER" >> /tmp/test-job.yaml && \
	echo "          value: clients_user" >> /tmp/test-job.yaml && \
	echo "        - name: DB_PASSWORD" >> /tmp/test-job.yaml && \
	echo "          valueFrom:" >> /tmp/test-job.yaml && \
	echo "            secretKeyRef:" >> /tmp/test-job.yaml && \
	echo "              name: personae-dev-secrets" >> /tmp/test-job.yaml && \
	echo "              key: CLIENTS_DB_PASSWORD" >> /tmp/test-job.yaml && \
	echo "        - name: KAFKA_BROKERS" >> /tmp/test-job.yaml && \
	echo "          value: kafka-cluster-kafka-bootstrap:9092" >> /tmp/test-job.yaml && \
	echo "        - name: CLIENT_ID" >> /tmp/test-job.yaml && \
	echo "          value: demo_client" >> /tmp/test-job.yaml && \
	kubectl apply -f /tmp/test-job.yaml && \
	echo "$(GREEN)Test job $$JOB_NAME created. Waiting for completion...$(NC)" && \
	kubectl wait --for=condition=complete --timeout=300s job/$$JOB_NAME -n $(PROJECT_NAME) || \
		(kubectl logs job/$$JOB_NAME -n $(PROJECT_NAME) && exit 1) && \
	echo "$(GREEN)Tests completed successfully!$(NC)"


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

