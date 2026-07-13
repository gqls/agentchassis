// test/integration/database/state_persistence_test.go
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOrchestratorStatePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	// Use test UUID generator for integration tests
	correlationID := helpers.TestUUIDWithType("integration")
	// This creates something like: 00000002-1234-5678-9abc-def012345678

	// Test state creation
	workflowPlan := map[string]interface{}{
		"start_step": "init",
		"steps": map[string]interface{}{
			"init": map[string]interface{}{
				"action":    "initialize",
				"next_step": "process",
			},
			"process": map[string]interface{}{
				"action":    "process_data",
				"next_step": "complete",
			},
			"complete": map[string]interface{}{
				"action": "complete_workflow",
			},
		},
	}

	workflowPlanJSON, _ := json.Marshal(workflowPlan)

	// Insert state with all required fields
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, current_step, workflow_plan, execution_metadata)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, correlationID, "test_client", "RUNNING", "init", workflowPlanJSON,
		json.RawMessage(`{"completed_steps": 0, "total_steps": 3}`))
	require.NoError(t, err)

	// Update state
	_, err = db.Exec(`
        UPDATE orchestrator_state 
        SET status = $1, current_step = $2, 
            execution_metadata = jsonb_set(execution_metadata, '{completed_steps}', '1')
        WHERE correlation_id = $3
    `, "RUNNING", "process", correlationID)
	require.NoError(t, err)

	// Read and verify
	var status, currentStep string
	var metadata json.RawMessage
	err = db.QueryRow(`
        SELECT status, current_step, execution_metadata
        FROM orchestrator_state
        WHERE correlation_id = $1
    `, correlationID).Scan(&status, &currentStep, &metadata)
	require.NoError(t, err)

	assert.Equal(t, "RUNNING", status)
	assert.Equal(t, "process", currentStep)

	var meta map[string]interface{}
	json.Unmarshal(metadata, &meta)
	assert.Equal(t, float64(1), meta["completed_steps"])
}

func TestCollectedDataAccumulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	correlationID := helpers.TestUUIDWithType("integration")

	// Initialize state WITH current_step (required field)
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, current_step, workflow_plan, collected_data)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, correlationID, "test_client", "RUNNING", "research", // Added current_step
		json.RawMessage(`{"start_step": "research", "steps": {"research": {"action": "research"}}}`),
		json.RawMessage(`{}`))
	require.NoError(t, err)

	// Accumulate data from multiple steps
	steps := []struct {
		step string
		data map[string]interface{}
	}{
		{
			step: "research",
			data: map[string]interface{}{
				"sources": []string{"source1", "source2"},
				"summary": "Research findings",
			},
		},
		{
			step: "analysis",
			data: map[string]interface{}{
				"insights": []string{"insight1", "insight2"},
				"score":    0.85,
			},
		},
		{
			step: "output",
			data: map[string]interface{}{
				"result":     "Final output",
				"confidence": 0.92,
			},
		},
	}

	for i, s := range steps {
		dataJSON, _ := json.Marshal(s.data)

		// Also update current_step as we progress
		currentStep := s.step
		if i == len(steps)-1 {
			currentStep = "complete"
		}

		_, err := db.Exec(`
            UPDATE orchestrator_state 
            SET collected_data = jsonb_set(
                collected_data, 
                $1::text[], 
                $2::jsonb
            ),
            current_step = $4
            WHERE correlation_id = $3
        `, fmt.Sprintf("{%s}", s.step), dataJSON, correlationID, currentStep)
		require.NoError(t, err)
	}

	// Verify accumulated data
	var collectedData json.RawMessage
	err = db.QueryRow(`
        SELECT collected_data FROM orchestrator_state WHERE correlation_id = $1
    `, correlationID).Scan(&collectedData)
	require.NoError(t, err)

	var data map[string]interface{}
	json.Unmarshal(collectedData, &data)

	assert.Contains(t, data, "research")
	assert.Contains(t, data, "analysis")
	assert.Contains(t, data, "output")

	research := data["research"].(map[string]interface{})
	assert.Equal(t, "Research findings", research["summary"])
}

func TestExecutionPathTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	correlationID := helpers.TestUUIDWithType("integration")

	// Initialize with empty execution path AND current_step (required)
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, current_step, workflow_plan, execution_path)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, correlationID, "test_client", "RUNNING", "init", // Added current_step
		json.RawMessage(`{"start_step": "init", "steps": {"init": {"action": "initialize"}}}`),
		json.RawMessage(`[]`))
	require.NoError(t, err)

	// Add execution steps
	executionSteps := []map[string]interface{}{
		{
			"step":       "init",
			"action":     "initialize",
			"start_time": time.Now().Add(-5 * time.Minute),
			"end_time":   time.Now().Add(-4 * time.Minute),
			"result":     "success",
		},
		{
			"step":       "process",
			"action":     "process_data",
			"start_time": time.Now().Add(-4 * time.Minute),
			"end_time":   time.Now().Add(-2 * time.Minute),
			"result":     "success",
			"output": map[string]interface{}{
				"processed_count": 100,
			},
		},
		{
			"step":       "validate",
			"action":     "validate_output",
			"start_time": time.Now().Add(-2 * time.Minute),
			"end_time":   time.Now().Add(-1 * time.Minute),
			"result":     "success",
		},
	}

	for i, step := range executionSteps {
		stepJSON, _ := json.Marshal(step)

		// Update current_step as we add execution steps
		currentStep := step["step"].(string)
		if i == len(executionSteps)-1 {
			currentStep = "complete"
		}

		_, err := db.Exec(`
            UPDATE orchestrator_state 
            SET execution_path = execution_path || $1::jsonb,
                current_step = $3
            WHERE correlation_id = $2
        `, stepJSON, correlationID, currentStep)
		require.NoError(t, err)
	}

	// Verify execution path
	var executionPath json.RawMessage
	var currentStep string
	err = db.QueryRow(`
        SELECT execution_path, current_step FROM orchestrator_state WHERE correlation_id = $1
    `, correlationID).Scan(&executionPath, &currentStep)
	require.NoError(t, err)

	var path []map[string]interface{}
	json.Unmarshal(executionPath, &path)

	assert.Len(t, path, 3)
	assert.Equal(t, "init", path[0]["step"])
	assert.Equal(t, "process", path[1]["step"])
	assert.Equal(t, "validate", path[2]["step"])
	assert.Equal(t, "complete", currentStep)
}

func TestStateRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	logger := zap.NewNop()
	repo := orchestration.NewStateRepository(db, logger)
	ctx := context.Background()

	// Test using the actual repository
	correlationID := helpers.TestUUIDWithType("integration")
	clientID := "test_client"

	plan := models.WorkflowPlan{
		StartStep: "init",
		Steps: map[string]models.Step{
			"init": {
				Action:   "initialize",
				NextStep: "process",
			},
			"process": {
				Action:   "process_data",
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	// Create initial state
	err := repo.CreateInitialState(ctx, correlationID, clientID, plan, nil)
	require.NoError(t, err)

	// Get state
	state, err := repo.GetState(ctx, correlationID)
	require.NoError(t, err)
	assert.Equal(t, correlationID, state.CorrelationID)
	assert.Equal(t, "init", state.CurrentStep)
	assert.Equal(t, orchestration.StatusRunning, state.Status)

	// Update state
	state.CurrentStep = "process"
	state.CollectedData = map[string]interface{}{
		"step1_result": "success",
		"data": map[string]interface{}{
			"count": 42,
			"items": []string{"a", "b", "c"},
		},
	}

	err = repo.UpdateState(ctx, state)
	require.NoError(t, err)

	// Verify update
	updatedState, err := repo.GetState(ctx, correlationID)
	require.NoError(t, err)
	assert.Equal(t, "process", updatedState.CurrentStep)
	assert.Equal(t, 42.0, updatedState.CollectedData["data"].(map[string]interface{})["count"])

	// Add execution record
	record := orchestration.ExecutionRecord{
		Step:      "init",
		Action:    "initialize",
		StartTime: time.Now().Add(-1 * time.Minute),
		EndTime:   timePtr(time.Now()),
		Result:    "success",
	}

	err = repo.AddExecutionRecord(ctx, updatedState, record)
	require.NoError(t, err)

	// Verify execution record was added
	finalState, err := repo.GetState(ctx, correlationID)
	require.NoError(t, err)
	assert.Len(t, finalState.ExecutionPath, 1)
	assert.Equal(t, "init", finalState.ExecutionPath[0].Step)
}

// Helper function
func timePtr(t time.Time) *time.Time {
	return &t
}
