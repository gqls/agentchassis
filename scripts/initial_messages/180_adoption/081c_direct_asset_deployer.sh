# Replace <S3_URI> with what we find — should look like:
#   s3://personae-prod-uk001-images/images/demo_client/20260510/3e4c4bae-9c49-47de-898b-48f5ca0944f8.png

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Direct asset-deployer test (spawn+call wrapper) ==="
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "  Target:        robot-hands.com hero_about"
echo "  TIMESTAMP=$TIMESTAMP"
echo ""

kubectl -n kafka run -i --rm kcat-deploy-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H message_type=request \
    -H client_id=$CLIENT_ID \
    -H action=orchestrate \
    -H sender_agent_type=cli \
    -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses \
    -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"generic","workflow":{"start_step":"spawn_deployer","steps":{"spawn_deployer":{"action":"spawn_agent","config":{"role":"deployer","agent_type":"asset-deployer"},"next_step":"call_deployer","output_field":"deployer_agent"},"call_deployer":{"action":"call_agent","config":{"target_role":"deployer","input_mapping":{"domain":"input_data.domain","s3_uri":"input_data.s3_uri","purpose":"input_data.purpose","asset_key?":"input_data.asset_key"},"timeout_seconds":180},"next_step":"complete","output_field":"deploy_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["deploy_result"]}}}}},"input_data":{"domain":"robot-hands.com","s3_uri":"s3://personae-prod-uk001-images/images/demo_client/20260510/3e4c4bae-9c49-47de-898b-48f5ca0944f8.png","purpose":"hero","asset_key":"hero_about"}}
JSON

echo ""
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo "  TIMESTAMP=$TIMESTAMP"




-------------------------------------------------------------
getting the url
SELECT
    content_data->>'hero_uri'           AS hero_uri,
    content_data->>'hero_about_uri'     AS hero_about_uri,
    content_data->>'logo_uri'           AS logo_uri
FROM sites
WHERE id = '00ff3af5-dad8-4770-9f70-3edc267a3c92';