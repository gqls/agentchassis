// FILE: platform/orchestration/actions/component_schema_fields_test.go
//
// Enforcement-path regression for bugs_open/026: the render-time required-field
// gate (missingRequiredLLMFields) must catch an empty required field even when
// the component is authored in the legacy JSON-Schema dialect. The reader itself
// (datahelpers.SchemaContentFields) is unit-tested in the datahelpers package;
// this asserts the gate that consumes it.

package actions

import (
	"sort"
	"testing"
)

// The EXACT input_schema news-listing carried when 026 was filed.
func legacyNewsListingSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":     "object",
		"required": []interface{}{"headline"},
		"properties": map[string]interface{}{
			"headline": map[string]interface{}{
				"type":        "string",
				"source":      "llm",
				"description": "Page headline for the news listing",
			},
		},
	}
}

func TestMissingRequiredLLMFields_LegacyDialect(t *testing.T) {
	schema := legacyNewsListingSchema()

	t.Run("empty required headline is flagged (the 026 blanking case)", func(t *testing.T) {
		if m := missingRequiredLLMFields(schema, map[string]interface{}{"headline": ""}); len(m) != 1 || m[0] != "headline" {
			t.Fatalf("legacy-dialect empty required field must be caught, got %v", m)
		}
	})

	t.Run("absent required headline is flagged", func(t *testing.T) {
		if m := missingRequiredLLMFields(schema, map[string]interface{}{}); len(m) != 1 || m[0] != "headline" {
			t.Fatalf("legacy-dialect absent required field must be caught, got %v", m)
		}
	})

	t.Run("populated headline passes", func(t *testing.T) {
		if m := missingRequiredLLMFields(schema, map[string]interface{}{"headline": "Latest watch news"}); len(m) != 0 {
			t.Fatalf("populated required field must pass, got %v", m)
		}
	})

	t.Run("legacy query-sourced field is not enforced as llm", func(t *testing.T) {
		schema := map[string]interface{}{
			"required": []interface{}{"items"},
			"properties": map[string]interface{}{
				"items": map[string]interface{}{"type": "array", "source": "query.news_archive"},
			},
		}
		if m := missingRequiredLLMFields(schema, map[string]interface{}{}); len(m) != 0 {
			t.Fatalf("query-sourced field must not be checked as an llm field, got %v", m)
		}
	})

	t.Run("multiple legacy required llm fields come back sorted", func(t *testing.T) {
		schema := map[string]interface{}{
			"required": []interface{}{"b_field", "a_field"},
			"properties": map[string]interface{}{
				"a_field": map[string]interface{}{"source": "llm"},
				"b_field": map[string]interface{}{"source": "llm"},
			},
		}
		m := missingRequiredLLMFields(schema, map[string]interface{}{})
		if len(m) != 2 || !sort.StringsAreSorted(m) {
			t.Fatalf("both required llm fields should be flagged and sorted, got %v", m)
		}
	})
}
