// FILE: platform/orchestration/actions/load_existing_component_action.go
//
// LoadExistingComponentAction looks up the canonical existing shared component
// for a section_type (active, non-forked, section-level) and, if found, outputs
// its input_schema field names as a ready-to-print string. component-creator's
// generate_template prompt uses this (via the {{if .existing_component.field_names}}
// block) to instruct the LLM to REUSE those field names on regeneration, so the
// shared field contract that dependents' content_data is keyed on is preserved.
//
// It is the pre-generation half of the field-name-preservation fix; the
// StoreGeneratedComponentAction field-contract guard is the store-time backstop.
//
// Advisory by design: a new section (no existing row) yields empty output and
// the prompt block stays dormant (component generated fresh); and ANY lookup
// problem degrades to blind generation rather than blocking, because the guard
// still catches drift. So this action never returns an error, and always
// returns a well-formed map so the prompt's {{if ...}} guard is safe even if an
// upstream error_step routed around an earlier step.
//
// Registration (registry.go):
//   "load_existing_component": {
//       Handler:     LoadExistingComponentAction,
//       Category:    "site",
//       Description: "Load existing component field names for regeneration preservation",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadExistingComponentInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"section_type"},
	Optional:    []string{},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_existing_component", LoadExistingComponentInputSpec)
}

func LoadExistingComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_existing_component"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Always return a well-formed map so the prompt's
	// {{if .existing_component.field_names}} guard is safe, and never block
	// generation on a lookup problem (the store-time guard is the backstop).
	empty := map[string]interface{}{"field_names": "", "function": "", "field_count": 0}

	if params.DB == nil {
		logger.Warn("load_existing_component: no DB — generating blind")
		return empty, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, LoadExistingComponentInputSpec, logger,
	)
	if err != nil {
		logger.Warn("load_existing_component: input extraction failed — generating blind", zap.Error(err))
		return empty, nil
	}

	sectionType := inputs.Get("section_type")
	if sectionType == "" {
		logger.Info("load_existing_component: no section_type — generating blind")
		return empty, nil
	}

	// Canonical shared section component for this section_type (matches the
	// selector index: active, non-forked, section-level). If several exist,
	// prefer most-used then most-recent — the row dependents are most likely
	// bound to.
	var function string
	var schemaJSON []byte
	err = params.DB.QueryRowContext(ctx, `
		SELECT function, input_schema
		FROM content_components
		WHERE section_type = $1
		  AND forked_from IS NULL
		  AND is_active = true
		  AND component_level = 'section'
		ORDER BY usage_count DESC NULLS LAST, updated_at DESC
		LIMIT 1
	`, sectionType).Scan(&function, &schemaJSON)
	if err != nil {
		// No existing component (sql.ErrNoRows) or a lookup failure — advisory
		// empty, generation proceeds fresh.
		logger.Info("load_existing_component: no existing component for section_type (or lookup failed) — generating fresh",
			zap.String("section_type", sectionType), zap.Error(err))
		return empty, nil
	}

	names := schemaFieldNamesSorted(schemaJSON)
	if len(names) == 0 {
		logger.Info("load_existing_component: existing component has no schema fields",
			zap.String("section_type", sectionType), zap.String("function", function))
		return map[string]interface{}{"field_names": "", "function": function, "field_count": 0}, nil
	}

	logger.Info("load_existing_component: found existing component — requesting field-name preservation",
		zap.String("section_type", sectionType),
		zap.String("function", function),
		zap.Int("field_count", len(names)))

	return map[string]interface{}{
		"field_names": strings.Join(names, ", "),
		"function":    function,
		"field_count": len(names),
	}, nil
}

// schemaFieldNamesSorted extracts the sorted field names from an
// input_schema.fields object. Empty or invalid input → nil.
func schemaFieldNamesSorted(inputSchemaJSON []byte) []string {
	if len(inputSchemaJSON) == 0 {
		return nil
	}
	var schema map[string]interface{}
	if err := json.Unmarshal(inputSchemaJSON, &schema); err != nil {
		return nil
	}
	fields, ok := schema["fields"].(map[string]interface{})
	if !ok {
		return nil
	}
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
