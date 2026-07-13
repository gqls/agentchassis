PROV_ID=$(kubectl -n ai-persona-system exec postgres-clients-0 -- \
  psql -U clients_user -d clients_db -t -c "
SELECT thunder_instance_id FROM thunder_instances
WHERE status NOT IN ('decommissioned','failed','reaped')
ORDER BY created_at DESC LIMIT 1;" | tr -d ' ')
echo "Will decommission Thunder identifier=$PROV_ID"

kubectl -n kafka run kcat-decomm-$(date +%s) \
  --rm -i --restart=Never \
  --image=edenhill/kcat:1.7.1 -- \
  kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.adapter.thunder.requests \
    -H correlation_id=$(cat /proc/sys/kernel/random/uuid) \
    -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_type=request \
    -H sender_agent_type=cli \
    -H step_name=manual_decomm_killswitch \
    -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<JSON
{"action":"decommission_instance","thunder_identifier":"$PROV_ID","reason":"killswitch_after_test","reply_to_topic":"system.agent.generic.responses"}
JSON