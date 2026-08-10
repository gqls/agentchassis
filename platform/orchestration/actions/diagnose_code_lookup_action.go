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
//	  symbol  — a name ("GenerateText", "Type.Method") or a path:Symbol handle
//	            ("internal/analysis/symbolbody.go:ReadSymbolBody"); the path
//	            half matches the path column, the name half token-matches the
//	            symbol column (receiver forms resolve), and a ":LINE" suffix
//	            degrades to a path check and says so (bugs_open/163 — the old
//	            contract line here said "match against the symbol name", and
//	            path-bearing queries were unsatisfiable by construction).
//	            Returns path, symbol, signature, lines; a path-qualified miss
//	            reports where the bare name DOES resolve before saying absent.
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
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gqls/agentchassis/internal/analysis"
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
	commits  int   // distinct commit_sha values across those rows (NULL counts as one)
	err      error // the scope read itself failed — then we know nothing

	// exts and kinds are what the corpus CAN REPRESENT, counted the same way and
	// for the same reason as total/withBody above — one census per run, rendered
	// into the answers (bugs_open/223).
	//
	// The two facts this census adds are the two the action had never stated. It
	// said WHEN the index was built (freshness) and WHETHER bodies were
	// searchable (bodyCoverageNote), and a reader could still not tell whether
	// their question was ANSWERABLE: on 2026-08-10 five of one run's eight checks
	// named Python scripts, a config-value string and a workflow step name, every
	// one of them absent from a corpus that holds 5,837 Go symbols and nothing
	// else, and all five rendered as "The query was RUN; this is not an
	// unanswered question."
	//
	// MEASURED, never assumed: both maps are read from the live table, so if the
	// corpus widens (D12's doc rows, or the var/const kinds the CHECK constraint
	// already permits) every sentence derived from them changes itself. A
	// hardcoded ".go only" would go on being printed after it stopped being true,
	// which is the stale-status failure this file already carries scars from.
	exts  map[string]int // ".go" → 5837; a path with no extension keys as "(none)"
	kinds map[string]int // "func" → 3653; a kind never written has NO entry
}

// loadCodeIndexScope counts the searchable corpus. Two queries per action run.
// Deliberately NOT folded into codeIndexFreshness: that function's pure half
// (freshnessBanner) is unit-tested against the stale/empty/error branches, and
// widening its signature to carry counts would be a change to a tested seam for
// the sake of saving one cheap COUNT.
//
// The representability census is a SECOND statement rather than more columns on
// the first: it is a GROUP BY over (kind, extension), and folding a grouped read
// into a single-row read would either lose the grouping or need a lateral join
// for no gain. A failure in it is non-fatal and deliberately does NOT set s.err —
// s.err means "we know nothing about scope", which would suppress the row counts
// the older guards depend on. Empty maps degrade to "representability unknown",
// which the classifier treats as "do not claim unanswerable" (fail open: the
// pre-223 wording, never a false NOT ANSWERABLE).
func loadCodeIndexScope(ctx context.Context, db *sql.DB, repoFilter string) codeIndexScope {
	var s codeIndexScope
	s.err = db.QueryRowContext(ctx, `
		SELECT count(*), count(body), count(DISTINCT COALESCE(commit_sha,'')) FROM code_symbols
		WHERE ($1 = '' OR repo = $1)`, repoFilter).Scan(&s.total, &s.withBody, &s.commits)
	if s.err != nil {
		return s
	}
	s.exts, s.kinds = make(map[string]int), make(map[string]int)
	rows, err := db.QueryContext(ctx, `
		SELECT kind,
		       lower(COALESCE(NULLIF(substring(path from '\.([^./]+)$'), ''), '(none)')) AS ext,
		       count(*)
		FROM code_symbols
		WHERE ($1 = '' OR repo = $1)
		GROUP BY 1, 2`, repoFilter)
	if err != nil {
		return s
	}
	defer rows.Close()
	for rows.Next() {
		var kind, ext string
		var n int
		if err := rows.Scan(&kind, &ext, &n); err != nil {
			return s
		}
		s.kinds[kind] += n
		s.exts["."+ext] += n
	}
	if err := rows.Err(); err != nil {
		// Partial maps are worse than none: a half-read census would report an
		// extension as absent because iteration stopped, and that is the exact
		// false "does not exist" this whole change exists to remove.
		s.exts, s.kinds = nil, nil
	}
	return s
}

// censusKnown reports whether the representability census was read at all. Every
// claim about what the index CANNOT hold is gated on it, so a failed census
// costs the new sentences and changes nothing else.
func (s codeIndexScope) censusKnown() bool { return len(s.exts) > 0 }

// representsExt reports whether the corpus in scope holds ANY row whose path
// carries this extension. False is only meaningful when censusKnown() is true.
func (s codeIndexScope) representsExt(ext string) bool {
	if !s.censusKnown() {
		return true // unknown ⇒ never claim unanswerable
	}
	return s.exts[strings.ToLower(ext)] > 0
}

// extSummary renders the extensions actually present, largest first, for the
// sentence that has to say what the reader's query was searched against.
func (s codeIndexScope) extSummary() string {
	if !s.censusKnown() {
		return "index composition unknown"
	}
	type pair struct {
		ext string
		n   int
	}
	pairs := make([]pair, 0, len(s.exts))
	for e, n := range s.exts {
		pairs = append(pairs, pair{e, n})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].n != pairs[j].n {
			return pairs[i].n > pairs[j].n
		}
		return pairs[i].ext < pairs[j].ext
	})
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, fmt.Sprintf("%s (%d rows)", p.ext, p.n))
	}
	return "the indexed corpus holds only: " + strings.Join(parts, ", ")
}

// missingCodeKinds names the members of codeKindList that the corpus holds NO
// row of — the fact the verifier needed on 2026-08-08 and invented a rename
// hypothesis without. Sorted for a stable sentence; empty when the census is
// unknown, so silence never implies completeness.
//
// It reads codeKindList rather than a second list, so the set it can complain
// about is exactly the set this file calls code (016b §9: two hand-maintained
// copies of one list is the drift class the council reviews for).
func (s codeIndexScope) missingCodeKinds() []string {
	if !s.censusKnown() {
		return nil
	}
	var missing []string
	for _, k := range codeKindList {
		if s.kinds[k] == 0 {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	return missing
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

// mixedCommitNote states that the corpus in scope does not all describe ONE
// commit, which the freshness banner above it cannot say: that banner reads the
// most recently written row (ORDER BY updated_at DESC LIMIT 1), so it reports the
// newest commit even when part of the index is older.
//
// It exists because the write side may now legitimately RETAIN rows instead of
// pruning them (bugs_open/135: a prune whose floor was not met is refused, so the
// rows this run did not confirm stay). Retention is the right call — deleting them
// is the catastrophe the guard exists to prevent — but a retained row is stale by
// construction, and an index that is silently part-stale while announcing one
// commit is the same lie bugs_closed/108 was filed about, one layer along. The
// pre-existing "no commit_sha, prune skipped" branch produces the same state, so
// this note describes a condition that was already reachable, not only the new one.
//
// Empty (no note) in the healthy single-commit case: a banner that always says
// something stops being read.
func (s codeIndexScope) mixedCommitNote() string {
	if s.err != nil || s.commits <= 1 {
		return ""
	}
	return fmt.Sprintf("!! INDEX NOT AT ONE COMMIT: the %d symbols in scope span %d different indexed commits, "+
		"so the freshness line above describes the NEWEST of them, not all of them — some rows describe older code. "+
		"Usual cause: an indexing run whose prune was REFUSED (its floor was not met) or skipped, which retains the rows it "+
		"could not confirm. Check `SELECT commit_sha, count(*) FROM code_symbols GROUP BY 1` and the doc_notes trail under "+
		"subject_key='index_code_symbols'. Treat a match on an older row as possibly-superseded, and an empty answer as UNKNOWN. !!\n",
		s.total, s.commits)
}

// emptyAnswer renders a zero-row result as an ANSWER rather than as silence.
//
// bugs_open/223 CHANGED WHAT "AN ANSWER" MEANS HERE, and the distinction is the
// whole fix. Until now every in-scope branch below closed with a sentence
// asserting the strongest available reading — "the query was RUN; this is not an
// unanswered question" — which is TRUE of every query this action executes and
// MISLEADING whenever the corpus could not have held the answer. Those words were
// written for bugs_closed/108 defect B, where empty answers were being read as
// silence, and they fixed it; they then became the mechanism of the opposite
// error. A guard can be wrong by being too confident.
//
// So the sentence is now earned, not assumed: `notAnswerableAnswer` replaces it
// outright when the census says the target class is unrepresentable, and the
// census-derived caveats below qualify it when the class is merely unlikely to be
// found. Every one of those caveats disappears on its own the day the corpus
// widens, because each is computed from the live census rather than written down.
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
			"(%d with bodies). The query was RUN and found nothing; this is not an unanswered question.\n%s",
			s.total, s.withBody, s.contentReachNote())
	case "ls":
		return fmt.Sprintf("  answered: 0 rows — no indexed path has that prefix, out of %d indexed symbols. "+
			"The query was RUN; this is not an unanswered question.\n", s.total)
	default:
		return fmt.Sprintf("  answered: 0 rows — searched the names of %d indexed symbols. "+
			"The query was RUN and matched none; this is not an unanswered question.\n%s",
			s.total, s.missingKindNote())
	}
}

// contentReachNote qualifies a 0-row `content` answer with what the search could
// have reached. A content check is the kind most often aimed at something that is
// not Go at all — the landmine-verifier's own derive_checks prompt DEFINES it as
// "a table name, a distinctive string, a command name" — and on 2026-08-10 three
// of one run's content checks named a `doc_notes` category, an agent type and a
// workflow step name, none of which can exist in a corpus of Go symbols.
//
// Empty when the corpus spans more than one language: at that point "only Go was
// searched" stops being the explanation and the reader is better served by the
// count alone. A note that always prints stops being read.
func (s codeIndexScope) contentReachNote() string {
	if !s.censusKnown() || len(s.exts) != 1 {
		return ""
	}
	return fmt.Sprintf("  note: %s — a footprint that lives in a script, a SQL file, a migration, "+
		"a database table, a config value or an agent definition CANNOT match here whatever its state. "+
		"For those classes this 0 is UNKNOWN, not absent.\n", s.extSummary())
}

// lsReachNote says what a NON-EMPTY `ls` listing is a listing of. See the call
// site for the measurement that motivates it: a prefix that holds indexed Go files
// in its subdirectories answers generously while the non-Go files the reviewer
// asked about are structurally absent from the result.
//
// Gated on a single-extension corpus for the same reason as contentReachNote —
// once the index spans several languages this sentence stops being the
// explanation, and a note that always prints stops being read.
func (s codeIndexScope) lsReachNote() string {
	if !s.censusKnown() || len(s.exts) != 1 {
		return ""
	}
	return fmt.Sprintf("  note: this lists INDEXED paths only — %s. A file of any other type under this "+
		"prefix is NOT listed, so its absence from the listing above is UNKNOWN, not evidence it is gone. "+
		"This is not a directory listing.\n", s.extSummary())
}

// missingKindNote states which kinds of declaration the corpus cannot hold, for a
// symbol answer that found nothing.
//
// This is the sentence bugs_open/223's third failure mode was filed for. Asked
// about `metaCommentaryPatterns` — a live `var` at validate_page_content.go:1229
// — the verifier reported it "no longer resolves as a standalone symbol (possibly
// inlined or renamed)". Nothing was renamed: code_symbols holds no `var` row at
// all, so the answer was structurally guaranteed and the hypothesis was
// manufactured. The remedy is not to ask the model to be careful; it is to hand
// it the census, so the absent kind is a stated fact instead of a gap it explains.
func (s codeIndexScope) missingKindNote() string {
	missing := s.missingCodeKinds()
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf("  note: the corpus in scope holds NO declarations of kind %s (kinds present: %s). "+
		"A package-level declaration of a missing kind is UNREPRESENTABLE here, so this 0-row answer supports "+
		"NEITHER removal NOR a rename-or-inline hypothesis — do not assert one without a row showing the new location.\n",
		strings.Join(missing, ", "), s.kindSummary())
}

// kindSummary lists the kinds the corpus does hold, so missingKindNote's claim
// can be read against its own evidence in the same sentence.
func (s codeIndexScope) kindSummary() string {
	if !s.censusKnown() {
		return "unknown"
	}
	present := make([]string, 0, len(s.kinds))
	for k := range s.kinds {
		present = append(present, k)
	}
	sort.Strings(present)
	return strings.Join(present, ", ")
}

// pathExt returns the lower-cased extension of the last segment of p, or "" when
// it has none. Deliberately conservative in the same direction as looksLikePath:
// a dot in a directory name ("docs/v1.2/notes") must not be read as a file
// extension, so only the final segment is considered.
func pathExt(p string) string {
	if p == "" {
		return ""
	}
	if i := strings.LastIndex(p, "/"); i >= 0 {
		p = p[i+1:]
	}
	i := strings.LastIndex(p, ".")
	if i <= 0 || i == len(p)-1 {
		return ""
	}
	return strings.ToLower(p[i:])
}

// unanswerableReason returns one clause naming why this check could NOT have
// matched, or "" when it could in principle have matched.
//
// It is the structural half of bugs_open/223: the blind spot is deterministic and
// knowable BEFORE the answer is read, so the answer must carry it rather than
// leaving a model to infer it. On 2026-08-10 a verifier run inferred it correctly
// — "the index covers Go symbols heavily but may not cover Python scripts" — and
// that is the problem, not the reassurance: the same 0 rows had already produced a
// flat "the entire described workflow has no footprint" three times, and nothing
// distinguishes the runs.
//
// Conservative by construction, in the same direction as looksLikePath: it speaks
// only about a FILE EXTENSION the census says is absent. A directory prefix, a
// bare identifier or an unknown census yields "" and the pre-223 wording — this
// must never invent an unanswerable, because a false NOT ANSWERABLE would suppress
// a real absence, which is the mirror of the bug being fixed.
func unanswerableReason(kind string, sq symbolQuery, query string, s codeIndexScope) string {
	if s.err != nil || !s.censusKnown() {
		return ""
	}
	target := query
	if kind == "symbol" {
		target = sq.path
	}
	ext := pathExt(target)
	if ext == "" || s.representsExt(ext) {
		return ""
	}
	return fmt.Sprintf("the corpus holds NO %s file at all — %s", ext, s.extSummary())
}

// notAnswerableAnswer renders the verdict that replaces a 0-row answer whose
// query could not have matched.
//
// The wording is the load-bearing part and is written for a MODEL reader, which
// is why it is explicit to the point of bluntness about all three readings it
// must block: removal, rename, and "does not exist". The bug this comes from
// records a verdict that chose each of those in turn from identical evidence.
func notAnswerableAnswer(reason string) string {
	return fmt.Sprintf("  NOT ANSWERABLE BY THIS INDEX: %s. The query was executed and returned 0 rows, "+
		"and it COULD NOT have returned a row whatever the state of the repository. This is UNKNOWN. It is "+
		"NOT evidence that the target is absent, removed, renamed or inlined, and it must not contribute to "+
		"a verdict of STALE — check it outside this index or record it as unverifiable.\n", reason)
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
			// A round that asked nothing has no code evidence either, and a
			// consumer branching on evidence must see the same answer here as it
			// would for a round whose every check was unanswerable. Omitting the
			// keys on this path would make the branch depend on WHY there was no
			// evidence, which is not a distinction the branch is entitled to.
			"checks_with_rows":    0,
			"checks_unanswerable": 0,
			"no_code_evidence":    true,
			"evidence_line":       "[code-lookup evidence: no code_checks were asked this round, so nothing about the code was verified either way.]",
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
	// And whether it is all ONE commit: the freshness line above names the newest
	// indexed commit, which is not the whole story once a prune has been refused
	// or skipped and part of the corpus is older (bugs_open/135). Silent when it is.
	b.WriteString(scope.mixedCommitNote())
	run, withRows, unanswerable := 0, 0, 0
	for i, c := range checks {
		fmt.Fprintf(&b, "\n[code_check %d] kind=%s query=%q — %s\n", i+1, c.Kind, c.Query, c.Why)
		outcome, err := answerCodeCheck(ctx, params.DB, c, repoFilter, rowCap, excerptChars, scope, &b)
		if err != nil {
			fmt.Fprintf(&b, "  (lookup failed: %v)\n", err)
			continue
		}
		run++
		if outcome.codeRows > 0 {
			withRows++
		}
		if outcome.unanswerable {
			unanswerable++
		}
	}
	if dropped > 0 {
		fmt.Fprintf(&b, "\n> %d further code_check(s) dropped (max_checks=%d) — coverage was capped, not complete.\n", dropped, maxChecks)
	}
	// The evidence line goes in the RENDERED text as well as the return map. The
	// map is for a workflow to branch on; the text is for the model that is about
	// to draw a conclusion, and it reads the text.
	evidence := codeEvidenceLine(run, withRows, unanswerable, scope)
	fmt.Fprintf(&b, "\n%s\n", evidence)

	logger.Info("diagnose_code_lookup: answered reviewer code checks",
		zap.Int("checks_run", run),
		zap.Int("checks_dropped", dropped),
		zap.Int("checks_with_rows", withRows),
		zap.Int("checks_unanswerable", unanswerable),
		zap.String("orchestration_id", orchIDForLog(params)))

	return map[string]interface{}{
		"results_text":   b.String(),
		"checks_run":     run,
		"checks_dropped": dropped,
		// ADDITIVE, and `checks_run` deliberately keeps its old meaning — checks
		// that EXECUTED without error, unanswerable ones included. Measured before
		// adding these: 0 live agent definitions reference checks_run, and 0
		// reference any of the four new keys, so nothing downstream changes
		// meaning and nothing collides (query in the lane's RUNBOOK).
		//
		// LANDMINE for the next reader: `checks_run > 0` does NOT mean anything was
		// verified. It never did; before bugs_open/223 there was simply no other
		// field to use. Branch on no_code_evidence.
		"checks_with_rows":    withRows,
		"checks_unanswerable": unanswerable,
		"no_code_evidence":    withRows == 0,
		"evidence_line":       evidence,
	}, nil
}

// codeEvidenceLine is the one-line mechanical census of what a round of checks
// actually established, composed HERE so that no model writes it, softens it or
// omits it (bugs_open/223).
//
// It exists because the persisted product of the landmine-verifier is PROSE in
// doc_notes, read months later by sessions and by council seats, and a verdict
// that rests on unanswerable checks is indistinguishable from one that rests on
// evidence once the run's inputs are gone. A consumer can append this to the row
// it persists (append_doc_note's note_body_suffix_field), and then the qualifier
// cannot be argued away by the same model that wrote the verdict.
func codeEvidenceLine(run, withRows, unanswerable int, scope codeIndexScope) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[code-lookup evidence: %d check(s) ran; %d matched indexed code; %d NOT ANSWERABLE by this index; %d ran and matched nothing in scope.",
		run, withRows, unanswerable, run-withRows-unanswerable)
	if withRows == 0 && run > 0 {
		b.WriteString(" NOTHING in this round was confirmed against indexed code, so absence claims here carry no weight in EITHER direction.")
	}
	switch {
	case scope.err != nil:
		b.WriteString(" Scope: UNKNOWN (the census could not be read).")
	case !scope.censusKnown():
		fmt.Fprintf(&b, " Scope: %d symbols; composition unknown.", scope.total)
	default:
		fmt.Fprintf(&b, " Scope: %d symbols, %s", scope.total, scope.extSummary())
		if missing := scope.missingCodeKinds(); len(missing) > 0 {
			fmt.Fprintf(&b, "; kinds with NO rows: %s", strings.Join(missing, ", "))
		}
		b.WriteString(".")
	}
	b.WriteString("]")
	return b.String()
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

// checkOutcome is what ONE check established, for a consumer that must branch on
// evidence rather than read prose (bugs_open/223).
//
// codeRows counts CODE-tier rows only. A `[doc]` row is deliberately not evidence:
// the D12 guard exists precisely because a document SAYS a mechanism exists where
// only code SHOWS it, and a field named "did this check find evidence" must honour
// the same distinction or it launders prose into proof one layer up.
//
// unanswerable and codeRows>0 are not exhaustive: a check can be answerable, run,
// and find nothing (the honest absence this action has always been able to report).
// Keeping the third state visible is the point — collapsing it into a boolean is
// how "no rows" became "does not exist" in the first place.
type checkOutcome struct {
	codeRows     int
	unanswerable bool
}

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
//
// ROW-CAP INVARIANT (bugs_open/181), and a fourth kind must honour BOTH halves:
// every arm binds probeLimit(rowCap) as its LIMIT — never rowCap itself — and
// every arm ends by rendering rowCapNotice(capped, rowCap), where `capped` is
// the OBSERVED arrival of the probe row, not an inference from the rendered
// count. An arm that binds rowCap directly cannot tell a complete answer from a
// truncated one, and that indistinguishability is the whole defect: a 40-of-305
// answer rendered identically to a complete one.
// RETURNS WHAT IT ESTABLISHED (bugs_open/223), not only prose. Both call sites
// need the outcome mechanically: the action aggregates it into the additive
// return keys a workflow can branch on, and the runtime lane folds it into the
// diagnosis bundle's evidence line. Every arm must set it — an arm that renders
// rows and forgets the outcome reports "no evidence found" on a successful check,
// which is a lie in the safe-looking direction.
func answerCodeCheck(ctx context.Context, db *sql.DB, c codeCheck, repoFilter string, rowCap, excerptChars int, scope codeIndexScope, b *strings.Builder) (checkOutcome, error) {
	var outcome checkOutcome
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
		// The NAME half is matched as an AND of identifier-token substrings,
		// NOT as the raw string. A reviewer writes Go methods as "Type.Method"
		// (e.g. "OllamaClient.GenerateText"), but the index stores the receiver
		// form "(*OllamaClient).GenerateText" — a raw ILIKE '%Type.Method%'
		// then misses it (false negative on run 90e989d5, 2026-07-17: the
		// Ollama-adapter check came back empty and a reviewer nearly read that
		// as "no such symbol"). Splitting on non-identifier chars and requiring
		// each token as a substring matches both forms.
		//
		// The PATH half (bugs_open/163) is matched against the path column,
		// because code_symbols.symbol never holds one. See symbolQuery.
		sq := parseSymbolQuery(c.Query)
		if sq.lineRef {
			fmt.Fprintf(b, "  note: %q names a LINE, not a symbol — answered as a path check over that file.\n", sq.raw)
		}
		clause, args, searched := symbolClauseFor(sq, repoFilter)
		codeRows, total, err := renderSymbolRows(ctx, db, clause, args, rowCap, b, &docs)
		if err != nil {
			return outcome, err
		}
		if total > 0 {
			outcome.codeRows = codeRows
			return outcome, nil
		}
		// A PATH-QUALIFIED query that found nothing is the exact shape that
		// produced 163's false verdict: the symbol existed and was reported
		// absent, because the answer could only speak about the pair. Ask the
		// name on its own before reporting absence, and report BOTH facts — a
		// moved or mis-cited file must not read as a missing symbol.
		if sq.path != "" && sq.name != "" {
			var elsewhere strings.Builder
			altClause, altArgs, _ := symbolClauseFor(symbolQuery{name: sq.name, raw: sq.raw}, repoFilter)
			altCode, _, err := renderSymbolRows(ctx, db, altClause, altArgs, rowCap, &elsewhere, &docs)
			if err != nil {
				return outcome, err
			}
			if altCode > 0 {
				fmt.Fprintf(b, "  answered: 0 rows AT THAT PATH — searched %s\n", searched)
				fmt.Fprintf(b, "  the NAME alone matches %d indexed symbol(s) ELSEWHERE, so the symbol EXISTS and the path is what did not match:\n", altCode)
				b.WriteString(elsewhere.String())
				outcome.codeRows = altCode
				return outcome, nil
			}
		}
		// The census verdict comes BEFORE the honest-absence wording, because for
		// an unrepresentable target the honest-absence wording is the error: it
		// says the question was answered when the question could not be asked.
		if reason := unanswerableReason("symbol", sq, c.Query, scope); reason != "" {
			b.WriteString(notAnswerableAnswer(reason))
			fmt.Fprintf(b, "  -- searched: %s\n", searched)
			outcome.unanswerable = true
			return outcome, nil
		}
		b.WriteString(scope.emptyAnswer("symbol"))
		fmt.Fprintf(b, "  -- searched: %s\n", searched)
		return outcome, nil

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
			LIMIT $3`, c.Query, repoFilter, probeLimit(rowCap))
		if err != nil {
			return outcome, err
		}
		defer rows.Close()
		n := 0
		capped := false
		for rows.Next() {
			// The probe row's ARRIVAL is the cap fact — break before scanning it.
			if rowCap > 0 && n >= rowCap {
				capped = true
				break
			}
			var path, symbol, body, content, sha, kind string
			var hasBody bool
			if err := rows.Scan(&path, &symbol, &body, &content, &sha, &hasBody, &kind); err != nil {
				return outcome, err
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
			} else {
				outcome.codeRows++
			}
			fmt.Fprintf(out, "  - %s : %s%s  (commit %s)\n    [%s] %s\n",
				path, symbol, docTag(kind), shortSHA(sha), where, matchingExcerpt(text, c.Query, excerptChars))
			n++
		}
		if err := rows.Err(); err != nil {
			return outcome, err
		}
		// n == 0 and capped are mutually exclusive, so these two branches cannot
		// both fire: an empty answer stays exactly the empty answer it was.
		if n == 0 {
			// A content query is free text, so the census can only speak about it
			// when the reviewer wrote a path-shaped one. When it can, it must:
			// "content: scripts/x.py" is the same unanswerable question as
			// "ls: scripts/x.py" and deserves the same verdict.
			if reason := unanswerableReason("content", symbolQuery{}, c.Query, scope); reason != "" {
				b.WriteString(notAnswerableAnswer(reason))
				outcome.unanswerable = true
			} else {
				b.WriteString(scope.emptyAnswer("content"))
			}
		}
		b.WriteString(rowCapNotice(capped, rowCap))
		return outcome, nil

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
			LIMIT $3`, c.Query, repoFilter, probeLimit(rowCap), codeKindsCSV)
		if err != nil {
			return outcome, err
		}
		defer rows.Close()
		n := 0
		capped := false
		for rows.Next() {
			// The probe row's ARRIVAL is the cap fact — break before scanning it.
			// This arm is ORDER BY path with nothing else to rank by, so its
			// discarded tail is purely alphabetical: the shape bugs_closed/164 was
			// filed for, and the reason an unreported cap here reads as absence.
			if rowCap > 0 && n >= rowCap {
				capped = true
				break
			}
			var path, sha string
			var hasCode bool
			if err := rows.Scan(&path, &sha, &hasCode); err != nil {
				return outcome, err
			}
			out, tag := b, ""
			if !hasCode {
				out, tag = &docs, " [doc]"
			} else {
				outcome.codeRows++
			}
			fmt.Fprintf(out, "  - %s%s  (commit %s)\n", path, tag, shortSHA(sha))
			n++
		}
		if err := rows.Err(); err != nil {
			return outcome, err
		}
		if n == 0 {
			if reason := unanswerableReason("ls", symbolQuery{}, c.Query, scope); reason != "" {
				b.WriteString(notAnswerableAnswer(reason))
				outcome.unanswerable = true
			} else {
				b.WriteString(scope.emptyAnswer("ls"))
			}
		} else {
			// THE ONE PLACE A CAVEAT RIDES ON A NON-EMPTY ANSWER, and it is the
			// nastier half of bugs_open/223 — a false POSITIVE, which the bug file
			// does not record because it was found while fixing it.
			//
			// `ls` is a path-PREFIX listing over an index of Go symbols, and it
			// presents as a directory listing. Measured 2026-08-10: `scripts/`
			// returns 110 indexed paths (Go programs under scripts/documentation_project/,
			// scripts/goscripts/, …) while every .py and .sh directly under scripts/
			// — including the three the entry actually named — is invisible. A
			// generous listing therefore reads as CONFIRMATION that a footprint
			// resolves, which is worse than a false STALE: a wrong accusation
			// invites checking, and a flattering partial confirmation reads as
			// diligence. So an `ls` answer says what it is a listing OF.
			b.WriteString(scope.lsReachNote())
		}
		b.WriteString(rowCapNotice(capped, rowCap))
		return outcome, nil
	}
	return outcome, fmt.Errorf("unrecognised kind %q", c.Kind)
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

// probeLimit returns the LIMIT to bind so a cap can be OBSERVED rather than
// inferred: one past rowCap. The extra row is never rendered — its arrival is
// the fact "more matches exist". Inferring the cap from `n == rowCap` instead
// would libel every answer that happens to hold exactly rowCap matches as
// incomplete, which is the same class of quiet lie as saying nothing
// (bugs_open/181). rowCap <= 0 is passed through unchanged so today's LIMIT 0
// -> emptyAnswer edge keeps its exact behaviour.
func probeLimit(rowCap int) int {
	if rowCap <= 0 {
		return rowCap
	}
	return rowCap + 1
}

// rowCapNotice is the one wording every arm renders when its lookup was capped
// (empty string when it was not — which is what keeps every uncapped answer
// byte-identical). capped is observed via probeLimit, never inferred from
// n == rowCap, so an answer with exactly rowCap matches carries no notice.
//
// bugs_open/181: the three arms of answerCodeCheck all capped at row_cap and
// none said so, while the same function's max_checks cap eight lines above the
// first arm DID report itself — so a reviewer reasonably read the absence of
// such a line as "nothing was capped". Live evidence: five rendered blocks at
// exactly 40 rows in llm_call_log prompts, true match counts 82/43/305/279.
// The wording says "this answer", never "the rows above", because a check's
// rows are split between the code builder and the deferred doc block: the
// notice lands after the code rows and before the doc block, and must stay
// true whatever the split.
func rowCapNotice(capped bool, rowCap int) string {
	if !capped {
		return ""
	}
	return fmt.Sprintf("  > this answer is CAPPED (row_cap=%d): the query matched more rows than are shown, "+
		"and rows are ordered by path — the missing matches sort after those shown. Treat absence from this "+
		"listing as UNKNOWN, not absent; narrow the query or raise row_cap.\n", rowCap)
}

// renderSymbolRows executes one symbol-arm SELECT and renders its rows: code
// rows to codeOut, prose rows to docs (the D12 guard — a doc-comment mention
// must never read as an implementation). Returns how many code rows and how
// many rows in total were rendered. Extracted so the 163 name-only fallback
// answers through the SAME renderer as the primary query — a second inline
// copy is how the [doc] tagging or a cap notice would silently apply to one
// branch and not the other.
//
// That cap notice now lives HERE (bugs_open/181), which is what makes it
// structural: the primary query and the 163 ELSEWHERE fallback both get it
// without either call site remembering, and a capped ELSEWHERE listing carries
// its notice inside the `elsewhere` builder it belongs to. `total` remains the
// RENDERED row count, so `total > 0` / `altCode > 0` at the call sites keep
// their meaning — the probe row is counted by nothing.
func renderSymbolRows(ctx context.Context, db *sql.DB, clause string, args []interface{}, rowCap int, codeOut *strings.Builder, docs *strings.Builder) (codeRows, total int, err error) {
	rows, err := db.QueryContext(ctx, `
		SELECT path, symbol, COALESCE(signature,''), COALESCE(line_start,0), COALESCE(line_end,0), COALESCE(commit_sha,''), kind
		FROM code_symbols
		WHERE `+clause+`
		ORDER BY path, symbol
		LIMIT `+fmt.Sprintf("$%d", len(args)+1), append(args, probeLimit(rowCap))...)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	capped := false
	for rows.Next() {
		// The probe row's ARRIVAL is the cap fact. Break before scanning it, so
		// it is observed and never rendered.
		if rowCap > 0 && total >= rowCap {
			capped = true
			break
		}
		var path, symbol, sig, sha, kind string
		var ls, le int
		if err := rows.Scan(&path, &symbol, &sig, &ls, &le, &sha, &kind); err != nil {
			return codeRows, total, err
		}
		out := codeOut
		if isCode(kind) {
			codeRows++
		} else {
			out = docs
		}
		fmt.Fprintf(out, "  - %s : %s%s  [L%d-%d]  %s  (commit %s)\n", path, symbol, docTag(kind), ls, le, truncateCell(sig, 160), shortSHA(sha))
		total++
	}
	if err := rows.Err(); err != nil {
		return codeRows, total, err
	}
	codeOut.WriteString(rowCapNotice(capped, rowCap))
	return codeRows, total, nil
}

// symbolQuery is a kind="symbol" check's query, parsed into the two columns
// that can actually answer it.
//
// bugs_open/163. The landmine-verifier's derive_checks prompt DEFINES kind
// "symbol" as `path:Symbol` — "a file path followed by a colon and a Go
// identifier" — and scopeFromCodeResults, resolveScopeEntries and the §7D scope
// resolver all compose that same handle. Until 2026-08-03 the clause builder
// tokenised the whole string and AND-ed EVERY token against the SYMBOL column,
// path fragments included, so `internal/analysis/symbolbody.go:ReadSymbolBody`
// executed as symbol ILIKE '%internal%' AND '%analysis%' AND '%go%' AND …
// code_symbols.symbol holds a Go identifier and NEVER a path — measured 0 of
// 4,992 rows contain a '/' — so a path-bearing symbol check was unsatisfiable
// BY CONSTRUCTION, returned 0 rows, and rendered as "the query was RUN and
// matched none". The producer's contract was right; the executor could not
// honour it. Splitting on the last colon is the fix already recorded in
// 016b §9 and in the bug file itself.
type symbolQuery struct {
	// path constrains code_symbols.path. Empty means "any path".
	path string
	// name is token-matched against code_symbols.symbol. Empty means "no name
	// constraint", reachable only via lineRef.
	name string
	// lineRef records that what followed the colon was a line number or range
	// (`spawn_actions.go:3066`, `run_checks_action.go:773-774`) rather than an
	// identifier. 12 LANDMINES footprints are written that way, and matching
	// "3066" against the symbol column would reproduce 163 in a new costume —
	// so the check degrades to a path question, and the answer says it did.
	lineRef bool
	// raw is the query as the reviewer wrote it, kept for narration.
	raw string
}

// parseSymbolQuery splits a symbol query into its path and name halves.
//
// Anything it does not recognise as path-bearing keeps the pre-2026-08-03
// behaviour exactly — the whole string is a name. That is the load-bearing
// property: "OllamaClient.GenerateText", "(*OllamaClient).GenerateText" and
// bare identifiers all land there, preserving the receiver-form match that
// 51e0776fb added for a MEASURED false negative (run 90e989d5, where a reviewer
// nearly read an empty result as "no such symbol"). This fix must not trade one
// false negative for another.
func parseSymbolQuery(query string) symbolQuery {
	q := symbolQuery{raw: query, name: query}
	pathPart, namePart := analysis.SplitSymbol(query)
	if namePart == "" || !looksLikePath(pathPart) {
		return q
	}
	q.path, q.name = pathPart, namePart
	if isLineReference(namePart) {
		q.name, q.lineRef = "", true
	}
	return q
}

// looksLikePath decides whether the text before the last colon is a file path
// rather than the first half of something like "Type.Method". A slash settles
// it; failing that a source-file extension does, because footprints routinely
// name a bare file ("claims.go:Foo"). Deliberately conservative: whatever it
// rejects falls back to the old whole-string-is-a-name path, so a wrong answer
// here can only fail to improve a query — never break one that works today.
func looksLikePath(s string) bool {
	if s == "" {
		return false
	}
	if strings.Contains(s, "/") {
		return true
	}
	lower := strings.ToLower(s)
	for _, ext := range []string{".go", ".sql", ".sh", ".py", ".md", ".ts", ".tsx", ".js"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// isLineReference reports whether s is a line number or line range — "3066",
// "773-774", optionally L-prefixed. Not a regexp: this file uses none, and the
// grammar is three characters wide.
func isLineReference(s string) bool {
	s = strings.TrimPrefix(strings.TrimPrefix(s, "L"), "l")
	if s == "" {
		return false
	}
	seenDigit, seenDash := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			seenDigit = true
		case c == '-' && seenDigit && !seenDash && i+1 < len(s):
			seenDash = true
		default:
			return false
		}
	}
	return seenDigit
}

// symbolClauseFor builds the WHERE clause for a parsed symbol query, its
// ordered bind args (the caller appends the LIMIT param), and a one-line human
// rendering of the predicate actually built.
//
// The narration is RETURNED rather than logged because the reader is a model. A
// 0-row answer that shows its own predicate is self-diagnosing, and — this is
// the point — it gives a per-check fact that can contradict the run-level
// staleness banner. The false verdict that produced 163 was not invented from
// nothing: the banner correctly warned that "a 'no matches' answer may mean NOT
// YET INDEXED", and the model generalised that whole-run caveat onto symbols
// that predated the indexed commit. There was nothing in the per-check output
// specific enough to override it. Now there is.
//
// The repo filter stays the LAST bind arg; a test pins that.
func symbolClauseFor(sq symbolQuery, repoFilter string) (clause string, args []interface{}, searched string) {
	var clauses, human []string
	if sq.path != "" {
		args = append(args, sq.path)
		clauses = append(clauses, fmt.Sprintf("path ILIKE '%%' || $%d || '%%'", len(args)))
		human = append(human, fmt.Sprintf("path ILIKE '%%%s%%'", sq.path))
	}
	tokens := identifierTokens(sq.name)
	if sq.name != "" && len(tokens) == 0 {
		// A name with no identifier token at all still gets a raw-substring
		// match, so the check does something predictable rather than nothing.
		tokens = []string{sq.name}
	}
	for _, t := range tokens {
		args = append(args, t)
		clauses = append(clauses, fmt.Sprintf("symbol ILIKE '%%' || $%d || '%%'", len(args)))
		human = append(human, fmt.Sprintf("symbol ILIKE '%%%s%%'", t))
	}
	args = append(args, repoFilter)
	clauses = append(clauses, fmt.Sprintf("($%d = '' OR repo = $%d)", len(args), len(args)))
	return strings.Join(clauses, " AND "), args, strings.Join(human, " AND ")
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
