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
//	  content — match against symbol source BODIES and declarations (ILIKE over
//	            the trigram-indexed body and content columns, e.g. "stop_reason");
//	            returns path, symbol + an excerpt taken AROUND the match, marked
//	            [body] or [decl] so a reviewer can see which it matched.
//	  ls      — path prefix listing (e.g. "platform/aiservice/"); returns the
//	            distinct indexed paths under it.
//
// Every kind is READ-ONLY by construction — the SQL is written HERE, the
// reviewer supplies only a pattern; nothing model-written is executed.
//
// WHAT THE INDEX IS A SNAPSHOT OF (state it, do not leave it to be inferred):
// the index mirrors the last PUSHED ref that the code-indexer fetched, never
// local HEAD and never uncommitted work. So "no matches" means "not in the code
// as pushed at commit <sha>". The freshness banner states this on every answer,
// and its verdict is keyed to the indexed COMMIT's own committer date
// (code_symbols.commit_time, bugs_open/108 defect A) — never to updated_at,
// which any refresh resets even when it re-fetches the same stale tip.
//
// CORRECTED 2026-07-27 (D11 layer 1, council 18fe4035): until this date the
// `content` kind's description above was FALSE. It claimed to match source
// bodies; code_symbols held declarations only (composeSymbolContent: kind +
// symbol + signature + doc + path), because bodies were read on demand by
// analysis.ReadSymbolBody, which the indexer never called. Its own worked
// example, "%stop_reason%", could not match anything — a string literal inside a
// function was structurally unfindable, and every such check returned zero rows
// that read as absence (bugs_open/108 defect B). code_symbols.body now carries
// the source text and this comment describes what the code does.
package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gqls/agentchassis/pkg/diagnose"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// codeIndexStaleAfter is when the index is loud-flagged because the REFRESH
// pipeline looks dead. The index is refreshed by the `code-index-refresh`
// scheduled task every 24h (SEED_code_index_refresh_cadence.sql, bugs_open/059);
// 48h therefore means "at least one scheduled refresh was missed". Deliberately a
// const, not step config: it is one platform-wide fact coupled to that cadence
// row, and the banner always prints the ACTUAL age, so a reader can judge for
// themselves whatever this threshold says.
const codeIndexStaleAfter = 48 * time.Hour

// codeIndexCommitStaleAfter flags the index STALE by the age of the COMMIT it
// describes — the verdict that bugs_open/108 defect A was about. Distinct from
// codeIndexStaleAfter (missed-refresh detection) because the two clocks fail
// independently: re-running the indexer resets updated_at but can never reset
// the committer date, so this verdict cannot be laundered by a refresh that
// re-fetches the same pushed tip. Measured on this fleet (~191 commits/day),
// 48h of commit age is ~380 commits of invisibility; a genuinely idle repo
// tripping it errs in the safe direction (readers are told to treat empty
// answers as unknown, which for an idle repo costs trust, not correctness).
const codeIndexCommitStaleAfter = 48 * time.Hour

// indexFreshness is the high-water-mark row the banner decides from. A zero
// commitTime means the commit's date is UNRECORDED (NULL) — rendered as a loud
// UNKNOWN, never as fresh: FRESH must be unprovable without evidence about the
// commit itself, or the row clock creeps back in as the verdict.
type indexFreshness struct {
	sha, ref   string
	commitTime time.Time
	updatedAt  time.Time
	err        error
}

// codeIndexFreshness reads the index's high-water mark and renders the freshness
// banner both answer tiers prepend to their output. One query per action run,
// scoped to the repo the checks will search (same filter loadCodeIndexScope
// applies — a banner describing another corpus's freshness is not a guard).
// Never fatal: an error degrades to an "unknown freshness" note (fail open) —
// the guard must not break the lookup it qualifies.
func codeIndexFreshness(ctx context.Context, db *sql.DB, repoFilter string) string {
	var f indexFreshness
	var ct sql.NullTime
	f.err = db.QueryRowContext(ctx,
		`SELECT COALESCE(commit_sha,''), COALESCE(ref,''), commit_time, updated_at FROM code_symbols
		 WHERE ($1 = '' OR repo = $1)
		 ORDER BY updated_at DESC LIMIT 1`, repoFilter).Scan(&f.sha, &f.ref, &ct, &f.updatedAt)
	if f.err == sql.ErrNoRows {
		f = indexFreshness{}
	}
	if ct.Valid {
		f.commitTime = ct.Time
	}
	return freshnessBanner(f, time.Now())
}

// refClause names the indexed ref when it is recorded, so the banner can state
// what the index mirrors rather than leaving "which branch is this even" to be
// guessed. Pre-migration rows have no ref; the clause degrades to nothing.
func refClause(ref string) string {
	if ref == "" {
		return ""
	}
	return fmt.Sprintf(" (ref %s)", ref)
}

// freshnessBanner is the pure decision+formatting half of codeIndexFreshness,
// split out so every branch is unit-testable without a DB
// (verify-the-failing-branch: a guard whose job is to catch a fault must be
// SEEN catching it). bugs_open/059's read-time half: a stale index answers
// "absent" identically to a genuine absence, so at read time the answer must
// carry its own freshness — an empty result against a stale index is UNKNOWN,
// not evidence of absence, and nothing upstream can add that qualifier later.
//
// The verdict is a function of the indexed COMMIT, not of updated_at
// (bugs_open/108 defect A: the refresh cadence re-fetches the last pushed tip,
// so keying on the row clock reported FRESH forever while the described code
// fell 1,003 commits behind — and the honest empty-answer wording shipped with
// defect B's fix made that false FRESH a stronger lie). updated_at keeps one
// job, its original one: detecting that the refresh pipeline itself is dead.
func freshnessBanner(f indexFreshness, now time.Time) string {
	if f.err != nil {
		return fmt.Sprintf("(index freshness UNKNOWN — %v; treat every empty answer below as unknown, not absent)\n", f.err)
	}
	if f.updatedAt.IsZero() {
		return "!! CODE INDEX EMPTY — no symbols indexed at all. Every answer below is UNKNOWN, not absent. Run index-orchestrator before trusting any of this. !!\n"
	}
	rowAge := now.Sub(f.updatedAt)
	if f.commitTime.IsZero() {
		return fmt.Sprintf(
			"!! CODE INDEX AGE UNKNOWN: it describes commit %s%s but that commit's date is unrecorded — the refresh clock (rows written %s ago) says when the INDEXER last ran, never how old the code is, so freshness CANNOT be judged. Treat every empty answer below as UNKNOWN, not absent. (Expected until the first reindex after migration 250.) !!\n",
			shortSHA(f.sha), refClause(f.ref), formatAge(rowAge))
	}
	commitAge := now.Sub(f.commitTime)
	if commitAge > codeIndexCommitStaleAfter {
		return fmt.Sprintf(
			"!! CODE INDEX STALE: it describes commit %s%s, committed %s (%s ago) — every change since then is INVISIBLE below, so a 'no matches' answer may mean NOT YET INDEXED, not absent. The index mirrors the last PUSHED tip%s; re-running the indexer cannot advance it — only a push can. (rows refreshed %s ago) !!\n",
			shortSHA(f.sha), refClause(f.ref), f.commitTime.Format("2006-01-02"), formatAge(commitAge),
			refOfClause(f.ref), formatAge(rowAge))
	}
	if rowAge > codeIndexStaleAfter {
		return fmt.Sprintf(
			"!! CODE INDEX REFRESH MISSED: the indexed commit %s%s is recent (committed %s ago) but no refresh has run for %s — the cadence looks dead. Run index-orchestrator to refresh. Treat empty answers below as unknown, not absent. !!\n",
			shortSHA(f.sha), refClause(f.ref), formatAge(commitAge), formatAge(rowAge))
	}
	return fmt.Sprintf(
		"(index freshness: commit %s%s, committed %s ago; refreshed %s ago. The index mirrors the last pushed tip%s — local unpushed work is never visible.)\n",
		shortSHA(f.sha), refClause(f.ref), formatAge(commitAge), formatAge(rowAge), refOfClause(f.ref))
}

// refOfClause is refClause for prose position: " of <ref>" mid-sentence.
func refOfClause(ref string) string {
	if ref == "" {
		return ""
	}
	return " of " + ref
}

// codeIndexScope is what was actually searched, read ONCE per action run (not
// per check) and rendered into every empty answer.
//
// It exists because of the defect underneath bugs_open/108: an empty section
// reads as "nothing there", and is indistinguishable from "nobody ran your
// query" — and the prompt that tells a reviewer to treat the two differently is
// NOT in front of them when they read the result. So the distinction has to
// travel WITH the data. "answered: 0 rows, searched 4,535 symbols" is a finding;
// an empty section is not.
type codeIndexScope struct {
	total    int   // rows in scope (after the repo filter)
	withBody int   // of those, rows whose source body is indexed
	err      error // the scope read itself failed — then we know nothing
}

// loadCodeIndexScope counts the searchable corpus. One query per action run.
// Deliberately NOT folded into codeIndexFreshness: that function's pure half
// (freshnessBanner) is unit-tested against the stale/empty/error branches, and
// widening its signature to carry counts would be a change to a tested seam for
// the sake of saving one cheap COUNT.
func loadCodeIndexScope(ctx context.Context, db *sql.DB, repoFilter string) codeIndexScope {
	var s codeIndexScope
	s.err = db.QueryRowContext(ctx, `
		SELECT count(*), count(body) FROM code_symbols
		WHERE ($1 = '' OR repo = $1)`, repoFilter).Scan(&s.total, &s.withBody)
	return s
}

// bodyCoverageNote is the one-line statement of whether bodies are searchable at
// all. During the window between migration 243 applying and the chassis image
// rolling, the column exists and is entirely NULL — content checks degrade to
// the old declaration-only behaviour, which is not a break but IS something a
// reviewer must be told, or they will read the degrade as absence.
func (s codeIndexScope) bodyCoverageNote() string {
	if s.err != nil {
		return fmt.Sprintf("(index scope UNKNOWN — %v)\n", s.err)
	}
	switch {
	case s.total == 0:
		return "(index scope: 0 symbols — nothing to search)\n"
	case s.withBody == 0:
		return fmt.Sprintf("(index scope: %d symbols, but source BODIES ARE NOT INDEXED (0 of %d) — "+
			"a `content` check can only match declarations, signatures, doc comments and paths. "+
			"A string inside a function CANNOT be found: read those zeros as UNKNOWN, not absent.)\n", s.total, s.total)
	case s.withBody < s.total:
		return fmt.Sprintf("(index scope: %d symbols, source bodies indexed for %d of them (%.0f%%) — "+
			"the rest can only match declarations)\n", s.total, s.withBody, 100*float64(s.withBody)/float64(s.total))
	default:
		return fmt.Sprintf("(index scope: %d symbols, source bodies indexed for all of them)\n", s.total)
	}
}

// emptyAnswer renders a zero-row result as an ANSWER rather than as silence.
func (s codeIndexScope) emptyAnswer(kind string) string {
	if s.err != nil {
		return fmt.Sprintf("  NOT ANSWERED: the index scope could not be read (%v) — this is UNKNOWN, not absent.\n", s.err)
	}
	if s.total == 0 {
		return "  NOT ANSWERED: the index holds 0 symbols in scope — this is UNKNOWN, not absent.\n"
	}
	switch kind {
	case "content":
		if s.withBody == 0 {
			return fmt.Sprintf("  answered: 0 rows — searched %d symbols, but NO source bodies are indexed, "+
				"so a string inside a function could not have matched. Treat as UNKNOWN, not absent.\n", s.total)
		}
		return fmt.Sprintf("  answered: 0 rows — searched the bodies and declarations of %d indexed symbols "+
			"(%d with bodies). The query was RUN and found nothing; this is not an unanswered question.\n",
			s.total, s.withBody)
	case "ls":
		return fmt.Sprintf("  answered: 0 rows — no indexed path has that prefix, out of %d indexed symbols. "+
			"The query was RUN; this is not an unanswered question.\n", s.total)
	default:
		return fmt.Sprintf("  answered: 0 rows — searched the names of %d indexed symbols. "+
			"The query was RUN and matched none; this is not an unanswered question.\n", s.total)
	}
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
	b.WriteString(codeIndexFreshness(ctx, params.DB, repoFilter))
	// WHAT was searched, beside WHEN it was indexed. Freshness alone cannot say
	// whether a `content` check could have matched at all.
	scope := loadCodeIndexScope(ctx, params.DB, repoFilter)
	b.WriteString(scope.bodyCoverageNote())
	run := 0
	for i, c := range checks {
		fmt.Fprintf(&b, "\n[code_check %d] kind=%s query=%q — %s\n", i+1, c.Kind, c.Query, c.Why)
		if err := answerCodeCheck(ctx, params.DB, c, repoFilter, rowCap, excerptChars, scope, &b); err != nil {
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

// kindDoc is the code_symbols.kind carried by an indexed markdown section.
// Declared here rather than in the indexer because THIS is the reader that must
// never present one as code, and the guard ships before the corpus does (D12).
const kindDoc = "doc"

// codeKinds is the set of kinds that ARE source code, and therefore static-tier
// evidence. It is an ALLOW-LIST, and that direction is the whole point.
//
// The obvious form of this guard is `if kind == kindDoc { tag it }` — and that is
// a denylist of one: every other value, including a kind nobody has invented yet,
// silently falls through to "render as code, untagged". That is the shape the
// bug_historian seat named on corr da1f9c81 (medium): a dispatch table's default
// branch is a silent bug factory, once per new enum value. Inverted, the failure
// mode is loud and safe instead — an unknown kind renders as NOT-code, which is
// wrong in the direction that costs a reviewer one wasted lookup rather than a
// CONFIRMED verdict resting on prose.
// ONE source of truth: the slice below builds both the Go lookup and the CSV the
// `ls` query binds, so the set cannot drift between them. (Two hand-maintained
// copies of one list is precisely the class this council reviews for.)
var codeKindList = []string{"func", "method", "struct", "interface", "alias", "type", "var", "const"}

var codeKinds = func() map[string]bool {
	m := make(map[string]bool, len(codeKindList))
	for _, k := range codeKindList {
		m[k] = true
	}
	return m
}()

// codeKindsCSV feeds `kind = ANY(string_to_array($4, ','))` — a plain string bind,
// so no driver-specific array type is involved.
var codeKindsCSV = strings.Join(codeKindList, ",")

// docTag marks a row the verdicter must NOT read as code. Empty for every known
// code kind, so code answers render byte-identically to before this change.
func docTag(kind string) string {
	if codeKinds[kind] {
		return ""
	}
	return " [doc]"
}

// isCode reports whether a row belongs in the code block rather than the prose
// block. Same allow-list, same reason.
func isCode(kind string) bool { return codeKinds[kind] }

// docBlockHeader introduces the prose rows inside ONE check's answer. It is
// deliberately wordy: the reader is a model, and the bundle's own comment records
// why the words are the guard — "the model reads the heading and not this comment".
const docBlockHeader = "    ── documentation matches: PROSE from this repository's own docs.\n" +
	"       NOT static-tier evidence. A document SAYS a mechanism exists;\n" +
	"       only code SHOWS it. Do not cite these as \"the code shows X\" —\n" +
	"       follow them to the code and cite that.\n"

// answerCodeCheck runs ONE check against code_symbols. The SQL is fixed per
// kind; the reviewer's query arrives only as a bind parameter — nothing
// model-written is executed as SQL (contrast run_checks, whose whole point is
// model-written SELECTs under containment; here the containment is that the
// model writes no SQL at all).
//
// Rows whose kind is not in codeKinds are buffered into a separate block emitted
// AFTER the code rows, under docBlockHeader, with a [doc] tag on each line. Both
// code-answer lanes reach this function (diagnose_load_runtime_action.go:479
// calls it — "same package, so this is reuse, not a second implementation"), so
// this one guard covers the council's verify tier and the runtime lane alike.
func answerCodeCheck(ctx context.Context, db *sql.DB, c codeCheck, repoFilter string, rowCap, excerptChars int, scope codeIndexScope, b *strings.Builder) error {
	// Prose rows for THIS check, emitted after its code rows. Nil-safe: with no
	// doc rows in the index (the state on the day this ships) it stays empty and
	// nothing is written, so the rendered answer is unchanged byte-for-byte.
	var docs strings.Builder
	defer func() {
		if docs.Len() > 0 {
			b.WriteString(docBlockHeader)
			b.WriteString(docs.String())
		}
	}()
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
			SELECT path, symbol, COALESCE(signature,''), COALESCE(line_start,0), COALESCE(line_end,0), COALESCE(commit_sha,''), kind
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
			var path, symbol, sig, sha, kind string
			var ls, le int
			if err := rows.Scan(&path, &symbol, &sig, &ls, &le, &sha, &kind); err != nil {
				return err
			}
			out := b
			if !isCode(kind) {
				out = &docs
			}
			fmt.Fprintf(out, "  - %s : %s%s  [L%d-%d]  %s  (commit %s)\n", path, symbol, docTag(kind), ls, le, truncateCell(sig, 160), shortSHA(sha))
			n++
		}
		if n == 0 {
			b.WriteString(scope.emptyAnswer("symbol"))
		}
		return rows.Err()

	case "content":
		// body OR content. body is the symbol's source text (added 2026-07-27);
		// content is the declaration line + doc + path. Both are searched because
		// the kind's ONLY working use until now was declaration matches, and
		// silently dropping that would break checks that currently succeed.
		//
		// NO COALESCE IN THE PREDICATE, and that is load-bearing rather than
		// stylistic. Wrapping the column in COALESCE(body,'') disqualifies
		// idx_code_symbols_body_trgm — the planner cannot match an expression to
		// a plain-column index — so the whole thing falls back to a Seq Scan.
		// MEASURED on the live index (4,535 rows, bodies populated, 2026-07-28):
		//   COALESCE(body,'') ILIKE .. OR content ILIKE ..  → Seq Scan, 125.9 ms
		//   body ILIKE ..            OR content ILIKE ..    → BitmapOr across BOTH
		//                                                     trigram indexes, 5.5 ms
		// The row sets are IDENTICAL. On a NULL body, `body ILIKE x` evaluates to
		// NULL rather than false, and a WHERE clause discards NULL exactly as it
		// discards false — so a not-yet-indexed row still falls through to the
		// content side of the OR, which is the only thing the COALESCE was for.
		// (Verified against the NULL branch explicitly: with 0 NULL bodies in the
		// table, comparing the two forms on live rows would have been vacuous.)
		// This is guardian's low objection settled by measurement — and its
		// suggested remedy, one index on (COALESCE(body,'') || ' ' || content),
		// is not needed: it would duplicate every byte of both columns on disk and
		// lose the body-vs-declaration distinction the [body]/[decl] marker needs.
		rows, err := db.QueryContext(ctx, `
			SELECT path, symbol, COALESCE(body,''), content, COALESCE(commit_sha,''), (body IS NOT NULL), kind
			FROM code_symbols
			WHERE (body ILIKE '%' || $1 || '%' OR content ILIKE '%' || $1 || '%')
			  AND ($2 = '' OR repo = $2)
			ORDER BY path, symbol
			LIMIT $3`, c.Query, repoFilter, rowCap)
		if err != nil {
			return err
		}
		defer rows.Close()
		n := 0
		for rows.Next() {
			var path, symbol, body, content, sha, kind string
			var hasBody bool
			if err := rows.Scan(&path, &symbol, &body, &content, &sha, &hasBody, &kind); err != nil {
				return err
			}
			// Say WHICH text matched. "[body]" is a fact about the source;
			// "[decl]" is a fact about a signature, doc comment or path, and a
			// reviewer who cannot tell them apart will read a doc-comment
			// mention as an implementation.
			text, where := content, "decl"
			if strings.Contains(strings.ToLower(body), strings.ToLower(c.Query)) {
				text, where = body, "body"
			}
			out := b
			if !isCode(kind) {
				out = &docs
			}
			fmt.Fprintf(out, "  - %s : %s%s  (commit %s)\n    [%s] %s\n",
				path, symbol, docTag(kind), shortSHA(sha), where, matchingExcerpt(text, c.Query, excerptChars))
			n++
		}
		if n == 0 {
			b.WriteString(scope.emptyAnswer("content"))
		}
		return rows.Err()

	case "ls":
		// GROUP BY, not SELECT DISTINCT ... kind. `ls` lists one row per PATH, and
		// a Go file holds several kinds — adding kind to the DISTINCT would return
		// one row per (path, kind) and multiply every source file in the listing.
		// bool_or collapses it back: a path is code if ANY row under it is code.
		rows, err := db.QueryContext(ctx, `
			SELECT path, COALESCE(commit_sha,''),
			       bool_or(kind = ANY(string_to_array($4, ','))) AS has_code
			FROM code_symbols
			WHERE path LIKE $1 || '%'
			  AND ($2 = '' OR repo = $2)
			GROUP BY path, COALESCE(commit_sha,'')
			ORDER BY path
			LIMIT $3`, c.Query, repoFilter, rowCap, codeKindsCSV)
		if err != nil {
			return err
		}
		defer rows.Close()
		n := 0
		for rows.Next() {
			var path, sha string
			var hasCode bool
			if err := rows.Scan(&path, &sha, &hasCode); err != nil {
				return err
			}
			out, tag := b, ""
			if !hasCode {
				out, tag = &docs, " [doc]"
			}
			fmt.Fprintf(out, "  - %s%s  (commit %s)\n", path, tag, shortSHA(sha))
			n++
		}
		if n == 0 {
			b.WriteString(scope.emptyAnswer("ls"))
		}
		return rows.Err()
	}
	return fmt.Errorf("unrecognised kind %q", c.Kind)
}

// matchingExcerpt returns a capped window AROUND the match — enough to see it in
// context without dumping whole symbol bodies into the round.
//
// The window matters more since bodies became searchable. Against declaration
// text (a line or two) "first matching line, truncated" was indistinguishable
// from "around the match". Against a 200-line function it is not: truncating
// from the head returns a useless prefix of a body whose match is 180 lines
// down, and the reviewer sees a hit they cannot read. So: matching line first,
// then centred within that line if the line itself is long, then a whole-text
// window if the match straddles a newline, and only then the head.
func matchingExcerpt(text, query string, cap_ int) string {
	if cap_ <= 0 {
		cap_ = 400
	}
	lower, lq := strings.ToLower(text), strings.ToLower(query)
	if lq != "" {
		// Line-level first: a whole line of source is the most readable unit.
		off := 0
		for _, line := range strings.Split(text, "\n") {
			if i := strings.Index(strings.ToLower(line), lq); i >= 0 {
				return windowAround(strings.TrimRight(line, " \t"), i, query, cap_)
			}
			off += len(line) + 1
		}
		// Not within any single line — the match straddles a newline, or ToLower
		// shifted byte offsets (non-ASCII). Locate it in the whole text.
		if i := strings.Index(lower, lq); i >= 0 && i <= len(text) {
			return windowAround(text, i, query, cap_)
		}
	}
	// No verbatim match in THIS text: with the OR predicate the row may have
	// matched on the other column entirely. The head is then the honest fallback.
	return truncateCell(strings.TrimSpace(text), cap_)
}

// windowAround returns at most cap_ runes of s centred on the match at byteIdx,
// with ellipses marking what was cut. Byte offset in, rune arithmetic out, so a
// multi-byte source line is never cut mid-character.
func windowAround(s string, byteIdx int, query string, cap_ int) string {
	runes := []rune(s)
	if len(runes) <= cap_ {
		return strings.TrimSpace(s)
	}
	if byteIdx < 0 {
		byteIdx = 0
	}
	if byteIdx > len(s) {
		byteIdx = len(s)
	}
	matchStart := utf8.RuneCountInString(s[:byteIdx])
	matchLen := utf8.RuneCountInString(query)

	from := matchStart - (cap_-matchLen)/2
	if from < 0 {
		from = 0
	}
	if from > len(runes)-cap_ {
		from = len(runes) - cap_
	}
	to := from + cap_

	out := string(runes[from:to])
	if from > 0 {
		out = "…" + strings.TrimLeft(out, " \t")
	}
	if to < len(runes) {
		out += "…"
	}
	return out
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
