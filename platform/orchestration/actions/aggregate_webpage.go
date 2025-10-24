// FILE: platform/orchestration/actions/aggregate_webpage.go
package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AggregateWebpageAction combines content sections into a complete HTML webpage
func AggregateWebpageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("AggregateWebpageAction executing",
		zap.Any("DEBUGaa: AggregateWebpage params", params),
	)

	config := params.StepConfig.Config

	// Get response fields to aggregate
	responseFields := getResponseFields(config)
	if len(responseFields) == 0 {
		return nil, fmt.Errorf("response_fields is required for aggregate_webpage")
	}

	// Collect responses from the specified fields
	responses := collectResponses(params.CollectedData, responseFields, params.Logger)

	// Get configuration
	sectionOrder := getConfigStringSlice(config, "section_order", responseFields)
	wrapper := getConfigMap(config, "wrapper")
	addSectionTags := getConfigBool(config, "add_section_tags", true)

	params.Logger.Info("AggregateWebpageAction collected data",
		zap.Any("responseFields", responseFields),
		zap.Any("responses", responses),
		zap.Any("sectionOrder", sectionOrder),
		zap.Any("wrapper", wrapper),
		zap.Any("addSectionTags", addSectionTags),
	)
	// Extract head and foot with defaults
	htmlHead := getConfigString(wrapper, "html_head",
		"<!DOCTYPE html>\n<html lang='en'>\n<head>\n<meta charset='utf-8'>\n<meta name='viewport' content='width=device-width, initial-scale=1'>\n<title>Generated Website</title>\n</head>\n<body>")
	htmlFoot := getConfigString(wrapper, "html_foot", "</body>\n</html>")

	// Build sections in order
	var sections []string
	sectionsProcessed := 0

	for _, sectionKey := range sectionOrder {
		// Remove "generate_" prefix if present for cleaner section names
		sectionName := strings.TrimPrefix(sectionKey, "generate_")

		if response, ok := responses[sectionKey]; ok {
			content := extractResponseContent(response)
			params.Logger.Info("AggregateWebpageAction in loop",
				zap.Any("DEBUGaa: extracted content", content),
				zap.String("DEBUGaa: which was for section key", sectionKey),
			)

			if content == "" {
				params.Logger.Warn("Empty content for section",
					zap.String("section_key", sectionKey))
				continue
			}

			if addSectionTags {
				sections = append(sections,
					fmt.Sprintf("<section id='%s' class='%s-section'>\n%s\n</section>",
						sectionName, sectionName, content))
			} else {
				sections = append(sections, content)
			}

			sectionsProcessed++
			params.Logger.Info("Added section to webpage",
				zap.String("section", sectionName),
				zap.Int("content_length", len(content)))
		} else {
			params.Logger.Warn("Section not found in responses",
				zap.String("section_key", sectionKey))
		}
	}

	if sectionsProcessed == 0 {
		return nil, fmt.Errorf("no sections were successfully processed")
	}

	// Assemble final HTML
	assembledHTML := htmlHead + "\n\n" + strings.Join(sections, "\n\n") + "\n\n" + htmlFoot

	params.Logger.Info("Webpage assembled successfully",
		zap.Int("sections_processed", sectionsProcessed),
		zap.Int("total_html_length", len(assembledHTML)))

	return map[string]interface{}{
		"timestamp":            time.Now(),
		"aggregation_strategy": "webpage",
		"assembled_html":       assembledHTML,
		"sections":             responses,
		"section_count":        sectionsProcessed,
		"total_length":         len(assembledHTML),
		"success":              true,
	}, nil
}

// Helper: Get response fields from config
func getResponseFields(config map[string]interface{}) []string {
	fields := []string{}
	if fieldsList, ok := config["response_fields"].([]interface{}); ok {
		for _, f := range fieldsList {
			if fieldName, ok := f.(string); ok {
				fields = append(fields, fieldName)
			}
		}
	}
	return fields
}

// collectResponses collects responses from CollectedData for aggregation
// It handles multiple storage patterns for compatibility
func collectResponses(collectedData map[string]interface{}, responseFields []string, logger *zap.Logger) map[string]interface{} {
	responses := make(map[string]interface{})

	for _, fieldName := range responseFields {
		logger.Debug("collectResponses: looking for field",
			zap.String("field", fieldName))

		// Check if field exists directly in CollectedData
		if stepData, ok := collectedData[fieldName].(map[string]interface{}); ok {
			logger.Debug("collectResponses: found step data",
				zap.String("field", fieldName),
				zap.Strings("step_data_keys", getMapKeys(stepData)))

			// Pattern 1: Look for explicit "response" field (from handleCompleteResponse)
			if response, ok := stepData["response"].(map[string]interface{}); ok {
				responses[fieldName] = response
				logger.Debug("collectResponses: found response field",
					zap.String("field", fieldName))
				continue
			}

			// Pattern 2: The step data itself IS the response (after our fix)
			// Look for actual content indicators
			hasContent := false
			for key := range stepData {
				if strings.Contains(key, "generate_") ||
					strings.Contains(key, "result") ||
					strings.Contains(key, "content") ||
					key == "input_data" {
					hasContent = true
					break
				}
			}

			if hasContent {
				responses[fieldName] = stepData
				logger.Debug("collectResponses: using step data as response",
					zap.String("field", fieldName))
				continue
			}
		}

		logger.Warn("collectResponses: field not found or empty",
			zap.String("field", fieldName))
	}

	logger.Info("collectResponses: collection complete",
		zap.Int("requested", len(responseFields)),
		zap.Int("found", len(responses)),
		zap.Strings("found_fields", getMapKeys(responses)))

	return responses
}

// Helper: Extract actual content from nested response structure
// extractResponseContent extracts actual content from nested response structure
// It handles multiple patterns:
// 1. Direct step results: response["step_name"]["result"]
// 2. Nested generate_content: response["generate_content"]["result"]
// 3. Direct result: response["result"]
// 4. Direct content: response["content"]
func extractResponseContent(response interface{}) string {
	if respMap, ok := response.(map[string]interface{}); ok {

		// Pattern 1: Look for step results from hierarchical workflows
		// Example: {"call_researcher": {"result": "..."}, "generate_hero_content": {"result": "..."}}
		for key, value := range respMap {
			// Look for keys that might contain step results
			if strings.Contains(key, "generate_") ||
				strings.Contains(key, "_content") ||
				strings.HasSuffix(key, "_result") {
				if stepMap, ok := value.(map[string]interface{}); ok {
					if result, ok := stepMap["result"].(string); ok && result != "" {
						return result
					}
				}
			}
		}

		// Pattern 2: Try nested generate_content.result (legacy structure)
		if genContent, ok := respMap["generate_content"].(map[string]interface{}); ok {
			if result, ok := genContent["result"].(string); ok {
				return result
			}
		}

		// Pattern 3: Try direct result field
		if result, ok := respMap["result"].(string); ok {
			return result
		}

		// Pattern 4: Try content field
		if content, ok := respMap["content"].(string); ok {
			return content
		}

		// Pattern 5: Look for ANY field with "result" in it as last resort
		for key, value := range respMap {
			if strings.Contains(strings.ToLower(key), "result") {
				if strVal, ok := value.(string); ok && strVal != "" {
					return strVal
				}
			}
		}
	}
	return ""
}

func extractResponseContentOld(response interface{}) string {
	if respMap, ok := response.(map[string]interface{}); ok {
		// Try nested generate_content.result (your structure)
		if genContent, ok := respMap["generate_content"].(map[string]interface{}); ok {
			if result, ok := genContent["result"].(string); ok {
				return result
			}
		}
		// Try direct result field
		if result, ok := respMap["result"].(string); ok {
			return result
		}
		// Try content field
		if content, ok := respMap["content"].(string); ok {
			return content
		}
	}
	return ""
}

// Helper: Get string slice from config with default
func getConfigStringSlice(config map[string]interface{}, key string, defaultVal []string) []string {
	if val, ok := config[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultVal
}

// Helper: Get map from config
func getConfigMap(config map[string]interface{}, key string) map[string]interface{} {
	if val, ok := config[key].(map[string]interface{}); ok {
		return val
	}
	return make(map[string]interface{})
}

// Helper: Get string from config with default
func getConfigString(config map[string]interface{}, key string, defaultVal string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultVal
}

// Helper: Get bool from config with default
func getConfigBool(config map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultVal
}
