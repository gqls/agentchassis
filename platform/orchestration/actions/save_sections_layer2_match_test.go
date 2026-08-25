// FILE: platform/orchestration/actions/save_sections_layer2_match_test.go
//
// bugs_open/385 §5c — the Layer 2 carry-forward's matcher.
//
// "Is this stored interactive row already represented in the incoming set?" is
// answered three times in the save path. matchLockedRow was given an identity
// arm (182/189/204); MergeLockedPageSlots was born mirroring those arms
// (LOCK-008); the Layer 2 matcher was the survivor still comparing by exact
// slot-name string. On the build arm the incoming names are the PLAN's function
// names, so a positionally-named stored tool (`tool-2`) read as "dropped
// entirely" while the same component sat in the set as `tool-loan-vs-savings` —
// and the re-append arm duplicated a locked calculator (2026-08-23, one page of
// a ten-page wave; the armed discriminator was the preload's
// build_status='deployed', which only that page's locked row carried).
//
// The unit tests call matchPreservedSectionIdx DIRECTLY — the extraction exists
// so a test executes the production decision rather than mirroring its
// construction (WRONG_CALLS 2026-08-19). The action-level test then pins the
// wiring on the motivating shape, because a helper nothing calls is not a fix.
package actions

import (
	"context"
	"go/ast"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Unit: the matcher's arms
// ---------------------------------------------------------------------------

func TestMatchPreservedSectionIdx_IdentityArmPairsPlanNamedSectionWithPositionalSlot(t *testing.T) {
	// THE MOTIVATING CASE, bugs_open/385 §5c. Stored: the locked calculator,
	// slot "tool-2", component 448422ce…, function "tool-loan-vs-savings".
	// Incoming: the plan's composition, which names the tool by function and —
	// enrichSectionsWithComponentIDs having run — carries the resolved id.
	//
	// MUTATION THAT MUST BREAK IT: remove the identity arm (or revert the
	// matcher to `sections[i].ComponentName == p.slot` alone). The pre-385 code
	// returns -1 here, and -1 is exactly the "dropped entirely" verdict that
	// appended the byte-identical orphan.
	componentID := "448422ce-fbf0-4e3d-98a1-fab0a6e856ed"
	sections := []SectionData{
		{ComponentName: "hero", ComponentID: "23f95f00-f293-466e-b43a-81791ea0fc6c"},
		{ComponentName: "tool-loan-vs-savings", ComponentID: componentID},
		{ComponentName: "ported-prose"},
	}
	p := preservedSection{slot: "tool-2", componentID: componentID, componentFunction: "tool-loan-vs-savings"}

	if got := matchPreservedSectionIdx(sections, p, map[int]bool{}); got != 1 {
		t.Errorf("matchPreservedSectionIdx = %d, want 1 — the stored tool IS in the set, under the "+
			"plan's name for it; judging it dropped is what duplicated the locked calculator "+
			"(bugs_open/385 §5c)", got)
	}
}

func TestMatchPreservedSectionIdx_FunctionArmDecidesWhenEnrichmentFailed(t *testing.T) {
	// Same shape, but the incoming side never resolved an id (enrichment can
	// fail without failing the save). The stored side still knows its
	// component's function — the preload's LEFT JOIN fetches it — and that is
	// the merge's arm 3, mirrored.
	sections := []SectionData{
		{ComponentName: "hero"},
		{ComponentName: "tool-loan-vs-savings"}, // no ComponentID
	}
	p := preservedSection{slot: "tool-2", componentID: "448422ce-fbf0-4e3d-98a1-fab0a6e856ed",
		componentFunction: "tool-loan-vs-savings"}

	if got := matchPreservedSectionIdx(sections, p, map[int]bool{}); got != 1 {
		t.Errorf("matchPreservedSectionIdx = %d, want 1 — the function arm must pair a stored "+
			"positional slot with the plan entry naming the same component even when the incoming "+
			"side carries no id", got)
	}
}

func TestMatchPreservedSectionIdx_SlotExactStillWins(t *testing.T) {
	// The pre-385 behaviour, kept: rerender-arm sets name sections by slot.
	sections := []SectionData{{ComponentName: "prose-0"}, {ComponentName: "tool-2"}}
	p := preservedSection{slot: "tool-2"}
	if got := matchPreservedSectionIdx(sections, p, map[int]bool{}); got != 1 {
		t.Errorf("matchPreservedSectionIdx = %d, want 1 (slot-exact arm)", got)
	}
}

func TestMatchPreservedSectionIdx_KebabArmNormalisesSlotVariants(t *testing.T) {
	// The 041 naming landmine, same as matchLockedRow's third arm: older rows
	// and plans carry snake_case/CamelCase variants of one slot.
	sections := []SectionData{{ComponentName: "loan-calc"}}
	p := preservedSection{slot: "loan_calc"}
	if got := matchPreservedSectionIdx(sections, p, map[int]bool{}); got != 0 {
		t.Errorf("matchPreservedSectionIdx = %d, want 0 (kebab-normalised slot arm)", got)
	}
}

func TestMatchPreservedSectionIdx_EmptyIdentitiesNeverPair(t *testing.T) {
	// Guarded on non-empty exactly as the sibling matchers are: an unresolved
	// incoming section (id "") must not be claimed by an idless stored row,
	// and an empty stored slot must not claim a nameless section.
	sections := []SectionData{{ComponentName: "", ComponentID: ""}}
	p := preservedSection{slot: "", componentID: "", componentFunction: ""}
	if got := matchPreservedSectionIdx(sections, p, map[int]bool{}); got != -1 {
		t.Errorf("matchPreservedSectionIdx = %d, want -1 — empty identities pairing is how every "+
			"unresolved section claims the first idless row", got)
	}
}

func TestMatchPreservedSectionIdx_OneRowClaimsOneSection(t *testing.T) {
	// The consumption rule both siblings already have. Two stored instances of
	// one component against a composition naming it once: the first claims the
	// entry, the second must NOT match — it is genuinely unrepresented and the
	// re-append arm preserves it.
	componentID := uuid.New().String()
	sections := []SectionData{{ComponentName: "tool-x", ComponentID: componentID}}
	first := preservedSection{slot: "tool-1", componentID: componentID}
	second := preservedSection{slot: "tool-2", componentID: componentID}

	claimed := map[int]bool{}
	if got := matchPreservedSectionIdx(sections, first, claimed); got != 0 {
		t.Fatalf("first stored instance: matchPreservedSectionIdx = %d, want 0", got)
	}
	claimed[0] = true
	if got := matchPreservedSectionIdx(sections, second, claimed); got != -1 {
		t.Errorf("second stored instance: matchPreservedSectionIdx = %d, want -1 — one incoming "+
			"entry cannot represent two stored instances; the second must be re-appended, not "+
			"silently judged present", got)
	}
}

// ---------------------------------------------------------------------------
// Wiring: the action must decide through the extracted matcher
// ---------------------------------------------------------------------------

// TestLayer2_UsesMatchPreservedSectionIdx keeps the seam wired to the decision
// the unit tests above pin — the same pattern as
// TestSplice_UsesAdoptCarriedProvenance, and needed for the same reason:
// someone re-inlining a slot-name comparison in the Layer 2 loop reverts
// production to the bugs_open/385 behaviour while every unit test here stays
// green, because the unit tests only know about the helper.
//
// MUTATION THAT MUST BREAK IT: replace the matchPreservedSectionIdx call in
// SavePageSectionsAction with an inline loop, or remove it.
func TestLayer2_UsesMatchPreservedSectionIdx(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)
	fd, ok := funcs["SavePageSectionsAction"]
	if !ok {
		t.Fatal("CONTROL FAILED: SavePageSectionsAction not found — the scan cannot see its target")
	}
	if !callsNamed(fd, "matchPreservedSectionIdx") {
		t.Error("SavePageSectionsAction no longer calls matchPreservedSectionIdx.\n" +
			"Layer 2's \"is this stored tool already represented?\" question must be answered by the " +
			"extracted matcher the unit tests above pin — an inline slot-name comparison is exactly " +
			"the bugs_open/385 defect returning. If the decision moved, point this test at its new " +
			"home; do not delete it.")
	}
}

// callsSelector reports whether fn's body calls pkg.name — the qualified form
// callsNamed cannot see.
func callsSelector(fn *ast.FuncDecl, pkg, name string) bool {
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == name {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == pkg {
				found = true
			}
		}
		return true
	})
	return found
}

// TestPairingRelation_BothMatchersCallTheSharedCore is the actions-side half
// of the drift guard (the datahelpers side pins MergeLockedPageSlots). The
// behaviour suites are equivalence proofs, so they stay green against a
// re-inlined private copy of the arms — re-inlining is precisely the drift
// that produced bugs_open/385 and drew council ece638fb's reuse gate.
//
// MUTATION THAT MUST BREAK IT: re-inline either matcher's arms.
func TestPairingRelation_BothMatchersCallTheSharedCore(t *testing.T) {
	funcs, _ := parsePackageFuncs(t)
	for fn, callee := range map[string]string{
		"matchLockedRow":           "PairIncomingToStored",
		"matchPreservedSectionIdx": "PairStoredToIncoming",
	} {
		fd, ok := funcs[fn]
		if !ok {
			t.Fatalf("CONTROL FAILED: %s not found — the scan cannot see its target", fn)
		}
		if !callsSelector(fd, "datahelpers", callee) {
			t.Errorf("%s no longer calls datahelpers.%s — the pairing relation must stay shared "+
				"(slot_pairing.go); a private copy of the arms is how the third asker drifted and "+
				"minted 385's orphan.", fn, callee)
		}
	}
}

// ---------------------------------------------------------------------------
// Whole action: the motivating case through the real wiring
// ---------------------------------------------------------------------------

// TestLayer2_PlanNamedToolIsNotReappendedBesideItsLockedRow drives the exact
// 2026-08-23 shape end to end: a locked, positionally-named,
// build_status='deployed' interactive row, and a build-arm composition that
// names the same component by function with the id enrichment already carries.
//
// Correct outcome: the incoming tool consumes the lock via matchLockedRow's
// identity arm (fresh copy DISCARDED, locked row repositioned — 058 working),
// and Layer 2 judges the stored tool represented, so nothing is appended. The
// only INSERT is the hero.
//
// ⚠ THIS TEST IS NOT THE MUTATION DETECTOR, and the reason is worth keeping:
// under the slot-name-only mutation Layer 2 re-appends the stored bytes, but
// the insert loop TOLERATES a failed INSERT (Warn + continue), so sqlmock's
// "unexpected call" error is swallowed, sections_saved still reads 1, and this
// test stays green — a guard in series hiding the mutation (measured while
// writing it, 2026-08-25). The mutation load is carried by the unit tests
// above (five of them fail on slot-only) and the wiring scan; this test pins
// the FIXED path end to end — the full expected call set of the correct
// outcome, with no second section INSERT among the expectations.
func TestLayer2_PlanNamedToolIsNotReappendedBesideItsLockedRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	lockedRowID := uuid.New()
	toolComponentID := uuid.New()

	// The stored row: the locked calculator, preloaded by Layer 2 because it is
	// deployed AND interactive — the fleet's one armed row, per §5c's census.
	preload := layer2PreloadWithFunction("tool-2", layer2ToolHTML, "",
		toolComponentID.String(), "tool-loan-vs-savings")

	// The same row, as loadActiveLockedRows returns it (identity arm data
	// included — the column matchLockedRow pairs on).
	locked := lockedRowSet().
		AddRow(lockedRowID, "tool-2", 2, "owner", "permanent", toolComponentID.String())

	// writable=1/locked=1/planned=2: the completeness floor sees the page it is
	// replacing (one writable hero + the locked tool) against a two-entry plan,
	// so both cohorts score and the save is not refused before the loop.
	expectSaveSlotReadsPreloading(mock, siteID, pageID, "tool-loan-vs-savings", locked,
		1, 1, 2, preload)

	// EXACTLY ONE section INSERT — the hero, first in the loop. The incoming
	// tool is discarded by the lock guard; the stored tool must NOT be
	// re-appended. A second section INSERT anywhere in this run is the 385
	// orphan, and sqlmock fails the run if it is attempted.
	expectSectionInsert(mock, pageID, 1, "hero")

	// The lock guard's branch for the incoming tool (i=1): reposition the
	// locked row to follow the composition, then surface the blocked change
	// (emitLockBlockedChangeItem runs behind insertWorkItem's two-strike
	// lookup, in its own transaction).
	mock.ExpectExec("UPDATE page_components SET position").
		WithArgs(lockedRowID, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-loan-vs-savings", []interface{}{
			map[string]interface{}{
				"rendered_html":    layer2HeroHTML,
				"stored_slot_name": "hero",
				"component_name":   "hero",
			},
			map[string]interface{}{
				// The plan's name for the tool, with the id the enricher
				// resolves — a fresh interactive render of the same component.
				"rendered_html":    layer2ToolHTML,
				"stored_slot_name": "tool-loan-vs-savings",
				"component_name":   "tool-loan-vs-savings",
				"component_id":     toolComponentID.String(),
			},
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := saveResult(t, out)
	if got := res["sections_saved"]; got != 1 {
		t.Errorf("sections_saved = %v, want 1 — a second saved section here is bugs_open/385's "+
			"byte-identical orphan being re-minted", got)
	}
	preservedSlots, _ := res["locked_sections_preserved"].([]string)
	if len(preservedSlots) != 1 || preservedSlots[0] != "tool-2" {
		t.Errorf("locked_sections_preserved = %v, want [tool-2] — the incoming tool must be the "+
			"copy the lock guard discards", res["locked_sections_preserved"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations — the fixed path did not run as pinned: %v", err)
	}
}
