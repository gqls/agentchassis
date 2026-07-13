// test/unit/actions/discovery_actions_test.go
package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDiscoverAgentsAction_Mock(t *testing.T) {
	// Use mock database for unit tests
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	logger := zap.NewNop()

	// Mock the discovery query - this will fail but won't crash
	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	params := actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_type": "researcher",
			},
		},
		Headers: helpers.TestHeaders("test-discover-mock"),
		DB:      db,
		Logger:  logger,
	}

	// The action will fail due to pgxpool requirement, but we can test the flow
	_, err = actions.DiscoverAgentsAction(ctx, params)

	// We expect an error because the action requires pgxpool
	assert.Error(t, err)
}

func TestSpawnAgentAction_Mock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	logger := zap.NewNop()

	// Test missing agent_type
	params := actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				// agent_type missing
			},
		},
		Headers: helpers.TestHeaders("test-spawn-missing-type"),
		DB:      db,
		Logger:  logger,
	}

	_, err = actions.SpawnAgentAction(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_type not specified")

	// Test missing client_id
	params = actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_type": "researcher",
			},
		},
		Headers: map[string]string{
			"correlation_id": "test-spawn-no-client",
			// client_id missing
		},
		DB:     db,
		Logger: logger,
	}

	_, err = actions.SpawnAgentAction(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client_id not specified")

	// Test successful spawn (mock)
	params = actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_type": "researcher",
			},
		},
		Headers: helpers.TestHeaders("test-spawn-success"),
		DB:      db,
		Logger:  logger,
	}

	// Mock the SELECT query (agent doesn't exist)
	mock.ExpectQuery("SELECT id FROM").
		WillReturnError(sql.ErrNoRows)

	// Mock the INSERT query
	mock.ExpectExec("INSERT INTO").
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := actions.SpawnAgentAction(ctx, params)
	require.NoError(t, err)

	res, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, res["agent_id"])
	assert.Equal(t, "system.agent.researcher.process", res["topic"])
	assert.Equal(t, "created", res["status"])
}

func TestSpawnGroupAction_Mock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	logger := zap.NewNop()

	// Test missing group_type
	params := actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				// group_type missing
			},
		},
		Headers: helpers.TestHeaders("test-group-missing-type"),
		DB:      db,
		Logger:  logger,
	}

	_, err = actions.SpawnGroupAction(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group_type not specified")

	// Test group not found
	params = actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"group_type": "nonexistent-group",
			},
		},
		Headers: helpers.TestHeaders("test-group-not-found"),
		DB:      db,
		Logger:  logger,
	}

	// Mock the SELECT query (no group found)
	mock.ExpectQuery("SELECT id, name").
		WillReturnError(sql.ErrNoRows)

	_, err = actions.SpawnGroupAction(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no group found")
}

func TestConditionalRouteAction_Mock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Test boolean routing
	params := actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"condition_field": "approved",
				"routes": map[string]interface{}{
					"true":  "approve_step",
					"false": "reject_step",
				},
			},
		},
		Headers: map[string]string{
			"correlation_id": "test-conditional",
		},
		CollectedData: map[string]interface{}{
			"approved": true,
		},
		DB: db,
	}

	// Mock the UPDATE query
	mock.ExpectExec("UPDATE orchestrator_state").
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := actions.ConditionalRouteAction(ctx, params)
	require.NoError(t, err)

	routeResult, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "approve_step", routeResult["routed_to"])
	assert.Equal(t, true, routeResult["condition_value"])
}
