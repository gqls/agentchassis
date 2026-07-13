#!/bin/bash
# test-website-build.sh

CID=$(cat /proc/sys/kernel/random/uuid)
RID=$(cat /proc/sys/kernel/random/uuid)

echo "Testing full website build workflow"
echo "CID: $CID"

# Send a website build request
kubectl exec -i kafkacat-test-pod -n ai-persona-system -- sh << EOF
echo '{
  "action": "build_website",
  "data": {
    "business_name": "Acme Corp",
    "domain": "acme.com",
    "description": "A technology company"
  }
}' | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.website-builder.process \
  -H "correlation_id=$CID" \
  -H "request_id=$RID" \
  -H "client_id=demo_client" \
  -H "agent_instance_id=9bb17944-0000-0000-0000-000000000001" \
  -H "fuel_budget=5000"
EOF

echo "Waiting for processing..."
sleep 10

# Check the workflow progress
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "
SELECT
    correlation_id,
    status,
    current_step,
    execution_metadata->>'completed_steps' as completed,
    execution_metadata->>'total_steps' as total
FROM orchestrator_state
WHERE correlation_id = '$CID';"