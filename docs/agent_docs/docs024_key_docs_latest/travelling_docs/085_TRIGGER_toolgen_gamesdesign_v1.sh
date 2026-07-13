#!/usr/bin/env bash
# 085_TRIGGER_toolgen_gamesdesign_v1.sh — Task-3 PROOF: create ONE new tool on
# gamesdesign.co.uk via tool-generator, and watch it write its own PLAN.
# Envelope mirrors 084 (kcat pod -> system.agent.generic.requests, action=orchestrate).
#
# REAL SIDE EFFECTS (accepted 2026-07-07: gamesdesign is the trial site):
#   - a content_components row (the tool), a pages row, a nav entry on the LIVE
#     domain, a rerender/deploy through the normal pipeline;
#   - AND (Task 3) a doc_plans row, source='tool-generator' — the proof.
#
# Default spec: tool-xp-curve-designer — on-theme for a games-design site,
# self-contained, no external data. Override any field via env.
#
# Usage:
#   ./085_TRIGGER_toolgen_gamesdesign_v1.sh
#   SPEC_FUNCTION=tool-drop-rate-tuner SPEC_NAME="Drop Rate Tuner" \
#     SPEC_DESC="..." ./085_TRIGGER_toolgen_gamesdesign_v1.sh
set -euo pipefail

TARGET_AGENT_TYPE='tool-generator'
DOMAIN="${DOMAIN:-gamesdesign.co.uk}"
SITE_ID="${SITE_ID:-e33263f4-74f8-494f-b191-546845dbbddf}"
SPEC_FUNCTION="${SPEC_FUNCTION:-tool-xp-curve-designer}"
SPEC_NAME="${SPEC_NAME:-XP Curve Designer}"
SPEC_DESC="${SPEC_DESC:-Design and compare XP level curves. Pick a curve type (linear, quadratic, exponential), set base XP and growth, and see a level-by-level XP table plus a total-XP chart so designers can tune progression pacing.}"
SPEC_COMPLEXITY="${SPEC_COMPLEXITY:-simple}"

# input_data — string concat, matching 082/084 (no jq dependency)
INPUT_DATA="{\"domain\":\"$DOMAIN\",\"site_id\":\"$SITE_ID\",\"spec\":{\"function\":\"$SPEC_FUNCTION\",\"name\":\"$SPEC_NAME\",\"description\":\"$SPEC_DESC\",\"complexity\":\"$SPEC_COMPLEXITY\"}}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'

echo "========================================="
echo "Manual tool creation via tool-generator (Task-3 PROOF run)"
echo "========================================="
echo "  Site:          ${DOMAIN}  (${SITE_ID})"
echo "  Function:      ${SPEC_FUNCTION}"
echo "  Name:          ${SPEC_NAME}"
echo "  SIDE EFFECTS:  REAL component + page + nav on the live site, plus the"
echo "                 Task-3 doc_plans row (source='tool-generator')"
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"
echo ""

kubectl -n kafka run -i --rm "kcat-toolgen-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-toolgen-$(date +%Y%m%d-%H%M%S)" \
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
echo "========================================="
echo "tool creation triggered."
echo "========================================="
echo ""
echo "1) Orchestration state (by correlation id, never by created_at):"
echo "  SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s,"
echo "         substring(COALESCE(error,''),1,200) AS err"
echo "  FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;"
echo ""
echo "2) TASK-3 PROOF (once COMPLETED) — the tool wrote its own PLAN:"
echo "  SELECT subject_key, is_current, body LIKE '%\`\`\`criteria%' AS has_fence,"
echo "         length(body) AS body_len, source, created_at"
echo "  FROM doc_plans WHERE source = 'tool-generator' ORDER BY created_at DESC LIMIT 3;"
echo ""
echo "3) The artefacts (component + page):"
echo "  SELECT id::text, function, component_level, is_active FROM content_components"
echo "  WHERE function = '$SPEC_FUNCTION';"
echo "  SELECT name, url, status FROM pages WHERE site_id = '$SITE_ID'::uuid"
echo "  ORDER BY created_at DESC LIMIT 3;"
echo ""
echo "4) COMPOSER REVIEW (one-time): paste the PLAN body —"
echo "  SELECT body FROM doc_plans WHERE subject_type='tool' AND subject_key='$SPEC_FUNCTION' AND is_current;"
echo ""
echo "Timing: two Sonnet calls (generate + compose_plan); workflow timeout 480s."
