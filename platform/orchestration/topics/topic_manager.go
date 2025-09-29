// FILE: platform/orchestration/topics/topic_manager.go
package topics

import (
	"fmt"
	"os/exec"
	"strings"
)

// GenerateJobTopic creates a unique topic name for a job
func GenerateJobTopic(correlationID, orchestrationID, stepName string) string {
	// Take first 8 chars for brevity
	if len(correlationID) > 8 {
		correlationID = correlationID[:8]
	}
	if len(orchestrationID) > 8 {
		orchestrationID = orchestrationID[:8]
	}

	stepName = sanitizeTopicPart(stepName)

	return fmt.Sprintf("job.%s.%s.%s", correlationID, orchestrationID, stepName)
}

// CreateJobTopic actually creates the topic using kafka-topics.sh
func CreateJobTopic(brokers []string, topicName string, partitions int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers provided")
	}

	// Use the first broker
	//broker := brokers[0]

	// Create the topic using kafka-topics.sh via kubectl
	cmd := exec.Command("kubectl", "-n", "kafka", "exec", "-i",
		"personae-kafka-cluster-combined-pool-prod-0", "--",
		"/opt/kafka/bin/kafka-topics.sh",
		"--bootstrap-server", "localhost:9092",
		"--create",
		"--topic", topicName,
		"--partitions", fmt.Sprintf("%d", partitions),
		"--replication-factor", "1",
		"--if-not-exists")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's because the topic already exists
		if strings.Contains(string(output), "already exists") {
			return nil // Topic exists, that's fine
		}
		return fmt.Errorf("failed to create topic %s: %v - output: %s", topicName, err, string(output))
	}

	return nil
}

func sanitizeTopicPart(s string) string {
	// Replace spaces with hyphens, remove other special chars
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
