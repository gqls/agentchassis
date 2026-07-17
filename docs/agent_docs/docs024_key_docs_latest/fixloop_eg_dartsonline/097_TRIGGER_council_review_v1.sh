#!/usr/bin/env bash
# 097_TRIGGER_council_review_v1.sh — submit a change to the COUNCIL GATE.
# Envelope mirrors 091_TRIGGER_fix_proposer_v1.sh (kcat pod ->
# system.agent.generic.requests, action=orchestrate), target agent_type
# council-gate (seed: 0NN_council_gate.sql — must be applied first; owner
# decision 2026-07-17: apply waits for more concept-register stage-3 seats).
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
TARGET_AGENT_TYPE='council-gate'
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
  -H "responses_topic=system.agent.generic.responses"

echo ""
echo "Submitted. Watch the run:"
echo "  SELECT new_current_step, new_status, changed_at FROM orchestration_state_audit"
echo "  WHERE orchestration_id='${ORCHESTRATION_ID}' ORDER BY changed_at;"
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
