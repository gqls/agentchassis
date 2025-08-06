// test/integration/kafka/message_flow_test.go
package kafka

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func TestKafkaMessageFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup Kafka writer and reader
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test.topic",
	})
	defer writer.Close()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "test.topic",
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	// Test message flow
	correlationID := "test-kafka-001"
	headers := helpers.TestHeaders(correlationID)

	// Convert headers to Kafka headers
	kafkaHeaders := make([]kafka.Header, 0)
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	// Send message
	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:     []byte(correlationID),
		Value:   []byte(`{"test": "message"}`),
		Headers: kafkaHeaders,
	})
	require.NoError(t, err)

	// Read message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg, err := reader.FetchMessage(ctx)
	require.NoError(t, err)

	assert.Equal(t, correlationID, string(msg.Key))
	assert.Contains(t, string(msg.Value), "test")

	err = reader.CommitMessages(context.Background(), msg)
	require.NoError(t, err)
}
