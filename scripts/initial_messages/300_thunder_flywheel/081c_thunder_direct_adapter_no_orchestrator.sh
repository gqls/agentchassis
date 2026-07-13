bash
cat << 'CMD'
# Provision through the adapter (direct topic). This creates an instance WITH our
# public_key (the whole point — probe v3 needs an adapter-provisioned box, not a
# manual tnr create). Note: bare provision (no training_run_id) → 2h reaper failsafe.
CORRELATION=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
echo "CORRELATION=$CORRELATION  REQ=$REQ  (write these down)"

kubectl -n kafka run kcat-prov-$(date +%s) --rm -i --restart=Never --image=edenhill/kcat:1.7.1 -- \
  kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.adapter.thunder.requests \
    -H "correlation_id=$CORRELATION" -H "request_id=$REQ" -H "message_type=request" <<'JSON'
{"body":{"action":"provision_instance","gpu":"a100","mode":"prototyping","num_gpus":1,"reply_to_topic":"system.agent.generic.responses"}}
JSON

echo ""
echo "Then watch for Provision complete + the db_row_id:"
echo "  kubectl -n ai-persona-system logs deploy/thunder-adapter --since=3m | grep -iE 'Provision complete|db_row_id|thunder_identifier'"
CMD

----------------------
Output
# Provision through the adapter (direct topic). This creates an instance WITH our
# public_key (the whole point — probe v3 needs an adapter-provisioned box, not a
# manual tnr create). Note: bare provision (no training_run_id) → 2h reaper failsafe.
CORRELATION=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
echo "CORRELATION=$CORRELATION  REQ=$REQ  (write these down)"

kubectl -n kafka run kcat-prov-$(date +%s) --rm -i --restart=Never --image=edenhill/kcat:1.7.1 -- \
  kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.adapter.thunder.requests \
    -H "correlation_id=$CORRELATION" -H "request_id=$REQ" -H "message_type=request" <<'JSON'
{"body":{"action":"provision_instance","gpu":"a100","mode":"prototyping","num_gpus":1,"reply_to_topic":"system.agent.generic.responses"}}
JSON

echo ""
echo "Then watch for Provision complete + the db_row_id:"
echo "  kubectl -n ai-persona-system logs deploy/thunder-adapter --since=3m | grep -iE 'Provision complete|db_row_id|thunder_identifier'"
