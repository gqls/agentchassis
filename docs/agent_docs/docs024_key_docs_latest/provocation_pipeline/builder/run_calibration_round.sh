#!/usr/bin/env bash
# Run ONE provocation-gate calibration round against calibration.vonc.com.
#
# Why this exists: §4a of HANDOFF_2026-08-08 — the judge is STOCHASTIC, so a
# single green calibration is not evidence. Declaring the gate calibrated needs
# at least three consecutive passing rounds, which means running the same round
# three times without re-typing the envelope and without forgetting the reset.
#
# The gate NEVER re-judges a row that already has gated_at set (on purpose), so
# a round that is not reset first will report the PREVIOUS round's scorecard and
# look like a successful run. That is the trap this script closes.
#
# Usage:  ./run_calibration_round.sh            # reset + dispatch
#         DRY=1 ./run_calibration_round.sh      # show what it would do
#
# It does NOT wait for the verdicts — score with the queries the script prints
# (or ./score_calibration_round.sh).
set -euo pipefail

NS=ai-persona-system
PG=postgres-clients-0
DOMAIN=calibration.vonc.com
AGENT=provocation-gate-calibration
TOPIC=system.agent.generic.requests
BOOTSTRAP=personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092

psql() { kubectl -n "$NS" exec -i "$PG" -- psql -U clients_user -d clients_db "$@"; }

# ── guard 1: never dispatch within ~300s of a chassis (re)start — the spawn is
# silently dropped, which is indistinguishable from a gate that judged nothing.
youngest=$(kubectl get pods -n "$NS" -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.status.startTime}{"\n"}{end}' | sort | tail -1)
age=$(( $(date -u +%s) - $(date -u -d "$youngest" +%s) ))
if [ "$age" -lt 300 ]; then
  echo "REFUSED: youngest chassis pod is ${age}s old (<300s) — the dispatch would be silently dropped." >&2
  exit 1
fi
echo "chassis age OK: ${age}s since youngest pod start ($youngest)"

# ── guard 2: the harness must be incapable of touching production. Migration
# 319 asserts this, but assert it again at dispatch time: the domain must NOT
# be a site, or render_provocation_feed's assertKnownDomain would accept it.
if [ "$(psql -tAc "SELECT count(*) FROM sites WHERE domain='$DOMAIN';")" != "0" ]; then
  echo "REFUSED: '$DOMAIN' now exists in sites — the harness could reach a real feed. Stop." >&2
  exit 1
fi
echo "isolation OK: '$DOMAIN' is absent from sites"

if [ "${DRY:-0}" = "1" ]; then
  echo "--- DRY: current scorecard (would be reset) ---"
  psql -c "SELECT status, count(*) FROM provocations WHERE domain='$DOMAIN' GROUP BY 1 ORDER BY 1;"
  exit 0
fi

# ── reset: un-judge every calibration row so the gate actually re-judges.
echo "--- reset ---"
psql -c "UPDATE provocations SET status='draft', gated_at=NULL, gate_verdict=NULL
          WHERE domain='$DOMAIN';"
psql -c "SELECT count(*) AS ungated_drafts FROM provocations
          WHERE domain='$DOMAIN' AND status='draft' AND gated_at IS NULL;"

# ── envelope (copied from 097_TRIGGER_council_review_v1.sh lines 168-188)
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID=00000000-0000-0000-0000-000000000000
ORCH_NAME="provocation-calibration-$(date +%H%M%S)"

# ONE LINE, non-negotiable: kcat -P splits stdin on newlines into separate messages.
PAYLOAD=$(jq -cn --arg agent "$AGENT" \
  '{action:"orchestrate", config:{agent_type:$agent}, input_data:{}}')

echo "========================================="
echo "  Orchestration:      $ORCHESTRATION_ID"
echo "  Orchestration name: $ORCH_NAME"
echo "  Correlation:        $CORRELATION_ID"
echo "========================================="
echo "SAVE: RUN_ORCH_ID=$ORCHESTRATION_ID"

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-prov-$(date +%s)" \
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
echo "Score it:"
echo "  ./score_calibration_round.sh"
