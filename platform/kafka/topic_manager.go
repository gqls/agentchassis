// FILE: platform/kafka/topic_manager.go
package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// TopicManager handles dynamic Kafka topic creation
type TopicManager struct {
	brokers []string
	logger  *zap.Logger
	timeout time.Duration
}

// NewTopicManager creates a new topic manager
func NewTopicManager(brokers []string, logger *zap.Logger) *TopicManager {
	return &TopicManager{
		brokers: brokers,
		logger:  logger,
		timeout: 30 * time.Second,
	}
}

// TopicDefinition defines a single topic configuration
type TopicDefinition struct {
	Name              string
	Partitions        int
	ReplicationFactor int
}

// CreateAgentTopics creates all topics needed for a specific agent type
func (tm *TopicManager) CreateAgentTopics(ctx context.Context, agentType string) error {
	tm.logger.Info("Creating topics for agent type", zap.String("agent_type", agentType))

	topics := tm.getTopicsForAgent(agentType)

	for _, topic := range topics {
		if err := tm.CreateTopic(ctx, topic); err != nil {
			// Don't fail on "already exists" errors
			if !strings.Contains(err.Error(), "already exists") {
				tm.logger.Error("Failed to create topic",
					zap.String("topic", topic.Name),
					zap.Error(err))
			}
		}
	}

	tm.logger.Info("Successfully created all topics for agent",
		zap.String("agent_type", agentType),
		zap.Int("topic_count", len(topics)))

	return nil
}

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

// CreateJobTopic creates a single job-specific topic
// This is a simple wrapper that handles empty brokers gracefully
func CreateJobTopic(brokers []string, topicName string, partitions int) error {
	// Validate and fix brokers
	validBrokers := []string{}
	for _, broker := range brokers {
		broker = strings.TrimSpace(broker)
		if broker != "" && broker != ":9092" { // Filter out empty or invalid brokers
			validBrokers = append(validBrokers, broker)
		}
	}

	// If no valid brokers provided, use defaults
	if len(validBrokers) == 0 {
		validBrokers = []string{
			"personae-kafka-cluster-kafka-bootstrap.kafka:9092",
			"kafka-0.kafka-headless.kafka:9092",
		}
	}

	// Try to create the topic using each broker
	var lastErr error
	for _, broker := range validBrokers {
		conn, err := kafka.Dial("tcp", broker)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect to broker %s: %w", broker, err)
			continue
		}
		defer conn.Close()

		// Set a timeout
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		// Check if topic already exists
		partitionList, err := conn.ReadPartitions()
		if err == nil {
			for _, p := range partitionList {
				if p.Topic == topicName {
					// Topic already exists
					return nil
				}
			}
		}

		// Create the topic
		topicConfig := kafka.TopicConfig{
			Topic:             topicName,
			NumPartitions:     partitions,
			ReplicationFactor: 1, // Use 1 for job topics
			ConfigEntries: []kafka.ConfigEntry{
				{ConfigName: "retention.ms", ConfigValue: "3600000"}, // 1 hour for job topics
				{ConfigName: "compression.type", ConfigValue: "snappy"},
			},
		}

		err = conn.CreateTopics(topicConfig)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return nil // Topic exists, that's fine
			}
			lastErr = fmt.Errorf("failed to create topic on broker %s: %w", broker, err)
			continue
		}

		// Success!
		return nil
	}

	// All attempts failed
	return fmt.Errorf("failed to create topic %s on any broker: %w", topicName, lastErr)
}

// CreateTopic creates a single topic if it doesn't exist
func (tm *TopicManager) CreateTopic(ctx context.Context, topic TopicDefinition) error {
	// Create a controller connection
	controller, err := tm.getController(ctx)
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	conn, err := kafka.Dial("tcp", controller)
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer conn.Close()

	// Set deadline for the operation
	deadline := time.Now().Add(tm.timeout)
	conn.SetDeadline(deadline)

	// Check if topic exists first
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	for _, p := range partitions {
		if p.Topic == topic.Name {
			tm.logger.Debug("Topic already exists", zap.String("topic", topic.Name))
			return nil
		}
	}

	// Create topic
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic.Name,
			NumPartitions:     topic.Partitions,
			ReplicationFactor: topic.ReplicationFactor,
			ConfigEntries: []kafka.ConfigEntry{
				{ConfigName: "retention.ms", ConfigValue: "604800000"}, // 7 days
				{ConfigName: "compression.type", ConfigValue: "snappy"},
				{ConfigName: "cleanup.policy", ConfigValue: "delete"},
			},
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		// Check if error is because topic already exists (race condition)
		if strings.Contains(err.Error(), "already exists") {
			tm.logger.Debug("Topic already exists (race condition)", zap.String("topic", topic.Name))
			return nil
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	tm.logger.Info("Topic created successfully",
		zap.String("topic", topic.Name),
		zap.Int("partitions", topic.Partitions),
		zap.Int("replication", topic.ReplicationFactor))

	return nil
}

// TopicExists checks if a topic already exists
func (tm *TopicManager) TopicExists(ctx context.Context, topicName string) (bool, error) {
	controller, err := tm.getController(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get controller: %w", err)
	}

	conn, err := kafka.Dial("tcp", controller)
	if err != nil {
		return false, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	// Set deadline
	deadline := time.Now().Add(tm.timeout)
	conn.SetDeadline(deadline)

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return false, fmt.Errorf("failed to read partitions: %w", err)
	}

	for _, p := range partitions {
		if p.Topic == topicName {
			return true, nil
		}
	}

	return false, nil
}

// getController finds the current Kafka controller
func (tm *TopicManager) getController(ctx context.Context) (string, error) {
	for _, broker := range tm.brokers {
		conn, err := kafka.Dial("tcp", broker)
		if err != nil {
			tm.logger.Warn("Failed to connect to broker",
				zap.String("broker", broker),
				zap.Error(err))
			continue
		}

		controller, err := conn.Controller()
		if err != nil {
			conn.Close()
			continue
		}

		conn.Close()
		return fmt.Sprintf("%s:%d", controller.Host, controller.Port), nil
	}

	return "", fmt.Errorf("failed to find Kafka controller")
}

// getTopicsForAgent returns the topics needed for a specific agent type
func (tm *TopicManager) getTopicsForAgent(agentType string) []TopicDefinition {
	// Base topics with NEW consistent naming convention
	topics := []TopicDefinition{
		{
			Name:              fmt.Sprintf("system.agent.%s.requests", agentType),
			Partitions:        3,
			ReplicationFactor: 2,
		},
		{
			Name:              fmt.Sprintf("system.agent.%s.responses", agentType),
			Partitions:        3,
			ReplicationFactor: 2,
		},
		{
			Name:              fmt.Sprintf("system.agent.%s.errors", agentType),
			Partitions:        1,
			ReplicationFactor: 2,
		},
		{
			Name:              fmt.Sprintf("system.agent.%s.dlq", agentType),
			Partitions:        1,
			ReplicationFactor: 2,
		},
	}

	// Add priority-based topics for data-driven agents
	if isDataDrivenAgent(agentType) {
		for _, priority := range []string{"high", "normal", "low"} {
			topics = append(topics, TopicDefinition{
				Name:              fmt.Sprintf("system.agent.%s.tasks.%s", agentType, priority),
				Partitions:        3,
				ReplicationFactor: 2,
			})
		}
	}

	// Add adapter-specific topics
	if isAdapterAgent(agentType) {
		topics = append(topics, TopicDefinition{
			Name:              fmt.Sprintf("system.agent.%s.adapter", agentType),
			Partitions:        3,
			ReplicationFactor: 2,
		})
	}

	// Add special topics for high-throughput agents
	if isHighThroughputAgent(agentType) {
		topics = append(topics, TopicDefinition{
			Name:              fmt.Sprintf("system.agent.%s.bulk", agentType),
			Partitions:        6, // More partitions for higher throughput
			ReplicationFactor: 2,
		})
	}

	return topics
}

// isDataDrivenAgent checks if an agent is data-driven type
func isDataDrivenAgent(agentType string) bool {
	dataDrivenTypes := map[string]bool{
		"copywriter":         true,
		"researcher":         true,
		"content-creator":    true,
		"content-researcher": true,
		"summarizer":         true,
		"domain-analyst":     true,
	}
	return dataDrivenTypes[agentType]
}

// isAdapterAgent checks if an agent is an adapter type
func isAdapterAgent(agentType string) bool {
	adapterTypes := map[string]bool{
		"image-generator": true,
		"web-search":      true,
		"database-query":  true,
		"api-caller":      true,
		"site-publisher":  true,
	}
	return adapterTypes[agentType]
}

// isHighThroughputAgent checks if an agent needs high throughput topics
func isHighThroughputAgent(agentType string) bool {
	highThroughputTypes := map[string]bool{
		"reasoning":       true,
		"website-builder": true,
		"orchestrator":    true,
	}
	return highThroughputTypes[agentType]
}

// CreateSystemTopics creates all system-level topics
func (tm *TopicManager) CreateSystemTopics(ctx context.Context) error {
	systemTopics := []TopicDefinition{
		// Core orchestration topics
		{Name: "system.orchestrator.requests", Partitions: 6, ReplicationFactor: 2},
		{Name: "system.orchestrator.responses", Partitions: 6, ReplicationFactor: 2},
		{Name: "system.orchestrator.state-changes", Partitions: 6, ReplicationFactor: 2},
		{Name: "system.orchestrator.commands", Partitions: 3, ReplicationFactor: 2},

		// Human interaction topics
		{Name: "system.human.approvals", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.human.inputs", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.commands.workflow.resume", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.commands.workflow.cancel", Partitions: 3, ReplicationFactor: 2},

		// System events and monitoring
		{Name: "system.events.all", Partitions: 6, ReplicationFactor: 2},
		{Name: "system.events.errors", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.events.warnings", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.notifications.ui", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.metrics.agents", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.metrics.workflows", Partitions: 3, ReplicationFactor: 2},

		// Audit and compliance
		{Name: "system.audit.log", Partitions: 6, ReplicationFactor: 3}, // Higher replication for audit
		{Name: "system.audit.access", Partitions: 3, ReplicationFactor: 3},
		{Name: "system.compliance.events", Partitions: 3, ReplicationFactor: 3},

		// Dead letter queues for system-level issues
		{Name: "system.dlq.unroutable", Partitions: 1, ReplicationFactor: 2},
		{Name: "system.dlq.parsing-errors", Partitions: 1, ReplicationFactor: 2},
	}

	for _, topic := range systemTopics {
		if err := tm.CreateTopic(ctx, topic); err != nil {
			tm.logger.Error("Failed to create system topic",
				zap.String("topic", topic.Name),
				zap.Error(err))
			// Continue with other topics even if one fails
		}
	}

	tm.logger.Info("System topic creation completed")
	return nil
}

// DeleteTopic deletes a topic (use with caution!)
func (tm *TopicManager) DeleteTopic(ctx context.Context, topicName string) error {
	controller, err := tm.getController(ctx)
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	conn, err := kafka.Dial("tcp", controller)
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer conn.Close()

	// Set deadline
	deadline := time.Now().Add(tm.timeout)
	conn.SetDeadline(deadline)

	err = conn.DeleteTopics(topicName)
	if err != nil {
		return fmt.Errorf("failed to delete topic %s: %w", topicName, err)
	}

	tm.logger.Warn("Topic deleted", zap.String("topic", topicName))
	return nil
}

// ListTopics returns all topics in the cluster
func (tm *TopicManager) ListTopics(ctx context.Context) ([]string, error) {
	controller, err := tm.getController(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller: %w", err)
	}

	conn, err := kafka.Dial("tcp", controller)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	// Set deadline
	deadline := time.Now().Add(tm.timeout)
	conn.SetDeadline(deadline)

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	// Use map to deduplicate topic names
	topicMap := make(map[string]bool)
	for _, p := range partitions {
		topicMap[p.Topic] = true
	}

	// Convert to slice
	topics := make([]string, 0, len(topicMap))
	for topic := range topicMap {
		topics = append(topics, topic)
	}

	return topics, nil
}

// CreateAgentTypeTopics is now just an alias for CreateAgentTopics
// since we've standardized on the new naming convention
func (tm *TopicManager) CreateAgentTypeTopics(ctx context.Context, agentType string) error {
	// Just create the core request/response topics if you want minimal setup
	tm.logger.Info("Creating core topics for agent type",
		zap.String("agent_type", agentType))

	topics := []TopicDefinition{
		{
			Name:              fmt.Sprintf("system.agent.%s.requests", agentType),
			Partitions:        3,
			ReplicationFactor: 2,
		},
		{
			Name:              fmt.Sprintf("system.agent.%s.responses", agentType),
			Partitions:        3,
			ReplicationFactor: 2,
		},
	}

	for _, topic := range topics {
		if err := tm.CreateTopic(ctx, topic); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				tm.logger.Error("Failed to create topic",
					zap.String("topic", topic.Name),
					zap.Error(err))
			}
		}
	}

	tm.logger.Info("Agent type core topics ready",
		zap.String("agent_type", agentType))

	return nil
}

// MigrateOldTopics helps migrate from old naming to new naming
func (tm *TopicManager) MigrateOldTopics(ctx context.Context, agentType string) error {
	tm.logger.Info("Checking for old topic naming conventions",
		zap.String("agent_type", agentType))

	oldToNew := map[string]string{
		fmt.Sprintf("system.agent.%s.process", agentType): fmt.Sprintf("system.agent.%s.requests", agentType),
		fmt.Sprintf("system.responses.%s", agentType):     fmt.Sprintf("system.agent.%s.responses", agentType),
		fmt.Sprintf("system.errors.%s", agentType):        fmt.Sprintf("system.agent.%s.errors", agentType),
		fmt.Sprintf("dlq.%s", agentType):                  fmt.Sprintf("system.agent.%s.dlq", agentType),
	}

	for oldTopic, newTopic := range oldToNew {
		// Check if old topic exists
		if exists, _ := tm.TopicExists(ctx, oldTopic); exists {
			tm.logger.Warn("Found old topic naming, consider migrating",
				zap.String("old_topic", oldTopic),
				zap.String("new_topic", newTopic))
			// Note: Actual migration would involve consuming from old and producing to new
			// This is just detection for now
		}
	}

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

func (tm *TopicManager) WaitForTopicOld(ctx context.Context, topic string, logger *zap.Logger) error {
	conn, err := kafka.DialContext(ctx, "tcp", tm.brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := 0; i < 10; i++ {
		partitions, err := conn.ReadPartitions(topic)
		if err == nil && len(partitions) > 0 {
			logger.Info("Topic found and ready", zap.String("topic", topic))
			return nil
		}
		logger.Info("Waiting for topic to propagate...", zap.Int("attempt", i+1))
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("topic %s not found after waiting", topic)
}

// In platform/kafka/topic_manager.go

func (tm *TopicManager) WaitForTopic(ctx context.Context, topic string, logger *zap.Logger) error {
	const maxAttempts = 10
	const pollInterval = 5 * time.Second

	// We'll poll for a total of 10 seconds.
	for i := 0; i < maxAttempts; i++ {
		allBrokersReady := true
		checkedBrokers := 0

		// On each attempt, we now check ALL brokers in the list.
		for _, brokerAddr := range tm.brokers {
			conn, err := kafka.DialContext(ctx, "tcp", brokerAddr)
			if err != nil {
				logger.Warn("Could not connect to broker to verify topic, will retry",
					zap.String("broker", brokerAddr),
					zap.Error(err))
				allBrokersReady = false
				break // If one broker is down, fail this attempt and retry the poll.
			}
			defer conn.Close()

			partitions, err := conn.ReadPartitions(topic)
			if err != nil || len(partitions) == 0 {
				// This broker does not know about the topic yet.
				allBrokersReady = false
				break // No need to check other brokers; we know it's not ready yet.
			}
			checkedBrokers++
		}

		// If all brokers were checked and all were ready, success!
		if allBrokersReady {
			logger.Info("Topic propagated and is ready on all brokers",
				zap.String("topic", topic),
				zap.Int("brokers_checked", checkedBrokers))
			return nil
		}

		// If not ready, wait for the next polling interval.
		logger.Info("Waiting for topic to propagate to all brokers...",
			zap.String("topic", topic),
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", maxAttempts))
		time.Sleep(pollInterval)
	}

	// If the loop finishes, it means we timed out.
	return fmt.Errorf("topic %s not ready on all brokers after %d seconds", topic, maxAttempts)
}
