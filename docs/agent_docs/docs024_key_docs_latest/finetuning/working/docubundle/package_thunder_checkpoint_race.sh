#!/bin/bash
#
# package_thunder_checkpoint_race.sh
#   Self-contained packager for the FLYWHEEL-C CHECKPOINT-UPLOAD / LOOP-AWAIT-RACE context.
#   Bundles the async-await + loop machinery of agent-chassis plus the checkpoint-upload
#   path into one context file for an AI assistant, copies the working docs alongside, and
#   (optionally) appends a read-only live capture: schema, the decisive presign/export/await
#   queries, the training-launcher/model-trainer/data-preparer workflows, and runtime state.
#
#   Patterned on package_page_build_debug.sh: same self-locating wrapper + helpers; one
#   hardcoded module scoped to this blocker.
#
# Usage:  ./package_thunder_checkpoint_race.sh [-o output_dir] [-e env] [--no-live] [--tar]
# Example: ./package_thunder_checkpoint_race.sh
#          ./package_thunder_checkpoint_race.sh --no-live      # code+docs only, no cluster
#          ./package_thunder_checkpoint_race.sh --tar          # also produce a .tar.gz
#
# Output:  <output_dir>/<environment>_thunder-checkpoint-race_context.txt   (code + live)
#          <output_dir>/*.md                                                (docs copied in)
#
# ---------------------------------------------------------------------
# SCOPE (the blast radius of the open thread):
#   - presign_checkpoints loop stalls intermittently at a later iteration (a RACE):
#     the local dispatch (dispatch_thunder_prepare_object_url) is SEND-before-REGISTER, so
#     a ~1s adapter reply can beat the awaited_requests INSERT and ClaimAwaitedRequest
#     (WHERE status='waiting') drops it -> timeout -> retry -> same race, forever.
#   - THE FIX (non-framework): pre-register the awaited request in the dispatch BEFORE the
#     send, reusing preRegisterAwaitedRequest (what spawn/call already do).
#   - FALLBACK (structural): one batch prepare_object_urls adapter call instead of the loop.
#
# REUSE-DISCOVERY (so we don't reinvent existing code): keeps registry.go (every registered
#   action), the await/loop machinery in coordinator.go + state.go, await_response.go,
#   spawn_actions.go (the helper to mirror), and call_agent.go (the register-before-send
#   contrast).
#
# PATHS TO VERIFY before/while running (this script does NOT know your exact tree):
#   - chassis actions are assumed under platform/orchestration/actions/ (per the page-build
#     packager). add_file() falls back to a find-by-basename if a literal path misses, so
#     distinctively-named files (thunder_prepare_object_url_dispatch.go, compute_checkpoint_
#     keys_action.go, ...) are located even if the path guess is wrong.
#   - the THUNDER-ADAPTER is a SEPARATE service: the adapter/storage paths below are
#     best-guess and literal-only (generic names like adapter.go/interface.go are NOT
#     find-resolved, to avoid grabbing the wrong file). If the adapter lives in its own repo,
#     run this script there too, or add those files manually. They are FALLBACK-route only.
#
# LIVE CAPTURE (read-only) is appended by default when kubectl is present:
#   \d awaited_requests / agent_definitions / model_lifecycle.* / training_exports.* ;
#   presign_one binding check ; export-rows LEFT JOIN ; current awaited_requests rows ;
#   recent training_runs ; live thunder_instances ; the three agent workflows ; pods/jobs.
#   Disable with --no-live.
# ---------------------------------------------------------------------

set -e

# --- Self-locating logic (verbatim from package_page_build_debug.sh) ---
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT="${PROJECT_ROOT:-$(realpath "$SCRIPT_DIR/../../")}"
if [ ! -f "$PROJECT_ROOT/go.mod" ] && [ -f "$PWD/go.mod" ]; then
  PROJECT_ROOT="$PWD"
fi
cd "$PROJECT_ROOT"

# --- Configuration ---------------------------------------------------
DEFAULT_OUTPUT_DIR="$SCRIPT_DIR/docs/agent_docs/docs024_key_docs_latest/finetuning/working/context_packages/thunder-checkpoint-race"
ENVIRONMENT="production"                       # only affects the output filename
COMPONENT_NAME="thunder-checkpoint-race"       # fixed; this script packages one module
WITH_LIVE=true                                 # append read-only live capture if kubectl present
MAKE_TAR=false                                 # also produce a .tar.gz of the package dir

# Where your markdown docs live in the repo (a hint for the copy step). If empty or wrong,
# the doc copy falls back to a find-by-basename under PROJECT_ROOT.
DOCS_SEARCH_ROOT="${DOCS_SEARCH_ROOT:-$PROJECT_ROOT}"

# --- Argument parsing ------------------------------------------------
OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do
  case $1 in
    -o | --output)      shift; OUTPUT_DIR="$1" ;;
    -e | --environment) shift; ENVIRONMENT="$1" ;;
    --no-live)          WITH_LIVE=false ;;
    --tar)              MAKE_TAR=true ;;
    -h | --help)
      echo "Usage: $0 [-o output_dir] [-e environment] [--no-live] [--tar]"
      echo "Packages the thunder checkpoint-upload / loop-await-race code + docs, plus a"
      echo "read-only live capture (schema/data/workflows/runtime) unless --no-live is given."
      exit 0
      ;;
  esac
  shift
done

# --- Helper functions (verbatim from package_page_build_debug.sh) ----
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

function write_directory() {
  local dir_path=$1
  local output_file=$2
  if [ ! -d "$dir_path" ]; then
    echo "Warning: Directory '$dir_path' not found in '$PWD'. Skipping." >&2
    return
  fi
  dir_path="${dir_path%/}"
  while IFS= read -r -d $'\0' file; do
    if [ "$(realpath "$file" 2>/dev/null)" = "$(realpath "$output_file" 2>/dev/null)" ]; then
      continue
    fi
    write_file "$file" "$output_file" "false"
  done < <(find "$dir_path" -type f \
    -not -path '*/.git/*' -not -path '*/vendor/*' -not -path '*/node_modules/*' \
    -not -path '*/output_contexts/*' -not -path '*/context_packages/*' \
    -not -name '*.log' -not -name 'go.sum' -not -name '*_test.go' \
    -not -name '.DS_Store' -print0)
}

# --- Extra helper: literal path first, else find-by-basename ---------
# Use ONLY for distinctively-named files whose exact path is uncertain.
function add_file() {
  local want=$1 output_file=$2
  local f=""
  if [ -f "$want" ]; then
    f="$want"
  else
    f=$(find . -type f -name "$(basename "$want")" \
        -not -path '*/.git/*' -not -path '*/vendor/*' -not -path '*/node_modules/*' \
        -not -path '*/output_contexts/*' -not -path '*/context_packages/*' \
        2>/dev/null | head -1)
    f="${f#./}"
    [ -n "$f" ] && echo "Note: '$want' not at literal path; using found '$f'." >&2
  fi
  if [ -n "$f" ] && [ -f "$f" ]; then
    write_file "${f#./}" "$output_file" "false"
  else
    echo "Warning: '$want' not found (by path or basename). Skipping." >&2
  fi
}

# --- Module definition: thunder-checkpoint-race ----------------------
# (1) Distinctively-named files — add_file (literal, else find-by-basename).
SRC_FILES_FUZZY=(
  # --- Orchestration engine: the async-await + loop machinery (the race lives here) ---
  "platform/orchestration/coordinator.go"            # ProcessResponse, handleCompleteResponse, continueExecution, handleRequestTimeout, route/failWorkflow, setLoopVariable, propagateIterationOutputs, shouldContinueLoopOnError, skipToNextLoopIterationForAsync, getTimeout
  "platform/orchestration/loop_expansion_handler.go" # setLoopVariable / propagateIterationOutputs loop-item mechanics
  "platform/orchestration/loop_error_handler.go"     # continue_on_error / skip path

  # --- Actions: await registration + the fix helper + the contrast + reuse-discovery ---
  "platform/orchestration/actions/await_response.go" # processAwaitResponse / createAwaitedRequest / extractRequestID / determineResponsesTopic (POST-return registration == send-before-register)
  "platform/orchestration/actions/spawn_actions.go"  # preRegisterAwaitedRequest (THE helper the fix reuses) + the register-before-send call site to mirror
  "platform/orchestration/actions/call_agent.go"     # register-before-send contrast (why call_agent loops don't race)
  "platform/orchestration/actions/registry.go"       # every registered action (reuse-discovery)

  # --- The file the fix EDITS + the checkpoint-upload path (chassis side) ---
  "platform/orchestration/actions/thunder_prepare_object_url_dispatch.go"  # FIX GOES HERE: call preRegisterAwaitedRequest just before ProduceWithValidation
  "platform/orchestration/actions/compute_checkpoint_keys_action.go"       # emits all 40 checkpoint keys (Phase B; also feeds the batch fallback)
  "platform/orchestration/actions/assemble_upload_manifest_action.go"      # Phase B
  "platform/orchestration/actions/flatten_presign_results_action.go"       # Phase B (loop flatten; the batch fallback removes this)
)

# (2) Generic-named engine/shared files — literal path only (no find; avoids wrong matches).
SRC_FILES_LITERAL=(
  "platform/orchestration/state.go"                  # StateRepository: ClaimAwaitedRequest / InsertAwaitedRequest / GetAwaitedRequest(+WithRetry) / GetAwaitedRequestStatus (awaited_requests table)
  "platform/orchestration/helpers.go"
  "platform/orchestration/actions/types.go"          # ActionParams — VERIFY CurrentStep holds the expanded loop-substep name at dispatch time
  "platform/orchestration/actions/helpers.go"
  "platform/database/postgres.go"
)

# (3) FALLBACK route only — the thunder-adapter (SEPARATE service). Best-guess paths,
#     literal-only. Adjust if your tree differs; may live in the adapter's own repo.
ADAPTER_FILES_LITERAL=(
  "services/thunder-adapter/adapter.go"                  # request routing + sendSuccessResponse (reply path)
  "services/thunder-adapter/data_url_actions.go"         # prepare_object_url handler; where prepare_object_urls (plural) would go
  "services/thunder-adapter/storage/s3.go"               # presign helpers
  "services/thunder-adapter/storage/url_helpers.go"      # GetPresignedPutURL / PresignPutObject
  "services/thunder-adapter/storage/interface.go"
)

# --- Package the code ------------------------------------------------
mkdir -p "$OUTPUT_DIR"
OUTPUT_FILE="${OUTPUT_DIR}/${ENVIRONMENT}_${COMPONENT_NAME}_context.txt"
> "$OUTPUT_FILE"

echo "Packaging '$COMPONENT_NAME' for environment '$ENVIRONMENT'"
echo "  project root: $PROJECT_ROOT"
echo "  output file:  $OUTPUT_FILE"

{
  echo "================================================================="
  echo "THUNDER CHECKPOINT-UPLOAD / LOOP-AWAIT-RACE — code context"
  echo "Fix: pre-register the awaited request in thunder_prepare_object_url_dispatch.go"
  echo "     (preRegisterAwaitedRequest) BEFORE ProduceWithValidation. See HANDOFF."
  echo "================================================================="
  echo ""
} >> "$OUTPUT_FILE"

for item in "${SRC_FILES_FUZZY[@]}";    do add_file "$item" "$OUTPUT_FILE"; done
for item in "${SRC_FILES_LITERAL[@]}";  do write_file "$item" "$OUTPUT_FILE" "false"; done
{ echo ""; echo "----- FALLBACK ROUTE: thunder-adapter (separate service; verify paths) -----"; } >> "$OUTPUT_FILE"
for item in "${ADAPTER_FILES_LITERAL[@]}"; do write_file "$item" "$OUTPUT_FILE" "false"; done

# --- Copy the working docs alongside the context file ----------------
# DOCS_FILES: basenames (or repo-relative paths). Located via literal path, else
# find-by-basename under DOCS_SEARCH_ROOT, then copied into the package dir.
#
# >>> PROPOSED SET — confirm/replace per the doc list in chat. The first group are the
#     versions uploaded 2026-06-08; the guideline versions are placeholders to confirm. <<<
DOCS_FILES=(
  # thread working docs (uploaded versions)
  "HANDOFF_2026-06-06_checkpoint_upload_loop_await_race.md"
  "docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase5/NOTES_phase5_training_launcher_running(35).md"
  "docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase5/RUNBOOK_phase_b_c_d_deploy(4).md"
  "docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase5/PLAN_checkpoint_and_artefact_upload_b2(4).md"
  "CONTEXT_PACK_thunder_checkpoint_race.md"
  "docs/agent_docs/docs024_key_docs_latest/finetuning/working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md"
  # fallback-route docs
  "STATUS_thunder_adapter_2026-06_04.md"
  "docs/agent_docs/docs024_key_docs_latest/finetuning/working/flywheel_docs/FOCUS_adapter_design(3).md"
  # guidelines (the constitution) — confirmed paths under docs/agent_docs/docs024_key_docs_latest/
  "docs/agent_docs/docs024_key_docs_latest/001_development_guide(3).md"
  "docs/agent_docs/docs024_key_docs_latest/002_system_architecture.md"
  "docs/agent_docs/docs024_key_docs_latest/003_contracts_and_standards.md"
  "docs/agent_docs/docs024_key_docs_latest/016_debugging_guide_v2_35.md"   # merged v34 + the loop-await race branch; replaces v2_33/34
  # the targeted code extract (optional; the context .txt above supersedes it)
  # "CHASSIS_await_loop_extract.txt"
)

echo "Copying docs into the package dir…"
for d in "${DOCS_FILES[@]}"; do
  src=""
  if [ -f "$d" ]; then
    src="$d"
  else
    src=$(find "$DOCS_SEARCH_ROOT" -type f -name "$(basename "$d")" \
          -not -path '*/.git/*' -not -path '*/context_packages/*' 2>/dev/null | head -1)
  fi
  if [ -n "$src" ] && [ -f "$src" ]; then
    cp -f "$src" "$OUTPUT_DIR/$(basename "$d")"
    echo "  + $(basename "$d")"
  else
    echo "  ! MISSING: $d (add manually)" >&2
  fi
done

# --- Optional live capture (read-only) -------------------------------
if [ "$WITH_LIVE" = true ] && command -v kubectl >/dev/null 2>&1; then
  echo "Capturing live data (read-only; disable with --no-live)…"
  set +e
  PG="kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db"

  {
    echo ""
    echo "================================================================="
    echo "LIVE CAPTURE — $(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "read-only: \\d schema, SELECTs, kubectl get"
    echo "================================================================="
  } >> "$OUTPUT_FILE"

  # 1) SCHEMA (\d) — awaited_requests is the central table for the race
  { echo ""; echo "----- SCHEMA (\\d) -----"; } >> "$OUTPUT_FILE"
  $PG >> "$OUTPUT_FILE" 2>&1 <<'SQL'
\d awaited_requests
\d agent_definitions
\d model_lifecycle.training_runs
\d model_lifecycle.thunder_instances
\d training_exports.runs
\d training_exports.rows
SQL

  # 2) LIVE DATA — the decisive checks
  { echo ""; echo "----- LIVE DATA -----"; } >> "$OUTPUT_FILE"
  $PG >> "$OUTPUT_FILE" 2>&1 <<'SQL'
\echo '== presign_one binding (expect key_path:"ckpt_key", NO input_mapping) =='
SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_checkpoints,config,sub_workflow,steps,presign_one,config}') AS presign_one_config
FROM agent_definitions
WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

\echo '== export rows (pick a NON-ZERO actual_rows; 146a9a12/fef7be6b good, a8484922 = 0) =='
SELECT r.id, r.rows_exported, count(x.*) AS actual_rows
FROM training_exports.runs r
LEFT JOIN training_exports.rows x ON x.export_id = r.id
GROUP BY r.id ORDER BY r.created_at DESC LIMIT 20;

\echo '== awaited_requests: current rows (during a run, shows stuck iterations) =='
SELECT request_id, orchestration_id, step_name, status, retry_version,
       sent_at, timeout_at, target_agent_type
FROM awaited_requests ORDER BY sent_at DESC LIMIT 40;

\echo '== recent training_runs =='
SELECT id, status, created_at, started_at, thunder_instance_id, export_id
FROM model_lifecycle.training_runs ORDER BY created_at DESC LIMIT 20;

\echo '== live thunder_instances (should be empty after cleanup) =='
SELECT id, status, training_run_id, created_at
FROM model_lifecycle.thunder_instances
WHERE status NOT IN ('decommissioned','reaped','failed') ORDER BY created_at DESC;
SQL

  # 3) AGENT WORKFLOWS (the launcher loop + its orchestrator + data-preparer)
  { echo ""; echo "----- AGENT WORKFLOWS (agent_definitions.default_config) -----"; } >> "$OUTPUT_FILE"
  $PG >> "$OUTPUT_FILE" 2>&1 <<'SQL'
SELECT type, jsonb_pretty(default_config) AS default_config
FROM agent_definitions
WHERE type IN ('training-launcher','model-trainer','data-preparer')
  AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL
ORDER BY type;
SQL

  # 4) RUNTIME (is anything stuck / are boxes live?)
  { echo ""; echo "----- RUNTIME (kubectl) -----"; } >> "$OUTPUT_FILE"
  echo "\$ kubectl -n ai-persona-system get pods | grep -iE 'training-launcher|thunder-adapter|model-trainer|data-preparer|provisioner'" >> "$OUTPUT_FILE"
  kubectl -n ai-persona-system get pods 2>&1 | grep -iE 'training-launcher|thunder-adapter|model-trainer|data-preparer|provisioner' >> "$OUTPUT_FILE" 2>&1
  { echo ""; echo "\$ kubectl -n ai-persona-system get jobs | grep -iE 'training-launcher|data-prep|provision|model-train'"; } >> "$OUTPUT_FILE"
  kubectl -n ai-persona-system get jobs 2>&1 | grep -iE 'training-launcher|data-prep|provision|model-train' >> "$OUTPUT_FILE" 2>&1
  { echo ""; echo "\$ kubectl -n kafka get pods"; } >> "$OUTPUT_FILE"
  kubectl -n kafka get pods >> "$OUTPUT_FILE" 2>&1

  set -e
  echo "  live capture appended."
elif [ "$WITH_LIVE" = true ]; then
  echo "Skipping live capture: kubectl not found on PATH (run with --no-live to silence)."
fi

# --- Optional tarball ------------------------------------------------
if [ "$MAKE_TAR" = true ]; then
  TAR_PATH="${OUTPUT_DIR%/}.tar.gz"
  tar -czf "$TAR_PATH" -C "$(dirname "$OUTPUT_DIR")" "$(basename "$OUTPUT_DIR")"
  echo "📦 Tarball: $TAR_PATH"
fi

echo "✅ Done."
echo "   Context : $OUTPUT_FILE  ($(du -h "$OUTPUT_FILE" | cut -f1))"
echo "   Docs    : $(ls -1 "$OUTPUT_DIR"/*.md 2>/dev/null | wc -l | tr -d ' ') markdown file(s) in $OUTPUT_DIR"
echo ""
echo "Sanity check the fix helper made it in:"
echo "  grep -n 'func preRegisterAwaitedRequest' \"$OUTPUT_FILE\" || echo 'MISSING — add the file that defines it'"
