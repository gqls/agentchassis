package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

// bugs_open/172 — gatherAgentState's two caps both reported success while
// discarding evidence, and the section's heading asserted the coverage they had
// just dropped.
//
// Every assertion here has to be INDUCED: the type cap has never fired in
// production (max ever listed is 4 against a default of 5), so a live run cannot
// exercise it. These drive gatherAgentState through sqlmock with the cap lowered,
// which is the only way to see the branch at all.

// agentStateMock wires the three queries gatherAgentState issues, in order.
// callLogRows is (agent_type, count) pairs rendered as log rows.
type agentStateMock struct {
	allTypes   []string
	callLog    [][2]string // {agent_type, step_name} per row, in the order the query returns
	skipConfig bool
}

func newAgentStateDB(t *testing.T, m agentStateMock) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}

	typeRows := sqlmock.NewRows([]string{"type"})
	for _, ty := range m.allTypes {
		typeRows.AddRow(ty)
	}
	mock.ExpectQuery("SELECT DISTINCT type FROM agent_definitions").WillReturnRows(typeRows)

	if !m.skipConfig {
		mock.ExpectQuery("FROM agent_definitions\\s*\\n?\\s*WHERE type = ANY").
			WillReturnRows(sqlmock.NewRows([]string{"type", "model", "max", "topmax", "hasroot"}))
		mock.ExpectQuery("LATERAL jsonb_each").
			WillReturnRows(sqlmock.NewRows([]string{"type", "key", "model", "max"}))
	}

	logRows := sqlmock.NewRows([]string{"created_at", "agent_type", "step_name", "model", "max_tokens", "output_tokens", "success"})
	for _, r := range m.callLog {
		logRows.AddRow("2026-08-02T00:00:00Z", r[0], r[1], "claude-opus-5", 4096, 512, true)
	}
	mock.ExpectQuery("ROW_NUMBER\\(\\) OVER \\(PARTITION BY agent_type").WillReturnRows(logRows)

	return db, mock
}

// TestGatherAgentStateReportsTypeCap induces the cap the bug is filed about:
// three named types, cap of two. The dropped type must be NAMED, the count
// stated, and the two survivors must still render.
func TestGatherAgentStateReportsTypeCap(t *testing.T) {
	// Sorted, as `ORDER BY type` now guarantees. The cap slices the TAIL, so
	// "zz-writer" is the casualty and that is now predictable rather than
	// plan-dependent.
	db, _ := newAgentStateDB(t, agentStateMock{
		allTypes: []string{"aa-router", "mm-planner", "zz-writer"},
		callLog:  [][2]string{{"aa-router", "step_a"}, {"mm-planner", "step_b"}},
	})
	defer db.Close()

	var b strings.Builder
	gatherAgentState(context.Background(), db,
		"aa-router handed to mm-planner and zz-writer never ran", &b, 2, 10, zap.NewNop())
	got := b.String()

	if !strings.Contains(got, "auto-gathered: 2 of 3 agent types named") {
		t.Errorf("heading must count kept-vs-named when the cap fires; got:\n%s", got)
	}
	if !strings.Contains(got, "1 further agent type(s) named in the symptom/hypothesis were NOT gathered (agent_state_cap=2): zz-writer") {
		t.Errorf("cap notice must state the count, the cap and the dropped TYPE; got:\n%s", got)
	}
	if !strings.Contains(got, "absence here is not evidence an agent was uninvolved") {
		t.Errorf("cap notice must warn against reading the absence as an answer; got:\n%s", got)
	}
	// The survivors still render — a marker that ate the evidence would be worse
	// than the silence it replaced.
	for _, want := range []string{"aa-router/step_a", "mm-planner/step_b"} {
		if !strings.Contains(got, want) {
			t.Errorf("kept type's rows must still render (%s); got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "zz-writer/") {
		t.Errorf("dropped type must not be gathered; got:\n%s", got)
	}
}

// TestGatherAgentStateNegativeControl is the control the bug asks for: below the
// cap, the section must be byte-identical to what it rendered before the fix.
// Written as a literal, not derived from the code, so a later edit to the heading
// or the log line has to change this string deliberately.
func TestGatherAgentStateNegativeControl(t *testing.T) {
	db, _ := newAgentStateDB(t, agentStateMock{
		allTypes: []string{"aa-router", "mm-planner"},
		callLog:  [][2]string{{"aa-router", "step_a"}, {"mm-planner", "step_b"}},
	})
	defer db.Close()

	var b strings.Builder
	gatherAgentState(context.Background(), db,
		"aa-router handed to mm-planner", &b, 5, 10, zap.NewNop())

	const want = "\n### agent state (auto-gathered: agent types named in the symptom/hypothesis)\n\n" +
		"- llm_call_log [2026-08-02T00:00:00Z] aa-router/step_a model=claude-opus-5 max_tokens=4096 output_tokens=512 success=true\n" +
		"- llm_call_log [2026-08-02T00:00:00Z] mm-planner/step_b model=claude-opus-5 max_tokens=4096 output_tokens=512 success=true\n"
	if got := b.String(); got != want {
		t.Errorf("uncapped section must be byte-identical to the pre-fix baseline\n got: %q\nwant: %q", got, want)
	}
}

// TestGatherAgentStateNamesStarvedTypes covers the half that was FIRING: a named
// type with no rows must be stated, so its absence cannot be read as "this agent
// made no LLM calls" when the truth is that the gather never covered it.
func TestGatherAgentStateNamesStarvedTypes(t *testing.T) {
	db, _ := newAgentStateDB(t, agentStateMock{
		allTypes: []string{"chatty-agent", "quiet-agent"},
		// Per-type allocation returns chatty's rows; quiet genuinely has none.
		callLog: [][2]string{{"chatty-agent", "s1"}, {"chatty-agent", "s2"}},
	})
	defer db.Close()

	var b strings.Builder
	gatherAgentState(context.Background(), db,
		"chatty-agent called quiet-agent", &b, 5, 10, zap.NewNop())
	got := b.String()

	if !strings.Contains(got, "no llm_call_log rows carry agent_type: quiet-agent") {
		t.Errorf("a gathered type with no rows must be stated by name; got:\n%s", got)
	}
	if !strings.Contains(got, "their budget was not spent and this is not a cap") {
		t.Errorf("an unspent budget must be distinguished from a capped one; got:\n%s", got)
	}
	// The council's llm_reliability seat: agent_type was relabelled 2026-07-26, so
	// "no rows" is a fact about the LABEL and must not read as "made no calls".
	if !strings.Contains(got, "relabelled 2026-07-26") {
		t.Errorf("the empty case must name the relabel boundary rather than assert a clean absence; got:\n%s", got)
	}
}

// TestGatherAgentStateOrdersTheTypeListing asserts the ORDER BY that makes the
// kept set reproducible.
//
// It has to assert the QUERY TEXT, and that limitation is the point: sqlmock
// replays rows in whatever order the test supplies, so no mock-driven test can
// observe the database's ordering — a run that asserted only on the returned
// slice passed unchanged when ORDER BY was deleted (measured, 2026-08-02). The
// ordering itself is Postgres's guarantee, verified against the live table in
// RUNBOOK_agent_state_cap.md; this test's job is to stop the clause being
// dropped, which is the failure mode a mock CAN catch.
func TestGatherAgentStateOrdersTheTypeListing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Strict: the ORDER BY must be present or this expectation does not match,
	// the query errors, and the section renders nothing.
	mock.ExpectQuery(`SELECT DISTINCT type FROM agent_definitions WHERE deleted_at IS NULL ORDER BY type`).
		WillReturnRows(sqlmock.NewRows([]string{"type"}).AddRow("aa-router"))
	mock.ExpectQuery("WHERE type = ANY").
		WillReturnRows(sqlmock.NewRows([]string{"type", "model", "max", "topmax", "hasroot"}))
	mock.ExpectQuery("LATERAL jsonb_each").
		WillReturnRows(sqlmock.NewRows([]string{"type", "key", "model", "max"}))
	mock.ExpectQuery("ROW_NUMBER").
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "agent_type", "step_name", "model", "max_tokens", "output_tokens", "success"}))

	var b strings.Builder
	gatherAgentState(context.Background(), db, "aa-router failed", &b, 5, 10, zap.NewNop())

	if !strings.Contains(b.String(), "agent state (auto-gathered") {
		t.Fatalf("section did not render — the type listing query no longer carries ORDER BY type; got:\n%s", b.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("query expectations unmet: %v", err)
	}
}

// TestCallLogCoverageNotice pins the two facts apart without a DB. A type at its
// budget and a type with nothing are opposite signals: one says "look further",
// the other says "there is nothing to find".
func TestCallLogCoverageNotice(t *testing.T) {
	cases := []struct {
		name        string
		matched     []string
		perType     map[string]int
		limit       int
		wantContain []string
		wantEmpty   bool
	}{
		{
			name:      "every type fully covered says nothing",
			matched:   []string{"a", "b"},
			perType:   map[string]int{"a": 3, "b": 1},
			limit:     10,
			wantEmpty: true,
		},
		{
			name:        "a type at its budget is a COVERAGE limit",
			matched:     []string{"a", "b"},
			perType:     map[string]int{"a": 10, "b": 2},
			limit:       10,
			wantContain: []string{"per-type llm_call_log cap of 10 was reached for: a", "capped, not exhaustive"},
		},
		{
			name:        "a type with nothing is an ANSWER",
			matched:     []string{"a", "b"},
			perType:     map[string]int{"a": 4},
			limit:       10,
			wantContain: []string{"no llm_call_log rows carry agent_type: b", "relabelled 2026-07-26"},
		},
		{
			name:        "both facts render, separately",
			matched:     []string{"a", "b", "c"},
			perType:     map[string]int{"a": 10, "b": 0, "c": 1},
			limit:       10,
			wantContain: []string{"no llm_call_log rows carry agent_type: b", "cap of 10 was reached for: a"},
		},
		{
			name:      "limit<=0 means no cap, so nothing is AT one",
			matched:   []string{"a"},
			perType:   map[string]int{"a": 99},
			limit:     0,
			wantEmpty: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := callLogCoverageNotice(tc.matched, tc.perType, tc.limit)
			if tc.wantEmpty {
				if got != "" {
					t.Errorf("want empty notice, got %q", got)
				}
				return
			}
			for _, w := range tc.wantContain {
				if !strings.Contains(got, w) {
					t.Errorf("notice missing %q; got %q", w, got)
				}
			}
		})
	}
}

// TestCappedTypeNotice — the marker is "" when nothing was dropped, which is what
// keeps the negative control byte-identical.
func TestCappedTypeNotice(t *testing.T) {
	if got := cappedTypeNotice(nil, 5); got != "" {
		t.Errorf("no drops must render nothing, got %q", got)
	}
	got := cappedTypeNotice([]string{"x-agent", "y-agent"}, 3)
	for _, w := range []string{"2 further agent type(s)", "agent_state_cap=3", "x-agent, y-agent"} {
		if !strings.Contains(got, w) {
			t.Errorf("notice missing %q; got %q", w, got)
		}
	}
}
