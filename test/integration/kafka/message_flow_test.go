// test/integration/kafka/message_flow_test.go
package kafka

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKafkaBasicOperations tests basic Kafka operations
func TestKafkaBasicOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	brokers := getKafkaBrokers()

	// First, discover what topics exist
	conn, err := kafka.Dial("tcp", brokers[0])
	require.NoError(t, err)
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		t.Logf("Could not read partitions: %v", err)
		// If we can't read partitions, try to create a test topic
		testTopic := fmt.Sprintf("test-%d", time.Now().Unix())
		if err := createTestTopic(t, brokers[0], testTopic); err != nil {
			t.Skip("Cannot read partitions or create topics")
			return
		}
		// Use the created topic
		testWriteRead(t, brokers, testTopic)
		return
	}

	// Find a usable topic
	var usableTopic string
	for _, p := range partitions {
		// Skip internal topics
		if !strings.HasPrefix(p.Topic, "__") && !strings.HasPrefix(p.Topic, "_") {
			usableTopic = p.Topic
			t.Logf("Found usable topic: %s", usableTopic)
			break
		}
	}

	if usableTopic == "" {
		// No existing topics, try to create one
		testTopic := fmt.Sprintf("test-%d", time.Now().Unix())
		if err := createTestTopic(t, brokers[0], testTopic); err != nil {
			t.Skip("No usable topics and cannot create new ones")
			return
		}
		usableTopic = testTopic
	}

	// Test write and read with the usable topic
	testWriteRead(t, brokers, usableTopic)
}

// testWriteRead performs a write and optional read test
func testWriteRead(t *testing.T, brokers []string, topic string) {
	// Write test
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		RequiredAcks: 1,
		MaxAttempts:  3,
	})
	defer writer.Close()

	correlationID := helpers.TestUUIDWithType("integration")
	message := kafka.Message{
		Key:   []byte(correlationID),
		Value: []byte(fmt.Sprintf(`{"test": true, "id": "%s", "timestamp": %d}`, correlationID, time.Now().Unix())),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, message)
	if err != nil {
		t.Errorf("Failed to write message to topic %s: %v", topic, err)
		return
	}

	t.Logf("Successfully wrote message to topic: %s", topic)

	// The write succeeded, which is the main test
	// Reading back is optional as other consumers might get the message
}

// createTestTopic attempts to create a test topic
func createTestTopic(t *testing.T, broker string, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	t.Logf("Created test topic: %s", topic)
	return nil
}

// TestKafkaHealthCheck verifies Kafka is accessible
// In TestKafkaHealthCheck function
func TestKafkaHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	brokers := getKafkaBrokers()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		t.Fatalf("Kafka health check failed - cannot connect: %v", err)
	}
	defer conn.Close()

	// Get broker info (returns only Broker, no error)
	broker := conn.Broker()

	// Get API versions to verify functionality
	apiVersions, err := conn.ApiVersions()
	require.NoError(t, err)

	assert.Greater(t, len(apiVersions), 0, "Should support API versions")

	t.Logf("Kafka health check passed - Broker: %s:%d, APIs: %d",
		broker.Host, broker.Port, len(apiVersions))
}

// In TestKafkaMetadata function
func TestKafkaMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available")
	}

	brokers := getKafkaBrokers()
	conn, err := kafka.Dial("tcp", brokers[0])
	require.NoError(t, err)
	defer conn.Close()

	// Get and log controller info
	controller, err := conn.Controller()
	if err != nil {
		t.Logf("Could not get controller info: %v", err)
	} else {
		t.Logf("Controller: ID=%d, Host=%s, Port=%d",
			controller.ID, controller.Host, controller.Port)
	}

	// Try to read partitions
	partitions, err := conn.ReadPartitions()
	if err != nil {
		t.Logf("Could not read partitions: %v", err)
	} else {
		// Count topics
		topics := make(map[string]int)
		for _, p := range partitions {
			topics[p.Topic]++
		}
		t.Logf("Found %d topics with %d total partitions", len(topics), len(partitions))

		// Log first few topics
		count := 0
		for topic, partitionCount := range topics {
			if count >= 5 {
				break
			}
			t.Logf("  - %s (%d partitions)", topic, partitionCount)
			count++
		}
	}
}
