CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

echo "=== Webdesign: $CORRELATION_ID ==="

kubectl -n kafka run -i --rm kcat-webdesign-$(date +%s) \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
-t system.agent.generic.requests \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H request_id=$REQUEST_ID \
-H message_id=$MESSAGE_ID \
-H message_type=request \
-H client_id=$CLIENT_ID \
-H action=process \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H responses_topic=system.agent.generic.responses \
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_webdesign_agent","steps":{"spawn_webdesign_agent":{"action":"spawn_agent","config":{"role":"webdesigner","agent_type":"webdesign-agent"},"description":"Spawn webdesign agent","next_step":"call_webdesign_agent","output_field":"webdesign_agent_info"},"call_webdesign_agent":{"action":"call_agent","config":{"agent_type":"webdesign-agent","target_role":"webdesigner","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain"},"timeout_seconds":240},"description":"Call webdesign agent to generate CSS","next_step":"complete","output_field":"webdesign_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["webdesign_result"]}}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo "Monitor: kubectl -n ai-persona-system logs -f -l agent-type=webdesign-agent --tail=50"


