#!/usr/bin/env bash
# 0NN_TRIGGER_feature_designer_v1.sh — manual feature-designer trigger (feature
# builder delta 1). Renumber 0NN when filing. Envelope mirrors
# 092_TRIGGER_fix_implementer_v1.sh (the proven kcat replay shape).
#
# TARGET: feature-designer (runs in-chassis — its only write surface is
# diagnosis_artifacts, so it needs no spawn-gate token, mirroring fix-proposer).
#
# INPUT: work_item_id — a site_work_items row with item_type='capability_gap'
# whose spec carries BOTH owner_approval AND code_pointers; the workflow
# refuses anything else (check_spec_approved gate). A FRESH fix_correlation_id
# is generated per run — SAVE IT: it keys the staged plan + council artifacts,
# and it is what you hand to the implementer trigger after approval.
#
# Owner approval, when granting it (spec jsonb, never a status value):
#   UPDATE site_work_items SET spec = spec ||
#     '{"owner_approval": {"approved_by": "<name>", "date": "YYYY-MM-DD"},
#       "code_pointers": [{"path": "platform/...", "why": "..."}]}'
#   WHERE id = '<work_item_id>';
#
# Usage:
#   ./0NN_TRIGGER_feature_designer_v1.sh <work_item_id>
#   WORK_ITEM=<uuid> ./0NN_TRIGGER_feature_designer_v1.sh
set -euo pipefail

WORK_ITEM="${1:-${WORK_ITEM:?work_item_id required (site_work_items.id of the approved capability_gap spec)}}"
TARGET_AGENT_TYPE='feature-designer'
CLIENT_ID='demo_client'

FIX_CORR=$(cat /proc/sys/kernel/random/uuid)
INPUT_DATA="{\"fix_correlation_id\":\"$FIX_CORR\",\"work_item_id\":\"$WORK_ITEM\"}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="featdesign-v1-$(date +%H%M%S)"

echo "========================================="
echo "Manual feature-designer trigger (staged plan + council)"
echo "========================================="
echo "  Work item (approved spec):     ${WORK_ITEM}"
echo "  FEATURE correlation (SAVE!):   ${FIX_CORR}"
echo "  Run correlation (envelope):    ${CORRELATION_ID}"
echo "  Orchestration:                 ${ORCHESTRATION_ID}"
echo "========================================="
echo "SAVE: FEATURE_CORR=${FIX_CORR}   RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-featdesign-$(date +%s)" \
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
echo "feature-designer triggered. Watch the run:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
echo ""
echo "The staged plan + council verdicts (keyed on the FEATURE correlation):"
echo "  SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts"
echo "  WHERE correlation_id='${FIX_CORR}' ORDER BY created_at;"
