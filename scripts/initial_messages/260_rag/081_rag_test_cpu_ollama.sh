CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

echo "CORRELATION_ID=$CORRELATION_ID"
echo "ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run -i --rm kcat-ragtest-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=rag-chassis-test-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=demo_client \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses <<'JSON'
{"action":"orchestrate","config":{"agent_type":"rag-test-agent"},"input_data":{"content":"French Bulldogs are brachycephalic breeds with shortened skulls. They are prone to Brachycephalic Obstructive Airway Syndrome (BOAS) which affects up to 50 percent of the breed and often requires surgical intervention. Labrador Retrievers are popular family dogs known for their friendly temperament and high energy levels. A grand piano produces sound through strings struck by felt hammers. Electric vehicles use rechargeable lithium-ion batteries.","query":"dog breed airway problems breathing difficulty"}}
JSON

echo ""
echo "Saved IDs for check:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"