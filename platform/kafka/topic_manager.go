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
			return fmt.Errorf("failed to create topic %s: %w", topic.Name, err)
		}
	}

	tm.logger.Info("Successfully created all topics for agent",
		zap.String("agent_type", agentType),
		zap.Int("topic_count", len(topics)))

	return nil
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
	// Base topics that all agents need
	topics := []TopicDefinition{
		{
			Name:              fmt.Sprintf("system.agent.%s.process", agentType),
			Partitions:        3,
			ReplicationFactor: 2,
		},
		{
			Name:              fmt.Sprintf("system.responses.%s", agentType),
			Partitions:        3,
			ReplicationFactor: 2,
		},
		{
			Name:              fmt.Sprintf("system.errors.%s", agentType),
			Partitions:        1,
			ReplicationFactor: 2,
		},
		{
			Name:              fmt.Sprintf("dlq.%s", agentType),
			Partitions:        1,
			ReplicationFactor: 2,
		},
	}

	// Add priority-based topics for data-driven agents
	if isDataDrivenAgent(agentType) {
		for _, priority := range []string{"high", "normal", "low"} {
			topics = append(topics, TopicDefinition{
				Name:              fmt.Sprintf("tasks.%s.%s", priority, agentType),
				Partitions:        3,
				ReplicationFactor: 2,
			})
		}
	}

	// Add adapter-specific topics
	if isAdapterAgent(agentType) {
		topics = append(topics, TopicDefinition{
			Name:              fmt.Sprintf("system.adapter.%s", strings.ReplaceAll(agentType, "-", ".")),
			Partitions:        3,
			ReplicationFactor: 2,
		})
	}

	// Add special topics for reasoning agents
	if agentType == "reasoning" {
		topics = append(topics, TopicDefinition{
			Name:              "system.agent.reasoning.process",
			Partitions:        6, // More partitions for higher throughput
			ReplicationFactor: 2,
		})
	}

	return topics
}

// isDataDrivenAgent checks if an agent is data-driven type
func isDataDrivenAgent(agentType string) bool {
	dataDrivenTypes := map[string]bool{
		"copywriter":      true,
		"researcher":      true,
		"content-creator": true,
		"summarizer":      true,
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
	}
	return adapterTypes[agentType]
}

// CreateSystemTopics creates all system-level topics
func (tm *TopicManager) CreateSystemTopics(ctx context.Context) error {
	systemTopics := []TopicDefinition{
		// Orchestration topics
		{Name: "orchestrator.state-changes", Partitions: 6, ReplicationFactor: 2},
		{Name: "orchestrator.commands", Partitions: 3, ReplicationFactor: 2},

		// Human interaction topics
		{Name: "human.approvals", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.commands.workflow.resume", Partitions: 3, ReplicationFactor: 2},

		// System events and monitoring
		{Name: "system.events", Partitions: 6, ReplicationFactor: 2},
		{Name: "system.notifications.ui", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.errors", Partitions: 3, ReplicationFactor: 2},
		{Name: "system.metrics", Partitions: 3, ReplicationFactor: 2},

		// Audit and compliance
		{Name: "audit.log", Partitions: 6, ReplicationFactor: 3}, // Higher replication for audit
		{Name: "compliance.events", Partitions: 3, ReplicationFactor: 3},
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
