#!/bin/bash
# ============================================================================
# PROBE_write_door_guard.sh — prove the WDS-018 write-door guard (bugs_open/291)
# is live and SELECTIVE in the running chassis, not merely deployed.
#
# WHY A DISPATCH AND NOT A QUERY. The guard lives in writeWorkItem, which no SQL
# can reach: a hand INSERT bypasses the Go door entirely (that is the guard's own
# stated limit — ~41 raw INSERT sites). The only way to exercise it is to make the
# running binary write a work item, which means driving the create_work_item
# action. The workflow travels INLINE in the message (selectWorkflow Priority 1,
# processor.go), so NO agent_definitions row is written; a misfire degrades to
# `generic`'s no-op complete step and is inert AND visible.
#
# THE TEST AND ITS CONTROL, in one dispatch:
#   probe_guard   : handler_agent = a name that is NOT in agent_definitions,
#                   at a status in the guard's trigger set.
#                   PASS  => row is born status='blocked' with
#                            error='Handler agent not registered: <name>' and
#                            claimed_at IS NULL — blocked AT WRITE, never claimed.
#                   Under the pre-guard binary this row is born 'claimed' with a
#                   NULL error, so the observable genuinely discriminates.
#   probe_control : handler_agent = 'tool-improver' (registered), same status.
#                   MUST be born unblocked. Without it a guard that blocked
#                   EVERYTHING would read as a pass.
#
# WHY status='claimed' AND NOT 'triaged'. Both are in the guard's trigger set
# (workItemStatusRequiresRegisteredHandler), but only triaged/approved are
# DISPATCHABLE — so a control row born 'claimed' exercises the guard's real branch
# while being structurally incapable of being picked up and actually run against a
# live site. The probe cleans both rows up at the end regardless.
#
# Usage: ./PROBE_write_door_guard.sh [site_id]
# ============================================================================
set -uo pipefail

NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

SITE_ID="${1:-}"
if [ -z "$SITE_ID" ]; then
  SITE_ID=$($PSQL -t -A -c "SELECT id FROM sites ORDER BY created_at LIMIT 1;" | tr -d '[:space:]')
fi
STAMP=$(date +%s)
FAKE_HANDLER="zzz-unregistered-probe-291"
KEY_GUARD="probe291_guard_${STAMP}"
KEY_CONTROL="probe291_control_${STAMP}"
CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)

echo "== WDS-018 write-door guard — demand control =="
echo "site_id:          $SITE_ID"
echo "fake handler:     $FAKE_HANDLER"
echo "correlation:      $CID"
echo "running image(s): $(kubectl -n $NS get pods -l app=agent-chassis -o jsonpath='{.items[*].spec.containers[0].image}')"
echo ""
echo "-- precondition: the fake handler must NOT be registered (else the probe is void) --"
REGISTERED=$($PSQL -t -A -c "SELECT count(*) FROM agent_definitions WHERE type='$FAKE_HANDLER' AND deleted_at IS NULL;" | tr -d '[:space:]')
if [ "$REGISTERED" != "0" ]; then echo "VOID: '$FAKE_HANDLER' exists in agent_definitions"; exit 2; fi
CONTROL_OK=$($PSQL -t -A -c "SELECT count(*) FROM agent_definitions WHERE type='tool-improver' AND deleted_at IS NULL;" | tr -d '[:space:]')
if [ "$CONTROL_OK" = "0" ]; then echo "VOID: control handler 'tool-improver' is not registered"; exit 2; fi
echo "fake handler registered: $REGISTERED (must be 0)   control handler: $CONTROL_OK (must be >0)"
echo ""

PAYLOAD_B64=$(python3 - "$SITE_ID" "$FAKE_HANDLER" "$KEY_GUARD" "$KEY_CONTROL" <<'PY'
import base64, json, sys
site_id, fake_handler, key_guard, key_control = sys.argv[1:5]
def step(name, handler, key, desc, nxt, out):
    return {
        "action": "create_work_item",
        "description": desc,
        "config": {
            "site_id": "input_data.site_id",
            "item_type": "probe_write_door_291",
            "handler_agent": handler,
            # 'claimed' is in the guard's trigger set but NOT dispatchable, so the
            # probe cannot cause real work to run.
            "status": "claimed",
            "severity": "low",
            "priority": 100,
            "source": "probe-291",
            "summary": "bugs_open/291 write-door guard probe — safe to cancel",
            "item_key_prefix": key,
            "spec_literal": {"probe": "291_write_door_guard"},
        },
        "next_step": nxt,
        "output_field": out,
    }
msg = {
    "action": "orchestrate",
    "config": {
        "agent_type": "write-door-guard-probe",
        "workflow": {
            "start_step": "probe_guard",
            "processing_mode": "orchestrator",
            "timeout_seconds": 120,
            "steps": {
                "probe_guard": step("probe_guard", fake_handler, key_guard,
                                    "THE TEST: unregistered handler at a trigger-set status",
                                    "probe_control", "guard_item"),
                "probe_control": step("probe_control", "tool-improver", key_control,
                                      "THE CONTROL: registered handler, must NOT be demoted",
                                      "finish", "control_item"),
                "finish": {
                    "action": "complete_workflow",
                    "description": "probe complete",
                    "config": {"multiple_output_fields": ["guard_item", "control_item"]},
                },
            },
        },
    },
    "input_data": {"site_id": site_id},
}
line = json.dumps(msg, separators=(",", ":"))
assert "\n" not in line
sys.stdout.write(base64.b64encode(line.encode()).decode())
PY
)

echo "publishing to system.agent.generic.requests ..."
kubectl -n kafka run "kcat-p291-${STAMP}-$RANDOM" \
  --rm --restart=Never --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "echo '$PAYLOAD_B64' | base64 -d | kcat -P \
    -b $BROKER \
    -t system.agent.generic.requests \
    -H correlation_id=$CID \
    -H orchestration_id=$OID \
    -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) \
    -H orchestration_name=write-door-guard-probe-291 \
    -H step_name=start \
    -H client_id=demo_client \
    -H message_type=request \
    -H action=orchestrate \
    -H from_agent_type=user \
    -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "No PUBLISH_OK above means NOTHING was published — kcat exits 0 on a drop."
echo "Waiting for the run ..."
for i in $(seq 1 20); do
  sleep 3
  STATUS=$($PSQL -t -A -c "SELECT status||'|'||current_step FROM orchestration_states WHERE correlation_id='$CID';" 2>/dev/null | tr -d '[:space:]' || true)
  [ -n "$STATUS" ] && echo "  [$i] $STATUS"
  case "$STATUS" in COMPLETED*|FAILED*|completed*|failed*) break ;; esac
done

echo ""
echo "== VERDICT — the two rows the running binary actually wrote =="
$PSQL -c "
SELECT
  CASE WHEN item_key LIKE 'probe291_guard%' THEN 'TEST    (unregistered)'
       ELSE                                       'CONTROL (registered)' END AS arm,
  handler_agent, status, error, (claimed_at IS NULL) AS never_claimed
FROM site_work_items
WHERE item_key IN ('${KEY_GUARD}_', '${KEY_CONTROL}_')
   OR item_key LIKE 'probe291_%${STAMP}%'
ORDER BY 1;"

echo "PASS requires ALL of:"
echo "  TEST    row: status='blocked', error='Handler agent not registered: $FAKE_HANDLER', never_claimed=t"
echo "  CONTROL row: status='claimed' (NOT blocked), error NULL"
echo "An all-green run with the CONTROL also blocked is NOT a pass — it means the"
echo "guard blocks everything and the test arm proved nothing."
echo ""
echo "-- cleanup: cancelling both probe rows --"
$PSQL -c "
UPDATE site_work_items SET status='cancelled', updated_at=now()
WHERE item_type='probe_write_door_291' AND item_key LIKE 'probe291_%${STAMP}%'
RETURNING item_key, status;"
