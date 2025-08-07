// test/integration/kafka/kafka_integration_test.go
package kafka

import (
	"context"
	"fmt"
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

	tests := []struct {
		name        string
		brokers     []string
		topic       string
		expectError bool
	}{
		{
			name:        "Valid broker",
			brokers:     []string{"localhost:9092"},
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
				assert.NoError(t, err)
			}
		})
	}
}

func TestKafkaPartitioning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create topic with multiple partitions
	topic := "test.partitioned.topic"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"localhost:9092"},
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

	err := writer.WriteMessages(context.Background(), messages...)
	require.NoError(t, err)

	// Verify messages with same key went to same partition
	// Read from specific partitions and verify
	verifyPartitioning(t, topic, messages)
}

func TestKafkaConsumerGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	topic := "test.consumer.group.topic"
	groupID := "test-consumer-group"

	// Create multiple consumers in same group
	consumer1 := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer consumer1.Close()

	consumer2 := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer consumer2.Close()

	// Send messages
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
	})
	defer writer.Close()

	for i := 0; i < 10; i++ {
		err := writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: []byte(fmt.Sprintf("value-%d", i)),
		})
		require.NoError(t, err)
	}

	// Verify load balancing between consumers
	var consumer1Count, consumer2Count int
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Read from both consumers
	go countMessages(ctx, consumer1, &consumer1Count)
	go countMessages(ctx, consumer2, &consumer2Count)

	time.Sleep(5 * time.Second)

	// Both should have received messages
	assert.Greater(t, consumer1Count, 0)
	assert.Greater(t, consumer2Count, 0)
	assert.Equal(t, 10, consumer1Count+consumer2Count)
}
