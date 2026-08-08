#!/bin/bash
# bugs_open/136 §9 — the deterministic runtime witness.
#
# Creates a one-shot agent definition (type alias-witness-136, category test)
# whose single create_work_item step carries ONLY the deprecated key
# `item_domain` with the NON-default value "content", then dispatches exactly
# one orchestration of it. The row it files is born status='cancelled' — outside
# idx_swi_dedup's partial index and outside every dispatcher's status predicate,
# so it is inert by construction (090's trap #4: NOT 'detected', which
# triage_detect_items launders to pipeline='build').
#
# The discriminator: row pipeline='content' → alias honoured at runtime;
# pipeline='build' → the alias fell through to the hardcoded default.
# Either outcome is decisive — this observation CAN come out otherwise,
# which the create_work_item-on-live-lanes route could not (§9).
set -euo pipefail
NS=ai-persona-system
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db)
SYSTEM_SITE=eac60db8-b032-432b-b36d-76f37632045d   # system.internal (090 trigger, note 2)
CORR=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_ID=$(cat /proc/sys/kernel/random/uuid)

echo "== 1. agent definition (refuses if the type already exists) =="
EXISTS=$(printf '%s' "SELECT count(*) FROM agent_definitions WHERE type='alias-witness-136';" | "${PSQL[@]}" -t -A)
if [ "$EXISTS" != "0" ]; then
  echo "alias-witness-136 already exists ($EXISTS rows) — not inserting again."
else
"${PSQL[@]}" <<'SQL'
INSERT INTO agent_definitions (type, display_name, category, description, default_config)
VALUES (
  'alias-witness-136',
  'Runtime witness for bugs_open/136 config-key alias (temporary)',
  'test',
  'One-shot witness for bugs_open/136 SS9: a create_work_item step whose config carries ONLY the deprecated item_domain key, with the NON-default value content. The row it files is born cancelled (inert everywhere). Deactivate after the run; owned by the bugfix_136_config_key_aliases lane.',
  '{"workflow":{"start_step":"file_witness_row","steps":{
     "file_witness_row":{"action":"create_work_item","config":{
         "source":"alias-witness-136",
         "site_id":"input_data.site_id",
         "summary":"bugs_open/136 runtime witness - deprecated item_domain=content must land as pipeline=content; born cancelled, no consumer",
         "priority":1000,"severity":"low",
         "item_type":"alias_witness_136",
         "item_domain":"content",
         "status":"cancelled",
         "handler_agent":"none",
         "item_key_prefix":"alias_witness_136"},
       "next_step":"complete","description":"the witness: only the deprecated key, non-default value","output_field":"witness_item_created"},
     "complete":{"action":"complete_workflow","config":{"output_fields":["witness_item_created"]},"description":"done"}
  }}}'::jsonb
);
SQL
fi

echo "== 2. publish (payload in container COMMAND — kcat -P on kubectl-run stdin drops 4 of 5) =="
JSON=$(printf '{"action":"orchestrate","config":{"agent_type":"alias-witness-136"},"input_data":{"site_id":"%s","witness_marker":"%s","correlation_id":"%s"}}' "$SYSTEM_SITE" "$CORR" "$CORR")
kubectl -n kafka run "kcat-witness136-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '$JSON' | kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCH_ID -H orchestration_name=alias-witness-136-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start -H client_id=demo_client -H message_type=request \
  -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "SAVE: WITNESS_CORR=$CORR"
echo "Find the run by payload, not the printed id:"
echo "  SELECT orchestration_id, current_step, status FROM orchestration_states"
echo "  WHERE collected_data->'input_data'->>'witness_marker' = '$CORR';"
