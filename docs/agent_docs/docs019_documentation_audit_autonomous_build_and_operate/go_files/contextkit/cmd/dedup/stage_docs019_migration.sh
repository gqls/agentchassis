#!/usr/bin/env bash
# stage_docs019_migration.sh — stage the docs019 cleanup SAFELY.
#
# The migration has THREE kinds of move; this script only AUTO-DOES the two
# deterministic ones and STAGES the editorial third for your approval.
#
#   (1) ARCHIVE-DIR MOVES (deterministic): directories whose name means
#       "superseded" — go_files_old/, thin_slice_run/, working/, a stray
#       top-level archive_april_26/ — go into _archive/. Unambiguous by name.
#   (2) DUPLICATE COLLAPSE (deterministic): the ~200 (N) download-duplicates
#       and exact copies → handed to the dedup tool, which moves them to
#       _archive/ with a manifest. (Run separately; this script prints the
#       exact command and can run it with --dedup.)
#   (3) EDITORIAL MOVES (NOT automated): the loose FOCUS_/PLAN_/RUNBOOK_/016_/
#       NOTES_ working docs → engines/ or runbooks/. WHICH file is canonical
#       and WHAT each is about is your call, so this script only writes
#       PROPOSED_MOVES.tsv for you to edit and apply by hand.
#
# DEFAULT IS REPORT ONLY. Nothing moves without --apply. Even --apply touches
# only (1); (2) needs --dedup; (3) is never auto-applied.
#
# Usage:
#   ./stage_docs019_migration.sh <docs019-dir>            # report what (1) and (3) would do
#   ./stage_docs019_migration.sh <docs019-dir> --apply    # do (1) archive-dir moves
#   ./stage_docs019_migration.sh <docs019-dir> --dedup    # also run the dedup tool (needs DEDUP_BIN)
#
# Reversible: (1) moves dirs into _archive/<name> — mv back to undo. (2) writes
# dedup-manifest.tsv. (3) is a report, applied only by you.

set -euo pipefail

ROOT="${1:?usage: stage_docs019_migration.sh <docs019-dir> [--apply] [--dedup]}"
APPLY=0; RUN_DEDUP=0
for a in "${@:2}"; do
  case "$a" in
    --apply) APPLY=1 ;;
    --dedup) RUN_DEDUP=1 ;;
    *) echo "unknown flag: $a" >&2; exit 2 ;;
  esac
done
cd "$ROOT"

ARCHIVE="_archive"
mkdir -p "$ARCHIVE"

# ── (1) Archive-dir moves — deterministic, by directory name ──
# Only these names; each only if present at top level and not already inside _archive.
ARCHIVE_DIRS="go_files_old thin_slice_run working archive_april_26 old go_files_OLD"
echo "== (1) archive-dir moves =="
for d in $ARCHIVE_DIRS; do
  if [ -d "$d" ] && [ "$d" != "$ARCHIVE" ]; then
    if [ -e "$ARCHIVE/$d" ]; then
      echo "  SKIP $d/  (already in $ARCHIVE/ — would collide; merge by hand)"
      continue
    fi
    if [ "$APPLY" -eq 1 ]; then
      mv "$d" "$ARCHIVE/$d"
      echo "  MOVED $d/  ->  $ARCHIVE/$d/"
    else
      echo "  would move $d/  ->  $ARCHIVE/$d/"
    fi
  fi
done

# ── (2) Duplicate collapse — delegate to the dedup tool ──
echo
echo "== (2) duplicate collapse (dedup tool) =="
DEDUP_CMD="go run ./go_files/contextkit/cmd/dedup ."
echo "  exact + (N) duplicates are handled by dedup, not this script."
echo "  report:  $DEDUP_CMD -ext .go,.md,.json,.sql"
echo "  apply :  $DEDUP_CMD -ext .go,.md,.json,.sql -move"
echo "  (add -near for near-duplicate prose; review before -move -near.)"
if [ "$RUN_DEDUP" -eq 1 ]; then
  if [ -d go_files/contextkit/cmd/dedup ]; then
    echo "  running dedup (report only — re-run with -move yourself once happy)..."
    ( cd . && eval "$DEDUP_CMD -ext .go,.md,.json,.sql" ) || echo "  (dedup run failed — run it by hand)"
  else
    echo "  dedup not found at go_files/contextkit/cmd/dedup — skipping --dedup."
  fi
fi

# ── (3) Editorial moves — STAGE ONLY, never auto-apply ──
# Heuristic destination by filename prefix; YOU confirm/edit before applying.
echo
echo "== (3) editorial moves — writing PROPOSED_MOVES.tsv (NOT applied) =="
MAP="PROPOSED_MOVES.tsv"
{
  printf "current\tproposed_dest\treason\tACTION(edit:keep|move|archive|skip)\n"
  for f in *; do
    [ -f "$f" ] || continue
    case "$f" in
      PROPOSED_MOVES.tsv|dedup-manifest.tsv|stage_docs019_migration.sh) continue ;;
    esac
    dest=""; reason=""
    case "$f" in
      RUNBOOK_*)  dest="runbooks/$f";                 reason="runbook" ;;
      016_*)      dest="runbooks/$f";                 reason="debug-guide (016)" ;;
      *adapter*|*code_symbols*|*analyser*)
                  dest="engines/analyser/$f";         reason="analyser engine" ;;
      *contextkit*|*assembler*|*resolve*|*embed*|*fuse*|*eval_targets*|*pipeline*|*B4a*)
                  dest="engines/contextkit-cli/$f";   reason="contextkit CLI" ;;
      *tool*|*component*)
                  dest="engines/tool-docs/$f";        reason="tools thread" ;;
      NOTES_*|PLAN_*|FOCUS_*|MASTER_*|MAPPING_*|ARCHITECTURAL_*|GUIDE_*)
                  dest="(decide)";                    reason="working doc — keep at root, file under an engine, or archive?" ;;
      *)          dest="(decide)";                    reason="unclassified" ;;
    esac
    # flag download-duplicates so they're obviously dedup's job, not an editorial move
    case "$f" in
      *\(*\)*) reason="$reason; (N)-DUPLICATE → leave for dedup" ;;
    esac
    printf "%s\t%s\t%s\t\n" "$f" "$dest" "$reason"
  done
} > "$MAP"
echo "  wrote $MAP ($(($(wc -l < "$MAP") - 1)) rows)."
echo "  EDIT the ACTION column (move|archive|skip|keep), then apply with the one-liner below."

cat <<'EOF'

== applying (3) AFTER you edit PROPOSED_MOVES.tsv ==
  # dry-run what your edits would do:
  awk -F'\t' 'NR>1 && $4!="" && $4!="keep" && $4!="skip"{print $4": "$1" -> "$2}' PROPOSED_MOVES.tsv
  # apply (move rows marked "move"; archive rows marked "archive"):
  awk -F'\t' 'NR>1 && $4=="move"{print $1"\t"$2}' PROPOSED_MOVES.tsv | while IFS=$'\t' read -r src dst; do
    mkdir -p "$(dirname "$dst")" && git mv "$src" "$dst" 2>/dev/null || mv "$src" "$dst"
  done
  awk -F'\t' 'NR>1 && $4=="archive"{print $1}' PROPOSED_MOVES.tsv | while read -r src; do
    mkdir -p "_archive/$(dirname "$src")" && mv "$src" "_archive/$src"
  done

REMEMBER: (3) is editorial. The script proposes; you decide. After moving,
re-point internal links and rebuild the index with `-exclude _archive/`.
EOF

echo
if [ "$APPLY" -eq 0 ]; then
  echo "REPORT ONLY (step 1 not applied). Re-run with --apply for the archive-dir moves."
fi
