# List all topics
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--list | grep response

# List topics without the newline issueue
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--list | grep "system.responses" | tr -d '\r'


# Check what's in each response topic
for topic in $(kubectl exec -it kafka-client-test -n kafka -- kafka-topics --bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 --list | grep response); do
echo "=== Topic: $topic ==="
kubectl exec -it kafka-client-test -n kafka -- kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic $topic \
--from-beginning \
--max-messages 2 \
--timeout-ms 5000 2>/dev/null || echo "No messages"
done
