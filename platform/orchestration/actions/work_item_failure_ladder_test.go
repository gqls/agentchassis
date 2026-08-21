package actions

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Tests for the work-item failure-write contract (bugs_open/307).
//
// THESE ARE THE FIRST TESTS THAT HAVE EVER DRIVEN THIS PATH. Before this file,
// `grep -rn "fail_work_item\|FailWorkItem" --include=*_test.go` returned two
// prose mentions and no test: the retry ladder, the status_override branch and
// the count-free AI branch were entirely unpinned while carrying every dispatch
// loop in the fleet. That absence is why the three defects in 307 could coexist
// for months.
//
// The suite is built around bug §5's acceptance shape, including its
// disconfirming case: a lone permanent failure must STILL reach `failed` in
// three attempts. A test suite that only proves items survive would pass on an
// implementation that never fails anything, which is the worse bug.

// ── sqlmock plumbing ────────────────────────────────────────────────────────

// ladderState is the row the pre-read returns.
type ladderState struct {
	status            string
	attemptCount      int
	maxAttempts       int
	itemType          string
	transientReleases int
}

func defaultLadderState() ladderState {
	return ladderState{status: "claimed", attemptCount: 0, maxAttempts: 3, itemType: "content_rewrite"}
}

func expectStateRead(mock sqlmock.Sqlmock, st ladderState) {
	mock.ExpectQuery(`SELECT status, attempt_count, max_attempts, item_type`).
		WillReturnRows(sqlmock.NewRows([]string{"status", "attempt_count", "max_attempts", "item_type", "releases"}).
			AddRow(st.status, st.attemptCount, st.maxAttempts, st.itemType, st.transientReleases))
}

// expectBurstProbe arms the agent_error_log read. Passing counts below the
// thresholds is how a test says "the fleet is healthy"; passing counts above
// them is how it says "an outage is in progress".
func expectBurstProbe(mock sqlmock.Sqlmock, errs, domains, types int) {
	mock.ExpectQuery(`FROM agent_error_log`).
		WillReturnRows(sqlmock.NewRows([]string{"errs", "domains", "types"}).AddRow(errs, domains, types))
}

func expectPolicyLookup(mock sqlmock.Sqlmock, backoffMinutes int) {
	mock.ExpectQuery(`FROM reaper_policies`).
		WillReturnRows(sqlmock.NewRows([]string{"backoff_minutes"}).AddRow(backoffMinutes))
}

// ladderWrite is what the UPDATE actually wrote, captured off the driver.
type ladderWrite struct {
	sql         string
	errorArg    string
	agentArg    string
	backoffArg  int64
	newStatus   string
	rowsMatched bool
}

// captureSQL records the statement text so a test can assert on the clauses
// that are the whole point (the guard, the CASE, the retry_after stamp) —
// asserting only on arguments would pass against a statement with the guard
// deleted, which is the mutation this file exists to catch.
type sqlCapture struct{ got *string }

func (c sqlCapture) Match(v driver.Value) bool { return true }

func expectLadderUpdate(mock sqlmock.Sqlmock, w *ladderWrite, newStatus string, attemptsLeft int, matched bool) {
	rows := sqlmock.NewRows([]string{"status", "attempts_left"})
	if matched {
		rows = rows.AddRow(newStatus, attemptsLeft)
	}
	mock.ExpectQuery(`UPDATE site_work_items`).
		WithArgs(sqlmock.AnyArg(), captureArg{got: &w.errorArg}, captureArg{got: &w.agentArg},
			sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)
}

// runLadder drives the helper against sqlmock. arm lets a test add its own
// expectations between the state read and the write.
func runLadder(t *testing.T, st ladderState, errMsg string, arm func(sqlmock.Sqlmock, *ladderWrite)) (failureLadderOutcome, *ladderWrite, error) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	w := &ladderWrite{}
	expectStateRead(mock, st)
	arm(mock, w)

	out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
		uuid.New(), errMsg, "build-dispatch-loop", nil)
	if mErr := mock.ExpectationsWereMet(); mErr != nil {
		t.Fatalf("expectations: %v", mErr)
	}
	return out, w, err
}

// ── The ladder: an attempt is counted, and only the LAST one is terminal ─────

func TestFailureLadder_CountsAttemptsAndOnlyTerminatesAtTheCeiling(t *testing.T) {
	// This is defect (b) of bugs_open/307, and the larger half of it: 141 of
	// 270 failed rows in the 14 days to 2026-08-19 died BEFORE exhausting
	// their budget, 139 of them on their first attempt of three, because
	// update_work_item_status wrote `failed` with no CASE at all.
	cases := []struct {
		name         string
		attemptCount int
		maxAttempts  int
		wantStatus   string
		wantBackoff  bool
	}{
		{"first of three — back to the queue, with a wait", 0, 3, "triaged", true},
		{"second of three — back to the queue, longer wait", 1, 3, "triaged", true},
		{"last of three — terminal, no wait to stamp", 2, 3, "failed", false},
		{"a one-shot lane is terminal on its only attempt", 0, 1, "failed", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := defaultLadderState()
			st.attemptCount, st.maxAttempts = tc.attemptCount, tc.maxAttempts
			out, _, err := runLadder(t, st, "save_page_sections: shrink guard refused", func(m sqlmock.Sqlmock, w *ladderWrite) {
				if tc.maxAttempts > 1 {
					expectBurstProbe(m, 1, 1, 1) // healthy fleet
				}
				if tc.wantBackoff {
					expectPolicyLookup(m, 30)
				}
				expectLadderUpdate(m, w, tc.wantStatus, tc.maxAttempts-tc.attemptCount-1, true)
			})
			if err != nil {
				t.Fatalf("ladder: %v", err)
			}
			if out.NewStatus != tc.wantStatus {
				t.Errorf("status = %q, want %q", out.NewStatus, tc.wantStatus)
			}
			if out.Released {
				t.Error("a plain failure must consume its attempt, not be released")
			}
			// The wait is the whole of defect (a): without it three attempts
			// land inside one outage, which is what killed 88 of 100 items.
			if tc.wantBackoff && out.BackoffMins <= 0 {
				t.Errorf("no backoff stamped on a non-terminal attempt — retry-without-delay is equivalent to no retry")
			}
			if !tc.wantBackoff && out.BackoffMins != 0 {
				t.Errorf("backoff %d stamped on a terminal attempt", out.BackoffMins)
			}
		})
	}
}

func TestFailureLadder_BackoffGrowsWithTheAttemptAndComesFromPolicy(t *testing.T) {
	// The numbers come from reaper_policies (RFC_018/SCH-024), scaled linearly
	// by attempt exactly as that mechanism's first consumer does. A literal
	// here would be the third hand-rolled copy of a shared mechanism.
	for _, tc := range []struct {
		attempt     int
		policyMins  int
		wantBackoff int
	}{
		{0, 30, 30},
		{1, 30, 60},
		{0, 5, 5},
		{1, 5, 10},
	} {
		t.Run(fmt.Sprintf("attempt_%d_policy_%d", tc.attempt, tc.policyMins), func(t *testing.T) {
			st := defaultLadderState()
			st.attemptCount = tc.attempt
			out, _, err := runLadder(t, st, "adapter said no", func(m sqlmock.Sqlmock, w *ladderWrite) {
				expectBurstProbe(m, 1, 1, 1)
				expectPolicyLookup(m, tc.policyMins)
				expectLadderUpdate(m, w, "triaged", 1, true)
			})
			if err != nil {
				t.Fatalf("ladder: %v", err)
			}
			if out.BackoffMins != tc.wantBackoff {
				t.Errorf("backoff = %d, want %d (policy %d × attempt %d)",
					out.BackoffMins, tc.wantBackoff, tc.policyMins, tc.attempt+1)
			}
		})
	}
}

func TestFailureLadder_FallsBackToTheCodeDefaultWhenNoPolicyRowExists(t *testing.T) {
	// reaper_policies is operator-owned config; a missing row must not make the
	// failure path fail, and must not silently mean "no wait".
	st := defaultLadderState()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	expectStateRead(mock, st)
	expectBurstProbe(mock, 1, 1, 1)
	mock.ExpectQuery(`FROM reaper_policies`).WillReturnError(errors.New("relation \"reaper_policies\" does not exist"))
	w := &ladderWrite{}
	expectLadderUpdate(mock, w, "triaged", 2, true)

	out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
		uuid.New(), "boom", "build-dispatch-loop", nil)
	if err != nil {
		t.Fatalf("a missing policy table must not break the failure path: %v", err)
	}
	if out.BackoffMins != defaultBackoffMinutes {
		t.Errorf("backoff = %d, want the code default %d", out.BackoffMins, defaultBackoffMinutes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

// ── The guard: a deliberate DECISION is never overwritten ───────────────────

func TestFailureLadder_GuardsEveryDecisionStatus(t *testing.T) {
	// Defect (c). CompleteWorkItemAction and failUnverifiedCompletion both
	// refuse to overwrite these; the failure path did not, so a handler's
	// `needs_human_review` or `wont_fix` could be replaced by triaged/failed —
	// and triaged means re-dispatched to be refused again.
	//
	// The write reports 0 rows because the guard matched. Asserting on the
	// OUTCOME (skipped, not an error) is the half that matters to a caller: a
	// saga that fails after a handler decided must not look like a crash.
	for _, status := range workItemDecisionStatuses {
		t.Run(status, func(t *testing.T) {
			st := defaultLadderState()
			st.status = status
			out, _, err := runLadder(t, st, "saga fell over on the way out", func(m sqlmock.Sqlmock, w *ladderWrite) {
				expectBurstProbe(m, 1, 1, 1)
				expectPolicyLookup(m, 30)
				expectLadderUpdate(m, w, "", 0, false) // guard matched: no row
			})
			if err != nil {
				t.Fatalf("a refused write is not an error: %v", err)
			}
			if !out.Skipped {
				t.Errorf("status %q was overwritten — this is the defect", status)
			}
			if !strings.Contains(out.SkipReason, status) {
				t.Errorf("skip reason %q does not name the status it preserved", out.SkipReason)
			}
		})
	}
}

func TestFailureLadder_GuardListIsNotTheCompletionPathList(t *testing.T) {
	// THE DIFFERENCE IS THE POINT, and copying the sibling list verbatim is the
	// naive move the §third-gap contribution to 307 explicitly warned against.
	// `failed` and `unresolved` are guarded on the COMPLETION path and must NOT
	// be guarded here: overwriting `failed` IS the retry ladder (failed →
	// triaged), and the admin retry door depends on it.
	for _, mustBeOverwritable := range []string{"failed", "unresolved", "triaged", "claimed", "detected"} {
		for _, guarded := range workItemDecisionStatuses {
			if guarded == mustBeOverwritable {
				t.Fatalf("%q is in the failure-path guard list, which would break the retry ladder", mustBeOverwritable)
			}
		}
	}
	// And the decision statuses must all actually be there — a shortened list
	// is the mutation this pins.
	for _, want := range []string{"needs_human_review", "wont_fix", "rejected", "verified", "blocked", "cancelled", "deferred"} {
		found := false
		for _, s := range workItemDecisionStatuses {
			if s == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q is missing from the failure-path guard list", want)
		}
	}
}

func TestFailureLadder_TheGuardIsActuallyInTheStatement(t *testing.T) {
	// A mock's own bookkeeping cannot assert a negative: the test above passes
	// against a statement with no guard at all, because the mock decides how
	// many rows come back. So read the SQL and require the clause.
	st := defaultLadderState()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	expectStateRead(mock, st)
	expectBurstProbe(mock, 1, 1, 1)
	expectPolicyLookup(mock, 30)

	var gotSQL string
	mock.ExpectQuery(`UPDATE site_work_items`).
		WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow("triaged", 2))
	_, _ = applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
		uuid.New(), "boom", "build-dispatch-loop", nil)
	_ = gotSQL

	// sqlmock's regexp matching is the readable way to assert on statement
	// text: an expectation that cannot match is a failed expectation.
	db2, mock2, _ := sqlmock.New()
	defer db2.Close()
	expectStateRead(mock2, st)
	expectBurstProbe(mock2, 1, 1, 1)
	expectPolicyLookup(mock2, 30)
	// Derived from the constant, NOT hard-coded: this test's job is "the clause is
	// IN the statement", and a literal here goes stale the moment the vocabulary
	// changes (it did — `deferred` was added answering a round-2 advisory, and
	// this line failed rather than the guard). MEMBERSHIP is pinned separately by
	// TestGuardLists_TheFailureAndCompletionListsDifferByExactlyTwo, which is the
	// test that must fail if someone shortens the list.
	mock2.ExpectQuery(regexp.QuoteMeta(`AND status NOT IN (` + sqlInList(workItemDecisionStatuses) + `)`)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow("triaged", 2))
	if _, err := applyWorkItemFailureLadder(context.Background(), db2, zap.NewNop(),
		uuid.New(), "boom", "build-dispatch-loop", nil); err != nil {
		t.Fatalf("the ladder UPDATE does not carry the decision-status guard: %v", err)
	}
	if err := mock2.ExpectationsWereMet(); err != nil {
		t.Fatalf("guard clause absent from the statement: %v", err)
	}
}

func TestFailureLadder_RetryAfterIsActuallyStamped(t *testing.T) {
	// Same reasoning as the guard: the backoff must be IN the statement, or the
	// outcome's BackoffMins is a number the database never saw.
	st := defaultLadderState()
	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectStateRead(mock, st)
	expectBurstProbe(mock, 1, 1, 1)
	expectPolicyLookup(mock, 30)
	mock.ExpectQuery(regexp.QuoteMeta(`retry_after = CASE WHEN attempt_count + 1 >= max_attempts OR $4::int <= 0 THEN NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow("triaged", 2))
	if _, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
		uuid.New(), "boom", "build-dispatch-loop", nil); err != nil {
		t.Fatalf("the ladder UPDATE does not stamp retry_after: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("retry_after clause absent: %v", err)
	}
}

// ── The transient release: back to the queue, attempt NOT consumed ──────────

func TestFailureLadder_BurstReleasesWithoutConsumingAnAttempt(t *testing.T) {
	// bug §5(a): while the burst holds, no item reaches `failed`. Note the
	// fixture is the item's LAST attempt — the case that would otherwise be
	// terminal — because that is the one the outage actually killed.
	st := defaultLadderState()
	st.attemptCount, st.maxAttempts = 2, 3

	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectStateRead(mock, st)
	expectBurstProbe(mock, 34, 5, 7) // the 2026-08-17 git outage's own shape
	mock.ExpectQuery(`UPDATE site_work_items`).
		WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow("triaged", 1))

	out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(), uuid.New(),
		`failed to get latest commit/base tree for branch "master": github API request failed with status: 404 Not Found`,
		"build-dispatch-loop", nil)
	if err != nil {
		t.Fatalf("ladder: %v", err)
	}
	if !out.Released {
		t.Fatal("an item failing inside a fleet-wide burst must be released, not charged an attempt (owner ruling 2026-08-18)")
	}
	if out.NewStatus != "triaged" {
		t.Errorf("status = %q, want triaged — 'returned to the queue' is the ruling's own wording", out.NewStatus)
	}
	if out.ReleaseReason != "burst" {
		t.Errorf("release reason = %q, want burst", out.ReleaseReason)
	}
	// Never count-free AND delay-free: that combination is bug §3's trap.
	if out.BackoffMins <= 0 {
		t.Error("released with no cooldown — it would be re-claimed straight back into the outage")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestFailureLadder_ALonePermanentFailureStillDiesInThreeAttempts(t *testing.T) {
	// bug §5(c) — THE DISCONFIRMING CASE, and the reason burst detection is
	// safe at all. A deleted repository produces error text byte-identical to
	// an outage's (compare bugs_open/131), so the text cannot be the signal. It
	// fails ALONE: one domain, one agent type, few errors. It must still exhaust
	// its budget honestly.
	st := defaultLadderState()
	st.attemptCount, st.maxAttempts = 2, 3

	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectStateRead(mock, st)
	expectBurstProbe(mock, 3, 1, 1) // same error, but only this item's own site
	w := &ladderWrite{}
	expectLadderUpdate(mock, w, "failed", 0, true)

	out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(), uuid.New(),
		`failed to get latest commit/base tree for branch "master": github API request failed with status: 404 Not Found`,
		"build-dispatch-loop", nil)
	if err != nil {
		t.Fatalf("ladder: %v", err)
	}
	if out.Released {
		t.Fatal("a lone permanent 404 was released — an infinite retry on a deleted repo is bug §3's named trap")
	}
	if out.NewStatus != "failed" {
		t.Errorf("status = %q, want failed — the honest end of three attempts", out.NewStatus)
	}
}

func TestFailureLadder_ReleaseIsCappedSoItCanNeverBeInfinite(t *testing.T) {
	// The pre-existing count-free branch retried FOREVER. Tolerable for LLM
	// credit; not for anything ambiguous. Past the cap, a burst stops earning
	// free attempts and the ladder takes over.
	st := defaultLadderState()
	st.transientReleases = defaultTransientReleaseCap // already at the ceiling
	st.attemptCount, st.maxAttempts = 2, 3

	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectStateRead(mock, st)
	// No burst probe at all: the cap short-circuits before the classifier runs.
	w := &ladderWrite{}
	expectLadderUpdate(mock, w, "failed", 0, true)

	out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(), uuid.New(),
		"github API request failed with status: 404 Not Found", "build-dispatch-loop", nil)
	if err != nil {
		t.Fatalf("ladder: %v", err)
	}
	if out.Released {
		t.Fatalf("released for the %dth time — the cap is not enforced", defaultTransientReleaseCap+1)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestFailureLadder_AOneShotLaneIsNeverReleased(t *testing.T) {
	// 67 live needs_diagnosis items run max_attempts=1 DELIBERATELY: a 26-minute
	// LLM diagnosis must not silently auto-retry, and WRONG_CALLS L1693 records
	// a session calling that value "a safe one-liner" to raise. A retry
	// mechanism that helpfully retries them breaks the lane.
	st := defaultLadderState()
	st.maxAttempts, st.itemType = 1, "needs_diagnosis"

	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectStateRead(mock, st)
	// No burst probe: max_attempts == 1 short-circuits first.
	w := &ladderWrite{}
	expectLadderUpdate(mock, w, "failed", 0, true)

	out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(), uuid.New(),
		"connection refused", "diagnose-dispatch-loop", nil)
	if err != nil {
		t.Fatalf("ladder: %v", err)
	}
	if out.Released {
		t.Fatal("a one-shot item was released for retry — max_attempts=1 is load-bearing, not a data error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestFailureLadder_ClassifiersAreLayeredNotSwapped(t *testing.T) {
	// isAIUnavailable and RetryDisposition's needle lists disagree in BOTH
	// directions, so both are asked. These fixtures are needles that exist in
	// exactly one of the two lists: swapping one classifier for the other (the
	// tempting "consolidation") makes one column of this table fail.
	cases := []struct {
		name        string
		errMsg      string
		wantRelease bool
		wantReason  string
	}{
		{"only isAIUnavailable knows a 402", "API request failed with status 402: out of credit", true, "ai_unavailable"},
		{"only isAIUnavailable knows a bare EOF", "read tcp: EOF", true, "ai_unavailable"},
		{"only isAIUnavailable knows an api key fault", "invalid api key supplied", true, "ai_unavailable"},
		{"only the shared classifier knows 'service unavailable'", "upstream returned service unavailable", true, "transient_classified"},
		{"only the shared classifier knows 'bad gateway'", "proxy said bad gateway", true, "transient_classified"},
		{"only the shared classifier knows 'temporary'", "temporary name resolution failure", true, "transient_classified"},
		{"a validation failure is permanent to both", "page name is required", false, ""},
		{"an ordinary handler refusal is permanent to both", "shrink guard refused this save", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := defaultLadderState()
			st.attemptCount, st.maxAttempts = 2, 3
			out, _, err := runLadder(t, st, tc.errMsg, func(m sqlmock.Sqlmock, w *ladderWrite) {
				expectBurstProbe(m, 1, 1, 1) // healthy fleet: only the string classifiers can release
				if !tc.wantRelease {
					expectLadderUpdate(m, w, "failed", 0, true)
				} else {
					m.ExpectQuery(`UPDATE site_work_items`).
						WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow("triaged", 1))
				}
			})
			if err != nil {
				t.Fatalf("ladder: %v", err)
			}
			if out.Released != tc.wantRelease {
				t.Fatalf("released = %v, want %v for %q", out.Released, tc.wantRelease, tc.errMsg)
			}
			if tc.wantRelease && !strings.HasPrefix(out.ReleaseReason, tc.wantReason) {
				t.Errorf("release reason = %q, want prefix %q", out.ReleaseReason, tc.wantReason)
			}
		})
	}
}

// ── Robustness: the failure path must never be the thing that breaks ────────

func TestFailureLadder_ABrokenBurstProbeFallsThroughToTheLadder(t *testing.T) {
	// A detector that can block the failure write is worse than no detector.
	st := defaultLadderState()
	st.attemptCount, st.maxAttempts = 2, 3

	db, mock, _ := sqlmock.New()
	defer db.Close()
	expectStateRead(mock, st)
	mock.ExpectQuery(`FROM agent_error_log`).WillReturnError(errors.New("statement timeout"))
	w := &ladderWrite{}
	expectLadderUpdate(mock, w, "failed", 0, true)

	out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(), uuid.New(),
		"adapter exploded", "build-dispatch-loop", nil)
	if err != nil {
		t.Fatalf("a failed burst probe must not fail the write: %v", err)
	}
	if out.NewStatus != "failed" {
		t.Errorf("status = %q, want failed", out.NewStatus)
	}
}

func TestFailureLadder_BurstThresholdsRequireTheConjunction(t *testing.T) {
	// The conjunction IS the discriminator (measured over 7 days to 2026-08-19:
	// with ≥2 domains AND ≥2 agent types, exactly three signatures fire and all
	// three were real infrastructure outages; zero single-item faults fire).
	// Each row below drops exactly one leg and must not fire.
	cases := []struct {
		name                   string
		errs, domains, agentTs int
		wantBurst              bool
	}{
		{"all three legs met", defaultBurstMinErrors, 2, 2, true},
		{"volume alone, one domain", 500, 1, 5, false},
		{"volume alone, one agent type", 500, 5, 1, false},
		{"spread but too few errors", defaultBurstMinErrors - 1, 5, 5, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			expectBurstProbe(mock, tc.errs, tc.domains, tc.agentTs)
			got := detectFailureBurst(context.Background(), db, zap.NewNop(), "some fleet-wide error")
			if got != tc.wantBurst {
				t.Errorf("burst = %v, want %v (errs=%d domains=%d types=%d)",
					got, tc.wantBurst, tc.errs, tc.domains, tc.agentTs)
			}
		})
	}
}

func TestFailureLadder_BurstProbeNormalisesBothSidesInSQL(t *testing.T) {
	// Normalising the incoming message in Go and the stored rows in SQL would be
	// two implementations of one rule — the drift class bugs_closed/034 closed.
	// So the probe must apply the SAME expression to $2 as to error_message,
	// and it must key on `domain` (89.8% NULL on site_id, measured) and never on
	// error_code (ClassifyError labels the git 404 LLM_API_ERROR).
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(`count(DISTINCT domain)`)).
		WillReturnRows(sqlmock.NewRows([]string{"e", "d", "t"}).AddRow(1, 1, 1))
	detectFailureBurst(context.Background(), db, zap.NewNop(), "x")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("burst probe does not group by domain: %v", err)
	}

	db2, mock2, _ := sqlmock.New()
	defer db2.Close()
	mock2.ExpectQuery(regexp.QuoteMeta(`$2::text`)).
		WillReturnRows(sqlmock.NewRows([]string{"e", "d", "t"}).AddRow(1, 1, 1))
	detectFailureBurst(context.Background(), db2, zap.NewNop(), "x")
	if err := mock2.ExpectationsWereMet(); err != nil {
		t.Fatalf("burst probe does not normalise the incoming message in SQL: %v", err)
	}

	if strings.Contains(normSigFragment, "error_code") {
		t.Error("the burst signature must not involve error_code")
	}
}

func TestFailureLadder_AnEmptyErrorIsNeverABurst(t *testing.T) {
	// Otherwise every blank-message failure fleet-wide collapses to one
	// signature and matches itself en masse.
	db, mock, _ := sqlmock.New()
	defer db.Close()
	if detectFailureBurst(context.Background(), db, zap.NewNop(), "   ") {
		t.Error("an empty error message must not be treated as a burst signature")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("an empty message must not even query: %v", err)
	}
}

// ── Kill switches (WDS-018 posture: armed, independently disarmable) ────────

func TestFailureLadder_EachNewBehaviourDisarmsIndependently(t *testing.T) {
	t.Run("guard disarmed — the clause leaves the statement", func(t *testing.T) {
		t.Setenv(envDisableDecisionGuard, "1")
		st := defaultLadderState()
		st.attemptCount, st.maxAttempts = 2, 3
		db, mock, _ := sqlmock.New()
		defer db.Close()
		expectStateRead(mock, st)
		expectBurstProbe(mock, 1, 1, 1)
		// sqlmock collapses whitespace, so "the guard is gone" is expressible
		// as the WHERE landing straight on the RETURNING.
		mock.ExpectQuery(`WHERE id = \$1 RETURNING`).
			WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow("failed", 0))
		if _, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
			uuid.New(), "boom", "build-dispatch-loop", nil); err != nil {
			t.Fatalf("disarm did not remove the guard: %v", err)
		}
	})

	t.Run("backoff disarmed — no wait is stamped", func(t *testing.T) {
		t.Setenv(envDisableRetryBackoff, "1")
		st := defaultLadderState()
		db, mock, _ := sqlmock.New()
		defer db.Close()
		expectStateRead(mock, st)
		expectBurstProbe(mock, 1, 1, 1)
		// No reaper_policies lookup at all when the backoff is disarmed.
		w := &ladderWrite{}
		expectLadderUpdate(mock, w, "triaged", 2, true)
		out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
			uuid.New(), "boom", "build-dispatch-loop", nil)
		if err != nil {
			t.Fatalf("ladder: %v", err)
		}
		if out.BackoffMins != 0 {
			t.Errorf("backoff = %d with the disarm set", out.BackoffMins)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations: %v", err)
		}
	})

	t.Run("transient release disarmed — no probe, always counts", func(t *testing.T) {
		t.Setenv(envDisableTransientRelease, "1")
		st := defaultLadderState()
		st.attemptCount, st.maxAttempts = 2, 3
		db, mock, _ := sqlmock.New()
		defer db.Close()
		expectStateRead(mock, st)
		// No burst probe, no classifier: the whole release arm is off.
		w := &ladderWrite{}
		expectLadderUpdate(mock, w, "failed", 0, true)
		out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
			uuid.New(), "connection refused", "build-dispatch-loop", nil)
		if err != nil {
			t.Fatalf("ladder: %v", err)
		}
		if out.Released {
			t.Error("released with the transient-release disarm set")
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations: %v", err)
		}
	})

	t.Run("burst disarmed but the AI classifier still releases", func(t *testing.T) {
		// The disarms are separate because they fail differently: turning off a
		// misbehaving burst detector must not also revoke the pre-existing
		// AI-unavailable behaviour, which predates this change.
		t.Setenv(envDisableBurstRelease, "1")
		st := defaultLadderState()
		st.attemptCount, st.maxAttempts = 2, 3
		db, mock, _ := sqlmock.New()
		defer db.Close()
		expectStateRead(mock, st)
		mock.ExpectQuery(`UPDATE site_work_items`).
			WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow("triaged", 1))
		out, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
			uuid.New(), "connection refused by the endpoint", "build-dispatch-loop", nil)
		if err != nil {
			t.Fatalf("ladder: %v", err)
		}
		if !out.Released || out.ReleaseReason != "ai_unavailable" {
			t.Errorf("released=%v reason=%q — disarming the burst must not revoke the AI-unavailable release",
				out.Released, out.ReleaseReason)
		}
	})
}

// ── The contract is reached from BOTH writers ───────────────────────────────

func TestFailureLadder_IsTheOnlyFailureWriterInThePackage(t *testing.T) {
	// The defect was four writers with four guarantees. This asserts the two Go
	// failure paths converge here, by reading the source: a future edit that
	// re-introduces a private ladder in either action fails this test rather
	// than silently re-splitting the contract.
	for _, f := range []string{"load_work_item_actions.go", "v3_site_actions.go"} {
		src := readSourceForLadderTest(t, f)
		if !strings.Contains(src, "applyWorkItemFailureLadder(") {
			t.Errorf("%s no longer routes its failure path through the shared contract", f)
		}
	}
	// And the old private ladder must be gone from FailWorkItemAction: the CASE
	// that used to live there is the thing that had no delay and no guard.
	src := readSourceForLadderTest(t, "load_work_item_actions.go")
	failStart := strings.Index(src, "func FailWorkItemAction")
	if failStart < 0 {
		t.Fatal("FailWorkItemAction not found")
	}
	body := src[failStart:]
	if end := strings.Index(body, "\nfunc "); end > 0 {
		body = body[:end]
	}
	if strings.Contains(body, "attempt_count + 1 >= max_attempts") {
		t.Error("FailWorkItemAction still carries its own ladder — two implementations of one rule")
	}
}

// readSourceForLadderTest reads a file in this package. A source-scanning test
// makes comments load-bearing (LANDMINES), so the assertions above are written
// against code shapes — a call expression and a SQL fragment — not prose.
func readSourceForLadderTest(t *testing.T, name string) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join(".", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(src)
}

// TestGuardLists_TheFailureAndCompletionListsDifferByExactlyTwo pins the defect
// the council's editquality seat caught in round 1 (corr 4cdec68b, gating): the
// complete arm of update_work_item_status reused the FAILURE list, which omits
// 'failed' and 'unresolved' so the retry ladder can move a row through them.
// On the completion path that omission means a `complete` write silently
// overwrites a row that already failed or was given up — the exact
// silent-overwrite class this change exists to close.
//
// The two lists must therefore differ, and differ in a specific way. Asserting
// "they are different" would pass on any divergence; asserting the exact delta
// is what makes this a pin rather than a smoke test.
func TestGuardLists_TheFailureAndCompletionListsDifferByExactlyTwo(t *testing.T) {
	inFailure := map[string]bool{}
	for _, s := range workItemDecisionStatuses {
		inFailure[s] = true
	}
	inCompletion := map[string]bool{}
	for _, s := range workItemCompletionGuardStatuses {
		inCompletion[s] = true
	}

	// The retry ladder needs these overwritable on the failure path and
	// protected on the completion path. This is the whole objection.
	for _, s := range []string{"failed", "unresolved"} {
		if inFailure[s] {
			t.Errorf("%q is in the FAILURE guard list — the ladder could not move a row through it, "+
				"and failed→triaged is what a retry IS", s)
		}
		if !inCompletion[s] {
			t.Errorf("%q is missing from the COMPLETION guard list — a complete write would silently "+
				"overwrite a row that already failed or was given up (the defect editquality gated on)", s)
		}
	}

	// Everything else must be protected on BOTH paths: a deliberate decision is
	// a deliberate decision whichever way the saga exits.
	for _, s := range workItemDecisionStatuses {
		if !inCompletion[s] {
			t.Errorf("%q is guarded on the failure path but not on the completion path — "+
				"a decision must not depend on which exit the saga takes", s)
		}
	}

	// And they must not be the same constant, which is how they drifted in the
	// first place.
	// `deferred` is in BOTH (a park is a decision either way), so the delta stays
	// exactly failed+unresolved regardless of what else is added to both.
	if len(workItemCompletionGuardStatuses) != len(workItemDecisionStatuses)+2 {
		t.Errorf("completion list has %d entries, failure list %d — expected exactly two more "+
			"(failed, unresolved); if that changed, the divergence needs re-arguing, not re-sizing",
			len(workItemCompletionGuardStatuses), len(workItemDecisionStatuses))
	}
}

// TestUpdateWorkItemStatus_CompleteArmUsesTheCompletionGuard reads the statement
// the action actually builds. The list constants above can be perfect while the
// call site interpolates the wrong one — which is precisely what happened.
func TestUpdateWorkItemStatus_CompleteArmUsesTheCompletionGuard(t *testing.T) {
	src := readSourceForLadderTest(t, "v3_site_actions.go")
	completeArm := src
	if i := strings.Index(completeArm, `if newStatus == "complete" {`); i >= 0 {
		completeArm = completeArm[i:]
		if j := strings.Index(completeArm, "execDB("); j > 0 {
			completeArm = completeArm[:j]
		}
	} else {
		t.Fatal("could not locate the complete arm in UpdateWorkItemStatusAction")
	}
	if !strings.Contains(completeArm, "sqlInList(workItemCompletionGuardStatuses)") {
		t.Error("the complete arm does not interpolate workItemCompletionGuardStatuses")
	}
	if strings.Contains(completeArm, "sqlInList(workItemDecisionStatuses)") {
		t.Error("the complete arm interpolates the FAILURE guard list — 'failed' and 'unresolved' " +
			"would be overwritable by a complete write (council corr 4cdec68b, editquality, gating)")
	}
}

// ── Statement/bind-list integrity: the 42P18 class ──────────────────────────

// TestLadderStatementPlaceholdersMatchBindList audits every assembled variant
// of both ladder statements: the set of $N placeholders in the text must be
// exactly $1..$len(args), contiguous. This is the check sqlmock structurally
// cannot perform — a mock matches text and never TYPES a placeholder — and its
// absence is how the terminal transition shipped broken: computeLadder passes
// backoff 0 exactly on the terminal attempt, the first cut's retry_after
// fragment collapsed to `retry_after = NULL,`, $4 left the text but not the
// bind list, and every terminal write died with SQLSTATE 42P18 ("could not
// determine data type of parameter $4") in production while all fifteen tests
// here passed (bugs_open/307 §9, 2026-08-21: one canary + one natural
// fail_work_item hit inside two minutes).
func TestLadderStatementPlaceholdersMatchBindList(t *testing.T) {
	type variant struct {
		name string
		q    string
		args []interface{}
	}
	id := uuid.New()
	var variants []variant
	for _, withCol := range []bool{true, false} {
		for _, backoff := range []int{0, 30} {
			q, args := countingLadderStatement(id, "boom", "loop", backoff, nil, "", withCol)
			variants = append(variants, variant{fmt.Sprintf("counting/col=%v/backoff=%d", withCol, backoff), q, args})
		}
		q, args := transientReleaseStatement(id, "boom", "ai_unavailable", 15, nil, "", withCol)
		variants = append(variants, variant{fmt.Sprintf("transient/col=%v", withCol), q, args})
	}

	ph := regexp.MustCompile(`\$([0-9]+)`)
	for _, v := range variants {
		seen := map[string]bool{}
		for _, m := range ph.FindAllStringSubmatch(v.q, -1) {
			seen[m[1]] = true
		}
		for n := 1; n <= len(v.args); n++ {
			key := fmt.Sprintf("%d", n)
			if !seen[key] {
				t.Errorf("%s: $%d is bound but never referenced — SQLSTATE 42P18 at PREPARE time", v.name, n)
			}
			delete(seen, key)
		}
		for n := range seen {
			t.Errorf("%s: $%s is referenced but never bound", v.name, n)
		}
	}
}

// TestLadderStatementTextIsValueInvariant pins the repair's shape itself: the
// zero-backoff case lives IN the SQL (`$4::int <= 0` → NULL), never in a
// Go-side text branch — a statement whose TEXT depends on a bound VALUE is the
// exact construction that produced the 42P18, because only one of its shapes
// ever meets the test suite.
func TestLadderStatementTextIsValueInvariant(t *testing.T) {
	id := uuid.New()
	q0, _ := countingLadderStatement(id, "boom", "loop", 0, nil, "", true)
	q30, _ := countingLadderStatement(id, "boom", "loop", 30, nil, "", true)
	if q0 != q30 {
		t.Fatal("counting-ladder SQL text varies with the backoff VALUE — the 42P18 construction is back")
	}
}

// TestFailureLadder_ZeroBackoffTriggersAllReachTheSameStatement drives the
// THREE trigger paths that make backoffMins 0 (census from the lane that built
// the ladder, 2026-08-21): the terminal attempt (computeLadder never assigns),
// the DISABLE_WORK_ITEM_RETRY_BACKOFF disarm (the other leg of the same `if` —
// pre-fix, the kill switch shipped as the SAFE way to quiet a misbehaving
// backoff broke EVERY failure write instead), and a reaper_policies row of 0
// (operator-set, live-editable, no build needed to trigger). Each case asserts
// the UPDATE carries the value-invariant CASE text (`OR $4::int <= 0`): the
// pre-fix code emitted `retry_after = NULL,` on all three, which is the
// dropped-$4 statement that died with SQLSTATE 42P18 in production.
func TestFailureLadder_ZeroBackoffTriggersAllReachTheSameStatement(t *testing.T) {
	cases := []struct {
		name  string
		state ladderState
		arm   func(t *testing.T, mock sqlmock.Sqlmock)
		want  string // status the mock returns, matching the arm's real outcome
	}{
		{
			name:  "terminal attempt (2 of 3): computeLadder skips the policy lookup",
			state: ladderState{status: "claimed", attemptCount: 2, maxAttempts: 3, itemType: "content_rewrite"},
			arm: func(t *testing.T, mock sqlmock.Sqlmock) {
				expectBurstProbe(mock, 1, 1, 1)
			},
			want: "failed",
		},
		{
			name:  "one-shot item (max_attempts=1): the burst probe is skipped too",
			state: ladderState{status: "claimed", attemptCount: 0, maxAttempts: 1, itemType: "content_rewrite"},
			arm:   func(t *testing.T, mock sqlmock.Sqlmock) {},
			want:  "failed",
		},
		{
			name:  "retry-backoff disarm set, NON-terminal attempt",
			state: defaultLadderState(),
			arm: func(t *testing.T, mock sqlmock.Sqlmock) {
				t.Setenv(envDisableRetryBackoff, "1") // disarmed: backoff stays 0
				expectBurstProbe(mock, 1, 1, 1)
			},
			want: "triaged",
		},
		{
			name:  "reaper_policies row of 0 minutes",
			state: defaultLadderState(),
			arm: func(t *testing.T, mock sqlmock.Sqlmock) {
				expectBurstProbe(mock, 1, 1, 1)
				expectPolicyLookup(mock, 0)
			},
			want: "triaged",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			expectStateRead(mock, tc.state)
			tc.arm(t, mock)
			mock.ExpectQuery(regexp.QuoteMeta(`OR $4::int <= 0 THEN NULL`)).
				WillReturnRows(sqlmock.NewRows([]string{"status", "attempts_left"}).AddRow(tc.want, 0))
			if _, err := applyWorkItemFailureLadder(context.Background(), db, zap.NewNop(),
				uuid.New(), "boom", "build-dispatch-loop", nil); err != nil {
				t.Fatalf("ladder errored on a zero-backoff trigger: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("the zero-backoff UPDATE does not carry the value-invariant CASE ($4 referenced): %v", err)
			}
		})
	}
}
