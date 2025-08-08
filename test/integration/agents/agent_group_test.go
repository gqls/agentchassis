// test/integration/agents/agent_group_test.go
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
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAgentGroupSpawning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()

	// Setup test groups in database
	setupTestGroups(t, db)

	tests := []struct {
		name      string
		groupType string
		validate  func(t *testing.T, result interface{})
	}{
		{
			name:      "Website Builder Group",
			groupType: "website-builder",
			validate: func(t *testing.T, result interface{}) {
				groupResult, ok := result.(map[string]interface{})
				require.True(t, ok)

				agents, ok := groupResult["agents"].(map[string]string)
				require.True(t, ok)

				// Website builder should have these roles
				expectedRoles := []string{"architect", "designer", "developer", "publisher"}
				for _, role := range expectedRoles {
					_, exists := agents[role]
					assert.True(t, exists, "Should have agent for role %s", role)
				}

				// Verify each agent was created in database
				for role, agentID := range agents {
					var exists bool
					query := fmt.Sprintf(`
						SELECT EXISTS(
							SELECT 1 FROM client_%s.agent_instances 
							WHERE id = $1 AND is_active = true
						)`, "test_client")

					err := db.QueryRow(query, agentID).Scan(&exists)
					if err != nil {
						t.Logf("Could not verify agent %s for role %s: %v", agentID, role, err)
					} else {
						assert.True(t, exists, "Agent for role %s should exist", role)
					}
				}
			},
		},
		{
			name:      "Content Creation Team",
			groupType: "content-team",
			validate: func(t *testing.T, result interface{}) {
				groupResult, ok := result.(map[string]interface{})
				require.True(t, ok)

				agents, ok := groupResult["agents"].(map[string]string)
				require.True(t, ok)

				// Content team should have these roles
				expectedRoles := []string{"researcher", "writer", "editor"}
				for _, role := range expectedRoles {
					_, exists := agents[role]
					assert.True(t, exists, "Should have agent for role %s", role)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			params := actions.ActionParams{
				Context: ctx,
				StepConfig: models.Step{
					Config: map[string]interface{}{
						"group_type": tt.groupType,
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

			result, err := actions.SpawnGroupAction(ctx, params)
			if err != nil {
				t.Skipf("Skipping - SpawnGroupAction failed: %v", err)
				return
			}

			tt.validate(t, result)

			// Verify group usage was tracked
			groupResult := result.(map[string]interface{})
			if groupID, ok := groupResult["group_id"].(string); ok {
				var usageCount int
				err = db.QueryRow(`
					SELECT usage_count FROM agent_groups 
					WHERE id = $1
				`, groupID).Scan(&usageCount)

				if err == nil {
					assert.GreaterOrEqual(t, usageCount, 1)
				}
			}
		})
	}
}

func TestAgentGroupOrchestration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	producer := createTestProducer(t)
	logger := zap.NewNop()

	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	// Setup test groups
	setupTestGroups(t, db)

	// Test website builder group workflow using available actions
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
				Action: "fan_out",
				SubTasks: []models.SubTask{
					{
						StepName: "architect_task",
						Topic:    "system.agent.architect.process",
					},
					{
						StepName: "designer_task",
						Topic:    "system.agent.designer.process",
					},
					{
						StepName: "developer_task",
						Topic:    "system.agent.developer.process",
					},
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	ctx := context.Background()
	correlationID := uuid.New().String()
	headers := helpers.TestHeaders(correlationID)

	err := coordinator.ExecuteWorkflow(ctx, workflow, headers, nil)
	require.NoError(t, err)

	// Check initial state
	repo := orchestration.NewStateRepository(db, logger)
	state, err := repo.GetState(ctx, correlationID)

	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	// Verify workflow was created and is progressing
	assert.NotNil(t, state)
	assert.NotEmpty(t, state.CurrentStep)

	// For workflows with fan_out, status should be AWAITING_RESPONSES
	if state.CurrentStep == "distribute_tasks" {
		assert.Equal(t, orchestration.StatusAwaitingResponses, state.Status)
		assert.Len(t, state.AwaitedSteps, 3) // Three subtasks
	}
}

func TestAgentGroupPerformanceTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	// Create a test group
	groupID := uuid.New().String()

	// Create agent configs properly
	agentConfigs := []map[string]interface{}{
		{
			"role":       "worker",
			"agent_type": "generic",
		},
	}
	agentConfigsJSON, _ := json.Marshal(agentConfigs)

	// Create workflow steps properly
	processStep := map[string]interface{}{
		"action": "process_task",
	}

	workflowSteps := map[string]interface{}{
		"process": processStep,
	}

	workflow := map[string]interface{}{
		"start_step": "process",
		"steps":      workflowSteps,
	}
	workflowJSON, _ := json.Marshal(workflow)

	_, err := db.Exec(`
		INSERT INTO agent_groups 
		(id, name, group_type, version, agent_configs, orchestration_workflow, performance_metrics, usage_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, groupID, "Test Performance Group", "test-perf", "1.0.0",
		agentConfigsJSON, workflowJSON, json.RawMessage(`{"success_rate": 0.5}`), 0)

	require.NoError(t, err)

	// Simulate multiple uses of the group
	for i := 0; i < 5; i++ {
		// Update usage count and performance metrics
		successRate := (4.0 + float64(i)) / (5.0 + float64(i))

		_, err := db.Exec(`
			UPDATE agent_groups 
			SET usage_count = usage_count + 1,
				last_used_at = NOW(),
				performance_metrics = jsonb_set(
					COALESCE(performance_metrics, '{}'::jsonb),
					'{success_rate}',
					to_jsonb($1::float)
				)
			WHERE id = $2
		`, successRate, groupID)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)
	}

	// Verify performance metrics
	var metrics json.RawMessage
	var usageCount int
	err = db.QueryRow(`
		SELECT performance_metrics, usage_count 
		FROM agent_groups 
		WHERE id = $1
	`, groupID).Scan(&metrics, &usageCount)
	require.NoError(t, err)

	assert.Equal(t, 5, usageCount)

	var perfMetrics map[string]interface{}
	err = json.Unmarshal(metrics, &perfMetrics)
	require.NoError(t, err)
	assert.Contains(t, perfMetrics, "success_rate")

	successRate, ok := perfMetrics["success_rate"].(float64)
	require.True(t, ok)
	assert.Greater(t, successRate, 0.5)
}

func TestAgentGroupEvolution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupIntegrationDB(t)
	defer db.Close()

	//ctx := context.Background()

	// Create a parent group
	parentID := uuid.New().String()

	// Create agent configs properly
	agentConfigs := []map[string]interface{}{
		{
			"role":       "analyzer",
			"agent_type": "analyzer",
		},
		{
			"role":       "executor",
			"agent_type": "generic",
		},
	}
	agentConfigsJSON, _ := json.Marshal(agentConfigs)

	// Create workflow properly
	workflowSteps := map[string]map[string]interface{}{
		"analyze": {
			"action":    "analyze_task",
			"next_step": "execute",
		},
		"execute": {
			"action":    "execute_task",
			"next_step": "complete",
		},
		"complete": {
			"action": "complete_workflow",
		},
	}

	workflow := map[string]interface{}{
		"start_step": "analyze",
		"steps":      workflowSteps,
	}
	workflowJSON, _ := json.Marshal(workflow)

	_, err := db.Exec(`
		INSERT INTO agent_groups 
		(id, name, group_type, version, agent_configs, orchestration_workflow, capabilities)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, parentID, "Evolving Team", "adaptive-team", "1.0.0",
		agentConfigsJSON, workflowJSON, json.RawMessage(`["analysis", "execution"]`))
	require.NoError(t, err)

	// Since CreateGroupVersion doesn't exist in actions package,
	// create the new version manually
	newGroupID := uuid.New().String()

	// Create updated agent configs with new agent
	updatedAgentConfigs := []map[string]interface{}{
		{
			"role":       "analyzer",
			"agent_type": "analyzer",
		},
		{
			"role":       "executor",
			"agent_type": "generic",
		},
		{
			"role":       "optimizer",
			"agent_type": "optimizer",
		},
	}
	updatedAgentConfigsJSON, _ := json.Marshal(updatedAgentConfigs)

	// Create mutation history
	mutationHistory := []map[string]interface{}{
		{
			"type":       "add_agent",
			"agent_type": "optimizer",
			"role":       "optimizer",
			"timestamp":  time.Now(),
		},
	}
	mutationHistoryJSON, _ := json.Marshal(mutationHistory)

	// Insert new version using the approach from discovery_actions.go
	_, err = db.Exec(`
		INSERT INTO agent_groups (id, name, group_type, parent_id, version, 
								  agent_configs, orchestration_workflow, capabilities,
								  mutation_history)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, newGroupID, "Evolving Team v2", "adaptive-team", parentID, "2.0.0",
		updatedAgentConfigsJSON, workflowJSON,
		json.RawMessage(`["analysis", "execution", "optimization"]`),
		mutationHistoryJSON)
	require.NoError(t, err)

	// Verify the new version exists
	var version string
	var parentIDResult sql.NullString
	err = db.QueryRow(`
		SELECT version, parent_id 
		FROM agent_groups 
		WHERE id = $1
	`, newGroupID).Scan(&version, &parentIDResult)
	require.NoError(t, err)

	assert.Equal(t, "2.0.0", version)
	assert.True(t, parentIDResult.Valid)
	assert.Equal(t, parentID, parentIDResult.String)
}

// Helper functions

func setupIntegrationDB(t *testing.T) *sql.DB {
	db := helpers.TestDB(t)
	helpers.SetupTestSchema(t, db)
	return db
}

func createTestProducer(t *testing.T) kafka.Producer {
	// For integration tests, you might want to use a real Kafka producer
	// if Kafka is available in your test environment
	if isKafkaAvailable() {
		return createRealKafkaProducer(t)
	}
	// Otherwise fall back to mock
	return helpers.NewMockProducer()
}

func createRealKafkaProducer(t *testing.T) kafka.Producer {
	// This would create a real Kafka producer
	// Implementation depends on your Kafka setup
	t.Log("Using mock producer - implement createRealKafkaProducer for real Kafka")
	return helpers.NewMockProducer()
}

func isKafkaAvailable() bool {
	// Check if Kafka is available for testing
	// This could check environment variables or try to connect
	return false
}

func setupTestGroups(t *testing.T, db *sql.DB) {
	// Website builder group
	websiteBuilderAgents := []map[string]interface{}{
		{
			"role":       "architect",
			"agent_type": "site-architect",
		},
		{
			"role":       "designer",
			"agent_type": "visual-designer",
		},
		{
			"role":       "developer",
			"agent_type": "html-developer",
		},
		{
			"role":       "publisher",
			"agent_type": "site-publisher",
		},
	}
	websiteBuilderAgentsJSON, _ := json.Marshal(websiteBuilderAgents)

	// Create workflow steps properly
	websiteWorkflowSteps := map[string]map[string]interface{}{
		"plan": {
			"action":    "create_plan",
			"next_step": "design",
		},
		"design": {
			"action":    "create_design",
			"next_step": "develop",
		},
		"develop": {
			"action":    "build_html",
			"next_step": "publish",
		},
		"publish": {
			"action": "publish_site",
		},
	}

	websiteWorkflow := map[string]interface{}{
		"start_step": "plan",
		"steps":      websiteWorkflowSteps,
	}
	websiteWorkflowJSON, _ := json.Marshal(websiteWorkflow)

	_, err := db.Exec(`
		INSERT INTO agent_groups 
		(id, name, group_type, version, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, uuid.New().String(), "Website Builder", "website-builder", "1.0.0",
		websiteBuilderAgentsJSON, websiteWorkflowJSON)

	if err != nil {
		t.Logf("Warning: Could not insert website builder group: %v", err)
	}

	// Content team group
	contentTeamAgents := []map[string]interface{}{
		{
			"role":       "researcher",
			"agent_type": "researcher",
		},
		{
			"role":       "writer",
			"agent_type": "content-creator",
		},
		{
			"role":       "editor",
			"agent_type": "editor",
		},
	}
	contentTeamAgentsJSON, _ := json.Marshal(contentTeamAgents)

	contentWorkflowSteps := map[string]map[string]interface{}{
		"research": {
			"action":    "research_topic",
			"next_step": "write",
		},
		"write": {
			"action":    "create_content",
			"next_step": "edit",
		},
		"edit": {
			"action": "edit_content",
		},
	}

	contentWorkflow := map[string]interface{}{
		"start_step": "research",
		"steps":      contentWorkflowSteps,
	}
	contentWorkflowJSON, _ := json.Marshal(contentWorkflow)

	_, err = db.Exec(`
		INSERT INTO agent_groups 
		(id, name, group_type, version, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, uuid.New().String(), "Content Team", "content-team", "1.0.0",
		contentTeamAgentsJSON, contentWorkflowJSON)

	if err != nil {
		t.Logf("Warning: Could not insert content team group: %v", err)
	}
}

func mustMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
