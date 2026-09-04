// FILE: platform/orchestration/actions/save_sections_reappend_identity_test.go
//
// bugs_open/479 — A RE-APPENDED TOOL MUST KEEP ITS OWN COMPONENT.
//
// Layer 2 has two carry arms and they were sharing one identity rule, which is
// why this bug existed. The arms are not the same case:
//
//	SPLICE      the plan named this slot and produced prose for it; Layer 2 puts
//	            the stored tool's bytes back. An incoming section is present and
//	            carries the PLAN's component. Overwriting that with the stored
//	            component would hold a legitimately-typed component at its old
//	            identity when the plan meant to swap it — three council seats
//	            objected to exactly that in RFC_046's first round, so the carry
//	            there is opt-in (`adopt_unidentified_fragments`) and narrowed to
//	            adopted fragments. THAT NARROWING IS CORRECT AND STAYS.
//
//	RE-APPEND   the plan named this slot NOTHING. The section is built wholly out
//	            of the stored row — bytes, content_data, slot name, stamp — and
//	            the component id was the one field of that row the copy dropped.
//	            There is no plan intent to override, because the plan never
//	            mentioned the slot. Dropping it is pure information loss.
//
// The loss is measured, not argued: `[MEASURED 2026-09-04]` 17 page_components
// rows fleet-wide hold rendered_html with no component reference, 5 of them tool
// slots serving working tools, and 3 of those were created AFTER bugs_open/385's
// matcher fix went live on 2026-08-26 — a live producer, not a backlog. Every one
// sits at its page's last position and was the last row of its write burst, which
// is this arm's signature.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// TestReappendedComponentID is the direct pin on the decision.
//
// MUTATION THAT MUST BREAK IT: return storedComponentID unconditionally (drops
// the active guard), or return "" unconditionally (restores the bug).
func TestReappendedComponentID(t *testing.T) {
	t.Run("an active stored component travels with its own bytes", func(t *testing.T) {
		if got := reappendedComponentID("comp-1", true); got != "comp-1" {
			t.Errorf("reappendedComponentID(active) = %q, want %q. The re-appended row IS the stored "+
				"row — same bytes, same slot — so its component is a fact already in the database, "+
				"not an inference about what the plan wanted.", got, "comp-1")
		}
	})

	t.Run("an inactive component is refused — a dangling id is worse than none", func(t *testing.T) {
		if got := reappendedComponentID("comp-1", false); got != "" {
			t.Errorf("reappendedComponentID(inactive) = %q, want empty. loadComponentSchemasByID drops "+
				"a row it cannot load and resolveComponent then returns invalidTemplate rather than "+
				"falling through to the slot-name map, which fails the WHOLE page. NULL at least "+
				"falls through.", got)
		}
	})

	t.Run("no stored component stays unknown", func(t *testing.T) {
		if got := reappendedComponentID("", true); got != "" {
			t.Errorf("reappendedComponentID(none) = %q, want empty", got)
		}
	})
}

// TestLayer2_ReappendedToolKeepsItsOwnComponent is the assertion at the seam.
//
// The stored tool's slot is absent from the incoming composition, so Layer 2 takes
// the re-append arm. The row it writes must bind the stored row's own component —
// pinned as a literal here, NOT AnyArg, because AnyArg is satisfied by the NULL
// this bug wrote for months.
//
// MUTATION THAT MUST BREAK IT: put carriedIdentity back in the re-append arm.
func TestLayer2_ReappendedToolKeepsItsOwnComponent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID, storedID := uuid.New(), uuid.New(), uuid.New()

	expectSaveSlotReadsPreloading(mock, siteID, pageID, "tool-repayment", lockedRowSet(), 1, 0, 1,
		layer2PreloadWithFunction("tool-calculator", layer2ToolHTML, "", storedID.String(), "tool-calculator"))

	// The incoming prose section goes in first and is not what this test is about.
	mock.ExpectExec("INSERT INTO page_components").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// The re-appended tool at position 2. Arg 5 is component_id ($5 of the INSERT).
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, 2, layer2ToolHTML, "tool-calculator",
			storedID.String(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-repayment", []interface{}{
			map[string]interface{}{
				"rendered_html":    layer2HeroHTML,
				"stored_slot_name": "prose-0",
				"component_name":   "prose-0",
			},
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := saveResult(t, out)["sections_saved"]; got != 2 {
		t.Errorf("sections_saved = %v, want 2 (the incoming section plus the re-appended tool)", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestLayer2_ReappendedToolWithADeadComponentStaysUnbound proves the guard is
// reachable through the real seam, not just in the unit test above. Without this
// the active flag could be dropped from the preload query and only the unit test
// would notice — and the unit test cannot see the query.
//
// MUTATION THAT MUST BREAK IT: drop `COALESCE(cc.is_active, false)` from the
// Layer 2 preload (and its scan), so every stored component reads as active.
func TestLayer2_ReappendedToolWithADeadComponentStaysUnbound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID, deadID := uuid.New(), uuid.New(), uuid.New()

	expectSaveSlotReadsPreloading(mock, siteID, pageID, "tool-repayment", lockedRowSet(), 1, 0, 1,
		layer2PreloadWithInactiveComponent("tool-calculator", layer2ToolHTML, deadID.String(), "tool-calculator"))

	mock.ExpectExec("INSERT INTO page_components").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, 2, layer2ToolHTML, "tool-calculator",
			nil, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-repayment", []interface{}{
			map[string]interface{}{
				"rendered_html":    layer2HeroHTML,
				"stored_slot_name": "prose-0",
				"component_name":   "prose-0",
			},
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := saveResult(t, out)["sections_saved"]; got != 2 {
		t.Errorf("sections_saved = %v, want 2", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestLayer2_SplicedToolStillTakesThePlansComponent is the BOUNDARY test, and it
// is the reason this change is narrow enough to be safe.
//
// The stored tool's component is active and would be carried by the re-append
// arm. On the SPLICE arm it must NOT be: an incoming section is present, it
// carries the plan's component, and the council's narrowing says the plan wins
// unless `adopt_unidentified_fragments` is armed AND the stored component is an
// adopted fragment. Neither holds here.
//
// MUTATION THAT MUST BREAK IT: use reappendedComponentID in the splice arm too —
// the "obvious" wider fix, which is the one three council seats rejected.
func TestLayer2_SplicedToolStillTakesThePlansComponent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	planID, storedID := uuid.New(), uuid.New()

	expectSaveSlotReadsPreloading(mock, siteID, pageID, "tool-repayment", lockedRowSet(), 1, 0, 1,
		layer2PreloadWithFunction("hero", layer2ToolHTML, "", storedID.String(), "hero"))

	// One row: the stored tool's bytes spliced into the plan's hero section,
	// still bound to the PLAN's component.
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, 1, layer2ToolHTML, "hero",
			planID.String(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-repayment", []interface{}{
			map[string]interface{}{
				"rendered_html":    layer2HeroHTML,
				"stored_slot_name": "hero",
				"component_name":   "hero",
				"component_id":     planID.String(),
			},
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := saveResult(t, out)["sections_saved"]; got != 1 {
		t.Errorf("sections_saved = %v, want 1", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
