# AI Persona System - Setup Scripts

## 1. Makefile

```makefile
# FILE: Makefile
.PHONY: all build push deploy clean test dev setup

REGISTRY ?= ai-persona-system
VERSION ?= latest

# Services
SERVICES = auth-service core-manager agent-chassis reasoning-agent image-generator-adapter web-search-adapter

# Default target
all: build deploy

# Build all Docker images
build:
	@echo "🔨 Building Docker images..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		docker build -f cmd/$$service/Dockerfile -t $(REGISTRY)/$$service:$(VERSION) . || exit 1; \
	done
	@echo "✅ All images built successfully"

# Push all images to registry
push:
	@echo "📤 Pushing images to registry..."
	@for service in $(SERVICES); do \
		echo "Pushing $$service..."; \
		docker push $(REGISTRY)/$$service:$(VERSION) || exit 1; \
	done
	@echo "✅ All images pushed successfully"

# Deploy to Kubernetes
deploy:
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/configmap-common.yaml
	kubectl apply -f k8s/postgres-clients.yaml
	kubectl apply -f k8s/postgres-templates.yaml
	kubectl apply -f k8s/mysql-auth.yaml
	kubectl apply -f k8s/kafka.yaml
	kubectl apply -f k8s/minio.yaml
	@echo "⏳ Waiting for infrastructure to be ready..."
	kubectl wait --for=condition=ready pod -l app=postgres-clients -n ai-persona-system --timeout=300s
	kubectl wait --for=condition=ready pod -l app=postgres-templates -n ai-persona-system --timeout=300s
	kubectl wait --for=condition=ready pod -l app=mysql-auth -n ai-persona-system --timeout=300s
	kubectl apply -f k8s/auth-service.yaml
	kubectl apply -f k8s/core-manager.yaml
	kubectl apply -f k8s/agent-chassis.yaml
	kubectl apply -f k8s/reasoning-agent.yaml
	kubectl apply -f k8s/image-generator-adapter.yaml
	kubectl apply -f k8s/monitoring/
	@echo "✅ Deployment complete"

# Initial setup (creates secrets and runs migrations)
setup:
	@echo "🚀 Running initial setup..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

# Run tests
test:
	@echo "🧪 Running unit tests..."
	go test -v ./...

# Run integration tests
test-integration:
	@echo "🧪 Running integration tests..."
	go test -v -tags=integration ./tests/integration/

# Local development with docker-compose
dev:
	@echo "🐳 Starting local development environment..."
	docker-compose up -d
	@echo "✅ Local environment started"
	@echo "Services available at:"
	@echo "  - Auth Service: http://localhost:8081"
	@echo "  - Core Manager: http://localhost:8088"

# Clean up everything
clean:
	@echo "🧹 Cleaning up..."
	docker-compose down -v || true
	kubectl delete namespace ai-persona-system --ignore-not-found=true || true
	@echo "✅ Cleanup complete"

# Database migrations
migrate-up:
	@echo "📝 Running database migrations..."
	@kubectl cp platform/database/migrations/001_enable_pgvector.sql ai-persona-system/postgres-clients-0:/tmp/
	@kubectl cp platform/database/migrations/002_create_templates_schema.sql ai-persona-system/postgres-templates-0:/tmp/
	@kubectl cp platform/database/migrations/003_create_client_schema.sql ai-persona-system/postgres-clients-0:/tmp/
	@kubectl cp platform/database/migrations/004_auth_schema.sql ai-persona-system/mysql-auth-0:/tmp/
	@kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -f /tmp/001_enable_pgvector.sql
	@kubectl exec -n ai-persona-system postgres-templates-0 -- psql -U templates_user -d templates_db -f /tmp/002_create_templates_schema.sql
	@kubectl exec -n ai-persona-system mysql-auth-0 -- mysql -u auth_user -p$$AUTH_DB_PASSWORD auth_db < /tmp/004_auth_schema.sql
	@echo "✅ Migrations complete"

# Create a new client schema
create-client:
	@read -p "Enter client ID: " client_id; \
	kubectl exec -n ai-persona-system postgres-clients-0 -- \
		psql -U clients_user -d clients_db -c "CREATE SCHEMA IF NOT EXISTS client_$$client_id"; \
	sed "s/{client_id}/$$client_id/g" platform/database/migrations/003_create_client_schema.sql | \
	kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
		psql -U clients_user -d clients_db

# Port forwarding for local development
port-forward:
	@echo "🔌 Setting up port forwarding..."
	@pkill -f "kubectl port-forward" || true
	kubectl port-forward -n ai-persona-system svc/auth-service 8081:8081 &
	kubectl port-forward -n ai-persona-system svc/core-manager 8088:8088 &
	kubectl port-forward -n ai-persona-system svc/grafana 3000:3000 &
	kubectl port-forward -n ai-persona-system svc/prometheus 9090:9090 &
	kubectl port-forward -n ai-persona-system svc/minio 9001:9001 &
	@echo "✅ Port forwarding active"
	@echo "Services available at:"
	@echo "  - Auth API: http://localhost:8081"
	@echo "  - Core API: http://localhost:8088"
	@echo "  - Grafana: http://localhost:3000"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - MinIO Console: http://localhost:9001"

# Stop port forwarding
stop-port-forward:
	@pkill -f "kubectl port-forward" || true
	@echo "✅ Port forwarding stopped"

# View logs for a service
logs:
	@read -p "Enter service name (e.g., auth-service, core-manager, agent-chassis): " service; \
	kubectl logs -n ai-persona-system -l app=$$service -f --tail=100

# Scale a deployment
scale:
	@read -p "Enter deployment name: " deployment; \
	read -p "Enter replica count: " replicas; \
	kubectl scale deployment -n ai-persona-system $$deployment --replicas=$$replicas

# Check system status
status:
	@echo "📊 System Status:"
	@echo "\n🏃 Deployments:"
	@kubectl get deployments -n ai-persona-system
	@echo "\n📦 Pods:"
	@kubectl get pods -n ai-persona-system
	@echo "\n🔌 Services:"
	@kubectl get services -n ai-persona-system
	@echo "\n💾 Persistent Volumes:"
	@kubectl get pvc -n ai-persona-system

# Create initial template data
seed-data:
	@echo "🌱 Seeding initial data..."
	@kubectl exec -i -n ai-persona-system postgres-templates-0 -- psql -U templates_user -d templates_db <<EOF
	INSERT INTO persona_templates (id, name, description, category, config, is_active) VALUES
	('00000000-0000-0000-0000-000000000001', 'Content Writer', 'General purpose content writing agent', 'writing',
	 '{"model": "claude-3-opus", "temperature": 0.7, "skills": ["blog_writing", "copywriting", "editing"]}', true),
	('00000000-0000-0000-0000-000000000002', 'Code Assistant', 'Programming and code review agent', 'technical',
	 '{"model": "claude-3-opus", "temperature": 0.2, "skills": ["code_generation", "debugging", "review"]}', true),
	('00000000-0000-0000-0000-000000000003', 'Research Analyst', 'Research and data analysis agent', 'research',
	 '{"model": "claude-3-opus", "temperature": 0.3, "skills": ["web_research", "data_analysis", "summarization"]}', true);
	EOF
	@echo "✅ Initial data seeded"

# Quick start - runs everything needed to get started
quickstart: setup build deploy migrate-up seed-data port-forward
	@echo "🎉 AI Persona System is ready!"
	@echo "Default services are now accessible - see above for URLs"

# Help
help:
	@echo "AI Persona System - Makefile Commands"
	@echo ""
	@echo "Setup & Deployment:"
	@echo "  make setup          - Initial setup (creates secrets, configures cluster)"
	@echo "  make build          - Build all Docker images"
	@echo "  make deploy         - Deploy to Kubernetes"
	@echo "  make quickstart     - Complete setup from scratch"
	@echo ""
	@echo "Development:"
	@echo "  make dev            - Start local development environment"
	@echo "  make test           - Run unit tests"
	@echo "  make port-forward   - Set up local port forwarding"
	@echo "  make logs           - View service logs"
	@echo ""
	@echo "Operations:"
	@echo "  make status         - Check system status"
	@echo "  make scale          - Scale a deployment"
	@echo "  make create-client  - Create a new client schema"
	@echo "  make seed-data      - Add initial template data"