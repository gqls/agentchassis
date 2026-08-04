package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

// Tests for the duplicate-section collapse at the save choke point
// (bugs_open/156). The behaviour lives in the pure seam, following
// scanSectionClaims and evaluateSectionShrink — the DB seam only reads the plan
// store and writes a record, and the package landmine recorded in
// save_sections_prune_floor_test.go applies here too (a best-effort writer
// swallows sqlmock's unexpected-call error, so "no INSERT was issued" is a
// vacuous assertion; the negatives are pinned at the pure seam instead).
//
// EVERY TEST HERE NAMES THE MUTATION THAT MAKES IT FAIL, because a test that
// could not have come out otherwise is not evidence. The mutations were run, not
// imagined; the two that matter most are:
//
//   - TestCollapseLeavesLegitimateRepeatedSlotsAlone fails if the key is reduced
//     to slot_name — i.e. it fails under the unique-index-on-(page_id, slot_name)
//     rule that bugs_open/156 records as the WRONG fix. 11 live pages depend on
//     that distinction, so this is the load-bearing negative control.
//   - TestCollapseLeavesNullContentDataWithDifferentHTMLAlone fails if
//     rendered_html is dropped from the key — i.e. it fails under the key the bug
//     file's own candidate 1 literally proposes. That shape (two rows, NULL
//     content_data on both) is live on finetuning.uk today.

func dedupSection(slot, html, componentID string, content map[string]interface{}) SectionData {
	return SectionData{
		ComponentName: slot,
		ComponentID:   componentID,
		HTML:          html,
		ContentData:   content,
	}
}

func slotsOf(sections []SectionData) []string {
	out := make([]string, 0, len(sections))
	for _, s := range sections {
		out = append(out, s.ComponentName)
	}
	return out
}

// The recorded incident, exactly: 12 entries as 6 adjacent byte-identical pairs,
// each pair equal on slot, HTML, component_id and content_data.
//
// MUTATION: make collapseDuplicateSections return (sections, nil, nil)
// unconditionally — kept becomes 12 and this fails.
func TestCollapseCatchesTheVoncAdjacentPairSignature(t *testing.T) {
	slots := []string{"hero-about", "content-block-about", "game-master-explanation",
		"platform-comparison", "differentiators", "gauntlet-cta"}

	var sections []SectionData
	for _, slot := range slots {
		s := dedupSection(slot, "<section>"+slot+"</section>", "",
			map[string]interface{}{"heading": slot, "body": "copy for " + slot})
		sections = append(sections, s, s) // the adjacent pair
	}

	kept, groups, planProtected := collapseDuplicateSections(sections, nil)

	if len(kept) != 6 {
		t.Fatalf("kept = %d sections (%v), want 6", len(kept), slotsOf(kept))
	}
	for i, slot := range slots {
		if kept[i].ComponentName != slot {
			t.Errorf("kept[%d] = %q, want %q — first-occurrence order is not preserved",
				i, kept[i].ComponentName, slot)
		}
	}
	if len(groups) != 6 {
		t.Fatalf("groups = %d, want 6", len(groups))
	}
	if len(planProtected) != 0 {
		t.Errorf("planProtected = %v, want none (no plan was supplied)", planProtected)
	}
	for _, g := range groups {
		if g.KeptArrivalPosition%2 != 1 {
			t.Errorf("group %s kept arrival position %d, want the odd (first) member",
				g.Slot, g.KeptArrivalPosition)
		}
		if len(g.RemovedArrivalPositions) != 1 ||
			g.RemovedArrivalPositions[0] != g.KeptArrivalPosition+1 {
			t.Errorf("group %s removed %v, want exactly [%d]",
				g.Slot, g.RemovedArrivalPositions, g.KeptArrivalPosition+1)
		}
	}
	if sig := dedupAdjacencySignature(groups); sig != "adjacent" {
		t.Errorf("adjacency signature = %q, want %q — this is the clue that ruled out the concurrent-save race", sig, "adjacent")
	}
}

// The negative control, and the reason the fix is not a unique index. This is a
// live fleet shape: 11 pages repeat a slot with DIFFERENT content
// (generic-text-block 2–3× on ai-agent-orchestration, leopardess,
// gaswholesalers, finetuning, idea.uk; info-card-grid ×2 on webdesign.co.uk).
//
// MUTATION: reduce sectionPersistIdentity to s.ComponentName — this pair
// collapses and the test fails. That mutation IS the rejected unique-index rule,
// so this test is what distinguishes the shipped guard from the documented
// landmine.
func TestCollapseLeavesLegitimateRepeatedSlotsAlone(t *testing.T) {
	sections := []SectionData{
		dedupSection("generic-text-block", "<section>pricing</section>", "",
			map[string]interface{}{"heading": "Pricing", "body": "how we price"}),
		dedupSection("generic-text-block", "<section>terms</section>", "",
			map[string]interface{}{"heading": "Terms", "body": "supply terms"}),
	}

	kept, groups, _ := collapseDuplicateSections(sections, nil)

	if len(kept) != 2 {
		t.Fatalf("kept = %d, want 2 — a legitimate repeated slot was destroyed", len(kept))
	}
	if len(groups) != 0 {
		t.Fatalf("groups = %+v, want none", groups)
	}
}

// The finetuning.uk shape: two rows, NULL content_data on BOTH, different
// markup. bugs_open/156's own candidate key — (slot_name, md5(content_data)) —
// calls these identical and would delete one.
//
// MUTATION: drop s.HTML from sectionPersistIdentity's join — these collapse and
// the test fails. That mutation IS the bug file's literal recommendation.
func TestCollapseLeavesNullContentDataWithDifferentHTMLAlone(t *testing.T) {
	sections := []SectionData{
		dedupSection("generic-text-block", "<section>our position on AI</section>", "", nil),
		dedupSection("generic-text-block", "<section>what that means for you</section>", "", nil),
	}

	kept, groups, _ := collapseDuplicateSections(sections, nil)

	if len(kept) != 2 {
		t.Fatalf("kept = %d, want 2 — two NULL-content sections with different markup are not the same section", len(kept))
	}
	if len(groups) != 0 {
		t.Fatalf("groups = %+v, want none", groups)
	}
}

// nil and an empty map both bind SQL NULL in the insert loop, so two entries
// differing only in that respect would write indistinguishable rows.
//
// MUTATION: marshal ContentData unconditionally (drop the len==0 → NULL
// normalisation) — "null" and "{}" differ as strings, nothing collapses, and
// the test fails.
func TestCollapseTreatsNilAndEmptyContentDataAsTheSameStoredValue(t *testing.T) {
	sections := []SectionData{
		dedupSection("hero", "<section>hero</section>", "", nil),
		dedupSection("hero", "<section>hero</section>", "", map[string]interface{}{}),
	}

	kept, groups, _ := collapseDuplicateSections(sections, nil)

	if len(kept) != 1 {
		t.Fatalf("kept = %d, want 1 — both entries bind NULL content_data and identical HTML", len(kept))
	}
	if len(groups) != 1 {
		t.Fatalf("groups = %d, want 1", len(groups))
	}
}

// An unparseable component_id binds NULL exactly as an empty one does.
//
// MUTATION: compare s.ComponentID raw instead of parse-else-NULL — nothing
// collapses and the test fails.
func TestCollapseNormalisesComponentIDTheWayTheInsertDoes(t *testing.T) {
	sections := []SectionData{
		dedupSection("hero", "<section>hero</section>", "", map[string]interface{}{"a": 1}),
		dedupSection("hero", "<section>hero</section>", "not-a-uuid", map[string]interface{}{"a": 1}),
	}

	kept, _, _ := collapseDuplicateSections(sections, nil)

	if len(kept) != 1 {
		t.Fatalf("kept = %d, want 1 — both component_ids bind SQL NULL", len(kept))
	}
}

// A real component_id difference is a real difference.
//
// MUTATION: drop the component_id part from the key — these collapse and the
// test fails.
func TestCollapseKeepsEntriesThatDifferOnlyByComponentID(t *testing.T) {
	sections := []SectionData{
		dedupSection("hero", "<section>hero</section>", "11111111-1111-1111-1111-111111111111", nil),
		dedupSection("hero", "<section>hero</section>", "22222222-2222-2222-2222-222222222222", nil),
	}

	kept, groups, _ := collapseDuplicateSections(sections, nil)

	if len(kept) != 2 {
		t.Fatalf("kept = %d, want 2", len(kept))
	}
	if len(groups) != 0 {
		t.Fatalf("groups = %+v, want none", groups)
	}
}

// The other doubling shape: the whole loop ran twice (1,2,3,1,2,3) rather than
// each iteration emitting twice (1,1,2,2,3,3). Both must collapse, and the
// record must tell them apart — that distinction is what ruled out the
// concurrent-save race in the original investigation.
//
// MUTATIONS: (a) compare each entry only against its immediate predecessor —
// nothing collapses; (b) hardcode dedupAdjacencySignature to "adjacent" — the
// signature assertion fails.
func TestCollapseCatchesNonAdjacentDuplicatesAndSaysSo(t *testing.T) {
	a := dedupSection("hero", "<section>a</section>", "", map[string]interface{}{"n": 1})
	b := dedupSection("body", "<section>b</section>", "", map[string]interface{}{"n": 2})
	c := dedupSection("cta", "<section>c</section>", "", map[string]interface{}{"n": 3})

	kept, groups, _ := collapseDuplicateSections([]SectionData{a, b, c, a, b, c}, nil)

	if len(kept) != 3 {
		t.Fatalf("kept = %d (%v), want 3", len(kept), slotsOf(kept))
	}
	if got := slotsOf(kept); got[0] != "hero" || got[1] != "body" || got[2] != "cta" {
		t.Errorf("kept order = %v, want [hero body cta]", got)
	}
	if len(groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(groups))
	}
	if sig := dedupAdjacencySignature(groups); sig != "non_adjacent" {
		t.Errorf("adjacency signature = %q, want %q", sig, "non_adjacent")
	}
}

// Plan parity with remove_duplicate_page_sections: a repetition the effective
// plan source specifies is not collapsed, and the fact is recorded rather than
// dropped.
//
// MUTATION: ignore the planned map — the pair collapses and the test fails.
func TestCollapseHonoursPlanSpecifiedRepetition(t *testing.T) {
	s := dedupSection("info-card-grid", "<section>cards</section>", "",
		map[string]interface{}{"cards": []interface{}{"one", "two"}})

	kept, groups, planProtected := collapseDuplicateSections(
		[]SectionData{s, s}, map[string]int{"info-card-grid": 2})

	if len(kept) != 2 {
		t.Fatalf("kept = %d, want 2 — the plan specifies two instances of this slot", len(kept))
	}
	if len(groups) != 0 {
		t.Fatalf("groups = %+v, want none — a plan-protected slot must not be reported as collapsed", groups)
	}
	if len(planProtected) != 1 {
		t.Fatalf("planProtected = %+v, want one entry", planProtected)
	}
	if planProtected[0]["slot_name"] != "info-card-grid" {
		t.Errorf("planProtected slot = %v, want info-card-grid", planProtected[0]["slot_name"])
	}
}

// The plan constrains the count, not the identity: four identical entries where
// the plan specifies two must collapse to two, not to one and not to four.
//
// MUTATION: treat a planned slot as wholly exempt (skip it entirely) — kept
// becomes 4 and the test fails.
func TestCollapseStopsAtThePlannedCountRatherThanExemptingTheSlot(t *testing.T) {
	s := dedupSection("info-card-grid", "<section>cards</section>", "",
		map[string]interface{}{"cards": []interface{}{"one"}})

	kept, groups, planProtected := collapseDuplicateSections(
		[]SectionData{s, s, s, s}, map[string]int{"info-card-grid": 2})

	// 4 rows, 3 would be removed by identity, leaving 1 < planned 2 — so the
	// slot is protected and nothing is removed. The guard deliberately does NOT
	// partially collapse: it has no basis for choosing how many of an identical
	// set the plan meant, and leaving them is the non-destructive direction.
	if len(kept) != 4 {
		t.Fatalf("kept = %d, want 4 — a partial collapse would be a guess about what the plan meant", len(kept))
	}
	if len(groups) != 0 {
		t.Fatalf("groups = %+v, want none", groups)
	}
	if len(planProtected) != 1 {
		t.Fatalf("planProtected = %+v, want one entry naming the slot", planProtected)
	}
	if planProtected[0]["would_remove"] != 3 {
		t.Errorf("would_remove = %v, want 3 — the record must say what was NOT done", planProtected[0]["would_remove"])
	}
}

// The overwhelmingly common case must be untouched and must not read the plan
// store. The DB-seam consequence is tested by inspection of the lazy call, but
// the pure seam's "nothing to do" answer is what gates it.
//
// MUTATION: drop the slot-count prefilter's early return — still passes (the
// answer is the same), which is why this test asserts identity of the slice
// contents rather than the prefilter's existence.
func TestCollapseIsANoOpOnAPageWithNoRepeatedSlot(t *testing.T) {
	sections := []SectionData{
		dedupSection("hero", "<section>a</section>", "", map[string]interface{}{"n": 1}),
		dedupSection("body", "<section>b</section>", "", map[string]interface{}{"n": 2}),
		dedupSection("cta", "<section>c</section>", "", map[string]interface{}{"n": 3}),
	}

	kept, groups, planProtected := collapseDuplicateSections(sections, nil)

	if len(kept) != 3 || len(groups) != 0 || len(planProtected) != 0 {
		t.Fatalf("kept=%d groups=%d planProtected=%d, want 3/0/0",
			len(kept), len(groups), len(planProtected))
	}
}

// Key ordering inside content_data must not decide identity: encoding/json sorts
// map keys, which is the property the file header's subset proof rests on. If
// that ever stopped holding, this guard would silently stop collapsing and the
// proof against SectionIdentityKey would be void.
//
// MUTATION: build the content part with fmt.Sprintf("%v", s.ContentData) —
// Go randomises map iteration order, so this becomes flaky and fails.
func TestCollapseIdentityDoesNotDependOnGoMapOrder(t *testing.T) {
	a := dedupSection("hero", "<section>hero</section>", "",
		map[string]interface{}{"alpha": 1, "beta": 2, "gamma": 3})
	b := dedupSection("hero", "<section>hero</section>", "",
		map[string]interface{}{"gamma": 3, "beta": 2, "alpha": 1})

	for i := 0; i < 50; i++ {
		if sectionPersistIdentity(a, 0) != sectionPersistIdentity(b, 1) {
			t.Fatalf("identity differed on iteration %d — json.Marshal is not ordering map keys", i)
		}
	}

	// And the property the subset proof needs: equal identity implies equal
	// marshalled documents, which is what makes the jsonb text equal after
	// persist.
	ja, _ := json.Marshal(a.ContentData)
	jb, _ := json.Marshal(b.ContentData)
	if string(ja) != string(jb) {
		t.Fatalf("marshalled blobs differ: %s vs %s", ja, jb)
	}
}

// A content_data that cannot be marshalled makes the identity uncomputable, so
// the guard abstains for that entry rather than guessing it is a duplicate. The
// insert loop would bind NULL for both, so collapsing would be defensible — but
// abstaining is the direction that cannot destroy a section, and it costs only a
// duplicate row the post-hoc detector still sees.
//
// MUTATION: fall through to dedupNullSentinel on a marshal error — these
// collapse and the test fails.
func TestCollapseAbstainsWhenContentDataCannotBeMarshalled(t *testing.T) {
	bad := map[string]interface{}{"ch": make(chan int)}
	sections := []SectionData{
		dedupSection("hero", "<section>hero</section>", "", bad),
		dedupSection("hero", "<section>hero</section>", "", bad),
	}

	kept, groups, _ := collapseDuplicateSections(sections, nil)

	if len(kept) != 2 {
		t.Fatalf("kept = %d, want 2 — an uncomputable identity must not be treated as a match", len(kept))
	}
	if len(groups) != 0 {
		t.Fatalf("groups = %+v, want none", groups)
	}
}

// The identity key must not be forgeable by moving bytes across the join —
// "a"+"bc" and "ab"+"c" must never produce one identity. Same discipline, and
// the same reason, as datahelpers.SectionIdentityKey's NUL join.
//
// MUTATION: join the parts with "|" — a slot named "hero|<section>x" collides
// and the test fails.
func TestCollapseIdentityJoinIsUnambiguous(t *testing.T) {
	a := sectionPersistIdentity(dedupSection("hero", "x", "", nil), 0)
	b := sectionPersistIdentity(dedupSection("herox", "", "", nil), 1)
	if a == b {
		t.Fatalf("identity collision across the join boundary: %q", a)
	}
	if !strings.Contains(a, dedupIdentitySep) {
		t.Fatalf("identity %q does not use the NUL separator", a)
	}
}

// A mixed shape — one adjacent pair and one whole-loop repeat on the same page —
// must be reported as "mixed" rather than silently reduced to either. The
// signature is a forensic claim about the producer; a wrong one sends the next
// investigator after the wrong mechanism.
//
// MUTATION: return early from dedupAdjacencySignature on the first adjacent
// group — the answer becomes "adjacent" and the test fails.
func TestAdjacencySignatureReportsMixedWhenBothShapesArePresent(t *testing.T) {
	groups := []collapsedDupGroup{
		{Slot: "hero", KeptArrivalPosition: 1, RemovedArrivalPositions: []int{2}},
		{Slot: "cta", KeptArrivalPosition: 3, RemovedArrivalPositions: []int{9}},
	}
	if sig := dedupAdjacencySignature(groups); sig != "mixed" {
		t.Fatalf("signature = %q, want %q", sig, "mixed")
	}
	if sig := dedupAdjacencySignature(nil); sig != "" {
		t.Fatalf("signature of no groups = %q, want empty", sig)
	}
}
