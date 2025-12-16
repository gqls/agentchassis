// FILE: platform/orchestration/datahelpers/unified_extractor.go
// THE MASTER EXTRACTOR - Uses all our helper functions together

package datahelpers

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ExtractFields is THE ONLY function you should use to extract fields
// It uses ALL our existing helpers in the right order
func ExtractFields(
	collectedData map[string]interface{},
	fieldNames []string,
	logger *zap.Logger,
) map[string]interface{} {

	logger.Info("=== MASTER EXTRACTOR START ===",
		zap.Strings("requested_fields", fieldNames),
		zap.Strings("available_keys", getMapKeys(collectedData)),
	)

	result := make(map[string]interface{})

	// Special handling for "input_data" field
	if contains(fieldNames, "input_data") {
		logger.Info("Special case: flattening input_data")

		// Use aggressive search to get core data
		coreData := ExtractCoreInputData(collectedData, logger)

		// Flatten it into result
		for k, v := range coreData {
			result[k] = v
			logger.Info("Flattened from core data",
				zap.String("key", k),
				zap.String("type", fmt.Sprintf("%T", v)),
			)
		}

		// Also try to get input_data map directly
		if inputMap := getInputDataMap(collectedData, logger); inputMap != nil {
			for k, v := range inputMap {
				if _, exists := result[k]; !exists {
					result[k] = v
					logger.Info("Added from input_data map",
						zap.String("key", k),
					)
				}
			}
		}
	}

	// Extract each specific field
	for _, fieldName := range fieldNames {
		if fieldName == "input_data" {
			continue // Already handled above
		}

		logger.Info(">>> Extracting field", zap.String("field", fieldName))

		value := extractSingleField(collectedData, fieldName, make(map[string]bool), logger)

		if value != nil {
			// Store with simple name (last part of path)
			parts := strings.Split(fieldName, ".")
			simpleKey := parts[len(parts)-1]
			result[simpleKey] = value

			logger.Info("✓ Field extracted",
				zap.String("requested", fieldName),
				zap.String("stored_as", simpleKey),
				zap.String("type", fmt.Sprintf("%T", value)),
			)
		} else {
			logger.Warn("✗ Field not found",
				zap.String("field", fieldName),
			)
		}
	}

	// CRITICAL: Always ensure domain and objective exist
	ensureCoreFields(result, collectedData, logger)

	logger.Info("=== MASTER EXTRACTOR COMPLETE ===",
		zap.Int("fields_extracted", len(result)),
		zap.Strings("result_keys", getMapKeys(result)),
	)

	return result
}

// extractSingleField tries multiple strategies to find ONE field
func extractSingleField(
	data map[string]interface{},
	fieldName string,
	seen map[string]bool,
	logger *zap.Logger,
) interface{} {

	logger.Debug("Trying extraction strategies", zap.String("field", fieldName))

	// Prevent infinite loops from circular aliases
	if seen[fieldName] {
		return nil // Already tried this field
	}
	seen[fieldName] = true

	// Strategy 1: Use FindByPath (handles unwrapping)
	if value := FindByPath(data, fieldName, logger); value != nil {
		logger.Info("Found via FindByPath", zap.String("field", fieldName))
		return value
	}

	// Strategy 2: Try with input_data prefix
	if !strings.HasPrefix(fieldName, "input_data.") {
		path := "input_data." + fieldName
		if value := FindByPath(data, path, logger); value != nil {
			logger.Info("Found via input_data prefix",
				zap.String("field", fieldName),
				zap.String("path", path))
			return value
		}
	}

	// Strategy 3: Look inside input_data map directly
	if inputMap := getInputDataMap(data, logger); inputMap != nil {
		if value, ok := inputMap[fieldName]; ok {
			logger.Info("Found in input_data map", zap.String("field", fieldName))
			return UnwrapDeep(value, logger)
		}
	}

	// Strategy 4: Aggressive recursive search
	logger.Info("Trying aggressive search", zap.String("field", fieldName))
	if value := findFieldRecursive(data, fieldName, 0, logger); value != nil {
		logger.Info("Found via aggressive search", zap.String("field", fieldName))
		return value
	}

	// Strategy 5: Check known aliases
	if alias := getFieldAlias(fieldName); alias != "" {
		logger.Info("Trying field alias",
			zap.String("field", fieldName),
			zap.String("alias", alias))
		return extractSingleField(data, alias, seen, logger)
	}

	return nil
}

// findFieldRecursive does aggressive recursive search
func findFieldRecursive(
	data interface{},
	fieldName string,
	depth int,
	logger *zap.Logger,
) interface{} {
	if depth > 20 {
		return nil
	}

	// Handle maps
	if m, ok := data.(map[string]interface{}); ok {
		// Direct match
		if val, ok := m[fieldName]; ok {
			logger.Debug("Recursive: found direct match",
				zap.String("field", fieldName),
				zap.Int("depth", depth),
			)
			return UnwrapDeep(val, logger)
		}

		// Try unwrapping first
		unwrapped := tryUnwrapMapPatterns(m, logger)
		if unwrapped != nil {
			if result := findFieldRecursive(unwrapped, fieldName, depth+1, logger); result != nil {
				return result
			}
		}

		// Recurse into all values
		for key, val := range m {
			if result := findFieldRecursive(val, fieldName, depth+1, logger); result != nil {
				logger.Debug("Recursive: found through key",
					zap.String("through_key", key),
					zap.Int("depth", depth),
				)
				return result
			}
		}
	}

	// Handle slices
	if slice, ok := data.([]interface{}); ok {
		for _, val := range slice {
			if result := findFieldRecursive(val, fieldName, depth+1, logger); result != nil {
				return result
			}
		}
	}

	return nil
}

// tryUnwrapMapPatterns tries to unwrap common nesting patterns
func tryUnwrapMapPatterns(m map[string]interface{}, logger *zap.Logger) interface{} {
	// Pattern 1: {field}_result.result
	for key, val := range m {
		if strings.HasSuffix(key, "_result") {
			if resultMap, ok := val.(map[string]interface{}); ok {
				if result, ok := resultMap["result"]; ok {
					if parsed := tryParseJSON(result); parsed != nil {
						return parsed
					}
					return result
				}
			}
		}
	}

	// Pattern 2: Direct result field
	if result, ok := m["result"]; ok {
		if parsed := tryParseJSON(result); parsed != nil {
			return parsed
		}
		return result
	}

	// Pattern 3: input_data field
	if inputData, ok := m["input_data"]; ok {
		return inputData
	}

	return nil
}

// ensureCoreFields makes absolutely sure domain and objective exist
func ensureCoreFields(
	result map[string]interface{},
	source map[string]interface{},
	logger *zap.Logger,
) {
	logger.Info("Ensuring core fields present")

	// Check domain
	if _, hasDomain := result["domain"]; !hasDomain {
		logger.Warn("Domain missing from result, searching aggressively")
		if domain := FindDomainAggressive(source, logger); domain != "" {
			result["domain"] = domain
			logger.Info("✓ Recovered domain via aggressive search", zap.String("domain", domain))
		} else {
			logger.Error("✗ Could not find domain anywhere")
		}
	}

	// Check objective
	if _, hasObjective := result["objective"]; !hasObjective {
		logger.Warn("Objective missing from result, searching aggressively")
		if objective := FindObjectiveAggressive(source, logger); objective != "" {
			result["objective"] = objective
			logger.Info("✓ Recovered objective via aggressive search",
				zap.Int("length", len(objective)))
		} else {
			logger.Error("✗ Could not find objective anywhere")
		}
	}

	// Check model
	if _, hasModel := result["model"]; !hasModel {
		if model := findFieldRecursive(source, "model", 0, logger); model != nil {
			if modelStr, ok := model.(string); ok {
				result["model"] = modelStr
				logger.Info("✓ Recovered model via aggressive search", zap.String("model", modelStr))
			}
		}
	}
}

// getInputDataMap extracts the input_data map if it exists
func getInputDataMap(data map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Try direct lookup
	if inputData, ok := data["input_data"].(map[string]interface{}); ok {
		// Check for double nesting
		if nested, ok := inputData["input_data"].(map[string]interface{}); ok {
			logger.Debug("Unwrapped double-nested input_data")
			return nested
		}
		return inputData
	}

	// Try unwrapping data first
	unwrapped := UnwrapDeep(data, logger)
	if unwrappedMap, ok := unwrapped.(map[string]interface{}); ok {
		if inputData, ok := unwrappedMap["input_data"].(map[string]interface{}); ok {
			return inputData
		}
	}

	return nil
}

// getFieldAlias returns alternative names for common fields
func getFieldAlias(fieldName string) string {
	aliases := map[string]string{
		"site_architecture":  "architecture",
		"architecture":       "site_architecture",
		"site_content":       "content",
		"content":            "site_content",
		"domain_analysis":    "analysis",
		"analysis":           "domain_analysis",
		"available_builders": "builders",
	}
	return aliases[fieldName]
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
