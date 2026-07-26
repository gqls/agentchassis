#!/usr/bin/env bash
# ============================================================================
# TRIGGER_rerender_page.sh — re-render ONE page from its stored content_data
# ============================================================================
# Why this exists rather than reusing
# docs024/cta_link_integrity/scripts/049b_deploy_single_page.sh:
#
#   That script's `section_data_resolved` branch FAILS as written. It sends
#   {page_id, site_id, domain} but `rerender_page_sections` declares
#   Required: []string{"target_site_id", "page_name"}
#   (rerender_page_sections_action.go:80), and nothing upstream derives
#   page_name from page_id. Result, observed 2026-07-26 on both oufe pages:
#
#     step rerender_sections failed: failed to execute action
#     rerender_page_sections: input extraction failed:
#     missing required fields: [page_name]
#
#   The assemble-only branch (no reason) does not touch that action, which is
#   why the gap survived — the failure only appears on the branch you need when
#   you have edited content_data and want it re-rendered. Reported to the
#   cta_link_integrity workstream; this file adds page_name and is otherwise
#   the same proven envelope.
#
# WHEN YOU NEED THE reason ARGUMENT
#   No reason  -> assemble-only: stitches STORED rendered_html. Cannot pick up
#                 a content_data edit or a component template change.
#   section_data_resolved -> re-renders every section from stored content_data
#                 through the CURRENT template, NO LLM call. This is the one you
#                 want after authoring copy directly into content_data.
#   cta_links_stale -> additionally recomputes CTA destinations.
#
# GOTCHA, checked before every run: if ANY section on the page has NULL
# content_data, the whole page escalates to the content writer and the copy IS
# REGENERATED — silently undoing hand-authored text.
#   SELECT slot_name, content_data IS NULL FROM page_components WHERE page_id='...';
#
# Usage: ./TRIGGER_rerender_page.sh <page_name> <domain> [reason]
# ============================================================================
set -euo pipefail

PAGE_NAME="${1:?Usage: $0 <page_name> <domain> [reason]}"
DOMAIN="${2:?missing domain}"
REASON="${3:-section_data_resolved}"

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc)

SITE_ID=$("${PSQL[@]}" "SELECT id FROM sites WHERE domain='${DOMAIN}';" | tr -d ' ')
[ -n "$SITE_ID" ] || { echo "no site row for ${DOMAIN}" >&2; exit 2; }

PAGE_ID=$("${PSQL[@]}" "SELECT id FROM pages WHERE site_id='${SITE_ID}' AND name='${PAGE_NAME}';" | tr -d ' ')
[ -n "$PAGE_ID" ] || { echo "no page '${PAGE_NAME}' on ${DOMAIN}" >&2; exit 2; }

NULLS=$("${PSQL[@]}" "SELECT count(*) FROM page_components WHERE page_id='${PAGE_ID}' AND content_data IS NULL;" | tr -d ' ')
if [ "${NULLS:-0}" -gt 0 ]; then
  echo "REFUSING: ${NULLS} section(s) on '${PAGE_NAME}' have NULL content_data." >&2
  echo "A ${REASON} re-render would escalate the page to the content writer and" >&2
  echo "regenerate the copy, silently discarding authored text." >&2
  exit 3
fi

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "corr=$CORR page=$PAGE_NAME ($PAGE_ID) domain=$DOMAIN reason=$REASON nulls=0"

kubectl -n kafka run -i --rm "kcat-rerender-$(date +%s)" \
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
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"page_id":"$PAGE_ID","page_name":"$PAGE_NAME","site_id":"$SITE_ID","target_site_id":"$SITE_ID","domain":"$DOMAIN","spec":{"reason":"$REASON","page_name":"$PAGE_NAME"}}}
JSON

echo "CORR=$CORR"
echo "  SELECT status, current_step, LEFT(error,160) FROM orchestration_states WHERE correlation_id='$CORR'::uuid;"
