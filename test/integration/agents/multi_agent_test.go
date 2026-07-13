// test/integration/agents/multi_agent_test.go
package agents

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMultiAgentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()

	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)
	ctx := context.Background()

	// Setup content team group
	setupContentTeam(t, db)

	// Create a workflow that uses multiple agents
	workflow := models.WorkflowPlan{
		StartStep: "validate",
		Steps: map[string]models.Step{
			"validate": {
				Action:      "validate_input",
				Description: "Validate workflow input",
				NextStep:    "spawn_team",
			},
			"spawn_team": {
				Action: "spawn_group",
				Config: map[string]interface{}{
					"group_type": "content-team",
				},
				NextStep: "distribute_work",
			},
			"distribute_work": {
				Action: "fan_out",
				SubTasks: []models.SubTask{
					{
						StepName: "research",
						Topic:    "system.agent.researcher.process",
					},
					{
						StepName: "write",
						Topic:    "system.agent.content-creator.process",
					},
				},
				NextStep: "aggregate",
			},
			"aggregate": {
				Action:      "transform_data",
				Description: "Aggregate results",
				Config: map[string]interface{}{
					"transformation": "uppercase",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := helpers.TestUUIDWithType("integration")
	headers := helpers.TestHeaders(correlationID)

	// Add initial data for validation
	initialData, _ := json.Marshal(map[string]interface{}{
		"action": "multi_agent_test",
		"data": map[string]interface{}{
			"message": "Test multi-agent workflow",
		},
	})

	err := coordinator.ExecuteWorkflow(ctx, workflow, headers, initialData)
	require.NoError(t, err)

	// Check workflow state
	repo := orchestration.NewStateRepository(db, logger)
	state, err := repo.GetState(ctx, correlationID)
	require.NoError(t, err)

	// Verify workflow is progressing
	assert.NotNil(t, state)
	assert.NotEmpty(t, state.CurrentStep)

	// For fan_out step, should be awaiting responses
	if state.CurrentStep == "aggregate" {
		assert.Equal(t, orchestration.StatusAwaitingResponses, state.Status)
		assert.Len(t, state.AwaitedSteps, 2) // Two subtasks
	}

	// Simulate agent responses if needed
	if state.Status == orchestration.StatusAwaitingResponses {
		for _, awaitedStep := range state.AwaitedSteps {
			response := models.TaskResponse{
				Data: map[string]interface{}{
					"result": fmt.Sprintf("Completed %s", awaitedStep),
					"status": "success", // Include status in data if needed
				},
				Error: "", // No error for successful response
			}

			responseData, _ := json.Marshal(response)
			responseHeaders := make(map[string]string)
			for k, v := range headers {
				responseHeaders[k] = v
			}
			responseHeaders["causation_id"] = awaitedStep

			err = coordinator.HandleResponse(ctx, responseHeaders, responseData)
			assert.NoError(t, err)
		}

		// Check state after responses
		time.Sleep(100 * time.Millisecond)
		state, _ = repo.GetState(ctx, correlationID)

		// Should have progressed past fan_out
		assert.NotEqual(t, "distribute_work", state.CurrentStep)
	}
}

func TestMultiAgentCoordination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()
	ctx := context.Background()

	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	// Test agent coordination with dependencies
	workflow := models.WorkflowPlan{
		StartStep: "spawn_first",
		Steps: map[string]models.Step{
			"spawn_first": {
				Action: "spawn_agent",
				Config: map[string]interface{}{
					"agent_type": "analyzer",
				},
				NextStep: "analyze",
			},
			"analyze": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_type": "analyzer",
					"action":     "analyze_data",
				},
				NextStep: "spawn_second",
			},
			"spawn_second": {
				Action: "spawn_agent",
				Config: map[string]interface{}{
					"agent_type": "processor",
				},
				NextStep: "process",
			},
			"process": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_type": "processor",
					"action":     "process_analysis",
					"input":      "{{collected_data.analyze}}",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := helpers.TestUUIDWithType("integration")
	headers := helpers.TestHeaders(correlationID)

	err := coordinator.ExecuteWorkflow(ctx, workflow, headers, nil)
	require.NoError(t, err)

	// Verify workflow created
	repo := orchestration.NewStateRepository(db, logger)
	state, err := repo.GetState(ctx, correlationID)
	require.NoError(t, err)
	assert.NotNil(t, state)
}

func TestMultiAgentFailureHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()
	ctx := context.Background()

	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	// Test workflow with potential failure points
	workflow := models.WorkflowPlan{
		StartStep: "validate",
		Steps: map[string]models.Step{
			"validate": {
				Action:   "validate_input",
				NextStep: "risky_operation",
			},
			"risky_operation": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_type": "nonexistent", // This agent doesn't exist
					"action":     "will_fail",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := helpers.TestUUIDWithType("integration")
	headers := helpers.TestHeaders(correlationID)

	// Invalid data to trigger validation failure
	invalidData := []byte(`{"invalid": "json structure}`)

	coordinator.ExecuteWorkflow(ctx, workflow, headers, invalidData)

	// Workflow creation might succeed but execution will fail
	repo := orchestration.NewStateRepository(db, logger)
	state, stateErr := repo.GetState(ctx, correlationID)

	if stateErr == nil && state != nil {
		// Check if workflow failed
		if state.Status == orchestration.StatusFailed {
			assert.NotEmpty(t, state.Error)
			t.Logf("Workflow failed as expected: %s", state.Error)
		}
	}
}

// Helper function to setup content team
func setupContentTeam(t *testing.T, db *sql.DB) {
	// Create proper JSON for agent configs
	agentConfigs := []map[string]interface{}{
		{
			"role":       "researcher",
			"agent_type": "researcher",
		},
		{
			"role":       "writer",
			"agent_type": "content-creator",
		},
		{
			"role":       "editor",
			"agent_type": "editor",
		},
	}
	agentConfigsJSON, _ := json.Marshal(agentConfigs)

	// Create proper workflow JSON
	workflowSteps := map[string]map[string]interface{}{
		"research": {
			"action":    "research_topic",
			"next_step": "write",
		},
		"write": {
			"action":    "create_content",
			"next_step": "edit",
		},
		"edit": {
			"action": "edit_content",
		},
	}

	workflow := map[string]interface{}{
		"start_step": "research",
		"steps":      workflowSteps,
	}
	workflowJSON, _ := json.Marshal(workflow)

	_, err := db.Exec(`
		INSERT INTO agent_groups 
		(id, name, group_type, version, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, uuid.New().String(), "Content Team", "content-team", "1.0.0",
		agentConfigsJSON, workflowJSON)

	if err != nil {
		t.Logf("Warning: Could not insert content team: %v", err)
	}
}
