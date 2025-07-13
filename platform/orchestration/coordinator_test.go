// FILE: platform/orchestration/coordinator_test.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/governance"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockKafkaProducer allows us to test the coordinator without a real Kafka connection.
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) Produce(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	args := m.Called(ctx, topic, headers, key, value)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// setupTest creates the coordinator with mocked dependencies for testing.
func setupTest(t *testing.T) (*SagaCoordinator, *MockKafkaProducer, *sql.DB, sqlmock.Sqlmock) {
	db, mockDB, err := sqlmock.New()
	require.NoError(t, err)

	mockProducer := new(MockKafkaProducer)
	logger := zap.NewNop()

	coordinator := NewSagaCoordinator(db, mockProducer, logger)
	require.NotNil(t, coordinator)

	return coordinator, mockProducer, db, mockDB
}

// TestExecuteWorkflow_InitialStep verifies the start of a new workflow.
func TestExecuteWorkflow_InitialStep(t *testing.T) {
	coordinator, mockProducer, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	headers := map[string]string{
		"correlation_id":      correlationID,
		"request_id":          uuid.NewString(),
		governance.FuelHeader: "1000",
	}
	initialData, _ := json.Marshal(map[string]string{"goal": "test"})

	plan := models.WorkflowPlan{
		StartStep: "step1",
		Steps: map[string]models.Step{
			"step1":  {Action: "do_something", Topic: "topic.do_something", NextStep: "finish"},
			"finish": {Action: "complete_workflow"},
		},
	}

	// First, expect check if state exists (it won't)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnError(sql.ErrNoRows)

	// Then expect creation of initial state
	mockDB.ExpectExec("INSERT INTO orchestrator_state").
		WithArgs(
			correlationID,    // correlation_id
			StatusRunning,    // status
			"step1",          // current_step
			sqlmock.AnyArg(), // awaited_steps
			sqlmock.AnyArg(), // collected_data
			initialData,      // initial_request_data
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Then expect fetch of the newly created state
	rows := sqlmock.NewRows([]string{
		"correlation_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, StatusRunning, "step1", "[]",
		"{}", initialData, nil, nil, // Use nil for NULL values
		time.Now(), time.Now(),
	)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Expect Kafka message production
	mockProducer.On("Produce", ctx, "topic.do_something", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	// Expect state update
	mockDB.ExpectExec("UPDATE orchestrator_state SET").
		WithArgs(
			correlationID,           // WHERE correlation_id = $1
			StatusAwaitingResponses, // status = $2
			"finish",                // current_step = $3
			sqlmock.AnyArg(),        // awaited_steps = $4
			sqlmock.AnyArg(),        // collected_data = $5
			sqlmock.AnyArg(),        // final_result = $6
			"",                      // error = $7
			sqlmock.AnyArg(),        // updated_at = $8
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := coordinator.ExecuteWorkflow(ctx, plan, headers, initialData)
	require.NoError(t, err)

	mockProducer.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestExecuteWorkflow_DependenciesNotMet verifies the workflow waits correctly.
func TestExecuteWorkflow_DependenciesNotMet(t *testing.T) {
	coordinator, _, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	headers := map[string]string{
		"correlation_id":      correlationID,
		governance.FuelHeader: "1000",
	}

	plan := models.WorkflowPlan{
		StartStep: "step2",
		Steps: map[string]models.Step{
			"step1": {Action: "do_something"},
			"step2": {Action: "do_something_else", Dependencies: []string{"step1"}},
		},
	}

	// First check - state doesn't exist
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnError(sql.ErrNoRows)

	// Create initial state
	mockDB.ExpectExec("INSERT INTO orchestrator_state").
		WithArgs(
			correlationID,    // correlation_id
			StatusRunning,    // status
			"step2",          // current_step
			sqlmock.AnyArg(), // awaited_steps
			sqlmock.AnyArg(), // collected_data
			sqlmock.AnyArg(), // initial_request_data (nil)
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Fetch state - missing step1 dependency
	stateJSON := `{}` // No step1 data
	rows := sqlmock.NewRows([]string{
		"correlation_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, StatusRunning, "step2", "[]",
		stateJSON, nil, nil, nil,
		time.Now(), time.Now(),
	)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Should not produce any messages or update state
	err := coordinator.ExecuteWorkflow(ctx, plan, headers, nil)
	require.NoError(t, err, "Waiting for dependencies should not be an error")

	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestExecuteWorkflow_FuelCheckFail verifies that a workflow stops if out of fuel.
func TestExecuteWorkflow_FuelCheckFail(t *testing.T) {
	coordinator, _, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	headers := map[string]string{
		"correlation_id":      correlationID,
		governance.FuelHeader: "5", // Low fuel (need 50 for claude opus)
	}

	plan := models.WorkflowPlan{
		StartStep: "step1",
		Steps: map[string]models.Step{
			"step1": {Action: "ai_text_generate_claude_opus", Topic: "topic.expensive"},
		},
	}

	// First check - state doesn't exist
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnError(sql.ErrNoRows)

	// Create initial state
	mockDB.ExpectExec("INSERT INTO orchestrator_state").
		WithArgs(
			correlationID,    // correlation_id
			StatusRunning,    // status
			"step1",          // current_step
			sqlmock.AnyArg(), // awaited_steps
			sqlmock.AnyArg(), // collected_data
			sqlmock.AnyArg(), // initial_request_data
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Fetch state
	rows := sqlmock.NewRows([]string{
		"correlation_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, StatusRunning, "step1", "[]",
		"{}", nil, nil, nil,
		time.Now(), time.Now(),
	)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Expect update to FAILED status
	mockDB.ExpectExec("UPDATE orchestrator_state SET").
		WithArgs(
			correlationID,    // WHERE correlation_id = $1
			StatusFailed,     // status = $2
			"step1",          // current_step = $3
			sqlmock.AnyArg(), // awaited_steps = $4
			sqlmock.AnyArg(), // collected_data = $5
			sqlmock.AnyArg(), // final_result = $6
			sqlmock.AnyArg(), // error = $7 (will contain "insufficient fuel")
			sqlmock.AnyArg(), // updated_at = $8
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute the workflow - it should fail with insufficient fuel error
	err := coordinator.ExecuteWorkflow(ctx, plan, headers, nil)

	// Assert that we got an error
	require.Error(t, err, "Expected an error for insufficient fuel")
	assert.Contains(t, err.Error(), "insufficient fuel", "Error should mention insufficient fuel")

	// Verify all expectations were met
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestHandleFanOut verifies that multiple messages are sent in parallel.
func TestHandleFanOut(t *testing.T) {
	coordinator, mockProducer, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	headers := map[string]string{
		"correlation_id":      correlationID,
		"request_id":          "parent_req_1",
		governance.FuelHeader: "1000",
	}

	step := models.Step{
		Action:   "fan_out",
		NextStep: "aggregate_results",
		SubTasks: []models.SubTask{
			{StepName: "get_research", Topic: "topic.research"},
			{StepName: "get_style", Topic: "topic.style"},
		},
	}
	state := &OrchestrationState{
		CorrelationID: correlationID,
		CollectedData: make(map[string]interface{}), // Initialize the map
	}

	// Expect messages to be produced
	mockProducer.On("Produce", ctx, "topic.research", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockProducer.On("Produce", ctx, "topic.style", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	// Expect state update
	mockDB.ExpectExec("UPDATE orchestrator_state SET").
		WithArgs(
			correlationID,           // WHERE correlation_id = $1
			StatusAwaitingResponses, // status = $2
			"aggregate_results",     // current_step = $3
			sqlmock.AnyArg(),        // awaited_steps = $4
			sqlmock.AnyArg(),        // collected_data = $5
			sqlmock.AnyArg(),        // final_result = $6
			"",                      // error = $7
			sqlmock.AnyArg(),        // updated_at = $8
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := coordinator.handleFanOut(ctx, headers, step, state)
	require.NoError(t, err)

	mockProducer.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}
