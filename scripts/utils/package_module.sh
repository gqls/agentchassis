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
# Example: ./scripts/utils/package_context.sh user-frontend

set -e

# --- Self-locating Logic ---
# Ensures the script can be run from anywhere in the project.
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT=$( realpath "$SCRIPT_DIR/../../" )
cd "$PROJECT_ROOT"

# --- Configuration ---
DEFAULT_OUTPUT_DIR="./output_contexts"
ENVIRONMENT="production"
REGION="uk001"

# --- Main Functions ---
# Helper function to write a single file's content to the output.
function write_file() {
  local file_path=$1
  local output_file=$2
  if [ -f "$file_path" ]; then
    echo "filepath = ./$file_path" >> "$output_file"
    cat "$file_path" >> "$output_file"
    echo "-------------------------------------------------" >> "$output_file"
  fi
}

# Helper function to write all files in a directory to the output.
function write_directory() {
  local dir_path=$1
  local output_file=$2
  if [ -d "$dir_path" ]; then
    # Using find with -print0 and while read is safe for filenames with spaces.
    while IFS= read -r -d $'\0' file; do
      write_file "$file" "$output_file"
    done < <(find "$dir_path" -type f -print0)
  fi
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
  echo "  - all                     # All project files"
  echo "  - platform                # Shared Go platform code"
  echo "  - deployment-all          # All deployment configurations (Terraform & Kustomize)"
  echo ""
  echo "  # Backend Services"
  echo "  - auth-service"
  echo "  - core-manager"
  echo "  - agent-chassis"
  echo "  - reasoning-agent"
  echo ""
  echo "  # Frontend Applications"
  echo "  - user-frontend"
  echo "  - admin-dashboard"
  echo ""
  echo "  # Infrastructure Layers"
  echo "  - infra-cluster           # The core Rackspace Kubernetes cluster"
  echo "  - infra-kafka             # The Kafka cluster deployment"
}


if [ -z "$COMPONENT_NAME" ]; then
  show_help
  exit 1
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
  all)
    MODULE_DIRS=(
      "cmd/" "configs/" "internal/" "pkg/" "platform/" "build/"
      "deployments/" "scripts/" "frontends/" "docs/"
    )
    MODULE_FILES=("${SHARED_ROOT_FILES[@]}")
    ;;

  platform)
    MODULE_DIRS=("platform/" "pkg/")
    MODULE_FILES=()
    ;;

  deployment-all)
    MODULE_DIRS=("deployments/" "build/docker/")
    MODULE_FILES=("Makefile")
    ;;

  # --- Backend Services ---
  auth-service)
    MODULE_DIRS=(
      "cmd/auth-service/" "internal/auth-service/"
      "deployments/kustomize/services/auth-service/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/core-platform/110-auth-service/"
      "${SHARED_PLATFORM_CODE[@]}" "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "build/docker/auth-service.dockerfile" "configs/auth-service.yaml"
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
      "build/docker/core-manager.dockerfile" "configs/core-manager.yaml"
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
      "build/docker/reasoning-agent.dockerfile" "configs/reasoning-agent.yaml"
      "${SHARED_ROOT_FILES[@]}"
    )
    ;;

  # --- Frontend Applications ---
  user-frontend)
    MODULE_DIRS=(
      "frontends/user-portal/" # Assuming user-portal is the main user-frontend
      "deployments/kustomize/frontends/user-portal/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/frontends/3320-user-portal/"
      "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "Makefile"
    )
    ;;

  admin-dashboard)
    MODULE_DIRS=(
      "frontends/admin-dashboard/"
      "deployments/kustomize/frontends/admin-dashboard/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/frontends/3310-admin-dashboard/"
      "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "Makefile"
    )
    ;;

  # --- Infrastructure Layers ---
  infra-cluster)
    MODULE_DIRS=(
      "deployments/terraform/modules/rackspace-kubernetes/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/010-infrastructure/"
    )
    MODULE_FILES=("Makefile")
    ;;

  infra-kafka)
    MODULE_DIRS=(
      "deployments/terraform/modules/strimzi-operator/"
      "deployments/terraform/modules/kafka-cluster/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/030-strimzi-operator/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/040-kafka-cluster/"
      "deployments/kustomize/infrastructure/kafka/"
    )
    MODULE_FILES=("Makefile")
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
  write_file "$file" "$OUTPUT_FILE"
done

echo "✅ Done. Component context saved to $OUTPUT_FILE"
FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
echo "📦 File size: $FILE_SIZE"
