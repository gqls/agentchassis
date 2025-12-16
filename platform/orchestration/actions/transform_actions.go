// platform/orchestration/actions/transform_actions.go
package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ParseJSONFieldAction extracts and parses JSON from a field in CollectedData
// Leverages existing datahelpers.FindByPath which already handles JSON parsing
func ParseJSONFieldAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ParseJSONFieldAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	config := params.StepConfig.Config

	// Get source field path
	sourceField, ok := config["source_field"].(string)
	if !ok || sourceField == "" {
		return nil, fmt.Errorf("parse_json_field action requires 'source_field' config")
	}

	params.Logger.Info("Parsing JSON from field",
		zap.String("source_field", sourceField),
	)

	// Use existing FindByPath which already:
	// 1. Navigates nested paths
	// 2. Unwraps common patterns (_result.result, etc.)
	// 3. Strips markdown code fences
	// 4. Parses JSON strings
	value := datahelpers.FindByPath(params.CollectedData, sourceField, params.Logger)

	if value == nil {
		return nil, fmt.Errorf("field '%s' not found or could not be parsed", sourceField)
	}

	params.Logger.Info("Found field value",
		zap.String("source_field", sourceField),
		zap.String("value_type", fmt.Sprintf("%T", value)),
	)

	// Use the public UnwrapDeep to handle all unwrapping patterns
	result := datahelpers.UnwrapDeep(value, params.Logger)

	if result == nil {
		// Check if it's a truncation issue
		if strVal, ok := value.(string); ok {
			return nil, fmt.Errorf("field '%s' contains invalid/truncated JSON - likely hit token limit. Original length: %d chars. Error: failed to parse",
				sourceField, len(strVal))
		}
		if mapVal, ok := value.(map[string]interface{}); ok {
			if resultStr, hasResult := mapVal["result"].(string); hasResult {
				return nil, fmt.Errorf("field '%s' contains invalid/truncated JSON in 'result' field - likely hit token limit. Result length: %d chars",
					sourceField, len(resultStr))
			}
		}
		return nil, fmt.Errorf("field '%s' could not be unwrapped/parsed", sourceField)
	}

	// Validate the parsed result has expected structure
	if err := validateParsedResult(result, params.Logger); err != nil {
		return nil, fmt.Errorf("field '%s' parsed but validation failed: %w - This may indicate truncated LLM output", sourceField, err)
	}

	params.Logger.Info("Successfully extracted and parsed JSON",
		zap.String("source_field", sourceField),
		zap.String("result_type", fmt.Sprintf("%T", result)),
	)

	return result, nil
}

// ExtractFieldAction extracts fields using the unified extractor
// This wraps datahelpers.ExtractFields
func ExtractFieldAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ExtractFieldAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	config := params.StepConfig.Config

	// Get field names to extract
	var fieldNames []string
	if names, ok := config["fields"].([]interface{}); ok {
		for _, name := range names {
			if str, ok := name.(string); ok {
				fieldNames = append(fieldNames, str)
			}
		}
	} else if name, ok := config["field"].(string); ok {
		// Single field
		fieldNames = []string{name}
	} else {
		return nil, fmt.Errorf("extract_field action requires 'fields' or 'field' config")
	}

	params.Logger.Info("Extracting fields using unified extractor",
		zap.Strings("fields", fieldNames),
	)

	// Use the MASTER EXTRACTOR from datahelpers
	result := datahelpers.ExtractFields(params.CollectedData, fieldNames, params.Logger)

	params.Logger.Info("Field extraction complete",
		zap.Int("fields_requested", len(fieldNames)),
		zap.Int("fields_extracted", len(result)),
	)

	return result, nil
}

// validateParsedResult checks if the parsed JSON has the expected structure
func validateParsedResult(result interface{}, logger *zap.Logger) error {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", result)
	}

	// Check for required top-level keys
	requiredKeys := []string{"sections", "component_details"}
	for _, key := range requiredKeys {
		if _, exists := resultMap[key]; !exists {
			logger.Error("Missing required key in parsed result",
				zap.String("missing_key", key),
				zap.Strings("available_keys", getMapKeys(resultMap)),
			)
			return fmt.Errorf("missing required key '%s' - JSON may be truncated", key)
		}
	}

	// Validate sections is an array
	if sections, ok := resultMap["sections"].([]interface{}); ok {
		if len(sections) == 0 {
			return fmt.Errorf("'sections' array is empty")
		}
		logger.Info("Validated sections array", zap.Int("count", len(sections)))
	} else {
		return fmt.Errorf("'sections' is not an array")
	}

	// Validate component_details is a map
	if details, ok := resultMap["component_details"].(map[string]interface{}); ok {
		if len(details) == 0 {
			return fmt.Errorf("'component_details' map is empty")
		}
		logger.Info("Validated component_details map", zap.Int("count", len(details)))
	} else {
		return fmt.Errorf("'component_details' is not a map")
	}

	return nil
}
