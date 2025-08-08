// test/e2e/scenarios/content_creation_test.go
package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestContentCreationPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	db := setupTestDB(t)
	producer := setupTestProducer(t)
	logger := zap.NewNop()
	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	// Test content creation pipeline using available actions
	workflow := models.WorkflowPlan{
		StartStep: "validate",
		Steps: map[string]models.Step{
			"validate": {
				Action:      "validate_input",
				Description: "Validate input for content creation",
				NextStep:    "spawn_creator",
			},
			"spawn_creator": {
				Action: "spawn_agent",
				Config: map[string]interface{}{
					"agent_type": "content-creator",
				},
				NextStep: "create_content",
			},
			"create_content": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_type": "content-creator",
					"action":     "generate_content",
					"data": map[string]interface{}{
						"topic":        "AI and Automation Trends",
						"content_type": "blog_post",
						"style":        "professional",
						"length":       "medium",
					},
				},
				NextStep: "transform",
			},
			"transform": {
				Action: "transform_data",
				Config: map[string]interface{}{
					"transformation": "uppercase",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := "test-e2e-content-" + uuid.New().String()
	headers := helpers.TestHeaders(correlationID)

	// Add required initial data
	initialData, _ := json.Marshal(map[string]interface{}{
		"action": "create_content",
		"data": map[string]interface{}{
			"message": "Create blog post about AI trends",
		},
	})

	// Execute
	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, initialData)
	require.NoError(t, err)

	// For local actions, check state immediately
	state := getWorkflowState(t, db, correlationID)
	assert.NotNil(t, state)

	// Verify workflow progressed
	assert.NotEqual(t, "", state.CurrentStep)

	// If workflow has remote actions, it will be in AWAITING_RESPONSES
	// Otherwise, it might be RUNNING or COMPLETED
	validStatuses := []string{
		string(orchestration.StatusRunning),
		string(orchestration.StatusAwaitingResponses),
		string(orchestration.StatusCompleted),
	}
	assert.Contains(t, validStatuses, string(state.Status))
}

func TestContentCreationWithMultipleFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	db := setupTestDB(t)
	producer := setupTestProducer(t)
	logger := zap.NewNop()
	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	formats := []string{"blog_post", "social_media", "email", "technical_doc"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			workflow := createContentWorkflow(format)
			correlationID := fmt.Sprintf("test-e2e-%s-%s", format, uuid.New().String())
			headers := helpers.TestHeaders(correlationID)

			initialData, _ := json.Marshal(map[string]interface{}{
				"action": "create_content",
				"data": map[string]interface{}{
					"message": fmt.Sprintf("Create %s content", format),
					"format":  format,
				},
			})

			err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, initialData)
			require.NoError(t, err)

			// Verify workflow was created
			state := getWorkflowState(t, db, correlationID)
			assert.NotNil(t, state)

			// Verify format is in the workflow config
			if state.WorkflowPlan.Steps["create_content"].Config != nil {
				config := state.WorkflowPlan.Steps["create_content"].Config
				if contentType, ok := config["content_type"]; ok {
					assert.Equal(t, format, contentType)
				}
			}
		})
	}
}

func createContentWorkflow(format string) models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "validate",
		Steps: map[string]models.Step{
			"validate": {
				Action:   "validate_input",
				NextStep: "create_content",
			},
			"create_content": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_type":   "content-creator",
					"content_type": format,
					"action":       "generate",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}
}

func verifyContentFormat(t *testing.T, correlationID string, format string) {
	// In a real implementation, this would check the actual content
	// For now, we just verify the workflow executed with the right format
	t.Logf("Verifying content format %s for correlation %s", format, correlationID)
}

func executeAndWait(t *testing.T, workflow models.WorkflowPlan, correlationID string) error {
	db := setupTestDB(t)
	producer := setupTestProducer(t)
	logger := zap.NewNop()
	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	headers := helpers.TestHeaders(correlationID)
	initialData, _ := json.Marshal(map[string]interface{}{
		"action": "execute",
		"data": map[string]interface{}{
			"message": "Execute workflow",
		},
	})

	return coordinator.ExecuteWorkflow(context.Background(), workflow, headers, initialData)
}
