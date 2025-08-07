// test/e2e/scenarios/website_builder_test.go
package scenarios

import (
	"context"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsiteBuilderWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	ctx := context.Background()
	coordinator := setupTestCoordinator(t)

	// Create website builder workflow
	workflow := models.WorkflowPlan{
		StartStep: "spawn_builders",
		Steps: map[string]models.Step{
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
					"agent_role": "architect",
					"action":     "create_site_plan",
					"data": map[string]interface{}{
						"domain":        "test-site.com",
						"business_name": "Test Business",
						"requirements":  []string{"responsive", "modern", "fast"},
					},
				},
				NextStep: "design_site",
			},
			"design_site": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_role": "designer",
					"action":     "create_design",
				},
				NextStep: "develop_site",
			},
			"develop_site": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_role": "developer",
					"action":     "build_html",
				},
				NextStep: "publish_site",
			},
			"publish_site": {
				Action: "call_agent",
				Config: map[string]interface{}{
					"agent_role": "publisher",
					"action":     "publish_site",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := "test-e2e-website-" + time.Now().Format("20060102150405")
	headers := helpers.TestHeaders(correlationID)

	// Execute workflow
	err := coordinator.ExecuteWorkflow(ctx, workflow, headers, nil)
	require.NoError(t, err)

	// Wait for completion
	helpers.WaitForCondition(t, 30*time.Second, func() bool {
		state := getWorkflowState(t, correlationID)
		return state.Status == "COMPLETED" || state.Status == "FAILED"
	})

	// Verify results
	state := getWorkflowState(t, correlationID)
	assert.Equal(t, "COMPLETED", state.Status)

	// Check collected data
	collectedData := state.CollectedData.(map[string]interface{})
	assert.NotEmpty(t, collectedData["site_plan"])
	assert.NotEmpty(t, collectedData["design"])
	assert.NotEmpty(t, collectedData["html_output"])
	assert.NotEmpty(t, collectedData["publish_result"])
}

func TestWebsiteBuilderWithErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	tests := []struct {
		name          string
		failAtStep    string
		expectedError string
	}{
		{
			name:          "Invalid domain",
			failAtStep:    "plan_site",
			expectedError: "invalid domain format",
		},
		{
			name:          "Design failure",
			failAtStep:    "design_site",
			expectedError: "design generation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test error scenarios
			// Implementation details...
		})
	}
}
