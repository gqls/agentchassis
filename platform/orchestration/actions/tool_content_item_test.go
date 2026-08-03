// FILE: platform/orchestration/actions/tool_content_item_test.go
//
// bugs_open/177. Every `tool_content:%` work item ever minted (9 of 9, across
// four sites and three weeks) died in needs_human_review, because the page it
// named declared no sections and page-build-handler had nothing to build. These
// tests pin the guard that decides whether the item is worth minting at all.
//
// The load-bearing case is NoSections: the assertion there is a NEGATIVE — that
// no transaction is begun and no row is written — and it is carried by
// mock.ExpectationsWereMet plus sqlmock's refusal of unexpected calls, not by
// counting anything the helper reports about itself. A helper's own bookkeeping
// cannot prove it did not write.
//
// The arm that must NOT change is DeployPathShape: the deployer declares three
// prose sections around the widget, so the guard resolves them and raises the
// item exactly as the hand-rolled INSERT did.
package actions

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func toolContentTestRequest() toolContentItemRequest {
	return toolContentItemRequest{
		siteID:       uuid.New(),
		pageID:       uuid.New(),
		pageName:     "tool-x",
		toolFunction: "tool-x",
		displayName:  "Tool X",
		description:  "Works out X",
		pageURL:      "/tools/tool-x/index.html",
		source:       "tool-generator",
	}
}

// expectNoPlanTableSections / expectNoSpecAspectSections stand in for the two
// higher-priority sources missing, which is the state a freshly minted tool page
// is always in. They are separate helpers because the ORDER matters: sqlmock
// matches in sequence, so a guard that consulted pages.sections first would fail
// here rather than silently agreeing by coincidence.
func expectNoPlanTableSections(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT sps.component_name").
		WillReturnRows(sqlmock.NewRows([]string{"component_name"}))
}

func expectNoSpecAspectSections(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}))
}

func expectPagesSections(mock sqlmock.Sqlmock, sectionsJSON string) {
	mock.ExpectQuery("SELECT sections FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(sectionsJSON)))
}

// expectToolContentInsert pins the fields that make the row dispatchable and
// dedupable, over the statement sequence recurrenceExpected produces: begin,
// insert, commit, with no two-strike probe between them.
//
// IT DOES NOT PROVE THE FLAG IS SET, and an earlier draft of this file claimed
// it did. Mutating `recurrenceExpected` to false left every test here green: in
// ordered mode sqlmock answers the unexpected `SELECT COUNT(*)` with an error,
// and writeWorkItem's own `if err == nil && terminalCount > 0` swallows it, so
// the probe is issued, refused, ignored, and the insert lands unchanged. The
// flag has its own test below, which asserts it by EFFECT.
//
// The argument count is load-bearing (sqlmock fails on a length mismatch), which
// is why there are sixteen: parent_item_id is appended as $17 only by callers
// that set one, and this caller does not.
func expectToolContentInsert(mock sqlmock.Sqlmock, rowsAffected int64) {
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"needs_content_page", // $4 item_type
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"page-build-handler", // $11 handler_agent
			"triaged",            // $12 status
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, rowsAffected))
	mock.ExpectCommit()
}

// The generator path as it stands today: the page row carries `[]` and the page
// is in no plan. Nothing may be written — this is the whole of bugs_open/177.
func TestRaiseToolContentItem_NoSectionsAnywhere_Skips(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectNoPlanTableSections(mock)
	expectNoSpecAspectSections(mock)
	expectPagesSections(mock, `[]`)
	// Deliberately no ExpectBegin and no ExpectExec: if the guard opens a
	// transaction or writes a row, sqlmock fails on the unexpected call.

	got := raiseToolContentItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), toolContentTestRequest())

	if got != "skipped_no_prose_sections" {
		t.Errorf("disposition = %q, want %q", got, "skipped_no_prose_sections")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A generated tool page that IS in the current plan is satisfiable, and must
// still get its item. 33 current-plan section rows for tool-named pages existed
// when this was written, so the edge is real rather than theoretical: a guard
// that trusted the creating action's own (empty) declaration would refuse these.
func TestRaiseToolContentItem_PlanTableSections_Raises(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.ExpectQuery("SELECT sps.component_name").
		WillReturnRows(sqlmock.NewRows([]string{"component_name"}).
			AddRow("hero").
			AddRow("generic-text-block"))
	// No lower-priority lookup is registered because none should happen. That is
	// a weak assertion on its own — sqlmock refuses an unexpected Query, and the
	// resolver treats a refused source as an empty one — so the priority contract
	// is pinned properly by TestToolPageDeclaredSections_PriorityOrder below.
	expectToolContentInsert(mock, 1)

	got := raiseToolContentItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), toolContentTestRequest())

	if got != "raised" {
		t.Errorf("disposition = %q, want %q", got, "raised")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The deployer's shape, reached through the LAST fallback: the page declares the
// widget plus three prose sections. This arm's behaviour must not change, and
// the three sections that are not the tool function are what keeps it raising.
func TestRaiseToolContentItem_DeployPathShape_Raises(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectNoPlanTableSections(mock)
	expectNoSpecAspectSections(mock)
	expectPagesSections(mock, `["hero-tool", "tool-guide-intro", "tool-x", "tool-cta"]`)
	expectToolContentInsert(mock, 1)

	req := toolContentTestRequest()
	req.source = "tool-deployer"

	got := raiseToolContentItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), req)

	if got != "raised" {
		t.Errorf("disposition = %q, want %q", got, "raised")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The widget alone is not prose. Raising here would be worse than raising on an
// empty page: the item would be SATISFIABLE, and the only section the writer
// could touch is the tool itself (TL-001's clobber surface).
func TestRaiseToolContentItem_WidgetIsTheOnlySection_Skips(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectNoPlanTableSections(mock)
	expectNoSpecAspectSections(mock)
	expectPagesSections(mock, `["tool-x"]`)

	got := raiseToolContentItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), toolContentTestRequest())

	if got != "skipped_no_prose_sections" {
		t.Errorf("disposition = %q, want %q", got, "skipped_no_prose_sections")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The priority contract, with all three sources holding DIFFERENT answers — the
// only input on which a wrong order is visible. The guard must ask the question
// page-build-handler will ask, in the order it will ask it
// (load_page_sections_from_spec_action.go fallbacks 1 to 3), or it starts
// refusing items the handler could have built.
//
// Unordered expectations, and ExpectationsWereMet is not called: two of the
// three are meant to go unfulfilled, and which two is the assertion.
func TestToolPageDeclaredSections_PriorityOrder(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT sps.component_name").
		WillReturnRows(sqlmock.NewRows([]string{"component_name"}).AddRow("from-plan-tables"))
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow([]byte(`{"pages":[{"name":"tool-x","sections":["from-spec-aspect"]}]}`)))
	mock.ExpectQuery("SELECT sections FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(`["from-pages-table"]`)))

	sections, source := toolPageDeclaredSections(context.Background(), db, zap.NewNop(), uuid.New(), "tool-x")

	if source != "site_plan_tables" {
		t.Errorf("source = %q, want %q", source, "site_plan_tables")
	}
	if len(sections) != 1 || sections[0] != "from-plan-tables" {
		t.Errorf("sections = %v, want [from-plan-tables]", sections)
	}
}

// The aspect is the middle rung, and the only one whose answer has to be dug out
// of a JSON document by page name. Five sites are still served from it.
func TestToolPageDeclaredSections_SpecAspectBeatsPagesTable(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.MatchExpectationsInOrder(false)

	expectNoPlanTableSections(mock)
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow([]byte(`{"pages":[{"name":"other","sections":["wrong-page"]},{"name":"tool-x","sections":["from-spec-aspect"]}]}`)))
	mock.ExpectQuery("SELECT sections FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(`["from-pages-table"]`)))

	sections, source := toolPageDeclaredSections(context.Background(), db, zap.NewNop(), uuid.New(), "tool-x")

	if source != "site_specs" {
		t.Errorf("source = %q, want %q", source, "site_specs")
	}
	if len(sections) != 1 || sections[0] != "from-spec-aspect" {
		t.Errorf("sections = %v, want [from-spec-aspect]", sections)
	}
}

// The flag, asserted by its effect rather than by the statement sequence (see
// expectToolContentInsert for why the sequence cannot carry it).
//
// A tool re-deployed onto a site that has already had two SUCCESSFUL content
// builds under this key is the shape bugs_open/024 recorded: the two-strike rule
// reads two terminal predecessors as two failed attempts and brands the new row
// `unresolved`, which is terminal and never dispatched. This item is an ACTION
// REQUEST, so both predecessors are successes and the third request is ordinary
// business.
//
// Expectations are UNORDERED here and ExpectationsWereMet is deliberately not
// called: the probe expectation is meant to go unfulfilled, and that is the
// assertion. If recurrenceExpected were dropped, the probe would be issued and
// answered with two terminal rows, the row would arrive branded `unresolved`
// with a rewritten summary, the insert expectation below would not match, and
// the helper would report insert_failed. Verified by mutation, 2026-08-03.
func TestRaiseToolContentItem_RecurrenceExpected_SurvivesTerminalPredecessors(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT sps.component_name").
		WillReturnRows(sqlmock.NewRows([]string{"component_name"}).AddRow("hero"))
	// Two terminal predecessors, the newest 5h old — past the within-cycle
	// window, so the branding arm is the one that would fire.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(2, 5.0))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"Write content for tool page: Tool X", // $6 summary, unbranded
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"triaged", // $12 status, dispatchable
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	got := raiseToolContentItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), toolContentTestRequest())

	if got != "raised" {
		t.Errorf("disposition = %q, want %q — a third content request must not be suppressed or branded", got, "raised")
	}
}

// An OPEN item already holds the key, so idx_swi_dedup's partial predicate takes
// the insert to nothing. Not a failure and not a raise: the caller's output map
// must be able to tell the two apart, since "already asked for" and "declined to
// ask" are different facts about the same page.
func TestRaiseToolContentItem_OpenItemHoldsKey_Deduped(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectNoPlanTableSections(mock)
	expectNoSpecAspectSections(mock)
	expectPagesSections(mock, `["hero-tool", "tool-guide-intro", "tool-x", "tool-cta"]`)
	expectToolContentInsert(mock, 0)

	got := raiseToolContentItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), toolContentTestRequest())

	if got != "deduped_open_item" {
		t.Errorf("disposition = %q, want %q", got, "deduped_open_item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
