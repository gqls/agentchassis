#!/bin/bash
#
# package_idea_uk_golive.sh
#   Self-contained packager for the IDEA.UK GO-LIVE context.
#   Bundles the idea.uk service (Go code + embedded page + deploy) and the
#   go-live docs (handoff, architecture, email, liability, Stripe plan, the
#   §11 deploy guide) into one context file for an AI assistant.
#
#   idea.uk is a single self-contained Go binary with file-based persistence
#   (store.go) — NO database, NO SQL/schema, NOT on Kubernetes. No live capture.
#
#   COPES WITH A MESSY FOLDER: the idea.uk docs accumulate as (N)-versioned
#   copies where the un-suffixed name is the OLDEST, so each doc is resolved to
#   the NEWEST variant by mtime. The code walk drops *.orig* backups + the binary.
#
# Usage:  ./package_idea_uk_golive.sh [-o output_dir] [--no-debugdoc]
#   Run it from the repo root (where go.mod is), or set REPO_ROOT / IDEA_ROOT.
#   Env overrides: REPO_ROOT, IDEA_ROOT, CODE_DIR, DOC_SEARCH_ROOT
#
# Output:  <output_dir>/idea-uk-golive_context.txt
#   (default under idea.uk/docubundle_idea_golive/package_module/output_contexts)
#
# Requires GNU find (Linux) for -printf. --no-debugdoc drops the large guide.
# ---------------------------------------------------------------------

set -e

# --- Self-locating (find the agentchassis repo root) -----------------
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
REPO_ROOT="${REPO_ROOT:-$SCRIPT_DIR}"
if [ ! -f "$REPO_ROOT/go.mod" ] && [ -f "$PWD/go.mod" ]; then REPO_ROOT="$PWD"; fi

# --- Configuration (override any as env vars) ------------------------
DOCS024="$REPO_ROOT/docs/agent_docs/docs024_key_docs_latest"
IDEA_ROOT="${IDEA_ROOT:-$DOCS024/idea.uk}"          # the idea.uk folder
CODE_DIR="${CODE_DIR:-$IDEA_ROOT/golang_files}"     # the Go service
DOC_SEARCH_ROOT="${DOC_SEARCH_ROOT:-$DOCS024}"      # docs tree (searched for newest variants)
PROJECT_ROOT="$REPO_ROOT"                            # ./relative headers are repo-relative
DEFAULT_OUTPUT_DIR="$IDEA_ROOT/docubundle_idea_golive/package_module/output_contexts"
WITH_DEBUGDOC=true

# Directories never worth walking/searching (old copies, isolation blobs, prior bundles).
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
      echo "Env: REPO_ROOT IDEA_ROOT CODE_DIR DOC_SEARCH_ROOT"
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

# Resolve a logical doc name to the NEWEST matching variant (name + name(N).ext).
function add_doc() {             # $1=logical name  $2=search_root  $3=list_only(optional)
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

# Walk a directory: every non-binary, non-test, non-.orig file.
function write_directory() {     # $1=dir
  local dir_path=$1
  if [ ! -d "$dir_path" ]; then echo "Warning: directory '$dir_path' not found. Skipping." >&2; MISSING+=("$dir_path/"); return; fi
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
OUTPUT_FILE="${OUTPUT_DIR}/idea-uk-golive_context.txt"
> "$OUTPUT_FILE"

echo "Packaging idea.uk go-live context"
echo "  repo root:   $REPO_ROOT"
echo "  idea root:   $IDEA_ROOT"
echo "  code dir:    $CODE_DIR"
echo "  output file: $OUTPUT_FILE"

# 1) The Go service (code + embedded page + deploy/). Tests + *.orig excluded.
#    This also captures terms_preview.html / privacy_preview.html (they live here).
write_directory "$CODE_DIR"

# 2) Go-live docs — each resolved to its NEWEST (N) variant under the docs tree.
DOCS=(
  "HANDOFF.md"
  "running_notes.md"
  "idea_uk_architecture_and_deployment.md"
  "EMAIL_identity_in_site_spec.md"
  "LIABILITY_AND_TERMS.md"
  "PLAN_stripe_billing_integration.md"
  "RUNBOOK_idea_uk.md"
  "DEVELOPMENT_RUNBOOK.md"
  "leopardess_uk_index.html"
)
for d in "${DOCS[@]}"; do add_doc "$d"; done

# 3) The large deploy/debug guide (skip with --no-debugdoc).
if [ "$WITH_DEBUGDOC" = true ]; then add_doc "016_debugging_guide_v2_32.md"
else echo "  (skipping 016_debugging_guide_v2_32.md per --no-debugdoc)"; fi

# --- Report ----------------------------------------------------------
echo ""
echo "Included ${#RESOLVED[@]} files."
if [ "${#MISSING[@]}" -gt 0 ]; then
  echo "MISSING (${#MISSING[@]}) — adjust REPO_ROOT/IDEA_ROOT/CODE_DIR/DOC_SEARCH_ROOT, or sync the file in:"
  printf '  - %s\n' "${MISSING[@]}"
fi
echo "Done. Context saved to $OUTPUT_FILE"
echo "File size: $(du -h "$OUTPUT_FILE" | cut -f1)"
