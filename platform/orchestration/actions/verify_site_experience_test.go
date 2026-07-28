package actions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

func TestSubstituteExperienceBindings_ReplacesAndReportsGaps(t *testing.T) {
	doc := `{"checks":[{"id":"a","type":"selector_exists","selector":"{{binding.list}}"},
	                   {"id":"b","type":"selector_exists","selector":"{{binding.row}}"}]}`

	out, unresolved := substituteExperienceBindings(doc, map[string]interface{}{
		"list": ".archive-list", "row": ".archive-row",
	})
	if len(unresolved) != 0 {
		t.Fatalf("a complete binding set reported gaps: %v", unresolved)
	}
	if strings.Contains(out, "{{") {
		t.Fatalf("placeholders survived substitution: %s", out)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("substitution broke the JSON: %v\n%s", err, out)
	}

	// A missing binding must be REPORTED, not silently left as literal text —
	// left in place it would become a selector nothing matches, and Tier 2 would
	// skip it, which reads as green.
	if _, gaps := substituteExperienceBindings(doc, map[string]interface{}{"list": ".x"}); len(gaps) != 1 || gaps[0] != "row" {
		t.Fatalf("a missing binding was not reported: %v", gaps)
	}

	// An EMPTY value counts as unresolved. It satisfies closure and produces a
	// check that cannot fail, which is the worse of the two failures.
	if _, gaps := substituteExperienceBindings(doc, map[string]interface{}{"list": ".x", "row": "   "}); len(gaps) != 1 {
		t.Fatalf("an empty binding was treated as resolved: %v", gaps)
	}
}

func TestSubstituteExperienceBindings_EscapesForJSON(t *testing.T) {
	// The value lands inside a JSON document; an unescaped quote would corrupt
	// it, and a corrupt document fails to parse — which the action reports, but
	// it should not happen from an ordinary selector.
	doc := `{"checks":[{"id":"a","selector":"{{binding.sel}}"}]}`
	out, _ := substituteExperienceBindings(doc, map[string]interface{}{"sel": `.a[data-x="y"]`})
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("a quoted selector broke the JSON: %v\n%s", err, out)
	}
}

// The load-bearing test in this file.
//
// A criteria document can produce zero executed checks — everything skipped
// (unanchorable selector, or an -EDIT id) or everything deferred. Such a run has
// NO FAILURES. If "no failures" were read as a pass, the register would verify a
// fork that asserted nothing, which is exactly the defect it exists to end.
func TestVerifyDecision_NoFailuresIsNotAPass(t *testing.T) {
	cases := []struct {
		name       string
		ev         discovery_checks.StaticEvaluation
		wantStatus string
	}{
		{
			name:       "everything skipped — no failures, nothing asserted",
			ev:         discovery_checks.StaticEvaluation{Skipped: []discovery_checks.StaticCheckOutcome{{ID: "a"}, {ID: "b"}}},
			wantStatus: "inconclusive",
		},
		{
			name:       "empty document — no checks at all",
			ev:         discovery_checks.StaticEvaluation{},
			wantStatus: "inconclusive",
		},
		{
			name:       "one real pass",
			ev:         discovery_checks.StaticEvaluation{Passed: []discovery_checks.StaticCheckOutcome{{ID: "a"}}},
			wantStatus: "verified",
		},
		{
			name: "a pass and a failure — a failure always wins",
			ev: discovery_checks.StaticEvaluation{
				Passed: []discovery_checks.StaticCheckOutcome{{ID: "a"}},
				Failed: []discovery_checks.StaticCheckOutcome{{ID: "b"}},
			},
			wantStatus: "broken",
		},
		{
			name: "passes plus skips is still a pass — skips are honest, not fatal",
			ev: discovery_checks.StaticEvaluation{
				Passed:  []discovery_checks.StaticCheckOutcome{{ID: "a"}},
				Skipped: []discovery_checks.StaticCheckOutcome{{ID: "b"}, {ID: "c"}},
			},
			wantStatus: "verified",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := experienceVerdict(tc.ev)
			if got != tc.wantStatus {
				t.Errorf("evaluation %+v gave %q, want %q", tc.ev, got, tc.wantStatus)
			}
		})
	}
}

func TestVerifyDecision_ProvenNeedsAPassNotJustAnApprovedEntry(t *testing.T) {
	// `proven` is the strongest word in the vocabulary and means the criteria ran
	// green against a real page. An inconclusive or broken run must never earn
	// it, however approved the entry is.
	for _, ev := range []discovery_checks.StaticEvaluation{
		{},
		{Skipped: []discovery_checks.StaticCheckOutcome{{ID: "a"}}},
		{Failed: []discovery_checks.StaticCheckOutcome{{ID: "a"}}},
	} {
		if experienceVerdict(ev) == "verified" {
			t.Errorf("a run with %d passed / %d failed earned 'verified'", len(ev.Passed), len(ev.Failed))
		}
	}
}
