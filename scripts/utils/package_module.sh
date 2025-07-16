#!/bin/bash
#
# package_context.sh - A script to package the relevant files for a specific
#                      microservice, frontend, or infrastructure component into a
#                      single context file for AI assistants.
#
# This script is designed to work with the new agent-managed project structure.
#
# Usage: ./scripts/utils/package_context.sh [-o /path/to/output_dir] [component_name]
# Example: ./scripts/utils/package_context.sh auth-service
# Example: ./scripts/utils/package_context.sh code-all

set -e

# --- Self-locating Logic ---
# Ensures the script can be run from anywhere in the project.
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# Get the full path to this script
SCRIPT_PATH="${SCRIPT_DIR}/$(basename "${BASH_SOURCE[0]}")"
PROJECT_ROOT=$( realpath "$SCRIPT_DIR/../../" )
cd "$PROJECT_ROOT"

# --- Configuration ---
DEFAULT_OUTPUT_DIR=$SCRIPT_DIR"/output_contexts"
ENVIRONMENT="production"
REGION="uk001"

# --- Component List ---
# List of all individual components for the 'all' option
ALL_COMPONENTS=(
    # Horizontal slices
    "code-all"
    "deployments-all"
    "environment-prod"

    # Full service stacks
    "auth-service"
    "core-manager"
    "agent-chassis"
    "reasoning-agent"
    "user-frontend"
    "admin-dashboard"
    "agent-playground"

    # Infrastructure ---
    "infra-cluster-provisioning"   # High-level: Terraform to build the k8s cluster
    "infra-kafka-stack"            # High-level: The entire Kafka setup (TF + Kustomize)
    "infra-terraform-rackspace-module" # Module for creating Rackspace K8s cluster
    "infra-terraform-kafka-modules"  # Modules for Strimzi Operator + Kafka Cluster
    "infra-terraform-environment"    # Top-level environment wiring for Terraform
    "infra-kustomize-kafka-instance" # Kustomize manifests for the Kafka cluster
    "infra-kustomize-ingress"        # Kustomize manifests for NGINX Ingress


    # Frontend development
    "frontend-user-portal-only"
    "frontend-admin-only"
    "frontend-playground-only"
    "frontend-shared-components"
    "frontend-all-apps"

    # Backend APIs
    "api-auth-only"
    "api-core-only"
    "api-agents-only"
    "api-all"

    # Platform libraries
    "platform-core"
    "platform-messaging"
    "platform-data"
    "platform-observability"

    # Agent development
    "agent-framework"
    "agent-reasoning-only"
    "agent-adapters"

    # Database and migrations
    "database-schemas"
    "database-auth"
    "database-clients"

    # Deployment specific
    "deploy-terraform-modules"
    "deploy-kustomize-base"
    "deploy-services"
    "deploy-frontends"

    # Development tools
    "dev-scripts"
    "dev-docker"
    "dev-local-env"

    # Testing
    "test-integration"
    "test-e2e"
    "test-all"
)

# --- Main Functions ---
# Helper function to write a single file's content to the output.
function write_file() {
  local file_path=$1
  local output_file=$2
  local list_only=$3

  if [ -f "$file_path" ]; then
    echo "filepath = ./$file_path" >> "$output_file"
    if [ "$list_only" = "true" ]; then
      echo "[File listed only - content not included]" >> "$output_file"
    else
      cat "$file_path" >> "$output_file"
    fi
    echo "-------------------------------------------------" >> "$output_file"
  fi
}

# Helper function to write all files in a directory to the output.
function write_directory() {
  local dir_path=$1
  local output_file=$2

  # Check if the directory exists before trying to find files in it.
  if [ ! -d "$dir_path" ]; then
    echo "Warning: Directory '$dir_path' not found in '$PWD'. Skipping." >&2
    return
  fi

  # Using find with -print0 and while read is safe for filenames with spaces.
  while IFS= read -r -d $'\0' file; do
    # Check if the file is in a strimzi-yaml* directory
    if [[ "$file" =~ strimzi-yaml[^/]*/[^/]+$ ]]; then
      write_file "$file" "$output_file" "true"
    else
      write_file "$file" "$output_file" "false"
    fi
  done < <(find "$dir_path" -type f \
    -not -path '*/.git/*' \
    -not -path '*/.terraform/*' \
    -not -path '*/.terraform.lock.hcl' \
    -not -path '*/node_modules/*' \
    -not -path '*/dist/*' \
    -not -path '*/build/*' \
    -not -path '*/target/*' \
    -not -path '*/vendor/*' \
    -not -path '*/.idea/*' \
    -not -path '*/.vscode/*' \
    -not -name '*.tfstate' \
    -not -name '*.tfstate.backup' \
    -not -name '*.log' \
    -not -name '*.zip' \
    -not -name '*.tar' \
    -not -name '*.gz' \
    -not -name '*.jar' \
    -not -name '*.war' \
    -not -name '*.exe' \
    -not -name '*.dll' \
    -not -name '*.so' \
    -not -name '*.dylib' \
    -not -name '*.pyc' \
    -not -name '*.pyo' \
    -not -name '__pycache__' \
    -not -name '*.class' \
    -not -name 'go.sum' \
    -not -name 'package-lock.json' \
    -not -name 'yarn.lock' \
    -not -name '*.secret' \
    -not -name '.DS_Store' \
    -not -name 'Thumbs.db' \
    -print0)
}

# --- Script Argument Parsing ---
OUTPUT_DIR=$DEFAULT_OUTPUT_DIR

while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do
  case $1 in
    -o | --output)
      shift
      OUTPUT_DIR=$1
      ;;
  esac
  shift
done

COMPONENT_NAME=$1

# --- Help and Usage ---
function show_help() {
  echo "Usage: $0 [-o /path/to/output_dir] [component_name]"
  echo "Please provide the name of the component to package."
  echo ""
  echo "Available components:"
  echo ""
  echo "  BULK OPERATIONS:"
  echo "    all                      # Package all individual components into separate files"
  echo ""
  echo "  HORIZONTAL SLICES:"
  echo "    code-all                 # All Go source code (cmd, internal, pkg, platform)"
  echo "    deployments-all          # All deployment configurations (Terraform & Kustomize)"
  echo "    environment-prod         # Production environment Terraform configurations"
  echo ""
  echo "  FULL SERVICE STACKS (code + deploy):"
  echo "    auth-service             # Complete auth service stack"
  echo "    core-manager             # Complete core manager stack"
  echo "    agent-chassis            # Complete agent chassis stack"
  echo "    reasoning-agent          # Complete reasoning agent stack"
  echo "    user-frontend            # Complete user portal stack"
  echo "    admin-dashboard          # Complete admin dashboard stack"
  echo "    agent-playground         # Complete agent playground stack"
  echo ""
  echo "  FRONTEND DEVELOPMENT:"
  echo "    frontend-user-portal-only    # Just user portal React code"
  echo "    frontend-admin-only          # Just admin dashboard React code"
  echo "    frontend-playground-only     # Just playground React code"
  echo "    frontend-shared-components   # Shared UI components and API client"
  echo "    frontend-all-apps           # All frontend applications"
  echo ""
  echo "  BACKEND API DEVELOPMENT:"
  echo "    api-auth-only            # Auth service API code only"
  echo "    api-core-only            # Core manager API code only"
  echo "    api-agents-only          # All agent-related API code"
  echo "    api-all                  # All backend API code"
  echo ""
  echo "  PLATFORM LIBRARIES:"
  echo "    platform-core            # Core utilities (config, errors, logging, validation)"
  echo "    platform-messaging       # Kafka and messaging infrastructure"
  echo "    platform-data           # Database, storage, and memory services"
  echo "    platform-observability  # Metrics, health, tracing, resilience"
  echo ""
  echo "  AGENT DEVELOPMENT:"
  echo "    agent-framework         # Agent base classes and orchestration"
  echo "    agent-reasoning-only    # Just reasoning agent implementation"
  echo "    agent-adapters         # Web search and image adapter code"
  echo ""
  echo "  DATABASE & MIGRATIONS:"
  echo "    database-schemas        # All migration files and seed data"
  echo "    database-auth          # Auth-related database code"
  echo "    database-clients       # Client/persona database code"
  echo ""
  echo "  DEPLOYMENT SPECIFIC:"
  echo "    deploy-terraform-modules    # Reusable Terraform modules"
  echo "    deploy-kustomize-base      # Base Kustomize configurations"
  echo "    deploy-services            # Service deployment configs"
  echo "    deploy-frontends          # Frontend deployment configs"
  echo ""
  echo "  DEVELOPMENT TOOLS:"
  echo "    dev-scripts             # All development scripts"
  echo "    dev-docker             # Docker configurations"
  echo "    dev-local-env          # Local development environment"
  echo ""
  echo "  TESTING:"
  echo "    test-integration       # Integration test code"
  echo "    test-e2e              # End-to-end test code"
  echo "    test-all              # All test code"
  echo ""
  echo "  INFRASTRUCTURE:"
  echo "    infra-cluster-provisioning       # Terraform for provisioning the base Rackspace K8s cluster"
  echo "    infra-kafka-stack                # The complete Kafka stack (Terraform modules & Kustomize instance)"
  echo "    infra-terraform-rackspace-module # Module for creating the Rackspace Kubernetes cluster"
  echo "    infra-terraform-kafka-modules    # Modules for Strimzi Operator and the Kafka Cluster"
  echo "    infra-terraform-environment      # Top-level Terraform config for the production environment"
  echo "    infra-kustomize-kafka-instance   # Kustomize definition for the Kafka cluster resource"
  echo "    infra-kustomize-ingress          # Kustomize definition for the NGINX Ingress"

}


if [ -z "$COMPONENT_NAME" ]; then
  show_help
  exit 1
fi

# If the component is 'all', loop and call the script for each component.
if [ "$COMPONENT_NAME" = "all" ]; then
  echo "Packaging all components into separate files..."
  mkdir -p "$OUTPUT_DIR"

  for component in "${ALL_COMPONENTS[@]}"; do
    echo "-------------------------------------------------"
    echo "--> Packaging component: $component"

    # Call the script recursively using its full path
    if [[ -n "$OUTPUT_DIR" && "$OUTPUT_DIR" != "$DEFAULT_OUTPUT_DIR" ]]; then
      bash "$SCRIPT_PATH" -o "$OUTPUT_DIR" "$component"
    else
      bash "$SCRIPT_PATH" "$component"
    fi

    # Display the file size for the component just created
    COMPONENT_FILE="${OUTPUT_DIR}/${component}_context.txt"
    if [ -f "$COMPONENT_FILE" ]; then
      FILE_SIZE=$(du -h "$COMPONENT_FILE" | cut -f1)
      echo "    📦 File size: $FILE_SIZE"
    fi
  done

  echo "-------------------------------------------------"
  echo "✅ All components packaged."
  echo ""
  echo "Summary of generated files:"
  for component in "${ALL_COMPONENTS[@]}"; do
    COMPONENT_FILE="${OUTPUT_DIR}/${component}_context.txt"
    if [ -f "$COMPONENT_FILE" ]; then
      FILE_SIZE=$(du -h "$COMPONENT_FILE" | cut -f1)
      printf "  %-25s %10s\n" "${component}_context.txt" "$FILE_SIZE"
    fi
  done
  exit 0
fi

mkdir -p "$OUTPUT_DIR"
OUTPUT_FILE="${OUTPUT_DIR}/${COMPONENT_NAME}_context.txt"
> "$OUTPUT_FILE"

echo "Packaging component '$COMPONENT_NAME' into $OUTPUT_FILE..."

# --- Component Definitions ---
# Each case defines the specific source code, build, and deployment files
# that make up a complete, independent component.

# Shared files are included where necessary to provide full context.
SHARED_PLATFORM_CODE=("platform/" "pkg/")
SHARED_DEPLOYMENT_MODULES=("deployments/terraform/modules/kustomize-apply/")
SHARED_KUSTOMIZE_BASE=("deployments/kustomize/base/")
SHARED_ROOT_FILES=("Makefile" "go.mod" "go.sum" "docker-compose.yaml")

case "$COMPONENT_NAME" in
  # --- Horizontal Slices ---
  code-all)
    MODULE_DIRS=( "cmd/" "internal/" "pkg/" "platform/" )
    MODULE_FILES=( "go.mod" "go.sum" )
    ;;

  deployments-all)
    MODULE_DIRS=( "deployments/" "build/docker/" )
    MODULE_FILES=( "Makefile" "docker-compose.yaml" )
    ;;

  environment-prod)
    MODULE_DIRS=( "deployments/terraform/environments/$ENVIRONMENT/" )
    MODULE_FILES=( "Makefile" )
    ;;

  # --- Full Service Stacks ---
  auth-service)
    MODULE_DIRS=(
      "cmd/auth-service/" "internal/auth-service/"
      "deployments/kustomize/services/auth-service/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/core-platform/110-auth-service/"
      "${SHARED_PLATFORM_CODE[@]}" "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "build/docker/backend/auth-service.dockerfile" "configs/auth-service.yaml"
      "${SHARED_ROOT_FILES[@]}"
    )
    ;;

  core-manager)
    MODULE_DIRS=(
      "cmd/core-manager/" "internal/core-manager/"
      "deployments/kustomize/services/core-manager/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/core-platform/120-core-manager/"
      "${SHARED_PLATFORM_CODE[@]}" "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "build/docker/backend/core-manager.dockerfile" "configs/core-manager.yaml"
      "${SHARED_ROOT_FILES[@]}"
    )
    ;;

  agent-chassis)
    MODULE_DIRS=(
      "cmd/agent-chassis/" "platform/agentbase/"
      "deployments/kustomize/services/agent-chassis/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2210-agent-chassis/"
      "${SHARED_PLATFORM_CODE[@]}" "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "build/docker/backend/agent-chassis.dockerfile" "configs/agent-chassis.yaml"
      "${SHARED_ROOT_FILES[@]}"
    )
    ;;

  reasoning-agent)
    MODULE_DIRS=(
      "cmd/reasoning-agent/" "internal/agents/reasoning/"
      "deployments/kustomize/services/reasoning-agent/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2220-reasoning-agent/"
      "${SHARED_PLATFORM_CODE[@]}" "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "build/docker/backend/reasoning-agent.dockerfile" "configs/reasoning-agent.yaml"
      "${SHARED_ROOT_FILES[@]}"
    )
    ;;

  # --- Frontend Applications (Full Stack) ---
  user-frontend)
    MODULE_DIRS=(
      "frontends/user-portal/"
      "deployments/kustomize/frontends/user-portal/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/frontends/3320-user-portal/"
      "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=( "Makefile" )
    ;;

  admin-dashboard)
    MODULE_DIRS=(
      "frontends/admin-dashboard/"
      "deployments/kustomize/frontends/admin-dashboard/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/frontends/3310-admin-dashboard/"
      "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=( "Makefile" )
    ;;

  agent-playground)
    MODULE_DIRS=(
      "frontends/agent-playground/"
      "deployments/kustomize/frontends/agent-playground/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/frontends/3330-agent-playground/"
      "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=( "Makefile" )
    ;;

  # --- Infrastructure Layers ---
  # --- Infrastructure Layers (Refined and Expanded) ---
  infra-cluster-provisioning) # Renamed from infra-cluster
    MODULE_DIRS=(
      "deployments/terraform/modules/rackspace-kubernetes/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/010-infrastructure/"
    )
    MODULE_FILES=("Makefile")
    ;;

  infra-kafka-stack) # New comprehensive stack
    MODULE_DIRS=(
      "deployments/terraform/modules/strimzi-operator/"
      "deployments/terraform/modules/kafka-cluster/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/030-strimzi-operator/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/040-kafka-cluster/"
      "deployments/kustomize/infrastructure/kafka/"
    )
    MODULE_FILES=("Makefile")
    ;;

  infra-terraform-rackspace-module)
    MODULE_DIRS=( "deployments/terraform/modules/rackspace-kubernetes/" )
    ;;

  infra-terraform-kafka-modules)
    MODULE_DIRS=(
      "deployments/terraform/modules/strimzi-operator/"
      "deployments/terraform/modules/kafka-cluster/"
    )
    ;;

  infra-terraform-environment)
    MODULE_DIRS=( "deployments/terraform/environments/$ENVIRONMENT/" )
    ;;

  infra-kustomize-kafka-instance)
    MODULE_DIRS=( "deployments/kustomize/infrastructure/kafka/" )
    ;;

  infra-kustomize-ingress)
    MODULE_DIRS=( "deployments/kustomize/infrastructure/nginx-ingress/" )
    ;;

  # --- Frontend Development Only ---
  frontend-user-portal-only)
    MODULE_DIRS=( "frontends/user-portal/" )
    MODULE_FILES=( "build/docker/frontend/react-nginx.dockerfile" )
    ;;

  frontend-admin-only)
    MODULE_DIRS=( "frontends/admin-dashboard/" )
    MODULE_FILES=( "build/docker/frontend/react-nginx.dockerfile" )
    ;;

  frontend-playground-only)
    MODULE_DIRS=( "frontends/agent-playground/" )
    MODULE_FILES=( "build/docker/frontend/react-nginx.dockerfile" )
    ;;

  frontend-shared-components)
    MODULE_DIRS=( "frontends/shared/" )
    ;;

  frontend-all-apps)
    MODULE_DIRS=( "frontends/" )
    MODULE_FILES=( "build/docker/frontend/react-nginx.dockerfile" "build/docker/frontend/react-dev.dockerfile" )
    ;;

  # --- Backend API Development ---
  api-auth-only)
    MODULE_DIRS=( "cmd/auth-service/" "internal/auth-service/" )
    MODULE_FILES=( "configs/auth-service.yaml" "go.mod" )
    ;;

  api-core-only)
    MODULE_DIRS=( "cmd/core-manager/" "internal/core-manager/" )
    MODULE_FILES=( "configs/core-manager.yaml" "go.mod" )
    ;;

  api-agents-only)
    MODULE_DIRS=(
      "cmd/agent-chassis/" "cmd/reasoning-agent/"
      "internal/agents/" "platform/agentbase/"
    )
    MODULE_FILES=( "configs/agent-chassis.yaml" "configs/reasoning-agent.yaml" "go.mod" )
    ;;

  api-all)
    MODULE_DIRS=( "cmd/" "internal/" "pkg/models/" )
    MODULE_FILES=( "configs/" "go.mod" )
    ;;

  # --- Platform Libraries ---
  platform-core)
    MODULE_DIRS=(
      "platform/config/" "platform/errors/" "platform/logger/"
      "platform/validation/" "platform/contracts/"
    )
    MODULE_FILES=( "go.mod" )
    ;;

  platform-messaging)
    MODULE_DIRS=(
      "platform/kafka/" "platform/messaging/"
    )
    MODULE_FILES=( "go.mod" )
    ;;

  platform-data)
    MODULE_DIRS=(
      "platform/database/" "platform/storage/" "platform/memory/"
    )
    MODULE_FILES=( "go.mod" )
    ;;

  platform-observability)
    MODULE_DIRS=(
      "platform/observability/" "platform/health/" "platform/resilience/"
    )
    MODULE_FILES=( "go.mod" )
    ;;

  # --- Agent Development ---
  agent-framework)
    MODULE_DIRS=(
      "platform/agentbase/" "platform/orchestration/"
      "platform/aiservice/" "platform/governance/"
    )
    MODULE_FILES=( "go.mod" )
    ;;

  agent-reasoning-only)
    MODULE_DIRS=(
      "cmd/reasoning-agent/" "internal/agents/reasoning/"
    )
    MODULE_FILES=( "configs/reasoning-agent.yaml" "go.mod" )
    ;;

  agent-adapters)
    MODULE_DIRS=(
      "cmd/web-search-adapter/" "cmd/image-generator-adapter/"
      "internal/adapters/"
    )
    MODULE_FILES=( "configs/web-search-adapter.yaml" "go.mod" )
    ;;

  # --- Database and Migrations ---
  database-schemas)
    MODULE_DIRS=( "platform/database/migrations/" )
    MODULE_FILES=( "scripts/docker/seed-data.sql" )
    ;;

  database-auth)
    MODULE_DIRS=(
      "internal/auth-service/user/" "internal/auth-service/project/"
      "internal/auth-service/subscription/" "platform/database/migrations/"
    )
    ;;

  database-clients)
    MODULE_DIRS=(
      "internal/core-manager/database/" "platform/database/"
    )
    MODULE_FILES=( "platform/database/migrations/003_create_client_schema.sql" )
    ;;

  # --- Deployment Specific ---
  deploy-terraform-modules)
    MODULE_DIRS=( "deployments/terraform/modules/" )
    ;;

  deploy-kustomize-base)
    MODULE_DIRS=(
      "deployments/kustomize/base/" "deployments/kustomize/components/"
    )
    ;;

  deploy-services)
    MODULE_DIRS=(
      "deployments/kustomize/services/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/"
    )
    ;;

  deploy-frontends)
    MODULE_DIRS=(
      "deployments/kustomize/frontends/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/frontends/"
    )
    ;;

  # --- Development Tools ---
  dev-scripts)
    MODULE_DIRS=( "scripts/" )
    MODULE_FILES=( "Makefile" )
    ;;

  dev-docker)
    MODULE_DIRS=( "build/docker/" )
    MODULE_FILES=( "docker-compose.yaml" ".env.example" )
    ;;

  dev-local-env)
    MODULE_DIRS=( "scripts/local/" "scripts/docker/" )
    MODULE_FILES=(
      "docker-compose.yaml" ".env.example"
      "Makefile" "scripts/setup.sh"
    )
    ;;

  # --- Testing ---
  test-integration)
    MODULE_DIRS=( "tests/integration/" )
    MODULE_FILES=( "go.mod" )
    ;;

  test-e2e)
    MODULE_DIRS=( "tests/e2e/" )
    MODULE_FILES=( "go.mod" )
    ;;

  test-all)
    MODULE_DIRS=( "tests/" )
    MODULE_FILES=( "go.mod" "Makefile" )
    ;;

  *)
    echo "Error: Unknown component '$COMPONENT_NAME'."
    show_help
    exit 1
    ;;
esac

# --- Packaging Logic ---
# This ensures that directories are processed before loose files.
for dir in "${MODULE_DIRS[@]}"; do
  write_directory "$dir" "$OUTPUT_FILE"
done
for file in "${MODULE_FILES[@]}"; do
  write_file "$file" "$OUTPUT_FILE" "false"
done

echo "✅ Done. Component context saved to $OUTPUT_FILE"
FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
echo "📦 File size: $FILE_SIZE"