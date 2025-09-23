// platform/orchestration/actions/workflow_actions.go

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// FILE: platform/orchestration/actions/workflow_actions.go
func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	current, caller := getFuncInfo(1)

	params.Logger.Info("In workflow_actions.go CompleteWorkflowAction",
		zap.String("function", current),
		zap.String("called_by workflow actions", caller),
		zap.String("container", os.Getenv("HOSTNAME")),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	)

	var callstack []string
	if params.Tracer != nil {
		callstack = params.Tracer.GetCallStack(12)
	}

	params.Logger.Info("Completing workflow CompleteWorkflowAction and GetCallStack",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("parent_orchestration_id", params.ExecutionContext.ParentOrchestrationID),
		zap.Any("collected_data_keys for CollectedData", getMapKeys(params.CollectedData)),
		zap.Strings("call stack", callstack),
	)

	var finalResult interface{}

	// If this agent's job was to run a "process" step (like the calculator),
	// its result is the only thing that matters.
	if processResult, ok := params.CollectedData["process"]; ok {
		finalResult = processResult
	} else if aggResult, ok := params.CollectedData["aggregate_results"]; ok {
		// If this is an orchestrator that just ran an aggregation, that is the result.
		finalResult = aggResult
	} else {
		// As a fallback, return all collected data, but clean out internal fields.
		filteredData := make(map[string]interface{})
		for key, value := range params.CollectedData {
			if key != "__raw_message__" && key != "__execution_context__" {
				filteredData[key] = value
			}
		}
		finalResult = filteredData
	}
	params.Logger.Info("CompleteWorkflowAction extracted result", zap.Any("result", finalResult))

	// Case 1: This is a CHILD orchestration completing its task.
	// It needs to notify its parent.
	if params.ExecutionContext.ParentOrchestrationID != "" {
		params.Logger.Info("Child orchestration needs to notify parent",
			zap.String("child_orch", params.ExecutionContext.OrchestrationID),
			zap.String("parent_orch", params.ExecutionContext.ParentOrchestrationID))

		// Get the original execution context stored when this child started
		if execCtxData, ok := params.CollectedData["__execution_context__"]; ok {
			var parentRequestID string
			var parentResponseTopic string
			var parentStepName string

			// Handle different types the context might be stored as
			switch ctx := execCtxData.(type) {
			case *types.ExecutionContext:
				parentRequestID = ctx.RequestID
				parentResponseTopic = ctx.ResponsesTopic
				parentStepName = ctx.StepName
			case map[string]interface{}:
				parentRequestID, _ = ctx["request_id"].(string)
				parentResponseTopic, _ = ctx["responses_topic"].(string)
				parentStepName, _ = ctx["step_name"].(string)
			}

			params.Tracer.TraceMessage(params.ExecutionContext, "EXTRACT_PARENT_TOPIC", "",
				map[string]interface{}{
					"extracted_response_topic": parentResponseTopic,
					"parent_request_id":        parentRequestID,
					"child_orch_id":            params.ExecutionContext.OrchestrationID,
					"parent_orch_id":           params.ExecutionContext.ParentOrchestrationID,
					"exec_ctx_type":            fmt.Sprintf("%T", execCtxData),
				})

			if parentRequestID == "" || parentResponseTopic == "" {
				params.Logger.Error("Cannot notify parent: parent request ID or response topic is missing from execution context",
					zap.String("parent_request_id", parentRequestID),
					zap.String("parent_response_topic", parentResponseTopic),
					zap.String("parent_step_name", parentStepName))

				params.Tracer.TraceMessage(params.ExecutionContext, "MISSING_PARENT_INFO", "",
					map[string]interface{}{
						"parent_request_id":     parentRequestID,
						"parent_response_topic": parentResponseTopic,
						"exec_ctx_data":         execCtxData,
					})
				return map[string]interface{}{"result": finalResult}, nil
			}

			params.Logger.Info("Sending completion response to parent",
				zap.String("parent_request_id", parentRequestID),
				zap.String("response_topic", parentResponseTopic))

			params.Tracer.TraceMessage(params.ExecutionContext, "SEND_TO_PARENT", parentResponseTopic,
				map[string]interface{}{
					"topic":      parentResponseTopic,
					"request_id": parentRequestID,
					"from_child": params.ExecutionContext.OrchestrationID,
					"to_parent":  params.ExecutionContext.ParentOrchestrationID,
				})

			// Build SIMPLE response message
			responseMsg := types.ResponseMessage{
				Headers: types.ResponseHeaders{
					Sender: types.AgentIdentity{
						AgentType:    params.ExecutionContext.Sender.AgentType,
						AgentID:      params.ExecutionContext.Sender.AgentID,
						PodName:      params.ExecutionContext.Sender.PodName,
						AgentVersion: params.ExecutionContext.Sender.AgentVersion,
					},
					OrchestrationID:            params.ExecutionContext.OrchestrationID,
					OrchestrationName:          params.ExecutionContext.OrchestrationName,
					InResponseToRequestID:      parentRequestID,
					InResponseToStepID:         params.ExecutionContext.StepID,
					InResponseToStepName:       parentStepName,
					InResponseToParentOrchID:   params.ExecutionContext.ParentOrchestrationID,
					InResponseToParentOrchName: params.ExecutionContext.ParentOrchestrationName,
					InResponseToMessageID:      params.ExecutionContext.MessageID,
					InResponseToAction:         params.ExecutionContext.Action,
					RetryCount:                 params.ExecutionContext.RetryVersion,
					MyOrchestrationID:          params.ExecutionContext.OrchestrationID,
					MyOrchestrationName:        params.ExecutionContext.OrchestrationName,
					CorrelationID:              params.ExecutionContext.CorrelationID,
					CorrelationName:            params.ExecutionContext.CorrelationName,
					ClientID:                   params.ExecutionContext.ClientID,
					ToAgent:                    params.ExecutionContext.FromAgentID,
					ToAgentType:                params.ExecutionContext.FromAgentType,
					MessageType:                "response",
					Status:                     "complete",
					IsComplete:                 true,
					IsError:                    false,
					TimeSent:                   time.Now(),
				},
				Body: types.ResponseBody{
					Success: true,
					Headers: nil,
					Body:    finalResult,
					Error:   nil,
				},
			}

			responseBytes, err := json.Marshal(responseMsg)
			if err != nil {
				params.Logger.Error("Failed to marshal response", zap.Error(err))
				return finalResult, fmt.Errorf("failed to marshal response: %w", err)
			}

			headers := responseMsg.Headers.ToMap()
			key := []byte(params.ExecutionContext.CorrelationID)

			err = params.Producer.Produce(ctx, parentResponseTopic, headers, key, responseBytes)
			if err != nil {
				params.Logger.Error("Failed to send response to parent",
					zap.Error(err),
					zap.String("topic", parentResponseTopic))
				return finalResult, fmt.Errorf("failed to send response: %w", err)
			}

			params.Logger.Info("Successfully sent response to parent orchestration",
				zap.String("topic", parentResponseTopic),
				zap.String("request_id", parentRequestID),
				zap.Any("result_sent", finalResult))
		}
		// Case 2: This is a ROOT orchestration completing its entire workflow.
		// It needs to notify the original client.
	} else {
		params.Logger.Info("Root orchestration completed. Sending final response to client.")

		originalResponseTopic := params.ExecutionContext.ResponsesTopic
		originalRequestID := params.ExecutionContext.RequestID
		if originalRequestID == "" {
			params.Logger.Warn("No request_id specified in initial request. Cannot send final client request id.")
			return map[string]interface{}{"result": finalResult}, nil
		}

		// If not there, check stored initial context
		if originalResponseTopic == "" {
			if initCtx, ok := params.CollectedData["__execution_context__"]; ok {
				switch ctx := initCtx.(type) {
				case *types.ExecutionContext:
					originalResponseTopic = ctx.ResponsesTopic
					originalRequestID = ctx.RequestID
				case map[string]interface{}:
					originalResponseTopic, _ = ctx["responses_topic"].(string)
					originalRequestID, _ = ctx["request_id"].(string)
				}
			}
		}

		if originalResponseTopic == "" {
			params.Logger.Warn("No responses_topic found for root orchestration")
			return map[string]interface{}{"result": finalResult}, nil
		}

		responseMsg := types.ResponseMessage{
			Headers: types.ResponseHeaders{
				Sender: types.AgentIdentity{
					AgentType:    params.ExecutionContext.Sender.AgentType,
					AgentID:      params.ExecutionContext.Sender.AgentID,
					PodName:      params.ExecutionContext.Sender.PodName,
					AgentVersion: params.ExecutionContext.Sender.AgentVersion,
				},
				OrchestrationID:            params.ExecutionContext.OrchestrationID,
				OrchestrationName:          params.ExecutionContext.OrchestrationName,
				InResponseToRequestID:      params.ExecutionContext.RequestID, // Use original request ID
				InResponseToStepID:         params.ExecutionContext.StepID,
				InResponseToStepName:       params.ExecutionContext.StepName,
				InResponseToParentOrchID:   params.ExecutionContext.ParentOrchestrationID,
				InResponseToParentOrchName: params.ExecutionContext.ParentOrchestrationName,
				InResponseToMessageID:      params.ExecutionContext.MessageID,
				InResponseToAction:         params.ExecutionContext.Action,
				RetryCount:                 params.ExecutionContext.RetryVersion,
				MyOrchestrationID:          params.ExecutionContext.OrchestrationID,
				MyOrchestrationName:        params.ExecutionContext.OrchestrationName,
				CorrelationID:              params.ExecutionContext.CorrelationID,
				CorrelationName:            params.ExecutionContext.CorrelationName,
				ClientID:                   params.ExecutionContext.ClientID,
				ToAgent:                    params.ExecutionContext.FromAgentID,
				ToAgentType:                params.ExecutionContext.FromAgentType,
				MessageType:                "response",
				Status:                     "complete",
				IsComplete:                 true,
				IsError:                    false,
				TimeSent:                   time.Now(),
			},
			Body: types.ResponseBody{
				Success: true,
				Headers: nil,
				Body:    map[string]interface{}{"result": finalResult},
				Error:   nil,
			},
		}

		responseBytes, err := json.Marshal(responseMsg)
		if err != nil {
			params.Logger.Error("Failed to marshal response", zap.Error(err))
			return finalResult, fmt.Errorf("failed to marshal response: %w", err)
		}

		headers := responseMsg.Headers.ToMap()
		key := []byte(params.ExecutionContext.CorrelationID)

		// Note: The original code had a variable scope issue here, using 'parentResponseTopic'.
		// This has been corrected to use 'responseTopic' which is defined in this 'else' block.
		err = params.Producer.Produce(ctx, originalResponseTopic, headers, key, responseBytes)
		if err != nil {
			params.Logger.Error("Failed to send response to client",
				zap.Error(err),
				zap.String("topic", originalResponseTopic))
			return finalResult, fmt.Errorf("failed to send response: %w", err)
		}

		params.Logger.Info("Successfully sent response to client",
			zap.String("topic", originalResponseTopic),
			zap.String("request_id", params.ExecutionContext.RequestID))
	}

	return map[string]interface{}{"result": finalResult}, nil
}

func CompleteWorkflowActionold2(ctx context.Context, params ActionParams) (interface{}, error) {
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
