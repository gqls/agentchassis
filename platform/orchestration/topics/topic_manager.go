package topics

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func GenerateJobTopic(correlationID, orchestrationID, stepName string) string {
	return fmt.Sprintf("job.%s.%s.%s",
		truncateID(correlationID, 8),
		truncateID(orchestrationID, 8),
		stepName)
}

func truncateID(id string, length int) string {
	if len(id) > length {
		return id[:length]
	}
	return id
}

func CreateJobTopic(brokers []string, topicName string, ttlHours int) error {
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return err
	}
	defer adminClient.Close()

	topicSpec := kafka.TopicSpecification{
		Topic:             topicName,
		NumPartitions:     1,
		ReplicationFactor: 1,
		Config: map[string]string{
			"retention.ms": fmt.Sprintf("%d", ttlHours*3600000),
		},
	}

	_, err = adminClient.CreateTopics(
		context.Background(),
		[]kafka.TopicSpecification{topicSpec},
		kafka.SetAdminOperationTimeout(10*time.Second))

	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTopicAlreadyExists {
			return nil // Ignore already exists
		}
	}
	return err
}
