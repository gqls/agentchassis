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
					if parsed := tryParseJSON(result, logger); parsed != nil {
						return unwrapRecursive(parsed, depth+1, logger)
					}
					return unwrapRecursive(result, depth+1, logger)
				}
			}
		}
	}

	// Pattern 2: Direct result field
	if result, hasResult := m["result"]; hasResult {
		if parsed := tryParseJSON(result, logger); parsed != nil {
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

// StripCodeFences removes markdown code fences from LLM output.
// Handles all variations: ```json, ```html, ```css, plain ``` etc.
// Safe to call on content that has no fences (returns it unchanged).
//
// Used by:
//   - tryParseJSON (JSON fence stripping before parsing)
//   - CreateToolComponentAction (HTML fence stripping from LLM-generated tools)
//   - Any action receiving raw LLM text output that may be fenced
func StripCodeFences(s string) string {
	s = strings.TrimSpace(s)

	// Strip leading fences: ```html\n, ```json\n, ```\n, etc.
	for strings.HasPrefix(s, "```") {
		newlineIdx := strings.Index(s, "\n")
		if newlineIdx > 0 {
			s = s[newlineIdx+1:]
		} else {
			s = strings.TrimPrefix(s, "```")
		}
		s = strings.TrimSpace(s)
	}

	// Strip trailing fences
	for strings.HasSuffix(s, "```") {
		lastNewline := strings.LastIndex(s, "\n```")
		if lastNewline > 0 {
			s = s[:lastNewline]
		} else {
			s = strings.TrimSuffix(s, "```")
		}
		s = strings.TrimSpace(s)
	}

	return s
}

// tryParseJSON attempts to parse a JSON string
// Returns nil for non-JSON strings (this is expected, not an error)
func tryParseJSON(value interface{}, logger *zap.Logger) interface{} {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	originalLen := len(str)
	str = StripCodeFences(str)

	// Quick check: if it doesn't look like JSON, don't try to parse
	// This avoids noisy logs for plain text LLM responses
	if len(str) == 0 {
		return nil
	}
	looksLikeJSON := strings.HasPrefix(str, "{") || strings.HasPrefix(str, "[")
	if !looksLikeJSON {
		return nil
	}

	// Attempt to parse
	var parsed interface{}
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		// Only log for content that looked like JSON but failed
		if isTruncatedJSON(str) {
			logger.Warn("Truncated JSON detected",
				zap.Int("original_len", originalLen),
				zap.Int("cleaned_len", len(str)),
				zap.String("last_100_chars", str[max(0, len(str)-100):]),
			)
		} else {
			logger.Debug("JSON syntax error",
				zap.Error(err),
				zap.Int("length", len(str)),
				zap.String("preview", str[:min(100, len(str))]),
			)
		}
		return nil
	}

	logger.Debug("JSON parsed successfully",
		zap.Int("original_len", originalLen),
		zap.Int("cleaned_len", len(str)),
	)
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
