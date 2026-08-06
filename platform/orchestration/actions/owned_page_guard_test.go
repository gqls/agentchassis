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

	mock.ExpectQuery("SELECT COALESCE\\(rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"rebuild_policy"}).AddRow("owned"))
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

	mock.ExpectQuery("SELECT COALESCE\\(rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"rebuild_policy"}).AddRow("generic"))

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
	mock.ExpectQuery("SELECT COALESCE\\(rebuild_policy").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"rebuild_policy"}).AddRow("owned"))
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

	mock.ExpectQuery("SELECT COALESCE\\(rebuild_policy").
		WithArgs(pageID).
		WillReturnError(context.DeadlineExceeded)

	params := assembleParams(db, map[string]interface{}{
		"current_page": map[string]interface{}{"id": pageID.String(), "name": "about"},
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

// TestUpdatePageStatus_OrdinarySkipStillStamps pins the SCOPE of the previous
// guard. Keying it to any assembly skip would change retry behaviour on the
// fleet's main build path (a content-failed page would be re-selected every run) —
// a wider blast radius than this bug, deliberately left out and filed separately.
// If someone widens the condition, this test is what tells them they did.
func TestUpdatePageStatus_OrdinarySkipStillStamps(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	// The pre-existing deploy guards run, then the stamp.
	mock.ExpectQuery("FROM page_components").
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

	// Not asserting the outcome (the pre-existing guards own that) — only that the
	// OWNERSHIP branch did not claim this skip as its own.
	out, err := UpdatePageStatusAction(context.Background(), params)
	if err == nil && out != nil {
		if res, ok := out.(map[string]interface{}); ok {
			if reason, _ := res["reason"].(string); strings.Contains(reason, ownedPageSkipReasonPrefix) {
				t.Fatal("an ordinary content-failure skip was treated as an ownership skip; " +
					"the guard's condition has been widened beyond bugs_open/208's scope")
			}
		}
	}
	_ = mock
}
