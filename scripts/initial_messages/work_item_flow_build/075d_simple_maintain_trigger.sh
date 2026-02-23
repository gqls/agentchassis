#!/bin/bash
# 075d_trigger_maintenance.sh — Trigger maintenance dispatch loop
# Usage: ./075d_trigger_maintenance.sh finetuning.uk

set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain>}"

-------------------

DOMAIN="finetuning.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "========================================="
echo "Maintenance Dispatch"
echo "  Domain:  $DOMAIN"
echo "  CorrID:  $CORRELATION_ID"
echo "========================================="

kubectl -n kafka run -i --rm kcat-maint-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_orchestrator","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_orchestrator":{"action":"spawn_agent","config":{"role":"site_orchestrator","agent_type":"site-work-orchestrator"},"output_field":"spawn_orchestrator","next_step":"call_orchestrator","description":"Spawn site work orchestrator"},"call_orchestrator":{"action":"call_agent","config":{"target_role":"site_orchestrator","input_mapping":{"domain":"input_data.domain","mode":"input_data.mode"},"timeout_seconds":600},"output_field":"orchestrator_result","next_step":"complete","description":"Run maintenance dispatch loop"},"complete":{"action":"complete_workflow","config":{"output_fields":["orchestrator_result"]},"description":"Maintenance complete"}}}},"input_data":{"domain":"${DOMAIN}","mode":"maintenance"}}
JSON

echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
echo ""
echo "CORRELATION_ID=$CORRELATION_ID"

echo "  "
echo "   SELECT item_type, handler_agent, status, priority, "
echo "          spec->>'purpose' as purpose, "
echo "          spec->>'asset_id' as asset_id, "
echo "          result->>'commit_sha' as commit_sha, "
echo "          created_at::timestamp(0) as updated "
echo "   FROM site_work_items "
echo "   WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN') "
echo "   ORDER BY priority, created_at; "
echo "  "
echo "  "

SELECT item_type, handler_agent, status, priority,
       spec->>'purpose' as purpose,
       spec->>'asset_id' as asset_id,
       result->>'commit_sha' as commit_sha,
       created_at::timestamp(0) as updated
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain = 'finetuning.uk')
ORDER BY priority, created_at;