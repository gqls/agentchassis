// test/unit/orchestration/state_test.go
package orchestration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestratorState(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	store := orchestration.NewStateStore(db)

	tests := []struct {
		name  string
		state models.OrchestratorState
	}{
		{
			name: "Simple state",
			state: models.OrchestratorState{
				CorrelationID: "test-state-001",
				ClientID:      "test_client",
				Status:        "RUNNING",
				CurrentStep:   "process",
				WorkflowPlan:  helpers.ValidWorkflow(),
				CollectedData: map[string]interface{}{
					"input": "test data",
				},
				ExecutionMetadata: map[string]interface{}{
					"started_at": time.Now(),
				},
			},
		},
		{
			name: "Complex state with execution path",
			state: models.OrchestratorState{
				CorrelationID: "test-state-002",
				ClientID:      "test_client",
				Status:        "AWAITING_RESPONSES",
				CurrentStep:   "fan_out",
				WorkflowPlan:  createComplexWorkflow(),
				ExecutionPath: []models.ExecutionStep{
					{
						Step:      "init",
						Action:    "validate_input",
						StartTime: time.Now().Add(-5 * time.Minute),
						EndTime:   time.Now().Add(-4 * time.Minute),
						Result:    "success",
					},
					{
						Step:      "spawn",
						Action:    "spawn_agents",
						StartTime: time.Now().Add(-4 * time.Minute),
						EndTime:   time.Now().Add(-3 * time.Minute),
						Result:    "success",
						Output: map[string]interface{}{
							"spawned_agents": []string{"agent-1", "agent-2"},
						},
					},
				},
				AwaitedSteps: []string{"response-1", "response-2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create state
			err := store.CreateState(tt.state)
			require.NoError(t, err)

			// Read state
			retrieved, err := store.GetState(tt.state.CorrelationID)
			require.NoError(t, err)

			assert.Equal(t, tt.state.CorrelationID, retrieved.CorrelationID)
			assert.Equal(t, tt.state.Status, retrieved.Status)
			assert.Equal(t, tt.state.CurrentStep, retrieved.CurrentStep)

			// Update state
			retrieved.Status = "COMPLETED"
			retrieved.CurrentStep = "complete"
			err = store.UpdateState(retrieved)
			require.NoError(t, err)

			// Verify update
			updated, err := store.GetState(tt.state.CorrelationID)
			require.NoError(t, err)
			assert.Equal(t, "COMPLETED", updated.Status)
			assert.Equal(t, "complete", updated.CurrentStep)
		})
	}
}

func TestStateTransitions(t *testing.T) {
	validTransitions := []struct {
		from  string
		to    string
		valid bool
	}{
		{"INITIALIZING", "RUNNING", true},
		{"RUNNING", "AWAITING_RESPONSES", true},
		{"RUNNING", "COMPLETED", true},
		{"RUNNING", "FAILED", true},
		{"AWAITING_RESPONSES", "RUNNING", true},
		{"AWAITING_RESPONSES", "COMPLETED", true},
		{"AWAITING_RESPONSES", "FAILED", true},
		{"COMPLETED", "RUNNING", false}, // Invalid
		{"FAILED", "RUNNING", false},    // Invalid
	}

	for _, tt := range validTransitions {
		t.Run(fmt.Sprintf("%s_to_%s", tt.from, tt.to), func(t *testing.T) {
			err := orchestration.ValidateStateTransition(tt.from, tt.to)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestExecutionMetadata(t *testing.T) {
	metadata := orchestration.NewExecutionMetadata()

	// Test step tracking
	metadata.StartStep("init")
	time.Sleep(100 * time.Millisecond)
	metadata.CompleteStep("init", "success")

	metadata.StartStep("process")
	time.Sleep(100 * time.Millisecond)
	metadata.CompleteStep("process", "success")

	assert.Equal(t, 2, metadata.CompletedSteps)
	assert.Equal(t, 0, metadata.FailedSteps)
	assert.Greater(t, metadata.EndTime.Sub(metadata.StartTime), 200*time.Millisecond)

	// Test checkpoint
	metadata.AddCheckpoint("halfway", map[string]interface{}{
		"processed": 50,
		"remaining": 50,
	})

	checkpoint, exists := metadata.GetCheckpoint("halfway")
	assert.True(t, exists)
	assert.Equal(t, 50, checkpoint["processed"])

	// Test retry tracking
	metadata.IncrementRetry("process")
	metadata.IncrementRetry("process")

	assert.Equal(t, 2, metadata.GetRetryCount("process"))
}

func TestCollectedDataManagement(t *testing.T) {
	collector := orchestration.NewDataCollector()

	// Add data from different steps
	collector.AddStepData("research", map[string]interface{}{
		"sources": []string{"source1", "source2"},
		"summary": "Research findings",
	})

	collector.AddStepData("analysis", map[string]interface{}{
		"insights":   []string{"insight1", "insight2"},
		"confidence": 0.85,
	})

	// Retrieve specific step data
	researchData := collector.GetStepData("research")
	assert.NotNil(t, researchData)
	assert.Equal(t, "Research findings", researchData["summary"])

	// Get all collected data
	allData := collector.GetAllData()
	assert.Len(t, allData, 2)
	assert.Contains(t, allData, "research")
	assert.Contains(t, allData, "analysis")

	// Test data merging
	collector.MergeData("analysis", map[string]interface{}{
		"additional_insights": []string{"insight3"},
		"confidence":          0.90, // Should update
	})

	analysisData := collector.GetStepData("analysis")
	assert.Equal(t, 0.90, analysisData["confidence"])
	assert.Contains(t, analysisData, "additional_insights")
}

func createComplexWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "init",
		Steps: map[string]models.Step{
			"init": {
				Action:   "validate_input",
				NextStep: "spawn",
			},
			"spawn": {
				Action: "spawn_agents",
				Config: map[string]interface{}{
					"count": 3,
				},
				NextStep: "fan_out",
			},
			"fan_out": {
				Action: "fan_out",
				SubTasks: []models.SubTask{
					{StepName: "task1", Topic: "system.agent.worker.process"},
					{StepName: "task2", Topic: "system.agent.analyzer.process"},
					{StepName: "task3", Topic: "system.agent.reviewer.process"},
				},
				NextStep: "aggregate",
			},
			"aggregate": {
				Action:   "aggregate_results",
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}
}
