package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func main() {
	// Use the Strimzi bootstrap service
	bootstrapServers := "personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"

	// Check if we're inside the cluster
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		// Outside cluster - use port forwarding
		bootstrapServers = "localhost:9092"
		fmt.Println("Running outside cluster, using localhost:9092")
	} else {
		fmt.Println("Running inside cluster, using Kafka bootstrap service")
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create Kafka writer
	writer := &kafka.Writer{
		Addr:     kafka.TCP(bootstrapServers),
		Topic:    "system.agent.content-creator.process",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Create reader for responses
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{bootstrapServers},
		Topic:   "system.responses.content-creator",
		GroupID: "test-consumer-" + uuid.NewString(),
	})
	defer reader.Close()

	// Test payload
	correlationID := uuid.NewString()
	payload := map[string]interface{}{
		"action": "generate_blog_post",
		"data": map[string]interface{}{
			"topic":           "The Future of AI in Healthcare",
			"content_type":    "blog_post",
			"keywords":        []string{"AI", "healthcare", "innovation"},
			"length":          "medium",
			"style":           "informative",
			"tone":            "professional",
			"target_audience": "Healthcare professionals",
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	// Send message
	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(correlationID),
		Value: payloadBytes,
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(correlationID)},
			{Key: "request_id", Value: []byte(uuid.NewString())},
			{Key: "client_id", Value: []byte("demo_client")},
			{Key: "agent_instance_id", Value: []byte("123e4567-e89b-12d3-a456-426614174001")},
			{Key: "fuel_budget", Value: []byte("1000")},
		},
	})

	if err != nil {
		log.Fatal("Failed to send message:", err)
	}

	fmt.Println("Message sent successfully! Waiting for response...")

	// Read response
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if string(msg.Key) == correlationID {
			var response map[string]interface{}
			json.Unmarshal(msg.Value, &response)

			fmt.Printf("\nResponse received:\n")
			fmt.Printf("Success: %v\n", response["success"])

			if response["success"] == true {
				fmt.Printf("\n--- Generated Content ---\n%s\n", response["generated_text"])

				if metadata, ok := response["metadata"].(map[string]interface{}); ok {
					fmt.Printf("\n--- Metadata ---\n")
					for k, v := range metadata {
						fmt.Printf("%s: %v\n", k, v)
					}
				}
			} else {
				fmt.Printf("Error: %v\n", response["error"])
			}

			reader.CommitMessages(ctx, msg)
			break
		}
	}
}
