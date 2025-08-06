// test/unit/orchestration/coordinator_test.go
package orchestration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWorkflowExecution(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	producer := helpers.NewMockProducer()
	logger := zap.NewNop()

	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	tests := []struct {
		name     string
		workflow models.WorkflowPlan
		wantErr  bool
	}{
		{
			name:     "Valid simple workflow",
			workflow: helpers.ValidWorkflow(),
			wantErr:  false,
		},
		{
			name: "Workflow with invalid start step",
			workflow: models.WorkflowPlan{
				StartStep: "nonexistent",
				Steps:     map[string]models.Step{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			correlationID := fmt.Sprintf("test-%d", time.Now().Unix())
			headers := helpers.TestHeaders(correlationID)

			err := coordinator.ExecuteWorkflow(context.Background(), tt.workflow, headers, nil)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify state was created
			var status string
			err = db.QueryRow("SELECT status FROM orchestrator_state WHERE correlation_id = $1", correlationID).Scan(&status)
			require.NoError(t, err)
			assert.NotEmpty(t, status)
		})
	}
}
