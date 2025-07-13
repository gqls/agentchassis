// FILE: platform/orchestration/coordinator_test.go
package orchestration

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/ai-persona-system/pkg/models"
	"github.com/gqls/ai-persona-system/platform/governance"
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
	db, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
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

	// Expect the coordinator to check if state exists (it won't).
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = ?").
		WithArgs(correlationID).
		WillReturnError(sql.ErrNoRows)

	// Expect it to create the initial state.
	mockDB.ExpectExec("INSERT INTO orchestrator_state").
		WithArgs(correlationID, StatusRunning, "step1", mock.Anything, mock.Anything, initialData, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect it to fetch the state again after creation.
	rows := sqlmock.NewRows([]string{"correlation_id", "status", "current_step", "awaited_steps", "collected_data", "initial_request_data"}).
		AddRow(correlationID, StatusRunning, "step1", "[]", "{}", initialData)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = ?").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Expect it to produce a Kafka message for the first step.
	mockProducer.On("Produce", ctx, "topic.do_something", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	// Expect it to update the state to AWAITING_RESPONSES.
	mockDB.ExpectExec("UPDATE orchestrator_state SET").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := coordinator.ExecuteWorkflow(ctx, plan, headers, initialData)
	require.NoError(t, err)

	mockProducer.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestExecuteWorkflow_DependenciesNotMet verifies the workflow waits correctly.
func TestExecuteWorkflow_DependenciesNotMet(t *testing.T) {
	coordinator, mockProducer, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	headers := map[string]string{"correlation_id": correlationID, governance.FuelHeader: "1000"}

	plan := models.WorkflowPlan{
		StartStep: "step2", // Start at a step that has dependencies
		Steps: map[string]models.Step{
			"step1": {Action: "do_something"},
			"step2": {Action: "do_something_else", Dependencies: []string{"step1"}},
		},
	}

	// Mock the state where step2 is current, but step1 has not responded.
	stateJSON := `{"get_other_data": {"status": "complete"}}` // Missing "step1" data
	rows := sqlmock.NewRows([]string{"correlation_id", "status", "current_step", "awaited_steps", "collected_data"}).
		AddRow(correlationID, StatusRunning, "step2", "[]", stateJSON)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = ?").
		WithArgs(correlationID).
		WillReturnRows(rows)

	err := coordinator.ExecuteWorkflow(ctx, plan, headers, nil)
	require.NoError(t, err, "Waiting for dependencies should not be an error")

	// Crucially, the producer should NOT have been called.
	mockProducer.AssertNotCalled(t, "Produce")
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestExecuteWorkflow_FuelCheckFail verifies that a workflow stops if out of fuel.
func TestExecuteWorkflow_FuelCheckFail(t *testing.T) {
	coordinator, mockProducer, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	// Provide a low fuel budget.
	headers := map[string]string{"correlation_id": correlationID, governance.FuelHeader: "5"}

	plan := models.WorkflowPlan{
		StartStep: "step1",
		Steps: map[string]models.Step{
			// The cost of this action (50) is higher than the budget (5).
			"step1": {Action: "ai_text_generate_claude_opus", Topic: "topic.expensive"},
		},
	}

	rows := sqlmock.NewRows([]string{"correlation_id", "status", "current_step", "awaited_steps", "collected_data"}).
		AddRow(correlationID, StatusRunning, "step1", "[]", "{}")
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = ?").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Expect an UPDATE to set the state to FAILED.
	mockDB.ExpectExec("UPDATE orchestrator_state SET status = ?, error = ?").
		WithArgs(StatusFailed, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := coordinator.ExecuteWorkflow(ctx, plan, headers, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient fuel")

	mockProducer.AssertNotCalled(t, "Produce")
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
	state := &OrchestrationState{CorrelationID: correlationID}
	repo := NewStateRepository(db, zap.NewNop())

	// Expect a message to be produced to each sub-task topic.
	mockProducer.On("Produce", ctx, "topic.research", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockProducer.On("Produce", ctx, "topic.style", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	// Expect the state to be updated to AWAITING_RESPONSES with the correct next step.
	mockDB.ExpectExec(regexp.QuoteMeta("UPDATE orchestrator_state SET status = ?, current_step = ?, awaited_steps = ?")).
		WithArgs(StatusAwaitingResponses, "aggregate_results", mock.Anything).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := coordinator.handleFanOut(ctx, headers, step, state, repo)
	require.NoError(t, err)

	mockProducer.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}
