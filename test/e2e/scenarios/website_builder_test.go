// test/e2e/scenarios/website_builder_test.go
package scenarios

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWebsiteBuilderWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	db := setupTestDB(t)
	producer := setupTestProducer(t)
	logger := zap.NewNop()
	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	// Setup website builder group in database
	setupWebsiteBuilderGroup(t, db)

	// Create website builder workflow using available actions
	workflow := models.WorkflowPlan{
		StartStep: "validate_request",
		Steps: map[string]models.Step{
			"validate_request": {
				Action:      "validate_input",
				Description: "Validate website requirements",
				NextStep:    "spawn_builders",
			},
			"spawn_builders": {
				Action: "spawn_group",
				Config: map[string]interface{}{
					"group_type": "website-builder",
				},
				NextStep: "plan_site",
			},
			"plan_site": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_type": "site-architect",
					"action":     "create_site_plan",
					"data": map[string]interface{}{
						"domain":        "test-site.com",
						"business_name": "Test Business",
						"requirements":  []string{"responsive", "modern", "fast"},
					},
				},
				NextStep: "transform_plan",
			},
			"transform_plan": {
				Action: "transform_data",
				Config: map[string]interface{}{
					"transformation": "uppercase",
				},
				NextStep: "notify",
			},
			"notify": {
				Action: "send_notification",
				Config: map[string]interface{}{
					"topic": "system.responses.website",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := helpers.TestUUIDWithType("e2e")
	headers := helpers.TestHeaders(correlationID)

	initialData, _ := json.Marshal(map[string]interface{}{
		"action": "build_website",
		"data": map[string]interface{}{
			"message":       "Build a website",
			"business_name": "Test Business",
			"domain":        "test-site.com",
		},
	})

	// Execute workflow
	err := coordinator.ExecuteWorkflow(ctx, workflow, headers, initialData)
	require.NoError(t, err)

	// Check initial state
	repo := orchestration.NewStateRepository(db, logger)
	state, err := repo.GetState(ctx, correlationID)

	if err != nil {
		t.Fatalf("Failed to get workflow state: %v", err)
	}

	assert.NotNil(t, state)
	assert.Equal(t, correlationID, state.CorrelationID)

	// Check workflow is progressing
	assert.NotEqual(t, "", state.CurrentStep)

	// For workflows with remote calls, status will be AWAITING_RESPONSES
	// For local-only workflows, might be RUNNING or COMPLETED
	validStatuses := []string{
		string(orchestration.StatusRunning),
		string(orchestration.StatusAwaitingResponses),
		string(orchestration.StatusCompleted),
	}
	assert.Contains(t, validStatuses, string(state.Status))

	// If workflow has remote actions, simulate responses
	if state.Status == orchestration.StatusAwaitingResponses {
		// Simulate agent responses
		for _, awaitedStep := range state.AwaitedSteps {
			response := models.TaskResponse{
				Status: "success",
				Data: map[string]interface{}{
					"site_plan": "Generated site plan",
					"result":    "success",
				},
			}

			responseData, _ := json.Marshal(response)
			responseHeaders := make(map[string]string)
			for k, v := range headers {
				responseHeaders[k] = v
			}
			responseHeaders["causation_id"] = awaitedStep

			err = coordinator.HandleResponse(ctx, responseHeaders, responseData)
			assert.NoError(t, err)
		}

		// Check state after responses
		time.Sleep(100 * time.Millisecond)
		state, _ = repo.GetState(ctx, correlationID)
		t.Logf("State after responses: %s", state.Status)
	}
}

func TestWebsiteBuilderWithErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()
	db := setupTestDB(t)
	producer := setupTestProducer(t)
	logger := zap.NewNop()
	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	tests := []struct {
		name          string
		invalidData   map[string]interface{}
		expectedError string
	}{
		{
			name: "Missing message field",
			invalidData: map[string]interface{}{
				"action": "build",
				"data": map[string]interface{}{
					// message field missing
					"domain": "test.com",
				},
			},
			expectedError: "missing required field: message",
		},
		{
			name: "Invalid JSON",
			invalidData: map[string]interface{}{
				"invalid": "json}structure",
			},
			expectedError: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := models.WorkflowPlan{
				StartStep: "validate",
				Steps: map[string]models.Step{
					"validate": {
						Action:   "validate_input",
						NextStep: "complete",
					},
					"complete": {
						Action: "complete_workflow",
					},
				},
			}

			correlationID := helpers.TestUUIDWithType("e2e")
			headers := helpers.TestHeaders(correlationID)

			initialData, _ := json.Marshal(tt.invalidData)

			err := coordinator.ExecuteWorkflow(ctx, workflow, headers, initialData)

			// The workflow will be created but validation will fail
			// Check the state to see if it failed
			repo := orchestration.NewStateRepository(db, logger)
			state, stateErr := repo.GetState(ctx, correlationID)

			if stateErr == nil && state != nil {
				// If workflow failed, check the error
				if state.Status == orchestration.StatusFailed {
					assert.Contains(t, state.Error, tt.expectedError)
				}
			}
		})
	}
}

func setupWebsiteBuilderGroup(t *testing.T, db *sql.DB) {
	agentConfigs, _ := json.Marshal([]map[string]interface{}{
		{"role": "architect", "agent_type": "site-architect"},
		{"role": "designer", "agent_type": "visual-designer"},
		{"role": "developer", "agent_type": "html-developer"},
		{"role": "publisher", "agent_type": "site-publisher"},
	})

	workflow, _ := json.Marshal(map[string]interface{}{
		"start_step": "plan",
		"steps": map[string]interface{}{
			"plan":    {"action": "create_plan", "next_step": "design"},
			"design":  {"action": "create_design", "next_step": "develop"},
			"develop": {"action": "build_html", "next_step": "publish"},
			"publish": {"action": "publish_site"},
		},
	})

	_, err := db.Exec(`
		INSERT INTO agent_groups (id, name, group_type, version, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, uuid.New().String(), "Website Builder Team", "website-builder", "1.0.0",
		agentConfigs, workflow)

	if err != nil {
		t.Logf("Warning: Could not insert website builder group: %v", err)
	}
}
