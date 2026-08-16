// FILE: platform/orchestration/datahelpers/locked_page_sections_test.go
//
// bugs_open/285 — the section-list assembler was lock-blind. These pin the two
// shared pieces the fix rests on: (1) the locked-row query uses THE GUARD'S
// predicate, not a lookalike; (2) the merge pairs list entries with locked
// rows the way save_page_sections' matchLockedRow does (slot exact → slot
// kebab → component function/name, consume-once) and inserts the unpaired
// rows at their live position. A wrong pairing is not a cosmetic bug: it is
// how bugs_open/189 turned a locked calculator into two rows.

package datahelpers

import (
	"reflect"
	"strings"
	"testing"
)

func TestLockedPageSlotsSQLUsesTheGuardsOwnPredicate(t *testing.T) {
	// The lock arm must be AgentWritableSQLFor("pc.") verbatim, negated —
	// LOCK-007's rule (read predicate DERIVED from the write predicate) and
	// LANDMINES "the PIN predicate is not the POOL predicate".
	want := "NOT " + AgentWritableSQLFor("pc.")
	if !strings.Contains(LockedPageSlotsSQL, want) {
		t.Fatalf("LockedPageSlotsSQL does not embed the guard predicate %q:\n%s", want, LockedPageSlotsSQL)
	}
	for _, frag := range []string{"pc.locked_at IS NULL", "pc.lock_type = 'timed'", "pc.lock_expires_at < NOW()"} {
		if !strings.Contains(LockedPageSlotsSQL, frag) {
			t.Errorf("predicate fragment %q missing", frag)
		}
	}
	// Membership condition is stated separately from the lock predicate.
	if !strings.Contains(LockedPageSlotsSQL, "<> 'removed'") {
		t.Errorf("membership filter (build_status <> 'removed') missing")
	}
	// A bare `locked_at IS NOT NULL` re-typing would ignore expiry — refuse it.
	if strings.Contains(LockedPageSlotsSQL, "locked_at IS NOT NULL") {
		t.Errorf("SQL re-types the lock test instead of negating the shared predicate")
	}
}

func TestMergeLockedPageSlots(t *testing.T) {
	row := func(slot, fn string, pos int) LockedPageSlot {
		return LockedPageSlot{Slot: slot, ComponentFunction: fn, ComponentName: fn, Position: pos, ComponentID: "cid-" + fn}
	}
	cases := []struct {
		name         string
		list         []string
		locked       []LockedPageSlot
		want         []string
		wantInserted []string
		wantAt       []int
	}{
		{
			// The webdesign.uk/contact case: plan [hero, contact-info], locked
			// chat-input-box at live position 3 → appended.
			name:         "append at tail (contact chat box)",
			list:         []string{"hero", "contact-info"},
			locked:       []LockedPageSlot{row("chat-input-box", "chat-input-box", 3)},
			want:         []string{"hero", "contact-info", "chat-input-box"},
			wantInserted: []string{"chat-input-box"}, wantAt: []int{2},
		},
		{
			// loancalculator tool-settlement-calculator: locked tool-2 at live
			// position 3 inside a four-entry plan → inserted at index 2.
			name:         "insert at live position",
			list:         []string{"hero", "ported-prose", "faq", "tool-cta"},
			locked:       []LockedPageSlot{row("tool-2", "tool-early-settlement", 3)},
			want:         []string{"hero", "ported-prose", "tool-2", "faq", "tool-cta"},
			wantInserted: []string{"tool-2"}, wantAt: []int{2},
		},
		{
			// An exiled row (position beyond the list) clamps to the tail —
			// membership restored, history not rewritten.
			name:         "position beyond list clamps to tail",
			list:         []string{"hero", "faq"},
			locked:       []LockedPageSlot{row("tool-3", "tool-x", 9)},
			want:         []string{"hero", "faq", "tool-3"},
			wantInserted: []string{"tool-3"}, wantAt: []int{2},
		},
		{
			name:   "already present by exact slot → no duplicate",
			list:   []string{"hero", "chat-input-box"},
			locked: []LockedPageSlot{row("chat-input-box", "chat-input-box", 2)},
			want:   []string{"hero", "chat-input-box"},
		},
		{
			// The 041 naming landmine: snake_case plan, kebab-case slot. Without
			// the kebab arm the merge would add a second entry, plan_sections
			// would resolve both to one component, and save would INSERT the
			// second as a fresh row beside the locked one — the 189 shape.
			name:   "already present by kebab-normalised slot → no duplicate",
			list:   []string{"social_proof"},
			locked: []LockedPageSlot{row("social-proof", "social-proof", 1)},
			want:   []string{"social_proof"},
		},
		{
			// Plan names the FUNCTION, the row carries a positional slot: the
			// guard would pair them by identity at save time (component_id),
			// so the merge must treat the function name as "present".
			name:   "already present by component function → no duplicate",
			list:   []string{"hero", "tool-loan-repayment"},
			locked: []LockedPageSlot{row("tool-3", "tool-loan-repayment", 2)},
			want:   []string{"hero", "tool-loan-repayment"},
		},
		{
			// Two locked rows rendering the same component; plan names it once.
			// Consume-once: the first row pairs, the second is inserted.
			name: "consume-once with duplicate component",
			list: []string{"evidence-chart", "mechanism-flow"},
			locked: []LockedPageSlot{
				row("evidence-chart", "evidence-chart", 1),
				row("evidence-chart-ofwat", "evidence-chart", 3),
			},
			want:         []string{"evidence-chart", "mechanism-flow", "evidence-chart-ofwat"},
			wantInserted: []string{"evidence-chart-ofwat"}, wantAt: []int{2},
		},
		{
			// Two locked rows, plan names the component twice → both pair, nothing inserted.
			name: "two rows two entries → both pair",
			list: []string{"generic-text-block", "generic-text-block"},
			locked: []LockedPageSlot{
				row("generic-text-block", "generic-text-block", 1),
				row("generic-text-block", "generic-text-block", 2),
			},
			want: []string{"generic-text-block", "generic-text-block"},
		},
		{
			// Ascending insertion: the later row indexes against the grown list.
			name: "two insertions keep live order",
			list: []string{"a", "b", "c"},
			locked: []LockedPageSlot{
				row("l5", "l5", 5),
				row("l2", "l2", 2),
			},
			want:         []string{"a", "l2", "b", "c", "l5"},
			wantInserted: []string{"l2", "l5"}, wantAt: []int{1, 4},
		},
		{
			// A row without slot_name contributes its function name.
			name:         "slotless row falls back to function",
			list:         []string{"hero"},
			locked:       []LockedPageSlot{{ComponentFunction: "chat-input-box", Position: 2}},
			want:         []string{"hero", "chat-input-box"},
			wantInserted: []string{"chat-input-box"}, wantAt: []int{1},
		},
		{
			name:   "no locked rows → list unchanged",
			list:   []string{"hero"},
			locked: nil,
			want:   []string{"hero"},
		},
		{
			// The merge itself is pure and will happily produce a locked-only
			// list; the LOADER refuses to call it when no tier served (see the
			// loader test). Pinned here so that rule is a loader decision, not
			// a helper accident.
			name:         "empty list still merges (caller decides whether to call)",
			list:         nil,
			locked:       []LockedPageSlot{row("x", "x", 1)},
			want:         []string{"x"},
			wantInserted: []string{"x"}, wantAt: []int{0},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := append([]string(nil), tc.list...)
			got, inserted, at := MergeLockedPageSlots(tc.list, tc.locked)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("merged = %v, want %v", got, tc.want)
			}
			var names []string
			for _, lr := range inserted {
				names = append(names, lr.MergedName())
			}
			if !reflect.DeepEqual(names, tc.wantInserted) {
				t.Errorf("inserted = %v, want %v", names, tc.wantInserted)
			}
			if !reflect.DeepEqual(at, tc.wantAt) {
				t.Errorf("insertedAt = %v, want %v", at, tc.wantAt)
			}
			// The input list must not be mutated (callers keep parallel slices).
			if !reflect.DeepEqual(append([]string(nil), tc.list...), before) {
				t.Errorf("input list mutated: %v → %v", before, tc.list)
			}
		})
	}
}

func TestNormalizeComponentFunctionMovedDown(t *testing.T) {
	// The actions package delegates here; these are its documented examples.
	cases := map[string]string{
		"social_proof": "social-proof", "call_to_action": "call-to-action",
		"SocialProof": "social-proof", "social-proof": "social-proof", "": "",
	}
	for in, want := range cases {
		if got := NormalizeComponentFunction(in); got != want {
			t.Errorf("NormalizeComponentFunction(%q) = %q, want %q", in, got, want)
		}
	}
	if !IsKebabCase("chat-input-box") || IsKebabCase("Chat_Input") {
		t.Errorf("IsKebabCase misclassifies")
	}
}
