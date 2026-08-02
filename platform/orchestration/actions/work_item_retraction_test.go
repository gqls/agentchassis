// FILE: platform/orchestration/actions/work_item_retraction_test.go
//
// Guards for the retraction seam (RFC_010, owner ruling 2026-08-02 Decision 1).

package actions

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
)

// TestResolveWorkItemsRefusesAnUnderspecifiedClaim pins the validation as a
// REFUSAL rather than a best guess.
//
// The wide branch closes every open item of a type for a site. Reaching it by
// leaving a field blank is exactly the failure the owner's 2026-08-02 ruling
// ("new authority on a shared seam ships as an opt-in field, unsafe default
// OFF") exists to prevent, so "no ItemKey" must NOT silently mean "all of them".
//
// Passing a nil *sql.Tx is deliberate and is itself an assertion: every case
// below must be rejected BEFORE the database is touched. If validation ever
// moves after the query, these panic instead of failing, which is a louder
// signal than a wrong count.
func TestResolveWorkItemsRefusesAnUnderspecifiedClaim(t *testing.T) {
	for _, c := range []struct {
		name string
		in   checks.ResolvedFinding
		want string
	}{
		{"no item type", checks.ResolvedFinding{Reason: "r", ItemKey: "k"}, "ItemType is empty"},
		{"no reason", checks.ResolvedFinding{ItemType: "t", ItemKey: "k"}, "Reason is empty"},
		{"neither key nor all", checks.ResolvedFinding{ItemType: "t", Reason: "r"}, "neither ItemKey nor AllOfType"},
		{"both key and all", checks.ResolvedFinding{ItemType: "t", Reason: "r", ItemKey: "k", AllOfType: true}, "pick one"},
	} {
		t.Run(c.name, func(t *testing.T) {
			n, err := resolveWorkItems(context.Background(), nil, uuid.New(), "some_check", uuid.New(), c.in, zap.NewNop())
			if err == nil {
				t.Fatalf("expected a refusal, got nil error (%d rows) — an underspecified retraction "+
					"must never be guessed at", n)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("error %q does not mention %q — the message is what tells the check author "+
					"which half of the claim is missing", err, c.want)
			}
		})
	}
}

// TestResolveWorkItemsClosesTheRightRows pins the SQL's three load-bearing
// predicates: the status set, the narrow/wide switch, and the batch guard.
func TestResolveWorkItemsClosesTheRightRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	site, batch := uuid.New(), uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs("backend_unreachable", "health recovered", site, "backend_unreachable", "", batch).
		WillReturnResult(sqlmock.NewResult(0, 3))
	tx, _ := db.Begin()

	n, err := resolveWorkItems(context.Background(), tx, site, "backend_unreachable", batch,
		checks.ResolvedFinding{ItemType: "backend_unreachable", AllOfType: true, Reason: "health recovered"},
		zap.NewNop())
	if err != nil {
		t.Fatalf("resolveWorkItems: %v", err)
	}
	if n != 3 {
		t.Errorf("resolved %d, want 3 — the caller counts what actually changed, not what it asked for", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query did not match: %v", err)
	}
}

// TestRetractionStatusSetIsTheTerminalSetMinusGaveUp is the lockstep guard
// between the two status vocabularies.
//
// They must differ, and differ by exactly `failed` and `unresolved` — the two
// "we gave up" states the owner's Decision 2 ruling makes retractable. If
// someone later "tidies" them into one list, retraction either stops reaching
// abandoned items (re-creating the landfill) or starts reopening settled ones.
func TestRetractionStatusSetIsTheTerminalSetMinusGaveUp(t *testing.T) {
	terminal := map[string]bool{}
	for _, s := range workItemTerminalStatuses {
		terminal[s] = true
	}
	closed := map[string]bool{}
	for _, s := range workItemClosedStatuses {
		closed[s] = true
	}

	for s := range closed {
		if !terminal[s] {
			t.Errorf("%q is in workItemClosedStatuses but not workItemTerminalStatuses — retraction "+
				"must never protect a status the dedup index treats as open", s)
		}
	}

	var onlyTerminal []string
	for s := range terminal {
		if !closed[s] {
			onlyTerminal = append(onlyTerminal, s)
		}
	}
	want := map[string]bool{"failed": true, "unresolved": true}
	if len(onlyTerminal) != len(want) {
		t.Fatalf("the two lists differ by %v, want exactly [failed unresolved] — that difference IS the "+
			"owner's Decision 2 ruling (RFC_010); changing it changes what a retraction may reach", onlyTerminal)
	}
	for _, s := range onlyTerminal {
		if !want[s] {
			t.Errorf("unexpected status %q retractable — see workItemClosedStatuses' comment", s)
		}
	}
}

// TestClosedStatusesNeverReachAnOnConflictClause is the 42P10 guard.
//
// Only workItemTerminalStatuses may be interpolated into an `ON CONFLICT …
// WHERE`, because only it matches idx_swi_dedup's predicate. Using the
// retraction list there would fail partial-index inference on EVERY keyed
// insert fleet-wide — the breakage migration 157 already caused once, recorded
// in workItemTerminalStatuses' own comment.
//
// A source scan, because the property is "these two things are never combined",
// which no value-level test can observe.
func TestClosedStatusesNeverReachAnOnConflictClause(t *testing.T) {
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	// Any ON CONFLICT whose formatting argument is the closed list.
	bad := regexp.MustCompile(`(?s)ON CONFLICT.{0,400}?sqlInList\(workItemClosedStatuses\)`)

	var scanned int
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".go") || strings.HasSuffix(f.Name(), "_test.go") {
			continue
		}
		src, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatalf("read %s: %v", f.Name(), err)
		}
		scanned++
		if bad.Match(src) {
			t.Errorf("%s interpolates workItemClosedStatuses into an ON CONFLICT clause.\n"+
				"Only workItemTerminalStatuses matches idx_swi_dedup's predicate; this fails partial-index "+
				"inference with SQLSTATE 42P10 on every keyed insert (see migration 157).", f.Name())
		}
	}
	if scanned == 0 {
		t.Fatal("scanned no files — this guard is vacuous, which is how it survives the thing it guards")
	}
}
