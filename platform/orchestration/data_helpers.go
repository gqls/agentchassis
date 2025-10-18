// internal/backend/agent-chassis/platform/orchestration/data_helpers.go
package orchestration

import (
	"fmt"
	"os"
	"strings"

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
	if msg == nil || msg.Body == nil {
		return map[string]interface{}{}
	}

	// Body could be a map or another type
	switch body := msg.Body.(type) {
	case map[string]interface{}:
		// Look for data field first
		if data, ok := body["data"].(map[string]interface{}); ok {
			logger.Debug("extractFromRequestMessage: found body.data")
			return data
		}
		// Look for input_data field
		if data, ok := body["input_data"].(map[string]interface{}); ok {
			logger.Debug("extractFromRequestMessage: found body.input_data")
			return data
		}
		// Return cleaned body
		return cleanDataMap(body)
	default:
		logger.Debug("extractFromRequestMessage: body is not a map")
		return map[string]interface{}{}
	}
}

// extractFromResponseMessage extracts data from a typed ResponseMessage
func extractFromResponseMessage(msg *types.ResponseMessage, logger *zap.Logger) map[string]interface{} {
	if msg == nil || msg.Body == nil {
		return map[string]interface{}{}
	}

	// Handle ResponseBody type
	if responseBody, ok := msg.Body.(types.ResponseBody); ok {
		// ResponseBody.Body contains the actual data
		if body, ok := responseBody.Body.(map[string]interface{}); ok {
			// Look for data field
			if data, ok := body["data"].(map[string]interface{}); ok {
				logger.Debug("extractFromResponseMessage: found body.body.data")
				return data
			}
			// Return the body itself if it's clean data
			return cleanDataMap(body)
		}
	}

	// Handle if Body is directly a map
	if body, ok := msg.Body.(map[string]interface{}); ok {
		if data, ok := body["data"].(map[string]interface{}); ok {
			return data
		}
		return cleanDataMap(body)
	}

	return map[string]interface{}{}
}

// extractFromRawMap handles raw map extraction (backward compatibility)
func extractFromRawMap(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Try different extraction strategies in order of preference

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
		cleaned := cleanDataMap(body)
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
		"action": action,
		"data":   data,
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
		Body:    map[string]interface{}{"data": responseData},
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
		"data": initialData,
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
func BuildCollectedData(
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
}

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
func NormalizeResponseData(responseBody types.ResponseBody, logger *zap.Logger) map[string]interface{} {
	if responseBody.Body == nil && responseBody.Error == nil {
		logger.Debug("NormalizeResponseData: responseBody is nil")
		return map[string]interface{}{}
	}

	// ResponseBody.Body contains the actual data
	if body, ok := responseBody.Body.(map[string]interface{}); ok {
		if data, ok := body["data"].(map[string]interface{}); ok {
			return data
		}
		return cleanDataMap(body)
	}

	return map[string]interface{}{}
}

// GetInputData safely retrieves input_data from CollectedData
func GetInputData(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if collectedData == nil {
		logger.Warn("GetInputData: collectedData is nil")
		return map[string]interface{}{}
	}

	// First try the normalized location
	if data, ok := collectedData["input_data"].(map[string]interface{}); ok {
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

// UpdateCollectedData safely updates CollectedData with new information
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
		cleanStepData := cleanDataMap(stepData)
		for k, v := range cleanStepData {
			currentData[k] = v
		}
		logger.Debug("UpdateCollectedData: merged step data",
			zap.String("step", stepName),
			zap.Int("merged_fields", len(cleanStepData)))
	}
}

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

// cleanDataMap removes system fields from a map
func cleanDataMap(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	systemFields := map[string]bool{
		"action":                     true,
		"config":                     true,
		"agent_config":               true,
		"agent_group":                true,
		"__execution_context__":      true,
		"__raw_message__":            true,
		"__my_requests_topic__":      true,
		"__parent_responses_topic__": true,
		"headers":                    true,
		"is_initialization":          true,
		"agent_info":                 true,
		"role":                       true,
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
	systemFields := []string{"action", "config", "agent_config", "__execution_context__"}
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
