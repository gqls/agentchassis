#!/bin/bash
# Fire quality-discovery-agent at one domain. Envelope copied from
# scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh
#
# DELIBERATELY OMITTED from that script:
#   - its `case "$2"` which rejects "quality" (only design|completeness)
#   - its tail, which runs an UNCONDITIONAL
#       UPDATE site_work_items SET status='triaged' WHERE ... domain='finetuning.uk' AND status='detected'
#     — a hardcoded OTHER domain, which would make that site's detected items
#     dispatchable and trigger real rebuilds. See bugs_open/201 lane notes.
# This script DETECTS ONLY. It never changes a work item's status.

set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain>}"
AGENT_TYPE="quality-discovery-agent"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "  Domain: $DOMAIN"
echo "  Agent:  $AGENT_TYPE"
echo "  CorrID: $CORRELATION_ID"
echo "========================================="

kubectl -n kafka run -i --rm kcat-qdisc-$(date +%s) \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_discovery","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_discovery":{"action":"spawn_agent","config":{"role":"discoverer","agent_type":"${AGENT_TYPE}"},"output_field":"discoverer","next_step":"call_discovery","description":"Spawn quality discovery agent"},"call_discovery":{"action":"call_agent","config":{"agent_type":"${AGENT_TYPE}","target_role":"discoverer","input_mapping":{"domain":"input_data.domain"},"timeout_seconds":600},"output_field":"discovery_result","next_step":"complete","description":"Run quality discovery checks"},"complete":{"action":"complete_workflow","config":{"output_fields":["discovery_result"]},"description":"Discovery complete"}}}},"input_data":{"domain":"${DOMAIN}"}}
JSON

echo ""
echo "CORRELATION_ID=$CORRELATION_ID"
