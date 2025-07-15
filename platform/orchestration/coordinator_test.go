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
	clientID := "test_client_123"
	headers := map[string]string{
		"correlation_id":      correlationID,
		"request_id":          uuid.NewString(),
		"client_id":           clientID,
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
			clientID,         // client_id
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
		"correlation_id", "client_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, clientID, StatusRunning, "step1", "[]",
		"{}", initialData, nil, nil,
		time.Now(), time.Now(),
	)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Expect Kafka message production - verify headers are properly set
	mockProducer.On("Produce", ctx, "topic.do_something", mock.MatchedBy(func(h map[string]string) bool {
		// Verify required headers are present and correct
		return h["correlation_id"] == correlationID &&
			h["causation_id"] == headers["request_id"] &&
			h["request_id"] != "" && h["request_id"] != headers["request_id"] &&
			h[governance.FuelHeader] == "999" // 1000 - 1 for default_step cost
	}), []byte(correlationID), mock.AnythingOfType("[]uint8")).Return(nil).Once()

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

// TestExecuteWorkflow_AlreadyCompleted verifies handling of already completed workflows.
func TestExecuteWorkflow_AlreadyCompleted(t *testing.T) {
	coordinator, _, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	clientID := "test_client_123"
	headers := map[string]string{
		"correlation_id": correlationID,
		"client_id":      clientID,
	}

	plan := models.WorkflowPlan{
		StartStep: "step1",
		Steps: map[string]models.Step{
			"step1": {Action: "do_something"},
		},
	}

	// State already exists and is completed
	rows := sqlmock.NewRows([]string{
		"correlation_id", "client_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, clientID, StatusCompleted, "step1", "[]",
		"{}", nil, []byte(`{"result": "done"}`), nil,
		time.Now(), time.Now(),
	)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Should not do anything else
	err := coordinator.ExecuteWorkflow(ctx, plan, headers, nil)
	require.NoError(t, err)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestExecuteWorkflow_MissingClientID verifies error when client_id is missing.
func TestExecuteWorkflow_MissingClientID(t *testing.T) {
	coordinator, _, db, _ := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	headers := map[string]string{
		"correlation_id": uuid.NewString(),
		// client_id is missing
	}

	plan := models.WorkflowPlan{
		StartStep: "step1",
		Steps:     map[string]models.Step{},
	}

	err := coordinator.ExecuteWorkflow(ctx, plan, headers, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client_id header is required")
}

// TestExecuteWorkflow_DependenciesNotMet verifies the workflow waits correctly.
func TestExecuteWorkflow_DependenciesNotMet(t *testing.T) {
	coordinator, _, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	clientID := "test_client_123"
	headers := map[string]string{
		"correlation_id":      correlationID,
		"client_id":           clientID,
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
			clientID,         // client_id
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
		"correlation_id", "client_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, clientID, StatusRunning, "step2", "[]",
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
	clientID := "test_client_123"
	headers := map[string]string{
		"correlation_id":      correlationID,
		"client_id":           clientID,
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
			clientID,         // client_id
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
		"correlation_id", "client_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, clientID, StatusRunning, "step1", "[]",
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
	clientID := "test_client_123"
	headers := map[string]string{
		"correlation_id":      correlationID,
		"client_id":           clientID,
		"request_id":          "parent_req_1",
		governance.FuelHeader: "995", // Already deducted by ExecuteWorkflow (1000 - 5)
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
		ClientID:      clientID,
		CollectedData: make(map[string]interface{}),
	}

	// Expect messages to be produced with validation
	var capturedRequestIDs []string
	mockProducer.On("Produce", ctx, "topic.research", mock.MatchedBy(func(h map[string]string) bool {
		if h["causation_id"] == "parent_req_1" && h["request_id"] != "parent_req_1" &&
			h[governance.FuelHeader] == "995" { // Fuel already deducted
			capturedRequestIDs = append(capturedRequestIDs, h["request_id"])
			return true
		}
		return false
	}), []byte(correlationID), mock.AnythingOfType("[]uint8")).Return(nil).Once()

	mockProducer.On("Produce", ctx, "topic.style", mock.MatchedBy(func(h map[string]string) bool {
		if h["causation_id"] == "parent_req_1" && h["request_id"] != "parent_req_1" &&
			h[governance.FuelHeader] == "995" { // Fuel already deducted
			capturedRequestIDs = append(capturedRequestIDs, h["request_id"])
			return true
		}
		return false
	}), []byte(correlationID), mock.AnythingOfType("[]uint8")).Return(nil).Once()

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

	// Verify state was updated correctly
	assert.Equal(t, StatusAwaitingResponses, state.Status)
	assert.Equal(t, "aggregate_results", state.CurrentStep)
	assert.Len(t, state.AwaitedSteps, 2)

	mockProducer.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestHandlePauseForHumanInput verifies human approval pause functionality.
func TestHandlePauseForHumanInput(t *testing.T) {
	coordinator, mockProducer, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	projectID := "project_123"
	clientID := "client_123"
	headers := map[string]string{
		"correlation_id": correlationID,
		"project_id":     projectID,
		"client_id":      clientID,
	}

	step := models.Step{
		Action:      "pause_for_human_input",
		NextStep:    "after_approval",
		Description: "Review generated content",
	}
	state := &OrchestrationState{
		CorrelationID: correlationID,
		ClientID:      clientID,
		CollectedData: map[string]interface{}{
			"generated_content": "Some content to review",
		},
	}

	// Expect state update
	mockDB.ExpectExec("UPDATE orchestrator_state SET").
		WithArgs(
			correlationID,        // WHERE correlation_id = $1
			StatusPausedForHuman, // status = $2
			"after_approval",     // current_step = $3
			sqlmock.AnyArg(),     // awaited_steps = $4
			sqlmock.AnyArg(),     // collected_data = $5
			sqlmock.AnyArg(),     // final_result = $6
			"",                   // error = $7
			sqlmock.AnyArg(),     // updated_at = $8
		).WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect notification to be sent
	mockProducer.On("Produce", ctx, NotificationTopic, headers, []byte(correlationID), mock.MatchedBy(func(payload []byte) bool {
		var notification map[string]interface{}
		json.Unmarshal(payload, &notification)
		return notification["event_type"] == "WORKFLOW_PAUSED_FOR_APPROVAL" &&
			notification["correlation_id"] == correlationID &&
			notification["project_id"] == projectID &&
			notification["client_id"] == clientID
	})).Return(nil).Once()

	err := coordinator.handlePauseForHumanInput(ctx, headers, step, state)
	require.NoError(t, err)

	assert.Equal(t, StatusPausedForHuman, state.Status)
	assert.Equal(t, "after_approval", state.CurrentStep)

	mockProducer.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestHandleResponse verifies processing of sub-task responses.
func TestHandleResponse(t *testing.T) {
	coordinator, _, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	causationID := "request_123"
	headers := map[string]string{
		"correlation_id": correlationID,
		"causation_id":   causationID,
	}

	taskResponse := models.TaskResponse{
		Success: true,
		Data: map[string]interface{}{
			"result": "task completed",
		},
	}
	responseBytes, _ := json.Marshal(taskResponse)

	// Existing state with awaited steps
	existingData := map[string]interface{}{
		"existing": "data",
	}
	existingDataJSON, _ := json.Marshal(existingData)
	awaitedSteps := []string{causationID, "another_request"}
	awaitedStepsJSON, _ := json.Marshal(awaitedSteps)

	rows := sqlmock.NewRows([]string{
		"correlation_id", "client_id", "status", "current_step", "awaited_steps",
		"collected_data", "initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		correlationID, "client_123", StatusAwaitingResponses, "aggregate", string(awaitedStepsJSON),
		string(existingDataJSON), nil, nil, nil,
		time.Now(), time.Now(),
	)
	mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
		WithArgs(correlationID).
		WillReturnRows(rows)

	// Expect state update with response data added
	mockDB.ExpectExec("UPDATE orchestrator_state SET").
		WithArgs(
			correlationID,           // WHERE correlation_id = $1
			StatusAwaitingResponses, // status = $2 (still awaiting one more)
			"aggregate",             // current_step = $3
			sqlmock.AnyArg(),        // awaited_steps = $4 (should have one less)
			sqlmock.AnyArg(),        // collected_data = $5 (should include new data)
			sqlmock.AnyArg(),        // final_result = $6
			"",                      // error = $7
			sqlmock.AnyArg(),        // updated_at = $8
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := coordinator.HandleResponse(ctx, headers, responseBytes)
	require.NoError(t, err)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

// TestResumeWorkflow verifies resuming after human approval.
func TestResumeWorkflow(t *testing.T) {
	coordinator, _, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()
	headers := map[string]string{
		"correlation_id": correlationID,
	}

	t.Run("approved", func(t *testing.T) {
		resumePayload := struct {
			Approved bool                   `json:"approved"`
			Feedback map[string]interface{} `json:"feedback,omitempty"`
		}{
			Approved: true,
			Feedback: map[string]interface{}{
				"comment": "Looks good!",
			},
		}
		resumeData, _ := json.Marshal(resumePayload)

		// Existing paused state
		rows := sqlmock.NewRows([]string{
			"correlation_id", "client_id", "status", "current_step", "awaited_steps",
			"collected_data", "initial_request_data", "final_result", "error",
			"created_at", "updated_at",
		}).AddRow(
			correlationID, "client_123", StatusPausedForHuman, "after_approval", "[]",
			"{}", nil, nil, nil,
			time.Now(), time.Now(),
		)
		mockDB.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
			WithArgs(correlationID).
			WillReturnRows(rows)

		// Expect state update to running
		mockDB.ExpectExec("UPDATE orchestrator_state SET").
			WithArgs(
				correlationID,    // WHERE correlation_id = $1
				StatusRunning,    // status = $2
				"after_approval", // current_step = $3
				sqlmock.AnyArg(), // awaited_steps = $4
				sqlmock.AnyArg(), // collected_data = $5 (should include feedback)
				sqlmock.AnyArg(), // final_result = $6
				"",               // error = $7
				sqlmock.AnyArg(), // updated_at = $8
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err := coordinator.ResumeWorkflow(ctx, headers, resumeData)
		require.NoError(t, err)
		require.NoError(t, mockDB.ExpectationsWereMet())
	})

	t.Run("rejected", func(t *testing.T) {
		// Create fresh mocks for this subtest
		db2, mockDB2, err := sqlmock.New()
		require.NoError(t, err)
		defer db2.Close()

		// Create a new coordinator with the fresh DB
		coordinator2 := NewSagaCoordinator(db2, coordinator.producer, coordinator.logger)

		resumePayload := struct {
			Approved bool `json:"approved"`
		}{
			Approved: false,
		}
		resumeData, _ := json.Marshal(resumePayload)

		// Existing paused state
		rows := sqlmock.NewRows([]string{
			"correlation_id", "client_id", "status", "current_step", "awaited_steps",
			"collected_data", "initial_request_data", "final_result", "error",
			"created_at", "updated_at",
		}).AddRow(
			correlationID, "client_123", StatusPausedForHuman, "after_approval", "[]",
			"{}", nil, nil, nil,
			time.Now(), time.Now(),
		)
		mockDB2.ExpectQuery("SELECT .* FROM orchestrator_state WHERE correlation_id = \\$1").
			WithArgs(correlationID).
			WillReturnRows(rows)

		// Expect state update to failed
		mockDB2.ExpectExec("UPDATE orchestrator_state SET").
			WithArgs(
				correlationID,               // WHERE correlation_id = $1
				StatusFailed,                // status = $2
				"after_approval",            // current_step = $3
				sqlmock.AnyArg(),            // awaited_steps = $4
				sqlmock.AnyArg(),            // collected_data = $5
				sqlmock.AnyArg(),            // final_result = $6
				"Workflow rejected by user", // error = $7
				sqlmock.AnyArg(),            // updated_at = $8
			).WillReturnResult(sqlmock.NewResult(1, 1))

		err = coordinator2.ResumeWorkflow(ctx, headers, resumeData)
		require.NoError(t, err)
		require.NoError(t, mockDB2.ExpectationsWereMet())
	})
}

// TestCompleteWorkflow verifies workflow completion.
func TestCompleteWorkflow(t *testing.T) {
	coordinator, _, db, mockDB := setupTest(t)
	defer db.Close()

	ctx := context.Background()
	correlationID := uuid.NewString()

	state := &OrchestrationState{
		CorrelationID: correlationID,
		ClientID:      "client_123",
		Status:        StatusRunning,
		CurrentStep:   "final",
		CollectedData: map[string]interface{}{
			"step1_result": "data1",
			"step2_result": "data2",
		},
	}

	// Expect state update to completed
	mockDB.ExpectExec("UPDATE orchestrator_state SET").
		WithArgs(
			correlationID,    // WHERE correlation_id = $1
			StatusCompleted,  // status = $2
			"final",          // current_step = $3
			sqlmock.AnyArg(), // awaited_steps = $4
			sqlmock.AnyArg(), // collected_data = $5
			sqlmock.AnyArg(), // final_result = $6 (should be marshaled collected_data)
			"",               // error = $7
			sqlmock.AnyArg(), // updated_at = $8
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := coordinator.completeWorkflow(ctx, state)
	require.NoError(t, err)

	assert.Equal(t, StatusCompleted, state.Status)
	assert.NotNil(t, state.FinalResult)

	require.NoError(t, mockDB.ExpectationsWereMet())
}
