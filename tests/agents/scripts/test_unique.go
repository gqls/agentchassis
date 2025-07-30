package main

// underwater basket weaving

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	bootstrapServers := "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
	correlationID := uuid.NewString()
	uniqueTopic := "The Amazing World of Underwater Basket Weaving" // Very unique topic

	fmt.Printf("Sending test with correlation ID: %s\n", correlationID)
	fmt.Printf("Topic: %s\n", uniqueTopic)

	writer := &kafka.Writer{
		Addr:     kafka.TCP(bootstrapServers),
		Topic:    "system.agent.content-creator.process",
		Balancer: &kafka.LeastBytes{},
	}

	request := map[string]interface{}{
		"action": "generate_content",
		"data": map[string]interface{}{
			"topic":        uniqueTopic,
			"content_type": "blog_post",
			"length":       "short",
			"style":        "humorous",
		},
	}

	reqBytes, _ := json.Marshal(request)
	writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(correlationID),
		Value: reqBytes,
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(correlationID)},
			{Key: "fuel_budget", Value: []byte("1000")},
			{Key: "client_id", Value: []byte("demo_client")},
			{Key: "agent_instance_id", Value: []byte("test-instance")},
			{Key: "request_id", Value: []byte(uuid.NewString())},
		},
	})

	fmt.Printf("\nNow check Kafka for a blog post about: %s\n", uniqueTopic)
	fmt.Printf("With correlation ID: %s\n", correlationID)
}
