// Tests for the spec finding cap (bugs_open/332, 2026-09-03).
//
// The cap exists because this check was the only one of its family writing an
// unbounded finding list into a work item spec, and 332 widens the pattern set.
// The two properties below are what make it safe rather than merely smaller:
// the reported total stays exact, and ROUTING never sees the capped view.

package discovery_checks

import "testing"

func mdFindings(n int, pattern string) []literalMarkdownFinding {
	out := make([]literalMarkdownFinding, n)
	for i := range out {
		out[i] = literalMarkdownFinding{
			SlotName: "news-listing",
			Pattern:  pattern,
			Source:   "rendered_html",
			Matched:  "x",
		}
	}
	return out
}

func TestCapSpecFindingsLeavesShortListsAlone(t *testing.T) {
	in := mdFindings(maxLiteralMarkdownSpecFindings, "bold")
	got := capSpecFindings(in, "page", nil)
	if len(got) != len(in) {
		t.Errorf("a list AT the cap must pass through whole: got %d, want %d", len(got), len(in))
	}
	// Exactly at the boundary is the case an off-by-one gets wrong, so assert
	// identity rather than only length.
	if &got[0] != &in[0] {
		t.Errorf("an uncapped list must be the same slice, not a copy")
	}
}

func TestCapSpecFindingsBoundsLongLists(t *testing.T) {
	in := mdFindings(maxLiteralMarkdownSpecFindings+7, "md_link")
	got := capSpecFindings(in, "page", nil)
	if len(got) != maxLiteralMarkdownSpecFindings {
		t.Errorf("capped length = %d, want %d", len(got), maxLiteralMarkdownSpecFindings)
	}
	// The INPUT must be untouched: the caller reports len(pf.Findings) as
	// findings_total and passes the same slice to transformRouteSlot. A cap that
	// truncated in place would silently make both wrong.
	if len(in) != maxLiteralMarkdownSpecFindings+7 {
		t.Errorf("cap mutated its input: len(in) = %d, want %d", len(in), maxLiteralMarkdownSpecFindings+7)
	}
}

// THE HAZARD IS REAL, AND THIS TEST IS WHAT PROVES IT.
//
// The caller's comment says routing must read the full slice, never the capped
// one. A comment enforces nothing, and the usual way to test this would be to
// assert that the capped slice routes the same as the full one — which would
// pass VACUOUSLY on any input where they happen to agree, i.e. almost all of
// them. So assert the opposite: construct the input where they DISAGREE, and
// fail if they ever stop disagreeing.
//
// If this test starts failing, the fixture no longer exercises the hazard and
// the comment above capSpecFindings has quietly become decorative — fix the
// fixture, do not delete the test.
//
// The shape: 25 code_span findings then one bold. transformRouteSlot asks "is
// EVERY finding a code_span?" The full slice answers no, correctly. The capped
// view answers yes, and would send the page to a transform that cannot repair
// what finding 26 found — an item that fails for a reason nothing records.
func TestCappedAndFullFindingsRouteDIFFERENTLY(t *testing.T) {
	full := append(mdFindings(maxLiteralMarkdownSpecFindings, "code_span"), mdFindings(1, "bold")...)
	slots := map[string]*slotRepro{"news-listing": {occurrences: 1, canRegenerate: false}}

	if _, ok := transformRouteSlot(full, slots); ok {
		t.Fatalf("premise wrong: the full slice carries a non-code_span finding and must NOT route")
	}
	if _, ok := transformRouteSlot(capSpecFindings(full, "page", nil), slots); !ok {
		t.Fatalf("premise wrong: the capped view is all code_span on one single-occurrence " +
			"unregenerable slot and must route — if it does not, this fixture no longer " +
			"demonstrates the hazard the cap's comment describes")
	}
}
