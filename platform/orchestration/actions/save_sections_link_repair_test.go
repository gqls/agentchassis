package actions

import (
	"regexp"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// The save-path seam's two decision branches, without a DB. The rewrite/unlink
// semantics themselves stay covered by datahelpers' link_repair_test.go — what
// is pinned here is that the PERSISTENCE path applies them per section
// (bugs_open/079: the gate computed the repair and save_sections discarded it),
// that a clean section is left byte-identical, and that the degrade persists
// every section untouched. The failure mode of the degrade is silent damage: an
// over-eager "repair" with no usable index would strip every internal link on
// every section of the page and write that to page_components.

func TestRepairSectionLinksRepairsEachSection(t *testing.T) {
	idx := datahelpers.NewPageURLIndex([]string{"/contact.html", "/about.html"})
	sections := []SectionData{
		{ComponentName: "hero", HTML: `<section><a href="/never-built">our pricing</a></section>`},
		{ComponentName: "cta", HTML: `<section><a href="/contact">Talk to us</a></section>`},
		{ComponentName: "clean", HTML: `<section><a href="/about.html">About</a></section>`},
		{ComponentName: "empty-href", HTML: `<section><a href="">Read more</a></section>`},
	}
	cleanBefore := sections[2].HTML

	rewritten, unlinked, repairs := repairSectionLinks(sections, idx, true)

	if len(repairs) != 3 {
		t.Fatalf("expected 3 repairs (phantom, rewrite, empty href), got %d: %+v", len(repairs), repairs)
	}
	if rewritten != 1 {
		t.Errorf("expected 1 rewrite, got %d", rewritten)
	}
	if unlinked != 2 {
		t.Errorf("expected 2 unlinks (phantom + empty href), got %d", unlinked)
	}

	if strings.Contains(sections[0].HTML, `href="/never-built"`) {
		t.Errorf("phantom link must be unlinked in the persisted section: %q", sections[0].HTML)
	}
	if !strings.Contains(sections[0].HTML, "our pricing") {
		t.Errorf("unlinking must keep the anchor text — body prose is content: %q", sections[0].HTML)
	}
	if !strings.Contains(sections[1].HTML, `href="/contact.html"`) {
		t.Errorf("extension-omitted href must be rewritten to the stored url: %q", sections[1].HTML)
	}
	if sections[2].HTML != cleanBefore {
		t.Errorf("a section with no dead links must be byte-identical:\n got %q\nwant %q", sections[2].HTML, cleanBefore)
	}
	if strings.Contains(sections[3].HTML, "<a ") {
		t.Errorf(`href="" must be unlinked — it renders as a control that goes nowhere: %q`, sections[3].HTML)
	}
	if !strings.Contains(sections[3].HTML, "Read more") {
		t.Errorf("unlinking an empty href must keep its text: %q", sections[3].HTML)
	}
}

func TestRepairSectionLinksFailsOpenWithoutTrustworthyIndex(t *testing.T) {
	idx := datahelpers.NewPageURLIndex([]string{"/about.html"})
	original := []string{
		`<section><a href="/never-built">gone?</a></section>`,
		`<section><a href="/about.html">About</a></section>`,
	}

	// indexOK=false: the page query failed. An empty page set would make every
	// link on the page a phantom, so the whole pass must stand down.
	sections := []SectionData{{HTML: original[0]}, {HTML: original[1]}}
	rewritten, unlinked, repairs := repairSectionLinks(sections, idx, false)
	if len(repairs) != 0 || rewritten != 0 || unlinked != 0 {
		t.Fatalf("untrustworthy index must produce NO repairs, got %d (%d/%d)", len(repairs), rewritten, unlinked)
	}
	for i := range sections {
		if sections[i].HTML != original[i] {
			t.Errorf("fail-open: section %d must be byte-identical, got %q", i, sections[i].HTML)
		}
	}

	// A nil/empty index reaching the seam with indexOK=true is the same hazard
	// wearing a disguise, and must degrade identically.
	sections = []SectionData{{HTML: original[0]}, {HTML: original[1]}}
	if _, _, repairs = repairSectionLinks(sections, nil, true); len(repairs) != 0 {
		t.Fatalf("empty index must produce NO repairs, got %d", len(repairs))
	}
	for i := range sections {
		if sections[i].HTML != original[i] {
			t.Errorf("empty index: section %d must be byte-identical, got %q", i, sections[i].HTML)
		}
	}
}

// The runtime-fill exemption is whole-DOCUMENT in RepairPageLinks; applied per
// section it becomes whole-SECTION. That narrowing is the point: a shell whose
// hrefs are hydrated client-side still exempts itself, while its statically
// linked neighbours no longer ride on its exemption.
func TestRepairSectionLinksNarrowsRuntimeFillExemptionToItsOwnSection(t *testing.T) {
	idx := datahelpers.NewPageURLIndex([]string{"/about.html"})
	sections := []SectionData{
		{ComponentName: "results", HTML: `<section data-runtime-fill><a href="">placeholder</a></section>`},
		{ComponentName: "prose", HTML: `<section><a href="/never-built">phantom</a></section>`},
	}
	runtimeFillBefore := sections[0].HTML

	_, _, repairs := repairSectionLinks(sections, idx, true)

	if sections[0].HTML != runtimeFillBefore {
		t.Errorf("a runtime-fill section must be left alone: %q", sections[0].HTML)
	}
	if strings.Contains(sections[1].HTML, `href="/never-built"`) {
		t.Errorf("a neighbouring static section must NOT inherit the exemption: %q", sections[1].HTML)
	}
	if len(repairs) != 1 {
		t.Errorf("expected exactly the one static repair, got %d: %+v", len(repairs), repairs)
	}
}

// The repair runs BEFORE the content-regression guard and the Layer 1
// interactivity guard, so both of them measure post-repair bytes. The claim
// that this is safe — "unlinking preserves the anchor text, so the guards'
// totals do not move" — was reasoning, not a check, and the council was right
// that this is the same shape of claim that "one call site" and "the ordering
// is at line N" turned out to be. So it is a test now.
//
// The two guards' inputs are pure functions of SectionData.HTML, which is what
// makes this testable without a DB: the content-regression guard computes
// len(TrimSpace(regexp `<[^>]*>` -> "")) per section and sums it
// (save_page_sections_action.go, "Content regression guard"), and the
// interactivity guard calls sectionHTMLIsInteractive per section. Both are
// reproduced here against the same expressions.
func TestRepairSectionLinksLeavesBothGuardMeasurementsInvariant(t *testing.T) {
	idx := datahelpers.NewPageURLIndex([]string{"/contact.html", "/about.html"})
	sections := []SectionData{
		{ComponentName: "prose", HTML: `<section><p>Read <a href="/never-built">our pricing</a> or <a href="/contact">talk to us</a>.</p></section>`},
		{ComponentName: "empty", HTML: `<section><a href="">Read more</a></section>`},
		{ComponentName: "tool", HTML: `<section class="tool-page"><canvas id="board"></canvas><a href="/never-built">rules</a></section>`},
		{ComponentName: "clean", HTML: `<section><a href="/about.html">About</a></section>`},
	}

	// The guards' own arithmetic, before the repair.
	tagStripper := regexp.MustCompile(`<[^>]*>`)
	strippedLen := func(s SectionData) int {
		return len(strings.TrimSpace(tagStripper.ReplaceAllString(s.HTML, "")))
	}
	beforeLen := make([]int, len(sections))
	beforeInteractive := make([]bool, len(sections))
	beforeTotal := 0
	for i, s := range sections {
		beforeLen[i] = strippedLen(s)
		beforeInteractive[i] = sectionHTMLIsInteractive(s.HTML)
		beforeTotal += beforeLen[i]
	}

	_, _, repairs := repairSectionLinks(sections, idx, true)
	if len(repairs) == 0 {
		t.Fatal("fixture must actually repair something, or this test asserts nothing")
	}

	afterTotal := 0
	for i, s := range sections {
		if got := strippedLen(s); got != beforeLen[i] {
			t.Errorf("section %d (%s): the content-regression guard's stripped-text length moved %d -> %d; unlinking must keep the text",
				i, s.ComponentName, beforeLen[i], got)
		}
		if got := sectionHTMLIsInteractive(s.HTML); got != beforeInteractive[i] {
			t.Errorf("section %d (%s): the interactivity guard's verdict moved %v -> %v; repair must not touch canvas/game-container/tool-page markers",
				i, s.ComponentName, beforeInteractive[i], got)
		}
		afterTotal += strippedLen(s)
	}
	if afterTotal != beforeTotal {
		t.Errorf("the content-regression guard's TOTAL moved %d -> %d — a repaired page could now trip a guard it used to pass", beforeTotal, afterTotal)
	}
	if !beforeInteractive[2] {
		t.Fatal("fixture bug: the tool section must read as interactive before the repair, or the invariant above is vacuous")
	}
}
