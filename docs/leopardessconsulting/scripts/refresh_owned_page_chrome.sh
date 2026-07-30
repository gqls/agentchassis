#!/bin/bash
# ============================================================================
# refresh_owned_page_chrome.sh — give a rebuild_policy='owned' page the current
# site chrome (header/footer), then put its protection back.
#
# WHY THIS DANCE IS NECESSARY. An owned page cannot be re-rendered in place:
# page-rerender's save_sections step refuses it outright —
#   "page X is rebuild_policy=owned (tool/widget-owned): a generic section save
#    would clobber it ... Refusing to overwrite."
# That refusal is correct and is what protects tool pages from the generic
# builder. But it also means a nav or chrome change never reaches them, so an
# owned tool page silently keeps a stale footer for ever.
#
# So: flip to 'generic', re-render in ASSEMBLE mode (no spec.reason, so stored
# section HTML is re-assembled with fresh chrome and no section is regenerated),
# then flip back to 'owned'.
#
# THE SAFETY PROPERTY, and the reason this is a script and not four commands:
# the restore runs from an EXIT trap, so the page gets its protection back even
# if the render fails, the wait times out, or this script is interrupted. A page
# left on 'generic' is a page any other session's wide rebuild may overwrite —
# and on a shared tree with many concurrent sessions that window is real.
#
# Assemble mode also needs page_id: rerender_single_page reads page_id/site_id/
# domain off the input and errors "page_id not found in input". page_name alone
# is only resolved by the section branch.
#
# Usage: ./refresh_owned_page_chrome.sh <site_id> <domain> <marker> <page_name> [...]
# ============================================================================
set -uo pipefail

S="${1:?site_id}"; DOMAIN="${2:?domain}"; MARKER="${3:?marker that must appear once chrome is current}"
shift 3
[ $# -gt 0 ] || { echo "give at least one page_name"; exit 2; }

PSQL=(kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)
HERE="$(cd "$(dirname "$0")" && pwd)"

RESTORE_IDS=()
restore_all() {
  local rc=$?
  for pid in "${RESTORE_IDS[@]:-}"; do
    [ -n "$pid" ] || continue
    "${PSQL[@]}" "UPDATE pages SET rebuild_policy='owned' WHERE id='$pid'" >/dev/null 2>&1
    local now
    now=$("${PSQL[@]}" "SELECT rebuild_policy FROM pages WHERE id='$pid'")
    printf '  restore %s -> %s\n' "$pid" "$now"
    [ "$now" = "owned" ] || echo "  !! FAILED TO RESTORE $pid — set rebuild_policy='owned' BY HAND NOW"
  done
  exit $rc
}
trap restore_all EXIT INT TERM

for PAGE in "$@"; do
  PID=$("${PSQL[@]}" "SELECT id FROM pages WHERE site_id='$S' AND name='$PAGE'")
  POLICY=$("${PSQL[@]}" "SELECT COALESCE(rebuild_policy,'generic') FROM pages WHERE id='$PID'")
  URL=$("${PSQL[@]}" "SELECT url FROM pages WHERE id='$PID'")
  echo ""
  echo "=== $PAGE ($URL) policy=$POLICY"
  [ -n "$PID" ] || { echo "  no such page — skipping"; continue; }
  if [ "$POLICY" != "owned" ]; then
    echo "  not owned; use the ordinary reconcile for this page — skipping"
    continue
  fi

  "${PSQL[@]}" "UPDATE pages SET rebuild_policy='generic' WHERE id='$PID'" >/dev/null
  RESTORE_IDS+=("$PID")           # armed BEFORE the render, so the trap covers it
  echo "  flipped to generic"

  OUT=$("$HERE/orchestrate_safe.sh" page-rerender \
        "{\"site_id\":\"$S\",\"domain\":\"$DOMAIN\",\"page_name\":\"$PAGE\",\"page_id\":\"$PID\"}" 2>&1)
  CID=$(printf '%s' "$OUT" | sed -n 's/^correlation: *//p')
  case "$OUT" in *PUBLISH_OK*) echo "  published ($CID)";;
                 *) echo "  !! PUBLISH DROPPED — nothing was sent; re-run for this page"; continue;; esac

  for _ in $(seq 1 24); do
    ST=$("${PSQL[@]}" "SELECT status FROM orchestration_states WHERE correlation_id='$CID'")
    case "$ST" in COMPLETED|FAILED) break;; esac
    sleep 5
  done
  ERR=$("${PSQL[@]}" "SELECT COALESCE(error,'') FROM orchestration_states WHERE correlation_id='$CID'")
  echo "  render: ${ST:-no-row} ${ERR:0:120}"
done

echo ""
echo "restoring ownership before verifying (protection first, cosmetics second)"
# The trap does the restore; verify AFTER it by re-reading the served pages.
trap - EXIT INT TERM
restore_verify() {
  for pid in "${RESTORE_IDS[@]:-}"; do
    "${PSQL[@]}" "UPDATE pages SET rebuild_policy='owned' WHERE id='$pid'" >/dev/null
    local now url; now=$("${PSQL[@]}" "SELECT rebuild_policy FROM pages WHERE id='$pid'")
    url=$("${PSQL[@]}" "SELECT url FROM pages WHERE id='$pid'")
    printf '  %-46s policy=%s\n' "$url" "$now"
    [ "$now" = "owned" ] || echo "  !! FAILED TO RESTORE — set rebuild_policy='owned' BY HAND NOW"
  done
}
restore_verify

echo ""
echo "served check (a first read can be stale; this retries once)"
for pid in "${RESTORE_IDS[@]:-}"; do
  url=$("${PSQL[@]}" "SELECT url FROM pages WHERE id='$pid'")
  hit=$(curl -s "https://${DOMAIN}${url}?cb=$(date +%s%N)" | grep -cF "$MARKER")
  [ "$hit" -eq 0 ] && { sleep 20; hit=$(curl -s "https://${DOMAIN}${url}?cb=$(date +%s%N)" | grep -cF "$MARKER"); }
  printf '  %-46s marker=%s\n' "$url" "$hit"
done
