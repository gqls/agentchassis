package discovery_checks

import (
	"testing"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// sec builds a section with pre-normalised text, as loadSiteSections would.
func sec(page uuid.UUID, pageName, slot string, pos int, rawJSON string) siteSection {
	text := datahelpers.NormaliseSectionText(rawJSON)
	return siteSection{
		ComponentID: uuid.New(),
		PageID:      page,
		PageName:    pageName,
		SlotName:    slot,
		Position:    pos,
		Text:        text,
		Tokens:      datahelpers.SectionTokenSet(text),
	}
}

const longA = `{"headline":"The rules are simple","body":"Holding your position is not. Every day one provocation drops into the arena and you defend it against an opponent on a clock, with a judge at the end who does not care how you feel about it."}`
const longB = `{"headline":"A provocation, not a prompt","body":"Every day one arguable statement drops into the arena. It is not a question and it is not a chat window; it is a position you either hold or lose on a twenty minute clock."}`

// THE DEFECT THIS CHECK EXISTS FOR: vonc.com's about page carried 6 pairs of
// content-identical rows and rendered every section twice for two days.
func TestFindIdenticalSamePage_FlagsContentIdenticalPairs(t *testing.T) {
	page := uuid.New()
	sections := []siteSection{
		sec(page, "about", "hero-about", 1, longA),
		sec(page, "about", "hero-about", 2, longA), // the duplicate
		sec(page, "about", "differentiators", 3, longB),
	}
	groups := findIdenticalSamePage(sections)
	if len(groups) != 1 {
		t.Fatalf("want 1 affected page, got %d", len(groups))
	}
	if got := len(groups[0].Groups); got != 1 {
		t.Fatalf("want 1 duplicate group, got %d", got)
	}
	redundant := groups[0].redundantComponentIDs()
	if len(redundant) != 1 {
		t.Fatalf("want exactly 1 redundant row, got %d", len(redundant))
	}
	// It must keep the EARLIEST and mark the later one. Keeping the later row
	// would silently reorder the page.
	if redundant[0] != sections[1].ComponentID.String() {
		t.Errorf("want the position-2 row marked redundant, got %s", redundant[0])
	}
	if groups[0].Keep.Position != 1 {
		t.Errorf("want to keep position 1, keeping %d", groups[0].Keep.Position)
	}
}

// THE FALSE POSITIVE THAT WOULD DELETE REAL CONTENT.
// Fleet-wide, 17 (page_id, slot_name) duplicate groups exist and 11 are
// legitimate — repeated slots carrying DIFFERENT content. A check keyed on slot
// name would delete real sections on five live sites (bugs_open/156).
func TestFindIdenticalSamePage_IgnoresRepeatedSlotWithDifferentContent(t *testing.T) {
	page := uuid.New()
	sections := []siteSection{
		sec(page, "index", "generic-text-block", 1, longA),
		sec(page, "index", "generic-text-block", 2, longB), // same slot, different words
	}
	if groups := findIdenticalSamePage(sections); len(groups) != 0 {
		t.Fatalf("repeated slot with differing content must NOT be flagged; got %d group(s)", len(groups))
	}
}

// Two sections on DIFFERENT pages are never in-remit, however identical. The
// deterministic repair deletes rows from one page; applying it across pages would
// remove a page's only copy of something.
func TestFindIdenticalSamePage_NeverCrossesPages(t *testing.T) {
	sections := []siteSection{
		sec(uuid.New(), "index", "cta", 1, longA),
		sec(uuid.New(), "about", "cta", 1, longA),
	}
	if groups := findIdenticalSamePage(sections); len(groups) != 0 {
		t.Fatalf("cross-page identity is residue, not in-remit; got %d group(s)", len(groups))
	}
}

// Cross-page identity IS residue, and residue must never carry a handler.
func TestFindResidue_CountsCrossPageIdentical(t *testing.T) {
	sections := []siteSection{
		sec(uuid.New(), "index", "cta", 1, longA),
		sec(uuid.New(), "about", "cta", 1, longA),
	}
	rep := findResidue(sections, nil)
	if rep.CrossPageIdentical != 1 {
		t.Fatalf("want 1 cross-page identical pair, got %d", rep.CrossPageIdentical)
	}
	if rep.total() == 0 {
		t.Fatal("residue total must be non-zero when a pair was found")
	}
}

// A small evidence base must be REPORTED as blind, not reported as clean.
// This is the trap measured on vonc: 4 approved facts, so no section can reach a
// 3-fact overlap, and a clean fact census reads identically to a clean site.
func TestFindResidue_FlagsFactCensusBlindBelowFloor(t *testing.T) {
	sections := []siteSection{sec(uuid.New(), "index", "hero", 1, longA)}
	rep := findResidue(sections, []string{"a", "b", "c", "d"})
	if !rep.FactCensusBlind {
		t.Fatal("4 facts is below the floor; the report must say the fact census was blind")
	}
	repOK := findResidue(sections, []string{"a", "b", "c", "d", "e", "f", "g"})
	if repOK.FactCensusBlind {
		t.Fatal("7 facts is above the floor; must not claim blindness")
	}
}

// Short sections are skipped. Without this, boilerplate one-liners ("Read more")
// make every page look duplicated and the report becomes noise.
func TestFindIdenticalSamePage_SkipsShortSections(t *testing.T) {
	page := uuid.New()
	sections := []siteSection{
		sec(page, "index", "note", 1, `{"body":"Read more"}`),
		sec(page, "index", "note", 2, `{"body":"Read more"}`),
	}
	if groups := findIdenticalSamePage(sections); len(groups) != 0 {
		t.Fatalf("sections under the length floor must be skipped; got %d group(s)", len(groups))
	}
}

// The normaliser must ignore asset/identifier keys. Two sections differing only
// by their image are NOT saying the same thing; conversely, captured chrome text
// once made two unrelated vonc sections match at 1.00 before this filter existed.
func TestNormaliseSectionText_IgnoresAssetKeys(t *testing.T) {
	a := datahelpers.NormaliseSectionText(`{"headline":"Same words here","image_url":"/a.png"}`)
	b := datahelpers.NormaliseSectionText(`{"headline":"Same words here","image_url":"/b.png"}`)
	if a != b {
		t.Errorf("differing image_url must not change the compared text:\n a=%q\n b=%q", a, b)
	}
	if a == "" {
		t.Fatal("headline text must survive normalisation")
	}
}

// Map iteration order in Go is randomised. Without the sort in the normaliser the
// same content_data would normalise differently between runs and "identical"
// would be a coin toss — so this asserts stability, not just correctness.
func TestNormaliseSectionText_IsStableAcrossRuns(t *testing.T) {
	raw := `{"z":"zebra section text goes here","a":"alpha section text goes here","m":"middle section text"}`
	first := datahelpers.NormaliseSectionText(raw)
	for i := 0; i < 50; i++ {
		if got := datahelpers.NormaliseSectionText(raw); got != first {
			t.Fatalf("unstable normalisation on run %d:\n want %q\n got  %q", i, first, got)
		}
	}
}

// Unparseable JSON must yield "" and callers must treat that as "no comparable
// text" — if it returned a shared sentinel, every broken blob would be identical
// to every other broken blob and they would all be flagged as duplicates.
func TestNormaliseSectionText_UnparseableIsEmpty(t *testing.T) {
	if got := datahelpers.NormaliseSectionText(`{not json`); got != "" {
		t.Fatalf("want empty string for unparseable JSON, got %q", got)
	}
	page := uuid.New()
	sections := []siteSection{
		sec(page, "index", "a", 1, `{not json`),
		sec(page, "index", "b", 2, `{also not json`),
	}
	if groups := findIdenticalSamePage(sections); len(groups) != 0 {
		t.Fatal("two unparseable blobs must not be flagged as duplicates of each other")
	}
}

func TestJaccard(t *testing.T) {
	a := datahelpers.SectionTokenSet("alpha beta gamma delta")
	if got := jaccard(a, a); got != 1.0 {
		t.Errorf("identical sets must be 1.0, got %v", got)
	}
	b := datahelpers.SectionTokenSet("epsilon zeta theta iota")
	if got := jaccard(a, b); got != 0 {
		t.Errorf("disjoint sets must be 0, got %v", got)
	}
	if got := jaccard(a, map[string]struct{}{}); got != 0 {
		t.Errorf("empty set must be 0, not NaN, got %v", got)
	}
}
