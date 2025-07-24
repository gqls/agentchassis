#!/bin/bash
# FILE: scripts/setup.sh
# Initial setup script for AI Persona System

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    echo "🔍 Checking prerequisites..."

    local missing_deps=()

    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    fi

    # Check docker
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi

    # Check if Kubernetes is accessible
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster. Please ensure kubectl is configured."
        exit 1
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        echo "Please install missing dependencies and try again."
        exit 1
    fi

    print_success "All prerequisites met"
}

# Create namespace
create_namespace() {
    print_info "Creating Kubernetes namespace..."
    kubectl apply -f k8s/namespace.yaml || {
        print_error "Failed to create namespace"
        exit 1
    }
    print_success "Namespace created"
}

# Generate random passwords
generate_password() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-25
}

# Create secrets
create_secrets() {
    print_info "Creating secrets..."

    # Database secrets
    kubectl create secret generic db-secrets -n ai-persona-system \
        --from-literal=clients-db-password=$(generate_password) \
        --from-literal=templates-db-password=$(generate_password) \
        --from-literal=auth-db-password=$(generate_password) \
        --from-literal=mysql-root-password=$(generate_password) \
        --dry-run=client -o yaml | kubectl apply -f - || {
        print_error "Failed to create database secrets"
        exit 1
    }

    # MinIO secrets
    kubectl create secret generic minio-secrets -n ai-persona-system \
        --from-literal=access-key=$(generate_password) \
        --from-literal=secret-key=$(generate_password) \
        --dry-run=client -o yaml | kubectl apply -f - || {
        print_error "Failed to create MinIO secrets"
        exit 1
    }

    # Auth secrets
    kubectl create secret generic auth-secrets -n ai-persona-system \
        --from-literal=jwt-secret=$(generate_password) \
        --dry-run=client -o yaml | kubectl apply -f - || {
        print_error "Failed to create auth secrets"
        exit 1
    }

    # Grafana secrets
    kubectl create secret generic grafana-secrets -n ai-persona-system \
        --from-literal=admin-password=$(generate_password) \
        --dry-run=client -o yaml | kubectl apply -f - || {
        print_error "Failed to create Grafana secrets"
        exit 1
    }

    print_success "Basic secrets created"
}

# Get API keys from user
get_api_keys() {
    print_info "Setting up AI service API keys..."
    echo "Please provide your API keys (or press Enter to skip):"

    # Anthropic
    read -p "Anthropic API Key (for Claude): " anthropic_key
    if [ -z "$anthropic_key" ]; then
        anthropic_key="dummy-key-replace-me"
        print_info "Skipping Anthropic API key - reasoning agent will not work"
    fi

    # Stability AI
    read -p "Stability AI API Key (for image generation): " stability_key
    if [ -z "$stability_key" ]; then
        stability_key="dummy-key-replace-me"
        print_info "Skipping Stability API key - image generation will not work"
    fi

    # SERP API
    read -p "SERP API Key (for web search): " serp_key
    if [ -z "$serp_key" ]; then
        serp_key="dummy-key-replace-me"
        print_info "Skipping SERP API key - web search will not work"
    fi

    # Create AI secrets
    kubectl create secret generic ai-secrets -n ai-persona-system \
        --from-literal=anthropic-api-key="$anthropic_key" \
        --from-literal=stability-api-key="$stability_key" \
        --from-literal=scraping-bee-key="$serp_key" \
        --dry-run=client -o yaml | kubectl apply -f - || {
        print_error "Failed to create AI secrets"
        exit 1
    }

    print_success "AI service secrets created"
}

# Create service account and RBAC
create_rbac() {
    print_info "Setting up RBAC..."

    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
  namespace: ai-persona-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
- apiGroups: [""]
  resources:
  - nodes
  - nodes/proxy
  - services
  - endpoints
  - pods
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
- kind: ServiceAccount
  name: prometheus
  namespace: ai-persona-system
EOF

    print_success "RBAC configured"
}

# Main setup flow
main() {
    echo "🚀 AI Persona System Setup"
    echo "========================="
    echo ""

    # Run checks
    check_prerequisites

    # Create namespace
    create_namespace

    # Create secrets
    create_secrets

    # Get API keys
    get_api_keys

    # Setup RBAC
    create_rbac

    echo ""
    print_success "Setup complete!"
    echo ""
    echo "Next steps:"
    echo "1. Build Docker images:    make build"
    echo "2. Deploy to Kubernetes:   make deploy"
    echo "3. Run migrations:         make migrate-up"
    echo "4. Seed initial data:      make seed-data"
    echo "5. Set up port forward:    make port-forward"
    echo ""
    echo "Or run everything at once: make quickstart"
    echo ""
}

# Run main function
main