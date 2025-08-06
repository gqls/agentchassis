// test/unit/helpers/test_helpers.go
package helpers

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestDB creates a test database connection
func TestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", "postgres://clients_user:password@localhost:5432/clients_db?sslmode=disable")
	require.NoError(t, err)

	// Ensure test schema exists
	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS client_test_client")
	require.NoError(t, err)

	return db
}

// MockProducer implements kafka.Producer for testing
type MockProducer struct {
	Messages []ProducedMessage
}

type ProducedMessage struct {
	Topic   string
	Headers map[string]string
	Key     []byte
	Value   []byte
}

func NewMockProducer() *MockProducer {
	return &MockProducer{
		Messages: make([]ProducedMessage, 0),
	}
}

func (m *MockProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	m.Messages = append(m.Messages, ProducedMessage{
		Topic:   topic,
		Headers: headers,
		Key:     key,
		Value:   value,
	})
	return nil
}

func (m *MockProducer) Close() error {
	return nil
}

// TestHeaders creates standard test headers
func TestHeaders(correlationID string) map[string]string {
	return map[string]string{
		"correlation_id":    correlationID,
		"request_id":        uuid.New().String(),
		"client_id":         "test_client",
		"user_id":           "test_user",
		"fuel_budget":       "1000",
		"agent_instance_id": "test-instance",
	}
}

// ValidWorkflow creates a valid test workflow
func ValidWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "validate",
		Steps: map[string]models.Step{
			"validate": {
				Action:      "validate_input",
				Description: "Validate input",
				NextStep:    "process",
			},
			"process": {
				Action:      "process_data",
				Description: "Process data",
				NextStep:    "complete",
			},
			"complete": {
				Action:      "complete_workflow",
				Description: "Complete workflow",
			},
		},
	}
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(t *testing.T, timeout time.Duration, check func() bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("Condition not met within timeout")
}

// LoadFixture loads a JSON fixture file
func LoadFixture(t *testing.T, path string, v interface{}) {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	err = json.Unmarshal(data, v)
	require.NoError(t, err)
}

// CleanupTestData removes test data from database
func CleanupTestData(t *testing.T, db *sql.DB, correlationID string) {
	_, err := db.Exec("DELETE FROM orchestrator_state WHERE correlation_id LIKE 'test-%'")
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM client_test_client.agent_instances WHERE name LIKE 'test-%'")
	require.NoError(t, err)
}
