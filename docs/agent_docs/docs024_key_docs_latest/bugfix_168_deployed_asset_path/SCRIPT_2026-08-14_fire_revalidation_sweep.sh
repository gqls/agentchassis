#!/usr/bin/env bash
# Fire the review-queue revalidation sweep ON DEMAND.
#
# Mirrors cmd/scheduler/main.go fireTrigger() exactly (body shape + headers), so
# the chassis sees what the daily schedule sends. Deliberately does NOT touch
# scheduled_tasks.last_triggered_at: winding that back would fire the sweep but
# would also move the daily anchor off 08:44Z permanently, and that row is
# shared state this lane does not own.
#
# The sender is named honestly — from_agent_type=cli and an orchestration_name
# of manual-... — so a reader of orchestration_states can tell this run from a
# scheduled one rather than having to infer it from the clock.
#
# ⚠ kcat -P can send NOTHING at exit 0 (LANDMINE). The proof is the row.
set -euo pipefail

AGENT_TYPE="diagnosis-review-queue-revalidator"
TOPIC="system.agent.generic.requests"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_NAME="manual-review-queue-revalidate-$(date -u +%Y%m%d-%H%M%S)"

echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo "SAVE: ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo "SAVE: ORCH_NAME=$ORCH_NAME"
echo ""

kubectl -n kafka run -i --rm "kcat-revalidate-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t "$TOPIC" \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=$ORCH_NAME" \
  -H "step_name=start" \
  -H "client_id=system" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=cli" \
  -H "from_agent_id=cli-manual-sweep" <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{}}
JSON

echo ""
echo "Verify:  SELECT status, current_step FROM orchestration_states WHERE orchestration_id='${ORCHESTRATION_ID}';"
