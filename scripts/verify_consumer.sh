#!/bin/bash
# FILE: scripts/verify-consumer.sh
# Check if the consumer is actually joining the group

echo "=== List all consumer groups ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list

echo ""
echo "=== Check for stale consumer groups ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list | while read group; do
    echo "Group: $group"
    kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
      bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
      --group "$group" --describe 2>/dev/null | head -5
    echo "---"
done

echo ""
echo "=== Force reset consumer group offset ==="
AGENT_ID="be288537-69e2-4505-89f9-1ce6cad9456e"
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --group "$AGENT_ID" \
  --topic system.agent.generic.requests \
  --reset-offsets --to-earliest --execute 2>&1

echo ""
echo "=== Create a test consumer to verify connectivity ==="
timeout 5 kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 \
  --topic system.agent.generic.requests \
  --group test-consumer-group \
  --from-beginning \
  --max-messages 1 2>&1 || echo "Timeout reached"

echo ""
echo "=== Check if test consumer group was created ==="
kubectl exec -n kafka personae-kafka-cluster-combined-pool-prod-0 -c kafka -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --group test-consumer-group --describe