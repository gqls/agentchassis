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
func TestSymbolTokenClauseGoReceiver(t *testing.T) {
	clause, args := symbolTokenClause("OllamaClient.GenerateText", "", 40)
	if !strings.Contains(clause, "symbol ILIKE '%' || $1 || '%'") ||
		!strings.Contains(clause, "symbol ILIKE '%' || $2 || '%'") {
		t.Fatalf("expected two AND-ed symbol substring clauses, got: %s", clause)
	}
	if len(args) != 3 || args[0] != "OllamaClient" || args[1] != "GenerateText" || args[2] != "" {
		t.Fatalf("args = %v, want [OllamaClient GenerateText \"\"]", args)
	}
}

func TestSymbolTokenClauseRepoFilter(t *testing.T) {
	_, args := symbolTokenClause("GenerateText", "gqls/agentchassis", 40)
	if args[len(args)-1] != "gqls/agentchassis" {
		t.Fatalf("repo filter should be the last arg, got %v", args)
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
