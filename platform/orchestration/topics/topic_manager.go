// FILE: platform/orchestration/topics/topic_manager.go
package topics

import (
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// GenerateJobTopic creates a unique topic name for a job
func GenerateJobTopic(correlationID, orchestrationID, stepName, agentType string) string {
	// Take first 8 chars for brevity
	if len(correlationID) > 8 {
		correlationID = correlationID[:8]
	}
	if len(orchestrationID) > 8 {
		orchestrationID = orchestrationID[:8]
	}
	if len(agentType) > 10 {
		agentType = agentType[:10]
	}

	stepName = sanitizeTopicPart(stepName)
	agentType = sanitizeTopicPart(agentType)

	return fmt.Sprintf("job.%s.%s.%s.%s", correlationID, orchestrationID, agentType, stepName)
}

// CreateJobTopic creates the topic using Kafka admin client
func CreateJobTopic(brokers []string, topicName string, partitions int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers provided")
	}

	// Create a connection to the Kafka cluster
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	// Get controller
	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	// Connect to controller
	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	// Create topic
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topicName,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		// Ignore "already exists" errors
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create topic %s: %w", topicName, err)
	}

	// Give Kafka a moment to propagate the topic
	time.Sleep(100 * time.Millisecond)

	return nil
}

func sanitizeTopicPart(s string) string {
	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	output := result.String()
	if len(output) > 30 {
		output = output[:30]
	}

	return strings.ToLower(output)
}
