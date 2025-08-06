// test/integration/agents/multi_agent_test.go
package agents

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/require"
)

func TestMultiAgentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	// This test requires actual Kafka
	producer := createRealKafkaProducer(t)
	defer producer.Close()

	coordinator := orchestration.NewSagaCoordinator(db, producer, zap.NewNop())

	// Create a workflow that uses multiple agents
	workflow := models.WorkflowPlan{
		StartStep: "spawn_team",
		Steps: map[string]models.Step{
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
					{StepName: "research", Topic: "system.agent.researcher.process"},
					{StepName: "write", Topic: "system.agent.content-creator.process"},
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := fmt.Sprintf("test-multi-%d", time.Now().Unix())
	headers := helpers.TestHeaders(correlationID)

	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
	require.NoError(t, err)

	// Wait for workflow to progress
	time.Sleep(10 * time.Second)

	// Verify workflow state
	var status string
	var collectedData []byte
	err = db.QueryRow(`
        SELECT status, collected_data 
        FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&status, &collectedData)

	require.NoError(t, err)
	assert.Contains(t, []string{"COMPLETED", "AWAITING_RESPONSES"}, status)
}
