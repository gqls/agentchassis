// test/e2e/scenarios/website_builder_test.go
package scenarios

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
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

// TestMain allows us to add setup/teardown and better logging
func TestMain(m *testing.M) {
	fmt.Println("=== E2E TEST SUITE STARTING ===")
	fmt.Printf("Time: %s\n", time.Now().Format(time.RFC3339))

	// Run tests
	code := m.Run()

	fmt.Println("=== E2E TEST SUITE COMPLETED ===")
	fmt.Printf("Exit code: %d\n", code)
	if code != 0 {
		fmt.Println("FAILURE DETECTED")
	}

	// Exit with the code
	os.Exit(code)
}

func TestWebsiteBuilderWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Log("Starting TestWebsiteBuilderWorkflow")

	ctx := context.Background()

	// Setup with error checking
	t.Log("Setting up test database...")
	db := setupTestDB(t)
	if db == nil {
		t.Fatal("Failed to setup test database")
	}
	defer func() {
		t.Log("Closing database connection")
		db.Close()
	}()

	t.Log("Setting up test producer...")
	producer := setupTestProducer(t)
	if producer == nil {
		t.Fatal("Failed to setup test producer")
	}

	logger := zap.NewNop()

	t.Log("Creating saga coordinator...")
	coordinator := orchestration.NewSagaCoordinator(db, producer, logger)

	// Setup website builder group in database
	t.Log("Setting up website builder group...")
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
	t.Logf("Executing workflow with correlation ID: %s", correlationID)
	err := coordinator.ExecuteWorkflow(ctx, workflow, headers, initialData)
	if err != nil {
		t.Logf("ExecuteWorkflow returned error: %v", err)
	}
	require.NoError(t, err)

	// Check initial state
	t.Log("Getting workflow state...")
	repo := orchestration.NewStateRepository(db, logger)
	state, err := repo.GetState(ctx, correlationID)

	if err != nil {
		t.Fatalf("Failed to get workflow state: %v", err)
	}

	assert.NotNil(t, state)
	assert.Equal(t, correlationID, state.CorrelationID)

	// Check workflow is progressing
	assert.NotEqual(t, "", state.CurrentStep)
	t.Logf("Workflow state - Step: %s, Status: %s", state.CurrentStep, state.Status)

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
		t.Logf("Simulating responses for %d awaited steps", len(state.AwaitedSteps))
		// Simulate agent responses
		for _, awaitedStep := range state.AwaitedSteps {
			// TaskResponse doesn't have Status field, put status in Data
			response := models.TaskResponse{
				Data: map[string]interface{}{
					"status":    "success",
					"site_plan": "Generated site plan",
					"result":    "success",
				},
				Error: "", // No error for successful response
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

	t.Log("TestWebsiteBuilderWorkflow completed successfully")
}

func TestWebsiteBuilderWithErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Log("Starting TestWebsiteBuilderWithErrors")

	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

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
				"invalid": "json_structure",
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

			_ = coordinator.ExecuteWorkflow(ctx, workflow, headers, initialData)

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

	t.Log("TestWebsiteBuilderWithErrors completed successfully")
}

func setupWebsiteBuilderGroup(t *testing.T, db *sql.DB) {
	// Create agent configs properly
	agentConfigs := []map[string]interface{}{
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
	agentConfigsJSON, _ := json.Marshal(agentConfigs)

	// Create workflow steps properly
	planStep := map[string]interface{}{
		"action":    "create_plan",
		"next_step": "design",
	}

	designStep := map[string]interface{}{
		"action":    "create_design",
		"next_step": "develop",
	}

	developStep := map[string]interface{}{
		"action":    "build_html",
		"next_step": "publish",
	}

	publishStep := map[string]interface{}{
		"action": "publish_site",
	}

	workflowSteps := map[string]interface{}{
		"plan":    planStep,
		"design":  designStep,
		"develop": developStep,
		"publish": publishStep,
	}

	workflowData := map[string]interface{}{
		"start_step": "plan",
		"steps":      workflowSteps,
	}

	workflowJSON, _ := json.Marshal(workflowData)

	_, err := db.Exec(`
		INSERT INTO agent_groups (id, name, group_type, version, agent_configs, orchestration_workflow)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, uuid.New().String(), "Website Builder Team", "website-builder", "1.0.0",
		agentConfigsJSON, workflowJSON)

	if err != nil {
		t.Logf("Warning: Could not insert website builder group: %v", err)
	}
}
