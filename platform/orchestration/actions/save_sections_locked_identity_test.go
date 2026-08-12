package actions

import (
	"testing"

	"github.com/google/uuid"
)

// The identity arm of matchLockedRow (bugs_open/241 rebuild lane, 2026-08-12).
//
// Companion to TestMatchLockedRowPositionalToolSlot, which settled the
// slot-NAME rule and drew the right conclusion from it: "the precondition to
// protect is 'the composition handed to the writer must name the tool slot'
// ... Seeding a site plan is the dangerous act; rerendering from stored
// sections is not." These tests cover what happens when that precondition
// CANNOT be met — a fresh site plan describes pages in semantic component
// names, so it never emits the positional `tool-2` a decomposed page stores.
//
// Without the identity arm the outcome is worse than the relocation that
// framing implies: the plan's own section for the SAME component is composed
// and inserted in place, while the locked original is repositioned to the page
// foot. One calculator becomes two, and the lock "held" while losing its
// position. Identity matching pairs them, so the locked row keeps the slot and
// the fresh copy is discarded — the behaviour the lock was always meant to buy.

func TestMatchLockedRowMatchesOnComponentIdentityAcrossPositionalSlots(t *testing.T) {
	toolComponent := uuid.New().String()

	// The shape loancalculator.co.uk actually stores: positional slot name,
	// linked to the tool component that renders the calculator.
	rows := []*lockedPageRow{{id: uuid.New(), slot: "tool-2", componentID: toolComponent}}

	// A fresh site plan names the component, not the positional slot.
	lr := matchLockedRow(rows, "tool-compare-loan-offers", toolComponent)
	if lr == nil {
		t.Fatal("a planned section naming the locked row's own component must match it — without this the plan inserts a duplicate and exiles the locked original to the page foot")
	}
	if lr.slot != "tool-2" {
		t.Fatalf("matched the wrong row: %+v", lr)
	}
}

func TestMatchLockedRowIdentityBeatsNaming(t *testing.T) {
	wanted := uuid.New().String()
	rows := []*lockedPageRow{
		{id: uuid.New(), slot: "tool-compare-loan-offers", componentID: uuid.New().String()},
		{id: uuid.New(), slot: "tool-2", componentID: wanted},
	}

	// The name matches row 0, the identity matches row 1. Identity wins —
	// same precedence as matchDecisionProtectedRow.
	lr := matchLockedRow(rows, "tool-compare-loan-offers", wanted)
	if lr == nil || lr.slot != "tool-2" {
		t.Fatalf("identity must beat naming, got %+v", lr)
	}
}

func TestMatchLockedRowEmptyIdentityDoesNotPairIdlessRows(t *testing.T) {
	// The guard the sibling documents: sections often arrive before
	// enrichSectionsWithComponentIDs has resolved an id. An empty id must not
	// pair with the first id-less locked row, or every unresolved section
	// would consume a lock at random.
	rows := []*lockedPageRow{{id: uuid.New(), slot: "prose-0", componentID: ""}}
	if lr := matchLockedRow(rows, "hero", ""); lr != nil {
		t.Fatalf("an empty component id must not match an id-less locked row by identity, got %+v", lr)
	}
	// ...but the slot-name arms still work, unchanged.
	if lr := matchLockedRow(rows, "prose-0", ""); lr == nil {
		t.Fatal("slot-name matching must still work when no identity is available")
	}
}

func TestMatchLockedRowIdentityConsumesOnlyOnce(t *testing.T) {
	shared := uuid.New().String()
	rows := []*lockedPageRow{{id: uuid.New(), slot: "tool-2", componentID: shared}}

	first := matchLockedRow(rows, "tool-x", shared)
	if first == nil {
		t.Fatal("first section must match")
	}
	first.consumed = true

	// A second section naming the same component must NOT swallow the same
	// lock — one locked row blocks at most one incoming section.
	if again := matchLockedRow(rows, "tool-x", shared); again != nil {
		t.Fatalf("a consumed locked row must not match again, got %+v", again)
	}
}
