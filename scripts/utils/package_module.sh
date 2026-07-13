#!/bin/bash
#
# package_module.sh - A script to package the relevant files for a specific
#                      microservice, frontend, or infrastructure component into a
#                      single context file for AI assistants.
#
# This script is designed to work with the new agent-managed project structure.
#
# Usage: ./scripts/utils/package_module.sh [-o /path/to/output_dir] [-e env] [component_name]
# Example: ./scripts/utils/package_module.sh auth-service
# Example: ./scripts/utils/package_module.sh -e production auth-service

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
# Default to development, allow override via flag
#ENVIRONMENT="development"
ENVIRONMENT="production"
REGION="uk_dev" # Assuming 'uk_dev' for development environment

# --- File exclusion (used by the 'remainder' targets) ---
# When non-empty, write_directory skips any file whose project-relative path
# matches an entry here. Lets us express "this whole directory EXCEPT these
# files" without enumerating the directory by hand. Empty for all other
# components, so it has no effect unless a case sets it.
EXCLUDE_FILES=()

# --- Component List ---
# List of all individual components for the 'all' option
ALL_COMPONENTS=(
    # Horizontal slices
    "code-all"
    "deployments-all"
    "environment-all"

    # Full service stacks
    "auth-service"
    "core-manager"
    "agent-chassis"
    "reasoning-agent"
    "web-search-adapter-full"
    "image-generator-adapter-full"

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

    # Backend API slices
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

    # Agent debugging contexts
    "agent-chassis-full"
    "agent-chassis-focused"
    "agent-chassis-core"
    "agent-chassis-orchestration"
    "agent-chassis-comms"
    "agent-chassis-services"
    "agent-chassis-deploy"
    "agent-chassis-actions-current"
    "agent-chassis-actions-remainder"
    "reasoning-agent-full"
    "web-search-adapter-full"
    "image-generator-adapter-full"
    "agents-interaction-kafka"
    "deploy-spawn-actions"

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

    # GRANULAR CONTEXTS ---
    "deploy-all-kustomize"
    "deploy-all-terraform"
    "tf-module-database"
    "tf-module-kafka"
    "tf-module-k8s-util"
    "tf-module-cloud"
    "tf-stack-strimzi-operator"
    "tf-stack-kafka-cluster"
    "tf-stack-kafka-config"

    # --- Testing Components ---
    "test-unit"
    "test-integration"
    "test-e2e"
    "test-tools"
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
# Returns 0 (true) if a file path is in the EXCLUDE_FILES list. Normalises a
# leading "./" on both sides so matches are robust regardless of how find or
# the case statement spell the path.
function is_excluded() {
  local f="${1#./}"
  local ex
  for ex in "${EXCLUDE_FILES[@]}"; do
    ex="${ex#./}"
    if [ "$f" = "$ex" ]; then
      return 0
    fi
  done
  return 1
}

function write_directory() {
  local dir_path=$1
  local output_file=$2

  # Check if the directory exists before trying to find files in it.
  if [ ! -d "$dir_path" ]; then
    echo "Warning: Directory '$dir_path' not found in '$PWD'. Skipping." >&2
    return
  fi

  # Remove trailing slash if present
  dir_path="${dir_path%/}"

  # Use find to get ALL files in the directory and subdirectories
  while IFS= read -r -d $'\0' file; do
    # Skip if this is the output file itself to avoid self-reference
    if [ "$(realpath "$file" 2>/dev/null)" = "$(realpath "$output_file" 2>/dev/null)" ]; then
      continue
    fi

    # Skip files explicitly excluded (used by the 'remainder' targets).
    if is_excluded "$file"; then
      continue
    fi

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
    -not -path '*/output_contexts/*' \
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

# Helper function to handle MODULE_FILES which can be individual files or directories
function process_module_files() {
  local item=$1
  local output_file=$2

  if [ -f "$item" ]; then
    # It's a file, write it directly
    write_file "$item" "$output_file" "false"
  elif [ -d "$item" ]; then
    # It's a directory, process all files in it
    write_directory "$item" "$output_file"
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
    -e | --environment)
      shift
      ENVIRONMENT=$1
      ;;
  esac
  shift
done

COMPONENT_NAME=$1

# --- Help and Usage ---
function show_help() {
  echo "Usage: $0 [-o /path/to/output_dir] [-e environment] [component_name]"
  echo "Please provide the name of the component to package."
  echo ""
  echo "Available components:"
  echo ""
  echo "  BULK OPERATIONS:"
  echo "    all                      # Package all individual components into separate files"
  echo "    all-in-one               # Package all components into a single combined file"
  echo ""
  echo "  HORIZONTAL SLICES:"
  echo "    code-all                 # All Go source code (cmd, internal, pkg, platform)"
  echo "    deployments-all          # All deployment configurations (Terraform & Kustomize)"
  echo "    environment-all          # All environment Terraform configurations"
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
  echo "  INFRASTRUCTURE:"
  echo "    infra-cluster-provisioning       # Terraform for provisioning the base Rackspace K8s cluster"
  echo "    infra-kafka-stack                # The complete Kafka stack (Terraform modules & Kustomize instance)"
  echo "    infra-terraform-rackspace-module # Module for creating the Rackspace Kubernetes cluster"
  echo "    infra-terraform-kafka-modules    # Modules for Strimzi Operator and the Kafka Cluster"
  echo "    infra-terraform-environment      # Top-level Terraform config for the specified environment"
  echo "    infra-kustomize-kafka-instance   # Kustomize definition for the Kafka cluster resource"
  echo "    infra-kustomize-ingress          # Kustomize definition for the NGINX Ingress"
  echo ""
  echo "  FRONTEND DEVELOPMENT:"
  echo "    frontend-user-portal-only    # Just user portal React code"
  echo "    frontend-admin-only          # Just admin dashboard React code"
  echo "    frontend-playground-only     # Just playground React code"
  echo "    frontend-shared-components   # Shared UI components and API client"
  echo "    frontend-all-apps            # All frontend applications"
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
  echo "    platform-data            # Database, storage, and memory services"
  echo "    platform-observability   # Metrics, health, tracing, resilience"
  echo ""
  echo "  AGENT DEVELOPMENT:"
  echo "    agent-framework          # Agent base classes and orchestration"
  echo "    agent-reasoning-only     # Just reasoning agent implementation"
  echo "    agent-adapters           # Web search and image adapter code"
  echo ""
  echo "  AGENT DEBUGGING:"
  echo "    agent-chassis-full           # Massive combined context for the chassis"
  echo "    agent-chassis-focused        # Truncated context for the chassis"
  echo "    agent-chassis-core           # Chassis core and configs"
  echo "    agent-chassis-orchestration  # Orchestration platform only"
  echo "    agent-chassis-comms          # Kafka and messaging"
  echo "    agent-chassis-services       # DB, discovery, storage, etc."
  echo "    agent-chassis-deploy         # Deployment manifests"
  echo "    agent-chassis-actions-current    # Orchestration core + actions used by live workflows"
  echo "    agent-chassis-actions-remainder  # All other action files (actions tree minus 'current')"
  echo "    reasoning-agent-full"
  echo "    web-search-adapter-full"
  echo "    image-generator-adapter-full"
  echo "    agents-interaction-kafka"
  echo "    deploy-spawn-actions"
  echo ""
  echo "  DATABASE & MIGRATIONS:"
  echo "    database-schemas         # All migration files and seed data"
  echo "    database-auth            # Auth-related database code"
  echo "    database-clients         # Client/persona database code"
  echo ""
  echo "  DEPLOYMENT SPECIFIC:"
  echo "    deploy-terraform-modules    # Reusable Terraform modules"
  echo "    deploy-kustomize-base       # Base Kustomize configurations"
  echo "    deploy-services             # Service deployment configs"
  echo "    deploy-frontends            # Frontend deployment configs"
  echo ""
  echo "  DEVELOPMENT TOOLS:"
  echo "    dev-scripts              # All development scripts"
  echo "    dev-docker               # Docker configurations"
  echo "    dev-local-env            # Local development environment"
  echo ""
  echo "  TESTING:"
  echo "    test-unit                # Unit test code"
  echo "    test-integration         # Integration test code"
  echo "    test-e2e                 # End-to-end test code"
  echo "    test-tools               # Test tools"
  echo "    test-all                 # All test code"
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
    echo "--> Packaging component: $component (Env: $ENVIRONMENT)"

    # Build the command with optional flags
    CMD="bash \"$SCRIPT_PATH\""
    if [[ -n "$OUTPUT_DIR" && "$OUTPUT_DIR" != "$DEFAULT_OUTPUT_DIR" ]]; then
      CMD+=" -o \"$OUTPUT_DIR\""
    fi
    if [[ "$ENVIRONMENT" != "development" ]]; then # Pass env if not default
        CMD+=" -e \"$ENVIRONMENT\""
    fi
    CMD+=" \"$component\""

    # Execute the command
    eval $CMD

    # Display the file size for the component just created
    COMPONENT_FILE="${OUTPUT_DIR}/${ENVIRONMENT}_${component}_context.txt"
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
    COMPONENT_FILE="${OUTPUT_DIR}/${ENVIRONMENT}_${component}_context.txt"
    if [ -f "$COMPONENT_FILE" ]; then
      FILE_SIZE=$(du -h "$COMPONENT_FILE" | cut -f1)
      printf "  %-35s %10s\n" "${ENVIRONMENT}_${component}_context.txt" "$FILE_SIZE"
    fi
  done
  exit 0
fi

# If the component is 'all-in-one', create a single file with everything
if [ "$COMPONENT_NAME" = "all-in-one" ]; then
  echo "Packaging all components into a single file..."
  mkdir -p "$OUTPUT_DIR"

  TEMP_DIR=$(mktemp -d)
  ALL_IN_ONE_FILE="${OUTPUT_DIR}/all-in-one_context.txt"
  > "$ALL_IN_ONE_FILE"

  echo "Environment: $ENVIRONMENT" >> "$ALL_IN_ONE_FILE"
  echo "Generated on: $(date)" >> "$ALL_IN_ONE_FILE"
  echo "=================================================" >> "$ALL_IN_ONE_FILE"
  echo "" >> "$ALL_IN_ONE_FILE"

  # First generate all individual component files in temp directory
  for component in "${ALL_COMPONENTS[@]}"; do
    echo "--> Processing component: $component"

    # Build the command with temp directory
    CMD="bash \"$SCRIPT_PATH\" -o \"$TEMP_DIR\""
    if [[ "$ENVIRONMENT" != "development" ]]; then
        CMD+=" -e \"$ENVIRONMENT\""
    fi
    CMD+=" \"$component\""

    # Execute the command
    eval $CMD 2>/dev/null
  done

  # Now concatenate all files with headers
  for component in "${ALL_COMPONENTS[@]}"; do
    COMPONENT_FILE="${TEMP_DIR}/${component}_context.txt"
    if [ -f "$COMPONENT_FILE" ]; then
      echo "" >> "$ALL_IN_ONE_FILE"
      echo "=================================================" >> "$ALL_IN_ONE_FILE"
      echo "COMPONENT: $component" >> "$ALL_IN_ONE_FILE"
      echo "=================================================" >> "$ALL_IN_ONE_FILE"
      echo "" >> "$ALL_IN_ONE_FILE"
      cat "$COMPONENT_FILE" >> "$ALL_IN_ONE_FILE"
    fi
  done

  # Clean up temp directory
  rm -rf "$TEMP_DIR"

  echo "-------------------------------------------------"
  echo "✅ All components packaged into single file."
  FILE_SIZE=$(du -h "$ALL_IN_ONE_FILE" | cut -f1)
  echo "📦 Output file: $ALL_IN_ONE_FILE"
  echo "📦 Total size: $FILE_SIZE"
  exit 0
fi

mkdir -p "$OUTPUT_DIR"
OUTPUT_FILE="${OUTPUT_DIR}/${ENVIRONMENT}_${COMPONENT_NAME}_context.txt"
> "$OUTPUT_FILE"

echo "Packaging component '$COMPONENT_NAME' for environment '$ENVIRONMENT' into $OUTPUT_FILE..."

# Adjust region based on environment
if [ "$ENVIRONMENT" = "production" ]; then
    REGION="uk001"
else
    REGION="uk_dev"
fi


# --- Component Definitions ---
# Each case defines the specific source code, build, and deployment files
# that make up a complete, independent component.

# Shared files are included where necessary to provide full context.
SHARED_PLATFORM_CODE=("platform/" "pkg/")
SHARED_DEPLOYMENT_MODULES=("deployments/terraform/modules/kustomize-apply/")
SHARED_KUSTOMIZE_BASE=("deployments/kustomize/base/")
SHARED_ROOT_FILES=("Makefile" "go.mod" "go.sum" "docker-compose.yaml")

# --- Action surface split (from the action-usage audit) ---
# CURRENT_ACTION_FILES: orchestration core + action files whose actions are
# referenced by at least one ACTIVE workflow. Used directly by
# 'agent-chassis-actions-current', and as the exclusion set for
# 'agent-chassis-actions-remainder' (= everything else in the actions tree).
# Keep this list as the single source of truth for the split — the remainder
# target derives from it automatically, so files only need listing here once.
CURRENT_ACTION_FILES=(
  # --- orchestration core (outside actions/) ---
  "platform/orchestration/coordinator.go"
  "platform/orchestration/state.go"
  "platform/orchestration/helpers.go"
  "platform/orchestration/agent_error_log.go"
  "platform/orchestration/loop_expansion_handler.go"
  "platform/orchestration/loop_error_handler.go"
  # --- data helpers ---
  "platform/orchestration/datahelpers/data_helpers.go"
  "platform/orchestration/datahelpers/action_inputs.go"
  "platform/orchestration/datahelpers/unified_extractor.go"
  "platform/orchestration/datahelpers/timeout_helpers.go"
  # --- queryresolve subpackage ---
  "platform/orchestration/actions/queryresolve/queryresolve.go"
  # --- site-build pipeline action files (actions/ dir, excludes dormant subsystems) ---
  "platform/orchestration/actions/ai_actions.go"
  "platform/orchestration/actions/ai_errors.go"
  "platform/orchestration/actions/apply_adoption_plan_action.go"
  "platform/orchestration/actions/apply_gap_plan_action.go"
  "platform/orchestration/actions/assemble_from_library.go"
  "platform/orchestration/actions/await_response.go"
  "platform/orchestration/actions/basic_actions.go"
  "platform/orchestration/actions/batch_webscrape_action.go"
  "platform/orchestration/actions/call_agent.go"
  "platform/orchestration/actions/check_endpoint_health_action.go"
  "platform/orchestration/actions/check_tool_completeness_action.go"
  "platform/orchestration/actions/checkpoint_for_review_action.go"
  "platform/orchestration/actions/claim_work_item_action.go"
  "platform/orchestration/actions/cleanup_stale_topics.go"
  "platform/orchestration/actions/color_util.go"
  "platform/orchestration/actions/component_library.go"
  "platform/orchestration/actions/component_selector.go"
  "platform/orchestration/actions/component_validation.go"
  "platform/orchestration/actions/compute_component_quality.go"
  "platform/orchestration/actions/conditional_branch_action.go"
  "platform/orchestration/actions/create_blog_posts_action.go"
  "platform/orchestration/actions/create_rerender_items_action.go"
  "platform/orchestration/actions/create_tool_component_action.go"
  "platform/orchestration/actions/create_tool_cross_link_items.go"
  "platform/orchestration/actions/create_work_item_action.go"
  "platform/orchestration/actions/css_templating.go"
  "platform/orchestration/actions/database_actions.go"
  "platform/orchestration/actions/deploy_image_asset_action.go"
  "platform/orchestration/actions/deploy_tool_action.go"
  "platform/orchestration/actions/design_actions.go"
  "platform/orchestration/actions/discovery_actions.go"
  "platform/orchestration/actions/discovery_checks.go"
  "platform/orchestration/actions/dispatch_actions.go"
  "platform/orchestration/actions/enrich_fingerprint_with_css_action.go"
  "platform/orchestration/actions/entity_state_actions.go"
  "platform/orchestration/actions/extract_css_vars.go"
  "platform/orchestration/actions/extract_design_fingerprint_action.go"
  "platform/orchestration/actions/extract_interactive_fingerprint_action.go"
  "platform/orchestration/actions/firecrawl_map_action.go"
  "platform/orchestration/actions/fix_component_template_action.go"
  "platform/orchestration/actions/fix_forced_text_colours_action.go"
  "platform/orchestration/actions/fix_harcoded_colours_action.go"
  "platform/orchestration/actions/fix_nav_link_templates_action.go"
  "platform/orchestration/actions/fork_theme_composition.go"
  "platform/orchestration/actions/fork_theme_from_site_action.go"
  "platform/orchestration/actions/format_crawl_for_analysis_action.go"
  "platform/orchestration/actions/generate_image_actions.go"
  "platform/orchestration/actions/generic_actions.go"
  "platform/orchestration/actions/get_pages_for_rerender_action.go"
  "platform/orchestration/actions/get_pages_to_build_actions.go"
  "platform/orchestration/actions/git_deployer_actions.go"
  "platform/orchestration/actions/helpers.go"
  "platform/orchestration/actions/hitl_actions.go"
  "platform/orchestration/actions/hitl_persistence.go"
  "platform/orchestration/actions/hitl_request_human_input.go"
  "platform/orchestration/actions/html_actions.go"
  "platform/orchestration/actions/http_request_logger.go"
  "platform/orchestration/actions/install_site_composition_action.go"
  "platform/orchestration/actions/link_constraints.go"
  "platform/orchestration/actions/link_site_components_action.go"
  "platform/orchestration/actions/llm_call_logger.go"
  "platform/orchestration/actions/load_component_library_actions.go"
  "platform/orchestration/actions/load_existing_content_action.go"
  "platform/orchestration/actions/load_page_record_action.go"
  "platform/orchestration/actions/load_page_sections_from_spec_action.go"
  "platform/orchestration/actions/load_site_pages_action.go"
  "platform/orchestration/actions/load_work_item_actions.go"
  "platform/orchestration/actions/lock_helpers.go"
  "platform/orchestration/actions/lock_policy.go"
  "platform/orchestration/actions/loop_actions.go"
  "platform/orchestration/actions/maintenance_actions.go"
  "platform/orchestration/actions/multipage_actions.go"
  "platform/orchestration/actions/nav_tables.go"
  "platform/orchestration/actions/page_growth_budget.go"
  "platform/orchestration/actions/plan_sections_action.go"
  "platform/orchestration/actions/populate_nav_tables_action.go"
  "platform/orchestration/actions/prepare_link_context_action.go"
  "platform/orchestration/actions/query_agent_definitions_actions.go"
  "platform/orchestration/actions/read_layout_taxonomy_action.go"
  "platform/orchestration/actions/rebuild_blog_listing_action.go"
  "platform/orchestration/actions/reconcile_site_plan_action.go"
  "platform/orchestration/actions/registry.go"
  "platform/orchestration/actions/render_css_composition_helpers.go"
  "platform/orchestration/actions/render_css_composition_loader.go"
  "platform/orchestration/actions/render_css_from_spec_action.go"
  "platform/orchestration/actions/render_js_snippets_for_site_action.go"
  "platform/orchestration/actions/render_site_components_action.go"
  "platform/orchestration/actions/rerender_pages_actions.go"
  "platform/orchestration/actions/rerender_single_page_action.go"
  "platform/orchestration/actions/research_actions.go"
  "platform/orchestration/actions/resolve_composition_helpers.go"
  "platform/orchestration/actions/resolve_composition_layout_action.go"
  "platform/orchestration/actions/resolve_composition_pallette_action.go"
  "platform/orchestration/actions/resolve_composition_typography_action.go"
  "platform/orchestration/actions/save_component_history_action.go"
  "platform/orchestration/actions/save_page_sections_action.go"
  "platform/orchestration/actions/section_editor_actions.go"
  "platform/orchestration/actions/seed_build_queue_action.go"
  "platform/orchestration/actions/site_db_actions.go"
  "platform/orchestration/actions/site_snapshots_actions.go"
  "platform/orchestration/actions/site_spec_actions.go"
  "platform/orchestration/actions/spawn_actions.go"
  "platform/orchestration/actions/spawn_group.go"
  "platform/orchestration/actions/storage_actions.go"
  "platform/orchestration/actions/store_generated_component_action.go"
  "platform/orchestration/actions/sync_site_identity_action.go"
  "platform/orchestration/actions/transform_actions.go"
  "platform/orchestration/actions/triage_detect_items_action.go"
  "platform/orchestration/actions/types.go"
  "platform/orchestration/actions/update_component_html_action.go"
  "platform/orchestration/actions/update_site_spec_from_item_action.go"
  "platform/orchestration/actions/v3_site_actions.go"
  "platform/orchestration/actions/validate_composition_inputs_action.go"
  "platform/orchestration/actions/validate_dark_sections.go"
  "platform/orchestration/actions/validate_page_content.go"
  "platform/orchestration/actions/web_search_action.go"
  "platform/orchestration/actions/webscrape_actions.go"
  "platform/orchestration/actions/work_items_common.go"
  "platform/orchestration/actions/write_audit_findings_action.go"
  "platform/orchestration/actions/write_site_plan_action.go"
)

case "$COMPONENT_NAME" in
  # --- Horizontal Slices ---
  code-all)
    MODULE_DIRS=(
    "cmd/"
    "internal/"
    "pkg/"
    "platform/"
    "configs/"
    "deployments/"
    )
    MODULE_FILES=( "makefile" )
    ;;

  deployments-all)
    MODULE_DIRS=( "deployments/" "build/docker/" )
    MODULE_FILES=( "Makefile" "docker-compose.yaml" )
    ;;

  environment-all)
    MODULE_DIRS=( "deployments/terraform/environments/" )
    MODULE_FILES=( "Makefile" )
    ;;

  # --- Full Service Stacks ---
  auth-service)
    MODULE_DIRS=(
      "cmd/auth-service/" "internal/auth-service/"
      "deployments/kustomize/services/auth-service/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/core-platform/110-auth-service/"
      "test/unit/auth/"
      "${SHARED_PLATFORM_CODE[@]}" "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "build/docker/backend/auth-service.dockerfile" "configs/auth-service.yaml"
      "test/e2e/scenarios/auth_test.go"
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

  # --- Infrastructure Layers (Refined and Expanded) ---
  infra-cluster-provisioning)
    MODULE_DIRS=(
      "deployments/terraform/modules/rackspace-kubernetes/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/010-infrastructure/"
    )
    MODULE_FILES=("Makefile")
    ;;

  infra-kafka-stack)
    MODULE_DIRS=(
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/030-strimzi-operator/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/040-kafka-cluster/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/045-kafka-users/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/080-kafka-topics/"
      "deployments/terraform/modules/strimzi-operator/"
      "deployments/terraform/modules/kafka-cluster/"
      "deployments/terraform/modules/kafka_topics/"
      # --- KRAFT DEBUGGING ---
      "deployments/terraform/modules/strimzi-operator/strimzi-0.47.0/"
      "deployments/terraform/modules/kafka-cluster/config/"
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


  # --- Agent-Specific Debugging ---
  agent-chassis-full)
    MODULE_DIRS=(
      "cmd/agent-chassis/"
      "internal/agents/"
      "configs/"
      "deployments/kustomize/services/agent-chassis/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2210-agent-chassis/"
      "platform/agentbase/"
      "platform/kafka/"
      "platform/messaging/"
      "platform/orchestration/"
      # "platform/database/"
      "platform/discovery/"
      "platform/evolution/"
      "platform/storage/"
      "platform/aiservice/"
      "platform/governance/"
      "platform/validation/"
    )
    MODULE_FILES=( "makefile" "build/docker/backend/agent-chassis.dockerfile" )
    ;;

  # --- Modular Agent Chassis Debugging (Context-Friendly) ---

  agent-chassis-core)
    MODULE_DIRS=(
      "cmd/agent-chassis/"
      "internal/agents/"
      "configs/"
      "platform/agentbase/"
    )
    MODULE_FILES=( "makefile" "build/docker/backend/agent-chassis.dockerfile" )
    ;;

  agent-chassis-orchestration)
    MODULE_DIRS=(
      "platform/orchestration/"
    )
    ;;

  agent-chassis-comms)
    MODULE_DIRS=(
      "platform/kafka/"
      "platform/messaging/"
    )
    ;;

  agent-chassis-services)
    MODULE_DIRS=(
      "platform/database/"
      "platform/discovery/"
      "platform/evolution/"
      "platform/storage/"
      "platform/aiservice/"
      "platform/governance/"
    )
    ;;

  agent-chassis-deploy)
    MODULE_DIRS=(
      "deployments/kustomize/services/agent-chassis/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2210-agent-chassis/"
    )
    ;;

  # --- Action surface, split by live-workflow usage ---
  # Orchestration core + action files referenced by at least one ACTIVE
  # workflow (per the action-usage audit). The file list lives in the shared
  # CURRENT_ACTION_FILES array above so the 'remainder' target can derive from
  # it without duplication.
  agent-chassis-actions-current)
    MODULE_FILES=( "${CURRENT_ACTION_FILES[@]}" )
    ;;

  # Everything else in the actions tree: the whole
  # platform/orchestration/actions/ directory MINUS the files already in
  # CURRENT_ACTION_FILES. Because it's computed by exclusion rather than a
  # hand-listed set, it stays correct as new action files are added — any new
  # file lands here automatically until it's promoted into CURRENT_ACTION_FILES.
  agent-chassis-actions-remainder)
    MODULE_DIRS=( "platform/orchestration/actions/" )
    EXCLUDE_FILES=( "${CURRENT_ACTION_FILES[@]}" )
    ;;

# =====================================================================
  # FOCUSED AGENT CHASSIS CONTEXT (~40k lines, ~1.4MB)
  # Down from 90k/3.1MB full version.
  #
  # Keeps: orchestration core, key site-building actions, discovery,
  #        kafka, work item pipeline, key fix actions, configs
  #
  # Drops: component_library, multipage, business_intel, companies_house,
  #        html_actions, section_editor, generate_image, fix_forced_text,
  #        storage_actions, loop_actions, migration READMEs + SQL,
  #        manual seeding, tests, makefile, topicflow.png, vet pipeline
  #        detail, internal agent implementations, HITL detail,
  #        calculate_actions, aggregate_*, entity_state
  #
  # If you need a dropped file, upload it directly to the chat.
  # =====================================================================
  agent-chassis-focused)
    MODULE_DIRS=(
      # Discovery checks (small files, all needed)
      "platform/orchestration/actions/discovery_checks/"

      # Kafka (essential)
      "platform/kafka/"
    )
    MODULE_FILES=(
      # --- Agent core ---
      "platform/agentbase/agent.go"
      "platform/agentbase/bootstrap.go"
      "platform/agentbase/server.go"
      "platform/messaging/processor.go"
      "platform/messaging/context.go"

      # --- Orchestration core ---
      "platform/orchestration/coordinator.go"
      "platform/orchestration/state.go"
      "platform/orchestration/helpers.go"
      "platform/orchestration/agent_error_log.go"
      "platform/orchestration/loop_expansion_handler.go"
      "platform/orchestration/loop_error_handler.go"
      "platform/orchestration/types/context.go"
      "platform/orchestration/types/trace_logger.go"
      "platform/orchestration/input_contracts/input_mapping.go"

      # --- Data helpers (essential subset) ---
      "platform/orchestration/datahelpers/data_helpers.go"
      "platform/orchestration/datahelpers/action_inputs.go"
      "platform/orchestration/datahelpers/unified_extractor.go"
      "platform/orchestration/datahelpers/timeout_helpers.go"

      # --- Action registry + types ---
      "platform/orchestration/actions/registry.go"
      "platform/orchestration/actions/types.go"
      "platform/orchestration/actions/helpers.go"

      # --- Core actions (spawning, calling, workflow) ---
      "platform/orchestration/actions/spawn_actions.go"
      "platform/orchestration/actions/call_agent.go"
      "platform/orchestration/actions/workflow_actions.go"
      "platform/orchestration/actions/basic_actions.go"
      "platform/orchestration/actions/conditional_branch_action.go"

      # --- Site building pipeline ---
      "platform/orchestration/actions/v3_site_actions.go"
      "platform/orchestration/actions/site_db_actions.go"
      "platform/orchestration/actions/site_spec_actions.go"
      "platform/orchestration/actions/maintenance_actions.go"
      "platform/orchestration/actions/validate_page_content.go"
      "platform/orchestration/actions/render_site_components_action.go"
      "platform/orchestration/actions/rerender_pages_actions.go"
      "platform/orchestration/actions/save_page_sections_action.go"
      "platform/orchestration/actions/render_css_from_spec_action.go"
      "platform/orchestration/actions/plan_sections_action.go"
      "platform/orchestration/actions/git_deployer_actions.go"

      # --- Work item + dispatch ---
      "platform/orchestration/actions/load_work_item_actions.go"
      "platform/orchestration/actions/claim_work_item_action.go"
      "platform/orchestration/actions/create_work_item_action.go"
      "platform/orchestration/actions/dispatch_actions.go"
      "platform/orchestration/actions/triage_detect_items_action.go"
      "platform/orchestration/actions/seed_build_queue_action.go"

      # --- Discovery + audit ---
      "platform/orchestration/actions/discovery_actions.go"
      "platform/orchestration/actions/discovery_checks.go"
      "platform/orchestration/actions/write_audit_findings_action.go"

      # --- Fix actions ---
      "platform/orchestration/actions/fix_component_template_action.go"
      "platform/orchestration/actions/fix_harcoded_colours_action.go"
      "platform/orchestration/actions/update_component_html_action.go"

      # --- Blog + content + tools ---
      "platform/orchestration/actions/create_blog_posts_action.go"
      "platform/orchestration/actions/apply_gap_plan_action.go"
      "platform/orchestration/actions/create_tool_component_action.go"
      "platform/orchestration/actions/deploy_tool_action.go"

      # --- LLM ---
      "platform/orchestration/actions/ai_actions.go"

      # --- Navigation ---
      "platform/orchestration/actions/nav_tables.go"
      "platform/orchestration/actions/populate_nav_tables_action.go"
      "platform/orchestration/actions/link_site_components_action.go"

      # --- Page loading ---
      "platform/orchestration/actions/load_page_record_action.go"
      "platform/orchestration/actions/load_site_pages_action.go"
      "platform/orchestration/actions/get_pages_to_build_actions.go"
      "platform/orchestration/actions/get_pages_for_rerender_action.go"

      # --- Database actions ---
      "platform/orchestration/actions/database_actions.go"

      # --- Design ---
      "platform/orchestration/actions/design_actions.go"
      "platform/orchestration/actions/assemble_from_library.go"
      "platform/orchestration/actions/deploy_image_asset_action.go"

      # --- Database Go code ---
      "platform/database/postgres.go"

      # --- AI service ---
      "platform/aiservice/anthropic.go"
      "platform/aiservice/model_aliases.go"

      # --- Entry point + configs ---
      "cmd/agent-chassis/main.go"
      "configs/agent-chassis.yaml"
      "configs/core-manager.yaml"

      # --- Deployment ---
      "deployments/kustomize/services/agent-chassis/overlays/production/uk_001/patch-deployment.yaml"
      "deployments/kustomize/services/agent-chassis/base/deployment.yaml"
    )
    ;;

  reasoning-agent-full)
    MODULE_DIRS=(
      "cmd/reasoning-agent/"
      "internal/agents/reasoning/"
      "deployments/kustomize/services/reasoning-agent/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2220-reasoning-agent/"
      "platform/agentbase/"
      "platform/kafka/"
      "platform/messaging/"
      "platform/orchestration/"
    )
    MODULE_FILES=(
      "makefile"
      "build/docker/backend/reasoning-agent.dockerfile"
      "configs/reasoning-agent.yaml"
    )
    ;;

  web-search-adapter-full)
    MODULE_DIRS=(
      "cmd/web-search-adapter/"
      "internal/adapters/websearch/"
      "deployments/kustomize/services/web-search-adapter/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2230-web-search-adapter/"
      "platform/agentbase/"
      "platform/kafka/"
      "platform/messaging/"
      "platform/orchestration/"
    )
    MODULE_FILES=(
      "makefile"
      "build/docker/backend/web-search-adapter.dockerfile"
      "configs/web-search-adapter.yaml"
    )
    ;;

  image-generator-adapter-full)
    MODULE_DIRS=(
      "cmd/image-generator-adapter/"
      "internal/adapters/imagegenerator/"
      "deployments/kustomize/services/image-generator-adapter/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2240-image-generator-adapter/"
      "platform/agentbase/"
      "platform/kafka/"
      "platform/messaging/"
      "platform/orchestration/"
    )
    MODULE_FILES=(
      "makefile"
      "build/docker/backend/image-generator-adapter.dockerfile"
      "configs/image-generator-adapter.yaml"
    )
    ;;

  content-creator-agent) # Add this new case
    MODULE_DIRS=(
      "cmd/content-creator-agent/" "internal/agents/contentcreator/"
      "deployments/kustomize/services/content-creator-agent/"
      # Assuming you'll add a Terraform module path for it later if needed, e.g.:
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/services/agents/2250-content-creator-agent/"
      "${SHARED_PLATFORM_CODE[@]}" "${SHARED_DEPLOYMENT_MODULES[@]}" "${SHARED_KUSTOMIZE_BASE[@]}"
    )
    MODULE_FILES=(
      "build/docker/backend/content-creator-agent.dockerfile" "configs/content-creator-agent.yaml"
      "${SHARED_ROOT_FILES[@]}"
    )
  ;;

  agents-interaction-kafka)
    MODULE_DIRS=(
      "cmd/core-manager/"
      "internal/core-manager/"
      "cmd/agent-chassis/"
      "cmd/reasoning-agent/"
      "cmd/web-search-adapter/"
      "cmd/image-generator-adapter/"
      "platform/kafka/"
      "platform/messaging/"
      "platform/agentbase"
      "platform/orchestration/"
      "platform/contracts/"
    )
    MODULE_FILES=( "configs/core-manager.yaml" "configs/kafka_topics.yaml" )
    ;;

  deploy-spawn-actions)
    MODULE_DIRS=(
      #"cmd/test-spawning/"
      "internal/core-manager/api/"
      #"internal/core-manager/admin/"
      "platform/orchestration/actions/"
      "platform/orchestration/"
      "platform/discovery/"
      "platform/database/migrations/"
      "platform/contracts/"
      "platform/kafka/"
      "platform/resilience/"
    )
    MODULE_FILES=(
      "cmd/core-manager/main.go"
      #"cmd/test-spawning/main.go"
      #"go.mod"
      #"go.sum"
      #"Makefile"
    )
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

  # --- Granular Breakdowns ---
  deploy-all-kustomize)
    MODULE_DIRS=( "deployments/kustomize/" )
    MODULE_FILES=( "makefile" )
    ;;

  deploy-all-terraform)
    MODULE_DIRS=( "deployments/terraform/" )
    MODULE_FILES=( "makefile" )
    ;;

  tf-module-database)
    MODULE_DIRS=(
      "deployments/terraform/modules/postgres-instance/"
      "deployments/terraform/modules/mysql-instance/"
    )
    ;;

  tf-module-kafka)
    MODULE_DIRS=(
      "deployments/terraform/modules/strimzi-operator/"
      "deployments/terraform/modules/kafka-cluster/"
      "deployments/terraform/modules/kafka_topics/"
    )
    ;;

  tf-module-k8s-util)
    MODULE_DIRS=(
      "deployments/terraform/modules/kustomize-apply/"
      "deployments/terraform/modules/nginx-ingress/"
      "deployments/terraform/modules/k8s-job-runner/"
    )
    ;;

  tf-module-cloud)
    MODULE_DIRS=(
      "deployments/terraform/modules/rackspace-kubernetes/"
      "deployments/terraform/modules/s3-buckets/"
    )
    ;;

  tf-stack-strimzi-operator)
    MODULE_DIRS=(
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/030-strimzi-operator/"
      "deployments/terraform/modules/strimzi-operator/"
    )
    MODULE_FILES=( "makefile" )
    ;;

  tf-stack-kafka-cluster)
    MODULE_DIRS=(
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/040-kafka-cluster/"
      "deployments/terraform/modules/kafka-cluster/"
    )
    MODULE_FILES=( "makefile" )
    ;;

  tf-stack-kafka-config)
    MODULE_DIRS=(
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/045-kafka-users/"
      "deployments/terraform/environments/$ENVIRONMENT/$REGION/080-kafka-topics/"
      "deployments/terraform/modules/kafka_topics/"
    )
    MODULE_FILES=( "makefile" )
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
  test-unit)
    MODULE_DIRS=( "test/unit/" )
    MODULE_FILES=( "go.mod" "Makefile" "scripts/run_unit_tests.sh" )
    ;;

  test-integration)
    MODULE_DIRS=( "test/integration/" )
    MODULE_FILES=( "go.mod" "Makefile" "scripts/run_integration_tests.sh" )
    ;;

  test-e2e)
    MODULE_DIRS=( "test/e2e/" )
    MODULE_FILES=( "go.mod" "Makefile" "scripts/run_e2e_tests.sh" "docker/test-harness.dockerfile" )
    ;;

  test-tools)
    MODULE_DIRS=( "test/tools/" "test/scripts/" )
    MODULE_FILES=( "Makefile" )
    ;;

  test-all)
    MODULE_DIRS=( "test/" "scripts/" "docker/" "k8s/" "migrations/" "fixtures/" )
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
for item in "${MODULE_FILES[@]}"; do
  process_module_files "$item" "$OUTPUT_FILE"
done

echo "✅ Done. Component context saved to $OUTPUT_FILE"
FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
echo "📦 File size: $FILE_SIZE"