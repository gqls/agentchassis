package datahelpers

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// Locks the contract that the experience-planner's prompts depend on, and that
// the 2026-07-18 council failure exposed.
//
// 001_development_guide records execute_llm_prompt's RAW collected_data shape as
// {type, result} — that is what config dot-paths (plan_body_field,
// review_fields) read, via ExtractNestedField, and they must keep ".result".
//
// But prompt TEMPLATE data goes through ExtractFields -> UnwrapDeep, which
// strips the {result,type} wrapper (unwrapRecursive Pattern 2). So inside a
// prompt template the step's output IS the value, and "{{.step.result}}" is a
// nested access on an unwrapped value — an execution error for a text step
// (001 §Step 6: "{{.missing.nested}} -> execution error"), and a silent
// "<no value>" for a JSON step. Both are wrong; the asymmetry is the trap.
func TestUnwrapDeep_TemplateVsConfigPaths(t *testing.T) {
	logger := zap.NewNop()

	// A text-output LLM step (the experience planner's compose step).
	textStep := map[string]interface{}{"type": "text", "result": "# EXPERIENCE_PLAN\nbody"}
	// A json-output LLM step (a council critic).
	jsonStep := map[string]interface{}{"type": "json", "result": map[string]interface{}{
		"reviewer": "journeys", "verdict": "object",
	}}
	collected := map[string]interface{}{"proposal": textStep, "review_journeys": jsonStep}

	// --- Config paths (raw collected_data) keep .result ---
	if got := ExtractNestedFieldString(collected, "proposal.result"); !strings.Contains(got, "EXPERIENCE_PLAN") {
		t.Fatalf("config path proposal.result must resolve the body, got %q", got)
	}
	if got := ExtractNestedField(collected, "review_journeys.result"); got == nil {
		t.Fatal("config path review_journeys.result must resolve the critic object")
	}

	// --- Template data is UNWRAPPED ---
	td := ExtractFields(collected, []string{"proposal", "review_journeys"}, logger)

	if _, isMap := td["proposal"].(map[string]interface{}); isMap {
		t.Fatal("proposal should be unwrapped to its value, not remain a {result,type} map")
	}
	if s, ok := td["proposal"].(string); !ok || !strings.Contains(s, "EXPERIENCE_PLAN") {
		t.Fatalf("proposal should unwrap to the plan string, got %T %v", td["proposal"], td["proposal"])
	}

	rev, ok := td["review_journeys"].(map[string]interface{})
	if !ok {
		t.Fatalf("review_journeys should unwrap to the critic object, got %T", td["review_journeys"])
	}
	if _, stillWrapped := rev["result"]; stillWrapped {
		t.Fatal("review_journeys must not still carry a 'result' key after unwrapping")
	}
	if rev["reviewer"] != "journeys" {
		t.Fatalf("unwrapped critic object lost its fields: %v", rev)
	}

	// --- The fix: bare {{.step}} renders; {{.step.result}} does not ---
	if out, err := RenderPromptTemplate("{{.proposal}}", td, *logger); err != nil || !strings.Contains(out, "EXPERIENCE_PLAN") {
		t.Fatalf("{{.proposal}} must render the plan: out=%q err=%v", out, err)
	}
	if _, err := RenderPromptTemplate("{{.proposal.result}}", td, *logger); err == nil {
		t.Fatal("{{.proposal.result}} must ERROR on an unwrapped text step (this was the live failure)")
	}

	// JSON steps fail SILENTLY rather than loudly — the more dangerous half.
	out, err := RenderPromptTemplate("[{{.review_journeys.result}}]", td, *logger)
	if err == nil && !strings.Contains(out, "journeys") {
		t.Logf("CONFIRMED silent degradation: {{.review_journeys.result}} rendered %q", out)
	}
	if out2, err2 := RenderPromptTemplate("{{.review_journeys}}", td, *logger); err2 != nil || !strings.Contains(out2, "journeys") {
		t.Fatalf("{{.review_journeys}} must render the critic object: out=%q err=%v", out2, err2)
	}
}
