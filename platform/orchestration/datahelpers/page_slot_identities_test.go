// FILE: platform/orchestration/datahelpers/page_slot_identities_test.go
//
// bugs_open/204 — these pin the three things the three call sites of this
// judgement now share, and the one place they deliberately DIVERGE:
//   1. the per-page projection is byte-identical to what plan_sections shipped
//      (commit 13252f714), so the move to this package is a pure move;
//   2. both projections read the SAME rows, because a predicate that drifts
//      between "which component is at this slot" and "does this page have this
//      slot" is the drift class this file was created to end;
//   3. SlotIDMap drops a slot whose repeats disagree — an ambiguous carry source
//      is no carry source;
//   4. SlotNameSet does NOT, and that difference is load-bearing rather than an
//      oversight: membership decides whether to KEEP a name the page visibly has,
//      so inheriting the id map's conflict rule would delete a real section.

package datahelpers

import (
	"reflect"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// The projection plan_sections shipped, before this loader moved here. Written
// out rather than referenced so the test fails if someone "tidies" the SQL: the
// point is that the moved query still asks for exactly this.
const shippedPerPageProjection = `SELECT COALESCE(pc.slot_name, ''), COALESCE(pc.component_id::text, '')`

func TestPageSlotIdentitiesSQL_PerPageProjectionIsTheOnePlanSectionsShipped(t *testing.T) {
	if !strings.Contains(PageSlotIdentitiesSQL, shippedPerPageProjection) {
		t.Fatalf("per-page projection changed — the move is no longer a pure move.\nwant to contain:\n%s\ngot:\n%s",
			shippedPerPageProjection, PageSlotIdentitiesSQL)
	}
	// Order is part of the contract: callers that carry stored content rely on
	// position order, and 13252f714's own tests mock rows in that order.
	if !strings.Contains(PageSlotIdentitiesSQL, "ORDER BY pc.position ASC") {
		t.Errorf("per-page read lost its position ordering")
	}
	// A page name is REQUIRED on this projection — it does not select p.name, so
	// rows could not be attributed if several pages came back.
	if strings.Contains(PageSlotIdentitiesSQL, "SELECT p.name") {
		t.Errorf("per-page projection now selects p.name; either use the site read or update the scan")
	}
}

func TestPageSlotIdentitiesSQL_BothProjectionsShareOnePredicate(t *testing.T) {
	// The disconfirming result this test exists for: someone edits one query's
	// WHERE clause (say, adds `AND p.status='active'`) and leaves the other, so
	// "which component is at this slot" and "does this page carry this slot"
	// silently start answering about different row sets.
	for name, sqlText := range map[string]string{
		"per-page": PageSlotIdentitiesSQL,
		"per-site": PageSlotIdentitiesForSiteSQL,
	} {
		if !strings.Contains(sqlText, pageSlotPredicate) {
			t.Errorf("%s read does not embed the shared predicate:\n%s", name, sqlText)
		}
	}
	// And the shared predicate must still be the narrow one this file argues for
	// in its header — no membership filter smuggled in, because adding one would
	// change plan_sections' behaviour under cover of a refactor.
	if strings.Contains(pageSlotPredicate, "build_status") {
		t.Errorf("a build_status filter appeared in the shared predicate; that is a behaviour change for plan_sections and needs its own measurement")
	}
}

func TestSlotIDMap_DropsDisagreeingRepeatsKeepsAgreeingOnes(t *testing.T) {
	rows := []PageSlotRow{
		{PageName: "index", Slot: "generic-text-block", ComponentID: "aaa"},
		{PageName: "index", Slot: "generic-text-block", ComponentID: "aaa"}, // legitimate repeat
		{PageName: "index", Slot: "conflicted", ComponentID: "bbb"},
		{PageName: "index", Slot: "conflicted", ComponentID: "ccc"}, // disagreeing
		{PageName: "index", Slot: "", ComponentID: "ddd"},           // no slot name
		{PageName: "index", Slot: "no-component", ComponentID: ""},  // no component id
		{PageName: "index", Slot: "prose-0", ComponentID: "eee"},
	}
	got := SlotIDMap(rows, "plan_sections", "index", zap.NewNop())
	want := map[string]string{"generic-text-block": "aaa", "prose-0": "eee"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SlotIDMap = %v, want %v", got, want)
	}
	// Named explicitly: the conflicted slot must be ABSENT, not present with an
	// arbitrary winner. A map that picked "bbb" would still satisfy a laxer
	// assertion and would silently pin the wrong component.
	if _, present := got["conflicted"]; present {
		t.Errorf("a slot whose repeats disagree must be absent so resolution falls back to the name path")
	}
}

func TestSlotIDMap_WarningNamesItsCaller(t *testing.T) {
	// 204's closure evidence pod-greps this substring to prove a binary carries
	// the fix. If it changes, somebody's verification silently retires.
	core, recorded := observer.New(zapcore.WarnLevel)
	SlotIDMap([]PageSlotRow{
		{Slot: "x", ComponentID: "1"},
		{Slot: "x", ComponentID: "2"},
		{Slot: "x", ComponentID: "3"}, // a third disagreement must not warn twice
	}, "plan_sections", "index", zap.New(core))

	entries := recorded.All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly one warning per conflicted slot, got %d", len(entries))
	}
	msg := entries[0].Message
	if !strings.Contains(msg, "slot_name repeats with different component_ids") {
		t.Errorf("the pod-greppable substring is gone from %q", msg)
	}
	if !strings.HasPrefix(msg, "plan_sections: ") {
		t.Errorf("warning does not name its caller: %q", msg)
	}
}

func TestSlotNameSet_KeepsRepeatsAndGroupsByPage(t *testing.T) {
	rows := []PageSlotRow{
		{PageName: "guide-a", Slot: "prose-0", ComponentID: "aaa"},
		{PageName: "guide-a", Slot: "prose-1", ComponentID: "aaa"},
		{PageName: "guide-b", Slot: "tool-1", ComponentID: "bbb"},
		{PageName: "guide-b", Slot: "", ComponentID: "ccc"}, // no slot name: contributes nothing
	}
	got := SlotNameSet(rows)
	want := map[string]map[string]bool{
		"guide-a": {"prose-0": true, "prose-1": true},
		"guide-b": {"tool-1": true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SlotNameSet = %v, want %v", got, want)
	}
}

// The divergence test, and the reason SlotNameSet is not implemented on top of
// SlotIDMap. A page really can carry the same slot name twice with different
// components; for the ID question that is an ambiguity to refuse, but for the
// MEMBERSHIP question the page plainly HAS the slot, and answering "no" would
// delete a live section from the plan — the exact damage 204 records.
func TestSlotNameSet_DoesNotInheritTheIDMapsConflictRule(t *testing.T) {
	rows := []PageSlotRow{
		{PageName: "index", Slot: "prose-0", ComponentID: "aaa"},
		{PageName: "index", Slot: "prose-0", ComponentID: "bbb"}, // disagreeing repeat
	}
	if ids := SlotIDMap(rows, "test", "index", zap.NewNop()); len(ids) != 0 {
		t.Fatalf("precondition: SlotIDMap should refuse the ambiguous slot, got %v", ids)
	}
	set := SlotNameSet(rows)
	if !set["index"]["prose-0"] {
		t.Fatalf("membership must still be true for a slot the page demonstrably carries; " +
			"inheriting the id map's conflict rule here would delete a real section")
	}
}

func TestSlotNameSet_MissingComponentIDStillCounts(t *testing.T) {
	// A row whose component_id is NULL means the LINK is broken, not that the
	// page lacks the slot. Dropping it here would let validate delete a section
	// that a separate repair (check_unlinked_components) exists to relink.
	set := SlotNameSet([]PageSlotRow{{PageName: "index", Slot: "prose-0", ComponentID: ""}})
	if !set["index"]["prose-0"] {
		t.Errorf("a slot with no component_id must still count as present")
	}
}

func TestLoadPageSlotRows_RefusesAnEmptyPageName(t *testing.T) {
	// The per-page projection cannot attribute rows to pages, so an empty page
	// name would silently read the WHOLE SITE and label every row with "".
	// Refusing is how that becomes impossible rather than merely unlikely.
	_, err := LoadPageSlotRows(nil, nil, [16]byte{}, "")
	if err == nil {
		t.Fatal("expected a refusal for an empty page name")
	}
	if !strings.Contains(err.Error(), "page name required") {
		t.Errorf("error should name the cause, got %q", err)
	}
}
