#!/usr/bin/env bash
# 096_TRIGGER_diagnosis_silent_check_v1.sh — fire the silent-failure verification
# checker (manual, v1; ships dry_run). Envelope mirrors 095/084: kcat pod ->
# system.agent.generic.requests, action=orchestrate,
# config.agent_type=diagnosis-silent-check — the proven replay envelope.
#
# TARGET: diagnosis-silent-check (runs in-chassis; its only write surfaces are
# site_work_items rows it owns (created_by='silent-check') and doc_notes).
#
# INPUT: none required. Verifies structural invariants (nav-linked pages never
# built; deployed pages with zero components report-only), emits INERT
# silent_failure items for triage unless dry_run, closes resolved ones, and
# persists a report. Read the latest with:
#   SELECT body FROM doc_notes WHERE categories ? 'silent-check'
#   ORDER BY created_at DESC LIMIT 1;
#
# GOTCHA (rebalance window): never fire within ~300s of a chassis pod restart —
# the spawn is silently dropped.
#
# Usage:
#   ./096_TRIGGER_diagnosis_silent_check_v1.sh
set -euo pipefail

TARGET_AGENT_TYPE='diagnosis-silent-check'
CLIENT_ID='demo_client'

INPUT_DATA='{}'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="silentchk-v1-$(date +%H%M%S)"

echo "========================================="
echo "Manual silent-check trigger (verification sweep)"
echo "========================================="
echo "  Run correlation (envelope): ${CORRELATION_ID}"
echo "  Orchestration:              ${ORCHESTRATION_ID}"
echo "  Orchestration name:         ${ORCH_NAME}"
echo "========================================="
echo "SAVE: RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-silentchk-$(date +%s)" \
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
echo "silent-check triggered. Watch it with the run orchestration id:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
echo ""
echo "Latest report:"
echo "  SELECT body FROM doc_notes WHERE categories ? 'silent-check'"
echo "  ORDER BY created_at DESC LIMIT 1;"
echo ""
echo "Findings written (live mode only):"
echo "  SELECT item_key, status, left(summary,90) FROM site_work_items"
echo "  WHERE created_by='silent-check' ORDER BY created_at DESC LIMIT 10;"
