#!/usr/bin/env bash
# 294_TRIGGER_improvement_loop_v1.sh — fire the improvement-loop at ONE site.
#
# WHY THIS EXISTS. 004_improvement_loop.md line 265 documents the manual path as
# `./trigger-audit.sh improvement-loop <site_id> <domain>`. That script is not in
# the repo and `find` across the tree returns nothing — the doc has outlived it.
# The improvement-loop is therefore reachable only by its scheduled task,
# `improvement-sweep`, which has been `enabled=f` since 2026-05-02. So on
# 2026-08-03 there was NO way to run discovery+triage+dispatch against a chosen
# site without hand-writing a Kafka envelope. This is that envelope, kept.
#
# WHAT THE LOOP DOES, and why it is the whole chain rather than just an audit:
# ensure_site_record -> the discovery agents (design, quality, completeness) and
# the audit agents -> triage_findings (the ONLY promoter of detected -> triaged
# in the platform) -> call_dispatch (build-dispatch-loop, which claims triaged
# items and runs their handler_agent). Running a discovery agent ALONE leaves
# every finding at 'detected', which the dispatcher's claim query
# (`status IN ('triaged','approved')`) cannot see — 004 line 305 says so and the
# fleet proves it: measured 2026-08-03, 204 detected items across 10 sites
# against 2 triaged.
#
# ENVELOPE mirrors 097_TRIGGER_council_review_v1.sh, which is the trigger on this
# estate that is known-good. Topic is system.agent.scheduled.requests because
# that is what the improvement-sweep row itself targets — read from
# scheduled_tasks.target_topic, not assumed.
#
# TWO PRE-FLIGHT RULES, both from CLAUDE.md, both enforced below rather than
# left to the operator:
#   - no dispatch within ~300s of a chassis pod (re)start; the spawn is silently
#     dropped, which looks exactly like a queue that is merely slow.
#   - check the QUEUE, not just the pod: another session may already have work in
#     flight against this site.
#
# Usage:  ./294_TRIGGER_improvement_loop_v1.sh <site_id> [domain]
#         FORCE=1 ./294_… to override the pre-flight refusals.
set -euo pipefail

SITE_ID="${1:?usage: $0 <site_id> [domain]}"
DOMAIN="${2:-}"
NS=ai-persona-system
CLIENT_ID='demo_client'

command -v jq >/dev/null || { echo "ERROR: jq is required" >&2; exit 1; }

psql_q() {
  kubectl -n "$NS" exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -c "$1" 2>/dev/null
}

# --- pre-flight 1: chassis pod age -------------------------------------------
YOUNGEST=$(kubectl get pods -n "$NS" -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.status.startTime}{"\n"}{end}' | sort | tail -1)
if [[ -n "$YOUNGEST" ]]; then
  AGE=$(( $(date -u +%s) - $(date -u -d "$YOUNGEST" +%s) ))
  if (( AGE < 300 )) && [[ "${FORCE:-0}" != "1" ]]; then
    echo "REFUSING: youngest agent-chassis pod is ${AGE}s old (<300s)." >&2
    echo "  A dispatch now is silently DROPPED. Wait $((300-AGE))s, or FORCE=1." >&2
    exit 1
  fi
  echo "pre-flight: youngest chassis pod ${AGE}s old — OK"
fi

# --- pre-flight 2: is anyone already working this site? ----------------------
INFLIGHT=$(psql_q "SELECT count(*) FROM site_work_items
  WHERE site_id='${SITE_ID}' AND status IN ('claimed','in_progress');")
if [[ "${INFLIGHT:-0}" != "0" ]] && [[ "${FORCE:-0}" != "1" ]]; then
  echo "REFUSING: ${INFLIGHT} work item(s) already claimed/in_progress on this site." >&2
  echo "  Another session may have a fix in flight. Read them, then FORCE=1." >&2
  exit 1
fi
echo "pre-flight: ${INFLIGHT:-0} claimed/in-flight item(s) — OK"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="improvement-loop-$(date +%H%M%S)"

# ONE LINE, non-negotiable: kcat -P splits stdin on newlines into separate
# messages, so a pretty-printed payload becomes N broken messages at exit 0.
PAYLOAD=$(jq -cn --arg sid "$SITE_ID" --arg dom "$DOMAIN" \
  '{action:"orchestrate",
    config:{agent_type:"improvement-loop"},
    input_data: ({site_id:$sid} + (if $dom == "" then {} else {domain:$dom} end))}')

echo "========================================="
echo "improvement-loop manual trigger"
echo "  site_id:        ${SITE_ID}"
echo "  domain:         ${DOMAIN:-(not supplied — optional in input_contract)}"
echo "  correlation:    ${CORRELATION_ID}"
echo "  orchestration:  ${ORCHESTRATION_ID}"
echo "========================================="

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-improve-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.scheduled.requests \
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
  -H "responses_topic=system.agent.generic.responses"

cat <<EOF

Submitted. SAVE: ORCH_ID=${ORCHESTRATION_ID}

Watch the run (a missing row for the first few minutes is LATENCY, not a drop —
do not re-fire on that evidence, it costs a duplicate pass):

  SELECT current_step, status, updated_at FROM orchestration_states
  WHERE orchestration_id = '${ORCHESTRATION_ID}';

Did triage actually promote anything?

  SELECT status, count(*) FROM site_work_items
  WHERE site_id = '${SITE_ID}' GROUP BY 1 ORDER BY 2 DESC;
EOF
