package actions

import (
	"strings"
	"testing"
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
