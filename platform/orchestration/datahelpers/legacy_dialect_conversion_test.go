// FILE: platform/orchestration/datahelpers/legacy_dialect_conversion_test.go
//
// bugs_closed/265 / migration 437. The migration converts the last three legacy
// JSON-Schema input_schema rows to the v2 `fields` dialect IN SQL, mirroring
// SchemaContentFields' projection, and then adds a CHECK constraint that refuses
// the legacy dialect for every producer. This test is the proof that the SQL
// projection and the Go projection agree: for each row, the field map the
// helper returns from the BEFORE (legacy) schema must deep-equal the field map
// it returns from the AFTER (converted) schema, and the AFTER schema must read
// as v2 (ok=true, fromLegacy=false).
//
// The fixture is not hand-written. `before` is the three rows as captured live
// on 2026-08-16; `after` is what migration 437's UPDATE produced for them in a
// probe run that was rolled back. If the migration's projection is ever edited,
// re-capture the fixture the same way — a fixture that is re-typed by hand
// proves only that the typist agreed with themselves.

package datahelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type legacyConversionFixture struct {
	Rows []struct {
		Function string                 `json:"function"`
		Before   map[string]interface{} `json:"before"`
		After    map[string]interface{} `json:"after"`
	} `json:"rows"`
}

func TestLegacyDialectConversionMatchesProjection(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "legacy_dialect_conversion_437.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fx legacyConversionFixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	if len(fx.Rows) != 3 {
		t.Fatalf("fixture should carry the 3 rows migration 437 converts, got %d", len(fx.Rows))
	}

	for _, row := range fx.Rows {
		t.Run(row.Function, func(t *testing.T) {
			// The BEFORE shape is the legacy dialect and the helper says so.
			wantFields, ok, fromLegacy := SchemaContentFields(row.Before)
			if !ok || !fromLegacy {
				t.Fatalf("before: want ok=true fromLegacy=true, got ok=%v fromLegacy=%v", ok, fromLegacy)
			}
			// The AFTER shape is v2: no top-level properties (the constraint's
			// predicate), read as-is, not flagged.
			if _, has := row.After["properties"]; has {
				t.Fatalf("after: still carries a top-level properties key — the constraint would refuse it")
			}
			gotFields, ok, fromLegacy := SchemaContentFields(row.After)
			if !ok || fromLegacy {
				t.Fatalf("after: want ok=true fromLegacy=false, got ok=%v fromLegacy=%v", ok, fromLegacy)
			}
			// And every consumer sees the SAME field set either way.
			if !reflect.DeepEqual(wantFields, gotFields) {
				t.Fatalf("projection mismatch for %s:\n before→ %v\n after → %v", row.Function, wantFields, gotFields)
			}
		})
	}
}

// IsLegacyInputSchemaDialect is what the birth path calls before it writes;
// pin its verdicts on the shapes that matter.
func TestIsLegacyInputSchemaDialect(t *testing.T) {
	cases := []struct {
		name   string
		schema map[string]interface{}
		want   bool
	}{
		{"nil", nil, false},
		{"empty", map[string]interface{}{}, false},
		{"v2", map[string]interface{}{"fields": map[string]interface{}{"x": map[string]interface{}{"source": "llm"}}}, false},
		{"legacy", legacyNewsListingSchema(), true},
		// A bare example-value map is the OLDEST shape, not the legacy JSON-Schema
		// dialect; the reader treats it as "no declared fields" and it stays storable.
		{"bare example map", map[string]interface{}{"headline": "string"}, false},
		// Both keys present: the reader takes fields; the constraint would still
		// refuse the row (top-level properties), so the birth path must too.
		{"fields and properties", map[string]interface{}{
			"fields":     map[string]interface{}{"x": map[string]interface{}{"source": "llm"}},
			"properties": map[string]interface{}{"x": map[string]interface{}{"type": "string"}},
		}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsLegacyInputSchemaDialect(c.schema); got != c.want {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}
