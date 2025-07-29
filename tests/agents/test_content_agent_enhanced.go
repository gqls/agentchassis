// FILE: test_content_agent_enhanced.go (updated section)
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

	kafkaGo "github.com/segmentio/kafka-go" // Import kafka-go directly to access WriterConfig options
)

func main() {
	// Configuration
	// IMPORTANT: Keep this as "localhost:9093" as this is what you're port-forwarding to.
	// We will override the internal Kafka client's Dial and Resolver.
	brokers := []string{"localhost:9093"} // Use the port you actually forwarded to.

	clientID := "demo_client"
	agentInstanceID := "123e4567-e89b-12d3-a456-426614174001"
	userID := "123e4567-e89b-12d3-a456-426614174000" // mysql auth db

	appLogger, _ := logger.New("debug")
	defer appLogger.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --- Setup Kafka Producer with custom Dial mechanism ---
	// Instead of NewProducer directly, we need to create a custom kafka.Writer
	writer := &kafkaGo.Writer{
		Addr:         kafkaGo.TCP(brokers...),
		Balancer:     &kafkaGo.LeastBytes{},
		RequiredAcks: kafkaGo.RequireAll,
		Async:        false,
		WriteTimeout: 10 * time.Second,
		// *** CRITICAL CHANGE ***
		// Override Dial with a custom dialer that resolves only to localhost
		Dialer: &kafkaGo.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true, // Allow both IPv4 and IPv6
			Resolver: &kafkaGo.DNSResolver{
				// This resolver explicitly tells kafka-go client how to resolve addresses.
				// We map all discovered broker IPs/hostnames back to localhost.
				// This is a hack for local testing with port-forwarding.
				LookupIP: func(ctx context.Context, host string) ([]kafkaGo.IPAddr, error) {
					// Always return localhost IP regardless of the host it tries to look up
					return []kafkaGo.IPAddr{{IP: []byte{127, 0, 0, 1}}}, nil
				},
			},
		},
	}
	// Now create your producer instance using this custom writer
	producer := &kafka.KafkaProducer{writer: writer, logger: appLogger} // Re-use your KafkaProducer struct

	// --- Setup Kafka Consumer with custom Dial mechanism ---
	readerConfig := kafkaGo.ReaderConfig{
		Brokers:        brokers,
		GroupID:        "test-consumer-" + uuid.NewString(),
		Topic:          "system.responses.content-creator",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 0,    // Manual commit
		// *** CRITICAL CHANGE ***
		// Apply the same custom Dial/Resolver to the consumer
		Dialer: &kafkaGo.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			Resolver: &kafkaGo.DNSResolver{
				LookupIP: func(ctx context.Context, host string) ([]kafkaGo.IPAddr, error) {
					return []kafkaGo.IPAddr{{IP: []byte{127, 0, 0, 1}}}, nil
				},
			},
		},
	}
	consumer = &kafka.Consumer{reader: kafkaGo.NewReader(readerConfig), logger: appLogger} // Re-use your KafkaConsumer struct

	// ... (rest of the test script remains the same) ...
	// Ensure you also apply the same Dialer logic to any errorConsumer if it's separate.
	// For the error consumer as well:
	errorReaderConfig := kafkaGo.ReaderConfig{
		Brokers:        brokers,
		GroupID:        "test-error-consumer-" + uuid.NewString(),
		Topic:          "system.errors.content-creator",
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
		Dialer: &kafkaGo.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			Resolver: &kafkaGo.DNSResolver{
				LookupIP: func(ctx context.Context, host string) ([]kafkaGo.IPAddr, error) {
					return []kafkaGo.IPAddr{{IP: []byte{127, 0, 0, 1}}}, nil
				},
			},
		},
	}
	errorConsumer := &kafka.Consumer{reader: kafkaGo.NewReader(errorReaderConfig), logger: appLogger}
	defer errorConsumer.Close()

	// ... (rest of the main function, especially the select block)
	// The select block should already try to fetch from both consumers
	// as it was structured previously. Just ensure errorConsumer is handled.
}
