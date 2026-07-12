#!/usr/bin/env bash
# 087_dispatch_work_items_vonc.sh
#
# Manually run ONE pass of the build-dispatch-loop for vonc.com — the same agent
# the (currently disabled) improvement-sweep scheduler drives. Loads up to 5
# dispatchable items (status triaged/approved, attempt_count < max_attempts,
# depends_on satisfied), claims each atomically, spawns the item's handler_agent,
# calls it, marks complete/failed.
#
# Ordering is enforced by data, not by this script: the three empty_section
# rebuilds carry depends_on -> their component's regeneration item, and the final
# needs_rerender depends on all three rebuilds. So the first pass runs only the
# regenerations; re-run this script when they complete to dispatch the next layer.
#
# Watch handlers in their spawned pods:
#   kubectl -n ai-persona-system get pods | grep -E 'component-creator|page-build|rerender'
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

SITE_ID="9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"
DOMAIN="vonc.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Dispatch pass: site_work_items  ($DOMAIN)"
echo "  Correlation: $CORRELATION_ID"
echo "========================================="
echo "SAVE: CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm "kcat-dispatch-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_dispatch","processing_mode":"orchestrator","timeout_seconds":3600,"steps":{"spawn_dispatch":{"action":"spawn_agent","config":{"role":"dispatcher","agent_type":"build-dispatch-loop"},"output_field":"dispatch_agent","next_step":"call_dispatch","description":"Spawn build-dispatch-loop"},"call_dispatch":{"action":"call_agent","config":{"agent_type":"build-dispatch-loop","target_role":"dispatcher","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain"},"timeout_seconds":3400},"output_field":"dispatch_result","next_step":"complete","description":"One dispatch pass: claim + run up to 5 dispatchable items"},"complete":{"action":"complete_workflow","config":{"output_fields":["dispatch_result"]},"description":"Dispatch pass complete"}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Watch items:"
echo "  SELECT item_key, status, attempt_count FROM site_work_items WHERE site_id='${SITE_ID}'::uuid AND status NOT IN ('rejected') ORDER BY priority, item_key;"
