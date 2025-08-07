// test/e2e/scenarios/content_creation_test.go
package scenarios

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"testing"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentCreationPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	coordinator := setupTestCoordinator(t)

	// Test content creation pipeline
	workflow := models.WorkflowPlan{
		StartStep: "research",
		Steps: map[string]models.Step{
			"research": {
				Action: "call_agent",
				Topic:  "system.agent.web-search.process",
				Config: map[string]interface{}{
					"action": "search",
					"data": map[string]interface{}{
						"query":       "latest trends in AI and automation",
						"max_results": 5,
					},
				},
				NextStep: "analyze",
			},
			"analyze": {
				Action: "call_agent",
				Topic:  "system.agent.reasoning.process",
				Config: map[string]interface{}{
					"action": "analyze",
					"data": map[string]interface{}{
						"content_to_review": "{{collected_data.research.results}}",
						"review_criteria": []string{
							"relevance",
							"accuracy",
							"comprehensiveness",
						},
					},
				},
				NextStep: "create_content",
			},
			"create_content": {
				Action: "call_agent",
				Topic:  "system.agent.content-creator.process",
				Config: map[string]interface{}{
					"action": "generate_content",
					"data": map[string]interface{}{
						"topic":         "AI and Automation Trends",
						"content_type":  "blog_post",
						"research_data": "{{collected_data.research}}",
						"analysis":      "{{collected_data.analyze}}",
						"style":         "professional",
						"length":        "long",
					},
				},
				NextStep: "review",
			},
			"review": {
				Action: "call_agent",
				Topic:  "system.agent.reasoning.process",
				Config: map[string]interface{}{
					"action": "review_content",
					"data": map[string]interface{}{
						"content": "{{collected_data.create_content.generated_text}}",
						"criteria": []string{
							"grammar",
							"clarity",
							"engagement",
						},
					},
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := "test-e2e-content-" + time.Now().Format("20060102150405")
	headers := helpers.TestHeaders(correlationID)

	// Execute
	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
	require.NoError(t, err)

	// Wait and verify
	helpers.WaitForCondition(t, 60*time.Second, func() bool {
		state := getWorkflowState(t, correlationID)
		return state.Status == "COMPLETED"
	})

	state := getWorkflowState(t, correlationID)
	assert.Equal(t, "COMPLETED", state.Status)

	// Verify content was created
	collectedData := state.CollectedData.(map[string]interface{})
	content := collectedData["create_content"].(map[string]interface{})
	assert.NotEmpty(t, content["generated_text"])
	assert.Contains(t, content["generated_text"].(string), "AI")
}

func TestContentCreationWithMultipleFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	formats := []string{"blog_post", "social_media", "email", "technical_doc"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			// Test different content formats
			workflow := createContentWorkflow(format)
			correlationID := fmt.Sprintf("test-e2e-%s-%d", format, time.Now().Unix())

			err := executeAndWait(t, workflow, correlationID)
			require.NoError(t, err)

			// Verify format-specific requirements
			verifyContentFormat(t, correlationID, format)
		})
	}
}
