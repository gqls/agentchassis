// FILE: platform/orchestration/actions/discovery_checks/check_hardcoded_section_colors_test.go
//
// The discriminator cases encode the remit-vs-detector trap (WRONG_CALLS.md
// 2026-07-20 — the page_rerender verifier was written, tested and HELD over
// exactly this): several fixtures MATCH the detector's broad predicate (any hex
// background + a <style> tag somewhere in the component) yet are OUTSIDE the
// handler's remit, and the verifier must call them resolved. If one of those
// starts failing, the transform's remit has widened — re-read the verdict's
// assumptions before "fixing" the test.

package discovery_checks

import (
	"strings"
	"testing"
)

func TestReplaceHardcodedColorsRemit(t *testing.T) {
	cases := []struct {
		name        string
		html        string
		withinRemit bool // true → the handler's transform would change it
	}{
		{"dark 6-digit background in style block",
			`<style>.s{background: #1a2b3c;}</style>`, true},
		{"dark 6-digit background-color in style block",
			`<style>.s{background-color: #0f172a;}</style>`, true},
		{"two-colour deg gradient in style block",
			`<style>.h{background: linear-gradient(135deg, #0a1a2a, #2a3a4a);}</style>`, true},

		// Detector matches ALL of these; the handler deliberately leaves them.
		{"light hex background — dark-only rule",
			`<style>.s{background: #f5f5f5;}</style>`, false},
		{"3-digit hex background — 6-digit-only rule",
			`<style>.s{background: #333;}</style>`, false},
		{"dark hex in inline style attribute — transform only enters <style> blocks",
			`<div style="background: #1a2b3c;"><style>.x{color:red}</style></div>`, false},
		{"already using variables",
			`<style>.s{background: var(--color-primary);}</style>`, false},
	}
	for _, tc := range cases {
		got := ReplaceHardcodedColors(tc.html) != tc.html
		if got != tc.withinRemit {
			t.Errorf("%s: transform changed=%v, want %v", tc.name, got, tc.withinRemit)
		}
	}
}

func TestHardcodedSectionColoursVerdict(t *testing.T) {
	dark := `<style>.s{background: #1a2b3c;}</style>`
	light := `<style>.s{background: #f5f5f5;}</style>`

	// An aggregate item with an empty sweep is the defect gone (no 032-style
	// missing-target ambiguity here — there is no single target to miss).
	if v := hardcodedSectionColoursVerdict(nil); !v.Resolved {
		t.Errorf("no candidates: want resolved, got %+v", v)
	}

	// The trap discriminator: in the DETECTOR's population, outside the
	// HANDLER's remit → must resolve, or every such item strands in 'failed'.
	if v := hardcodedSectionColoursVerdict([]hardcodedColourCandidate{
		{Page: "home", Slot: "hero", HTML: light},
	}); !v.Resolved {
		t.Errorf("out-of-remit candidate: want resolved, got %+v", v)
	}

	v := hardcodedSectionColoursVerdict([]hardcodedColourCandidate{
		{Page: "home", Slot: "hero", HTML: light},
		{Page: "about", Slot: "cta", HTML: dark},
	})
	if v.Resolved {
		t.Errorf("within-remit candidate: want unresolved, got %+v", v)
	}
	if !strings.Contains(v.Detail, "1 component") || !strings.Contains(v.Detail, "about/cta") {
		t.Errorf("detail should count only within-remit components and name one, got %q", v.Detail)
	}
}
