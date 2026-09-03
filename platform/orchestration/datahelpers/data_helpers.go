// internal/backend/agent-chassis/platform/orchestration/data_helpers.go
package datahelpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
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

		// Get parent responses topic from message body if not in execCtx
		if execCtx.ReplyToTopic == "" {
			// Try parent_responses_topic in message body
			if val, ok := unnestedBody["parent_responses_topic"].(string); ok && val != "" {
				parentResponsesTopic = val
				logger.Info("BuildCollectedData: found parent_responses_topic in message body",
					zap.String("parent_responses_topic", parentResponsesTopic))
			}
			// Try __execution_context__.reply_to_topic in message body
			if parentResponsesTopic == "" {
				if embeddedExecCtx, ok := unnestedBody["__execution_context__"].(map[string]interface{}); ok {
					if val, ok := embeddedExecCtx["reply_to_topic"].(string); ok && val != "" {
						parentResponsesTopic = val
						logger.Info("BuildCollectedData: found reply_to_topic in embedded __execution_context__",
							zap.String("parent_responses_topic", parentResponsesTopic))
					}
				}
			}
		} else {
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

	// ============================================================================
	// Check if __work_request__ already exists in the message body
	// If it does, preserve it instead of creating a new one
	// ============================================================================
	var existingWorkRequest map[string]interface{}
	if unnestedBody != nil {
		if wr, ok := unnestedBody["__work_request__"].(map[string]interface{}); ok {
			logger.Info("BuildCollectedData: found existing __work_request__ in message, preserving it",
				zap.String("existing_request_id", fmt.Sprintf("%v", wr["request_id"])),
				zap.Any("existing_work_request", wr))
			existingWorkRequest = wr
		}
	}

	// Only create new work request metadata if one doesn't already exist
	if existingWorkRequest != nil {
		collectedData["__work_request__"] = existingWorkRequest
		logger.Info("BuildCollectedData: preserved existing work request metadata",
			zap.String("request_id", fmt.Sprintf("%v", existingWorkRequest["request_id"])),
			zap.String("parent_topic", fmt.Sprintf("%v", existingWorkRequest["parent_responses_topic"])))
	} else {
		// Create new work request metadata from current execution context
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

		logger.Info("BuildCollectedData: created new work request metadata",
			zap.String("request_id", workRequestMetadata["request_id"].(string)),
			zap.String("parent_topic", workRequestMetadata["parent_responses_topic"].(string)),
			zap.Any("work_request_metadata", workRequestMetadata))
	}

	logger.Info("DEBUGaa: Data helpers: built CollectedData requestIDs",
		zap.String("what agent am I on", os.Getenv("AGENT_TYPE")),
		zap.String("what should the request id be - where am I in the process and what was the request id should I move it to parent", "nothing to see here"),
		//zap.Any("DEBUGaa: _execution_context__ look for requestid and parent request id", execCtx),
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

// SafeCut returns at most n BYTES of s, never splitting a multi-byte rune.
// The one truncation primitive in this package: TruncateString is expressed in
// terms of it, and callers that need a cut WITHOUT an ellipsis (e.g. image
// prompt composition, where a trailing "..." would reach the model as an
// instruction rather than a display affordance) use it directly.
// Added 2026-07-20, bugs_open/027 §4b (council correlation 0a07f5ed).
func SafeCut(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	// was: s[:maxLength] + "..." — a raw byte slice, which split multi-byte
	// runes landing on the boundary and emitted invalid UTF-8 into logs and
	// prompt previews. SafeCut backs off to the last rune start instead.
	return SafeCut(s, maxLength) + "..."
}

// UpperFirst upper-cases the FIRST RUNE of s and leaves the rest byte-identical.
//
// It is the rune-safe replacement for the `strings.ToUpper(s[:1]) + s[1:]`
// idiom this estate hand-rolled at eight call sites (census 2026-09-02). That
// idiom is a BYTE slice: when the first rune is multi-byte it hands ToUpper a
// lone lead byte — which decodes as U+FFFD — and then re-attaches the orphaned
// continuation bytes, so the result is invalid UTF-8. Postgres REFUSES invalid
// UTF-8, so the corrupted string does not degrade quietly; it kills whatever
// statement tries to persist it.
//
// bugs_open/423 is the worked case, and it is worth stating because the trigger
// looks harmless: a word-splitter made a standalone em-dash in a page title
// ("Boxing Quiz — Test Your Knowledge") its own word, w[:1] cut it after one
// byte, and the site's footer store failed with `invalid byte sequence for
// encoding "UTF8": 0x80` — on two live sites, for ten days and two weeks
// respectively, reported as nothing at all.
//
// On ASCII input the output is byte-identical to the idiom it replaces, which
// is what made converting every call site in one pass safe.
func UpperFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		// Already invalid UTF-8 on the way in. Return it untouched rather than
		// substituting U+FFFD: this helper's job is to not CREATE the defect,
		// and silently rewriting a caller's bad bytes would hide one that a
		// validator upstream should be naming.
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// invalidUTF8Window is how many bytes either side of a bad byte the report
// carries. Wide enough to recognise the offending text, short enough to log.
const invalidUTF8Window = 48

// InvalidUTF8At returns the byte offset of the first invalid UTF-8 sequence in
// s together with a printable window around it, or found=false when s is clean.
//
// It exists because Postgres names the offending BYTE and never its position —
// `invalid byte sequence for encoding "UTF8": 0x80` — so a store refusal on a
// 40 KB document tells you a multi-byte rune was cut somewhere in it and
// nothing more, and the only way to find out where has been to bisect the
// pipeline by hand (bugs_open/423 cost two sessions roughly an hour before the
// mechanism was even named).
//
// The window is QuoteToASCII'd deliberately: the result is pure ASCII with each
// bad byte shown as \x80, so the report can be logged, persisted and put in a
// work item WITHOUT reproducing the defect it describes. A diagnostic that
// cannot itself be stored is the shape this bug already bit us with once, at
// the summary truncations in the chrome failure-reporting path.
func InvalidUTF8At(s string) (offset int, window string, found bool) {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			lo := i - invalidUTF8Window
			if lo < 0 {
				lo = 0
			}
			hi := i + invalidUTF8Window
			if hi > len(s) {
				hi = len(s)
			}
			return i, strconv.QuoteToASCII(s[lo:hi]), true
		}
		i += size
	}
	return 0, "", false
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
	// tmpl := template.New("agent_prompt")
	tmpl := template.New("agent_prompt").Funcs(PromptTemplateFuncs())
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

	rendered := buf.String()

	// ========================================================================
	// NEW: Check for <no value> placeholders which indicate missing data
	// ========================================================================
	if strings.Contains(rendered, "<no value>") {
		// Count occurrences
		count := strings.Count(rendered, "<no value>")
		logger.Warn("TEMPLATE RENDERED WITH MISSING DATA - Found <no value> placeholders",
			zap.Int("count", count),
			zap.String("preview", rendered[:min(500, len(rendered))]),
		)

		// Find which fields are missing by checking context around <no value>
		// This helps identify which template variables weren't populated
		parts := strings.Split(rendered, "<no value>")
		for i := 0; i < len(parts)-1 && i < 5; i++ {
			// Show context before each <no value>
			contextStart := len(parts[i]) - 50
			if contextStart < 0 {
				contextStart = 0
			}
			logger.Warn("Missing value context",
				zap.Int("occurrence", i+1),
				zap.String("before", parts[i][contextStart:]),
			)
		}
	}

	return rendered, nil
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
// Paths like "site_plan.validated_plan.needs_logo" will:
// 1. First try: data["site_plan"]["validated_plan"]["needs_logo"]
// 2. If not found, try: data["site_plan"]["response"]["validated_plan"]["needs_logo"]
//
// A path segment that is a bare non-negative integer indexes an array when the
// value at that point is a []interface{} — "search_results.results.0.url".
// Map access is tried FIRST and is unchanged, so a map carrying a literal "0"
// key resolves exactly as it always did; the array branch is only reachable
// where the walk previously returned nil. Bracket notation ("results[0].url")
// is NOT parsed here: the config-side convention in this fleet is the bare dot
// segment. (Bracket paths do exist as an EMITTED convention elsewhere — see
// walkContentDataLinks in content_data_links.go — but nothing feeds them back
// into this resolver.)
func ExtractNestedField(data map[string]interface{}, fieldPath string) interface{} {
	if fieldPath == "" {
		return nil
	}

	parts := strings.Split(fieldPath, ".")
	var current interface{} = data

	for _, part := range parts {
		// Array indexing: only when the current value is a slice and the
		// segment is a bare, in-range, non-negative integer. The two branches
		// are mutually exclusive BY TYPE — a slice can never satisfy the map
		// assertion below — so this ordering is a readability choice, not a
		// precedence rule, and saying otherwise overstates it. What actually
		// keeps a map with a literal "0" key resolving by key is that such a
		// map takes the branch below and never reaches this one.
		if arr, isArr := current.([]interface{}); isArr {
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(arr) {
				return nil
			}
			current = arr[idx]
			continue
		}

		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}

		// Try direct access first
		if val, exists := currentMap[part]; exists {
			current = val
			continue
		}

		// Auto-unwrap: try through .response (call_agent/spawn_agent wrapper)
		// This handles the case where the map has metadata at top level
		// and actual response data under "response"
		if response, hasResponse := currentMap["response"].(map[string]interface{}); hasResponse {
			if val, exists := response[part]; exists {
				current = val
				continue
			}
		}

		// Part not found
		return nil
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

// ExtractNestedFieldInt is a convenience wrapper that returns 0 if the
// field is missing or not numeric. Handles all three numeric types that
// can appear in a map[string]interface{}: float64 (JSON unmarshal default),
// int (Go-native), and int64.
func ExtractNestedFieldInt(data map[string]interface{}, fieldPath string) int {
	value := ExtractNestedField(data, fieldPath)
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
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
// Handles any language identifier (```json, ```html, ```css, etc.)
func CleanHTMLString(s string) string {
	// Remove markdown code blocks with any language identifier
	// Matches: ```json, ```html, ```css, ```, etc.
	codeBlockRe := regexp.MustCompile("```(?:\\w*)?\\s*\n?([\\s\\S]*?)```")
	s = codeBlockRe.ReplaceAllString(s, "$1")

	// Trim leading/trailing whitespace
	s = strings.TrimSpace(s)

	return s
}

func ExtractHTMLFromResponse(result interface{}) string {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return ""
	}

	response, _ := resultMap["result"].(string)

	// Remove markdown code blocks
	htmlBlockRe := regexp.MustCompile("```html\\s*([\\s\\S]*?)```")
	if matches := htmlBlockRe.FindStringSubmatch(response); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Look for DOCTYPE or html tag
	if strings.Contains(response, "<!DOCTYPE") || strings.Contains(response, "<html") {
		startIdx := strings.Index(response, "<!DOCTYPE")
		if startIdx == -1 {
			startIdx = strings.Index(response, "<html")
		}
		if startIdx >= 0 {
			return response[startIdx:]
		}
	}

	return response
}

func ExtractBusinessInfo(collectedData map[string]interface{}) map[string]interface{} {
	info := make(map[string]interface{})

	// Try to get from input data (checking for response field)
	if inputStep, ok := collectedData["input_data"]; ok {
		extracted := ExtractStepData(inputStep)
		if inputData, ok := extracted.(map[string]interface{}); ok {
			if businessName, ok := inputData["business_name"].(string); ok {
				info["business_name"] = businessName
			}
			if domain, ok := inputData["domain"].(string); ok {
				info["domain"] = domain
			}
			if desc, ok := inputData["description"].(string); ok {
				info["description"] = desc
			}
		}
	}

	// Try to get from headers (checking for response field)
	if headersStep, ok := collectedData["headers"]; ok {
		extracted := ExtractStepData(headersStep)
		if headers, ok := extracted.(map[string]interface{}); ok {
			if clientID, ok := headers["client_id"].(string); ok {
				info["client_id"] = clientID
			}
		}
	}

	return info
}

func BuildHTMLPrompt(context map[string]interface{}, agentConfig map[string]interface{}) string {
	contextJSON, _ := json.MarshalIndent(context, "", "  ")

	return fmt.Sprintf(`Generate a complete, modern, responsive HTML website based on the following context:

%s

Requirements:
1. Create a complete HTML5 document with proper structure
2. Include inline CSS for styling (modern, clean design)
3. Make it fully responsive with mobile-first approach
4. Include proper meta tags for SEO
5. Use semantic HTML elements
6. Include a navigation menu
7. Create sections based on the site structure provided
8. Make it production-ready

Output only the HTML code, starting with <!DOCTYPE html>.`, string(contextJSON))
}

func EnsureHTMLStructure(doc *goquery.Document) {
	// Ensure DOCTYPE (goquery doesn't preserve it, we'll add it back later)

	// Ensure html element has lang attribute
	html := doc.Find("html")
	if html.Length() > 0 {
		if lang, exists := html.Attr("lang"); !exists || lang == "" {
			html.SetAttr("lang", "en")
		}
	}

	// Ensure head exists
	if doc.Find("head").Length() == 0 {
		doc.Find("html").PrependHtml("<head></head>")
	}

	// Ensure body exists
	if doc.Find("body").Length() == 0 {
		doc.Find("html").AppendHtml("<body></body>")
	}
}

func AddMetaTags(doc *goquery.Document, businessInfo map[string]interface{}) {
	head := doc.Find("head")

	// Charset
	if doc.Find("meta[charset]").Length() == 0 {
		head.PrependHtml(`<meta charset="UTF-8">`)
	}

	// Viewport
	if doc.Find("meta[name='viewport']").Length() == 0 {
		head.AppendHtml(`<meta name="viewport" content="width=device-width, initial-scale=1.0">`)
	}

	// Description
	if desc, ok := businessInfo["description"].(string); ok && desc != "" {
		if doc.Find("meta[name='description']").Length() == 0 {
			head.AppendHtml(fmt.Sprintf(`<meta name="description" content="%s">`, desc))
		}
	}

	// Open Graph
	if name, ok := businessInfo["business_name"].(string); ok && name != "" {
		head.AppendHtml(fmt.Sprintf(`<meta property="og:title" content="%s">`, name))
		head.AppendHtml(`<meta property="og:type" content="website">`)

		if desc, ok := businessInfo["description"].(string); ok {
			head.AppendHtml(fmt.Sprintf(`<meta property="og:description" content="%s">`, desc))
		}
	}
}

func EnsureResponsiveDesign(doc *goquery.Document) {
	// Check if viewport meta exists
	if doc.Find("meta[name='viewport']").Length() == 0 {
		doc.Find("head").AppendHtml(`<meta name="viewport" content="width=device-width, initial-scale=1.0">`)
	}

	// Add responsive CSS if not present
	style := doc.Find("style")
	if style.Length() == 0 {
		doc.Find("head").AppendHtml(`
        <style>
            * { box-sizing: border-box; margin: 0; padding: 0; }
            body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; line-height: 1.6; }
            img { max-width: 100%; height: auto; }
            .container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
            @media (max-width: 768px) {
                .container { padding: 0 15px; }
            }
        </style>`)
	}
}

func OptimizeImages(doc *goquery.Document) {
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		// Add loading="lazy" for images not in viewport
		if loading, exists := s.Attr("loading"); !exists || loading == "" {
			s.SetAttr("loading", "lazy")
		}

		// Ensure images have dimensions to prevent layout shift
		if _, exists := s.Attr("width"); !exists {
			s.SetAttr("width", "auto")
		}
		if _, exists := s.Attr("height"); !exists {
			s.SetAttr("height", "auto")
		}
	})
}

func AddStructuredData(doc *goquery.Document, businessInfo map[string]interface{}) {
	if name, ok := businessInfo["business_name"].(string); ok {
		structuredData := map[string]interface{}{
			"@context": "https://schema.org",
			"@type":    "Organization",
			"name":     name,
		}

		if desc, ok := businessInfo["description"].(string); ok {
			structuredData["description"] = desc
		}

		if domain, ok := businessInfo["domain"].(string); ok {
			structuredData["url"] = "https://" + domain
		}

		jsonLD, _ := json.Marshal(structuredData)
		doc.Find("head").AppendHtml(fmt.Sprintf(`<script type="application/ld+json">%s</script>`, string(jsonLD)))
	}
}

func MinifyHTML(htmlContent string, logger *zap.Logger) string {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)

	minified, err := m.String("text/html", htmlContent)
	if err != nil {
		logger.Warn("Failed to minify HTML", zap.Error(err))
		return htmlContent
	}

	// Ensure DOCTYPE is preserved
	if !strings.HasPrefix(minified, "<!DOCTYPE") {
		minified = "<!DOCTYPE html>" + minified
	}

	return minified
}

func CountElements(doc *goquery.Document) map[string]int {
	return map[string]int{
		"images":   doc.Find("img").Length(),
		"links":    doc.Find("a").Length(),
		"headings": doc.Find("h1,h2,h3,h4,h5,h6").Length(),
		"sections": doc.Find("section,article,main,aside").Length(),
	}
}

// ExtractStepData checks if stepData contains a response field and extracts it.
// This is used throughout actions to handle both direct data and step responses.
func ExtractStepData(stepData interface{}) interface{} {
	if stepMap, ok := stepData.(map[string]interface{}); ok {
		// Check if this step has a response field
		if response, hasResponse := stepMap["response"]; hasResponse {
			return response
		}
		// No response field, return the step data itself
		return stepMap
	}
	// Not a map, return as-is
	return stepData
}

func GetStringField(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key].(string); ok && val != "" {
		return val
	}
	return defaultVal
}

func GetIntField(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	if val, ok := m[key].(int); ok {
		return val
	}
	return defaultVal
}

func GetBoolField(m map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return defaultVal
}

func ExtractSectionNamesHelper(sectionsRaw interface{}) []string {
	var names []string
	switch sections := sectionsRaw.(type) {
	case []interface{}:
		for _, s := range sections {
			switch v := s.(type) {
			case string:
				names = append(names, v)
			case map[string]interface{}:
				for _, key := range []string{"name", "type", "component", "component_name"} {
					if name, ok := v[key].(string); ok && name != "" {
						names = append(names, name)
						break
					}
				}
			}
		}
	case []string:
		names = sections
	}
	return names
}

// ExtractStringListHelper coerces a value to []string. It accepts the two decoded
// list shapes ([]interface{} of strings, []string) and — see below — a []byte or
// string holding a JSON array.
//
// bugs_open/174: the JSON-text arms are not a convenience. A jsonb column does NOT
// survive the orchestration data path as a list. QueryDatabaseAction scans every
// column into interface{} and stringifies any []byte it gets back
// (database_actions.go, "if b, ok := values[i].([]byte)"), so `spec->'seed_scope'`
// arrives in collected_data as the STRING `["a","b"]`, not as a slice. Nothing in
// input_mapping re-types it — ResolveInputMapping passes values through unchanged —
// so it reaches the action still a string. Before this change every such value
// returned nil here: indistinguishable from "the caller supplied nothing", which is
// what made 174's lost seed_scope invisible for three lanes rather than an error.
//
// This is deliberately WIDENING only. A string previously yielded nil, so no
// existing caller can be relying on a different answer; a string that is not a JSON
// array still yields nil. It is also why the fix is correct whichever shape the
// driver hands back — pgx may decode a jsonb column or may not, and neither the
// callers nor this helper have to know which.
func ExtractStringListHelper(val interface{}) []string {
	var result []string
	switch v := val.(type) {
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	case []string:
		result = v
	case []byte:
		return stringListFromJSON(v)
	case string:
		return stringListFromJSON([]byte(v))
	}
	return result
}

// stringListFromJSON decodes a JSON array of strings. Anything else — a scalar, an
// object, malformed text, a JSON array of non-strings — yields nil, matching what
// this helper has always returned for a value it could not read as a list.
//
// Decodes through SafeUnmarshal (same package) rather than calling
// json.Unmarshal directly: that helper already exists precisely to attempt an
// optional parse without propagating an error, which is exactly this contract.
// The council's reuse seat raised this on corr 081d98b3 and was right — it was
// added here as a private json.Unmarshal first.
//
// The leading-'[' guard is kept in front of it. SafeUnmarshal into
// []interface{} would reject a scalar or object anyway, so this is not
// load-bearing for correctness; it states the contract at the top of the
// function and keeps the overwhelmingly common case (an ordinary non-JSON
// string, e.g. every config value the other four callers pass) off the decode
// path entirely.
func stringListFromJSON(raw []byte) []string {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil
	}
	var decoded []interface{}
	if !SafeUnmarshal(trimmed, &decoded) {
		return nil
	}
	var result []string
	for _, item := range decoded {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// Helper function to convert interface{} to float64
func ToFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case float32:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func ToInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int:
		return int64(val), true
	case int64:
		return val, true
	case int32:
		return int64(val), true
	case float32:
		return int64(val), true
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func NullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// FindResultsArray searches for a results array in various common locations
func FindResultsArray(collectedData map[string]interface{}, basePath string, logger *zap.Logger) []interface{} {
	// Common paths to try, in order of preference
	pathsToTry := []string{
		basePath + ".results",      // search_results.results (flattened format)
		basePath + ".data.results", // search_results.data.results (wrapped format)
		basePath,                   // search_results (if it's directly an array)
		basePath + ".data",         // search_results.data (if data is the array)
		basePath + ".items",        // search_results.items (alternative naming)
	}

	for _, path := range pathsToTry {
		extracted := ExtractNestedField(collectedData, path)
		if extracted == nil {
			continue
		}

		// Direct array
		if arr, ok := extracted.([]interface{}); ok {
			logger.Debug("findResultsArray: Found array at path",
				zap.String("path", path),
				zap.Int("count", len(arr)))
			return arr
		}

		// Map that might contain results
		if m, ok := extracted.(map[string]interface{}); ok {
			for _, key := range []string{"results", "items", "data"} {
				if inner, exists := m[key]; exists {
					if arr, ok := inner.([]interface{}); ok {
						logger.Debug("findResultsArray: Found array in map",
							zap.String("path", path),
							zap.String("key", key),
							zap.Int("count", len(arr)))
						return arr
					}
				}
			}
		}
	}

	logger.Warn("findResultsArray: No array found",
		zap.String("basePath", basePath),
		zap.Strings("tried_paths", pathsToTry))

	return nil
}

func ToStringSlice(items []interface{}) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// InterfaceToString converts various types to string for template substitution
func InterfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case bool:
		return strconv.FormatBool(v)
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, InterfaceToString(item))
		}
		return strings.Join(parts, ", ")
	case []string:
		return strings.Join(v, ", ")
	case map[string]interface{}:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

type ErrorDetails struct {
	Message string
	Code    string
	Step    string
	Action  string
}

// extractErrorDetails parses error information from message body
func ExtractErrorDetails(body []byte) ErrorDetails {
	details := ErrorDetails{
		Message: "(unable to parse error)",
		Code:    "",
	}

	if len(body) == 0 {
		details.Message = "(empty error body)"
		return details
	}

	// Try to parse as standard response format
	var response struct {
		Body struct {
			Error  string `json:"error"`
			Status string `json:"status"`
		} `json:"body"`
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err == nil {
		// Check nested body.error first (common format)
		if response.Body.Error != "" {
			details.Message = response.Body.Error
		}
		// Check top-level error object
		if response.Error.Message != "" {
			details.Message = response.Error.Message
			details.Code = response.Error.Code
		}
	}

	// If still no message, try simpler format
	if details.Message == "(unable to parse error)" {
		var simple map[string]interface{}
		if err := json.Unmarshal(body, &simple); err == nil {
			if errMsg, ok := simple["error"].(string); ok {
				details.Message = errMsg
			}
			if errMap, ok := simple["error"].(map[string]interface{}); ok {
				if msg, ok := errMap["message"].(string); ok {
					details.Message = msg
				}
			}
		}
	}

	// Truncate very long error messages
	if len(details.Message) > 500 {
		details.Message = details.Message[:500] + "..."
	}

	return details
}

// templateToJSON serialises any value as indented JSON for use in prompt templates.
// Usage in templates: {{.some_map | toJSON}}
// Falls back to Go default formatting if JSON marshalling fails.
// Strings pass through unchanged to avoid double-encoding.
func templateToJSON(v interface{}) string {
	if v == nil {
		return "null"
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// NormaliseToKebab ensures a string is valid kebab-case for the function column.
func NormaliseToKebab(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == '-' || r == '_' || r == ' ' {
			return '-'
		}
		return -1
	}, s)
	// Remove double hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}

// BuildSemanticTags generates initial semantic tags from section_type and site_type
func BuildSemanticTags(sectionType, siteType string) string {
	tags := []string{}

	// Add section type parts as tags
	for _, part := range strings.Split(sectionType, "-") {
		if part != "" {
			tags = append(tags, part)
		}
	}

	// Add site type if provided
	if siteType != "" {
		tags = append(tags, siteType)
	}

	// Add "generated" provenance tag
	tags = append(tags, "generated")

	tagsJSON, _ := json.Marshal(tags)
	return string(tagsJSON)
}

// FunctionToDisplayName converts "spark-provocation-card" to "Spark Provocation Card"
func FunctionToDisplayName(function string) string {
	parts := strings.Split(function, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = UpperFirst(p)
		}
	}
	return strings.Join(parts, " ")
}

// templatePlaceholder emits a literal Go-template placeholder string for
// use in prompts that need to illustrate template syntax to an LLM.
// Usage in a prompt: {{placeholder "field_name"}}
// Rendered output: {{.field_name}}
//
// Without this helper, literal {{.x}} in a prompt gets parsed as an
// action by RenderPromptTemplate and resolves to "<no value>" at
// execute time, which teaches the LLM the wrong syntax. This helper
// lets prompt authors embed the target syntax for the LLM to copy.
func templatePlaceholder(name string) string {
	return "{{." + name + "}}"
}

// templateRangeStart emits a literal Go-template {{range .field}} action
// for use in prompts that illustrate iteration syntax to an LLM. Without
// this helper, a literal {{range .items}} in a prompt would be parsed as
// an action by RenderPromptTemplate, which would then attempt to range
// over .items in the prompt's data context and fail with "missing value
// for range" when there is no such field. Pair with templateRangeEnd.
//
// Usage in a prompt: {{rangeStart "items"}}
// Rendered output:   {{range .items}}
func templateRangeStart(field string) string {
	return "{{range ." + field + "}}"
}

// templateRangeEnd emits a literal Go-template {{end}} action.
// Companion to templateRangeStart.
//
// Usage in a prompt: {{rangeEnd}}
// Rendered output:   {{end}}
func templateRangeEnd() string {
	return "{{end}}"
}
