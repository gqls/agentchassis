// FILE: platform/orchestration/actions/invalid_banned_claim_pattern_test.go
//
// RFC_060 §1e/§3e (2026-09-02): claims.go:348 silently degrades a per-site
// banned_claims pattern that fails to compile to a literal match of its own
// source text — no logger, no error path — and nothing else in the estate
// was positioned to notice, since TestEveryGlobalPatternIsAValidRegex pins
// only the Go-authored fleet-wide set. This file pins the daily-refresh-loop
// detector added to close that: checkBannedClaimPatterns (pure) and the
// keyed-per-FINDING write path (createInvalidBannedClaimPatternItems).

package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestCheckBannedClaimPatternsFindsOnlyWhatFailsToCompile(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"pattern": `\bno credit checks?\b`, "reason": "predatory promise"},
		map[string]interface{}{"pattern": `guaranteed(`, "reason": "unbalanced paren — the invalid one"},
		map[string]interface{}{"pattern": `[apr`, "reason": "unterminated character class — the invalid one"},
	}

	got := checkBannedClaimPatterns(raw)
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 invalid patterns (index 1 and 2), got %d: %+v", len(got), got)
	}
	if got[0].Index != 1 || got[0].Pattern != "guaranteed(" {
		t.Fatalf("first invalid entry wrong: %+v", got[0])
	}
	if got[1].Index != 2 || got[1].Pattern != "[apr" {
		t.Fatalf("second invalid entry wrong: %+v", got[1])
	}
	for _, inv := range got {
		if inv.Error == "" {
			t.Fatalf("invalid pattern %q reported no compile error", inv.Pattern)
		}
	}
}

// TestCheckBannedClaimPatternsCleanRegisterIsNil is the discriminating
// control: a register where every pattern compiles must report NOTHING, or
// the detector is noise, not signal — exactly claims.go's own standard
// ("a scanner that always reports something is one people stop reading").
func TestCheckBannedClaimPatternsCleanRegisterIsNil(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"pattern": `\bno credit checks?\b`},
		map[string]interface{}{"pattern": `guaranteed acceptance`},
	}
	if got := checkBannedClaimPatterns(raw); len(got) != 0 {
		t.Fatalf("clean register reported findings: %+v", got)
	}
}

func TestCheckBannedClaimPatternsHandlesAbsentOrMalformedBannedClaims(t *testing.T) {
	if got := checkBannedClaimPatterns(nil); got != nil {
		t.Fatalf("nil banned_claims should report nothing, got %+v", got)
	}
	if got := checkBannedClaimPatterns("not an array"); got != nil {
		t.Fatalf("malformed banned_claims should report nothing (not panic), got %+v", got)
	}
}

// TestInvalidBannedClaimPatternKeyedPerFinding is the mechanism the whole
// point of this file rests on: a SECOND, DIFFERENT bad pattern must get its
// OWN item even while an earlier one for the SAME site is already open — the
// exact shape bugs_open/091 measured going wrong for the sibling
// stale_evidence item (per-site key, four of five open items naming the
// wrong fact). Proven here by asserting the two keys differ, which is what
// ON CONFLICT (site_id, item_key) actually dedups on.
func TestInvalidBannedClaimPatternKeyedPerFinding(t *testing.T) {
	siteID := uuid.New()
	keyA := bannedClaimPatternItemKey(siteID, "guaranteed(")
	keyB := bannedClaimPatternItemKey(siteID, "[apr]{2,")
	if keyA == keyB {
		t.Fatalf("two different bad patterns on the same site produced the SAME item_key — "+
			"the second finding would silently dedup against the first: %q", keyA)
	}
	// Same pattern, same site: must be STABLE (a re-run dedups against itself).
	if again := bannedClaimPatternItemKey(siteID, "guaranteed("); again != keyA {
		t.Fatalf("item_key is not stable across calls: %q vs %q", keyA, again)
	}
}

// TestInvalidBannedClaimPatternItemsUsesDropOnConflict pins the PRODUCTION
// write path via sqlmock against the real INSERT — DO NOTHING, never DO
// UPDATE. A daily re-write via DO UPDATE would bump updated_at on every
// pass, making the reaper that keys on it never reap this row
// (bugs_closed/213) — the finding is a standing defect until a human edits
// the pattern, not a value with a description that needs keeping current.
func TestInvalidBannedClaimPatternItemsUsesDropOnConflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	invalid := []invalidBannedClaimPattern{
		{Index: 0, Pattern: "guaranteed(", Error: "missing closing )"},
	}

	mock.ExpectBegin()
	// The load-bearing assertion: DO NOTHING, not DO UPDATE. sqlmock matches
	// the query by regex substring, so this fails loudly if the write path
	// is ever pointed at refreshOnConflict instead.
	mock.ExpectExec(`INSERT INTO site_work_items \(.*\)\s*ON CONFLICT \(site_id, item_key\)\s*WHERE.*DO NOTHING`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	created, err := createInvalidBannedClaimPatternItems(
		context.Background(), db, siteID, "example.co.uk", invalid, "evidence-refresher", zap.NewNop())
	if err != nil {
		t.Fatalf("createInvalidBannedClaimPatternItems: %v", err)
	}
	if created != 1 {
		t.Fatalf("expected 1 item created, got %d", created)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("production write path did not match the expected DO-NOTHING insert: %v", err)
	}
}

// TestInvalidBannedClaimPatternItemsHonoursExistingOpenItem is the control
// for the test above: when the row already exists (RowsAffected 0, the
// ON CONFLICT ... DO NOTHING branch), the function must report ZERO created,
// not error and not silently claim success — a second run finding the same
// bad pattern must not double-count.
func TestInvalidBannedClaimPatternItemsHonoursExistingOpenItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	invalid := []invalidBannedClaimPattern{
		{Index: 0, Pattern: "guaranteed(", Error: "missing closing )"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnResult(sqlmock.NewResult(0, 0)) // conflict: an open item already holds this key
	mock.ExpectCommit()

	created, err := createInvalidBannedClaimPatternItems(
		context.Background(), db, siteID, "example.co.uk", invalid, "evidence-refresher", zap.NewNop())
	if err != nil {
		t.Fatalf("createInvalidBannedClaimPatternItems: %v", err)
	}
	if created != 0 {
		t.Fatalf("an already-open item must report 0 newly created, got %d", created)
	}
	// The discriminating assertion: under dropOnConflict, a 0-rows INSERT is
	// the END of it — no further query. Under refreshOnConflict (the mutation
	// this file guards against), the conflict branch issues a SECOND query
	// (an UPDATE) this mock never set up, and sqlmock fails IT rather than
	// silently allowing it through — but only if we actually check.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected extra query after the conflict — the write path is not on dropOnConflict: %v", err)
	}
}
