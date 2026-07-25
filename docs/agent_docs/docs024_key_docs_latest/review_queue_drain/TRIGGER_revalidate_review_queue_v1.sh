#!/usr/bin/env bash
# TRIGGER_revalidate_review_queue_v1.sh — fire the needs_human_review DRAIN
# (bugs_open/033; manual, v1; ships dry_run). Envelope mirrors
# TRIGGER_superseded_reviews_v1.sh — the proven replay envelope.
#
# TARGET: diagnosis-review-queue-revalidator (runs in-chassis; write surfaces,
# LIVE MODE ONLY: site_work_items.status/completed_at/resolution_path for
# 'resolved' verdicts, and result.revalidation on every item it judges).
#
# GOTCHA (rebalance window): never fire within ~300s of a chassis pod restart —
# the spawn is silently dropped.
#
# GOTCHA (dispatch latency): the sweep itself takes seconds, but the dispatch
# queues behind the fleet. Budget ~30 minutes to see a run start, and find your
# run by orchestration_id, not by created_at.
#
# Usage:
#   ./TRIGGER_revalidate_review_queue_v1.sh
set -euo pipefail

TARGET_AGENT_TYPE='diagnosis-review-queue-revalidator'
CLIENT_ID='demo_client'

INPUT_DATA='{}'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="revalqueue-v1-$(date +%H%M%S)"

echo "========================================="
echo "Manual review-queue revalidation (the 033 drain)"
echo "========================================="
echo "  Run correlation (envelope): ${CORRELATION_ID}"
echo "  Orchestration:              ${ORCHESTRATION_ID}"
echo "  Orchestration name:         ${ORCH_NAME}"
echo "========================================="
echo "SAVE: RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-revalqueue-$(date +%s)" \
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
echo "revalidation sweep triggered. Watch it:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
echo ""
echo "Per-item verdicts (in the run's payload):"
echo "  SELECT jsonb_pretty(collected_data->'complete'->'result'->'response') FROM orchestration_states"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}';"
echo ""
echo "What it closed (live mode only):"
echo "  SELECT id, item_type, result->'revalidation'->>'reason' FROM site_work_items"
echo "  WHERE resolution_path='auto:revalidated' ORDER BY completed_at DESC LIMIT 20;"
echo ""
echo "Queue depth:"
echo "  SELECT count(*) FROM site_work_items WHERE status='needs_human_review';"
