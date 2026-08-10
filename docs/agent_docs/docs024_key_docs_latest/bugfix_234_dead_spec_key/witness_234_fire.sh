#!/bin/bash
# bugs_open/234 — the strict-validation runtime witness (RUNBOOK's canary).
#
# Creates a one-shot agent definition (type strict-witness-234, category test)
# whose single create_work_item step carries one BOGUS key
# (zzz_strict_witness_234) alongside a minimal valid config, then dispatches
# exactly one orchestration of it. Modelled on witness_136_fire.sh, including
# its kcat trap workaround (payload in the container COMMAND, not stdin).
#
# The discriminator — either outcome is decisive, so this observation CAN come
# out otherwise:
#   OLD binary (warn-only): workflow validates, the row is filed
#     (item_type='strict_witness_234', born status='cancelled' — outside
#     idx_swi_dedup's partial index and every dispatcher predicate, inert).
#   NEW binary (StrictConfig live): the canary pod logs
#     "Invalid workflow configuration" whose cause names zzz_strict_witness_234
#     and "declares its config contract as complete"; NO row is filed.
#
# Row absence ALONE is NOT the proof (the spawn->call handshake drops ~half of
# dispatches fleet-wide — memory: spawn-call-handshake-races); the pod log line
# is the positive signal, and witness_234_poll.sh captures it every 3s.
#
# Deactivate the definition after the run (witness_234_poll.sh prints the SQL).
set -euo pipefail
NS=ai-persona-system
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db)
SYSTEM_SITE=eac60db8-b032-432b-b36d-76f37632045d   # system.internal
CORR=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_ID=$(cat /proc/sys/kernel/random/uuid)

echo "== 1. agent definition (refuses if the type already exists) =="
EXISTS=$(printf '%s' "SELECT count(*) FROM agent_definitions WHERE type='strict-witness-234';" | "${PSQL[@]}" -t -A)
if [ "$EXISTS" != "0" ]; then
  echo "strict-witness-234 already exists ($EXISTS rows) — not inserting again."
else
"${PSQL[@]}" <<'SQL'
INSERT INTO agent_definitions (type, display_name, category, description, default_config)
VALUES (
  'strict-witness-234',
  'Runtime witness for bugs_open/234 strict config validation (temporary)',
  'test',
  'One-shot witness for bugs_open/234: a create_work_item step carrying one bogus config key (zzz_strict_witness_234). On the StrictConfig binary the workflow must be REJECTED at validation (pod log: Invalid workflow configuration, cause names the key); on the old binary it files an inert row born cancelled. Deactivate after the run; owned by the bugfix_234_dead_spec_key lane.',
  '{"workflow":{"start_step":"file_witness_row","steps":{
     "file_witness_row":{"action":"create_work_item","config":{
         "source":"strict-witness-234",
         "site_id":"input_data.site_id",
         "summary":"bugs_open/234 strict witness - this row must NEVER be filed on the strict binary",
         "priority":1000,"severity":"low",
         "item_type":"strict_witness_234",
         "status":"cancelled",
         "handler_agent":"none",
         "item_key_prefix":"strict_witness_234",
         "zzz_strict_witness_234":"bogus key the strict contract must refuse"},
       "next_step":"complete","description":"the witness: one bogus key on a strict action","output_field":"witness_item_created"},
     "complete":{"action":"complete_workflow","config":{"output_fields":["witness_item_created"]},"description":"done"}
  }}}'::jsonb
);
SQL
fi

echo "== 2. publish (payload in container COMMAND — kcat -P on kubectl-run stdin drops 4 of 5) =="
JSON=$(printf '{"action":"orchestrate","config":{"agent_type":"strict-witness-234"},"input_data":{"site_id":"%s","witness_marker":"%s","correlation_id":"%s"}}' "$SYSTEM_SITE" "$CORR" "$CORR")
kubectl -n kafka run "kcat-witness234-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '$JSON' | kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCH_ID -H orchestration_name=strict-witness-234-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start -H client_id=demo_client -H message_type=request \
  -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "SAVE: WITNESS_CORR=$CORR"
