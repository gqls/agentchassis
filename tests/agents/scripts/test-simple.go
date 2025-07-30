package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	bootstrapServers := "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

	fmt.Println("Starting content creator test...")

	// Create writer
	writer := &kafka.Writer{
		Addr:     kafka.TCP(bootstrapServers),
		Topic:    "system.agent.content-creator.process",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Create reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{bootstrapServers},
		Topic:     "system.responses.content-creator",
		GroupID:   "test-" + uuid.NewString(),
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	// Prepare message
	correlationID := uuid.NewString()
	fmt.Printf("Correlation ID: %s\n", correlationID)

	payload := map[string]interface{}{
		"action": "generate_content",
		"data": map[string]interface{}{
			"topic":        "The Future of AI in Healthcare",
			"content_type": "blog_post",
			"keywords":     []string{"AI", "healthcare", "innovation"},
			"length":       "medium",
			"style":        "informative",
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	// Send message
	ctx := context.Background()
	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(correlationID),
		Value: payloadBytes,
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(correlationID)},
			{Key: "request_id", Value: []byte(uuid.NewString())},
			{Key: "client_id", Value: []byte("demo_client")},
			{Key: "agent_instance_id", Value: []byte("test-instance")},
			{Key: "fuel_budget", Value: []byte("1000")},
		},
	})

	if err != nil {
		log.Fatal("Failed to send:", err)
	}

	fmt.Println("Message sent! Waiting for response...")

	// Read with timeout
	readCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		msg, err := reader.ReadMessage(readCtx)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		if string(msg.Key) == correlationID {
			var response map[string]interface{}
			json.Unmarshal(msg.Value, &response)

			fmt.Printf("\n=== RESPONSE ===\n")
			fmt.Printf("Success: %v\n", response["success"])

			if text, ok := response["generated_text"].(string); ok {
				fmt.Printf("\n--- Generated Content ---\n%s\n", text)
			}

			if metadata, ok := response["metadata"].(map[string]interface{}); ok {
				fmt.Printf("\n--- Metadata ---\n")
				for k, v := range metadata {
					fmt.Printf("%s: %v\n", k, v)
				}
			}
			break
		}
	}
}
