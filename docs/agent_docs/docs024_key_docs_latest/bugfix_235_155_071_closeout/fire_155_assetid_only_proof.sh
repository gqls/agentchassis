#!/usr/bin/env bash
# bugs_open/155 closure proof — deploy ONE asset by asset_id ALONE (no s3_uri
# anywhere in the mapping; that absence IS the test). Run twice with two
# same-purpose assets, then sha256 the committed artefacts: distinct bytes each
# matching its own origin_prompt is the PASS the file's recipe demands.
# Usage: ./fire_155_assetid_only_proof.sh <domain> <purpose> <asset_key> <asset_id>
# Robust kcat pattern (memory: kcat-publish-silently-drops): PUBLISH_OK or it did not send.
set -euo pipefail

DOMAIN="${1:?domain}"; PURPOSE="${2:?purpose}"; ASSET_KEY="${3:?asset_key}"; ASSET_ID="${4:?asset_id}"
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.generic.requests"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

JSON='{"action":"orchestrate","config":{"agent_type":"generic","workflow":{"start_step":"spawn_deployer","steps":{"spawn_deployer":{"action":"spawn_agent","config":{"role":"deployer","agent_type":"asset-deployer"},"next_step":"call_deployer","output_field":"deployer_agent"},"call_deployer":{"action":"call_agent","config":{"agent_type":"asset-deployer","target_role":"deployer","input_mapping":{"domain":"input_data.domain","purpose":"input_data.purpose","asset_key":"input_data.asset_key","asset_id":"input_data.asset_id"},"timeout_seconds":180},"next_step":"complete","output_field":"deploy_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["deploy_result"]}}}}},"input_data":{"domain":"'$DOMAIN'","purpose":"'$PURPOSE'","asset_key":"'$ASSET_KEY'","asset_id":"'$ASSET_ID'"}}'

CMD="printf '%s' '${JSON}' | kcat -P -b ${BROKER} -t ${TOPIC} \
 -H correlation_id=${CORRELATION_ID} -H orchestration_id=${ORCHESTRATION_ID} \
 -H request_id=${REQUEST_ID} -H message_id=${MESSAGE_ID} -H message_type=request \
 -H client_id=demo_client -H action=orchestrate -H sender_agent_type=cli \
 -H sender_agent_id=cli-user -H responses_topic=system.agent.generic.responses \
 -H timestamp=${TIMESTAMP} && echo PUBLISH_OK"

echo "ASSET_KEY=$ASSET_KEY  ASSET_ID=$ASSET_ID"
echo "CORRELATION_ID=$CORRELATION_ID"
echo "ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run "kcat-155-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "$CMD"
