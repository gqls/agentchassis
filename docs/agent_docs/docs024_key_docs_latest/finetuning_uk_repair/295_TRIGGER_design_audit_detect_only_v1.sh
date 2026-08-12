#!/usr/bin/env bash
# 295_TRIGGER_design_audit_detect_only_v1.sh — run the DESIGN AUDIT at one site,
# and NOTHING ELSE. Sibling of 294 (the full improvement-loop trigger).
#
# WHY THIS EXISTS, and why it is not just "294 again".
# 294 fires the whole improvement-loop: discovery + audits -> triage_findings ->
# call_dispatch. `triage_findings` is the ONLY promoter of detected -> triaged in
# the platform, and `call_dispatch` then claims triaged items and runs their
# handlers — including the content-REGENERATING ones (`tone_shift`,
# `content_rewrite`, `cta_improvement`). On 2026-08-09 exactly that sequence,
# fired to serve five images, regenerated finetuning.uk's `case-studies-grid`,
# dropped every `card*_image_url` and shipped five empty <img src=""> to the live
# homepage (`bugs_open/238`). The audit was wanted; the rewrite was not.
#
# This trigger runs `design-audit-agent` DIRECTLY. Its live workflow is six steps:
#   ensure_site_record -> spawn_visual_auditor -> call_visual_auditor
#                      -> spawn_content_auditor -> call_content_auditor -> complete
# There is no triage step in it, nor in either child (verified against
# `agent_definitions`, not a seed, 2026-08-12). Findings are written by
# `write_audit_findings_action.go:677`, which hardcodes status **'detected'**, and
# the dispatcher's claim query is `status IN ('triaged','approved')`. So the
# findings CANNOT be claimed until a human promotes them. That is the whole point:
# you get the report, you choose the repairs.
#
# ⚠ TRAP — `scripts/initial_messages/290_design_discovery/081b_design_audit_agent_robot_hands.sh`
# is the same envelope but its COMMENTS ARE STALE. It states the agent runs
# `triage_detected_items` and that "dispatch picks them up automatically on the
# next 30s tick". That was true of an older definition; the live one has no triage
# step. If you copy that script, copy the envelope and not the expectations —
# and re-read the live workflow before believing either file, including this one.
#
# WHAT YOU GET: `visual-design-auditor` + `content-quality-auditor`, both LLM
# passes, writing `detected` items with `spec->>'audit_source'`. Read them, then
# promote the ones you want with an explicit UPDATE to 'triaged'.
#
# Usage:  ./295_TRIGGER_design_audit_detect_only_v1.sh <site_id> [domain]
#         FORCE=1 to override the pre-flight refusals.
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

# --- pre-flight 1: chassis pod age (CLAUDE.md: <300s dispatches are DROPPED) ---
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

# --- baseline: identity, not just a count ------------------------------------
# A post-run count alone cannot tell "the audit ran and found nothing" from "the
# audit never ran". Record the newest existing finding so the delta is provable.
BEFORE=$(psql_q "SELECT count(*) FROM site_work_items
  WHERE site_id='${SITE_ID}' AND status='detected';")
NEWEST=$(psql_q "SELECT COALESCE(max(created_at)::text,'(none)') FROM site_work_items
  WHERE site_id='${SITE_ID}' AND item_key LIKE 'design-audit%';")
echo "baseline: ${BEFORE:-0} detected item(s); newest design-audit finding: ${NEWEST}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

# ONE LINE, non-negotiable: kcat -P splits stdin on newlines into separate
# messages, so a pretty-printed payload becomes N broken messages at exit 0.
PAYLOAD=$(jq -cn --arg sid "$SITE_ID" --arg dom "$DOMAIN" \
  '{action:"orchestrate",
    config:{agent_type:"design-audit-agent"},
    input_data: ({site_id:$sid} + (if $dom == "" then {} else {domain:$dom} end))}')

echo "========================================="
echo "design-audit-agent — DETECT ONLY"
echo "  site_id:        ${SITE_ID}"
echo "  domain:         ${DOMAIN:-(not supplied)}"
echo "  correlation:    ${CORRELATION_ID}"
echo "  orchestration:  ${ORCHESTRATION_ID}"
echo "========================================="

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-audit-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "sender_agent_type=cli" \
  -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" \
  -H "timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

cat <<EOF

Submitted. SAVE: ORCH_ID=${ORCHESTRATION_ID}

A missing orchestration row for the first minutes is LATENCY, not a drop — do not
re-fire on that evidence.

  SELECT current_step, status, updated_at FROM orchestration_states
  WHERE orchestration_id = '${ORCHESTRATION_ID}';

The findings (they stay 'detected' — nothing will claim them):

  SELECT item_type, severity, priority, left(summary,90) AS summary
  FROM site_work_items
  WHERE site_id = '${SITE_ID}' AND status = 'detected'
    AND created_at > now() - interval '30 minutes'
  ORDER BY priority, item_type;

Baseline was ${BEFORE:-0} detected / newest design-audit finding ${NEWEST}, so a
non-zero delta here is the audit's own output and not history.

To act on one, promote it EXPLICITLY (this is the step that arms dispatch):

  UPDATE site_work_items SET status='triaged', priority=1
  WHERE id = '<the id you chose>';
EOF
