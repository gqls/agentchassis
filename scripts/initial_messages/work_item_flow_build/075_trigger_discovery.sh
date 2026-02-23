#!/bin/bash
# trigger_discovery.sh
# Test the discovery agents against a site.
# Start here — these are the simplest agents (DB queries only, no LLM/git).
#
# Usage:
#   ./trigger_discovery.sh finetuning.uk design         # design checks only
#   ./trigger_discovery.sh finetuning.uk completeness   # content checks only
#   ./trigger_discovery.sh finetuning.uk                # defaults to design

set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain> [design|completeness]}"
AGENT_TYPE="${2:-design}-discovery-agent"

# Validate agent type
case "$2" in
    design|"")      AGENT_TYPE="design-discovery-agent" ;;
    completeness)   AGENT_TYPE="completeness-discovery-agent" ;;
    *)              echo "Unknown check type: $2 (use: design, completeness)"; exit 1 ;;
esac

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Discovery Agent Trigger"
echo "========================================="
echo "  Domain:         $DOMAIN"
echo "  Agent:          $AGENT_TYPE"
echo "  Correlation ID: $CORRELATION_ID"
echo "========================================="

# The discovery agent needs site_id or domain in input_data.
# We pass domain — ensure_site_record resolves to site_id.
INPUT_DATA="{\"domain\":\"${DOMAIN}\"}"

kubectl -n kafka run -i --rm kcat-discovery-$(date +%s) \
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
  -H client_id=$CLIENT_ID \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_discovery","processing_mode":"orchestrator","timeout_seconds":120,"steps":{"spawn_discovery":{"action":"spawn_agent","config":{"role":"discoverer","agent_type":"${AGENT_TYPE}"},"output_field":"discoverer","next_step":"call_discovery","description":"Spawn discovery agent"},"call_discovery":{"action":"call_agent","config":{"agent_type":"${AGENT_TYPE}","target_role":"discoverer","input_mapping":{"domain":"input_data.domain"},"timeout_seconds":60},"output_field":"discovery_result","next_step":"complete","description":"Run discovery checks"},"complete":{"action":"complete_workflow","config":{"output_fields":["discovery_result"]},"description":"Discovery complete"}}}},"input_data":${INPUT_DATA}}
JSON

echo ""
echo "========================================="
echo "Discovery triggered"
echo "========================================="
echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
echo ""
echo "Check findings:"
echo "  kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c \\"
echo "    \"SELECT id, item_type, severity, summary, handler_agent, status, priority FROM site_work_items WHERE batch_id IS NOT NULL ORDER BY created_at DESC LIMIT 20;\""
echo ""
echo "CORRELATION_ID=$CORRELATION_ID"
