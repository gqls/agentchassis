// test/e2e/scenarios/human_in_loop_test.go
package scenarios

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gqls/agentchassis/test/unit/helpers"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanInLoopWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	coordinator := setupTestCoordinator(t)
	db := getTestDB(t)

	// Workflow that requires human approval
	workflow := models.WorkflowPlan{
		StartStep: "generate_proposal",
		Steps: map[string]models.Step{
			"generate_proposal": {
				Action: "call_agent",
				Topic:  "system.agent.content-creator.process",
				Config: map[string]interface{}{
					"action": "generate_proposal",
					"data": map[string]interface{}{
						"type":     "marketing_campaign",
						"budget":   50000,
						"duration": "3 months",
					},
				},
				NextStep: "request_approval",
			},
			"request_approval": {
				Action: "pause_for_human",
				Config: map[string]interface{}{
					"approval_type": "proposal_review",
					"required_role": "manager",
					"timeout":       "24h",
					"notification": map[string]interface{}{
						"type":     "email",
						"template": "proposal_approval_request",
					},
					"data_to_review": "{{collected_data.generate_proposal}}",
				},
				NextStep: "check_approval",
			},
			"check_approval": {
				Action: "conditional",
				Config: map[string]interface{}{
					"condition":  "{{human_response.approved}} == true",
					"true_step":  "execute_campaign",
					"false_step": "revise_proposal",
				},
			},
			"revise_proposal": {
				Action: "call_agent",
				Topic:  "system.agent.content-creator.process",
				Config: map[string]interface{}{
					"action": "revise_proposal",
					"data": map[string]interface{}{
						"original_proposal": "{{collected_data.generate_proposal}}",
						"feedback":          "{{human_response.feedback}}",
					},
				},
				NextStep: "request_approval", // Loop back for re-approval
			},
			"execute_campaign": {
				Action: "call_agent",
				Topic:  "system.agent.campaign-executor.process",
				Config: map[string]interface{}{
					"action": "launch_campaign",
					"data":   "{{collected_data.generate_proposal}}",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := "test-human-loop-" + uuid.New().String()
	headers := helpers.TestHeaders(correlationID)

	// Start workflow
	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
	require.NoError(t, err)

	// Wait for workflow to pause for human input
	helpers.WaitForCondition(t, 10*time.Second, func() bool {
		var status string
		db.QueryRow(`
            SELECT status FROM orchestrator_state 
            WHERE correlation_id = $1
        `, correlationID).Scan(&status)
		return status == "PAUSED_FOR_HUMAN"
	})

	// Verify pause state
	var state models.OrchestratorState
	var pauseData json.RawMessage
	err = db.QueryRow(`
        SELECT status, current_step, execution_metadata
        FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&state.Status, &state.CurrentStep, &pauseData)

	require.NoError(t, err)
	assert.Equal(t, "PAUSED_FOR_HUMAN", state.Status)
	assert.Equal(t, "request_approval", state.CurrentStep)

	// Verify human task was created
	var taskID string
	var taskData json.RawMessage
	err = db.QueryRow(`
        SELECT id, task_data 
        FROM human_tasks 
        WHERE correlation_id = $1 AND status = 'PENDING'
    `, correlationID).Scan(&taskID, &taskData)
	require.NoError(t, err)

	var task map[string]interface{}
	json.Unmarshal(taskData, &task)
	assert.Equal(t, "proposal_review", task["approval_type"])
	assert.Contains(t, task, "data_to_review")

	// Simulate human approval
	humanResponse := map[string]interface{}{
		"approved":    true,
		"feedback":    "Looks good, please proceed",
		"approved_by": "test_manager",
		"approved_at": time.Now(),
	}

	err = completeHumanTask(db, taskID, correlationID, humanResponse)
	require.NoError(t, err)

	// Resume workflow
	err = coordinator.ResumeWorkflow(correlationID, humanResponse)
	require.NoError(t, err)

	// Wait for workflow to complete
	helpers.WaitForCondition(t, 30*time.Second, func() bool {
		var status string
		db.QueryRow(`
            SELECT status FROM orchestrator_state 
            WHERE correlation_id = $1
        `, correlationID).Scan(&status)
		return status == "COMPLETED"
	})

	// Verify final state
	var finalStatus string
	var collectedData json.RawMessage
	err = db.QueryRow(`
        SELECT status, collected_data
        FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&finalStatus, &collectedData)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", finalStatus)

	var data map[string]interface{}
	json.Unmarshal(collectedData, &data)
	assert.Contains(t, data, "human_response")
	assert.Contains(t, data, "execute_campaign")
}

func TestHumanInLoopTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	coordinator := setupTestCoordinator(t)
	db := getTestDB(t)

	// Workflow with short timeout for testing
	workflow := models.WorkflowPlan{
		StartStep: "generate",
		Steps: map[string]models.Step{
			"generate": {
				Action: "call_agent",
				Topic:  "system.agent.generic.process",
				Config: map[string]interface{}{
					"action": "generate_data",
				},
				NextStep: "approve",
			},
			"approve": {
				Action: "pause_for_human",
				Config: map[string]interface{}{
					"timeout":    "5s", // Very short timeout
					"on_timeout": "timeout_handler",
				},
				NextStep: "complete",
			},
			"timeout_handler": {
				Action: "call_agent",
				Topic:  "system.agent.generic.process",
				Config: map[string]interface{}{
					"action": "handle_timeout",
					"data": map[string]interface{}{
						"reason": "human_approval_timeout",
					},
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}

	correlationID := "test-human-timeout-" + uuid.New().String()
	headers := helpers.TestHeaders(correlationID)

	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
	require.NoError(t, err)

	// Wait for timeout
	time.Sleep(6 * time.Second)

	// Check that workflow proceeded with timeout handler
	var status, currentStep string
	err = db.QueryRow(`
        SELECT status, current_step 
        FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&status, &currentStep)

	require.NoError(t, err)
	assert.NotEqual(t, "PAUSED_FOR_HUMAN", status)

	// Verify timeout was handled
	var collectedData json.RawMessage
	err = db.QueryRow(`
        SELECT collected_data 
        FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&collectedData)

	require.NoError(t, err)
	var data map[string]interface{}
	json.Unmarshal(collectedData, &data)
	assert.Contains(t, data, "timeout_handler")
}

func TestHumanInLoopRejectionFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test")
	}

	coordinator := setupTestCoordinator(t)
	db := getTestDB(t)

	// Workflow that handles rejection
	workflow := createHumanApprovalWorkflow()

	correlationID := "test-human-rejection-" + uuid.New().String()
	headers := helpers.TestHeaders(correlationID)

	err := coordinator.ExecuteWorkflow(context.Background(), workflow, headers, nil)
	require.NoError(t, err)

	// Wait for pause
	helpers.WaitForCondition(t, 10*time.Second, func() bool {
		var status string
		db.QueryRow(`
            SELECT status FROM orchestrator_state 
            WHERE correlation_id = $1
        `, correlationID).Scan(&status)
		return status == "PAUSED_FOR_HUMAN"
	})

	// Get task
	var taskID string
	err = db.QueryRow(`
        SELECT id FROM human_tasks 
        WHERE correlation_id = $1 AND status = 'PENDING'
    `, correlationID).Scan(&taskID)
	require.NoError(t, err)

	// Simulate rejection with feedback
	humanResponse := map[string]interface{}{
		"approved":    false,
		"feedback":    "Budget too high, please reduce by 20%",
		"rejected_by": "test_manager",
		"rejected_at": time.Now(),
	}

	err = completeHumanTask(db, taskID, correlationID, humanResponse)
	require.NoError(t, err)

	// Resume workflow
	err = coordinator.ResumeWorkflow(correlationID, humanResponse)
	require.NoError(t, err)

	// Verify workflow went to revision step
	time.Sleep(2 * time.Second)

	var currentStep string
	err = db.QueryRow(`
        SELECT current_step 
        FROM orchestrator_state 
        WHERE correlation_id = $1
    `, correlationID).Scan(&currentStep)

	require.NoError(t, err)
	assert.Equal(t, "revise_proposal", currentStep)
}

// Helper functions
func completeHumanTask(db *sql.DB, taskID, correlationID string, response map[string]interface{}) error {
	responseJSON, _ := json.Marshal(response)

	_, err := db.Exec(`
        UPDATE human_tasks 
        SET status = 'COMPLETED',
            response_data = $1,
            completed_at = NOW()
        WHERE id = $2
    `, responseJSON, taskID)

	if err != nil {
		return err
	}

	// Update orchestrator state with human response
	_, err = db.Exec(`
        UPDATE orchestrator_state 
        SET collected_data = jsonb_set(
            collected_data,
            '{human_response}',
            $1::jsonb
        )
        WHERE correlation_id = $2
    `, responseJSON, correlationID)

	return err
}

func createHumanApprovalWorkflow() models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "prepare",
		Steps: map[string]models.Step{
			"prepare": {
				Action: "call_agent",
				Topic:  "system.agent.content-creator.process",
				Config: map[string]interface{}{
					"action": "prepare_document",
				},
				NextStep: "review",
			},
			"review": {
				Action: "pause_for_human",
				Config: map[string]interface{}{
					"review_type":   "document_approval",
					"reviewers":     []string{"manager", "director"},
					"min_approvals": 1,
					"timeout":       "48h",
				},
				NextStep: "decision",
			},
			"decision": {
				Action: "conditional",
				Config: map[string]interface{}{
					"condition":  "{{human_response.approved}} == true",
					"true_step":  "publish",
					"false_step": "revise",
				},
			},
			"revise": {
				Action: "call_agent",
				Topic:  "system.agent.content-creator.process",
				Config: map[string]interface{}{
					"action":   "revise_document",
					"feedback": "{{human_response.feedback}}",
				},
				NextStep: "review", // Loop back
			},
			"publish": {
				Action: "call_agent",
				Topic:  "system.agent.publisher.process",
				Config: map[string]interface{}{
					"action": "publish_document",
				},
				NextStep: "complete",
			},
			"complete": {
				Action: "complete_workflow",
			},
		},
	}
}
