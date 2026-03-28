DOMAIN="leopardessconsulting.co.uk"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"

SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

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
  -H action=orchestrate \
  -H to_agent_type=webdesign-agent \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"webdesign-agent"},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}"}}
JSON