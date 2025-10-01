// platform/orchestration/actions/workflow_actions.go

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("parent_orchestration_id", params.ExecutionContext.ParentOrchestrationID))

	// Extract final result
	var finalResult interface{}
	if processResult, ok := params.CollectedData["process"]; ok {
		finalResult = processResult
	} else if aggResult, ok := params.CollectedData["aggregate_results"]; ok {
		finalResult = aggResult
	} else {
		// Return filtered collected data
		filteredData := make(map[string]interface{})
		for key, value := range params.CollectedData {
			if !strings.HasPrefix(key, "__") {
				filteredData[key] = value
			}
		}
		finalResult = filteredData
	}

	// Case 1: Child orchestration completing
	if params.ExecutionContext.ParentOrchestrationID != "" {
		params.Logger.Info("Child orchestration completing")

		// Get parent's response topic from stored context
		var parentResponsesTopic string
		var parentRequestID string

		// Check stored parent context
		if parentTopic, ok := params.CollectedData["__parent_responses_topic__"]; ok {
			parentResponsesTopic, _ = parentTopic.(string)
		}

		// Get parent request ID from execution context
		if execCtx, ok := params.CollectedData["__execution_context__"]; ok {
			switch ctx := execCtx.(type) {
			case *types.ExecutionContext:
				if parentRequestID == "" {
					parentRequestID = ctx.RequestID
				}
				if parentResponsesTopic == "" {
					parentResponsesTopic = ctx.ResponsesTopic
				}
			case map[string]interface{}:
				if parentRequestID == "" {
					parentRequestID, _ = ctx["request_id"].(string)
				}
				if parentResponsesTopic == "" {
					parentResponsesTopic, _ = ctx["responses_topic"].(string)
				}
			}
		}

		if parentRequestID == "" || parentResponsesTopic == "" {
			params.Logger.Error("Missing parent info",
				zap.String("request_id", parentRequestID),
				zap.String("responses_topic", parentResponsesTopic))
			return map[string]interface{}{"result": finalResult}, nil
		}

		// Send response to parent
		responseMsg := buildResponseMessage(params, parentRequestID, finalResult)

		responseBytes, err := json.Marshal(responseMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		headers := responseMsg.Headers.ToMap()
		key := []byte(params.ExecutionContext.CorrelationID)

		err = params.Producer.Produce(ctx, parentResponsesTopic, headers, key, responseBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to send response: %w", err)
		}

		params.Logger.Info("Sent response to parent",
			zap.String("topic", parentResponsesTopic),
			zap.String("request_id", parentRequestID))

	} else {
		// Case 2: Root orchestration completing
		params.Logger.Info("Root orchestration completing")

		var originalResponsesTopic string
		var originalRequestID string

		// Check for stored original request
		if origReq, ok := params.CollectedData["__original_request__"]; ok {
			if reqMap, ok := origReq.(map[string]interface{}); ok {
				originalResponsesTopic, _ = reqMap["responses_topic"].(string)
				originalRequestID, _ = reqMap["request_id"].(string)
			}
		}

		// Fallback to current context
		if originalResponsesTopic == "" {
			originalResponsesTopic = params.ExecutionContext.ResponsesTopic
		}
		if originalRequestID == "" {
			originalRequestID = params.ExecutionContext.RequestID
		}

		if originalResponsesTopic == "" || originalRequestID == "" {
			params.Logger.Warn("No response destination",
				zap.String("responses_topic", originalResponsesTopic),
				zap.String("request_id", originalRequestID))
			return map[string]interface{}{"result": finalResult}, nil
		}

		// Send final response
		responseMsg := buildResponseMessage(params, originalRequestID, finalResult)

		responseBytes, err := json.Marshal(responseMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		headers := responseMsg.Headers.ToMap()
		key := []byte(params.ExecutionContext.CorrelationID)

		err = params.Producer.Produce(ctx, originalResponsesTopic, headers, key, responseBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to send response: %w", err)
		}

		params.Logger.Info("Sent final response",
			zap.String("topic", originalResponsesTopic),
			zap.String("request_id", originalRequestID))
	}

	return map[string]interface{}{"result": finalResult}, nil
}

// Helper to build response message
func buildResponseMessage(params ActionParams, requestID string, result interface{}) types.ResponseMessage {
	return types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender:                     params.ExecutionContext.Sender,
			OrchestrationID:            params.ExecutionContext.OrchestrationID,
			OrchestrationName:          params.ExecutionContext.OrchestrationName,
			InResponseToRequestID:      requestID,
			InResponseToStepID:         params.ExecutionContext.StepID,
			InResponseToStepName:       params.ExecutionContext.StepName,
			InResponseToParentOrchID:   params.ExecutionContext.ParentOrchestrationID,
			InResponseToParentOrchName: params.ExecutionContext.ParentOrchestrationName,
			MyOrchestrationID:          params.ExecutionContext.OrchestrationID,
			MyOrchestrationName:        params.ExecutionContext.OrchestrationName,
			CorrelationID:              params.ExecutionContext.CorrelationID,
			ClientID:                   params.ExecutionContext.ClientID,
			MessageType:                "response",
			Status:                     "complete",
			IsComplete:                 true,
			IsError:                    false,
			TimeSent:                   time.Now(),
		},
		Body: types.ResponseBody{
			Success: true,
			Body:    result,
			Error:   nil,
		},
	}
}
