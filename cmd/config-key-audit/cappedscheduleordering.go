// FILE: cmd/config-key-audit/cappedscheduleordering.go
//
// --capped-schedule-ordering (bugs_open/316). Which capped query_database step
// picks its work by a STATIC sort over a candidate set that the CLOCK refills?
//
// `content-feed-trigger.find_news_sites` ended `ORDER BY s.domain LIMIT 5` over
// the nine news-feed sites whose sources were due. The runs are 6-hourly and the
// sort is stable, so the same five names won every time more than five were in
// contention. Measured 2026-08-22 over five consecutive cap-hitting runs: the
// alphabetically LAST eligible site was selected ZERO times while continuously
// due, reaching 31 hours of staleness on a 6-hour cadence — 419% of its own
// configured interval — with 128 deployed pages and 5 of 5 sources due at every
// run it missed.
//
// WHY LCO-009 CANNOT SEE THIS, even though it looks directly at the same step.
// That check counts rows and warns when a result reaches its own ceiling. The
// row count is IDENTICAL whether the ordering is fair or not: five of nine is
// five of nine either way. It correctly reported "this step hit its cap" on
// every run for days, and that statement was true and insufficient — the defect
// is not the size of the cut, it is that the cut always falls in the same place.
//
// THE DISTINCTION THIS MODE ENCODES, and it is the whole rule. A cap over a
// candidate set that is CONSUMED — a row leaves once served — is only a batching
// delay, and "coverage is eventual" is true of it.
// `meta-description-backfiller.load_pages_missing_meta` is `ORDER BY p.name
// LIMIT 25`, alphabetical AND capped, and is deliberately NOT a finding: a page
// that gains a meta description leaves `WHERE COALESCE(p.meta_description,”)=”`
// and cannot be starved. A cap over a set REPLENISHED BY THE CLOCK — rows never
// leave, they merely acquire a later due-time — starves its tail permanently
// under a static sort. Only the second is reported, and the `NOW()` predicate is
// what tells them apart in config alone.
//
// THE PLATFORM ALREADY HAS THE RIGHT IDIOM, one layer down: both
// dispatch_feed_sources_action.go:101 and feed_actions.go:1016 order feed work
// `ORDER BY next_fetch_at ASC NULLS FIRST`. Those select SOURCES within a site.
// The SITE-selection query lives in config rather than in Go, and config is the
// one layer that skipped the convention — which is exactly why this check has to
// read live agent_definitions rather than the repo.
//
// "CAPPED" IS NOT REDEFINED HERE. It is actions.QueryRowCap, the same function
// LCO-009's runtime WARN uses, so the two checks cannot disagree about which
// steps are even capped — including its exclusions (LIMIT 1, LIMIT 0, a
// parameterised `LIMIT $2`, a non-trailing limit, tolerated trailing comments).
// A second hand-written regex here would be bugs_open/144's shape precisely.
//
// Same validation.WalkSteps traversal as every other mode in this binary, top
// level and nested, for the same reason.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/validation"
)

// cappedOrderingFinding is one capped, clock-replenished step whose ordering
// decides who waits by something other than how long they have waited.
type cappedOrderingFinding struct {
	Agent   string `json:"agent"`
	Path    string `json:"path"`
	RowCap  int    `json:"row_cap"`     // the query's own trailing LIMIT
	OrderBy string `json:"order_by"`    // the top-level ORDER BY, "" when absent
	DueCols string `json:"due_columns"` // the clock-relative columns it ignores
	// EffectiveCap is min(RowCap, the consuming loop's max_iterations), and it is
	// reported because RowCap alone is misleading — see effectiveCapFor.
	EffectiveCap int    `json:"effective_cap"`
	CappedBy     string `json:"effective_cap_set_by"`
	// ConsumingLoop names the loop step that caps the same fan-out again, and is
	// reported WHENEVER one exists — not only when it is the smaller of the two.
	// When the two caps are EQUAL (find_news_sites LIMIT 5 / process_sites
	// max_iterations 5, which is the live case) reporting only the winner would
	// hide the second cap from the very reader about to raise the first.
	ConsumingLoop string `json:"consuming_loop,omitempty"`
}

// dueColumnRe finds a clock-relative predicate: a column compared against NOW()
// or CURRENT_TIMESTAMP. Its capture is the COLUMN, which is what the ORDER BY is
// then judged against.
//
// This is the "replenished by the clock" test, and it is a TEXT test on purpose:
// the property is visible in config, which is what makes the whole class
// checkable offline before a run pays for it.
var dueColumnRe = regexp.MustCompile(`(?is)([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)?)\s*(?:<=|<|>=|>)\s*(?:NOW\s*\(\s*\)|CURRENT_TIMESTAMP)`)

// topLevelOrderByRe takes the LAST ORDER BY in the statement, which is the one
// that orders the returned rows. An ORDER BY inside a subquery or a CTE orders
// something else, and judging on it would acquit a query whose OUTER sort is
// static. Ends at a trailing LIMIT/OFFSET/FETCH or the end of the statement.
var topLevelOrderByRe = regexp.MustCompile(`(?is)ORDER\s+BY\s+(.*?)(?:\s+LIMIT\b|\s+OFFSET\b|\s+FETCH\b|\s*;?\s*$)`)

// nondeterministicOrderRe recognises orderings that do not systematically
// starve. `random()` is fair only in expectation — a site can lose several draws
// running — but it cannot produce the fixed, reproducible priority list this
// mode exists to catch, so it is not reported. The sibling
// model-directory-trigger uses it and has never shown the pattern.
var nondeterministicOrderRe = regexp.MustCompile(`(?i)\brandom\s*\(\s*\)|\bnewid\s*\(\s*\)`)

// identifierRe pulls the bare identifiers out of an ORDER BY clause so it can be
// asked whether it names a due column.
//
// ⚠ NOTE WHICH ERROR IS THE SAFE ONE, because an earlier draft of this file had
// it backwards in a comment. For a DETECTOR, over-acquitting is blindness and
// over-convicting is noise a human clears with one query. So the alias
// resolution below is deliberately EXACT (a real `AS` binding whose expression
// mentions the due column) rather than a loose substring match. The loose
// version was tried first, and it was both unprincipled and useless: it could
// not relate `due_at` to `next_fetch_at`, which is precisely the alias the fix
// for bugs_open/316 introduces. The test for the fixed query caught it.
var identifierRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// aliasRe finds `AS <name>` bindings. Each one is a candidate name for a derived
// column, which an ORDER BY may legitimately sort by instead of naming the
// underlying due column.
var aliasRe = regexp.MustCompile(`(?is)\bAS\s+([A-Za-z_][A-Za-z0-9_]*)`)

// selectItemBefore returns the select-item expression that an `AS` at asPos
// binds, by walking BACKWARDS with paren awareness: stop at a comma seen at
// depth 0, or when we step out of the enclosing parenthesis.
//
// Paren awareness is the whole point and a naive "text back to the previous
// comma" does not work here. The bugs_open/316 fix binds
// `(SELECT min(COALESCE(cs.next_fetch_at, '-infinity'::timestamptz)) FROM ...) AS due_at`,
// whose nearest preceding comma is INSIDE the COALESCE — so the naive window
// excludes the very column that justifies the alias, and the check would report
// its own fix as a defect.
func selectItemBefore(q string, asPos int) string {
	depth := 0
	i := asPos - 1
	for ; i >= 0; i-- {
		switch q[i] {
		case ')':
			depth++
		case '(':
			if depth == 0 {
				// Stepped out of the enclosing expression.
				return q[i+1 : asPos]
			}
			depth--
		case ',':
			if depth == 0 {
				return q[i+1 : asPos]
			}
		}
	}
	return q[0:asPos]
}

// dueAliasesIn returns the alias names that are BOUND to an expression
// mentioning one of the due columns — so an ORDER BY naming the alias is
// ordering by the schedule, just through a derived column.
func dueAliasesIn(query string, dueCols []string) map[string]bool {
	aliases := map[string]bool{}
	lowerQ := strings.ToLower(query)
	for _, loc := range aliasRe.FindAllStringSubmatchIndex(query, -1) {
		alias := strings.ToLower(query[loc[2]:loc[3]])
		expr := strings.ToLower(selectItemBefore(lowerQ, loc[0]))
		for _, c := range dueCols {
			if strings.Contains(expr, c) {
				aliases[alias] = true
				break
			}
		}
	}
	return aliases
}

// lastSegment reduces "cs.next_fetch_at" to "next_fetch_at" so that an ORDER BY
// naming the column under a different alias — or unqualified, as a derived
// column often is — still counts as ordering by the schedule. The fixed query
// for bugs_open/316 orders by a derived `due_at` computed from `next_fetch_at`,
// and an alias-blind test would have called the FIX a finding.
func lastSegment(col string) string {
	if i := strings.LastIndex(col, "."); i >= 0 {
		return col[i+1:]
	}
	return col
}

// dueColumnsIn returns the distinct clock-compared columns in a query, by their
// last segment, sorted. Empty means the query is not clock-replenished and is
// therefore out of scope however it is ordered.
func dueColumnsIn(query string) []string {
	seen := map[string]bool{}
	for _, m := range dueColumnRe.FindAllStringSubmatch(query, -1) {
		c := strings.ToLower(lastSegment(m[1]))
		// A comparison against NOW() whose left side is a function call or a
		// literal produces no useful column name; skip the obvious noise.
		if c == "" || c == "now" || c == "current_timestamp" {
			continue
		}
		seen[c] = true
	}
	out := make([]string, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// topLevelOrderBy returns the outermost ORDER BY clause, or "" when the query
// has none.
func topLevelOrderBy(query string) string {
	ms := topLevelOrderByRe.FindAllStringSubmatch(query, -1)
	if len(ms) == 0 {
		return ""
	}
	return strings.TrimSpace(ms[len(ms)-1][1])
}

// orderByNamesADueColumn reports whether the ordering is anchored to the
// schedule — i.e. it mentions (an alias of) a column the query compares against
// NOW(). This is the acquitting condition, and matching an identifier loosely is
// the safe error.
func orderByNamesADueColumn(query, orderBy string, dueCols []string) bool {
	if orderBy == "" {
		return false
	}
	idents := map[string]bool{}
	for _, id := range identifierRe.FindAllString(strings.ToLower(orderBy), -1) {
		idents[id] = true
	}
	// Direct: the ORDER BY names the due column itself (any alias qualifier is
	// already stripped, since identifierRe splits on the dot).
	for _, c := range dueCols {
		if idents[c] {
			return true
		}
	}
	// Indirect: the ORDER BY names a derived column whose defining expression
	// mentions a due column.
	for alias := range dueAliasesIn(query, dueCols) {
		if idents[alias] {
			return true
		}
	}
	return false
}

// effectiveCapFor resolves the cap a fan-out ACTUALLY runs at.
//
// ⚠ REPORTING THE QUERY'S LIMIT ALONE IS MISLEADING, and this is not
// hypothetical: content-feed-trigger caps find_news_sites at LIMIT 5 and then
// caps the loop that consumes it at max_iterations 5. Raising only the LIMIT
// changes throughput by NOTHING — the query returns 10 rows and the loop
// processes the first 5 — while the cap-hit census (which measures the step's
// own output length) stops reporting a cap hit. The instrument would show relief
// that never happened, which is the worst possible outcome for a check.
//
// A loop cap is NOT itself a finding and does not enter the predicate. It is
// reported so that a reader who acts on the number acts on the honest one.
func effectiveCapFor(steps map[string]models.Step, producer models.Step, rowCap int) (eff int, setBy, consumingLoop string) {
	field := producer.OutputField
	if field == "" {
		return rowCap, "query LIMIT", ""
	}
	eff, setBy = rowCap, "query LIMIT"
	for path, s := range steps {
		if s.Action != "loop" {
			continue
		}
		itemsField, _ := s.Config["items_field"].(string)
		if itemsField == "" {
			continue
		}
		// The loop consumes this producer when its items_field IS the output
		// field or dots into it ("news_sites.rows").
		if itemsField != field && !strings.HasPrefix(itemsField, field+".") {
			continue
		}
		n, ok := toInt(s.Config["max_iterations"])
		if !ok || n <= 0 {
			continue
		}
		// Recorded whether or not it wins, so an EQUAL second cap is still
		// visible. Only the strictly smaller one changes the effective number.
		consumingLoop = fmt.Sprintf("%s max_iterations=%d", path, n)
		if n < eff {
			eff, setBy = n, "loop max_iterations at "+path
		}
	}
	return eff, setBy, consumingLoop
}

// toInt reads a JSON-decoded number, which arrives as float64 through
// encoding/json but may be an int through other decoders.
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	}
	return 0, false
}

// findCappedScheduleOrdering is the pure check (I/O split off, this binary's
// convention).
//
// A step is a finding when ALL THREE hold:
//
//	(a) CAPPED      — actions.QueryRowCap finds a trailing literal LIMIT n>=2.
//	(b) CLOCK-FED   — the query compares a column against NOW()/CURRENT_TIMESTAMP,
//	                  so its candidate set refills with time instead of draining
//	                  as work is done.
//	(c) STATIC SORT — the top-level ORDER BY names no due column and is not
//	                  random(). NO ORDER BY AT ALL COUNTS: an unordered capped
//	                  query returns rows in a heap/index order that is stable in
//	                  practice, which starves in exactly the same way while
//	                  looking arbitrary rather than alphabetical.
//
// (a) on its own is LCO-009's question, not this one. The conjunction is the
// rule, and every clause can acquit.
//
// STATED FALSE NEGATIVES, here rather than in a later bug file: a parameterised
// `LIMIT $n` and a LIMIT inside a subquery (both inherited from QueryRowCap, and
// both deliberate there); a due predicate that reaches the query as a bound
// parameter rather than a literal NOW(); and an ordering that is fair in
// expectation but not in fact, which `random()` is.
//
// STATED FALSE POSITIVES: a NOW() comparison used as a sanity bound rather than
// a schedule, and a static sort over a set that in fact drains for a reason the
// SQL cannot show. Both cost a human one query; the silence they replace cost
// one site 31 hours of staleness for as long as nobody looked.
func findCappedScheduleOrdering(agents []liveAgent) []cappedOrderingFinding {
	findings := []cappedOrderingFinding{}

	for _, agent := range agents {
		steps := map[string]models.Step{}
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			steps[path] = step
		})

		for path, step := range steps {
			if step.Action != "query_database" {
				continue
			}
			query, _ := step.Config["query"].(string)
			if query == "" {
				continue
			}

			// (a) capped — the SAME definition the runtime WARN uses.
			rowCap, capped := actions.QueryRowCap(query)
			if !capped {
				continue
			}

			// (b) clock-replenished.
			dueCols := dueColumnsIn(query)
			if len(dueCols) == 0 {
				continue
			}

			// (c) statically ordered.
			orderBy := topLevelOrderBy(query)
			if nondeterministicOrderRe.MatchString(orderBy) {
				continue
			}
			if orderByNamesADueColumn(query, orderBy, dueCols) {
				continue
			}

			eff, setBy, loop := effectiveCapFor(steps, step, rowCap)
			findings = append(findings, cappedOrderingFinding{
				Agent:         agent.Type,
				Path:          path,
				RowCap:        rowCap,
				OrderBy:       orderBy,
				DueCols:       strings.Join(dueCols, ", "),
				EffectiveCap:  eff,
				CappedBy:      setBy,
				ConsumingLoop: loop,
			})
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

// emitCappedScheduleOrdering is the I/O half: DB route when PG_CLIENTS_HOST is
// set (the CronJob), stdin otherwise (a session at a terminal). Same refusals as
// every mode here — a DB error when the env asked for the DB is fatal, and a
// 0-agent fleet is never reported as clean.
func emitCappedScheduleOrdering(args []string) {
	report := false
	for _, a := range args {
		switch a {
		case "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --capped-schedule-ordering: unknown argument %q\n", a)
			os.Exit(2)
		}
	}

	var agents []liveAgent
	var failed int
	var err error
	if db, derr := dbConn(); derr != nil || db != nil {
		if derr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --capped-schedule-ordering: %v\n", derr)
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDB(db, "--capped-schedule-ordering")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		raw, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --capped-schedule-ordering: reading stdin: %v\n", rerr)
			os.Exit(2)
		}
		agents, failed, err = decodeLiveAgents(raw, "--capped-schedule-ordering")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --capped-schedule-ordering: 0 live agents decoded — refusing to print a clean report over an empty fleet.\n")
		os.Exit(2)
	}

	findings := findCappedScheduleOrdering(agents)

	out := map[string]interface{}{
		"agents_scanned":   len(agents),
		"agents_undecoded": failed,
		"findings":         findings,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)

	if report {
		writeDocNote("capped-schedule-ordering",
			cappedOrderingRunSummary(len(agents), failed, findings),
			"config-integrity", "capped_schedule_ordering_check")
	}

	if len(findings) > 0 {
		fmt.Fprintf(os.Stderr,
			"%d capped step(s) select clock-replenished work by a static sort: the same rows win every run and the tail starves (bugs_open/316).\n",
			len(findings))
		os.Exit(1)
	}
}

// cappedOrderingRunSummary is the doc_notes body. It states the SCOPE as well as
// the result, because "0 findings over 3 agents" and "0 findings over 177" have
// opposite meanings — and one row is written per run, clean runs included, so a
// MISSING row means the job did not run and must not read as "nothing is wrong".
func cappedOrderingRunSummary(scanned, undecoded int, findings []cappedOrderingFinding) string {
	var b strings.Builder
	if len(findings) == 0 {
		fmt.Fprintf(&b, "capped-schedule-ordering check CLEAN: across %d live agents, every capped query_database step whose candidate set is refilled by the clock orders its work by the schedule rather than by a static sort.", scanned)
	} else {
		fmt.Fprintf(&b, "capped-schedule-ordering: %d capped step(s) across %d live agents pick clock-replenished work by a static sort, so the same rows win every run and the tail starves (bugs_open/316): ", len(findings), scanned)
		for i, f := range findings {
			if i > 0 {
				b.WriteString("; ")
			}
			ob := f.OrderBy
			if ob == "" {
				ob = "(no ORDER BY)"
			}
			fmt.Fprintf(&b, "%s %s — ORDER BY %s, cap %d (effective %d, set by %s), ignoring due column(s) %s",
				f.Agent, f.Path, ob, f.RowCap, f.EffectiveCap, f.CappedBy, f.DueCols)
			if f.ConsumingLoop != "" {
				fmt.Fprintf(&b, " [also capped in series by %s — raising one alone changes nothing]", f.ConsumingLoop)
			}
		}
		b.WriteString(".")
	}
	if undecoded > 0 {
		fmt.Fprintf(&b, " %d agent row(s) failed to decode and were not scanned.", undecoded)
	}
	return b.String()
}
