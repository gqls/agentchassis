package actions

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// bugs_open/149 B4 follow-up: an ERRORING discovery check must leave a durable
// record, not only a line in the step output map. Pod logs roll and
// collected_data is pruned at ~24h, so without this "which check silently
// stopped working, and when" is unanswerable a day later.
//
// THESE TESTS ASSERT THE MECHANISM FIRED, not merely that nothing blew up: each
// asserts on the INSERT and its column values, never on "no error returned".
//
// BUT NOTE THE DIVISION, because I got this wrong first and the mutation check
// is what caught it. The cases in this first group call
// writeDiscoveryCheckErrorLog DIRECTLY, so they pin the row's SHAPE and are
// blind to the WIRING — with the call site in RunDiscoveryChecksAction deleted,
// all nine still passed. That is exactly the failure LANDMINES.md records as "a
// quiet-test passes when the RULE is gone, not when the guard works", and it was
// true of this file until TestRunDiscoveryChecksWritesDurableRecordForErroringCheck
// (at the bottom) was added to drive the real action.
//
// Mutation-checked 2026-07-31, both directions: removing the call site leaves
// this group green and fails the wiring test on the unmet INSERT expectation
// plus an empty captured message. If you add a case here, ask which of the two
// things it pins — shape or wiring — and do not assume shape implies wiring.

// ---------------------------------------------------------------------------
// One row per failed check, with the columns that make it queryable
// ---------------------------------------------------------------------------

func TestDiscoveryCheckErrorLogWritesOneRowPerCheck(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	batchID := uuid.New()

	failed := []map[string]string{
		{"check": "orphan_tool_pages", "error": "dial tcp: connection refused"},
		{"check": "nav_drift", "error": "pq: relation \"nav_links\" does not exist"},
	}

	// TWO inserts, not one batched row: the question is "which check stopped
	// working", and a single row would force a reader to parse a list back out
	// of context to answer it.
	for range failed {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	writeDiscoveryCheckErrorLog(context.Background(), params, siteID,
		"gamesdesign.co.uk", "design-discovery-agent", "design", batchID,
		failed, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected one agent_error_log INSERT per failed check: %v", err)
	}
}

// ---------------------------------------------------------------------------
// The column values are the whole point — assert them, not just the verb
// ---------------------------------------------------------------------------

func TestDiscoveryCheckErrorLogColumnValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	batchID := uuid.New()

	var gotContext string
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WithArgs(
			siteID,
			"gamesdesign.co.uk",
			"design-discovery-agent",
			"scan_site", // step_name comes from the ExecutionContext, not the default
			sqlmock.AnyArg(),
			discoveryCheckErrorCode,
			captureArg{&gotContext},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "scan_site"},
	}
	writeDiscoveryCheckErrorLog(context.Background(), params, siteID,
		"gamesdesign.co.uk", "design-discovery-agent", "design", batchID,
		[]map[string]string{{"check": "nav_drift", "error": "boom"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("column values did not match: %v", err)
	}

	// The context jsonb must carry what makes the row joinable back to the run.
	for _, want := range []string{"nav_drift", "boom", "design", batchID.String()} {
		if !strings.Contains(gotContext, want) {
			t.Errorf("context jsonb missing %q; got: %s", want, gotContext)
		}
	}
}

// The error code must be distinct from every other code the estate writes.
// A shared code makes "which path caught this" unanswerable — the drift
// bugs_open/097 exists to keep answerable.
func TestDiscoveryCheckErrorCodeIsDistinct(t *testing.T) {
	if discoveryCheckErrorCode == claimsFloorErrorCode {
		t.Error("discovery check code must not reuse the claims floor code")
	}
	for _, taken := range []string{
		"CONTENT_LINK_REPAIR_DETAIL", "CONTENT_LINK_REPAIR_SKIPPED",
		"CONTENT_CLAIMS_FLOOR_DETAIL", "CONTENT_VALIDATION_FAILED",
		"CONTENT_VALIDATION_BLOCKER_DETAIL", "TRUNCATION_DEGRADED_REVIEW",
		"VALIDATION_ERROR_DROPPED", "UNKNOWN",
	} {
		if discoveryCheckErrorCode == taken {
			t.Errorf("discovery check code collides with live code %q", taken)
		}
	}
}

// ---------------------------------------------------------------------------
// Severity: `warning`, and the reason is a rule the estate already states
// ---------------------------------------------------------------------------
//
// validate_page_content_stats.go: "Severity `error` must never mean 'we could
// not check this'". A check erroring is exactly that case, so it must not be
// graded alongside a real content finding.

func TestDiscoveryCheckErrorLogUsesWarningSeverity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("'warning'").WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	writeDiscoveryCheckErrorLog(context.Background(), params, uuid.New(),
		"x.com", "a", "design", uuid.New(),
		[]map[string]string{{"check": "c", "error": "e"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the INSERT must grade the row `warning`, not `error`: %v", err)
	}
}

// ---------------------------------------------------------------------------
// No failures → no rows. A diagnostic that writes on a clean run is noise.
// ---------------------------------------------------------------------------

func TestDiscoveryCheckErrorLogSilentWhenNothingFailed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// No ExpectExec at all: any INSERT is an unexpected call and fails the test.
	params := ActionParams{DB: db, Logger: zap.NewNop()}
	writeDiscoveryCheckErrorLog(context.Background(), params, uuid.New(),
		"x.com", "a", "design", uuid.New(), nil, zap.NewNop())
	writeDiscoveryCheckErrorLog(context.Background(), params, uuid.New(),
		"x.com", "a", "design", uuid.New(), []map[string]string{}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a clean run must write nothing: %v", err)
	}
}

// ---------------------------------------------------------------------------
// agent_type is NOT NULL on agent_error_log — never send an empty string
// ---------------------------------------------------------------------------

func TestDiscoveryCheckErrorLogDefaultsAgentType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WithArgs(siteID, "x.com", "unknown", "run_discovery_checks",
			sqlmock.AnyArg(), discoveryCheckErrorCode, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	writeDiscoveryCheckErrorLog(context.Background(), params, siteID,
		"x.com", "", "design", uuid.New(),
		[]map[string]string{{"check": "c", "error": "e"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("empty agent_type must default to a non-empty label: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Best-effort: a failing diagnostic write must not stop the remaining rows,
// and must not panic. The run's real output is already committed by this point.
// ---------------------------------------------------------------------------

func TestDiscoveryCheckErrorLogToleratesWriteFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// First insert fails; the SECOND must still be attempted.
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WillReturnError(errors.New("pq: deadlock detected"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	writeDiscoveryCheckErrorLog(context.Background(), params, uuid.New(),
		"x.com", "a", "design", uuid.New(),
		[]map[string]string{
			{"check": "first", "error": "e1"},
			{"check": "second", "error": "e2"},
		}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a failed row must not abandon the rest: %v", err)
	}
}

// A nil DB is the unit-test and odd-adoption path; it must be a no-op, not a
// nil dereference.
func TestDiscoveryCheckErrorLogNilDBIsNoOp(t *testing.T) {
	params := ActionParams{DB: nil, Logger: zap.NewNop()}
	writeDiscoveryCheckErrorLog(context.Background(), params, uuid.New(),
		"x.com", "a", "design", uuid.New(),
		[]map[string]string{{"check": "c", "error": "e"}}, zap.NewNop())
}

// ---------------------------------------------------------------------------
// The message must name the consequence, not just the failure
// ---------------------------------------------------------------------------
//
// "check X errored" tells an operator nothing actionable. "the site was NOT
// checked for this class" is the fact that matters when reading the row months
// later, and it is why a zero finding count on that site is not clearance.

func TestDiscoveryCheckErrorLogMessageNamesTheConsequence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	var captured string
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			captureArg{&captured},
			discoveryCheckErrorCode, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	writeDiscoveryCheckErrorLog(context.Background(), params, uuid.New(),
		"x.com", "a", "design", uuid.New(),
		[]map[string]string{{"check": "nav_drift", "error": "boom"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("insert did not match: %v", err)
	}
	for _, want := range []string{"nav_drift", "NOT checked", "boom"} {
		if !strings.Contains(captured, want) {
			t.Errorf("error_message missing %q; got: %s", want, captured)
		}
	}
}

// ---------------------------------------------------------------------------
// THE WIRING TEST — the one that actually catches a removed call site
// ---------------------------------------------------------------------------
//
// Written after a mutation check exposed that the tests above do NOT catch it:
// they call writeDiscoveryCheckErrorLog directly, so deleting the call in
// RunDiscoveryChecksAction left all nine passing. That is precisely the failure
// LANDMINES.md records — "a quiet-test passes when the RULE is gone, not when
// the guard works" — and it was true of my own suite until this test existed.
//
// This drives the real action with a real registered check (`empty_sections`)
// whose query is made to fail, so the erroring-check path is reached the way
// production reaches it. No registry mutation: registering a fake check would
// leave a phantom name visible to the coverage tests in this same binary.
func TestRunDiscoveryChecksWritesDurableRecordForErroringCheck(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()

	mock.ExpectQuery("SELECT domain FROM sites").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("gamesdesign.co.uk"))

	mock.ExpectBegin()

	// The check's own query fails — a transient DB fault, the condition the
	// report-and-continue arm exists for.
	mock.ExpectQuery("FROM page_components").
		WillReturnError(errors.New("pq: canceling statement due to statement timeout"))

	mock.ExpectCommit()

	// THE ASSERTION: the erroring check left a durable row.
	var gotMessage string
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO agent_error_log")).
		WithArgs(siteID, "gamesdesign.co.uk", sqlmock.AnyArg(), sqlmock.AnyArg(),
			captureArg{&gotMessage}, discoveryCheckErrorCode, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		StepConfig: models.Step{Config: map[string]interface{}{
			"checks": []interface{}{"empty_sections"},
		}},
		CollectedData: map[string]interface{}{"site_id": siteID.String()},
	}

	out, err := RunDiscoveryChecksAction(context.Background(), params)
	if err != nil {
		t.Fatalf("the run must not fail because one check errored: %v", err)
	}

	// The step output still reports it — the durable row is in ADDITION to that,
	// not a replacement for it.
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected output type %T", out)
	}
	failed, _ := m["checks_failed"].([]map[string]string)
	if len(failed) != 1 {
		t.Errorf("checks_failed should name the one erroring check, got %v", m["checks_failed"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("an erroring check must leave an agent_error_log row: %v", err)
	}
	if !strings.Contains(gotMessage, "empty_sections") {
		t.Errorf("the durable row must name the check; got: %s", gotMessage)
	}
}
