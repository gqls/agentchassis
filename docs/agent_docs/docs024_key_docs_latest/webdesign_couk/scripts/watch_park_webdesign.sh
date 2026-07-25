#!/bin/bash
# ============================================================================
# watch_park_webdesign.sh — the airlock for webdesign.co.uk's build cascade
# ============================================================================
# WHY THIS EXISTS. The onboarding cascade emits each next work item at
# status='triaged' (create_work_item_action.go:141 defaults to "triaged"), which
# is immediately dispatchable. We want to inspect what each leg wrote before the
# next leg runs — above all, to land the design_intent pin BEFORE webdesign-agent
# ever runs, because an unpinned first run invents its own palette.
#
# THE SITE LOCK IS NOT A PARK. Verified 2026-07-25 by reading both SQL sites:
#   - scheduled_tasks('build-pipeline-trigger').pre_query counts only sites with
#     locked_at IS NULL — so the lock decides whether the trigger FIRES.
#   - the trigger's find_dispatchable_site step then runs:
#       SELECT DISTINCT ON (wi.site_id) ... FROM site_work_items wi
#       JOIN sites s ... WHERE wi.status IN ('triaged','approved') ...
#       ORDER BY wi.site_id, wi.priority ASC LIMIT 1
#     with NO locked_at filter. So when the trigger fires for ANY other site's
#     backlog, it can select ours — locked or not.
# The lock is belt. Item status is the braces, and that is what this watches.
#
# WHAT IT DOES. Every INTERVAL seconds, any work item on this site that is
# dispatchable (detected/triaged/approved) and NOT allowlisted is flipped to
# status='blocked'. 'blocked' is a real, respected status: it has its own index
# (idx_work_items_blocked) and CompleteWorkItemAction explicitly refuses to
# overwrite it (load_work_item_actions.go:808).
#
# NOTE — a blocked item still HOLDS its dedup slot. idx_swi_dedup excludes
# complete/verified/rejected/wont_fix/failed/unresolved/cancelled — 'blocked' is
# not in that list. That is what we want (no duplicate item can be created
# behind a parked one), but it means a forgotten park blocks re-emission.
#
# Items already 'claimed' are left alone — they are in flight and flipping them
# would only confuse the completion path.
#
# RELEASING A LEG (the airlock):
#   1. echo "<item-uuid>" >> <allowlist file>
#   2. WAIT one full interval. The loop reads the allowlist ONCE per cycle, so a
#      flip that lands mid-cycle races the UPDATE built from the previous read
#      and the item is re-parked instantly. That happened to needs_briefing on
#      2026-07-25 and cost a confused minute: the allowlist was correct, the
#      item was blocked, and neither fact explained the other.
#   3. psql: UPDATE site_work_items SET status='triaged' WHERE id='<item-uuid>';
#   4. watch orchestration_states (NOT site_work_items.updated_at — bugs_open/035)
#   5. when it completes, the next leg's item appears and is parked within
#      INTERVAL seconds. Read what the leg wrote, then release the next one.
#
# Usage:
#   ./watch_park_webdesign.sh <domain> [allowlist-file] [interval-seconds]
# Defaults: allowlist=./park_allowlist.txt  interval=5
#
# Keep parks SHORT. stale-work-item-reaper runs hourly and keys on row age
# (bugs_open/070) — a multi-day park invites it. Check its status filter before
# parking anything overnight.
# ============================================================================
set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain> [allowlist-file] [interval-seconds]}"
ALLOWLIST="${2:-$(dirname "$0")/park_allowlist.txt}"
INTERVAL="${3:-5}"

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A)

touch "$ALLOWLIST"

echo "[park] domain=$DOMAIN allowlist=$ALLOWLIST interval=${INTERVAL}s"
echo "[park] waiting for the sites row to appear..."

SITE_ID=""
while [ -z "$SITE_ID" ]; do
  SITE_ID="$("${PSQL[@]}" -c "SELECT id FROM sites WHERE domain = '$DOMAIN';" 2>/dev/null | tr -d '[:space:]' || true)"
  [ -z "$SITE_ID" ] && sleep "$INTERVAL"
done
echo "[park] site_id=$SITE_ID — parking active"

while true; do
  # Allowlist as a SQL literal list. Empty file => a sentinel that matches nothing.
  # The `|| true` is load-bearing: under `set -o pipefail` a grep that matches
  # nothing (i.e. an empty allowlist — the normal state) fails the whole pipeline
  # and `set -e` kills the watcher. That killed the first run on 2026-07-25.
  IDS="$(grep -oE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}' "$ALLOWLIST" 2>/dev/null | sort -u | sed "s/^/'/;s/$/'/" | paste -sd, - || true)"
  [ -z "$IDS" ] && IDS="'00000000-0000-0000-0000-000000000000'"

  FLIPPED="$("${PSQL[@]}" -c "
    UPDATE site_work_items
       SET status = 'blocked'
     WHERE site_id = '$SITE_ID'
       AND status IN ('detected','triaged','approved')
       AND id NOT IN ($IDS)
    RETURNING item_type || ' [' || COALESCE(handler_agent,'(no handler)') || '] ' || id;
  " 2>/dev/null || true)"

  # psql prints its command tag ("UPDATE 0") alongside the RETURNING rows even
  # under -t -A, so filter it out. Logging it made every idle cycle look like a
  # park and would have hidden a real one in the noise.
  while IFS= read -r line; do
    case "$line" in
      ""|UPDATE\ [0-9]*) continue ;;
    esac
    echo "[park $(date -u +%H:%M:%SZ)] PARKED $line"
  done <<< "$FLIPPED"

  sleep "$INTERVAL"
done
