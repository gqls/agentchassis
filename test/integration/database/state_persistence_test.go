// test/integration/database/state_persistence_test.go
package database

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestratorStatePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	correlationID := "test-db-" + uuid.New().String()

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

	// Insert state
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

	correlationID := "test-collect-" + uuid.New().String()

	// Initialize state
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, workflow_plan, collected_data)
        VALUES ($1, $2, $3, $4, $5)
    `, correlationID, "test_client", "RUNNING",
		json.RawMessage(`{"start_step": "init"}`),
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

	for _, s := range steps {
		dataJSON, _ := json.Marshal(s.data)
		_, err := db.Exec(`
            UPDATE orchestrator_state 
            SET collected_data = jsonb_set(
                collected_data, 
                $1::text[], 
                $2::jsonb
            )
            WHERE correlation_id = $3
        `, fmt.Sprintf("{%s}", s.step), dataJSON, correlationID)
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

	correlationID := "test-path-" + uuid.New().String()

	// Initialize with empty execution path
	_, err := db.Exec(`
        INSERT INTO orchestrator_state 
        (correlation_id, client_id, status, workflow_plan, execution_path)
        VALUES ($1, $2, $3, $4, $5)
    `, correlationID, "test_client", "RUNNING",
		json.RawMessage(`{"start_step": "init"}`),
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

	for _, step := range executionSteps {
		stepJSON, _ := json.Marshal(step)
		_, err := db.Exec(`
            UPDATE orchestrator_state 
            SET execution_path = execution_path || $1::jsonb
            WHERE correlation_id = $2
        `, stepJSON, correlationID)
		require.NoError(t, err)
	}

	// Verify execution path
	var executionPath json.RawMessage
	err = db.QueryRow(`
        SELECT execution_path FROM orchestrator_state WHERE correlation_id = $1
    `, correlationID).Scan(&executionPath)
	require.NoError(t, err)

	var path []map[string]interface{}
	json.Unmarshal(executionPath, &path)

	assert.Len(t, path, 3)
	assert.Equal(t, "init", path[0]["step"])
	assert.Equal(t, "process", path[1]["step"])
	assert.Equal(t, "validate", path[2]["step"])
}
