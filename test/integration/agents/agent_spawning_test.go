// test/integration/agents/agent_spawning_test.go
package agents

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentSpawningIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	producer := createRealKafkaProducer(t)
	defer producer.Close()

	tests := []struct {
		name      string
		agentType string
		verify    func(t *testing.T, agentID string)
	}{
		{
			name:      "Spawn researcher agent",
			agentType: "researcher",
			verify: func(t *testing.T, agentID string) {
				// Verify agent exists in database
				var exists bool
				err := db.QueryRow(`
                    SELECT EXISTS(
                        SELECT 1 FROM client_demo_client.agent_instances 
                        WHERE id = $1 AND is_active = true
                    )`, agentID).Scan(&exists)
				require.NoError(t, err)
				assert.True(t, exists)
			},
		},
		{
			name:      "Spawn content creator",
			agentType: "content-creator",
			verify: func(t *testing.T, agentID string) {
				// Verify agent configuration
				var config map[string]interface{}
				err := db.QueryRow(`
                    SELECT config FROM client_demo_client.agent_instances 
                    WHERE id = $1
                `, agentID).Scan(&config)
				require.NoError(t, err)
				assert.Equal(t, "content-creator", config["agent_type"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": tt.agentType,
					},
				},
				Headers: map[string]string{
					"correlation_id": "test-spawn-" + uuid.New().String(),
					"client_id":      "demo_client",
					"user_id":        "test_user",
				},
				DB:       db,
				Producer: producer,
				Logger:   zap.NewNop(),
			}

			result, err := actions.SpawnAgentAction(context.Background(), params)
			require.NoError(t, err)

			agentInfo := result.(map[string]interface{})
			agentID := agentInfo["agent_id"].(string)

			// Verify agent was spawned correctly
			tt.verify(t, agentID)

			// Test agent responds to messages
			testAgentResponse(t, agentID, tt.agentType)
		})
	}
}

func TestAgentLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	// Test full agent lifecycle
	agentID := uuid.New().String()

	// 1. Create agent
	_, err := db.Exec(`
        INSERT INTO client_demo_client.agent_instances 
        (id, template_id, owner_user_id, name, config, is_active)
        VALUES ($1, $2, $3, $4, $5, true)
    `, agentID, uuid.New(), "test_user", "lifecycle-test-agent",
		map[string]interface{}{
			"agent_type": "generic",
			"topic":      "system.agent.generic.process",
		})
	require.NoError(t, err)

	// 2. Update heartbeat
	for i := 0; i < 3; i++ {
		_, err = db.Exec(`
            UPDATE client_demo_client.agent_instances 
            SET last_heartbeat = NOW() 
            WHERE id = $1
        `, agentID)
		require.NoError(t, err)
		time.Sleep(1 * time.Second)
	}

	// 3. Deactivate agent
	_, err = db.Exec(`
        UPDATE client_demo_client.agent_instances 
        SET is_active = false 
        WHERE id = $1
    `, agentID)
	require.NoError(t, err)

	// 4. Verify agent is deactivated
	var isActive bool
	err = db.QueryRow(`
        SELECT is_active FROM client_demo_client.agent_instances 
        WHERE id = $1
    `, agentID).Scan(&isActive)
	require.NoError(t, err)
	assert.False(t, isActive)
}

func TestConcurrentAgentSpawning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	// Test concurrent spawning
	concurrency := 10
	results := make(chan string, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			params := actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": "worker",
					},
				},
				Headers: map[string]string{
					"correlation_id": fmt.Sprintf("test-concurrent-%d", index),
					"client_id":      "demo_client",
					"user_id":        "test_user",
				},
				DB:     db,
				Logger: zap.NewNop(),
			}

			result, err := actions.SpawnAgentAction(context.Background(), params)
			if err != nil {
				errors <- err
				return
			}

			agentInfo := result.(map[string]interface{})
			results <- agentInfo["agent_id"].(string)
		}(i)
	}

	// Collect results
	spawnedAgents := make([]string, 0, concurrency)
	for i := 0; i < concurrency; i++ {
		select {
		case agentID := <-results:
			spawnedAgents = append(spawnedAgents, agentID)
		case err := <-errors:
			t.Errorf("Spawning error: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent spawning")
		}
	}

	// Verify all agents were created
	assert.Len(t, spawnedAgents, concurrency)

	// Verify no duplicates
	agentSet := make(map[string]bool)
	for _, id := range spawnedAgents {
		if agentSet[id] {
			t.Errorf("Duplicate agent ID: %s", id)
		}
		agentSet[id] = true
	}
}
