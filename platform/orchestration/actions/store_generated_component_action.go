// FILE: platform/orchestration/actions/store_generated_component_action.go
//
// StoreGeneratedComponentAction stores an LLM-generated component template
// into the content_components table with full selection metadata.
//
// Used by the component-creator handler agent after execute_llm_prompt
// generates the html_template and input_schema.
//
// Registration:
//   "store_generated_component": {
//       Handler:     StoreGeneratedComponentAction,
//       Category:    "site",
//       Description: "Store a generated component template in the component library",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "store_component": {
//       "action": "store_generated_component",
//       "config": {
//           "section_type": "input_data.section_type",
//           "site_type": "input_data.site_type",
//           "generated_template": "generate_template"
//       },
//       "next_step": "complete",
//       "output_field": "stored_component"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var StoreGeneratedComponentInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"section_type"},
	Optional:   []string{"site_type", "page_context", "description", "design_direction", "generated_template"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("store_generated_component", StoreGeneratedComponentInputSpec)
}

func StoreGeneratedComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "store_generated_component"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		StoreGeneratedComponentInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	sectionType := inputs.Get("section_type")
	siteType := inputs.Get("site_type")
	pageContext := inputs.Get("page_context")
	description := inputs.Get("description")

	// The LLM output is in collected_data under the output_field of the generate step.
	// extract the generated template — look for the LLM result which contains
	// html_template and input_schema as structured output.
	generatedRaw := inputs.GetRaw("generated_template")

	htmlTemplate, inputSchemaJSON, functionName, isDark, err := parseGeneratedTemplate(generatedRaw, sectionType, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse generated template: %w", err)
	}

	if htmlTemplate == "" {
		return nil, fmt.Errorf("generated template is empty for section_type %q", sectionType)
	}

	// Build suitable_site_types from the site_type that triggered the creation
	suitableSiteTypes := []string{}
	if siteType != "" {
		suitableSiteTypes = append(suitableSiteTypes, siteType)
	}
	suitableSiteTypesJSON, _ := json.Marshal(suitableSiteTypes)

	// Build suitable_page_types from page context if available
	suitablePageTypes := []string{}
	if pageContext != "" {
		suitablePageTypes = append(suitablePageTypes, pageContext)
	}
	suitablePageTypesJSON, _ := json.Marshal(suitablePageTypes)

	// Build display name from function
	displayName := datahelpers.FunctionToDisplayName(functionName)

	// Determine category from site_type
	category := "custom"
	if siteType != "" {
		category = siteType
	}

	logger.Info("store_generated_component: storing component",
		zap.String("section_type", sectionType),
		zap.String("function", functionName),
		zap.String("category", category),
		zap.Int("template_length", len(htmlTemplate)),
		zap.Bool("is_dark", isDark))

	// Check if a component with this function already exists
	var existingID string
	err = params.DB.QueryRowContext(ctx, `
		SELECT id::text FROM content_components
		WHERE function = $1 AND is_active = true AND forked_from IS NULL
		LIMIT 1
	`, functionName).Scan(&existingID)

	if err == nil {
		// Component already exists — don't overwrite, return the existing one
		logger.Info("store_generated_component: component already exists, skipping",
			zap.String("function", functionName),
			zap.String("existing_id", existingID))
		return map[string]interface{}{
			"component_id": existingID,
			"function":     functionName,
			"status":       "already_exists",
		}, nil
	}

	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing component: %w", err)
	}

	// Insert the new component
	var newID string
	err = params.DB.QueryRowContext(ctx, `
		INSERT INTO content_components (
			name, display_name, function, category, component_level,
			section_type, suitable_site_types, suitable_page_types,
			description, html_template, input_schema,
			is_dark_section, render_mode, created_from, is_active,
			usage_count, avg_quality_score,
			semantic_tags
		) VALUES (
			$1, $2, $3, $4, 'section',
			$5, $6::jsonb, $7::jsonb,
			$8, $9, $10::jsonb,
			$11, 'template', 'generated', true,
			0, NULL,
			$12::jsonb
		)
		RETURNING id::text
	`,
		functionName,                                         // $1 name
		displayName,                                          // $2 display_name
		functionName,                                         // $3 function
		category,                                             // $4 category
		sectionType,                                          // $5 section_type
		string(suitableSiteTypesJSON),                        // $6 suitable_site_types
		string(suitablePageTypesJSON),                        // $7 suitable_page_types
		description,                                          // $8 description
		htmlTemplate,                                         // $9 html_template
		inputSchemaJSON,                                      // $10 input_schema
		isDark,                                               // $11 is_dark_section
		datahelpers.BuildSemanticTags(sectionType, siteType), // $12 semantic_tags
	).Scan(&newID)

	if err != nil {
		return nil, fmt.Errorf("failed to insert component: %w", err)
	}

	logger.Info("store_generated_component: component created",
		zap.String("component_id", newID),
		zap.String("function", functionName),
		zap.String("section_type", sectionType))

	return map[string]interface{}{
		"component_id":  newID,
		"function":      functionName,
		"section_type":  sectionType,
		"display_name":  displayName,
		"category":      category,
		"status":        "created",
		"template_size": len(htmlTemplate),
	}, nil
}

// parseGeneratedTemplate extracts html_template, input_schema, and metadata
// from the LLM output. The LLM is instructed to return JSON with these fields,
// but we handle various formats defensively.
func parseGeneratedTemplate(raw interface{}, sectionType string, logger *zap.Logger) (
	htmlTemplate string, inputSchemaJSON string, functionName string, isDark bool, err error,
) {
	if raw == nil {
		return "", "{}", "", false, fmt.Errorf("generated template is nil")
	}

	// The LLM output might be:
	// 1. A map with "result" containing the JSON string (from execute_llm_prompt)
	// 2. A map with html_template/input_schema directly
	// 3. A string containing JSON (possibly wrapped in markdown code blocks)

	var data map[string]interface{}

	switch v := raw.(type) {
	case map[string]interface{}:
		// Check for execute_llm_prompt's "result" wrapper
		if result, ok := v["result"]; ok {
			switch r := result.(type) {
			case string:
				data = parseJSONStringToMap(r, logger)
			case map[string]interface{}:
				data = r
			}
		} else {
			data = v
		}
	case string:
		data = parseJSONStringToMap(v, logger)
	default:
		return "", "{}", "", false, fmt.Errorf("unexpected type for generated template: %T", raw)
	}

	// Extract html_template
	if ht, ok := data["html_template"].(string); ok {
		htmlTemplate = strings.TrimSpace(ht)
	}

	// Safety check: if htmlTemplate still looks like a JSON blob wrapping an
	// html_template field, extract the actual HTML. This catches cases where
	// json.Unmarshal failed and the fallback stored the entire JSON string.
	htmlTemplate = unwrapJSONBlobIfNeeded(htmlTemplate, logger)

	// Extract input_schema
	inputSchemaJSON = "{}"
	if schema, ok := data["input_schema"]; ok {
		if schemaMap, ok := schema.(map[string]interface{}); ok {
			schemaBytes, _ := json.Marshal(schemaMap)
			inputSchemaJSON = string(schemaBytes)
		} else if schemaStr, ok := schema.(string); ok {
			inputSchemaJSON = schemaStr
		}
	}

	// Extract or derive function name
	if fn, ok := data["function"].(string); ok && fn != "" {
		functionName = fn
	} else {
		// Derive from section_type — prefix with a category hint
		functionName = sectionType
	}

	// Validate kebab-case
	functionName = datahelpers.NormaliseToKebab(functionName)

	// Extract is_dark_section
	if dark, ok := data["is_dark_section"].(bool); ok {
		isDark = dark
	}

	return htmlTemplate, inputSchemaJSON, functionName, isDark, nil
}

// parseJSONStringToMap takes a raw string (possibly with markdown code blocks)
// and tries to parse it as a JSON map. If standard parsing fails, it falls back
// to field-level extraction from broken JSON. If the string is not JSON at all,
// it treats it as raw HTML.
//
// Uses datahelpers.StripCodeFences (shared with content_search, create_tool, etc.)
// and datahelpers.SafeUnmarshalString for safe parsing.
func parseJSONStringToMap(s string, logger *zap.Logger) map[string]interface{} {
	// Strip markdown code blocks using the shared helper from datahelpers
	cleaned := datahelpers.StripCodeFences(s)

	// Try standard JSON parse using the shared SafeUnmarshalString
	var data map[string]interface{}
	if datahelpers.SafeUnmarshalString(cleaned, &data) {
		logger.Info("store_generated_component: parsed LLM output as JSON",
			zap.Int("fields", len(data)))
		return data
	}

	// JSON parse failed — try field-level extraction from broken JSON.
	// LLMs often produce JSON with unescaped characters in HTML/SVG content
	// that breaks json.Unmarshal, but the structure is still recoverable.
	if strings.Contains(cleaned, `"html_template"`) {
		logger.Info("store_generated_component: json.Unmarshal failed, attempting field extraction",
			zap.Int("length", len(cleaned)),
			zap.String("first_80", truncateStr(cleaned, 80)))

		result := map[string]interface{}{}

		if ht, ok := extractJSONStringField(cleaned, "html_template"); ok {
			result["html_template"] = ht
			logger.Info("store_generated_component: extracted html_template from broken JSON",
				zap.Int("length", len(ht)),
				zap.String("first_40", truncateStr(ht, 40)))
		}
		if fn, ok := extractJSONStringField(cleaned, "function"); ok {
			result["function"] = fn
		}
		// is_dark_section is a bool, not a string — check with simple contains
		if strings.Contains(cleaned, `"is_dark_section": true`) || strings.Contains(cleaned, `"is_dark_section":true`) {
			result["is_dark_section"] = true
		}

		if _, hasHTML := result["html_template"]; hasHTML {
			return result
		}
	}

	// Not JSON at all — treat as raw HTML
	logger.Info("store_generated_component: treating as raw HTML",
		zap.Int("length", len(cleaned)),
		zap.String("first_50", truncateStr(cleaned, 50)))
	return map[string]interface{}{
		"html_template": cleaned,
	}
}

// extractJSONStringField extracts the value of a string field from a JSON-like
// string, even when json.Unmarshal fails due to unescaped characters elsewhere.
// It manually scans the JSON string value, handling standard escape sequences.
func extractJSONStringField(s, fieldName string) (string, bool) {
	key := `"` + fieldName + `"`
	keyIdx := strings.Index(s, key)
	if keyIdx == -1 {
		return "", false
	}

	// Find the colon after the key
	rest := s[keyIdx+len(key):]
	colonIdx := strings.IndexByte(rest, ':')
	if colonIdx == -1 {
		return "", false
	}
	rest = rest[colonIdx+1:]

	// Skip whitespace
	rest = strings.TrimLeft(rest, " \t\n\r")

	// Must start with a quote
	if len(rest) == 0 || rest[0] != '"' {
		return "", false
	}
	rest = rest[1:] // skip opening quote

	// Scan for the closing unescaped quote, handling escape sequences
	var result strings.Builder
	result.Grow(len(rest))
	i := 0
	for i < len(rest) {
		if rest[i] == '\\' && i+1 < len(rest) {
			// JSON escape sequence
			switch rest[i+1] {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case '/':
				result.WriteByte('/')
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case 'b':
				result.WriteByte('\b')
			case 'f':
				result.WriteByte('\f')
			default:
				// Unknown escape — preserve as-is
				result.WriteByte(rest[i])
				result.WriteByte(rest[i+1])
			}
			i += 2
		} else if rest[i] == '"' {
			// Unescaped closing quote — end of field value
			return result.String(), true
		} else {
			result.WriteByte(rest[i])
			i++
		}
	}

	// No closing quote found. The JSON is severely broken, but we likely have
	// the content up to the end of the string. Return what we collected.
	extracted := result.String()
	if len(extracted) > 0 {
		return strings.TrimSpace(extracted), true
	}
	return "", false
}

// unwrapJSONBlobIfNeeded checks if a string looks like it's still a JSON blob
// wrapping an html_template field, and extracts the actual HTML if so.
// This is a safety net that catches any case where earlier parsing failed.
func unwrapJSONBlobIfNeeded(s string, logger *zap.Logger) string {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "{") {
		return s // Not a JSON blob
	}
	if !strings.Contains(trimmed, `"html_template"`) {
		return s // Doesn't contain the field
	}

	logger.Info("store_generated_component: unwrapping JSON blob from html_template",
		zap.Int("length", len(trimmed)),
		zap.String("first_60", truncateStr(trimmed, 60)))

	// Try standard JSON parse first using shared SafeUnmarshalString
	var wrapper map[string]interface{}
	if datahelpers.SafeUnmarshalString(trimmed, &wrapper) {
		if ht, ok := wrapper["html_template"].(string); ok && strings.TrimSpace(ht) != "" {
			return strings.TrimSpace(ht)
		}
	}

	// Standard parse failed — use field extraction
	if ht, ok := extractJSONStringField(trimmed, "html_template"); ok && ht != "" {
		return ht
	}

	return s // Couldn't unwrap — return as-is
}

// truncateStr returns the first n characters of s, appending "..." if truncated.
// Named truncateStr to avoid conflict with any future stdlib truncate.
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
