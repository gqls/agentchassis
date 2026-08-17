// FILE: platform/orchestration/actions/query_row_cap_test.go
//
// bugs_open/275 — a row-count LIMIT feeding an LLM prompt is a SILENT cap.
//
// The fixtures below are the REAL live queries, copied verbatim from
// agent_definitions on 2026-08-16, not invented shapes. That matters: the whole
// value of this check is that it fires on the estate's actual configs, and a
// test built from hand-written SQL would prove only that the regex works on SQL
// I wrote to make it work.
package actions

import (
	"context"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// The live tool-suggester query, verbatim (migration 406's requires-backend gate
// included — the bug warns explicitly not to work from an older ungated sketch).
const liveToolSuggesterQuery = `SELECT id::text, function, display_name, category, description FROM content_components WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true AND html_template != '' AND (NOT (COALESCE(semantic_tags, '[]'::jsonb) ? 'requires-backend') OR EXISTS (SELECT 1 FROM sites s WHERE s.id = $1 AND COALESCE(s.deploy_config->'capabilities', '[]'::jsonb) ? 'backend')) ORDER BY display_name LIMIT 30`

func TestQueryRowCapFindsTheLiveToolSuggesterCap(t *testing.T) {
	n, ok := queryRowCap(liveToolSuggesterQuery)
	if !ok {
		t.Fatal("did not find a cap in the live tool-suggester query — this is the exact query " +
			"bugs_open/275 is about, so a miss here means the check cannot see the case it was written for")
	}
	if n != 30 {
		t.Fatalf("found LIMIT %d, wanted 30", n)
	}
}

// THE SIGNAL-TO-NOISE DECISION, asserted explicitly because it is the arm most
// likely to be "simplified" away by someone who reads the regex and not the
// reasoning. 19 of 26 live hits are LIMIT 1 — fetch-one and claim-one — and they
// return exactly one row on EVERY execution by design.
func TestFetchOneIdiomIsNotReportedAsACap(t *testing.T) {
	for _, q := range []string{
		`SELECT body FROM doc_notes WHERE id = $1 LIMIT 1`,
		`SELECT * FROM site_work_items WHERE status='pending' ORDER BY created_at LIMIT 1`,
		`SELECT 1 LIMIT 0`,
	} {
		if n, ok := queryRowCap(q); ok {
			t.Errorf("LIMIT %d reported as a cap for %q — a lookup that returns one row by design "+
				"would warn on every run for ever, and a channel that always fires is a channel "+
				"nobody reads", n, q)
		}
	}
}

// A LIMIT that does not bound the whole result set must NOT be read as one.
func TestOnlyATrailingLimitCounts(t *testing.T) {
	cases := []struct {
		name string
		q    string
	}{
		{"limit inside a subquery", `SELECT * FROM a WHERE id IN (SELECT id FROM b ORDER BY x LIMIT 5)`},
		{"limit inside a CTE", `WITH t AS (SELECT * FROM b ORDER BY x LIMIT 6) SELECT * FROM t JOIN c USING (id)`},
		{"parameterised limit", `SELECT * FROM a ORDER BY x LIMIT $2`},
		{"no limit at all", `SELECT * FROM a ORDER BY x`},
	}
	for _, tc := range cases {
		if n, ok := queryRowCap(tc.q); ok {
			t.Errorf("%s: reported LIMIT %d; only a limit bounding the WHOLE statement may be "+
				"reported, because anything else bounds a different set and the warning would be false",
				tc.name, n)
		}
	}
}

// Trailing whitespace/semicolon and case must not defeat it — real configs carry both.
func TestTrailingFormsAreStillFound(t *testing.T) {
	for _, q := range []string{
		`SELECT * FROM a LIMIT 15`,
		`SELECT * FROM a LIMIT 15;`,
		"SELECT * FROM a LIMIT 15\n",
		`SELECT * FROM a limit 15`,
		"SELECT * FROM a\n  LIMIT 15\n;\n",
	} {
		n, ok := queryRowCap(q)
		if !ok || n != 15 {
			t.Errorf("missed the cap in %q (got %d, %v)", q, n, ok)
		}
	}
}

// The condition the action actually branches on.
func TestResultHitItsRowCapOnlyFiresOnEquality(t *testing.T) {
	// exactly at the ceiling -> suspicious, warn
	if n, hit := resultHitItsRowCap(liveToolSuggesterQuery, 30); !hit || n != 30 {
		t.Fatalf("30 rows under LIMIT 30 must be reported (got n=%d hit=%v)", n, hit)
	}
	// under the ceiling -> the population is fully visible, say nothing
	if _, hit := resultHitItsRowCap(liveToolSuggesterQuery, 29); hit {
		t.Fatal("29 rows under LIMIT 30 is NOT a truncation — warning here would be noise on every " +
			"under-full result in the estate")
	}
	if _, hit := resultHitItsRowCap(liveToolSuggesterQuery, 0); hit {
		t.Fatal("an empty result is not a capped one")
	}
	// AT OR BEYOND the ceiling. A driver cannot return more rows than the LIMIT,
	// so this is unreachable from QueryDatabaseAction today — it is asserted
	// because it is the arm that decides `>=` vs `==`, and without it that choice
	// is untested and reads as arbitrary (mutation M3 survived until this existed).
	if _, hit := resultHitItsRowCap(liveToolSuggesterQuery, 31); !hit {
		t.Fatal("a count BEYOND the declared cap must still report as capped — the signal must not " +
			"go quiet for a caller whose row count did not come straight from the driver")
	}

	// no cap declared -> nothing to compare against
	if _, hit := resultHitItsRowCap(`SELECT * FROM a`, 100); hit {
		t.Fatal("a query with no LIMIT can never be reported as capped")
	}
	// the fetch-one idiom, end to end
	if _, hit := resultHitItsRowCap(`SELECT * FROM a WHERE id=$1 LIMIT 1`, 1); hit {
		t.Fatal("the fetch-one idiom must never warn — see TestFetchOneIdiomIsNotReportedAsACap")
	}
}

// The other live multi-row caps censused 2026-08-16, so this test fails if the
// check stops seeing the real population it was built for.
func TestEveryLiveMultiRowCapIsDetected(t *testing.T) {
	live := map[string]int{
		"tool-suggester.load_library_tools":            30,
		"internal-linker.load_candidate_pages":         15,
		"model-directory-trigger.find_directory_sites": 12,
		"tool-recreation-handler.load_related_context": 10,
		"content-feed-trigger.find_news_sites":         5,
		"visual-design-auditor.load_design_context":    5,
		"fix-proposer.load_last_bundle":                2,
	}
	for step, want := range live {
		q := "SELECT a, b FROM t WHERE x = $1 ORDER BY a LIMIT " + strconv.Itoa(want)
		got, ok := queryRowCap(q)
		if !ok || got != want {
			t.Errorf("%s: cap %d not detected (got %d, %v)", step, want, got, ok)
		}
	}
}

// ============================================================================
// THE CALL-SITE TEST, and it is the one that matters.
//
// Without it every test above is vacuous in the way this estate keeps getting
// caught by: they prove `resultHitItsRowCap` computes the right answer, and they
// ALL STILL PASS if `QueryDatabaseAction` never calls it. The helper being
// correct and the helper being USED are independent facts, and only the second
// one puts a warning in front of a human.
//
// Found by mutation M7 on 2026-08-16: disabling the branch in the action left
// the whole suite green. This test kills it.
//
// Behavioural, not a source scan — register OPP-003 records a source-scanning
// detector that examined zero files and printed a clean result, and "a clean
// result and an unrun check are byte-identical output".
// ============================================================================

func TestQueryDatabaseActionActuallyWarnsWhenAResultHitsItsCap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Three rows under LIMIT 3 — a result exactly at its ceiling.
	rows := sqlmock.NewRows([]string{"id", "display_name"}).
		AddRow("1", "Alpha").AddRow("2", "Beta").AddRow("3", "Gamma")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	core, logs := observer.New(zap.WarnLevel)
	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.New(core),
		AgentType:        "tool-suggester",
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "load_library_tools"},
		CollectedData:    map[string]interface{}{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"query": "SELECT id, display_name FROM content_components ORDER BY display_name LIMIT 3",
		}},
	}

	if _, err := QueryDatabaseAction(context.Background(), params); err != nil {
		t.Fatalf("QueryDatabaseAction: %v", err)
	}

	found := logs.FilterMessageSnippet("EQUALS the query's LIMIT").All()
	if len(found) != 1 {
		t.Fatalf("expected exactly ONE cap warning, got %d. The action is not calling "+
			"resultHitItsRowCap — every other test in this file passes anyway, which is the "+
			"whole reason this one exists (mutation M7).", len(found))
	}
	// The warning must NAME the step, or a reader cannot act on it.
	fields := found[0].ContextMap()
	if fields["step"] != "load_library_tools" {
		t.Errorf("warning does not name the step (got %v) — an unattributed cap warning "+
			"tells nobody which query to go and look at", fields["step"])
	}
	if fields["limit"] != int64(3) {
		t.Errorf("warning does not carry the limit (got %v)", fields["limit"])
	}
}

// NEGATIVE CONTROL for the test above: an UNDER-full result must stay silent, or
// the check would warn on most queries in the estate and be ignored within a day.
func TestQueryDatabaseActionStaysSilentBelowTheCap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Two rows under LIMIT 3 — the population is fully visible.
	rows := sqlmock.NewRows([]string{"id", "display_name"}).
		AddRow("1", "Alpha").AddRow("2", "Beta")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	core, logs := observer.New(zap.WarnLevel)
	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.New(core),
		AgentType:        "tool-suggester",
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process", StepName: "load_library_tools"},
		CollectedData:    map[string]interface{}{},
		StepConfig: models.Step{Config: map[string]interface{}{
			"query": "SELECT id, display_name FROM content_components ORDER BY display_name LIMIT 3",
		}},
	}

	if _, err := QueryDatabaseAction(context.Background(), params); err != nil {
		t.Fatalf("QueryDatabaseAction: %v", err)
	}
	if n := logs.FilterMessageSnippet("EQUALS the query's LIMIT").Len(); n != 0 {
		t.Fatalf("an under-full result warned %d time(s); it must be silent, or the channel "+
			"becomes noise and stops being read", n)
	}
}
