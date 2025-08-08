// test/unit/actions/spawn_actions_test.go
package actions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
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

	ctx := context.Background()
	logger := zap.NewNop()

	// Clean up any existing test agents first
	_, _ = db.Exec(`DELETE FROM client_test_client.agent_instances WHERE name LIKE 'test-%'`)

	// Setup schema
	_, err := db.Exec(`CREATE SCHEMA IF NOT EXISTS client_test_client`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS client_test_client.agent_instances (
			id UUID PRIMARY KEY,
			template_id UUID NOT NULL,
			owner_user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			config JSONB NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	tests := []struct {
		name    string
		params  actions.ActionParams
		wantErr bool
		check   func(t *testing.T, result interface{})
	}{
		{
			name: "Create new agent",
			params: actions.ActionParams{
				Context: ctx,
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": "researcher",
					},
				},
				Headers: map[string]string{
					"correlation_id": "test-spawn-001",
					"client_id":      "test_client",
					"user_id":        "test_user",
				},
				DB:     db,
				Logger: logger,
			},
			wantErr: false,
			check: func(t *testing.T, result interface{}) {
				res, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.NotEmpty(t, res["agent_id"])
				assert.Equal(t, "system.agent.researcher.process", res["topic"])
				assert.Equal(t, "created", res["status"])
			},
		},
		{
			name: "Missing agent_type",
			params: actions.ActionParams{
				Context: ctx,
				StepConfig: models.Step{
					Config: map[string]interface{}{},
				},
				Headers: helpers.TestHeaders("test-spawn-002"),
				DB:      db,
				Logger:  logger,
			},
			wantErr: true,
		},
		{
			name: "Missing client_id",
			params: actions.ActionParams{
				Context: ctx,
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": "researcher",
					},
				},
				Headers: map[string]string{
					"correlation_id": "test-spawn-003",
					// client_id missing
				},
				DB:     db,
				Logger: logger,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actions.SpawnAgentAction(ctx, tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestSpawnGroupAction(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	ctx := context.Background()
	logger := zap.NewNop()

	// Setup test data
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_groups (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			group_type VARCHAR(100) NOT NULL,
			agent_configs JSONB NOT NULL,
			orchestration_workflow JSONB NOT NULL,
			usage_count INTEGER DEFAULT 0,
			last_used_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// Insert test group
	agentConfigs := []map[string]interface{}{
		{"role": "worker", "agent_type": "generic"},
	}
	agentConfigsJSON, _ := json.Marshal(agentConfigs)

	workflow := map[string]interface{}{
		"start_step": "process",
		"steps": map[string]interface{}{
			"process": map[string]interface{}{
				"action": "process_task",
			},
		},
	}
	workflowJSON, _ := json.Marshal(workflow)

	_, err = db.Exec(`
		INSERT INTO agent_groups (id, name, group_type, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New().String(), "Test Group", "test-group", agentConfigsJSON, workflowJSON)
	require.NoError(t, err)

	// Setup client schema for spawned agents
	_, err = db.Exec(`CREATE SCHEMA IF NOT EXISTS client_test_client`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS client_test_client.agent_instances (
			id UUID PRIMARY KEY,
			template_id UUID NOT NULL,
			owner_user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			config JSONB NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	params := actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"group_type": "test-group",
			},
		},
		Headers: map[string]string{
			"correlation_id": "test-group-001",
			"client_id":      "test_client",
			"user_id":        "test_user",
		},
		DB:     db,
		Logger: logger,
	}

	result, err := actions.SpawnGroupAction(ctx, params)
	require.NoError(t, err)

	groupResult, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, groupResult["group_id"])
	assert.Equal(t, "Test Group", groupResult["group_name"])

	agents, ok := groupResult["agents"].(map[string]string)
	require.True(t, ok)
	assert.NotEmpty(t, agents["worker"])
}
