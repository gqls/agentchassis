CORRELATION=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"
echo "CORRELATION=$CORRELATION  ORCH=$ORCH  REQ=$REQ   (write these down)"

kubectl -n kafka run kcat-iter0-$(date +%s) --rm -i --restart=Never \
  --image=edenhill/kcat:1.7.1 -- \
  -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -k "$CORRELATION" \
  -H correlation_id=$CORRELATION \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_type=request \
  -H action=orchestrate \
  -H client_id=$CLIENT_ID \
  -H step_name=iter0_training_run \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H from_agent_type=user \
  -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<'JSON'
{"action":"orchestrate","config":{"agent_type":"model-trainer"},"input_data":{"export_id":"146a9a12-c953-48eb-bf1f-c1856e5f13b7","hyperparameters":{"epochs":3,"batch":1,"grad_accum":8,"lr":0.0002,"lora_r":16,"lora_alpha":16,"max_seq":4096}}}
JSON