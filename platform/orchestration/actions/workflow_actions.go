// platform/orchestration/actions/workflow_actions.go

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow CompleteWorkflowAction",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("parent_orchestration_id", params.ExecutionContext.ParentOrchestrationID),
		zap.Any("collected_data_keys", getMapKeys(params.CollectedData)))

	// Find the actual calculation/process result
	var actualResult interface{}

	// Look for "process" step result (calculator's actual work)
	if processData, ok := params.CollectedData["process"].(map[string]interface{}); ok {
		actualResult = processData
	} else {
		// Fallback - return everything except internal fields
		filteredData := make(map[string]interface{})
		for key, value := range params.CollectedData {
			if key != "__raw_message__" && key != "__execution_context__" {
				filteredData[key] = value
			}
		}
		actualResult = filteredData
	}

	params.Logger.Info("CompleteWorkflowAction extracted result",
		zap.Any("actual_result", actualResult))

	// If this is a child orchestration, we need to send response to parent
	if params.ExecutionContext.ParentOrchestrationID != "" {
		params.Logger.Info("Child orchestration needs to notify parent",
			zap.String("child_orch", params.ExecutionContext.OrchestrationID),
			zap.String("parent_orch", params.ExecutionContext.ParentOrchestrationID))

		// Get the original execution context stored when this child started
		if execCtxData, ok := params.CollectedData["__execution_context__"]; ok {
			var parentRequestID string
			var parentResponseTopic string

			// Handle different types the context might be stored as
			switch ctx := execCtxData.(type) {
			case *types.ExecutionContext:
				parentRequestID = ctx.RequestID
				parentResponseTopic = ctx.ResponsesTopic
			case map[string]interface{}:
				parentRequestID, _ = ctx["request_id"].(string)
				parentResponseTopic, _ = ctx["responses_topic"].(string)
			}

			if parentRequestID != "" && parentResponseTopic != "" {
				params.Logger.Info("Sending completion response to parent",
					zap.String("parent_request_id", parentRequestID),
					zap.String("response_topic", parentResponseTopic))

				// Build SIMPLE response message
				responseMsg := types.ResponseMessage{
					Headers: types.ResponseHeaders{
						Sender: types.AgentIdentity{
							AgentType:    params.ExecutionContext.Sender.AgentType,
							AgentID:      params.ExecutionContext.Sender.AgentID,
							PodName:      params.ExecutionContext.Sender.PodName,
							AgentVersion: params.ExecutionContext.Sender.AgentVersion,
						},

						OrchestrationID:   params.ExecutionContext.OrchestrationID,
						OrchestrationName: params.ExecutionContext.OrchestrationName,

						// Response tracking
						InResponseToRequestID:      parentRequestID,
						InResponseToStepID:         params.ExecutionContext.StepID,
						InResponseToStepName:       params.ExecutionContext.StepName,
						InResponseToParentOrchID:   params.ExecutionContext.ParentOrchestrationID,
						InResponseToParentOrchName: params.ExecutionContext.ParentOrchestrationName,
						InResponseToMessageID:      params.ExecutionContext.MessageID,
						InResponseToAction:         params.ExecutionContext.Action,
						RetryCount:                 params.ExecutionContext.RetryVersion,

						// Context
						MyOrchestrationID:   params.ExecutionContext.OrchestrationID,
						MyOrchestrationName: params.ExecutionContext.OrchestrationName,
						CorrelationID:       params.ExecutionContext.CorrelationID,
						CorrelationName:     params.ExecutionContext.CorrelationName,
						ClientID:            params.ExecutionContext.ClientID,

						ToAgent:     params.ExecutionContext.FromAgentID,
						ToAgentType: params.ExecutionContext.FromAgentType,

						// Status
						MessageType: "response",
						Status:      "complete",
						IsComplete:  true,
						IsError:     false,

						// Timing
						TimeSent: time.Now(),
					},
					Body: types.ResponseBody{
						Success: true,
						Headers: nil,
						Body:    actualResult, // Direct assignment - no wrapping!
						Error:   nil,
					},
				}

				// Send the response
				responseBytes, err := json.Marshal(responseMsg)
				if err != nil {
					params.Logger.Error("Failed to marshal response", zap.Error(err))
					return actualResult, fmt.Errorf("failed to marshal response: %w", err)
				}

				headers := responseMsg.Headers.ToMap()
				key := []byte(params.ExecutionContext.CorrelationID)

				err = params.Producer.Produce(ctx, parentResponseTopic, headers, key, responseBytes)
				if err != nil {
					params.Logger.Error("Failed to send response to parent",
						zap.Error(err),
						zap.String("topic", parentResponseTopic))
					return actualResult, fmt.Errorf("failed to send response: %w", err)
				}

				params.Logger.Info("Successfully sent response to parent orchestration",
					zap.String("topic", parentResponseTopic),
					zap.String("request_id", parentRequestID))
			}
		}
	} else {
		params.Logger.Info("Root orchestration completed - no parent to notify")
	}

	return actualResult, nil
}

func CompleteWorkflowActionold(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow CompleteWorkflowAction",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("parent_orchestration_id", params.ExecutionContext.ParentOrchestrationID),
		zap.Any("collected_data_keys", getMapKeys(params.CollectedData)))

	// Filter out internal/system fields that might have cycles
	filteredData := make(map[string]interface{})
	for key, value := range params.CollectedData {
		// Skip internal fields that can cause cycles
		if key == "__raw_message__" || key == "__execution_context__" {
			continue
		}
		filteredData[key] = value
	}

	// Prepare the result with filtered data
	result := map[string]interface{}{
		"status":    "completed",
		"result":    filteredData, // Use filtered data instead
		"timestamp": time.Now(),
	}

	params.Logger.Info("CompleteWorkflowAction Filtered data workflow_actions",
		zap.Any("filtered data result", result),
	)

	// If this is a child orchestration, we need to send response to parent
	if params.ExecutionContext.ParentOrchestrationID != "" {
		params.Logger.Info("Child orchestration needs to notify parent",
			zap.String("child_orch", params.ExecutionContext.OrchestrationID),
			zap.String("parent_orch", params.ExecutionContext.ParentOrchestrationID))

		// Find the actual calculation/process result
		var actualResult interface{}

		// Look for "process" step result (calculator's actual work)
		if processData, ok := params.CollectedData["process"].(map[string]interface{}); ok {
			actualResult = processData
		} else {
			// Fallback - return everything except internal fields
			filteredData := make(map[string]interface{})
			for key, value := range params.CollectedData {
				if key != "__raw_message__" && key != "__execution_context__" {
					filteredData[key] = value
				}
			}
			actualResult = filteredData
		}

		// Get the original execution context stored when this child started
		if execCtxData, ok := params.CollectedData["__execution_context__"]; ok {
			var parentRequestID string
			var parentResponseTopic string

			// Handle different types the context might be stored as
			switch ctx := execCtxData.(type) {
			case *types.ExecutionContext:
				parentRequestID = ctx.RequestID
				parentResponseTopic = ctx.ResponsesTopic
			case map[string]interface{}:
				parentRequestID, _ = ctx["request_id"].(string)
				parentResponseTopic, _ = ctx["responses_topic"].(string)
			}

			if parentRequestID != "" && parentResponseTopic != "" {
				params.Logger.Info("Sending completion response to parent",
					zap.String("parent_request_id", parentRequestID),
					zap.String("response_topic", parentResponseTopic))

				// Build response message
				responseMsg := types.ResponseMessage{
					Headers: types.ResponseHeaders{
						Sender: types.AgentIdentity{
							AgentType:    params.ExecutionContext.Sender.AgentType,
							AgentID:      params.ExecutionContext.Sender.AgentID,
							PodName:      params.ExecutionContext.Sender.PodName,
							AgentVersion: params.ExecutionContext.Sender.AgentVersion,
						},

						OrchestrationID:   params.ExecutionContext.OrchestrationID,
						OrchestrationName: params.ExecutionContext.OrchestrationName,

						// Response tracking
						InResponseToRequestID:      parentRequestID,
						InResponseToStepID:         params.ExecutionContext.StepID,
						InResponseToStepName:       params.ExecutionContext.StepName,
						InResponseToParentOrchID:   params.ExecutionContext.ParentOrchestrationID,
						InResponseToParentOrchName: params.ExecutionContext.ParentOrchestrationName,
						InResponseToMessageID:      params.ExecutionContext.MessageID,
						InResponseToAction:         params.ExecutionContext.Action,
						RetryCount:                 params.ExecutionContext.RetryVersion,

						// Context
						MyOrchestrationID:   params.ExecutionContext.OrchestrationID,
						MyOrchestrationName: params.ExecutionContext.OrchestrationName,
						CorrelationID:       params.ExecutionContext.CorrelationID,
						CorrelationName:     params.ExecutionContext.CorrelationName,
						ClientID:            params.ExecutionContext.ClientID,

						ToAgent:     params.ExecutionContext.FromAgentID,
						ToAgentType: params.ExecutionContext.FromAgentType,

						// Status
						MessageType: "response",
						Status:      "complete",
						IsComplete:  true,
						IsError:     false,

						// Timing
						TimeSent: time.Now(),
					},
					Body: types.ResponseBody{
						Success: true,
						Headers: nil,
						Body: struct {
							Result      interface{}      `json:"result"`
							Calculation interface{}      `json:"calculation,omitempty"`
							Error       *types.ErrorInfo `json:"error,omitempty"`
						}{
							Result: actualResult,
							Error:  nil,
						},
					},
				}

				// Send the response
				responseBytes, err := json.Marshal(responseMsg)
				if err != nil {
					params.Logger.Error("Failed to marshal response", zap.Error(err))
					return result, fmt.Errorf("failed to marshal response: %w", err)
				}

				headers := responseMsg.Headers.ToMap()
				key := []byte(params.ExecutionContext.CorrelationID)

				err = params.Producer.Produce(ctx, parentResponseTopic, headers, key, responseBytes)
				if err != nil {
					params.Logger.Error("Failed to send response to parent",
						zap.Error(err),
						zap.String("topic", parentResponseTopic))
					return result, fmt.Errorf("failed to send response: %w", err)
				}

				params.Logger.Info("Successfully sent response to parent orchestration",
					zap.String("topic", parentResponseTopic),
					zap.String("request_id", parentRequestID),
					zap.Any("responseMsg", responseMsg),
				)
			}
		}
	} else {
		params.Logger.Info("Root orchestration completed - no parent to notify")
		// Root orchestration - the processor will handle sending final response
	}

	return result, nil
}
