package actions

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func TestAnyPatternMatches(t *testing.T) {
	cases := []struct {
		name     string
		patterns []string
		files    []string
		corpus   string
		want     bool
	}{
		{"file path substring", []string{"aiservice/"}, []string{"platform/aiservice/anthropic.go"}, "", true},
		{"case-insensitive", []string{"ANTHROPIC.GO"}, []string{"platform/aiservice/anthropic.go"}, "", true},
		{"corpus fallback (diagnosis names a table)", []string{"llm_call_log"}, []string{"cmd/main.go"}, "the bug shows in llm_call_log rows", true},
		{"no match", []string{"adoption"}, []string{"platform/aiservice/anthropic.go"}, "unrelated text", false},
		{"empty pattern skipped, no match", []string{"", "  "}, []string{"a.go"}, "b", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			lf := make([]string, len(c.files))
			for i, f := range c.files {
				lf[i] = lower(f)
			}
			got := anyPatternMatches(c.patterns, lf, lower(c.corpus))
			if got != c.want {
				t.Fatalf("anyPatternMatches(%v, %v, %q) = %v, want %v", c.patterns, c.files, c.corpus, got, c.want)
			}
		})
	}
}

// lower is a tiny local helper mirroring the action's lower-casing so the test
// exercises anyPatternMatches on the same shape the action feeds it.
func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

func TestSelectReviewPanel_FailOpenAndMatch(t *testing.T) {
	logger := zap.NewNop()
	collected := map[string]interface{}{
		"plan_persisted": map[string]interface{}{
			"files":     []interface{}{"platform/aiservice/anthropic.go"},
			"plan_json": `{"summary":"decode stop_reason","edits":[{"file":"platform/aiservice/anthropic.go"}]}`,
		},
		"diagnosis_row": map[string]interface{}{
			"conclusion": "GenerateText never decodes stop_reason; 17 llm_call_log rows show it",
		},
	}
	params := ActionParams{
		CollectedData: collected,
		Logger:        logger,
		StepConfig: models.Step{Config: map[string]interface{}{
			"plan_field":        "plan_persisted",
			"extra_text_fields": []interface{}{"diagnosis_row.conclusion"},
			"footprints": map[string]interface{}{
				"llm_reliability": []interface{}{"aiservice/", "ai_actions.go", "llm_call_log"}, // should fire (file + corpus)
				"render":          []interface{}{"rerender", "save_page_sections"},              // should NOT fire
				"misconfigured":   []interface{}{},                                              // empty footprint -> fail-open -> fires
			},
		}},
	}

	out, err := SelectReviewPanelAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", out)
	}
	if m["run_llm_reliability"] != true {
		t.Errorf("llm_reliability should fire (matches aiservice/ file + llm_call_log corpus): %v", m["run_llm_reliability"])
	}
	if m["run_render"] != false {
		t.Errorf("render should NOT fire (no rerender/save_page_sections anywhere): %v", m["run_render"])
	}
	if m["run_misconfigured"] != true {
		t.Errorf("empty footprint must fail-open (run), never silently skip: %v", m["run_misconfigured"])
	}
}

func TestToStringSlice(t *testing.T) {
	if got := toStringSlice([]interface{}{"a", "b", ""}); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("toStringSlice dropped/mangled: %v", got)
	}
	if got := toStringSlice("solo"); len(got) != 1 || got[0] != "solo" {
		t.Fatalf("toStringSlice(string): %v", got)
	}
	if got := toStringSlice(nil); got != nil {
		t.Fatalf("toStringSlice(nil) should be nil: %v", got)
	}
}

// sanity: ExtractNestedField reaches plan_persisted.files as the action expects.
func TestExtractStringListReachesFiles(t *testing.T) {
	data := map[string]interface{}{
		"plan_persisted": map[string]interface{}{
			"files": []interface{}{"x.go", "y.go"},
		},
	}
	got := extractStringList(data, "plan_persisted.files")
	if len(got) != 2 {
		t.Fatalf("expected 2 files via %T path, got %v (ExtractNestedField=%v)", data, got, datahelpers.ExtractNestedField(data, "plan_persisted.files"))
	}
}
