Testing Agent Communication

Here’s a sample test run for your web-search-adapter:

    Start a temporary pod with Kafka tools inside your cluster.

kubectl run -it --rm kafka-tools -n kafka --image=confluentinc/cp-kafka:latest -- bash

From the new pod's shell, send a test message. We'll send a message to the system.adapter.web.search topic, which is what the web-search-adapter is listening on.

# Run this inside the kafka-tools pod
kafka-console-producer --bootstrap-server personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092 --topic system.adapter.web.search

The producer will wait for input. Paste a JSON message like the one below and press Enter. This simulates a request to search the web.

{"request_id": "test-123", "query": "What is Gemini?"}

Listen for the result. In another terminal, start another Kafka tools pod and use the kafka-console-consumer to listen on a topic where you expect to see results, for example system.events.

# In a second terminal
kubectl run -it --rm kafka-consumer -n kafka --image=confluentinc/cp-kafka:latest -- bash

# Inside the new consumer pod
kafka-console-consumer --bootstrap-server personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092 --topic system.events --from-beginning

## 1. Check the Topic's Status

From inside your kafka-tools pod, run the kafka-topics command with the --describe flag. This will give us a detailed report on the topic's health.

# Run this inside the kafka-tools pod
kafka-topics --bootstrap-server personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092 --describe --topic system.adapter.web.search

In the output, look at the Leader and Isr (In-Sync Replicas) columns. You will likely see that the Leader is none, which confirms the issue.

2. Recreate the Topic with the Correct Replication Factor

The simplest fix is to delete and recreate the topic with a replication-factor of 1, which is suitable for a single-broker cluster.

# Run these inside the kafka-tools pod

# First, delete the misconfigured topic
kafka-topics --bootstrap-server personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092 --delete --topic system.adapter.web.search

# Second, recreate it with replication factor 1
kafka-topics --bootstrap-server personae-kafka-cluster-kafka-bootstrap.kafka.svc:9092 --create --topic system.adapter.web.search --partitions 1 --replication-factor 1

