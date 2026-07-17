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
