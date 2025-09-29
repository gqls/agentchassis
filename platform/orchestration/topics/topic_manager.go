// FILE: platform/orchestration/topics/topic_manager.go
package topics

import (
	"fmt"
	"strings"
)

// GenerateJobTopic creates a unique topic name for a job
func GenerateJobTopic(correlationID, orchestrationID, stepName string) string {
	// Sanitize the IDs to be Kafka-topic safe
	correlationID = sanitizeTopicPart(correlationID)
	orchestrationID = sanitizeTopicPart(orchestrationID)
	stepName = sanitizeTopicPart(stepName)

	return fmt.Sprintf("job.%s.%s.%s", correlationID, orchestrationID, stepName)
}

// CreateJobTopic creates the topic - simplified version using existing Kafka setup
func CreateJobTopic(brokers []string, topicName string, partitions int) error {
	// For now, topics are auto-created when first used
	// If you need explicit creation, you'll need to implement using your Kafka client
	// or shell out to kafka-topics.sh
	return nil
}

// sanitizeTopicPart makes a string safe for use in Kafka topic names
func sanitizeTopicPart(s string) string {
	// Kafka topic names can contain alphanumeric, '.', '_', and '-'
	// Replace any other characters with '_'
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	// Limit length to prevent overly long topic names
	output := result.String()
	if len(output) > 50 {
		output = output[:50]
	}

	return output
}
