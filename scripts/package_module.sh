#!/bin/bash
#
# package_module.sh - A script to package the relevant files for a specific
#                     microservice or module into a single context file for AI assistants.
#
# Usage: ./scripts/package_module.sh [-o /path/to/output_dir] [module_name]
# Example: ./scripts/package_module.sh all

set -e

# --- Self-locating Logic ---
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT=$( dirname "$SCRIPT_DIR" )
cd "$PROJECT_ROOT"

# --- Configuration ---
DEFAULT_OUTPUT_DIR="./output"

# --- Main Functions ---
function write_file() {
  local file_path=$1
  local output_file=$2
  if [ -f "$file_path" ]; then
    echo "filepath = ./$file_path" >> "$output_file"
    cat "$file_path" >> "$output_file"
    echo "-------------------------------------------------" >> "$output_file"
  fi
}

function write_directory() {
  local dir_path=$1
  local output_file=$2
  if [ -d "$dir_path" ]; then
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

MODULE_NAME=$1

if [ -z "$MODULE_NAME" ]; then
  echo "Usage: $0 [-o /path/to/output_dir] [module_name]"
  echo "Please provide the name of the module to package."
  echo ""
  echo "Available modules:"
  echo "  - all" # <<< CHANGE: Added 'all' to the list
  echo "  - platform"
  echo "  - auth-service"
  echo "  - core-manager"
  echo "  - agent-chassis"
  echo "  - reasoning-agent"
  echo "  - image-generator-adapter"
  echo "  - web-search-adapter"
  echo "  - deployment"
  exit 1
fi

mkdir -p "$OUTPUT_DIR"
OUTPUT_FILE="${OUTPUT_DIR}/${MODULE_NAME}_context.txt"
> "$OUTPUT_FILE"

echo "Packaging module '$MODULE_NAME' into $OUTPUT_FILE..."

# Core shared files (almost always needed)
SHARED_DIRS=("platform/")
SHARED_FILES=(
  "k8s/namespace.yaml"
  "k8s/configmap-common.yaml"
  "k8s/secrets-template.yaml"
  "k8s/rbac-security.yaml"
  "k8s/network-policies.yaml"
  "k8s/jobs/kafka-topics-job.yaml"
  "k8s/jobs/database-init-job.yaml"
  "k8s/kafka.yaml"
  "Makefile"
)

# --- Module Definitions ---
case "$MODULE_NAME" in
  # <<< CHANGE: Added the 'all' case
  all)
    # Define all top-level source directories and root-level files
    MODULE_DIRS=(
      "cmd/"
      "configs/"
      "internal/"
      "k8s/"
      "pkg/"
      "platform/"
      "scripts/"
    )
    MODULE_FILES=(
      ".env"
      "docker-compose.yaml"
      "Dockerfile"
      "Makefile"
    )
    # The 'all' module does not need to also add the shared files
    SHARED_DIRS=()
    SHARED_FILES=()
    ;;
  platform)
    MODULE_DIRS=("platform/")
    MODULE_FILES=()
    ;;
  auth-service)
    MODULE_DIRS=("internal/auth-service/" "cmd/auth-service/")
    MODULE_FILES=(
      "configs/auth-service.yaml"
      "Dockerfile.auth-service"
      "k8s/auth-service.yaml"
      "k8s/mysql-auth.yaml"
    )
    ;;
  core-manager)
    MODULE_DIRS=("internal/core-manager/" "cmd/core-manager/")
    MODULE_FILES=(
      "configs/core-manager.yaml"
      "Dockerfile.core-manager"
      "k8s/core-manager.yaml"
      "k8s/postgres-templates.yaml"
      "k8s/postgres-clients.yaml"
    )
    ;;
  agent-chassis)
    MODULE_DIRS=("cmd/agent-chassis/")
    MODULE_FILES=(
      "configs/agent-chassis.yaml"
      "Dockerfile.agent-chassis"
      "k8s/agent-chassis.yaml"
      "pkg/models/contracts.go"
    )
    ;;
  reasoning-agent)
    MODULE_DIRS=("internal/agents/reasoning/" "cmd/reasoning-agent/")
    MODULE_FILES=(
      "configs/reasoning-agent.yaml"
      "cmd/reasoning-agent/Dockerfile"
      "k8s/reasoning-agent.yaml"
    )
    ;;
  image-generator-adapter)
    MODULE_DIRS=("internal/adapters/imagegenerator/" "cmd/image-generator-adapter/")
    MODULE_FILES=(
      "configs/image-adapter.yaml"
      "Dockerfile.image-generator-adapter"
      "k8s/image-generator-adapter.yaml"
      "k8s/minio.yaml"
    )
    ;;
  web-search-adapter)
    MODULE_DIRS=("internal/adapters/websearch/" "cmd/web-search-adapter/")
    MODULE_FILES=(
      "configs/web-search-adapter.yaml"
      "Dockerfile.web-search-adapter"
      "k8s/web-search-adapter.yaml"
    )
    ;;
  deployment)
    MODULE_DIRS=("scripts/" "k8s/")
    MODULE_FILES=("Makefile" "docker-compose.yaml")
    SHARED_DIRS=()
    SHARED_FILES=()
    ;;
  *)
    echo "Error: Unknown module '$MODULE_NAME'."
    exit 1
    ;;
esac

# --- Packaging Logic ---
for dir in "${MODULE_DIRS[@]}"; do
  write_directory "$dir" "$OUTPUT_FILE"
done
for file in "${MODULE_FILES[@]}"; do
  write_file "$file" "$OUTPUT_FILE"
done

if [ "$MODULE_NAME" != "platform" ] && [ "$MODULE_NAME" != "deployment" ] && [ "$MODULE_NAME" != "all" ]; then
    for dir in "${SHARED_DIRS[@]}"; do
      write_directory "$dir" "$OUTPUT_FILE"
    done
    for file in "${SHARED_FILES[@]}"; do
      write_file "$file" "$OUTPUT_FILE"
    done
fi

echo "✅ Done. Module context saved to $OUTPUT_FILE"
FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
echo "📦 File size: $FILE_SIZE"