# Preconditions

-- Confirm gating still open
SELECT can_provision, denial_reason FROM thunder_provision_check;
-- Expect: t | NULL

-- Clean up the row from the failed 17:10 run
UPDATE awaited_requests
SET status='expired', processed_at=NOW()
WHERE orchestration_id = 'c09c94a1-b7be-4644-ae55-31b4cd5a358a'
  AND status='waiting';

# Open the log watchers first
# Terminal A — adapter:

kubectl -n ai-persona-system logs deploy/thunder-adapter -f --since=10s \
  | grep -vE "Failed to fetch message from Kafka|context deadline exceeded"

# Terminal B — DB poll on awaited_requests + thunder_instances:

# Set ORCH after firing, then run:
$ORCH=
watch -n 5 "kubectl -n ai-persona-system exec postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c \"
SELECT 'awaited' AS kind, status, sent_at, processed_at,
       processed_at - sent_at AS latency
FROM awaited_requests WHERE orchestration_id = '\$ORCH'
UNION ALL
SELECT 'instance', status, created_at, decommissioned_at, NULL
FROM thunder_instances ORDER BY created_at DESC LIMIT 5;\""

# Fire
CORRELATION=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
echo "CORRELATION=$CORRELATION  ORCH=$ORCH  REQ=$REQ"
echo "(write these down)"

kubectl -n kafka run kcat-prov-$(date +%s) \
  --rm -i --restart=Never \
  --image=edenhill/kcat:1.7.1 -- \
  kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION \
    -H orchestration_id=$ORCH \
    -H request_id=$REQ \
    -H message_type=request \
    -H action=orchestrate \
    -H client_id=demo_client \
    -H step_name=manual_provision_test \
    -H sender_agent_type=cli \
    -H sender_agent_id=cli-user \
    -H from_agent_type=user \
    -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<'JSON'
{"action":"orchestrate","config":{"agent_type":"gpu-provisioner"},"input_data":{"gpu":"a100","mode":"prototyping","num_gpus":1}}
JSON