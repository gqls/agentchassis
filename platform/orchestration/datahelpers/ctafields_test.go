// FILE: platform/orchestration/datahelpers/ctafields_test.go
//
// The two named regression cases pin the corrections from doc_notes 2026-07-20
// (commit b6e374fc2): bare-stem sibling pairing, and the site_assets guard.

package datahelpers

import (
	"encoding/json"
	"reflect"
	"testing"
)

func schemaOf(fields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"fields": fields}
}

func field(source string) map[string]interface{} {
	return map[string]interface{}{"source": source}
}

func TestDeriveCTAURLFields_SuffixAndBareStemForms(t *testing.T) {
	s := schemaOf(map[string]interface{}{
		// suffix form
		"cta_primary_url":   field("site_specs.cta.primary_url"),
		"cta_primary_label": field("static"),
		// _text form (hero.cta_url ↔ cta_text)
		"cta_url":  field("renderer"),
		"cta_text": field("static"),
		// REGRESSION (b6e374fc2 correction 2a): bare-stem form —
		// hero.secondary_cta_url pairs with bare secondary_cta.
		"secondary_cta_url": field("renderer"),
		"secondary_cta":     field("llm"),
		// image url with no sibling: not derived
		"guide_image_url": field("site_assets.image"),
		"guide_image_alt": field("llm"),
	})
	got := DeriveCTAURLFields(s)
	want := []CTAField{
		{URLField: "cta_primary_url", LabelField: "cta_primary_label", Source: "site_specs.cta.primary_url"},
		{URLField: "cta_url", LabelField: "cta_text", Source: "renderer"},
		{URLField: "secondary_cta_url", LabelField: "secondary_cta", Source: "renderer"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("derive mismatch:\n got %+v\nwant %+v", got, want)
	}
}

func TestDeriveCTAURLFields_SiteAssetsGuard(t *testing.T) {
	// REGRESSION (b6e374fc2 correction 2b): header-leopardess.logo_url has
	// sibling logo_text but source site_assets.logo — it is an image and must
	// NOT derive, or the resolver could overwrite a logo with a page link.
	s := schemaOf(map[string]interface{}{
		"logo_url":  field("site_assets.logo"),
		"logo_text": field("site_specs.identity.company_name"),
	})
	if got := DeriveCTAURLFields(s); got != nil {
		t.Fatalf("site_assets url derived as CTA: %+v", got)
	}
}

func TestDeriveCTAURLFields_NilAndEmptySafe(t *testing.T) {
	if got := DeriveCTAURLFields(nil); got != nil {
		t.Fatalf("nil schema: got %+v", got)
	}
	if got := DeriveCTAURLFields(map[string]interface{}{}); got != nil {
		t.Fatalf("empty schema: got %+v", got)
	}
	if got := DeriveCTAURLFields(schemaOf(map[string]interface{}{})); got != nil {
		t.Fatalf("empty fields: got %+v", got)
	}
}

func TestUncoveredCTAURLFields(t *testing.T) {
	s := schemaOf(map[string]interface{}{
		// covered: derives, so not uncovered
		"cta_primary_url":   field("site_specs.cta.primary_url"),
		"cta_primary_label": field("static"),
		// uncovered: query-resolved source, no sibling in any form
		"orphan_cta_url": field("site_specs.cta.secondary_url"),
		// image url, no sibling: excluded (site_assets is not query-resolved here)
		"guide_image_url": field("site_assets.image"),
		// renderer-sourced url, no sibling: not query-resolved, excluded
		"free_url": field("renderer"),
	})
	got := UncoveredCTAURLFields(s)
	want := []string{"orphan_cta_url"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uncovered mismatch: got %v want %v", got, want)
	}
}

func TestParseInputSchemaValue(t *testing.T) {
	m := map[string]interface{}{"fields": map[string]interface{}{}}
	if got := ParseInputSchemaValue(m); !reflect.DeepEqual(got, m) {
		t.Fatalf("map form not passed through")
	}
	b, _ := json.Marshal(m)
	if got := ParseInputSchemaValue(string(b)); got == nil || got["fields"] == nil {
		t.Fatalf("string form not parsed: %+v", got)
	}
	for _, bad := range []interface{}{nil, "", "not-json", 42, []interface{}{}} {
		if got := ParseInputSchemaValue(bad); got != nil {
			t.Fatalf("garbage %v parsed to %+v", bad, got)
		}
	}
}

func TestDeriveCTAURLFields_Deterministic(t *testing.T) {
	s := schemaOf(map[string]interface{}{
		"b_url": field("renderer"), "b_label": field("static"),
		"a_url": field("renderer"), "a_label": field("static"),
		"c_url": field("renderer"), "c": field("static"),
	})
	first := DeriveCTAURLFields(s)
	for i := 0; i < 10; i++ {
		if got := DeriveCTAURLFields(s); !reflect.DeepEqual(got, first) {
			t.Fatalf("nondeterministic output: %+v vs %+v", got, first)
		}
	}
	if first[0].URLField != "a_url" || first[2].URLField != "c_url" {
		t.Fatalf("not sorted: %+v", first)
	}
}
