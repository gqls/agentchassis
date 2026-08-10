// FILE: platform/orchestration/actions/write_site_plan_dedup_test.go
//
// bugs_open/215 — two emitted pages canonicalise to ONE name and the whole
// replan write dies on the unique index.
//
// These tests cover two layers, deliberately:
//
//  1. TestCanonicalisePage_CollapseFamilies pins the PREMISE — that distinct
//     LLM spellings really do collapse onto one canonical name. If the
//     canonicaliser ever stops collapsing, the dedup below becomes dead code
//     and this test says so rather than leaving it quietly inert.
//  2. The dedupePlanPageRows tests cover the FIX itself.
//
// The fixture spellings are the ones from the live 2026-08-08 incident and
// its 2026-08-07 predecessor (fundamentallyai): tool-prefixed twins, the
// tools/tool-tools pair, and the homepage/section-index families found by
// reading page_canonical.go rather than by waiting for them to bite.
package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// TestCanonicalisePage_CollapseFamilies asserts the premise the dedup exists
// to handle: two DIFFERENT descriptors yielding the SAME canonical name.
func TestCanonicalisePage_CollapseFamilies(t *testing.T) {
	cases := []struct {
		family string
		a, b   datahelpers.PageDescriptor
		want   string
	}{
		{
			family: "tool prefix collapse (the live 2026-08-08 collision)",
			a:      datahelpers.PageDescriptor{Role: "tool", Slug: "llm-cost-calculator"},
			b:      datahelpers.PageDescriptor{Role: "tool", Slug: "tool-llm-cost-calculator"},
			want:   "tool-llm-cost-calculator",
		},
		{
			family: "guide prefix collapse",
			a:      datahelpers.PageDescriptor{Role: "guide", Slug: "ai-readiness"},
			b:      datahelpers.PageDescriptor{Role: "guide", Slug: "guide-ai-readiness"},
			want:   "guide-ai-readiness",
		},
		{
			family: "homepage collapse",
			a:      datahelpers.PageDescriptor{Role: "index", Slug: "whatever"},
			b:      datahelpers.PageDescriptor{Role: "content", Slug: "home"},
			want:   "index",
		},
		{
			family: "section-index collapse",
			a:      datahelpers.PageDescriptor{Role: "section-index", Slug: "guides"},
			b:      datahelpers.PageDescriptor{Role: "section-index", Slug: "guides-index"},
			want:   "guides-index",
		},
	}

	for _, tc := range cases {
		t.Run(tc.family, func(t *testing.T) {
			nameA, _, _ := datahelpers.CanonicalisePage(tc.a)
			nameB, _, _ := datahelpers.CanonicalisePage(tc.b)
			if nameA != tc.want || nameB != tc.want {
				t.Fatalf("expected both descriptors to canonicalise to %q, got %q and %q\n"+
					"if this now differs, the collision family changed — re-check dedupePlanPageRows is still needed",
					tc.want, nameA, nameB)
			}
		})
	}
}

func sec(names ...string) []sectionEntry {
	out := make([]sectionEntry, 0, len(names))
	for _, n := range names {
		out = append(out, sectionEntry{Name: n})
	}
	return out
}

// TestDedupePlanPageRows_StubLosesToComposedPage is the live incident's exact
// shape: a composed page and a zero-section stub of the same identity.
func TestDedupePlanPageRows_StubLosesToComposedPage(t *testing.T) {
	// Stub emitted FIRST, so "keep the first" would be the wrong answer and
	// only a section-count rule passes.
	rows := []planPageRow{
		{RawName: "tool-llm-cost-calculator", Name: "tool-llm-cost-calculator", Role: "tool"},
		{RawName: "llm-cost-calculator", Name: "tool-llm-cost-calculator", Role: "tool",
			Sections: sec("hero", "calculator", "faq")},
	}

	out, merges, _ := dedupePlanPageRows(rows, zap.NewNop())

	if len(out) != 1 {
		t.Fatalf("expected 1 surviving row, got %d", len(out))
	}
	if merges != 1 {
		t.Fatalf("expected 1 merge, got %d", merges)
	}
	if got := len(out[0].Sections); got != 3 {
		t.Fatalf("the composed entry must win over the stub: expected 3 sections, got %d", got)
	}
	if out[0].RawName != "llm-cost-calculator" {
		t.Fatalf("survivor should be the composed entry, got raw name %q", out[0].RawName)
	}
}

// TestDedupePlanPageRows_PreservesOrderAndDistinctPages guards the no-op case:
// a plan with no collisions must come through completely untouched. A dedup
// that quietly dropped or reordered good pages would be far worse than the
// bug it fixes.
func TestDedupePlanPageRows_PreservesOrderAndDistinctPages(t *testing.T) {
	rows := []planPageRow{
		{Name: "index", Sections: sec("hero")},
		{Name: "tool-llm-cost-calculator", Sections: sec("calculator")},
		{Name: "guides-index", Sections: sec("list")},
		{Name: "contact", Sections: sec("form")},
	}

	out, merges, _ := dedupePlanPageRows(rows, zap.NewNop())

	if merges != 0 {
		t.Fatalf("no collisions present, expected 0 merges, got %d", merges)
	}
	if len(out) != len(rows) {
		t.Fatalf("expected all %d pages preserved, got %d", len(rows), len(out))
	}
	for i := range rows {
		if out[i].Name != rows[i].Name {
			t.Fatalf("order changed at %d: expected %q, got %q", i, rows[i].Name, out[i].Name)
		}
	}
}

// TestDedupePlanPageRows_TwoComposedTieKeepsFirst covers the lossy branch on a
// genuine TIE: both entries carry the same number of sections, so one list IS
// discarded and the first must win. Pinned so the choice stays deliberate.
func TestDedupePlanPageRows_TwoComposedTieKeepsFirst(t *testing.T) {
	rows := []planPageRow{
		{RawName: "tools", Name: "tools", Sections: sec("intro", "grid")},
		{RawName: "tool-tools", Name: "tools", Sections: sec("hero", "list")},
	}

	out, merges, _ := dedupePlanPageRows(rows, zap.NewNop())

	if len(out) != 1 || merges != 1 {
		t.Fatalf("expected 1 row and 1 merge, got %d rows, %d merges", len(out), merges)
	}
	if out[0].RawName != "tools" {
		t.Fatalf("tie must keep the FIRST entry, got %q", out[0].RawName)
	}
	if len(out[0].Sections) != 2 || out[0].Sections[0].Name != "intro" {
		t.Fatalf("first entry's own sections must survive intact, got %v", sectionNames(out[0].Sections))
	}
}

// TestDedupePlanPageRows_RicherComposedWinsEvenWhenSecond is the case that
// caught a real defect while this fix was being written: an earlier draft let
// the "both are composed" branch force first-wins, which discarded the RICHER
// page. The count rule decides; the loud log only reports the loss.
func TestDedupePlanPageRows_RicherComposedWinsEvenWhenSecond(t *testing.T) {
	rows := []planPageRow{
		{RawName: "tools", Name: "tools", Sections: sec("intro", "grid")},
		{RawName: "tool-tools", Name: "tools", Sections: sec("hero", "list", "cta")},
	}

	out, _, _ := dedupePlanPageRows(rows, zap.NewNop())

	if len(out) != 1 {
		t.Fatalf("expected 1 row, got %d", len(out))
	}
	if out[0].RawName != "tool-tools" || len(out[0].Sections) != 3 {
		t.Fatalf("the richer composed entry must win regardless of position: got %q with %v",
			out[0].RawName, sectionNames(out[0].Sections))
	}
}

// TestDedupePlanPageRows_BackfillsOnlyBlankMetadata: the backfill may fill a
// blank, never overwrite an authored value.
func TestDedupePlanPageRows_BackfillsOnlyBlankMetadata(t *testing.T) {
	rows := []planPageRow{
		// Stub carries metadata the composed entry lacks, plus a title that
		// must NOT win.
		{RawName: "stub", Name: "tool-x", Title: "Stub Title",
			MetaDescription: "from the stub", NavLabel: "Stub Nav"},
		{RawName: "composed", Name: "tool-x", Title: "Authored Title",
			Sections: sec("hero")},
	}

	out, _, _ := dedupePlanPageRows(rows, zap.NewNop())

	if len(out) != 1 {
		t.Fatalf("expected 1 row, got %d", len(out))
	}
	if out[0].Title != "Authored Title" {
		t.Fatalf("backfill must never overwrite an authored value, got %q", out[0].Title)
	}
	if out[0].MetaDescription != "from the stub" {
		t.Fatalf("blank meta_description should be backfilled, got %q", out[0].MetaDescription)
	}
	if out[0].NavLabel != "Stub Nav" {
		t.Fatalf("blank nav_label should be backfilled, got %q", out[0].NavLabel)
	}
}

// TestDedupePlanPageRows_ThreeWayAndMultipleCollisions checks the counter and
// the map-based survivor tracking under more than one collision, including a
// name colliding three times.
func TestDedupePlanPageRows_ThreeWayAndMultipleCollisions(t *testing.T) {
	rows := []planPageRow{
		{RawName: "home", Name: "index"},
		{RawName: "index", Name: "index", Sections: sec("hero")},
		{RawName: "landing", Name: "index", Sections: sec("hero", "proof")},
		{RawName: "guides", Name: "guides-index", Sections: sec("list")},
		{RawName: "guides-index", Name: "guides-index"},
		{RawName: "contact", Name: "contact", Sections: sec("form")},
	}

	out, merges, _ := dedupePlanPageRows(rows, zap.NewNop())

	if merges != 3 {
		t.Fatalf("expected 3 merges (2 onto index, 1 onto guides-index), got %d", merges)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 surviving pages, got %d: %v", len(out), namesOf(out))
	}
	byName := map[string]planPageRow{}
	for _, r := range out {
		byName[r.Name] = r
	}
	if got := len(byName["index"].Sections); got != 2 {
		t.Fatalf("richest index entry must win: expected 2 sections, got %d", got)
	}
	if got := len(byName["guides-index"].Sections); got != 1 {
		t.Fatalf("composed guides entry must beat the later stub, got %d sections", got)
	}
	if got := len(byName["contact"].Sections); got != 1 {
		t.Fatalf("uncollided page must be untouched, got %d sections", got)
	}
}

// TestDedupePlanPageRows_NoDuplicateNamesSurvive is the property the two
// unique indexes actually require: whatever goes in, no name may appear twice
// on the way out. site_plan_pages is UNIQUE(plan_id, name) and
// site_plan_sections is UNIQUE(plan_id, page_name, ordering), so a survivor
// pair aborts the entire transaction.
func TestDedupePlanPageRows_NoDuplicateNamesSurvive(t *testing.T) {
	rows := []planPageRow{
		{Name: "index"}, {Name: "index", Sections: sec("a")},
		{Name: "index", Sections: sec("a", "b")},
		{Name: "tool-x"}, {Name: "tool-x"},
		{Name: "guides-index", Sections: sec("l")},
		{Name: "contact"},
	}

	out, _, _ := dedupePlanPageRows(rows, zap.NewNop())

	seen := map[string]bool{}
	for _, r := range out {
		if seen[r.Name] {
			t.Fatalf("duplicate name %q survived — this is exactly the state the "+
				"unique index rejects, and it would abort the whole plan write", r.Name)
		}
		seen[r.Name] = true
	}
}

func namesOf(rows []planPageRow) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Name)
	}
	return out
}

// A merge that discards a composed page's sections must come back as a lossy
// detail for the caller to persist (owner ruling 2026-08-10 on bugs_open/215:
// richer-wins is ratified on condition the loss is durably recorded — a
// chassis Warn rotates away in under a second). A stub merge is not lossy and
// must not be reported as one.
func TestDedupePlanPageRows_LossyMergeDetailReturned(t *testing.T) {
	rows := []planPageRow{
		{RawName: "tools", Name: "tools-index", Role: "section_index",
			Sections: sec("hero", "tool-grid")},
		{RawName: "tools-index", Name: "tools-index", Role: "section_index",
			Sections: sec("hero", "tool-grid", "faq")},
	}
	out, merges, lossy := dedupePlanPageRows(rows, zap.NewNop())
	if len(out) != 1 || merges != 1 {
		t.Fatalf("out=%d merges=%d, want 1/1", len(out), merges)
	}
	if len(lossy) != 1 {
		t.Fatalf("a composed-vs-composed merge must return a lossy detail, got %#v", lossy)
	}
	l := lossy[0]
	if l.CanonicalName != "tools-index" || l.KeptRawName != "tools-index" || l.DroppedRawName != "tools" {
		t.Errorf("lossy identity = %+v, want kept=tools-index dropped=tools under tools-index", l)
	}
	if len(l.KeptSections) != 3 || len(l.DroppedSections) != 2 {
		t.Errorf("lossy sections = kept %v / dropped %v, want the FULL lists so the loss is reconstructable", l.KeptSections, l.DroppedSections)
	}
}

func TestDedupePlanPageRows_StubMergeIsNotLossy(t *testing.T) {
	rows := []planPageRow{
		{RawName: "llm-cost-calculator", Name: "tool-llm-cost-calculator", Role: "tool",
			Sections: sec("hero", "calculator")},
		{RawName: "tool-llm-cost-calculator", Name: "tool-llm-cost-calculator", Role: "tool"},
	}
	_, merges, lossy := dedupePlanPageRows(rows, zap.NewNop())
	if merges != 1 {
		t.Fatalf("merges=%d, want 1", merges)
	}
	if len(lossy) != 0 {
		t.Errorf("a stub merge discards nothing and must not be reported lossy, got %#v", lossy)
	}
}
