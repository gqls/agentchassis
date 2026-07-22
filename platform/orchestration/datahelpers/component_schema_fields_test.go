// FILE: platform/orchestration/datahelpers/component_schema_fields_test.go
//
// Regression tests for the schema-dialect fail-open fixed in bugs_open/026: a
// component authored in the legacy JSON-Schema dialect must have its fields
// projected onto the v2 view (so callers plan for and enforce them), and the
// projection must be flagged (fromLegacy) so a build/audit path can trip the
// fleet-wide fail-loud rather than absorb the extinct dialect silently.

package datahelpers

import (
	"reflect"
	"testing"
)

// The EXACT input_schema news-listing carried when 026 was filed: legacy
// JSON-Schema dialect, `headline` required and source:llm.
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
	t.Run("v2 fields dialect returned as-is, not flagged legacy", func(t *testing.T) {
		f := map[string]interface{}{"content": map[string]interface{}{"source": "llm", "required": true}}
		got, ok, fromLegacy := SchemaContentFields(map[string]interface{}{"fields": f})
		if !ok || fromLegacy {
			t.Fatalf("v2 schema: want ok=true fromLegacy=false, got ok=%v fromLegacy=%v", ok, fromLegacy)
		}
		if !reflect.DeepEqual(got, f) {
			t.Fatalf("v2 fields should pass through unchanged, got %v", got)
		}
	})

	t.Run("empty v2 fields is ok=true, not legacy", func(t *testing.T) {
		got, ok, fromLegacy := SchemaContentFields(map[string]interface{}{"fields": map[string]interface{}{}})
		if !ok || fromLegacy || len(got) != 0 {
			t.Fatalf("empty fields: want (empty,true,false), got (%v,%v,%v)", got, ok, fromLegacy)
		}
	})

	t.Run("legacy properties dialect is projected and flagged", func(t *testing.T) {
		got, ok, fromLegacy := SchemaContentFields(legacyNewsListingSchema())
		if !ok || !fromLegacy {
			t.Fatalf("legacy schema: want ok=true fromLegacy=true, got ok=%v fromLegacy=%v", ok, fromLegacy)
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

	t.Run("legacy property with no explicit source defaults to llm", func(t *testing.T) {
		schema := map[string]interface{}{
			"required":   []interface{}{"title"},
			"properties": map[string]interface{}{"title": map[string]interface{}{"type": "string"}},
		}
		got, _, _ := SchemaContentFields(schema)
		if got["title"].(map[string]interface{})["source"] != "llm" {
			t.Fatalf("unspecified source should default to llm, got %v", got["title"])
		}
	})

	t.Run("empty schema is ok=false, not legacy", func(t *testing.T) {
		if _, ok, fl := SchemaContentFields(map[string]interface{}{}); ok || fl {
			t.Fatalf("empty {}: want ok=false fromLegacy=false, got ok=%v fl=%v", ok, fl)
		}
		if _, ok, fl := SchemaContentFields(nil); ok || fl {
			t.Fatalf("nil: want ok=false fromLegacy=false, got ok=%v fl=%v", ok, fl)
		}
	})

	t.Run("bare example-value schema (legacy core sections) is ok=false, not legacy", func(t *testing.T) {
		// hero/header/footer et al: no `fields`, no `properties`, no requiredness.
		bare := map[string]interface{}{"headline": "string", "primary_cta": "string"}
		if _, ok, fl := SchemaContentFields(bare); ok || fl {
			t.Fatalf("bare example-value map declares no fields — want ok=false fromLegacy=false, got ok=%v fl=%v", ok, fl)
		}
	})
}

// WarnLegacyDialect must tolerate a nil logger (test / non-logging call sites).
func TestWarnLegacyDialectNilLogger(t *testing.T) {
	WarnLegacyDialect(nil, "test", "news-listing") // must not panic
}
