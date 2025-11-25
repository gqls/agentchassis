# Comprehensive Makefile for Agent-Managed Microservices

# Load environment variables from .env file
include .env
export

export TMPDIR := $(HOME)/kind-tmp

# Project variables - project name is also used for namespaces
PROJECT_NAME := ai-persona-system
ENVIRONMENT ?= production
REGION ?= uk001
REGION_PATH ?= uk_001
REGISTRY ?= docker.io/aqls
#IMAGE_TAG ?= latest
IMAGE_TAG ?= v1.0.471

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

ifeq ($(ENVIRONMENT),production)
    OVERLAY_PATH := $(ENVIRONMENT)/$(REGION_PATH)
else
    OVERLAY_PATH := $(ENVIRONMENT)
endif

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

.PHONY: build-web-scrape-adapter
build-web-scrape-adapter: ## Build web-scrape-adapter image
	@echo "$(YELLOW)Building web-scrape-adapter...$(NC)"
	docker build -t $(REGISTRY)/web-scrape-adapter:$(IMAGE_TAG) \
		-f build/docker/backend/web-scrape-adapter.dockerfile .

.PHONY: build-git-adapter
build-git-adapter: ## Build git-adapter image
	@echo "$(YELLOW)Building git-adapter...$(NC)"
	docker build -t $(REGISTRY)/git-adapter:$(IMAGE_TAG) \
		-f build/docker/backend/git-adapter.dockerfile .

.PHONY: build-image-generator-adapter
build-image-generator-adapter: ## Build image-generator-adapter image
	@echo "$(YELLOW)Building image-generator-adapter...$(NC)"
	docker build -t $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG) \
		-f build/docker/backend/image-generator-adapter.dockerfile .

.PHONY: build-content-creator-agent
build-content-creator-agent: ## Build content-creator-agent image
	@echo "$(YELLOW)Building content-creator-agent...$(NC)"
	docker build -t $(REGISTRY)/content-creator-agent:$(IMAGE_TAG) \
		-f build/docker/backend/content-creator-agent.dockerfile . # NEW


# Agent targets
.PHONY: build-agents
build-agents: build-agent-chassis build-reasoning-agent build-content-creator-agent ## Build all agents

.PHONY: build-adapters
build-adapters: build-web-search-adapter build-web-scrape-adapter build-git-adapter build-image-generator-adapter ## Build all adapters

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
	docker push $(REGISTRY)/web-scrape-adapter:$(IMAGE_TAG)
	docker push $(REGISTRY)/git-adapter:$(IMAGE_TAG)
	docker push $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG)
	docker push $(REGISTRY)/content-creator-agent:$(IMAGE_TAG)

.PHONY: push-frontends
push-frontends: ## Push all frontend images
	@echo "$(YELLOW)Pushing frontend images...$(NC)"
	docker push $(REGISTRY)/admin-dashboard:$(IMAGE_TAG)
	docker push $(REGISTRY)/user-portal:$(IMAGE_TAG)
	docker push $(REGISTRY)/agent-playground:$(IMAGE_TAG)

#################################
# Infrastructure Deployment
#################################
KUBECONFIG_PATH := $(HOME)/.kube/config_$(ENVIRONMENT)_$(REGION)

.PHONY: deploy-cluster-only
deploy-cluster-only: ## Deploy just the Kubernetes cluster
	@echo "$(GREEN)Deploying Kubernetes cluster...$(NC)"
	@cd $(TERRAFORM_DIR)/010-infrastructure && \
		terraform init && \
		terraform apply -auto-approve -var-file=terraform.tfvars.secret

.PHONY: deploy-infrastructure-old
deploy-infrastructure-old: ## Deploy all infrastructure components
	@echo "$(YELLOW)Deploying infrastructure to $(ENVIRONMENT)/$(REGION)...$(NC)"
	@$(MAKE) deploy-010-infrastructure
	@$(MAKE) deploy-020-ingress
	@$(MAKE) deploy-030-strimzi-operator
	@$(MAKE) deploy-040-kafka-cluster
	@$(MAKE) deploy-045-kafka-users
	@$(MAKE) deploy-047-base-configs
	@$(MAKE) deploy-050-storage
	@$(MAKE) deploy-060-databases
	@$(MAKE) deploy-070-database-schemas
	@$(MAKE) deploy-080-kafka-topics
	@$(MAKE) deploy-090-monitoring

.PHONY: deploy-infrastructure
deploy-infrastructure: ## Deploy all infrastructure components
	@echo "$(YELLOW)Deploying infrastructure to $(ENVIRONMENT)/$(REGION)...$(NC)"
	@echo "$(GREEN)Step 1: Deploying Kubernetes cluster...$(NC)"
	@cd $(TERRAFORM_DIR)/010-infrastructure && \
		terraform init && \
		terraform apply -auto-approve -var-file=terraform.tfvars.secret && \
		terraform output -raw kubeconfig_raw > $(KUBECONFIG_PATH)
	@echo "$(GREEN)Cluster deployed! Using kubeconfig: $(KUBECONFIG_PATH)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-020-ingress
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-030-strimzi-operator
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-040-kafka-cluster
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-045-kafka-users
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-047-base-configs
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-050-storage
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-060-databases
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-070-database-schemas
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-080-kafka-topics
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-090-monitoring
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-100-bootstrap-agents
	@echo "$(GREEN)Infrastructure deployment complete!$(NC)"
	@echo "$(YELLOW)To use this cluster, run: export KUBECONFIG=$(KUBECONFIG_PATH)$(NC)"

# Add this new target that skips cluster creation
.PHONY: deploy-infrastructure-from-ingress
deploy-infrastructure-from-ingress: ## Deploy infrastructure starting from ingress (assumes cluster exists)
	@echo "$(YELLOW)Deploying infrastructure from ingress for $(ENVIRONMENT)/$(REGION)...$(NC)"
	@echo "$(GREEN)Using existing kubeconfig: $(KUBECONFIG_PATH)$(NC)"
	@if [ ! -f "$(KUBECONFIG_PATH)" ]; then \
		echo "$(RED)Error: Kubeconfig not found at $(KUBECONFIG_PATH)$(NC)"; \
		echo "$(YELLOW)Manually set up Kubeconfig first - export KUBECONFIG=~/.kube/config_$(ENVIRONMENT)_$(REGION)    $(NC)"; \
		exit 1; \
	fi
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-020-ingress
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-030-strimzi-operator
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-040-kafka-cluster
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-045-kafka-users
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-047-base-configs
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-050-storage
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-060-databases
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-070-database-schemas
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-080-kafka-topics
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-090-monitoring
	@KUBECONFIG=$(KUBECONFIG_PATH) $(MAKE) deploy-100-bootstrap-agents
	@echo "$(GREEN)Infrastructure deployment complete!$(NC)"
	@echo "$(YELLOW)To use this cluster, run: export KUBECONFIG=$(KUBECONFIG_PATH)$(NC)"

# Quick helper for your current situation
.PHONY: continue-deployment
continue-deployment: deploy-infrastructure-from-ingress ## Continue deployment from where cluster creation finished


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

# Export KUBECONFIG for all terraform commands in this section
.PHONY: deploy-020-ingress
deploy-020-ingress: ## Deploy ingress controller
	@echo "$(GREEN)Deploying 020-ingress-nginx...$(NC)"
	@cd $(TERRAFORM_DIR)/020-ingress-nginx && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-030-strimzi-operator
deploy-030-strimzi-operator: ## Deploy Strimzi operator
	@echo "$(GREEN)Deploying 030-strimzi-operator...$(NC)"
	@cd $(TERRAFORM_DIR)/030-strimzi-operator && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-040-kafka-cluster
deploy-040-kafka-cluster: ## Deploy Kafka cluster
	@echo "$(GREEN)Deploying 040-kafka-cluster...$(NC)"
	@cd $(TERRAFORM_DIR)/040-kafka-cluster && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-045-kafka-users
deploy-045-kafka-users: deploy-040-kafka-cluster ## Fixed dependency name
	@echo "$(GREEN)Deploying 045-kafka-users...$(NC)"
	cd $(TERRAFORM_DIR)/045-kafka-users && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-047-base-configs
deploy-047-base-configs: ## Deploy base ConfigMaps and Secrets
	@echo "$(GREEN)Deploying 047-base-configs...$(NC)"
	@cd $(TERRAFORM_DIR)/047-base-configs && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-050-storage
deploy-050-storage: ## Deploy S3/storage buckets
	@echo "$(GREEN)Deploying 050-storage...$(NC)"
	@cd $(TERRAFORM_DIR)/050-storage && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-060-databases
deploy-060-databases: ## Deploy database instances
	@echo "$(GREEN)Deploying 060-databases...$(NC)"
	@cd $(TERRAFORM_DIR)/060-databases && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-070-database-schemas
deploy-070-database-schemas: ## Run database migrations
	@echo "$(GREEN)Deploying 070-database-schemas...$(NC)"
	@cd $(TERRAFORM_DIR)/070-database-schemas && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-080-kafka-topics
deploy-080-kafka-topics: ## Create Kafka topics
	@echo "$(GREEN)Deploying 080-kafka-topics...$(NC)"
	@cd $(TERRAFORM_DIR)/080-kafka-topics && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

.PHONY: deploy-090-monitoring
deploy-090-monitoring: ## Deploy monitoring stack
	@echo "$(GREEN)Deploying 090-monitoring...$(NC)"
	@cd $(TERRAFORM_DIR)/090-monitoring && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi


.PHONY: deploy-100-bootstrap-agents
deploy-100-bootstrap-agents: ## Deploy bootstrap agents (generic orchestrator) with image updates
	@echo "$(GREEN)Deploying 100-bootstrap-agents...$(NC)"
	@echo "$(YELLOW)First updating agent definitions with current image...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -i postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"UPDATE agent_definitions SET image_repository = '$(REGISTRY)/agent-chassis', image_tag = '$(IMAGE_TAG)', updated_at = NOW(); SELECT COUNT(*) as updated_count FROM agent_definitions;" 2>/dev/null || true
	@cd $(TERRAFORM_DIR)/100-bootstrap-agents && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve \
				-var-file=terraform.tfvars.secret \
				-var="image_tag=$(IMAGE_TAG)" \
				-var="registry=$(REGISTRY)"; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve \
				-var="image_tag=$(IMAGE_TAG)" \
				-var="registry=$(REGISTRY)"; \
		fi
	@echo "$(GREEN)Bootstrap agents deployed with image $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"


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
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform apply -auto-approve; \
		fi

# Generic target for destroying any service via Terraform
.PHONY: destroy-service
destroy-service:
	@echo "$(RED)Destroying service at $(path)...$(NC)"
	@cd $(path) && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform init -upgrade && \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve; \
		fi

# Core Platform Services
.PHONY: deploy-core
deploy-core: update-kustomization-images deploy-047-base-configs deploy-auth-service deploy-core-manager ## Deploy core platform services using Terraform

.PHONY: deploy-auth-service
deploy-auth-service:  ## Deploy auth-service using Terraform
	# Update the image tag in kustomization.yaml FIRST
	@echo "$(YELLOW)Updating auth-service image tag to $(IMAGE_TAG)...$(NC)"
	@cd $(KUSTOMIZE_DIR)/services/auth-service/overlays/$(OVERLAY_PATH) && \
		sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' kustomization.yaml

	@$(MAKE) deploy-service path=$(TERRAFORM_DIR)/services/core-platform/1110-auth-service

.PHONY: deploy-core-manager
deploy-core-manager:  ## Deploy core-manager using Terraform
# Update the image tag in kustomization.yaml FIRST
	@echo "$(YELLOW)Updating core-manager image tag to $(IMAGE_TAG)...$(NC)"
	@cd $(KUSTOMIZE_DIR)/services/core-manager/overlays/$(OVERLAY_PATH) && \
		sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' kustomization.yaml


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


# Update all agent images
.PHONY: update-kustomization-images
update-kustomization-images: ## Update image tags in kustomization.yaml files
	@echo "$(YELLOW)Updating kustomization.yaml files with image tag $(IMAGE_TAG)...$(NC)"
	@for agent in agent-chassis reasoning-agent web-search-adapter web-scrape-adapter git-adapter image-generator-adapter content-creator-agent; do \
		kust_file="$(KUSTOMIZE_DIR)/services/$$agent/overlays/$(OVERLAY_PATH)/kustomization.yaml"; \
		if [ -f "$$kust_file" ]; then \
			echo "Updating $$agent kustomization.yaml..."; \
			if grep -q "images:" "$$kust_file"; then \
				sed -i.bak '/images:/,/^[^ ]/{/newTag:/s/newTag:.*/newTag: $(IMAGE_TAG)/}' "$$kust_file"; \
			else \
				echo "" >> "$$kust_file"; \
				echo "images:" >> "$$kust_file"; \
				echo "  - name: docker.io/aqls/$$agent" >> "$$kust_file"; \
				echo "    newTag: $(IMAGE_TAG)" >> "$$kust_file"; \
			fi; \
		fi; \
	done

# Deploy agents with automatic image update
# Update ConfigMap with new image tag
.PHONY: update-agent-image-tag
update-agent-image-tag: ## Update the agent image tag in ConfigMap
	@echo "$(YELLOW)Updating agent image tag to $(IMAGE_TAG)...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl patch configmap personae-prod-config \
		-n ai-persona-system \
		--type merge \
		-p '{"data":{"AGENT_IMAGE_TAG":"$(IMAGE_TAG)","agent_image_tag":"$(IMAGE_TAG)"}}'

# Deploy agents with automatic image update
.PHONY: deploy-agents
deploy-agents: ## Deploy all agent services with dynamic image tag
	@echo "$(YELLOW)Deploying agent services with image tag $(IMAGE_TAG)...$(NC)"

	# Update agent-chassis kustomization.yaml
	@echo "Updating agent-chassis to $(IMAGE_TAG)..."
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)/kustomization.yaml
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)

	# Update reasoning-agent kustomization.yaml
	@echo "Updating reasoning-agent to $(IMAGE_TAG)..."
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/reasoning-agent/overlays/$(OVERLAY_PATH)/kustomization.yaml 2>/dev/null || true
	@if [ -d "$(KUSTOMIZE_DIR)/services/reasoning-agent/overlays/$(OVERLAY_PATH)" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/reasoning-agent/overlays/$(OVERLAY_PATH); \
	fi

	# Update web-search-adapter kustomization.yaml
	@echo "Updating web-search-adapter to $(IMAGE_TAG)..."
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/web-search-adapter/overlays/$(OVERLAY_PATH)/kustomization.yaml 2>/dev/null || true
	@if [ -d "$(KUSTOMIZE_DIR)/services/web-search-adapter/overlays/$(OVERLAY_PATH)" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/web-search-adapter/overlays/$(OVERLAY_PATH); \
	fi

	# Update web-scrape-adapter kustomization.yaml
	@echo "Updating web-scrape-adapter to $(IMAGE_TAG)..."
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/web-scrape-adapter/overlays/$(OVERLAY_PATH)/kustomization.yaml 2>/dev/null || true
	@if [ -d "$(KUSTOMIZE_DIR)/services/web-scrape-adapter/overlays/$(OVERLAY_PATH)" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/web-scrape-adapter/overlays/$(OVERLAY_PATH); \
	fi

	# Update git-adapter kustomization.yaml
	@echo "Updating git-adapter to $(IMAGE_TAG)..."
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/git-adapter/overlays/$(OVERLAY_PATH)/kustomization.yaml 2>/dev/null || true
	@if [ -d "$(KUSTOMIZE_DIR)/services/git-adapter/overlays/$(OVERLAY_PATH)" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/git-adapter/overlays/$(OVERLAY_PATH); \
	fi

	# Update image-generator-adapter kustomization.yaml
	@echo "Updating image-generator-adapter to $(IMAGE_TAG)..."
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/image-generator-adapter/overlays/$(OVERLAY_PATH)/kustomization.yaml 2>/dev/null || true
	@if [ -d "$(KUSTOMIZE_DIR)/services/image-generator-adapter/overlays/$(OVERLAY_PATH)" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/image-generator-adapter/overlays/$(OVERLAY_PATH); \
	fi

	# Update content-creator-agent kustomization.yaml
	@echo "Updating content-creator-agent to $(IMAGE_TAG)..."
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/content-creator-agent/overlays/$(OVERLAY_PATH)/kustomization.yaml 2>/dev/null || true
	@if [ -d "$(KUSTOMIZE_DIR)/services/content-creator-agent/overlays/$(OVERLAY_PATH)" ]; then \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/content-creator-agent/overlays/$(OVERLAY_PATH); \
	fi

	# Update database agent definitions
	@$(MAKE) update-agent-images-v2 IMAGE_TAG=$(IMAGE_TAG)

	# Force rollout restart to pick up new images
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/agent-chassis -n ai-persona-system 2>/dev/null || true

	@echo "$(GREEN)All agents deployed with image tag $(IMAGE_TAG)$(NC)"

#  kubectl apply -k deployments/kustomize/services/agent-chassis/overlays/production/uk_001/

.PHONY: redeploy-agents
redeploy-agents:  ## Forces a rolling restart of all agent deployments
	@echo "$(YELLOW)Forcing rollout restart of agent deployments...$(NC)"
	kubectl rollout restart deployment agent-chassis -n ai-persona-system
	kubectl rollout restart deployment reasoning-agent -n ai-persona-system
	kubectl rollout restart deployment web-search-adapter -n ai-persona-system
	kubectl rollout restart deployment web-scrape-adapter -n ai-persona-system
	kubectl rollout restart deployment git-adapter -n ai-persona-system
	kubectl rollout restart deployment image-generator-adapter -n ai-persona-system
	kubectl rollout restart deployment content-creator-agent -n ai-persona-system


.PHONY: deploy-frontends
deploy-frontends: ## Deploy all frontend applications
	@echo "$(YELLOW)Deploying frontend applications...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/admin-dashboard/overlays/$(OVERLAY_PATH)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/user-portal/overlays/$(OVERLAY_PATH)
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/agent-playground/overlays/$(OVERLAY_PATH)

.PHONY: deploy-admin-dashboard
deploy-admin-dashboard: ## Deploy admin-dashboard only
	@echo "$(GREEN)Deploying admin-dashboard...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/admin-dashboard/overlays/$(OVERLAY_PATH)

.PHONY: deploy-user-portal
deploy-user-portal: ## Deploy user-portal only
	@echo "$(GREEN)Deploying user-portal...$(NC)"
	kubectl apply -k $(KUSTOMIZE_DIR)/frontends/user-portal/overlays/$(OVERLAY_PATH)

#################################
# Full Stack Operations
#################################
.PHONY: full-deploy
full-deploy: build-all push-all deploy-all ## Build, push, and deploy everything

.PHONY: quick-deploy
quick-deploy:  ## Deploy applications without building (uses existing images)
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
	kind load docker-image $(REGISTRY)/web-scrape-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/git-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/image-generator-adapter:$(IMAGE_TAG) --name personae-dev
	kind load docker-image $(REGISTRY)/content-creator-agent:$(IMAGE_TAG) --name personae-dev

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
create-dev-secrets: ## Create all development secrets (personae-dev-secrets and docker-hub-creds)
	@echo "$(YELLOW)Creating development namespace...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl create namespace $(PROJECT_NAME) --dry-run=client -o yaml | KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f -
	@echo "$(YELLOW)Creating personae-dev-secrets...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl create secret generic personae-dev-secrets \
		--from-literal=CLIENTS_DB_PASSWORD=$CLIENTS_DB_PASSWORD${} \
		--from-literal=TEMPLATES_DB_PASSWORD=$${TEMPLATES_DB_PASSWORD} \
		--from-literal=AUTH_DB_PASSWORD=$${AUTH_DB_PASSWORD} \
		--from-literal=MINIO_ACCESS_KEY=$${MINIO_ACCESS_KEY} \
		--from-literal=SECRET_KEY=$${SECRET_KEY} \
		--from-literal=JWT_SECRET_KEY=$${JWT_SECRET_KEY} \
		--from-literal=ANTHROPIC_API_KEY=$${ANTHROPIC_API_KEY} \
		--from-literal=SERP_API_KEY=$${SERP_API_KEY} \
		--from-literal=SCRAPING_BEE_API_KEY=$${SCRAPING_BEE_API_KEY} \
		--from-literal=FIRECRAWL_API_KEY=$${FIRECRAWL_API_KEY} \
		--from-literal=STABILITY_API_KEY=$${STABILITY_API_KEY:-not-a-real-key} \
		-n $(PROJECT_NAME) --dry-run=client -o yaml | KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f -
	@echo "$(GREEN)✓ personae-dev-secrets created$(NC)"
	@echo "$(YELLOW)Creating docker-hub-creds secret...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl create secret docker-registry docker-hub-creds \
		--namespace=$(PROJECT_NAME) \
		--docker-server=docker.io \
		--docker-username="$$(echo $${DOCKER_USERNAME} | tr -d '"')" \
		--docker-password="$$(echo $${DOCKER_PASSWORD} | tr -d '"')" \
		--docker-email="$$(echo $${DOCKER_EMAIL} | tr -d '"')" \
		--dry-run=client -o yaml | KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -f -
	@echo "$(GREEN)✓ docker-hub-creds created$(NC)"
	@echo "$(GREEN)All development secrets created successfully!$(NC)"

# Delete all development secrets
.PHONY: delete-dev-secrets
delete-dev-secrets: ## Delete all development secrets
	@echo "$(YELLOW)Deleting development secrets...$(NC)"
	@kubectl delete secret personae-dev-secrets -n $(PROJECT_NAME) --ignore-not-found
	@kubectl delete secret docker-hub-creds -n $(PROJECT_NAME) --ignore-not-found
	@echo "$(GREEN)Development secrets deleted$(NC)"

# Verify all development secrets
.PHONY: verify-dev-secrets
verify-dev-secrets: ## Verify all development secrets exist
	@echo "$(YELLOW)Verifying development secrets...$(NC)"
	@kubectl get secret personae-dev-secrets -n $(PROJECT_NAME) -o name && echo "$(GREEN)✓ personae-dev-secrets exists$(NC)" || echo "$(RED)✗ personae-dev-secrets missing$(NC)"
	@kubectl get secret docker-hub-creds -n $(PROJECT_NAME) -o name && echo "$(GREEN)✓ docker-hub-creds exists$(NC)" || echo "$(RED)✗ docker-hub-creds missing$(NC)"
	@echo "$(YELLOW)Docker registry config:$(NC)"
	@kubectl get secret docker-hub-creds -n $(PROJECT_NAME) -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d | jq -r '.auths."docker.io" | {username, email}' || true

#################################
# ConfigMap Management
#################################
.PHONY: create-dev-configs
create-dev-configs: ## Create development configmaps
	@echo "$(YELLOW)Creating development configmaps...$(NC)"
	kubectl create namespace ai-persona-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployments/kustomize/infrastructure/configs/development/configmap-dev.yaml -n ai-persona-system


#################################
# Workflow Monitoring
#################################
.PHONY: build-workflow-monitor
build-workflow-monitor: ## Build workflow-monitor image
	@echo "$(YELLOW)Building workflow-monitor...$(NC)"
	docker build -t $(REGISTRY)/workflow-monitor:$(IMAGE_TAG) \
		-f build/docker/backend/workflow-monitor.dockerfile .

.PHONY: push-workflow-monitor
push-workflow-monitor: ## Push workflow-monitor image
	docker push $(REGISTRY)/workflow-monitor:$(IMAGE_TAG)

# Quick monitoring commands
.PHONY: monitor-workflows
monitor-workflows: ## Run workflow monitor as a one-off command
	@echo "$(YELLOW)Checking workflow status...$(NC)"
	kubectl run workflow-monitor-$(shell date +%s) \
		--image=$(REGISTRY)/workflow-monitor:$(IMAGE_TAG) \
		--rm -it --restart=Never \
		-n $(PROJECT_NAME) \
		--env="DATABASE_URL=postgresql://clients_user:password@postgres-clients:5432/clients_db?sslmode=disable" \
		--env="CLIENT_ID=demo_client" \
		-- /workflow-monitor -stuck-hours=1

.PHONY: monitor-stuck
monitor-stuck: ## Check for stuck workflows
	@echo "$(YELLOW)Checking for stuck workflows...$(NC)"
	kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT correlation_id, current_step, status, \
		 EXTRACT(EPOCH FROM (NOW() - updated_at))/3600 as hours_stuck \
		 FROM orchestrator_state \
		 WHERE status IN ('RUNNING', 'AWAITING_RESPONSES') \
		 AND updated_at < NOW() - INTERVAL '1 hour' \
		 ORDER BY updated_at ASC;"

.PHONY: monitor-active
monitor-active: ## Show active workflows
	@echo "$(YELLOW)Active workflows:$(NC)"
	kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT correlation_id, current_step, status, \
		 execution_metadata->>'completed_steps' as completed, \
		 execution_metadata->>'total_steps' as total, \
		 ROUND(((execution_metadata->>'completed_steps')::numeric / \
		        NULLIF((execution_metadata->>'total_steps')::numeric, 0)) * 100, 1) as progress_pct \
		 FROM orchestrator_state \
		 WHERE status NOT IN ('COMPLETED', 'FAILED') \
		 ORDER BY updated_at DESC \
		 LIMIT 20;"

.PHONY: monitor-metrics
monitor-metrics: ## Show workflow metrics for last 24 hours
	@echo "$(YELLOW)Workflow metrics (24h):$(NC)"
	kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT \
		 COUNT(*) as total, \
		 COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed, \
		 COUNT(CASE WHEN status = 'FAILED' THEN 1 END) as failed, \
		 COUNT(CASE WHEN status IN ('RUNNING', 'AWAITING_RESPONSES') THEN 1 END) as active, \
		 ROUND(100.0 * COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) / NULLIF(COUNT(*), 0), 1) as success_rate \
		 FROM orchestrator_state \
		 WHERE created_at > NOW() - INTERVAL '24 hours';"

# Add these targets to your Makefile

#################################
# Database Operations - Runtime Management
#################################

# Quick SQL execution for runtime changes
.PHONY: db-exec-templates
db-exec-templates: ## Execute SQL in templates DB
	@echo "$(YELLOW)Executing SQL in templates DB...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- \
		psql -U templates_user -d templates_db

.PHONY: db-exec-clients
db-exec-clients: ## Execute SQL in clients DB
	@echo "$(YELLOW)Executing SQL in clients DB...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-clients-0 -n $(PROJECT_NAME) -- \
		psql -U clients_user -d clients_db

# Create new agent definition on the fly
.PHONY: agent-create
agent-create: ## Create a new agent definition (usage: make agent-create TYPE=analyzer NAME="Data Analyzer")
	@echo "$(YELLOW)Creating agent definition: $(TYPE)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		INSERT INTO agent_definitions (type, display_name, description, category, default_config, capabilities) VALUES \
		('$(TYPE)', '$(NAME)', '$(DESC)', 'data-driven', \
		'{\"model\": \"claude-3-5-sonnet-20241022\", \"temperature\": 0.5, \"processing_mode\": \"task\", \
		  \"workflow\": {\"start_step\": \"process\", \"steps\": { \
		    \"process\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \
		    \"complete\": {\"action\": \"complete_workflow\"}}}}', \
		'[\"analysis\", \"$(TYPE)\"]'::jsonb) \
		ON CONFLICT (type) DO UPDATE SET \
		  display_name = EXCLUDED.display_name, \
		  updated_at = NOW() \
		RETURNING id, type, display_name;"

# Update agent configuration
.PHONY: agent-update-config
agent-update-config: ## Update agent config (usage: make agent-update-config TYPE=analyzer CONFIG='{"temperature": 0.7}')
	@echo "$(YELLOW)Updating agent config for: $(TYPE)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		UPDATE agent_definitions \
		SET default_config = default_config || '$(CONFIG)'::jsonb, \
		    updated_at = NOW() \
		WHERE type = '$(TYPE)' \
		RETURNING type, default_config;"

# List all agent definitions
.PHONY: agent-list
agent-list: ## List all agent definitions
	@echo "$(YELLOW)Agent Definitions:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT type, display_name, category, \
		       array_length(capabilities::text[], 1) as cap_count, \
		       is_active, \
		       to_char(updated_at, 'YYYY-MM-DD HH24:MI') as last_updated \
		FROM agent_definitions \
		ORDER BY updated_at DESC;"

# Show agent performance
.PHONY: agent-performance
agent-performance: ## Show agent performance metrics
	@echo "$(YELLOW)Agent Performance Metrics:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT agent_type, \
		       total_tasks, \
		       ROUND(success_rate * 100, 1) || '%' as success_rate, \
		       avg_response_time_ms || 'ms' as avg_time, \
		       ROUND(avg_quality_score, 2) as quality \
		FROM agent_metrics \
		WHERE total_tasks > 0 \
		ORDER BY success_rate DESC;"

# Create agent group dynamically
.PHONY: group-create
group-create: ## Create agent group (usage: make group-create NAME="Analysis Team" TYPE=analysis)
	@echo "$(YELLOW)Creating agent group: $(NAME)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		INSERT INTO agent_groups (name, group_type, agent_configs, orchestration_workflow) \
		VALUES ('$(NAME)', '$(TYPE)', \
		'[{\"role\": \"lead\", \"agent_type\": \"$(TYPE)-leader\"}, \
		  {\"role\": \"worker\", \"agent_type\": \"$(TYPE)-worker\"}]'::jsonb, \
		'{\"start_step\": \"validate\", \"steps\": {}}'::jsonb) \
		RETURNING id, name, group_type;"

# Hot reload agent configuration (notifies running agents)
.PHONY: agent-hot-reload
agent-hot-reload: ## Hot reload agent config (usage: make agent-hot-reload AGENT_ID=xxx CONFIG='{"key": "value"}')
	@echo "$(YELLOW)Hot reloading config for agent: $(AGENT_ID)$(NC)"
	@echo '{"type": "config_update", "agent_id": "$(AGENT_ID)", "config": $(CONFIG)}' | \
		KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -i kafka-cluster-kafka-0 -n $(PROJECT_NAME) -- \
		/opt/kafka/bin/kafka-console-producer.sh \
		--broker-list localhost:9092 \
		--topic system.agent.$(AGENT_ID).control

# Test discovery functions
.PHONY: agent-discover
agent-discover: ## Test agent discovery (usage: make agent-discover CAPS="analysis,reporting")
	@echo "$(YELLOW)Discovering agents with capabilities: $(CAPS)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT * FROM find_agents_by_capability('{$(CAPS)}'::text[], 'demo_client');"

# Recommend agents for task
.PHONY: agent-recommend
agent-recommend: ## Get agent recommendations (usage: make agent-recommend TASK=website-builder)
	@echo "$(YELLOW)Recommending agents for task: $(TASK)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -it postgres-templates-0 -n $(PROJECT_NAME) -- psql -U templates_user -d templates_db -c "\
		SELECT agent_type, display_name, \
		       ROUND(performance_score * 100) || '%' as score, \
		       recommendation_reason \
		FROM recommend_agents_for_task('$(TASK)', NULL);"

# Quick agent spawn via API call
.PHONY: agent-spawn
agent-spawn: ## Spawn an agent instance (usage: make agent-spawn TYPE=analyzer CLIENT=demo_client)
	@echo "$(YELLOW)Spawning agent: $(TYPE) for client: $(CLIENT)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl run spawn-agent-$(shell date +%s) --rm -i --restart=Never \
		--image=curlimages/curl -n $(PROJECT_NAME) -- \
		curl -X POST http://core-manager:8088/api/v1/agents/spawn \
		-H "Content-Type: application/json" \
		-d '{"agent_type": "$(TYPE)", "client_id": "$(CLIENT)", "spawn_job": true}'

# Monitor agent jobs
.PHONY: agent-jobs
agent-jobs: ## Show running agent jobs
	@echo "$(YELLOW)Running Agent Jobs:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl get jobs -n $(PROJECT_NAME) -l spawned-by=orchestrator \
		-o custom-columns=NAME:.metadata.name,TYPE:.metadata.labels.agent-type,STATUS:.status.conditions[0].type,AGE:.metadata.creationTimestamp

# Clean up completed agent jobs
.PHONY: agent-cleanup
agent-cleanup: ## Clean up completed agent jobs
	@echo "$(YELLOW)Cleaning up completed agent jobs...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl delete jobs -n $(PROJECT_NAME) -l spawned-by=orchestrator \
		--field-selector status.successful=1

# Add a specific target to just deploy/update the bootstrap agents
.PHONY: bootstrap-agents
bootstrap-agents: deploy-100-bootstrap-agents ## Deploy or update bootstrap agents

# Destroy bootstrap agents if needed
.PHONY: destroy-bootstrap-agents
destroy-bootstrap-agents: ## Destroy bootstrap agents
	@echo "$(RED)Destroying bootstrap agents...$(NC)"
	@cd $(TERRAFORM_DIR)/100-bootstrap-agents && \
		if [ -f terraform.tfvars.secret ]; then \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve -var-file=terraform.tfvars.secret; \
		else \
			KUBECONFIG=$(KUBECONFIG_PATH) terraform destroy -auto-approve; \
		fi

# Check bootstrap agent status
.PHONY: bootstrap-status
bootstrap-status: ## Check status of bootstrap agents
	@echo "$(YELLOW)Bootstrap Agent Status:$(NC)"
	@kubectl get statefulset -n $(PROJECT_NAME) generic-orchestrator
	@echo "\n$(YELLOW)Bootstrap Agent Pods:$(NC)"
	@kubectl get pods -n $(PROJECT_NAME) -l app=generic-orchestrator
	@echo "\n$(YELLOW)Bootstrap Agent Logs (last 20 lines):$(NC)"
	@kubectl logs -n $(PROJECT_NAME) -l app=generic-orchestrator --tail=20


#################################
# Agent Image Management
#################################

# Update agent definitions with current image tag
.PHONY: update-agent-images
update-agent-images: ## Update all agent definitions with current image tag
	@echo "$(YELLOW)Updating agent definitions with image: $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -i postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"UPDATE agent_definitions SET image_repository = '$(REGISTRY)/agent-chassis', image_tag = '$(IMAGE_TAG)', updated_at = NOW();"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -i postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c \
		"SELECT type, image_repository, image_tag FROM agent_definitions ORDER BY type LIMIT 5;"
	@echo "$(GREEN)Agent definitions updated with $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"

# Alternative version using a single command
.PHONY: update-agent-images-v2
update-agent-images-v2: ## Update all agent definitions with current image tag (alternative)
	@echo "$(YELLOW)Updating agent definitions with image: $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -i postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -c "\
		UPDATE agent_definitions \
		SET image_repository = '$(REGISTRY)/agent-chassis', \
		    image_tag = '$(IMAGE_TAG)', \
		    updated_at = NOW(); \
		SELECT type, image_repository, image_tag \
		FROM agent_definitions \
		ORDER BY type \
		LIMIT 5;"
	@echo "$(GREEN)Agent definitions updated$(NC)"

# Update agent images and restart orchestrator
.PHONY: update-generic-orchestrator
update-generic-orchestrator: ## Update generic orchestrator image to current IMAGE_TAG
	@echo "$(YELLOW)Updating generic orchestrator to $(REGISTRY)/agent-chassis:$(IMAGE_TAG)...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) set image statefulset/generic-orchestrator \
		orchestrator=$(REGISTRY)/agent-chassis:$(IMAGE_TAG)
	@echo "$(GREEN)Waiting for rollout to complete...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) rollout status statefulset/generic-orchestrator --timeout=120s
	@echo "$(GREEN)Generic orchestrator updated to $(IMAGE_TAG)$(NC)"

.PHONY: restart-generic-orchestrator
restart-generic-orchestrator: ## Restart generic orchestrator pod
	@echo "$(YELLOW)Restarting generic orchestrator...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete pod generic-orchestrator-0
	@echo "$(GREEN)Waiting for pod to be ready...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) wait --for=condition=ready pod/generic-orchestrator-0 --timeout=120s
	@echo "$(GREEN)Generic orchestrator restarted$(NC)"

.PHONY: update-and-restart-orchestrator
update-and-restart-orchestrator: update-generic-orchestrator restart-generic-orchestrator ## Update and restart generic orchestrator
	@echo "$(GREEN)Generic orchestrator updated and restarted with $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"

.PHONY: sync-all-agents
sync-all-agents: update-agent-images-v2 update-generic-orchestrator ## Update database and generic orchestrator to same image
	@echo "$(YELLOW)Cleaning up old agent pods to force respawn with new image...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete jobs -l app=dynamic-agent 2>/dev/null || true
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) delete pods -l app=dynamic-agent 2>/dev/null || true
	@echo "$(GREEN)All agents synced to $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"

.PHONY: verify-agent-images
verify-agent-images: ## Verify all agent images are consistent
	@echo "$(YELLOW)Checking agent image versions...$(NC)"
	@echo "$(CYAN)Database agent definitions:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl exec -i postgres-clients-0 -n $(PROJECT_NAME) -- psql -U clients_user -d clients_db -t -c \
		"SELECT DISTINCT image_repository || ':' || image_tag as image FROM agent_definitions WHERE is_active = true;" 2>/dev/null || echo "Failed to query database"
	@echo "$(CYAN)Generic orchestrator:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get statefulset generic-orchestrator -o jsonpath='{.spec.template.spec.containers[0].image}' && echo
	@echo "$(CYAN)Running dynamic agents:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get pods -l app=dynamic-agent -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image 2>/dev/null || echo "No dynamic agents running"
	@echo "$(CYAN)Agent chassis deployment:$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl -n $(PROJECT_NAME) get deployment agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null && echo || echo "No agent-chassis deployment"

.PHONY: quick-agent-update
quick-agent-update: ## Build, push and deploy agent-chassis with current IMAGE_TAG
	@echo "$(YELLOW)Building agent-chassis:$(IMAGE_TAG)...$(NC)"
	@$(MAKE) build-agent-chassis IMAGE_TAG=$(IMAGE_TAG)
	@echo "$(YELLOW)Pushing agent-chassis:$(IMAGE_TAG)...$(NC)"
	@docker push $(REGISTRY)/agent-chassis:$(IMAGE_TAG)
	@echo "Updating agent-chassis kustomization to $(IMAGE_TAG)...$(NC)"
	@sed -i.bak 's/newTag:.*/newTag: $(IMAGE_TAG)/' $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)/kustomization.yaml
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)
	@echo "$(YELLOW)Deploying...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k $(KUSTOMIZE_DIR)/services/agent-chassis/overlays/$(OVERLAY_PATH)
	@echo "$(YELLOW)Updating database...$(NC)"
	@$(MAKE) update-agent-images-v2 IMAGE_TAG=$(IMAGE_TAG)
	@echo "$(YELLOW)Restarting pods...$(NC)"
	@KUBECONFIG=$(KUBECONFIG_PATH) kubectl rollout restart deployment/agent-chassis -n ai-persona-system
	@echo "$(GREEN)Deployment complete with $(REGISTRY)/agent-chassis:$(IMAGE_TAG)$(NC)"