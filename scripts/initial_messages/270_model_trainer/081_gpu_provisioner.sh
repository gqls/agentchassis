# In a third terminal
CORRELATION=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
FIRE_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "CORRELATION=$CORRELATION"
echo "ORCH=$ORCH"
echo "REQ=$REQ"
echo "FIRE_TIME=$FIRE_TIME"
echo "(write these down — they're our anchors for everything that follows)"

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
    -H timestamp=$FIRE_TIME <<'JSON'
{"action":"orchestrate","config":{"agent_type":"gpu-provisioner"},"input_data":{"gpu":"a100","mode":"prototyping","num_gpus":1}}
JSON



----------------------


--------------------------------------------------------------

clients_db=# -- Cap and gating
SELECT daily_cap_usd, estimated_new_run_cost_usd, is_paused FROM thunder_config WHERE singleton='X';
SELECT spend_24h_usd FROM thunder_spend_24h;
SELECT can_provision, denial_reason FROM thunder_provision_check;
-- Expect: can_provision=t. If not, lower estimated_new_run_cost_usd to 2 and retry.

-- gpu-provisioner workflow swap is still in place
SELECT default_config->'workflow'->>'start_step' AS start_step
FROM agent_definitions
WHERE type='gpu-provisioner' AND is_active=true;
-- Expect: dispatch_provision

-- Clean up the stuck awaited_requests from yesterday's failed runs
UPDATE awaited_requests
SET status='expired', processed_at=NOW()
WHERE orchestration_id IN (
  'a0129348-c48f-4dff-85e6-300b8fba2d39',
  'b0afa101-159d-4854-b58e-51b1342d1f37'
) AND status='waiting';
 daily_cap_usd | estimated_new_run_cost_usd | is_paused
---------------+----------------------------+-----------
            15 |                         25 | f
(1 row)

ERROR:  column "spend_24h_usd" does not exist
LINE 1: SELECT spend_24h_usd FROM thunder_spend_24h;
               ^
 can_provision |     denial_reason
---------------+-----------------------
 f             | cost_cap_would_exceed
(1 row)

     start_step
--------------------
 dispatch_provision
(1 row)

UPDATE 0

--
Step 2: open the live log watcher BEFORE firing
# Terminal A: chassis pods, tail live, prefix with pod name
( for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
    kubectl -n ai-persona-system logs $pod -f --since=10s --prefix=true 2>&1 \
      | sed "s|^|[$pod] |" &
  done
  wait )

kubectl -n ai-persona-system logs --tail=300 -l app=agent-chassis -f | tee logs-agent-chassis.json

--

kubectl -n ai-persona-system logs --tail=500 -l app=thunder-adapter -f --max-log-requests 20 | tee logs-thunder-adapter.json

# Terminal B
kubectl -n ai-persona-system logs deploy/thunder-adapter -f --since=10s \
  | grep -vE "Failed to fetch message from Kafka|context deadline exceeded"


--

# Quicker live filter once you have $ORCH:
( for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
    kubectl -n ai-persona-system logs $pod -f --since=10s 2>&1 \
      | grep --line-buffered -E "$ORCH|gpu-provisioner|dispatch_thunder|Response consumer|ClaimAwaitedRequest|Failed to create ExecutionContext|Awaited request not found" \
      | sed "s|^|[$pod] |" &
  done
  wait )

  ----

  Step 5: regardless of outcome, capture state

-- The orchestration
SELECT status, current_step, LEFT(error,200) AS err, updated_at
FROM orchestration_states
WHERE orchestration_id = '<ORCH from step 3>';

-- The awaited_request from dispatch_provision
SELECT request_id, status, sent_at, processed_at, timeout_at,
       processed_at - sent_at AS response_latency
FROM awaited_requests
WHERE orchestration_id = '<ORCH from step 3>';

-- If a real instance was provisioned
SELECT id, thunder_instance_id, status, instance_ip, ssh_user, cost_usd
FROM thunder_instances
ORDER BY created_at DESC LIMIT 1;

---

What we'll know from the results
OutcomeDiagnosisChassis logs show Response consumer received message for the adapter's reply + ClaimAwaitedRequest: successfully claimed + workflow advancesThe system is fine; previous test had a transient or unrelated issueResponse consumer received message but ClaimAwaitedRequest: not claimed status_before=waitingRace condition or matcher bug — different debugging pathNo Response consumer received message for our request_id within ~5s of adapter sendingThe chassis didn't consume our response — points to topic/partition routingFailed to create ExecutionContext from headersHeader-shape mismatch in thunder-adapter's response builder — concrete fixProvision denied for capsJust need to adjust thunder_config and refire

----------------------
-- set caps
UPDATE thunder_config
SET daily_cap_usd = 100,
    estimated_new_run_cost_usd = 25
WHERE singleton = 'X';


-----------------

CORRELATION=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
FIRE_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
echo "CORRELATION=$CORRELATION  ORCH=$ORCH  REQ=$REQ  FIRE=$FIRE_TIME"

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
    -H timestamp=$FIRE_TIME <<'JSON'
{"action":"orchestrate","config":{"agent_type":"gpu-provisioner"},"input_data":{"gpu":"a100","mode":"prototyping","num_gpus":1}}
JSON