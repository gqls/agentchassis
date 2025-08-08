// test/unit/orchestration/state_repository_test.go
package orchestration

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStateRepository_CreateInitialState(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := orchestration.NewStateRepository(db, zap.NewNop())

	plan := models.WorkflowPlan{
		StartStep: "init",
		Steps: map[string]models.Step{
			"init": {
				Action:   "validate",
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	mock.ExpectExec("INSERT INTO orchestrator_state").
		WithArgs(
			"test-123",
			"client-1",
			sqlmock.AnyArg(), // status
			"init",           // start_step
			sqlmock.AnyArg(), // awaited_steps
			sqlmock.AnyArg(), // collected_data
			sqlmock.AnyArg(), // initial_request_data
			sqlmock.AnyArg(), // workflow_plan
			sqlmock.AnyArg(), // execution_metadata
			sqlmock.AnyArg(), // execution_path
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.CreateInitialState(context.Background(), "test-123", "client-1", plan, nil)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestStateRepository_GetState(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := orchestration.NewStateRepository(db, zap.NewNop())

	rows := sqlmock.NewRows([]string{
		"correlation_id", "client_id", "status", "current_step",
		"awaited_steps", "collected_data", "initial_request_data",
		"final_result", "error", "workflow_plan", "execution_metadata",
		"execution_path", "created_at", "updated_at",
	}).AddRow(
		"test-123", "client-1", "RUNNING", "init",
		[]byte("[]"), []byte("{}"), nil,
		nil, nil, []byte("{}"), []byte("{}"),
		[]byte("[]"), time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT correlation_id").
		WithArgs("test-123").
		WillReturnRows(rows)

	state, err := repo.GetState(context.Background(), "test-123")

	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, "test-123", state.CorrelationID)
	assert.Equal(t, "client-1", state.ClientID)
}
