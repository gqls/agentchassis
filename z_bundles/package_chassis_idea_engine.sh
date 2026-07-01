#!/bin/bash
#
# package_chassis_idea_engine.sh
#   Self-contained packager for the CHASSIS IDEA-ENGINE-WORKFLOW context.
#   (A) the idea.uk engine to port — generator/validator method source + docs —
#   and (B) the chassis framework to build it in: orchestration engine, the
#   action catalogue (reuse-discovery), agent-creation guidelines, schemas, seed.
#   Goal: re-implement the engine as a SEPARATE chassis workflow (agent owns a
#   workflow of steps that call actions), to merge into site-build later.
#
#   COPES WITH A MESSY FOLDER: docs resolve to the NEWEST (N) variant by mtime;
#   the engine-source/code walks drop *.orig* backups + the binary.
#
# Usage:  ./package_chassis_idea_engine.sh [-o output_dir] [-e env] [--no-live] [--no-debugdoc]
#   Run from the repo root (go.mod), or set REPO_ROOT.
#   Env: REPO_ROOT DOC_SEARCH_ROOT SQL_ROOT IDEA_GO_DIR
#
# Output:  <output_dir>/<env>_chassis-idea-engine_context.txt
#
# Chassis Go paths follow package_page_build_debug.sh (platform/orchestration/…).
# Docs/SQL/engine-source resolve by name (newest variant) under their roots, with
# a name search fallback. Unresolved items print in a MISSING report. Needs GNU find.
#
# LIVE CAPTURE (read-only; --no-live to skip): \d of the agent/workflow tables +
# the EXISTING agent_definitions workflows (reuse-discovery, keep responsibilities distinct).
# ---------------------------------------------------------------------

set -e

# --- Self-locating (agentchassis repo root) --------------------------
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
REPO_ROOT="${REPO_ROOT:-$SCRIPT_DIR}"
if [ ! -f "$REPO_ROOT/go.mod" ] && [ -f "$PWD/go.mod" ]; then REPO_ROOT="$PWD"; fi
cd "$REPO_ROOT"

# --- Configuration (override any as env vars) ------------------------
DOCS024="$REPO_ROOT/docs/agent_docs/docs024_key_docs_latest"
DOC_SEARCH_ROOT="${DOC_SEARCH_ROOT:-$DOCS024}"            # chassis docs + method docs live here
IDEA_GO_DIR="${IDEA_GO_DIR:-$DOCS024/idea.uk/golang_files}"  # engine source
SQL_ROOT="${SQL_ROOT:-$REPO_ROOT}"                        # schemas searched here
PROJECT_ROOT="$REPO_ROOT"                                  # ./relative headers
ENVIRONMENT="production"
DEFAULT_OUTPUT_DIR="$DOCS024/idea.uk/docubundle_idea_within_chassis/package_module/output_contexts"
WITH_LIVE=true
WITH_DEBUGDOC=true

NOISE=( -not -path '*/.git/*' -not -path '*/vendor/*' -not -path '*/node_modules/*'
        -not -path '*/output_contexts/*' -not -path '*/old_golang_files/*'
        -not -path '*/python_files/*' -not -path '*/_iso/*' )

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
      echo "Env: REPO_ROOT DOC_SEARCH_ROOT SQL_ROOT IDEA_GO_DIR"
      exit 0 ;;
  esac
  shift
done

RESOLVED=(); MISSING=()

# --- Helpers ---------------------------------------------------------
function emit() {                # $1=path  $2=list_only
  local path=$1 lo=$2 rel
  rel="${path#$PROJECT_ROOT/}"
  echo "filepath = ./$rel" >> "$OUTPUT_FILE"
  if [ "$lo" = "true" ]; then echo "[File listed only - content not included]" >> "$OUTPUT_FILE"
  else cat "$path" >> "$OUTPUT_FILE"; fi
  echo "-------------------------------------------------" >> "$OUTPUT_FILE"
  RESOLVED+=("$rel")
}

# Logical doc/file name -> NEWEST matching variant (name + name(N).ext) by mtime.
function add_doc() {             # $1=name  $2=search_root  $3=list_only(optional)
  local want=$1 root=${2:-$DOC_SEARCH_ROOT} lo=${3:-false}
  local base stem ext pat2 hits path n
  base=$(basename "$want"); stem="${base%.*}"; ext="${base##*.}"
  if [ "$stem" = "$base" ]; then pat2="${stem}(*)"; else pat2="${stem}(*).${ext}"; fi
  hits=$(find "$root" -type f \( -name "$base" -o -name "$pat2" \) "${NOISE[@]}" \
           -printf '%T@ %p\n' 2>/dev/null | sort -rn || true)
  if [ -n "$hits" ]; then
    path=$(printf '%s\n' "$hits" | head -n1 | cut -d' ' -f2-)
    n=$(printf '%s\n' "$hits" | wc -l | tr -d ' ')
    [ "$n" -gt 1 ] && echo "  note: '$base' -> newest of $n variants: ${path#$PROJECT_ROOT/}" >&2
    emit "$path" "$lo"
  else MISSING+=("$want"); echo "  MISSING: $want" >&2; fi
}

# Explicit repo-relative path (chassis Go); exact, else basename search; else MISSING.
function add_path() {            # $1=repo-relative path  $2=list_only(optional)
  local want=$1 lo=${2:-false} path="" base
  if   [ -f "$PROJECT_ROOT/$want" ]; then path="$PROJECT_ROOT/$want"
  elif [ -f "$want" ];               then path="$want"
  else
    base=$(basename "$want")
    path=$(find "$PROJECT_ROOT" -type f -name "$base" "${NOISE[@]}" 2>/dev/null | head -n1 || true)
  fi
  if [ -n "$path" ] && [ -f "$path" ]; then emit "$path" "$lo"
  else MISSING+=("$want"); echo "  MISSING: $want" >&2; fi
}

function write_directory() {     # $1=dir
  local dir_path=$1
  if [ ! -d "$dir_path" ]; then echo "Warning: dir '$dir_path' not found. Skipping." >&2; MISSING+=("$dir_path/"); return; fi
  dir_path="${dir_path%/}"
  while IFS= read -r -d $'\0' file; do
    [ "$(realpath "$file" 2>/dev/null)" = "$(realpath "$OUTPUT_FILE" 2>/dev/null)" ] && continue
    emit "$file" "false"
  done < <(find "$dir_path" -type f "${NOISE[@]}" \
    -not -iname '*orig*' -not -name '*~' -not -name '*.bak' \
    -not -name 'go.sum' -not -name '*_test.go' -not -name 'idea' \
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
echo "  repo root:   $REPO_ROOT"
echo "  engine src:  $IDEA_GO_DIR"
echo "  output file: $OUTPUT_FILE"

# === A. THE ENGINE TO PORT (idea.uk source + method docs) ===========
# stages = future workflow steps: challenge audience -> generate across 4 lenses
# -> cut vs the free alternative -> verify via web search -> score (Risk col) -> rank.
for f in engine.go prompts.go audience_check.go; do add_doc "$f" "$IDEA_GO_DIR"; done
for f in idea_uk_method_v0.md idea_method_prompt.md idea_uk_testrun_v2.md \
         PARALLEL_engine_deployment_and_layer5.md CONSOLIDATION_where_it_all_fits.md; do
  add_doc "$f" "$DOC_SEARCH_ROOT"; done

# === B. THE CHASSIS FRAMEWORK ========================================
# B1. Agent-creation guidelines + current platform state (newest variant).
for f in 000_documentation_index.md 001_development_guide.md 002_system_architecture.md \
         003_contracts_and_standards.md 019_tool_library.md 020_tool_lifecycle.md \
         023_llm_quality_testing.md 009_model_infrastructure.md \
         running_notes_16_content_quality_and_internal_linking.md \
         HANDOFF_2026-06-09_sections_durability_and_content_quality.md; do
  add_doc "$f" "$DOC_SEARCH_ROOT"; done

# B2. Orchestration engine + action catalogue (Go) — REUSE, don't recreate.
write_directory "platform/orchestration/types"
write_directory "platform/orchestration/input_contracts"
CHASSIS_GO=(
  "platform/orchestration/datahelpers/content_search.go"     # search/extract for the verify stage
  "platform/orchestration/datahelpers/deep_search.go"
  "platform/orchestration/datahelpers/unified_extractor.go"
  "platform/orchestration/datahelpers/safe_unmarshal.go"
  "platform/orchestration/datahelpers/sql_helpers.go"
  "platform/orchestration/coordinator.go"
  "platform/orchestration/state.go"
  "platform/orchestration/helpers.go"
  "platform/orchestration/agent_error_log.go"
  "platform/orchestration/actions/registry.go"               # catalogue: every registered action
  "platform/orchestration/actions/types.go"
  "platform/orchestration/actions/helpers.go"
  "platform/orchestration/actions/workflow_actions.go"
  "platform/orchestration/actions/conditional_branch_action.go"
  "platform/orchestration/actions/basic_actions.go"
  "platform/orchestration/actions/generic_actions.go"
  "platform/orchestration/actions/spawn_actions.go"          # spawn sub-agents (no-subworkflows rule)
  "platform/orchestration/actions/call_agent.go"
  "platform/orchestration/actions/await_response.go"
  "platform/database/postgres.go"
  "cmd/agent-chassis/main.go"
)
for f in "${CHASSIS_GO[@]}"; do add_path "$f"; done
# The LLM-call action (engine stages are prompts). Filename varies → match by pattern.
while IFS= read -r -d $'\0' f; do emit "$f" "false"; done < <(
  find "$PROJECT_ROOT/platform/orchestration/actions" -maxdepth 1 -type f -name '*.go' \
    \( -iname '*llm*' -o -iname '*prompt*' \) -not -name '*_test.go' -print0 2>/dev/null)

# B3. Schemas — CHECK THE SCHEMA before writing SQL (newest variant).
for f in 002_intake_orchestrator.sql agent_definition_types.sql 018_briefing_questionnaire.sql \
         019_model_lifecycle_schema.sql 021_model_swap_and_rollback.sql schemas_all schemas_some; do
  add_doc "$f" "$SQL_ROOT"; done

# B4. Seed / table content — the message envelope (headers+body) pattern.
for f in initial_messages__without_current_ids_ initial_vet_practice_check_message; do
  add_doc "$f" "$SQL_ROOT"; done

# === C. FOR THE LATER SITE-BUILD MERGE (listed only) =================
for f in 021_site_spec_and_classifier.md 029_site_plan_and_reconciler.md; do add_doc "$f" "$DOC_SEARCH_ROOT" "true"; done
for f in platform/orchestration/actions/v3_site_actions.go \
         platform/orchestration/actions/write_site_plan_action.go \
         platform/orchestration/actions/site_spec_actions.go \
         platform/orchestration/actions/site_db_actions.go; do add_path "$f" "true"; done

# === Debug guide (large; skip with --no-debugdoc) ====================
if [ "$WITH_DEBUGDOC" = true ]; then add_doc "016_debugging_guide_v2.md" "$DOC_SEARCH_ROOT"
else echo "  (skipping 016_debugging_guide_v2.md per --no-debugdoc)"; fi

# === Live capture (read-only reuse-discovery) ========================
if [ "$WITH_LIVE" = true ] && command -v kubectl >/dev/null 2>&1; then
  echo "Capturing live data (read-only; --no-live to skip)…"
  set +e
  PG="kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db"
  { echo ""; echo "================================================================="
    echo "LIVE CAPTURE — $(date -u +%Y-%m-%dT%H:%M:%SZ)  (read-only)"
    echo "================================================================="
    echo ""; echo "----- SCHEMA (\\d) of the agent/workflow tables -----"; } >> "$OUTPUT_FILE"
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
SELECT type, default_config FROM agent_definitions ORDER BY type LIMIT 5;
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
  echo "MISSING (${#MISSING[@]}) — fix REPO_ROOT/DOC_SEARCH_ROOT/SQL_ROOT/IDEA_GO_DIR, or the path:"
  printf '  - %s\n' "${MISSING[@]}"
fi
echo "Done. Context saved to $OUTPUT_FILE"
echo "File size: $(du -h "$OUTPUT_FILE" | cut -f1)"
