#!/usr/bin/env bash
# TRIGGER_superseded_reviews_v1.sh — fire the review-bypass reconciler
# (bugs_open/056 regen; manual, v1; ships dry_run). Envelope mirrors
# 096_TRIGGER_diagnosis_silent_check_v1.sh — the proven replay envelope.
#
# TARGET: diagnosis-superseded-reviews (runs in-chassis; write surfaces:
# agent_error_log REVIEW_SUPERSEDED_BY_PASSING_SAVE rows + the parked items'
# result.superseded_by_passing_save annotation — live mode only).
#
# GOTCHA (rebalance window): never fire within ~300s of a chassis pod restart —
# the spawn is silently dropped.
#
# Usage:
#   ./TRIGGER_superseded_reviews_v1.sh
set -euo pipefail

TARGET_AGENT_TYPE='diagnosis-superseded-reviews'
CLIENT_ID='demo_client'

INPUT_DATA='{}'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="supersedrev-v1-$(date +%H%M%S)"

echo "========================================="
echo "Manual superseded-reviews trigger (reconciliation sweep)"
echo "========================================="
echo "  Run correlation (envelope): ${CORRELATION_ID}"
echo "  Orchestration:              ${ORCHESTRATION_ID}"
echo "  Orchestration name:         ${ORCH_NAME}"
echo "========================================="
echo "SAVE: RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-supersedrev-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=$ORCH_NAME" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":$INPUT_DATA}
JSON

echo ""
echo "superseded-reviews sweep triggered. Watch it:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
echo ""
echo "Pair report (in the run's payload):"
echo "  SELECT jsonb_pretty(collected_data->'complete'->'result'->'response') FROM orchestration_states"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}';"
echo ""
echo "Evidence written (live mode only):"
echo "  SELECT occurred_at, error_message FROM agent_error_log"
echo "  WHERE error_code='REVIEW_SUPERSEDED_BY_PASSING_SAVE' ORDER BY occurred_at DESC LIMIT 10;"
