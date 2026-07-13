// test/integration/kafka/kafka_integration_test.go
package kafka

import (
	"context"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

// getKafkaBrokers returns the appropriate Kafka brokers
func getKafkaBrokers() []string {
	if kafkaBroker := os.Getenv("KAFKA_BROKERS"); kafkaBroker != "" {
		return []string{kafkaBroker}
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// In Kubernetes, use the service name
		return []string{"personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"}
	}

	return []string{"localhost:9092"}
}

// isKafkaAvailable checks if Kafka is reachable
func isKafkaAvailable() bool {
	brokers := getKafkaBrokers()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		return false
	}
	defer conn.Close()

	// Try to get API versions as a health check
	_, err = conn.ApiVersions()
	return err == nil
}
