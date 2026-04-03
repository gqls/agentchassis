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
	displayName := functionToDisplayName(functionName)

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
		functionName,                             // $1 name
		displayName,                              // $2 display_name
		functionName,                             // $3 function
		category,                                 // $4 category
		sectionType,                              // $5 section_type
		string(suitableSiteTypesJSON),            // $6 suitable_site_types
		string(suitablePageTypesJSON),            // $7 suitable_page_types
		description,                              // $8 description
		htmlTemplate,                             // $9 html_template
		inputSchemaJSON,                          // $10 input_schema
		isDark,                                   // $11 is_dark_section
		buildSemanticTags(sectionType, siteType), // $12 semantic_tags
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
				// Strip markdown code blocks if present
				cleaned := stripCodeBlocks(r)
				// Try parsing the result string as JSON
				if err := json.Unmarshal([]byte(cleaned), &data); err != nil {
					// Not JSON — maybe it's raw HTML
					logger.Info("store_generated_component: result is not JSON, treating as raw HTML",
						zap.Int("length", len(cleaned)),
						zap.String("first_50", truncate(cleaned, 50)))
					data = map[string]interface{}{
						"html_template": cleaned,
					}
				}
			case map[string]interface{}:
				data = r
			}
		} else {
			data = v
		}
	case string:
		// Strip markdown code blocks if present
		cleaned := stripCodeBlocks(v)
		// Try JSON parse
		if err := json.Unmarshal([]byte(cleaned), &data); err != nil {
			// Raw HTML string
			data = map[string]interface{}{
				"html_template": cleaned,
			}
		}
	default:
		return "", "{}", "", false, fmt.Errorf("unexpected type for generated template: %T", raw)
	}

	// Extract html_template
	if ht, ok := data["html_template"].(string); ok {
		htmlTemplate = strings.TrimSpace(ht)
	}

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
	functionName = normaliseToKebab(functionName)

	// Extract is_dark_section
	if dark, ok := data["is_dark_section"].(bool); ok {
		isDark = dark
	}

	return htmlTemplate, inputSchemaJSON, functionName, isDark, nil
}

// stripCodeBlocks removes markdown code block wrappers (```json ... ``` or ``` ... ```)
// that LLMs commonly add around JSON output.
func stripCodeBlocks(s string) string {
	s = strings.TrimSpace(s)
	// Handle ```json\n...\n``` or ```\n...\n```
	if strings.HasPrefix(s, "```") {
		// Find end of first line (the opening ```)
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		// Remove trailing ```
		if strings.HasSuffix(s, "```") {
			s = s[:len(s)-3]
		}
		s = strings.TrimSpace(s)
	}
	return s
}

// normaliseToKebab ensures a string is valid kebab-case for the function column.
func normaliseToKebab(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == '-' || r == '_' || r == ' ' {
			return '-'
		}
		return -1
	}, s)
	// Remove double hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}

// functionToDisplayName converts "spark-provocation-card" to "Spark Provocation Card"
func functionToDisplayName(function string) string {
	parts := strings.Split(function, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// buildSemanticTags generates initial semantic tags from section_type and site_type
func buildSemanticTags(sectionType, siteType string) string {
	tags := []string{}

	// Add section type parts as tags
	for _, part := range strings.Split(sectionType, "-") {
		if part != "" {
			tags = append(tags, part)
		}
	}

	// Add site type if provided
	if siteType != "" {
		tags = append(tags, siteType)
	}

	// Add "generated" provenance tag
	tags = append(tags, "generated")

	tagsJSON, _ := json.Marshal(tags)
	return string(tagsJSON)
}
