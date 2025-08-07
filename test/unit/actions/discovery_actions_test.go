// test/unit/actions/discovery_actions_test.go
package actions

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/discovery"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDiscoverAgentsAction(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	logger := zap.NewNop()

	// Seed test agents
	seedAgents := []struct {
		agentType    string
		capabilities []string
	}{
		{"researcher", []string{"web_search", "summarization"}},
		{"analyzer", []string{"data_analysis", "reporting"}},
		{"content-creator", []string{"writing", "editing"}},
	}

	for i, agent := range seedAgents {
		_, err := db.Exec(`
            INSERT INTO client_demo_client.agent_instances 
            (id, template_id, owner_user_id, name, config, is_active)
            VALUES ($1, $2, $3, $4, $5, true)
        `,
			fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1),
			uuid.New().String(),
			"test_user",
			fmt.Sprintf("test-%s", agent.agentType),
			map[string]interface{}{
				"agent_type":   agent.agentType,
				"capabilities": agent.capabilities,
				"topic":        fmt.Sprintf("system.agent.%s.process", agent.agentType),
			},
		)
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		params    actions.ActionParams
		wantCount int
		wantTypes []string
		wantErr   bool
	}{
		{
			name: "Discover by type",
			params: actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": "researcher",
					},
				},
				Headers: helpers.TestHeaders("test-discover-001"),
				DB:      db,
				Logger:  logger,
			},
			wantCount: 1,
			wantTypes: []string{"researcher"},
			wantErr:   false,
		},
		{
			name: "Discover by capabilities",
			params: actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"required_capabilities": []string{"web_search"},
					},
				},
				Headers: helpers.TestHeaders("test-discover-002"),
				DB:      db,
				Logger:  logger,
			},
			wantCount: 1,
			wantTypes: []string{"researcher"},
			wantErr:   false,
		},
		{
			name: "Discover multiple agents",
			params: actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"min_agents": 2,
					},
				},
				Headers: helpers.TestHeaders("test-discover-003"),
				DB:      db,
				Logger:  logger,
			},
			wantCount: 3, // Should find all agents
			wantErr:   false,
		},
		{
			name: "No matching agents",
			params: actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": "nonexistent",
					},
				},
				Headers: helpers.TestHeaders("test-discover-004"),
				DB:      db,
				Logger:  logger,
			},
			wantCount: 0,
			wantErr:   false, // Not an error, just empty result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := actions.DiscoverAgentsAction(context.Background(), tt.params)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			discovered := result.(map[string]interface{})
			agents := discovered["agents"].([]discovery.AgentMatch)

			assert.Len(t, agents, tt.wantCount)

			if len(tt.wantTypes) > 0 {
				foundTypes := make(map[string]bool)
				for _, agent := range agents {
					foundTypes[agent.AgentType] = true
				}

				for _, wantType := range tt.wantTypes {
					assert.True(t, foundTypes[wantType], "Expected to find agent type %s", wantType)
				}
			}
		})
	}
}

func TestSelectBestAgentAction(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	// Create agents with different scores
	agents := []struct {
		id    string
		score float64
	}{
		{"agent-1", 0.9},
		{"agent-2", 0.7},
		{"agent-3", 0.95},
	}

	for _, agent := range agents {
		_, err := db.Exec(`
            INSERT INTO agent_metrics 
            (agent_id, success_rate, avg_response_time, total_requests)
            VALUES ($1, $2, $3, $4)
        `, agent.id, agent.score, 1000, 100)
		require.NoError(t, err)
	}

	params := actions.ActionParams{
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_candidates":   []string{"agent-1", "agent-2", "agent-3"},
				"selection_criteria": "highest_score",
			},
		},
		Headers: helpers.TestHeaders("test-select-001"),
		DB:      db,
		Logger:  zap.NewNop(),
	}

	result, err := actions.SelectBestAgentAction(context.Background(), params)
	require.NoError(t, err)

	selection := result.(map[string]interface{})
	assert.Equal(t, "agent-3", selection["selected_agent_id"])
	assert.Equal(t, 0.95, selection["score"])
}

func TestAgentHealthCheck(t *testing.T) {
	db := helpers.TestDB(t)
	defer db.Close()

	// Create agent with heartbeat
	agentID := "health-check-agent"
	_, err := db.Exec(`
        INSERT INTO client_demo_client.agent_instances 
        (id, template_id, owner_user_id, name, config, is_active, last_heartbeat)
        VALUES ($1, $2, $3, $4, $5, true, $6)
    `, agentID, uuid.New().String(), "test_user", "health-test",
		map[string]interface{}{"agent_type": "generic"},
		time.Now().Add(-30*time.Second)) // Recent heartbeat
	require.NoError(t, err)

	params := actions.ActionParams{
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_id":             agentID,
				"health_check_timeout": "1m",
			},
		},
		Headers: helpers.TestHeaders("test-health-001"),
		DB:      db,
		Logger:  zap.NewNop(),
	}

	result, err := actions.CheckAgentHealthAction(context.Background(), params)
	require.NoError(t, err)

	health := result.(map[string]interface{})
	assert.True(t, health["is_healthy"].(bool))
	assert.Less(t, health["last_heartbeat_seconds"].(float64), 60.0)
}
