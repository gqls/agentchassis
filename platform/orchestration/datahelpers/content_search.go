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
			if val, ok := v[part]; ok {
				current = val
			} else {
				// Try recursive unwrapping
				unwrapped := unwrapDeep(v, logger)
				if unwrappedMap, ok := unwrapped.(map[string]interface{}); ok {
					if val, ok := unwrappedMap[part]; ok {
						current = val
						continue
					}
				}
				logger.Warn("Path part not found",
					zap.String("part", part),
					zap.Strings("available_keys", getKeys(v)),
				)
				return nil
			}
		default:
			logger.Warn("Cannot traverse non-map",
				zap.String("part", part),
				zap.String("type", fmt.Sprintf("%T", current)),
			)
			return nil
		}
	}

	// Final unwrapping
	return unwrapDeep(current, logger)
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
func unwrapDeep(data interface{}, logger *zap.Logger) interface{} {
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

	return data
}

// tryParseJSON attempts to parse a JSON string
func tryParseJSON(value interface{}) interface{} {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	// Remove markdown fences
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "```json\n")
	str = strings.TrimPrefix(str, "```json")
	str = strings.TrimPrefix(str, "```\n")
	str = strings.TrimPrefix(str, "```")
	str = strings.TrimSuffix(str, "\n```")
	str = strings.TrimSuffix(str, "```")
	str = strings.TrimSpace(str)

	var parsed interface{}
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		return nil
	}

	return parsed
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
