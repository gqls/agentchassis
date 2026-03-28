CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
AGENT_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "========================================="
echo "Remote Job Spawner — Smoke Test"
echo "========================================="
echo "  Correlation ID: $CORRELATION_ID"
echo "  Agent ID:       $AGENT_ID"
echo "  Time:           $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  AGENT_ID=$AGENT_ID"
echo ""

kubectl -n kafka run -i --rm kcat-dispatch-test-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.dispatch.requests \
  -H correlation_id=$CORRELATION_ID \
  -H agent_id=$AGENT_ID \
  -H agent_type=copywriter \
  -H target_cluster=uk_001 \
  -H message_type=dispatch_request <<JSON
{"agent_id":"$AGENT_ID","agent_type":"copywriter","agent_name":"copywriter-dispatch-test","role":"writer","client_id":"default","image_repository":"docker.io/aqls/agent-chassis","image_tag":"v1.0.813","command":null,"resources":{"requests":{"cpu":"100m","memory":"256Mi"},"limits":{"cpu":"500m","memory":"1Gi"}},"health_config":{"port":8080,"liveness_path":"/health","readiness_path":"/ready","initial_delay_seconds":30},"env_vars":[],"category":"data-driven","requests_topic":"job.test-dispatch.requests","responses_topic":"job.test-dispatch.responses","parent_responses_topic":"system.agent.generic.responses","target_cluster":"uk_001","kafka_brokers":"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092","database_host":"pgbouncer.ai-persona-system.svc.cluster.local","database_port":"6432","database_user":"clients_user","database_name":"clients_db","dispatched_at":"$TIMESTAMP"}
JSON



then check:

# Spawner logs
kubectl -n ai-persona-system logs -l app=remote-job-spawner --tail=20

# Job created?
kubectl -n ai-persona-system get jobs | grep copywriter


--

The dispatch path works end to end:
message hits Kafka → spawner picks it up → K8s Job created in ~640ms.