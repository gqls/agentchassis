// platform/orchestration/actions/workflow_actions.go - NEW FILE

package actions

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
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

	// Prepare completion data
	completionData := map[string]interface{}{
		"status":         "completed",
		"collected_data": collectedDataCopy,
		"timestamp":      time.Now().UTC(),
	}

	// Check if this is a child orchestration
	if parentCorrelationID := params.Headers["parent_correlation_id"]; parentCorrelationID != "" {
		params.Logger.Info("Notifying parent orchestration",
			zap.String("parent_correlation_id", parentCorrelationID))

		// Send completion to parent
		parentResponse := models.TaskResponse{
			Success: true,
			Data:    completionData,
		}

		responseBytes, _ := json.Marshal(parentResponse)

		err := params.Producer.Produce(ctx,
			"system.generic.responses",
			map[string]string{
				"correlation_id": parentCorrelationID,
				"causation_id":   params.Headers["correlation_id"],
			},
			[]byte(parentCorrelationID),
			responseBytes)

		if err != nil {
			params.Logger.Error("Failed to notify parent", zap.Error(err))
		}
	}

	return completionData, nil
}
