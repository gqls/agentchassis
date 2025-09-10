// platform/orchestration/actions/workflow_actions.go

package actions

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Get execution context
	execCtx, err := types.FromHeaders(params.Headers)
	if err != nil {
		params.Logger.Error("Failed to get execution context", zap.Error(err))
		return nil, err
	}

	params.Logger.Info("Completing workflow",
		zap.String("orchestration_id", execCtx.OrchestrationID))

	// Clean collected data
	result := make(map[string]interface{})
	for k, v := range params.CollectedData {
		if !strings.HasPrefix(k, "__") && !strings.HasSuffix(k, "_request") {
			result[k] = v
		}
	}

	// Check if this is a child needing to respond to parent
	if execCtx.IsChildOrchestration() {
		// Look for stored parent context
		if parentCtxData, ok := params.CollectedData["__execution_context__"]; ok {
			if _, ok := parentCtxData.(*types.ExecutionContext); ok {
				// Create response context
				responseCtx := execCtx.CreateResponseContext("completed", 0)

				// Build response
				response := models.TaskResponse{
					Success: true,
					Data: map[string]interface{}{
						"final_result":     result,
						"status":           "completed",
						"orchestration_id": execCtx.OrchestrationID,
					},
				}

				responseBytes, _ := json.Marshal(response)

				// Send to parent
				err := params.Producer.Produce(ctx,
					responseCtx.ReplyToTopic,
					responseCtx.ToHeaders(),
					[]byte(responseCtx.CorrelationID),
					responseBytes)

				if err != nil {
					params.Logger.Error("Failed to send response to parent",
						zap.Error(err),
						zap.String("parent_orchestration_id", responseCtx.OrchestrationID))
				} else {
					params.Logger.Info("Sent completion to parent",
						zap.String("parent_orchestration_id", responseCtx.OrchestrationID))
				}
			}
		}
	}

	return map[string]interface{}{
		"status":    "completed",
		"result":    result,
		"timestamp": time.Now().UTC(),
	}, nil
}
