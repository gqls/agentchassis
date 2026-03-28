#!/bin/bash
# ==========================================================================
# Trigger the vet-pipeline-orchestrator via the vet-intel pod
# ==========================================================================
#
# Usage:
#   bash 072e_trigger_vet_pipeline.sh                     # defaults
#   bash 072e_trigger_vet_pipeline.sh --area-code BT      # Belfast only
#   bash 072e_trigger_vet_pipeline.sh --limit 5           # only 5 areas
#   bash 072e_trigger_vet_pipeline.sh --verify-limit 20   # only verify 20
#   bash 072e_trigger_vet_pipeline.sh --dry-run            # show message
#
# What happens:
#   1. Sweeps unswept postcode districts for vet practices
#   2. Promotes pending discovery candidates into businesses
#   3. Ensures collection tasks exist for pending businesses
#   4. Runs batch verification (scrape + LLM extraction)
#
# This is a rolling pipeline — each run advances work from previous runs.
# ==========================================================================

set -euo pipefail

# Defaults
SWEEP_LIMIT=0
PROMOTE_LIMIT=500
VERIFY_LIMIT=100
AREA_CODE=""
DELAY_MS=5000
COUNTRY="GB"
BUSINESS_TYPE="veterinary practice"
VERTICAL_SLUG="veterinary"
DRY_RUN=false
CLIENT_ID="vetcomparison"

# Parse arguments
while [[ $# -gt 0 ]]; do
case $1 in
--area-code)     AREA_CODE="$2"; shift 2 ;;
--limit)         SWEEP_LIMIT="$2"; shift 2 ;;
--promote-limit) PROMOTE_LIMIT="$2"; shift 2 ;;
--verify-limit)  VERIFY_LIMIT="$2"; shift 2 ;;
--delay-ms)      DELAY_MS="$2"; shift 2 ;;
--dry-run)       DRY_RUN=true; shift ;;
--client-id)     CLIENT_ID="$2"; shift 2 ;;
*) echo "Unknown option: $1"; exit 1 ;;
esac
done

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
ORCHESTRATION_NAME="vet-pipeline-$(date +%Y%m%d-%H%M%S)"

KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.vet-intel.requests"
RESPONSES_TOPIC="system.agent.vet-intel.responses"

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
echo "  Topic:           ${TOPIC}"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

if [ "$DRY_RUN" = true ]; then
echo "DRY RUN — message:"
echo "{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"vet-pipeline-orchestrator\"},\"input_data\":${INPUT_DATA}}" | python3 -m json.tool 2>/dev/null || echo "$INPUT_DATA"
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
    -H "responses_topic=$RESPONSES_TOPIC" <<JSON
{"action":"orchestrate","config":{"agent_type":"vet-pipeline-orchestrator"},"input_data":${INPUT_DATA}}
JSON

echo ""
echo "Pipeline started!"
echo ""
echo "Monitor with:  make logs-vet-intel"
echo ""
echo "DB status:"
echo "  SELECT"
echo "      (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'completed') as tasks_done,"
echo "      (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'in_progress') as tasks_active,"
echo "      (SELECT COUNT(*) FROM business_intel.collection_tasks WHERE status = 'pending') as tasks_pending,"
echo "      (SELECT COUNT(*) FROM business_intel.businesses b"
echo "       JOIN business_intel.business_verticals bv ON bv.id = b.vertical_id"
echo "       WHERE bv.slug = 'veterinary' AND b.verification_status = 'verified') as verified,"
echo "      (SELECT COUNT(*) FROM business_intel.business_prices WHERE is_current = TRUE) as current_prices;"
echo ""
echo "  SELECT orchestration_id, status, current_step"
echo "  FROM orchestration_states"
echo "  WHERE orchestration_id = '$ORCHESTRATION_ID';"
echo ""

