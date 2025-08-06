// test/unit/actions/spawn_actions_test.go
package actions

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSpawnAgentAction(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	logger := zap.NewNop()
	producer := helpers.NewMockProducer()

	tests := []struct {
		name    string
		params  actions.ActionParams
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Create new agent",
			params: actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": "researcher",
					},
				},
				Headers:  helpers.TestHeaders("test-001"),
				DB:       db,
				Logger:   logger,
				Producer: producer,
			},
			want: map[string]string{
				"status": "created",
				"topic":  "system.agent.researcher.process",
			},
			wantErr: false,
		},
		{
			name: "Missing agent_type",
			params: actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{},
				},
				Headers: helpers.TestHeaders("test-002"),
				DB:      db,
				Logger:  logger,
			},
			wantErr: true,
		},
		{
			name: "Reuse existing agent",
			params: actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": "generic", // This exists in test data
					},
				},
				Headers: map[string]string{
					"correlation_id": "test-003",
					"client_id":      "demo_client",
					"user_id":        "test_user",
				},
				DB:       db,
				Logger:   logger,
				Producer: producer,
			},
			want: map[string]string{
				"status": "reused",
				"topic":  "system.agent.generic.process",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actions.SpawnAgentAction(context.Background(), tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			resultMap := result.(map[string]interface{})

			assert.Equal(t, tt.want["status"], resultMap["status"])
			assert.Equal(t, tt.want["topic"], resultMap["topic"])
			assert.NotEmpty(t, resultMap["agent_id"])
		})
	}
}

func TestSpawnGroupAction(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	logger := zap.NewNop()
	producer := helpers.NewMockProducer()

	params := actions.ActionParams{
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"group_type": "test-group",
			},
		},
		Headers:  helpers.TestHeaders("test-group-001"),
		DB:       db,
		Logger:   logger,
		Producer: producer,
	}

	result, err := actions.SpawnGroupAction(context.Background(), params)
	require.NoError(t, err)

	groupResult := result.(map[string]interface{})
	assert.Equal(t, "33333333-3333-3333-3333-333333333333", groupResult["group_id"])
	assert.Equal(t, "Test Team", groupResult["group_name"])

	agents := groupResult["agents"].(map[string]string)
	assert.NotEmpty(t, agents["worker"])
}
