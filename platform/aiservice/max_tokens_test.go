// FILE: platform/aiservice/max_tokens_test.go
//
// Wire-shape tests for bugs_open/257: the output-token budget must be resolved
// from the config the client was CONSTRUCTED with, not only from the per-call
// options map.
//
// Every assertion is against the request BODY, for the reason prompt_caching_test
// gives: a client that silently dropped the configured budget would pass any
// response-only test while quietly truncating every reply — which is exactly the
// failure this seam exists to prevent, and it is invisible from the outside
// because the answers keep coming back, just shorter.
//
// Each test states what its DISCONFIRMING result would be. A check that passes on
// a binary which still ignores the config is not a check (bug 257 §6).
package aiservice

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
)

// --- harnesses: built through the REAL constructors, so the test covers config
// parsing and request building together. Only the transport is swapped, which is
// the one thing a unit test must not do for real.

const okAnthropicBody = `{"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`
const okGeminiBody = `{"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`
const okOllamaBody = `{"message":{"role":"assistant","content":"ok"},"done":true}`

func anthropicFromConfig(t *testing.T, cfg map[string]interface{}) (*AnthropicClient, *capturingTransport) {
	t.Helper()
	t.Setenv("MAXTOK_TEST_KEY", "test-key")
	cfg["api_key_env_var"] = "MAXTOK_TEST_KEY"
	if _, ok := cfg["model"]; !ok {
		cfg["model"] = "claude-test"
	}
	c, err := NewAnthropicClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewAnthropicClient: %v", err)
	}
	tr := &capturingTransport{body: okAnthropicBody}
	c.httpClient = &http.Client{Transport: tr}
	return c, tr
}

func geminiFromConfig(t *testing.T, cfg map[string]interface{}) (*GeminiClient, *capturingTransport) {
	t.Helper()
	t.Setenv("MAXTOK_TEST_KEY", "test-key")
	cfg["api_key_env_var"] = "MAXTOK_TEST_KEY"
	if _, ok := cfg["model"]; !ok {
		// A non-thinking model, so maxOutputTokens is the visible budget with no
		// reserve added. The thinking case is asserted separately.
		cfg["model"] = "gemini-2.0-flash-lite"
	}
	c, err := NewGeminiClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewGeminiClient: %v", err)
	}
	tr := &capturingTransport{body: okGeminiBody}
	c.httpClient = &http.Client{Transport: tr}
	return c, tr
}

func ollamaFromConfig(t *testing.T, cfg map[string]interface{}) (*OllamaClient, *capturingTransport) {
	t.Helper()
	if _, ok := cfg["model"]; !ok {
		cfg["model"] = "llama-test"
	}
	cfg["api_url"] = "http://ollama.test"
	c, err := NewOllamaClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewOllamaClient: %v", err)
	}
	tr := &capturingTransport{body: okOllamaBody}
	c.httpClient = &http.Client{Transport: tr}
	return c, tr
}

// bodyField digs a top-level numeric field out of a captured request body.
// Returns (0, false) when the key is ABSENT — a distinction that matters for
// ollama, where "no num_predict at all" is the correct unconfigured behaviour
// and is not the same as "num_predict: 0".
func bodyField(t *testing.T, captured []byte, path ...string) (float64, bool) {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(captured, &m); err != nil {
		t.Fatalf("captured body is not valid JSON: %v\n%s", err, captured)
	}
	var cur interface{} = m
	for _, k := range path {
		asMap, ok := cur.(map[string]interface{})
		if !ok {
			return 0, false
		}
		cur, ok = asMap[k]
		if !ok {
			return 0, false
		}
	}
	n, ok := cur.(float64)
	return n, ok
}

// ============================================================================
// THE BUG ITSELF: a configured budget must reach the API through a nil options
// map. Pre-fix this sent 2048 no matter what the config said.
// ============================================================================

func TestAnthropicConfiguredBudgetReachesTheWireThroughNilOptions(t *testing.T) {
	c, tr := anthropicFromConfig(t, map[string]interface{}{"max_tokens": float64(8000)})

	// nil — the exact shape internal/agents/reasoning/agent.go:127 uses, and the
	// shape bug 257 says "reproduces this bug by construction".
	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	got, ok := bodyField(t, tr.captured, "max_tokens")
	if !ok {
		t.Fatal("request carries no max_tokens at all")
	}
	// DISCONFIRMING RESULT: 2048. That is what a binary without this fix sends,
	// and it is what this assertion was written to be able to fail on.
	if int(got) != 8000 {
		t.Fatalf("configured ai_service.max_tokens=8000 did not reach the wire: sent %d.\n"+
			"%d means the client is still ignoring its construction config "+
			"(bugs_open/257) — the config key is inert again.", int(got), int(got))
	}
}

// THE NEGATIVE CONTROL, and the test that stops the one above proving nothing.
// If the fix were "always send 8000" this would fail; if the assertion above
// could not fail, this one shows the same code path producing a different number.
func TestAnthropicUnconfiguredClientStillSendsTheFallback(t *testing.T) {
	c, tr := anthropicFromConfig(t, map[string]interface{}{}) // no max_tokens

	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	got, _ := bodyField(t, tr.captured, "max_tokens")
	if int(got) != DefaultMaxOutputTokens {
		t.Fatalf("an unconfigured client must send the unchanged fallback %d, sent %d — "+
			"this change is opt-in and must be byte-identical for configs that never "+
			"named max_tokens (owner ruling 2026-08-02 §2)", DefaultMaxOutputTokens, int(got))
	}
}

// The canonical ExecuteLLMPromptAction path in one assertion: it ALWAYS passes
// options built from the same config, so it must be unaffected. If per-call
// options ever stopped winning, 127 live steps would silently re-cap.
func TestPerCallOptionsStillOverrideTheConfiguredBudget(t *testing.T) {
	c, tr := anthropicFromConfig(t, map[string]interface{}{"max_tokens": float64(8000)})

	opts := map[string]interface{}{"max_tokens": 500}
	if _, err := c.GenerateText(context.Background(), "p", opts); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}

	got, _ := bodyField(t, tr.captured, "max_tokens")
	if int(got) != 500 {
		t.Fatalf("per-call options must win over the construction config: sent %d, wanted 500. "+
			"Precedence is options > ai_service config > fallback", int(got))
	}
	// Telemetry parity: llm_call_log.max_tokens is fed solely by this key.
	if v, ok := opts["__sent_max_tokens"].(int); !ok || v != 500 {
		t.Fatalf("__sent_max_tokens must record what was actually sent, got %v", opts["__sent_max_tokens"])
	}
}

// The write-back must also reflect a budget that came from the CONFIG, or
// llm_call_log.max_tokens under-reports exactly the calls this fix enables.
func TestSentMaxTokensRecordsAConfigSourcedBudget(t *testing.T) {
	c, _ := anthropicFromConfig(t, map[string]interface{}{"max_tokens": float64(8000)})

	opts := map[string]interface{}{} // non-nil, but carries no budget
	if _, err := c.GenerateText(context.Background(), "p", opts); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	if v, ok := opts["__sent_max_tokens"].(int); !ok || v != 8000 {
		t.Fatalf("__sent_max_tokens should be the config-sourced 8000, got %v — "+
			"llm_call_log would under-report the budget for every direct caller", opts["__sent_max_tokens"])
	}
}

// ============================================================================
// THE TYPE TRAP: the same key arrives as float64 from jsonb (agent_definitions)
// and as int from YAML (configs/*.yaml via viper). ai_actions.go:361 reads
// float64 ONLY, which is why configs/reasoning-agent.yaml's int was doubly dead.
// ============================================================================

func TestJSONBFloatAndYAMLIntProduceTheSameBudget(t *testing.T) {
	cases := []struct {
		name string
		val  interface{}
	}{
		{"jsonb float64", float64(16000)},
		{"yaml int", int(16000)},
		{"int64", int64(16000)},
		{"json.Number", json.Number("16000")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, tr := anthropicFromConfig(t, map[string]interface{}{"max_tokens": tc.val})
			if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
				t.Fatalf("GenerateText: %v", err)
			}
			got, _ := bodyField(t, tr.captured, "max_tokens")
			if int(got) != 16000 {
				t.Fatalf("%s config gave %d, wanted 16000 — a reader that accepts only "+
					"one numeric type silently drops whole config paths", tc.name, int(got))
			}
		})
	}
}

// Values that are present but unusable must fall back, not be sent. `max_tokens: 0`
// is a 400 from Anthropic, so treating it as "configured" would turn a typo into
// an outage.
func TestUnusableConfiguredValuesFallBackRatherThanBeingSent(t *testing.T) {
	for _, bad := range []interface{}{float64(0), int(0), float64(-1), "8000", nil, true} {
		c, tr := anthropicFromConfig(t, map[string]interface{}{"max_tokens": bad})
		if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
			t.Fatalf("GenerateText: %v", err)
		}
		got, _ := bodyField(t, tr.captured, "max_tokens")
		if int(got) != DefaultMaxOutputTokens {
			t.Fatalf("max_tokens=%#v (unusable) produced %d; it must be treated as unset "+
				"and fall back to %d", bad, int(got), DefaultMaxOutputTokens)
		}
	}
}

// configMaxTokens asserted DIRECTLY, because the wire-level test above cannot
// see this guard: the floor in generate() sits in SERIES with it, so a resolver
// that wrongly returned (0, true) still produced a 2048 request and the test
// passed for the wrong reason. Found by mutation M9 on 2026-08-16 — the mutation
// SURVIVED, which is the only reason this test exists.
//
// The (int, bool) pair is the thing under test: `false` means "no operator chose
// a budget", and only the resolver can say that. Downstream everything is an int,
// and an int cannot carry the distinction — which is why ollama (omit the field)
// and anthropic (send a floor) can disagree about what to do about it.
func TestConfigMaxTokensDistinguishesAChosenBudgetFromAnUnusableOne(t *testing.T) {
	chosen := []struct {
		name string
		val  interface{}
		want int
	}{
		{"jsonb float64", float64(8000), 8000},
		{"yaml int", int(8000), 8000},
		{"int64", int64(8000), 8000},
		{"json.Number", json.Number("8000"), 8000},
		{"one", int(1), 1},
	}
	for _, tc := range chosen {
		got, ok := configMaxTokens(map[string]interface{}{"max_tokens": tc.val})
		if !ok || got != tc.want {
			t.Errorf("%s: configMaxTokens = (%d, %v), wanted (%d, true)", tc.name, got, ok, tc.want)
		}
	}

	unusable := []struct {
		name string
		cfg  map[string]interface{}
	}{
		{"zero float64", map[string]interface{}{"max_tokens": float64(0)}},
		{"zero int", map[string]interface{}{"max_tokens": int(0)}},
		{"zero int64", map[string]interface{}{"max_tokens": int64(0)}},
		{"zero json.Number", map[string]interface{}{"max_tokens": json.Number("0")}},
		{"negative", map[string]interface{}{"max_tokens": float64(-1)}},
		{"negative int", map[string]interface{}{"max_tokens": int(-1)}},
		{"string", map[string]interface{}{"max_tokens": "8000"}},
		{"bool", map[string]interface{}{"max_tokens": true}},
		{"nil value", map[string]interface{}{"max_tokens": nil}},
		{"absent", map[string]interface{}{}},
		{"nil config", nil},
		{"unparseable json.Number", map[string]interface{}{"max_tokens": json.Number("eight thousand")}},
	}
	for _, tc := range unusable {
		got, ok := configMaxTokens(tc.cfg)
		if ok {
			t.Errorf("%s: configMaxTokens reported a CHOSEN budget of %d. Zero is not a budget, "+
				"it is a 400 from Anthropic; anything unusable must read as unset so the "+
				"caller can apply its own policy", tc.name, got)
		}
		if got != 0 {
			t.Errorf("%s: configMaxTokens returned %d alongside ok=false; the int must be 0 "+
				"so a caller ignoring the bool cannot silently send it", tc.name, got)
		}
	}
}

// resolvedMaxTokens is the anthropic/gemini policy over that answer: chosen wins,
// anything else becomes the fleet floor. Asserted separately from ollama's, which
// deliberately implements the OTHER policy.
func TestResolvedMaxTokensAppliesTheFloorOnlyWhenNothingWasChosen(t *testing.T) {
	if got := resolvedMaxTokens(map[string]interface{}{"max_tokens": float64(8000)}); got != 8000 {
		t.Errorf("a chosen budget must survive: got %d, wanted 8000", got)
	}
	if got := resolvedMaxTokens(map[string]interface{}{"max_tokens": float64(0)}); got != DefaultMaxOutputTokens {
		t.Errorf("an unusable budget must become the floor: got %d, wanted %d", got, DefaultMaxOutputTokens)
	}
	if got := resolvedMaxTokens(nil); got != DefaultMaxOutputTokens {
		t.Errorf("a nil config must become the floor: got %d, wanted %d", got, DefaultMaxOutputTokens)
	}
}

// The struct-literal guard, with its failing case demonstrated at the moment it
// is written (register OPP-003: a detector whose failing case was never shown is
// indistinguishable from one that never runs).
//
// Every fake in this package builds a client as a STRUCT LITERAL, bypassing the
// constructor, so maxTokens is the zero value. Without the floor in generate(),
// this sends `"max_tokens": 0` — a 400 — and the whole existing suite passes
// anyway, because nothing else asserts the field. Measured 2026-08-16.
// ⚠ POPULATION MEASURED 2026-08-16, after the same seat asked whether struct-literal
// constructions exist outside test files (advisory low): `grep -rn '&(AnthropicClient|
// GeminiClient|OllamaClient){' --include=*.go` minus tests returns exactly THREE hits
// — anthropic.go:69, gemini.go:186, ollama.go:62 — and all three ARE the constructors
// themselves. So there is no production code building these clients as a bare struct.
// The floor's remaining population is the 7 test fakes across 5 files in this package,
// which is what it was written for. The seat was right to ask: the GenerateText call
// sites were enumerated and this axis was not, until now.
func TestStructLiteralClientCannotSendAZeroBudget(t *testing.T) {
	tr := &capturingTransport{body: okAnthropicBody}
	c := &AnthropicClient{ // deliberately NOT through the constructor
		apiKey:     "k",
		model:      "m",
		httpClient: &http.Client{Transport: tr},
	}
	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	got, _ := bodyField(t, tr.captured, "max_tokens")
	if int(got) != DefaultMaxOutputTokens {
		t.Fatalf("a struct-literal client sent max_tokens=%d; it must be floored to %d. "+
			"Zero is not 'no budget', it is a 400.", int(got), DefaultMaxOutputTokens)
	}
}

// ============================================================================
// THE OTHER TWO PROVIDERS. The bug filing read only the Anthropic client;
// gemini carries the identical hardcoded 2048 and was never mentioned.
// ============================================================================

func TestGeminiConfiguredBudgetReachesMaxOutputTokens(t *testing.T) {
	c, tr := geminiFromConfig(t, map[string]interface{}{"max_tokens": float64(9000)})
	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	got, ok := bodyField(t, tr.captured, "generationConfig", "maxOutputTokens")
	if !ok {
		t.Fatal("gemini request carries no generationConfig.maxOutputTokens")
	}
	if int(got) != 9000 {
		t.Fatalf("gemini sent maxOutputTokens=%d, wanted 9000 (non-thinking model, so no "+
			"reserve is added); %d means the config never reached the client", int(got), int(got))
	}
}

func TestGeminiUnconfiguredStillSendsTheFallback(t *testing.T) {
	c, tr := geminiFromConfig(t, map[string]interface{}{})
	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	got, _ := bodyField(t, tr.captured, "generationConfig", "maxOutputTokens")
	if int(got) != DefaultMaxOutputTokens {
		t.Fatalf("unconfigured gemini must send the unchanged %d, sent %d",
			DefaultMaxOutputTokens, int(got))
	}
}

// A thinking model adds thinkingReserve ON TOP of the visible budget. The
// configured value must behave exactly as a caller-supplied one does here, or
// the reserve silently eats the operator's budget.
func TestGeminiThinkingModelAddsTheReserveToAConfiguredBudget(t *testing.T) {
	c, tr := geminiFromConfig(t, map[string]interface{}{
		"max_tokens": float64(9000),
		"model":      "gemini-pro-latest",
	})
	if !c.thinks() {
		t.Skip("model no longer classified as thinking; the reserve arithmetic is asserted elsewhere")
	}
	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	got, _ := bodyField(t, tr.captured, "generationConfig", "maxOutputTokens")
	if int(got) != 9000+c.thinkingReserve {
		t.Fatalf("thinking model sent maxOutputTokens=%d, wanted visible 9000 + reserve %d = %d",
			int(got), c.thinkingReserve, 9000+c.thinkingReserve)
	}
}

func TestOllamaConfiguredBudgetSetsNumPredict(t *testing.T) {
	c, tr := ollamaFromConfig(t, map[string]interface{}{"max_tokens": float64(4096)})
	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	got, ok := bodyField(t, tr.captured, "options", "num_predict")
	if !ok {
		t.Fatal("ollama request carries no options.num_predict")
	}
	if int(got) != 4096 {
		t.Fatalf("ollama sent num_predict=%d, wanted 4096", int(got))
	}
}

// NEGATIVE CONTROL, and a deliberate asymmetry with anthropic/gemini: an
// unconfigured ollama client must send NO num_predict at all, letting the server
// choose. 257 is about a configured budget failing to arrive; imposing a 2048
// ceiling where there never was one would be a regression dressed as a fix.
func TestOllamaUnconfiguredOmitsNumPredictEntirely(t *testing.T) {
	c, tr := ollamaFromConfig(t, map[string]interface{}{})
	if _, err := c.GenerateText(context.Background(), "p", nil); err != nil {
		t.Fatalf("GenerateText: %v", err)
	}
	if _, present := bodyField(t, tr.captured, "options", "num_predict"); present {
		t.Fatal("unconfigured ollama must omit num_predict entirely, not send a ceiling — " +
			"the server's own default is the pre-existing behaviour and this change does not own it")
	}
}

// ============================================================================
// THE VISION PATH — raised by the council's bug_historian seat (corr 366efae9),
// which was right to ask and is answered here with tests rather than a sentence.
//
// The objection: LANDMINES pairs `.GenerateText(` and `.GenerateWithImages(` as
// the two shapes that bypass ExecuteLLMPromptAction, the submission spoke only of
// GenerateText, and this platform's most-repeated failure is exactly "one call
// site of a shared judgement gets the rigorous fix, the sibling stays heuristic"
// (016b §9; bugs_closed/012, 046, 076 are all silent-truncation cases).
//
// The answer is that both providers route vision through the SAME generate()
// this change fixes — anthropic.go GenerateWithImages -> c.generate(ctx, content,
// options), gemini.go GenerateWithImages -> c.generate(ctx, parts, options). So
// the budget arrives by construction. But "it shares the path" is an assertion
// about code that someone can refactor on a Tuesday, and this estate's own rule
// is that asserted-not-verified IS the objection. These tests make the sharing
// load-bearing: split the paths and they fail.
// ============================================================================

func TestAnthropicVisionInheritsTheConfiguredBudgetThroughNilOptions(t *testing.T) {
	c, tr := anthropicFromConfig(t, map[string]interface{}{"max_tokens": float64(12000)})

	imgs := []ImageInput{{MediaType: "image/png", Data: []byte("not-a-real-png")}}
	if _, err := c.GenerateWithImages(context.Background(), "describe", imgs, nil); err != nil {
		t.Fatalf("GenerateWithImages: %v", err)
	}

	got, ok := bodyField(t, tr.captured, "max_tokens")
	if !ok {
		t.Fatal("vision request carries no max_tokens at all")
	}
	if int(got) != 12000 {
		t.Fatalf("vision sent max_tokens=%d, wanted the configured 12000. %d means the "+
			"image path has stopped sharing generate() and now has its own budget — the "+
			"exact sibling-left-heuristic shape 016b §9 names (bug_historian, corr 366efae9)",
			int(got), int(got))
	}
}

func TestGeminiVisionInheritsTheConfiguredBudgetThroughNilOptions(t *testing.T) {
	c, tr := geminiFromConfig(t, map[string]interface{}{"max_tokens": float64(12000)})

	imgs := []ImageInput{{MediaType: "image/png", Data: []byte("not-a-real-png")}}
	if _, err := c.GenerateWithImages(context.Background(), "describe", imgs, nil); err != nil {
		t.Fatalf("GenerateWithImages: %v", err)
	}

	got, ok := bodyField(t, tr.captured, "generationConfig", "maxOutputTokens")
	if !ok {
		t.Fatal("gemini vision request carries no generationConfig.maxOutputTokens")
	}
	if int(got) != 12000 {
		t.Fatalf("gemini vision sent maxOutputTokens=%d, wanted the configured 12000 "+
			"(non-thinking model, no reserve added)", int(got))
	}
}

// NEGATIVE CONTROL for the pair above: an unconfigured client must still send the
// unchanged fallback on the vision path too, or the two tests prove only that
// SOMETHING is being sent.
func TestVisionUnconfiguredStillSendsTheFallback(t *testing.T) {
	imgs := []ImageInput{{MediaType: "image/png", Data: []byte("x")}}

	ac, atr := anthropicFromConfig(t, map[string]interface{}{})
	if _, err := ac.GenerateWithImages(context.Background(), "d", imgs, nil); err != nil {
		t.Fatalf("anthropic GenerateWithImages: %v", err)
	}
	if got, _ := bodyField(t, atr.captured, "max_tokens"); int(got) != DefaultMaxOutputTokens {
		t.Fatalf("unconfigured anthropic vision sent %d, wanted %d", int(got), DefaultMaxOutputTokens)
	}

	gc, gtr := geminiFromConfig(t, map[string]interface{}{})
	if _, err := gc.GenerateWithImages(context.Background(), "d", imgs, nil); err != nil {
		t.Fatalf("gemini GenerateWithImages: %v", err)
	}
	if got, _ := bodyField(t, gtr.captured, "generationConfig", "maxOutputTokens"); int(got) != DefaultMaxOutputTokens {
		t.Fatalf("unconfigured gemini vision sent %d, wanted %d", int(got), DefaultMaxOutputTokens)
	}
}

// ============================================================================
// CANDIDATE 4, at the layer that can't be fooled by a comment.
//
// A source scan for "who calls GenerateText" was the filing's suggestion, but
// this estate has been bitten twice by source-scanning detectors (OPP-003 read
// zero files and printed a clean result; a clean result and an unrun check are
// byte-identical output). So this binds the CONTRACT behaviourally, and uses a
// source scan only to notice a provider that has no behavioural coverage.
// ============================================================================

// ⚠ SCOPE OF THIS GUARD, recorded after the council's bug_historian seat flagged
// it (advisory low, corr 366efae9 round 2): this scan guards the PROVIDER-NAME
// axis only — a provider reachable from factory.go's `case` list. A client
// constructed OUTSIDE the factory would not be caught by it.
//
// Measured 2026-08-16, because the seat's concern deserves a number rather than
// a reassurance. Five production sites bypass the factory and call a concrete
// constructor directly:
//
//	platform/orchestration/actions/rag_actions.go:374   NewOllamaClient   (embeddings)
//	internal/agents/reasoning/agent.go:71               NewAnthropicClient
//	internal/tools-api/handlers/defend.go:83            NewAnthropicClient
//	internal/tools-api/handlers/gripper.go:63           NewAnthropicClient
//	internal/tools-api/handlers/position.go:77          NewAnthropicClient
//
// All five are covered anyway, and NOT by this scan: the per-provider tests above
// call NewAnthropicClient / NewGeminiClient / NewOllamaClient DIRECTLY rather than
// through the factory, so the constructor axis is already bound behaviourally.
// What remains genuinely unguarded is a FUTURE provider whose constructor exists
// but is never added to factory.go's switch — reachable only by a direct call,
// invisible to this scan, and with no per-provider test because nobody wrote one.
// That is the residual; it is small, and it is stated rather than papered over.
func TestEveryImplementedProviderResolvesItsBudgetFromConfig(t *testing.T) {
	t.Setenv("MAXTOK_TEST_KEY", "test-key")

	// Providers whose budget contract is asserted behaviourally above.
	covered := map[string]bool{"anthropic": true, "gemini": true, "ollama": true}

	src, err := os.ReadFile("factory.go")
	if err != nil {
		t.Fatalf("read factory.go: %v", err)
	}
	caseRe := regexp.MustCompile(`(?m)^\s*case\s+"([a-z0-9_-]+)":`)
	matches := caseRe.FindAllStringSubmatch(string(src), -1)
	if len(matches) == 0 {
		// A scan that finds nothing passes for the wrong reason.
		t.Fatal("factory.go has no `case \"provider\":` lines — this test is no longer " +
			"watching anything; find the provider switch and repoint it")
	}

	for _, m := range matches {
		provider := m[1]
		if covered[provider] {
			continue
		}
		// Not covered: it is only acceptable if the factory refuses to build it
		// (an unimplemented stub, e.g. openai). A provider that CONSTRUCTS but
		// has no budget-contract test is the gap this test exists to catch.
		cfg := map[string]interface{}{
			"provider":        provider,
			"model":           "test-model",
			"api_key_env_var": "MAXTOK_TEST_KEY",
			"api_url":         "http://test.invalid",
		}
		if _, err := NewClient(context.Background(), cfg); err == nil {
			t.Errorf("factory.go implements provider %q but no test in this file asserts that "+
				"ai_service.max_tokens reaches its wire request. Add it to `covered` AND write "+
				"the assertion — bugs_open/257 is that a budget silently not arriving looks "+
				"exactly like one that did.", provider)
		}
	}
}

// Guards the guard: if DefaultMaxOutputTokens is ever changed, that is a
// fleet-wide floor change for every unconfigured call on two providers, and it
// should be a deliberate act with a reason, not a drive-by edit.
func TestDefaultMaxOutputTokensIsTheDocumentedFleetFloor(t *testing.T) {
	if DefaultMaxOutputTokens != 2048 {
		t.Fatalf("DefaultMaxOutputTokens is %d, not 2048. That raises or lowers the floor for "+
			"EVERY unconfigured call on anthropic and gemini. If deliberate, update this test, "+
			"bugs_open/205, bugs_open/257 and the LANDMINES entry, which all quote 2048 as the "+
			"fleet's smallest number.", DefaultMaxOutputTokens)
	}
	// The constant must actually be the one the clients use — not a lookalike
	// sitting unread beside a surviving literal.
	if strings.Contains(mustRead(t, "anthropic.go"), `"max_tokens": 2048`) {
		t.Fatal("anthropic.go still hardcodes 2048 in the request body; the constant is decorative")
	}
	if strings.Contains(mustRead(t, "gemini.go"), "visibleBudget := 2048") {
		t.Fatal("gemini.go still hardcodes 2048 as the visible budget; the constant is decorative")
	}
}

func mustRead(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}
