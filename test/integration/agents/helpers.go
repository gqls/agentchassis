// test/integration/agents/helpers.go
package agents

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/test/unit/helpers"
)

// setupIntegrationDB sets up the database for integration tests
func setupIntegrationDB(t *testing.T) *sql.DB {
	db := helpers.TestDB(t)
	helpers.SetupTestSchema(t, db)
	return db
}

// createTestProducer creates a Kafka producer for testing
func createTestProducer(t *testing.T) kafka.Producer {
	// For integration tests, you might want to use a real Kafka producer
	// if Kafka is available in your test environment
	if isKafkaAvailable() {
		return createRealKafkaProducer(t)
	}
	// Otherwise fall back to mock
	return helpers.NewMockProducer()
}

// createRealKafkaProducer creates a real Kafka producer if available
func createRealKafkaProducer(t *testing.T) kafka.Producer {
	// This would create a real Kafka producer
	// Implementation depends on your Kafka setup
	// For now, fall back to mock
	t.Log("Real Kafka not configured, using mock producer")
	return helpers.NewMockProducer()
}

// isKafkaAvailable checks if Kafka is available for testing
func isKafkaAvailable() bool {
	// Check if Kafka is available
	// This could check environment variables or try to connect
	// For now, return false to use mocks
	return false
}

// Helper functions for creating test data
func createAgentConfig(role, agentType string) map[string]interface{} {
	return map[string]interface{}{
		"role":       role,
		"agent_type": agentType,
	}
}

func createStep(action string, nextStep string) map[string]interface{} {
	step := map[string]interface{}{
		"action": action,
	}
	if nextStep != "" {
		step["next_step"] = nextStep
	}
	return step
}

func createWorkflowJSON(startStep string, steps map[string]map[string]interface{}) []byte {
	workflow := map[string]interface{}{
		"start_step": startStep,
		"steps":      steps,
	}
	data, _ := json.Marshal(workflow)
	return data
}
