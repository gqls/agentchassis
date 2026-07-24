// FILE: platform/orchestration/actions/diagnose_code_lookup_action.go
//
// F2.3b(c) of the diagnosis→fix loop: the council's CODE-shaped verify tier.
//
// Why it exists (real case, 2026-07-17, run ca064df2 / fix_correlation
// e505f70f): the bug-historian's blocking objection hinged on ONE code fact —
// "does the codebase have other LLM provider adapters?" — and the council's
// only verify step (diagnose_run_checks) executes SQL against the platform DB,
// not questions about the codebase. Three revise rounds burned without the one
// answer that would have settled the round; the router escalated (honestly).
// The answer existed the whole time in the code_symbols index:
//
//	SELECT path, symbol FROM code_symbols WHERE symbol LIKE '%GenerateText%'
//	→ anthropic.go AND ollama.go.
//
// This action gives reviewers a `code_checks` channel beside their SQL
// `checks`: structured code questions answered from the code_symbols index —
// a plain DB read, so it runs in the SHARED chassis pod, which deliberately
// never holds the GitHub read token (only spawned repo-cloning pods do; see
// analyse_repo_local_action.go). No tarball fetch, no token, no new trust
// surface: the index is refreshed by index-orchestrator and carries commit_sha
// per row, which is rendered so staleness is visible, never hidden.
//
// Wire format (mirrors checks:[{sql,why}] — one convention per tier):
//
//	code_checks: [{"kind": "symbol|content|ls", "query": "...", "why": "..."}]
//	  symbol  — match against the symbol name (ILIKE, e.g. "GenerateText" or
//	            "(*OllamaClient).%"); returns path, symbol, signature, lines.
//	  content — match against symbol source bodies (ILIKE over the
//	            trigram-indexed content column, e.g. "%stop_reason%");
//	            returns path, symbol + a capped matching-line excerpt.
//	  ls      — path prefix listing (e.g. "platform/aiservice/"); returns the
//	            distinct indexed paths under it.
//
// Every kind is READ-ONLY by construction — the SQL is written HERE, the
// reviewer supplies only a pattern; nothing model-written is executed.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/pkg/diagnose"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// codeIndexStaleAfter is when the code_symbols index is loud-flagged as STALE in
// every rendered answer. The index is refreshed by the `code-index-refresh`
// scheduled task every 24h (SEED_code_index_refresh_cadence.sql, bugs_open/059);
// 48h therefore means "at least one scheduled refresh was missed". Deliberately a
// const, not step config: it is one platform-wide fact coupled to that cadence
// row, and the banner always prints the ACTUAL age, so a reader can judge for
// themselves whatever this threshold says.
const codeIndexStaleAfter = 48 * time.Hour

// codeIndexFreshness reads the index's high-water mark and renders the freshness
// banner both answer tiers prepend to their output. One query per action run.
// Never fatal: an error degrades to an "unknown freshness" note (fail open) —
// the guard must not break the lookup it qualifies.
func codeIndexFreshness(ctx context.Context, db *sql.DB) string {
	var sha string
	var updatedAt time.Time
	err := db.QueryRowContext(ctx,
		`SELECT COALESCE(commit_sha,''), updated_at FROM code_symbols
		 ORDER BY updated_at DESC LIMIT 1`).Scan(&sha, &updatedAt)
	if err == sql.ErrNoRows {
		return freshnessBanner("", time.Time{}, time.Now(), codeIndexStaleAfter, nil)
	}
	return freshnessBanner(sha, updatedAt, time.Now(), codeIndexStaleAfter, err)
}

// freshnessBanner is the pure decision+formatting half of codeIndexFreshness,
// split out so the stale/empty/error branches are unit-testable without a DB
// (verify-the-failing-branch: a guard whose job is to catch a fault must be
// SEEN catching it). bugs_open/059's read-time half: a stale index answers
// "absent" identically to a genuine absence, so at read time the answer must
// carry its own freshness — an empty result against a stale index is UNKNOWN,
// not evidence of absence, and nothing upstream can add that qualifier later.
func freshnessBanner(sha string, updatedAt, now time.Time, staleAfter time.Duration, queryErr error) string {
	if queryErr != nil {
		return fmt.Sprintf("(index freshness UNKNOWN — %v; treat every empty answer below as unknown, not absent)\n", queryErr)
	}
	if updatedAt.IsZero() {
		return "!! CODE INDEX EMPTY — no symbols indexed at all. Every answer below is UNKNOWN, not absent. Run index-orchestrator before trusting any of this. !!\n"
	}
	age := now.Sub(updatedAt)
	if age > staleAfter {
		return fmt.Sprintf(
			"!! CODE INDEX STALE: last refreshed %s ago (%s) at commit %s — code newer than that is NOT in the index, so a 'no matches' answer below may mean NOT YET INDEXED, not absent. Run index-orchestrator to refresh. !!\n",
			formatAge(age), updatedAt.Format("2006-01-02"), shortSHA(sha))
	}
	return fmt.Sprintf("(index freshness: refreshed %s ago at commit %s)\n", formatAge(age), shortSHA(sha))
}

// formatAge renders a duration at banner altitude: days for old, hours/minutes
// for recent. Precision beyond this is noise in a provenance line.
func formatAge(d time.Duration) string {
	switch {
	case d >= 48*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
}

var DiagnoseCodeLookupInputSpec = datahelpers.ActionInputSpec{
	Optional: []string{
		"code_check_fields", "max_checks", "row_cap", "excerpt_chars", "repo",
	},
	Defaults: map[string]interface{}{
		"max_checks":    8,
		"row_cap":       40,
		"excerpt_chars": 400,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("diagnose_code_lookup", DiagnoseCodeLookupInputSpec)
}

// codeCheck is one reviewer-attached code question.
type codeCheck struct {
	Kind  string
	Query string
	Why   string
}

// DiagnoseCodeLookupAction answers the reviewers' code_checks from the
// code_symbols index and hands the results to the next repropose.
func DiagnoseCodeLookupAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config
	logger := params.Logger.With(zap.String("action", "diagnose_code_lookup"))
	if params.ExecutionContext != nil && params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("diagnose_code_lookup: no DB handle")
	}

	fields := configStringSlice(config, "code_check_fields", nil)
	if len(fields) == 0 {
		return nil, fmt.Errorf("no code_check_fields configured — a code-verify step with nowhere to look is a wiring mistake")
	}

	var checks []codeCheck
	for _, f := range fields {
		checks = append(checks, codeChecksFromCollected(params.CollectedData, f)...)
	}
	// Dedup identical (kind, query) checks BEFORE the cap. Reviewers reviewing
	// the same plan independently ask the same question — the first live run
	// (90e989d5, 2026-07-17) requested 13 checks of which several were exact
	// duplicates (two "content:stop_reason", two "content:done_reason"), and the
	// cap then dropped 5 DISTINCT questions to make room for repeats. Dedup
	// reclaims that budget for genuinely different questions.
	checks = dedupCodeChecks(checks)

	maxChecks := datahelpers.GetIntField(config, "max_checks", 8)
	dropped := 0
	if maxChecks > 0 && len(checks) > maxChecks {
		dropped = len(checks) - maxChecks
		checks = checks[:maxChecks]
	}

	if len(checks) == 0 {
		logger.Info("diagnose_code_lookup: reviewers asked no code questions")
		return map[string]interface{}{
			"results_text":   "(reviewers asked no code_checks this round)",
			"checks_run":     0,
			"checks_dropped": 0,
		}, nil
	}

	rowCap := datahelpers.GetIntField(config, "row_cap", 40)
	excerptChars := datahelpers.GetIntField(config, "excerpt_chars", 400)
	repoFilter := datahelpers.GetStringField(config, "repo", "")

	var b strings.Builder
	b.WriteString("Code questions your reviewers asked, answered from the code_symbols index\n")
	b.WriteString("(an INDEXED snapshot — each answer names its commit_sha; treat a stale or\nempty answer as 'unknown', not 'absent'):\n")
	// The read-time freshness guard (bugs_open/059): the header above SAYS to
	// treat a stale answer as unknown, but until now nothing COMPUTED staleness,
	// so a reader had no way to apply the rule. One query; loud when stale.
	b.WriteString(codeIndexFreshness(ctx, params.DB))
	run := 0
	for i, c := range checks {
		fmt.Fprintf(&b, "\n[code_check %d] kind=%s query=%q — %s\n", i+1, c.Kind, c.Query, c.Why)
		if err := answerCodeCheck(ctx, params.DB, c, repoFilter, rowCap, excerptChars, &b); err != nil {
			fmt.Fprintf(&b, "  (lookup failed: %v)\n", err)
			continue
		}
		run++
	}
	if dropped > 0 {
		fmt.Fprintf(&b, "\n> %d further code_check(s) dropped (max_checks=%d) — coverage was capped, not complete.\n", dropped, maxChecks)
	}

	logger.Info("diagnose_code_lookup: answered reviewer code checks",
		zap.Int("checks_run", run),
		zap.Int("checks_dropped", dropped),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"results_text":   b.String(),
		"checks_run":     run,
		"checks_dropped": dropped,
	}, nil
}

// codeChecksFromCollected extracts code_checks entries from a collected-data
// field (dot path to a []interface{} of {kind, query, why} maps). Malformed
// entries are skipped, never fatal — a reviewer that fumbles the format loses
// that check, not the run. Kinds are validated here (allowlist) so an
// unrecognised kind is dropped loudly in the caller's rendering, not guessed.
func codeChecksFromCollected(collected map[string]interface{}, field string) []codeCheck {
	raw := datahelpers.ExtractNestedField(collected, field)
	list, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var out []codeCheck
	for _, e := range list {
		m, ok := e.(map[string]interface{})
		if !ok {
			continue
		}
		kind, _ := m["kind"].(string)
		query, _ := m["query"].(string)
		why, _ := m["why"].(string)
		kind = strings.ToLower(strings.TrimSpace(kind))
		query = strings.TrimSpace(query)
		if query == "" || !validCodeCheckKind(kind) {
			continue
		}
		out = append(out, codeCheck{Kind: kind, Query: query, Why: why})
	}
	return out
}

func validCodeCheckKind(kind string) bool {
	switch kind {
	case "symbol", "content", "ls":
		return true
	}
	return false
}

// answerCodeCheck runs ONE check against code_symbols. The SQL is fixed per
// kind; the reviewer's query arrives only as a bind parameter — nothing
// model-written is executed as SQL (contrast run_checks, whose whole point is
// model-written SELECTs under containment; here the containment is that the
// model writes no SQL at all).
func answerCodeCheck(ctx context.Context, db *sql.DB, c codeCheck, repoFilter string, rowCap, excerptChars int, b *strings.Builder) error {
	switch c.Kind {
	case "symbol":
		// Match every identifier token in the query as an AND of substrings,
		// NOT the raw string. A reviewer writes Go methods as "Type.Method"
		// (e.g. "OllamaClient.GenerateText"), but the index stores the receiver
		// form "(*OllamaClient).GenerateText" — a raw ILIKE '%Type.Method%'
		// then misses it (false negative on run 90e989d5, 2026-07-17: the
		// Ollama-adapter check came back empty and a reviewer nearly read that
		// as "no such symbol"). Splitting on non-identifier chars and requiring
		// each token as a substring matches both forms and is still anchored to
		// the symbol column (never the body).
		clause, args := symbolTokenClause(c.Query, repoFilter, rowCap)
		rows, err := db.QueryContext(ctx, `
			SELECT path, symbol, COALESCE(signature,''), COALESCE(line_start,0), COALESCE(line_end,0), COALESCE(commit_sha,'')
			FROM code_symbols
			WHERE `+clause+`
			ORDER BY path, symbol
			LIMIT `+fmt.Sprintf("$%d", len(args)+1), append(args, rowCap)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		n := 0
		for rows.Next() {
			var path, symbol, sig, sha string
			var ls, le int
			if err := rows.Scan(&path, &symbol, &sig, &ls, &le, &sha); err != nil {
				return err
			}
			fmt.Fprintf(b, "  - %s : %s  [L%d-%d]  %s  (commit %s)\n", path, symbol, ls, le, truncateCell(sig, 160), shortSHA(sha))
			n++
		}
		if n == 0 {
			b.WriteString("  (no symbol matches in the index)\n")
		}
		return rows.Err()

	case "content":
		rows, err := db.QueryContext(ctx, `
			SELECT path, symbol, content, COALESCE(commit_sha,'')
			FROM code_symbols
			WHERE content ILIKE '%' || $1 || '%'
			  AND ($2 = '' OR repo = $2)
			ORDER BY path, symbol
			LIMIT $3`, c.Query, repoFilter, rowCap)
		if err != nil {
			return err
		}
		defer rows.Close()
		n := 0
		for rows.Next() {
			var path, symbol, content, sha string
			if err := rows.Scan(&path, &symbol, &content, &sha); err != nil {
				return err
			}
			fmt.Fprintf(b, "  - %s : %s  (commit %s)\n    | %s\n",
				path, symbol, shortSHA(sha), matchingExcerpt(content, c.Query, excerptChars))
			n++
		}
		if n == 0 {
			b.WriteString("  (no content matches in the index)\n")
		}
		return rows.Err()

	case "ls":
		rows, err := db.QueryContext(ctx, `
			SELECT DISTINCT path, COALESCE(commit_sha,'')
			FROM code_symbols
			WHERE path LIKE $1 || '%'
			  AND ($2 = '' OR repo = $2)
			ORDER BY path
			LIMIT $3`, c.Query, repoFilter, rowCap)
		if err != nil {
			return err
		}
		defer rows.Close()
		n := 0
		for rows.Next() {
			var path, sha string
			if err := rows.Scan(&path, &sha); err != nil {
				return err
			}
			fmt.Fprintf(b, "  - %s  (commit %s)\n", path, shortSHA(sha))
			n++
		}
		if n == 0 {
			b.WriteString("  (no indexed paths under that prefix)\n")
		}
		return rows.Err()
	}
	return fmt.Errorf("unrecognised kind %q", c.Kind)
}

// matchingExcerpt returns the first line of content containing the (case-
// insensitive) query, capped — enough to see the match in context without
// dumping whole symbol bodies into the round. Falls back to the content head
// when the match is multi-line-split or not found verbatim.
func matchingExcerpt(content, query string, cap_ int) string {
	lq := strings.ToLower(query)
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(strings.ToLower(line), lq) {
			return truncateCell(strings.TrimSpace(line), cap_)
		}
	}
	return truncateCell(strings.TrimSpace(content), cap_)
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	if sha == "" {
		return "?"
	}
	return sha
}

// dedupCodeChecks removes exact (kind, query) duplicates, preserving order and
// keeping the first occurrence's `why`. Independent reviewers reviewing one
// plan ask the same question; running it once is enough and reclaims the cap.
func dedupCodeChecks(in []codeCheck) []codeCheck {
	seen := make(map[string]bool, len(in))
	var out []codeCheck
	for _, c := range in {
		key := diagnose.CodeRequestKey(c.Kind, c.Query)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, c)
	}
	return out
}

// symbolTokenClause builds a WHERE clause matching every identifier token in
// the query as a case-insensitive substring of the symbol column (AND), plus
// the optional repo filter. Returns the clause and its ordered bind args; the
// caller appends the LIMIT param. Tokens are the maximal [A-Za-z0-9_] runs, so
// "OllamaClient.GenerateText", "(*OllamaClient).GenerateText", and
// "OllamaClient GenerateText" all yield {OllamaClient, GenerateText} and match
// the stored receiver form. A query with no identifier token falls back to a
// single raw-substring match so it still does something predictable.
func symbolTokenClause(query, repoFilter string, _ int) (string, []interface{}) {
	tokens := identifierTokens(query)
	var clauses []string
	var args []interface{}
	if len(tokens) == 0 {
		args = append(args, query)
		clauses = append(clauses, fmt.Sprintf("symbol ILIKE '%%' || $%d || '%%'", len(args)))
	} else {
		for _, t := range tokens {
			args = append(args, t)
			clauses = append(clauses, fmt.Sprintf("symbol ILIKE '%%' || $%d || '%%'", len(args)))
		}
	}
	args = append(args, repoFilter)
	clauses = append(clauses, fmt.Sprintf("($%d = '' OR repo = $%d)", len(args), len(args)))
	return strings.Join(clauses, " AND "), args
}

// identifierTokens splits s into maximal [A-Za-z0-9_] runs.
func identifierTokens(s string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			cur.WriteByte(c)
		} else {
			flush()
		}
	}
	flush()
	return out
}
