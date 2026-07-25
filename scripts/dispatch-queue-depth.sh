#!/usr/bin/env bash
# dispatch-queue-depth.sh — is my dispatch QUEUED, or was it LOST?
#
# bugs_open/030 exists because those two are indistinguishable from every
# surface an operator has: `kcat` exits 0, the trigger prints "Submitted", and
# then `orchestration_states` returns 0 rows for tens of minutes. That reads as
# failure. It has cost real work twice — a duplicate paid council round, and an
# abandoned investigation — so the trigger scripts call this after publishing.
#
# WHAT IT ANSWERS (and what it deliberately does not):
#   - LAG > 0            → your message is QUEUED. It is not lost. Do NOT re-fire.
#   - no group member    → a REAL fault (nothing is consuming the lane).
#   - what is at the head → whether an expensive orchestration is in front of you.
#   - it does NOT print an ETA. There is no drain rate to quote: throughput is
#     1/(duration of the segment executing inline at the head), which ranges from
#     milliseconds to >15 minutes. Three threads computed three "measured" rates
#     (0.21, 2.4, 0.62 msg/min) from this queue on one afternoon, all
#     arithmetically correct, all useless as forecasts. See bugs_open/030
#     "CORRECTION OF THIS CORRECTION" and WRONG_CALLS.md 7/8/9.
#
# ADVISORY BY CONSTRUCTION: every probe is timeout-wrapped and fail-soft, and
# the script exits 0 even when it can reach nothing. It is called from the
# publish path of 090/097 — it must never be able to fail a submission.
#
# Usage: scripts/dispatch-queue-depth.sh [--topic T] [--group G] [--brief]
#   --brief   omit the in-flight orchestration list (Kafka figures only)

TOPIC="system.agent.generic.requests"
GROUP="generic-requests-group"
KAFKA_NS="kafka"
KAFKA_POD="personae-kafka-cluster-combined-pool-prod-0"   # NOT ...-dual-role-0, which
                                                          # does not exist here and
                                                          # fails quietly (030 gotcha)
PG_NS="ai-persona-system"
PG_POD="postgres-clients-0"
BRIEF=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --topic) TOPIC="$2"; shift 2 ;;
    --group) GROUP="$2"; shift 2 ;;
    --brief) BRIEF=1; shift ;;
    -h|--help) sed -n '2,30p' "$0"; exit 0 ;;
    *) echo "dispatch-queue-depth: ignoring unknown argument '$1'" >&2; shift ;;
  esac
done

psql_q() {  # $1 = SQL; prints nothing on any failure
  timeout 25 kubectl -n "$PG_NS" exec -i "$PG_POD" -- \
    psql -U clients_user -d clients_db -X -q "$@" 2>/dev/null
}

echo "-----------------------------------------------------------------------"
echo "DISPATCH LANE  ${TOPIC}"
echo "  group ${GROUP} · bugs_open/030 · queued != lost"
echo "-----------------------------------------------------------------------"

if ! command -v kubectl >/dev/null 2>&1; then
  echo "  kubectl not on PATH — cannot read the lane. Nothing is wrong with your"
  echo "  submission; this report is just unavailable."
  exit 0
fi

# --describe on ONE named group. Never --all-groups: it iterates every group in
# the cluster and takes >120s (030 gotcha). The CLI is under /opt/kafka/bin and
# is not on $PATH.
DESCRIBE="$(timeout 60 kubectl -n "$KAFKA_NS" exec "$KAFKA_POD" -- \
  /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group "$GROUP" 2>/dev/null)" || DESCRIBE=""

if [[ -z "$DESCRIBE" ]]; then
  echo "  Could not reach the broker (pod ${KAFKA_POD} in ns ${KAFKA_NS})."
  echo "  Unknown queue depth — this says nothing about your dispatch."
  exit 0
fi

read -r LAG CUR END MEMBERS <<<"$(awk -v t="$TOPIC" '
  $2 == t { lag += ($6 ~ /^[0-9]+$/ ? $6 : 0); cur = $4; end = $5;
            if ($7 != "-" && $7 != "") members++ }
  END { printf "%d %s %s %d", lag, (cur==""?"?":cur), (end==""?"?":end), members }
' <<<"$DESCRIBE")"

echo "  consumer position : ${CUR}"
echo "  lane end          : ${END}"
echo "  QUEUE DEPTH (LAG) : ${LAG}"

if (( MEMBERS == 0 )); then
  echo ""
  echo "  ** NO ACTIVE CONSUMER in group ${GROUP}. **"
  echo "  This one IS a fault, not a wait: nothing is reading the lane, so nothing"
  echo "  queued on it will run until a consumer joins. Check the chassis pod:"
  echo "    kubectl -n ${PG_NS} get pods -l app=agent-chassis"
elif (( LAG > 1 )); then
  echo ""
  echo "  Your message was published to the back of this lane, so roughly"
  echo "  $((LAG - 1)) message(s) are ahead of it. It is QUEUED, NOT LOST:"
  echo "  an absent orchestration_states row means 'not started yet'."
  echo "  DO NOT re-fire — a duplicate spends the same LLM credits and lands"
  echo "  even further back in this same lane (bugs_open/030 landmine)."
else
  echo ""
  echo "  Lane is clear (LAG ${LAG}); expect your run to start shortly."
fi

if (( BRIEF )); then
  exit 0
fi

# Liveness, the honest way. A static CURRENT-OFFSET is worthless as a liveness
# signal — it holds still for the whole duration of the longest orchestration on
# the lane, and reading that as a dead consumer is a recorded wrong call
# (bugs_open/030 landmine, 2026-07-20). A recent state transition is the real
# signal, so ask Postgres, not Kafka.
LIVE="$(psql_q -t -A -c "SELECT EXTRACT(EPOCH FROM (now() - max(updated_at)))::int
  FROM orchestration_states WHERE updated_at > now() - interval '2 hours';")"
if [[ -n "${LIVE//[[:space:]]/}" ]]; then
  echo ""
  echo "  Consumer liveness: last orchestration step advanced ${LIVE//[[:space:]]/}s ago."
fi

echo ""
echo "  In flight now (a long one here is what you are waiting behind):"
psql_q -c "SELECT status,
       COALESCE(NULLIF(orchestration_name,''), '(unnamed)') AS name,
       current_step,
       EXTRACT(EPOCH FROM (now() - created_at))::int AS age_s,
       EXTRACT(EPOCH FROM (now() - updated_at))::int AS since_step_s
  FROM orchestration_states
 WHERE status IN ('EXECUTING_STEP','AWAITING_RESPONSES')
 ORDER BY created_at DESC LIMIT 8;" | sed 's/^/    /'

echo ""
echo "  No ETA is printed, deliberately: this lane has no stable drain rate."
echo "  Re-run this script to watch the depth; see the workstream runbook"
echo "  docs/agent_docs/docs024_key_docs_latest/dispatch_queue_serialisation/."
echo "-----------------------------------------------------------------------"
exit 0
