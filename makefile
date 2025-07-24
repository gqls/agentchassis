# Comprehensive Makefile for Agent-Managed Microservices

export TMPDIR := $(HOME)/kind-tmp

# Project variables
PROJECT_NAME := ai-persona-system
ENVIRONMENT ?= production
REGION ?= uk001
REGISTRY ?= docker.io/aqls
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
	@$(MAKE) deploy-045-kafka-users
	@$(MAKE) deploy-047-base-configs
	@$(MAKE) deploy-050-storage
	@$(MAKE) deploy-060-databases
	@$(MAKE) deploy-070-schemas
	@$(MAKE) deploy-080-topics
	@$(MAKE) deploy-090-monitoring


# Individual infrastructure components
.PHONY: deploy-010-infrastructure
deploy-010-infrastructure: ## Deploy core infrastructure (Kubernetes cluster)
	@echo "$(GREEN)Deploying 010-infrastructure...$(NC)"
	@cd $(TERRAFORM_DIR)/010-infrastructure && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-020-ingress
deploy-020-ingress: ## Deploy ingress controller
	@echo "$(GREEN)Deploying 020-ingress-nginx...$(NC)"
	@cd $(TERRAFORM_DIR)/020-ingress-nginx && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-030-strimzi
deploy-030-strimzi: ## Deploy Strimzi operator
	@echo "$(GREEN)Deploying 030-strimzi-operator...$(NC)"
	@cd $(TERRAFORM_DIR)/030-strimzi-operator && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-040-kafka
deploy-040-kafka: ## Deploy Kafka cluster
	@echo "$(GREEN)Deploying 040-kafka-cluster...$(NC)"
	@cd $(TERRAFORM_DIR)/040-kafka-cluster && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-045-kafka-users
deploy-045-kafka-users: deploy-040-kafka ## Fixed dependency name
	@echo "$(GREEN)Deploying 045-kafka-users...$(NC)"
	cd $(TERRAFORM_DIR)/045-kafka-users && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: destroy-045-kafka-users
destroy-045-kafka-users: ## Destroy Kafka users
	@echo "$(RED)Destroying 045-kafka-users...$(NC)"
	cd $(TERRAFORM_DIR)/045-kafka-users && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform destroy -auto-approve; \
		fi

.PHONY: deploy-047-base-configs
deploy-047-base-configs: ## Deploy base ConfigMaps and Secrets
	@echo "$(GREEN)Deploying 047-base-configs...$(NC)"
	@cd $(TERRAFORM_DIR)/047-base-configs && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi


.PHONY: deploy-050-storage
deploy-050-storage: ## Deploy S3/storage buckets
	@echo "$(GREEN)Deploying 050-storage...$(NC)"
	@cd $(TERRAFORM_DIR)/050-storage && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-060-databases
deploy-060-databases: ## Deploy database instances
	@echo "$(GREEN)Deploying 060-databases...$(NC)"
	@cd $(TERRAFORM_DIR)/060-databases && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-070-schemas
deploy-070-schemas: ## Run database migrations
	@echo "$(GREEN)Deploying 070-database-schemas...$(NC)"
	@cd $(TERRAFORM_DIR)/070-database-schemas && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-080-topics
deploy-080-topics: ## Create Kafka topics
	@echo "$(GREEN)Deploying 080-kafka-topics...$(NC)"
	@cd $(TERRAFORM_DIR)/080-kafka-topics && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

.PHONY: deploy-090-monitoring
deploy-090-monitoring: ## Deploy monitoring stack
	@echo "$(GREEN)Deploying 090-monitoring...$(NC)"
	@cd $(TERRAFORM_DIR)/090-monitoring && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init && \
			terraform apply -auto-approve; \
		fi

#################################
# Application Deployment (Terraform Workflow)
#################################
# Generic target for deploying any service via Terraform
.PHONY: deploy-all
deploy-all: deploy-infrastructure deploy-core deploy-agents ## deploy-frontends ## Deploy everything

.PHONY: deploy-service
deploy-service:
	@echo "$(GREEN)Deploying service at $(path)...$(NC)"
	@cd $(path) && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init -upgrade && \
			terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init -upgrade && \
			terraform apply -auto-approve; \
		fi

# Generic target for destroying any service via Terraform
.PHONY: destroy-service
destroy-service:
	@echo "$(RED)Destroying service at $(path)...$(NC)"
	@cd $(path) && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform init -upgrade && \
			terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform init -upgrade && \
			terraform destroy -auto-approve; \
		fi

# Core Platform Services
.PHONY: deploy-core
deploy-core: deploy-047-base-configs deploy-auth-service deploy-core-manager ## Deploy core platform services using Terraform

.PHONY: deploy-auth-service
deploy-auth-service: ## Deploy auth-service using Terraform
	@$(MAKE) deploy-service path=$(TERRAFORM_DIR)/services/core-platform/1110-auth-service

.PHONY: deploy-core-manager
deploy-core-manager: ## Deploy core-manager using Terraform
	@$(MAKE) deploy-service path=$(TERRAFORM_DIR)/services/core-platform/1120-core-manager

# Corresponding destroy targets
.PHONY: destroy-core
destroy-core: destroy-core-manager destroy-auth-service ## Destroy core platform services using Terraform

.PHONY: destroy-auth-service
destroy-auth-service: ## Destroy auth-service using Terraform
	@$(MAKE) destroy-service path=$(TERRAFORM_DIR)/services/core-platform/1110-auth-service

.PHONY: destroy-core-manager
destroy-core-manager: ## Destroy core-manager using Terraform
	@$(MAKE) destroy-service path=$(TERRAFORM_DIR)/services/core-platform/1120-core-manager


.PHONY: deploy-agents
deploy-agents: ## Deploy all agent services
	@echo "$(YELLOW)Deploying agent services...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/services/reasoning-agent/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/services/web-search-adapter/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/services/image-generator-adapter/overlays/$(ENVIRONMENT)


.PHONY: redeploy-agents
redeploy-agents: ## Forces a rolling restart of all agent deployments
	@echo "$(YELLOW)Forcing rollout restart of agent deployments...$(NC)"
	kubectl rollout restart deployment agent-chassis -n ai-persona-system
	kubectl rollout restart deployment reasoning-agent -n ai-persona-system
	kubectl rollout restart deployment web-search-adapter -n ai-persona-system
	kubectl rollout restart deployment image-generator-adapter -n ai-persona-system


.PHONY: deploy-frontends
deploy-frontends: ## Deploy all frontend applications
	@echo "$(YELLOW)Deploying frontend applications...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/admin-dashboard/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/user-portal/overlays/$(ENVIRONMENT)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/agent-playground/overlays/$(ENVIRONMENT)

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
	@for dir in $(TERRAFORM_DIR)/0*; do \  # This pattern already includes 045-kafka-users
		echo "$(GREEN)Planning $$dir...$(NC)"; \
		cd $$dir && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform plan -var-file=terraform.tfvars.secret; \
		else \
			terraform plan; \
		fi; \
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
		cd $(TERRAFORM_DIR)/$$dir && \
		if [ -f terraform.tfvars.secret ]; then \
			terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			terraform destroy -auto-approve; \
		fi; \
	done

#################################
# Swagger/Documentation
#################################

# Install swagger tools
.PHONY: install-swagger
install-swagger: ## Install swagger generation tools
	@echo "$(YELLOW)Installing swagger tools...$(NC)"
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger documentation for auth-service
.PHONY: swagger-auth
swagger-auth: ## Generate swagger documentation for auth-service
	@echo "$(YELLOW)Generating swagger documentation for auth-service...$(NC)"
	@cd cmd/auth-service && swag init -g main.go -o docs --parseDependency --parseInternal --parseDepth 2
	@echo "$(GREEN)Auth service swagger documentation generated$(NC)"

# Generate swagger documentation for core-manager
.PHONY: swagger-core
swagger-core: ## Generate swagger documentation for core-manager
	@echo "$(YELLOW)Generating swagger documentation for core-manager...$(NC)"
	@cd cmd/core-manager && swag init -g main.go -o docs --parseDependency --parseInternal --parseDepth 2
	@echo "$(GREEN)Core manager swagger documentation generated$(NC)"

# Generate swagger for all services
.PHONY: swagger
swagger: swagger-auth swagger-core ## Generate swagger documentation for all services
	@echo "$(GREEN)All swagger documentation generated$(NC)"

# Backwards compatibility alias
.PHONY: swagger-all
swagger-all: swagger ## Alias for swagger target

# Run the comprehensive documentation generation script
.PHONY: docs
docs: swagger ## Generate comprehensive API documentation
	@echo "$(YELLOW)Running comprehensive documentation generation...$(NC)"
	@if [ -f "$(SCRIPTS_DIR)/docs/generate-docs.sh" ]; then \
		$(SCRIPTS_DIR)/docs/generate-docs.sh; \
	else \
		echo "$(YELLOW)Documentation script not found, skipping$(NC)"; \
	fi

# Start swagger UI servers
.PHONY: swagger-ui
swagger-ui: ## Start Swagger UI, Redoc, and Swagger Editor
	@echo "$(YELLOW)Starting documentation servers...$(NC)"
	@if [ -f "deployments/docker-compose/docker-compose.swagger.yml" ]; then \
		docker-compose -f deployments/docker-compose/docker-compose.swagger.yml up -d; \
		echo "$(GREEN)Documentation servers started:$(NC)"; \
		echo "  • Swagger UI: http://localhost:8082"; \
		echo "  • Redoc: http://localhost:8083"; \
		echo "  • Swagger Editor: http://localhost:8084"; \
	else \
		echo "$(YELLOW)Creating swagger docker-compose file...$(NC)"; \
		$(MAKE) create-swagger-compose; \
		docker-compose -f deployments/docker-compose/docker-compose.swagger.yml up -d; \
	fi

# Create swagger docker-compose file if it doesn't exist
.PHONY: create-swagger-compose
create-swagger-compose: ## Create swagger docker-compose file
	@mkdir -p deployments/docker-compose
	@echo "version: '3.8'" > deployments/docker-compose/docker-compose.swagger.yml
	@echo "services:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "  swagger-ui:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    image: swaggerapi/swagger-ui" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    ports:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - \"8082:8080\"" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    environment:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      SWAGGER_JSON: /docs/swagger.json" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    volumes:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - ./../../cmd/auth-service/docs:/docs" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "  redoc:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    image: redocly/redoc" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    ports:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - \"8083:80\"" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    environment:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      SPEC_URL: /docs/swagger.json" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    volumes:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - ./../../cmd/auth-service/docs:/usr/share/nginx/html/docs" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "  swagger-editor:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    image: swaggerapi/swagger-editor" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "    ports:" >> deployments/docker-compose/docker-compose.swagger.yml
	@echo "      - \"8084:8080\"" >> deployments/docker-compose/docker-compose.swagger.yml

# Stop swagger UI servers
.PHONY: swagger-down
swagger-down: ## Stop documentation servers
	@echo "$(YELLOW)Stopping documentation servers...$(NC)"
	@if [ -f "deployments/docker-compose/docker-compose.swagger.yml" ]; then \
		docker-compose -f deployments/docker-compose/docker-compose.swagger.yml down; \
	fi

# Validate swagger specs
.PHONY: validate-swagger
validate-swagger: ## Validate swagger specifications
	@echo "$(YELLOW)Validating swagger specifications...$(NC)"
	@if [ -f "cmd/auth-service/docs/swagger.json" ]; then \
		echo "$(GREEN)Validating auth-service swagger...$(NC)"; \
		docker run --rm -v ${PWD}/cmd/auth-service/docs:/spec redocly/cli lint /spec/swagger.json || true; \
	fi
	@if [ -f "cmd/core-manager/docs/swagger.json" ]; then \
		echo "$(GREEN)Validating core-manager swagger...$(NC)"; \
		docker run --rm -v ${PWD}/cmd/core-manager/docs:/spec redocly/cli lint /spec/swagger.json || true; \
	fi

# Generate API documentation (HTML)
.PHONY: generate-api-docs
generate-api-docs: swagger ## Generate HTML API documentation
	@echo "$(YELLOW)Generating HTML API documentation...$(NC)"
	@mkdir -p docs/api
	@if [ -f "cmd/auth-service/docs/swagger.json" ]; then \
		docker run --rm -v ${PWD}:/app redocly/cli build-docs /app/cmd/auth-service/docs/swagger.json -o /app/docs/api/auth-service.html; \
		echo "$(GREEN)Auth service documentation generated at docs/api/auth-service.html$(NC)"; \
	fi
	@if [ -f "cmd/core-manager/docs/swagger.json" ]; then \
		docker run --rm -v ${PWD}:/app redocly/cli build-docs /app/cmd/core-manager/docs/swagger.json -o /app/docs/api/core-manager.html; \
		echo "$(GREEN)Core manager documentation generated at docs/api/core-manager.html$(NC)"; \
	fi

# Serve API documentation locally
.PHONY: serve-docs
serve-docs: ## Serve API documentation locally on port 8080
	@echo "$(YELLOW)Serving API documentation...$(NC)"
	@if command -v python3 > /dev/null; then \
		cd docs/api && python3 -m http.server 8080; \
	else \
		echo "$(RED)Python3 not found. Please install Python3 to serve docs locally.$(NC)"; \
	fi

# Clean swagger generated files
.PHONY: clean-swagger
clean-swagger: ## Clean swagger generated files
	@echo "$(YELLOW)Cleaning swagger files...$(NC)"
	rm -rf cmd/auth-service/docs
	rm -rf cmd/core-manager/docs
	rm -rf docs/api

# Quick documentation workflow
.PHONY: docs-quick
docs-quick: swagger swagger-ui ## Quick swagger generation and UI startup
	@echo "$(GREEN)Documentation ready at http://localhost:8082$(NC)"

# Generate and view documentation
.PHONY: docs-view
docs-view: generate-api-docs ## Generate and open HTML documentation
	@echo "$(GREEN)Opening documentation...$(NC)"
	@if [ -f "docs/api/auth-service.html" ]; then \
		if command -v xdg-open > /dev/null; then \
			xdg-open docs/api/auth-service.html; \
		elif command -v open > /dev/null; then \
			open docs/api/auth-service.html; \
		else \
			echo "$(YELLOW)Please open docs/api/auth-service.html in your browser$(NC)"; \
		fi \
	fi

#################################
# Kind Cluster Management
#################################
.PHONY: kind-create
kind-create: ## Create Kind cluster for development
	@echo "$(YELLOW)Creating Kind cluster using Terraform...$(NC)"
	cd deployments/terraform/environments/development/uk_dev/010-infrastructure && \
		terraform init && \
		terraform apply -auto-approve

.PHONY: kind-delete
kind-delete: ## Delete Kind cluster
	@echo "$(RED)Deleting Kind cluster...$(NC)"
	cd deployments/terraform/environments/development/uk_dev/010-infrastructure && \
		terraform destroy -auto-approve

.PHONY: kind-status
kind-status: ## Check Kind cluster status
	@echo "$(YELLOW)Kind cluster status:$(NC)"
	kind get clusters
	kubectl config use-context kind-personae-dev && kubectl get nodes

.PHONY: kind-load-images
kind-load-images: ## Load Docker images into Kind
	@echo "$(YELLOW)Loading images into Kind...$(NC)"
	@mkdir -p $(TMPDIR)
	kind load docker-image $(REGISTRY)/auth-service:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/core-manager:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/agent-chassis:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/reasoning-agent:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/web-search-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG) --name personae-dev

.PHONY: reload-auth-service
reload-auth-service: ## Rebuild and reload auth-service in Kind
	@echo "$(YELLOW)Rebuilding auth-service...$(NC)"
	@$(MAKE) build-auth-service
	@mkdir -p $(TMPDIR)
	kind load docker-image $(REGISTRY)/auth-service:$(IMAGE_TAG) --name personae-dev
	kubectl delete pod -n ai-persona-system -l app=auth-service
	@echo "$(GREEN)auth-service reloaded$(NC)"

.PHONY: reload-core-manager
reload-core-manager: ## Rebuild and reload core-manager in Kind
	@echo "$(YELLOW)Rebuilding core-manager...$(NC)"
	@$(MAKE) build-core-manager
	@mkdir -p $(TMPDIR)
	kind load docker-image $(REGISTRY)/core-manager:$(IMAGE_TAG) --name personae-dev
	kubectl delete pod -n ai-persona-system -l app=core-manager
	@echo "$(GREEN)core-manager reloaded$(NC)"

# Add a new helper target
.PHONY: kind-load-auth
kind-load-auth: ## Load auth-service image into Kind
	@mkdir -p $(TMPDIR)
	kind load docker-image auth-service:local --name personae-dev

.PHONY: kind-load-core
kind-load-core: ## Load core-manager image into Kind
	@mkdir -p $(TMPDIR)
	kind load docker-image core-manager:local --name personae-dev

#################################
# Environment Specific Helpers
#################################
.PHONY: use-dev-context
use-dev-context: ## Switch to development Kubernetes context
	kubectl config use-context kind-personae-dev

.PHONY: use-prod-context
use-prod-context: ## Switch to production Kubernetes context
	kubectl config use-context personae-$(REGION)-prod-cluster

#################################
# Secrets Management
#################################
.PHONY: create-dev-secrets
create-dev-secrets: ## Create development secrets
	@echo "$(YELLOW)Creating development secrets...$(NC)"
	kubectl create namespace ai-persona-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl create secret generic personae-dev-secrets \
		--from-literal=clients-db-password=dev-clients-password \
		--from-literal=templates-db-password=dev-templates-password \
		--from-literal=auth-db-password=agent-chassis123! \
		--from-literal=minio-access-key=minio \
		--from-literal=secret-key=minioadmin \
		--from-literal=JWT_SECRET_KEY=dev-secret \
		-n ai-persona-system --dry-run=client -o yaml | kubectl apply -f -

#################################
# ConfigMap Management
#################################
.PHONY: create-dev-configs
create-dev-configs: ## Create development configmaps
	@echo "$(YELLOW)Creating development configmaps...$(NC)"
	kubectl create namespace ai-persona-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployments/kustomize/infrastructure/configs/development/configmap-dev.yaml -n ai-persona-system