# Using kubectl exec into a Kafka pod - List all Kafka topics:
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--list

# 2. Check specific topic details:
# Replace <topic-name> with one of your topics
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--describe \
--topic <topic-name>

# 3. Test producing and consuming messages:
#   Produce a test message:
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- bin/kafka-console-producer.sh \
   --bootstrap-server localhost:9092 \
   --topic <topic-name>
# Type a message and press Enter, then Ctrl+C to exit

# Consume messages:
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- bin/kafka-console-consumer.sh \
--bootstrap-server localhost:9092 \
--topic <topic-name> \
--from-beginning

# 4. Check the job logs to see what was created:
kubectl logs -n kafka kafka-system-topics-init-ea6cbe30-5h5vq

# 5. Using Kafka's built-in tools via port-forward:
# Port forward to access Kafka from your local machine
kubectl port-forward -n kafka svc/personae-kafka-cluster-kafka-bootstrap 9092:9092

# In another terminal, if you have Kafka tools installed locally:
kafka-topics --bootstrap-server localhost:9092 --list

# 6. Check Strimzi Kafka Topic resources (if using Strimzi):
kubectl get kafkatopic -n kafka

# Expected topics might include:
# agent.requests
# agent.responses
# orchestrator.events
# workflow.executions