CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_NAME="vet-batch-$(date +%Y%m%d-%H%M%S)"
CLIENT_ID="vetcomparison"
BATCH_SIZE=100

kubectl -n kafka run -i --rm kcat-batch-$$ \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H "correlation_id=$CORRELATION_ID" \
    -H "request_id=$REQUEST_ID" \
    -H "message_id=$MESSAGE_ID" \
    -H "orchestration_id=$ORCHESTRATION_ID" \
    -H "orchestration_name=$ORCHESTRATION_NAME" \
    -H "step_name=start" \
    -H "client_id=$CLIENT_ID" \
    -H "message_type=request" \
    -H "action=orchestrate" \
    -H "from_agent_type=user" \
    -H "from_agent_id=cli" \
    -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"vet-batch-processor"},"input_data":{"batch_size":$BATCH_SIZE,"task_type":"initial_verification","vertical_slug":"veterinary"}}
JSON