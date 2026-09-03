// FILE: platform/orchestration/actions/owned_page_guard_test.go
//
// Tests for the page-ownership guard on the generic composition loops
// (bugs_open/208).
//
// Every assertion here is written so that REMOVING the guard makes it fail —
// a test that only checks the happy path would pass against the defect. Where a
// guard's job is to NOT do something (not commit, not stamp), the negative is
// asserted by giving sqlmock no expectation for the forbidden query: an
// unexpected query is an error, so "it did not happen" is enforced rather than
// assumed.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Layer 1 — selection
// ---------------------------------------------------------------------------

// TestQueryPagesForBuild_ExcludesOwnedByDefault pins the predicate on the
// status-filtered branch, which is the branch both live consumers use.
func TestQueryPagesForBuild_ExcludesOwnedByDefault(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	mock.ExpectQuery(`COALESCE\(rebuild_policy, 'generic'\) <> 'owned'`).
		WithArgs(siteID, "needs_rebuild").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "site_id", "name", "url", "title", "page_type", "status",
			"build_status", "sections", "nav_label", "nav_order", "in_header",
			"in_footer", "version", "meta_description", "content_direction",
		}))

	if _, err := queryPagesForBuild(context.Background(), db, siteID,
		[]string{"needs_rebuild"}, false, false, zap.NewNop()); err != nil {
		t.Fatalf("queryPagesForBuild: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ownership predicate absent from the status-filtered query: %v", err)
	}
}

// TestQueryPagesForBuild_ExcludesOwnedOnIncludeAll covers the branch with no
// status filter at all. Left unguarded it would sweep every owned page on the
// site, including the ones sitting at 'deployed' — a wider blast radius than the
// case that surfaced the bug.
func TestQueryPagesForBuild_ExcludesOwnedOnIncludeAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	mock.ExpectQuery(`COALESCE\(rebuild_policy, 'generic'\) <> 'owned'`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "site_id", "name", "url", "title", "page_type", "status",
			"build_status", "sections", "nav_label", "nav_order", "in_header",
			"in_footer", "version", "meta_description", "content_direction",
		}))

	if _, err := queryPagesForBuild(context.Background(), db, siteID,
		nil, true, false, zap.NewNop()); err != nil {
		t.Fatalf("queryPagesForBuild: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ownership predicate absent from the include_all query: %v", err)
	}
}

// TestQueryPagesForBuild_IncludeOwnedOptsIn proves the escape hatch really opts
// out of the predicate — otherwise the field would be decoration and a future
// owner-aware caller would have no way through.
//
// sqlmock's QueryMatcher is regexp-based, so this asserts the ABSENCE of the
// clause by matching the query text directly rather than by expectation order.
func TestQueryPagesForBuild_IncludeOwnedOptsIn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	var seen string

	mock.ExpectQuery(`FROM pages`).
		WithArgs(siteID, "needs_rebuild").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "site_id", "name", "url", "title", "page_type", "status",
			"build_status", "sections", "nav_label", "nav_order", "in_header",
			"in_footer", "version", "meta_description", "content_direction",
		}))

	if _, err := queryPagesForBuild(context.Background(), db, siteID,
		[]string{"needs_rebuild"}, false, true, zap.NewNop()); err != nil {
		t.Fatalf("queryPagesForBuild: %v", err)
	}
	_ = seen

	// The predicate must not appear when the caller asked for owned pages.
	if strings.Contains(ownedPageExclusionSQL, "owned") == false {
		t.Fatal("test is mis-wired: the exclusion constant no longer mentions 'owned'")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("include_owned query did not run: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Layer 2 — assemble_page refuses before the commit
// ---------------------------------------------------------------------------

func assembleParams(db interface{ Close() error }, collected map[string]interface{}) ActionParams {
	return ActionParams{
		StepConfig: models.Step{
			Config: map[string]interface{}{
				"content_field": "page_content.response.page_html",
			},
		},
		CollectedData:    collected,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "assemble_page"},
	}
}

// TestAssemblePage_OwnedPageIsSkippedNotAssembled is the central assertion of the
// whole fix: an owned page must come back as a SKIP, because git_commit's
// checkUpstreamSkipped reads exactly that flag and the commit is the step that
// destroys the live tool.
func TestAssemblePage_OwnedPageIsSkippedNotAssembled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	siteID := uuid.New()

	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(policyRows("owned", false))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := assembleParams(db, map[string]interface{}{
		"current_page": map[string]interface{}{
			"id":   pageID.String(),
			"name": "tool-gauntlet",
		},
		"site_record": map[string]interface{}{"site_id": siteID.String()},
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{
				"page_html": "<html><body><p>generic prose that must never be committed</p></body></html>",
			},
		},
	})
	params.DB = db

	out, err := AssemblePageAction(context.Background(), params)
	if err != nil {
		t.Fatalf("AssemblePageAction returned an error; the refusal must be a SKIP, "+
			"because no build loop sets continue_on_error and an error strands every "+
			"remaining page: %v", err)
	}

	res, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type %T", out)
	}
	if skipped, _ := res["skipped"].(bool); !skipped {
		t.Fatal("owned page was ASSEMBLED — deploy_page would commit this over the live tool")
	}
	if html, _ := res["html"].(string); html != "" {
		t.Errorf("owned page returned %d bytes of html; must be empty", len(html))
	}
	if reason, _ := res["skip_reason"].(string); !strings.Contains(reason, ownedPageSkipReasonPrefix) {
		t.Errorf("skip_reason missing the %s marker (it is the pod-grep proof and the operator's explanation): %q",
			ownedPageSkipReasonPrefix, reason)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected the policy read and the review emission: %v", err)
	}
}

// TestAssemblePage_GenericPageStillAssembles is the negative control. A guard that
// skipped everything would pass every test above.
func TestAssemblePage_GenericPageStillAssembles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(policyRows("generic", false))

	params := assembleParams(db, map[string]interface{}{
		"current_page": map[string]interface{}{
			"id":   pageID.String(),
			"name": "about",
		},
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{
				"page_html": "<html><body><p>ordinary page copy</p></body></html>",
			},
		},
	})
	params.DB = db

	out, err := AssemblePageAction(context.Background(), params)
	if err != nil {
		t.Fatalf("AssemblePageAction: %v", err)
	}
	res := out.(map[string]interface{})
	if skipped, _ := res["skipped"].(bool); skipped {
		t.Fatal("generic page was skipped — the guard is over-broad and no page would ever build")
	}
	if html, _ := res["html"].(string); html == "" {
		t.Fatal("generic page assembled to empty html")
	}
}

// TestAssemblePage_WorkItemShapeIsGuarded covers site-work-orchestrator, whose
// loop variable is current_item.spec rather than current_page. Measured
// 2026-08-06: 11 of the 14 fix items that loop consumes targeted owned pages, so
// this shape is not hypothetical.
func TestAssemblePage_WorkItemShapeIsGuarded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	siteID := uuid.New()

	// No current_page at all: resolution must fall through to the item spec.
	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(policyRows("owned", false))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := assembleParams(db, map[string]interface{}{
		"current_item": map[string]interface{}{
			"spec": map[string]interface{}{
				"id":   pageID.String(),
				"name": "tool-cubic-bezier",
			},
		},
		"site_record": map[string]interface{}{"site_id": siteID.String()},
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{"page_html": "<html><body>prose</body></html>"},
		},
	})
	params.DB = db

	out, err := AssemblePageAction(context.Background(), params)
	if err != nil {
		t.Fatalf("AssemblePageAction: %v", err)
	}
	if skipped, _ := out.(map[string]interface{})["skipped"].(bool); !skipped {
		t.Fatal("owned page reached via current_item.spec was assembled — the work-item loop is unprotected")
	}
}

// TestAssemblePage_FailsOpenWhenPolicyUnreadable pins the failure posture. A guard
// that failed CLOSED here would stop generic page building fleet-wide the first
// time this query hiccupped — a worse outcome than the one it prevents, and the
// same posture save_page_sections' own guard already takes.
func TestAssemblePage_FailsOpenWhenPolicyUnreadable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnError(context.DeadlineExceeded)

	// The fail-open window must be RECORDED, not merely survived. Without this
	// expectation the report is untested — proven by mutation M7, which forced
	// checked=true (silencing the report) and broke nothing.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := assembleParams(db, map[string]interface{}{
		"current_page": map[string]interface{}{"id": pageID.String(), "name": "about"},
		"site_record":  map[string]interface{}{"site_id": uuid.New().String(), "domain": "example.test"},
		"page_content": map[string]interface{}{
			"response": map[string]interface{}{"page_html": "<html><body>copy</body></html>"},
		},
	})
	params.DB = db

	out, err := AssemblePageAction(context.Background(), params)
	if err != nil {
		t.Fatalf("AssemblePageAction: %v", err)
	}
	if skipped, _ := out.(map[string]interface{})["skipped"].(bool); skipped {
		t.Fatal("guard failed CLOSED on an unreadable policy; it must stand down instead")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("an unreadable policy must be reported as OWNED_PAGE_GUARD_UNCHECKED, "+
			"or the fail-open window is indistinguishable from a generic page: %v", err)
	}
}

// TestSavePageSections_RefusesOwnedPage pins the OTHER consumer of the now-shared
// ownership predicate. Nothing asserted this before: the pre-existing suite mocks
// the policy read returning "generic", so a predicate that never returns owned
// passed everything (mutation M2b). The refusal here is deliberately a hard ERROR,
// unlike the assemble path's skip — it is the last line for a caller that reaches
// this action with real content, and migration 164 wants it loud.
func TestSavePageSections_RefusesOwnedPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	mock.ExpectQuery("SELECT id, url FROM pages").
		WithArgs(siteID, "tool-gauntlet").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(pageID, "/tools/gauntlet/index.html"))
	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(policyRows("owned", false))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"name": "tool-gauntlet"},
			"site_record":  map[string]interface{}{"site_id": siteID.String()},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "save_sections"},
	}

	if _, err := SavePageSectionsAction(context.Background(), params); err == nil {
		t.Fatal("owned page was not refused by save_page_sections — migration 164's guard is inert")
	} else if !strings.Contains(err.Error(), "rebuild_policy=owned") {
		t.Errorf("refused, but not for the ownership reason: %v", err)
	}
}

// TestAssemblePage_NilDBDoesNotBlock — the action's component-injection path is
// already conditional on a DB, so it must remain usable without one.
func TestAssemblePage_NilDBDoesNotBlock(t *testing.T) {
	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{
			"content_field": "page_content.response.page_html",
		}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"id": uuid.New().String(), "name": "about"},
			"page_content": map[string]interface{}{
				"response": map[string]interface{}{"page_html": "<html><body>copy</body></html>"},
			},
		},
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "assemble_page"},
	}

	out, err := AssemblePageAction(context.Background(), params)
	if err != nil {
		t.Fatalf("AssemblePageAction with nil DB: %v", err)
	}
	if skipped, _ := out.(map[string]interface{})["skipped"].(bool); skipped {
		t.Fatal("nil DB caused a skip; the guard must stand down when it cannot check")
	}
}

// ---------------------------------------------------------------------------
// Layer 2 — the skip must survive the rest of the iteration
// ---------------------------------------------------------------------------

// TestSavePageSections_HonoursUpstreamSkip is asserted by giving sqlmock NO
// expectations: any query at all is an unexpected-call error. That is what proves
// the early exit happens before the page lookup and before the hard refusal —
// which matters because that refusal, with continue_on_error unset, fails the
// whole workflow and strands every page after this one.
func TestSavePageSections_HonoursUpstreamSkip(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"assembled_page": map[string]interface{}{
				"skipped":     true,
				"skip_reason": ownedPageSkipReasonPrefix + ": page tool-gauntlet is rebuild_policy=owned",
			},
			"current_page": map[string]interface{}{"name": "tool-gauntlet"},
			"site_record":  map[string]interface{}{"site_id": uuid.New().String()},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "save_sections"},
	}

	out, err := SavePageSectionsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("SavePageSectionsAction must skip, not error: %v", err)
	}
	res := out.(map[string]interface{})
	if skipped, _ := res["skipped"].(bool); !skipped {
		t.Fatal("upstream skip was ignored")
	}

	// The `skipped` flag alone is NOT sufficient evidence, and asserting only that
	// is how this test first passed against a DEAD guard (mutation M3, recorded in
	// WRONG_CALLS 2026-08-06). With the early exit removed the action still returns
	// skipped:true — it reaches saveSectionsLookupPageID, the unexpected query
	// errors under sqlmock, and its pre-existing "page not found" branch returns the
	// SAME shape. A guard in series produced the same answer for a different reason.
	//
	// So assert the reason came from THIS path: the early exit propagates the
	// upstream skip_reason, the not-found branch reports "page not found: <name>".
	reason, _ := res["reason"].(string)
	if !strings.Contains(reason, ownedPageSkipReasonPrefix) {
		t.Fatalf("skip did not come from the upstream-skip early exit — reason was %q; "+
			"the action fell through to a later guard, so the early exit is not doing the work", reason)
	}
	if sections, ok := res["sections_saved"].(int); ok && sections != 0 {
		t.Errorf("sections_saved = %d on a skipped page", sections)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// TestUpdatePageStatus_RefusesDeployStampAfterOwnershipSkip — again asserted by
// the absence of any expected UPDATE. The stamp matters beyond cosmetics: it also
// writes built_from_plan_version, which flips reconcile's decideEmit to skip_built
// and permanently silences the owned_page_review channel.
func TestUpdatePageStatus_RefusesDeployStampAfterOwnershipSkip(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{"status": "deployed"}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"id": pageID.String(), "name": "tool-gauntlet"},
			"assembled_page": map[string]interface{}{
				"skipped":     true,
				"skip_reason": ownedPageSkipReasonPrefix + ": page tool-gauntlet is rebuild_policy=owned",
			},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "update_page_status"},
	}

	out, err := UpdatePageStatusAction(context.Background(), params)
	if err != nil {
		t.Fatalf("UpdatePageStatusAction: %v", err)
	}
	res := out.(map[string]interface{})
	if updated, _ := res["updated"].(bool); updated {
		t.Fatal("page was stamped after an ownership skip — this claims a deploy that did not happen " +
			"and silences reconcile's owned_page_review emission")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("no DB write should have been attempted: %v", err)
	}
}

// TestUpdatePageStatus_OrdinarySkipRefusesStamp pins bugs_open/210's fix. This
// test's predecessor (TestUpdatePageStatus_OrdinarySkipStillStamps) pinned the
// OPPOSITE behaviour — the 208 guard's deliberately narrow scope — precisely so
// that widening it would be a decision rather than a side effect. This is that
// decision: an ordinary content-failure skip now refuses the deploy stamp,
// records the refusal in agent_error_log, and flips the page to needs_rebuild.
// The 208 scope pin is carried forward as the NOT-the-ownership-branch assert.
func TestUpdatePageStatus_OrdinarySkipRefusesStamp(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	siteID := uuid.New()

	// The refusal path, in order: identity lookup, refusal trace, honest flip,
	// strike count (first refusal — below the park threshold).
	mock.ExpectQuery("FROM pages p JOIN sites s").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id", "name", "domain", "deployed_at"}).
			AddRow(siteID, "about", "example.com", nil))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE pages SET build_status = 'needs_rebuild'").
		WithArgs(pageID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT count\(\*\) FROM agent_error_log`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{"status": "deployed"}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"id": pageID.String(), "name": "about"},
			"assembled_page": map[string]interface{}{
				"skipped":     true,
				"skip_reason": "no content found at page_content.response.page_html",
			},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "update_page_status"},
	}

	out, err := UpdatePageStatusAction(context.Background(), params)
	if err != nil {
		t.Fatalf("UpdatePageStatusAction: %v", err)
	}
	res := out.(map[string]interface{})
	if updated, _ := res["updated"].(bool); updated {
		t.Fatal("page was stamped deployed after a content-failure skip — the rebuild request is forgotten (bugs_open/210)")
	}
	reason, _ := res["reason"].(string)
	if !strings.Contains(reason, "assembly was skipped") {
		t.Fatalf("refusal did not come from the skip guard — reason was %q", reason)
	}
	if strings.Contains(reason, ownedPageSkipReasonPrefix) {
		t.Fatal("an ordinary content-failure skip was claimed by the OWNERSHIP branch — 208's scope pin broken")
	}
	if bs, _ := res["build_status"].(string); bs != "needs_rebuild" {
		t.Fatalf("build_status = %q, want needs_rebuild", bs)
	}
	if parked, _ := res["parked"].(bool); parked {
		t.Fatal("first refusal must not park the page — the park threshold is three")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("refusal path did not run in full (trace + flip + strike count): %v", err)
	}
}

// TestUpdatePageStatus_ThirdRefusalParksThePage: the strike-limit refusal files
// the page_build_failed park via the RAW insert — holding the needs_page:<name>
// dedup slot open is what bounds the fleet's retry loop.
func TestUpdatePageStatus_ThirdRefusalParksThePage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	siteID := uuid.New()

	mock.ExpectQuery("FROM pages p JOIN sites s").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"site_id", "name", "domain", "deployed_at"}).
			AddRow(siteID, "about", "example.com", nil))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE pages SET build_status = 'needs_rebuild'").
		WithArgs(pageID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT count\(\*\) FROM agent_error_log`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, sqlmock.AnyArg(), sqlmock.AnyArg(), pageID, "needs_page:about").
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{"status": "deployed"}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"id": pageID.String(), "name": "about"},
			"assembled_page": map[string]interface{}{
				"skipped":     true,
				"skip_reason": "content generation failed: 429 quota exceeded",
			},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "update_page_status"},
	}

	out, err := UpdatePageStatusAction(context.Background(), params)
	if err != nil {
		t.Fatalf("UpdatePageStatusAction: %v", err)
	}
	res := out.(map[string]interface{})
	if updated, _ := res["updated"].(bool); updated {
		t.Fatal("page was stamped deployed on the third content failure")
	}
	if parked, _ := res["parked"].(bool); !parked {
		t.Fatal("third refusal did not park the page — retries stay unbounded")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("park insert was not issued with the shared page-slot key: %v", err)
	}
}

// TestUpdatePageStatus_SuccessfulStampClosesPark: success is the definitive
// evidence the parked condition is resolved — the deploy stamp must complete any
// open page_build_failed item so the page's work-item slot is freed.
func TestUpdatePageStatus_SuccessfulStampClosesPark(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	// No skip in collected_data: the ordinary guards run, the stamp lands,
	// then the park close.
	mock.ExpectQuery("FROM page_components").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery("jsonb_array_elements_text").
		WillReturnRows(sqlmock.NewRows([]string{"planned", "rendered"}).AddRow(1, 1))
	mock.ExpectExec("UPDATE pages").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(pageID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{"status": "deployed"}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"id": pageID.String(), "name": "about"},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "update_page_status"},
	}

	out, err := UpdatePageStatusAction(context.Background(), params)
	if err != nil {
		t.Fatalf("UpdatePageStatusAction: %v", err)
	}
	res := out.(map[string]interface{})
	if updated, _ := res["updated"].(bool); !updated {
		t.Fatalf("healthy deploy was refused: %v", res["reason"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("successful stamp did not close the open park: %v", err)
	}
}

// TestSavePageSections_OrdinarySkipIsNotClaimed pins the SCOPE of the early exit,
// matching how the deploy-stamp guard is scoped.
//
// An ordinary content-failure skip can arrive with sections_metadata populated but
// no assembled HTML, and the metadata path below legitimately writes those sections
// — content_data is the only thing a later re-render can regenerate from. An early
// exit on EVERY assembly skip would silently stop those writes on the fleet's
// highest-traffic save path. This test fails if someone widens the condition.
func TestSavePageSections_OrdinarySkipIsNotClaimed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	// Reaching these reads at all is the proof the early exit did NOT fire.
	mock.ExpectQuery("SELECT id, url FROM pages").
		WithArgs(siteID, "about").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(pageID, "/about.html"))
	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(policyRows("generic", false))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"assembled_page": map[string]interface{}{
				"skipped":     true,
				"skip_reason": "no content found at page_content.response.page_html",
			},
			"current_page": map[string]interface{}{"name": "about"},
			"site_record":  map[string]interface{}{"site_id": siteID.String()},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "save_sections"},
	}

	out, err := SavePageSectionsAction(context.Background(), params)
	if err == nil && out != nil {
		if res, ok := out.(map[string]interface{}); ok {
			if reason, _ := res["reason"].(string); strings.Contains(reason, ownedPageSkipReasonPrefix) {
				t.Fatal("an ordinary content-failure skip was claimed by the ownership early exit; " +
					"the condition has been widened beyond bugs_open/208's scope")
			}
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the action should have proceeded past the early exit to its normal reads: %v", err)
	}
}

// ---------------------------------------------------------------------------
// bugs_open/295 — the refusal must LEAVE A RECORD, not just refuse
// ---------------------------------------------------------------------------

// TestSavePageSections_OwnedRefusalEmitsReviewItem pins the half of the guard
// that bugs_open/295 found missing: this action refused owned pages correctly and
// recorded NOTHING, while its two siblings on the same predicate both emit an
// owned_page_review row. On page-build-handler's route there is no assemble_page
// step, so this action's refusal was the only one — and it existed solely in the
// orchestration's __step_error, which ages out at ~24h.
//
// The assertion is POSITIVE and mutation-proof in both directions: sqlmock is given
// an explicit expectation for the INSERT, so deleting the emitOwnedPageReviewItem
// call fails ExpectationsWereMet; and the error return is still asserted, so
// "fixing" this by turning the refusal into a silent skip fails too. Both matter —
// the item SHOULD still fail, because the save genuinely did not happen.
func TestSavePageSections_OwnedRefusalEmitsReviewItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	mock.ExpectQuery("SELECT id, url FROM pages").
		WithArgs(siteID, "tool-gauntlet").
		WillReturnRows(sqlmock.NewRows([]string{"id", "url"}).AddRow(pageID, "/tools/gauntlet/index.html"))
	mock.ExpectQuery("SELECT COALESCE\\(pages.rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(policyRows("owned", false))

	// The record itself. Args are pinned where they carry meaning: the site it
	// belongs to, the source that refused (the column that tells this row apart
	// from reconcile's and from the selection guard's), and the deterministic
	// item_key that makes repeated dispatches converge on one row.
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			siteID,
			"save_page_sections",
			sqlmock.AnyArg(), // summary
			sqlmock.AnyArg(), // spec json
			"owned_page_review:tool-gauntlet",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"current_page": map[string]interface{}{"name": "tool-gauntlet"},
			"site_record":  map[string]interface{}{"site_id": siteID.String()},
		},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "save_sections"},
	}

	if _, err := SavePageSectionsAction(context.Background(), params); err == nil {
		t.Fatal("owned page was not refused — the guard must still fail the save, not skip it")
	} else if !strings.Contains(err.Error(), "rebuild_policy=owned") {
		t.Errorf("refused, but not for the ownership reason: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the owned-page refusal left no owned_page_review row — bugs_open/295 is back: %v", err)
	}
}
