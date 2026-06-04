
-- get provisioning id and | training id
SELECT id AS provisioning_id, training_run_id
FROM thunder_instances WHERE status='running' AND training_run_id IS NOT NULL;

paste them into the json
fabfd7fa-ac84-4476-86f3-f7ac57862214|1cd65dd7-ad74-4f0d-b509-75f821c29d46


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
{"action":"orchestrate","config":{"agent_type":"thunder-training-monitor-worker"},"input_data":{"provisioning_id":"fabfd7fa-ac84-4476-86f3-f7ac57862214","training_run_id":"1cd65dd7-ad74-4f0d-b509-75f821c29d46"}}
JSON

echo "CORRELATION=$CORRELATION  ORCH=$ORCH  REQ=$REQ  "