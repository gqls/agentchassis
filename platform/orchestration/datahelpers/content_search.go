package datahelpers

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// FindHTML searches recursively for HTML content in nested data structures
// Returns the first HTML string found that starts with <!DOCTYPE or <html
func FindHTML(data interface{}, logger *zap.Logger) string {
	result := findContentRecursive(data, isHTML, 0, logger)
	if result != nil {
		if html, ok := result.(string); ok {
			return html
		}
	}
	return ""
}

// FindByPath extracts data at a specific path, with recursive unwrapping
// Path can be like "final_html.final_html" or just "final_html"
// FindByPath extracts data at a specific path, with recursive unwrapping
// Supports auto-unwrapping of call_agent/spawn_agent .response wrappers
//
// Path can be like "site_plan.validated_plan" or "final_html.final_html"
func FindByPath(data map[string]interface{}, path string, logger *zap.Logger) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for i, part := range parts {
		logger.Debug("Traversing path",
			zap.String("part", part),
			zap.Int("depth", i),
			zap.String("current_type", fmt.Sprintf("%T", current)),
		)

		switch v := current.(type) {
		case map[string]interface{}:
			// Try direct access first
			if val, ok := v[part]; ok {
				current = val
				continue
			}

			// Auto-unwrap: try through .response (call_agent/spawn_agent wrapper)
			if response, hasResponse := v["response"].(map[string]interface{}); hasResponse {
				if val, ok := response[part]; ok {
					logger.Debug("Found part via .response auto-unwrap",
						zap.String("part", part),
					)
					current = val
					continue
				}
			}

			// Try recursive unwrapping (existing behavior)
			unwrapped := UnwrapDeep(v, logger)
			if unwrappedMap, ok := unwrapped.(map[string]interface{}); ok {
				if val, ok := unwrappedMap[part]; ok {
					current = val
					continue
				}
			}

			// If path starts with "input_data." and we couldn't find it,
			// try the path without the prefix (data might be flattened)
			if i == 0 && part == "input_data" && len(parts) > 1 {
				// Skip input_data prefix and try remaining path directly
				remainingPath := strings.Join(parts[1:], ".")
				logger.Debug("Trying path without input_data prefix",
					zap.String("original_path", path),
					zap.String("trying", remainingPath),
				)
				if result := FindByPath(data, remainingPath, logger); result != nil {
					return result
				}
			}

			// Also try adding input_data prefix if not present
			// (data might be wrapped when we expect it flat)
			if i == 0 && part != "input_data" {
				if inputData, hasInputData := v["input_data"].(map[string]interface{}); hasInputData {
					// Try to find the entire path inside input_data
					if result := FindByPath(inputData, path, logger); result != nil {
						logger.Debug("Found via input_data wrapper",
							zap.String("path", path),
						)
						return result
					}
				}
			}

			logger.Warn("Path part not found",
				zap.String("part", part),
				zap.Strings("available_keys", getKeys(v)),
			)
			return nil

		default:
			logger.Warn("Cannot traverse non-map",
				zap.String("part", part),
				zap.String("type", fmt.Sprintf("%T", current)),
			)
			return nil
		}
	}

	// Final unwrapping
	return UnwrapDeep(current, logger)
}

// FindHTMLWithFallback tries path first, then content search
func FindHTMLWithFallback(data map[string]interface{}, path string, logger *zap.Logger) string {
	logger.Info("Searching for HTML",
		zap.String("primary_path", path),
	)

	// Try path first
	if path != "" {
		if result := FindByPath(data, path, logger); result != nil {
			if html, ok := result.(string); ok && isHTML(html) {
				logger.Info("Found HTML via path",
					zap.String("path", path),
					zap.Int("length", len(html)),
				)
				return html
			}
		}
	}

	// Fallback to content search
	logger.Info("Path failed, searching for HTML content recursively")
	html := FindHTML(data, logger)
	if html != "" {
		logger.Info("Found HTML via content search",
			zap.Int("length", len(html)),
		)
	} else {
		logger.Warn("No HTML found in data structure")
	}
	return html
}

// findContentRecursive searches for content matching a predicate
func findContentRecursive(data interface{}, predicate func(string) bool, depth int, logger *zap.Logger) interface{} {
	if depth > 15 {
		return nil
	}

	// Check if current data matches
	if str, ok := data.(string); ok {
		if predicate(str) {
			logger.Debug("Found matching content",
				zap.Int("depth", depth),
				zap.Int("length", len(str)),
			)
			return str
		}
		return nil
	}

	// Recurse into maps
	if m, ok := data.(map[string]interface{}); ok {
		for key, val := range m {
			if result := findContentRecursive(val, predicate, depth+1, logger); result != nil {
				logger.Debug("Found content in map",
					zap.String("key", key),
					zap.Int("depth", depth),
				)
				return result
			}
		}
	}

	// Recurse into slices
	if slice, ok := data.([]interface{}); ok {
		for i, val := range slice {
			if result := findContentRecursive(val, predicate, depth+1, logger); result != nil {
				logger.Debug("Found content in slice",
					zap.Int("index", i),
					zap.Int("depth", depth),
				)
				return result
			}
		}
	}

	return nil
}

// unwrapDeep recursively unwraps nested structures
func UnwrapDeep(data interface{}, logger *zap.Logger) interface{} {
	return unwrapRecursive(data, 0, logger)
}

// unwrapRecursive handles the actual unwrapping logic
func unwrapRecursive(data interface{}, depth int, logger *zap.Logger) interface{} {
	if depth > 10 {
		return data
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	// Pattern 1: {field}_result.result
	for key, val := range m {
		if strings.HasSuffix(key, "_result") {
			if resultMap, ok := val.(map[string]interface{}); ok {
				if result, hasResult := resultMap["result"]; hasResult {
					// Try to parse JSON string
					if parsed := tryParseJSON(result); parsed != nil {
						return unwrapRecursive(parsed, depth+1, logger)
					}
					return unwrapRecursive(result, depth+1, logger)
				}
			}
		}
	}

	// Pattern 2: Direct result field
	if result, hasResult := m["result"]; hasResult {
		if parsed := tryParseJSON(result); parsed != nil {
			return unwrapRecursive(parsed, depth+1, logger)
		}
		return unwrapRecursive(result, depth+1, logger)
	}

	// Pattern 3: Single wrapper key
	if len(m) == 1 {
		for key, val := range m {
			if strings.HasSuffix(key, "_data") || strings.HasSuffix(key, "_result") ||
				strings.HasSuffix(key, "_output") || strings.HasSuffix(key, "_response") {
				return unwrapRecursive(val, depth+1, logger)
			}
		}
	}

	// Pattern 4: call_agent/spawn_agent response wrapper
	// These have "response" alongside metadata keys like "request_id", "action_sent"
	if response, hasResponse := m["response"].(map[string]interface{}); hasResponse {
		// Check for call_agent metadata markers to confirm this is a wrapped response
		_, hasRequestID := m["request_id"]
		_, hasActionSent := m["action_sent"]
		_, hasAwaitResponse := m["await_response"]
		if hasRequestID || hasActionSent || hasAwaitResponse {
			logger.Debug("UnwrapDeep: unwrapping call_agent response wrapper")
			return unwrapRecursive(response, depth+1, logger)
		}
	}

	return data
}

// tryParseJSON attempts to parse a JSON string
func tryParseJSON(value interface{}) interface{} {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	// Store original for debugging
	// originalStr := str
	originalLen := len(str)
	// Remove markdown fences
	str = strings.TrimSpace(str)

	// Handle all fence variations
	for strings.HasPrefix(str, "```") {
		// Find end of first line
		newlineIdx := strings.Index(str, "\n")
		if newlineIdx > 0 {
			str = str[newlineIdx+1:] // Skip past ```json\n or ```\n
		} else {
			str = strings.TrimPrefix(str, "```")
		}
		str = strings.TrimSpace(str)
	}

	// Remove trailing fences
	for strings.HasSuffix(str, "```") {
		// Find start of last line
		lastNewline := strings.LastIndex(str, "\n```")
		if lastNewline > 0 {
			str = str[:lastNewline]
		} else {
			str = strings.TrimSuffix(str, "```")
		}
		str = strings.TrimSpace(str)
	}

	cleanedLen := len(str)

	var parsed interface{}
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		// Log the error with context for debugging
		fmt.Printf("JSON parsing failed: %v. Original length: %d, Cleaned length: %d, Preview: %s\n",
			err, originalLen, len(str), str[max(0, len(str)-100):])
		// LOG THE ACTUAL FAILURE
		fmt.Printf("JSON PARSE FAILED\n")
		fmt.Printf("  Original length: %d\n", originalLen)
		fmt.Printf("  Cleaned length: %d\n", cleanedLen)
		fmt.Printf("  Error: %v\n", err)
		fmt.Printf("  First 200 chars: %s\n", str[:min(200, len(str))])
		fmt.Printf("  Last 200 chars: %s\n", str[max(0, len(str)-200):])

		// Check WHY it failed
		if isTruncatedJSON(str) {
			fmt.Printf("  Cause: TRUNCATED JSON (unmatched brackets/braces)\n")
			fmt.Printf("  Last 300 chars: ...%s\n", str[max(0, len(str)-300):])
		} else {
			fmt.Printf("  Cause: SYNTAX ERROR (not truncation)\n")
			fmt.Printf("  First 200 chars: %s\n", str[:min(200, len(str))])
		}

		return nil
	}

	fmt.Printf("JSON parse success (original: %d bytes, cleaned: %d bytes)\n", originalLen, cleanedLen)
	return parsed
}

// isTruncatedJSON detects if a string appears to be truncated JSON
func isTruncatedJSON(str string) bool {
	trimmed := strings.TrimSpace(str)

	// Check for unmatched brackets/braces
	openBraces := strings.Count(trimmed, "{")
	closeBraces := strings.Count(trimmed, "}")
	openBrackets := strings.Count(trimmed, "[")
	closeBrackets := strings.Count(trimmed, "]")

	if openBraces != closeBraces || openBrackets != closeBrackets {
		return true
	}

	// Check if ends with incomplete structure
	if strings.HasSuffix(trimmed, ",") ||
		strings.HasSuffix(trimmed, ":") ||
		strings.HasSuffix(trimmed, "\":") {
		return true
	}

	return false
}

// isHTML checks if a string looks like HTML
func isHTML(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "<!DOCTYPE") ||
		strings.HasPrefix(s, "<!doctype") ||
		strings.HasPrefix(s, "<html") ||
		strings.HasPrefix(s, "<HTML")
}

// getKeys returns keys from a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
