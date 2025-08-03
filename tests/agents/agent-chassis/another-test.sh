cat > test.go << 'EOL'

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
			{Key: "correlation_id", Value: []byte("123e4567-e89b-12d3-a456-426614174000")},
			{Key: "request_id", Value: []byte("223e4567-e89b-12d3-a456-426614174000")},
			{Key: "client_id", Value: []byte("demo_client")},
			{Key: "agent_instance_id", Value: []byte("00000000-0000-0000-0000-000000000001")},
			{Key: "fuel_budget", Value: []byte("100")},
		},
		Key:   []byte("test-key-3"),
		Value: []byte(`{"action": "test_action", "data": {"message": "Hello with proper UUIDs"}}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Fatal("failed to write message:", err)
	}

	fmt.Println("Message sent with valid UUIDs!")
}
EOL