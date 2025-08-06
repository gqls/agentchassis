// test/e2e/scenarios/complete_workflow_test.go
package scenarios

import (
	"context"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/stretchr/testify/require"
)

func TestCompleteWorkflowWithSpawning(t *testing.T) {
	// Setup
	ctx := context.Background()
	coordinator := setupTestCoordinator(t)

	tests := []struct {
		name     string
		workflow models.WorkflowPlan
		wantErr  bool
	}{
		{
			name:     "Simple workflow",
			workflow: loadWorkflow(t, "simple.yaml"),
			wantErr:  false,
		},
		{
			name:     "Multi-agent workflow",
			workflow: loadWorkflow(t, "multi_agent.yaml"),
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
			correlationID := generateTestCorrelationID()
			headers := createTestHeaders(correlationID)

			err := coordinator.ExecuteWorkflow(ctx, tt.workflow, headers, nil)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Wait and verify completion
			time.Sleep(5 * time.Second)
			state := getWorkflowState(t, correlationID)
			require.Equal(t, "COMPLETED", state.Status)
		})
	}
}

func createSpawningWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "spawn_agents",
		Steps: map[string]models.Step{
			"spawn_agents": {
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
