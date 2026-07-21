// FILE: platform/orchestration/actions/component_schema_fields_test.go
//
// Regression tests for the schema-dialect fail-open fixed in bugs_open/026: a
// component authored in the legacy JSON-Schema dialect must have its fields both
// planned for AND enforced, not silently read as "no fields".

package actions

import (
	"reflect"
	"sort"
	"testing"
)

// The EXACT input_schema news-listing carried when 026 was filed: legacy
// JSON-Schema dialect, `headline` required and source:llm. This is the shape
// that was invisible to both the planner and the render gate.
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

func TestSchemaContentFields(t *testing.T) {
	t.Run("v2 fields dialect returned as-is", func(t *testing.T) {
		f := map[string]interface{}{"content": map[string]interface{}{"source": "llm", "required": true}}
		schema := map[string]interface{}{"fields": f}
		got, ok := schemaContentFields(schema)
		if !ok {
			t.Fatal("v2 schema must report ok=true")
		}
		if !reflect.DeepEqual(got, f) {
			t.Fatalf("v2 fields should pass through unchanged, got %v", got)
		}
	})

	t.Run("empty v2 fields is ok=true (caller treats as no declared fields)", func(t *testing.T) {
		got, ok := schemaContentFields(map[string]interface{}{"fields": map[string]interface{}{}})
		if !ok || len(got) != 0 {
			t.Fatalf("empty fields should be (empty,true), got (%v,%v)", got, ok)
		}
	})

	t.Run("legacy properties dialect is projected with required folded in", func(t *testing.T) {
		got, ok := schemaContentFields(legacyNewsListingSchema())
		if !ok {
			t.Fatal("legacy properties schema must report ok=true")
		}
		hl, ok := got["headline"].(map[string]interface{})
		if !ok {
			t.Fatalf("headline field missing from projection: %v", got)
		}
		if hl["source"] != "llm" {
			t.Errorf("source: want llm, got %v", hl["source"])
		}
		if hl["required"] != true {
			t.Errorf("required: want true (folded from required[]), got %v", hl["required"])
		}
		if hl["type"] != "text" {
			t.Errorf("type: want text (mapped from string), got %v", hl["type"])
		}
		if hl["llm_guidance"] != "Page headline for the news listing" {
			t.Errorf("llm_guidance: want the description text, got %v", hl["llm_guidance"])
		}
	})

	t.Run("property with no explicit source defaults to llm", func(t *testing.T) {
		schema := map[string]interface{}{
			"required":   []interface{}{"title"},
			"properties": map[string]interface{}{"title": map[string]interface{}{"type": "string"}},
		}
		got, _ := schemaContentFields(schema)
		if got["title"].(map[string]interface{})["source"] != "llm" {
			t.Fatalf("unspecified source should default to llm, got %v", got["title"])
		}
	})

	t.Run("empty schema is ok=false", func(t *testing.T) {
		if _, ok := schemaContentFields(map[string]interface{}{}); ok {
			t.Fatal("empty {} should report ok=false")
		}
		if _, ok := schemaContentFields(nil); ok {
			t.Fatal("nil schema should report ok=false")
		}
	})

	t.Run("bare example-value schema (legacy core sections) is ok=false", func(t *testing.T) {
		// hero/header/footer et al: no `fields`, no `properties`, no requiredness.
		bare := map[string]interface{}{"headline": "string", "primary_cta": "string"}
		if _, ok := schemaContentFields(bare); ok {
			t.Fatal("a bare example-value map declares no fields — must be ok=false so the caller's no-fields path is preserved")
		}
	})
}

// The end-to-end regression: the 026 mechanism. Under the legacy dialect the
// render gate must now catch an empty required field, and must still pass a
// populated one.
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
}

// Guard the sort so callers get a stable field ordering across both dialects.
func TestMissingRequiredLLMFields_LegacyDialectSorted(t *testing.T) {
	schema := map[string]interface{}{
		"required": []interface{}{"b_field", "a_field"},
		"properties": map[string]interface{}{
			"a_field": map[string]interface{}{"source": "llm"},
			"b_field": map[string]interface{}{"source": "llm"},
		},
	}
	m := missingRequiredLLMFields(schema, map[string]interface{}{})
	if !sort.StringsAreSorted(m) {
		t.Fatalf("missing fields should be sorted, got %v", m)
	}
	if len(m) != 2 {
		t.Fatalf("both required llm fields should be flagged, got %v", m)
	}
}
