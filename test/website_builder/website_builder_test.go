// test/website_builder_test.go
package test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

func TestWebsiteBuilder(t *testing.T) {
	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create a test request
	request := map[string]interface{}{
		"action": "build_website",
		"data": map[string]interface{}{
			"business_name": "Test Company",
			"domain":        "testcompany.com",
		},
	}

	requestBytes, _ := json.Marshal(request)

	// Create headers
	headers := map[string]string{
		"correlation_id":    uuid.NewString(),
		"request_id":        uuid.NewString(),
		"client_id":         "demo_client",
		"agent_instance_id": "00000000-0000-0000-0000-000000001000",
		"fuel_budget":       "1000",
	}

	// Send to website builder topic
	producer, err := kafka.NewProducer([]string{"localhost:9092"}, logger)
	if err != nil {
		t.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	err = producer.Produce(
		context.Background(),
		"system.agent.website-builder.process",
		headers,
		[]byte(headers["correlation_id"]),
		requestBytes,
	)

	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	t.Logf("Request sent successfully with correlation_id: %s", headers["correlation_id"])
}
