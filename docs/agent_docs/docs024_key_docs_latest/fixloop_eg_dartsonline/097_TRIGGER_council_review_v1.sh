#!/usr/bin/env bash
# 097_TRIGGER_council_review_v1.sh — submit a change to the COUNCIL GATE.
# Envelope mirrors 091_TRIGGER_fix_proposer_v1.sh (kcat pod ->
# system.agent.council-gate.requests, action=orchestrate), target agent_type
# council-gate (seed: 0NN_council_gate.sql — must be applied first; owner
# decision 2026-07-17: apply waits for more concept-register stage-3 seats).
#
# WHY THIS TOPIC IS NOT system.agent.generic.requests (changed 2026-07-28,
# bugs_open/096). A council step is a single LLM call that runs for minutes, and
# the generic topic was a single lane every other dispatch shared, so one council
# head-of-line blocked the fleet. Council traffic now has its OWN topic, consumed
# by its own group in the same chassis (EXTRA_REQUEST_TOPICS). Responses still go
# to system.agent.generic.responses on purpose — the response topic was never the
# congested one, and moving it would strand every reply.
#
# IF YOU EVER CHANGE THIS TOPIC, THE ORDER IS LOAD-BEARING (agentbase/agent.go
# :429-432): ship a chassis that CONSUMES the new topic first, confirm the
# consumer group exists, and only then point this producer at it. A producer
# aimed at an unconsumed topic piles messages up where nothing will ever run
# them, and they look exactly like queue latency.
#
# WHAT A SUBMISSION IS. One JSON file:
#   {
#     "rationale":  "why this change: the problem, the evidence, the intent",
#     "submitter":  "thread/session name (optional but be a good citizen)",
#     "plan": {
#       "summary":     "one paragraph: what is broken and what this changes",
#       "edits": [
#         {"file": "platform/x/y.go", "symbol": "FuncOrStep",
#          "operation": "modify|add|remove|config_change",
#          "rationale": "why THIS edit",
#          "sketch": "the change, precisely — real diff hunks welcome here"}
#       ],
#       "grounded_in": ["verbatim evidence quotes this change rests on"],
#       "risks": "what could this break; what a reviewer should check"
#     }
#   }
# The plan block is exactly the fix_plan artifact schema (validated server-side
# by diagnose_persist_fix_plan: <=8 edits, <=64KB, no no-op edits, repo-relative
# paths). The rationale is what reviewers judge the plan AGAINST — an empty or
# lazy rationale gets you objections, and the seed cannot catch it (missingkey
# =zero renders a missing field as blank, silently), so this script refuses it.
#
# SCOPE (owner ruling 2026-07-17): platform/, internal/, pkg/. A submission
# touching none of those is refused (docs/site content never spends council
# credits) — FORCE=1 overrides if you know better.
#
# RESUBMISSION: pass the previous correlation as RESUBMIT_CORR (or arg 2) so
# the artifact trail (fix_plan rounds, council_reports, verdict notes)
# accumulates in one place. Round counting is orchestration-scoped, so a
# resubmission is judged fresh — the correlation is for the TRAIL, not the cap.
#
# APPROVED? Commit with the trailer line (this is what makes the 098 report's
# commit<->verdict join exact):
#   Council-Reviewed: <the fix_correlation_id printed below>
#
# Usage:
#   ./097_TRIGGER_council_review_v1.sh submission.json
#   RESUBMIT_CORR=<uuid> ./097_TRIGGER_council_review_v1.sh submission.json
#   ./097_TRIGGER_council_review_v1.sh submission.json <uuid>
set -euo pipefail

SUBMISSION_FILE="${1:?usage: $0 <submission.json> [resubmit_correlation_id]}"
FIX_CORR="${2:-${RESUBMIT_CORR:-}}"
# bugs_open/096: council-gate runs its 16 seats INLINE on the shared chassis
# request lane, so one submission blocks every other dispatch for 4-9 minutes and
# two councils cannot overlap. `council-gate-orchestrator` is a thin wrapper that
# spawns the council into its own pod (0NN_council_gate_orchestrator.sql) and
# releases the lane in ~8s. Overridable so the wrapper can be exercised without
# changing the default for every other session mid-rollout:
#   TARGET_AGENT_TYPE=council-gate-orchestrator ./097_TRIGGER_council_review_v1.sh sub.json
# The verdict, artifacts and correlation trail are identical either way — the
# wrapper forwards the submission verbatim and the CHILD writes the artifacts.
TARGET_AGENT_TYPE="${TARGET_AGENT_TYPE:-council-gate}"
CLIENT_ID='demo_client'
SCOPE_RE='^(platform|internal|pkg)/'

[ -f "$SUBMISSION_FILE" ] || { echo "ERROR: no such file: $SUBMISSION_FILE" >&2; exit 1; }
command -v jq >/dev/null || { echo "ERROR: jq is required" >&2; exit 1; }

# ── client-side validation: fail HERE, before any credits are spent ──────────
jq -e 'type == "object"' "$SUBMISSION_FILE" >/dev/null \
  || { echo "ERROR: submission is not a JSON object" >&2; exit 1; }
jq -e '.rationale | type == "string" and (gsub("\\s";"") | length) > 40' "$SUBMISSION_FILE" >/dev/null \
  || { echo "ERROR: .rationale missing or too thin (<40 non-space chars) — the reviewers judge the plan AGAINST the rationale; write the real why" >&2; exit 1; }
jq -e '.plan | type == "object"' "$SUBMISSION_FILE" >/dev/null \
  || { echo "ERROR: .plan missing — see the header for the fix_plan schema" >&2; exit 1; }
jq -e '.plan.summary | type == "string" and length > 0' "$SUBMISSION_FILE" >/dev/null \
  || { echo "ERROR: .plan.summary is empty" >&2; exit 1; }
jq -e '.plan.edits | type == "array" and length > 0 and length <= 8' "$SUBMISSION_FILE" >/dev/null \
  || { echo "ERROR: .plan.edits must be a non-empty array of <=8 edits (wider than 8 files is architecture-shaped — take it to a human, not this gate)" >&2; exit 1; }
jq -e '[.plan.edits[] | select((.file//"")=="" or (.operation//"")=="" or (.rationale//"")=="" or (.sketch//"")=="")] | length == 0' "$SUBMISSION_FILE" >/dev/null \
  || { echo "ERROR: every edit needs file, operation, rationale, sketch" >&2; exit 1; }
jq -e '.plan.grounded_in | type == "array" and length > 0' "$SUBMISSION_FILE" >/dev/null \
  || { echo "ERROR: .plan.grounded_in is empty — quote the evidence this change rests on" >&2; exit 1; }

# Scope pre-filter (the cheap deterministic filter from the design — keeps
# docs/site-content submissions from spending council credits at all).
if ! jq -e --arg re "$SCOPE_RE" '[.plan.edits[].file | select(test($re))] | length > 0' "$SUBMISSION_FILE" >/dev/null; then
  if [ "${FORCE:-0}" != "1" ]; then
    echo "REFUSED: no edit touches the review scope (platform/, internal/, pkg/ — owner ruling 2026-07-17)." >&2
    echo "Docs and site content do not spend council credits. FORCE=1 to override." >&2
    exit 2
  fi
  echo "WARN: FORCE=1 — submitting despite no in-scope files" >&2
fi

BYTES=$(jq -c '.plan' "$SUBMISSION_FILE" | wc -c)
[ "$BYTES" -le 65536 ] || { echo "ERROR: plan is ${BYTES} bytes (cap 65536) — trim the diff sketches" >&2; exit 1; }

# ── envelope ─────────────────────────────────────────────────────────────────
[ -n "$FIX_CORR" ] || FIX_CORR=$(cat /proc/sys/kernel/random/uuid)
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="council-gate-$(date +%H%M%S)"

# ONE LINE, non-negotiable: kcat -P splits stdin on newlines into separate
# messages (this bit the travelling-docs thread) — jq -c keeps the whole
# payload a single line whatever the submission file's formatting was.
PAYLOAD=$(jq -c --arg corr "$FIX_CORR" --arg agent "$TARGET_AGENT_TYPE" \
  '{action: "orchestrate",
    config: {agent_type: $agent},
    input_data: {
      fix_correlation_id: $corr,
      rationale: .rationale,
      submitter: (.submitter // "unnamed"),
      plan: .plan
    }}' "$SUBMISSION_FILE")

echo "========================================="
echo "Council-gate submission (advisory review)"
echo "========================================="
echo "  Submission correlation (the TRAIL id):  ${FIX_CORR}"
echo "  Run correlation (envelope):             ${CORRELATION_ID}"
echo "  Orchestration:                          ${ORCHESTRATION_ID}"
echo "  Orchestration name:                     ${ORCH_NAME}"
echo "  Plan bytes:                             ${BYTES}"
echo "========================================="
echo "SAVE: SUBMISSION_CORR=${FIX_CORR}  RUN_ORCH_ID=${ORCHESTRATION_ID}"
echo ""

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-cgate-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.council-gate.requests \
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

echo ""
echo "Submitted. Watch the run:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"

# bugs_open/030: that query returns 0 rows until the dispatch reaches the head of
# the single generic requests lane, which reads as a drop and has twice cost a
# duplicate paid round. Print the lane depth so nobody re-fires. Advisory and
# fail-soft — never let it break a submission (hence `|| true` under set -e).
QUEUE_REPORT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && git rev-parse --show-toplevel 2>/dev/null || true)/scripts/dispatch-queue-depth.sh"
if [[ -x "$QUEUE_REPORT" ]]; then
  echo ""
  "$QUEUE_REPORT" || true
fi
echo ""
echo "The verdict (keyed on the submission correlation):"
echo "  SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts"
echo "  WHERE correlation_id='${FIX_CORR}' AND kind='council_report' ORDER BY created_at;"
echo ""
echo "Human-readable verdict note:"
echo "  SELECT body FROM doc_notes WHERE categories ? 'council-gate'"
echo "  ORDER BY created_at DESC LIMIT 1;"
echo ""
echo "If APPROVED, commit with:   Council-Reviewed: ${FIX_CORR}"
