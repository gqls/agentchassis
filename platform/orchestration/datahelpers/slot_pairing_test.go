// FILE: platform/orchestration/datahelpers/slot_pairing_test.go
//
// The shared pairing relation (slot_pairing.go) is the single point of truth
// for three askers, so two properties are pinned HERE, once:
//
//   - ARM PRIORITY. Identity beats slot beats kebab beats function — and the
//     arm loop sits OUTSIDE the candidate loop, so a weaker match on an early
//     candidate never beats a stronger match on a later one. The three
//     call-site suites cannot pin this: each drives fixtures where only one
//     arm can fire.
//   - WIRING. All three askers actually call the shared functions. The
//     behaviour suites are equivalence proofs and therefore stay green if
//     someone re-inlines a private copy — which is exactly the drift that
//     produced bugs_open/385 and drew council ece638fb's reuse gate.
package datahelpers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestSlotPairing_ArmPriorityAcrossCandidates pins the arm-outside-candidate
// structure. Candidate 0 matches only by kebab; candidate 1 matches by
// identity. A per-candidate arm walk would return 0 (first candidate, any
// arm); the canonical relation must return 1 (strongest arm, any candidate).
//
// MUTATION THAT MUST BREAK IT: swap the loops (candidates outer, arms inner),
// or reorder the arms in slotPairArmMatches.
func TestSlotPairing_ArmPriorityAcrossCandidates(t *testing.T) {
	stored := []SlotIdentity{
		{Slot: "loan_calc"},                      // kebab-matches "loan-calc"
		{Slot: "tool-2", ComponentID: "id-1234"}, // identity-matches
	}
	if got := PairIncomingToStored("loan-calc", "id-1234", stored, nil); got != 1 {
		t.Errorf("PairIncomingToStored = %d, want 1 — identity (arm 1) on a later candidate must beat "+
			"kebab (arm 3) on an earlier one; a per-candidate arm walk gets this wrong", got)
	}

	incoming := []IncomingSection{
		{Name: "loan-calc"},                          // kebab vs stored slot below
		{Name: "tool-loan", ComponentID: "id-1234"},  // identity
	}
	if got := PairStoredToIncoming(SlotIdentity{Slot: "loan_calc", ComponentID: "id-1234"}, incoming, nil); got != 1 {
		t.Errorf("PairStoredToIncoming = %d, want 1 — same priority rule in the inverted direction", got)
	}
}

// TestSlotPairing_EmptySidesNeverPair pins the non-empty guards on every arm
// in one place: empty ids, names and slots must never pair, or every
// unresolved section claims the first idless row.
func TestSlotPairing_EmptySidesNeverPair(t *testing.T) {
	stored := []SlotIdentity{{}}
	if got := PairIncomingToStored("", "", stored, nil); got != -1 {
		t.Errorf("PairIncomingToStored on all-empty = %d, want -1", got)
	}
	if got := PairStoredToIncoming(SlotIdentity{}, []IncomingSection{{}}, nil); got != -1 {
		t.Errorf("PairStoredToIncoming on all-empty = %d, want -1", got)
	}
}

// TestSlotPairing_ConsumedCandidatesAreSkipped pins that the consumption
// predicate is honoured on every arm, not only the first.
func TestSlotPairing_ConsumedCandidatesAreSkipped(t *testing.T) {
	stored := []SlotIdentity{{Slot: "tool-2"}, {Slot: "tool-2"}}
	consumed := map[int]bool{0: true}
	if got := PairIncomingToStored("tool-2", "", stored, func(i int) bool { return consumed[i] }); got != 1 {
		t.Errorf("PairIncomingToStored = %d, want 1 — a consumed candidate must be skipped", got)
	}
}

// TestSlotPairing_MergeCallsTheSharedRelation is the datahelpers-side wiring
// pin: MergeLockedPageSlots must decide through PairIncomingToStored. Its
// behaviour table (TestMergeLockedPageSlots) is an equivalence proof and stays
// green against a re-inlined private copy; this test is what goes red.
//
// MUTATION THAT MUST BREAK IT: re-inline the pair() arms inside
// MergeLockedPageSlots, or stop calling PairIncomingToStored there.
func TestSlotPairing_MergeCallsTheSharedRelation(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "locked_page_sections.go", nil, 0)
	if err != nil {
		t.Fatalf("parse locked_page_sections.go: %v", err)
	}
	var merge *ast.FuncDecl
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "MergeLockedPageSlots" {
			merge = fd
		}
	}
	if merge == nil {
		t.Fatal("CONTROL FAILED: MergeLockedPageSlots not found — the scan cannot see its target")
	}
	found := false
	ast.Inspect(merge, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "PairIncomingToStored" {
			found = true
		}
		return true
	})
	if !found {
		t.Error("MergeLockedPageSlots no longer calls PairIncomingToStored.\n" +
			"The pairing relation is shared across three askers (slot_pairing.go); a private copy " +
			"here is the drift that produced bugs_open/385. If the decision moved, point this test " +
			"at its new home — do not delete it.")
	}
}
