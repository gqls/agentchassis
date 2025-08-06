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

# Check if workflows are working from database
# Check if workflows were created
kubectl exec -it postgres-0 -n postgres -- psql -U dbuser -d clients_db -c "SELECT correlation_id, status, current_step, created_at FROM orchestrator_state ORDER BY created_at DESC LIMIT 5;"