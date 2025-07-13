#!/bin/bash
# FILE: scripts/deploy-system.sh
# Complete deployment script that orchestrates the entire system deployment

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
NAMESPACE="ai-persona-system"
TIMEOUT="300s"
DOCKER_REGISTRY="ai-persona-system"

# Function to print colored output
print_header() {
    echo -e "${BLUE}=============================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=============================================${NC}"
}

print_step() {
    echo -e "${YELLOW}🔧 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to wait for pods
wait_for_pods() {
    local label=$1
    local description=$2
    local timeout=${3:-300s}

    print_step "Waiting for $description to be ready..."
    if kubectl wait --for=condition=ready pod -l "$label" -n "$NAMESPACE" --timeout="$timeout"; then
        print_success "$description is ready"
    else
        print_error "$description failed to become ready"
        return 1
    fi
}

# Function to wait for jobs to complete
wait_for_job() {
    local job_name=$1
    local description=$2
    local timeout=${3:-600}

    print_step "Waiting for $description to complete..."
    local attempt=1
    local max_attempts=$((timeout / 10))

    while [ $attempt -le $max_attempts ]; do
        local status=$(kubectl get job "$job_name" -n "$NAMESPACE" -o jsonpath='{.status.conditions[0].type}' 2>/dev/null || echo "NotFound")

        if [ "$status" = "Complete" ]; then
            print_success "$description completed successfully"
            return 0
        elif [ "$status" = "Failed" ]; then
            print_error "$description failed"
            kubectl logs -n "$NAMESPACE" job/"$job_name" --tail=50
            return 1
        fi

        echo -e "${YELLOW}⏳ Attempt $attempt/$max_attempts - $description still running...${NC}"
        sleep 10
        ((attempt++))
    done

    print_error "$description timed out"
    return 1
}

# Check prerequisites
check_prerequisites() {
    print_header "CHECKING PREREQUISITES"

    print_step "Checking required tools..."
    for tool in kubectl docker; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            print_error "$tool is required but not installed"
            exit 1
        fi
    done

    print_step "Checking Kubernetes connection..."
    if ! kubectl cluster-info >/dev/null 2>&1; then
        print_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    print_success "Prerequisites check passed"
}

# Phase 1: Infrastructure
deploy_infrastructure() {
    print_header "PHASE 1: DEPLOYING INFRASTRUCTURE"

    print_step "Creating namespace..."
    kubectl apply -f k8s/namespace.yaml

    print_step "Waiting for namespace to be active..."
    kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/"$NAMESPACE" --timeout=60s

    print_step "Applying RBAC and security policies..."
    kubectl apply -f k8s/rbac-security.yaml

    print_step "Creating ConfigMaps..."
    kubectl apply -f k8s/configmap-common.yaml

    print_success "Infrastructure deployed"
}

# Phase 2: Storage
deploy_storage() {
    print_header "PHASE 2: DEPLOYING STORAGE SYSTEMS"

    print_step "Deploying PostgreSQL databases..."
    kubectl apply -f k8s/postgres-clients.yaml
    kubectl apply -f k8s/postgres-templates.yaml

    print_step "Deploying MySQL database..."
    kubectl apply -f k8s/mysql-auth.yaml

    print_step "Deploying MinIO object storage..."
    kubectl apply -f k8s/minio.yaml

    print_step "Deploying backup storage..."
    kubectl apply -f k8s/backup-cronjob.yaml

    print_success "Storage systems deployed"
}

# Phase 3: Messaging
deploy_messaging() {
    print_header "PHASE 3: DEPLOYING MESSAGING SYSTEM"

    print_step "Deploying Kafka cluster..."
    kubectl apply -f k8s/kafka.yaml

    print_success "Messaging system deployed"
}

# Phase 4: Wait for infrastructure
wait_for_infrastructure() {
    print_header "PHASE 4: WAITING FOR INFRASTRUCTURE"

    wait_for_pods "app=postgres-clients" "PostgreSQL Clients"
    wait_for_pods "app=postgres-templates" "PostgreSQL Templates"
    wait_for_pods "app=mysql-auth" "MySQL Auth"
    wait_for_pods "app=minio" "MinIO"
    wait_for_pods "app=kafka" "Kafka Cluster" "600s"

    print_success "All infrastructure components are ready"
}

# Phase 5: Initialize system
initialize_system() {
    print_header "PHASE 5: INITIALIZING SYSTEM"

    print_step "Running database migrations..."
    kubectl apply -f k8s/jobs/database-init-job.yaml
    wait_for_job "database-init" "Database Migrations" 600

    print_step "Creating Kafka topics..."
    kubectl apply -f k8s/jobs/kafka-topics-job.yaml
    wait_for_job "kafka-topics-init" "Kafka Topics Creation" 300

    print_step "Seeding initial data..."
    wait_for_job "data-seeder" "Data Seeding" 300

    print_success "System initialization completed"
}

# Phase 6: Core services
deploy_core_services() {
    print_header "PHASE 6: DEPLOYING CORE SERVICES"

    print_step "Deploying Auth Service..."
    kubectl apply -f k8s/auth-service.yaml

    print_step "Deploying Core Manager..."
    kubectl apply -f k8s/core-manager.yaml

    wait_for_pods "app=auth-service" "Auth Service"
    wait_for_pods "app=core-manager" "Core Manager"

    print_success "Core services deployed"
}

# Phase 7: Agents
deploy_agents() {
    print_header "PHASE 7: DEPLOYING AGENT SERVICES"

    print_step "Deploying Agent Chassis..."
    kubectl apply -f k8s/agent-chassis.yaml

    print_step "Deploying Reasoning Agent..."
    kubectl apply -f k8s/reasoning-agent.yaml

    print_step "Deploying Image Generator Adapter..."
    kubectl apply -f k8s/image-generator-adapter.yaml

    print_step "Deploying Web Search Adapter..."
    kubectl apply -f k8s/web-search-adapter.yaml

    wait_for_pods "app=agent-chassis" "Agent Chassis" "600s"
    wait_for_pods "app=reasoning-agent" "Reasoning Agent"
    wait_for_pods "app=image-generator-adapter" "Image Generator"
    wait_for_pods "app=web-search-adapter" "Web Search"

    print_success "Agent services deployed"
}

# Phase 8: Monitoring and ingress
deploy_monitoring_ingress() {
    print_header "PHASE 8: DEPLOYING MONITORING AND INGRESS"

    print_step "Deploying monitoring stack..."
    kubectl apply -f k8s/monitoring/

    print_step "Deploying ingress..."
    kubectl apply -f k8s/ingress.yaml

    wait_for_pods "app=prometheus" "Prometheus" 300s
    wait_for_pods "app=grafana" "Grafana" 300s

    print_success "Monitoring and ingress deployed"
}

# System verification
verify_system() {
    print_header "SYSTEM VERIFICATION"

    print_step "Checking system status..."
    kubectl get pods -n "$NAMESPACE"

    print_step "Checking services..."
    kubectl get services -n "$NAMESPACE"

    print_step "Verifying Kafka topics..."
    kubectl exec -n "$NAMESPACE" kafka-0 -- kafka-topics --bootstrap-server localhost:9092 --list | head -10

    print_step "Checking database tables..."
    kubectl exec -n "$NAMESPACE" postgres-templates-0 -- psql -U templates_user -d templates_db -c "\dt" >/dev/null 2>&1 && \
        echo -e "${GREEN}✅ Templates database OK${NC}" || echo -e "${RED}❌ Templates database issue${NC}"

    kubectl exec -n "$NAMESPACE" postgres-clients-0 -- psql -U clients_user -d clients_db -c "\dt" >/dev/null 2>&1 && \
        echo -e "${GREEN}✅ Clients database OK${NC}" || echo -e "${RED}❌ Clients database issue${NC}"

    print_success "System verification completed"
}

# Cleanup function
cleanup_failed_deployment() {
    print_error "Deployment failed. Cleaning up failed jobs..."
    kubectl delete jobs --field-selector status.successful=0 -n "$NAMESPACE" 2>/dev/null || true
}

# Main execution
main() {
    print_header "AI PERSONA SYSTEM DEPLOYMENT"
    echo -e "${BLUE}Starting complete system deployment...${NC}"
    echo ""

    # Set up error handling
    trap cleanup_failed_deployment ERR

    # Execute deployment phases
    check_prerequisites
    deploy_infrastructure
    deploy_storage
    deploy_messaging
    wait_for_infrastructure
    initialize_system
    deploy_core_services
    deploy_agents
    deploy_monitoring_ingress
    verify_system

    print_header "DEPLOYMENT COMPLETED SUCCESSFULLY!"
    echo ""
    echo -e "${GREEN}🎉 AI Persona System is now running!${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "1. Set up port forwarding: ${BLUE}make port-forward${NC}"
    echo -e "2. Create your first client: ${BLUE}make create-client-schema CLIENT_ID=demo_client${NC}"
    echo -e "3. Test the system: ${BLUE}make test-api${NC}"
    echo ""
    echo -e "${YELLOW}Access URLs (after port forwarding):${NC}"
    echo -e "- Auth API: ${BLUE}http://localhost:8081${NC}"
    echo -e "- Core API: ${BLUE}http://localhost:8088${NC}"
    echo -e "- Grafana: ${BLUE}http://localhost:3000${NC} (admin/admin)"
    echo -e "- Kafka UI: ${BLUE}http://localhost:8080${NC}"
    echo ""
}

# Execute main function
main "$@"