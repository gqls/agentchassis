#!/bin/bash
# ==========================================================================
# Trigger the vet-pipeline-orchestrator
# ==========================================================================
#
# Usage:
#   bash trigger_vet_pipeline.sh                     # defaults: 50 areas, 500 promote, 100 verify
#   bash trigger_vet_pipeline.sh --area-code BT      # Belfast only
#   bash trigger_vet_pipeline.sh --limit 5           # only 5 areas to sweep
#   bash trigger_vet_pipeline.sh --verify-limit 20   # only verify 20 businesses
#   bash trigger_vet_pipeline.sh --dry-run            # show message without sending
#
# What happens:
#   1. Dispatches area-sweep-discoverer for unswept postcode districts (fire-and-forget)
#   2. Promotes pending discovery_candidates into businesses (from previous sweeps)
#   3. Dispatches vet-practice-verifier for pending businesses (fire-and-forget)
#
# This is a rolling pipeline — each run advances work from previous runs.
# ==========================================================================

set -euo pipefail

# Defaults
SWEEP_LIMIT=50
PROMOTE_LIMIT=500
VERIFY_LIMIT=100
AREA_CODE=""
DELAY_MS=200
DRY_RUN=false
CLIENT_ID="vetcomparison"

# Parse arguments
while [[ $# -gt 0 ]]; do
case $1 in
--area-code)    AREA_CODE="$2"; shift 2 ;;
--limit)        SWEEP_LIMIT="$2"; shift 2 ;;
--promote-limit) PROMOTE_LIMIT="$2"; shift 2 ;;
--verify-limit) VERIFY_LIMIT="$2"; shift 2 ;;
--delay-ms)     DELAY_MS="$2"; shift 2 ;;
--dry-run)      DRY_RUN=true; shift ;;
--client-id)    CLIENT_ID="$2"; shift 2 ;;
*) echo "Unknown option: $1"; exit 1 ;;
esac
done
----------------------
------- v 1  ------------

SWEEP_LIMIT=50
PROMOTE_LIMIT=500
VERIFY_LIMIT=100
AREA_CODE=""
DELAY_MS=5000
COUNTRY="GB"
BUSINESS_TYPE="veterinary practice"
VERTICAL_SLUG="veterinary"
DRY_RUN=false
CLIENT_ID="vetcomparison"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
ORCHESTRATION_NAME="vet-pipeline-$(date +%Y%m%d-%H%M%S)"

KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.generic.requests"

# Build input_data JSON
INPUT_DATA="{\"limit\":${SWEEP_LIMIT},\"promote_limit\":${PROMOTE_LIMIT},\"verify_limit\":${VERIFY_LIMIT},\"delay_ms\":${DELAY_MS},\"country\":\"${COUNTRY}\",\"business_type\":\"${BUSINESS_TYPE}\",\"vertical_slug\":\"${VERTICAL_SLUG}\""
if [ -n "$AREA_CODE" ]; then
  INPUT_DATA="${INPUT_DATA},\"area_code\":\"${AREA_CODE}\""
fi
INPUT_DATA="${INPUT_DATA}}"


echo "========================================="
echo "Vet Pipeline Orchestrator"
echo "========================================="
echo "  Sweep limit:     ${SWEEP_LIMIT}"
echo "  Promote limit:   ${PROMOTE_LIMIT}"
echo "  Verify limit:    ${VERIFY_LIMIT}"
echo "  Area code:       ${AREA_CODE:-all}"
echo "  Delay (ms):      ${DELAY_MS}"
echo "  Client:          ${CLIENT_ID}"
echo "  Orchestration:   ${ORCHESTRATION_NAME}"
echo "  Country:         ${COUNTRY}"
echo "  Business type:   ${BUSINESS_TYPE}"
echo "  Vertical:        ${VERTICAL_SLUG}"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

if [ "$DRY_RUN" = true ]; then
echo "DRY RUN — message body:"
echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
echo ""
echo "Headers:"
echo "  correlation_id=$CORRELATION_ID"
echo "  orchestration_id=$ORCHESTRATION_ID"
echo "  orchestration_name=$ORCHESTRATION_NAME"
echo "  client_id=$CLIENT_ID"
echo "  action=orchestrate"
exit 0
fi

kubectl -n kafka run -i --rm kcat-pipeline-$$ \
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
    -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"vet-pipeline-orchestrator"},"input_data":${INPUT_DATA}}
JSON

echo ""
echo "========================================="
echo "Pipeline started!"
echo "========================================="
echo ""
echo "MONITORING:"
echo ""
echo "1. Watch logs:"
echo "   kubectl logs -n ai-persona-system -l app=agent-chassis -f | grep '$CORRELATION_ID'"
echo ""
echo "2. Check orchestration status:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT status, current_step FROM orchestration_states WHERE orchestration_id = '$ORCHESTRATION_ID'\""
echo ""
echo "3. Check pipeline output (after completion):"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT final_result FROM orchestration_states WHERE orchestration_id = '$ORCHESTRATION_ID'\""
echo ""
echo "4. Monitor discovery candidates:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT status, COUNT(*) FROM business_intel.discovery_candidates GROUP BY status\""
echo ""
echo "5. Monitor business verification:"
echo "   kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "     \"SELECT verification_status, COUNT(*) FROM business_intel.businesses b JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id WHERE bv.slug = 'veterinary' GROUP BY verification_status\""
echo ""

echo "  -- Did ensure_tasks actually create them?"
echo "  SELECT status, COUNT(*) FROM business_intel.collection_tasks GROUP BY status;"
echo "  "
echo "  -- What did the batch processor's child orchestration do?"
echo "  SELECT orchestration_id, status, current_step, error"
echo "  FROM orchestration_states"
echo "  WHERE parent_orchestration_id = '$ORCHESTRATION_ID'"
echo "  ORDER BY created_at;"
echo "  "
echo "  -- And check the batch processor's own collected_data"
echo "  SELECT collected_data->'batch' as batch_data"
echo "  FROM orchestration_states"
echo "  WHERE owner_agent_type = 'vet-batch-processor'"
echo "  ORDER BY created_at DESC"
echo "  LIMIT 1;"
echo ""