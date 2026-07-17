#!/bin/bash
# 092_TRIGGER_experience_plan.sh — fire the experience-planner + challenge
# council for a named experience (RUNBOOK T3.3). Modelled on the 090
# needs_diagnosis trigger: a durable intake work item (inert status,
# ON CONFLICT DO NOTHING) + a kcat orchestrate envelope on the SAME
# correlation_id, so the plan, the council_report artifacts and the
# experience-council doc_notes all join on one key.
#
# Usage:
#   ./092_TRIGGER_experience_plan.sh <domain> <experience_key> [experience_name]
#   ./092_TRIGGER_experience_plan.sh vonc.com vonc-spark-game "the Spark daily-provocation game"
#
# PRECONDITIONS (image → seed → fire):
#   1. chassis image carries the docResolveSubject 'experience' fix (66d32477d)
#   2. 167_experience_planner_and_council.sql applied
#   3. no chassis pod restart within ~300s (spawn silently dropped otherwise)
set -euo pipefail

DOMAIN="${1:?Usage: $0 <domain> <experience_key> [experience_name]}"
EXPERIENCE_KEY="${2:?experience_key required, e.g. vonc-spark-game}"
EXPERIENCE_NAME="${3:-$EXPERIENCE_KEY}"

PSQL=(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -Atc)

SITE_ID=$("${PSQL[@]}" "SELECT id FROM sites WHERE domain='${DOMAIN}' LIMIT 1;")
if [[ -z "$SITE_ID" ]]; then echo "No site for domain ${DOMAIN}"; exit 1; fi

CID=$(cat /proc/sys/kernel/random/uuid)
OID=$(cat /proc/sys/kernel/random/uuid)
MID=$(cat /proc/sys/kernel/random/uuid)
RID=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "========================================="
echo "Experience Planner Trigger"
echo "  Domain:          $DOMAIN"
echo "  Site:            $SITE_ID"
echo "  Experience key:  $EXPERIENCE_KEY"
echo "  Correlation ID:  $CID"
echo "========================================="

# Durable intake row (inert to build sweeps; idempotent via idx_swi_dedup).
"${PSQL[@]}" "INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, status, created_by, item_key)
  VALUES ('${SITE_ID}', 'experience-trigger', 'experience', 'needs_experience_plan', 'medium',
          'Compose + council-review the ${EXPERIENCE_NAME} EXPERIENCE_PLAN', '{}'::jsonb, 50,
          'awaiting_experience_plan', 'cli', 'needs_experience_plan:${EXPERIENCE_KEY}')
  ON CONFLICT DO NOTHING;" >/dev/null || true

INPUT="{\"domain\":\"${DOMAIN}\",\"site_id\":\"${SITE_ID}\",\"experience_key\":\"${EXPERIENCE_KEY}\",\"experience_name\":\"${EXPERIENCE_NAME}\",\"experience_domain\":\"${DOMAIN}\",\"experience_correlation_id\":\"${CID}\"}"

# Single-line JSON envelope (kcat -P splits on newlines — never multi-line the payload).
kubectl -n kafka run -i --rm kcat-expplan-$(date +%s) \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CID -H orchestration_id=$OID -H request_id=$RID \
  -H message_id=$MID -H message_type=request -H client_id=demo_client \
  -H action=process -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$TS <<JSON
{"headers":{"correlation_id":"${CID}","orchestration_id":"${OID}","request_id":"${RID}","message_id":"${MID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TS}"},"config":{"workflow":{"start_step":"spawn_planner","processing_mode":"orchestrator","timeout_seconds":1200,"steps":{"spawn_planner":{"action":"spawn_agent","config":{"role":"planner","agent_type":"experience-planner"},"output_field":"planner","next_step":"call_planner","description":"Spawn experience-planner"},"call_planner":{"action":"call_agent","config":{"agent_type":"experience-planner","target_role":"planner","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","experience_key":"input_data.experience_key","experience_name":"input_data.experience_name","experience_domain":"input_data.experience_domain","experience_correlation_id":"input_data.experience_correlation_id"},"timeout_seconds":1100},"output_field":"plan_result","next_step":"complete","description":"Run the experience-planner + council"},"complete":{"action":"complete_workflow","config":{"output_fields":["plan_result"]},"description":"done"}}}},"input_data":${INPUT}}
JSON

echo "CORRELATION_ID=$CID"
echo "Watch:  SELECT status,current_step FROM orchestration_states WHERE correlation_id='${CID}';"
echo "Plan:   SELECT is_current,length(body) FROM doc_plans WHERE subject_type='experience' AND subject_key='${EXPERIENCE_KEY}' ORDER BY created_at DESC;"
