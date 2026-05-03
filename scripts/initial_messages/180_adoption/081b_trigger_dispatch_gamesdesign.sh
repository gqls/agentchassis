#!/usr/bin/env bash
# Manually point build-dispatch-loop at gamesdesign.co.uk so its 9 open
# work items get claimed and handled. Needed because build-pipeline-trigger
# is scheduler-driven with a fixed input that defaults to system.internal,
# so the dispatcher never fires for real sites on its own.

set -euo pipefail

SITE_ID="055e6ef4-971f-4419-a19c-8fad11646e59"
DOMAIN="gamesdesign.co.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"

echo "========================================="
echo "Manual Dispatch Trigger"
echo "========================================="
echo "  Target site_id:      $SITE_ID"
echo "  Target domain:       $DOMAIN"
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm kcat-dispatch-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=manual-dispatch-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"build-dispatch-loop"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON

