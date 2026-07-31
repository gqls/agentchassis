#!/bin/bash
# landmines-verify-dispatch.sh — RFC_005 §3.2's automatic trigger.
#
# Runs landmines-sync.py --apply (the normal sync — writes doc_notes as
# always), then fires landmine-verifier ONCE per entry that sync reported as
# new or content-changed (the "NEEDS_VERIFICATION:<source>" lines it now
# prints — see landmines-sync.py's `changed` detection, added alongside this
# script). Unchanged entries are not re-verified on every run — that would be
# the weekly staleness sweep's job (RFC_005 §3.3, not this).
#
# Usage:
#   ./scripts/landmines-verify-dispatch.sh [ref]
#   ref defaults to the current branch, passed through to each trigger for
#   the verdict body only (see trigger-landmine-verifier.sh's own note).
#
# Run this instead of landmines-sync.py --apply directly when you want new/
# changed entries verified as part of the sync. Plain `landmines-sync.py
# --apply` still works standalone (this script does not replace it, only
# wraps it) for anyone who wants the sync without the dispatch.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REF="${1:-$(git -C "$REPO_ROOT" rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)}"

echo "========================================="
echo "landmines-sync.py --apply"
echo "========================================="
SYNC_OUT=$(python3 "$REPO_ROOT/scripts/landmines-sync.py" --apply)
echo "$SYNC_OUT"

SOURCES=$(echo "$SYNC_OUT" | grep '^NEEDS_VERIFICATION:' | cut -d: -f2- || true)

if [ -z "$SOURCES" ]; then
  echo ""
  echo "Nothing needs verification (no new or changed entries this run)."
  exit 0
fi

COUNT=$(echo "$SOURCES" | grep -c . || true)
echo ""
echo "========================================="
echo "Dispatching landmine-verifier for $COUNT entr$([ "$COUNT" = "1" ] && echo y || echo ies) (ref: $REF)"
echo "========================================="

FAILED=0
while IFS= read -r SOURCE; do
  [ -z "$SOURCE" ] && continue
  echo ""
  echo "--- $SOURCE ---"
  if ! "$REPO_ROOT/scripts/trigger-landmine-verifier.sh" "$SOURCE" "$REF"; then
    echo "  !! dispatch failed for $SOURCE" >&2
    FAILED=$((FAILED + 1))
  fi
  # A small gap between dispatches, not a rate limit this platform actually
  # needs — just avoids bursting many kubectl-run pods into the kafka
  # namespace at once when a sync catches several changed entries together.
  sleep 2
done <<< "$SOURCES"

echo ""
echo "========================================="
echo "Dispatched $COUNT, $FAILED failed to publish."
echo "Verdicts land in doc_notes (categories ? 'landmine-verification') as each"
echo "run completes -- these are async orchestrations, not awaited here. Check:"
echo "  SELECT subject_key, left(body,100), created_at FROM doc_notes"
echo "  WHERE categories ? 'landmine-verification' ORDER BY created_at DESC LIMIT $COUNT;"
echo "========================================="

[ "$FAILED" -eq 0 ]
