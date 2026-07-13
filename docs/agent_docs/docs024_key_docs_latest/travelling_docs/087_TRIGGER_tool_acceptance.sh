#!/usr/bin/env bash
# 087_TRIGGER_tool_acceptance.sh — run a Tier-4 acceptance cycle for one tool
# via tool-acceptance-agent (migration 145). Envelope mirrors 084/085/086
# (kcat → system.agent.generic.requests, action=orchestrate).
#
# GATED ON A CHASSIS DEPLOY: the agent's request_browser_run /
# judge_acceptance_results actions ride the chassis image built from the
# commit that added platform/orchestration/actions/tool_acceptance_actions.go.
# On an older image the workflow FAILS at request_run (unknown action). The
# script checks the registry cannot be probed from here, so it prints the
# ancestry check to run first.
#
# SAFETY: DRY-RUN BY DEFAULT — builds + validates the message and stops.
#     SEND=1 ./087_TRIGGER_tool_acceptance.sh
#
# Effects when SEND=1: a browser run against the LIVE page (read-only), then
# doc_notes writes — acceptance-run on pass, acceptance-fail + ONE improve_tool
# work item (handler tool-improver) on failure. That item WILL be dispatched by
# the build loop: a failing run starts the fix cycle. That is the point.
#
# Usage:
#   SPEC_FUNCTION=tool-xp-curve-designer SEND=1 ./087_TRIGGER_tool_acceptance.sh
set -euo pipefail

TARGET_AGENT_TYPE='tool-acceptance-agent'
DOMAIN="${DOMAIN:-gamesdesign.co.uk}"
SITE_ID="${SITE_ID:-e33263f4-74f8-494f-b191-546845dbbddf}"
SPEC_FUNCTION="${SPEC_FUNCTION:-tool-xp-curve-designer}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

# Single line — kcat -P is line-delimited (the 086 lesson).
MESSAGE_BODY="{\"action\":\"orchestrate\",\"config\":{\"agent_type\":\"$TARGET_AGENT_TYPE\"},\"input_data\":{\"site_id\":\"$SITE_ID\",\"domain\":\"$DOMAIN\",\"spec\":{\"function\":\"$SPEC_FUNCTION\"}}}"
case "$MESSAGE_BODY" in *$'\n'*) echo "ERROR: body is multi-line" >&2; exit 1;; esac

echo "========================================="
echo "Tier-4 acceptance cycle via tool-acceptance-agent"
echo "========================================="
echo "  Function:      ${SPEC_FUNCTION}"
echo "  Site:          ${DOMAIN}  (${SITE_ID})"
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "========================================="
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"
echo ""

if [ "${SEND:-0}" != "1" ]; then
  echo ">>> DRY RUN (SEND != 1). Nothing produced."
  echo ">>> FIRST verify the deployed chassis carries the actions:"
  echo ">>>   git log --oneline -1 -- platform/orchestration/actions/tool_acceptance_actions.go"
  echo ">>>   git merge-base --is-ancestor <that-commit> <release-commit> && echo carried"
  echo ">>> Then:  SEND=1 $0"
  exit 0
fi

kubectl -n kafka run -i --rm "kcat-accept-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-acceptance-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=demo_client" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
$MESSAGE_BODY
JSON

echo ""
echo "acceptance cycle triggered."
echo ""
echo "1) Orchestration (by correlation, never created_at):"
echo "  SELECT status, current_step, EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s,"
echo "         substring(COALESCE(error,''),1,200) AS err"
echo "  FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid ORDER BY created_at;"
echo ""
echo "2) The verdict artifacts:"
echo "  SELECT subject_key, categories, source, left(body,120) AS head, created_at"
echo "  FROM doc_notes WHERE source='tool-acceptance' ORDER BY created_at DESC LIMIT 3;"
echo "  SELECT item_key, status FROM site_work_items WHERE item_key LIKE 'acceptance_fail:%';"
echo ""
echo "Timing: ~10-30s (browser launch + navigate + settle + judge)."
