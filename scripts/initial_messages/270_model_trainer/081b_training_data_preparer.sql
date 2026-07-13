AGENT_TYPE="training-data-preparer"
EXPORT_ID="146a9a12-c953-48eb-bf1f-c1856e5f13b7"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

read -r -d '' HYPERPARAMETERS <<'HP'
{"base_model":"unsloth/Llama-3.3-70B-Instruct-bnb-4bit","epochs":3,"batch":1,"grad_accum":8,"lr":2e-4,"lora_r":16,"lora_alpha":32,"max_seq":4096,"seed":3407}
HP
HYPERPARAMETERS_INLINE=$(echo "$HYPERPARAMETERS" | jq -c .)

echo "=== Direct test: training-data-preparer ==="
echo "  Export:        $EXPORT_ID"
echo "  Correlation:   $CORRELATION_ID"
echo "  Orchestration: $ORCHESTRATION_ID"
echo "  Timestamp: $TIMESTAMP"
echo ""

kubectl -n kafka run -i --rm kcat-prep-$(date +%s) \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
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
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"export_id":"${EXPORT_ID}","hyperparameters":${HYPERPARAMETERS_INLINE}}}
JSON