SELECT id, thunder_instance_id, instance_ip, status, running_since
FROM thunder_instances
WHERE status = 'running'
ORDER BY created_at DESC LIMIT 3;

PROVISIONING_ID='paste-the-id-here'
PROVISIONING_ID='fabfd7fa-ac84-4476-86f3-f7ac57862214'
CORRELATION=$(uuidgen)

printf '{"headers":{"correlation_id":"%s","client_id":"demo_client","message_type":"request","action":"decommission_instance"},"body":{"action":"decommission_instance","provisioning_id":"%s","reason":"manual cleanup: iter0 orphan, call_launcher failed before training","reply_to_topic":"system.generic.responses"}}\n' "$CORRELATION" "$PROVISIONING_ID" \
| kubectl -n kafka run kcat-decom-$RANDOM --rm -i --restart=Never --image=edenhill/kcat:1.7.1 -- \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.adapter.thunder.requests \
    -P -k "$CORRELATION" \
    -H correlation_id="$CORRELATION" -H client_id=demo_client -H message_type=request -H action=decommission_instance


check:
clients_db=# SELECT id, thunder_instance_id, status, decommissioned_at
FROM thunder_instances
WHERE status IN ('provisioning','running','decommissioning')
ORDER BY created_at DESC;
 id | thunder_instance_id | status | decommissioned_at
----+---------------------+--------+-------------------
(0 rows)

SELECT * FROM thunder_provision_check;   -- can_provision + reason
SELECT * FROM thunder_config;            -- daily_cap_usd, max_concurrent, etc.


lifecycle training running_since
SELECT id, status, started_at FROM model_lifecycle.training_runs WHERE status='running';
SELECT id, thunder_instance_id, status FROM thunder_instances
WHERE status NOT IN ('decommissioned','reaped','failed');

UPDATE model_lifecycle.training_runs
SET status='failed',
    error_message='superseded by iter_0 run 1cd65dd7; box already gone — reconciled by hand',
    completed_at=now()
WHERE id='e6ab9fad...' AND status='running';