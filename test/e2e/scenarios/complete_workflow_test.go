// test/e2e/scenarios/complete_workflow_test.go
package scenarios

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCompleteWorkflowWithSpawning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup
	ctx := context.Background()
	db := setupTestDB(t)
	producer := setupTestProducer(t)
	logger := zap.NewNop()

	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	tests := []struct {
		name     string
		workflow models.WorkflowPlan
		wantErr  bool
	}{
		{
			name:     "Simple workflow",
			workflow: createSimpleWorkflow(),
			wantErr:  false,
		},
		{
			name:     "Multi-agent workflow",
			workflow: createMultiAgentWorkflow(),
			wantErr:  false,
		},
		{
			name:     "Workflow with spawning",
			workflow: createSpawningWorkflow(),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correlationID := helpers.TestUUIDWithType("e2e")
			headers := helpers.TestHeaders(correlationID)
			headers["client_id"] = "test_client"
			headers["fuel_budget"] = "1000"

			err := coordinator.ExecuteWorkflow(ctx, tt.workflow, headers, nil)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			// For local actions, workflow might complete immediately
			// For remote actions, we need to wait
			if hasRemoteActions(tt.workflow) {
				// Simulate response handling for remote actions
				time.Sleep(100 * time.Millisecond)

				// In a real test, you'd consume from Kafka and send responses
				// For now, we'll just check the state
			}

			// Verify state was created
			state := getWorkflowState(t, db, correlationID)
			assert.NotNil(t, state)

			// Check status is either running, awaiting, or completed
			validStatuses := []string{
				string(orchestration.StatusRunning),
				string(orchestration.StatusAwaitingResponses),
				string(orchestration.StatusCompleted),
			}
			assert.Contains(t, validStatuses, string(state.Status))
		})
	}
}

func createSimpleWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "validate",
		Steps: map[string]models.Step{
			"validate": {
				Action:      "validate_input",
				Description: "Validate input data",
				NextStep:    "transform",
			},
			"transform": {
				Action:      "transform_data",
				Description: "Transform data",
				Config: map[string]interface{}{
					"transformation": "uppercase",
				},
				NextStep: "complete",
			},
			"complete": {
				Action:      "complete_workflow",
				Description: "Complete the workflow",
			},
		},
	}
}

func createMultiAgentWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "spawn_agents",
		Steps: map[string]models.Step{
			"spawn_agents": {
				Action: "spawn_agent",
				Config: map[string]interface{}{
					"agent_type": "generic",
				},
				NextStep: "call_agents",
			},
			"call_agents": {
				Action: "fan_out",
				SubTasks: []models.SubTask{
					{
						StepName: "task1",
						Topic:    "system.agent.generic.process",
					},
					{
						StepName: "task2",
						Topic:    "system.agent.generic.process",
					},
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}
}

func createSpawningWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "spawn_group",
		Steps: map[string]models.Step{
			"spawn_group": {
				Action: "spawn_group",
				Config: map[string]interface{}{
					"group_type": "test-group",
				},
				NextStep: "process",
			},
			"process": {
				Action: "fan_out",
				SubTasks: []models.SubTask{
					{StepName: "task1", Topic: "system.agent.worker.process"},
					{StepName: "task2", Topic: "system.agent.analyzer.process"},
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}
}

func hasRemoteActions(workflow models.WorkflowPlan) bool {
	for _, step := range workflow.Steps {
		if step.Topic != "" || step.Action == "fan_out" || step.Action == "call_agent" {
			return true
		}
	}
	return false
}

func getWorkflowState(t *testing.T, db *sql.DB, correlationID string) *orchestration.OrchestrationState {
	repo := orchestration.NewStateRepository(db, zap.NewNop())
	state, err := repo.GetState(context.Background(), correlationID)
	if err != nil {
		t.Logf("Failed to get state: %v", err)
		return nil
	}
	return state
}

func setupTestDB(t *testing.T) *sql.DB {
	db := helpers.TestDB(t)
	helpers.SetupTestSchema(t, db)

	// Ensure agent_groups table exists with test data
	_, err := db.Exec(`
		INSERT INTO agent_groups (id, name, group_type, version, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, uuid.New().String(), "Test Group", "test-group", "1.0.0",
		`[{"role": "worker", "agent_type": "generic"}]`,
		`{"start_step": "process", "steps": {"process": {"action": "process_task"}}}`)

	if err != nil {
		t.Logf("Warning: Could not insert test group: %v", err)
	}

	return db
}

func setupTestProducer(t *testing.T) kafka.Producer {
	// For E2E tests, you might want to use a real Kafka producer
	// For now, we'll use the mock
	return helpers.NewMockProducer()
}
