package gripper

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// The vocabulary must be the cluster's. This test pins the names the live
// report-builder workflow reads (load_request: spec->>'…') so a rename here
// cannot silently break the pipeline downstream.
func TestFieldNamesAreTheClustersVocabulary(t *testing.T) {
	want := map[string]bool{
		"mass_kg": true, "travel_mm": true, "surface_material": true, "cycle_rate": true,
		"ip_min": true, "mounting": true, "part_geometry": true, "application": true,
	}
	have := map[string]bool{}
	for _, n := range FieldNames() {
		have[n] = true
	}
	for n := range want {
		if !have[n] {
			t.Errorf("cluster reads spec->>'%s' but the vocabulary does not carry it", n)
		}
	}
	// score_grippers' hard requirements must be required here too, or a
	// "complete" spec could still fail scoring.
	for _, n := range []string{"mass_kg", "travel_mm", "surface_material"} {
		f, _ := fieldByName(n)
		if !f.Required {
			t.Errorf("%s is required by score_grippers but optional here", n)
		}
	}
	// The design's OLD names must not have crept back in.
	for _, old := range []string{"payload_kg", "part_surface", "environment", "notes"} {
		if have[old] {
			t.Errorf("design-doc name %q present; the cluster does not read it", old)
		}
	}
}

func TestNormaliseDropsUnknownAndMistyped(t *testing.T) {
	raw := map[string]interface{}{
		"mass_kg":          2.5,
		"travel_mm":        "sixty", // wrong type → dropped
		"surface_material": "Steel", // case-normalised
		"ip_min":           65.0,
		"cycle_rate":       0, // not > 0 → dropped
		"mounting":         "  UR5e ",
		"application":      nil,     // null → absent
		"budget":           "",      // empty → absent
		"email":            "x@y.z", // unknown → dropped
		"__proto__":        "x",
	}
	got := Normalise(raw)
	want := Spec{"mass_kg": 2.5, "surface_material": "steel", "ip_min": 65, "mounting": "UR5e"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Normalise = %#v, want %#v", got, want)
	}
}

func TestNormaliseMaterialAliasesAndRejects(t *testing.T) {
	if s := Normalise(map[string]interface{}{"surface_material": "ALUMINUM"}); s["surface_material"] != "aluminium" {
		t.Errorf("aluminum alias not accepted: %#v", s)
	}
	if s := Normalise(map[string]interface{}{"surface_material": "titanium"}); len(s) != 0 {
		t.Errorf("off-list material accepted: %#v", s)
	}
}

func TestNormaliseIntRejectsFractions(t *testing.T) {
	if s := Normalise(map[string]interface{}{"ip_min": 65.5}); len(s) != 0 {
		t.Errorf("fractional ip_min accepted: %#v", s)
	}
	if s := Normalise(map[string]interface{}{"ip_min": -1}); len(s) != 0 {
		t.Errorf("negative ip_min accepted: %#v", s)
	}
	if s := Normalise(map[string]interface{}{"ip_min": 0}); s["ip_min"] != 0 {
		t.Errorf("ip_min 0 (no requirement) should be accepted: %#v", s)
	}
}

func TestNormaliseCapsText(t *testing.T) {
	long := strings.Repeat("x", maxTextLen*3)
	s := Normalise(map[string]interface{}{"application": long})
	if got := len(s["application"].(string)); got != maxTextLen {
		t.Errorf("text not capped: len=%d", got)
	}
}

func TestMergeNonNullNeverRegresses(t *testing.T) {
	stored := Spec{"mass_kg": 2.5, "mounting": "UR5e"}
	turn := Normalise(map[string]interface{}{"mass_kg": nil, "travel_mm": 60.0}) // model forgot mass this turn
	got := Merge(stored, turn)
	if got["mass_kg"] != 2.5 {
		t.Errorf("mass_kg regressed to %v", got["mass_kg"])
	}
	if got["travel_mm"] != 60.0 || got["mounting"] != "UR5e" {
		t.Errorf("merge lost a value: %#v", got)
	}
	// A genuine change of mind arrives non-null and wins.
	turn2 := Normalise(map[string]interface{}{"mass_kg": 3.0})
	if Merge(got, turn2)["mass_kg"] != 3.0 {
		t.Errorf("new non-null value did not win")
	}
}

func TestMissingIsInAskOrderAndCompleteAgrees(t *testing.T) {
	s := Spec{}
	if !reflect.DeepEqual(Missing(s), RequiredNames()) {
		t.Fatalf("empty spec Missing = %v, want %v", Missing(s), RequiredNames())
	}
	if Complete(s) {
		t.Fatal("empty spec reported complete")
	}
	for _, n := range RequiredNames() {
		switch n {
		case "surface_material":
			s[n] = "steel"
		case "part_geometry", "mounting":
			s[n] = "x"
		default:
			s[n] = 1.0
		}
	}
	if !Complete(s) || len(Missing(s)) != 0 {
		t.Fatalf("all required set but Missing=%v", Missing(s))
	}
	// Optional fields never appear in missing.
	for _, m := range Missing(Spec{}) {
		f, _ := fieldByName(m)
		if !f.Required {
			t.Errorf("optional %s listed as missing", m)
		}
	}
	// Missing on a complete spec is [] not nil, so JSON renders [] not null.
	if b, _ := json.Marshal(Missing(s)); string(b) != "[]" {
		t.Errorf("Missing(complete) marshals as %s, want []", b)
	}
}

func TestForClusterAddsNoIdentifiers(t *testing.T) {
	out := ForCluster(Spec{"mass_kg": 1.0})
	for _, k := range []string{"request_id", "submitted_at"} {
		if _, ok := out[k]; ok {
			t.Errorf("ForCluster added %s; the puller owns that key", k)
		}
	}
}
