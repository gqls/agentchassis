#!/usr/bin/env bash
# 092_TRIGGER_fix_implementer_v1.sh — manual fix-implementer trigger (F1.1b(c) write step).
# Envelope mirrors 084_TRIGGER_diagnose_v1.sh: kcat pod -> system.agent.generic.requests,
# action=orchestrate, config.agent_type=fix-proposer. Recovered verbatim from the
# raw envelope of run 5ca5dacb (2026-07-10) so the replay is exact.
#
# TARGET: fix-proposer (runs directly — its only write surface is diagnosis_artifacts,
# so it needs no spawn-gate token yet; that arrives with F1.1b(c)).
#
# INPUT: fix_correlation_id — the correlation of a plan the council APPROVED.
# The workflow refuses anything else (check_approved gate). Flow: read current
# file bodies -> whole-file implementation (LLM) -> deterministic allowlist ->
# create fix/* branch -> commit -> gofmt+build gate in a golang k8s Job ->
# (green) PR into main | (red) NO PR, branch + build log left for inspection.
#
# ROUND-COUNTING CAVEAT (deployed v1.0.1107): the shipped binary counts council
# rounds per fix_correlation_id, so PRE-EXISTING council_report rows on that
# correlation inflate the count and shorten the loop. For a fair run, clear them
# first (orchestration_id-scoped counting is fixed in source, rides next image):
#   DELETE FROM diagnosis_artifacts
#   WHERE correlation_id='<fix_corr>' AND kind='council_report';
#
# Usage:
#   ./091_TRIGGER_fix_proposer_v1.sh [fix_correlation_id]
#   FIX_CORR=e08c5b01-01ef-42ad-80d0-b77c50ec9e84 ./091_TRIGGER_fix_proposer_v1.sh
set -euo pipefail

FIX_CORR="${1:-${FIX_CORR:-e08c5b01-01ef-42ad-80d0-b77c50ec9e84}}"
TARGET_AGENT_TYPE='fix-implementer'
CLIENT_ID='demo_client'

INPUT_DATA="{\"fix_correlation_id\":\"$FIX_CORR\"}"

# Envelope correlation is the RUN's own id (separate from the fix target), matching
# how 5ca5dacb was fired. A fresh uuid keeps this run's orchestration_states distinct.
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="fiximpl-v1-$(date +%H%M%S)"

echo "========================================="
echo "Manual fix-implementer trigger (write step)"
echo "========================================="
echo "  Fix correlation (target diagnosis): ${FIX_CORR}"
echo "  Run correlation (envelope):         ${CORRELATION_ID}"
echo "  Orchestration:                      ${ORCHESTRATION_ID}"
echo "  Orchestration name:                 ${ORCH_NAME}"
echo "========================================="
echo "SAVE: RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-fiximpl-$(date +%s)" \
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
echo "fix-proposer triggered. Watch it with the run orchestration id:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
echo ""
echo "Council rounds (keyed on the fix correlation):"
echo "  SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts"
echo "  WHERE correlation_id='${FIX_CORR}' AND kind='council_report' ORDER BY created_at;"
