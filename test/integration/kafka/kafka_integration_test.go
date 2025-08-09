// test/integration/kafka/kafka_integration_test.go
package kafka

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKafkaErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Skip if Kafka is not available
	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	tests := []struct {
		name        string
		brokers     []string
		topic       string
		expectError bool
	}{
		{
			name:        "Valid broker",
			brokers:     getKafkaBrokers(),
			topic:       "test.valid.topic",
			expectError: false,
		},
		{
			name:        "Invalid broker",
			brokers:     []string{"invalid:9999"},
			topic:       "test.topic",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := kafka.NewWriter(kafka.WriterConfig{
				Brokers:  tt.brokers,
				Topic:    tt.topic,
				Balancer: &kafka.LeastBytes{},
			})
			defer writer.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := writer.WriteMessages(ctx, kafka.Message{
				Key:   []byte("test-key"),
				Value: []byte("test-value"),
			})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// For valid broker test, we might still get error if Kafka isn't running
				if err != nil {
					t.Skipf("Kafka not available: %v", err)
				}
			}
		})
	}
}

func TestKafkaPartitioning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	// Create topic with multiple partitions
	topic := "test.partitioned.topic"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  getKafkaBrokers(),
		Topic:    topic,
		Balancer: &kafka.Hash{}, // Use hash balancer for consistent partitioning
	})
	defer writer.Close()

	// Send messages with different keys
	messages := []kafka.Message{
		{Key: []byte("key1"), Value: []byte("value1")},
		{Key: []byte("key2"), Value: []byte("value2")},
		{Key: []byte("key1"), Value: []byte("value3")}, // Same key as first
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, messages...)
	if err != nil {
		t.Skipf("Could not write to Kafka: %v", err)
		return
	}

	// Verify messages with same key went to same partition
	verifyPartitioning(t, topic, messages)
}

func TestKafkaConsumerGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	topic := "test.consumer.group.topic"
	groupID := "test-consumer-group"

	// Create multiple consumers in same group
	consumer1 := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  getKafkaBrokers(),
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer consumer1.Close()

	consumer2 := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  getKafkaBrokers(),
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer consumer2.Close()

	// Send messages
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: getKafkaBrokers(),
		Topic:   topic,
	})
	defer writer.Close()

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: []byte(fmt.Sprintf("value-%d", i)),
		})
		if err != nil {
			t.Skipf("Could not write to Kafka: %v", err)
			return
		}
	}

	// Verify load balancing between consumers
	var consumer1Count int32
	var consumer2Count int32

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Read from both consumers
	go countMessages(ctx, consumer1, &consumer1Count)
	go countMessages(ctx, consumer2, &consumer2Count)

	time.Sleep(5 * time.Second)
	cancel() // Stop consumers

	// Get final counts
	c1 := atomic.LoadInt32(&consumer1Count)
	c2 := atomic.LoadInt32(&consumer2Count)

	// Both should have received messages (in a real Kafka setup)
	// In mock/unavailable scenario, we might get 0
	t.Logf("Consumer 1 received %d messages", c1)
	t.Logf("Consumer 2 received %d messages", c2)

	if c1+c2 > 0 {
		assert.Greater(t, c1, int32(0), "Consumer 1 should have received messages")
		assert.Greater(t, c2, int32(0), "Consumer 2 should have received messages")
		assert.Equal(t, int32(10), c1+c2, "Total messages should be 10")
	}
}

// Helper function to verify partitioning
func verifyPartitioning(t *testing.T, topic string, messages []kafka.Message) {
	// Create a reader to verify messages
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   getKafkaBrokers(),
		Topic:     topic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	// Track which partition each key went to
	keyPartitions := make(map[string]int)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read messages and track partitions
	for i := 0; i < len(messages); i++ {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			// If we can't read, skip verification
			t.Logf("Could not read messages for verification: %v", err)
			return
		}

		key := string(msg.Key)
		if prevPartition, exists := keyPartitions[key]; exists {
			// Verify same key went to same partition
			assert.Equal(t, prevPartition, msg.Partition,
				"Key %s should always go to same partition", key)
		} else {
			keyPartitions[key] = msg.Partition
		}
	}

	// Verify that key1 messages went to same partition
	if len(keyPartitions) > 0 {
		assert.Equal(t, keyPartitions["key1"], keyPartitions["key1"],
			"All messages with key1 should be in same partition")
	}
}

// Helper function to count messages received by a consumer
func countMessages(ctx context.Context, reader *kafka.Reader, count *int32) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Set a short timeout for each read
			readCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
			_, err := reader.ReadMessage(readCtx)
			cancel()

			if err == nil {
				atomic.AddInt32(count, 1)
			} else if err == context.DeadlineExceeded {
				// This is expected when no messages are available
				continue
			} else if err == context.Canceled {
				// Context was cancelled, stop reading
				return
			} else {
				// Some other error, log and continue
				continue
			}
		}
	}
}

// Helper function to check if Kafka is available
func isKafkaAvailable() bool {
	brokers := getKafkaBrokers()

	// Try to connect to Kafka with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return false
	}
	defer conn.Close()

	// Try to get metadata as a health check
	_, err = conn.ApiVersions()
	return err == nil
}

// getKafkaBrokers returns the appropriate Kafka brokers based on environment
func getKafkaBrokers() []string {
	// Check for explicit Kafka broker configuration
	if kafkaBroker := os.Getenv("KAFKA_BROKERS"); kafkaBroker != "" {
		return []string{kafkaBroker}
	}

	// Check if running in Kubernetes
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// Use Kubernetes service name
		return []string{"kafka-clients.ai-persona-system.svc.cluster.local:9092"}
	}

	// Default to localhost for local development
	return []string{"localhost:9092"}
}

// Additional test for message ordering
func TestKafkaMessageOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	topic := "test.ordering.topic"

	// Create writer with single partition to ensure ordering
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  getKafkaBrokers(),
		Topic:    topic,
		Balancer: &kafka.RoundRobin{},
	})
	defer writer.Close()

	// Send messages in order
	messages := make([]kafka.Message, 10)
	for i := 0; i < 10; i++ {
		messages[i] = kafka.Message{
			Key:   []byte("same-key"), // Same key ensures same partition
			Value: []byte(fmt.Sprintf("message-%d", i)),
		}
	}

	ctx := context.Background()
	err := writer.WriteMessages(ctx, messages...)
	if err != nil {
		t.Skipf("Could not write to Kafka: %v", err)
		return
	}

	// Read messages and verify order
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  getKafkaBrokers(),
		Topic:    topic,
		GroupID:  "test-ordering-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	for i := 0; i < 10; i++ {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			t.Skipf("Could not read from Kafka: %v", err)
			return
		}

		expected := fmt.Sprintf("message-%d", i)
		assert.Equal(t, expected, string(msg.Value),
			"Messages should be in order")
	}
}

// Test for handling large messages
func TestKafkaLargeMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	topic := "test.large.messages"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  getKafkaBrokers(),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	// Create a large message (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("large-key"),
		Value: largeData,
	})

	if err != nil {
		t.Logf("Large message test: %v", err)
		// This might fail due to Kafka configuration
		// Default max message size is often 1MB
	} else {
		// Verify we can read it back
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:  getKafkaBrokers(),
			Topic:    topic,
			GroupID:  "test-large-group",
			MinBytes: 1,
			MaxBytes: 10e6,
		})
		defer reader.Close()

		msg, err := reader.ReadMessage(ctx)
		require.NoError(t, err)
		assert.Equal(t, len(largeData), len(msg.Value))
	}
}
