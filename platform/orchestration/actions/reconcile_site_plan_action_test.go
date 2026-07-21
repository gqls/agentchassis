package actions

import (
	"testing"

	"github.com/google/uuid"
)

// Tests for decideEmit — the per-page rebuild/skip/re-stamp decision at the
// heart of /bugs_open/038. The bug: a re-plan writes a NEW site_plans row, so
// every deployed page fails the plain built_from == planID check and was being
// rebuilt (regenerating its copy) even when the plan left the page untouched.
// The fix re-stamps such pages instead of rebuilding, gated on the new plan
// proposing the SAME composition the page was built from (plan-to-plan on
// site_plan_sections; see decideEmit's doc for why NOT pages.sections).

func deployedFrom(planID uuid.UUID) realisedPageRecord {
	return realisedPageRecord{
		BuildStatus:          "deployed",
		BuiltFromPlanVersion: uuid.NullUUID{UUID: planID, Valid: true},
		RebuildPolicy:        "generic",
	}
}

func TestDecideEmit(t *testing.T) {
	current := uuid.New()
	older := uuid.New()

	secs := func(s ...string) []string { return s } // nil when called with none

	cases := []struct {
		name      string
		realised  realisedPageRecord
		curSecs   []string
		builtSecs []string
		want      string
	}{
		{
			name:     "no pages row -> missing",
			realised: realisedPageRecord{BuildStatus: ""},
			want:     "missing",
		},
		{
			name:     "planned but not deployed -> not_built",
			realised: realisedPageRecord{BuildStatus: "planned"},
			want:     "not_built",
		},
		{
			name:     "needs_rebuild -> not_built (owned by /bugs_open/037, not this path)",
			realised: realisedPageRecord{BuildStatus: "needs_rebuild"},
			want:     "not_built",
		},
		{
			name:     "deployed, no plan-version stamp -> stale",
			realised: realisedPageRecord{BuildStatus: "deployed", BuiltFromPlanVersion: uuid.NullUUID{}},
			want:     "stale",
		},
		{
			name:      "deployed at current plan -> skip_built",
			realised:  deployedFrom(current),
			curSecs:   secs("hero", "call-to-action"),
			builtSecs: secs("hero", "call-to-action"),
			want:      "skip_built",
		},
		{
			name:      "deployed at older plan, same composition -> restamp",
			realised:  deployedFrom(older),
			curSecs:   secs("hero-about", "about-content", "differentiators", "call-to-action"),
			builtSecs: secs("hero-about", "about-content", "differentiators", "call-to-action"),
			want:      "restamp",
		},
		{
			name:      "deployed at older plan, composition changed -> stale",
			realised:  deployedFrom(older),
			curSecs:   secs("hero", "product-grid", "features", "call-to-action"),
			builtSecs: secs("hero", "product-grid", "call-to-action"),
			want:      "stale",
		},
		{
			name:      "deployed at older plan, sections reordered -> stale",
			realised:  deployedFrom(older),
			curSecs:   secs("hero", "features", "product-grid"),
			builtSecs: secs("hero", "product-grid", "features"),
			want:      "stale",
		},
		{
			name:      "deployed at older plan, built-from list unknown -> stale (no evidence to skip on)",
			realised:  deployedFrom(older),
			curSecs:   secs("hero", "call-to-action"),
			builtSecs: nil,
			want:      "stale",
		},
		{
			name:      "deployed at older plan, current now sectionless but old had sections -> stale",
			realised:  deployedFrom(older),
			curSecs:   nil,
			builtSecs: secs("hero", "call-to-action"),
			want:      "stale",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := decideEmit(tc.realised, current, tc.curSecs, tc.builtSecs)
			if got != tc.want {
				t.Fatalf("decideEmit = %q, want %q", got, tc.want)
			}
		})
	}
}

// The re-stamp comparison must use the SAME normalisation the loader applies,
// or a plan that wrote "call_to_action" and one that wrote "call-to-action"
// (the same component) would false-negative and force a needless rebuild.
func TestDecideEmit_NormalisationUnifiesEquivalentNames(t *testing.T) {
	current := uuid.New()
	older := uuid.New()

	// Mirror loadPlanSections: normalise each raw plan name before comparing.
	norm := func(raw ...string) []string {
		out := make([]string, len(raw))
		for i, r := range raw {
			out[i] = NormalizeComponentFunction(r)
		}
		return out
	}

	cur := norm("hero", "call-to-action", "social-proof")
	built := norm("hero", "call_to_action", "SocialProof")

	if got := decideEmit(deployedFrom(older), current, cur, built); got != "restamp" {
		t.Fatalf("decideEmit = %q, want restamp (normalisation should unify the names)", got)
	}
}

func TestSectionsEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"both nil", nil, nil, true},
		{"nil vs empty", nil, []string{}, true},
		{"identical", []string{"a", "b"}, []string{"a", "b"}, true},
		{"different length", []string{"a"}, []string{"a", "b"}, false},
		{"different order", []string{"a", "b"}, []string{"b", "a"}, false},
		{"different element", []string{"a", "b"}, []string{"a", "c"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sectionsEqual(tc.a, tc.b); got != tc.want {
				t.Fatalf("sectionsEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
