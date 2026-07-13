// FILE: debug_collected_data.go
// Helper function to log collected_data structure with truncated values
// Add to platform/orchestration/datahelpers/ or similar
//
// This helps debug path issues by showing the structure without overwhelming logs

package datahelpers

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// LogCollectedDataStructure logs the keys and structure of collected_data
// with truncated string values for debugging path issues
// prefix is optional - call with or without it
func LogCollectedDataStructure(data map[string]interface{}, logger *zap.Logger, prefix ...string) {
	p := ""
	if len(prefix) > 0 {
		p = prefix[0]
	}
	summary := summarizeMap(data, "", 0, 3) // max depth 3
	logger.Info("CollectedData structure",
		zap.String("prefix", p),
		zap.String("structure", summary),
	)
}

// GetCollectedDataKeys returns a summary of all keys and their types
func GetCollectedDataKeys(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	collectKeys(data, "", result)
	return result
}

// summarizeMap creates a readable summary of a map structure
func summarizeMap(data interface{}, path string, depth int, maxDepth int) string {
	if depth > maxDepth {
		return "..."
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		var parts []string
		for key, val := range v {
			keyPath := key
			if path != "" {
				keyPath = path + "." + key
			}
			valSummary := summarizeMap(val, keyPath, depth+1, maxDepth)
			parts = append(parts, fmt.Sprintf("%s: %s", key, valSummary))
		}
		return "{ " + strings.Join(parts, ", ") + " }"

	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		// Just show first element type and count
		firstType := fmt.Sprintf("%T", v[0])
		return fmt.Sprintf("[%s x%d]", firstType, len(v))

	case string:
		if len(v) > 50 {
			return fmt.Sprintf(`"%s..."[%d chars]`, v[:47], len(v))
		}
		return fmt.Sprintf(`"%s"`, v)

	case float64:
		return fmt.Sprintf("%v", v)

	case bool:
		return fmt.Sprintf("%v", v)

	case nil:
		return "null"

	default:
		return fmt.Sprintf("<%T>", v)
	}
}

// collectKeys recursively collects all key paths and their value types
func collectKeys(data interface{}, path string, result map[string]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			keyPath := key
			if path != "" {
				keyPath = path + "." + key
			}

			// Record the type at this path
			switch innerVal := val.(type) {
			case map[string]interface{}:
				result[keyPath] = "map"
				collectKeys(innerVal, keyPath, result)
			case []interface{}:
				if len(innerVal) > 0 {
					result[keyPath] = fmt.Sprintf("array[%T x%d]", innerVal[0], len(innerVal))
				} else {
					result[keyPath] = "array[empty]"
				}
			case string:
				result[keyPath] = fmt.Sprintf("string[%d]", len(innerVal))
			default:
				result[keyPath] = fmt.Sprintf("%T", val)
			}
		}
	}
}

// Example usage in coordinator.go or actions:
//
// Before executing an action:
//   datahelpers.LogCollectedDataStructure(state.CollectedData, logger, "before_"+stepName)
//
// Output example:
//   CollectedData structure: prefix="before_assemble_page"
//   structure="{ page_content: { page_html: "<!DOCTYPE html>..."[5234 chars] },
//                reviewed_content: { auto_eval_content: { result: { approved: false, ... } } },
//                current_page: { id: "...", slug: "index" } }"
