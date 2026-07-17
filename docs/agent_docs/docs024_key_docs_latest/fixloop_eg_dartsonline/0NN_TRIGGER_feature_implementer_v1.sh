#!/usr/bin/env bash
# 0NN_TRIGGER_feature_implementer_v1.sh — manual stage-loop implementer trigger
# (feature builder delta 2). Renumber 0NN when filing. Envelope mirrors
# 092_TRIGGER_fix_implementer_v1.sh.
#
# TARGET: feature-implementer-orchestrator (NEVER feature-implementer directly —
# the wrapper spawns the dedicated pod so the spawn gate injects the read
# token; fired generically it dies at stage_read with no token, the fix loop's
# proven 2026-07-13 failure).
#
# INPUT: fix_correlation_id — the FEATURE correlation whose latest council
# report is 'approved' (from the designer trigger's output). Optional BASE_REF
# (default main). The run REFUSES a pre-existing feat/<short-corr> branch (E4)
# — delete stale ones first; that refusal is deliberate, not a bug.
#
# Usage:
#   ./0NN_TRIGGER_feature_implementer_v1.sh <feature_correlation_id>
#   FEATURE_CORR=<uuid> BASE_REF=main ./0NN_TRIGGER_feature_implementer_v1.sh
set -euo pipefail

FEATURE_CORR="${1:-${FEATURE_CORR:?feature correlation required (an APPROVED staged plan)}}"
BASE_REF="${BASE_REF:-main}"
TARGET_AGENT_TYPE='feature-implementer-orchestrator'
CLIENT_ID='demo_client'

INPUT_DATA="{\"fix_correlation_id\":\"$FEATURE_CORR\",\"base_ref\":\"$BASE_REF\"}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="featimpl-v1-$(date +%H%M%S)"

echo "========================================="
echo "Manual feature-implementer trigger (stage loop -> ONE PR)"
echo "========================================="
echo "  Feature correlation (approved plan): ${FEATURE_CORR}"
echo "  Base ref:                            ${BASE_REF}"
echo "  Run correlation (envelope):          ${CORRELATION_ID}"
echo "  Orchestration:                       ${ORCHESTRATION_ID}"
echo "========================================="
echo "SAVE: RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-featimpl-$(date +%s)" \
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
echo "feature-implementer triggered. Watch the stage loop:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
echo ""
echo "Per-stage commits land on feat/<first-8-of-corr>; the PR link is in the"
echo "run's collected_data (pr_result) when the end test gate passes."
