// FILE: platform/orchestration/actions/plan_sections_section_identity_test.go
//
// IMG-075 round 2, answering the council's gating objection (corr
// 2979c27f-1545-47c5-b28d-f8a700bb1cb0, bug_historian HIGH + reuse_agent HIGH):
//
//   - the ordinal's identity was computed by three separately-written pieces of
//     arithmetic reading three different orderings, which is the very shape the
//     submission invoked to REJECT a position integer; and
//   - no test asserted that the build path and the re-render path land on the
//     same occurrence for the same physical section, nor covered the population
//     where the plan's order and the live order have drifted apart — the one
//     population where the binding mis-binds SILENTLY instead of falling back.
//
// Both are now structural: occurrence comes from InstanceCounter (the estate's
// existing rule, the one that also assigns element-id tokens), and a binding is
// only attempted when the two orderings agree. These tests pin both, and both
// are mutation-sensitive: re-hand-roll either counter on the raw name and
// TestSectionIdentity_MatchesTheEstatesInstanceTokenRule fails; delete the
// agreement guard and TestSectionOrderAgrees_* plus the drift case fail.
package actions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/gqls/agentchassis/platform/storage"
)

// occurrenceFromToken recovers the occurrence InstanceToken encoded, so the
// imagery identity can be compared against the token rule rather than against a
// restatement of itself. "c-x" is occurrence 0; "c-x-3" is occurrence 2.
func occurrenceFromToken(t *testing.T, token string) int {
	t.Helper()
	i := strings.LastIndex(token, "-")
	if i < 0 {
		t.Fatalf("token %q has no separator", token)
	}
	n, err := strconv.Atoi(token[i+1:])
	if err != nil {
		return 0 // no trailing number: first instance
	}
	return n - 1
}

// TestSectionIdentity_MatchesTheEstatesInstanceTokenRule ties per-section
// imagery identity to the occurrence rule that already governs element ids. The
// spellings differ deliberately: a plan naming a slot "Article-Body " and a
// stored row naming it "article-body" are ONE slot to the counter, and a
// hand-rolled map keyed on the raw string would call them two — which is the
// reuse objection made falsifiable rather than argued.
func TestSectionIdentity_MatchesTheEstatesInstanceTokenRule(t *testing.T) {
	page := []string{"hero", "Illustrated-Text-Block", "illustrated-text-block", " ILLUSTRATED-TEXT-BLOCK ", "call-to-action"}

	tokens := InstanceTokensForPage(page)

	counter := NewInstanceCounter()
	for i, name := range page {
		ref := newSectionRef(name, counter.NextOccurrence(name))
		wantOccurrence := occurrenceFromToken(t, tokens[i])
		if ref.Occurrence != wantOccurrence {
			t.Errorf("section %d (%q): imagery identity says occurrence %d, the element-id rule says %d — the two derivations have diverged",
				i, name, ref.Occurrence, wantOccurrence)
		}
		if ref.Name != strings.ToLower(strings.TrimSpace(name)) {
			t.Errorf("section %d: identity name %q is not normalised the way the counter keys it", i, ref.Name)
		}
	}

	// And the three spellings must be three DIFFERENT sections of one slot,
	// not one section counted three times.
	seen := map[sectionRef]bool{}
	c2 := NewInstanceCounter()
	for _, name := range page {
		ref := newSectionRef(name, c2.NextOccurrence(name))
		if seen[ref] {
			t.Fatalf("two sections collapsed onto one identity at %q — every section on a page must be distinguishable", name)
		}
		seen[ref] = true
	}
}

// TestSectionOrderAgrees_TruthTable pins the guard that decides whether an
// ordinal into the plan's list may be turned into an occurrence over the live
// list. Each row is a way the two orderings come apart in production.
func TestSectionOrderAgrees_TruthTable(t *testing.T) {
	cases := []struct {
		name  string
		plan  []string
		live  []string
		agree bool
		why   string
	}{
		{"identical", []string{"hero", "itb", "itb"}, []string{"hero", "itb", "itb"}, true,
			"the ordinary case: the page was built from this plan and nothing has moved"},
		{"site-level slot in the plan only", []string{"hero", "itb", "site-footer"}, []string{"hero", "itb"}, true,
			"the build loop filters header/footer out before iterating, so the plan carries one the live list never can"},
		{"spelling differs", []string{"Hero", " itb "}, []string{"hero", "itb"}, true,
			"the occurrence counter lower-cases and trims, so the guard must not call this a mismatch"},
		{"reordered", []string{"itb", "hero", "itb"}, []string{"itb", "itb", "hero"}, false,
			"a manual reorder: every ordinal past the swap names a different section than it did"},
		{"section deleted live", []string{"hero", "itb", "itb"}, []string{"hero", "itb"}, false,
			"an earlier section edit removed one; ordinals past it now point one section too far"},
		{"section inserted live", []string{"hero", "itb"}, []string{"hero", "itb", "itb"}, false,
			"a locked or hand-added section the plan does not know about"},
		{"live list unknown", []string{"hero", "itb"}, nil, false,
			"a caller that never said what it is iterating gets no binding, not a guess"},
	}
	for _, c := range cases {
		if got := sectionOrderAgrees(c.plan, c.live); got != c.agree {
			t.Errorf("%s: sectionOrderAgrees = %v, want %v — %s", c.name, got, c.agree, c.why)
		}
	}
}

// TestPlanSections_DriftedLiveOrderStandsDownToPageWide is the population
// bug_historian named: the plan and the built page disagree, so an ordinal
// cannot be trusted to name a live section. The figures must NOT be bound
// per-section — every section takes the page-wide first-wins value, which is
// what this page resolved before binding existed.
//
// Without the guard this test does not merely fail, it fails INTERESTINGLY: the
// second section receives the shark-grip figure that belongs to a section which
// is no longer where the plan thinks it is.
func TestPlanSections_DriftedLiveOrderStandsDownToPageWide(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	componentID := uuid.New().String()

	expectIllustratedComponents(mock, componentID)
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(slotRows())
	expectPostLoopReads(mock)
	// Two figures, planned for ordinals 0 and 2 of a THREE-section plan — while
	// the live page has two sections, both illustrated blocks. The plan's middle
	// section is gone from the live page.
	expectAssetLookups(mock, sectionImageryRows().
		AddRow("illustration", "grip-styles:0", "illustration_ring_grip", "illustration").
		AddRow("illustration", "grip-styles:2", "illustration_shark_grip", "illustration"),
		"illustrated-text-block", "call-to-action", "illustrated-text-block")
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "grip-styles",
			[]string{"illustrated-text-block", "illustrated-text-block"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 2 {
		t.Fatalf("expected two ready sections, got %d", len(items))
	}
	// Page-wide first-wins: ordered by (kind, ordering), the ring grip is first.
	pageWide := storage.DeployedWebPath("illustration_ring_grip", "illustration")
	for i, item := range items {
		got, _ := item.ResolvedData["image_url"].(string)
		if got != pageWide {
			t.Errorf("section %d: image_url = %q, want the page-wide %q — a drifted page must fall back, never bind",
				i, got, pageWide)
		}
	}
}

// TestSectionRefForOrdinal_RefusesWhatItCannotName covers the ordinal arms that
// keep a bad reference from becoming a confident binding. Each is a live shape:
// bugs_open/214 records orphaned and malformed refs on current plans.
func TestSectionRefForOrdinal_RefusesWhatItCannotName(t *testing.T) {
	order := []string{"hero", "illustrated-text-block", "illustrated-text-block"}

	if ref, ok := sectionRefForOrdinal(order, "grip-styles:2"); !ok || ref.Occurrence != 1 || ref.Name != "illustrated-text-block" {
		t.Errorf("the in-range case must name the second illustrated block, got %+v ok=%v", ref, ok)
	}
	for _, bad := range []string{"grip-styles:3", "grip-styles:99", "grip-styles:x", "grip-styles:", "grip-styles", "grip-styles:-1"} {
		if _, ok := sectionRefForOrdinal(order, bad); ok {
			t.Errorf("scope_ref %q names no section of this page and must not bind", bad)
		}
	}
	// A page-name containing a colon must still resolve by its LAST colon.
	if ref, ok := sectionRefForOrdinal(order, fmt.Sprintf("%s:%d", "odd:name", 0)); !ok || ref.Name != "hero" {
		t.Errorf("the ordinal is the suffix after the last colon; got %+v ok=%v", ref, ok)
	}
}
