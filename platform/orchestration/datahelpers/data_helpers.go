// internal/backend/agent-chassis/platform/orchestration/data_helpers.go
package datahelpers

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ============================================================================
// DATA HELPERS - Working with existing types.RequestMessage and types.ResponseMessage
// ============================================================================

// ============================================================================
// EXTRACTION FUNCTIONS - Get clean data from any message format
// ============================================================================

// ExtractDataFromMessage extracts clean data from ANY message format
// Works with types.RequestMessage, types.ResponseMessage, or raw maps
func ExtractDataFromMessage(source interface{}, logger *zap.Logger) map[string]interface{} {
	logger.Info("Now into ExtractDataFromMessage")

	if source == nil {
		logger.Debug("ExtractDataFromMessage: source is nil")
		return map[string]interface{}{}
	}

	// Handle typed messages first
	switch msg := source.(type) {
	case *types.RequestMessage:
		return extractFromRequestMessage(msg, logger)
	case types.RequestMessage:
		return extractFromRequestMessage(&msg, logger)
	case *types.ResponseMessage:
		return extractFromResponseMessage(msg, logger)
	case types.ResponseMessage:
		return extractFromResponseMessage(&msg, logger)
	case map[string]interface{}:
		return extractFromRawMap(msg, logger)
	default:
		logger.Warn("ExtractDataFromMessage: unknown source type",
			zap.String("type", fmt.Sprintf("%T", source)))
		return map[string]interface{}{}
	}
}

// extractFromRequestMessage extracts data from a typed RequestMessage
func extractFromRequestMessage(msg *types.RequestMessage, logger *zap.Logger) map[string]interface{} {
	logger.Info("extractDataFromRequestMessage",
		zap.Any("DEBUGaa: msg was - look for empty Body", msg),
	)

	if msg == nil || msg.Body == nil {
		return map[string]interface{}{}
	}

	logger.Info("extractDataFromRequestMessage wasnt empty")

	// Body could be a map or another type
	switch body := msg.Body.(type) {
	case map[string]interface{}:
		// Look for input_data field
		if data, ok := body["input_data"].(map[string]interface{}); ok {
			logger.Debug("extractFromRequestMessage: found body.input_data")
			return data
		}
		// Return cleaned body
		return CleanDataMap(body)
	default:
		logger.Debug("extractFromRequestMessage: body is not a map")
		return map[string]interface{}{}
	}
}

// extractFromResponseMessage extracts data from a typed ResponseMessage
func extractFromResponseMessage(msg *types.ResponseMessage, logger *zap.Logger) map[string]interface{} {
	logger.Info("extractDataFromResponseMessage",
		zap.Any("DEBUGaa: msg was ", msg),
	)

	if msg == nil {
		return map[string]interface{}{}
	}

	logger.Info("extractDataFromResponseMessage wasnt empty")

	// ResponseBody is a struct, not a pointer, so we work with it directly
	responseBody := msg.Body

	// Check if the Body field within ResponseBody has content
	if responseBody.Body == nil && responseBody.Error == nil {
		logger.Debug("extractFromResponseMessage: empty response body")
		return map[string]interface{}{}
	}

	// ResponseBody.Body contains the actual data
	if body, ok := responseBody.Body.(map[string]interface{}); ok {
		// Look for data field
		if data, ok := body["data"].(map[string]interface{}); ok {
			logger.Debug("extractFromResponseMessage: found body.data")
			return data
		}
		// Return the body itself if it's clean data
		logger.Debug("extractFromResponseMessage: using cleaned body")
		return CleanDataMap(body)
	}

	logger.Debug("extractFromResponseMessage: body is not a map")
	return map[string]interface{}{}
}

// extractFromRawMap handles raw map extraction (backward compatibility)
func extractFromRawMap(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Try different extraction strategies in order of preference

	logger.Info("extractFromRawMap",
		zap.Any("DEBUGaa: source was ", source),
	)

	// Strategy 1: Look for body.data (clean format)
	if data := extractFromPath(source, "body.data"); data != nil {
		logger.Debug("extractFromRawMap: found body.data")
		return data
	}

	// Strategy 2: Look for body.input_data (current format)
	if data := extractFromPath(source, "body.input_data"); data != nil {
		logger.Debug("extractFromRawMap: found body.input_data")
		return data
	}

	// Strategy 3: Look for input_data at any level
	if data := findDataField(source, "input_data"); data != nil {
		logger.Debug("extractFromRawMap: found input_data")
		return data
	}

	// Strategy 4: Look for data field at any level
	if data := findDataField(source, "data"); data != nil {
		logger.Debug("extractFromRawMap: found data")
		return data
	}

	// Strategy 5: Clean the body itself
	if body, ok := source["body"].(map[string]interface{}); ok {
		cleaned := CleanDataMap(body)
		if len(cleaned) > 0 {
			logger.Debug("extractFromRawMap: using cleaned body")
			return cleaned
		}
	}

	logger.Debug("extractFromRawMap: no data found")
	return map[string]interface{}{}
}

// ============================================================================
// MESSAGE BUILDING FUNCTIONS - Using existing types
// ============================================================================

// BuildRequestMessage creates a types.RequestMessage for sending to another agent
func BuildRequestMessage(
	execCtx *types.ExecutionContext,
	toAgentType string,
	action string,
	data map[string]interface{},
	config map[string]interface{},
	logger *zap.Logger,
) *types.RequestMessage {

	// Update context for this request
	execCtx.Action = action
	execCtx.ToAgentType = toAgentType

	// Build headers using existing method
	headers := execCtx.ToRequestHeaders()

	// Build body
	body := map[string]interface{}{
		"action":     action,
		"input_data": data,
	}

	if config != nil && len(config) > 0 {
		body["config"] = config
	}

	logger.Debug("BuildRequestMessage: created message",
		zap.String("action", action),
		zap.String("to_agent_type", toAgentType),
		zap.Int("data_fields", len(data)))

	return &types.RequestMessage{
		Headers: headers,
		Body:    body,
	}
}

// BuildResponseMessage creates a types.ResponseMessage for responding to parent
func BuildResponseMessage(
	execCtx *types.ExecutionContext,
	success bool,
	responseData map[string]interface{},
	errorInfo *types.ErrorInfo,
	logger *zap.Logger,
) *types.ResponseMessage {

	// Create response context
	responseCtx := execCtx.CreateResponseContext(
		determineStatus(success, errorInfo),
		100, // fuel used - calculate properly in production
	)

	// Build headers using existing method
	headers := responseCtx.ToResponseHeaders()

	// Build body
	responseBody := types.ResponseBody{
		Success: success,
		Body:    map[string]interface{}{"input_data": responseData},
		Error:   errorInfo,
	}

	logger.Debug("BuildResponseMessage: created response",
		zap.Bool("success", success),
		zap.Int("data_fields", len(responseData)))

	return &types.ResponseMessage{
		Headers: headers,
		Body:    responseBody,
	}
}

// BuildInitializationRequest creates an initialization request for a spawned agent
func BuildInitializationRequest(
	parentCtx *types.ExecutionContext,
	childAgentType string,
	role string,
	initialData map[string]interface{},
	agentConfig map[string]interface{},
	logger *zap.Logger,
) *types.RequestMessage {

	// Create child context
	childCtx := parentCtx.CreateChildContext(childAgentType)
	childCtx.Action = "initialize"

	// Set functional role if specified
	if role != "" {
		childCtx.FunctionalRole = role
	}

	// Build headers
	headers := childCtx.ToRequestHeaders()

	// Build initialization body
	body := map[string]interface{}{
		"action":            "initialize",
		"is_initialization": true,
		"agent_info": map[string]interface{}{
			"agent_id":   childCtx.OrchestrationID,
			"agent_type": childAgentType,
			"agent_name": childCtx.OrchestrationName,
		},
		"input_data": initialData,
	}

	if role != "" {
		body["role"] = role
	}

	if agentConfig != nil {
		body["config"] = agentConfig
	}

	logger.Debug("BuildInitializationRequest: created",
		zap.String("child_type", childAgentType),
		zap.String("child_orch_id", childCtx.OrchestrationID))

	return &types.RequestMessage{
		Headers: headers,
		Body:    body,
	}
}

// ============================================================================
// COLLECTED DATA MANAGEMENT - Works with ExecutionContext
// ============================================================================

// BuildCollectedData builds CollectedData from an incoming message
/*func BuildCollectedData(
	message interface{},
	execCtx *types.ExecutionContext,
	logger *zap.Logger,
) map[string]interface{} {

	collected := map[string]interface{}{
		"__execution_context__": execCtx,
	}

	// Extract data based on message type
	data := ExtractDataFromMessage(message, logger)
	if len(data) > 0 {
		collected["input_data"] = data
	}

	// Add action from context
	if execCtx.Action != "" {
		collected["action"] = execCtx.Action
	}

	// Extract config if it's a raw message with config
	if rawMsg, ok := message.(map[string]interface{}); ok {
		if config := extractSystemConfig(rawMsg); config != nil {
			collected["config"] = config
		}
		// Store raw for debugging
		collected["__raw_message__"] = rawMsg
	}

	// Add routing information
	collected["__my_requests_topic__"] = execCtx.RequestsTopic
	collected["__my_responses_topic__"] = execCtx.ResponsesTopic

	if execCtx.ReplyToTopic != "" {
		collected["__parent_responses_topic__"] = execCtx.ReplyToTopic
	}

	logger.Debug("BuildCollectedData: built structure",
		zap.Int("data_fields", len(data)),
		zap.String("action", execCtx.Action))

	return collected
}*/

// ============================================================================
// ORIGINAL FUNCTIONS - Maintained for backward compatibility
// ============================================================================

// NormalizeInputData extracts clean input_data from any source structure
func NormalizeInputData(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	logger.Info("NormalizeInputData: source:",
		zap.Any("source is:", source))

	if source == nil {
		logger.Debug("NormalizeInputData: source is nil, returning empty map")
		return map[string]interface{}{}
	}

	return ExtractDataFromMessage(source, logger)
}

// NormalizeCollectedData builds a clean CollectedData structure from a message
func NormalizeCollectedData(
	message map[string]interface{},
	execCtx *types.ExecutionContext,
	requestsTopic string,
	logger *zap.Logger,
) map[string]interface{} {

	logger.Debug("NormalizeCollectedData: building clean structure")

	// Use the new builder
	collectedData := BuildCollectedData(message, execCtx, logger)

	// Override requests topic if provided
	if requestsTopic != "" {
		collectedData["__my_requests_topic__"] = requestsTopic
	}

	// Add parent responses topic from environment if needed
	if execCtx.ReplyToTopic == "" {
		if parentTopic := os.Getenv("PARENT_RESPONSES_TOPIC"); parentTopic != "" {
			execCtx.ReplyToTopic = parentTopic
			collectedData["__parent_responses_topic__"] = parentTopic
		}
	}

	// Extract additional fields from message if needed
	if agentConfig, ok := message["agent_config"].(map[string]interface{}); ok {
		collectedData["agent_config"] = agentConfig
	}

	if agentGroup, ok := message["agent_group"].(map[string]interface{}); ok {
		collectedData["agent_group"] = agentGroup
	}

	if agentType, ok := message["agent_type"].(string); ok {
		collectedData["agent_type"] = agentType
	}

	if prompt, ok := message["prompt"].(string); ok {
		collectedData["prompt"] = prompt
	}

	logger.Debug("NormalizeCollectedData: structure complete",
		zap.Int("fields", len(collectedData)))

	return collectedData
}

// NormalizeResponseData cleans response data before storing in CollectedData
/*func NormalizeResponseData(responseBodyData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	logger.Info("In NormalizeResponseData",
		zap.Any("responseBodyData", responseBodyData),
	)
	if responseBodyData.Body == nil && responseBodyData.Error == nil {
		logger.Info("NormalizeResponseData: responseBody is nil")
		return map[string]interface{}{}
	}

	// ResponseBody.Body contains the actual data
	if body, ok := responseBodyData.Body.(map[string]interface{}); ok {
		logger.Info("NormalizeResponseData: processing body",
			zap.Any("body", body),
		)

		cleanResult := CleanDataMap(body)

		logger.Info("NormalizeResponseData: preserved fields",
			zap.Any("clean result", cleanResult),
		)

		return cleanResult
	}
	logger.Info("NormalizeResponseData: body is not a map")
	return map[string]interface{}{}
}*/

// GetInputData safely retrieves input_data from CollectedData
func GetInputData(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if collectedData == nil {
		logger.Warn("GetInputData: collectedData is nil")
		return map[string]interface{}{}
	}

	// First try the normalized location
	if data, ok := collectedData["input_data"].(map[string]interface{}); ok {
		logger.Info("in GetInputData - we should have input data here",
			zap.Any("input_data", data),
		)
		return data
	}

	// Fallback to extraction
	return ExtractDataFromMessage(collectedData, logger)
}

// GetStepData safely retrieves data from a specific step's response
func GetStepData(collectedData map[string]interface{}, stepName string, logger *zap.Logger) (map[string]interface{}, bool) {
	if collectedData == nil {
		logger.Warn("GetStepData: collectedData is nil", zap.String("step", stepName))
		return nil, false
	}

	if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
		logger.Debug("GetStepData: found step data", zap.String("step", stepName))
		return stepData, true
	}

	logger.Debug("GetStepData: step data not found", zap.String("step", stepName))
	return nil, false
}

// GetMultipleStepData retrieves data from multiple steps
func GetMultipleStepData(collectedData map[string]interface{}, stepNames []string, logger *zap.Logger) map[string]interface{} {
	if collectedData == nil {
		logger.Warn("GetMultipleStepData: collectedData is nil")
		return map[string]interface{}{}
	}

	result := make(map[string]interface{})

	for _, stepName := range stepNames {
		if stepData, ok := collectedData[stepName].(map[string]interface{}); ok {
			result[stepName] = stepData
			logger.Debug("GetMultipleStepData: collected step data", zap.String("step", stepName))
		}
	}

	logger.Debug("GetMultipleStepData: collection complete",
		zap.Int("requested", len(stepNames)),
		zap.Int("found", len(result)))

	return result
}

// GetFieldFromPath retrieves a value using dot-notation path
func GetFieldFromPath(collectedData map[string]interface{}, path string, logger *zap.Logger) (interface{}, error) {
	if collectedData == nil {
		return nil, fmt.Errorf("collectedData is nil")
	}

	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	val := getFieldByPath(collectedData, path)
	if val == nil {
		return nil, fmt.Errorf("field not found: %s", path)
	}

	logger.Debug("GetFieldFromPath: found value", zap.String("path", path))
	return val, nil
}

// GetFieldFromPathWithDefault retrieves a value with a fallback default
func GetFieldFromPathWithDefault(collectedData map[string]interface{}, path string, defaultValue interface{}, logger *zap.Logger) interface{} {
	val, err := GetFieldFromPath(collectedData, path, logger)
	if err != nil {
		logger.Debug("GetFieldFromPathWithDefault: using default",
			zap.String("path", path),
			zap.Error(err))
		return defaultValue
	}
	return val
}

// MergeInputData merges new input data with existing
func MergeInputData(existing, new map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing
	for k, v := range existing {
		result[k] = v
	}

	// Overlay new (overwrites)
	for k, v := range new {
		result[k] = v
	}

	logger.Debug("MergeInputData: merged data",
		zap.Int("existing_fields", len(existing)),
		zap.Int("new_fields", len(new)),
		zap.Int("result_fields", len(result)))

	return result
}

/*// UpdateCollectedData safely updates CollectedData with new information
func UpdateCollectedData(
	collected map[string]interface{},
	stepName string,
	stepData map[string]interface{},
	logger *zap.Logger,
) {
	if collected == nil || stepData == nil {
		return
	}

	// Store step results
	collected[stepName] = stepData

	// Also merge into input_data if it's a data update
	if currentData, ok := collected["input_data"].(map[string]interface{}); ok {
		// Only merge actual data fields, not system fields
		cleanStepData := CleanDataMap(stepData)
		for k, v := range cleanStepData {
			currentData[k] = v
		}
		logger.Debug("UpdateCollectedData: merged step data",
			zap.String("step", stepName),
			zap.Int("merged_fields", len(cleanStepData)))
	}
}*/

// ============================================================================
// TRANSFORMATION FUNCTIONS
// ============================================================================

// TransformDataForAction prepares data for a specific action (dynamic)
func TransformDataForAction(
	sourceData map[string]interface{},
	transformSpec map[string]interface{},
	logger *zap.Logger,
) map[string]interface{} {

	if transformSpec == nil {
		return sourceData
	}

	result := make(map[string]interface{})

	// Apply field mappings if specified
	if mappings, ok := transformSpec["field_mappings"].(map[string]interface{}); ok {
		for targetField, sourceField := range mappings {
			if sourceFieldStr, ok := sourceField.(string); ok {
				if val := getFieldByPath(sourceData, sourceFieldStr); val != nil {
					result[targetField] = val
				}
			}
		}
	}

	// Apply field filters if specified
	if fields, ok := transformSpec["include_fields"].([]interface{}); ok {
		for _, field := range fields {
			if fieldStr, ok := field.(string); ok {
				if val, exists := sourceData[fieldStr]; exists {
					result[fieldStr] = val
				}
			}
		}
	}

	// If no specific transformation, use source data
	if len(result) == 0 {
		return sourceData
	}

	// Add any additional fields specified
	if additions, ok := transformSpec["add_fields"].(map[string]interface{}); ok {
		for k, v := range additions {
			result[k] = v
		}
	}

	logger.Debug("TransformDataForAction: applied transformation",
		zap.Int("source_fields", len(sourceData)),
		zap.Int("result_fields", len(result)))

	return result
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// extractFromPath extracts data from a nested path like "body.data"
func extractFromPath(source map[string]interface{}, path string) map[string]interface{} {
	current := source
	segments := strings.Split(path, ".")

	for _, segment := range segments {
		if next, ok := current[segment].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return current
}

// findDataField recursively finds a data field in the structure
func findDataField(source map[string]interface{}, fieldName string) map[string]interface{} {
	// Check current level
	if data, ok := source[fieldName].(map[string]interface{}); ok {
		// Make sure it's actually data, not a wrapper
		if !hasSystemFields(data) {
			return data
		}
		// It might be nested, recurse
		if nested := findDataField(data, fieldName); nested != nil {
			return nested
		}
	}

	// Check one level deeper
	for key, value := range source {
		if key == "__raw_message__" || key == "__execution_context__" {
			continue
		}
		if nested, ok := value.(map[string]interface{}); ok {
			if result := findDataField(nested, fieldName); result != nil {
				return result
			}
		}
	}

	return nil
}

// CleanDataMap removes system fields from a map
func CleanDataMap(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	systemFields := map[string]bool{
		"action":                       true,
		"config":                       true,
		"agent_config":                 true,
		"agent_group":                  true,
		"__execution_context__":        true,
		"__raw_message__":              true,
		"__my_requests_topic__":        true,
		"__reply_to_responses_topic__": true,
		"__parent_responses_topic__":   true,
		"headers":                      true,
		"is_initialization":            true,
		"agent_info":                   true,
		//"role":                       true,
	}

	for k, v := range source {
		if !systemFields[k] {
			result[k] = v
		}
	}

	return result
}

// hasSystemFields checks if a map contains system fields
func hasSystemFields(data map[string]interface{}) bool {
	systemFields := []string{"action", "config", "agent_config", "__execution_context__", "workflow", "headers"}
	for _, field := range systemFields {
		if _, exists := data[field]; exists {
			return true
		}
	}
	return false
}

// extractAction finds the action from various possible locations
func extractAction(message map[string]interface{}) string {
	// Try body.action first
	if body, ok := message["body"].(map[string]interface{}); ok {
		if action, ok := body["action"].(string); ok {
			return action
		}
	}

	// Try top-level action
	if action, ok := message["action"].(string); ok {
		return action
	}

	// Try headers.action
	if headers, ok := message["headers"].(map[string]interface{}); ok {
		if action, ok := headers["action"].(string); ok {
			return action
		}
	}

	return ""
}

// extractSystemConfig extracts workflow/system configuration
func extractSystemConfig(message map[string]interface{}) map[string]interface{} {
	// Try body.config first
	if body, ok := message["body"].(map[string]interface{}); ok {
		if config, ok := body["config"].(map[string]interface{}); ok {
			if isSystemConfig(config) {
				return config
			}
		}
	}

	// Try top-level config
	if config, ok := message["config"].(map[string]interface{}); ok {
		if isSystemConfig(config) {
			return config
		}
	}

	return nil
}

// isSystemConfig checks if config is system/workflow config vs user data
func isSystemConfig(config map[string]interface{}) bool {
	systemKeys := []string{"workflow", "group_type", "processing_mode", "ai_service"}
	for _, key := range systemKeys {
		if _, exists := config[key]; exists {
			return true
		}
	}
	return false
}

// getFieldByPath retrieves a field using dot notation
// possibly legacy / unused
func getFieldByPath(data map[string]interface{}, path string) interface{} {
	segments := strings.Split(path, ".")
	var current interface{} = data

	for _, segment := range segments {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[segment]
		} else {
			return nil
		}
	}

	return current
}

// determineStatus converts success/error info to status string
func determineStatus(success bool, errorInfo *types.ErrorInfo) string {
	if success {
		return "complete"
	}
	if errorInfo != nil && errorInfo.Recoverable {
		return "error_recoverable"
	}
	return "error_unrecoverable"
}

// BuildCollectedData with double body.body unwrapping
func BuildCollectedData(
	messageBody interface{},
	execCtx *types.ExecutionContext,
	logger *zap.Logger,
) map[string]interface{} {

	logger.Info("Data helpers - BuildCollectedData inwards",
		zap.Any("DEBUGaa: message body", messageBody),
		zap.Any("DEBUGaa: exectCtx", execCtx))

	collectedData := make(map[string]interface{})

	// Add execution context (own exec ctx unless it's a child then overwrite later)
	collectedData["__execution_context__"] = execCtx

	// Add system topics
	collectedData["__my_requests_topic__"] = execCtx.RequestsTopic
	collectedData["__my_responses_topic__"] = execCtx.ResponsesTopic

	// Add parent response topic if available
	if execCtx.ReplyToTopic != "" {
		collectedData["__parent_responses_topic__"] = execCtx.ReplyToTopic
	}

	// Preserve the parent request ID that we need to respond to
	if execCtx.ReplyToRequestID != "" {
		collectedData["__reply_to_request_id__"] = execCtx.ReplyToRequestID
	}

	// Extract from message body based on its structure
	var unnestedBody map[string]interface{}
	var parentResponsesTopic string
	switch body := messageBody.(type) {
	case map[string]interface{}:
		// Check if there's a nested "body" key
		if bodyThatWasNested, hasBodyKey := body["body"].(map[string]interface{}); hasBodyKey {
			logger.Info("BuildCollectedData: found nested 'body' key, unwrapping once")
			unnestedBody = bodyThatWasNested

			// Check if there's ANOTHER nested "body" key (double nesting from ResponseBody structure)
			// This happens when responses have: { body: { body: {actual data}, success: true } }
			// The outer "body" is the ResponseMessage.Body field
			// The inner "body" is the ResponseBody.Body field
			if bodyThatWasDoubleNested, hasSecondBodyKey := unnestedBody["body"].(map[string]interface{}); hasSecondBodyKey {
				logger.Info("BuildCollectedData: found DOUBLE nested 'body' key, unwrapping again",
					zap.Bool("has_success_field", unnestedBody["success"] != nil))
				unnestedBody = bodyThatWasDoubleNested
			}
		} else {
			unnestedBody = body
		}

		// Extract the actual input_data if it exists
		if inputData, ok := unnestedBody["input_data"].(map[string]interface{}); ok {
			logger.Info("Processor: found input_data in message body",
				zap.Any("input_data", inputData))
			collectedData["input_data"] = inputData
		} else {
			// No explicit input_data field
			// Check if body itself is the data (doesn't have system fields)
			if !hasSystemFields(unnestedBody) {
				logger.Info("Processor: treating entire unnested body as input_data")
				collectedData["input_data"] = unnestedBody
			} else {
				// Body has system fields, don't use it as input_data
				logger.Info("Processor: no input_data found, using empty map")
				collectedData["input_data"] = make(map[string]interface{})
			}
		}

		// get responses topic from message value if it isn't in execCtx
		if execCtx.ReplyToTopic == "" {
			if val, ok := unnestedBody["parent_responses_topic"].(string); ok {
				parentResponsesTopic = val
				logger.Info("Processor: buildCollecteData helper: found parent_responses_topic in message body",
					zap.Any("parent responses topic", parentResponsesTopic))
			}
		} else {
			logger.Info("Processor buildCollecteData helper: used parentResponsesTopic from execCtx",
				zap.Any("parent responses topic", parentResponsesTopic))
			parentResponsesTopic = execCtx.ReplyToTopic
		}
		if parentResponsesTopic == "" {
			parentResponsesTopic = os.Getenv("PARENT_RESPONSES_TOPIC")
		}
		collectedData["__parent_responses_topic__"] = parentResponsesTopic

		// Store action separately if present
		if action, ok := unnestedBody["action"].(string); ok {
			collectedData["action"] = action
		}

		// Store config separately if present
		if config, ok := unnestedBody["config"].(map[string]interface{}); ok {
			collectedData["config"] = config
		}

		// Store prompt separately if present
		if prompt, ok := unnestedBody["prompt"].(string); ok {
			collectedData["prompt"] = prompt
		}

		// Add agent_config extraction
		if agentConfig, ok := unnestedBody["agent_config"].(map[string]interface{}); ok {
			collectedData["agent_config"] = agentConfig
		}

		// Store the raw message for debugging (use original body, not unnested)
		collectedData["__raw_message__"] = body

	default:
		// Non-map body, store as-is
		logger.Warn("Data helpers: message body is not a map",
			zap.Any("body_type", fmt.Sprintf("%T", messageBody)))
		collectedData["input_data"] = make(map[string]interface{})
		collectedData["__raw_message__"] = messageBody
	}

	workRequestMetadata := map[string]interface{}{
		"request_id":             execCtx.RequestID,
		"parent_responses_topic": parentResponsesTopic,
		"requester_agent_id":     execCtx.Sender.AgentID,
		"requester_agent_type":   execCtx.Sender.AgentType,
		"step_id":                execCtx.StepID,
		"step_name":              execCtx.StepName,
		"action":                 execCtx.Action,
		"correlation_id":         execCtx.CorrelationID,
		"timestamp":              time.Now().Format(time.RFC3339),
	}
	collectedData["__work_request__"] = workRequestMetadata

	logger.Info("BuildCollectedData: stored work request metadata for initialize",
		zap.String("request_id", workRequestMetadata["request_id"].(string)),
		zap.String("parent_topic", workRequestMetadata["parent_responses_topic"].(string)),
		zap.Any("work_request_metadata", workRequestMetadata),
	)

	logger.Info("DEBUGaa: Data helpers: built CollectedData requestIDs",
		zap.String("what agent am I on", os.Getenv("AGENT_TYPE")),
		zap.String("what should the request id be - where am I in the process and what was the request id should I move it to parent", "nothing to see here"),
		zap.Any("DEBUGaa: _execution_context__ look for requestid and parent request id", execCtx),
	)

	logger.Info("Data helpers: built CollectedData",
		zap.Any("collected_keys", GetMapKeys(collectedData)),
		zap.Any("input_data", collectedData["input_data"]))

	return collectedData
}

// GetMapKeys returns the keys of a map as a string slice
func GetMapKeys(m map[string]interface{}) []string {
	if m == nil {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

func GetValueByPath(data map[string]interface{}, path string, logger *zap.Logger) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = data

	logger.Info("In GetValueByPath",
		zap.String("path", path),
		zap.Any("keys", keys),
		zap.Any("DEBUGaa: data", current),
	)

	// traverses down the dotted keys
	for _, key := range keys {
		currMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := currMap[key]
		if !exists {
			return nil, false
		}
		// write to current and start loop again looking for next key
		current = val
	}
	return current, true
}

// searchNestedForJSON recursively searches nested maps for common result field patterns.
// maxDepth prevents infinite recursion.
func SearchNestedForJSON(m map[string]interface{}, resultFields []string, logger *zap.Logger, depth int) string {
	const maxDepth = 3 // Prevent going too deep

	if depth > maxDepth {
		return ""
	}

	// For each value in the map, check if it's a nested map
	for key, val := range m {
		if nestedMap, ok := val.(map[string]interface{}); ok {
			// Check if this nested map has any of the result fields
			for _, fieldName := range resultFields {
				if result, ok := nestedMap[fieldName].(string); ok {
					cleaned := CleanMarkdownJSON(result)
					logger.Debug("Found result in nested structure",
						zap.String("parent_key", key),
						zap.String("result_field", fieldName),
						zap.Int("depth", depth),
						zap.Int("cleaned_length", len(cleaned)),
					)
					return cleaned
				}
			}

			// Recurse deeper
			if jsonStr := SearchNestedForJSON(nestedMap, resultFields, logger, depth+1); jsonStr != "" {
				return jsonStr
			}
		}
	}

	return ""
}

// cleanMarkdownJSON removes markdown code fences and other LLM artifacts.
func CleanMarkdownJSON(s string) string {
	s = strings.TrimSpace(s)

	// Remove ```json ... ``` wrappers (most common)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimSpace(s) // Remove newline after ```json

	// Remove plain ``` wrappers
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSpace(s)

	// Remove trailing ```
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// Extra cleanup: remove any leading/trailing quotes if the entire thing is quoted
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		// This might be double-encoded JSON
		s = s[1 : len(s)-1]
	}

	return strings.TrimSpace(s)
}

func RenderPromptTemplate(templateStr string, data map[string]interface{}, logger zap.Logger) (string, error) {
	tmpl := template.New("agent_prompt")
	parsedTemplate, err := tmpl.Parse(templateStr)
	logger.Info("DEBUGaa: parsing template in renderTemplate",
		zap.String("template", templateStr),
		zap.Any("data", data),
		zap.Any("tmpl", tmpl),
		zap.Any("parsedTemplate", parsedTemplate),
	)
	if err != nil {
		return "", fmt.Errorf("failed to parse template in render template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ExtractNestedField extracts a value from nested map using dot notation path.
// Returns nil if any part of the path is not found.
//
// Example usage:
//
//	value := ExtractNestedField(data, "input_data.domain")
//	if value != nil {
//	    domain := value.(string)
//	}
//
// Paths like "input_data.content_json.sections.component_header_0.brand_name"
// will traverse: data["input_data"]["content_json"]["sections"]["component_header_0"]["brand_name"]
func ExtractNestedField(data map[string]interface{}, fieldPath string) interface{} {
	parts := strings.Split(fieldPath, ".")
	var current interface{} = data

	for _, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[part]
			if current == nil {
				return nil
			}
		} else {
			return nil
		}
	}

	return current
}

// ExtractNestedFieldString is a convenience wrapper that returns empty string if not found or not a string
func ExtractNestedFieldString(data map[string]interface{}, fieldPath string) string {
	value := ExtractNestedField(data, fieldPath)
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// ExtractNestedFieldMap is a convenience wrapper that returns nil if not found or not a map
func ExtractNestedFieldMap(data map[string]interface{}, fieldPath string) map[string]interface{} {
	value := ExtractNestedField(data, fieldPath)
	if value == nil {
		return nil
	}
	if m, ok := value.(map[string]interface{}); ok {
		return m
	}
	return nil
}

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

// CleanHTMLString removes markdown code blocks and extra whitespace
func CleanHTMLString(s string) string {
	// Remove markdown code blocks
	codeBlockRe := regexp.MustCompile("```(?:html|css)?\n([\\s\\S]*?)```")
	s = codeBlockRe.ReplaceAllString(s, "$1")

	// Trim leading/trailing whitespace
	s = strings.TrimSpace(s)

	return s
}
