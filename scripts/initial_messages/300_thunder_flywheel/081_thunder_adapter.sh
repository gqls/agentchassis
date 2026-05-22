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
FROM thunder_instances ORDER BY sent_at DESC LIMIT 5;\""

# Fire
CORRELATION=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"
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
    -H client_id=$CLIENT_ID \
    -H step_name=manual_provision_test \
    -H sender_agent_type=cli \
    -H sender_agent_id=cli-user \
    -H from_agent_type=user \
    -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<'JSON'
{"action":"orchestrate","config":{"agent_type":"gpu-provisioner"},"input_data":{"gpu":"a100","mode":"prototyping","num_gpus":1}}
JSON

echo ""
echo "========================================="
echo "Thunder adapter provision"
echo "  Orchestration Id:  $ORCH"
echo "  Correlation Id:  $CORRELATION"
echo "  Request Id:  $REQ"
echo "  Client Id:  $CLIENT"
echo "  {action:orchestrate,config:{agent_type:gpu-provisioner},input_data:{gpu:a100,mode:prototyping,num_gpus:1}}"
echo "========================================="
echo ""

--------------------------------------------------------

SELECT thunder_instance_id, status, instance_ip, provisioned_at, requested_by
FROM thunder_instances
WHERE status = 'running'
ORDER BY provisioned_at DESC;

tnr status
tnr delete 0
tnr delete 1

# mark is as deleted:
UPDATE thunder_instances
SET status='decommissioned', decommissioned_at=NOW(),
    cost_usd = GREATEST(0, EXTRACT(EPOCH FROM (NOW()-running_since))/3600.0)*hourly_rate_usd
WHERE thunder_instance_id='0' AND status='running'
  AND provisioned_at > '2026-05-22';

----------------------------
clients_db=# SELECT can_provision, denial_reason FROM thunder_provision_check;
-- Expect: t | NULL

UPDATE awaited_requests
SET status='expired', processed_at=NOW()
WHERE status='waiting' AND sent_at < NOW() - INTERVAL '15 min';
 can_provision |     denial_reason
---------------+-----------------------
 f             | cost_cap_would_exceed
(1 row)

UPDATE 0
clients_db=#

-- Raise cap and check the gating view
UPDATE thunder_config SET daily_cap_usd = 15 WHERE singleton = 'X';

SELECT spend_24h_usd FROM thunder_spend_24h;
SELECT can_provision, denial_reason FROM thunder_provision_check;

--
UPDATE thunder_config SET daily_cap_usd = 100, updated_at = NOW() WHERE singleton = 'X';
SELECT can_provision, denial_reason FROM thunder_provision_check;
-- should now be t | NULL  ($64.89 + $2 < $100)

-----

-- Mark the orphaned rows decommissioned so the spend view clears.
-- Compute a cost from running_since so the 24h spend reflects real (past) usage.
UPDATE thunder_instances
SET status = 'decommissioned',
    decommissioned_at = NOW(),
    cost_usd = COALESCE(cost_usd,
                        GREATEST(0, EXTRACT(EPOCH FROM (NOW() - running_since))/3600.0) * hourly_rate_usd)
WHERE status NOT IN ('decommissioned','failed','reaped');


-----

-- See what the $64.89 is made of — confirms it's yesterday's tests, not a miscalculation
SELECT thunder_instance_id, status, running_since, decommissioned_at,
       cost_usd,
       EXTRACT(EPOCH FROM (decommissioned_at - running_since))/3600.0 AS billed_hours
FROM thunder_instances
WHERE decommissioned_at > NOW() - INTERVAL '24 hours'
ORDER BY decommissioned_at DESC;


---


