#!/usr/bin/env bash
# Run ONE provocation generation round against vonc.com: generate drafts with a
# real model, then judge them with the real gate, in one dispatch.
#
# Sibling of run_calibration_round.sh. The difference that matters: the
# calibration harness is isolated by DATA (its domain is absent from `sites`, so
# nothing it produces can ever be served), and this one is NOT — it writes into
# the live pool. Its containment is different and worth stating, because a reader
# who knows the calibration script will assume the same protection:
#
#   generate -> status='draft', publish_on NULL, human_approved_at NULL
#   gate     -> may set status='approved'; writes neither the date nor the stamp
#
# The publisher requires approved AND publish_on AND human_approved_at (mig 320),
# so nothing this script produces can reach the site until a person stamps it and
# the operator scheduler (mig 321) dates it. That is the whole of the protection.
# If you are about to add a step here that writes either field, don't.
#
# Usage:  ./run_generation_round.sh          # dispatch
#         DRY=1 ./run_generation_round.sh    # show state, dispatch nothing
set -euo pipefail

NS=ai-persona-system
PG=postgres-clients-0
DOMAIN=vonc.com
AGENT=provocation-generator-manual
TOPIC=system.agent.generic.requests
BOOTSTRAP=personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092

psql() { kubectl -n "$NS" exec -i "$PG" -- psql -U clients_user -d clients_db "$@"; }

# ── guard 1: never dispatch within ~300s of a chassis (re)start — the spawn is
# silently dropped, which is indistinguishable from a generator that wrote nothing.
youngest=$(kubectl get pods -n "$NS" -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.status.startTime}{"\n"}{end}' | sort | tail -1)
age=$(( $(date -u +%s) - $(date -u -d "$youngest" +%s) ))
if [ "$age" -lt 300 ]; then
  echo "REFUSED: youngest chassis pod is ${age}s old (<300s) — the dispatch would be silently dropped." >&2
  exit 1
fi
echo "chassis age OK: ${age}s since youngest pod start ($youngest)"

# ── guard 2: the seat must still be operator-invoked. Migration 371 asserts this
# at apply time; assert it again at dispatch time, because the failure it guards
# against is someone adding a schedule LATER, which no applied migration re-checks.
n_sched=$(psql -tAc "SELECT count(*) FROM scheduled_tasks WHERE target_agent_type='$AGENT';")
if [ "$n_sched" != "0" ]; then
  echo "REFUSED: '$AGENT' has $n_sched scheduled_tasks row(s). It is meant to be operator-invoked; something has automated it. Stop and read migration 371." >&2
  exit 1
fi
echo "containment OK: '$AGENT' has no schedule"

# ── guard 3: the gate step must judge with the model the calibration measured.
gate_model=$(psql -tAc "SELECT default_config->'workflow'->'steps'->'gate'->'config'->'ai_service'->>'model' FROM agent_definitions WHERE type='$AGENT' AND is_active AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;")
cal_model=$(psql -tAc "SELECT default_config->'workflow'->'steps'->'gate'->'config'->'ai_service'->>'model' FROM agent_definitions WHERE type='provocation-gate-calibration' AND is_active AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;")
if [ -n "$cal_model" ] && [ "$gate_model" != "$cal_model" ]; then
  echo "REFUSED: gate model '$gate_model' != calibrated model '$cal_model' — the calibration is not evidence about this run." >&2
  exit 1
fi
echo "gate model OK: $gate_model (matches the calibration)"

echo "--- pool before ---"
psql -c "SELECT status, count(*) FROM provocations WHERE domain='$DOMAIN' GROUP BY 1 ORDER BY 1;"
psql -c "SELECT max(publish_on) AS shelf_ends, max(publish_on) - CURRENT_DATE AS days_left
           FROM provocations WHERE domain='$DOMAIN' AND status='approved';"

# Take the cut from the DATABASE clock, not this shell's. A count cannot
# distinguish "the model wrote eight new ones" from "the model repeated eight it
# wrote last week and ON CONFLICT DO NOTHING silently dropped them all" — both
# leave the same total when a previous batch is still in the pool. So the read-out
# below asks for row IDENTITY created after this instant.
#
# Deliberately NOT a `comm` against a pre-run slug list: comm and psql do not
# agree on collation, so a diff of two sorted lists can silently report rows that
# are in both (LANDMINES: "ask git about git, not comm").
RUN_START=$(psql -tAc "SELECT now();" | head -1)   # -tAc is unaligned: no padding to trim
echo "cut taken from the database clock: $RUN_START"

if [ "${DRY:-0}" = "1" ]; then
  echo "--- DRY: would dispatch $AGENT; nothing sent ---"
  exit 0
fi

# ── envelope (copied from run_calibration_round.sh, which copied 097_TRIGGER)
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID=00000000-0000-0000-0000-000000000000
ORCH_NAME="provocation-generation-$(date +%H%M%S)"

# ONE LINE, non-negotiable: kcat -P splits stdin on newlines into separate messages.
PAYLOAD=$(jq -cn --arg agent "$AGENT" \
  '{action:"orchestrate", config:{agent_type:$agent}, input_data:{}}')

echo "========================================="
echo "  Orchestration:      $ORCHESTRATION_ID"
echo "  Orchestration name: $ORCH_NAME"
echo "  Correlation:        $CORRELATION_ID"
echo "  Cut (db clock):     $RUN_START"
echo "========================================="

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-gen-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b "$BOOTSTRAP" \
  -t "$TOPIC" \
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
echo "Dispatched. Watch it:"
echo "  SELECT current_step, status FROM orchestration_states WHERE id='$ORCHESTRATION_ID';"
echo ""
echo "Read what it actually produced (NEW rows only — identity, not count):"
echo "  SELECT slug, status, left(title,60) FROM provocations"
echo "   WHERE domain='$DOMAIN' AND created_at > '$RUN_START' ORDER BY created_at;"
echo ""
echo "Then the verdicts, including on rows this run did not create:"
echo "  SELECT slug, status, gate_verdict->'reasons' FROM provocations"
echo "   WHERE domain='$DOMAIN' AND gated_at > '$RUN_START' ORDER BY status, slug;"
echo ""
echo "Zero new rows with a completed run means every slug already existed"
echo "(ON CONFLICT DO NOTHING), NOT that the model failed. Check 'generated' in the"
echo "step result before concluding anything about the model."
