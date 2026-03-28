#!/bin/bash
# ==========================================================================
# Trigger vet-practice-verifier for a single business via the vet-intel pod
# ==========================================================================
#
# Usage:
#   bash 071c_vet_verifier.sh <business_id>
#
# Get a business_id with:
#   psql -c "SELECT id, name, postcode FROM business_intel.businesses
#            WHERE verification_status = 'pending' LIMIT 5"
# ==========================================================================

set -euo pipefail

BUSINESS_ID=${1:?"Usage: $0 <business_id>"}

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="vet-single-$(date +%Y%m%d-%H%M%S)"
CLIENT_ID="vetcomparison"

KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.vet-intel.requests"
RESPONSES_TOPIC="system.agent.vet-intel.responses"

echo "========================================="
echo "Single Practice Verification"
echo "========================================="
echo "  Business ID:     $BUSINESS_ID"
echo "  Topic:           $TOPIC"
echo "  Orchestration:   $ORCHESTRATION_NAME"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm kcat-vet-single-$$ \
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
{"action":"orchestrate","config":{"agent_type":"vet-practice-verifier"},"input_data":{"business_id":"$BUSINESS_ID"}}
JSON

echo ""
echo "Verifier started!"
echo "Monitor with:  make logs-vet-intel"
echo ""
echo "  SELECT status, current_step FROM orchestration_states"
echo "  WHERE orchestration_id = '$ORCHESTRATION_ID';"
echo ""