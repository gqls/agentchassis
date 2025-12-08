// FILE: platform/orchestration/datahelpers/deep_search.go
// Aggressive deep search for finding domain and objective anywhere in nested structures

package datahelpers

import (
	"encoding/json"
	"strings"

	"go.uber.org/zap"
)

// FindDomainAggressive searches everywhere for domain field
func FindDomainAggressive(data interface{}, logger *zap.Logger) string {
	result := findFieldAggressive(data, "domain", 0, logger)
	if result != "" {
		logger.Info("Found domain via aggressive search", zap.String("domain", result))
	}
	return result
}

// FindObjectiveAggressive searches everywhere for objective field
func FindObjectiveAggressive(data interface{}, logger *zap.Logger) string {
	result := findFieldAggressive(data, "objective", 0, logger)
	if result != "" {
		logger.Info("Found objective via aggressive search", zap.Int("length", len(result)))
	}
	return result
}

// findFieldAggressive recursively searches for a field with NO depth limit initially
func findFieldAggressive(data interface{}, fieldName string, depth int, logger *zap.Logger) string {
	// Only limit after we've gone really deep
	if depth > 20 {
		return ""
	}

	// Handle strings directly
	if str, ok := data.(string); ok {
		// If it's JSON, try to parse it
		if strings.HasPrefix(strings.TrimSpace(str), "{") || strings.HasPrefix(strings.TrimSpace(str), "[") {
			var parsed interface{}
			if err := json.Unmarshal([]byte(str), &parsed); err == nil {
				return findFieldAggressive(parsed, fieldName, depth+1, logger)
			}
		}
		return ""
	}

	// Handle maps
	if m, ok := data.(map[string]interface{}); ok {
		// Check direct field first
		if val, ok := m[fieldName]; ok {
			if str, ok := val.(string); ok && str != "" {
				logger.Debug("Found field in map",
					zap.String("field", fieldName),
					zap.Int("depth", depth),
					zap.String("value_preview", str[:min(50, len(str))]),
				)
				return str
			}
		}

		// Unwrap common patterns first
		if unwrapped := tryUnwrapMap(m, logger); unwrapped != nil {
			if result := findFieldAggressive(unwrapped, fieldName, depth+1, logger); result != "" {
				return result
			}
		}

		// Recurse into ALL values
		for key, val := range m {
			if result := findFieldAggressive(val, fieldName, depth+1, logger); result != "" {
				logger.Debug("Found field via recursion through key",
					zap.String("key", key),
					zap.Int("depth", depth),
				)
				return result
			}
		}
	}

	// Handle slices
	if slice, ok := data.([]interface{}); ok {
		for _, val := range slice {
			if result := findFieldAggressive(val, fieldName, depth+1, logger); result != "" {
				return result
			}
		}
	}

	return ""
}

// tryUnwrapMap attempts to unwrap common nesting patterns
func tryUnwrapMap(m map[string]interface{}, logger *zap.Logger) interface{} {
	// Pattern 1: {field}_result.result
	for key, val := range m {
		if strings.HasSuffix(key, "_result") {
			if resultMap, ok := val.(map[string]interface{}); ok {
				if result, hasResult := resultMap["result"]; hasResult {
					// Try to parse JSON
					if str, ok := result.(string); ok {
						str = strings.TrimSpace(str)
						str = strings.TrimPrefix(str, "```json\n")
						str = strings.TrimPrefix(str, "```json")
						str = strings.TrimPrefix(str, "```")
						str = strings.TrimSuffix(str, "\n```")
						str = strings.TrimSuffix(str, "```")
						str = strings.TrimSpace(str)

						var parsed interface{}
						if err := json.Unmarshal([]byte(str), &parsed); err == nil {
							return parsed
						}
					}
					return result
				}
			}
		}
	}

	// Pattern 2: Direct result field
	if result, ok := m["result"]; ok {
		if str, ok := result.(string); ok {
			str = strings.TrimSpace(str)
			if strings.HasPrefix(str, "{") || strings.HasPrefix(str, "[") {
				var parsed interface{}
				if err := json.Unmarshal([]byte(str), &parsed); err == nil {
					return parsed
				}
			}
		}
		return result
	}

	// Pattern 3: input_data field (common nesting)
	if inputData, ok := m["input_data"]; ok {
		return inputData
	}

	// Pattern 4: Single wrapper key
	if len(m) == 1 {
		for key, val := range m {
			if strings.HasSuffix(key, "_data") || strings.HasSuffix(key, "_output") {
				return val
			}
		}
	}

	return nil
}

// ExtractCoreInputData finds domain and objective and creates a clean input_data map
func ExtractCoreInputData(data interface{}, logger *zap.Logger) map[string]interface{} {
	inputData := make(map[string]interface{})

	domain := FindDomainAggressive(data, logger)
	if domain != "" {
		inputData["domain"] = domain
	}

	objective := FindObjectiveAggressive(data, logger)
	if objective != "" {
		inputData["objective"] = objective
	}

	// Also look for model field
	if model := findFieldAggressive(data, "model", 0, logger); model != "" {
		inputData["model"] = model
	}

	logger.Info("Extracted core input data",
		zap.Bool("has_domain", domain != ""),
		zap.Bool("has_objective", objective != ""),
		zap.Int("objective_length", len(objective)),
	)

	return inputData
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
