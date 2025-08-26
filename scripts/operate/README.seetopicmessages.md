kubectl exec -it -n kafka personae-kafka-cluster-combined-pool-prod-0 -- bash

# Once inside, you can use the built-in Kafka console consumer which is already available:
# Inside the pod, use kafka-console-consumer
/opt/kafka/bin/kafka-console-consumer.sh \
--bootstrap-server localhost:9092 \
--topic system.agent.website-builder.requests \
--from-offset end \
--max-messages 5 \
--property print.headers=true \
--property print.key=true \
--property print.timestamp=true \
--property print.offset=true


system.agent.content-creator.process
system.agent.content-creator.requests
system.agent.content-creator.responses
system.agent.content-researcher.dlq
system.agent.content-researcher.errors
system.agent.content-researcher.process
system.agent.content-researcher.requests
system.agent.content-researcher.responses
system.agent.content_researcher.process
system.agent.copywriter.process
system.agent.domain-analyst.dlq
system.agent.domain-analyst.errors
system.agent.domain-analyst.process
system.agent.domain-analyst.requests
system.agent.domain-analyst.responses
system.agent.generic.dlq
system.agent.generic.errors
system.agent.generic.process
system.agent.generic.requests
system.agent.generic.responses
system.agent.html-developer.dlq
system.agent.html-developer.errors
system.agent.html-developer.process
system.agent.html-developer.requests
system.agent.html-developer.responses
system.agent.image-generator.process
system.agent.reasoning.process
system.agent.researcher.process
system.agent.site-architect.dlq
system.agent.site-architect.errors
system.agent.site-architect.process
system.agent.site-architect.requests
system.agent.site-architect.responses
system.agent.site-publisher.dlq
system.agent.site-publisher.errors
system.agent.site-publisher.process
system.agent.site-publisher.requests
system.agent.site-publisher.responses
system.agent.visual-designer.dlq
system.agent.visual-designer.errors
system.agent.visual-designer.process
system.agent.visual-designer.requests
system.agent.visual-designer.responses
system.agent.web-search.process
system.agent.website-builder.dlq
system.agent.website-builder.errors
system.agent.website-builder.process
system.agent.website-builder.requests
system.agent.website-builder.responses


----------------------
to delete messages delete and recreate topic

kubectl exec -it -n kafka personae-kafka-cluster-combined-pool-prod-0 -- bash

# Delete the topic
/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --delete --topic system.agent.website-builder.requests

# Recreate it
/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --create --topic system.agent.website-builder.requests --partitions 3 --replication-factor 3

# or

kubectl exec -it -n kafka personae-kafka-cluster-combined-pool-prod-0 -- bash

# Set retention to 1ms
/opt/kafka/bin/kafka-configs.sh --bootstrap-server localhost:9092 --entity-type topics --entity-name system.agent.website-builder.requests --alter --add-config retention.ms=1

# Wait a few seconds for cleanup
sleep 10

# Reset retention back to default (7 days = 604800000ms)
/opt/kafka/bin/kafka-configs.sh --bootstrap-server localhost:9092 --entity-type topics --entity-name system.agent.website-builder.requests --alter --add-config retention.ms=604800000


-------
# run a debug pod

kubectl run -it --rm kafka-debug -n kafka \
--image=confluentinc/cp-kafka:latest \
--restart=Never -- \
kafka-console-consumer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.agent.website-builder.requests \
--from-beginning \
--max-messages 5 \
--property print.headers=true \
--property print.key=true

--------------------
# Or if you want to install kafkacat in the pod temporarily:
# Inside the pod
apt-get update && apt-get install -y kafkacat

# Then use kafkacat
kafkacat -b localhost:9092 -t system.agent.website-builder.requests -C \
-f '\nKey: %k\nHeaders: %h\nOffset: %o\nTimestamp: %T\nPayload: %s\n---\n' \
-o end -c 5






You can use kafkacat (also known as kcat) to inspect Kafka topics. Here are the commands:
To consume messages from a topic with headers and metadata:

# View messages from website-builder requests topic
kafkacat -b localhost:9092 -t system.agent.website-builder.requests -C -f '\nKey: %k\nHeaders: %h\nOffset: %o\nPartition: %p\nTimestamp: %T\nPayload: %s\n---\n' -o end -c 5

# View messages from website-builder responses topic
kafkacat -b localhost:9092 -t system.agent.website-builder.responses -C -f '\nKey: %k\nHeaders: %h\nOffset: %o\nPartition: %p\nTimestamp: %T\nPayload: %s\n---\n' -o end -c 5

Format options explained:

-C: Consumer mode
-f: Format string

%k: Key
%h: Headers (will show all headers with their values)
%o: Offset
%p: Partition
%T: Timestamp
%s: Message payload


-o end: Start from end of topic (use -o beginning for all messages)
-c 5: Consume only 5 messages (remove to consume continuously)

To see just the latest message with pretty-printed JSON:

# Latest message with JSON pretty-printing
kafkacat -b localhost:9092 -t system.agent.website-builder.requests -C -o -1 -c 1 -f 'Headers: %h\nPayload:\n' -e | jq '.'

If you're in Kubernetes, you can run kafkacat in a pod:

kubectl run -it --rm debug-kafka --image=confluentinc/cp-kafkacat:latest --restart=Never -- \
kafkacat -b personae-kafka-cluster-combined-pool-prod-0 \
-t system.agent.website-builder.requests \
-C -f '\nHeaders: %h\nPayload: %s\n---\n' -o end -c 5

Replace kafka-broker.ai-persona-system.svc.cluster.local:9092 with your actual Kafka broker address.