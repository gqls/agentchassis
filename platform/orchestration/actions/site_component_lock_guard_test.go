package actions

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Covers the CHROME write-side lock gate (bugs_open/069): automated writers
// must not overwrite site_components rows carrying an active human lock, the
// refusal must be race-free (the predicate is in the WHERE, not only in a
// pre-check), and a blocked change must surface as a work item rather than a
// silent skip.
//
// Helper names are deliberately distinct from lock_gate_test.go's — same
// package, so a shared name would collide, and depending on that file's
// helpers would couple this test to a file other sessions are editing.

func chromeLockCols() []string {
	return []string{"locked_at", "locked_by", "lock_type", "lock_expires_at",
		"agent_writable", "component_id", "id", "has_html"}
}

// ---------------------------------------------------------------------------
// The shared predicate, aliased for the ON CONFLICT ... WHERE clauses
// ---------------------------------------------------------------------------

func TestSiteComponentAgentWritablePredicateQualified(t *testing.T) {
	// Bare column names in an ON CONFLICT DO UPDATE ... WHERE are ambiguous
	// against EXCLUDED, so the chrome upserts must qualify the predicate.
	q := pageComponentAgentWritableSQL("site_components.")
	for _, want := range []string{
		"site_components.locked_at IS NULL",
		"site_components.lock_type = 'timed'",
		"site_components.lock_expires_at IS NOT NULL",
		"site_components.lock_expires_at < NOW()",
	} {
		if !strings.Contains(q, want) {
			t.Errorf("qualified predicate missing %q: %s", want, q)
		}
	}
}

func TestChromeLockItemKeyNamesTheSurface(t *testing.T) {
	// Not "…:chrome:<slot>": the page-side key is lock_blocked_change:<page>:<slot>,
	// so a page named "chrome" would dedup against the header slot.
	if got := chromeLockItemKey("header"); got != "lock_blocked_change:site_component:header" {
		t.Errorf("unexpected chrome item key: %s", got)
	}
}

// ---------------------------------------------------------------------------
// CheckSiteComponentLock — expiry-aware classification, keyed on site+slot
// ---------------------------------------------------------------------------

func TestCheckSiteComponentLockPermanentIsHard(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	compID := uuid.New()
	rowID := uuid.New()
	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()).
			AddRow(now, "069-verify", "permanent", nil, false, compID.String(), rowID.String(), true))

	lock, err := CheckSiteComponentLock(context.Background(), db, uuid.New(), "header", zap.NewNop())
	if err != nil {
		t.Fatalf("CheckSiteComponentLock: %v", err)
	}
	if !lock.IsLocked || !lock.IsHard {
		t.Errorf("permanent lock must be IsLocked+IsHard, got %+v", lock)
	}
	if !lock.RowExists || !lock.HasHTML {
		t.Errorf("row facts lost: RowExists=%v HasHTML=%v", lock.RowExists, lock.HasHTML)
	}
	if !lock.ComponentID.Valid || lock.ComponentID.UUID != compID {
		t.Errorf("component id lost: %+v", lock.ComponentID)
	}
	if lock.LockedBy != "069-verify" {
		t.Errorf("LockedBy lost: %+v", lock)
	}
}

func TestCheckSiteComponentLockUnstampedTypeIsHard(t *testing.T) {
	// A locked row with no lock_type predates the policy stamp: treat as hard,
	// conservatively — never silently overwrite what we cannot classify.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()).
			AddRow(time.Now(), "182_legal_pages", nil, nil, false, nil, uuid.New().String(), true))

	lock, err := CheckSiteComponentLock(context.Background(), db, uuid.New(), "footer", zap.NewNop())
	if err != nil {
		t.Fatalf("CheckSiteComponentLock: %v", err)
	}
	if !lock.IsLocked || !lock.IsHard {
		t.Errorf("unstamped lock must be treated as hard, got %+v", lock)
	}
	if lock.ComponentID.Valid {
		t.Errorf("NULL component_id must scan as invalid, got %+v", lock.ComponentID)
	}
}

func TestCheckSiteComponentLockExpiredTimedIsWritable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	past := time.Now().Add(-24 * time.Hour)
	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()).
			AddRow(past, "deploy", "timed", past, true, nil, uuid.New().String(), true))

	lock, err := CheckSiteComponentLock(context.Background(), db, uuid.New(), "head", zap.NewNop())
	if err != nil {
		t.Fatalf("CheckSiteComponentLock: %v", err)
	}
	if lock.IsLocked {
		t.Errorf("an expired timed lock must not block automation, got %+v", lock)
	}
	if !lock.RowExists {
		t.Errorf("RowExists must be true for an existing row: %+v", lock)
	}
}

func TestCheckSiteComponentLockMissingRow(t *testing.T) {
	// A slot with no row yet: not locked, and RowExists=false so the caller can
	// tell "about to be created" from "exists and writable".
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()))

	lock, err := CheckSiteComponentLock(context.Background(), db, uuid.New(), "header", zap.NewNop())
	if err != nil {
		t.Fatalf("CheckSiteComponentLock: %v", err)
	}
	if lock.IsLocked || lock.RowExists || lock.HasHTML {
		t.Errorf("missing row must report empty status, got %+v", lock)
	}
}

// ---------------------------------------------------------------------------
// The guarded write seams
// ---------------------------------------------------------------------------

func TestSetSiteComponentHTMLLockedRowRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE site_components SET rendered_html").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = setSiteComponentHTML(context.Background(), db, uuid.New(), "header", "<header/>")
	if !errors.Is(err, errSiteComponentLocked) {
		t.Fatalf("expected errSiteComponentLocked, got %v", err)
	}
}

func TestSetSiteComponentHTMLUnlockedRowWritten(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE site_components SET rendered_html").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := setSiteComponentHTML(context.Background(), db, uuid.New(), "header", "<header/>"); err != nil {
		t.Fatalf("unlocked write must succeed, got %v", err)
	}
}

func TestAppendSiteComponentHTMLLockedRowRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE site_components SET rendered_html = rendered_html").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = appendSiteComponentHTML(context.Background(), db, uuid.New(), "footer", "<style/>")
	if !errors.Is(err, errSiteComponentLocked) {
		t.Fatalf("expected errSiteComponentLocked, got %v", err)
	}
}

func TestRelinkSiteComponentReportsWhetherAnythingChanged(t *testing.T) {
	// 0 rows is NOT an error here: the upsert's IS DISTINCT FROM guard makes
	// "already linked correctly" the normal outcome. Only the caller's lock
	// read can tell that from a refusal — which is why it does one.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO site_components").WillReturnResult(sqlmock.NewResult(0, 0))
	changed, err := relinkSiteComponent(context.Background(), db, uuid.New(), "header", uuid.New())
	if err != nil || changed {
		t.Fatalf("no-op relink must report changed=false, no error: changed=%v err=%v", changed, err)
	}

	mock.ExpectExec("INSERT INTO site_components").WillReturnResult(sqlmock.NewResult(0, 1))
	changed, err = relinkSiteComponent(context.Background(), db, uuid.New(), "header", uuid.New())
	if err != nil || !changed {
		t.Fatalf("real relink must report changed=true, no error: changed=%v err=%v", changed, err)
	}
}

// ---------------------------------------------------------------------------
// renderAndStoreSiteComponent — where the gate sits, and what it costs
// ---------------------------------------------------------------------------

// expectChromeLockItem sets up the work-item emit sequence
// (BeginTx -> two-strike count -> INSERT -> Commit) used by
// emitChromeLockBlockedChangeItem.
func expectChromeLockItem(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func TestRenderLockedSlotIsNotRewritten(t *testing.T) {
	// The whole point of 069: a forced re-render of a locked slot must issue NO
	// write. If the gate were removed or moved after the template lookup, the
	// mock would be asked for queries it does not expect and the returns would
	// change.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()).
			AddRow(time.Now(), "069-verify", "permanent", nil, false, nil, uuid.New().String(), true))
	expectChromeLockItem(mock)

	ok, locked := renderAndStoreSiteComponent(context.Background(), db, uuid.New(), "header", nil, true, zap.NewNop())
	if !locked {
		t.Errorf("a locked slot must be reported as locked")
	}
	if !ok {
		t.Errorf("a locked slot that still HOLDS chrome is serving, so ok must be true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected or missing queries: %v", err)
	}
}

func TestRenderLockedButEmptySlotReportsNotServing(t *testing.T) {
	// A lock over an empty slot preserves nothing and leaves the site with no
	// chrome in that slot — honest reporting matters more here than anywhere.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()).
			AddRow(time.Now(), "069-verify", "permanent", nil, false, nil, uuid.New().String(), false))
	expectChromeLockItem(mock)

	ok, locked := renderAndStoreSiteComponent(context.Background(), db, uuid.New(), "footer", nil, true, zap.NewNop())
	if !locked || ok {
		t.Errorf("locked+empty must report locked=true, ok=false; got ok=%v locked=%v", ok, locked)
	}
}

func TestRenderUnforcedExitsBeforeTheLockCheck(t *testing.T) {
	// Regression guard for the placement decision: the lock check MUST stay
	// below the !force idempotence exit. Above it, every ordinary build of a
	// site with a locked slot would file a work item claiming a writer wanted
	// to change something, for a call that was never going to write.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT EXISTS").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	ok, locked := renderAndStoreSiteComponent(context.Background(), db, uuid.New(), "header", nil, false, zap.NewNop())
	if !ok || locked {
		t.Errorf("an already-rendered slot must short-circuit as ok, not locked: ok=%v locked=%v", ok, locked)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the unforced path must issue only the EXISTS query: %v", err)
	}
}

// ---------------------------------------------------------------------------
// fix_component_template's shared skip
// ---------------------------------------------------------------------------

func TestChromeFixLockSkipSpeaksNeedsReview(t *testing.T) {
	// fixed:false alone lets the dispatch loop record the item as done;
	// action:needs_review is what stops it, and the two-strike rule from
	// parking a re-detected item 'unresolved' two cycles later.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()).
			AddRow(time.Now(), "069-verify", "permanent", nil, false, nil, uuid.New().String(), true))
	expectChromeLockItem(mock)

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	skip := chromeFixLockSkip(context.Background(), params, uuid.New(), "header", "remove_element", zap.NewNop())
	if skip == nil {
		t.Fatal("a locked slot must produce a skip-result")
	}
	if skip["fixed"] != false || skip["locked"] != true || skip["action"] != "needs_review" {
		t.Errorf("unexpected skip-result shape: %+v", skip)
	}
}

func TestChromeFixLockSkipProceedsWhenUnlocked(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(chromeLockCols()).
			AddRow(nil, nil, nil, nil, true, nil, uuid.New().String(), true))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	if skip := chromeFixLockSkip(context.Background(), params, uuid.New(), "header", "responsive_fix", zap.NewNop()); skip != nil {
		t.Errorf("an unlocked slot must not be skipped, got %+v", skip)
	}
}
