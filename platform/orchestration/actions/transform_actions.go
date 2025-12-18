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

// validateParsedResult checks if the parsed JSON has a valid structure
// Supports multiple formats:
//   - v1 (landing page): requires "sections" (array) + "component_details" (map)
//   - v2 (multipage): requires "pages" (array of page objects with components)
func validateParsedResult(result interface{}, logger *zap.Logger) error {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{}, got %T", result)
	}

	availableKeys := getMapKeys(resultMap)
	logger.Info("Validating parsed result", zap.Strings("available_keys", availableKeys))

	// Check for v2 format first (pages-based multipage structure)
	if pages, hasPages := resultMap["pages"]; hasPages {
		return validateV2Format(pages, logger)
	}

	// Check for v1 format (sections + component_details for landing pages)
	_, hasSections := resultMap["sections"]
	_, hasComponentDetails := resultMap["component_details"]

	if hasSections && hasComponentDetails {
		return validateV1Format(resultMap, logger)
	}

	// Neither format matched
	logger.Error("Parsed result doesn't match v1 or v2 format",
		zap.Strings("available_keys", availableKeys),
		zap.Bool("has_pages", false),
		zap.Bool("has_sections", hasSections),
		zap.Bool("has_component_details", hasComponentDetails),
	)

	return fmt.Errorf("invalid structure: need either 'pages' (v2) or 'sections'+'component_details' (v1) - got keys: %v", availableKeys)
}

// validateV1Format validates the landing page format (sections + component_details)
func validateV1Format(resultMap map[string]interface{}, logger *zap.Logger) error {
	logger.Info("Validating v1 format (sections + component_details)")

	// Validate sections is a non-empty array
	sections, ok := resultMap["sections"].([]interface{})
	if !ok {
		return fmt.Errorf("'sections' is not an array")
	}
	if len(sections) == 0 {
		return fmt.Errorf("'sections' array is empty")
	}
	logger.Info("Validated sections array", zap.Int("count", len(sections)))

	// Validate component_details is a non-empty map
	details, ok := resultMap["component_details"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("'component_details' is not a map")
	}
	if len(details) == 0 {
		return fmt.Errorf("'component_details' map is empty")
	}
	logger.Info("Validated component_details map", zap.Int("count", len(details)))

	return nil
}

// validateV2Format validates the multipage format (pages array)
func validateV2Format(pages interface{}, logger *zap.Logger) error {
	logger.Info("Validating v2 format (pages array)")

	pagesArray, ok := pages.([]interface{})
	if !ok {
		return fmt.Errorf("'pages' is not an array")
	}

	if len(pagesArray) == 0 {
		return fmt.Errorf("'pages' array is empty")
	}

	// Validate each page has required fields
	for i, page := range pagesArray {
		pageMap, ok := page.(map[string]interface{})
		if !ok {
			return fmt.Errorf("page[%d] is not a map", i)
		}

		// Each page must have a name
		name, hasName := pageMap["name"].(string)
		if !hasName || name == "" {
			return fmt.Errorf("page[%d] missing 'name' field", i)
		}

		// Each page should have components (but we'll be lenient here)
		if components, hasComponents := pageMap["components"]; hasComponents {
			if compArray, ok := components.([]interface{}); ok {
				logger.Debug("Page has components",
					zap.String("page", name),
					zap.Int("component_count", len(compArray)),
				)
			}
		}
	}

	logger.Info("Validated pages array",
		zap.Int("page_count", len(pagesArray)),
	)

	return nil
}
