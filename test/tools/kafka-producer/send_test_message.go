// test/tools/kafka-producer/send_test_message.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var (
	brokers     = flag.String("brokers", "localhost:9092", "Kafka brokers")
	topic       = flag.String("topic", "", "Topic to send to")
	agent       = flag.String("agent", "", "Agent type")
	payloadFile = flag.String("payload", "", "JSON payload file")
	headers     = flag.String("headers", "", "Additional headers (key=value,key=value)")
	wait        = flag.Bool("wait", false, "Wait for response")
)

func main() {
	flag.Parse()

	if *topic == "" && *agent == "" {
		log.Fatal("Either -topic or -agent must be specified")
	}

	// Determine topic
	targetTopic := *topic
	if targetTopic == "" {
		targetTopic = fmt.Sprintf("system.agent.%s.process", *agent)
	}

	// Load payload
	var payload interface{}
	if *payloadFile != "" {
		data, err := os.ReadFile(*payloadFile)
		if err != nil {
			log.Fatal("Failed to read payload file:", err)
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			log.Fatal("Failed to parse payload:", err)
		}
	} else {
		// Default test payload
		payload = map[string]interface{}{
			"action": "test",
			"data": map[string]interface{}{
				"message":   "Test message",
				"timestamp": time.Now().Unix(),
			},
		}
	}

	// Create writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: parsebrokers(*brokers),
		Topic:   targetTopic,
	})
	defer writer.Close()

	// Prepare message
	correlationID := uuid.New().String()
	payloadBytes, _ := json.Marshal(payload)

	// Parse additional headers
	msgHeaders := []kafka.Header{
		{Key: "correlation_id", Value: []byte(correlationID)},
		{Key: "request_id", Value: []byte(uuid.New().String())},
		{Key: "client_id", Value: []byte("test-client")},
		{Key: "fuel_budget", Value: []byte("1000")},
	}

	if *headers != "" {
		// Parse additional headers
		for _, h := range parseHeaders(*headers) {
			msgHeaders = append(msgHeaders, h)
		}
	}

	// Send message
	fmt.Printf("Sending to topic: %s\n", targetTopic)
	fmt.Printf("Correlation ID: %s\n", correlationID)
	fmt.Printf("Payload: %s\n", string(payloadBytes))

	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:     []byte(correlationID),
		Value:   payloadBytes,
		Headers: msgHeaders,
	})

	if err != nil {
		log.Fatal("Failed to send message:", err)
	}

	fmt.Println("✓ Message sent successfully")

	if *wait && *agent != "" {
		waitForResponse(correlationID, *agent)
	}
}

func waitForResponse(correlationID, agent string) {
	fmt.Println("\nWaiting for response...")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   parsebrokers(*brokers),
		Topic:     fmt.Sprintf("system.responses.%s", agent),
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				fmt.Println("✗ Timeout waiting for response")
				return
			}
			continue
		}

		// Check correlation ID
		for _, h := range msg.Headers {
			if h.Key == "correlation_id" && string(h.Value) == correlationID {
				fmt.Println("\n✓ Received response:")
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

func parsebrokers(s string) []string {
	// Split comma-separated brokers
	return strings.Split(s, ",")
}

func parseHeaders(s string) []kafka.Header {
	// Parse key=value,key=value format
	headers := []kafka.Header{}
	for _, pair := range strings.Split(s, ",") {
		parts := strings.Split(pair, "=")
		if len(parts) == 2 {
			headers = append(headers, kafka.Header{
				Key:   strings.TrimSpace(parts[0]),
				Value: []byte(strings.TrimSpace(parts[1])),
			})
		}
	}
	return headers
}
