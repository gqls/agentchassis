https://claude.ai/chat/2d57dc36-1686-4021-8382-919342d4fa6e

To ensure the content creator only processes the intended messages, here are several strategies:
1. Clear the Queue Before Starting
   Always consume all old messages before starting the agent:

# Clear all messages from the topic using the agent's consumer group
kubectl exec -it kafka-client-test -n kafka -- kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.agent.content-creator.process \
--group content-creator-agent-group \
--from-beginning \
--timeout-ms 5000

# Verify the queue is empty
kubectl exec -it kafka-client-test -n kafka -- kafka-consumer-groups \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--group content-creator-agent-group \
--describe

2. Use a Different Topic for Testing
   Create a test-specific topic:
# Create test topic
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--create \
--topic system.agent.content-creator.test \
--partitions 1 \
--replication-factor 2

# Configure agent to use test topic (in config)

3. Send Messages with Unique Consumer Groups
   When testing, use a unique consumer group each time:
# In your agent config, use a timestamp-based group
kafka_consumer_group: "content-creator-agent-group-$(date +%s)"

4. Implement Message Filtering
   Add headers to identify test vs production messages:
   // In your test message
   Headers: []kafka.Header{
   {Key: "environment", Value: []byte("test")},
   {Key: "max_messages", Value: []byte("1")},
   // ... other headers
   }
   Then in the agent, check these headers before processing.
5. Use Single Partition for Testing
   Send test messages to a specific partition:
# Create single-partition test topic
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--create \
--topic system.agent.content-creator.singletest \
--partitions 1 \
--replication-factor 1

6. Best Practice: Test Workflow

Stop the agent
Check what's in the queue
Clear if needed
Send your test message
Start agent with replicas=1
Monitor logs
Stop agent after processing

# Full safe test sequence
kubectl scale deployment content-creator-agent -n ai-persona-system --replicas=0

# Check queue
kubectl exec -it kafka-client-test -n kafka -- kafka-consumer-groups \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--group content-creator-agent-group \
--describe

# Send ONE test message
# ... your test script ...

# Start with single replica
kubectl scale deployment content-creator-agent -n ai-persona-system --replicas=1

# Watch logs
kubectl logs -f -n ai-persona-system deployment/content-creator-agent

# Stop after seeing completion
kubectl scale deployment content-creator-agent -n ai-persona-system --replicas=0