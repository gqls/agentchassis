#!/bin/bash
# FILE: scripts/test-kafka-consumer.sh
# Test if messages are actually in the topic

echo "=== Checking if messages exist in the topic ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-run-class.sh kafka.tools.GetOffsetShell \
  --broker-list localhost:9092 \
  --topic system.agent.generic.requests

echo ""
echo "=== Checking last 3 messages in the topic ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.agent.generic.requests \
  --max-messages 3 \
  --property print.key=true \
  --property print.headers=true \
  --property print.timestamp=true \
  --from-beginning \
  --timeout-ms 5000 2>&1

echo ""
echo "=== Checking consumer group details ==="
# Get the agent pod name/ID
AGENT_POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
AGENT_ID=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].spec.containers[0].env[?(@.name=="AGENT_ID")].value}')

# If AGENT_ID is empty, try using the pod name
if [ -z "$AGENT_ID" ]; then
    AGENT_ID="be288537-69e2-4505-89f9-1ce6cad9456e"  # From your logs
fi

echo "Agent Pod: $AGENT_POD"
echo "Agent ID (consumer group): $AGENT_ID"

kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group "$AGENT_ID" \
  --describe 2>&1

echo ""
echo "=== Checking if agent is actually calling Consume() ==="
kubectl logs -n ai-persona-system deployment/agent-chassis --tail=50 | grep -E "(Consume|consume|DEBUG|processRequests)"