// FILE: platform/orchestration/actions/work_item_owned_page_door_test.go
//
// The policy door in writeWorkItem (bugs_open/333): a content finding whose page
// is rebuild_policy='owned', filed at a handler that DECLARES it refuses owned
// pages, is parked at 'deferred' instead of being routed into a guaranteed
// refusal.
//
// THE BUG THIS PINS. ~26 producers hard-code handler_agent='page-build-handler'
// for content findings and none reads pages.rebuild_policy. On an owned page the
// handler refuses, the item terminates `wont_fix` — the status meaning "we
// decided not to fix this" — idx_swi_dedup excludes `wont_fix` so the detector
// re-files, and the only human-visible trail (owned_page_review) dedups per PAGE,
// so the second distinct defect on a reviewed page leaves no trace at all.
// 83 findings ended that way between 2026-08-19 and 2026-08-24 alone.
//
// PROBE ORDER: policy read (a `pages` PK lookup) FIRST, declaration probe (the
// jsonb_path_exists over agent_definitions) SECOND and only for an owned page —
// swapped from the reverse after the council's guardian seat objected that
// running the novel SQL on every write through a shared seam injects a new
// systemic failure mode into a hot path. Tests scripting only ONE expectation
// are therefore the generic-page cases, not oversights.
//
// HOW THESE TESTS ARE BUILT, and why it matters more than usual here. Both probe
// expectations are rendered FROM the shared renderers
// (workItemHandlerRefusesOwnedPagesSQL / readRebuildPolicy's statement), the same
// coupling proof the 291 guard's tests use: hand-write either statement in
// writeWorkItem and the expectation stops matching.
//
// ⚠ AND THE TRAP THAT MAKES A NAIVE VERSION OF THIS FILE WORTHLESS. The door
// FAILS OPEN: a probe error is logged and swallowed so the finding is never lost
// to a pod log. sqlmock reports an unscripted query as an error ON THAT QUERY —
// which the fall-through then swallows — so a test that simply omits the probe
// expectations does not fail. It passes, having proved nothing, and so does the
// mutation. Every negative below therefore either scripts the probes and calls
// ExpectationsWereMet, or uses the TRIPWIRE idiom: script a probe that would
// demote if consumed, and assert the outcome is undemoted. The 291 guard's file
// records learning this the hard way ("the widened-set mutation PASSED that
// version of this test"); this is the same lesson, one door along.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"go.uber.org/zap"
)

// rebuildPolicyReadSQL is the statement readGenericBuildPolicy runs. It is
// written here ONCE, as the expectation source for every test in this file, and
// it is RENDERED FROM the shared predicate rather than transcribed — hand-write
// the shell fragment anywhere else and this expectation stops matching, which is
// the coupling proof the file's header describes.
//
// It reads TWO columns since bugs_open/450: the policy, and whether the page is
// a tool with no tool component. Both fixture helpers below therefore return two
// columns, and a one-column fixture will not compile.
var rebuildPolicyReadSQL = `
		SELECT COALESCE(pages.rebuild_policy, 'generic'),
		       ` + toolShellPredicateFor("pages") + `
		FROM pages WHERE pages.id = $1
	`

// policyRows builds the door's first-question result: (rebuild_policy, tool_shell).
func policyRows(policy string, toolShell bool) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"rebuild_policy", "tool_shell"}).AddRow(policy, toolShell)
}

// expectOwnedPageDeclarationProbe scripts the door's SECOND question, reached only
// for a page the generic builder may not write: does this handler declare that it
// refuses such pages?
func expectOwnedPageDeclarationProbe(mock sqlmock.Sqlmock, handler string, declares bool) {
	mock.ExpectQuery(regexp.QuoteMeta("SELECT " + workItemHandlerRefusesOwnedPagesSQL("$1"))).
		WithArgs(handler).
		WillReturnRows(sqlmock.NewRows([]string{"declares"}).AddRow(declares))
}

// expectRebuildPolicyRead scripts the door's FIRST question, run for every
// page-bearing write: is this page owned? (Not a tool shell — see
// expectToolShellPolicyRead for that arm.)
func expectRebuildPolicyRead(mock sqlmock.Sqlmock, pageID uuid.UUID, policy string) {
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WithArgs(pageID).
		WillReturnRows(policyRows(policy, false))
}

// expectToolShellPolicyRead scripts a GENERIC page that is a tool with no tool
// component — bugs_open/450's population, and the one the old door waved through.
func expectToolShellPolicyRead(mock sqlmock.Sqlmock, pageID uuid.UUID) {
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WithArgs(pageID).
		WillReturnRows(policyRows("generic", true))
}

// expectRebuildPolicyReadError scripts a failed policy read.
func expectRebuildPolicyReadError(mock sqlmock.Sqlmock, pageID uuid.UUID, err error) {
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WithArgs(pageID).
		WillReturnError(err)
}

// ── Helpers for tests that are NOT about the door but now pass through it ─────
//
// Any producer test whose item carries a page id and a handler at a dispatchable
// status reaches the door. If it does not script the probes, sqlmock reports an
// unexpected query, THE DOOR'S OWN FALL-THROUGH SWALLOWS THAT ERROR, and the
// test goes green having silently stopped exercising the statement sequence it
// was written to pin.
//
// That is not a hypothetical: instrumenting the door to announce an unscripted
// probe found 21 such cases across 8 files the day this landed. Both helpers
// below exist so those tests keep proving what they claim, and so the next
// producer test has an obvious thing to call.
//
// Neither uses WithArgs: these tests are not about WHICH handler or page, only
// that the door was consulted and stood down.

// expectWorkItemDoorStandsDown scripts a GENERIC page — the common case, and the
// point at which the door stops for the overwhelming majority of writes. One
// primary-key read and nothing else: the declaration probe (the novel SQL) is
// never reached, which is the whole reason the policy read goes first.
func expectWorkItemDoorStandsDown(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WillReturnRows(policyRows("generic", false))
}

// expectWorkItemDoorGenericPage is expectWorkItemDoorStandsDown by another name,
// kept as a separate call so a test says WHICH fact makes the door stand down.
// Since the policy read gates everything, a generic page needs one expectation
// whatever its handler declares.
func expectWorkItemDoorGenericPage(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WillReturnRows(policyRows("generic", false))
}

// doorItem is a content finding of the shape the producers actually file:
// a page id, a real item_key, and page-build-handler.
//
// recurrenceExpected is TRUE so the two-strike block cannot interfere with what
// these tests are measuring — the same isolation the 291 guard's fixture uses.
// The parked path sets the flag itself, so this changes nothing about it.
func doorItem(status, handler string, pageID *uuid.UUID) workItem {
	return workItem{
		siteID:             uuid.New(),
		source:             "discovery",
		pipeline:           "content",
		itemType:           "content_rewrite",
		severity:           "high",
		summary:            "Literal markdown is visible in the hero copy",
		spec:               `{"page_name":"tool-mortgage-calculator","detail":"**bold** rendered raw"}`,
		pageID:             pageID,
		priority:           50,
		handlerAgent:       handler,
		status:             status,
		createdBy:          "quality-discovery-agent",
		itemKey:            "content_rewrite:tool-mortgage-calculator",
		recurrenceExpected: true,
	}
}

// expectParkedInsert asserts the row lands at 'deferred' ($12) with NO handler
// ($11) and the conditional 17th error argument — the same conditional-append
// idiom 291 uses, so callers that never trip the door keep their sixteen args.
// $4 (item_type) and $14 (item_key) are asserted UNCHANGED, which is the whole
// retraction argument: a re-typed row could never be closed by its own detector.
func expectParkedInsert(mock sqlmock.Sqlmock, itemType, itemKey, errText string) {
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			itemType, // $4  — NOT re-typed
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"",                    // $11 — handler cleared
			ownedPageParkedStatus, // $12
			sqlmock.AnyArg(),
			itemKey, // $14 — NOT re-keyed
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			errText, // $17 — only present because the door fired
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// ── The bug's exact shape, at both statuses that reach a handler ──────────────
//
// Mutation proof: delete the door block in writeWorkItem and both fail on $11/$12
// (the row is written at its requested status with page-build-handler intact).

func TestWriteWorkItem_OwnedPage_ParkedAtDeferred(t *testing.T) {
	for _, status := range []string{"detected", "triaged"} {
		t.Run(status, func(t *testing.T) {
			db, mock := newInsertMock(t)
			defer db.Close()

			pageID := uuid.New()
			item := doorItem(status, "page-build-handler", &pageID)

			mock.ExpectBegin()
			expectRebuildPolicyRead(mock, pageID, "owned")
			expectOwnedPageDeclarationProbe(mock, "page-build-handler", true)
			expectParkedInsert(mock, "content_rewrite", "content_rewrite:tool-mortgage-calculator",
				ownedPageSkipReasonPrefix+": page-build-handler declares refuse_owned_page and page "+
					pageID.String()+" is rebuild_policy=owned — content_rewrite finding parked at "+
					"deferred, not dispatched (bugs_open/333)")

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}

			w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
			if err != nil {
				t.Fatalf("writeWorkItem: %v", err)
			}
			if !w.Inserted {
				t.Fatal("a parked item must still be INSERTED — losing the finding is the failure mode " +
					"the park exists to avoid (the finding is real; only its route is missing)")
			}
			if !w.OwnedPageParked {
				t.Fatal("OwnedPageParked must report the park to the caller, or a producer counting " +
					"its own successes reports a filing that will never be dispatched (bugs_open/177's shape)")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations: %v", err)
			}
		})
	}
}

// ── The discriminating negative, and the reason the door reads a DECLARATION ──
//
// page-rerender completes 5,040 items on owned pages (measured 2026-08-24, live
// + archive). If the door demoted by "the page is owned" alone it would park the
// estate's single most productive owned-page route. It must ask the handler.
func TestWriteWorkItem_OwnedPage_NonDeclaringHandler_Untouched(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("triaged", "page-rerender", &pageID)

	mock.ExpectBegin()
	expectRebuildPolicyRead(mock, pageID, "owned")
	expectOwnedPageDeclarationProbe(mock, "page-rerender", false)
	expectHandlerRegisteredProbe(mock, "page-rerender", true)
	expectInsertWithSummaryAndStatus(mock, item.summary, "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.OwnedPageParked {
		t.Fatal("a handler that does not declare refuse_owned_page must be left alone — " +
			"page-rerender completes thousands of items on owned pages")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// A generic page at the declaring handler is the ordinary case and must be
// byte-identical to before: sixteen args, no error column, requested status.
func TestWriteWorkItem_GenericPage_DeclaringHandler_Untouched(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("triaged", "page-build-handler", &pageID)

	mock.ExpectBegin()
	expectRebuildPolicyRead(mock, pageID, "generic")
	expectHandlerRegisteredProbe(mock, "page-build-handler", true)
	expectInsertWithSummaryAndStatus(mock, item.summary, "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.OwnedPageParked {
		t.Fatal("a generic page must route normally — this is the path that repairs most of the estate")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// A page_id that no longer resolves is ORDINARY at this seam: config-driven
// create_work_item takes whatever input_data.spec.page_id says. A page that does
// not exist cannot be owned, so the row routes normally rather than being parked
// on the strength of a stale id.
func TestWriteWorkItem_PageDoesNotResolve_RoutesNormally(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("triaged", "page-build-handler", &pageID)

	mock.ExpectBegin()
	expectRebuildPolicyReadError(mock, pageID, sql.ErrNoRows)
	expectHandlerRegisteredProbe(mock, "page-build-handler", true)
	expectInsertWithSummaryAndStatus(mock, item.summary, "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.OwnedPageParked {
		t.Fatal("ErrNoRows means the page is gone, not that it is owned")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// A policy read that fails for a REAL reason falls through to the handler's own
// refusal (bugs_closed/301), which stays the backstop — the same posture 291's
// probe takes towards the claim path.
func TestWriteWorkItem_PolicyReadFails_FallsThroughToHandlerRefusal(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("triaged", "page-build-handler", &pageID)

	mock.ExpectBegin()
	expectRebuildPolicyReadError(mock, pageID, errors.New("connection reset"))
	expectHandlerRegisteredProbe(mock, "page-build-handler", true)
	expectInsertWithSummaryAndStatus(mock, item.summary, "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("a probe failure must not fail the write — the finding would be lost to a pod log: %v", err)
	}
	if w.OwnedPageParked {
		t.Fatal("an unreadable policy is not an owned page")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// ── Shapes the door must NEVER probe (tripwire idiom) ─────────────────────────
//
// Each of these scripts a declaration probe that WOULD park the row if consumed,
// and asserts it was not. ExpectationsWereMet is deliberately NOT called: the
// probe is meant to go unconsumed. A bare "no expectations" version of this test
// cannot catch the mutation, because the door's own fall-through swallows
// sqlmock's unexpected-query error and inserts normally.
func TestWriteWorkItem_ShapesTheDoorMustNotTouch(t *testing.T) {
	pageID := uuid.New()

	cases := []struct {
		name string
		item workItem
	}{
		{"no page id — nothing to ask about", doorItem("triaged", "page-build-handler", nil)},
		{"no handler — the flag-only idiom (bugs_closed/284)", doorItem("detected", "", &pageID)},
		{"already parked at needs_human_review", doorItem("needs_human_review", "page-build-handler", &pageID)},
		{"already parked at deferred", doorItem(ownedPageParkedStatus, "page-build-handler", &pageID)},
		{"already blocked (291's demotion)", doorItem("blocked", "page-build-handler", &pageID)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newInsertMock(t)
			defer db.Close()
			mock.MatchExpectationsInOrder(false)

			mock.ExpectBegin()
			expectRebuildPolicyRead(mock, pageID, "owned")
			expectOwnedPageDeclarationProbe(mock, tc.item.handlerAgent, true)
			mock.ExpectExec("INSERT INTO site_work_items").
				WillReturnResult(sqlmock.NewResult(1, 1))

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			w, err := writeWorkItem(context.Background(), tx, tc.item, dropOnConflict, zap.NewNop())
			if err != nil {
				t.Fatalf("writeWorkItem: %v", err)
			}
			if w.OwnedPageParked {
				t.Fatalf("the door must not touch %s — parking a row that is already parked, or one "+
					"deliberately filed without a handler, is bugs_closed/284's regression", tc.name)
			}
		})
	}
}

// The kill switch disarms the door fleet-wide without a redeploy, falling back to
// exactly the pre-guard behaviour. Ships ARMED; this proves the lever works.
func TestWriteWorkItem_KillSwitch_DisarmsThePolicyDoor(t *testing.T) {
	t.Setenv("DISABLE_OWNED_PAGE_DOOR_DEMOTION", "1")

	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	pageID := uuid.New()
	item := doorItem("triaged", "page-build-handler", &pageID)

	mock.ExpectBegin()
	expectRebuildPolicyRead(mock, pageID, "owned")                    // tripwire
	expectOwnedPageDeclarationProbe(mock, "page-build-handler", true) // tripwire
	expectHandlerRegisteredProbe(mock, "page-build-handler", true)
	expectInsertWithSummaryAndStatus(mock, item.summary, "triaged")

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.OwnedPageParked {
		t.Fatal("the kill switch must disarm the door completely")
	}
}

// A parked row whose key is already held reports the park even though nothing was
// inserted. The two answer different questions, and a producer that conflates
// them says "raised at page-build-handler" about a finding that was not.
func TestWriteWorkItem_OwnedPage_DedupedRowStillReportsThePark(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("triaged", "page-build-handler", &pageID)

	mock.ExpectBegin()
	expectRebuildPolicyRead(mock, pageID, "owned")
	expectOwnedPageDeclarationProbe(mock, "page-build-handler", true)
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 0)) // ON CONFLICT DO NOTHING

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	w, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop())
	if err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if w.Inserted {
		t.Fatal("nothing was inserted — Inserted must stay honest")
	}
	if !w.OwnedPageParked {
		t.Fatal("the finding was still parked rather than routed; a producer must be able to see that")
	}
}

// ── The pure transform, checkable without a database ─────────────────────────

func TestOwnedPageParkedItem_KeepsIdentityTakesTheSignal(t *testing.T) {
	pageID := uuid.New()
	orig := doorItem("triaged", "page-build-handler", &pageID)
	orig.recurrenceExpected = false

	parked, errText := ownedPageParkedItem(orig, refusalOwned)

	// THE RETRACTION CONTRACT. resolveWorkItems closes by (item_type, item_key)
	// and `deferred` is not a closed status, so a row that keeps its identity is
	// retracted normally when the finding stops reproducing. Re-type or re-key it
	// and nothing can ever close it: it holds its dedup slot for ever, so the
	// detector cannot re-file either. This assertion IS that argument.
	if parked.itemType != orig.itemType {
		t.Errorf("item_type must survive the park (retraction matches on it): got %q want %q",
			parked.itemType, orig.itemType)
	}
	if parked.itemKey != orig.itemKey {
		t.Errorf("item_key must survive the park (retraction matches on it): got %q want %q",
			parked.itemKey, orig.itemKey)
	}
	if parked.severity != orig.severity {
		t.Errorf("severity is the finding's, not the park's: got %q want %q", parked.severity, orig.severity)
	}

	if parked.status != ownedPageParkedStatus {
		t.Errorf("status: got %q want %q", parked.status, ownedPageParkedStatus)
	}
	if parked.handlerAgent != "" {
		t.Errorf("handler must be cleared, not pointed elsewhere: got %q", parked.handlerAgent)
	}
	if parked.priority != ownedPageParkedPriority {
		t.Errorf("priority: got %d want %d", parked.priority, ownedPageParkedPriority)
	}
	// Without this, two retract/re-detect cycles inside 7 days brand the third
	// finding "[unresolved after 2 attempts]" — 333's own loop under a new label.
	if !parked.recurrenceExpected {
		t.Error("recurrenceExpected must be set, or the two-strike rule counts retractions as strikes")
	}
	if !strings.HasPrefix(parked.summary, ownedPageParkedPrefix) {
		t.Errorf("summary must lead with the marker: %q", parked.summary)
	}
	if !strings.HasPrefix(errText, ownedPageSkipReasonPrefix+":") {
		t.Errorf("error must LEAD with %q so existing ownership-refusal readers find it: %q",
			ownedPageSkipReasonPrefix, errText)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(parked.spec), &spec); err != nil {
		t.Fatalf("parked spec is not valid JSON: %v", err)
	}
	// The producer's own spec survives at the top level: a human reading the row
	// wants the finding, not a wrapper.
	if spec["page_name"] != "tool-mortgage-calculator" {
		t.Errorf("the producer's spec must be preserved: %v", spec)
	}
	if spec["gap_kind"] != ownedPageParkedGapKind {
		t.Errorf("gap_kind: got %v", spec["gap_kind"])
	}
	// The roadmap sweep GROUPS by builder_needed, so it must be stable per
	// handler — a per-finding string would produce N buckets of one.
	builder, _ := spec["builder_needed"].(string)
	if !strings.Contains(builder, "page-build-handler") || strings.Contains(builder, "tool-mortgage-calculator") {
		t.Errorf("builder_needed must name the HANDLER and not the finding, so the roadmap groups: %q", builder)
	}
	if _, ok := spec["what_to_do"]; !ok {
		t.Error("a parked row must say what DOES work on an owned page")
	}
	guard, ok := spec["owned_page_guard"].(map[string]interface{})
	if !ok || guard["refused_handler"] != "page-build-handler" || guard["requested_status"] != "triaged" {
		t.Errorf("owned_page_guard must record what was refused and what was asked for: %v", spec["owned_page_guard"])
	}
}

// The advice splits on whether the finding needs a section REWRITTEN or ADDED,
// because apply_section_edit is the measured route for the first and a documented
// dead end for the second (bugs_closed/295's residual). Sending a human to the
// wrong one is the failure this arm exists to avoid.
func TestOwnedPageParkedAdvice_SplitsRewriteFromAdd(t *testing.T) {
	rewrite := ownedPageParkedAdvice("content_rewrite")
	if !strings.Contains(rewrite, "section-editor") {
		t.Errorf("a rewrite finding must name the route that works: %q", rewrite)
	}
	add := ownedPageParkedAdvice("needs_content_page")
	if strings.Contains(add, "44 complete") {
		t.Errorf("an ADD finding must not be sent to apply_section_edit, which cannot add a section: %q", add)
	}
	if !strings.Contains(add, "tool pipeline") {
		t.Errorf("an ADD finding must name what can actually do it: %q", add)
	}
}

// A keyless item would park unbounded — idx_swi_dedup's predicate is
// `item_key IS NOT NULL`, so nothing collapses repeats. The synthesised key gives
// its parked form a dedup slot the original never had; that is a deliberate
// narrowing and is pinned so it stays deliberate.
func TestOwnedPageParkedItem_KeylessItemGetsABoundedKey(t *testing.T) {
	pageID := uuid.New()
	orig := doorItem("triaged", "page-build-handler", &pageID)
	orig.itemKey = ""

	parked, _ := ownedPageParkedItem(orig, refusalOwned)

	want := "content_rewrite:owned_page:" + pageID.String()
	if parked.itemKey != want {
		t.Errorf("keyless item must get a bounded key: got %q want %q", parked.itemKey, want)
	}
	// workItemKey's contract: the prefix equals the item_type.
	if !strings.HasPrefix(parked.itemKey, parked.itemType+":") {
		t.Errorf("synthesised key must keep the {itemType}:{target} contract: %q", parked.itemKey)
	}
}

// An unparsable spec is kept verbatim rather than dropped: the finding's detail
// is the thing a human needs, and losing it to a marshalling error would repeat
// the defect this whole change is about.
func TestOwnedPageParkedItem_UnparsableSpecIsKept(t *testing.T) {
	pageID := uuid.New()
	orig := doorItem("triaged", "page-build-handler", &pageID)
	orig.spec = "not json at all"

	parked, _ := ownedPageParkedItem(orig, refusalOwned)

	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(parked.spec), &spec); err != nil {
		t.Fatalf("the parked spec must still be valid JSON: %v", err)
	}
	if spec["spec_raw"] != "not json at all" {
		t.Errorf("an unparsable spec must be preserved verbatim: %v", spec)
	}
}

// A very long summary is truncated on the ORIGINAL, then prefixed — truncating
// the concatenation would eat the marker that says why the row is parked.
func TestOwnedPageParkedItem_TruncatesWithoutEatingTheMarker(t *testing.T) {
	pageID := uuid.New()
	orig := doorItem("triaged", "page-build-handler", &pageID)
	orig.summary = strings.Repeat("x", 400)

	parked, _ := ownedPageParkedItem(orig, refusalOwned)

	if !strings.HasPrefix(parked.summary, ownedPageParkedPrefix) {
		t.Error("the marker must survive truncation")
	}
	if len(parked.summary) > workItemSummaryMaxLen {
		t.Errorf("summary must fit the column: %d chars", len(parked.summary))
	}
}

// ── The status set, pinned against the bugs_closed/284 regression ─────────────
//
// `detected` is IN this set and OUT of 291's, and the asymmetry is load-bearing:
// registration is re-judged at promotion, ownership is re-judged by nobody. This
// test is what stops someone "tidying" the two into one predicate.
func TestStatusHeadsForDispatch_IsRegistrationsSetPlusDetected(t *testing.T) {
	heads := map[string]bool{
		"detected": true, "triaged": true, "approved": true, "claimed": true,
		"needs_human_review": false, "deferred": false, "blocked": false,
		"complete": false, "failed": false, "wont_fix": false, "unresolved": false,
		"cancelled": false, "verified": false, "rejected": false, "diagnosing": false,
	}
	for status, want := range heads {
		if got := workItemStatusHeadsForDispatch(status); got != want {
			t.Errorf("workItemStatusHeadsForDispatch(%q) = %v, want %v", status, got, want)
		}
		// The one relationship that must hold in both directions.
		if workItemStatusRequiresRegisteredHandler(status) && !workItemStatusHeadsForDispatch(status) {
			t.Errorf("%q requires a registered handler but does not head for dispatch — "+
				"the door would skip a status 291 guards", status)
		}
	}
	if workItemStatusRequiresRegisteredHandler("detected") {
		t.Error("`detected` must stay OUT of 291's set — widening it demotes the flag-only " +
			"findings to blocked, which is bugs_closed/284's regression")
	}
}

// ── The declaration renderer, and its parity with migration 488 ──────────────

func TestOwnedPageRefusalRendererIsDelegatedNotCopied(t *testing.T) {
	for _, expr := range []string{"$1", "wi.handler_agent", "'page-build-handler'"} {
		got := workItemHandlerRefusesOwnedPagesSQL(expr)
		want := checks.HandlerDeclaresOwnedPageRefusalSQL(expr)
		if got != want {
			t.Errorf("actions and discovery_checks render DIFFERENT declaration predicates for %q:\n"+
				"  actions:          %s\n  discovery_checks: %s", expr, got, want)
		}
		if !strings.Contains(got, "agent_definitions") || !strings.Contains(got, "jsonb_path_exists") {
			t.Errorf("rendered predicate lost its substance for %q: %s", expr, got)
		}
		// The opposite posture to HandlerRegisteredSQL, deliberately: a snapshot
		// definition will never run, so it must not license a park.
		if !strings.Contains(got, "is_active") || !strings.Contains(got, "is_snapshot") {
			t.Errorf("the declaration probe must read the definition that will RUN: %s", got)
		}
	}
}

// The door and migration 488 must agree about WHERE a handler declares the
// refusal: 488 wrote it with jsonb_set at
// '{workflow,steps,load_page_record,config,refuse_owned_page}' and the door reads
// it with the jsonpath in OwnedPageRefusalDeclarationPath.
//
// ⚠ This test used to READ 488's SQL file to extract that path, and the livespec
// guard (TestNoNewMigrationFileReadersOutsideTheAllowList, bugs_open/363) rightly
// refused that at HEAD: an applied migration is checksummed history, so an
// assertion about its TEXT can never fail, while the LIVE object it wrote can
// drift away from it. The path is therefore quoted below as a Go literal — a copy
// of a FROZEN artefact, the one kind of hand-kept copy that cannot go stale — and
// what this test guards is the only side that CAN drift here: the Go constant.
// The live tie (does page-build-handler's live config still carry the key at
// that path?) is a livespec Declaration checked by the daily auditor, OWED to
// platform/livespec once the 363 lane's in-flight rename lands there (2026-08-25,
// recorded in bugs_open/363 and this lane's NOTES). Until then WII-028's live
// declaration census is the manual check.
//
// Segment 3 being a STEP NAME is what pins the object shape: if `workflow.steps`
// were ever an array, 488's path would carry an index there and the door's
// wildcard would silently match nothing.
func TestOwnedPageRefusalPathMatchesMigration488(t *testing.T) {
	// Verbatim from docs/agent_docs/sql_for_agents/488_page_build_handler_refuses_owned_pages_before_the_writer.sql
	// — resolve by SLUG, 488 is a collided number. Checksummed on apply; it cannot change.
	const migration488Path = "{workflow,steps,load_page_record,config,refuse_owned_page}"

	segs := strings.Split(strings.Trim(migration488Path, "{}"), ",")
	if len(segs) != 5 {
		t.Fatalf("488's path has %d segments, want 5: %q", len(segs), migration488Path)
	}
	stepName := segs[2]
	if _, err := strconv.Atoi(stepName); err == nil {
		t.Fatalf("488's path segment 3 is the numeric index %q, so workflow.steps would be an ARRAY — "+
			"the door's jsonpath %q assumes an object keyed by step name and would match nothing",
			stepName, checks.OwnedPageRefusalDeclarationPath)
	}
	for _, seg := range []string{"workflow", "steps", "config", "refuse_owned_page"} {
		if !strings.Contains(checks.OwnedPageRefusalDeclarationPath, seg) {
			t.Errorf("the door's jsonpath is missing segment %q that 488 wrote: %q",
				seg, checks.OwnedPageRefusalDeclarationPath)
		}
	}
	// The door reads a WILDCARD step — the only shape that matches 488's named
	// step AND any later declarer — and it requires the value to be TRUE: a
	// handler that sets the key false has not opted in.
	if !strings.Contains(checks.OwnedPageRefusalDeclarationPath, ".steps.*.") {
		t.Errorf("the door's jsonpath must wildcard the step so 488's %q and any later declarer both match: %q",
			stepName, checks.OwnedPageRefusalDeclarationPath)
	}
	if !strings.Contains(checks.OwnedPageRefusalDeclarationPath, "(@ == true)") {
		t.Errorf("the door's jsonpath must require the value TRUE: %q", checks.OwnedPageRefusalDeclarationPath)
	}
}
