// Guards for the half-cadence due look-ahead (bugs_open/410).
//
// The defect these tests hold shut: a due predicate of bare `next_fetch_at <=
// NOW()` phase-locks any source whose fetch_interval equals the trigger
// cadence — the post-fetch stamp lands seconds after the next trigger fires,
// so the source is served every OTHER pass, and every run still completes
// green. Measured 2026-08-26: 10 of 12 news sites at half their labelled
// cadence.
//
// TWO TESTS, DELIBERATELY DIFFERENT IN KIND, because each is vacuous alone.
// The sqlmock tests prove the ACTIONS send the shared predicate to the
// database — but they match the query against the same constant the query is
// built from, so they cannot see the constant itself going wrong. The shape
// test pins the constant against independent literals — but it cannot see an
// action that stops using the constant. Delete the look-ahead from the
// constant and the shape test goes red; restate the query in an action without
// the constant and its sqlmock test goes red. Both mutations were run before
// this file was committed.

package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// TestDispatchFeedSourcesQueriesWithTheDueLookahead proves the dispatcher's
// due-sources query reaches the database carrying the shared predicate. The
// expectation matches the predicate text verbatim; if the query loses it —
// reverted to bare NOW(), or rewritten without the constant — the mock refuses
// the query, the action errors, and this test fails naming the cause.
func TestDispatchFeedSourcesQueriesWithTheDueLookahead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// isFeedEnabled: settings say the feed is on, so the action proceeds to
	// the due-sources query this test is actually about.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(settings, '{}'::jsonb) FROM sites")).
		WillReturnRows(sqlmock.NewRows([]string{"settings"}).
			AddRow(`{"maintenance_profile":{"content_feed":{"enabled":true}}}`))

	mock.ExpectQuery(regexp.QuoteMeta(feedSourceDuePredicate)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_type", "name", "config"}))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		AgentType:        "content-feed-orchestrator",
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "dispatch_sources"},
		CollectedData:    map[string]interface{}{"site_id": "1244516d-014d-421c-88c6-090bb1e9552a"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
	}

	out, err := DispatchFeedSourcesAction(context.Background(), params)
	if err != nil {
		t.Fatalf("DispatchFeedSourcesAction with the shared due predicate: %v — "+
			"if this names an unmatched query, the dispatcher has lost the look-ahead", err)
	}
	result, ok := out.(map[string]interface{})
	if !ok || result["dispatched"] != 0 {
		t.Fatalf("expected a zero-dispatch result over an empty source set, got %#v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestLoadDueSourcesQueriesWithTheDueLookahead holds the second Go reader to
// the same predicate. It has no live workflow caller as of 2026-08-26, and
// that is exactly why it needs its own guard: an unused path is where the
// predicate would quietly diverge before a future caller inherits it.
func TestLoadDueSourcesQueriesWithTheDueLookahead(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(feedSourceDuePredicate)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_type", "name", "config", "fetch_interval"}))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		AgentType:        "content-feed-orchestrator",
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "load_due_sources"},
		CollectedData:    map[string]interface{}{"site_id": "1244516d-014d-421c-88c6-090bb1e9552a"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
	}

	if _, err := LoadDueSourcesAction(context.Background(), params); err != nil {
		t.Fatalf("LoadDueSourcesAction with the shared due predicate: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// TestFeedDueLookaheadShape pins the constant against literals this test owns,
// so it cannot pass by construction the way the sqlmock tests can. Each
// required fragment is one load-bearing half of the design:
//   - the live-cadence subquery, so a capacity change in scheduled_tasks
//     propagates without a code change;
//   - the halving, which is what makes the window "nearest tick" rather than
//     "always fetch early";
//   - the named task row, which is the one source of truth for the cadence;
//   - the COALESCE fallback to 3 hours (half of the 21600 s cadence this was
//     written against — migration 653's guard asserts that equality against
//     the live row), so a renamed task degrades to today's designed value and
//     never to the bare NOW() that produced the phase lock.
func TestFeedDueLookaheadShape(t *testing.T) {
	for _, fragment := range []string{
		"SELECT make_interval(secs => interval_seconds / 2.0) FROM scheduled_tasks",
		"WHERE name = 'content-feed-refresh'",
		"COALESCE(",
		"interval '3 hours'",
	} {
		if !strings.Contains(feedDueLookaheadSQL, fragment) {
			t.Errorf("feedDueLookaheadSQL lost %q — the look-ahead no longer %s",
				fragment, "does what feed_due_lookahead.go says it does")
		}
	}
	if !strings.Contains(feedSourceDuePredicate, "next_fetch_at IS NULL OR next_fetch_at <= NOW() + ") {
		t.Errorf("feedSourceDuePredicate no longer applies the look-ahead to next_fetch_at: %q",
			feedSourceDuePredicate)
	}
	if strings.Contains(feedSourceDuePredicate, "next_fetch_at <= NOW())") {
		t.Errorf("feedSourceDuePredicate carries a bare next_fetch_at <= NOW() arm — " +
			"that is the exact predicate bugs_open/410 exists to remove")
	}
}
