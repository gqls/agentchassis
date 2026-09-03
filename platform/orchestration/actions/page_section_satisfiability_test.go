// FILE: platform/orchestration/actions/page_section_satisfiability_test.go
//
// bugs_open/187. Twenty-eight `needs_page` items were minted for pages
// page-build-handler could resolve no sections for; every one no-oped into
// needs_human_review, and no revalidator covered the type, so they parked for
// ever. These tests pin both ends of the fix: the shared resolver, the emit-side
// guard in each of the two 177-shaped emitters, and the drain.
//
// EXPECTATIONS ARE UNORDERED THROUGHOUT, deliberately. In ordered mode sqlmock
// answers an out-of-sequence call with an ERROR, and every resolver arm here
// treats a source it could not read as a source that declares nothing — so an
// ordered mock can turn a broken read into a passing test by agreeing with it
// (the 177 lane's own two-strike lesson, recorded in tool_content_item_test.go).
//
// Each test that exists to kill a specific mutation says which one in a comment
// above it. The negative assertions — "no work item is written" — are carried by
// the ABSENCE of an insert expectation plus sqlmock's refusal of unexpected
// calls, never by anything the code under test reports about itself.
package actions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ── Shared expectation helpers ───────────────────────────────────────────────

func expectNoDeclaredSections(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT sps.component_name").
		WillReturnRows(sqlmock.NewRows([]string{"component_name"}))
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}))
	mock.ExpectQuery("SELECT sections FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(`[]`)))
}

func expectPlanTableSections(mock sqlmock.Sqlmock, names ...string) {
	rows := sqlmock.NewRows([]string{"component_name"})
	for _, n := range names {
		rows = rows.AddRow(n)
	}
	mock.ExpectQuery("SELECT sps.component_name").WillReturnRows(rows)
}

func expectPlanMembership(mock sqlmock.Sqlmock, member bool) {
	mock.ExpectQuery("FROM site_plan_pages spp").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(member))
}

// expectEscalationOwnershipGuard scripts the owned-page check
// escalateRerenderToWriter runs before minting anything (bugs_open/333): the
// (site_id, name) → id lookup, then the rebuild_policy read. `policy` says
// which way the guard answers. rebuildPolicyReadSQL is the door test's pin of
// the same statement — one literal, both guards.
func expectEscalationOwnershipGuard(mock sqlmock.Sqlmock, policy string) {
	expectEscalationBuildPolicyGuard(mock, policy, false)
}

// expectEscalationBuildPolicyGuard is the same fixture with the tool-shell
// column exposed, so a test can script bugs_open/450's population: a GENERIC
// tool page carrying no tool component.
func expectEscalationBuildPolicyGuard(mock sqlmock.Sqlmock, policy string, toolShell bool) {
	mock.ExpectQuery("SELECT id, url FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(uuid.New(), "/x"))
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WillReturnRows(policyRows(policy, toolShell))
}

// expectNeedsPageInsert covers the statement sequence a needs_page emit
// produces: begin, the two-strike probe (these emitters do NOT set
// recurrenceExpected), the insert, commit. Sixteen arguments, because neither
// caller sets a parent_item_id.
// expectNeedsPageInsertActionRequest is expectNeedsPageInsert for a producer
// that declares recurrenceExpected (bugs_open/326 option E, 2026-08-24):
// flag_page_image_rebuild's emit skips the anti-churn COUNT probe entirely, so
// scripting it would leave an unmet expectation and fail ExpectationsWereMet on
// a query the production code is CORRECT not to issue. The flag itself is pinned
// elsewhere (action_request_producers_recurrence_test.go, ratchet + effect), so
// this helper stays a satisfiability fixture and asserts nothing about the brake.
func expectNeedsPageInsertActionRequest(mock sqlmock.Sqlmock, source string) {
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(),
			source, // $2 source
			sqlmock.AnyArg(),
			"needs_page", // $4 item_type
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"page-build-handler", // $11 handler_agent
			"triaged",            // $12 status
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func expectNeedsPageInsert(mock sqlmock.Sqlmock, source string) {
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(),
			source, // $2 source
			sqlmock.AnyArg(),
			"needs_page", // $4 item_type
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"page-build-handler", // $11 handler_agent
			"triaged",            // $12 status
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

// ── declaredPageSections: the fallback order IS the contract ─────────────────

// All three sources hold DIFFERENT answers — the only input on which a wrong
// order is visible at all. Kills: reordering the three blocks, or dropping the
// `if len(sections) > 0 { return }` after the plan-table rung.
func TestDeclaredPageSections_PlanTablesWinOverBothOthers(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPlanTableSections(mock, "from-plan-tables")
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow([]byte(`{"pages":[{"name":"guides","sections":["from-spec-aspect"]}]}`)))
	mock.ExpectQuery("SELECT sections FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(`["from-pages-table"]`)))

	sections, source := declaredPageSections(context.Background(), db, zap.NewNop(), uuid.New(), "guides")

	if source != "site_plan_tables" {
		t.Fatalf("source = %q, want site_plan_tables", source)
	}
	if len(sections) != 1 || sections[0] != "from-plan-tables" {
		t.Fatalf("sections = %v, want [from-plan-tables]", sections)
	}
}

// The middle rung, with the bottom one holding a different answer. Kills:
// swapping the site_specs and pages.sections blocks.
func TestDeclaredPageSections_SpecAspectWinsOverPagesTable(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT sps.component_name").
		WillReturnRows(sqlmock.NewRows([]string{"component_name"}))
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow([]byte(`{"pages":[{"name":"guides","sections":["from-spec-aspect"]}]}`)))
	mock.ExpectQuery("SELECT sections FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(`["from-pages-table"]`)))

	sections, source := declaredPageSections(context.Background(), db, zap.NewNop(), uuid.New(), "guides")

	if source != "site_specs" {
		t.Fatalf("source = %q, want site_specs", source)
	}
	if len(sections) != 1 || sections[0] != "from-spec-aspect" {
		t.Fatalf("sections = %v, want [from-spec-aspect]", sections)
	}
}

// The bottom rung still has to serve when the two above it are empty — a
// resolver that only ever read the plan tables would pass every test above.
func TestDeclaredPageSections_PagesTableServesWhenNothingAboveIt(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT sps.component_name").
		WillReturnRows(sqlmock.NewRows([]string{"component_name"}))
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}))
	mock.ExpectQuery("SELECT sections FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"sections"}).AddRow([]byte(`["hero","article-body"]`)))

	sections, source := declaredPageSections(context.Background(), db, zap.NewNop(), uuid.New(), "guides")

	if source != "pages_table" {
		t.Fatalf("source = %q, want pages_table", source)
	}
	if len(sections) != 2 {
		t.Fatalf("sections = %v, want two", sections)
	}
}

// The 187 population's own shape: nothing anywhere. "none" is what the guard
// keys on, so it is asserted rather than inferred from the empty list.
func TestDeclaredPageSections_EmptyEverywhere(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectNoDeclaredSections(mock)

	sections, source := declaredPageSections(context.Background(), db, zap.NewNop(), uuid.New(), "directory-index")

	if len(sections) != 0 {
		t.Fatalf("sections = %v, want none", sections)
	}
	if source != "none" {
		t.Fatalf("source = %q, want none", source)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a source went unread: %v", err)
	}
}

// ── pageInCurrentPlan: fail-open is the whole point of the function ──────────

func TestPageInCurrentPlan_MemberAndNonMember(t *testing.T) {
	for _, want := range []bool{true, false} {
		t.Run(fmt.Sprintf("member=%v", want), func(t *testing.T) {
			db, mock := newInsertMock(t)
			defer db.Close()
			mock.MatchExpectationsInOrder(false)

			expectPlanMembership(mock, want)

			if got := pageInCurrentPlan(context.Background(), db, zap.NewNop(), uuid.New(), "guides"); got != want {
				t.Fatalf("pageInCurrentPlan = %v, want %v", got, want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

// A read that FAILED is not evidence of absence. Kills: changing the error arm
// to `return false`, which would turn every transient DB error into a silent
// refusal to emit — the expensive direction of being wrong, and invisible.
func TestPageInCurrentPlan_QueryErrorFailsOpen(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("FROM site_plan_pages spp").WillReturnError(errors.New("connection reset"))

	if !pageInCurrentPlan(context.Background(), db, zap.NewNop(), uuid.New(), "guides") {
		t.Fatal("a failed membership read must fail OPEN (true), or a DB blip silently suppresses emits")
	}
}

// ── pageSectionsSatisfiable: the guard's own logic ───────────────────────────

// A page that declares sections is satisfiable WITHOUT consulting plan
// membership, and the source it reports must be the one that served.
//
// The membership expectation is registered but meant to go UNFULFILLED, so
// ExpectationsWereMet is deliberately not called. Kills: deleting the
// `if len(declared) > 0` early return — the membership query would then run,
// answer false, and the sectioned page would be judged unsatisfiable.
func TestPageSectionsSatisfiable_DeclaredSectionsShortCircuitMembership(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPlanTableSections(mock, "hero", "article-body")
	expectPlanMembership(mock, false)

	ok, sections, source := pageSectionsSatisfiable(context.Background(), db, zap.NewNop(), uuid.New(), "guides")

	if !ok {
		t.Fatal("a page declaring two sections is satisfiable regardless of plan membership")
	}
	if source != "site_plan_tables" {
		t.Fatalf("source = %q, want site_plan_tables — the declaring source, not the membership arm", source)
	}
	if len(sections) != 2 {
		t.Fatalf("sections = %v, want two", sections)
	}
}

// The conservative arm. A plan member declaring nothing is STILL satisfiable,
// because the handler's fallback 4 may synthesise a layout for it.
//
// Kills: dropping the pageInCurrentPlan arm entirely (the page would be judged
// unsatisfiable and a buildable page would lose its item).
func TestPageSectionsSatisfiable_PlanMemberWithNoDeclarations(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectNoDeclaredSections(mock)
	expectPlanMembership(mock, true)

	ok, sections, source := pageSectionsSatisfiable(context.Background(), db, zap.NewNop(), uuid.New(), "brands-index")

	if !ok {
		t.Fatal("a current-plan member is satisfiable: sibling synthesis can serve it")
	}
	if source != "current_plan_member" {
		t.Fatalf("source = %q, want current_plan_member", source)
	}
	if len(sections) != 0 {
		t.Fatalf("no sections are resolved on this arm, got %v", sections)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPageSectionsSatisfiable_SectionlessAndPlanless(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectNoDeclaredSections(mock)
	expectPlanMembership(mock, false)

	ok, _, source := pageSectionsSatisfiable(context.Background(), db, zap.NewNop(), uuid.New(), "directory-index")

	if ok {
		t.Fatal("no sections and no plan membership is the unsatisfiable case — this is the whole of bugs_open/187")
	}
	if source != "none" {
		t.Fatalf("source = %q, want none", source)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ── Emitter 1: flag_page_image_rebuild ───────────────────────────────────────

// Inputs travel in CollectedData, not StepConfig.Config: ExtractActionInputs
// treats a config STRING as a reference to resolve against collected data
// (bugs_open/042 is the family), so a literal site_id placed in config is
// silently unresolved. This is also how the live workflow delivers them.
func flagRebuildParams(db *sql.DB, siteID uuid.UUID, pageName string) ActionParams {
	return ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id":   siteID.String(),
			"scope":     "page",
			"scope_ref": pageName,
		},
	}
}

// The 14 parked rows' shape. NOTHING may be written: no ExpectBegin and no
// insert expectation are registered, so sqlmock fails the call if the guard
// emits anyway.
//
// Kills: deleting the guard, or inverting it to `if satisfiable { skip }`
// (which would fail the positive test below instead).
func TestFlagPageImageRebuild_SectionlessPage_SkipsEmit(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()

	// The needs_rebuild flag half is UNCONDITIONAL and must still run — a page
	// that cannot be re-rendered by an item still wants its stale-build marker.
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 1))
	expectNoDeclaredSections(mock)
	expectPlanMembership(mock, false)

	out, err := FlagPageImageRebuildAction(context.Background(), flagRebuildParams(db, siteID, "directory-index"))
	if err != nil {
		t.Fatalf("the skip must be a clean return, got error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result shape %#v", out)
	}
	if m["needs_page_emit"] != "skipped_sectionless_page" {
		t.Errorf("needs_page_emit = %v, want skipped_sectionless_page — a silent skip is indistinguishable from a guard that never ran", m["needs_page_emit"])
	}
	if m["rebuilt"] != false {
		t.Errorf("rebuilt = %v, want false", m["rebuilt"])
	}
	if m["flagged"] != true {
		t.Errorf("flagged = %v, want true — the needs_rebuild half is not guarded", m["flagged"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (a write may have been attempted): %v", err)
	}
}

// The positive control. Without it, a guard that refused everything would look
// like a working guard.
//
// The membership expectation is registered and meant to go unfulfilled, so
// ExpectationsWereMet is not called; sections_source is asserted instead, which
// is what distinguishes "the declaring source served" from "the membership arm
// waved it through".
func TestFlagPageImageRebuild_SectionedPage_StillEmits(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 1))
	expectPlanTableSections(mock, "hero", "article-body", "call-to-action")
	expectPlanMembership(mock, false)
	expectNeedsPageInsertActionRequest(mock, "image-build-handler")
	// bugs_open/384: the card-derive probe is unscripted here (it reads as
	// skipped_lookup_failed, i.e. no derive raised), so the listings must be
	// told in this same transaction — one consumer page, one reasoned item.
	expectPageListReresolve(mock, "image-build-handler", siteID)

	out, err := FlagPageImageRebuildAction(context.Background(), flagRebuildParams(db, siteID, "board-setup"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := out.(map[string]interface{})
	if m["needs_page_emit"] != "raised" {
		t.Errorf("needs_page_emit = %v, want raised", m["needs_page_emit"])
	}
	if m["rebuilt"] != true {
		t.Errorf("rebuilt = %v, want true", m["rebuilt"])
	}
	if m["sections_source"] != "site_plan_tables" {
		t.Errorf("sections_source = %v, want site_plan_tables — a page that DECLARES sections must not be credited to the membership arm", m["sections_source"])
	}
	if m["page_list_reresolve"] != "queued" {
		t.Errorf("page_list_reresolve = %v, want queued — the listing that renders this page's image was not told (bugs_open/384)", m["page_list_reresolve"])
	}
}

// The conservative edge, end to end: declares nothing, but the plan names it, so
// the handler may synthesise. The item must still be minted.
func TestFlagPageImageRebuild_PlanMemberWithoutSections_StillEmits(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 1))
	expectNoDeclaredSections(mock)
	expectPlanMembership(mock, true)
	expectNeedsPageInsertActionRequest(mock, "image-build-handler")
	expectPageListReresolve(mock, "image-build-handler", siteID) // bugs_open/384

	out, err := FlagPageImageRebuildAction(context.Background(), flagRebuildParams(db, siteID, "brands-index"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, _ := out.(map[string]interface{})
	if m["needs_page_emit"] != "raised" {
		t.Errorf("needs_page_emit = %v, want raised — the guard must not out-guess the handler's sibling synthesis", m["needs_page_emit"])
	}
	if m["page_list_reresolve"] != "queued" {
		t.Errorf("page_list_reresolve = %v, want queued (bugs_open/384)", m["page_list_reresolve"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ── Emitter 2: escalateRerenderToWriter ──────────────────────────────────────

// A tool page's widget slot legitimately carries no content_data, which is what
// triggers this escalation — on a section-less page the trigger is a false
// alarm and the item is unbuildable. Nothing may be written.
//
// Kills: deleting the guard (the insert would be attempted and refused), and
// returning the disposition as "raised" on the skip path.
func TestEscalateRerenderToWriter_SectionlessPage_SkipsEmit(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectNoDeclaredSections(mock)
	expectPlanMembership(mock, false)

	disposition, err := escalateRerenderToWriter(context.Background(), db, uuid.New(), "grip-styles-tool", "a section had no stored content_data", zap.NewNop())
	if err != nil {
		t.Fatalf("the skip must be a clean return, got error: %v", err)
	}
	if disposition != "skipped_sectionless_page" {
		t.Errorf("disposition = %q, want skipped_sectionless_page", disposition)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (a write may have been attempted): %v", err)
	}
}

// The positive control: a real content_data gap on a page with a section plan
// still reaches the writer.
func TestEscalateRerenderToWriter_SectionedPage_StillEmits(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPlanTableSections(mock, "hero", "article-body")
	expectPlanMembership(mock, false)
	expectEscalationOwnershipGuard(mock, "generic")
	expectNeedsPageInsert(mock, "page-rerender")

	disposition, err := escalateRerenderToWriter(context.Background(), db, uuid.New(), "tungsten-guide", "a section had no stored content_data", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disposition != "raised" {
		t.Errorf("disposition = %q, want raised", disposition)
	}
}

func TestEscalateRerenderToWriter_PlanMemberWithoutSections_StillEmits(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectNoDeclaredSections(mock)
	expectPlanMembership(mock, true)
	expectEscalationOwnershipGuard(mock, "generic")
	expectNeedsPageInsert(mock, "page-rerender")

	disposition, err := escalateRerenderToWriter(context.Background(), db, uuid.New(), "brands-index", "a section had no stored content_data", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disposition != "raised" {
		t.Errorf("disposition = %q, want raised", disposition)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// An owned page's empty content_data is its NORMAL state — that is what
// rebuild_policy='owned' means — so the escalation must not mint the needs_page
// its target's ownership guard is certain to refuse (bugs_open/333: 13 such
// items in the door's first 14 hours, every one wont_fix). What it leaves
// instead is the per-page owned_page_review trail, so the suppression is
// auditable if an owned page ever genuinely loses content.
// Kills: deleting the ownership guard (the needs_page emit then opens a
// transaction the mock refuses, surfacing as an escalate error AND unmet guard
// expectations), returning "raised" on the skip path, and dropping the
// owned_page_review emit (unmet INSERT expectation).
func TestEscalateRerenderToWriter_OwnedPage_SkipsEmit(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPlanTableSections(mock, "hero", "tool-widget")
	expectEscalationOwnershipGuard(mock, "owned")
	// The skip writes the per-page owned_page_review trail (round 70a1e557,
	// bug_historian: a suppression only a pod log records is invisible to any
	// later audit). Pinning source and item_key proves it is THAT trail — and
	// dropping the emit leaves this expectation unmet, which fails the test.
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(sqlmock.AnyArg(), "page-rerender", sqlmock.AnyArg(), sqlmock.AnyArg(),
			"owned_page_review:grip-styles-tool").
		WillReturnResult(sqlmock.NewResult(1, 1))

	disposition, err := escalateRerenderToWriter(context.Background(), db, uuid.New(), "grip-styles-tool", "a section had no stored content_data", zap.NewNop())
	if err != nil {
		t.Fatalf("the skip must be a clean return, got error: %v", err)
	}
	if disposition != "skipped_owned_page" {
		t.Errorf("disposition = %q, want skipped_owned_page", disposition)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (a write may have been attempted): %v", err)
	}
}

// A page that does not resolve yet is the legitimate build-request case — the
// whole point of a needs_page — so the guard fails OPEN on ErrNoRows and the
// escalation still emits. Kills: treating a missed lookup as owned, and any
// fail-closed rewrite of the guard.
func TestEscalateRerenderToWriter_PageNotYetBuilt_FailsOpen(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPlanTableSections(mock, "hero")
	mock.ExpectQuery("SELECT id, url FROM pages").WillReturnError(sql.ErrNoRows)
	expectNeedsPageInsert(mock, "page-rerender")

	disposition, err := escalateRerenderToWriter(context.Background(), db, uuid.New(), "not-yet-built", "a section had no stored content_data", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disposition != "raised" {
		t.Errorf("disposition = %q, want raised", disposition)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// An unreadable policy must not suppress a real escalation: the lookup
// resolves, the policy read errors, pageIsOwnedForGuard answers checked=false,
// and the guard stands down. Kills: a fail-closed rewrite (skipping when the
// policy cannot be read — mutation-proven with `owned || !checked`), and
// propagating the probe error. It does NOT kill dropping `checked` from the
// condition: every checked=false path also returns owned=false, so that clause
// is documentation of intent, not a separately observable branch.
func TestEscalateRerenderToWriter_PolicyUnreadable_FailsOpen(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPlanTableSections(mock, "hero")
	mock.ExpectQuery("SELECT id, url FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(uuid.New(), "/x"))
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WillReturnError(errors.New("policy read lost its connection"))
	expectNeedsPageInsert(mock, "page-rerender")

	disposition, err := escalateRerenderToWriter(context.Background(), db, uuid.New(), "tungsten-guide", "a section had no stored content_data", zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if disposition != "raised" {
		t.Errorf("disposition = %q, want raised", disposition)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// ── The drain: revalidateNeedsPage ───────────────────────────────────────────

func needsPageItem(spec map[string]interface{}) parkedReviewItem {
	return parkedReviewItem{
		ID:       uuid.New().String(),
		SiteID:   uuid.New(),
		ItemType: "needs_page",
		ItemKey:  "needs_page:tungsten-guide",
		Spec:     spec,
	}
}

func expectPageRow(mock sqlmock.Sqlmock, pageID, status string) {
	mock.ExpectQuery("SELECT p.id::text").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(pageID, status))
}

func expectSlots(mock sqlmock.Sqlmock, slots ...string) {
	rows := sqlmock.NewRows([]string{"slot_name"})
	for _, s := range slots {
		rows = rows.AddRow(s)
	}
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(rows)
}

// The item_key carries the page name too, and parsing it would be the obvious
// shortcut — but its prefix differs per producer, so it is not read. An item
// whose spec names no page is UNKNOWN, never a close.
//
// Kills: adding an ItemKey fallback that guesses the page name.
func TestRevalidateNeedsPage_NoPageNameInSpec_Unknown(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)
	// No expectations at all: nothing may be queried on a spec that names nothing.

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"reason": "image_landed"}), zap.NewNop())

	if v.Verdict != revalidationUnknown {
		t.Fatalf("verdict = %q, want unknown", v.Verdict)
	}
	if v.Reason == "" {
		t.Error("an unknown verdict must say why — the reason is what a human reads in the queue")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The one arm where "still true" is PROVABLE: the page asked for does not exist.
//
// Kills: collapsing the ErrNoRows arm into the generic error arm (unknown),
// which would leave a genuinely unsatisfied ask indistinguishable from a
// lookup that failed.
func TestRevalidateNeedsPage_PageAbsent_StillHolds(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery("SELECT p.id::text").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}))

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"page_name": "never-built"}), zap.NewNop())

	if v.Verdict != revalidationStillHolds {
		t.Fatalf("verdict = %q, want still_holds — the ask (build this page) is provably unmet", v.Verdict)
	}
}

// Archiving a page is not the same fact as satisfying the ask, and closing here
// would assert the second while only the first happened.
//
// Kills: treating archived as resolved, or as still_holds — either would take
// the human's call away.
func TestRevalidateNeedsPage_ArchivedPage_Unknown(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPageRow(mock, uuid.New().String(), "archived")

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"page_name": "retired-guide"}), zap.NewNop())

	if v.Verdict != revalidationUnknown {
		t.Fatalf("verdict = %q, want unknown", v.Verdict)
	}
	if !strings.Contains(v.Reason, "archived") || !strings.Contains(v.Reason, "human") {
		t.Errorf("reason = %q, want it to name the archive and the human call", v.Reason)
	}
	// Nothing beyond the page row may be read: an archived page's sections are
	// not the question.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// 187's own parked population: the page exists but resolves nothing, so the
// handler would no-op it again. No positive evidence either way.
func TestRevalidateNeedsPage_SectionlessPage_Unknown(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPageRow(mock, uuid.New().String(), "active")
	expectNoDeclaredSections(mock)

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"page_name": "directory-index"}), zap.NewNop())

	if v.Verdict != revalidationUnknown {
		t.Fatalf("verdict = %q, want unknown — a section-less page is ambiguous, never a close", v.Verdict)
	}
	if v.Evidence["sections_source"] != "none" {
		t.Errorf("evidence.sections_source = %v, want none", v.Evidence["sections_source"])
	}
}

// The close, and the audit trail that is its safety case. tungsten-guide's real
// shape: three declared sections, three matching slots (measured live
// 2026-08-03).
//
// Kills: relaxing the match to a COUNT comparison — the missing arm below shares
// this test's slot count.
func TestRevalidateNeedsPage_EveryDeclaredSectionBuilt_Resolved(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	pageID := uuid.New().String()
	expectPageRow(mock, pageID, "active")
	expectPlanTableSections(mock, "hero", "article-body", "call-to-action")
	expectSlots(mock, "hero", "article-body", "call-to-action")

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"page_name": "tungsten-guide"}), zap.NewNop())

	if v.Verdict != revalidationResolved {
		t.Fatalf("verdict = %q (%s), want resolved", v.Verdict, v.Reason)
	}
	if v.Evidence["sections_source"] != "site_plan_tables" {
		t.Errorf("evidence.sections_source = %v, want site_plan_tables", v.Evidence["sections_source"])
	}
	declared, _ := v.Evidence["declared_sections"].([]string)
	if len(declared) != 3 {
		t.Errorf("evidence.declared_sections = %v, want the three declarations the verdict judged", v.Evidence["declared_sections"])
	}
	matched, _ := v.Evidence["matched_slots"].([]string)
	if len(matched) != 3 {
		t.Errorf("evidence.matched_slots = %v, want the three slots that matched — the close's audit trail is its safety case", v.Evidence["matched_slots"])
	}
	if v.Evidence["page_status"] != "active" {
		t.Errorf("evidence.page_status = %v, want active", v.Evidence["page_status"])
	}
}

// Same three slots, but one of them is NOT one of the declared sections. A count
// comparison would close this; exact-name matching leaves it queued and says so
// in the words a human working the queue can act on.
//
// Kills: matching on count, and dropping "satisfiable now" from the reason.
func TestRevalidateNeedsPage_SatisfiableButUnbuilt_StillHolds(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	pageID := uuid.New().String()
	expectPageRow(mock, pageID, "active")
	expectPlanTableSections(mock, "hero", "article-body", "call-to-action")
	expectSlots(mock, "hero", "article-body", "unrelated-widget")

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"page_name": "grip-styles"}), zap.NewNop())

	if v.Verdict != revalidationStillHolds {
		t.Fatalf("verdict = %q, want still_holds", v.Verdict)
	}
	if !strings.Contains(v.Reason, "satisfiable now") {
		t.Errorf("reason = %q, want it to carry 'satisfiable now' — that phrase is what tells the queue's owner this row is real work", v.Reason)
	}
	missing, _ := v.Evidence["missing_sections"].([]string)
	if len(missing) != 1 || missing[0] != "call-to-action" {
		t.Errorf("evidence.missing_sections = %v, want [call-to-action]", v.Evidence["missing_sections"])
	}
}

// 31 live sections entries are written in the underscore dialect while no
// slot_name is (measured 2026-08-03), so both sides go through
// NormalizeComponentFunction. Without it these 31 would never resolve.
//
// Kills: comparing the raw strings.
func TestRevalidateNeedsPage_UnderscoreDialectStillMatches(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPageRow(mock, uuid.New().String(), "active")
	expectPlanTableSections(mock, "call_to_action")
	expectSlots(mock, "call-to-action")

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"page_name": "careers"}), zap.NewNop())

	if v.Verdict != revalidationResolved {
		t.Fatalf("verdict = %q (%s), want resolved — call_to_action and call-to-action are the same section", v.Verdict, v.Reason)
	}
}

// A lookup that failed is not a finding. Kills: letting a component-query error
// fall through to the resolved arm with an empty slot set.
func TestRevalidateNeedsPage_ComponentLookupFails_Unknown(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPageRow(mock, uuid.New().String(), "active")
	expectPlanTableSections(mock, "hero")
	mock.ExpectQuery("FROM page_components pc").WillReturnError(errors.New("connection reset"))

	v := revalidateNeedsPage(context.Background(), db,
		needsPageItem(map[string]interface{}{"page_name": "tungsten-guide"}), zap.NewNop())

	if v.Verdict != revalidationUnknown {
		t.Fatalf("verdict = %q, want unknown", v.Verdict)
	}
}
