#!/usr/bin/env bash
# republish_gauntlet_js.sh
#
# Republish the gauntlet-interface JS asset (/tools/assets/gauntlet-interface.js)
# from content_components.js_content, by driving the page-rerender agent's
# assemble-only path via a DIRECT orchestrator envelope (the 086 pattern that
# routes reliably — NOT the bare action=orchestrate 049b envelope, which hit the
# kubectl-run stdin race and never ingested here).
#
# WHY: apply_section_edit reassembled + deployed the correct page HTML but does
# NOT run collectJSAssets. Only rerender_single_page (page-rerender's render_page
# step, no reason -> else branch) emits js_content to /tools/assets/{function}.js.
# The rendered HTML is already correct, so an assemble-only pass is a safe no-op
# for the HTML and republishes the fresh JS.
set -euo pipefail

SITE_ID="9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"
PAGE_ID="ecb637c1-845f-46bf-b174-9c92a43f9586"
DOMAIN="vonc.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "SAVE: CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm "kcat-gauntlet-js-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerender","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_rerender":{"action":"spawn_agent","config":{"role":"rerenderer","agent_type":"page-rerender"},"output_field":"rr_agent","next_step":"call_rerender","description":"Spawn page-rerender"},"call_rerender":{"action":"call_agent","config":{"agent_type":"page-rerender","target_role":"rerenderer","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","page_id":"input_data.page_id"},"timeout_seconds":260},"output_field":"rr_result","next_step":"complete","description":"Assemble-only rerender: republish js_content, redeploy assembled HTML"},"complete":{"action":"complete_workflow","config":{"output_fields":["rr_result"]},"description":"done"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","page_id":"${PAGE_ID}"}}
JSON

echo ""
echo "Watch:  SELECT status,current_step FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid;"
