#!/usr/bin/env bash
# VERIFY_034_post_roll.sh — end-to-end proof for bugs_open/034.
#
# Run AFTER a chassis roll that carries 94c4ff471 + 56e77a501.
#
# What it does, in the order that matters:
#   1. pod-greps the symbols the fix CREATED (with a positive AND a negative
#      control — a bare grep that finds something proves nothing about which
#      binary you are looking at);
#   2. publishes a deliberately malformed request — valid correlation_id and
#      orchestration_id, message_type=request, and NO client_id;
#   3. reads agent_error_log back for that exact correlation.
#
# Before the fix, step 3 returns nothing: an error response goes out, the
# message is acked, and the database never hears about it. That silence IS the
# bug (bugs_open/002 error F spent two days as "accepted, never executed, no
# error anywhere"). A row here is the first end-to-end proof this bug has had.
#
# GOTCHAS, each one paid for:
#   - kcat -P splits stdin on newlines into SEPARATE messages. `-c 1` and a
#     single-line payload are both required.
#   - The column is `occurred_at`. The original handoff's verification recipe
#     says `created_at`, which does not exist on this table.
#   - Do NOT omit orchestration_id as well as client_id if you want to prove
#     site 1 specifically — omitting both still lands on site 1, but the row's
#     missing_headers list is what tells you which gate you hit.
#   - Give the roll ~300s before dispatching anything at a restarted chassis
#     pod; spawns inside that window are silently dropped.
set -euo pipefail

NS=ai-persona-system
POD=$(kubectl -n "$NS" get pods -o name | grep -m1 'agent-chassis' | cut -d/ -f2)
[ -n "$POD" ] || { echo "ERROR: no agent-chassis pod found" >&2; exit 1; }

echo "=== 1. pod-grep: $POD ==="
kubectl -n "$NS" get deploy agent-chassis -o jsonpath='{.spec.template.spec.containers[0].image}'; echo
kubectl exec -n "$NS" "$POD" -- sh -c '
  for s in INCOMING_MESSAGE_REJECTED MISSING_ORCHESTRATION_ID VALIDATION_ERROR_DROPPED \
           "agentbase.processMessage/ValidateIncomingMessage" "messaging.handleError"; do
    printf "  %-56s %s\n" "$s" "$(strings /app/agent-chassis | grep -c "$s")"
  done
  printf "  %-56s %s  (positive control, expect >=1)\n" MARK_COMPLETE_FAILED \
    "$(strings /app/agent-chassis | grep -c MARK_COMPLETE_FAILED)"
  printf "  %-56s %s  (negative control, expect 0)\n" zzz_not_a_real_symbol_034 \
    "$(strings /app/agent-chassis | grep -c zzz_not_a_real_symbol_034)"' || true

CORR="034verify-$(date +%s)"
echo
echo "=== 2. induce: publishing a request with NO client_id (corr=$CORR) ==="

# ONE LINE — see the kcat gotcha above.
PAYLOAD='{"action":"orchestrate","config":{"agent_type":"council-gate"},"input_data":{"probe":"bugs_open/034 post-roll verification"}}'

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-034v-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORR" \
  -H "request_id=$(cat /proc/sys/kernel/random/uuid)" \
  -H "message_id=$(cat /proc/sys/kernel/random/uuid)" \
  -H "orchestration_id=$(cat /proc/sys/kernel/random/uuid)" \
  -H "orchestration_name=034-post-roll-verify" \
  -H "step_name=start" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli-034-verify" \
  -H "responses_topic=system.agent.generic.responses" >/dev/null 2>&1

echo "published; waiting 20s for the chassis to consume and record"
sleep 20

echo
echo "=== 3. the row (this is the whole point) ==="
kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT occurred_at,
       agent_type,
       error_code,
       severity,
       left(error_message, 70)          AS err,
       context->>'missing_headers'      AS missing,
       context->>'dropped_at'           AS site
FROM agent_error_log
WHERE context->>'correlation_id' = '$CORR'
ORDER BY occurred_at DESC;"

echo
echo "PASS if one INCOMING_MESSAGE_REJECTED row names client_id in 'missing'."
echo "Zero rows means the fix is NOT in the running binary — check step 1's counts,"
echo "not git and not the image tag."
