#!/usr/bin/env bash
# 260_TRIGGER_experience_approval_v1.sh — put ONE register entry through the
# per-experience approval council (migration 259).
#
# WHAT IT SUBMITS
#   The entry AS STORED in experience_patterns — not the harvest file. That
#   distinction is the whole point: a verdict has to attach to what is actually
#   in the register, or approval describes a document nobody can select.
#
# WHAT COMES BACK
#   A verdict recorded as a council_report in diagnosis_artifacts, keyed on the
#   correlation printed below. It does NOT change the entry: nothing in this
#   lane can write status='approved'. Applying a verdict is a separate action
#   that does not exist yet, deliberately — this council can be run and read
#   before it is given the power to change anything.
#
# Usage:
#   ./260_TRIGGER_experience_approval_v1.sh <pattern-name>
#   ./260_TRIGGER_experience_approval_v1.sh feed-driven-teaser-list
set -euo pipefail

PATTERN="${1:?usage: $0 <pattern-name>}"
CLIENT_ID='demo_client'
TARGET_AGENT_TYPE='experience-approval-council'
PSQL=(kubectl exec -i -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -tA)

command -v jq >/dev/null || { echo "ERROR: jq is required" >&2; exit 1; }

# ── the entry, as STORED ─────────────────────────────────────────────────────
ENTRY=$("${PSQL[@]}" -c "SELECT to_jsonb(p) - 'id' - 'created_at' - 'updated_at' FROM experience_patterns p WHERE name = '${PATTERN//\'/\'\'}';" | tr -d '\n')
if [ -z "$ENTRY" ] || [ "$ENTRY" = "" ]; then
  echo "ERROR: no register entry named '$PATTERN'." >&2
  echo "Entries currently in the register:" >&2
  "${PSQL[@]}" -c "SELECT '  ' || name || '  [' || status || ']' FROM experience_patterns ORDER BY name;" >&2
  exit 1
fi

STATUS=$(printf '%s' "$ENTRY" | jq -r '.status')
if [ "$STATUS" != "draft" ]; then
  echo "NOTE: '$PATTERN' is already '$STATUS', not draft. Continuing — a re-review is legitimate," >&2
  echo "      e.g. after a contract change demoted it, but the verdict will not change the row." >&2
fi

# ── the register's current contents, for the prior-art and deferral seats ────
# Names, kinds, statuses and check counts only: enough to spot duplication and
# to judge one entry's deferral ratio against its siblings, without pasting nine
# full entries into every prompt.
SUMMARY=$("${PSQL[@]}" -c "
  SELECT string_agg(line, E'\n' ORDER BY line) FROM (
    SELECT '- ' || name || ' (' || kind || ', ' || status || '): '
           || executable_checks || ' executable, '
           || jsonb_array_length(deferred_checks) || ' deferred'
           || CASE WHEN jsonb_array_length(requires_invariant) > 0
                   THEN ', requires ' || (SELECT string_agg(v::text, '+') FROM jsonb_array_elements_text(requires_invariant) v)
                   ELSE '' END AS line
    FROM experience_patterns) t;")
INVARIANTS=$("${PSQL[@]}" -c "SELECT string_agg('- ' || name || ': ' || clause, E'\n') FROM experience_invariants;")
REGISTER_SUMMARY=$(printf 'ENTRIES\n%s\n\nSHARED INVARIANTS (reference these by name; do not restate them)\n%s\n' "$SUMMARY" "$INVARIANTS")

FIX_CORR=$(cat /proc/sys/kernel/random/uuid)
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="exp-approve-${PATTERN:0:26}"

# ONE LINE: kcat -P splits stdin on newlines into separate messages.
PAYLOAD=$(jq -cn --arg agent "$TARGET_AGENT_TYPE" --arg corr "$FIX_CORR" \
  --arg pattern "$PATTERN" --arg entry "$ENTRY" --arg reg "$REGISTER_SUMMARY" \
  '{action: "orchestrate",
    config: {agent_type: $agent},
    input_data: {
      fix_correlation_id: $corr,
      pattern_name: $pattern,
      entry_json: $entry,
      register_summary: $reg
    }}')

echo "========================================="
echo "Experience approval council"
echo "========================================="
echo "  entry:        $PATTERN  (currently: $STATUS)"
echo "  correlation:  $FIX_CORR"
echo "  payload:      $(printf '%s' "$PAYLOAD" | wc -c) bytes"
echo "  seats:        observable_outcome, honesty (VETO), checkability, deferral_honesty, prior_art"
echo "========================================="
echo "SAVE: APPROVAL_CORR=$FIX_CORR"
echo ""

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-expappr-$(date +%s%N | tail -c 8)" \
  --image=edenhill/kcat:1.7.1 --restart=Never --command -- \
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
  -H "responses_topic=system.agent.generic.responses" && echo "PUBLISH_OK"

cat <<EOS

Read the verdict (there are THREE outcomes — the third has no verdict in it):

  SELECT current_step, status, error FROM orchestration_states
  WHERE collected_data->'input_data'->>'fix_correlation_id' = '$FIX_CORR';
  -- error LIKE 'reaper: stale%' means the run DIED; that is not a REVISE.

  SELECT metadata::text FROM diagnosis_artifacts
  WHERE correlation_id='$FIX_CORR' AND kind='council_report';
  -- check unreadable:0 FIRST — a filtered/failed seat counts as abstained

  SELECT body FROM diagnosis_artifacts
  WHERE correlation_id='$FIX_CORR' AND kind='council_report';
EOS
