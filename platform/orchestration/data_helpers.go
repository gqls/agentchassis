// internal/backend/agent-chassis/platform/orchestration/data_helpers.go
package orchestration

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ============================================================================
// NORMALIZATION FUNCTIONS
// These are called when building or modifying CollectedData
// ============================================================================

// NormalizeInputData extracts clean input_data from any source structure
// This handles all the messy nested cases and returns a clean map
func NormalizeInputData(source map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if source == nil {
		logger.Debug("NormalizeInputData: source is nil, returning empty map")
		return map[string]interface{}{}
	}

	// CASE 1: Try body.input_data first (messages from parent to child)
	if body, ok := source["body"].(map[string]interface{}); ok {
		if data, ok := body["input_data"].(map[string]interface{}); ok {
			logger.Debug("NormalizeInputData: extracted from body.input_data")
			return data
		}
	}

	// CASE 2: Try direct input_data
	if data, ok := source["input_data"].(map[string]interface{}); ok {
		// Check if it's already clean (doesn't contain action/config/input_data)
		_, hasAction := data["action"]
		_, hasConfig := data["config"]
		_, hasInputData := data["input_data"]

		// If it has none of these system fields, it's clean user data
		if !hasAction && !hasConfig && !hasInputData {
			logger.Debug("NormalizeInputData: using clean input_data")
			return data
		}

		// CASE 3: It's nested - extract the real input_data
		if nestedData, ok := data["input_data"].(map[string]interface{}); ok {
			logger.Debug("NormalizeInputData: extracted from input_data.input_data")
			return nestedData
		}

		// CASE 4: It has system fields but no nested input_data
		// This shouldn't happen but handle gracefully - extract only non-system fields
		cleaned := make(map[string]interface{})
		for k, v := range data {
			if k != "action" && k != "config" && k != "agent_config" && k != "agent_group" {
				cleaned[k] = v
			}
		}
		if len(cleaned) > 0 {
			logger.Debug("NormalizeInputData: cleaned system fields from input_data")
			return cleaned
		}
	}

	// CASE 5: No input_data found anywhere
	logger.Debug("NormalizeInputData: no input_data found, returning empty map")
	return map[string]interface{}{}
}

// NormalizeCollectedData builds a clean CollectedData structure from a message
// This is the ONLY function that should build CollectedData from messages
func NormalizeCollectedData(
	message map[string]interface{},
	execCtx interface{},
	requestsTopic string,
	logger *zap.Logger,
) map[string]interface{} {

	logger.Debug("NormalizeCollectedData: building clean structure")

	// Start with system metadata
	collectedData := map[string]interface{}{
		"__execution_context__": execCtx,
		"__my_requests_topic__": requestsTopic,
		"__raw_message__":       message, // Keep for debugging, but never use for data access
	}

	// Extract action
	if action, ok := message["action"].(string); ok {
		collectedData["action"] = action
		logger.Debug("NormalizeCollectedData: extracted action", zap.String("action", action))
	}

	// Extract config (only if it's workflow/system config, not user data)
	if config, ok := message["config"].(map[string]interface{}); ok {
		// Check if this looks like system config
		_, hasWorkflow := config["workflow"]
		_, hasGroupType := config["group_type"]

		if hasWorkflow || hasGroupType {
			collectedData["config"] = config
			logger.Debug("NormalizeCollectedData: extracted system config")
		}
	}

	// CRITICAL: Normalize input_data to top level
	// This is the user/business data that flows through the system
	collectedData["input_data"] = NormalizeInputData(message, logger)

	// Extract other standard fields
	if agentConfig, ok := message["agent_config"].(map[string]interface{}); ok {
		collectedData["agent_config"] = agentConfig
		logger.Debug("NormalizeCollectedData: extracted agent_config")
	}

	if agentGroup, ok := message["agent_group"].(map[string]interface{}); ok {
		collectedData["agent_group"] = agentGroup
		logger.Debug("NormalizeCollectedData: extracted agent_group")
	}

	if agentType, ok := message["agent_type"].(string); ok {
		collectedData["agent_type"] = agentType
		logger.Debug("NormalizeCollectedData: extracted agent_type", zap.String("type", agentType))
	}

	if prompt, ok := message["prompt"].(string); ok {
		collectedData["prompt"] = prompt
		logger.Debug("NormalizeCollectedData: extracted prompt")
	}

	logger.Debug("NormalizeCollectedData: structure complete",
		zap.Int("fields", len(collectedData)))

	return collectedData
}

// NormalizeResponseData cleans response data before storing in CollectedData
// Called when a child agent returns a response
func NormalizeResponseData(responseBody map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if responseBody == nil {
		logger.Debug("NormalizeResponseData: responseBody is nil, returning empty map")
		return map[string]interface{}{}
	}

	// CASE 1: Response has explicit "data" field
	if data, ok := responseBody["data"].(map[string]interface{}); ok {
		logger.Debug("NormalizeResponseData: using response.body.data")
		return data
	}

	// CASE 2: Response has explicit "result" field
	if result, ok := responseBody["result"].(map[string]interface{}); ok {
		logger.Debug("NormalizeResponseData: using response.body.result")
		return result
	}

	// CASE 3: Clean the body by removing system fields
	cleaned := make(map[string]interface{})
	systemFields := map[string]bool{
		"action":                true,
		"config":                true,
		"__execution_context__": true,
		"__raw_message__":       true,
		"__my_requests_topic__": true,
		"agent_config":          true,
		"agent_group":           true,
	}

	for k, v := range responseBody {
		if !systemFields[k] {
			cleaned[k] = v
		}
	}

	logger.Debug("NormalizeResponseData: cleaned response body",
		zap.Int("original_fields", len(responseBody)),
		zap.Int("cleaned_fields", len(cleaned)))

	return cleaned
}

// ============================================================================
// SAFE ACCESS FUNCTIONS
// ALL actions should use these instead of direct CollectedData access
// ============================================================================

// GetInputData safely retrieves input_data from CollectedData
// This is the primary way to access user/business data
func GetInputData(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	if collectedData == nil {
		logger.Warn("GetInputData: collectedData is nil")
		return map[string]interface{}{}
	}

	// After normalization, input_data should always be at top level
	if data, ok := collectedData["input_data"].(map[string]interface{}); ok {
		logger.Debug("GetInputData: found input_data at top level")
		return data
	}

	// This shouldn't happen after normalization
	logger.Warn("GetInputData: input_data not found at top level, attempting normalization")
	return NormalizeInputData(collectedData, logger)
}

// GetStepData safely retrieves data from a specific step's response
// Used when one step needs data from a previous step
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
// Used by aggregation actions that combine results from multiple steps
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
		} else {
			logger.Warn("GetMultipleStepData: step data not found", zap.String("step", stepName))
			// Store nil or skip? Let's skip missing steps
		}
	}

	logger.Debug("GetMultipleStepData: collection complete",
		zap.Int("requested", len(stepNames)),
		zap.Int("found", len(result)))

	return result
}

// GetFieldFromPath retrieves a value using dot-notation path (e.g., "input_data.business_type")
// Used for template resolution and deep data access
func GetFieldFromPath(collectedData map[string]interface{}, path string, logger *zap.Logger) (interface{}, error) {
	if collectedData == nil {
		return nil, fmt.Errorf("collectedData is nil")
	}

	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	parts := strings.Split(path, ".")
	current := collectedData

	logger.Debug("GetFieldFromPath: navigating path",
		zap.String("path", path),
		zap.Int("depth", len(parts)))

	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty segment in path: %s", path)
		}

		// Last part - return the value
		if i == len(parts)-1 {
			if val, ok := current[part]; ok {
				logger.Debug("GetFieldFromPath: found value", zap.String("path", path))
				return val, nil
			}
			return nil, fmt.Errorf("field not found: %s", path)
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, fmt.Errorf("cannot navigate path at segment '%s' in path: %s", part, path)
		}
	}

	return nil, fmt.Errorf("unexpected end of path navigation")
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

// MergeInputData merges new input data with existing, useful for enrichment
// New data takes precedence over existing
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
