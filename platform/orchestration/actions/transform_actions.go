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

	params.Logger.Info("Successfully extracted and parsed JSON",
		zap.String("source_field", sourceField),
		zap.String("result_type", fmt.Sprintf("%T", value)),
	)

	return value, nil
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
