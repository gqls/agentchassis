#!/bin/bash
#
# package_chassis_idea_engine.sh
#   Self-contained packager for the CHASSIS IDEA-ENGINE-WORKFLOW context.
#   Bundles (A) the idea.uk engine to port — the generator/validator method
#   source + method docs — and (B) the chassis framework to build it in:
#   the orchestration engine, the action catalogue (reuse-discovery), the
#   agent-creation guidelines, schemas, and seed data. Goal: re-implement the
#   engine as a SEPARATE chassis workflow (agent owns a workflow of steps that
#   call actions), to merge into the site-build workflow later.
#
# Usage:  ./package_chassis_idea_engine.sh [-o output_dir] [-e env] [--no-live] [--no-debugdoc]
#   Env overrides: PROJECT_ROOT, DOCS_ROOT, SQL_ROOT, IDEA_GO_DIR, IDEA_DOCS_DIR
#
# Output:  <output_dir>/<env>_chassis-idea-engine_context.txt
#
# REUSE-DISCOVERY (so we don't reinvent): includes registry.go (every
#   registered action), types.go/helpers.go, the orchestration engine, and the
#   spawn/call/await/branch/workflow actions a new workflow composes from.
#
# PATHS: chassis Go paths follow package_page_build_debug.sh conventions
#   (platform/orchestration/...). Docs/SQL/engine-source are resolved by name
#   under their roots, with a repo-wide name search as a fallback. Anything not
#   found is printed in a MISSING report at the end — fix the roots and re-run.
#
# LIVE CAPTURE (read-only, appended if kubectl present; --no-live to skip):
#   \d of the agent/workflow tables, and the EXISTING agent_definitions
#   workflows — reuse-discovery so the new workflow matches the house pattern
#   and keeps responsibilities distinct.
# ---------------------------------------------------------------------

set -e

# --- Self-locating logic (find the agent-chassis repo root) ----------
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT="${PROJECT_ROOT:-$(realpath "$SCRIPT_DIR/../../" 2>/dev/null || echo "$SCRIPT_DIR")}"
if [ ! -f "$PROJECT_ROOT/go.mod" ] && [ -f "$PWD/go.mod" ]; then PROJECT_ROOT="$PWD"; fi
cd "$PROJECT_ROOT"

# --- Configuration (override any as env vars) ------------------------
DOCS_ROOT="${DOCS_ROOT:-$PROJECT_ROOT/docs/agent_docs/docs024_key_docs_latest}"
SQL_ROOT="${SQL_ROOT:-$PROJECT_ROOT}"                                   # schemas searched here
IDEA_GO_DIR="${IDEA_GO_DIR:-$DOCS_ROOT/idea.uk/golang_files}"           # engine source
IDEA_DOCS_DIR="${IDEA_DOCS_DIR:-$DOCS_ROOT/idea.uk}"                    # method docs
ENVIRONMENT="production"          # affects the output filename only
DEFAULT_OUTPUT_DIR="$DOCS_ROOT/adoption/docubundle/package_module/output_contexts"
WITH_LIVE=true
WITH_DEBUGDOC=true

# --- Argument parsing ------------------------------------------------
OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do
  case $1 in
    -o | --output)       shift; OUTPUT_DIR="$1" ;;
    -e | --environment)  shift; ENVIRONMENT="$1" ;;
    --no-live)           WITH_LIVE=false ;;
    --no-debugdoc)       WITH_DEBUGDOC=false ;;
    -h | --help)
      echo "Usage: $0 [-o output_dir] [-e env] [--no-live] [--no-debugdoc]"
      echo "Env: PROJECT_ROOT DOCS_ROOT SQL_ROOT IDEA_GO_DIR IDEA_DOCS_DIR"
      exit 0
      ;;
  esac
  shift
done

RESOLVED=()
MISSING=()

# --- Helpers ---------------------------------------------------------
function emit() {                # $1=path  $2=list_only
  local path=$1 list_only=$2
  local rel="${path#$PROJECT_ROOT/}"
  echo "filepath = ./$rel" >> "$OUTPUT_FILE"
  if [ "$list_only" = "true" ]; then echo "[File listed only - content not included]" >> "$OUTPUT_FILE"
  else cat "$path" >> "$OUTPUT_FILE"; fi
  echo "-------------------------------------------------" >> "$OUTPUT_FILE"
  RESOLVED+=("$rel")
}

# Resolve want (path or filename) under a root; fall back to name search.
function add_named() {           # $1=want  $2=search_root  $3=list_only(optional)
  local want=$1 root=${2:-$PROJECT_ROOT} list_only=${3:-false} path="" base
  base=$(basename "$want")
  if   [ -f "$root/$want" ];      then path="$root/$want"
  elif [ -f "$PROJECT_ROOT/$want" ]; then path="$PROJECT_ROOT/$want"
  elif [ -f "$want" ];           then path="$want"
  else
    local hits n
    hits=$(find "$root" -type f -name "$base" \
      -not -path '*/.git/*' -not -path '*/vendor/*' -not -path '*/node_modules/*' \
      -not -path '*/output_contexts/*' 2>/dev/null || true)
    if [ -n "$hits" ]; then
      path=$(printf '%s\n' "$hits" | head -n1)
      n=$(printf '%s\n' "$hits" | wc -l | tr -d ' ')
      [ "$n" -gt 1 ] && echo "  note: '$base' matched $n files; using ${path#$PROJECT_ROOT/}" >&2
    fi
  fi
  if [ -n "$path" ] && [ -f "$path" ]; then emit "$path" "$list_only"
  else MISSING+=("$want"); echo "  MISSING: $want" >&2; fi
}

function write_directory() {     # $1=dir
  local dir_path=$1
  if [ ! -d "$dir_path" ]; then echo "Warning: dir '$dir_path' not found. Skipping." >&2; MISSING+=("$dir_path/"); return; fi
  dir_path="${dir_path%/}"
  while IFS= read -r -d $'\0' file; do
    [ "$(realpath "$file" 2>/dev/null)" = "$(realpath "$OUTPUT_FILE" 2>/dev/null)" ] && continue
    emit "$file" "false"
  done < <(find "$dir_path" -type f \
    -not -path '*/.git/*' -not -path '*/vendor/*' -not -path '*/node_modules/*' \
    -not -name 'go.sum' -not -name '*_test.go' \
    -not -name '*.log' -not -name '*.zip' -not -name '*.tar' -not -name '*.gz' \
    -not -name '*.exe' -not -name '*.so' -not -name '*.dylib' \
    -not -name '.DS_Store' -not -name 'Thumbs.db' \
    -print0)
}

# --- The bundle ------------------------------------------------------
mkdir -p "$OUTPUT_DIR"
OUTPUT_FILE="${OUTPUT_DIR}/${ENVIRONMENT}_chassis-idea-engine_context.txt"
> "$OUTPUT_FILE"

echo "Packaging chassis idea-engine-workflow context ($ENVIRONMENT)"
echo "  project root: $PROJECT_ROOT"
echo "  engine src:   $IDEA_GO_DIR"
echo "  output file:  $OUTPUT_FILE"

# === A. THE ENGINE TO PORT (idea.uk source + method docs) ===========
# Method stages = future workflow steps: challenge audience -> generate across
# four lenses -> cut vs the free alternative -> verify via web search -> score
# (incl. the Risk column) -> rank.
ENGINE_SRC=( "engine.go" "prompts.go" "audience_check.go" )
for f in "${ENGINE_SRC[@]}"; do add_named "$f" "$IDEA_GO_DIR" "false"; done

ENGINE_DOCS=(
  "idea_uk_method_v0.md"                       # the method, written up
  "idea_method_prompt.md"
  "idea_uk_testrun_v2.md"                      # worked example — validate the port against this
  "PARALLEL_engine_deployment_and_layer5.md"
  "CONSOLIDATION_where_it_all_fits.md"
)
for f in "${ENGINE_DOCS[@]}"; do add_named "$f" "$IDEA_DOCS_DIR" "false"; done

# === B. THE CHASSIS FRAMEWORK TO BUILD IT IN ========================
# B1. Agent-creation guidelines + current platform state (docs).
GUIDE_DOCS=(
  "000_documentation_index.md"
  "001_development_guide_3_.md"                # agent-creation guidelines
  "002_system_architecture.md"
  "003_contracts_and_standards.md"             # workflow/action contracts + variable conventions
  "019_tool_library.md"
  "020_tool_lifecycle.md"
  "023_llm_quality_testing.md"                 # relevant to the validator/scoring half
  "009_model_infrastructure.md"                # best-model selection + vendor swap
  "running_notes_16_content_quality_and_internal_linking.md"   # recent platform state
  "HANDOFF_2026-06-09_sections_durability_and_content_quality.md"
)
for f in "${GUIDE_DOCS[@]}"; do add_named "$f" "$DOCS_ROOT" "false"; done

# B2. Orchestration engine + action catalogue (Go) — REUSE, don't recreate.
#     Paths per package_page_build_debug.sh.
write_directory "platform/orchestration/types"
write_directory "platform/orchestration/input_contracts"
CHASSIS_GO=(
  # search/extract helpers for the verify-with-web-search stage + action plumbing
  "platform/orchestration/datahelpers/content_search.go"
  "platform/orchestration/datahelpers/deep_search.go"
  "platform/orchestration/datahelpers/unified_extractor.go"
  "platform/orchestration/datahelpers/safe_unmarshal.go"
  "platform/orchestration/datahelpers/sql_helpers.go"
  "platform/orchestration/coordinator.go"               # how steps/workflows run
  "platform/orchestration/state.go"
  "platform/orchestration/helpers.go"
  "platform/orchestration/agent_error_log.go"
  "platform/orchestration/actions/registry.go"          # catalogue: every registered action (reuse-discovery)
  "platform/orchestration/actions/types.go"
  "platform/orchestration/actions/helpers.go"
  "platform/orchestration/actions/workflow_actions.go"  # complete_workflow / control flow
  "platform/orchestration/actions/conditional_branch_action.go"
  "platform/orchestration/actions/basic_actions.go"
  "platform/orchestration/actions/generic_actions.go"
  "platform/orchestration/actions/spawn_actions.go"     # spawn sub-agents (per the no-subworkflows rule)
  "platform/orchestration/actions/call_agent.go"
  "platform/orchestration/actions/await_response.go"
  "platform/database/postgres.go"
  "cmd/agent-chassis/main.go"
)
for f in "${CHASSIS_GO[@]}"; do add_named "$f" "$PROJECT_ROOT" "false"; done
# The LLM-call action (engine stages are prompts). The exact filename varies, so
# include any action source mentioning llm/prompt; registry.go already names it.
while IFS= read -r -d $'\0' f; do emit "$f" "false"; done < <(
  find "$PROJECT_ROOT/platform/orchestration/actions" -maxdepth 1 -type f -name '*.go' \
    \( -iname '*llm*' -o -iname '*prompt*' \) -not -name '*_test.go' -print0 2>/dev/null)

# B3. Schemas — CHECK THE SCHEMA before writing SQL.
SCHEMAS=(
  "002_intake_orchestrator.sql"
  "agent_definition_types.sql"
  "018_briefing_questionnaire.sql"
  "019_model_lifecycle_schema.sql"
  "021_model_swap_and_rollback.sql"
  "schemas_all"
  "schemas_some"
)
for f in "${SCHEMAS[@]}"; do add_named "$f" "$SQL_ROOT" "false"; done

# B4. Seed / table content — the message envelope (headers+body) pattern to log.
SEED=( "initial_messages__without_current_ids_" "initial_vet_practice_check_message" )
for f in "${SEED[@]}"; do add_named "$f" "$SQL_ROOT" "false"; done

# === C. FOR THE LATER SITE-BUILD MERGE (listed only; pull when merging) ===
LATER=(
  "021_site_spec_and_classifier.md"
  "029_site_plan_and_reconciler.md"
  "platform/orchestration/actions/v3_site_actions.go"
  "platform/orchestration/actions/write_site_plan_action.go"
  "platform/orchestration/actions/site_spec_actions.go"
  "platform/orchestration/actions/site_db_actions.go"
)
for f in "${LATER[@]}"; do add_named "$f" "$PROJECT_ROOT" "true"; done   # list_only

# === Debug guide (large; skip with --no-debugdoc) ====================
if [ "$WITH_DEBUGDOC" = true ]; then add_named "016_debugging_guide_v2.md" "$DOCS_ROOT" "false"
else echo "  (skipping 016_debugging_guide_v2.md per --no-debugdoc)"; fi

# === Live capture (read-only reuse-discovery) ========================
if [ "$WITH_LIVE" = true ] && command -v kubectl >/dev/null 2>&1; then
  echo "Capturing live data (read-only; --no-live to skip)…"
  set +e
  PG="kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db"
  {
    echo ""
    echo "================================================================="
    echo "LIVE CAPTURE — $(date -u +%Y-%m-%dT%H:%M:%SZ)  (read-only)"
    echo "================================================================="
    echo ""; echo "----- SCHEMA (\\d) of the agent/workflow tables -----"
  } >> "$OUTPUT_FILE"
  $PG >> "$OUTPUT_FILE" 2>&1 <<'SQL'
\d agent_definitions
\d agents
\d agent_error_log
SQL
  { echo ""; echo "----- EXISTING agent types (avoid overlap; reuse patterns) -----"; } >> "$OUTPUT_FILE"
  $PG -A -t >> "$OUTPUT_FILE" 2>&1 <<'SQL'
SELECT type FROM agent_definitions ORDER BY type;
SQL
  { echo ""; echo "----- A FEW agent_definitions workflows (the house step/action pattern) -----"; } >> "$OUTPUT_FILE"
  $PG -A -t >> "$OUTPUT_FILE" 2>&1 <<'SQL'
SELECT type, default_config
FROM agent_definitions
ORDER BY type
LIMIT 5;
SQL
  set -e
  echo "  live capture appended."
elif [ "$WITH_LIVE" = true ]; then
  echo "Skipping live capture: kubectl not found (use --no-live to silence)."
fi

# --- Report ----------------------------------------------------------
echo ""
echo "Included ${#RESOLVED[@]} files."
if [ "${#MISSING[@]}" -gt 0 ]; then
  echo "MISSING (${#MISSING[@]}) — fix the roots (PROJECT_ROOT/DOCS_ROOT/SQL_ROOT/IDEA_GO_DIR) or paths:"
  printf '  - %s\n' "${MISSING[@]}"
fi
echo "Done. Context saved to $OUTPUT_FILE"
echo "File size: $(du -h "$OUTPUT_FILE" | cut -f1)"
