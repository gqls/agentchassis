#!/bin/bash

# Test Workflow Script
echo "=== Testing Agent Workflow Execution ==="

# Generate correlation ID for tracking
CID=$(cat /proc/sys/kernel/random/uuid)
echo "Using Correlation ID: $CID"

# Step 1: Send a simple validate_input message with all required headers
echo "Step 1: Testing simple validate_input action..."
kubectl exec -it kafkacat-test-pod -n ai-persona-system -- sh -c "
echo '{\"action\":\"validate_input\",\"data\":{\"message\":\"test workflow execution\"}}' | \
kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.process \
  -H 'correlation_id:$CID' \
  -H 'request_id:$(cat /proc/sys/kernel/random/uuid)' \
  -H 'client_id:demo_client' \
  -H 'agent_instance_id:00000000-0000-0000-0000-000000000001' \
  -H 'fuel_budget:1000'"

echo "Message sent. Correlation ID: $CID"

# Step 2: Wait and check logs
echo "Step 2: Checking agent logs..."
sleep 3
kubectl logs -n ai-persona-system deployment/agent-chassis --since=30s | grep -E "$CID|ProcessMessage|Action extracted|Workflow|ExecuteWorkflow" | tail -20

# Step 3: Check orchestrator state in database
echo "Step 3: Checking orchestrator state..."
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "
SELECT correlation_id, status, current_step,
       execution_metadata->>'completed_steps' as completed,
       created_at, updated_at
FROM orchestrator_state
WHERE correlation_id = '$CID'::uuid;"

echo ""
echo "=== Step 4: Testing spawn_group action ==="

# Generate new correlation ID
CID2=$(cat /proc/sys/kernel/random/uuid)
echo "Using Correlation ID: $CID2"

kubectl exec -it kafkacat-test-pod -n ai-persona-system -- sh -c "
echo '{\"action\":\"spawn_group\",\"data\":{\"group_type\":\"website-builder\",\"business_name\":\"Test Company\",\"domain\":\"test.com\"}}' | \
kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.process \
  -H 'correlation_id:$CID2' \
  -H 'request_id:$(cat /proc/sys/kernel/random/uuid)' \
  -H 'client_id:demo_client' \
  -H 'agent_instance_id:00000000-0000-0000-0000-000000000001' \
  -H 'fuel_budget:1000'"

echo "Spawn group message sent. Correlation ID: $CID2"

# Wait and check logs for spawn_group
echo "Step 5: Checking spawn_group logs..."
sleep 5
kubectl logs -n ai-persona-system deployment/agent-chassis --since=60s | grep -E "$CID2|spawn_group|SpawnGroupAction|agents spawned" | tail -20

# Check database for both workflows
echo "Step 6: Checking all workflow states..."
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "
SELECT correlation_id, status, current_step,
       execution_metadata->>'completed_steps' as completed,
       created_at, updated_at
FROM orchestrator_state
WHERE correlation_id IN ('$CID'::uuid, '$CID2'::uuid)
ORDER BY created_at DESC;"

# Check agent instances created
echo "Step 7: Checking agent instances..."
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "
SELECT id, name, config->>'agent_type' as type, is_active, created_at
FROM client_demo_client.agent_instances
WHERE created_at > NOW() - INTERVAL '10 minutes'
ORDER BY created_at DESC;"

# Check Kubernetes jobs spawned
echo "Step 8: Checking Kubernetes jobs..."
kubectl get jobs -n ai-persona-system -l spawned-by=orchestrator --sort-by='.metadata.creationTimestamp' | tail -10

echo ""
echo "=== Test Complete ==="
echo "Correlation IDs used:"
echo "  Simple workflow: $CID"
echo "  Spawn group: $CID2"