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

type ImageGenerationRequest struct {
	Action string `json:"action"`
	Data   struct {
		Prompt      string  `json:"prompt"`
		AspectRatio string  `json:"aspect_ratio,omitempty"`
		Style       string  `json:"style,omitempty"`
		Seed        float64 `json:"seed,omitempty"`
	} `json:"data"`
}

func main() {
	// Kafka configuration
	brokers := []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}
	requestTopic := "system.adapter.image.generate"
	responseTopic := "system.responses.image"

	// Create a unique correlation ID
	correlationID := uuid.New().String()
	requestID := uuid.New().String()

	// Create the request
	request := ImageGenerationRequest{
		Action: "generate_image",
	}
	request.Data.Prompt = "A beautiful sunset over mountains with a lake in the foreground, photorealistic style"

	requestBytes, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	// Create Kafka writer
	writer := kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    requestTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Create headers
	headers := []kafka.Header{
		{Key: "correlation_id", Value: []byte(correlationID)},
		{Key: "request_id", Value: []byte(requestID)},
		{Key: "client_id", Value: []byte("test-client")},
		{Key: "agent_instance_id", Value: []byte("test-instance")},
		{Key: "timestamp", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
	}

	// Send the message
	msg := kafka.Message{
		Key:     []byte(correlationID),
		Value:   requestBytes,
		Headers: headers,
	}

	fmt.Printf("Sending image generation request...\n")
	fmt.Printf("Correlation ID: %s\n", correlationID)
	fmt.Printf("Prompt: %s\n", request.Data.Prompt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := writer.WriteMessages(ctx, msg); err != nil {
		log.Fatalf("Failed to write message: %v", err)
	}

	fmt.Println("Message sent successfully!")
	fmt.Printf("\nNow listening for response on topic: %s\n", responseTopic)

	// Create a reader for responses
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   responseTopic,
		GroupID: "test-consumer-" + uuid.New().String(),
	})
	defer reader.Close()

	// Read the response
	ctx2, cancel2 := context.WithTimeout(context.Background(), 120*time.Second) // 2 minutes for image generation
	defer cancel2()

	for {
		msg, err := reader.FetchMessage(ctx2)
		if err != nil {
			log.Printf("Timeout or error waiting for response: %v", err)
			break
		}

		// Check if this is our response
		for _, h := range msg.Headers {
			if h.Key == "correlation_id" && string(h.Value) == correlationID {
				fmt.Println("\nReceived response!")

				var response map[string]interface{}
				if err := json.Unmarshal(msg.Value, &response); err != nil {
					log.Printf("Failed to unmarshal response: %v", err)
				} else {
					responseJSON, _ := json.MarshalIndent(response, "", "  ")
					fmt.Printf("Response: %s\n", responseJSON)
				}

				// Commit the message
				reader.CommitMessages(context.Background(), msg)
				return
			}
		}
	}
}
