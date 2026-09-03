// FILE: platform/orchestration/actions/work_item_retraction_test.go
//
// Guards for the retraction seam (RFC_010, owner ruling 2026-08-02 Decision 1).

package actions

import (
	"context"
	"database/sql"
	"errors"
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
// predicates by asserting they are IN THE QUERY TEXT, not merely that some
// UPDATE ran.
//
// CORRECTED after council round 1 (`846f4f3d`, editquality, medium). The first
// version matched only `"UPDATE site_work_items"` and checked the arguments,
// then the submission claimed it "pins the query's three load-bearing
// predicates". It did not: dropping the status filter, the narrow/wide switch or
// the batch guard from the SQL would all have left it green, because none of
// them is an argument. The seat was right, and this is the assertion the claim
// described.
func TestResolveWorkItemsClosesTheRightRows(t *testing.T) {
	// Every fragment whose removal changes which rows are closed.
	const wantSQL = `(?s)UPDATE site_work_items.*` +
		`status NOT IN \('complete','verified','rejected','wont_fix','cancelled'\).*` +
		`batch_id IS DISTINCT FROM`

	for _, c := range []struct {
		name    string
		in      checks.ResolvedFinding
		wantKey string
	}{
		{"wide", checks.ResolvedFinding{ItemType: "backend_unreachable", AllOfType: true, Reason: "recovered"}, ""},
		{"narrow", checks.ResolvedFinding{ItemType: "undeployed_asset", ItemKey: "undeployed_asset:abc", Reason: "serves 200"}, "undeployed_asset:abc"},
	} {
		t.Run(c.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock.New: %v", err)
			}
			defer db.Close()

			site, batch := uuid.New(), uuid.New()
			mock.ExpectBegin()
			mock.ExpectExec(wantSQL).
				WithArgs("chk", c.in.Reason, site, c.in.ItemType, c.wantKey, batch).
				WillReturnResult(sqlmock.NewResult(0, 3))
			tx, _ := db.Begin()

			n, err := resolveWorkItems(context.Background(), tx, site, "chk", batch, c.in, zap.NewNop())
			if err != nil {
				t.Fatalf("resolveWorkItems: %v", err)
			}
			if n != 3 {
				t.Errorf("resolved %d, want 3 — the caller counts what actually changed, not what it asked for", n)
			}
			// The narrow/wide switch is the ItemKey argument: empty means the
			// `$5 = '' OR item_key = $5` disjunct opens to every row of the type.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("query or args did not match: %v", err)
			}
		})
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

// ─── The receipt coupling (bugs_open/469) ───────────────────────────────────
//
// WHY THESE TESTS EXIST AT THE SEAM AND NOT IN THE CHECK. Resolved's safety
// property — a retraction fires only on a positive observation — is necessary
// and not sufficient. For a divergence check, "the finding no longer
// reproduces" and "the damage completed" can be the SAME observation: the
// stores agree again precisely because the build overwrote a human's edit. A
// close on that observation alone launders destruction into "resolved".
//
// ResolvedFinding.Receipt makes the record of what was destroyed a
// PRECONDITION of the close, in the same transaction. It lives here because
// resolveWorkItems has two callers and a control in either one protects only
// that one.
//
// A MOCK CANNOT ASSERT A NEGATIVE, so each of these was proven by MUTATION:
//
//	mutation in work_items_common.go                     test that must FAIL
//	--------------------------------------------------------------------------
//	drop the `return 0, err` on a receipt insert error   TestReceiptFailureWithholdsTheRetraction
//	skip the presence SELECT when !inserted             TestDedupedReceiptMustBeConfirmedPresent
//	write the receipt AFTER the UPDATE                  TestReceiptIsWrittenBeforeTheClose
//	append the evidence arg unconditionally             TestResolveWorkItemsClosesTheRightRows
//
// All four were applied and observed to fail on 2026-09-03 — but only after the
// FIRST one was caught surviving. Deleting the insert-error return left
// TestReceiptFailureWithholdsTheRetraction green, because the flow fell through
// to the `!inserted` arm whose unmocked presence SELECT errors too: a guard in
// SERIES did the work and the test could not tell. The safety property held
// throughout; the test's claim to pin that line did not. It now asserts the
// specific message. Recording it because "the mutation passed, so the line is
// fine" is the wrong reading, and it is the reading this table exists to stop.

// receiptFinding is a lossy retraction: the drift item may only close if the
// record of the destroyed section becomes durable first.
func receiptFinding() checks.ResolvedFinding {
	return checks.ResolvedFinding{
		ItemType: "section_source_drift",
		ItemKey:  "section_source_drift:gripper-catalog",
		Reason:   "stores agree again; the authority won",
		Evidence: map[string]interface{}{"direction": "authority_won", "lost_sections": []string{"gripper-spec-sheet"}},
		Receipt: &checks.WorkItemSpec{
			SiteID:       uuid.New(),
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "section_composition_lost",
			Severity:     "high",
			Summary:      "Composition LOST on page 'gripper-catalog'",
			SpecJSON:     `{"lost_sections":["gripper-spec-sheet"]}`,
			Status:       "needs_human_review",
			CreatedBy:    "completeness-discovery-agent",
			ItemKey:      "section_composition_lost:gripper-catalog:abc123",
			Priority:     30,
			HandlerAgent: "",
		},
	}
}

// TestReceiptIsWrittenBeforeTheClose. Order is the property, not an accident of
// reading order: a run that dies between the two writes must leave the finding
// OPEN, never the loss unrecorded. sqlmock's expectations are ordered, so an
// INSERT after the UPDATE fails here.
func TestReceiptIsWrittenBeforeTheClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	site, batch := uuid.New(), uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))
	tx, _ := db.Begin()

	n, err := resolveWorkItems(context.Background(), tx, site, "section_source_drift", batch, receiptFinding(), zap.NewNop())
	if err != nil {
		t.Fatalf("resolveWorkItems: %v", err)
	}
	if n != 1 {
		t.Errorf("resolved %d, want 1", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the receipt was not written before the close: %v", err)
	}
}

// TestReceiptFailureWithholdsTheRetraction is the whole point of the field. If
// the record of the destruction cannot be made durable, the finding must stay
// OPEN — a silent close here is bugs_open/469 automated.
func TestReceiptFailureWithholdsTheRetraction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	site, batch := uuid.New(), uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnError(errors.New("disk on fire"))
	// DELIBERATELY NO ExpectExec for the UPDATE. sqlmock fails an unexpected
	// call, so this is what makes "the close did not happen" assertable rather
	// than merely unobserved.
	tx, _ := db.Begin()

	n, err := resolveWorkItems(context.Background(), tx, site, "section_source_drift", batch, receiptFinding(), zap.NewNop())
	if err == nil {
		t.Fatal("the receipt could not be written and the retraction was applied anyway — a destroyed page just closed as 'resolved'")
	}
	if n != 0 {
		t.Errorf("resolved %d, want 0", n)
	}
	if !strings.Contains(err.Error(), "WITHHELD") {
		t.Errorf("error does not say the retraction was withheld, so a reader cannot tell a refusal from bad luck: %v", err)
	}
	// ⚠ THIS ASSERTION IS WHY THE TEST DISCRIMINATES, and it was added after the
	// mutation table caught the first version. Deleting the insert-error return
	// entirely left this test GREEN: the flow fell through to the `!inserted`
	// arm, whose presence SELECT is unmocked and errors too, so the retraction
	// was still withheld — by a guard in SERIES. The property held; the test's
	// claim to pin THIS line did not. Asserting on the specific message is what
	// separates the two.
	if !strings.Contains(err.Error(), "could not be written") {
		t.Errorf("the refusal did not come from the insert-error guard — a second guard in series may be doing the work,\n"+
			"which would leave this line free to be deleted with the test still green: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// TestDedupedReceiptMustBeConfirmedPresent. insertWorkItem reports FALSE for
// several different reasons — an open row already holds the key, the anti-churn
// brake held it back, the two-strike rule dropped it. Only the first means a
// durable record exists. Assuming it turns a DROPPED receipt into a silent
// close, which is exactly the hole this field closes.
func TestDedupedReceiptMustBeConfirmedPresent(t *testing.T) {
	t.Run("an open row already holds the key → close proceeds", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		site, batch := uuid.New(), uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery("SELECT 1 FROM site_work_items").
			WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))
		mock.ExpectExec("UPDATE site_work_items").WillReturnResult(sqlmock.NewResult(0, 1))
		tx, _ := db.Begin()

		if _, err := resolveWorkItems(context.Background(), tx, site, "section_source_drift", batch, receiptFinding(), zap.NewNop()); err != nil {
			t.Fatalf("a confirmed-present receipt should permit the close: %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})

	t.Run("nothing holds the key → the receipt was DROPPED → close withheld", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		site, batch := uuid.New(), uuid.New()
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery("SELECT 1 FROM site_work_items").WillReturnError(sql.ErrNoRows)
		// No UPDATE expectation: the close must not happen.
		tx, _ := db.Begin()

		_, err := resolveWorkItems(context.Background(), tx, site, "section_source_drift", batch, receiptFinding(), zap.NewNop())
		if err == nil {
			t.Fatal("the receipt was neither inserted nor present, and the retraction was applied anyway")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectations: %v", err)
		}
	})
}

// TestReceiptWithoutAnItemKeyIsRefused — a receipt with no key could never be
// found again, so "it is durable" would be unverifiable by construction.
func TestReceiptWithoutAnItemKeyIsRefused(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	site, batch := uuid.New(), uuid.New()
	mock.ExpectBegin()
	tx, _ := db.Begin()

	f := receiptFinding()
	f.Receipt.ItemKey = ""
	if _, err := resolveWorkItems(context.Background(), tx, site, "section_source_drift", batch, f, zap.NewNop()); err == nil {
		t.Fatal("a keyless receipt was accepted; nothing could ever confirm it")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// TestEvidenceIsAppendedOnlyWhenSet is the REGRESSION GUARD FOR OTHER LANES.
// Every existing caller's test asserts a six-argument UPDATE. If the evidence
// argument were appended unconditionally — even as NULL — every one of them
// would break, and the breakage would look like this change's fault in a lane
// that never asked for it.
func TestEvidenceIsAppendedOnlyWhenSet(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	site, batch := uuid.New(), uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs("chk", "recovered", site, "backend_unreachable", "", batch).
		WillReturnResult(sqlmock.NewResult(0, 1))
	tx, _ := db.Begin()

	in := checks.ResolvedFinding{ItemType: "backend_unreachable", AllOfType: true, Reason: "recovered"}
	if _, err := resolveWorkItems(context.Background(), tx, site, "chk", batch, in, zap.NewNop()); err != nil {
		t.Fatalf("resolveWorkItems: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a retraction with no Evidence must send exactly the six arguments it always has: %v", err)
	}
}

// TestEvidenceLandsInResultWhenSet — the other direction, so the conditional is
// pinned from both sides rather than only in its off position.
func TestEvidenceLandsInResultWhenSet(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	site, batch := uuid.New(), uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(`(?s)UPDATE site_work_items.*\|\| \$7::jsonb`).
		WithArgs("chk", "reason", site, "empty_section", "empty_section:x", batch,
			sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	tx, _ := db.Begin()

	in := checks.ResolvedFinding{
		ItemType: "empty_section", ItemKey: "empty_section:x", Reason: "reason",
		Evidence: map[string]interface{}{"direction": "cache_held"},
	}
	if _, err := resolveWorkItems(context.Background(), tx, site, "chk", batch, in, zap.NewNop()); err != nil {
		t.Fatalf("resolveWorkItems: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("evidence did not reach the result jsonb: %v", err)
	}
}
