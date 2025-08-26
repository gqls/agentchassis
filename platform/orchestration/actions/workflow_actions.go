// platform/orchestration/actions/workflow_actions.go - NEW FILE

package actions

import (
	"context"
	"time"
)

func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow")

	// Create a clean copy to avoid circular references
	collectedDataCopy := make(map[string]interface{})
	for k, v := range params.CollectedData {
		// Skip the current step and 'complete' to avoid self-reference
		if k != params.CurrentStep && k != "complete" {
			collectedDataCopy[k] = v
		}
	}

	// Just return the completion data
	// Let the coordinator handle parent notification
	return map[string]interface{}{
		"status":         "completed",
		"collected_data": collectedDataCopy,
		"timestamp":      time.Now().UTC(),
	}, nil
}
