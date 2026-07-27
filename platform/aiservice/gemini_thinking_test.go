// FILE: platform/aiservice/gemini_thinking_test.go
//
// Regression net for the reason the Gemini switch was reverted on 2026-07-24
// (commit 4dd5d6378): Gemini's maxOutputTokens is a TOTAL output ceiling that
// thinking is drawn from first, so passing this platform's Anthropic-sized
// max_tokens straight through starved the answer — zero visible text at the
// 100-token tier, ~85 characters at the 500-token tier. That presented as the
// model being incapable rather than as a budget being mis-sized.
//
// These tests assert the two properties that make the failure impossible to
// reproduce silently: the visible budget is provisioned for, and when a cut
// happens anyway the error names thinking as the consumer.
package aiservice

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// capturingTransport records the request body so a test can assert what went on
// the wire, then returns a canned response.
type capturingTransport struct {
	body     string
	status   int
	captured []byte
}

func (c *capturingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		c.captured, _ = io.ReadAll(req.Body)
	}
	status := c.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(c.body)),
		Header:     make(http.Header),
	}, nil
}

func geminiWith(model, body string) (*GeminiClient, *capturingTransport) {
	tr := &capturingTransport{body: body}
	return &GeminiClient{
		apiKey:          "test-key",
		model:           model,
		httpClient:      &http.Client{Transport: tr},
		thinkingReserve: defaultGeminiThinkingReserve,
		embeddingModel:  "text-embedding-004",
	}, tr
}

// sentMaxOutputTokens pulls generationConfig.maxOutputTokens out of a captured
// request body.
func sentMaxOutputTokens(t *testing.T, captured []byte) int {
	t.Helper()
	var sent struct {
		GenerationConfig struct {
			MaxOutputTokens int                    `json:"maxOutputTokens"`
			ThinkingConfig  map[string]interface{} `json:"thinkingConfig"`
		} `json:"generationConfig"`
	}
	if err := json.Unmarshal(captured, &sent); err != nil {
		t.Fatalf("request body was not valid JSON: %v", err)
	}
	return sent.GenerationConfig.MaxOutputTokens
}

const geminiOKBody = `{
	"candidates": [{
		"content": {"parts": [{"text": "a complete answer"}]},
		"finishReason": "STOP"
	}],
	"usageMetadata": {"promptTokenCount": 10, "candidatesTokenCount": 4, "thoughtsTokenCount": 300, "totalTokenCount": 314}
}`

func TestGeminiReservesOutputBudgetForThinking(t *testing.T) {
	// The exact tier that produced ZERO visible text in production.
	client, tr := geminiWith("gemini-pro-latest", geminiOKBody)
	options := map[string]interface{}{"max_tokens": 100}

	if _, err := client.GenerateText(context.Background(), "prompt", options); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := sentMaxOutputTokens(t, tr.captured)
	want := 100 + defaultGeminiThinkingReserve
	if got != want {
		t.Errorf("maxOutputTokens = %d, want %d — the caller's 100 tokens are a VISIBLE-text budget, and on a thinking model thinking is spent from the same ceiling first. Sending 100 is what produced zero text on 2026-07-24", got, want)
	}
	if options["__sent_visible_budget_tokens"] != 100 {
		t.Errorf("__sent_visible_budget_tokens = %v, want 100 — the caller's ask must stay legible in llm_call_log alongside what was sent", options["__sent_visible_budget_tokens"])
	}
	if options["__sent_max_tokens"] != want {
		t.Errorf("__sent_max_tokens = %v, want %d (what actually went on the wire, matching anthropic.go)", options["__sent_max_tokens"], want)
	}
}

func TestGeminiDoesNotReserveForANonThinkingModel(t *testing.T) {
	// flash-lite was measured on 2026-07-24 as having no thinking overhead:
	// it works cleanly at every budget, so the cap means what it says.
	client, tr := geminiWith("gemini-flash-lite-latest", geminiOKBody)

	if _, err := client.GenerateText(context.Background(), "prompt", map[string]interface{}{"max_tokens": 500}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := sentMaxOutputTokens(t, tr.captured); got != 500 {
		t.Errorf("maxOutputTokens = %d, want 500 unchanged — reserving budget on a model that does not think inflates the ceiling for nothing", got)
	}
}

func TestGeminiTreatsAnUnknownModelAsThinking(t *testing.T) {
	// The polarity that matters: an unrecognised Gemini model is almost always a
	// NEWER one, and every generation since 2.5 thinks by default. An allow-list
	// here would under-provision each new model on the day it ships, and the
	// symptom is bad copy rather than an error.
	client, tr := geminiWith("gemini-9-ultra-preview", geminiOKBody)

	if _, err := client.GenerateText(context.Background(), "prompt", map[string]interface{}{"max_tokens": 500}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := sentMaxOutputTokens(t, tr.captured); got != 500+defaultGeminiThinkingReserve {
		t.Errorf("maxOutputTokens = %d, want %d — an unknown model must be assumed to think", got, 500+defaultGeminiThinkingReserve)
	}
}

func TestGeminiTruncationNamesThinkingAsTheConsumer(t *testing.T) {
	// The production shape from 2026-07-24: the whole ceiling went on thinking
	// and not one visible token came back. The old message said only
	// "finishReason=MAX_TOKENS", which is indistinguishable from a prompt that
	// simply wanted to write more.
	const body = `{
		"candidates": [{
			"content": {"parts": []},
			"finishReason": "MAX_TOKENS"
		}],
		"usageMetadata": {"promptTokenCount": 900, "candidatesTokenCount": 0, "thoughtsTokenCount": 100, "totalTokenCount": 1000}
	}`

	client, _ := geminiWith("gemini-pro-latest", body)
	_, err := client.GenerateText(context.Background(), "prompt", nil)
	if err == nil {
		t.Fatal("a completion cut at MAX_TOKENS must return an error")
	}
	te, ok := IsTruncated(err)
	if !ok {
		t.Fatalf("want a *TruncatedError callers can detect by type, got %T: %v", err, err)
	}
	if te.Provider != "gemini" {
		t.Errorf("Provider = %q, want gemini — a mixed-provider deployment must be able to tell where a cut came from", te.Provider)
	}
	for _, want := range []string{"thinking", "thinking_reserve_tokens"} {
		if !strings.Contains(te.Reason, want) {
			t.Errorf("Reason %q must contain %q: a cut with zero visible text and non-zero thinking has a config fix, and the message is where it gets found", te.Reason, want)
		}
	}
}

func TestGeminiDropsThoughtPartsFromTheAnswer(t *testing.T) {
	// Gemini returns reasoning and answer in the SAME parts array. Concatenating
	// every part splices the model's thinking into published page copy, and
	// nothing above this layer reads it closely enough to notice.
	const body = `{
		"candidates": [{
			"content": {"parts": [
				{"text": "The user wants three bullets. I should avoid em dashes.", "thought": true},
				{"text": "We build agent systems."}
			]},
			"finishReason": "STOP"
		}],
		"usageMetadata": {"promptTokenCount": 10, "candidatesTokenCount": 5, "thoughtsTokenCount": 12, "totalTokenCount": 27}
	}`

	client, _ := geminiWith("gemini-pro-latest", body)
	got, err := client.GenerateText(context.Background(), "prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "We build agent systems." {
		t.Errorf("got %q, want only the answer part — a part flagged thought:true is reasoning, and it must never reach a page", got)
	}
}

func TestGeminiReportsThinkingTokensSeparately(t *testing.T) {
	client, _ := geminiWith("gemini-pro-latest", geminiOKBody)
	options := map[string]interface{}{}
	if _, err := client.GenerateText(context.Background(), "prompt", options); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Visible output stays in __usage_output_tokens so the field means the same
	// thing across providers; thinking is billed as output but is not answer, so
	// folding it in silently would make Gemini look verbose rather than pensive.
	if options["__usage_output_tokens"] != 4 {
		t.Errorf("__usage_output_tokens = %v, want 4 (VISIBLE tokens, as for every other provider)", options["__usage_output_tokens"])
	}
	if options["__usage_thinking_tokens"] != 300 {
		t.Errorf("__usage_thinking_tokens = %v, want 300 — thinking is billed, so it has to be visible to cost accounting", options["__usage_thinking_tokens"])
	}
}

func TestGeminiRefusesAModelClosedToThisKey(t *testing.T) {
	// A pinned 2.5 snapshot answers 404 "no longer available to new users" for a
	// key issued after Google gated that generation, on every call, mid-run.
	// It is a config error, so it fails at construction with the fix named.
	t.Setenv("GEMINI_TEST_KEY", "test-key")
	_, err := NewGeminiClient(context.Background(), map[string]interface{}{
		"api_key_env_var": "GEMINI_TEST_KEY",
		"model":           "gemini-2.5-pro",
	})
	if err == nil {
		t.Fatal("a model closed to this key must be refused at construction, not on every generate call")
	}
	if !strings.Contains(err.Error(), "gemini-pro-latest") {
		t.Errorf("error %q must name the replacement model — the message is the whole point of failing early", err)
	}
}

func TestGeminiRefusesAnAbsentModelRatherThanDefaulting(t *testing.T) {
	// The old default was gemini-2.5-pro, which has since been closed to this
	// key: a default that rots into a 404 reads as an outage, not a config error.
	t.Setenv("GEMINI_TEST_KEY", "test-key")
	_, err := NewGeminiClient(context.Background(), map[string]interface{}{
		"api_key_env_var": "GEMINI_TEST_KEY",
	})
	if err == nil {
		t.Fatal("an absent ai_service.model must be refused, not silently defaulted")
	}
}

func TestGeminiRejectsBothThinkingKnobsAtOnce(t *testing.T) {
	// thinking_level is the 3.x control and thinking_budget_tokens the 2.5 one.
	// Sending both is a 400 on every call; catching it at construction turns a
	// fleet-wide outage into a startup error.
	t.Setenv("GEMINI_TEST_KEY", "test-key")
	_, err := NewGeminiClient(context.Background(), map[string]interface{}{
		"api_key_env_var":        "GEMINI_TEST_KEY",
		"model":                  "gemini-pro-latest",
		"thinking_level":         "low",
		"thinking_budget_tokens": 512,
	})
	if err == nil {
		t.Fatal("setting both thinking knobs must be refused at construction")
	}
}

func TestGeminiSendsTheConfiguredThinkingKnobAndNothingOtherwise(t *testing.T) {
	// Default: no thinkingConfig at all. The two generations take incompatible
	// knobs and guessing wrong fails every call, so the reserve carries the
	// default case and the knob is opt-in.
	client, tr := geminiWith("gemini-pro-latest", geminiOKBody)
	if _, err := client.GenerateText(context.Background(), "prompt", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Contains(tr.captured, []byte("thinkingConfig")) {
		t.Errorf("no thinkingConfig may be sent by default; got %s", tr.captured)
	}

	client, tr = geminiWith("gemini-pro-latest", geminiOKBody)
	client.thinkingLevel = "low"
	if _, err := client.GenerateText(context.Background(), "prompt", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(tr.captured, []byte(`"thinkingLevel":"low"`)) {
		t.Errorf("configured thinking_level must reach the wire; got %s", tr.captured)
	}
}
