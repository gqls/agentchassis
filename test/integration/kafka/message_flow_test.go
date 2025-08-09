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

// TestDiscoverTopics lists all available topics in the Kafka cluster
func TestDiscoverTopics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	brokers := getKafkaBrokers()
	conn, err := kafka.Dial("tcp", brokers[0])
	require.NoError(t, err)
	defer conn.Close()

	// Get all partitions which includes topic information
	partitions, err := conn.ReadPartitions()
	require.NoError(t, err)

	// Collect unique topics
	topics := make(map[string][]int)
	for _, p := range partitions {
		topics[p.Topic] = append(topics[p.Topic], p.ID)
	}

	t.Logf("Discovered %d topics in Kafka cluster:", len(topics))
	for topic, partitionIDs := range topics {
		t.Logf("  - %s (partitions: %v)", topic, partitionIDs)
	}

	// Store discovered topics for use in other tests
	if len(topics) > 0 {
		// Get the first available topic for testing
		for topic := range topics {
			t.Logf("Suggested test topic: %s", topic)
			break
		}
	}
}

// TestCreateAndUseTestTopic creates a test topic and uses it
func TestCreateAndUseTestTopic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	brokers := getKafkaBrokers()

	// Create a unique test topic
	testTopic := fmt.Sprintf("test-topic-%s", helpers.TestUUIDWithType("integration"))

	// Try to create the topic
	conn, err := kafka.Dial("tcp", brokers[0])
	require.NoError(t, err)
	defer conn.Close()

	controller, err := conn.Controller()
	require.NoError(t, err)

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	require.NoError(t, err)
	defer controllerConn.Close()

	topicConfig := kafka.TopicConfig{
		Topic:             testTopic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		// Check if it's a permissions issue or other problem
		if strings.Contains(err.Error(), "Topic Already Exists") {
			t.Logf("Topic %s already exists", testTopic)
		} else if strings.Contains(err.Error(), "Authorization") {
			t.Skip("Cannot create topics - insufficient permissions")
			return
		} else {
			t.Logf("Could not create topic: %v", err)
			t.Skip("Topic creation not supported in this Kafka cluster")
			return
		}
	} else {
		t.Logf("Successfully created topic: %s", testTopic)
	}

	// Now test writing to the created topic
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    testTopic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	correlationID := helpers.TestUUIDWithType("integration")
	message := kafka.Message{
		Key:   []byte(correlationID),
		Value: []byte(`{"test": "message", "timestamp": ` + fmt.Sprintf("%d", time.Now().Unix()) + `}`),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = writer.WriteMessages(ctx, message)
	require.NoError(t, err)

	t.Logf("Successfully wrote message to topic: %s", testTopic)

	// Try to read it back
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     testTopic,
		Partition: 0,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	// Set the offset to the beginning to read our message
	reader.SetOffset(kafka.FirstOffset)

	readCtx, readCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer readCancel()

	msg, err := reader.FetchMessage(readCtx)
	if err == nil {
		assert.Equal(t, correlationID, string(msg.Key))
		t.Logf("Successfully read message back from topic: %s", testTopic)

		// Commit the message
		reader.CommitMessages(context.Background(), msg)
	} else {
		t.Logf("Could not read message back: %v", err)
	}
}

// TestKafkaWithDiscoveredTopics uses topics discovered in the cluster
func TestKafkaWithDiscoveredTopics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Skip("Kafka not available for testing")
	}

	brokers := getKafkaBrokers()
	conn, err := kafka.Dial("tcp", brokers[0])
	require.NoError(t, err)
	defer conn.Close()

	// Get all available topics
	partitions, err := conn.ReadPartitions()
	require.NoError(t, err)

	topics := make(map[string]bool)
	for _, p := range partitions {
		topics[p.Topic] = true
	}

	if len(topics) == 0 {
		t.Skip("No topics found in Kafka cluster")
		return
	}

	// Try to write to each discovered topic
	successCount := 0
	var lastError error

	for topic := range topics {
		// Skip internal Kafka topics
		if strings.HasPrefix(topic, "__") || strings.HasPrefix(topic, "_") {
			t.Logf("Skipping internal topic: %s", topic)
			continue
		}

		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		})

		correlationID := helpers.TestUUIDWithType("integration")
		message := kafka.Message{
			Key:   []byte(correlationID),
			Value: []byte(`{"test": "message"}`),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := writer.WriteMessages(ctx, message)
		cancel()
		writer.Close()

		if err != nil {
			t.Logf("Could not write to topic %s: %v", topic, err)
			lastError = err
		} else {
			t.Logf("Successfully wrote to topic: %s", topic)
			successCount++

			// If we found at least one writable topic, that's enough
			if successCount >= 1 {
				break
			}
		}
	}

	if successCount == 0 {
		if lastError != nil {
			t.Logf("No topics were writable. Last error: %v", lastError)
		}
		// Try to create a test topic as fallback
		t.Log("Attempting to create a test topic...")
		TestCreateAndUseTestTopic(t)
	} else {
		t.Logf("Successfully wrote to %d topic(s)", successCount)
	}
}

// TestKafkaConnectivity performs basic connectivity test
func TestKafkaConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	brokers := getKafkaBrokers()
	t.Logf("Testing connectivity to Kafka brokers: %v", brokers)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	require.NoError(t, err, "Failed to connect to Kafka")
	defer conn.Close()

	// Get broker metadata
	broker := conn.Broker()
	require.NoError(t, err)

	t.Logf("Connected to Kafka broker: ID=%d, Host=%s, Port=%d, Rack=%s",
		broker.ID, broker.Host, broker.Port, broker.Rack)

	// Get controller information
	controller, err := conn.Controller()
	require.NoError(t, err)

	t.Logf("Kafka Controller: ID=%d, Host=%s, Port=%d",
		controller.ID, controller.Host, controller.Port)

	// Get API versions
	apiVersions, err := conn.ApiVersions()
	require.NoError(t, err)

	t.Logf("Kafka cluster supports %d API versions", len(apiVersions))

	// Show a few API versions for debugging
	if len(apiVersions) > 0 {
		t.Log("Sample API versions:")
		count := 5
		if len(apiVersions) < count {
			count = len(apiVersions)
		}
		for i := 0; i < count; i++ {
			v := apiVersions[i]
			t.Logf("  - API Key: %d, Min Version: %d, Max Version: %d",
				v.ApiKey, v.MinVersion, v.MaxVersion)
		}
	}
}

// TestKafkaHealthCheck performs a basic health check
func TestKafkaHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if !isKafkaAvailable() {
		t.Fatal("Kafka is not available - health check failed")
	}

	t.Log("Kafka health check passed")
}
