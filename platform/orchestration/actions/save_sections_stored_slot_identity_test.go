// FILE: platform/orchestration/actions/save_sections_stored_slot_identity_test.go
//
// bugs_open/189 — a locked, positionally-named section is DUPLICATED rather than
// preserved the first time it becomes resolvable.
//
// Two pre-existing defects compound. extractSectionsFromMetadata preferred
// component_function over the stored slot name, so a slot called "tool-2" was
// persisted as "tool-loan-vs-savings" the moment bugs_closed/182 made it
// resolvable; matchLockedRow then matched that RENAMED section against locked
// rows still called "tool-2", found nothing, and the fresh copy was INSERTed
// beside the locked row the guard exists to protect. The live incident left five
// rows where there should have been four and rendered the calculator twice.
//
// The fix threads the page_components row's own slot identity through the
// metadata as stored_slot_name and has the save prefer it VERBATIM. These tests
// pin the four properties that must hold:
//
//   - the locked-row guard fires again on a positional slot (the regression);
//   - a positional name survives the save (the diagnostic property the
//     decomposed sites were designed around — "which paragraph vanished");
//   - the field's ABSENCE is byte-for-byte today's behaviour, normalisation
//     included (tool-recreation has no structured slot identity, and any
//     orchestration expanded before the producers gained the field has none
//     either);
//   - the field's PRESENCE is never normalised — it is a row identity being
//     matched back to the row that issued it, so a legacy spelling must survive
//     intact or the match it exists to restore is lost again.
//
// LEVELS. The four properties are each pinned twice: once as a unit
// (extractSectionsFromMetadata, and its composition with matchLockedRow — the
// two functions the bug file quotes) and once through the whole of
// SavePageSectionsAction against sqlmock, which is what proves the slot name
// reaching the INSERT is the one the producer sent. The unit tests fail crisply
// under the mutation; the action test proves the same thing where it matters.
package actions

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// storedSlotHTML is a resolved, non-stub section: real visible text and no
// section--generic marker, so neither the bugs_open/039 stub guard nor the
// interactivity guards have anything to say about it.
const storedSlotHTML = `<section class="section section--tool"><div class="container">` +
	`<h2>Loan vs savings</h2><p>Compare the cost of borrowing against the return on saving.</p></div></section>`

// metaEntry builds one sections_metadata entry. A key given as "" is OMITTED, so
// a test can express "this producer never set the field" as something different
// from "it set it to an empty string" — which is the distinction the fallthrough
// rule turns on.
func metaEntry(storedSlot, componentFunction, componentName, componentID string) map[string]interface{} {
	m := map[string]interface{}{"rendered_html": storedSlotHTML}
	if storedSlot != "" {
		m["stored_slot_name"] = storedSlot
	}
	if componentFunction != "" {
		m["component_function"] = componentFunction
	}
	if componentName != "" {
		m["component_name"] = componentName
	}
	if componentID != "" {
		m["component_id"] = componentID
	}
	return m
}

func onlySection(t *testing.T, meta ...interface{}) SectionData {
	t.Helper()
	got := extractSectionsFromMetadata(meta, zap.NewNop())
	if len(got) != 1 {
		t.Fatalf("expected exactly one section out of %d metadata entries, got %d", len(meta), len(got))
	}
	return got[0]
}

// ---------------------------------------------------------------------------
// Unit: extractSectionsFromMetadata's precedence
// ---------------------------------------------------------------------------

// (e) MUTATION THAT MUST BREAK IT: delete the stored_slot_name branch. The
// component function wins again and the section is renamed — defect #1 of the
// bug file, in one assertion.
func TestExtractSectionsFromMetadata_StoredSlotNameWinsOverFunctionAndName(t *testing.T) {
	s := onlySection(t, metaEntry("tool-2", "tool-loan-vs-savings", "tool-2", uuid.NewString()))
	if s.ComponentName != "tool-2" {
		t.Fatalf("the page's own slot identity must win: want tool-2, got %q", s.ComponentName)
	}
}

// (e) The empty-string case, which is NOT the absent case: a producer that sets
// the key but has nothing to put in it must fall through to today's precedence
// rather than persisting a nameless slot.
func TestExtractSectionsFromMetadata_EmptyStoredSlotNameFallsThroughToTodaysPrecedence(t *testing.T) {
	s := onlySection(t, map[string]interface{}{
		"rendered_html":      storedSlotHTML,
		"stored_slot_name":   "",
		"component_function": "ported-prose",
		"component_name":     "prose-0",
	})
	if s.ComponentName != "ported-prose" {
		t.Fatalf("an empty stored_slot_name must fall through to component_function, got %q", s.ComponentName)
	}
}

// (d) MUTATION THAT MUST BREAK IT: move the NormalizeComponentFunction call back
// outside the else, so it applies to the stored name too. A legacy snake_case
// slot is then rewritten to kebab and no longer matches the row that issued it —
// the un-match this field exists to prevent.
func TestExtractSectionsFromMetadata_StoredSlotNameIsNeverNormalised(t *testing.T) {
	s := onlySection(t, metaEntry("legacy_snake_slot", "ported-prose", "", uuid.NewString()))
	if s.ComponentName != "legacy_snake_slot" {
		t.Fatalf("a stored slot name is a row identity and must survive verbatim, got %q", s.ComponentName)
	}
}

// (c) Absence is today's behaviour, and "today" includes the kebab contract that
// bugs_closed/041 put on the DERIVED name. A test that only checked the new
// branch would let the else branch be simplified into a rename.
func TestExtractSectionsFromMetadata_AbsentFieldIsTodaysBehaviour(t *testing.T) {
	cases := []struct {
		name string
		meta map[string]interface{}
		want string
	}{
		{
			name: "component_function first",
			meta: metaEntry("", "ported-prose", "prose-0", ""),
			want: "ported-prose",
		},
		{
			name: "component_name second, when no function arrived (the carry path)",
			meta: metaEntry("", "", "prose-0", ""),
			want: "prose-0",
		},
		{
			name: "derived names still normalise to kebab (041)",
			meta: metaEntry("", "ported_prose", "", ""),
			want: "ported-prose",
		},
		{
			name: "neither key: the generic default, normalised",
			meta: map[string]interface{}{"rendered_html": storedSlotHTML},
			want: "section",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := onlySection(t, tc.meta).ComponentName; got != tc.want {
				t.Errorf("ComponentName = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Unit: the composition the bug file actually describes
// ---------------------------------------------------------------------------

// (a) The 189 mechanism at its crispest — the two quoted functions, composed.
// The locked row is called "tool-2"; the incoming section resolves to the
// component "tool-loan-vs-savings". Whether the guard fires is decided entirely
// by which of those two names extractSectionsFromMetadata put on the section.
//
// MUTATION THAT MUST BREAK IT: delete the stored_slot_name branch. ComponentName
// becomes "tool-loan-vs-savings", matchLockedRow returns nil, and the caller
// goes on to INSERT a duplicate beside the locked row.
func TestStoredSlotNameRestoresTheLockedRowMatch(t *testing.T) {
	locked := []*lockedPageRow{{id: uuid.New(), slot: "tool-2", position: 3, lockedBy: "admin", lockType: "permanent"}}

	s := onlySection(t, metaEntry("tool-2", "tool-loan-vs-savings", "tool-2", uuid.NewString()))
	if lr := matchLockedRow(locked, s.ComponentName, ""); lr == nil {
		t.Fatalf("the locked row must be matched by the stored slot name; section came through as %q", s.ComponentName)
	}

	// The control that makes the assertion above mean something: the SAME locked
	// rows, the SAME component, but no stored slot identity — which is the state
	// of the world this bug was filed from, and it still misses. Without this,
	// the test above would pass against a matchLockedRow that matched anything.
	locked[0].consumed = false
	bare := onlySection(t, metaEntry("", "tool-loan-vs-savings", "tool-2", uuid.NewString()))
	if lr := matchLockedRow(locked, bare.ComponentName, ""); lr != nil {
		t.Fatalf("control failed: %q should not match a row called tool-2 — the test above proves nothing", bare.ComponentName)
	}
}

// ---------------------------------------------------------------------------
// Producers: the re-render path, which holds the stored row in hand
// ---------------------------------------------------------------------------
//
// Kept beside the consumer deliberately. The field is only ever useful as a pair
// — a producer that stops emitting it and a save that keeps preferring it fail
// silently, back to exactly the rename this bug is about.

// MUTATION THAT MUST BREAK IT: drop stored_slot_name from the success entry in
// rerender_page_sections_action.go. The metadata then names only the component,
// and the save renames tool-2 to tool-loan-vs-savings — the live incident.
func TestRerenderPageSections_SuccessEntryCarriesTheStoredSlotName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	pinnedID := uuid.New().String()

	// The 189 shape: a positionally-named, locked slot whose component resolves
	// only by id (bugs_closed/182's repair population).
	expectPageAndSections(mock, siteID, pageID, "tool-loan-vs-savings", "loancalculator.co.uk",
		// id and parent_instance_id lead the column list (035 P1): loadStoredSections
		// now selects them so the composition walk can group children under parents.
		// "" for parent_instance_id is an ordinary top-level section, which is what
		// every row on the estate is today.
		*sqlmock.NewRows([]string{"id", "parent_instance_id", "component_id", "slot_name", "content_data", "rendered_html", "position", "component_version_id"}).
			AddRow(uuid.New().String(), "", pinnedID, "tool-2", []byte(`{"headline":"Loan vs savings"}`), "<section>old</section>", 3, ""))

	// Name and function passes both miss — "tool-2" is nobody's identity.
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("FROM content_components").WillReturnRows(emptyComponentRows())
	// by-id resolves to the calculator component.
	mock.ExpectQuery("FROM content_components").WillReturnRows(
		emptyComponentRows().AddRow(
			pinnedID, "Loan vs Savings Calculator", "Loan vs Savings Calculator", "tool-loan-vs-savings",
			"", nil, nil, "<section>{{.headline}}</section>", nil, "template", nil, "section"))
	expectBaseSiteData(mock)

	out, err := RerenderPageSectionsAction(context.Background(), rerenderParams(db, siteID, "tool-loan-vs-savings"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sections, _ := out.(map[string]interface{})["sections_metadata"].([]map[string]interface{})
	if len(sections) != 1 {
		t.Fatalf("expected exactly one section, got %d", len(sections))
	}
	if got, _ := sections[0]["stored_slot_name"].(string); got != "tool-2" {
		t.Fatalf("stored_slot_name = %q, want tool-2 (entry: %v)", got, sections[0])
	}
	// The component's own identity is still reported beside it — the two facts
	// travel together now, which is the whole point.
	if fn, _ := sections[0]["component_function"].(string); fn != "tool-loan-vs-savings" {
		t.Errorf("component_function = %q, want tool-loan-vs-savings", fn)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// carryStoredSection is the second producer in that file. Stating the field here
// is a no-op today (a carry sets no component_function, so the save already fell
// through to component_name) — which is exactly why it needs a test: a no-op is
// invisible, and the next edit to either producer must not be able to leave the
// two disagreeing about what they promise the save.
//
// MUTATION THAT MUST BREAK IT: drop the key from carryStoredSection.
func TestCarryStoredSection_StatesTheStoredSlotName(t *testing.T) {
	m := carryStoredSection(storedSection{
		slotName:     "prose-1",
		componentID:  uuid.NewString(),
		renderedHTML: storedSlotHTML,
	})
	if got, _ := m["stored_slot_name"].(string); got != "prose-1" {
		t.Fatalf("stored_slot_name = %q, want prose-1 (entry: %v)", got, m)
	}
	if got, _ := m["component_name"].(string); got != "prose-1" {
		t.Errorf("component_name must be unchanged for the diagnostic readers, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Whole action, against sqlmock
// ---------------------------------------------------------------------------

// saveSlotParams rigs SavePageSectionsAction on the structured-metadata path,
// with the three config keys every live caller sets.
func saveSlotParams(db *sql.DB, siteID uuid.UUID, pageName string, meta []interface{}) ActionParams {
	return ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		StepConfig: models.Step{Config: map[string]interface{}{
			"sections_metadata_field": "rerender_sections.sections_metadata",
			"page_name_field":         "current_page.name",
			"site_id_field":           "site_record.site_id",
		}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"name": pageName},
			"site_record":  map[string]interface{}{"site_id": siteID.String()},
			"rerender_sections": map[string]interface{}{
				"sections_metadata": meta,
			},
		},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "44444444-4444-4444-4444-444444444444",
			StepName:        "save_page_sections",
		},
	}
}

// expectSaveSlotReads wires every read SavePageSectionsAction performs before the
// insert loop, for a page whose refusal guards all stand down: not owned, no
// deployed text to regress against, no interactive markup, no shrink-floor
// population. `locked` is what loadActiveLockedRows returns; writable/lockedRows/
// planned feed the completeness floor so both its cohorts score 100% — this suite
// is about slot IDENTITY, and a floor refusal aborts before the loop the
// assertions read.
//
// Expectations are ORDERED deliberately. This action issues several queries
// against page_components, and an unordered set would let one regexp match the
// wrong one and still pass.
// expectSaveSlotReads keeps the original signature — a page with NO stored
// interactive section, which is what every case in this file wants. The Layer 2
// preload is parameterised in the variant below because bugs_open/357's stamp
// hygiene has to stage a stored row there, and until it did, nothing in the suite
// ever returned one: the splice arm had no test at all.
func expectSaveSlotReads(mock sqlmock.Sqlmock, siteID, pageID uuid.UUID, pageName string,
	locked *sqlmock.Rows, writable, lockedRows, planned int) {

	expectSaveSlotReadsPreloading(mock, siteID, pageID, pageName, locked,
		writable, lockedRows, planned, layer2PreloadRows())
}

// layer2PreloadRows is the Layer 2 preload's column set, in one place so a change
// to that query's SELECT list does not have to be chased through two files.
func layer2PreloadRows() *sqlmock.Rows {
	// ⚠ This list must match the Layer 2 preload query's SELECT exactly. A short
	// row makes rows.Scan fail, the loop logs and skips, and Layer 2 NEVER RUNS —
	// every assertion downstream then passes while testing nothing. Documented in
	// save_sections_layer2_provenance_test.go, and it has happened once already.
	return sqlmock.NewRows([]string{"slot_name", "rendered_html", "content_data", "component_version_id", "component_id", "function", "is_active"})
}

func expectSaveSlotReadsPreloading(mock sqlmock.Sqlmock, siteID, pageID uuid.UUID, pageName string,
	locked *sqlmock.Rows, writable, lockedRows, planned int, preload *sqlmock.Rows) {

	// saveSectionsLookupPageID
	mock.ExpectQuery("SELECT id, url FROM pages").
		WithArgs(siteID, pageName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(pageID, "/tools/loan-vs-savings.html"))

	// Page-ownership guard (rebuild_policy)
	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(policyRows("generic", false))

	// Layer 2 interactive-section preload. ⚠ The matcher must track the query's
	// actual text: when phase 2 added a LEFT JOIN and qualified the columns, the
	// old "SELECT slot_name, rendered_html, content_data" pattern stopped matching
	// this query and started matching a LATER page_components read, which
	// desynchronised every ordered expectation after it and produced a refusal
	// about the text floor. The failure names a completely unrelated guard.
	mock.ExpectQuery("SELECT pc.slot_name, pc.rendered_html, pc.content_data").
		WithArgs(pageID).
		WillReturnRows(preload)

	// Link repair: the site's page URL index
	mock.ExpectQuery("FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"url"}))

	// Claims guard: the site's registered evidence base
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}))

	// Page-total text floor — nothing deployed, so it stands down.
	// It reads rendered_html now rather than SUM()ing a tag-stripped length in
	// SQL (bugs_open/293): visible text needs a real parse, and REGEXP_REPLACE
	// cannot exclude what is inside <style>. So this is a THIRD read of
	// page_components and the ordered expectations below count from here.
	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}))

	// Per-slot shrink floor (bugs_open/178)
	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}))

	// Per-slot component floor (bugs_open/253). A SECOND read of
	// page_components, immediately after the shrink floor's. It must be
	// expected separately even though the regexp above would also match it:
	// expectations here are ORDERED (see this helper's comment), so the shrink
	// query consumes the first and this one consumes the second. Omitting it
	// does not merely leave a query unexpected — the guard fails CLOSED on a
	// measurement error, so the whole save is refused and every assertion after
	// the insert loop dies with a misleading "could not measure" error.
	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}))

	// Interactivity regression guard
	mock.ExpectQuery("SELECT COALESCE\\(bool_or").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"bool_or"}).AddRow(false))

	// loadActiveLockedRows (bugs_open/058)
	mock.ExpectQuery("SELECT id, COALESCE").
		WithArgs(pageID).
		WillReturnRows(locked)

	// Completeness floor (bugs_open/165)
	mock.ExpectQuery("FROM pages p WHERE p.id").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"writable", "locked", "planned", "suppressed"}).
			AddRow(writable, lockedRows, planned, 0))

	// content_data regression record (bugs_open/194): records, never refuses
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM page_components").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// content_brief population
	mock.ExpectQuery("SELECT COALESCE\\(page_spec").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"purpose"}).AddRow(""))

	// History snapshot, then the reconciliation DELETE
	mock.ExpectExec("INSERT INTO page_component_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM page_components").WillReturnResult(sqlmock.NewResult(0, 0))
}

// component_id is the 6th column loadActiveLockedRows selects (the identity arm
// of matchLockedRow, 2026-08-12). Callers of this helper AddRow five values and
// so leave it empty, which is deliberate: these fixtures exist to exercise the
// SLOT-NAME branch, and giving a locked row the same component id as the
// incoming section would route them through the identity branch instead —
// passing while silently no longer testing what their names claim. The identity
// branch has its own fixtures in save_sections_locked_identity_test.go.
func lockedRowSet() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "slot_name", "position", "locked_by", "lock_type", "component_id"})
}

// expectSectionInsert asserts the slot_name the INSERT actually receives — the
// 4th bind, the column this whole change is about.
func expectSectionInsert(mock sqlmock.Sqlmock, pageID uuid.UUID, position int, slotName string) {
	// The trailing AnyArg is component_version_id (RFC_046's provenance stamp).
	// It is AnyArg rather than nil because this helper is used by cases that do
	// and do not carry a stamp; the stamp's own behaviour is asserted in
	// save_sections_component_version_test.go, not smuggled in here.
	mock.ExpectExec("INSERT INTO page_components").
		WithArgs(pageID, position, sqlmock.AnyArg(), slotName,
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

func saveResult(t *testing.T, out interface{}) map[string]interface{} {
	t.Helper()
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected a result map, got %T", out)
	}
	return m
}

// (a) THE 189 REGRESSION, through the whole action.
//
// A page holding one human-locked row, slot_name "tool-2" (bugs_closed/182 made
// it resolvable; the component it resolves to is called "tool-loan-vs-savings").
// The correct outcome is the locked-slot branch: the locked row is repositioned
// to follow the new composition, a lock_blocked_change item is raised, and the
// fresh copy is DISCARDED — no INSERT for that section at all.
//
// MUTATION THAT MUST BREAK IT: delete the stored_slot_name branch from
// extractSectionsFromMetadata. The section arrives as "tool-loan-vs-savings",
// matchLockedRow misses, and the loop takes the INSERT path instead — which this
// expectation set does not contain, so the UPDATE/work-item expectations go
// unfulfilled and ExpectationsWereMet reports them. (The reposition's position
// bind changes too: the in-loop branch writes 1, the trailing retained-rows loop
// writes len(sections)+1 = 2, so even a mutation that somehow reached the UPDATE
// would fail on the argument.)
func TestSavePageSections_LockedPositionalSlotIsPreservedNotDuplicated(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID, lockedID := uuid.New(), uuid.New(), uuid.New()

	// One locked row; nothing agent-writable. The floor's plan cohort therefore
	// nets to zero planned sections, which is correct: a save whose only section
	// is held by a lock writes nothing and must not be refused for it.
	locked := lockedRowSet().AddRow(lockedID.String(), "tool-2", 3, "admin", "permanent", "")
	expectSaveSlotReads(mock, siteID, pageID, "tool-loan-vs-savings", locked, 0, 1, 1)

	// The locked-slot branch: reposition to follow the new composition (position
	// 1 — this is the first incoming section), then surface the blocked change.
	mock.ExpectExec("UPDATE page_components SET position").
		WithArgs(lockedID, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// emitLockBlockedChangeItem runs its insert in its own transaction, behind
	// insertWorkItem's two-strike lookup.
	mock.ExpectBegin()
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Deliberately NOT expected: INSERT INTO page_components. Its absence from
	// this set is the assertion that the duplicate row is never written.

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-loan-vs-savings", []interface{}{
			metaEntry("tool-2", "tool-loan-vs-savings", "tool-2", uuid.NewString()),
		}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := saveResult(t, out)
	if got := m["sections_saved"]; got != 0 {
		t.Errorf("a locked slot must write no row: sections_saved = %v, want 0", got)
	}
	preserved, _ := m["locked_sections_preserved"].([]string)
	if len(preserved) != 1 || preserved[0] != "tool-2" {
		t.Errorf("locked_sections_preserved = %v, want [tool-2]", m["locked_sections_preserved"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations — the locked-slot branch did not run: %v", err)
	}
}

// (b) Positional identity survives the save. This is the property the
// loancalculator decomposition was designed around ("so that a dropped-section
// warning names which paragraph vanished"), and the one that regresses
// fleet-wide, locked or not, the moment a positional slot becomes resolvable.
//
// MUTATION THAT MUST BREAK IT: delete the stored_slot_name branch. slot_name is
// then "ported-prose" and the WithArgs bind fails.
func TestSavePageSections_PositionalSlotNameSurvivesTheInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	expectSaveSlotReads(mock, siteID, pageID, "tool-loan-vs-savings", lockedRowSet(), 1, 0, 1)
	expectSectionInsert(mock, pageID, 1, "prose-0")

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-loan-vs-savings", []interface{}{
			metaEntry("prose-0", "ported-prose", "prose-0", uuid.NewString()),
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

// (c) Absence is today, byte for byte — including the kebab normalisation the
// derived path has carried since bugs_closed/041. Three save consumers were
// measured live and one of them (tool-recreation) has no structured slot
// identity at all; this is the test that keeps working for it.
//
// MUTATION THAT MUST BREAK IT: drop the else branch and default the name to the
// stored slot (i.e. make absence mean "section"), or move the normalise call
// inside the stored-slot branch and out of the derived one.
func TestSavePageSections_AbsentStoredSlotNameKeepsTodaysDerivedName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	expectSaveSlotReads(mock, siteID, pageID, "tool-loan-vs-savings", lockedRowSet(), 1, 0, 1)
	// snake_case in, kebab-case out: the derived-name contract, unchanged.
	expectSectionInsert(mock, pageID, 1, "ported-prose")

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-loan-vs-savings", []interface{}{
			metaEntry("", "ported_prose", "prose-0", uuid.NewString()),
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

// (d) Verbatim, all the way to the column. A stored name is a row identity being
// matched back to the row that issued it; normalising a legacy spelling would
// un-match it, which is the failure this whole change exists to stop.
//
// MUTATION THAT MUST BREAK IT: apply NormalizeComponentFunction to the stored
// name. slot_name becomes "legacy-snake-slot" and the bind fails.
func TestSavePageSections_StoredSlotNameIsInsertedVerbatim(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()
	expectSaveSlotReads(mock, siteID, pageID, "tool-loan-vs-savings", lockedRowSet(), 1, 0, 1)
	expectSectionInsert(mock, pageID, 1, "legacy_snake_slot")

	out, err := SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-loan-vs-savings", []interface{}{
			metaEntry("legacy_snake_slot", "ported-prose", "", uuid.NewString()),
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
