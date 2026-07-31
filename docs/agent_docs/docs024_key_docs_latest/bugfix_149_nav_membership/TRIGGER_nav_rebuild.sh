#!/usr/bin/env bash
# ============================================================================
# TRIGGER_nav_rebuild.sh — dispatch nav-updater at ONE site, directly
# ============================================================================
# WHY THIS EXISTS. The ordinary route is a work item (`item_type` nav_drift,
# `handler_agent` nav-updater, `status` triaged) picked up by the build dispatch
# loop — which is what RequestNavRebuild files, and what you should normally
# rely on. This script is the documented bypass for when that loop is not
# claiming: it publishes the orchestrate envelope itself.
#
# It was written on 2026-07-31 because the loop had stopped claiming FLEET-WIDE
# at 13:21 that day (9 items stalled, oldest 13:56, nothing claimed for two
# hours) while `scheduled_tasks.last_triggered_at` for build-pipeline-trigger
# kept advancing. **That stamp is fire-and-forget and proves only that the
# scheduler fired, never that a dispatch-loop orchestration was created.**
#
# WHAT IT DOES, in order (nav-updater's live workflow — read it, do not assume:
#   SELECT key, value->>'action' FROM agent_definitions a,
#          jsonb_each(a.default_config->'workflow'->'steps')
#   WHERE a.type='nav-updater' AND a.is_active
#     AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL; )
#
#   populate_nav_tables      — DELETEs the site's nav rows and rebuilds from `pages`
#   render_site_components   — re-renders header/footer INTO site_components
#   render_js_snippets_for_site + git_commit
#   get_pages_for_rerender + create_rerender_items — propagate to deployed pages
#
# THREE THINGS TO KNOW BEFORE YOU RUN IT
#
# 1. **It DELETEs and rebuilds `site_nav_items` for the whole site.** Count what
#    a rebuild cannot reproduce first — a child-path row whose page declares
#    NEITHER `in_header` nor `in_footer` is derived by nothing and will not come
#    back (RUNBOOK R4). Post-2026-07-31 a child-path row WITH a flag is
#    reproduced; before that chassis (v1.0.1215) none were.
#
# 2. **The final step only FILES rerender items.** If the dispatch loop is the
#    reason you are running this script, those items will sit unclaimed too — so
#    the nav tables and `site_components` will be correct while the SERVED pages
#    still carry the old chrome. Verify at the artefact and say which of the two
#    you have (RUNBOOK R7).
#
# 3. No orchestration dispatch within ~300s of a chassis pod (re)start — the
#    spawn is silently dropped. Check pod age first.
#
# Usage: ./TRIGGER_nav_rebuild.sh <domain>
# ============================================================================
set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain>}"

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)

SITE_ID=$("${PSQL[@]}" "SELECT id FROM sites WHERE domain='${DOMAIN}';" | tr -d ' ')
[ -n "$SITE_ID" ] || { echo "no site row for ${DOMAIN}" >&2; exit 2; }

# Pre-flight: what would a rebuild fail to reproduce? (see note 1)
ORPHANED=$("${PSQL[@]}" "
  SELECT count(*) FROM site_nav_items ni
  JOIN pages p ON p.site_id = ni.site_id AND (p.id = ni.page_id OR p.url = ni.url)
  WHERE ni.site_id='${SITE_ID}' AND ni.status='active'
    AND NOT (COALESCE(p.in_header,false) OR COALESCE(p.in_footer,false))
    AND lower(p.name) !~ '^(privacy|terms|cookie|disclaimer|legal)';" | tr -d ' ')

if [ "${ORPHANED:-0}" -gt 0 ]; then
  echo "WARNING: ${ORPHANED} active nav row(s) on ${DOMAIN} point at a page declaring NO nav flag." >&2
  echo "A rebuild derives from the flags, so those rows will NOT come back." >&2
  echo "Fix the flags first (that is the repair), or accept the loss knowingly." >&2
  echo "  RUNBOOK R4 has the per-row query." >&2
  [ "${FORCE:-0}" = "1" ] || { echo "Refusing. Re-run with FORCE=1 once you have read them." >&2; exit 3; }
fi

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "corr=$CORR domain=$DOMAIN site=$SITE_ID unreproducible_rows=${ORPHANED:-0}"

kubectl -n kafka run -i --rm "kcat-navrebuild-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_id=$MSG \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"nav-updater"},"input_data":{"site_id":"$SITE_ID","target_site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON

echo "CORR=$CORR"
echo "  SELECT status, current_step, LEFT(error,200) FROM orchestration_states WHERE correlation_id='$CORR'::uuid;"
echo "  -- then the nav rows it derived:"
echo "  SELECT ng.group_type, ni.label, ni.url FROM site_nav_items ni"
echo "    JOIN site_nav_groups ng ON ng.id=ni.group_id"
echo "   WHERE ni.site_id='$SITE_ID' AND ni.status='active' ORDER BY ng.group_type, ni.position;"
