check if the consumer group already has a committed offset past the message. Try resetting:

# Check consumer groups
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list

# Check the offset for your consumer group
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
--group generic-request-consumers --describe

# Reset if needed
kubectl -n kafka exec -i personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
--group generic-request-consumers --reset-offsets --to-earliest \
--topic system.agent.generic.requests --execute


kubectl -n ai-persona-system get pod agent-chassis-7556c89d7d-4cdtv -o yaml | grep -A5 "env:"