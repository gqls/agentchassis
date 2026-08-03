package actions

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// The extractor is the seam between reviewer JSON and the fixed-SQL executor:
// malformed entries must drop silently (a fumbled check loses the check, not
// the run), unknown kinds must not reach the SQL switch.
func TestCodeChecksFromCollected(t *testing.T) {
	collected := map[string]interface{}{
		"review_bug_historian": map[string]interface{}{
			"result": map[string]interface{}{
				"code_checks": []interface{}{
					map[string]interface{}{"kind": "symbol", "query": "GenerateText", "why": "adapter count"},
					map[string]interface{}{"kind": "LS", "query": "platform/aiservice/", "why": "case-folded kind"},
					map[string]interface{}{"kind": "content", "query": "stop_reason"},
					map[string]interface{}{"kind": "drop_table", "query": "x", "why": "bad kind"},
					map[string]interface{}{"kind": "symbol", "query": "   ", "why": "blank query"},
					"not a map",
				},
			},
		},
	}
	got := codeChecksFromCollected(collected, "review_bug_historian.result.code_checks")
	if len(got) != 3 {
		t.Fatalf("want 3 valid checks, got %d: %+v", len(got), got)
	}
	if got[0].Kind != "symbol" || got[0].Query != "GenerateText" {
		t.Fatalf("check 0 mangled: %+v", got[0])
	}
	if got[1].Kind != "ls" {
		t.Fatalf("kind not case-folded: %+v", got[1])
	}
	if got[2].Why != "" {
		t.Fatalf("missing why should stay empty, got %q", got[2].Why)
	}
}

func TestCodeChecksFromCollectedAbsentField(t *testing.T) {
	if got := codeChecksFromCollected(map[string]interface{}{}, "nope.nothing"); got != nil {
		t.Fatalf("absent field must yield nil, got %+v", got)
	}
}

func TestValidCodeCheckKind(t *testing.T) {
	for _, ok := range []string{"symbol", "content", "ls"} {
		if !validCodeCheckKind(ok) {
			t.Fatalf("%q should be valid", ok)
		}
	}
	for _, bad := range []string{"", "sql", "grep", "exec"} {
		if validCodeCheckKind(bad) {
			t.Fatalf("%q must not be valid", bad)
		}
	}
}

func TestMatchingExcerpt(t *testing.T) {
	content := "func (c *AnthropicClient) GenerateText(...) {\n\t// no stop_reason decoded\n\treturn text, nil\n}"
	got := matchingExcerpt(content, "STOP_REASON", 100)
	if !strings.Contains(got, "stop_reason") {
		t.Fatalf("excerpt should carry the matching line, got %q", got)
	}
	// no match → capped head, not empty
	head := matchingExcerpt(content, "zzz-absent", 20)
	if head == "" || len(head) > 24 {
		t.Fatalf("fallback head wrong: %q", head)
	}
}

func TestIdentifierTokens(t *testing.T) {
	cases := map[string][]string{
		"OllamaClient.GenerateText":    {"OllamaClient", "GenerateText"},
		"(*OllamaClient).GenerateText": {"OllamaClient", "GenerateText"},
		"OllamaClient GenerateText":    {"OllamaClient", "GenerateText"},
		"GenerateText":                 {"GenerateText"},
		"snake_case_name":              {"snake_case_name"},
		"":                             nil,
		"...":                          nil,
	}
	for in, want := range cases {
		got := identifierTokens(in)
		if len(got) != len(want) {
			t.Fatalf("identifierTokens(%q) = %v, want %v", in, got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("identifierTokens(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

// The regression that motivated the token match: "Type.Method" must produce a
// clause that would match the stored "(*Type).Method" receiver form. We can't
// hit the DB here, but we can assert the clause ANDs both tokens as symbol
// substrings (which the stored form satisfies) rather than one literal string.
// (Ported from symbolTokenClause to the parse+clause pair for bugs_open/163;
// the assertions are byte-identical to the 51e0776fb originals — a no-colon
// query must build exactly the clause it always built.)
func TestSymbolTokenClauseGoReceiver(t *testing.T) {
	clause, args, _ := symbolClauseFor(parseSymbolQuery("OllamaClient.GenerateText"), "")
	if !strings.Contains(clause, "symbol ILIKE '%' || $1 || '%'") ||
		!strings.Contains(clause, "symbol ILIKE '%' || $2 || '%'") {
		t.Fatalf("expected two AND-ed symbol substring clauses, got: %s", clause)
	}
	if len(args) != 3 || args[0] != "OllamaClient" || args[1] != "GenerateText" || args[2] != "" {
		t.Fatalf("args = %v, want [OllamaClient GenerateText \"\"]", args)
	}
	if strings.Contains(clause, "path ILIKE") {
		t.Fatalf("a dotted method name is not a path and must not constrain the path column: %s", clause)
	}
}

func TestSymbolTokenClauseRepoFilter(t *testing.T) {
	_, args, _ := symbolClauseFor(parseSymbolQuery("GenerateText"), "gqls/agentchassis")
	if args[len(args)-1] != "gqls/agentchassis" {
		t.Fatalf("repo filter should be the last arg, got %v", args)
	}
}

// bugs_open/163. The landmine-verifier's derive_checks prompt DEFINES kind
// "symbol" as `path:Symbol`, and until 2026-08-03 the clause builder AND-ed
// every token of that string — path fragments included — against the symbol
// column, which never holds a path (measured: 0 of 4,992 rows contain '/').
// Every path-bearing symbol check ever run returned 0 rows by construction and
// rendered as a confident "matched none"; 16 of 20 landmine verifications came
// back NEEDS_HUMAN_REVIEW and none ever CONFIRMED. The path half must bind to
// the path column, the name half to the symbol column.
func TestSymbolTokenClausePathBearing(t *testing.T) {
	sq := parseSymbolQuery("internal/analysis/symbolbody.go:ReadSymbolBody")
	if sq.path != "internal/analysis/symbolbody.go" || sq.name != "ReadSymbolBody" || sq.lineRef {
		t.Fatalf("parse mangled the handle: %+v", sq)
	}
	clause, args, searched := symbolClauseFor(sq, "")
	if !strings.Contains(clause, "path ILIKE '%' || $1 || '%'") {
		t.Fatalf("the path half must constrain the path column: %s", clause)
	}
	if !strings.Contains(clause, "symbol ILIKE '%' || $2 || '%'") {
		t.Fatalf("the name half must constrain the symbol column: %s", clause)
	}
	// The 163 failure shape: no path fragment may reach the symbol column.
	for i, a := range args[:len(args)-1] { // last arg is the repo filter
		if s, ok := a.(string); ok && i > 0 && (s == "internal" || s == "analysis" || s == "go" || strings.Contains(s, "/")) {
			t.Fatalf("path fragment %q bound as a symbol token — the 163 defect: args=%v", s, args)
		}
	}
	if len(args) != 3 || args[0] != "internal/analysis/symbolbody.go" || args[1] != "ReadSymbolBody" || args[2] != "" {
		t.Fatalf("args = %v, want [path name repo]", args)
	}
	// The narration is the per-check fact that can override the run-level
	// staleness banner; a 0-row answer must be able to show its own predicate.
	if !strings.Contains(searched, "path ILIKE '%internal/analysis/symbolbody.go%'") ||
		!strings.Contains(searched, "symbol ILIKE '%ReadSymbolBody%'") {
		t.Fatalf("searched narration does not state the predicate: %q", searched)
	}
}

// Receiver-qualified name AFTER a path — both halves must survive: the path to
// the path column, both name tokens to the symbol column (receiver form again).
func TestSymbolTokenClausePathWithDottedName(t *testing.T) {
	sq := parseSymbolQuery("platform/aiservice/ollama.go:OllamaClient.GenerateText")
	clause, args, _ := symbolClauseFor(sq, "")
	if sq.path != "platform/aiservice/ollama.go" {
		t.Fatalf("path half lost: %+v", sq)
	}
	if len(args) != 4 || args[1] != "OllamaClient" || args[2] != "GenerateText" {
		t.Fatalf("dotted name after a path must still token-split: %v", args)
	}
	if strings.Count(clause, "symbol ILIKE") != 2 {
		t.Fatalf("want two symbol token clauses: %s", clause)
	}
}

// 12 LANDMINES footprints write a LINE reference after the colon, not a symbol
// (`spawn_actions.go:3066`, `run_checks_action.go:773-774`). Matching "3066"
// against the symbol column would reproduce 163 in a new costume; the check
// must degrade to a path question and say so.
func TestSymbolTokenClauseLineReference(t *testing.T) {
	for _, q := range []string{
		"platform/orchestration/actions/spawn_actions.go:3066",
		"platform/orchestration/actions/run_checks_action.go:773-774",
		"internal/analysis/symbolbody.go:L105",
	} {
		sq := parseSymbolQuery(q)
		if !sq.lineRef {
			t.Fatalf("%q should parse as a line reference: %+v", q, sq)
		}
		clause, args, _ := symbolClauseFor(sq, "")
		if strings.Contains(clause, "symbol ILIKE") {
			t.Fatalf("a line number must not become a symbol predicate (%q): %s", q, clause)
		}
		if len(args) != 2 { // path + repo filter only
			t.Fatalf("line-ref args = %v, want [path repo]", args)
		}
	}
	// Negative control for the detector itself: a hyphenated identifier-less
	// name that is NOT numeric must stay a name, and a genuine symbol whose
	// file half lacks any path shape must keep the old whole-string behaviour.
	if sq := parseSymbolQuery("claims.go:773x"); sq.lineRef {
		t.Fatalf("773x is not a line reference: %+v", sq)
	}
	if sq := parseSymbolQuery("NotAPath:GenerateText"); sq.path != "" {
		t.Fatalf("a left half with no path shape must not bind the path column: %+v", sq)
	}
}

// The parser's conservatism is load-bearing: anything not recognised as
// path-bearing must keep the pre-163 behaviour byte-for-byte, or queries that
// work today change under people's feet.
func TestParseSymbolQueryFallsBackToNameOnly(t *testing.T) {
	for _, q := range []string{
		"GenerateText",
		"OllamaClient.GenerateText",
		"(*OllamaClient).GenerateText",
		"snake_case_name",
		"NotAPath:GenerateText", // colon, but the left half is not a path
	} {
		sq := parseSymbolQuery(q)
		if sq.path != "" || sq.name != q || sq.lineRef {
			t.Fatalf("parseSymbolQuery(%q) = %+v, want name-only passthrough", q, sq)
		}
	}
}

func TestDedupCodeChecks(t *testing.T) {
	in := []codeCheck{
		{Kind: "content", Query: "stop_reason", Why: "first"},
		{Kind: "content", Query: "stop_reason", Why: "dup — dropped"},
		{Kind: "content", Query: "STOP_REASON", Why: "case-fold dup — dropped"},
		{Kind: "symbol", Query: "stop_reason", Why: "different kind — kept"},
		{Kind: "content", Query: "done_reason", Why: "distinct — kept"},
	}
	got := dedupCodeChecks(in)
	if len(got) != 3 {
		t.Fatalf("want 3 after dedup, got %d: %+v", len(got), got)
	}
	if got[0].Why != "first" {
		t.Fatalf("first occurrence's why should survive, got %q", got[0].Why)
	}
}

// The read-time freshness guard (bugs_open/059, reworked for bugs_open/108
// defect A). The header text has always SAID "treat a stale or empty answer as
// unknown, not absent" — these tests verify the guard that COMPUTES that,
// branch by induced branch (a guard whose job is to catch a fault must be seen
// catching it, not only passing when healthy). The verdict keys on the indexed
// commit's own date: the exact live failure — a fresh refresh clock over a
// days-old commit — is the first case, because it is the branch the old
// updated_at-keyed verdict could not represent.
func TestFreshnessBanner(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

	t.Run("THE 108 failure: fresh refresh clock over a stale commit is loud-STALE", func(t *testing.T) {
		// Measured live 2026-07-28: rows written 2h ago, commit e19aa5d from
		// four days and 1,003 commits earlier. The old verdict rendered this
		// "(index freshness: refreshed 2h ago …)".
		got := freshnessBanner(indexFreshness{
			sha: "e19aa5d108e", ref: "086_experience_loop",
			commitTime: now.Add(-96 * time.Hour), updatedAt: now.Add(-2 * time.Hour),
		}, now)
		for _, must := range []string{"STALE", "e19aa5d1", "086_experience_loop", "4d", "NOT YET INDEXED", "PUSHED", "only a push"} {
			if !strings.Contains(got, must) {
				t.Fatalf("stale-commit banner must contain %q: %q", must, got)
			}
		}
	})

	t.Run("unrecorded commit date: loud UNKNOWN, never fresh", func(t *testing.T) {
		// Pre-migration rows, or the CommitInfo call failed. However recent the
		// refresh, FRESH must be unprovable without evidence about the commit.
		got := freshnessBanner(indexFreshness{sha: "adb00fd1234", updatedAt: now.Add(-time.Hour)}, now)
		if !strings.Contains(got, "UNKNOWN") || !strings.Contains(got, "!!") {
			t.Fatalf("unrecorded commit date must be loud-UNKNOWN: %q", got)
		}
		if strings.Contains(got, "index freshness:") {
			t.Fatalf("unrecorded commit date must not render the quiet fresh line: %q", got)
		}
	})

	t.Run("fresh index: quiet one-liner naming both ages, sha, ref and the mirror fact", func(t *testing.T) {
		got := freshnessBanner(indexFreshness{
			sha: "adb00fd1234", ref: "main",
			commitTime: now.Add(-5 * time.Hour), updatedAt: now.Add(-3 * time.Hour),
		}, now)
		if strings.Contains(got, "STALE") || strings.Contains(got, "!!") {
			t.Fatalf("a fresh index must not be flagged: %q", got)
		}
		for _, must := range []string{"5h", "3h", "adb00fd", "main", "pushed tip", "never visible"} {
			if !strings.Contains(got, must) {
				t.Fatalf("banner must name %q: %q", must, got)
			}
		}
	})

	t.Run("missed refresh: recent commit but a dead cadence stays loud", func(t *testing.T) {
		got := freshnessBanner(indexFreshness{
			sha: "abc1234", commitTime: now.Add(-3 * time.Hour), updatedAt: now.Add(-72 * time.Hour),
		}, now)
		for _, must := range []string{"REFRESH MISSED", "3d", "index-orchestrator"} {
			if !strings.Contains(got, must) {
				t.Fatalf("missed-refresh banner must contain %q: %q", must, got)
			}
		}
	})

	t.Run("boundary: commit age exactly at the threshold is not yet stale", func(t *testing.T) {
		got := freshnessBanner(indexFreshness{
			sha: "abc1234", commitTime: now.Add(-codeIndexCommitStaleAfter), updatedAt: now.Add(-time.Hour),
		}, now)
		if strings.Contains(got, "STALE") {
			t.Fatalf("commit age == threshold must not flag (only >): %q", got)
		}
	})

	t.Run("empty index: loudest — every answer is unknown", func(t *testing.T) {
		got := freshnessBanner(indexFreshness{}, now)
		for _, must := range []string{"EMPTY", "UNKNOWN, not absent", "index-orchestrator"} {
			if !strings.Contains(got, must) {
				t.Fatalf("empty banner must contain %q: %q", must, got)
			}
		}
	})

	t.Run("query error: fail open with an unknown-freshness note, never silent", func(t *testing.T) {
		got := freshnessBanner(indexFreshness{err: fmt.Errorf("connection refused")}, now)
		for _, must := range []string{"UNKNOWN", "connection refused", "unknown, not absent"} {
			if !strings.Contains(got, must) {
				t.Fatalf("error banner must contain %q: %q", must, got)
			}
		}
	})

	t.Run("missing ref degrades to no clause, not to an empty parenthesis", func(t *testing.T) {
		got := freshnessBanner(indexFreshness{
			sha: "abc1234", commitTime: now.Add(-96 * time.Hour), updatedAt: now.Add(-time.Hour),
		}, now)
		if strings.Contains(got, "(ref ") || strings.Contains(got, " of )") {
			t.Fatalf("pre-migration rows have no ref; the clause must vanish cleanly: %q", got)
		}
	})
}

func TestFormatAge(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Minute, "30m"},
		{3 * time.Hour, "3h"},
		{47 * time.Hour, "47h"},
		{49 * time.Hour, "2d"},
		{21 * 24 * time.Hour, "21d"},
	}
	for _, c := range cases {
		if got := formatAge(c.d); got != c.want {
			t.Fatalf("formatAge(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

// Bodies made the excerpt window load-bearing. Against declaration text (a line
// or two) "first matching line, truncated" and "around the match" are the same
// answer; against a 200-line function they are not — truncating from the head
// hands a reviewer a hit they cannot read. The failing branch is the one worth
// exercising, so the match here sits deep in a long body.
func TestMatchingExcerptWindowsAroundADeepMatch(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&b, "\tfiller line %d that is here only to push the match a long way down\n", i)
	}
	b.WriteString("\tif resp.StopReason == \"max_tokens\" { // the_needle_we_want\n")
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&b, "\tmore filler %d\n", i)
	}
	got := matchingExcerpt(b.String(), "the_needle_we_want", 120)
	if !strings.Contains(got, "the_needle_we_want") {
		t.Fatalf("excerpt lost the match in a long body: %q", got)
	}
	if strings.Contains(got, "filler line 0 ") {
		t.Fatalf("excerpt came from the head of the body, not around the match: %q", got)
	}
	if n := len([]rune(got)); n > 130 { // 120 + the two ellipses and trimming
		t.Fatalf("excerpt overran the cap: %d runes", n)
	}
}

// A single line longer than the cap must still be centred on the match, or a
// minified/generated line hides the very thing that matched.
func TestMatchingExcerptCentresWithinALongLine(t *testing.T) {
	line := strings.Repeat("x", 500) + "the_needle_we_want" + strings.Repeat("y", 500)
	got := matchingExcerpt(line, "the_needle_we_want", 80)
	if !strings.Contains(got, "the_needle_we_want") {
		t.Fatalf("long line was truncated from the head, losing the match: %q", got)
	}
}

// The whole point of the scope struct: a zero must never render as silence, and
// "0 rows because bodies are not indexed" must not read like "0 rows because the
// code does not do that". Both degraded branches are asserted, not just the
// healthy one.
func TestCodeIndexScopeEmptyAnswer(t *testing.T) {
	healthy := codeIndexScope{total: 4535, withBody: 4535}
	if got := healthy.emptyAnswer("content"); !strings.Contains(got, "answered: 0 rows") ||
		!strings.Contains(got, "4535") || strings.Contains(got, "UNKNOWN") {
		t.Fatalf("a genuine searched-and-empty answer should be stated as one: %q", got)
	}

	noBodies := codeIndexScope{total: 4535, withBody: 0}
	if got := noBodies.emptyAnswer("content"); !strings.Contains(got, "UNKNOWN") {
		t.Fatalf("with no bodies indexed a content zero is UNKNOWN, not absent: %q", got)
	}

	empty := codeIndexScope{total: 0}
	if got := empty.emptyAnswer("symbol"); !strings.Contains(got, "NOT ANSWERED") {
		t.Fatalf("an empty index cannot answer anything: %q", got)
	}

	broken := codeIndexScope{err: fmt.Errorf("connection refused")}
	if got := broken.emptyAnswer("ls"); !strings.Contains(got, "NOT ANSWERED") ||
		!strings.Contains(got, "connection refused") {
		t.Fatalf("a failed scope read must say so, with the reason: %q", got)
	}
}

// The coverage note is what a reviewer reads BEFORE any answer, so the
// between-migration-and-roll state (column exists, every row NULL) has to be
// legible as a degrade rather than as a healthy empty result.
func TestBodyCoverageNote(t *testing.T) {
	if got := (codeIndexScope{total: 4535, withBody: 0}).bodyCoverageNote(); !strings.Contains(got, "BODIES ARE NOT INDEXED") {
		t.Fatalf("the no-bodies degrade must be loud: %q", got)
	}
	if got := (codeIndexScope{total: 4535, withBody: 4535}).bodyCoverageNote(); !strings.Contains(got, "all of them") {
		t.Fatalf("full coverage should say so plainly: %q", got)
	}
	if got := (codeIndexScope{total: 4535, withBody: 2000}).bodyCoverageNote(); !strings.Contains(got, "44%") {
		t.Fatalf("partial coverage should report the percentage: %q", got)
	}
}

// bugs_open/135's read-side half. The freshness banner reads ONE row (the newest
// by updated_at) and names its commit, so a part-stale index — the state a refused
// or skipped prune deliberately leaves behind — announces a single commit it does
// not entirely describe. This note is the only thing that contradicts it, so its
// silence in the healthy case matters as much as its noise in the mixed one.
func TestMixedCommitNote(t *testing.T) {
	if got := (codeIndexScope{total: 4992, withBody: 4992, commits: 1}).mixedCommitNote(); got != "" {
		t.Errorf("a single-commit index must produce NO note (a banner that always fires stops being read): %q", got)
	}
	if got := (codeIndexScope{total: 4992, withBody: 4992, commits: 0}).mixedCommitNote(); got != "" {
		t.Errorf("an empty index has no commit spread to report: %q", got)
	}
	if got := (codeIndexScope{err: fmt.Errorf("connection refused"), commits: 3}).mixedCommitNote(); got != "" {
		t.Errorf("a failed scope read already says UNKNOWN via bodyCoverageNote; do not assert a spread we did not measure: %q", got)
	}
	mixed := (codeIndexScope{total: 4992, withBody: 4992, commits: 2}).mixedCommitNote()
	for _, want := range []string{"NOT AT ONE COMMIT", "4992", "2 different indexed commits", "REFUSED", "index_code_symbols"} {
		if !strings.Contains(mixed, want) {
			t.Errorf("the mixed-commit note is missing %q; got:\n%s", want, mixed)
		}
	}
}
