#!/usr/bin/env bash
# regenerate_via_tool_improver.sh <site_id> <component_id> <domain>
#
# Regenerate a TRUNCATED tool template in place via tool-improver (bugs_open/046).
# Why this mechanism (and not tool-generator / needs_tool_recreation):
#   - tool-generator's create_tool_component NO-OPs when an active tool exists on
#     the site (already_exists) and its birth path always INSERTs a new page —
#     wrong shape for repairing an existing placed tool.
#   - needs_tool_recreation regenerates from adoption-crawl interactive_features;
#     these tools are platform-generated, there is no adoption fingerprint.
#   - tool-improver rewrites the existing component's html_template from the
#     current (truncated) template + an issue statement, through the component
#     write guard (whose comparative checks deliberately allow a rewrite to land
#     on an already-broken row), and its migration-195 tail auto-emits the
#     section_edit delivery item (section-editor, sanctioned owned-page path).
# NB the section_edit delivery item rides the cron-starved build-dispatch-loop
# (bugs_open/030); if delivery lags, drive deliver_via_section_editor.sh directly.
set -euo pipefail
SITE_ID="${1:?site_id}"; COMPONENT_ID="${2:?component_id}"; DOMAIN="${3:?domain}"

ISSUE="This tool's html_template is a TRUNCATED LLM generation: a <script> block is opened but never closed - the original completion was cut mid-stream, so the tool's JavaScript never runs and, on the live page, all markup after the cut is swallowed as script text. Rewrite this tool as a COMPLETE self-contained fragment: keep the design, controls and behaviour already visible in the current HTML (the markup and CSS are largely intact; the JavaScript is cut partway through), complete the JavaScript so every control declared in the markup actually works end-to-end, and terminate every tag - every <script>, <style>, <section>, <div> and <fieldset> you open must be closed. The fragment must end on a closing tag. Do not invent external data sources or embed fabricated datasets; the tool is fully self-contained."

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm "kcat-toolimp-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" -H "client_id=demo_client" -H "action=process" \
  -H "sender_agent_type=cli" -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" -H "timestamp=$TIMESTAMP" <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_improver","processing_mode":"orchestrator","timeout_seconds":700,"steps":{"spawn_improver":{"action":"spawn_agent","config":{"role":"improver","agent_type":"tool-improver"},"output_field":"imp_agent","next_step":"call_improver","description":"Spawn tool-improver"},"call_improver":{"action":"call_agent","config":{"agent_type":"tool-improver","target_role":"improver","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","component_id":"input_data.component_id","issue":"input_data.issue"},"timeout_seconds":650},"output_field":"imp_result","next_step":"complete","description":"Rewrite the truncated tool template in place"},"complete":{"action":"complete_workflow","config":{"output_fields":["imp_result"]},"description":"done"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","component_id":"${COMPONENT_ID}","issue":"${ISSUE}"}}
JSON
echo ""
echo "Watch: SELECT status,current_step FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid ORDER BY updated_at DESC;"
