// platform/orchestration/actions/workflow_actions.go

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
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

// EvaluateConditionAction evaluates a condition for conditional workflow branching.
// Supports two config formats
// supports two config formats:
//
// Format 1 (Original - template-based):
//
//	config: {
//	  "condition": "{{.input_data.some_field}}",  // Go template
//	  "default": false                             // boolean default
//	}
//	Returns: {"result": bool, "condition": string, "evaluated": string}
//
// Format 2 (New - field + conditions map, used by briefing-agent):
//
//	config: {
//	  "condition_field": "input_data.hitl_mode",   // dot-path to field
//	  "conditions": {                              // value -> next_step map
//	    "interactive": "collect_via_hitl",
//	    "auto": "infer_via_llm"
//	  },
//	  "default": "infer_via_llm"                   // default next_step (string)
//	}
//	Returns: {"next_step": string, "condition_value": string, "matched": bool}
/*{
"condition_field": "input_data.hitl_mode",
"conditions": {"interactive": "step1", "auto": "step2"},
"default": "step2"
}*/
func EvaluateConditionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing EvaluateConditionAction",
		zap.String("step_name", params.ExecutionContext.StepName))

	config := params.StepConfig.Config

	// Detect which format we're using
	if conditionField, ok := config["condition_field"].(string); ok {
		// Format 2: condition_field + conditions map
		return evaluateConditionFieldFormat(params, conditionField, config)
	}

	// Format 1: template-based condition
	return evaluateConditionTemplateFormat(params, config)
}

// evaluateConditionFieldFormat handles the new format used by briefing-agent
// config: {"condition_field": "path.to.field", "conditions": {"value1": "step1", ...}, "default": "default_step"}
func evaluateConditionFieldFormat(params ActionParams, conditionField string, config map[string]interface{}) (interface{}, error) {
	logger := params.Logger

	// Get the conditions map
	conditionsRaw, ok := config["conditions"]
	if !ok {
		return nil, fmt.Errorf("evaluate_condition with condition_field requires 'conditions' map in config")
	}

	conditions, ok := conditionsRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'conditions' must be a map[string]interface{}, got %T", conditionsRaw)
	}

	// Get default next step
	defaultStep := ""
	if def, ok := config["default"].(string); ok {
		defaultStep = def
	}

	// Resolve the condition field value from collected data
	conditionValue := resolveFieldPath(params.CollectedData, conditionField)

	logger.Debug("Evaluating condition field",
		zap.String("condition_field", conditionField),
		zap.Any("condition_value", conditionValue),
		zap.Any("conditions", conditions),
		zap.String("default", defaultStep))

	// Convert condition value to string for lookup
	valueStr := fmt.Sprintf("%v", conditionValue)
	if conditionValue == nil {
		valueStr = ""
	}

	// Look up the next step based on condition value
	var nextStep string
	var matched bool

	if step, exists := conditions[valueStr]; exists {
		nextStep = fmt.Sprintf("%v", step)
		matched = true
	} else if defaultStep != "" {
		nextStep = defaultStep
		matched = false
	} else {
		// No match and no default - this is an error
		return nil, fmt.Errorf("condition value '%s' not found in conditions map and no default specified", valueStr)
	}

	logger.Info("Condition evaluated (field format)",
		zap.String("condition_field", conditionField),
		zap.String("condition_value", valueStr),
		zap.String("next_step", nextStep),
		zap.Bool("matched", matched))

	return map[string]interface{}{
		"next_step":       nextStep,
		"condition_value": valueStr,
		"condition_field": conditionField,
		"matched":         matched,
	}, nil
}

// evaluateConditionTemplateFormat handles the original template-based format
// config: {"condition": "{{.field}}", "default": bool}
func evaluateConditionTemplateFormat(params ActionParams, config map[string]interface{}) (interface{}, error) {
	logger := params.Logger

	// Get the condition template
	conditionTemplate, ok := config["condition"].(string)
	if !ok {
		return nil, fmt.Errorf("condition not specified in config (expected 'condition' template string or 'condition_field' path)")
	}

	// Get default value if condition fails or is empty
	defaultValue := false
	if def, ok := config["default"].(bool); ok {
		defaultValue = def
	}

	// Parse and execute the condition template
	tmpl, err := template.New("condition").Parse(conditionTemplate)
	if err != nil {
		logger.Error("Failed to parse condition template",
			zap.Error(err),
			zap.String("template", conditionTemplate))
		return map[string]interface{}{
			"result": defaultValue,
			"error":  err.Error(),
		}, nil
	}

	// Execute template with collected data
	var result strings.Builder
	if err := tmpl.Execute(&result, params.CollectedData); err != nil {
		logger.Warn("Condition evaluation failed, using default",
			zap.Error(err),
			zap.Bool("default", defaultValue))
		return map[string]interface{}{
			"result": defaultValue,
			"error":  err.Error(),
		}, nil
	}

	// Evaluate the result
	conditionResult := evaluateConditionString(result.String(), defaultValue, logger)

	logger.Debug("Condition evaluated (template format)",
		zap.String("condition", conditionTemplate),
		zap.String("evaluated_to", result.String()),
		zap.Bool("result", conditionResult))

	// Return the result for workflow branching
	return map[string]interface{}{
		"result":    conditionResult,
		"condition": conditionTemplate,
		"evaluated": result.String(),
	}, nil
}

// resolveFieldPath gets a value from nested maps using dot notation
// e.g., "input_data.hitl_mode" -> collectedData["input_data"]["hitl_mode"]
func resolveFieldPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		if current == nil {
			return nil
		}

		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case map[interface{}]interface{}:
			current = v[part]
		default:
			return nil
		}
	}

	return current
}

// evaluateConditionString converts various string values to boolean
func evaluateConditionString(value string, defaultValue bool, logger *zap.Logger) bool {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	// Handle empty values
	if trimmed == "" || trimmed == "<no value>" {
		return defaultValue
	}

	// Handle boolean strings
	switch trimmed {
	case "true", "yes", "1", "on", "enabled":
		return true
	case "false", "no", "0", "off", "disabled":
		return false
	}

	// Try parsing as number
	if num, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return num != 0
	}

	// Non-empty string = true, unless it explicitly looks like a negative
	if strings.HasPrefix(trimmed, "not ") || strings.HasPrefix(trimmed, "no ") {
		return false
	}

	// Any other non-empty string is considered true
	return trimmed != ""
}

// SplitURLsAction splits a list of URLs for batch processing
func SplitURLsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing SplitURLsAction")

	config := params.StepConfig.Config
	urlsField := "target_urls"
	if field, ok := config["urls_field"].(string); ok {
		urlsField = field
	}

	maxParallel := 3
	if max, ok := config["max_parallel"].(float64); ok {
		maxParallel = int(max)
	}

	// Get URLs from input data
	var urls []string
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		// Handle both single URL and multiple URLs
		if singleURL, ok := inputData["target_url"].(string); ok {
			urls = []string{singleURL}
		} else if urlList, ok := inputData[urlsField].([]interface{}); ok {
			for _, u := range urlList {
				if urlStr, ok := u.(string); ok {
					urls = append(urls, urlStr)
				}
			}
		} else if urlList, ok := inputData[urlsField].([]string); ok {
			urls = urlList
		}
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs found in %s", urlsField)
	}

	// Split URLs into batches
	batches := make(map[string]interface{})
	batchSize := (len(urls) + maxParallel - 1) / maxParallel

	for i := 0; i < maxParallel && i < len(urls); i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(urls) {
			end = len(urls)
		}

		if start < len(urls) {
			batchKey := fmt.Sprintf("batch_%d", i+1)
			if end-start == 1 {
				// Single URL in batch
				batches[batchKey] = urls[start]
			} else {
				// Multiple URLs in batch
				batches[batchKey] = urls[start:end]
			}
		}
	}

	params.Logger.Debug("URLs split into batches",
		zap.Int("total_urls", len(urls)),
		zap.Int("batches", len(batches)))

	return batches, nil
}

// AggregateDataAction combines data from multiple workflow steps
func AggregateScrapedDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing AggregateDataAction")

	config := params.StepConfig.Config

	// Get fields to aggregate
	responseFields, ok := config["response_fields"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("response_fields not specified")
	}

	aggregated := make(map[string]interface{})

	// Collect data from specified fields
	for _, field := range responseFields {
		fieldName, ok := field.(string)
		if !ok {
			continue
		}

		if data, exists := params.CollectedData[fieldName]; exists && data != nil {
			aggregated[fieldName] = data
		}
	}

	// Add metadata if requested
	if includeMetadata, ok := config["include_metadata"].(bool); ok && includeMetadata {
		aggregated["metadata"] = map[string]interface{}{
			"aggregated_at":    params.ExecutionContext.StepName,
			"fields_count":     len(aggregated),
			"orchestration_id": params.ExecutionContext.OrchestrationID,
		}
	}

	// Apply formatting if specified
	if formatConfig, ok := config["format_output"].(map[string]interface{}); ok {
		if formatConfig["summary"] == true {
			aggregated["summary"] = generateSummary(aggregated)
		}
		if formatConfig["s3_links"] == true {
			aggregated["s3_links"] = extractS3Links(aggregated)
		}
	}

	params.Logger.Debug("Data aggregated",
		zap.Int("fields_aggregated", len(aggregated)))

	return aggregated, nil
}

// Helper functions
func generateSummary(data map[string]interface{}) map[string]interface{} {
	summary := map[string]interface{}{
		"total_items": len(data),
		"has_errors":  false,
	}

	// Check for errors in the data
	for key, value := range data {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if _, hasError := valueMap["error"]; hasError {
				summary["has_errors"] = true
				break
			}
		}
		summary[key+"_present"] = value != nil
	}

	return summary
}

func extractS3Links(data map[string]interface{}) []string {
	var links []string

	var extractLinks func(interface{})
	extractLinks = func(v interface{}) {
		switch val := v.(type) {
		case map[string]interface{}:
			for key, value := range val {
				if strings.Contains(key, "_uri") || strings.Contains(key, "_url") {
					if str, ok := value.(string); ok && strings.Contains(str, "s3://") {
						links = append(links, str)
					}
				}
				extractLinks(value)
			}
		case []interface{}:
			for _, item := range val {
				extractLinks(item)
			}
		}
	}

	extractLinks(data)
	return links
}
