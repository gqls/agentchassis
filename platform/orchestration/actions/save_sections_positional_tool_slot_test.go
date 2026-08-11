package actions

import (
	"testing"

	"github.com/google/uuid"
)

// Settles the question HANDOFF_2026-08-10c §6 left marked [INFERRED] for the
// decomposition lanes, and that Track A could not answer: what happens to a
// decomposed page's LOCKED calculator row when a rebuild composes that page.
//
// The worry as originally written was "a positional `tool-1` never matches, so
// the calculator is moved to the bottom of the page". That framing is wrong and
// the distinction matters, because the two readings call for opposite actions:
//
//   - matchLockedRow matches the locked row's slot against the INCOMING SECTION
//     NAME. A positional name matches perfectly well — it is a string like any
//     other. So a composition built from pages.sections (which for a decomposed
//     tool page IS ["prose-0","tool-1","prose-2"]) matches `tool-1` exactly, on
//     the first branch, and the row is consumed and left where it is.
//   - The trap fires only when the incoming composition OMITS the tool slot —
//     which is what a SEEDED SITE PLAN would do if it describes the page in
//     semantic section names. Then the row is unconsumed, and
//     save_page_sections_action.go:975 repositions it to len(sections)+1 and
//     emits a lock_blocked change item.
//
// So the precondition to protect is "the composition handed to the writer must
// name the tool slot", NOT "avoid positional names". Seeding a site plan is the
// dangerous act; rerendering from stored sections is not.
//
// Written as a test rather than proved by driving a rebuild at a live
// consumer-finance calculator: it answers the same question deterministically,
// cannot move a real widget to the bottom of a real page, and stays behind as a
// regression guard. The live half that remains genuinely unmeasured is stated in
// the lane handoff — this covers the matching rule, not the whole pipeline.
func TestMatchLockedRowPositionalToolSlot(t *testing.T) {
	// The shape loans-consolidation actually has on disk today.
	newRows := func() []*lockedPageRow {
		return []*lockedPageRow{{id: uuid.New(), slot: "tool-1"}}
	}

	// 1. A rerender composed from pages.sections carries the positional name.
	//    It MUST match, or every rebuild of a decomposed tool page would shunt
	//    the calculator to the bottom.
	rows := newRows()
	if lr := matchLockedRow(rows, "tool-1"); lr == nil || lr.slot != "tool-1" {
		t.Fatalf("positional tool slot must match its own name exactly, got %+v", lr)
	}

	// 2. A composition that omits the tool slot must NOT match it. This is the
	//    trap: the row falls through to the reposition-and-emit-lock_blocked
	//    path. Named semantic slots are what a seeded site_plan would produce.
	rows = newRows()
	for _, incoming := range []string{"hero", "calculator", "mortgage-tool", "prose-0", "tool", "tool-2"} {
		if lr := matchLockedRow(rows, incoming); lr != nil {
			t.Fatalf("incoming %q must not match locked slot tool-1, got %+v", incoming, lr)
		}
	}

	// 3. The whole composition, in order, with the tool slot present: the
	//    locked row is consumed exactly once and by the right section.
	rows = newRows()
	consumed := 0
	for _, incoming := range []string{"prose-0", "tool-1", "prose-2"} {
		if lr := matchLockedRow(rows, incoming); lr != nil {
			lr.consumed = true
			consumed++
			if incoming != "tool-1" {
				t.Fatalf("locked tool row consumed by the wrong section %q", incoming)
			}
		}
	}
	if consumed != 1 {
		t.Fatalf("expected the locked row to be consumed exactly once, got %d", consumed)
	}

	// 4. The same composition with the tool slot dropped leaves the row
	//    UNCONSUMED — which is precisely the state that triggers repositioning
	//    to len(sections)+1 and the lock_blocked work item.
	rows = newRows()
	for _, incoming := range []string{"prose-0", "prose-2"} {
		if lr := matchLockedRow(rows, incoming); lr != nil {
			lr.consumed = true
		}
	}
	if rows[0].consumed {
		t.Fatal("a composition omitting the tool slot must leave the locked row unconsumed")
	}
}

// The kebab fallback is a real widening and worth pinning: `tool_1` and `tool-1`
// must be treated as the same slot, or a producer that emits snake_case silently
// drops the calculator to the bottom of the page.
func TestMatchLockedRowPositionalKebabEquivalence(t *testing.T) {
	rows := []*lockedPageRow{{id: uuid.New(), slot: "tool_1"}}
	if lr := matchLockedRow(rows, "tool-1"); lr == nil {
		t.Fatal("snake_case locked slot must match a kebab-case incoming name")
	}
}
