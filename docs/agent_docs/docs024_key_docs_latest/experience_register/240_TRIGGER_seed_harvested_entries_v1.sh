#!/usr/bin/env bash
# 240_TRIGGER_seed_harvested_entries_v1.sh — put the harvested entries into the
# experience register, through the validating write path.
#
# WHAT THIS DOES
#   For each harvest/entries/*.json, dispatches one experience-register-writer
#   run (migration 238). That workflow validates the entry, writes it as DRAFT,
#   and writes its travelling doc in the same run — an entry cannot land without
#   provenance because there is no path that produces one without the other.
#
# WHAT IT DELIBERATELY DOES NOT DO
#   It does not INSERT. Migration 218 seeded no entries by raw INSERT precisely
#   so that the register's first rows would be ones its own contract accepted;
#   seeding them here by SQL would throw away the only evidence that the
#   contract is enforceable.
#
# THE FILES ARE DOCUMENTS, NOT PAYLOADS
#   A harvested entry file carries commentary keys (`_status`, `_harvested`,
#   `_SCHEMA_DELTA`) and a `status` of its own recording that it is a harvest
#   draft. Both are stripped here: the action REFUSES a supplied status by
#   design, because 'approved' is a council verdict and 'proven' is a live green
#   run — things that happened, not fields a writer fills in.
#
# PRECONDITION — THE ROLL. Check it, do not assume it:
#   kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
#     'strings /app/agent-chassis | grep -c "contract must be an array of clauses"'
#   must be >= 1, AND the old shape must be GONE:
#     ... | grep -c "contract.triggers must be a non-empty array"   -> 0
#   The second is the load-bearing one. v1.0.1175 carries the INVENTED contract
#   shape and would refuse all nine entries; a positive-only grep cannot tell
#   the two builds apart if both strings were ever present.
#
# Usage:
#   ./240_TRIGGER_seed_harvested_entries_v1.sh            # all entries
#   ./240_TRIGGER_seed_harvested_entries_v1.sh CC-001     # one, by filename prefix
#   DRY_RUN=1 ./240_TRIGGER_seed_harvested_entries_v1.sh  # build payloads, publish nothing
set -euo pipefail

HERE="$(cd "$(dirname "$0")" && pwd)"
ENTRIES_DIR="$HERE/harvest/entries"
FILTER="${1:-}"
CLIENT_ID='demo_client'
TARGET_AGENT_TYPE='experience-register-writer'

command -v jq >/dev/null || { echo "ERROR: jq is required" >&2; exit 1; }
[ -d "$ENTRIES_DIR" ] || { echo "ERROR: no entries dir at $ENTRIES_DIR" >&2; exit 1; }

# ── precondition: the deployed binary must read the harvested shape ──────────
if [ "${SKIP_PRECHECK:-0}" != "1" ]; then
  POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis \
        -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
  if [ -z "$POD" ]; then
    echo "ERROR: no agent-chassis pod found (SKIP_PRECHECK=1 to override)" >&2; exit 1
  fi
  # grep -c exits 1 on zero matches while still printing "0"; swallow it inside the
  # pod and strip to digits, or the comparison below becomes a syntax error.
  NEW=$(kubectl exec -n ai-persona-system "$POD" -- sh -c \
        'strings /app/agent-chassis | grep -c "contract must be an array of clauses" || true' 2>/dev/null | tr -dc '0-9')
  OLD=$(kubectl exec -n ai-persona-system "$POD" -- sh -c \
        'strings /app/agent-chassis | grep -c "contract.triggers must be a non-empty array" || true' 2>/dev/null | tr -dc '0-9')
  if [ "$NEW" -lt 1 ] || [ "$OLD" -ne 0 ]; then
    echo "REFUSED: the running chassis does not carry the corrected contract shape." >&2
    echo "  pod=$POD  new-shape hits=$NEW (want >=1)  old-shape hits=$OLD (want 0)" >&2
    echo "  It would refuse every harvested entry. Roll the image first (commit 799c0c97e)." >&2
    exit 2
  fi
  echo "precheck ok: $POD carries the corrected shape (new=$NEW, old=$OLD)"

  # A chassis that restarted moments ago silently drops the spawn (~300s).
  STARTED=$(kubectl get pod -n ai-persona-system "$POD" -o jsonpath='{.status.startTime}')
  AGE=$(( $(date -u +%s) - $(date -u -d "$STARTED" +%s) ))
  if [ "$AGE" -lt 300 ]; then
    echo "REFUSED: chassis pod is ${AGE}s old; dispatches within ~300s of a restart are silently dropped." >&2
    echo "  Wait $((300 - AGE))s and re-run." >&2
    exit 2
  fi
fi

shopt -s nullglob
FILES=("$ENTRIES_DIR"/*.json)
[ ${#FILES[@]} -gt 0 ] || { echo "ERROR: no entry files" >&2; exit 1; }

SENT=0
for f in "${FILES[@]}"; do
  base="$(basename "$f")"
  if [ -n "$FILTER" ] && [[ "$base" != "$FILTER"* ]]; then continue; fi

  NAME=$(jq -r '.name // empty' "$f")
  [ -n "$NAME" ] || { echo "SKIP $base: no .name" >&2; continue; }

  # Strip the document's own metadata: underscore keys are commentary, and
  # `status` is refused by the action on purpose.
  ENTRY=$(jq -c 'with_entries(select(.key | startswith("_") | not)) | del(.status)' "$f")

  # The travelling doc. Provenance in prose, keyed by the pattern name — the
  # same machinery a tool's doc uses.
  DOC=$(jq -r --arg base "$base" '
    "# " + (.display_name // .name) + "\n\n" +
    "**Register entry:** `" + .name + "` (" + .kind + ")\n\n" +
    "**Harvested from:** " + ((._harvested // .harvested_from // "see the harvest report")|tostring) + "\n\n" +
    "**Source file:** `docs/agent_docs/docs024_key_docs_latest/experience_register/harvest/entries/" + $base + "`\n\n" +
    "## What it is\n\n" + ((.description // "—")|tostring) + "\n\n" +
    "## Why it is in the register\n\n" +
    "Written by the harvest of 2026-07-26 from a LIVE implementation, not from a design. Every clause below was read out of shipped code or a served page; see HARVEST_01 and HARVEST_02 in this directory for the evidence and for the ten (then six more) corrections the harvest forced on the 2026-07-24 design.\n\n" +
    "## Status\n\n" +
    "Written as **draft** by `experience-register-writer`. It is UNSELECTABLE until approved, and approval is a council verdict — not a field any writer can set. `proven` comes later still, from the first live green run of its bound criteria.\n"
  ' "$f")

  PAYLOAD=$(jq -cn --arg agent "$TARGET_AGENT_TYPE" --argjson entry "$ENTRY" --arg doc "$DOC" \
    '{action: "orchestrate",
      config: {agent_type: $agent},
      input_data: {experience_pattern: $entry, doc_plan_body: $doc}}')

  CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
  ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
  REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
  MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
  ORCH_NAME="exp-register-${NAME:0:28}"

  if [ "${DRY_RUN:-0}" = "1" ]; then
    echo "DRY_RUN $base -> $NAME ($(echo -n "$PAYLOAD" | wc -c) bytes)"
    SENT=$((SENT+1))
    continue
  fi

  # ONE LINE, non-negotiable: kcat -P splits stdin on newlines into separate
  # messages. jq -c keeps the payload a single line whatever the file's
  # formatting was.
  printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-expreg-$(date +%s%N | tail -c 8)" \
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
    -H "responses_topic=system.agent.generic.responses" && echo "PUBLISH_OK $NAME  orch=$ORCHESTRATION_ID"

  SENT=$((SENT+1))
  sleep 2   # one lane, one consumer — do not stampede it
done

echo
echo "dispatched: $SENT"
cat <<'EOS'

Verify (a row is not proof the ENTRY is right — read the counts and the docs):

  SELECT name, kind, status, executable_checks,
         jsonb_array_length(deferred_checks) AS deferred
  FROM experience_patterns ORDER BY name;

  -- every entry must have its travelling doc; a mismatch means the workflow
  -- completed the write and lost the doc, which is what mig 238 exists to stop
  SELECT (SELECT count(*) FROM experience_patterns) AS entries,
         (SELECT count(*) FROM doc_plans
           WHERE subject_type='experience-pattern' AND is_current) AS docs;

  -- refusals surface as FAILED orchestrations, not as missing rows
  SELECT current_step, status, left(error, 300)
  FROM orchestration_states
  WHERE collected_data->'input_data'->'experience_pattern' IS NOT NULL
  ORDER BY created_at DESC LIMIT 12;
EOS
