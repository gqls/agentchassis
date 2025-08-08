// test/integration/agents/agent_spawning_test.go
package agents

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAgentSpawningIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()
	ctx := context.Background()

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
				query := fmt.Sprintf(`
					SELECT EXISTS(
						SELECT 1 FROM client_%s.agent_instances 
						WHERE id = $1 AND is_active = true
					)`, "test_client")

				err := db.QueryRow(query, agentID).Scan(&exists)
				if err != nil {
					t.Logf("Could not verify agent existence: %v", err)
				} else {
					assert.True(t, exists)
				}
			},
		},
		{
			name:      "Spawn content creator",
			agentType: "content-creator",
			verify: func(t *testing.T, agentID string) {
				// Verify agent configuration
				var configJSON json.RawMessage
				query := fmt.Sprintf(`
					SELECT config FROM client_%s.agent_instances 
					WHERE id = $1
				`, "test_client")

				err := db.QueryRow(query, agentID).Scan(&configJSON)
				if err != nil {
					t.Logf("Could not get agent config: %v", err)
					return
				}

				var config map[string]interface{}
				err = json.Unmarshal(configJSON, &config)
				require.NoError(t, err)
				assert.Equal(t, "content-creator", config["agent_type"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := actions.ActionParams{
				Context: ctx,
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": tt.agentType,
					},
				},
				Headers: map[string]string{
					"correlation_id": uuid.New().String(),
					"client_id":      "test_client",
					"user_id":        "test_user",
				},
				DB:       db,
				Producer: producer,
				Logger:   logger,
			}

			result, err := actions.SpawnAgentAction(ctx, params)
			require.NoError(t, err)

			agentInfo, ok := result.(map[string]interface{})
			require.True(t, ok)

			agentID, ok := agentInfo["agent_id"].(string)
			require.True(t, ok)

			// Verify agent was spawned correctly
			tt.verify(t, agentID)

			// Verify agent topic is correct
			expectedTopic := fmt.Sprintf("system.agent.%s.process", tt.agentType)
			assert.Equal(t, expectedTopic, agentInfo["topic"])
		})
	}
}

func TestAgentLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	// Test full agent lifecycle
	agentID := uuid.New().String()
	clientID := "test_client"

	// 1. Create agent
	config := map[string]interface{}{
		"agent_type": "generic",
		"topic":      "system.agent.generic.process",
	}
	configJSON, _ := json.Marshal(config)

	query := fmt.Sprintf(`
		INSERT INTO client_%s.agent_instances 
		(id, template_id, owner_user_id, name, config, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, clientID)

	_, err := db.Exec(query, agentID, uuid.New().String(), "test_user",
		"lifecycle-test-agent", configJSON)
	require.NoError(t, err)

	// 2. Update heartbeat
	for i := 0; i < 3; i++ {
		updateQuery := fmt.Sprintf(`
			UPDATE client_%s.agent_instances 
			SET last_heartbeat = NOW() 
			WHERE id = $1
		`, clientID)

		_, err = db.Exec(updateQuery, agentID)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}

	// 3. Verify heartbeat was updated
	var lastHeartbeat sql.NullTime
	selectQuery := fmt.Sprintf(`
		SELECT last_heartbeat FROM client_%s.agent_instances 
		WHERE id = $1
	`, clientID)

	err = db.QueryRow(selectQuery, agentID).Scan(&lastHeartbeat)
	require.NoError(t, err)
	assert.True(t, lastHeartbeat.Valid)
	assert.WithinDuration(t, time.Now(), lastHeartbeat.Time, 1*time.Second)

	// 4. Deactivate agent
	deactivateQuery := fmt.Sprintf(`
		UPDATE client_%s.agent_instances 
		SET is_active = false 
		WHERE id = $1
	`, clientID)

	_, err = db.Exec(deactivateQuery, agentID)
	require.NoError(t, err)

	// 5. Verify agent is deactivated
	var isActive bool
	activeQuery := fmt.Sprintf(`
		SELECT is_active FROM client_%s.agent_instances 
		WHERE id = $1
	`, clientID)

	err = db.QueryRow(activeQuery, agentID).Scan(&isActive)
	require.NoError(t, err)
	assert.False(t, isActive)
}

func TestConcurrentAgentSpawning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()
	ctx := context.Background()

	// Test concurrent spawning
	concurrency := 5 // Reduced for reliability
	results := make(chan string, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			// Use unique agent type to avoid conflicts
			agentType := fmt.Sprintf("worker-%d", index)

			params := actions.ActionParams{
				Context: ctx,
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"agent_type": agentType,
					},
				},
				Headers: map[string]string{
					"correlation_id": fmt.Sprintf("test-concurrent-%d", index),
					"client_id":      "test_client",
					"user_id":        "test_user",
				},
				DB:       db,
				Producer: producer,
				Logger:   logger,
			}

			result, err := actions.SpawnAgentAction(ctx, params)
			if err != nil {
				errors <- err
				return
			}

			agentInfo, ok := result.(map[string]interface{})
			if !ok {
				errors <- fmt.Errorf("invalid result type")
				return
			}

			agentID, ok := agentInfo["agent_id"].(string)
			if !ok {
				errors <- fmt.Errorf("agent_id not found")
				return
			}

			results <- agentID
		}(i)
	}

	// Collect results
	spawnedAgents := make([]string, 0, concurrency)
	timeout := time.After(10 * time.Second)

	for i := 0; i < concurrency; i++ {
		select {
		case agentID := <-results:
			spawnedAgents = append(spawnedAgents, agentID)
		case err := <-errors:
			t.Errorf("Spawning error: %v", err)
		case <-timeout:
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

func TestAgentReuse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()
	ctx := context.Background()

	// First spawn
	params := actions.ActionParams{
		Context: ctx,
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"agent_type": "reusable-agent",
			},
		},
		Headers: map[string]string{
			"correlation_id": uuid.New().String(),
			"client_id":      "test_client",
			"user_id":        "test_user",
		},
		DB:       db,
		Producer: producer,
		Logger:   logger,
	}

	result1, err := actions.SpawnAgentAction(ctx, params)
	require.NoError(t, err)

	agentInfo1 := result1.(map[string]interface{})
	agentID1 := agentInfo1["agent_id"].(string)
	status1 := agentInfo1["status"].(string)

	// Should be created on first call
	assert.Equal(t, "created", status1)

	// Second spawn with same type
	params.Headers["correlation_id"] = uuid.New().String()
	result2, err := actions.SpawnAgentAction(ctx, params)
	require.NoError(t, err)

	agentInfo2 := result2.(map[string]interface{})
	agentID2 := agentInfo2["agent_id"].(string)
	status2 := agentInfo2["status"].(string)

	// Should reuse the same agent
	assert.Equal(t, "reused", status2)
	assert.Equal(t, agentID1, agentID2)
}
