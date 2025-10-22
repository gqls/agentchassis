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

// ReplyToMetadata contains everything needed to send a response back to the requester
type ReplyToMetadata struct {
	RequestID string // The request_id we're responding to
	Topic     string // Where to send the response
	Requester string // Who asked us (for logging)
	StepID    string // Which step in the requester's workflow
	StepName  string // Name of that step
}

func CompleteWorkflowAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompleteWorkflowAction: starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.Any("DEBUGaa: input params for CompleteWorkflowAction", params),
	)

	// Step 1: Extract the final result from CollectedData
	finalResult := extractFinalResult(params.CollectedData, params.Logger)
	params.Logger.Info("CompleteWorkflowAction",
		zap.Any("finalResult", finalResult),
	)

	// Step 2: Get reply-to metadata (WHO asked us to do work? WHERE do they expect the response?)
	replyTo, err := extractReplyToMetadata(params.CollectedData, params.ExecutionContext, params.Logger)
	params.Logger.Info("CompleteWorkflowAction ReplytoMetadata",
		zap.Any("replyTo", replyTo),
	)
	if err != nil {
		params.Logger.Error("CompleteWorkflowAction: cannot determine where to send response",
			zap.Error(err),
			zap.String("orchestration_id", params.ExecutionContext.OrchestrationID))
		return map[string]interface{}{"result": finalResult}, err
	}

	params.Logger.Info("CompleteWorkflowAction: sending response",
		zap.String("reply_to_request_id", replyTo.RequestID),
		zap.String("reply_to_topic", replyTo.Topic), // blank
		zap.String("requester", replyTo.Requester))

	// Step 3: Build and send the response
	responseMsg := buildResponseMessage(params.ExecutionContext, replyTo, finalResult)

	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	headers := responseMsg.Headers.ToMap()
	key := []byte(params.ExecutionContext.CorrelationID)

	err = params.Producer.Produce(ctx, replyTo.Topic, headers, key, responseBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send response: %w", err)
	}

	params.Logger.Info("CompleteWorkflowAction: response sent successfully",
		zap.String("topic", replyTo.Topic),
		zap.String("request_id", replyTo.RequestID))

	return map[string]interface{}{"result": finalResult}, nil
}

// extractFinalResult gets the workflow result from CollectedData
func extractFinalResult(collectedData map[string]interface{}, logger *zap.Logger) interface{} {
	// Try common result locations
	if processResult, ok := collectedData["process"]; ok {
		return processResult
	}
	if aggResult, ok := collectedData["aggregate_results"]; ok {
		return aggResult
	}

	// Return all non-system data
	filteredData := make(map[string]interface{})
	for key, value := range collectedData {
		if !strings.HasPrefix(key, "__") && key != "agent_config" {
			filteredData[key] = value
		}
	}

	if len(filteredData) == 0 {
		logger.Warn("CompleteWorkflowAction: no result data found in CollectedData")
		return map[string]interface{}{"message": "workflow completed"}
	}

	return filteredData
}

// extractReplyToMetadata finds where to send the response
// Priority order:
//  1. __work_request__ (stored when work request received) - PREFERRED
//  2. __execution_context__ (from the work request)
//  3. Current ExecutionContext (for inline workflows)
func extractReplyToMetadata(collectedData map[string]interface{}, execCtx *types.ExecutionContext, logger *zap.Logger) (*ReplyToMetadata, error) {

	// Priority 1: Check for explicitly stored work request metadata
	if workReqData, ok := collectedData["__work_request__"].(map[string]interface{}); ok {
		logger.Info("CompleteWorkflowAction: using __work_request__ metadata")
		return &ReplyToMetadata{
			RequestID: getStringField(workReqData, "request_id"),
			Topic:     getStringField(workReqData, "parent_responses_topic"),
			Requester: getStringField(workReqData, "requester_agent_id"),
			StepID:    getStringField(workReqData, "step_id"),
			StepName:  getStringField(workReqData, "step_name"),
		}, nil
	}

	// Priority 2: Extract from stored execution context
	if execCtxData, ok := collectedData["__execution_context__"]; ok {
		logger.Info("CompleteWorkflowAction: extracting from __execution_context__",
			zap.Any("DEBUGaa: exec_ctx", execCtxData),
		)

		var storedExecCtx *types.ExecutionContext
		switch v := execCtxData.(type) {
		case *types.ExecutionContext:
			storedExecCtx = v
		case map[string]interface{}:
			storedExecCtx = mapToExecutionContext(v, logger)
		default:
			logger.Warn("CompleteWorkflowAction: unexpected __execution_context__ type",
				zap.String("type", fmt.Sprintf("%T", v)))
		}

		if storedExecCtx != nil && storedExecCtx.RequestID != "" {
			// Get parent responses topic from CollectedData or environment
			parentTopic := getStringField(collectedData, "__parent_responses_topic__")
			if parentTopic == "" {
				parentTopic = storedExecCtx.ReplyToTopic
			}

			return &ReplyToMetadata{
				RequestID: storedExecCtx.RequestID, // The work request we received
				Topic:     parentTopic,             // Where parent is listening
				Requester: storedExecCtx.Sender.AgentID,
				StepID:    storedExecCtx.StepID,
				StepName:  storedExecCtx.StepName,
			}, nil
		}
	}

	// Priority 3: Use current execution context (for simple/inline cases)
	if execCtx.RequestID != "" && execCtx.ReplyToTopic != "" {
		logger.Debug("CompleteWorkflowAction: using current ExecutionContext")
		return &ReplyToMetadata{
			RequestID: execCtx.RequestID,
			Topic:     execCtx.ReplyToTopic,
			Requester: execCtx.Sender.AgentID,
			StepID:    execCtx.StepID,
			StepName:  execCtx.StepName,
		}, nil
	}

	// If we get here, we don't have enough information
	return nil, fmt.Errorf("cannot determine reply-to metadata: no work request info found in CollectedData or ExecutionContext")
}

/*func CompleteWorkflowActionOld(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("parent_orchestration_id", params.ExecutionContext.ParentOrchestrationID),
		zap.Any("DEBUGaa: CompleteWorkflowAction params - look for _execution_context__ look for request id and parent request id", params),
	)

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

	var parentResponsesTopic string
	parentResponsesTopic = os.Getenv("PARENT_RESPONSES_TOPIC")

	// Case 1: Child orchestration completing
	if params.ExecutionContext.ParentOrchestrationID != "" {
		params.Logger.Info("Child orchestration completing which agent am I on?",
			zap.String("I am on agent", os.Getenv("AGENT_TYPE")),
		)
		var replyToRequestID string
		replyToRequestID = params.ExecutionContext.ReplyToRequestID
		// Get reply to request ID from execution context that was in collectedData
		if replyToRequestID == "" {
			if collDataExecCtx, ok := params.CollectedData["__execution_context__"]; ok {
				switch storedExecCtx := collDataExecCtx.(type) {
				case *types.ExecutionContext:
					replyToRequestID = storedExecCtx.ReplyToRequestID
				case map[string]interface{}:
					replyToRequestID, _ = storedExecCtx["reply_to_request_id"].(string)
				}
			}
		}

		// 3. If still empty, check CollectedData directly for stored reply to request ID
		if replyToRequestID == "" {
			if storedReplyToRequestID, ok := params.CollectedData["__reply_to_request_id__"].(string); ok {
				replyToRequestID = storedReplyToRequestID
			}
		}

		if replyToRequestID == "" {
			params.Logger.Error("Missing reply to request ID",
				zap.String("replyToRequestID", replyToRequestID),
			)
			return map[string]interface{}{"result": finalResult}, nil
		}

		// Send response to parent
		responseMsg := buildResponseMessage(params, replyToRequestID, finalResult)

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

		params.Logger.Info("Sent response to parent in CompleteWorkflowAction",
			zap.String("topic", parentResponsesTopic),
			zap.String("DEBUGaa: reply to request_id", replyToRequestID))

	} else {
		// Case 2: Root orchestration completing
		params.Logger.Info("Root orchestration completing")

		var originalRequestID string

		// Check for stored original request
		if origReq, ok := params.CollectedData["__original_request__"]; ok {
			if reqMap, ok := origReq.(map[string]interface{}); ok {
				originalRequestID, _ = reqMap["request_id"].(string)
			}
		}

		if originalRequestID == "" {
			originalRequestID = params.ExecutionContext.RequestID
		}

		// Send final response
		responseMsg := buildResponseMessage(params, originalRequestID, finalResult)

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

		params.Logger.Info("Sent final response",
			zap.String("topic that it was sent to (parentResponsesTopic from environment)", parentResponsesTopic),
			zap.String("request_id", originalRequestID))
	}

	return map[string]interface{}{"result": finalResult}, nil
}
*/
// Helper to build response message
func buildResponseMessage(execCtx *types.ExecutionContext, replyTo *ReplyToMetadata, result interface{}) types.ResponseMessage {
	return types.ResponseMessage{
		Headers: types.ResponseHeaders{
			Sender:                     execCtx.Sender,
			OrchestrationID:            execCtx.OrchestrationID,
			OrchestrationName:          execCtx.OrchestrationName,
			InResponseToRequestID:      replyTo.RequestID,
			InResponseToStepID:         replyTo.StepID,
			InResponseToStepName:       replyTo.StepName,
			InResponseToParentOrchID:   execCtx.ParentOrchestrationID,
			InResponseToParentOrchName: execCtx.ParentOrchestrationName,
			MyOrchestrationID:          execCtx.OrchestrationID,
			MyOrchestrationName:        execCtx.OrchestrationName,
			CorrelationID:              execCtx.CorrelationID,
			ClientID:                   execCtx.ClientID,
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

// mapToExecutionContext converts a map to ExecutionContext
func mapToExecutionContext(m map[string]interface{}, logger *zap.Logger) *types.ExecutionContext {
	// Marshal to JSON then unmarshal to struct
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		logger.Error("Failed to marshal execution context map", zap.Error(err))
		return nil
	}

	var execCtx types.ExecutionContext
	if err := json.Unmarshal(jsonBytes, &execCtx); err != nil {
		logger.Error("Failed to unmarshal execution context", zap.Error(err))
		return nil
	}

	return &execCtx
}

// getStringField safely gets a string field from a map
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
