kubectl -n ai-persona-system exec -it agent-chassis-test -- sh



cd /app
cat << 'EOF' > /test-agent-chassis.go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

func main() {
    writer := kafka.NewWriter(kafka.WriterConfig{
        Brokers: []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"},
        Topic:   "system.agent.generic.process",
    })
    defer writer.Close()

    msg := kafka.Message{
        Headers: []kafka.Header{
            {Key: "correlation_id", Value: []byte("test-123")},
            {Key: "request_id", Value: []byte("req-456")},
            {Key: "client_id", Value: []byte("demo_client")},
            {Key: "agent_instance_id", Value: []byte("test-instance")},
            {Key: "fuel_budget", Value: []byte("100")},
        },
        Key:   []byte("test-key"),
        Value: []byte(`{"action": "test_action", "data": {"message": "Hello from test"}}`),
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err := writer.WriteMessages(ctx, msg)
    if err != nil {
        log.Fatal("failed to write message:", err)
    }

    fmt.Println("Message sent successfully!")
}
EOF

    go mod init test && \
    go mod tidy && \
    go get github.com/segmentio/kafka-go && \
    go run test-agent-chassis.go


##
# error topic
kubectl exec -it kafka-client-test -n kafka -- kafka-console-consumer   --bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092   --topic system.errors.generic   --from-beginning   --max-messages 1
# process topic
kubectl exec -it kafka-client-test -n kafka -- kafka-console-consumer   --bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092   --topic system.agent.generic.process   --from-beginning   --max-messages 1

# what topics exist
kubectl exec -it kafka-client-test -n kafka -- kafka-topics \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--list | grep generic

system.agent.generic.process
system.errors.generic


kubectl exec -it kafka-client-test -n kafka -- kafka-console-producer \
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092 \
--topic system.agent.generic.process \
--property "parse.headers=true" \
--property "headers.separator=|"

kafka-console-consumer
--bootstrap-server personae-kafka-cluster-kafka-bootstrap:9092
--topic system.errors.generic
--from-beginning
--max-messages 1


