#!/usr/bin/env bash
# rebuild_pages.sh — the operator entry point features_open/021 asks for:
# "an operator with a legitimate reason to rebuild a set of pages has no
# supported operation for it."
#
# WHAT THIS ACTUALLY DOES, AND WHY IT NEEDED NO NEW GO CODE (2026-08-05 finding)
# -------------------------------------------------------------------------
# features_open/021's own draft assumed the paved road (`maintenance_queue` →
# `maintenance-triage` → `page-rebuild`) would need (a) a new item_type that
# "says it is a rebuild" so it isn't confused with a resurrected historic row,
# and (b) protection from `stale-work-item-reaper` (bugs_open/070) reaping a
# freshly-queued request. Read the actual code before building either:
#
#   - `flagPagesForRebuild` (maintenance_actions.go:962) is a pure
#     `UPDATE pages SET build_status='needs_rebuild'`. This whole path NEVER
#     touches `site_work_items` at all — the "row lies about what happened"
#     problem in 021's filing was specific to the operator's AD HOC workaround
#     (`UPDATE site_work_items ... -- an OLD row, reused`), not to this
#     mechanism. Routing through `maintenance_queue` sidesteps it entirely.
#   - `stale-work-item-reaper` only ages `site_work_items` rows
#     (`WHERE status='triaged' AND pipeline='build'`). `maintenance_queue` is a
#     different table with its own status column; the reaper cannot see it,
#     let alone reap it. bugs_open/070 was a real prerequisite for the
#     operator's ORIGINAL by-hand workaround, but is orthogonal to this path.
#
# So this script is the whole of the missing piece: insert a `maintenance_queue`
# row naming the real reason and requester, then dispatch `maintenance-triage`
# directly (it has no scheduled_tasks row driving it — confirmed still true
# 2026-08-05 — so nothing else will ever pick this row up).
#
# WHAT IS NOT WIRED — READ BEFORE ASSUMING (2026-08-05)
# -------------------------------------------------------------------------
# 1. INTENT (recompose vs re-render, features_open/012's vocabulary) is NOT
#    read by this pipeline. `page-rebuild`'s `build_pages_loop` always calls
#    `plan_sections` + a content-writer step per page — there is no
#    re-render-only fast path today. `recompose_pages` (v3_site_actions.go)
#    is a DIFFERENT mechanism entirely, living in the site-PLAN validation
#    pipeline, not this one. This script still writes `payload.intent` (grep
#    for it below) so a future Go change has somewhere to read it from, but
#    right now every rebuild dispatched here is a full recompose. If you only
#    wanted a cheap re-render, this script does more work than you asked for.
# 2. SWEEP-IN: `page-rebuild`'s `get_pages_to_build` step claims EVERY page on
#    the target site currently at `build_status='needs_rebuild'`, not just the
#    ones this run flags. If other pages already sit `needs_rebuild` for
#    unrelated reasons, a live (non-dry-run) dispatch rebuilds those too. This
#    script prints that count before you commit to a real dispatch — read it.
# 3. A live (DRY_RUN=0) dispatch is LONG-RUNNING: `page-rebuild`'s own
#    per-site step timeout is 5400s (90 min), and `rebuild_loop` runs sites
#    SEQUENTIALLY. On the shared `system.agent.generic.requests` lane (this
#    script does not have, and has not been given, a dedicated topic — unlike
#    council-gate, which got one in bugs_open/096 specifically because of
#    head-of-line blocking), a real dispatch can sit behind, or hold up,
#    whatever else is queued. Check `scripts/dispatch-queue-depth.sh` first;
#    consider whether this needs its own topic if it ever gets used often.
#
# Usage:
#   ./rebuild_pages.sh <domain> <page1,page2,...> "<reason>"
#
#   # dry run (DEFAULT — reports what would happen, dispatches nothing real):
#   ./rebuild_pages.sh gaswholesalers.com index,capabilities "new hero-card-carousel in plan"
#
#   # the real thing, once you have read the dry run's report:
#   DRY_RUN=0 ./rebuild_pages.sh gaswholesalers.com index,capabilities "new hero-card-carousel in plan"
#
# Env:
#   DRY_RUN       1 (default, SAFE) = maintenance-triage scans and reports,
#                 dispatches NOTHING further (`check_dry_run`'s own branch).
#                 0 = a real dispatch: pages get flagged and page-rebuild runs.
#   PRIORITY      maintenance_queue.priority (default 3 — lower than the
#                 automated scan's own default of 5, so an explicit operator
#                 ask is claimed first if both are pending; ASC order).
#   REQUESTED_BY  free text identifying who/what asked (default: whoami + this
#                 script's name — name yourself properly, this is the field
#                 021 filed this whole feature over the ABSENCE of).
#   INTENT        recompose|rerender (default: recompose — see note 1 above;
#                 written to payload but NOT YET consumed by any code).
set -euo pipefail

DOMAIN="${1:?usage: $0 <domain> <page1,page2,...> \"<reason>\"}"
PAGES_CSV="${2:?usage: $0 <domain> <page1,page2,...> \"<reason>\"}"
REASON="${3:?usage: $0 <domain> <page1,page2,...> \"<reason>\"}"

DRY_RUN="${DRY_RUN:-1}"
PRIORITY="${PRIORITY:-3}"
REQUESTED_BY="${REQUESTED_BY:-$(whoami)@rebuild_pages.sh}"
INTENT="${INTENT:-recompose}"

TARGET_AGENT_TYPE='maintenance-triage'
CLIENT_ID='demo_client'

PSQL=(kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1)

# comma-separated -> JSON array of strings. No spaces, no quotes in entries.
csv_to_json_array() {
  printf '%s' "$1" | awk -F, '{for(i=1;i<=NF;i++){printf "%s\"%s\"", (i>1?",":""), $i}}'
}
PAGES_JSON="[$(csv_to_json_array "$PAGES_CSV")]"

echo "-- resolving site_id for ${DOMAIN} ..."
SITE_ID=$(printf '%s' "SELECT id FROM sites WHERE domain = '${DOMAIN}';" | "${PSQL[@]}" -t -A)
if [ -z "$SITE_ID" ]; then
  echo "REFUSING: no site found for domain '${DOMAIN}'." >&2
  exit 2
fi
echo "   site_id: ${SITE_ID}"

# ── pre-flight: what else already sits needs_rebuild for this site ──────────
# page-rebuild's get_pages_to_build claims EVERY needs_rebuild page on the
# site, not just the ones this run names. Report it before a real dispatch.
EXISTING_NEEDS_REBUILD=$(printf '%s' "
  SELECT count(*) FROM pages WHERE site_id = '${SITE_ID}' AND build_status = 'needs_rebuild';
" | "${PSQL[@]}" -t -A)
if [ "${EXISTING_NEEDS_REBUILD:-0}" != "0" ]; then
  echo ""
  echo "   NOTE: ${EXISTING_NEEDS_REBUILD} page(s) on this site ALREADY sit"
  echo "   build_status='needs_rebuild' for reasons unrelated to this run. A REAL"
  echo "   (DRY_RUN=0) dispatch will rebuild those too — page-rebuild claims every"
  echo "   needs_rebuild page on the site, not only the ones named below:"
  printf '%s' "
    SELECT name, updated_at FROM pages
    WHERE site_id = '${SITE_ID}' AND build_status = 'needs_rebuild'
    ORDER BY updated_at;
  " | "${PSQL[@]}" | sed 's/^/     /'
fi

# ── pre-flight: existing pending/claimed page_rebuild tasks for this site ───
EXISTING_TASKS=$(printf '%s' "
  SELECT count(*) FROM maintenance_queue
  WHERE site_id = '${SITE_ID}' AND task_type = 'page_rebuild' AND status IN ('pending','claimed');
" | "${PSQL[@]}" -t -A)
[ "${EXISTING_TASKS:-0}" != "0" ] && \
  echo "   NOTE: ${EXISTING_TASKS} page_rebuild task(s) already pending/claimed for this site — this run's pages will be UNIONED with theirs by PrepareRebuildDispatchesAction, not run separately."

echo ""
echo "========================================="
echo "operator page-rebuild request"
echo "========================================="
echo "  Domain:        ${DOMAIN}"
echo "  Pages:          ${PAGES_CSV}"
echo "  Reason:        ${REASON}"
echo "  Requested by:  ${REQUESTED_BY}"
echo "  Intent:        ${INTENT}  (NOT YET consumed by any code — see script header)"
echo "  DRY_RUN:       ${DRY_RUN}"
echo "========================================="

# ── DRY_RUN=1: report only, INSERT NOTHING, DISPATCH NOTHING ────────────────
# Found by testing this script, not by reading the workflow (2026-08-05): the
# existing `check_dry_run` step's dry-run branch (`complete_dry_run`) skips
# straight past `prepare_rebuild_dispatches` — the ONLY step that ever reads a
# maintenance_queue row. So maintenance-triage's own dry_run mode previews the
# AUTOMATED scanner's findings (stale_pages/missing_content/orphan_nav), never
# an operator-supplied page list. Inserting a row under DRY_RUN=1 would just
# leave inert debris nobody previews (a live test run of an earlier version of
# this script did exactly that — cleaned up by hand, not by any mechanism).
# So the honest dry run is local: print what WOULD be inserted/dispatched, and
# do neither.
if [ "$DRY_RUN" = "1" ]; then
  cat <<EOF

=========================================
DRY RUN — nothing inserted, nothing dispatched.
=========================================
Would INSERT into maintenance_queue:
  site_id:      ${SITE_ID}
  task_type:    page_rebuild
  priority:     ${PRIORITY}
  reason:       ${REASON}
  requested_by: ${REQUESTED_BY}
  payload:      {"pages": ${PAGES_JSON}, "intent": "${INTENT}"}

Would then dispatch maintenance-triage with input_data.domain='${DOMAIN}',
input_data.dry_run=false — which claims ALL pending page_rebuild tasks for
this site (not scoped to this one), flags their pages needs_rebuild, and runs
page-rebuild per site (sequential, up to 90 min/site — see script header).

This does NOT preview maintenance-triage's own dry_run report (that shows the
AUTOMATED scanner's findings, unrelated to the pages you named — see script
header note). If you want to see that too:
  DRY_RUN=0 with input_data.dry_run left true is not exposed by this script;
  fire maintenance-triage directly with {"domain":"${DOMAIN}","dry_run":true}
  if you specifically want the automated-scan preview.

Re-run with DRY_RUN=0 once you have read the pre-flight notes above.
EOF
  exit 0
fi

# ── 1. the durable queue row — this IS the fresh, self-describing record ────
# task_type/status/retry_count/max_retries/payload all take their table
# DEFAULTs except what we set explicitly (schema checked 2026-08-05).
# Wrapped in a CTE, not a bare INSERT ... RETURNING: psql -t suppresses a
# SELECT's header/footer but NOT a non-SELECT's command tag ("INSERT 0 1"),
# which leaked into the captured value on this script's own first test run
# (090's script documents the same landmine — read it before trusting `-t`
# on anything but a SELECT).
echo ""
echo "-- inserting maintenance_queue row ..."
TASK_ID=$(printf '%s' "
  WITH ins AS (
    INSERT INTO maintenance_queue (site_id, task_type, priority, reason, payload, requested_by)
    VALUES (
      '${SITE_ID}', 'page_rebuild', ${PRIORITY},
      \$reason\$${REASON}\$reason\$,
      jsonb_build_object('pages', '${PAGES_JSON}'::jsonb, 'intent', '${INTENT}',
                          'detected_at', to_char(now() AT TIME ZONE 'UTC', 'YYYY-MM-DD\"T\"HH24:MI:SS\"Z\"')),
      '${REQUESTED_BY}'
    )
    RETURNING id
  )
  SELECT id::text FROM ins;
" | "${PSQL[@]}" -t -A)
if ! printf '%s' "$TASK_ID" | grep -Eq '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'; then
  echo "INSERT DID NOT RETURN A ROW ID (got: '${TASK_ID}') — refusing to dispatch." >&2
  exit 2
fi
echo "   task_id: ${TASK_ID}"

# ── 2. dispatch maintenance-triage directly ──────────────────────────────────
# No scheduled_tasks row targets this agent (confirmed 2026-08-05) — nothing
# else will ever claim this row, so a direct dispatch is not a double-fire
# risk the way it would be for a live-scheduled loop (contrast 090's dance
# with diagnose-dispatch-loop).
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
INPUT_DATA="{\"domain\":\"${DOMAIN}\",\"dry_run\":false,\"correlation_id\":\"${CORRELATION_ID}\"}"

echo ""
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo "-- dispatching to ${TARGET_AGENT_TYPE} (REAL — pages will be rebuilt) ..."
kubectl -n kafka run -i --rm "kcat-rebuild-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=operator-rebuild-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":$INPUT_DATA}
JSON

cat <<EOF

=========================================
dispatched. Verify — this trigger printing cleanly is NOT proof it ran
(kcat's own known silent-drop shape; check for an orchestration_states row
within a few minutes, not just once).
=========================================

1) Orchestration state (by correlation id):
   SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - updated_at))::int AS since_s
   FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;

2) The queue row (pending -> claimed -> complete):
   SELECT status, claimed_at, completed_at, result, error_message
   FROM maintenance_queue WHERE id = '$TASK_ID';

3) Did the pages actually get flagged / rebuilt:
   SELECT name, build_status, updated_at FROM pages
   WHERE site_id = '$SITE_ID' AND name = ANY(string_to_array('$PAGES_CSV', ','));

4) Watch it run (page-rebuild's own per-site timeout is 5400s):
   kubectl -n ai-persona-system logs -f -l agent-type=page-rebuild --tail=200 | grep '$CORRELATION_ID'
EOF
