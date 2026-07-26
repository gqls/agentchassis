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
	if v := hardcodedSectionColoursVerdict([]RemitCandidate{
		{Key: "home/hero", Body: light},
	}); !v.Resolved {
		t.Errorf("out-of-remit candidate: want resolved, got %+v", v)
	}

	v := hardcodedSectionColoursVerdict([]RemitCandidate{
		{Key: "home/hero", Body: light},
		{Key: "about/cta", Body: dark},
	})
	if v.Resolved {
		t.Errorf("within-remit candidate: want unresolved, got %+v", v)
	}
	if !strings.Contains(v.Detail, "1 component") || !strings.Contains(v.Detail, "about/cta") {
		t.Errorf("detail should count only within-remit components and name one, got %q", v.Detail)
	}
}

// TestHardcodedColourPartitionIsWhatGetsFiled is bugs_open/077's acceptance
// criterion as a test: the check must file a dispatchable item ONLY for what the
// handler can change, and must not lose the rest.
//
// The measured motivation, 2026-07-26: on finetuning.uk 8 components matched the
// detector and 0 were inside the fixer's remit; likewise gaswholesalers.com 6/0,
// ai-agent-orchestration.com 4/0, dartsonline.com 1/0. Under the old code each of
// those sites carried an item labelled "[unresolved after 2 attempts]" — blaming
// a handler that had never had anything to do.
func TestHardcodedColourPartitionIsWhatGetsFiled(t *testing.T) {
	dark := `<style>.s{background: #1a2b3c;}</style>`
	light := `<style>.s{background: #f5f5f5;}</style>`
	inlineDark := `<div style="background: #1a2b3c;"><style>.x{color:red}</style></div>`

	cases := []struct {
		name        string
		population  []RemitCandidate
		wantInRemit int
		wantResidue int
	}{
		{"the zero-remit site — 077's whole point", []RemitCandidate{
			{Key: "home/hero", Body: light},
			{Key: "home/cta", Body: inlineDark},
		}, 0, 2},
		{"the fully-fixable site", []RemitCandidate{
			{Key: "home/hero", Body: dark},
		}, 1, 0},
		{"mixed — both items get filed", []RemitCandidate{
			{Key: "home/hero", Body: dark},
			{Key: "about/band", Body: light},
			{Key: "about/cta", Body: inlineDark},
		}, 1, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inRemit, residue := PartitionByRemit(tc.population, ReplaceHardcodedColors)
			if len(inRemit) != tc.wantInRemit {
				t.Errorf("in remit = %d, want %d", len(inRemit), tc.wantInRemit)
			}
			if len(residue) != tc.wantResidue {
				t.Errorf("residue = %d, want %d", len(residue), tc.wantResidue)
			}
			// Nothing may be dropped: a check that partitions and then forgets half
			// its population is a quieter version of the bug it is fixing.
			if len(inRemit)+len(residue) != len(tc.population) {
				t.Errorf("partition lost candidates: %d + %d != %d",
					len(inRemit), len(residue), len(tc.population))
			}
			// The verifier must agree with the half the item was filed against, or a
			// completion is judged by a different rule than the one that filed it.
			resolved := hardcodedSectionColoursVerdict(tc.population).Resolved
			if resolved != (len(inRemit) == 0) {
				t.Errorf("verifier resolved=%v but in-remit count is %d — the two ends disagree",
					resolved, len(inRemit))
			}
		})
	}
}

// TestCapabilityGapItemIsUndispatchable pins the two independent guards that stop
// a residue item being handed to a handler that cannot act on it. If either is
// ever relaxed, bugs_open/077 comes back wearing a different item type.
func TestCapabilityGapItemIsUndispatchable(t *testing.T) {
	item := CapabilityGapItem(DiscoveryCheckContext{}, CapabilityGap{
		Check:         "hardcoded_section_colors",
		Pipeline:      "design",
		BuilderNeeded: "color-variable-fixer",
		GapKind:       GapHandlerRemit,
		Population:    8,
		Residue:       8,
		Examples:      []RemitCandidate{{Key: "home/hero"}},
	})

	// claim_work_item_action.go and load_work_item_actions.go both filter
	// status IN ('triaged','approved'); 'deferred' is in neither.
	if item.Status != "deferred" {
		t.Errorf("status = %q, want deferred — anything dispatchable re-creates 077", item.Status)
	}
	if item.HandlerAgent != "" {
		t.Errorf("handler_agent = %q, want empty — naming a real agent invites promotion", item.HandlerAgent)
	}
	if item.ItemType != "capability_gap" {
		t.Errorf("item_type = %q, want capability_gap", item.ItemType)
	}
	// One open row per site per check: idx_swi_dedup is UNIQUE on
	// (site_id, item_key) excluding terminal statuses, and 'deferred' is not one.
	if item.ItemKey != "capability_gap:hardcoded_section_colors" {
		t.Errorf("item_key = %q — must be per-check so the population stays bounded", item.ItemKey)
	}
	// The summary is the whole legibility fix. It must not read as handler failure.
	if !strings.Contains(item.Summary, "capability gap, not a handler failure") {
		t.Errorf("summary must say it is not a handler failure, got %q", item.Summary)
	}
	if !strings.Contains(item.SpecJSON, `"builder_needed":"color-variable-fixer"`) {
		t.Errorf("spec must carry builder_needed — it is the roadmap grouping key, got %s", item.SpecJSON)
	}
}

func TestCapabilityGapSummaryDistinguishesMissingFromNarrow(t *testing.T) {
	missing := capabilityGapSummary(CapabilityGap{
		GapKind: GapHandlerMissing, Check: "forced_text_colors",
		BuilderNeeded: "forced-text-color-fixer", Residue: 3, Population: 3,
	})
	if !strings.Contains(missing, "not registered") {
		t.Errorf("a missing handler must say so, got %q", missing)
	}

	narrow := capabilityGapSummary(CapabilityGap{
		GapKind: GapHandlerRemit, Check: "hardcoded_section_colors",
		BuilderNeeded: "color-variable-fixer", Residue: 8, Population: 8,
	})
	if strings.Contains(narrow, "not registered") {
		t.Errorf("a narrow remit must not be reported as a missing agent, got %q", narrow)
	}
}
