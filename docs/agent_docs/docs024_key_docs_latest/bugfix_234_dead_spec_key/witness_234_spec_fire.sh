#!/bin/bash
# bugs_open/234 — the SPEC-DELIVERY witness (the data half's filed-row proof).
#
# WHY THIS EXISTS RATHER THAN WAITING. The case file's proof is "a filed row
# carries refresh_site_components" — never a definition, because a config that
# LOOKS right is exactly what this bug was. The natural producer is
# improvement-loop's insert_rerender_item, but that step is conditional (it only
# fires when an audit found and promoted issues) and improvement-loop has filed
# nothing since 2026-08-09 14:56Z. Waiting is not a plan, and dispatching
# improvement-loop at a real site to force it would run audits, fixes and a
# rerender on live customer pages to prove a one-key change.
#
# So this witness carries improvement-loop's EXACT live step config, copied
# verbatim from agent_definitions (only `source`, `item_type`, `handler_agent`,
# `status` and `item_key_prefix` differ, so the row is inert and traceable):
#
#   "spec_literal": {"refresh_site_components": true}
#
# THE DISCRIMINATOR — this observation can come out otherwise, which is what
# makes it evidence:
#   spec = {"refresh_site_components": true}  → spec_literal IS delivered; the
#          bug's mechanism is fixed and 051's input_mapping
#          (pending.first_item.spec.refresh_site_components) has something to read.
#   spec = {}                                 → still broken; the rename did not
#          take effect and 364 fixed nothing.
#
# What this does NOT prove on its own: that improvement-loop's own run reaches
# this step. That half is established by reading its live config (identical
# spec_literal, verified the same day) — the step is unchanged apart from the
# key rename this bug is about.
#
# The row is born status='cancelled' — outside idx_swi_dedup's partial index and
# every dispatcher's status predicate, so no handler ever sees it. Deactivate the
# definition after the run (the poll script prints the SQL); while active it is
# harmless (nothing dispatches it) but it is still a live definition.
set -euo pipefail
NS=ai-persona-system
PSQL=(kubectl -n "$NS" exec -i postgres-clients-0 -- psql -U clients_user -d clients_db)
SYSTEM_SITE=eac60db8-b032-432b-b36d-76f37632045d   # system.internal
CORR=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
ORCH_ID=$(cat /proc/sys/kernel/random/uuid)

echo "== 1. agent definition (refuses if the type already exists in ANY state) =="
EXISTS=$(printf '%s' "SELECT count(*) FROM agent_definitions WHERE type='spec-witness-234';" | "${PSQL[@]}" -t -A)
if [ "$EXISTS" != "0" ]; then
  echo "spec-witness-234 already exists ($EXISTS rows) — DELETE it before re-firing."
else
"${PSQL[@]}" <<'SQL'
INSERT INTO agent_definitions (type, display_name, category, description, default_config)
VALUES (
  'spec-witness-234',
  'Spec-delivery witness for bugs_open/234 (temporary)',
  'test',
  'One-shot witness for bugs_open/234: a create_work_item step carrying improvement-loop insert_rerender_item''s EXACT live spec_literal. The filed row must carry {"refresh_site_components": true}; an empty spec means the rename did not take. Row born cancelled (inert everywhere). Deactivate after the run; owned by the bugfix_234_dead_spec_key lane.',
  '{"workflow":{"start_step":"file_witness_row","steps":{
     "file_witness_row":{"action":"create_work_item","config":{
         "source":"spec-witness-234",
         "site_id":"input_data.site_id",
         "summary":"bugs_open/234 spec-delivery witness - the filed spec must carry refresh_site_components; born cancelled, no consumer",
         "priority":1000,
         "severity":"medium",
         "item_type":"spec_witness_234",
         "spec_literal":{"refresh_site_components":true},
         "status":"cancelled",
         "handler_agent":"none",
         "item_pipeline":"build",
         "item_key_prefix":"spec_witness_234"},
       "next_step":"complete","description":"improvement-loop insert_rerender_item config, verbatim on the spec key","output_field":"witness_item_created"},
     "complete":{"action":"complete_workflow","config":{"output_fields":["witness_item_created"]},"description":"done"}
  }}}'::jsonb
);
SQL
fi

echo "== 2. publish (payload in container COMMAND — kcat -P on kubectl-run stdin drops 4 of 5) =="
JSON=$(printf '{"action":"orchestrate","config":{"agent_type":"spec-witness-234"},"input_data":{"site_id":"%s","witness_marker":"%s","correlation_id":"%s"}}' "$SYSTEM_SITE" "$CORR" "$CORR")
kubectl -n kafka run "kcat-specwitness234-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '$JSON' | kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCH_ID -H orchestration_name=spec-witness-234-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start -H client_id=demo_client -H message_type=request \
  -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"

echo ""
echo "SAVE: WITNESS_CORR=$CORR"
echo "The proof (row, not definition):"
echo "  SELECT item_key, spec, created_at FROM site_work_items WHERE item_type='spec_witness_234';"
echo "Cleanup: UPDATE agent_definitions SET is_active=false, deleted_at=now() WHERE type='spec-witness-234';"
