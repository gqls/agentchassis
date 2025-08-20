#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Generate IDs
CID=$(cat /proc/sys/kernel/random/uuid)
RID=$(cat /proc/sys/kernel/random/uuid)

echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}Testing Agent Spawning (Complete Flow)${NC}"
echo -e "${GREEN}================================================${NC}"
echo "Correlation ID: $CID"
echo "Request ID: $RID"
echo ""

# Send the message
echo -e "${YELLOW}Sending spawn_group message...${NC}"
kubectl exec -i kafkacat-test-pod -n ai-persona-system -- sh << EOF
echo '{"action":"spawn_group","data":{"group_type":"website-builder"}}' | \
kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.process \
  -H "correlation_id=$CID" \
  -H "request_id=$RID" \
  -H "client_id=demo_client" \
  -H "user_id=test-user" \
  -H "agent_instance_id=00000000-0000-0000-0000-000000000001" \
  -H "fuel_budget=1000"
EOF

echo -e "${GREEN}✓ Message sent${NC}"
sleep 5

# Check orchestrator state
echo -e "${YELLOW}Checking orchestrator state:${NC}"
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "
SELECT
    correlation_id,
    status,
    current_step,
    execution_metadata->>'completed_steps' as completed,
    substring(error, 1, 50) as error_preview
FROM orchestrator_state
WHERE correlation_id = '$CID';"

# Check for spawned jobs
echo -e "${YELLOW}Checking for spawned Kubernetes jobs:${NC}"
kubectl get jobs -n ai-persona-system -l spawned-by=orchestrator 2>/dev/null || echo "No jobs found"

# Check if agents were created in database
echo -e "${YELLOW}Checking spawned agents in database:${NC}"
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c "
SELECT id, name, config->>'agent_type' as type
FROM client_demo_client.agent_instances
WHERE created_at > NOW() - INTERVAL '1 minute';"

echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}Test Complete${NC}"
echo -e "${GREEN}================================================${NC}"