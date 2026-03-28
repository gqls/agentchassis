#!/bin/bash
# ==========================================================================
# Trigger the vet-batch-processor directly via the vet-intel pod
# ==========================================================================
#
# Usage:
#   bash 076_trigger_vet_batch.sh           # default batch_size=100
#   bash 076_trigger_vet_batch.sh 50        # batch_size=50
#
# Skips sweeps/promote — just claims pending tasks and verifies them.
# ==========================================================================

set -euo pipefail

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="vet-batch-$(date +%Y%m%d-%H%M%S)"
CLIENT_ID="vetcomparison"
BATCH_SIZE=${1:-100}

KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.vet-intel.requests"
RESPONSES_TOPIC="system.agent.vet-intel.responses"

echo "========================================="
echo "Vet Batch Processor"
echo "========================================="
echo "  Batch size:      $BATCH_SIZE"
echo "  Client:          $CLIENT_ID"
echo "  Topic:           $TOPIC"
echo "  Orchestration:   $ORCHESTRATION_NAME"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm kcat-batch-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b "$KAFKA_BOOTSTRAP" \
    -t "$TOPIC" \
    -H "correlation_id=$CORRELATION_ID" \
    -H "request_id=$REQUEST_ID" \
    -H "message_id=$MESSAGE_ID" \
    -H "orchestration_id=$ORCHESTRATION_ID" \
    -H "orchestration_name=$ORCHESTRATION_NAME" \
    -H "step_name=start" \
    -H "client_id=$CLIENT_ID" \
    -H "message_type=request" \
    -H "action=orchestrate" \
    -H "from_agent_type=user" \
    -H "from_agent_id=cli" \
    -H "responses_topic=$RESPONSES_TOPIC" <<JSON
{"action":"orchestrate","config":{"agent_type":"vet-batch-processor"},"input_data":{"batch_size":$BATCH_SIZE,"task_type":"initial_verification","vertical_slug":"veterinary"}}
JSON

echo ""
echo "Batch processor started!"
echo "Monitor with:  make logs-vet-intel"
echo ""

