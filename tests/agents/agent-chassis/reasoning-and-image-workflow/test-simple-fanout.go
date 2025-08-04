// test-fanout.go
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

	correlationID := uuid.New().String()
	requestID := uuid.New().String()

	payload := map[string]interface{}{
		"action": "multi_agent_test",
		"data": map[string]interface{}{
			"message": "Test fan-out to multiple agents",
			"request": "Analyze this and create an image",
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	msg := kafka.Message{
		Headers: []kafka.Header{
			{Key: "correlation_id", Value: []byte(correlationID)},
			{Key: "request_id", Value: []byte(requestID)},
			{Key: "client_id", Value: []byte("demo_client")},
			{Key: "agent_instance_id", Value: []byte("00000000-0000-0000-0000-000000000050")}, // Fan out test
			{Key: "fuel_budget", Value: []byte("1000")},
		},
		Key:   []byte("fanout-test"),
		Value: payloadBytes,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Fatal("failed to write message:", err)
	}

	fmt.Printf("Fan-out workflow initiated!\n")
	fmt.Printf("Correlation ID: %s\n", correlationID)
	fmt.Printf("\nMonitor with:\n")
	fmt.Printf("curl http://localhost:8080/monitor/workflow/%s | jq .\n", correlationID)
}
