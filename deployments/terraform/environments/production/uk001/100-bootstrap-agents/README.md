# Reset consumer group to latest (skip all old messages)
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- /opt/kafka/bin/kafka-consumer-groups.sh \
--bootstrap-server localhost:9092 \
--group generic-orchestrator-group \
--topic system.agent.generic.process \
--reset-offsets \
--to-latest \
--execute

# Or delete and recreate the topic (more drastic)
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- /opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--delete \
--topic system.agent.generic.process

kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- /opt/kafka/bin/kafka-topics.sh \
--bootstrap-server localhost:9092 \
--create \
--topic system.agent.generic.process \
--partitions 3 \
--replication-factor 1

--

Looking at the logs, everything is working correctly:
✅ Configuration loaded successfully - Database configs are populated correctly
✅ Database connections established - Both clients_db and templates_db connected
✅ Kafka consumers and producers created - Connected to the Kafka cluster
✅ Health and metrics servers started - Running on ports 8080 and 9090
✅ Message consumption loop started - Ready to process messages
The key fix was mounting the custom config to a different path (/app/custom-configs/) and explicitly telling the application to use it with the command override. This avoided the conflict with the default config file built into the Docker image.
Now your generic orchestrator is:

Running as a bootstrap agent
Listening on topic system.agent.generic.process
Ready to orchestrate other agents
Connected to all required infrastructure

You can verify it's healthy by checking:
# Check health endpoint
kubectl -n ai-persona-system exec generic-orchestrator-0 -- wget -q -O- http://localhost:8080/health

# Check metrics
kubectl -n ai-persona-system exec generic-orchestrator-0 -- wget -q -O- http://localhost:9090/metrics | head -20

# Check if it's processing messages (if any are in the topic)
kubectl -n ai-persona-system logs generic-orchestrator-0 -f

The agent is now ready to:

Receive work requests from the Kafka topic
Process them according to the orchestration rules
Spawn specialized agents as Kubernetes jobs when needed
Coordinate saga workflows


3. Send a properly formatted test message
   Let's send a valid message to test the agent:
# Create a test message file
cat > /tmp/test-message.json <<EOF
{
"action": "test",
"workflow_id": "test-workflow-001",
"client_id": "demo_client",
"payload": {
"message": "Test message for generic orchestrator"
}
}
EOF

# Send the message
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- /opt/kafka/bin/kafka-console-producer.sh \
--bootstrap-server localhost:9092 \
--topic system.agent.generic.process \
--property "parse.headers=true" \
--property "headers.separator=|" < /tmp/test-message.json

kubectl -n ai-persona-system logs generic-orchestrator-0 -f --tail=20