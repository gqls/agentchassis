#!/usr/bin/env bash
# 209/231 behavioural proof — fire ONE legacy workflow at the sacrificial domain.
# Usage: ./fire_209_proof.sh pageflow|swo
# Robust kcat pattern (memory: kcat-publish-silently-drops): payload in the
# container COMMAND, --attach, PUBLISH_OK marker. No PUBLISH_OK => not published.
set -euo pipefail

MODE="${1:?pageflow|swo}"
SITE_ID="11c884e5-4174-4841-9607-20919cf4f3d5"
DOMAIN="cookly.uk"
CLIENT_ID="demo_client"
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.generic.requests"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

if [ "$MODE" = "pageflow" ]; then
  # From scripts/initial_messages/090_new_build/071b_new_build "trigger just pageflow-builder"
  JSON='{"headers":{"correlation_id":"'$CORRELATION_ID'","orchestration_id":"'$ORCHESTRATION_ID'","request_id":"'$REQUEST_ID'","message_id":"'$MESSAGE_ID'","message_type":"request","client_id":"'$CLIENT_ID'","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"'$TIMESTAMP'"},"config":{"workflow":{"start_step":"load_site","steps":{"load_site":{"action":"ensure_site_record","config":{},"description":"Load site record with content_data from database","output_field":"site_record","next_step":"spawn_builder"},"spawn_builder":{"action":"spawn_agent","config":{"role":"builder","agent_type":"pageflow-builder"},"output_field":"builder_agent","next_step":"call_builder"},"call_builder":{"action":"call_agent","config":{"agent_type":"pageflow-builder","target_role":"builder","input_mapping":{"site_id":"site_record.site_id","domain":"site_record.domain","repo_name":"input_data.repo_name","reviewed_brief":"site_record.content_data","input_data":"input_data"},"timeout_seconds":1800},"output_field":"build_result","next_step":"complete"},"complete":{"action":"complete_workflow","config":{"output_fields":["build_result"]}}}}},"input_data":{"site_id":"'$SITE_ID'","domain":"'$DOMAIN'","repo_name":"sites"}}'
else
  # site-work-orchestrator, mode=build (check_mode: != maintenance -> full pipeline)
  JSON='{"headers":{"correlation_id":"'$CORRELATION_ID'","orchestration_id":"'$ORCHESTRATION_ID'","request_id":"'$REQUEST_ID'","message_id":"'$MESSAGE_ID'","message_type":"request","client_id":"'$CLIENT_ID'","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"'$TIMESTAMP'"},"config":{"workflow":{"start_step":"spawn_orchestrator","steps":{"spawn_orchestrator":{"action":"spawn_agent","config":{"role":"site_orchestrator","agent_type":"site-work-orchestrator"},"output_field":"orchestrator","next_step":"call_orchestrator","description":"Spawn site work orchestrator"},"call_orchestrator":{"action":"call_agent","config":{"target_role":"site_orchestrator","input_mapping":{"domain":"input_data.domain","mode":"input_data.mode","site_id":"input_data.site_id","reviewed_brief":"input_data.reviewed_brief"},"timeout_seconds":1800},"output_field":"orchestrator_result","next_step":"complete"},"complete":{"action":"complete_workflow","config":{"output_fields":["orchestrator_result"]},"description":"Build complete"}}}},"input_data":{"domain":"'$DOMAIN'","mode":"build","site_id":"'$SITE_ID'"}}'
fi

CMD="printf '%s' '${JSON}' | kcat -P -b ${BROKER} -t ${TOPIC} \
 -H correlation_id=${CORRELATION_ID} -H orchestration_id=${ORCHESTRATION_ID} \
 -H request_id=${REQUEST_ID} -H message_id=${MESSAGE_ID} -H message_type=request \
 -H client_id=${CLIENT_ID} -H action=process -H sender_agent_type=cli \
 -H sender_agent_id=cli-user -H timestamp=${TIMESTAMP} && echo PUBLISH_OK"

echo "MODE=$MODE"
echo "CORRELATION_ID=$CORRELATION_ID"
echo "ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run "kcat-209-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "$CMD"
