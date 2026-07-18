#!/usr/bin/env bash
# TRIGGER_discovery_agent.sh — run a discovery (auditor) agent against one site.
#
#   ./TRIGGER_discovery_agent.sh <agent-type> <site_id>
#
# e.g. ./TRIGGER_discovery_agent.sh completeness-discovery-agent 1244516d-014d-421c-88c6-090bb1e9552a
#
# The three auditors and what they cover (from agent_definitions.default_config):
#   completeness-discovery-agent — dead_controls, phantom_internal_links,
#       misdirected_cta, empty_sections, orphan_pages, sectionless_pages,
#       unresolved_sections, required_fields_missing, section_source_drift, …
#   design-discovery-agent       — content_image_missing, image_url_404,
#       placeholder_image_in_use, missing_css, missing_tools, tool_health, …
#   quality-discovery-agent      — broken_nav_links, placeholder_contact,
#       generic_theme, unverified_claims, voice_tells
#
# Findings land in site_work_items with source='discovery' (most at
# status='needs_human_review'). Read them, don't assume the run did nothing:
#   SELECT item_type, severity, summary FROM site_work_items
#   WHERE site_id=<id> AND source='discovery' ORDER BY created_at DESC;
#
# ⚠️ The payload MUST stay on ONE line — kcat -P splits on newlines and would
#    publish each line as a separate (invalid) message.
set -euo pipefail

AGENT_TYPE="${1:?usage: $0 <agent-type> <site_id>}"
SITE_ID="${2:?usage: $0 <agent-type> <site_id>}"
CLIENT_ID="${CLIENT_ID:-demo_client}"

# These workflows open with ensure_site_record, which resolves BY DOMAIN and
# fails "domain not found in input_data" if given only a site_id (cost one run,
# 2026-07-18). Look the domain up so the caller only needs the site_id.
DOMAIN=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -tA \
  -c "SELECT domain FROM sites WHERE id='${SITE_ID}'" 2>/dev/null | tr -d '[:space:]')
[ -n "$DOMAIN" ] || { echo "ERROR: no sites row for site_id ${SITE_ID}"; exit 1; }

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

INPUT_DATA="{\"site_id\":\"${SITE_ID}\",\"domain\":\"${DOMAIN}\",\"correlation_id\":\"${CORRELATION_ID}\"}"

echo "========================================="
echo "discovery run: ${AGENT_TYPE}"
echo "  site_id:     ${SITE_ID}"
echo "  domain:      ${DOMAIN}"
echo "  correlation: ${CORRELATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"

kubectl -n kafka run -i --rm "kcat-discovery-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=discovery-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$AGENT_TYPE"},"input_data":$INPUT_DATA}
JSON

echo
echo "dispatched. Findings (give it a minute or two):"
echo "  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "    \"SELECT item_type, severity, left(summary,110) FROM site_work_items WHERE site_id='${SITE_ID}' AND source='discovery' ORDER BY created_at DESC;\""
