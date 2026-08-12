// FILE: platform/orchestration/actions/archived_page_guard_test.go
//
// Tests for the archived-page deploy guard (bugs_open/266).
//
// Every assertion is written so that REMOVING the guard makes it fail. Where the
// guard's job is to NOT do something (not commit, not stamp), the negative is
// asserted by giving sqlmock no expectation for the forbidden query — an
// unexpected query is an error, so "it did not happen" is enforced rather than
// assumed. Same discipline as owned_page_guard_test.go, deliberately.
//
// The control that matters most is TestGitCommit_LivePageStillDeploys: a guard
// that refuses EVERYTHING would satisfy every other test in this file.
package actions

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Layer 1 — the status read itself
// ---------------------------------------------------------------------------

func TestPageIsArchivedForGuard_ReadsStatus(t *testing.T) {
	for _, tc := range []struct {
		name         string
		status       string
		wantArchived bool
	}{
		{"archived page is archived", "archived", true},
		{"active page is not", "active", false},
		// 'deleted' is a different retirement state with its own handling; this
		// guard must not silently annex it.
		{"deleted is not this guard's business", "deleted", false},
		{"empty status is not archived", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			pageID := uuid.New()
			mock.ExpectQuery(`SELECT COALESCE\(status, ''\) FROM pages WHERE id = \$1`).
				WithArgs(pageID).
				WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(tc.status))

			archived, checked := pageIsArchivedForGuard(context.Background(), db, pageID, zap.NewNop())
			if !checked {
				t.Fatal("checked=false on a successful read")
			}
			if archived != tc.wantArchived {
				t.Errorf("status %q: archived=%v, want %v", tc.status, archived, tc.wantArchived)
			}
		})
	}
}

// TestPageIsArchivedForGuard_FailsOpenAndSaysSo pins the posture: a read failure
// must NOT be reported as "not archived" alone, because the caller has to be able
// to log the window. checked=false is the whole point of the second return.
func TestPageIsArchivedForGuard_FailsOpenAndSaysSo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery(`SELECT COALESCE\(status, ''\)`).
		WithArgs(pageID).
		WillReturnError(errors.New("connection reset"))

	archived, checked := pageIsArchivedForGuard(context.Background(), db, pageID, zap.NewNop())
	if archived {
		t.Error("failed CLOSED on a read error — this would stop deploys fleet-wide on a hiccup")
	}
	if checked {
		t.Error("checked=true after a failed read — the fail-open window would be invisible")
	}
}

// ---------------------------------------------------------------------------
// Layer 2 — the commit seam. This is where the damage actually happens.
// ---------------------------------------------------------------------------

func gitCommitParams(db interface{ Close() error }, collected map[string]interface{}) ActionParams {
	return ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{
			// Short-circuits the sites lookup so the only DB traffic in these
			// tests is the guard's own query.
			"repo_name": "sites",
		}},
		CollectedData:    collected,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "deploy_page"},
	}
}

// TestGitCommit_ArchivedPageIsRefusedBeforeCommit is the central assertion of the
// whole fix.
func TestGitCommit_ArchivedPageIsRefusedBeforeCommit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery(`SELECT COALESCE\(status, ''\) FROM pages WHERE id = \$1`).
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("archived"))

	params := gitCommitParams(db, map[string]interface{}{
		"page_id":   pageID.String(),
		"page_name": "tool-llm-cost-calculator",
		"domain":    "fundamentallyai.com",
	})
	params.DB = db

	out, err := GitCommitAction(context.Background(), params)
	if err != nil {
		t.Fatalf("GitCommitAction returned an error; a refusal must not fail the loop: %v", err)
	}

	res, ok := out.(GitCommitResult)
	if !ok {
		t.Fatalf("unexpected result type %T", out)
	}
	if !res.Success {
		t.Error("Success=false — a refusal is not a failure, and the loop must continue")
	}
	if res.AwaitResponse {
		t.Error("AwaitResponse=true — the commit was dispatched to the git adapter anyway")
	}
	if got := res.Metadata["status"]; got != "skipped" {
		t.Errorf("status=%v, want skipped — the archived page was committed", got)
	}
	reason, _ := res.Metadata["skip_reason"].(string)
	if !strings.Contains(reason, archivedPageSkipReasonPrefix) {
		t.Errorf("skipped, but not by this guard: %q", reason)
	}
}

// TestGitCommit_LivePageStillDeploys — THE CONTROL. Without it every other test
// here is satisfied by a guard that refuses unconditionally.
//
// A live page must get PAST the guard. It then stops at "no files to commit",
// which is a different skip with a different reason and, importantly, is reached
// only after the guard has let it through.
func TestGitCommit_LivePageStillDeploys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery(`SELECT COALESCE\(status, ''\) FROM pages WHERE id = \$1`).
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))

	params := gitCommitParams(db, map[string]interface{}{
		"page_id":   pageID.String(),
		"page_name": "index",
		"domain":    "fundamentallyai.com",
	})
	params.DB = db

	out, err := GitCommitAction(context.Background(), params)
	if err != nil {
		t.Fatalf("GitCommitAction: %v", err)
	}
	res, ok := out.(GitCommitResult)
	if !ok {
		t.Fatalf("unexpected result type %T", out)
	}
	reason, _ := res.Metadata["skip_reason"].(string)
	if strings.Contains(reason, archivedPageSkipReasonPrefix) {
		t.Fatalf("a LIVE page was refused by the archived guard: %q", reason)
	}
	if !strings.Contains(reason, "no files") {
		t.Errorf("expected the live page to reach the no-files branch, got %q", reason)
	}
}

// TestGitCommit_NonPageCommitDoesNotTouchPages — most git_commit steps deploy no
// page at all (CSS, JS snippets, RSS, reports, whole-site). The guard must be
// completely invisible to them.
//
// Asserted by giving sqlmock NO query expectation: if the guard queries `pages`
// for a commit with no page in scope, that is an unexpected query and the test
// fails.
func TestGitCommit_NonPageCommitDoesNotTouchPages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	params := gitCommitParams(db, map[string]interface{}{
		"domain": "fundamentallyai.com",
	})
	params.DB = db

	if _, err := GitCommitAction(context.Background(), params); err != nil {
		t.Fatalf("GitCommitAction: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected DB traffic on a page-less commit: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Layer 3 — the stamp seam. Different damage: the row claiming a deploy that
// the commit guard just prevented.
// ---------------------------------------------------------------------------

// TestUpdatePageStatus_ArchivedPageIsNotStampedDeployed asserts the negative the
// only way that cannot be faked: no UPDATE expectation is registered, so any
// write is an unexpected query and fails the test.
func TestUpdatePageStatus_ArchivedPageIsNotStampedDeployed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery(`SELECT COALESCE\(status, ''\) FROM pages WHERE id = \$1`).
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("archived"))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{
			"status":        "deployed",
			"page_id_field": "page_id",
		}},
		CollectedData:    map[string]interface{}{"page_id": pageID.String()},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "update_status"},
	}

	out, err := UpdatePageStatusAction(context.Background(), params)
	if err != nil {
		t.Fatalf("UpdatePageStatusAction: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected result type %T", out)
	}
	if updated, _ := m["updated"].(bool); updated {
		t.Error("updated=true — the archived page was stamped deployed")
	}
	reason, _ := m["reason"].(string)
	if !strings.Contains(reason, archivedPageSkipReasonPrefix) {
		t.Errorf("refused, but not by this guard: %q", reason)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a write reached the database despite the refusal: %v", err)
	}
}

// TestUpdatePageStatus_ArchivedDoesNotBlockOtherStatuses — the guard is scoped to
// the DEPLOY stamp. Bookkeeping writes on an archived page (needs_rebuild, etc.)
// must still work, or ordinary maintenance of retired pages breaks.
func TestUpdatePageStatus_ArchivedDoesNotBlockOtherStatuses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	// No status read is expected at all: newStatus != "deployed" must not even
	// reach the guard.
	mock.ExpectExec(`UPDATE pages SET build_status = \$2`).
		WithArgs(pageID, "needs_rebuild").
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := ActionParams{
		StepConfig: models.Step{Config: map[string]interface{}{
			"status":        "needs_rebuild",
			"page_id_field": "page_id",
		}},
		CollectedData:    map[string]interface{}{"page_id": pageID.String()},
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "update_status"},
	}

	if _, err := UpdatePageStatusAction(context.Background(), params); err != nil {
		t.Fatalf("UpdatePageStatusAction: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("non-deploy status write did not go through unchanged: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Layer 4 — page resolution at the deploy seam
// ---------------------------------------------------------------------------

// TestResolveDeployTargetPage_FlatDispatchShape covers the shape the two newest
// producers actually arrive in (page-rerender and section-editor are dispatched
// with page_id/site_id/domain, NOT with current_page.*). If this regresses, the
// guard silently stops seeing exactly the producers it was built for.
func TestResolveDeployTargetPage_FlatDispatchShape(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	for _, tc := range []struct {
		name      string
		collected map[string]interface{}
	}{
		{"bare page_id", map[string]interface{}{"page_id": pageID.String()}},
		{"nested under input_data", map[string]interface{}{
			"input_data": map[string]interface{}{"page_id": pageID.String()},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _, ok := resolveDeployTargetPage(context.Background(), db, tc.collected, zap.NewNop())
			if !ok {
				t.Fatal("page not resolved — the guard would be invisible to this producer")
			}
			if got != pageID {
				t.Errorf("resolved %s, want %s", got, pageID)
			}
		})
	}
}

// TestResolveDeployTargetPage_NoPageIsNotAnError — the page-less commit case,
// which must resolve to ok=false rather than guessing.
func TestResolveDeployTargetPage_NoPageIsNotAnError(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	if _, _, ok := resolveDeployTargetPage(context.Background(), db,
		map[string]interface{}{"domain": "fundamentallyai.com"}, zap.NewNop()); ok {
		t.Error("resolved a page from a commit that names none")
	}
}
