// FILE: platform/orchestration/actions/store_generated_component_guard_test.go
//
// Deterministic unit tests for the F1 field-contract guard's decision logic:
// schemaFieldSet (the fields parse) and the stranded-field diff the guard runs
// on a regeneration. No DB required. The full reject path
// (recordValidationRejection + error return + no-overwrite) is covered by the
// integration test described in RUNBOOK step F2 (Tier 2).

package actions

import (
	"sort"
	"testing"
)

func TestSchemaFieldSet(t *testing.T) {
	got := schemaFieldSet(`{"fields":{"a":{"type":"text"},"b":{"type":"text"}}}`)
	if len(got) != 2 || !got["a"] || !got["b"] {
		t.Fatalf("expected {a,b}, got %v", got)
	}
	for _, empty := range []string{"", "{}", "not-json", `{"no_fields":1}`, `{"fields":[]}`} {
		if s := schemaFieldSet(empty); len(s) != 0 {
			t.Errorf("expected empty set for %q, got %v", empty, s)
		}
	}
}

// stranded mirrors the guard's inline computation in StoreGeneratedComponentAction:
// retained fields that disappear from the new schema.
func stranded(oldJSON, newJSON string) []string {
	oldF := schemaFieldSet(oldJSON)
	newF := schemaFieldSet(newJSON)
	var s []string
	for name := range oldF {
		if !newF[name] {
			s = append(s, name)
		}
	}
	sort.Strings(s)
	return s
}

func TestRegenFieldContractStranded(t *testing.T) {
	cases := []struct {
		name    string
		oldJSON string
		newJSON string
		want    []string
	}{
		{
			// the real system-stats rename: number->value, eyebrow->eyebrow_label
			name:    "rename strands the old names",
			oldJSON: `{"fields":{"eyebrow":{},"stat_1_number":{},"heading":{}}}`,
			newJSON: `{"fields":{"eyebrow_label":{},"stat1_value":{},"section_headline":{}}}`,
			want:    []string{"eyebrow", "heading", "stat_1_number"},
		},
		{
			name:    "addition only strands nothing",
			oldJSON: `{"fields":{"a":{},"b":{}}}`,
			newJSON: `{"fields":{"a":{},"b":{},"cta_url":{}}}`,
			want:    nil,
		},
		{
			name:    "drop strands the removed field",
			oldJSON: `{"fields":{"a":{},"stat_1_icon":{}}}`,
			newJSON: `{"fields":{"a":{}}}`,
			want:    []string{"stat_1_icon"},
		},
		{
			name:    "identical strands nothing",
			oldJSON: `{"fields":{"a":{},"b":{}}}`,
			newJSON: `{"fields":{"a":{},"b":{}}}`,
			want:    nil,
		},
		{
			name:    "new component (no old schema) strands nothing",
			oldJSON: `{}`,
			newJSON: `{"fields":{"a":{},"b":{}}}`,
			want:    nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := stranded(c.oldJSON, c.newJSON)
			if len(got) != len(c.want) {
				t.Fatalf("stranded=%v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("stranded=%v, want %v", got, c.want)
				}
			}
		})
	}
}
