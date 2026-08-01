package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// ── Edit A: the declared output key was dead ────────────────────────────────
//
// bugs_open/119. The code read `output_type`; the fleet writes `output_format`.
// Measured 2026-08-01 over all 134 active execute_llm_prompt steps: output_type
// on 6, output_format on 100 (90 of them "json", across 32 agents — every
// council seat included). So ~75% of LLM steps declared a contract nothing
// honoured.

func TestGetOutputTypeReadsTheKeyTheFleetActuallyWrites(t *testing.T) {
	// The regression this pins: before the fix this returned "" and the step
	// silently received getDefaultOutputInstructions(), which says nothing about
	// JSON at all.
	if got := getOutputType(map[string]interface{}{"output_format": "json"}); got != "json" {
		t.Fatalf("output_format json must resolve to json, got %q", got)
	}
	if got := getOutputType(map[string]interface{}{"output_type": "json"}); got != "json" {
		t.Fatalf("output_type json must still resolve to json, got %q", got)
	}
}

func TestGetOutputTypePrefersOutputTypeAndIgnoresEmpties(t *testing.T) {
	// output_type wins when both are present and non-empty.
	cfg := map[string]interface{}{"output_type": "text", "output_format": "json"}
	if got := getOutputType(cfg); got != "text" {
		t.Fatalf("output_type must win over output_format, got %q", got)
	}
	// A present-but-empty output_type must not shadow a good output_format —
	// the platform's recurring "present-but-empty passes silently" shape.
	cfg = map[string]interface{}{"output_type": "", "output_format": "json"}
	if got := getOutputType(cfg); got != "json" {
		t.Fatalf("empty output_type must fall through to output_format, got %q", got)
	}
	if got := getOutputType(map[string]interface{}{}); got != "" {
		t.Fatalf("no declaration must resolve to empty, got %q", got)
	}
}

// The vocabulary gate is load-bearing, not tidiness: query_database reads the
// SAME KEY with a DIFFERENT vocabulary (array/object, database_actions.go:26).
// A blind pass-through would one day select a wrong instruction set.
func TestGetOutputTypeRejectsAnotherActionsVocabulary(t *testing.T) {
	for _, foreign := range []string{"array", "object", "csv", "ARRAY"} {
		if got := getOutputType(map[string]interface{}{"output_format": foreign}); got != "" {
			t.Fatalf("output_format %q belongs to query_database and must NOT be honoured here, got %q", foreign, got)
		}
	}
	// ...while the LLM vocabulary is accepted case/space-insensitively, because
	// config is hand-written.
	for raw, want := range map[string]string{
		" json ": "json", "JSON": "json", "Text": "text", "html": "html", "markdown": "markdown",
	} {
		if got := getOutputType(map[string]interface{}{"output_format": raw}); got != want {
			t.Fatalf("output_format %q must resolve to %q, got %q", raw, want, got)
		}
	}
}

func TestResolveOutputTypePrecedence(t *testing.T) {
	step := map[string]interface{}{"output_format": "json"}
	ai := map[string]interface{}{"output_format": "text"}
	if got := resolveOutputType(step, ai); got != "json" {
		t.Fatalf("step config must win over ai_service, got %q", got)
	}
	if got := resolveOutputType(map[string]interface{}{}, ai); got != "text" {
		t.Fatalf("ai_service must be the fallback, got %q", got)
	}
}

// The end-to-end assertion for Edit A, at the seam that actually matters: a
// council seat's real config shape must now receive the JSON syntax instruction.
// This is the one sentence the 90 JSON-declaring steps were missing, and it is
// the exact failure bugs_open/119 documents (a review object closed one bracket
// early).
func TestSeatConfigNowReceivesTheJSONSyntaxInstruction(t *testing.T) {
	seatConfig := map[string]interface{}{
		"output_format":       "json", // exactly what council-gate's review_* steps carry
		"tolerate_truncation": true,
	}
	got := appendOutputInstructions("REVIEW THE PLAN.", map[string]interface{}{}, seatConfig, zap.NewNop())

	if !strings.Contains(got, "Ensure valid JSON syntax") {
		t.Fatalf("a step declaring output_format json must be told to ensure valid JSON syntax; got:\n%s", got)
	}
	if !strings.Contains(got, "REVIEW THE PLAN.") {
		t.Fatal("the original prompt must be preserved")
	}
	// Negative control: a step that declared nothing must NOT be given the JSON
	// instructions, or this test would pass for the wrong reason.
	plain := appendOutputInstructions("WRITE SOME PROSE.", map[string]interface{}{}, map[string]interface{}{}, zap.NewNop())
	if strings.Contains(plain, "Ensure valid JSON syntax") {
		t.Fatal("an undeclared step must keep the default instructions, not gain JSON ones")
	}
}

// ── Edit B: the corrective re-ask ───────────────────────────────────────────

// The two failure modes have OPPOSITE remedies, so one generic "please return
// valid JSON" would be useless against a truncation. Pin that they differ.
func TestCorrectiveReaskPromptDiffersByFailureMode(t *testing.T) {
	base := "ORIGINAL PROMPT: judge this plan."

	truncated := correctiveReaskPrompt(base, true)
	malformed := correctiveReaskPrompt(base, false)

	if truncated == malformed {
		t.Fatal("a truncation and a malformed answer must not get the same corrective instruction")
	}
	// Both must carry the original question — a reviewer cannot re-judge a plan
	// it can no longer see.
	for name, p := range map[string]string{"truncated": truncated, "malformed": malformed} {
		if !strings.HasPrefix(p, base) {
			t.Fatalf("%s re-ask must append to the original prompt, not replace it", name)
		}
		if !strings.Contains(p, "RETRY") {
			t.Fatalf("%s re-ask must tell the model this is a retry", name)
		}
	}

	// Truncation: the remedy is brevity WITHOUT losing findings. It must not ask
	// for a bigger budget — platform/aiservice/truncation.go:26-29 is explicit
	// that raising caps does not fix the class.
	if !strings.Contains(truncated, "SHORTER") {
		t.Fatal("a truncated re-ask must ask for a shorter answer")
	}
	if !strings.Contains(truncated, "Cut words, never findings") &&
		!strings.Contains(truncated, "cut words, never findings") {
		t.Fatal("a truncated re-ask must protect findings while cutting length")
	}
	for _, capTalk := range []string{"max_tokens", "token limit will be raised", "larger budget"} {
		if strings.Contains(truncated, capTalk) {
			t.Fatalf("the re-ask must not promise or request a cap change (%q)", capTalk)
		}
	}

	// Malformed: the remedy is structural care, and bracket balance specifically
	// — the documented 119 failure is a stray "]".
	if !strings.Contains(malformed, "NOT VALID JSON") {
		t.Fatal("a malformed re-ask must name the actual failure")
	}
	if !strings.Contains(malformed, "closed by exactly one") {
		t.Fatal("a malformed re-ask must address bracket balance, the documented 119 failure")
	}
	if strings.Contains(malformed, "SHORTER") {
		t.Fatal("a malformed answer was not too long — asking for brevity aims at the wrong cause")
	}
}
