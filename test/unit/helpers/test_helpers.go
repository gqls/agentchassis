// test/unit/helpers/test_helpers.go
package helpers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestDB creates a test database connection
// It first checks if we're in a real environment with DB_HOST set,
// otherwise returns a mock database for unit tests
func TestDB(t *testing.T) *sql.DB {
	// Check if we're running in Kubernetes/Docker with real DB
	dbHost := os.Getenv("DB_HOST")
	if dbHost != "" {
		// Real database connection
		dbUser := os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "clients_user"
		}
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "clients_db"
		}

		connStr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable",
			dbUser, dbPass, dbHost, dbName)

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			t.Logf("Failed to connect to database: %v, using mock", err)
			return MockDB(t)
		}

		// Test the connection
		if err := db.Ping(); err != nil {
			t.Logf("Failed to ping database: %v, using mock", err)
			return MockDB(t)
		}

		// Setup test database schema
		SetupTestSchema(t, db)

		return db
	}

	// Default to mock database for unit tests
	return MockDB(t)
}

// MockDB creates a mock database for unit tests
func MockDB(t *testing.T) *sql.DB {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Set up common expectations that many tests need
	mock.ExpectExec("CREATE SCHEMA IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE SCHEMA IF NOT EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))

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
	if err != nil {
		// Ignore errors in cleanup
		t.Logf("Cleanup warning: %v", err)
	}

	_, err = db.Exec("DELETE FROM client_test_client.agent_instances WHERE name LIKE 'test-%'")
	if err != nil {
		// Ignore errors in cleanup
		t.Logf("Cleanup warning: %v", err)
	}
}

// SetupTestSchema creates necessary tables for testing
func SetupTestSchema(t *testing.T, db *sql.DB) {
	// Create schemas
	schemas := []string{
		"CREATE SCHEMA IF NOT EXISTS client_test_client",
		"CREATE SCHEMA IF NOT EXISTS client_demo_client",
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			t.Logf("Warning: failed to create schema: %v", err)
		}
	}

	// Create orchestrator_state table with proper JSONB defaults
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS orchestrator_state (
        correlation_id UUID PRIMARY KEY,
        client_id VARCHAR(255) NOT NULL,
        status VARCHAR(50) NOT NULL,
        current_step VARCHAR(255),
        awaited_steps JSONB NOT NULL DEFAULT '[]'::jsonb,
        collected_data JSONB NOT NULL DEFAULT '{}'::jsonb,
        initial_request_data JSONB,
        final_result JSONB,
        error TEXT,
        workflow_plan JSONB NOT NULL DEFAULT '{}'::jsonb,
        execution_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
        execution_path JSONB NOT NULL DEFAULT '[]'::jsonb,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    )`

	if _, err := db.Exec(createTableSQL); err != nil {
		// If table exists, try to update it
		alterStatements := []string{
			`ALTER TABLE orchestrator_state ALTER COLUMN awaited_steps SET NOT NULL`,
			`ALTER TABLE orchestrator_state ALTER COLUMN awaited_steps SET DEFAULT '[]'::jsonb`,
			`ALTER TABLE orchestrator_state ALTER COLUMN collected_data SET NOT NULL`,
			`ALTER TABLE orchestrator_state ALTER COLUMN collected_data SET DEFAULT '{}'::jsonb`,
			`ALTER TABLE orchestrator_state ALTER COLUMN workflow_plan SET NOT NULL`,
			`ALTER TABLE orchestrator_state ALTER COLUMN workflow_plan SET DEFAULT '{}'::jsonb`,
			`ALTER TABLE orchestrator_state ALTER COLUMN execution_metadata SET NOT NULL`,
			`ALTER TABLE orchestrator_state ALTER COLUMN execution_metadata SET DEFAULT '{}'::jsonb`,
			`ALTER TABLE orchestrator_state ALTER COLUMN execution_path SET NOT NULL`,
			`ALTER TABLE orchestrator_state ALTER COLUMN execution_path SET DEFAULT '[]'::jsonb`,
		}

		for _, stmt := range alterStatements {
			_, _ = db.Exec(stmt) // Ignore errors, some columns might already be correct
		}

		t.Logf("Note: orchestrator_state table already exists, attempted updates")
	}

	// Create agent_instances tables
	createAgentTables := []string{
		`CREATE TABLE IF NOT EXISTS client_test_client.agent_instances (
            id UUID PRIMARY KEY,
            template_id UUID NOT NULL,
            owner_user_id VARCHAR(255) NOT NULL,
            name VARCHAR(255) NOT NULL,
            config JSONB NOT NULL DEFAULT '{}'::jsonb,
            is_active BOOLEAN DEFAULT true,
            last_heartbeat TIMESTAMP,
            created_at TIMESTAMP DEFAULT NOW(),
            updated_at TIMESTAMP DEFAULT NOW()
        )`,
		`CREATE TABLE IF NOT EXISTS client_demo_client.agent_instances (
            id UUID PRIMARY KEY,
            template_id UUID NOT NULL,
            owner_user_id VARCHAR(255) NOT NULL,
            name VARCHAR(255) NOT NULL,
            config JSONB NOT NULL DEFAULT '{}'::jsonb,
            is_active BOOLEAN DEFAULT true,
            last_heartbeat TIMESTAMP,
            created_at TIMESTAMP DEFAULT NOW(),
            updated_at TIMESTAMP DEFAULT NOW()
        )`,
	}

	for _, createTable := range createAgentTables {
		if _, err := db.Exec(createTable); err != nil {
			t.Logf("Warning: failed to create agent table: %v", err)
		}
	}

	// Create agent_groups table with version column
	createGroupsSQL := `
    CREATE TABLE IF NOT EXISTS agent_groups (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL,
        group_type VARCHAR(100) NOT NULL,
        version VARCHAR(20) DEFAULT '1.0.0',
        parent_id UUID,
        agent_configs JSONB NOT NULL DEFAULT '[]'::jsonb,
        orchestration_workflow JSONB NOT NULL DEFAULT '{}'::jsonb,
        capabilities JSONB DEFAULT '[]'::jsonb,
        usage_count INTEGER DEFAULT 0,
        performance_metrics JSONB DEFAULT '{}'::jsonb,
        mutation_history JSONB DEFAULT '[]'::jsonb,
        last_used_at TIMESTAMP,
        is_active BOOLEAN DEFAULT true,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW()
    )`

	if _, err := db.Exec(createGroupsSQL); err != nil {
		// Try to add version column if table exists but column doesn't
		alterSQL := `ALTER TABLE agent_groups ADD COLUMN IF NOT EXISTS version VARCHAR(20) DEFAULT '1.0.0'`
		if _, err2 := db.Exec(alterSQL); err2 != nil {
			t.Logf("Warning: failed to create/update agent_groups table: %v, %v", err, err2)
		}
	}

	// Create agent_metrics table
	createMetricsSQL := `
    CREATE TABLE IF NOT EXISTS agent_metrics (
        agent_id UUID NOT NULL,
        success_rate FLOAT DEFAULT 0.5,
        avg_response_time INTEGER,
        total_requests INTEGER DEFAULT 0,
        failed_requests INTEGER DEFAULT 0,
        last_updated TIMESTAMP DEFAULT NOW(),
        PRIMARY KEY (agent_id)
    )`

	if _, err := db.Exec(createMetricsSQL); err != nil {
		t.Logf("Warning: failed to create agent_metrics table: %v", err)
	}
}
