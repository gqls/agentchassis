#!/bin/bash
#
# package_idea_uk_golive.sh
#   Self-contained packager for the IDEA.UK GO-LIVE context.
#   Bundles the idea.uk service (Go code + embedded page + deploy) and the
#   go-live docs (handoff, architecture, email, liability, Stripe plan, the
#   §11 deploy guide) into one context file for an AI assistant.
#
#   idea.uk is a single self-contained Go binary with file-based persistence
#   (store.go) — there is NO database, NO SQL/schema, and it is NOT on
#   Kubernetes (it runs on a Hetzner box). So there is NO live capture here.
#
# Usage:  ./package_idea_uk_golive.sh [-o output_dir] [--no-debugdoc]
#   IDEA_ROOT=/path/to/idea.uk ./package_idea_uk_golive.sh   # override the root
#
# Layout it expects (override with the env vars below):
#   IDEA_ROOT/                 (default: this script's own directory)
#     golang_files/            -> the Go service (CODE_DIR)
#     *.md / *.html            -> the docs (searched under DOCS_ROOT)
#
# Output:  <output_dir>/idea-uk-golive_context.txt
#
# --no-debugdoc drops 016_debugging_guide_v2_32.md (the large one) for a
# leaner bundle; everything else is small.
# ---------------------------------------------------------------------

set -e

# --- Self-locating logic ---------------------------------------------
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# --- Configuration (override any of these as env vars) ---------------
IDEA_ROOT="${IDEA_ROOT:-$SCRIPT_DIR}"          # folder holding golang_files/ + the docs
CODE_DIR="${CODE_DIR:-$IDEA_ROOT/golang_files}"  # the Go service
DOCS_ROOT="${DOCS_ROOT:-$IDEA_ROOT}"           # where the .md/.html docs live (searched)
PROJECT_ROOT="$IDEA_ROOT"                        # for clean ./relative headers
DEFAULT_OUTPUT_DIR="$IDEA_ROOT/output_contexts"
WITH_DEBUGDOC=true

# --- Argument parsing ------------------------------------------------
OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"
while [[ "$1" =~ ^- && ! "$1" == "--" ]]; do
  case $1 in
    -o | --output)  shift; OUTPUT_DIR="$1" ;;
    --no-debugdoc)  WITH_DEBUGDOC=false ;;
    -h | --help)
      echo "Usage: $0 [-o output_dir] [--no-debugdoc]"
      echo "Env:   IDEA_ROOT, CODE_DIR, DOCS_ROOT may be set to override paths."
      echo "Bundles the idea.uk go-live code + docs (no live capture: no DB, not on k8s)."
      exit 0
      ;;
  esac
  shift
done

RESOLVED=()
MISSING=()

# --- Helpers ---------------------------------------------------------

# Write one file: a "filepath = ./rel" header, the content, a separator.
function emit() {
  local path=$1 list_only=$2
  local rel="${path#$PROJECT_ROOT/}"
  echo "filepath = ./$rel" >> "$OUTPUT_FILE"
  if [ "$list_only" = "true" ]; then
    echo "[File listed only - content not included]" >> "$OUTPUT_FILE"
  else
    cat "$path" >> "$OUTPUT_FILE"
  fi
  echo "-------------------------------------------------" >> "$OUTPUT_FILE"
  RESOLVED+=("$rel")
}

# Resolve a wanted item (path or bare filename) under a search root; if the
# exact path misses, fall back to a name search; record MISSING if not found.
function add_named() {           # $1=want  $2=search_root  $3=list_only(optional)
  local want=$1 root=${2:-$PROJECT_ROOT} list_only=${3:-false} path="" base
  base=$(basename "$want")
  if   [ -f "$root/$want" ]; then path="$root/$want"
  elif [ -f "$want" ];       then path="$want"
  else
    local hits n
    hits=$(find "$root" -type f -name "$base" \
      -not -path '*/.git/*' -not -path '*/vendor/*' -not -path '*/node_modules/*' \
      -not -path '*/output_contexts/*' 2>/dev/null || true)
    if [ -n "$hits" ]; then
      path=$(printf '%s\n' "$hits" | head -n1)
      n=$(printf '%s\n' "$hits" | wc -l | tr -d ' ')
      [ "$n" -gt 1 ] && echo "  note: '$base' matched $n files; using ${path#$root/}" >&2
    fi
  fi
  if [ -n "$path" ] && [ -f "$path" ]; then emit "$path" "$list_only"
  else MISSING+=("$want"); echo "  MISSING: $want" >&2; fi
}

# Write every (non-binary, non-test) file under a directory tree.
function write_directory() {
  local dir_path=$1
  if [ ! -d "$dir_path" ]; then
    echo "Warning: directory '$dir_path' not found. Skipping." >&2
    MISSING+=("$dir_path/"); return
  fi
  dir_path="${dir_path%/}"
  while IFS= read -r -d $'\0' file; do
    [ "$(realpath "$file" 2>/dev/null)" = "$(realpath "$OUTPUT_FILE" 2>/dev/null)" ] && continue
    emit "$file" "false"
  done < <(find "$dir_path" -type f \
    -not -path '*/.git/*' -not -path '*/node_modules/*' \
    -not -name 'go.sum' -not -name '*_test.go' \
    -not -name 'idea' \
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
echo "  idea root:   $IDEA_ROOT"
echo "  code dir:    $CODE_DIR"
echo "  output file: $OUTPUT_FILE"

# 1) The Go service (code + embedded page + deploy/). Tests excluded by the walk.
write_directory "$CODE_DIR"

# 2) Start-here + go-live docs (resolved by name under DOCS_ROOT).
DOCS=(
  "HANDOFF.md"                                # state + the exact next steps
  "running_notes.md"                          # the cross-session journal
  "idea_uk_architecture_and_deployment.md"    # architecture, hosting, deploy, email live-state
  "EMAIL_identity_in_site_spec.md"            # email design
  "LIABILITY_AND_TERMS.md"                     # legal posture (solicitor review pending)
  "PLAN_stripe_billing_integration.md"        # the Stripe step (taking real money)
  "RUNBOOK_idea_uk.md"                         # the phased plan
  "DEVELOPMENT_RUNBOOK.md"
  # public pages + policy previews (content also embedded in service.go)
  "idea_uk_fakedoor.html"
  "leopardess_uk_index.html"
  "terms_preview.html"
  "refund_policy_preview.html"
  "privacy_preview.html"
)
for d in "${DOCS[@]}"; do add_named "$d" "$DOCS_ROOT" "false"; done

# 3) The large deploy/debug guide (skip with --no-debugdoc).
if [ "$WITH_DEBUGDOC" = true ]; then
  add_named "016_debugging_guide_v2_32.md" "$DOCS_ROOT" "false"
else
  echo "  (skipping 016_debugging_guide_v2_32.md per --no-debugdoc)"
fi

# --- Report ----------------------------------------------------------
echo ""
echo "Included ${#RESOLVED[@]} files."
if [ "${#MISSING[@]}" -gt 0 ]; then
  echo "MISSING (${#MISSING[@]}) — adjust IDEA_ROOT/CODE_DIR/DOCS_ROOT or paths:"
  printf '  - %s\n' "${MISSING[@]}"
fi
echo "Done. Context saved to $OUTPUT_FILE"
echo "File size: $(du -h "$OUTPUT_FILE" | cut -f1)"
