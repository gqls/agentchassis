# Get a known site to replan against
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"  # robot-hands.com
DOMAIN="robot-hands.com"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-step2-$(date +%s) \
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
    -H client_id=demo_client \
    -H action=orchestrate \
    -H sender_agent_type=cli \
    -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses \
    -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"build-site-planner"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN"}}
JSON


# Check all three pods since we don't know which one will pick up the work
for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  echo "=== $pod ==="
  kubectl -n ai-persona-system logs $pod | grep -E "flattenImageryBlock|WriteSitePlanAction: Complete" | tail -5
done