// test/scripts/test_individual_agent.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var (
	agentType = flag.String("agent", "", "Agent type to test")
	topic     = flag.String("topic", "", "Override topic")
	payload   = flag.String("payload", "", "JSON payload file")
)

type TestConfig struct {
	Agent   string
	Topic   string
	Payload interface{}
	Timeout time.Duration
}

var agentConfigs = map[string]TestConfig{
	"content-creator": {
		Agent: "content-creator",
		Topic: "system.agent.content-creator.process",
		Payload: map[string]interface{}{
			"action": "generate_content",
			"data": map[string]interface{}{
				"topic":        "Test Content Generation",
				"content_type": "blog_post",
				"style":        "informative",
			},
		},
		Timeout: 30 * time.Second,
	},
	"reasoning": {
		Agent: "reasoning",
		Topic: "system.agent.reasoning.process",
		Payload: map[string]interface{}{
			"action": "analyze",
			"data": map[string]interface{}{
				"content_to_review": "Test content for analysis",
				"review_criteria":   []string{"clarity", "accuracy"},
			},
		},
		Timeout: 20 * time.Second,
	},
	"web-search": {
		Agent: "web-search",
		Topic: "system.agent.web-search.process",
		Payload: map[string]interface{}{
			"action": "search",
			"data": map[string]interface{}{
				"query":       "kubernetes best practices",
				"max_results": 5,
			},
		},
		Timeout: 15 * time.Second,
	},
}

func main() {
	flag.Parse()

	if *agentType == "" {
		log.Fatal("Please specify agent type with -agent flag")
	}

	config, ok := agentConfigs[*agentType]
	if !ok {
		log.Fatalf("Unknown agent type: %s", *agentType)
	}

	if *topic != "" {
		config.Topic = *topic
	}

	testAgent(config)
}

func testAgent(config TestConfig) {
	brokers := []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}

	// Create writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   config.Topic,
	})
	defer writer.Close()

	// Create reader for responses
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     fmt.Sprintf("system.responses.%s", config.Agent),
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	// Prepare message
	correlationID := uuid.New().String()
	payloadBytes, _ := json.Marshal(config.Payload)

	fmt.Printf("Testing %s agent\n", config.Agent)
	fmt.Printf("Correlation ID: %s\n", correlationID)
	fmt.Printf("Payload: %s\n", string(payloadBytes))

	// Send message
	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(correlationID),
		Value: payloadBytes,
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(correlationID)},
			{Key: "request_id", Value: []byte(uuid.New().String())},
			{Key: "client_id", Value: []byte("test-client")},
			{Key: "agent_instance_id", Value: []byte("test-instance")},
			{Key: "fuel_budget", Value: []byte("1000")},
		},
	})

	if err != nil {
		log.Fatal("Failed to send message:", err)
	}

	fmt.Println("Message sent, waiting for response...")

	// Wait for response
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				fmt.Println("Timeout waiting for response")
				return
			}
			continue
		}

		// Check correlation ID
		for _, h := range msg.Headers {
			if h.Key == "correlation_id" && string(h.Value) == correlationID {
				fmt.Println("\nReceived response:")
				var response map[string]interface{}
				json.Unmarshal(msg.Value, &response)
				pretty, _ := json.MarshalIndent(response, "", "  ")
				fmt.Println(string(pretty))
				reader.CommitMessages(context.Background(), msg)
				return
			}
		}
		reader.CommitMessages(context.Background(), msg)
	}
}
