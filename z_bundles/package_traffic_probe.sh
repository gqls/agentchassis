#!/bin/bash
#
# package_traffic_probe.sh
#   Self-contained packager for the DOMAIN TRAFFIC-PROBE context (a SEPARATE
#   project from idea.uk that reuses idea.uk's VM/nginx/Go deployment model).
#   Bundles: the task brief + the domain list, the reusable Go service to fork,
#   and the deploy/persistence/VM docs — into one context file for a new chat.
#
#   No live capture: the probe boxes don't exist yet, there's no DB, not on k8s.
#   Copes with the messy folder: docs resolve to the NEWEST (N) variant by mtime;
#   the code walk drops *.orig* backups + the idea binary.
#
# Usage:  ./package_traffic_probe.sh [-o output_dir] [--no-debugdoc]
#   Run from the repo root (go.mod), or set REPO_ROOT.
#   Put the brief + domain list in  docs024_key_docs_latest/traffic_probe/  (or set TASK_DIR).
#   Env: REPO_ROOT TASK_DIR DOC_SEARCH_ROOT IDEA_GO_DIR
#
# Output:  <output_dir>/traffic-probe_context.txt   (default under traffic_probe/output_contexts)
#
# Requires GNU find (Linux) for -printf.
# ---------------------------------------------------------------------

set -e

# --- Self-locating (agentchassis repo root) --------------------------
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
REPO_ROOT="${REPO_ROOT:-$SCRIPT_DIR}"
if [ ! -f "$REPO_ROOT/go.mod" ] && [ -f "$PWD/go.mod" ]; then REPO_ROOT="$PWD"; fi

# --- Configuration (override any as env vars) ------------------------
DOCS024="$REPO_ROOT/docs/agent_docs/docs024_key_docs_latest"
TASK_DIR="${TASK_DIR:-$DOCS024/domains_relojistas}"          # where the brief + domain list live
DOC_SEARCH_ROOT="${DOC_SEARCH_ROOT:-$DOCS024}"          # docs tree (searched for newest variants)
IDEA_GO_DIR="${IDEA_GO_DIR:-$DOCS024/idea.uk/golang_files}"  # the reusable Go service to fork
PROJECT_ROOT="$REPO_ROOT"                                # ./relative headers are repo-relative
DEFAULT_OUTPUT_DIR="$TASK_DIR/docubundle/output_contexts"
WITH_DEBUGDOC=true

NOISE=( -not -path '*/.git/*' -not -path '*/vendor/*' -not -path '*/node_modules/*'
        -not -path '*/output_contexts/*' -not -path '*/old_golang_files/*'
        -not -path '*/python_files/*' -not -path '*/_iso/*' )

# --- Argument parsing ------------------------------------------------
OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do
  case $1 in
    -o | --output)  shift; OUTPUT_DIR="$1" ;;
    --no-debugdoc)  WITH_DEBUGDOC=false ;;
    -h | --help)
      echo "Usage: $0 [-o output_dir] [--no-debugdoc]"
      echo "Env: REPO_ROOT TASK_DIR DOC_SEARCH_ROOT IDEA_GO_DIR"
      echo "Put TASK_traffic_probe_brief.md + traffic_probe_domains.tsv in \$TASK_DIR."
      exit 0 ;;
  esac
  shift
done

RESOLVED=(); MISSING=()

# --- Helpers (same resolution logic as the other packagers) ----------
function emit() {                # $1=path  $2=list_only
  local path=$1 lo=$2 rel
  rel="${path#$PROJECT_ROOT/}"
  echo "filepath = ./$rel" >> "$OUTPUT_FILE"
  if [ "$lo" = "true" ]; then echo "[File listed only - content not included]" >> "$OUTPUT_FILE"
  else cat "$path" >> "$OUTPUT_FILE"; fi
  echo "-------------------------------------------------" >> "$OUTPUT_FILE"
  RESOLVED+=("$rel")
}

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
OUTPUT_FILE="${OUTPUT_DIR}/traffic-probe_context.txt"
> "$OUTPUT_FILE"

echo "Packaging domain traffic-probe context"
echo "  repo root:   $REPO_ROOT"
echo "  task dir:    $TASK_DIR"
echo "  engine src:  $IDEA_GO_DIR"
echo "  output file: $OUTPUT_FILE"

# 1) The new task definition: brief + domain list.
add_doc "TASK_traffic_probe_brief.md" "$TASK_DIR"
add_doc "traffic_probe_domains.tsv"   "$TASK_DIR"

# 2) The reusable VM/nginx/Go service to fork (drops *.orig + the binary).
#    Keep service.go/store.go/page.html/main.go/deploy/; the idea.uk-specific
#    engine/prompts/billing/audience_check ride along as Go+Anthropic reference.
write_directory "$IDEA_GO_DIR"

# 3) Deploy / persistence / VM docs (newest variant under the docs tree).
for f in idea_uk_architecture_and_deployment.md VM_LAUNCH_PLAN.md README_setup_box.md \
         PERSISTENCE_design.md PLAN_checkpoint_and_artefact_upload_b2.md; do
  add_doc "$f" "$DOC_SEARCH_ROOT"; done

# 4) The large deploy/debug guide (skip with --no-debugdoc).
if [ "$WITH_DEBUGDOC" = true ]; then add_doc "016_debugging_guide_v2_32.md" "$DOC_SEARCH_ROOT"
else echo "  (skipping 016_debugging_guide_v2_32.md per --no-debugdoc)"; fi

# --- Report ----------------------------------------------------------
echo ""
echo "Included ${#RESOLVED[@]} files."
if [ "${#MISSING[@]}" -gt 0 ]; then
  echo "MISSING (${#MISSING[@]}) — fix REPO_ROOT/TASK_DIR/DOC_SEARCH_ROOT/IDEA_GO_DIR, or drop the file in:"
  printf '  - %s\n' "${MISSING[@]}"
fi
echo "Done. Context saved to $OUTPUT_FILE"
echo "File size: $(du -h "$OUTPUT_FILE" | cut -f1)"
