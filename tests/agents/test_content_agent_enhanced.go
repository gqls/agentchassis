package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/governance"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/logger"
	"go.uber.org/zap"
)

func main() {
	// Configuration
	brokers := []string{"localhost:9092"}
	if os.Getenv("KAFKA_BROKERS") != "" {
		brokers = strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	}

	clientID := "demo_client"
	agentInstanceID := "123e4567-e89b-12d3-a456-426614174001"
	userID := "123e4567-e89b-12d3-a456-426614174000" // mysql auth db

	appLogger, _ := logger.New("debug")
	defer appLogger.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Setup Kafka
	producer, err := kafka.NewProducer(brokers, appLogger)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	consumer, err := kafka.NewConsumer(brokers, "system.responses.content-creator", "test-consumer-"+uuid.NewString(), appLogger)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Test different content types
	testCases := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "Blog Post",
			payload: map[string]interface{}{
				"action": "generate_blog_post",
				"data": map[string]interface{}{
					"topic":           "The Future of AI in Healthcare",
					"prompt":          "Focus on recent breakthroughs and ethical considerations",
					"content_type":    "blog_post",
					"keywords":        []string{"AI", "healthcare", "medical diagnosis", "patient care"},
					"length":          "medium",
					"style":           "informative",
					"tone":            "professional",
					"target_audience": "Healthcare professionals and tech enthusiasts",
					"format":          "markdown",
				},
			},
		},
		{
			name: "Product Description",
			payload: map[string]interface{}{
				"action": "generate_product_description",
				"data": map[string]interface{}{
					"topic":           "Smart Home Security Camera",
					"content_type":    "product_description",
					"keywords":        []string{"security", "AI-powered", "night vision", "mobile app"},
					"length":          "short",
					"style":           "persuasive",
					"tone":            "friendly",
					"target_audience": "Homeowners concerned about security",
					"context": map[string]interface{}{
						"price":    "$149.99",
						"features": []string{"4K resolution", "AI motion detection", "Two-way audio"},
					},
				},
			},
		},
		{
			name: "Social Media Post",
			payload: map[string]interface{}{
				"action": "generate_social_media",
				"data": map[string]interface{}{
					"topic":        "Announcing our new AI Writing Assistant",
					"content_type": "social_media",
					"platform":     "linkedin",
					"length":       "short",
					"style":        "professional",
					"tone":         "enthusiastic",
					"keywords":     []string{"AI", "productivity", "writing", "innovation"},
					"max_length":   1300,
				},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		fmt.Printf("\n=== Testing: %s ===\n", tc.name)

		correlationID := uuid.NewString()
		requestID := uuid.NewString()

		payloadBytes, _ := json.Marshal(tc.payload)

		headers := map[string]string{
			"correlation_id":      correlationID,
			"request_id":          requestID,
			"client_id":           clientID,
			"user_id":             userID,
			"agent_instance_id":   agentInstanceID,
			governance.FuelHeader: "1000",
		}

		// Send message
		appLogger.Info("Sending request", zap.String("test_case", tc.name))
		if err := producer.Produce(ctx, "system.agent.content-creator.process", headers,
			[]byte(correlationID), payloadBytes); err != nil {
			log.Printf("Failed to send message: %v", err)
			continue
		}

		// Wait for response
		responseCtx, responseCancel := context.WithTimeout(ctx, 15*time.Second)
		defer responseCancel()

		msg, err := consumer.FetchMessage(responseCtx)
		if err != nil {
			log.Printf("Failed to receive response: %v", err)
			continue
		}

		// Process response
		var response map[string]interface{}
		if err := json.Unmarshal(msg.Value, &response); err != nil {
			log.Printf("Failed to unmarshal response: %v", err)
		} else {
			if response["success"].(bool) {
				fmt.Println("\n--- GENERATED CONTENT ---")
				fmt.Printf("%s\n", response["generated_text"])
				fmt.Println("\n--- METADATA ---")
				if metadata, ok := response["metadata"].(map[string]interface{}); ok {
					fmt.Printf("Tokens Used: %v\n", metadata["tokens_used"])
					fmt.Printf("Word Count: %v\n", metadata["word_count"])
					fmt.Printf("Memories Used: %v\n", metadata["memories_used"])
					fmt.Printf("Model Used: %v\n", metadata["model_used"])
					fmt.Printf("Generation Time: %v seconds\n", metadata["generation_time"])
				}
			} else {
				fmt.Printf("Error: %v\n", response["error"])
			}
		}

		consumer.CommitMessages(context.Background(), msg)

		// Small delay between tests
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\nAll tests completed!")
}
