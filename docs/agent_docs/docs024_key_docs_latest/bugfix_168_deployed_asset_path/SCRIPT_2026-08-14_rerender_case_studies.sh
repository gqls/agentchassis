#!/usr/bin/env bash
# Assemble-only rerender + deploy of leopardessconsulting.co.uk /case-studies.
#
# Mirrors scripts/initial_messages/210_vonc_trigger/083_rerender-index-vonc.sh —
# spawn page-rerender, call it with {domain, page_id, site_id}. NO content
# rebuild: it re-assembles the existing page_components and deploys, so it will
# reprint the content_data this session edited and nothing else.
#
# ⚠ kcat -P can send NOTHING at exit 0 (LANDMINE). The proof of dispatch is the
# orchestration_states row, never this script's exit code.
set -euo pipefail

DOMAIN="leopardessconsulting.co.uk"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
PAGE_ID="ff5e75e3-547b-4a91-b00b-766d144ea4b9"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo "SAVE: ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-rerender-lc-casestudies-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerender","processing_mode":"orchestrator","timeout_seconds":240,"steps":{"spawn_rerender":{"action":"spawn_agent","config":{"role":"page_renderer","agent_type":"page-rerender"},"output_field":"rerender_agent","next_step":"call_rerender","description":"Spawn page-rerender"},"call_rerender":{"action":"call_agent","config":{"agent_type":"page-rerender","target_role":"page_renderer","input_mapping":{"domain":"input_data.domain","page_id":"input_data.page_id","site_id":"input_data.site_id"},"timeout_seconds":200},"output_field":"rerender_result","next_step":"complete","description":"Assemble stored components + deploy"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result"]},"description":"Rerender complete"}}}},"input_data":{"domain":"${DOMAIN}","page_id":"${PAGE_ID}","site_id":"${SITE_ID}"}}
JSON

echo ""
echo "Verify (the row, not the exit code):"
echo "  SELECT status, current_step FROM orchestration_states WHERE orchestration_id='${ORCHESTRATION_ID}';"
