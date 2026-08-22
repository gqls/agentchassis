// FILE: cmd/config-key-audit/cappedscheduleordering_test.go
//
// Tests for --capped-schedule-ordering (bugs_open/316).
//
// THE CENTRAL FIXTURE IS THE VERBATIM PRE-FIX QUERY, captured from live
// `agent_definitions` on 2026-08-22 before migration 552 was applied and saved
// alongside the lane docs as
// docs/agent_docs/docs024_key_docs_latest/bugfix_316_news_feed_ordering/
// PREFIX_find_news_sites_query_2026-08-22.sql.
//
// That is not decoration. Once the migration applies, the live row can no longer
// produce this check's POSITIVE control, and a detector whose only evidence is a
// post-fix zero has been silenced by its own author's action — a documented
// failure class in this estate. Keeping the motivating case as a fixture means
// the check can be re-proven to fire at any time, by anyone, for ever.
//
// The negative controls are the real siblings the fleet census turned up, not
// invented ones — each is a query this mode must NOT report, and each acquits on
// a DIFFERENT clause of the rule. A test suite whose negatives all fail the same
// way cannot tell you the conjunction is doing any work.
package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// The live query as it stood when bugs_open/316 was re-verified: alphabetical
// sort, trailing LIMIT 5, and two NOW() comparisons on content_sources.
const preFixFindNewsSites = `SELECT DISTINCT s.id::text as site_id, s.domain FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW()))) ORDER BY s.domain LIMIT 5`

// The query migration 552 installs. Ordering by a DERIVED column named for the
// due column it is computed from — this is the case an alias-blind check would
// wrongly report, so it is a first-class test rather than an afterthought.
const postFixFindNewsSites = `SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain, (SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) AS due_at FROM sites s JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'classification' AND ss.is_current = true AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id = s.id AND p.build_status = 'deployed') AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true) OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id = s.id AND cs.is_active = true AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))) q ORDER BY due_at ASC NULLS LAST, domain ASC LIMIT 5`

// meta-description-backfiller: alphabetical AND capped, and NOT a finding. Its
// candidate set is CONSUMED — a page that gains a meta description leaves the
// WHERE clause — and there is no NOW() predicate. This is the near-miss that
// forced the rule's central distinction.
const metaBackfillerQuery = `SELECT p.id, p.name FROM pages p JOIN page_components pc ON pc.page_id = p.id WHERE p.site_id = $1 AND p.status = 'active' AND COALESCE(p.meta_description, '') = '' GROUP BY p.id, p.name ORDER BY p.name LIMIT 25`

// model-directory-trigger: recurring, capped, but ORDER BY random(). Fair in
// expectation, never a fixed priority list. Not a finding.
const directoryTriggerQuery = `SELECT site_id, domain FROM (SELECT DISTINCT s.id::text AS site_id, s.domain FROM sites s WHERE s.updated_at <= NOW()) due ORDER BY random() LIMIT 12`

// agentWith builds a one-step fixture through the package's existing oneAgent
// decoder, so these tests exercise the SAME decode path the live export uses
// rather than a second hand-built construction of liveAgent.
func agentWith(t *testing.T, stepName, query string) []liveAgent {
	t.Helper()
	wf := map[string]interface{}{
		"steps": map[string]interface{}{
			stepName: map[string]interface{}{
				"action":       "query_database",
				"output_field": "rows_out",
				"config":       map[string]interface{}{"query": query},
			},
		},
	}
	b, err := json.Marshal(wf)
	if err != nil {
		t.Fatalf("fixture does not marshal: %v", err)
	}
	return oneAgent(t, string(b))
}

// TestFiresOnTheMotivatingCase is the POSITIVE CONTROL. If this ever goes quiet
// the check is not checking anything, whatever the live fleet reports.
func TestFiresOnTheMotivatingCase(t *testing.T) {
	got := findCappedScheduleOrdering(agentWith(t, "find_news_sites", preFixFindNewsSites))
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 finding on the pre-fix live query, got %d: %+v", len(got), got)
	}
	f := got[0]
	if f.RowCap != 5 {
		t.Errorf("row cap: got %d want 5", f.RowCap)
	}
	if !strings.Contains(f.OrderBy, "s.domain") {
		t.Errorf("order_by: got %q, expected it to name s.domain", f.OrderBy)
	}
	if !strings.Contains(f.DueCols, "next_fetch_at") {
		t.Errorf("due_columns: got %q, expected next_fetch_at", f.DueCols)
	}
}

// TestGoesQuietOnTheFix is the other half of the pair. Same binary, same rule,
// only the query changed — which is what makes a zero mean something.
func TestGoesQuietOnTheFix(t *testing.T) {
	got := findCappedScheduleOrdering(agentWith(t, "find_news_sites", postFixFindNewsSites))
	if len(got) != 0 {
		t.Fatalf("the fixed query must not be a finding, got %d: %+v", len(got), got)
	}
}

// TestNegativeControlsEachAcquitOnADifferentClause. If they all acquitted the
// same way, the conjunction would be untested and any one clause could be
// deleted without a test noticing.
func TestNegativeControlsEachAcquitOnADifferentClause(t *testing.T) {
	cases := []struct {
		name, query, why string
	}{
		{"consumed set, no clock predicate", metaBackfillerQuery, "clause (b): no NOW() comparison"},
		{"random ordering", directoryTriggerQuery, "clause (c): random() does not systematically starve"},
		{"uncapped", strings.Replace(preFixFindNewsSites, " LIMIT 5", "", 1), "clause (a): no trailing LIMIT"},
		{"LIMIT 1 fetch-one", strings.Replace(preFixFindNewsSites, "LIMIT 5", "LIMIT 1", 1), "clause (a): LIMIT 1 is excluded by QueryRowCap"},
		{"already due-ordered", strings.Replace(preFixFindNewsSites, "ORDER BY s.domain", "ORDER BY cs.next_fetch_at", 1), "clause (c): ordered by the schedule"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := findCappedScheduleOrdering(agentWith(t, "s", tc.query))
			if len(got) != 0 {
				t.Fatalf("expected clean (%s), got %+v", tc.why, got)
			}
		})
	}
}

// TestAbsentOrderByIsAFinding — an unordered capped query over a clock-refilled
// set returns rows in a stable heap/index order and starves identically. It just
// looks arbitrary instead of alphabetical, which is worse to diagnose, not
// better.
func TestAbsentOrderByIsAFinding(t *testing.T) {
	q := strings.Replace(preFixFindNewsSites, "ORDER BY s.domain ", "", 1)
	got := findCappedScheduleOrdering(agentWith(t, "s", q))
	if len(got) != 1 {
		t.Fatalf("a capped clock-fed query with NO ORDER BY must be a finding, got %d", len(got))
	}
	if got[0].OrderBy != "" {
		t.Errorf("order_by should be empty, got %q", got[0].OrderBy)
	}
}

// TestEffectiveCapReportsTheLoopInSeries. Reporting the query's LIMIT alone was
// measured to be misleading on the very agent this check was written for:
// find_news_sites caps at 5 and process_sites caps the consuming loop at 5, so
// raising only the LIMIT changes throughput by nothing while the cap-hit census
// stops reporting. The honest number has to come out of the check, not out of a
// reader's second query.
func TestEffectiveCapReportsTheLoopInSeries(t *testing.T) {
	wf := `{"steps":{
	  "find_news_sites":{"action":"query_database","output_field":"news_sites",
	    "config":{"query":` + strconv.Quote(strings.Replace(preFixFindNewsSites, "LIMIT 5", "LIMIT 10", 1)) + `}},
	  "process_sites":{"action":"loop",
	    "config":{"items_field":"news_sites.rows","max_iterations":5}}
	}}`
	agents := oneAgent(t, wf)
	got := findCappedScheduleOrdering(agents)
	if len(got) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(got))
	}
	if got[0].RowCap != 10 {
		t.Errorf("row_cap: got %d want 10 (the query's own LIMIT)", got[0].RowCap)
	}
	if got[0].EffectiveCap != 5 {
		t.Errorf("effective_cap: got %d want 5 — the loop caps the same fan-out in SERIES, "+
			"and reporting 10 would tell a reader that raising the LIMIT had worked when it had not",
			got[0].EffectiveCap)
	}
	if !strings.Contains(got[0].CappedBy, "max_iterations") {
		t.Errorf("effective_cap_set_by should name the loop, got %q", got[0].CappedBy)
	}
}

// TestEqualCapsInSeriesAreStillBothReported. The LIVE case has query LIMIT 5 and
// loop max_iterations 5 — equal, so the loop never "wins" and an
// effective-cap-only report would say "query LIMIT" and hide the second gate
// from exactly the reader about to raise the first. That reader would then
// measure no change in throughput and a cap-hit census that had gone quiet.
func TestEqualCapsInSeriesAreStillBothReported(t *testing.T) {
	wf := `{"steps":{
	  "find_news_sites":{"action":"query_database","output_field":"news_sites",
	    "config":{"query":` + strconv.Quote(preFixFindNewsSites) + `}},
	  "process_sites":{"action":"loop",
	    "config":{"items_field":"news_sites.rows","max_iterations":5}}
	}}`
	got := findCappedScheduleOrdering(oneAgent(t, wf))
	if len(got) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(got))
	}
	if got[0].EffectiveCap != 5 {
		t.Errorf("effective_cap: got %d want 5", got[0].EffectiveCap)
	}
	if !strings.Contains(got[0].ConsumingLoop, "max_iterations=5") {
		t.Fatalf("an EQUAL second cap must still be reported, got consuming_loop=%q", got[0].ConsumingLoop)
	}
}

// TestCleanRunSummaryStatesItsScope — "0 findings" is meaningless without the
// denominator, and this estate has been bitten by clean reports that were clean
// because they were blind.
func TestCleanRunSummaryStatesItsScope(t *testing.T) {
	s := cappedOrderingRunSummary(177, 0, nil)
	if !strings.Contains(s, "177") {
		t.Errorf("a clean summary must state how many agents it scanned: %q", s)
	}
	if !strings.Contains(strings.ToUpper(s), "CLEAN") {
		t.Errorf("a clean summary must say so plainly: %q", s)
	}
	if u := cappedOrderingRunSummary(177, 3, nil); !strings.Contains(u, "3 agent row(s) failed to decode") {
		t.Errorf("undecoded rows must be reported beside the zero: %q", u)
	}
}

// --- MUTATION TESTS -------------------------------------------------------
//
// A test that only feeds the check its expected inputs proves the check runs, not
// that it DISCRIMINATES. Each mutation below changes the FIXTURE in the one way
// that should flip the verdict; a mutation that fails to flip it means the clause
// is not load-bearing and could be deleted silently.
//
// ⚠ These mutate the QUERY, not the check. Mutating the check and re-running its
// own tests is the shape that can pass by hitting a guard in series; mutating the
// input drives the real predicate end to end.

func TestMutation_RemovingTheClockPredicateAcquits(t *testing.T) {
	// The ONLY change: the two NOW() comparisons become a static bound. Everything
	// else — the alphabetical sort, the LIMIT 5 — is untouched. If this still
	// reports, clause (b) is doing nothing and the check would flag every capped
	// alphabetical query in the fleet, including the consumed ones.
	mutated := strings.ReplaceAll(preFixFindNewsSites, "<= NOW()", "<= '2020-01-01'::timestamptz")
	if strings.Contains(mutated, "NOW()") {
		t.Fatalf("mutation did not apply — the fixture no longer contains the text it patches")
	}
	if got := findCappedScheduleOrdering(agentWith(t, "s", mutated)); len(got) != 0 {
		t.Fatalf("clause (b) is not load-bearing: a query with no clock predicate still reported %+v", got)
	}
}

func TestMutation_OrderingByTheDueColumnAcquits(t *testing.T) {
	// The ONLY change: the sort key. Cap and clock predicate untouched.
	mutated := strings.Replace(preFixFindNewsSites, "ORDER BY s.domain", "ORDER BY cs.next_fetch_at ASC NULLS FIRST", 1)
	if mutated == preFixFindNewsSites {
		t.Fatalf("mutation did not apply")
	}
	if got := findCappedScheduleOrdering(agentWith(t, "s", mutated)); len(got) != 0 {
		t.Fatalf("clause (c) is not load-bearing: a due-ordered query still reported %+v", got)
	}
}

func TestMutation_RemovingTheCapAcquits(t *testing.T) {
	// The ONLY change: no trailing LIMIT. This also pins the boundary to
	// QueryRowCap rather than to a private copy of it.
	mutated := strings.Replace(preFixFindNewsSites, " LIMIT 5", "", 1)
	if mutated == preFixFindNewsSites {
		t.Fatalf("mutation did not apply")
	}
	if got := findCappedScheduleOrdering(agentWith(t, "s", mutated)); len(got) != 0 {
		t.Fatalf("clause (a) is not load-bearing: an uncapped query still reported %+v", got)
	}
}

func TestMutation_ASubqueryLimitIsNotACap(t *testing.T) {
	// Moving the LIMIT inside the derived table bounds a different set. This is
	// QueryRowCap's documented exclusion, and the point of the test is that this
	// mode INHERITS it rather than re-deciding it.
	mutated := strings.Replace(preFixFindNewsSites, "ORDER BY s.domain LIMIT 5",
		"ORDER BY s.domain) q_outer SELECT 1 FROM (SELECT 1 LIMIT 5) x", 1)
	if got := findCappedScheduleOrdering(agentWith(t, "s", mutated)); len(got) != 0 {
		t.Fatalf("a non-trailing LIMIT must not count as a cap, got %+v", got)
	}
}

// TestAliasResolutionIsExactNotGreedy. Alias resolution is the acquitting path,
// so a window that is too wide is BLINDNESS, not noise: an unrelated `AS x`
// sitting near a due column would acquit a genuinely starving query.
//
// The positive half is TestGoesQuietOnTheFix (a derived `due_at` bound to an
// expression over `next_fetch_at` must acquit). This is the negative half: an
// alias bound to something else must NOT acquit, even though the due column
// appears elsewhere in the same statement.
func TestAliasResolutionIsExactNotGreedy(t *testing.T) {
	// `sort_key` is bound to s.domain. next_fetch_at appears in the WHERE clause
	// but is not part of sort_key's defining expression, so ordering by sort_key
	// is still ordering by the alphabet and must remain a finding.
	q := `SELECT site_id, sort_key FROM (SELECT s.id::text AS site_id, s.domain AS sort_key ` +
		`FROM sites s WHERE EXISTS (SELECT 1 FROM content_sources cs ` +
		`WHERE cs.site_id = s.id AND cs.next_fetch_at <= NOW())) q ORDER BY sort_key ASC LIMIT 5`

	got := findCappedScheduleOrdering(agentWith(t, "s", q))
	if len(got) != 1 {
		t.Fatalf("an alias bound to the DOMAIN must not acquit just because a due column "+
			"appears elsewhere in the statement — expected 1 finding, got %d: %+v", len(got), got)
	}
}

// TestSelectItemBeforeIsParenAware pins the mechanism the case above rests on.
// A naive "text back to the previous comma" stops inside COALESCE(...) and loses
// the column that justifies the alias — which is how the check came to report
// its own fix as a defect the first time it was run.
func TestSelectItemBeforeIsParenAware(t *testing.T) {
	q := `select a, (select min(coalesce(cs.next_fetch_at, 'x')) from t) as due_at from u`
	i := strings.Index(q, "as due_at")
	item := selectItemBefore(q, i)
	if !strings.Contains(item, "next_fetch_at") {
		t.Fatalf("the select-item window must reach past the comma inside COALESCE; got %q", item)
	}
	if strings.Contains(item, "select a") {
		t.Fatalf("the window must stop at the depth-0 comma, not run to the start; got %q", item)
	}
}
