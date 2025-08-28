// platform/orchestration/actions/workflow_actions.go

package actions

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow")

	// Clean collected data - remove internal keys
	result := make(map[string]interface{})
	for k, v := range params.CollectedData {
		// Skip internal keys
		if k == "__parent_context__" ||
			strings.HasSuffix(k, "_started") ||
			strings.HasSuffix(k, "_request") {
			continue
		}
		result[k] = v
	}

	// If this is a child orchestration, send response to parent
	if parentCtx, ok := params.CollectedData["__parent_context__"].(map[string]interface{}); ok {
		parentOrchestratorID, _ := parentCtx["orchestration_id"].(string)
		replyToTopic, _ := parentCtx["reply_to_topic"].(string)
		requestID, _ := parentCtx["request_id"].(string)

		if parentOrchestratorID != "" && replyToTopic != "" && requestID != "" {
			response := models.TaskResponse{
				Success: true,
				Data: map[string]interface{}{
					"final_result":     result,
					"status":           "completed",
					"orchestration_id": params.Headers["orchestration_id"],
				},
			}

			responseHeaders := map[string]string{
				"orchestration_id": parentOrchestratorID, // Route to parent
				"in_response_to":   requestID,            // What we're responding to
				"correlation_id":   params.Headers["correlation_id"],
				"message_type":     "response",
				"from_agent_id":    params.Headers["agent_id"],
				"causation_id":     requestID,
			}

			responseBytes, _ := json.Marshal(response)
			err := params.Producer.Produce(ctx, replyToTopic, responseHeaders,
				[]byte(params.Headers["orchestration_id"]), responseBytes)

			if err != nil {
				params.Logger.Error("Failed to send response to parent",
					zap.Error(err),
					zap.String("parent_orchestration_id", parentOrchestratorID),
					zap.String("reply_to_topic", replyToTopic),
					zap.String("in_response_to", requestID))
			} else {
				params.Logger.Info("Sent completion to parent",
					zap.String("parent_orchestration_id", parentOrchestratorID),
					zap.String("in_response_to", requestID),
					zap.String("reply_to_topic", replyToTopic))
			}
		}
	}

	return map[string]interface{}{
		"status":    "completed",
		"result":    result,
		"timestamp": time.Now().UTC(),
	}, nil
}
