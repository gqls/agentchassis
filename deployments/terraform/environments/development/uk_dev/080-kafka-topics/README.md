List topics:
error 137 is probably memory

# Execute directly in the Kafka broker pod
kubectl exec -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-topics.sh --bootstrap-server localhost:9092 --list
audit.log
system.commands.workflow.resume
system.errors
system.events
system.events.workflow.completed
system.events.workflow.paused

# Terminal 1 - Start Consumer:
kubectl exec -it -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 \
--topic test-messages --from-beginning

# Terminal 2 - Start Producer:
kubectl exec -it -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 \
--topic test-messages --from-beginning

# Create test topic
kubectl exec -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-topics.sh --bootstrap-server localhost:9092 \
--create --topic test-messages --partitions 3 --replication-factor 1

# test consumer groups
# Consume as part of a consumer group
kubectl exec -it kafka-client -n kafka -- \
kafkacat -b personae-kafka-cluster-kafka-bootstrap:9092 \
-t test-messages -G my-consumer-group -f '%s\n'

# test json messages
# Send JSON message
kubectl exec -n kafka personae-kafka-cluster-kafka-0 -c kafka -- sh -c '
echo '\''{"id": 1, "type": "agent-event", "timestamp": "2024-01-15T10:00:00Z", "data": "test"}'\'' | \
bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic personae-agent-events
'

# Consume it
kubectl exec -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 \
--topic personae-agent-events --from-beginning --max-messages 1

6. Check consumer groups:
7. # List consumer groups
kubectl exec -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list

# Create a consumer with a group
kubectl exec -it -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 \
--topic test-messages --group test-consumer-group --from-beginning

7. Performance test:
# Run a simple performance test
kubectl exec -n kafka personae-kafka-cluster-kafka-0 -c kafka -- \
bin/kafka-producer-perf-test.sh --topic test-messages \
--num-records 1000 --record-size 100 --throughput 100 \
--producer-props bootstrap.servers=localhost:9092