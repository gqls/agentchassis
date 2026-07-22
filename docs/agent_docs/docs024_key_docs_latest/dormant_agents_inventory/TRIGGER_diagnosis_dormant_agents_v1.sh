#!/usr/bin/env bash
# TRIGGER_diagnosis_dormant_agents_v1.sh — fire the dormant-agents capability
# inventory sweep (manual, v1; ships dry_run). Envelope mirrors the sibling
# 096_TRIGGER_diagnosis_silent_check: kcat pod -> system.agent.generic.requests,
# action=orchestrate, config.agent_type=diagnosis-dormant-agents.
#
# TARGET: diagnosis-dormant-agents (runs in-chassis; its only write surfaces are
# site_work_items rows it owns (created_by='diagnosis-dormant-agents',
# item_type='dormant_agent') and doc_notes). bugs_open/044.
#
# INPUT: none required. Inventories active agents by the step-fingerprint method;
# emits INERT dormant_agent items for human triage unless dry_run; closes ones
# whose agent has since run; persists a report. Read the latest with:
#   SELECT body FROM doc_notes WHERE categories ? 'dormant-agents'
#   ORDER BY created_at DESC LIMIT 1;
#
# GOTCHA (rebalance window): never fire within ~300s of a chassis pod restart —
# the spawn is silently dropped.
#
# Usage:
#   ./TRIGGER_diagnosis_dormant_agents_v1.sh
set -euo pipefail

TARGET_AGENT_TYPE='diagnosis-dormant-agents'
CLIENT_ID='demo_client'

INPUT_DATA='{}'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="dormantchk-v1-$(date +%H%M%S)"

echo "========================================="
echo "Manual dormant-agents trigger (capability inventory sweep)"
echo "========================================="
echo "  Run correlation (envelope): ${CORRELATION_ID}"
echo "  Orchestration:              ${ORCHESTRATION_ID}"
echo "  Orchestration name:         ${ORCH_NAME}"
echo "========================================="
echo "SAVE: RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-dormantchk-$(date +%s)" \
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
echo "dormant-agents triggered. Watch it with the run orchestration id:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
echo ""
echo "Latest report:"
echo "  SELECT body FROM doc_notes WHERE categories ? 'dormant-agents'"
echo "  ORDER BY created_at DESC LIMIT 1;"
echo ""
echo "Findings written (live mode only):"
echo "  SELECT item_key, status, left(summary,90) FROM site_work_items"
echo "  WHERE created_by='diagnosis-dormant-agents' ORDER BY created_at DESC LIMIT 10;"
