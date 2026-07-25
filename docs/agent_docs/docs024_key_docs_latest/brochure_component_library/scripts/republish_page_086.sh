#!/usr/bin/env bash
# republish_page_086.sh <page_id> <site_id> <domain>
#
# Assemble-only page rerender via a DIRECT orchestrator envelope (the 086
# pattern). Use this INSTEAD OF 049b_deploy_single_page.sh: the bare
# `action=orchestrate` envelope hits the kubectl-run stdin race and can fail
# silently — no orchestration row, no work item, no log. That happened here on
# 2026-07-25: four 049b calls for fundamentallyai.com produced zero
# orchestration rows, so a completed link fix never reached the live pages.
#
# Assemble-only = no `reason` stamped, so page-rerender takes the render_page
# else-branch: it reuses stored content_data/rendered_html (no LLM, no content
# regeneration) and redeploys the assembled HTML. Safe for a data fix applied
# directly to page_components.
#
# Verify by PAYLOAD, not by the printed correlation id:
#   SELECT status,current_step FROM orchestration_states
#    WHERE initial_request_data->'input_data'->>'page_id' = '<page_id>'
#    ORDER BY created_at DESC LIMIT 1;
set -euo pipefail

PAGE_ID="${1:?page_id required}"
SITE_ID="${2:?site_id required}"
DOMAIN="${3:?domain required}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "page=$PAGE_ID domain=$DOMAIN CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm "kcat-republish-$(date +%s)-$RANDOM" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" \
  -H "client_id=$CLIENT_ID" \
  -H "action=process" \
  -H "sender_agent_type=cli" \
  -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" \
  -H "timestamp=$TIMESTAMP" <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerender","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_rerender":{"action":"spawn_agent","config":{"role":"rerenderer","agent_type":"page-rerender"},"output_field":"rr_agent","next_step":"call_rerender","description":"Spawn page-rerender"},"call_rerender":{"action":"call_agent","config":{"agent_type":"page-rerender","target_role":"rerenderer","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","page_id":"input_data.page_id"},"timeout_seconds":260},"output_field":"rr_result","next_step":"complete","description":"Assemble-only rerender: redeploy assembled HTML from stored content"},"complete":{"action":"complete_workflow","config":{"output_fields":["rr_result"]},"description":"done"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","page_id":"${PAGE_ID}"}}
JSON

echo "dispatched $PAGE_ID"
