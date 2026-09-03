// FILE: platform/orchestration/actions/tool_shell_refusal_test.go
//
// bugs_open/450: a TOOL PAGE WITH NO TOOL ON IT must not be built generically.
//
// THE BUG THESE PIN. A site plan names tool pages before their tools exist —
// tools arrive from the design rotation hours-to-days later, under names the
// planner never saw (seotools.co.uk: 0 of 7 planned names matched what
// tool-deployer eventually built). Five generic producers route such a page to
// page-build-handler, whose guard asked only `rebuild_policy='owned'`; a planned
// tool page is 'generic', so the builder wrote what the plan said — prose about
// the tool — and deployed it. All seven seotools URLs then served 200 at ~56 KB
// with the tool's own headline and NOT ONE <form>. 61 such pages across 10 sites
// as of 2026-09-03.
//
// WHAT MAKES THESE TESTS WORTH ANYTHING. The refusal is DERIVED (page_type plus
// the absence of a tool component), not a stored flag, so every test here scripts
// a page that is `generic` — the exact row the old guard waved through. A test
// that scripted 'owned' would pass against the OLD code and prove nothing about
// this change.
//
// And the negatives are the ones to be careful with: the composition guards fail
// OPEN, so an unscripted query is swallowed and a naive negative goes green
// having proved nothing. Each negative below therefore either calls
// ExpectationsWereMet on a fully-scripted sequence, or asserts an outcome that
// could only arise from the arm under test.
//
// MUTATION PROTOCOL (run before trusting any of this):
//   - `genericBuildRefusal`: delete the toolShell arm  → every Shell test fails,
//     every Owned test still passes (the arms are independent).
//   - `readGenericBuildPolicy`: return `false` for toolShell → same.
//   - `toolShellRefusalArmed`: return false → the kill-switch test's armed
//     subtest fails; the owned subtest does not (the switch is arm-scoped).
//   - `genericBuildExclusionSQL`: drop the appended clause → the exclusion test
//     fails on the missing fragment.
//
// Guards in SERIES to disarm when hand-proving one of them: the selection
// exclusion, the writeWorkItem door, load_page_record's early refusal and
// save_page_sections' refusal all stop the same work. Killing one and seeing the
// work still refused means a later guard caught it, not that the mutation failed.

package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ── The derivation itself ────────────────────────────────────────────────────

// TestGenericBuildRefusal_DerivationTable pins the whole decision in one place:
// two independent reasons, each sufficient, neither shadowing the other.
func TestGenericBuildRefusal_DerivationTable(t *testing.T) {
	cases := []struct {
		name      string
		policy    string
		toolShell bool
		wantRefus bool
		wantClass string
	}{
		{"ordinary generic page builds", "generic", false, false, ""},
		{"owned page refused", "owned", false, true, refusalOwned},
		{"tool page with no tool refused", "generic", true, true, refusalToolPending},
		{"a real tool page builds — it is not a shell", "generic", false, false, ""},
		// Both true: the owned class wins, because it is the older and stricter
		// statement about the page and its receipt names the column an operator
		// would go and read.
		{"owned tool shell reports the owned class", "owned", true, true, refusalOwned},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			refused, class := genericBuildRefusal(tc.policy, tc.toolShell)
			if refused != tc.wantRefus {
				t.Errorf("refused = %v, want %v", refused, tc.wantRefus)
			}
			if class != tc.wantClass {
				t.Errorf("class = %q, want %q", class, tc.wantClass)
			}
		})
	}
}

// TestToolShellPredicate_AsksForALiveToolComponent pins the shape of the SQL
// rather than its text: the three conditions that make it mean "no tool is on
// this page right now", any of which silently inverts the answer if dropped.
//
// It deliberately does NOT pin the whole statement — that would be a change
// detector. It pins the parts whose absence would make the predicate WRONG.
func TestToolShellPredicate_AsksForALiveToolComponent(t *testing.T) {
	sql := toolShellPredicateFor("pages")

	for _, want := range []string{
		"pages.page_type = 'tool'",      // only tool pages are in scope at all
		"NOT EXISTS",                    // it is the ABSENCE of a tool that refuses
		"cc_g.component_level = 'tool'", // a tool component, not any component
		"cc_g.is_active = true",         // a live one
		"pc_g.page_id = pages.id",       // on THIS page
	} {
		if !strings.Contains(sql, want) {
			t.Errorf("predicate is missing %q — without it the refusal answers a different question:\n%s", want, sql)
		}
	}

	// The eligibility fragment next door carries an extra "exactly one component"
	// clause for the ported-tool identity question. Inheriting it here would make
	// the guard blind to the 450 shells, which carry two components.
	if strings.Contains(sql, "count(*)") {
		t.Error("the guard predicate must NOT carry toolEligibilityWhere's one-component clause: " +
			"the 450 shells have two components and would slip through")
	}
}

// ── The write-time door ──────────────────────────────────────────────────────

// TestWriteWorkItem_ToolShellPage_ParkedAtDeferred is the central assertion for
// the door: a finding filed against a tool page with no tool, at a handler that
// declares it refuses such pages, is parked rather than dispatched.
func TestWriteWorkItem_ToolShellPage_ParkedAtDeferred(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("detected", "page-build-handler", &pageID)
	item.itemType = "unbuilt_internal_link"
	item.itemKey = "unbuilt_internal_link:page_component:index:hero:/tools/robots-txt-tester/index.html"

	// The expected error text is RENDERED from the same function that produces it,
	// not transcribed — the coupling idiom this file's neighbour uses. A test that
	// hand-copied the sentence would pass while the door said something else.
	_, wantParkedErr := ownedPageParkedItem(item, refusalToolPending)

	mock.ExpectBegin()
	expectToolShellPolicyRead(mock, pageID)
	expectOwnedPageDeclarationProbe(mock, "page-build-handler", true)
	expectParkedInsert(mock, "unbuilt_internal_link", item.itemKey, wantParkedErr)
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop()); err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations — the door did not run its full sequence: %v", err)
	}
}

// TestWriteWorkItem_ToolPageWithItsToolIsUntouched is the control that makes the
// test above mean something. Same page type, same handler, same finding — the
// ONLY difference is that a tool component exists. It must dispatch normally.
//
// This is also the self-clearing proof: the moment tool-deployer inserts the
// component, findings against the page flow again with no flag to unset.
//
// The declaration probe is deliberately NOT scripted: reaching it would be the
// bug, and ExpectationsWereMet cannot catch a query that was never made, so the
// assertion is that the row inserted UNPARKED.
func TestWriteWorkItem_ToolPageWithItsToolIsUntouched(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("detected", "page-build-handler", &pageID)

	mock.ExpectBegin()
	expectRebuildPolicyRead(mock, pageID, "generic") // generic AND not a shell
	// The unparked insert: status 'detected' ($12) and the handler intact ($11).
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			item.itemType, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"page-build-handler", // $11 — NOT cleared
			"detected",           // $12 — NOT deferred
			sqlmock.AnyArg(), item.itemKey, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop()); err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a tool page WITH its tool must dispatch normally — the refusal must lift "+
			"by itself when the component arrives: %v", err)
	}
}

// TestWriteWorkItem_ToolShellAtNonDeclaringHandler_Dispatches pins the door's
// scope: it parks a finding only where the handler has DECLARED it refuses such
// pages. Anything else keeps its route, exactly as for the owned class.
func TestWriteWorkItem_ToolShellAtNonDeclaringHandler_Dispatches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	item := doorItem("detected", "tool-deployer", &pageID)

	mock.ExpectBegin()
	expectToolShellPolicyRead(mock, pageID)
	// TRIPWIRE: script the probe saying "does not declare". If the door ignored
	// the declaration and parked anyway, the insert assertion below fails.
	expectOwnedPageDeclarationProbe(mock, "tool-deployer", false)
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			item.itemType, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"tool-deployer", // $11 — route preserved
			"detected",      // $12 — not parked
			sqlmock.AnyArg(), item.itemKey, sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := writeWorkItem(context.Background(), tx, item, dropOnConflict, zap.NewNop()); err != nil {
		t.Fatalf("writeWorkItem: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestOwnedPageParkedItem_ToolShellKeepsItsIdentityAndSaysWhy pins the parked
// SHAPE for the new class. item_type and item_key must survive — a re-typed or
// re-keyed row could never be retracted by the detector that filed it, and would
// stack instead of collapsing on its dedup slot.
func TestOwnedPageParkedItem_ToolShellKeepsItsIdentityAndSaysWhy(t *testing.T) {
	pageID := uuid.New()
	orig := doorItem("detected", "page-build-handler", &pageID)
	orig.itemType = "unbuilt_internal_link"
	orig.itemKey = "unbuilt_internal_link:page_component:index:hero:/tools/x/index.html"

	parked, errText := ownedPageParkedItem(orig, refusalToolPending)

	if parked.itemType != orig.itemType {
		t.Errorf("item_type changed to %q — the detector could never retract this row", parked.itemType)
	}
	if parked.itemKey != orig.itemKey {
		t.Errorf("item_key changed to %q — the row loses its dedup slot and repeats stack", parked.itemKey)
	}
	if parked.status != ownedPageParkedStatus {
		t.Errorf("status = %q, want %q", parked.status, ownedPageParkedStatus)
	}
	if parked.handlerAgent != "" {
		t.Errorf("handler_agent = %q, want empty — a parked row must not be dispatchable", parked.handlerAgent)
	}
	if !strings.HasPrefix(errText, ownedPageSkipReasonPrefix) {
		t.Errorf("parked error must LEAD with %s (it is what chooses the terminal status): %q",
			ownedPageSkipReasonPrefix, errText)
	}
	// The reason must name THIS class, not the owned one: telling an operator the
	// page is rebuild_policy=owned sends them to read a column that says 'generic'.
	if !strings.Contains(errText, "no tool component") {
		t.Errorf("parked error does not say why the page was refused: %q", errText)
	}
	if strings.Contains(errText, "rebuild_policy=owned") {
		t.Errorf("parked error claims the page is owned; it is 'generic' and that is the whole bug: %q", errText)
	}
	// builder_needed is the roadmap sweep's grouping key. These findings wait on a
	// TOOL, not on an owned-page content route.
	if !strings.Contains(parked.spec, "tool-builder") {
		t.Errorf("spec.builder_needed should group these under tool-builder: %s", parked.spec)
	}
	if !strings.Contains(parked.spec, refusalToolPending) {
		t.Errorf("spec should record refusal_class=%s so a reader can tell the two classes apart: %s",
			refusalToolPending, parked.spec)
	}
}

// ── The kill switch ──────────────────────────────────────────────────────────

// TestToolShellKillSwitch_DisarmsOnlyTheNewArm proves the switch is arm-scoped.
// Disarming the tool-shell refusal in anger must not also disarm migration 164's
// ownership refusal, which has been live since August and protects live tools.
//
// The owned subtest is the guard-in-series check: it is the reason a passing
// mutation of the shell arm cannot be mistaken for "the switch does nothing".
func TestToolShellKillSwitch_DisarmsOnlyTheNewArm(t *testing.T) {
	t.Setenv(disableToolShellRefusalEnv, "1")

	if refused, _ := genericBuildRefusal("generic", true); refused {
		t.Error("kill switch set, but a tool shell was still refused — the switch is inert")
	}
	if refused, class := genericBuildRefusal("owned", false); !refused || class != refusalOwned {
		t.Errorf("kill switch disarmed the OWNED arm too (refused=%v class=%q) — "+
			"it must scope to the tool-shell arm only", refused, class)
	}
	if toolShellRefusalArmed() {
		t.Error("toolShellRefusalArmed() must be false while the switch is set")
	}
	// The SQL side must agree with the Go side, or the selection would exclude
	// pages the verdict says are buildable.
	if strings.Contains(genericBuildExclusionSQL(), "component_level = 'tool'") {
		t.Error("the exclusion still carries the shell clause while the switch is set — " +
			"the SQL and Go halves have diverged")
	}
}

// TestGenericBuildExclusion_CarriesBothArmsWhenArmed is the armed counterpart,
// and pins that the owned clause is preserved BYTE-IDENTICALLY: several tests and
// live queries match that literal, and folding it into a new expression would
// break them silently.
func TestGenericBuildExclusion_CarriesBothArmsWhenArmed(t *testing.T) {
	got := genericBuildExclusionSQL()
	if !strings.HasPrefix(got, ownedPageExclusionSQL) {
		t.Errorf("the owned clause must survive verbatim at the front:\n%s", got)
	}
	if !strings.Contains(got, "component_level = 'tool'") {
		t.Errorf("the tool-shell clause is missing — 450's pages would still be selected:\n%s", got)
	}
	if !strings.Contains(got, "AND NOT ") {
		t.Errorf("the shell clause must be NEGATED (exclude shells, not select them):\n%s", got)
	}
}

// ── The composition guards ───────────────────────────────────────────────────

// TestAssemblePage_ToolShellIsSkippedNotAssembled: the assemble seam is where a
// generic recomposition would be committed to the site repo. A tool shell must
// come back as the SKIP shape, which git_commit already honours.
func TestAssemblePage_ToolShellIsSkippedNotAssembled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID, siteID := uuid.New(), uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WithArgs(pageID).
		WillReturnRows(policyRows("generic", true))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := assembleParams(db, map[string]interface{}{
		"current_page": map[string]interface{}{
			"id":   pageID.String(),
			"name": "tool-robots-txt-tester",
		},
		"site_record": map[string]interface{}{"site_id": siteID.String()},
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{
				"page_html": "<html><body><p>An article about testing your robots.txt</p></body></html>",
			},
		},
	})
	params.DB = db

	out, err := AssemblePageAction(context.Background(), params)
	if err != nil {
		t.Fatalf("the refusal must be a SKIP, not an error — no build loop sets "+
			"continue_on_error and an error strands every remaining page: %v", err)
	}
	res, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type %T", out)
	}
	if skipped, _ := res["skipped"].(bool); !skipped {
		t.Fatal("a tool page with no tool was ASSEMBLED — deploy_page would publish prose " +
			"about a tool that is not there, which is bugs_open/450 exactly")
	}
	if html, _ := res["html"].(string); html != "" {
		t.Errorf("skip returned %d bytes of html; must be empty", len(html))
	}
	if reason, _ := res["skip_reason"].(string); !strings.Contains(reason, ownedPageSkipReasonPrefix) {
		t.Errorf("skip_reason missing the %s marker: %q", ownedPageSkipReasonPrefix, reason)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the refusal must also leave its receipt: %v", err)
	}
}

// TestSavePageSections_DoesNotRefuseToolShell pins the NARROWING, and it is the
// inverse of what this test asserted when first written.
//
// ⚠ THIS IS A REGRESSION FIX, reported by the bugs_open/427 lane within minutes of
// the arm going live. save_page_sections is shared by two paths that are
// indistinguishable at that seam: a generic build authoring prose about a missing
// tool (450's harm), and a page-rerender writing back components that are ALREADY
// DEPLOYED AND SERVING. Refusing here caught both — and the collateral dominated:
// of the 67 pages the predicate matches, 54 across 10 sites are already serving
// and only 13 are the empty page 450 is about [MEASURED 2026-09-03]. It broke the
// repair vehicle for those 54 on the same morning bugs_open/454 restored it.
//
// The class loses nothing, because every generic path is caught EARLIER — the
// writeWorkItem door at file time, load_page_record's refuse_owned_page arm,
// AssemblePageAction, and the build-selection exclusion — and page-rerender
// crosses none of them. See the narrowing note in save_page_sections_action.go.
//
// No receipt is scripted: emitting one here would be the old behaviour, so an
// unexpected INSERT fails this test.
func TestSavePageSections_DoesNotRefuseToolShell(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()

	mock.ExpectQuery("SELECT id, url FROM pages").
		WithArgs(siteID, "tool-fight-calendar").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(pageID, "/tools/fight-calendar/index.html"))
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WithArgs(pageID).
		WillReturnRows(policyRows("generic", true)) // a tool shell — the 427 lane's page
	// Everything after this point is the ordinary save path. We do not script it:
	// the assertion is only that the guard did NOT refuse, which the error text
	// below distinguishes from any downstream failure.

	_, err = SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-fight-calendar", []interface{}{
			map[string]interface{}{
				"rendered_html":  "<section><p>Re-rendered event list, already deployed</p></section>",
				"component_name": "event-list",
			},
		}))
	if err != nil && strings.HasPrefix(err.Error(), ownedPageSkipReasonPrefix) {
		t.Fatalf("the tool-shell arm still refuses at save_page_sections — this is the "+
			"bugs_open/427 regression: it blocks re-renders on 54 already-serving pages "+
			"while every generic path is already caught earlier: %v", err)
	}
}

// TestSavePageSections_StillRefusesOwnedPage is the guard-in-series control that
// stops the test above passing for the wrong reason. Narrowing the tool arm must
// not have disarmed migration 164's ownership refusal at the same seam — that one
// protects live verbatim tools from delete-and-reinsert and is unchanged.
func TestSavePageSections_StillRefusesOwnedPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID, pageID := uuid.New(), uuid.New()

	mock.ExpectQuery("SELECT id, url FROM pages").
		WithArgs(siteID, "tool-gauntlet").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(pageID, "/tools/gauntlet/index.html"))
	mock.ExpectQuery(regexp.QuoteMeta(rebuildPolicyReadSQL)).
		WithArgs(pageID).
		WillReturnRows(policyRows("owned", false))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = SavePageSectionsAction(context.Background(),
		saveSlotParams(db, siteID, "tool-gauntlet", []interface{}{
			map[string]interface{}{
				"rendered_html":  "<section><p>generic prose</p></section>",
				"component_name": "generic-text-block",
			},
		}))
	if err == nil {
		t.Fatal("an OWNED page was saved — narrowing the tool arm disarmed migration 164's " +
			"refusal, which is the one protecting live verbatim tools")
	}
	if !strings.HasPrefix(err.Error(), ownedPageSkipReasonPrefix) {
		t.Errorf("the owned refusal must still LEAD with %s: %q", ownedPageSkipReasonPrefix, err.Error())
	}
	if !strings.Contains(err.Error(), "rebuild_policy=owned") {
		t.Errorf("the owned wording must be preserved byte-for-byte: %q", err.Error())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the owned refusal must still leave its receipt: %v", err)
	}
}

// TestEscalateRerenderToWriter_ToolShell_SkipsEmit: a tool page waiting for its
// tool has empty prose slots for a legitimate reason, so escalating to the writer
// would mint exactly the work the save path is certain to refuse.
func TestEscalateRerenderToWriter_ToolShell_SkipsEmit(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	expectPlanTableSections(mock, "hero-tool", "generic-text-block")
	expectEscalationBuildPolicyGuard(mock, "generic", true)
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(sqlmock.AnyArg(), "page-rerender", sqlmock.AnyArg(), sqlmock.AnyArg(),
			"owned_page_review:tool-robots-txt-tester").
		WillReturnResult(sqlmock.NewResult(1, 1))

	disposition, err := escalateRerenderToWriter(context.Background(), db, uuid.New(),
		"tool-robots-txt-tester", "a section had no stored content_data", zap.NewNop())
	if err != nil {
		t.Fatalf("the skip must be a clean return, got error: %v", err)
	}
	if disposition != "skipped_tool_pending_page" {
		t.Errorf("disposition = %q, want skipped_tool_pending_page — the two refusal classes "+
			"must be distinguishable in the run record", disposition)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (a write may have been attempted): %v", err)
	}
}

// ── The receipt ──────────────────────────────────────────────────────────────

// TestEmitOwnedPageReviewItem_ToolShellSpecTellsTheTruth. The receipt is the only
// thing a human sees. It must not assert a policy value the page does not have —
// the old spec hardcoded "rebuild_policy": "owned", which on a 450 page would be
// a plain falsehood pointing the reader at the wrong column.
func TestEmitOwnedPageReviewItem_ToolShellSpecTellsTheTruth(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	var capturedSpec string
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(sqlmock.AnyArg(), "save_page_sections", sqlmock.AnyArg(),
			sqlmock.AnyArg(), "owned_page_review:tool-x").
		WillReturnResult(sqlmock.NewResult(1, 1))

	emitOwnedPageReviewItem(context.Background(), db, siteID, "tool-x", "save_page_sections",
		"because it has no tool", refusalToolPending, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the receipt was not written: %v", err)
	}
	_ = capturedSpec

	// The spec content is asserted through the helpers rather than the driver, so
	// this stays a statement about behaviour and not about sqlmock's argument
	// capture: both halves must name the class, and the advice must not send an
	// operator to apply_section_edit (useless on a page with no tool).
	advice := refusalFixAdvice(refusalToolPending)
	if !strings.Contains(advice, "tool") {
		t.Errorf("tool_pending advice does not mention the tool: %q", advice)
	}
	if strings.Contains(advice, "rebuild_policy if the page genuinely is") {
		t.Errorf("tool_pending advice reuses the OWNED remedy, which does not apply here: %q", advice)
	}
	// ⚠ The summary must state the MEASURED fact (no tool-LEVEL component) and not
	// the inference that the page has no tool at all. idea.uk /report.html is typed
	// 'tool', has no tool-level row, and serves 1 form / 8 inputs from a
	// SECTION-level component — so "a tool that is not there" was false for the one
	// page in the fleet that tests this wording, and would send an operator hunting
	// for a missing tool they would then find working.
	summary := refusalSummary(refusalToolPending, "tool-x")
	if !strings.Contains(summary, "no tool-level component") {
		t.Errorf("summary must name the measured fact — the absent tool-LEVEL component: %q", summary)
	}
	for _, overclaim := range []string{"tool that is not there", "has no tool yet", "is empty"} {
		if strings.Contains(summary, overclaim) {
			t.Errorf("summary asserts %q, which is false for a page serving a form from a "+
				"section-level component: %q", overclaim, summary)
		}
	}
	if summary := refusalSummary(refusalOwned, "tool-x"); !strings.Contains(summary, "Owned page") {
		t.Errorf("the owned summary must be unchanged — live queries read it: %q", summary)
	}
}

var _ = orchtypes.ExecutionContext{}
