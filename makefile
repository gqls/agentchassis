# Comprehensive Makefile for Agent-Managed Microservices

```makefile
# Project variables
PROJECT_NAME := personae-system
ENVIRONMENT ?= production
REGION ?= uk001
REGISTRY ?= registry.personae.io
IMAGE_TAG ?= latest

# Paths
TERRAFORM_DIR := deployments/terraform/environments/$(ENVIRONMENT)/$(REGION)
KUSTOMIZE_DIR := deployments/kustomize
SCRIPTS_DIR := scripts

# Colors for output
YELLOW := \033[1;33m
GREEN := \033[1;32m
RED := \033[1;31m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

#################################
# Help
#################################
.PHONY: help
help: ## Show this help message
	@echo '$(YELLOW)Personae System - Makefile Commands$(NC)'
	@echo ''
	@echo 'Usage:'
	@echo '  make $(GREEN)<target>$(NC) $(YELLOW)[ENVIRONMENT=production] [REGION=uk001] [IMAGE_TAG=latest]$(NC)'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-30s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

#################################
# Development Environment
#################################
.PHONY: dev-up
dev-up: ## Start local development environment
	@echo "$(YELLOW)Starting local development environment...$(NC)"
	docker-compose -f deployments/docker-compose/docker-compose.yaml up -d

.PHONY: dev-down
dev-down: ## Stop local development environment
	@echo "$(YELLOW)Stopping local development environment...$(NC)"
	docker-compose -f deployments/docker-compose/docker-compose.yaml down

.PHONY: dev-logs
dev-logs: ## Show logs from development environment
	docker-compose -f deployments/docker-compose/docker-compose.yaml logs -f

.PHONY: dev-reset
dev-reset: dev-down ## Reset development environment (removes volumes)
	@echo "$(YELLOW)Resetting development environment...$(NC)"
	docker-compose -f deployments/docker-compose/docker-compose.yaml down -v

#################################
# Building
#################################
.PHONY: build-all
build-all: build-backend build-frontends ## Build all images

.PHONY: build-backend
build-backend: build-auth-service build-core-manager build-agents build-adapters ## Build all backend services

.PHONY: build-frontends
build-frontends: build-admin-dashboard build-user-portal build-agent-playground ## Build all frontend applications

# Backend services
.PHONY: build-auth-service
build-auth-service: ## Build auth-service image
	@echo "$(YELLOW)Building auth-service...$(NC)"
	docker build -t $(REGISTRY)/auth-service:$(IMAGE_TAG) \
		-f build/docker/backend/auth-service.dockerfile .

.PHONY: build-core-manager
build-core-manager: ## Build core-manager image
	@echo "$(YELLOW)Building core-manager...$(NC)"
	docker build -t $(REGISTRY)/core-manager:$(IMAGE_TAG) \
		-f build/docker/backend/core-manager.dockerfile .

.PHONY: build-agent-chassis
build-agent-chassis: ## Build agent-chassis image
	@echo "$(YELLOW)Building agent-chassis...$(NC)"
	docker build -t $(REGISTRY)/agent-chassis:$(IMAGE_TAG) \
		-f build/docker/backend/agent-chassis.dockerfile .

.PHONY: build-reasoning-agent
build-reasoning-agent: ## Build reasoning-agent image
	@echo "$(YELLOW)Building reasoning-agent...$(NC)"
	docker build -t $(REGISTRY)/reasoning-agent:$(IMAGE_TAG) \
		-f build/docker/backend/reasoning-agent.dockerfile .

.PHONY: build-web-search-adapter
build-web-search-adapter: ## Build web-search-adapter image
	@echo "$(YELLOW)Building web-search-adapter...$(NC)"
	docker build -t $(REGISTRY)/web-search-adapter:$(IMAGE_TAG) \
		-f build/docker/backend/web-search-adapter.dockerfile .

.PHONY: build-image-generator-adapter
build-image-generator-adapter: ## Build image-generator-adapter image
	@echo "$(YELLOW)Building image-generator-adapter...$(NC)"
	docker build -t $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG) \
		-f build/docker/backend/image-generator-adapter.dockerfile .

# Agent targets
.PHONY: build-agents
build-agents: build-agent-chassis build-reasoning-agent ## Build all agents

.PHONY: build-adapters
build-adapters: build-web-search-adapter build-image-generator-adapter ## Build all adapters

# Frontend applications
.PHONY: build-admin-dashboard
build-admin-dashboard: ## Build admin-dashboard image
	@echo "$(YELLOW)Building admin-dashboard...$(NC)"
	cd frontends/admin-dashboard && npm install && npm run build
	docker build -t $(REGISTRY)/admin-dashboard:$(IMAGE_TAG) \
		-f frontends/admin-dashboard/Dockerfile frontends/admin-dashboard

.PHONY: build-user-portal
build-user-portal: ## Build user-portal image
	@echo "$(YELLOW)Building user-portal...$(NC)"
	cd frontends/user-portal && npm install && npm run build
	docker build -t $(REGISTRY)/user-portal:$(IMAGE_TAG) \
		-f frontends/user-portal/Dockerfile frontends/user-portal

.PHONY: build-agent-playground
build-agent-playground: ## Build agent-playground image
	@echo "$(YELLOW)Building agent-playground...$(NC)"
	cd frontends/agent-playground && npm install && npm run build
	docker build -t $(REGISTRY)/agent-playground:$(IMAGE_TAG) \
		-f frontends/agent-playground/Dockerfile frontends/agent-playground

#################################
# Push Images
#################################
.PHONY: push-all
push-all: push-backend push-frontends ## Push all images to registry

.PHONY: push-backend
push-backend: ## Push all backend images
	@echo "$(YELLOW)Pushing backend images...$(NC)"
	docker push $(REGISTRY)/auth-service:$(IMAGE_TAG)
	docker push $(REGISTRY)/core-manager:$(IMAGE_TAG)
	docker push $(REGISTRY)/agent-chassis:$(IMAGE_TAG)
	docker push $(REGISTRY)/reasoning-agent:$(IMAGE_TAG)
	docker push $(REGISTRY)/web-search-adapter:$(IMAGE_TAG)
	docker push $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG)

.PHONY: push-frontends
push-frontends: ## Push all frontend images
	@echo "$(YELLOW)Pushing frontend images...$(NC)"
	docker push $(REGISTRY)/admin-dashboard:$(IMAGE_TAG)
	docker push $(REGISTRY)/user-portal:$(IMAGE_TAG)
	docker push $(REGISTRY)/agent-playground:$(IMAGE_TAG)

#################################
# Infrastructure Deployment
#################################
.PHONY: deploy-infrastructure
deploy-infrastructure: ## Deploy all infrastructure components
	@echo "$(YELLOW)Deploying infrastructure to $(ENVIRONMENT)/$(REGION)...$(NC)"
	@$(MAKE) deploy-010-infrastructure
	@$(MAKE) deploy-020-ingress
	@$(MAKE) deploy-030-strimzi
	@$(MAKE) deploy-040-kafka
	@$(MAKE) deploy-050-storage
	@$(MAKE) deploy-060-databases
	@$(MAKE) deploy-070-schemas
	@$(MAKE) deploy-080-topics
	@$(MAKE) deploy-090-monitoring

# Individual infrastructure components
.PHONY: deploy-010-infrastructure
deploy-010-infrastructure: ## Deploy core infrastructure (Kubernetes cluster)
	@echo "$(GREEN)Deploying 010-infrastructure...$(NC)"
	cd $(TERRAFORM_DIR)/010-infrastructure && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-020-ingress
deploy-020-ingress: ## Deploy ingress controller
	@echo "$(GREEN)Deploying 020-ingress-nginx...$(NC)"
	cd $(TERRAFORM_DIR)/020-ingress-nginx && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-030-strimzi
deploy-030-strimzi: ## Deploy Strimzi operator
	@echo "$(GREEN)Deploying 030-strimzi-operator...$(NC)"
	cd $(TERRAFORM_DIR)/030-strimzi-operator && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-040-kafka
deploy-040-kafka: ## Deploy Kafka cluster
	@echo "$(GREEN)Deploying 040-kafka-cluster...$(NC)"
	cd $(TERRAFORM_DIR)/040-kafka-cluster && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-050-storage
deploy-050-storage: ## Deploy S3/storage buckets
	@echo "$(GREEN)Deploying 050-storage...$(NC)"
	cd $(TERRAFORM_DIR)/050-storage && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-060-databases
deploy-060-databases: ## Deploy database instances
	@echo "$(GREEN)Deploying 060-databases...$(NC)"
	cd $(TERRAFORM_DIR)/060-databases && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-070-schemas
deploy-070-schemas: ## Run database migrations
	@echo "$(GREEN)Deploying 070-database-schemas...$(NC)"
	cd $(TERRAFORM_DIR)/070-database-schemas && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-080-topics
deploy-080-topics: ## Create Kafka topics
	@echo "$(GREEN)Deploying 080-kafka-topics...$(NC)"
	cd $(TERRAFORM_DIR)/080-kafka-topics && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: deploy-090-monitoring
deploy-090-monitoring: ## Deploy monitoring stack
	@echo "$(GREEN)Deploying 090-monitoring...$(NC)"
	cd $(TERRAFORM_DIR)/090-monitoring && \
		terraform init && \
		terraform apply -auto-approve

#################################
# Application Deployment
#################################
.PHONY: deploy-all
deploy-all: deploy-infrastructure deploy-core deploy-agents deploy-frontends ## Deploy everything

.PHONY: deploy-core
deploy-core: ## Deploy core platform services
	@echo "$(YELLOW)Deploying core platform services...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/services/auth-service/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/services/core-manager/overlays/$(ENVIRONMENT)

.PHONY: deploy-agents
deploy-agents: ## Deploy all agent services
	@echo "$(YELLOW)Deploying agent services...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/services/reasoning-agent/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/services/web-search-adapter/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/services/image-generator-adapter/overlays/$(ENVIRONMENT)

.PHONY: deploy-frontends
deploy-frontends: ## Deploy all frontend applications
	@echo "$(YELLOW)Deploying frontend applications...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/admin-dashboard/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/user-portal/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/agent-playground/overlays/$(ENVIRONMENT)

# Individual service deployments
.PHONY: deploy-auth-service
deploy-auth-service: ## Deploy auth-service only
	@echo "$(GREEN)Deploying auth-service...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/services/auth-service/overlays/$(ENVIRONMENT)

.PHONY: deploy-core-manager
deploy-core-manager: ## Deploy core-manager only
	@echo "$(GREEN)Deploying core-manager...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/services/core-manager/overlays/$(ENVIRONMENT)

.PHONY: deploy-admin-dashboard
deploy-admin-dashboard: ## Deploy admin-dashboard only
	@echo "$(GREEN)Deploying admin-dashboard...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/admin-dashboard/overlays/$(ENVIRONMENT)

.PHONY: deploy-user-portal
deploy-user-portal: ## Deploy user-portal only
	@echo "$(GREEN)Deploying user-portal...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/user-portal/overlays/$(ENVIRONMENT)

#################################
# Full Stack Operations
#################################
.PHONY: full-deploy
full-deploy: build-all push-all deploy-all ## Build, push, and deploy everything

.PHONY: quick-deploy
quick-deploy: ## Deploy applications without building (uses existing images)
	@echo "$(YELLOW)Quick deployment using existing images...$(NC)"
	@$(MAKE) deploy-core
	@$(MAKE) deploy-agents
	@$(MAKE) deploy-frontends

#################################
# Status and Monitoring
#################################
.PHONY: status
status: ## Show status of all deployments
	@echo "$(YELLOW)Deployment Status:$(NC)"
	kubectl get deployments -n $(PROJECT_NAME)
	@echo "\n$(YELLOW)Services:$(NC)"
	kubectl get services -n $(PROJECT_NAME)
	@echo "\n$(YELLOW)Pods:$(NC)"
	kubectl get pods -n $(PROJECT_NAME)

.PHONY: logs
logs: ## Tail logs from all pods
	kubectl logs -f -n $(PROJECT_NAME) -l app.kubernetes.io/part-of=$(PROJECT_NAME) --all-containers=true

.PHONY: logs-auth
logs-auth: ## Tail logs from auth-service
	kubectl logs -f -n $(PROJECT_NAME) -l app=auth-service --all-containers=true

.PHONY: logs-core
logs-core: ## Tail logs from core-manager
	kubectl logs -f -n $(PROJECT_NAME) -l app=core-manager --all-containers=true

#################################
# Rollback Operations
#################################
.PHONY: rollback-auth-service
rollback-auth-service: ## Rollback auth-service deployment
	kubectl rollout undo deployment/auth-service -n $(PROJECT_NAME)

.PHONY: rollback-core-manager
rollback-core-manager: ## Rollback core-manager deployment
	kubectl rollout undo deployment/core-manager -n $(PROJECT_NAME)

#################################
# Testing
#################################
.PHONY: test
test: test-unit test-integration ## Run all tests

.PHONY: test-unit
test-unit: ## Run unit tests
	@echo "$(YELLOW)Running unit tests...$(NC)"
	go test ./... -v -short

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	go test ./tests/integration/... -v

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(YELLOW)Running E2E tests...$(NC)"
	go test ./tests/e2e/... -v

#################################
# Database Operations
#################################
.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(YELLOW)Running database migrations...$(NC)"
	$(SCRIPTS_DIR)/migration/run-migrations.sh

.PHONY: db-seed
db-seed: ## Seed database with test data
	@echo "$(YELLOW)Seeding database...$(NC)"
	kubectl exec -it deployment/postgres-clients -n $(PROJECT_NAME) -- \
		psql -U postgres -f /scripts/seed-data.sql

#################################
# Utility Commands
#################################
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf dist/
	rm -rf frontends/*/build/
	rm -rf frontends/*/dist/

.PHONY: setup-registry
setup-registry: ## Set up local Docker registry
	@echo "$(YELLOW)Setting up local Docker registry...$(NC)"
	$(SCRIPTS_DIR)/utils/setup-local-registry.sh

.PHONY: generate-secrets
generate-secrets: ## Generate required secrets
	@echo "$(YELLOW)Generating secrets...$(NC)"
	$(SCRIPTS_DIR)/utils/generate-jwt-secret.sh

.PHONY: port-forward-admin
port-forward-admin: ## Port forward admin dashboard to localhost:3000
	kubectl port-forward -n $(PROJECT_NAME) svc/admin-dashboard 3000:80

.PHONY: port-forward-grafana
port-forward-grafana: ## Port forward Grafana to localhost:3001
	kubectl port-forward -n $(PROJECT_NAME) svc/grafana 3001:3000

#################################
# Individual Service Builds & Deploys
#################################
# Convenience targets for individual service development
.PHONY: auth-service
auth-service: build-auth-service push-auth-service deploy-auth-service ## Build, push and deploy auth-service

.PHONY: core-manager
core-manager: build-core-manager push-core-manager deploy-core-manager ## Build, push and deploy core-manager

.PHONY: admin-dashboard
admin-dashboard: build-admin-dashboard push-admin-dashboard deploy-admin-dashboard ## Build, push and deploy admin-dashboard

# Push individual services
.PHONY: push-auth-service
push-auth-service: ## Push auth-service image
	docker push $(REGISTRY)/auth-service:$(IMAGE_TAG)

.PHONY: push-core-manager
push-core-manager: ## Push core-manager image
	docker push $(REGISTRY)/core-manager:$(IMAGE_TAG)

.PHONY: push-admin-dashboard
push-admin-dashboard: ## Push admin-dashboard image
	docker push $(REGISTRY)/admin-dashboard:$(IMAGE_TAG)

#################################
# Terraform Operations
#################################
.PHONY: tf-plan
tf-plan: ## Run terraform plan for all infrastructure
	@echo "$(YELLOW)Running Terraform plan...$(NC)"
	@for dir in $(TERRAFORM_DIR)/0*; do \
		echo "$(GREEN)Planning $$dir...$(NC)"; \
		cd $$dir && terraform plan; \
	done

.PHONY: tf-destroy-apps
tf-destroy-apps: ## Destroy all applications (keeps infrastructure)
	@echo "$(RED)Destroying all applications...$(NC)"
	kubectl delete -k $(KUSTOMIZE_DIR)/services --recursive
	kubectl delete -k $(KUSTOMIZE_DIR)/frontends --recursive

.PHONY: tf-destroy-all
tf-destroy-all: ## Destroy everything (WARNING: This will delete everything!)
	@echo "$(RED)WARNING: This will destroy all infrastructure and data!$(NC)"
	@echo "Press Ctrl+C within 5 seconds to cancel..."
	@sleep 5
	@for dir in $$(ls -r $(TERRAFORM_DIR)/); do \
		echo "$(RED)Destroying $$dir...$(NC)"; \
		cd $(TERRAFORM_DIR)/$$dir && terraform destroy -auto-approve; \
	done


#################################
# Swagger
#################################

# Install swagger tools
.PHONY: install-swagger
install-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation
.PHONY: swagger
swagger:
	swag init -g cmd/auth-service/main.go -o cmd/auth-service/docs --parseDependency --parseInternal

# Generate swagger for all services
.PHONY: swagger-all
swagger-all: swagger
	@echo "Swagger documentation generated for auth-service"

# Validate OpenAPI spec
.PHONY: validate-openapi
validate-openapi:
	docker run --rm -v ${PWD}:/spec redocly/cli lint /spec/internal/auth-service/api/openapi.yaml

# Generate API documentation (HTML)
.PHONY: generate-api-docs
generate-api-docs:
	docker run --rm -v ${PWD}:/spec redocly/cli build-docs /spec/internal/auth-service/api/openapi.yaml -o /spec/docs/api-reference.html

# Clean swagger generated files
.PHONY: clean-swagger
clean-swagger:
	rm -rf cmd/auth-service/docs