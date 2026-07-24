#!/usr/bin/env bash
# deliver_via_section_editor.sh <site_id> <page_component_id> <domain> <page_name> <slot_name>
#
# Deliver a repaired tool template to the live page via the SANCTIONED
# section-editor path (bugs_closed/024 + features_open/009). content_edit with
# field_updates={} is a pure re-render from the CURRENT html_template (no LLM),
# respects the experience-loop ownership guard (tool pages are rebuild_policy=owned),
# git-commits and deploys. Used to deliver the grip-force restore (bugs_open/046).
#
# For a tool with an INLINE <script> this is sufficient. For a tool referencing an
# EXTERNAL /tools/assets/{fn}.js asset, follow with the assemble-only JS republish
# (gauntlet_dead_cta/scripts/republish_gauntlet_js.sh pattern) — apply_section_edit
# does NOT run collectJSAssets.
set -euo pipefail
SITE_ID="${1:?site_id}"; PCID="${2:?page_component_id}"; DOMAIN="${3:?domain}"
PAGE_NAME="${4:?page_name}"; SLOT_NAME="${5:?slot_name}"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "CORRELATION_ID=$CORRELATION_ID"
kubectl -n kafka run -i --rm "kcat-secedit-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" -H "client_id=demo_client" -H "action=process" \
  -H "sender_agent_type=cli" -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" -H "timestamp=$TIMESTAMP" <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_editor","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_editor":{"action":"spawn_agent","config":{"role":"editor","agent_type":"section-editor"},"output_field":"ed_agent","next_step":"call_editor","description":"Spawn section-editor"},"call_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"editor","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","page_component_id":"input_data.page_component_id","page_name":"input_data.page_name","slot_name":"input_data.slot_name","edit_type":"input_data.edit_type","field_updates":"input_data.field_updates"},"timeout_seconds":260},"output_field":"ed_result","next_step":"complete","description":"content_edit re-render from current template + deploy"},"complete":{"action":"complete_workflow","config":{"output_fields":["ed_result"]},"description":"done"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","page_component_id":"${PCID}","page_name":"${PAGE_NAME}","slot_name":"${SLOT_NAME}","edit_type":"content_edit","field_updates":{}}}
JSON
echo ""
echo "Watch: SELECT status,current_step FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid ORDER BY updated_at DESC;"
