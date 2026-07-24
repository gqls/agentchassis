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

// Covers the write-side lock gate (bugs_open/058): automated writers must not
// overwrite page_components rows carrying an active human lock, and the
// helpers that decide "active" must agree with the canonical predicate from
// the applied 053/115 schema migration.

// ---------------------------------------------------------------------------
// pageComponentAgentWritableSQL: the shared predicate fragment
// ---------------------------------------------------------------------------

func TestPageComponentAgentWritableSQL(t *testing.T) {
	bare := pageComponentAgentWritableSQL("")
	if strings.Contains(bare, ".") {
		t.Errorf("bare predicate should carry no alias dots: %s", bare)
	}
	for _, want := range []string{"locked_at IS NULL", "lock_type = 'timed'", "lock_expires_at < NOW()"} {
		if !strings.Contains(bare, want) {
			t.Errorf("predicate missing %q: %s", want, bare)
		}
	}

	aliased := pageComponentAgentWritableSQL("pc.")
	for _, want := range []string{"pc.locked_at", "pc.lock_type", "pc.lock_expires_at"} {
		if !strings.Contains(aliased, want) {
			t.Errorf("aliased predicate missing %q: %s", want, aliased)
		}
	}
}

// ---------------------------------------------------------------------------
// matchLockedRow: slot matching in the save_page_sections rebuild guard
// ---------------------------------------------------------------------------

func TestMatchLockedRowExactAndKebab(t *testing.T) {
	rows := []*lockedPageRow{
		{id: uuid.New(), slot: "hero-section"},
		{id: uuid.New(), slot: "social_proof"},
	}

	if lr := matchLockedRow(rows, "hero-section"); lr == nil || lr.slot != "hero-section" {
		t.Fatalf("exact match failed: %+v", lr)
	}
	// Kebab-normalised fallback (the 041 naming landmine): snake_case row,
	// kebab-case incoming name.
	if lr := matchLockedRow(rows, "social-proof"); lr == nil || lr.slot != "social_proof" {
		t.Fatalf("kebab-normalised match failed: %+v", lr)
	}
	if lr := matchLockedRow(rows, "faq"); lr != nil {
		t.Fatalf("unrelated name must not match, got %+v", lr)
	}
	if lr := matchLockedRow(rows, ""); lr != nil {
		t.Fatalf("empty name must not match, got %+v", lr)
	}
}

func TestMatchLockedRowConsumesOnce(t *testing.T) {
	rows := []*lockedPageRow{{id: uuid.New(), slot: "content-block"}}

	lr := matchLockedRow(rows, "content-block")
	if lr == nil {
		t.Fatal("first match failed")
	}
	lr.consumed = true

	// A page with duplicate slot names: one lock must not swallow the second
	// incoming section of the same name.
	if again := matchLockedRow(rows, "content-block"); again != nil {
		t.Fatalf("consumed row matched again: %+v", again)
	}
}

// ---------------------------------------------------------------------------
// CheckComponentLock: expiry-aware classification
// ---------------------------------------------------------------------------

func lockCols() []string {
	return []string{"locked_at", "locked_by", "lock_type", "lock_expires_at", "agent_writable"}
}

func TestCheckComponentLockPermanent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(lockCols()).
			AddRow(now, "182_legal_pages", "permanent", nil, false))

	lock, err := CheckComponentLock(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("CheckComponentLock: %v", err)
	}
	if !lock.IsLocked || !lock.IsHard {
		t.Errorf("permanent lock must be IsLocked+IsHard, got %+v", lock)
	}
	if lock.LockedBy != "182_legal_pages" {
		t.Errorf("LockedBy lost: %+v", lock)
	}
}

func TestCheckComponentLockUnstampedTypeIsHard(t *testing.T) {
	// A locked row with no lock_type predates the policy stamp (e.g. the old
	// admin handler): conservative — treat as hard. The old locked_by switch
	// would have called this soft unless locked_by was one of three strings.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(lockCols()).
			AddRow(time.Now(), "admin", nil, nil, false))

	lock, err := CheckComponentLock(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("CheckComponentLock: %v", err)
	}
	if !lock.IsLocked || !lock.IsHard {
		t.Errorf("unstamped lock must be IsLocked+IsHard, got %+v", lock)
	}
}

func TestCheckComponentLockExpiredTimedIsOpen(t *testing.T) {
	// agent_writable=true comes from the SQL predicate itself for an expired
	// timed lock; the helper must report it as not locked.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	past := time.Now().Add(-24 * time.Hour)
	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(lockCols()).
			AddRow(past.Add(-30*24*time.Hour), "deploy", "timed", past, true))

	lock, err := CheckComponentLock(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("CheckComponentLock: %v", err)
	}
	if lock.IsLocked || lock.IsHard {
		t.Errorf("expired timed lock must be open, got %+v", lock)
	}
}

func TestCheckComponentLockMissingRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT locked_at, locked_by, lock_type, lock_expires_at").
		WillReturnRows(sqlmock.NewRows(lockCols()))

	lock, err := CheckComponentLock(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("CheckComponentLock: %v", err)
	}
	if lock.IsLocked {
		t.Errorf("missing row must not read as locked: %+v", lock)
	}
}

// ---------------------------------------------------------------------------
// updatePageComponentAfterEdit: the race-free WHERE guard
// ---------------------------------------------------------------------------

func TestUpdatePageComponentAfterEditLockedRowRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The guarded UPDATE matches no row (lock predicate excluded it).
	mock.ExpectExec("UPDATE page_components").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<section>new</section>", nil)
	if !errors.Is(err, errComponentLocked) {
		t.Fatalf("want errComponentLocked, got %v", err)
	}
}

func TestUpdatePageComponentAfterEditUnlockedRowWritten(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE page_components").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := updatePageComponentAfterEdit(context.Background(), db, uuid.New(), "<section>new</section>", nil); err != nil {
		t.Fatalf("unlocked row must be written: %v", err)
	}
}

func TestUpdatePageComponentSwapLockedRowRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE page_components").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = updatePageComponentSwap(context.Background(), db, uuid.New(), uuid.New(),
		"new-slot", "<section>new</section>", map[string]interface{}{"k": "v"})
	if !errors.Is(err, errComponentLocked) {
		t.Fatalf("want errComponentLocked, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// loadActiveLockedRows: preload for the rebuild guard
// ---------------------------------------------------------------------------

func TestLoadActiveLockedRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	idA, idB := uuid.New(), uuid.New()
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slot_name", "position", "locked_by", "lock_type"}).
			AddRow(idA.String(), "hero-section", 1, "admin", "permanent").
			AddRow(idB.String(), "cta-band", 4, "182_legal_pages", "permanent"))

	rows := loadActiveLockedRows(context.Background(), db, uuid.New(), zap.NewNop())
	if len(rows) != 2 {
		t.Fatalf("want 2 locked rows, got %d", len(rows))
	}
	if rows[0].slot != "hero-section" || rows[0].position != 1 || rows[1].slot != "cta-band" {
		t.Errorf("rows mis-scanned: %+v %+v", rows[0], rows[1])
	}
}

func TestLoadActiveLockedRowsQueryFailureIsBestEffort(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnError(errors.New("boom"))

	if rows := loadActiveLockedRows(context.Background(), db, uuid.New(), zap.NewNop()); rows != nil {
		t.Fatalf("preload failure must return nil (save proceeds, DELETE predicate still protects), got %+v", rows)
	}
}
