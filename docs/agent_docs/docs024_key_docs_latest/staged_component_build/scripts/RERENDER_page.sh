#!/usr/bin/env bash
# RERENDER_page.sh <site_id> <domain> <page_id> — parameterised single-page rerender.
#
# The committed, argument-taking version of
# scripts/initial_messages/001_assemble_all_pages_rerender/081b_… , which hard-codes
# one gaswholesalers page. Same message, same topic, same agent.
#
# WHY THIS PATH MATTERS for a component change: a rerender is what republishes
# BOTH the page html AND `/tools/assets/<function>.js` (collectJSAssets,
# rerender_single_page_action.go:338) — so a component whose js_content changed
# is only live on pages that have since re-rendered. It also re-runs
# sanitiseFormAction, which is what fills `form_action` from `sites.email`.
#
# VERIFY AT THE ARTEFACT, never at the orchestration status: curl the page and
# the asset. `complete` is not proof the work happened.
set -euo pipefail

if [ $# -lt 3 ]; then
  echo "usage: RERENDER_page.sh <site_id> <domain> <page_id> [reason]" >&2
  echo >&2
  echo "  reason  omit for the ASSEMBLE-ONLY path (republishes the page from each" >&2
  echo "          section's STORED rendered_html, plus every js asset)." >&2
  echo "          Pass 'section_data_resolved' to RE-RENDER each section from its" >&2
  echo "          component html_template first." >&2
  exit 2
fi

SITE_ID="$1"; DOMAIN="$2"; PAGE_ID="$3"; REASON="${4:-}"

# WHY THE REASON MATTERS, and it cost a wasted cycle to learn (2026-08-09).
# page-rerender has TWO paths, chosen by a conditional on input_data.spec.reason:
#
#   reason in (image_landed | section_data_resolved | cta_links_stale)
#        -> rerender_sections: each section is re-rendered from its component's
#           html_template, through the same RenderTemplate path as a build. This
#           is the ONLY path that picks up a TEMPLATE change, and the only one
#           that re-runs sanitiseFormAction.
#   anything else (including no reason at all)
#        -> render_page: assembles the page from each section's STORED
#           rendered_html.
#
# The trap is that the assemble-only path STILL republishes /tools/assets/*.js
# from js_content (collectJSAssets). So after a component change that touched
# BOTH template and js, a reason-less rerender ships the new SCRIPT against the
# OLD MARKUP, reports COMPLETED, and the served asset verifies green. Measured
# here: the contact-block asset went 2,100 -> 7,345 bytes while the form tag on
# the page was untouched and page_components.updated_at stayed at 2026-07-31.
# **Check the MARKUP, not just the asset.**
# AND page_name IS REQUIRED on the section path, which nothing tells you.
# `rerender_sections` ran happily (rerendered: 3, escalated: false) and then
# `save_sections` returned {"skipped": true, "success": true,
# "sections_saved": 0, "reason": "no page name"} — so three freshly rendered
# sections were computed and thrown away while the orchestration reported
# COMPLETED and the page redeployed from the OLD stored html. The real trigger
# is a work item, which carries it. And the path is EXACT, read out of the live
# workflow rather than guessed — save_sections is configured with
# `"page_name_field": "input_data.spec.page_name"`, so page_name must sit INSIDE
# spec. Putting it at the top of input_data (the obvious place, and the first
# thing tried here) produces the identical silent skip. Same family as
# bugs_open/095, which that action's own file cites. Look it up here rather than
# leave it to be forgotten: the failure is invisible in every status field.
#
#   SELECT jsonb_pretty(default_config->'workflow'->'steps'->'save_sections')
#     FROM agent_definitions WHERE type='page-rerender' AND is_active
#      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
PAGE_NAME=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -tAc \
  "SELECT name FROM pages WHERE id='${PAGE_ID}';" | tr -d ' \r\n')
if [ -z "$PAGE_NAME" ]; then
  echo "ABORT: no page row for ${PAGE_ID}" >&2
  exit 2
fi

if [ -n "$REASON" ]; then
  SPEC=",\"spec\":{\"reason\":\"${REASON}\",\"page_name\":\"${PAGE_NAME}\"}"
else
  SPEC=""
fi

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "=== page-rerender: ${DOMAIN} page ${PAGE_ID} (reason=${REASON:-<none: assemble-only>}) ==="
echo "  page_name=${PAGE_NAME}"
echo "  CORRELATION_ID=${CORRELATION_ID}"
echo "  ORCHESTRATION_ID=${ORCHESTRATION_ID}"

kubectl -n kafka run -i --rm "kcat-rerender-$(date +%s)-$RANDOM" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","page_id":"${PAGE_ID}","page_name":"${PAGE_NAME}"${SPEC}}}
JSON

echo
echo "  SELECT status, current_step FROM orchestration_states WHERE orchestration_id='${ORCHESTRATION_ID}';"
echo "  ALSO read collected_data->'sections_saved': {\"skipped\":true,\"success\":true}"
echo "  means the re-rendered sections were DISCARDED and the page shipped stale."
echo "  then CURL THE PAGE AND THE ASSET — a 'complete' row is not a rendered artefact."
