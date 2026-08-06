// FILE: platform/orchestration/actions/plan_sections_slot_identity_test.go
//
// bugs_open/204 — the BUILD path's half of the blindness bugs_open/182 fixed on
// the re-render path. PlanSectionsAction resolved a page's section names against
// component name/function only, so on a "decomposed" site — where pages.sections
// holds POSITIONAL slot names ("prose-0", "tool-2") that are no component's name
// or function — every section fell through to the selector, deferred, and filed
// junk needs_new_component work items asking the fleet to build components that
// already exist and are already pinned to the page.
//
// The page's own page_components rows carry the answer: component_id is the
// row's identity and does not depend on slot naming at all. These tests pin the
// fix (Path 0), and pin it to the SAME semantics the re-render path settled on,
// so the two call sites of this judgement cannot drift apart again:
//   - id first, and the id WINS when both routes resolve and disagree;
//   - an id whose row failed the template guard must NOT fall back to a
//     coincidentally name-matched component (the silent substitution, one level
//     down), and must not file needs_new_component for a component that exists;
//   - an id resolving to no active row falls through to the name/selector paths;
//   - a page with no stored rows plans exactly as it did before the change.
package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// planSectionsSchema is the minimal v2 input_schema that makes planSection mark
// a section ready without any resolver queries: one LLM-sourced field, which is
// "always available" and therefore never reaches the source resolver.
const planSectionsSchema = `{"fields":{"content":{"type":"rich_text","source":"llm","required":true}}}`

// planParams builds the ActionParams shape plan_sections reads. The live caller
// (page-build-handler's plan_sections step) maps all three keys via config
// dot-paths, which is what Strategy 0 of ExtractActionInputs resolves.
func planParams(db *sql.DB, siteID uuid.UUID, pageName string, sections []string) ActionParams {
	return ActionParams{
		Context: context.Background(),
		DB:      db,
		Logger:  zap.NewNop(),
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{
				"site_id":   siteID.String(),
				"page_name": pageName,
				"sections":  sections,
			},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id":   "input_data.site_id",
			"page_name": "input_data.page_name",
			"sections":  "input_data.sections",
		}},
		ExecutionContext: &orchtypes.ExecutionContext{
			OrchestrationID: "33333333-3333-3333-3333-333333333333",
			StepName:        "plan_sections",
		},
	}
}

// componentRow builds one content_components row in scanSectionComponentRow's
// column order (componentColumns, shared with the rerender resolve tests).
func componentRow(id, name, function, htmlTemplate, level string) *sqlmock.Rows {
	return sqlmock.NewRows(componentColumns).AddRow(
		id, name, name, function, "", nil, nil, htmlTemplate, planSectionsSchema, "template", nil, level)
}

func slotRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"slot_name", "component_id"})
}

// expectPostLoopReads wires the two reads every planning run does between the
// component loads and the section loop: site_specs (resolver.ensureSpecs) and
// the open needs_section_data items (loadOpenSectionDataRequests).
func expectPostLoopReads(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"aspect", "data"}))
	mock.ExpectQuery("FROM site_work_items").
		WillReturnRows(sqlmock.NewRows([]string{"section_name", "summary"}))
}

// readyItems digs the typed plan items out of the action's result map.
func readyItems(t *testing.T, out interface{}) []sectionPlanItem {
	t.Helper()
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected a result map, got %T", out)
	}
	items, ok := m["sections_ready"].([]sectionPlanItem)
	if !ok {
		t.Fatalf("expected sections_ready to be []sectionPlanItem, got %T", m["sections_ready"])
	}
	return items
}

func deferredItems(t *testing.T, out interface{}) []sectionPlanItem {
	t.Helper()
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected a result map, got %T", out)
	}
	items, ok := m["sections_deferred"].([]sectionPlanItem)
	if !ok {
		t.Fatalf("expected sections_deferred to be []sectionPlanItem, got %T", m["sections_deferred"])
	}
	return items
}

// TestPlanSections_PositionalSlotResolvesByStoredComponentID is the bugs_open/204
// regression: a positional slot name matches no component by name OR function,
// but the page's own page_components row names the component exactly. Before the
// fix this section deferred and filed a needs_new_component item for a component
// that was already built and already pinned to this very page.
func TestPlanSections_PositionalSlotResolvesByStoredComponentID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	componentA := uuid.New().String()

	// loadComponentSchemas: name pass then function pass, both miss — "prose-0"
	// is a positional slot name, not any component's identity.
	mock.ExpectQuery("WHERE name IN").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("WHERE function IN").WillReturnRows(emptyComponentRows())

	// The new identity read, scoped by (site_id, page name).
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "tool-loan-vs-savings").
		WillReturnRows(slotRows().AddRow("prose-0", componentA))

	// …and the by-id schema load for the ids it named.
	mock.ExpectQuery("id = ANY").WillReturnRows(
		componentRow(componentA, "ported-prose", "ported-prose", "<section>hi</section>", "section"))

	expectPostLoopReads(mock)

	// One ready section ⇒ persistSectionSkips runs its (no-op-guarded) merge.
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	// Deliberately NOT expected: the selector query, and any
	// INSERT INTO site_work_items. sqlmock errors on an unexpected call, so
	// their absence from this list is the assertion that no junk
	// needs_new_component item was filed.

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "tool-loan-vs-savings", []string{"prose-0"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := out.(map[string]interface{})
	names, _ := m["ready_names"].([]string)
	if len(names) != 1 || names[0] != "prose-0" {
		t.Fatalf("expected prose-0 to plan ready via its stored component_id, got ready_names=%v deferred=%v",
			m["ready_names"], m["sections_deferred"])
	}
	items := readyItems(t, out)
	if items[0].ComponentID != componentA {
		t.Errorf("expected the stored component_id %s on the plan item, got %s", componentA, items[0].ComponentID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_StoredIDWinsOverNameWhenBothResolve pins the resolution
// ORDER, mirroring the re-render path: when a slot name coincidentally matches a
// generic component AND the page's own row names a different one, the page's row
// wins. The name match is the coincidence; the id is the page saying which
// component it is.
func TestPlanSections_StoredIDWinsOverNameWhenBothResolve(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pinnedA := uuid.New().String()
	genericB := uuid.New().String()

	// Pass 1 (name) resolves — to the generic, site-agnostic hero.
	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(genericB, "hero", "hero", "<section>GENERIC</section>", "section"))

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows().AddRow("hero", pinnedA))

	mock.ExpectQuery("id = ANY").WillReturnRows(
		componentRow(pinnedA, "webdesign.uk Two-Column Hero", "hero", "<section>PINNED</section>", "section"))

	expectPostLoopReads(mock)
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected exactly one ready section, got %d", len(items))
	}
	if items[0].ComponentID != pinnedA {
		t.Errorf("the page's own pinned component must win over the name match: want %s, got %s",
			pinnedA, items[0].ComponentID)
	}
	if items[0].ComponentID == genericB {
		t.Error("the generic name-matched component must not be planned when component_id resolves")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_InvalidPinnedTemplateDefersLoudlyAndFilesNoNewComponentItem
// covers the second route into an empty by-id map (bugs_open/024's class): the
// pinned id resolves to a row that EXISTS but fails the template-truncation
// guard. Filing needs_new_component here would ask the fleet to build what is
// already there, so the section defers with a reason naming the component to
// repair — and, unlike the re-render path, does not fail the whole run.
func TestPlanSections_InvalidPinnedTemplateDefersLoudlyAndFilesNoNewComponentItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	componentA := uuid.New().String()
	// Long, opened <section> with no closing tag — the truncated-generation
	// signature sectionTemplateValid rejects.
	brokenTemplate := "<section>" + strings.Repeat("x", 120)

	mock.ExpectQuery("WHERE name IN").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("WHERE function IN").WillReturnRows(emptyComponentRows())

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "tool-loan-vs-savings").
		WillReturnRows(slotRows().AddRow("prose-0", componentA))

	mock.ExpectQuery("id = ANY").WillReturnRows(
		componentRow(componentA, "ported-prose", "ported-prose", brokenTemplate, "section"))

	expectPostLoopReads(mock)

	// The deferred section's durable trace: a needs_section_data item, NOT a
	// needs_new_component one. createDeferredItems writes exactly one insert.
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(1, 1))

	// No UPDATE pages: nothing planned ready and nothing skipped, so
	// persistSectionSkips is not called at all.

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "tool-loan-vs-savings", []string{"prose-0"}))
	if err != nil {
		t.Fatalf("a single broken pinned template must not fail the run: %v", err)
	}

	items := deferredItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected exactly one deferred section, got %d", len(items))
	}
	if !strings.Contains(items[0].Reason, componentA) {
		t.Errorf("the deferral must name the component to repair, got: %s", items[0].Reason)
	}
	if !strings.Contains(items[0].Reason, "do not create a new one") {
		t.Errorf("the deferral must say the component already exists, got: %s", items[0].Reason)
	}
	if len(readyItems(t, out)) != 0 {
		t.Error("a component dropped by the template guard must not plan ready")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_NoStoredRowsPreservesNamePath is the no-op case: an initial
// build has no page_components rows yet, so the identity map is empty and
// planning must behave exactly as it did before the change — including issuing
// no by-id query at all.
func TestPlanSections_NoStoredRowsPreservesNamePath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	componentC := uuid.New().String()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(componentC, "hero", "hero", "<section>hi</section>", "section"))

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows())

	// Deliberately NOT expected: the by-id load. An empty identity map must not
	// cost a query.

	expectPostLoopReads(mock)
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"hero"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 || items[0].ComponentID != componentC {
		t.Fatalf("the name path must still resolve when there are no stored rows, got %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_ConflictingDuplicateSlotFallsBackToName pins the repeated
// slot_name rule. A slot_name used two or three times on one page is normal
// (generic-text-block does exactly that); repeats that DISAGREE about which
// component they are, however, give no basis for picking one, so the slot drops
// out of the identity map and the name path governs — rather than resolving to
// whichever row the ORDER BY happened to yield first.
func TestPlanSections_ConflictingDuplicateSlotFallsBackToName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	componentA := uuid.New().String()
	componentB := uuid.New().String()
	componentC := uuid.New().String()

	// Unordered, because the by-id expectation below is a TRAP that must be
	// allowed to go unfulfilled — in order mode an unfulfilled expectation
	// derails every query after it and the failure would be unreadable.
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("WHERE name IN").WillReturnRows(
		componentRow(componentC, "x", "x", "<section>hi</section>", "section"))

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows().
			AddRow("x", componentA).
			AddRow("x", componentB))

	// THE TRAP, and the reason this test can come out false: if the conflict
	// were resolved by taking whichever row arrived first, the identity map
	// would be non-empty, a by-id load WOULD be issued, and it would resolve to
	// componentA — which the assertion below then catches. Simply omitting the
	// expectation proves nothing: loadContentComponentsByID swallows its own
	// query error, so an unexpected call would leave every observable identical
	// to the correct behaviour. This expectation is deliberately never asserted
	// as met.
	mock.ExpectQuery("id = ANY").WillReturnRows(
		componentRow(componentA, "a-component", "a-component", "<section>A</section>", "section"))

	expectPostLoopReads(mock)
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 0))

	out, err := PlanSectionsAction(context.Background(),
		planParams(db, siteID, "index", []string{"x"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items := readyItems(t, out)
	if len(items) != 1 {
		t.Fatalf("expected exactly one ready section, got %d", len(items))
	}
	if items[0].ComponentID != componentC {
		t.Errorf("a disagreeing repeated slot must fall back to the name path (want %s), got %s",
			componentC, items[0].ComponentID)
	}
	if items[0].ComponentID == componentA || items[0].ComponentID == componentB {
		t.Error("neither disagreeing row may be picked — there is no basis for choosing one")
	}
	// No ExpectationsWereMet here, deliberately: see THE TRAP above.
}

// TestLoadPageSlotComponentIDs_DropsDisagreeingRepeatsKeepsAgreeingOnes is the
// crisp unit of the rule the action-level test can only observe indirectly. A
// slot_name repeated with the SAME component_id is normal and must map; the same
// slot_name repeated with DIFFERENT ids must map to nothing at all.
func TestLoadPageSlotComponentIDs_DropsDisagreeingRepeatsKeepsAgreeingOnes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	agreeing := uuid.New().String()
	first := uuid.New().String()
	second := uuid.New().String()

	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "index").
		WillReturnRows(slotRows().
			AddRow("generic-text-block", agreeing).
			AddRow("generic-text-block", agreeing).
			AddRow("conflicted", first).
			AddRow("conflicted", second).
			AddRow("", first).             // a NULL/empty slot_name is skipped
			AddRow("no-component-id", "")) // as is a row with no component_id

	slotIDs, err := loadPageSlotComponentIDs(context.Background(), db, siteID, "index", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := slotIDs["generic-text-block"]; got != agreeing {
		t.Errorf("a repeat AGREEING on component_id must map: want %s, got %q", agreeing, got)
	}
	if got, ok := slotIDs["conflicted"]; ok {
		t.Errorf("a repeat DISAGREEING on component_id must map to nothing, got %q", got)
	}
	if _, ok := slotIDs["no-component-id"]; ok {
		t.Error("a row with no component_id must not enter the identity map")
	}
	if len(slotIDs) != 1 {
		t.Errorf("expected exactly one usable slot, got %d: %v", len(slotIDs), slotIDs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}

// TestPlanSections_SlotQueryErrorFailsTheAction: planning against a silently
// empty identity map is exactly the 204 defect — on a decomposed site it files
// junk needs_new_component items for every section. A transient read failure is
// therefore loud, not swallowed.
func TestPlanSections_SlotQueryErrorFailsTheAction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	mock.ExpectQuery("WHERE name IN").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("WHERE function IN").WillReturnRows(emptyComponentRows())
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID, "tool-loan-vs-savings").
		WillReturnError(sql.ErrConnDone)

	_, err = PlanSectionsAction(context.Background(),
		planParams(db, siteID, "tool-loan-vs-savings", []string{"prose-0"}))
	if err == nil {
		t.Fatal("a failed slot-identity read must fail the action, not plan against an empty map")
	}
	if !strings.Contains(err.Error(), "slot") {
		t.Errorf("the error must name the slot-identity read, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet sqlmock expectations: %v", err)
	}
}
