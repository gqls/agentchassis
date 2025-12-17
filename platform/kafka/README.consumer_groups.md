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


# Check what topics the pods are subscribed to
# Chief Strategist env
kubectl -n ai-persona-system exec agent-chief-strategist-e556f9ef-pb2p7 -- env | grep KAFKA_TOPICS

# Multipage env
kubectl -n ai-persona-system exec agent-multipage-website-builder-87a57b76-8rg2p -- env | grep KAFKA_TOPICS



kubectl -n ai-persona-system get pod agent-chassis-7556c89d7d-4cdtv -o yaml | grep -A5 "env:"

# actually consuming from right topic?
# Check chief strategist's RESPONSE consumer group
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh \
--bootstrap-server localhost:9092 \
--describe --group e556f9ef-c74f-437d-9291-bcd21636e357

# Check multipage's RESPONSE consumer group
kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
/opt/kafka/bin/kafka-consumer-groups.sh \
--bootstrap-server localhost:9092 \
--describe --group 87a57b76-0b5a-40b9-a11b-d8fd06da782a







groups:
git.adapter.group
1917fe4e-4be9-41a7-8b6f-f7654af2cea9
image-generator-adapter-group
chief-strategist-group-e556f9ef
87a57b76-0b5a-40b9-a11b-d8fd06da782a
generic-requests-group
briefing-agent-group-d58305be
multipage-website-builder-group-87a57b76
e556f9ef-c74f-437d-9291-bcd21636e357
8091fe06-6ed3-4600-be47-27b65fcfd946
site-classifier-group-1917fe4e
webscrape-adapter-group
d58305be-1187-40f6-9ea8-af3882cf4515


builder topic
job.a7d48f21-47867661-multipage-website-builder-spawn_builder.responses
job.a7d48f21-a4d089a2-chief-strategist-spawn_strategist.requests

3. Check strategist's consumer group status
   kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
   /opt/kafka/bin/kafka-consumer-groups.sh \
   --bootstrap-server localhost:9092 \
   --describe --group chief-strategist-group-e556f9ef

2. Check if there are messages in strategist's responses topic
   kubectl -n kafka exec -it personae-kafka-cluster-combined-pool-prod-0 -- \
   /opt/kafka/bin/kafka-console-consumer.sh \
   --bootstrap-server localhost:9092 \
   --topic job.a7d48f21-a4d089a2-chief-strategist-spawn_strategist.responses \
   --from-beginning --max-messages 5

