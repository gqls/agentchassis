// test/integration/agents/agent_group_test.go
package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentGroupSpawning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	producer := createRealKafkaProducer(t)
	defer producer.Close()

	tests := []struct {
		name      string
		groupType string
		validate  func(t *testing.T, result map[string]interface{})
	}{
		{
			name:      "Website Builder Group",
			groupType: "website-builder",
			validate: func(t *testing.T, result map[string]interface{}) {
				agents := result["agents"].(map[string]string)
				assert.Contains(t, agents, "architect")
				assert.Contains(t, agents, "designer")
				assert.Contains(t, agents, "developer")
				assert.Contains(t, agents, "publisher")

				// Verify each agent was created
				for role, agentID := range agents {
					var exists bool
					err := db.QueryRow(`
                        SELECT EXISTS(
                            SELECT 1 FROM client_demo_client.agent_instances 
                            WHERE id = $1 AND is_active = true
                        )`, agentID).Scan(&exists)
					require.NoError(t, err)
					assert.True(t, exists, "Agent for role %s should exist", role)
				}
			},
		},
		{
			name:      "Content Creation Team",
			groupType: "content-team",
			validate: func(t *testing.T, result map[string]interface{}) {
				agents := result["agents"].(map[string]string)
				assert.Contains(t, agents, "researcher")
				assert.Contains(t, agents, "writer")
				assert.Contains(t, agents, "editor")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := actions.ActionParams{
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"group_type": tt.groupType,
					},
				},
				Headers: map[string]string{
					"correlation_id": "test-group-" + uuid.New().String(),
					"client_id":      "demo_client",
					"user_id":        "test_user",
				},
				DB:       db,
				Producer: producer,
				Logger:   zap.NewNop(),
			}

			result, err := actions.SpawnGroupAction(context.Background(), params)
			require.NoError(t, err)

			groupResult := result.(map[string]interface{})
			assert.NotEmpty(t, groupResult["group_id"])
			assert.Equal(t, tt.groupType, groupResult["group_type"])

			tt.validate(t, groupResult)

			// Verify group usage was tracked
			var usageCount int
			err = db.QueryRow(`
                SELECT usage_count FROM agent_groups 
                WHERE id = $1
            `, groupResult["group_id"]).Scan(&usageCount)
			require.NoError(t, err)
			assert.Equal(t, 1, usageCount)
		})
	}
}

func TestAgentGroupOrchestration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	coordinator := orchestration.NewSagaCoordinator(db, createRealKafkaProducer(t), zap.NewNop())

	// Test website builder group workflow
	workflow := models.WorkflowPlan{
		StartStep: "spawn_builders",
		Steps: map[string]models.Step{
			"spawn_builders": {
				Action: "spawn_group",
				Config: map[string]interface{}{
					"group_type": "website-builder",
				},
				NextStep: "distribute_tasks",
			},
			"distribute_tasks": {
				Action: "fan_out_to_group",
				Config: map[string]interface{}{
					"tasks": []map[string]interface{}{
						{
							"role":   "architect",
							"action": "create_site_plan",
							"data": map[string]interface{}{
								"domain":       "test-group-site.com",
								"requirements": []string{"modern", "responsive"},
							},
						},
						{
							"role":   "designer",
							"action": "create_design",
							"data": map[string]interface{}{
								"style":        "minimalist",
								"color_scheme": "blue",
							},
						},
						{
							"role":   "developer",
							"action": "build_html",
							"data": map[string]interface{}{
								"framework": "vanilla",
								"optimize":  true,
							},
						},
					},
				},
				NextStep: "aggregate_results",
			},
			"aggregate_results": {
				Action: "wait_for_responses",
				Config: map[string]interface{}{
					"expected_responses": 3,
					"timeout":            "5m",
				},
				NextStep: "publish",
			},
			"publish": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"use_group_agent": true,
					"role":            "publisher",
					"action":          "publish_site",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := "test-group-orchestration-" + uuid.New().String()
	headers := helpers.TestHeaders(correlationID)

	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
	require.NoError(t, err)

	// Wait for workflow to complete
	helpers.WaitForCondition(t, 30*time.Second, func() bool {
		var status string
		db.QueryRow(`
            SELECT status FROM orchestrator_state 
            WHERE correlation_id = $1
        `, correlationID).Scan(&status)
		return status == "COMPLETED" || status == "FAILED"
	})

	// Verify workflow completed successfully
	var state models.OrchestratorState
	err = db.QueryRow(`
        SELECT status, collected_data, execution_metadata
        FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&state.Status, &state.CollectedData, &state.ExecutionMetadata)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", state.Status)

	// Verify all group agents participated
	collectedData := state.CollectedData.(map[string]interface{})
	assert.Contains(t, collectedData, "architect_response")
	assert.Contains(t, collectedData, "designer_response")
	assert.Contains(t, collectedData, "developer_response")
	assert.Contains(t, collectedData, "publisher_response")
}

func TestAgentGroupPerformanceTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	groupID := "11111111-1111-1111-1111-111111111111" // Website builder group

	// Simulate multiple uses of the group
	for i := 0; i < 5; i++ {
		// Update usage count and performance metrics
		_, err := db.Exec(`
            UPDATE agent_groups 
            SET usage_count = usage_count + 1,
                last_used_at = NOW(),
                performance_metrics = jsonb_set(
                    performance_metrics,
                    '{success_rate}',
                    to_jsonb((4.0 + $1::float) / (5.0 + $1::float))
                )
            WHERE id = $2
        `, i, groupID)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
	}

	// Verify performance metrics
	var metrics json.RawMessage
	var usageCount int
	err := db.QueryRow(`
        SELECT performance_metrics, usage_count 
        FROM agent_groups 
        WHERE id = $1
    `, groupID).Scan(&metrics, &usageCount)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, usageCount, 5)

	var perfMetrics map[string]interface{}
	json.Unmarshal(metrics, &perfMetrics)
	assert.Contains(t, perfMetrics, "success_rate")
}

func TestAgentGroupMutation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	db := helpers.TestDB(t)
	defer db.Close()

	// Create a new group that will mutate
	groupID := uuid.New().String()
	_, err := db.Exec(`
        INSERT INTO agent_groups 
        (id, name, group_type, agent_configs, orchestration_workflow, capabilities)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, groupID, "Evolving Team", "adaptive-team",
		json.RawMessage(`[
            {"role": "analyzer", "agent_type": "analyzer"},
            {"role": "executor", "agent_type": "generic"}
        ]`),
		json.RawMessage(`{
            "start_step": "analyze",
            "steps": {
                "analyze": {"action": "analyze_task", "next_step": "execute"},
                "execute": {"action": "execute_task", "next_step": "complete"},
                "complete": {"action": "complete_workflow"}
            }
        }`),
		json.RawMessage(`["analysis", "execution"]`))
	require.NoError(t, err)

	// Simulate group mutation based on performance
	mutation := map[string]interface{}{
		"timestamp": time.Now(),
		"reason":    "performance_optimization",
		"changes": map[string]interface{}{
			"added_agent": map[string]string{
				"role":       "optimizer",
				"agent_type": "optimizer",
			},
			"modified_workflow": map[string]interface{}{
				"added_step": "optimize",
				"position":   "between_analyze_and_execute",
			},
		},
	}

	// Apply mutation
	_, err = db.Exec(`
        UPDATE agent_groups 
        SET agent_configs = agent_configs || $1::jsonb,
            orchestration_workflow = jsonb_set(
                orchestration_workflow,
                '{steps,analyze,next_step}',
                '"optimize"'
            ),
            mutation_history = mutation_history || $2::jsonb,
            version = '1.1.0'
        WHERE id = $3
    `,
		json.RawMessage(`[{"role": "optimizer", "agent_type": "optimizer"}]`),
		json.RawMessage(fmt.Sprintf(`[%s]`, mustMarshal(mutation))),
		groupID)
	require.NoError(t, err)

	// Verify mutation
	var agentConfigs json.RawMessage
	var mutationHistory json.RawMessage
	var version string
	err = db.QueryRow(`
        SELECT agent_configs, mutation_history, version 
        FROM agent_groups 
        WHERE id = $1
    `, groupID).Scan(&agentConfigs, &mutationHistory, &version)
	require.NoError(t, err)

	assert.Equal(t, "1.1.0", version)

	var agents []map[string]interface{}
	json.Unmarshal(agentConfigs, &agents)
	assert.Len(t, agents, 3) // Original 2 + 1 new

	var mutations []map[string]interface{}
	json.Unmarshal(mutationHistory, &mutations)
	assert.Len(t, mutations, 1)
}
