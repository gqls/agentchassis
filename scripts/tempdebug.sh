#!/bin/bash
# FILE: scripts/debug-kafka-messages.sh
# Script to debug why messages aren't being processed

echo "=== Checking Kafka Topics ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-topics.sh --bootstrap-server localhost:9092 --list | grep -E "(system\.agent|system\.responses)"

echo ""
echo "=== Checking Messages in system.agent.generic.requests ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 \
  --topic system.agent.generic.requests \
  --from-beginning \
  --max-messages 5 \
  --timeout-ms 5000 2>/dev/null || echo "No messages or timeout"

echo ""
echo "=== Checking Consumer Groups ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list

echo ""
echo "=== Checking Consumer Group Lag ==="
AGENT_ID=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}' | cut -d'-' -f3-)
echo "Agent ID: $AGENT_ID"
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group "$AGENT_ID" 2>/dev/null || echo "Consumer group not found"

echo ""
echo "=== Checking if agent pod can reach Kafka ==="
kubectl exec -n ai-persona-system deployment/agent-chassis -- \
  nc -zv personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local 9092

echo ""
echo "=== Last 20 lines from agent logs ==="
kubectl logs -n ai-persona-system deployment/agent-chassis --tail=20

echo ""
echo "=== Sending a test message directly ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bash -c 'echo "{\"test\":\"message\"}" | bin/kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic system.agent.generic.requests'

echo "Waiting 3 seconds to check if test message is processed..."
sleep 3

echo ""
echo "=== Checking agent logs after test message ==="
kubectl logs -n ai-persona-system deployment/agent-chassis --tail=10
