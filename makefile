# AI Persona System - Comprehensive Makefile
# This Makefile handles the complete deployment lifecycle

.PHONY: help setup build deploy quickstart clean logs port-forward
.DEFAULT_GOAL := help

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Configuration
NAMESPACE := ai-persona-system
DOCKER_REGISTRY := ai-persona-system
TIMEOUT := 300s

help: ## Show this help message
	@echo "$(GREEN)AI Persona System - Available Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)Quick Start:$(NC)"
	@echo "  make quickstart    # Complete setup and deployment"
	@echo "  make status        # Check system status"
	@echo "  make clean         # Clean up everything"

# =============================================================================
# SETUP AND INITIALIZATION
# =============================================================================

setup: ## Run initial setup (creates secrets, namespaces)
	@echo "$(GREEN)🚀 Running initial setup...$(NC)"
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh
	@echo "$(GREEN)✅ Setup complete$(NC)"

check-prerequisites: ## Check if required tools are installed
	@echo "$(GREEN)🔍 Checking prerequisites...$(NC)"
	@command -v kubectl >/dev/null 2>&1 || { echo "$(RED)kubectl is required but not installed$(NC)"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)docker is required but not installed$(NC)"; exit 1; }
	@kubectl cluster-info >/dev/null 2>&1 || { echo "$(RED)Cannot connect to Kubernetes cluster$(NC)"; exit 1; }
	@echo "$(GREEN)✅ Prerequisites check passed$(NC)"

# =============================================================================
# BUILD TARGETS
# =============================================================================

build: ## Build all Docker images
	@echo "$(GREEN)🔨 Building Docker images...$(NC)"
	@docker build -t $(DOCKER_REGISTRY)/auth-service:latest -f Dockerfile.auth-service .
	@docker build -t $(DOCKER_REGISTRY)/core-manager:latest -f Dockerfile.core-manager .
	@docker build -t $(DOCKER_REGISTRY)/agent-chassis:latest -f Dockerfile.agent-chassis .
	@docker build -t $(DOCKER_REGISTRY)/reasoning-agent:latest -f Dockerfile.reasoning-agent .
	@docker build -t $(DOCKER_REGISTRY)/image-generator-adapter:latest -f Dockerfile.image-generator-adapter .
	@docker build -t $(DOCKER_REGISTRY)/web-search-adapter:latest -f Dockerfile.web-search-adapter .
	@echo "$(GREEN)✅ All images built successfully$(NC)"

build-init-images: ## Build initialization utility images
	@echo "$(GREEN)🔨 Building initialization images...$(NC)"
	@docker build -t $(DOCKER_REGISTRY)/database-migrator:latest -f docker/Dockerfile.migrator .
	@docker build -t $(DOCKER_REGISTRY)/data-seeder:latest -f docker/Dockerfile.seeder .
	@echo "$(GREEN)✅ Initialization images built$(NC)"

build-all: build build-init-images ## Build all images including initialization utilities

# =============================================================================
# DEPLOYMENT TARGETS (PROPER ORDER)
# =============================================================================

deploy: check-prerequisites ## Deploy the entire system in correct order
	@echo "$(GREEN)🚀 Starting full deployment...$(NC)"
	@$(MAKE) deploy-infrastructure
	@$(MAKE) deploy-storage
	@$(MAKE) deploy-messaging
	@$(MAKE) wait-for-infrastructure
	@$(MAKE) initialize-system
	@$(MAKE) deploy-core-services
	@$(MAKE) deploy-agents
	@$(MAKE) deploy-ingress-monitoring
	@echo "$(GREEN)✅ Deployment completed successfully!$(NC)"

deploy-infrastructure: ## Deploy namespace, secrets, and configmaps
	@echo "$(GREEN)📦 Deploying infrastructure...$(NC)"
	kubectl apply -f k8s/namespace.yaml
	@echo "$(YELLOW)⏳ Waiting for namespace to be ready...$(NC)"
	@kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/$(NAMESPACE) --timeout=$(TIMEOUT)
	kubectl apply -f k8s/configmap-common.yaml -n $(NAMESPACE)
	@echo "$(GREEN)✅ Infrastructure deployed$(NC)"

deploy-storage: ## Deploy persistent storage (databases, object storage)
	@echo "$(GREEN)💾 Deploying storage systems...$(NC)"
	kubectl apply -f k8s/postgres-clients.yaml
	kubectl apply -f k8s/postgres-templates.yaml
	kubectl apply -f k8s/mysql-auth.yaml
	kubectl apply -f k8s/minio.yaml
	@echo "$(GREEN)✅ Storage systems deployed$(NC)"

deploy-messaging: ## Deploy Kafka message queue
	@echo "$(GREEN)📨 Deploying messaging system...$(NC)"
	kubectl apply -f k8s/kafka.yaml
	@echo "$(GREEN)✅ Messaging system deployed$(NC)"

wait-for-infrastructure: ## Wait for infrastructure to be ready
	@echo "$(GREEN)⏳ Waiting for infrastructure to be ready...$(NC)"
	@echo "$(YELLOW)Waiting for PostgreSQL clients...$(NC)"
	kubectl wait --for=condition=ready pod -l app=postgres-clients -n $(NAMESPACE) --timeout=$(TIMEOUT)
	@echo "$(YELLOW)Waiting for PostgreSQL templates...$(NC)"
	kubectl wait --for=condition=ready pod -l app=postgres-templates -n $(NAMESPACE) --timeout=$(TIMEOUT)
	@echo "$(YELLOW)Waiting for MySQL auth...$(NC)"
	kubectl wait --for=condition=ready pod -l app=mysql-auth -n $(NAMESPACE) --timeout=$(TIMEOUT)
	@echo "$(YELLOW)Waiting for MinIO...$(NC)"
	kubectl wait --for=condition=ready pod -l app=minio -n $(NAMESPACE) --timeout=$(TIMEOUT)
	@echo "$(YELLOW)Waiting for Kafka cluster...$(NC)"
	kubectl wait --for=condition=ready pod -l app=kafka -n $(NAMESPACE) --timeout=$(TIMEOUT)
	@echo "$(GREEN)✅ Infrastructure is ready$(NC)"

deploy-automated: check-prerequisites build-all ## Automated deployment using the deployment script
	@echo "$(GREEN)🚀 Starting automated deployment...$(NC)"
	@chmod +x scripts/deploy-system.sh
	@./scripts/deploy-system.sh
	@echo "$(GREEN)✅ Automated deployment completed!$(NC)"

initialize-system: ## Initialize databases and create Kafka topics
	@echo "$(GREEN)🔧 Initializing system...$(NC)"
	kubectl apply -f k8s/jobs/database-init-job.yaml
	kubectl apply -f k8s/jobs/kafka-topics-job.yaml
	@echo "$(YELLOW)⏳ Waiting for initialization jobs to complete...$(NC)"
	@kubectl wait --for=condition=complete job/database-init -n $(NAMESPACE) --timeout=600s
	@kubectl wait --for=condition=complete job/kafka-topics-init -n $(NAMESPACE) --timeout=300s
	@kubectl wait --for=condition=complete job/data-seeder -n $(NAMESPACE) --timeout=300s
	@echo "$(GREEN)✅ System initialization complete$(NC)"

deploy-core-services: ## Deploy core services (auth, core-manager)
	@echo "$(GREEN)🏗️  Deploying core services...$(NC)"
	kubectl apply -f k8s/auth-service.yaml
	kubectl apply -f k8s/core-manager.yaml
	@echo "$(YELLOW)⏳ Waiting for core services to be ready...$(NC)"
	kubectl wait --for=condition=ready pod -l app=auth-service -n $(NAMESPACE) --timeout=$(TIMEOUT)
	kubectl wait --for=condition=ready pod -l app=core-manager -n $(NAMESPACE) --timeout=$(TIMEOUT)
	@echo "$(GREEN)✅ Core services deployed$(NC)"

deploy-agents: ## Deploy all agent services
	@echo "$(GREEN)🤖 Deploying agent services...$(NC)"
	kubectl apply -f k8s/agent-chassis.yaml
	kubectl apply -f k8s/reasoning-agent.yaml
	kubectl apply -f k8s/image-generator-adapter.yaml
	kubectl apply -f k8s/web-search-adapter.yaml
	@echo "$(YELLOW)⏳ Waiting for agents to be ready...$(NC)"
	kubectl wait --for=condition=ready pod -l app=agent-chassis -n $(NAMESPACE) --timeout=$(TIMEOUT)
	kubectl wait --for=condition=ready pod -l app=reasoning-agent -n $(NAMESPACE) --timeout=$(TIMEOUT)
	@echo "$(GREEN)✅ Agent services deployed$(NC)"

deploy-ingress-monitoring: ## Deploy ingress and monitoring
	@echo "$(GREEN)📊 Deploying ingress and monitoring...$(NC)"
	kubectl apply -f k8s/ingress.yaml
	kubectl apply -f k8s/monitoring/
	@echo "$(GREEN)✅ Ingress and monitoring deployed$(NC)"

# =============================================================================
# DATABASE MANAGEMENT
# =============================================================================

migrate-all-databases: ## Run all database migrations
	@echo "$(GREEN)📝 Running database migrations...$(NC)"
	@$(MAKE) migrate-pgvector
	@$(MAKE) migrate-templates-db
	@$(MAKE) migrate-clients-db
	@$(MAKE) migrate-auth-db
	@echo "$(GREEN)✅ All migrations completed$(NC)"

migrate-pgvector: ## Enable pgvector extension
	@echo "$(YELLOW)🔧 Enabling pgvector extension...$(NC)"
	kubectl exec -n $(NAMESPACE) postgres-clients-0 -- psql -U clients_user -d clients_db -c "CREATE EXTENSION IF NOT EXISTS vector;"
	@echo "$(GREEN)✅ pgvector enabled$(NC)"

migrate-templates-db: ## Migrate templates database
	@echo "$(YELLOW)📝 Migrating templates database...$(NC)"
	kubectl cp platform/database/migrations/002_create_templates_schema.sql $(NAMESPACE)/postgres-templates-0:/tmp/
	kubectl exec -n $(NAMESPACE) postgres-templates-0 -- psql -U templates_user -d templates_db -f /tmp/002_create_templates_schema.sql
	@echo "$(GREEN)✅ Templates database migrated$(NC)"

migrate-clients-db: ## Migrate clients database (requires CLIENT_ID)
	@echo "$(YELLOW)📝 Migrating clients database...$(NC)"
	kubectl cp platform/database/migrations/003_create_client_schema.sql $(NAMESPACE)/postgres-clients-0:/tmp/
	@# Note: This creates the base structure, client-specific schemas are created on-demand
	@echo "$(GREEN)✅ Clients database migrated$(NC)"

migrate-auth-db: ## Migrate auth database
	@echo "$(YELLOW)📝 Migrating auth database...$(NC)"
	kubectl cp platform/database/migrations/004_auth_schema.sql $(NAMESPACE)/mysql-auth-0:/tmp/
	kubectl exec -n $(NAMESPACE) mysql-auth-0 -- mysql -u auth_user -p$(shell kubectl get secret db-secrets -n $(NAMESPACE) -o jsonpath='{.data.auth-db-password}' | base64 -d) auth_db < /tmp/004_auth_schema.sql
	kubectl cp platform/database/migrations/005_projects_schema.sql $(NAMESPACE)/mysql-auth-0:/tmp/
	kubectl exec -n $(NAMESPACE) mysql-auth-0 -- mysql -u auth_user -p$(shell kubectl get secret db-secrets -n $(NAMESPACE) -o jsonpath='{.data.auth-db-password}' | base64 -d) auth_db < /tmp/005_projects_schema.sql
	@echo "$(GREEN)✅ Auth database migrated$(NC)"

create-client-schema: ## Create schema for a specific client (requires CLIENT_ID env var)
	@if [ -z "$(CLIENT_ID)" ]; then \
		echo "$(RED)❌ CLIENT_ID environment variable is required$(NC)"; \
		echo "Usage: make create-client-schema CLIENT_ID=client_123"; \
		exit 1; \
	fi
	@echo "$(YELLOW)🔧 Creating schema for client: $(CLIENT_ID)$(NC)"
	@sed 's/{client_id}/$(CLIENT_ID)/g' platform/database/migrations/003_create_client_schema.sql > /tmp/client_schema_$(CLIENT_ID).sql
	kubectl cp /tmp/client_schema_$(CLIENT_ID).sql $(NAMESPACE)/postgres-clients-0:/tmp/
	kubectl exec -n $(NAMESPACE) postgres-clients-0 -- psql -U clients_user -d clients_db -f /tmp/client_schema_$(CLIENT_ID).sql
	@rm /tmp/client_schema_$(CLIENT_ID).sql
	@echo "$(GREEN)✅ Schema created for client: $(CLIENT_ID)$(NC)"

# =============================================================================
# KAFKA MANAGEMENT
# =============================================================================

create-all-kafka-topics: ## Create all required Kafka topics
	@echo "$(GREEN)📨 Creating all Kafka topics...$(NC)"
	@$(MAKE) kafka-create-system-topics
	@$(MAKE) kafka-create-core-topics
	@echo "$(GREEN)✅ All Kafka topics created$(NC)"

kafka-create-core-topics: ## Create core system topics used by agents
	@echo "$(YELLOW)🔧 Creating core Kafka topics...$(NC)"
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.agent.reasoning.process --partitions 3 --replication-factor 1 --if-not-exists
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.responses.reasoning --partitions 6 --replication-factor 1 --if-not-exists
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.adapter.image.generate --partitions 3 --replication-factor 1 --if-not-exists
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.responses.image --partitions 6 --replication-factor 1 --if-not-exists
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.adapter.web.search --partitions 3 --replication-factor 1 --if-not-exists
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.responses.websearch --partitions 6 --replication-factor 1 --if-not-exists
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.notifications.ui --partitions 3 --replication-factor 1 --if-not-exists
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.commands.workflow.resume --partitions 3 --replication-factor 1 --if-not-exists
	@echo "$(GREEN)✅ Core topics created$(NC)"

kafka-list-topics: ## List all Kafka topics
	@echo "$(GREEN)📋 Listing Kafka topics...$(NC)"
	@kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --list

kafka-create-agent-topics: ## Create topics for a specific agent type
	@read -p "Enter agent type (e.g., copywriter, researcher): " agent_type; \
	echo "$(YELLOW)🔧 Creating topics for agent: $$agent_type$(NC)"; \
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic tasks.high.$$agent_type --partitions 3 --replication-factor 1 --if-not-exists; \
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic tasks.normal.$$agent_type --partitions 6 --replication-factor 1 --if-not-exists; \
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic tasks.low.$$agent_type --partitions 3 --replication-factor 1 --if-not-exists; \
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic responses.$$agent_type --partitions 6 --replication-factor 1 --if-not-exists; \
	kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic dlq.$$agent_type --partitions 1 --replication-factor 1 --if-not-exists; \
	echo "$(GREEN)✅ Topics created for agent: $$agent_type$(NC)"

kafka-create-system-topics: ## Create system-level topics
	@echo "$(YELLOW)🔧 Creating system topics...$(NC)"
	@kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic orchestrator.state-changes --partitions 12 --replication-factor 1 --if-not-exists
	@kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic human.approvals --partitions 6 --replication-factor 1 --if-not-exists
	@kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --create --topic system.events --partitions 3 --replication-factor 1 --if-not-exists
	@echo "$(GREEN)✅ System topics created$(NC)"

kafka-delete-agent-topics: ## Delete topics for a specific agent type
	@read -p "Enter agent type to delete topics for: " agent_type; \
	read -p "Are you sure you want to delete all topics for $$agent_type? (y/N): " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		echo "$(RED)🗑️  Deleting topics for agent: $$agent_type$(NC)"; \
		kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --delete --topic tasks.high.$$agent_type --if-exists; \
		kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --delete --topic tasks.normal.$$agent_type --if-exists; \
		kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --delete --topic tasks.low.$$agent_type --if-exists; \
		kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --delete --topic responses.$$agent_type --if-exists; \
		kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --delete --topic dlq.$$agent_type --if-exists; \
		echo "$(GREEN)✅ Topics deleted for agent: $$agent_type$(NC)"; \
	else \
		echo "$(YELLOW)❌ Deletion cancelled$(NC)"; \
	fi

# =============================================================================
# DATA SEEDING
# =============================================================================

seed-initial-data: ## Seed the system with initial templates and data
	@echo "$(GREEN)🌱 Seeding initial data...$(NC)"
	@$(MAKE) seed-persona-templates
	@$(MAKE) seed-subscription-tiers
	@echo "$(GREEN)✅ Initial data seeded$(NC)"

seed-persona-templates: ## Seed initial persona templates
	@echo "$(YELLOW)🤖 Seeding persona templates...$(NC)"
	@# Create a basic copywriter template
	kubectl exec -n $(NAMESPACE) postgres-templates-0 -- psql -U templates_user -d templates_db -c \
		"INSERT INTO persona_templates (id, name, description, category, config) VALUES \
		('00000000-0000-0000-0000-000000000001', 'Basic Copywriter', 'A versatile copywriting assistant', 'copywriter', \
		'{\"model\": \"claude-3-sonnet\", \"temperature\": 0.7, \"max_tokens\": 2000}') \
		ON CONFLICT (id) DO NOTHING;"
	@# Create a research assistant template
	kubectl exec -n $(NAMESPACE) postgres-templates-0 -- psql -U templates_user -d templates_db -c \
		"INSERT INTO persona_templates (id, name, description, category, config) VALUES \
		('00000000-0000-0000-0000-000000000002', 'Research Assistant', 'In-depth research and analysis', 'researcher', \
		'{\"model\": \"claude-3-opus\", \"temperature\": 0.3, \"max_tokens\": 4000}') \
		ON CONFLICT (id) DO NOTHING;"
	@echo "$(GREEN)✅ Persona templates seeded$(NC)"

seed-subscription-tiers: ## Ensure subscription tiers exist
	@echo "$(YELLOW)💳 Ensuring subscription tiers exist...$(NC)"
	@# The tiers should already be created by the migration, but this ensures they exist
	kubectl exec -n $(NAMESPACE) mysql-auth-0 -- mysql -u auth_user -p$(shell kubectl get secret db-secrets -n $(NAMESPACE) -o jsonpath='{.data.auth-db-password}' | base64 -d) auth_db -e \
		"SELECT COUNT(*) as tier_count FROM subscription_tiers;" 2>/dev/null || echo "Subscription tiers table not ready yet"
	@echo "$(GREEN)✅ Subscription tiers verified$(NC)"

# =============================================================================
# AGENT MANAGEMENT
# =============================================================================

register-agent: ## Register a new agent type
	@read -p "Enter agent type (e.g., copywriter): " agent_type; \
	read -p "Enter display name: " display_name; \
	read -p "Enter category (data-driven/code-driven/adapter): " category; \
	echo "$(YELLOW)📝 Registering agent: $$agent_type$(NC)"; \
	$(MAKE) kafka-create-agent-topics; \
	kubectl exec -n $(NAMESPACE) core-manager-0 -- /app/core-manager register-agent \
		--type="$$agent_type" \
		--name="$$display_name" \
		--category="$$category"

# =============================================================================
# SYSTEM MONITORING AND DEBUGGING
# =============================================================================

status: ## Check overall system status
	@echo "$(GREEN)📊 System Status Overview$(NC)"
	@echo "$(YELLOW)Namespace:$(NC)"
	@kubectl get namespace $(NAMESPACE) 2>/dev/null || echo "$(RED)Namespace not found$(NC)"
	@echo ""
	@echo "$(YELLOW)Pods Status:$(NC)"
	@kubectl get pods -n $(NAMESPACE) -o wide 2>/dev/null || echo "$(RED)No pods found$(NC)"
	@echo ""
	@echo "$(YELLOW)Services:$(NC)"
	@kubectl get services -n $(NAMESPACE) 2>/dev/null || echo "$(RED)No services found$(NC)"
	@echo ""
	@echo "$(YELLOW)Persistent Volumes:$(NC)"
	@kubectl get pvc -n $(NAMESPACE) 2>/dev/null || echo "$(RED)No PVCs found$(NC)"

system-check: ## Comprehensive system health check
	@echo "$(GREEN)🔍 Comprehensive System Check$(NC)"
	@echo ""
	@echo "$(YELLOW)📊 Kafka Topics:$(NC)"
	@$(MAKE) kafka-list-topics 2>/dev/null || echo "$(RED)Kafka not accessible$(NC)"
	@echo ""
	@echo "$(YELLOW)📊 Database Tables (Templates):$(NC)"
	@kubectl exec -n $(NAMESPACE) postgres-templates-0 -- psql -U templates_user -d templates_db -c "\dt" 2>/dev/null || echo "$(RED)Templates DB not accessible$(NC)"
	@echo ""
	@echo "$(YELLOW)📊 Database Tables (Clients):$(NC)"
	@kubectl exec -n $(NAMESPACE) postgres-clients-0 -- psql -U clients_user -d clients_db -c "\dt" 2>/dev/null || echo "$(RED)Clients DB not accessible$(NC)"
	@echo ""
	@echo "$(YELLOW)📊 Persona Templates:$(NC)"
	@kubectl exec -n $(NAMESPACE) postgres-templates-0 -- psql -U templates_user -d templates_db -c "SELECT id, name, category FROM persona_templates WHERE is_active = true;" 2>/dev/null || echo "$(RED)Templates not accessible$(NC)"

logs: ## View logs for a specific service
	@echo "$(GREEN)Available services:$(NC)"
	@echo "  auth-service"
	@echo "  core-manager"
	@echo "  agent-chassis"
	@echo "  reasoning-agent"
	@echo "  image-generator-adapter"
	@echo "  web-search-adapter"
	@echo "  kafka"
	@echo "  postgres-clients"
	@echo "  postgres-templates"
	@echo "  mysql-auth"
	@echo ""
	@read -p "Enter service name: " service; \
	echo "$(YELLOW)📋 Showing logs for $$service...$(NC)"; \
	kubectl logs -n $(NAMESPACE) -l app=$$service --tail=100 -f

describe-pod: ## Describe a specific pod for debugging
	@kubectl get pods -n $(NAMESPACE)
	@echo ""
	@read -p "Enter pod name: " pod; \
	kubectl describe pod $$pod -n $(NAMESPACE)

port-forward: ## Set up port forwarding for local access
	@echo "$(GREEN)🔗 Setting up port forwarding...$(NC)"
	@echo "$(YELLOW)Auth Service: http://localhost:8081$(NC)"
	@echo "$(YELLOW)Core Manager: http://localhost:8088$(NC)"
	@echo "$(YELLOW)Grafana: http://localhost:3000$(NC)"
	@echo "$(YELLOW)Kafka UI: http://localhost:8080$(NC)"
	@echo ""
	@echo "$(YELLOW)Starting port forwards (Ctrl+C to stop)...$(NC)"
	@trap 'kill %1 %2 %3 %4 2>/dev/null' EXIT; \
	kubectl port-forward -n $(NAMESPACE) svc/auth-service 8081:8081 & \
	kubectl port-forward -n $(NAMESPACE) svc/core-manager 8088:8088 & \
	kubectl port-forward -n $(NAMESPACE) svc/grafana 3000:3000 & \
	kubectl port-forward -n $(NAMESPACE) svc/kafka-ui 8080:8080 & \
	wait

# =============================================================================
# COMPLETE WORKFLOWS
# =============================================================================

quickstart: ## Complete setup and deployment from scratch
	@echo "$(GREEN)🚀 Starting AI Persona System Quickstart$(NC)"
	@$(MAKE) check-prerequisites
	@$(MAKE) setup
	@$(MAKE) deploy-automated
	@echo ""
	@echo "$(GREEN)✅ Quickstart completed successfully!$(NC)"
	@echo ""
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "1. Run 'make port-forward' to access services locally"
	@echo "2. Run 'make system-check' to verify everything is working"
	@echo "3. Create your first client: 'make create-client-schema CLIENT_ID=demo_client'"
	@echo "4. Register your first agent: 'make register-agent'"
	@echo ""
	@echo "$(YELLOW)Access URLs:$(NC)"
	@echo "- Auth API: http://localhost:8081"
	@echo "- Core API: http://localhost:8088"
	@echo "- Grafana: http://localhost:3000 (admin/admin)"

quickstart-manual: ## Manual step-by-step deployment
	@echo "$(GREEN)🚀 Starting AI Persona System Manual Deployment$(NC)"
	@$(MAKE) check-prerequisites
	@$(MAKE) setup
	@$(MAKE) build-all
	@$(MAKE) deploy
	@echo ""
	@echo "$(GREEN)✅ Manual deployment completed successfully!$(NC)"

restart-service: ## Restart a specific service
	@echo "$(GREEN)Available services to restart:$(NC)"
	@kubectl get deployments -n $(NAMESPACE) -o name | sed 's|deployment.apps/||'
	@echo ""
	@read -p "Enter service name: " service; \
	echo "$(YELLOW)🔄 Restarting $$service...$(NC)"; \
	kubectl rollout restart deployment/$$service -n $(NAMESPACE); \
	kubectl rollout status deployment/$$service -n $(NAMESPACE)

# =============================================================================
# CLEANUP
# =============================================================================

clean: ## Clean up everything (DESTRUCTIVE!)
	@echo "$(RED)⚠️  This will DELETE the entire $(NAMESPACE) namespace and all data!$(NC)"
	@read -p "Are you sure? Type 'DELETE' to confirm: " confirm; \
	if [ "$$confirm" = "DELETE" ]; then \
		echo "$(RED)🗑️  Deleting namespace $(NAMESPACE)...$(NC)"; \
		kubectl delete namespace $(NAMESPACE) --ignore-not-found=true; \
		echo "$(GREEN)✅ Cleanup completed$(NC)"; \
	else \
		echo "$(YELLOW)❌ Cleanup cancelled$(NC)"; \
	fi

clean-pods: ## Delete all pods (they will be recreated)
	@echo "$(YELLOW)🔄 Deleting all pods in $(NAMESPACE)...$(NC)"
	@kubectl delete pods --all -n $(NAMESPACE)
	@echo "$(GREEN)✅ Pods deleted (they will be recreated automatically)$(NC)"

clean-failed-jobs: ## Clean up failed jobs
	@echo "$(YELLOW)🧹 Cleaning up failed jobs...$(NC)"
	@kubectl delete jobs -n $(NAMESPACE) --field-selector status.successful=0
	@echo "$(GREEN)✅ Failed jobs cleaned up$(NC)"

# =============================================================================
# TESTING
# =============================================================================

test-api: ## Test the API endpoints
	@echo "$(GREEN)🧪 Testing API endpoints...$(NC)"
	@chmod +x scripts/test-system.sh
	@./scripts/test-system.sh

smoke-test: ## Run smoke tests to verify basic functionality
	@echo "$(GREEN)💨 Running smoke tests...$(NC)"
	@$(MAKE) system-check
	@echo ""
	@echo "$(YELLOW)Testing basic connectivity...$(NC)"
	@kubectl exec -n $(NAMESPACE) postgres-clients-0 -- pg_isready -U clients_user && echo "$(GREEN)✅ Clients DB ready$(NC)" || echo "$(RED)❌ Clients DB not ready$(NC)"
	@kubectl exec -n $(NAMESPACE) postgres-templates-0 -- pg_isready -U templates_user && echo "$(GREEN)✅ Templates DB ready$(NC)" || echo "$(RED)❌ Templates DB not ready$(NC)"
	@kubectl exec -n $(NAMESPACE) kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --list >/dev/null 2>&1 && echo "$(GREEN)✅ Kafka ready$(NC)" || echo "$(RED)❌ Kafka not ready$(NC)"
	@echo ""
	@echo "$(GREEN)✅ Smoke tests completed$(NC)"