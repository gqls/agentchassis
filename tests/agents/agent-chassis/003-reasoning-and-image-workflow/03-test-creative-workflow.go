// test-creative-workflow.go
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
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"},
		Topic:   "system.agent.generic.process",
	})
	defer writer.Close()

	// Test the Creative Orchestrator
	correlationID := uuid.New().String()
	requestID := uuid.New().String()

	payload := map[string]interface{}{
		"action": "create_artwork",
		"data": map[string]interface{}{
			"prompt": "A futuristic cyberpunk city with neon lights reflecting on wet streets",
			"style":  "digital art, highly detailed",
			"mood":   "mysterious, atmospheric",
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	msg := kafka.Message{
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(correlationID)},
			{Key: "request_id", Value: []byte(requestID)},
			{Key: "client_id", Value: []byte("demo_client")},
			{Key: "agent_instance_id", Value: []byte("00000000-0000-0000-0000-000000000010")}, // Creative Orchestrator
			{Key: "fuel_budget", Value: []byte("1000")},
		},
		Key:   []byte("creative-request"),
		Value: payloadBytes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Fatal("failed to write message:", err)
	}

	fmt.Printf("Creative workflow initiated!\n")
	fmt.Printf("Correlation ID: %s\n", correlationID)
	fmt.Printf("\nMonitor progress:\n")
	fmt.Printf("curl http://localhost:8080/monitor/workflow/%s | jq .\n", correlationID)

	// Test the Fan-out Orchestrator
	time.Sleep(2 * time.Second)

	correlationID2 := uuid.New().String()
	requestID2 := uuid.New().String()

	payload2 := map[string]interface{}{
		"action": "research_and_write",
		"data": map[string]interface{}{
			"topic":        "artificial intelligence in healthcare",
			"requirements": "comprehensive research with marketing copy",
		},
	}

	payloadBytes2, _ := json.Marshal(payload2)

	msg2 := kafka.Message{
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(correlationID2)},
			{Key: "request_id", Value: []byte(requestID2)},
			{Key: "client_id", Value: []byte("demo_client")},
			{Key: "agent_instance_id", Value: []byte("00000000-0000-0000-0000-000000000040")}, // Multi-Agent Orchestrator
			{Key: "fuel_budget", Value: []byte("2000")},
		},
		Key:   []byte("research-request"),
		Value: payloadBytes2,
	}

	err = writer.WriteMessages(ctx, msg2)
	if err != nil {
		log.Fatal("failed to write second message:", err)
	}

	fmt.Printf("\nMulti-agent workflow initiated!\n")
	fmt.Printf("Correlation ID: %s\n", correlationID2)
	fmt.Printf("\nMonitor progress:\n")
	fmt.Printf("curl http://localhost:8080/monitor/workflow/%s | jq .\n", correlationID2)
}
