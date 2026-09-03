// FILE: platform/orchestration/actions/llm_options.go
//
// The options map an action must build if it calls an AI client DIRECTLY.
//
// WHY THIS EXISTS
// `ai_service` config looks like it configures the call. For most steps it does,
// because `ExecuteAIStepAction` reads it and builds the options map that the
// provider clients actually consult (`ai_actions.go:351-372`). An action that
// bypasses that path and calls `client.GenerateText(ctx, prompt, nil)` — or,
// worse, `map[string]interface{}{}`, which looks deliberate — silently gets the
// provider's hardcoded fallback instead, whatever the config says.
//
// For Anthropic that fallback is **2048 output tokens**, the smallest number in
// the estate.
//
// > **CORRECTED 2026-08-16 — the paragraph above was true when written and is
// > now HALF true, so it is corrected in place rather than deleted.**
// > `bugs_open/257` Path A moved budget resolution INTO the provider clients
// > (`platform/aiservice/max_tokens.go`). A direct caller passing `nil` or an
// > empty map no longer "silently gets the provider's hardcoded fallback
// > whatever the config says" — it now inherits `ai_service.max_tokens` from the
// > config the client was CONSTRUCTED with, and reaches the 2048 fallback only
// > when nobody configured a budget at all.
// >
// > **This helper is still the right thing to call**, for the two reasons the
// > client cannot cover: (a) it reads the STEP's config, which is never passed
// > to a client constructor, and (b) it forwards `budget_tokens`. What changed
// > is the CONSEQUENCE of forgetting it — a wrong-but-configured budget, not a
// > silent 2048.
// >
// > The line references in this file were also stale: the hardcoded fallback
// > moved from `anthropic.go:109` to `:185` when the prompt-caching work
// > (`bugs_open/244`) landed, and it is now `DefaultMaxOutputTokens`. Line
// > numbers in comments rot; the constant name will not.
//
// MEASURED, 2026-08-10, and this is why the file exists rather than a comment:
// `generate_provocations` passed an empty options map. Its step config was given
// `max_tokens: 8000` (migration 372) and the very next run still died at
// `output_tokens=2048`. Two more config changes were applied against a value that
// could not reach the API, because "the config says 8000" and "the request sent
// 8000" are independent facts and only the second one is the request. A config
// key that nothing reads looks exactly like a config key that works.
//
// This is the same class `bugs_open/205` counted from the other end — 8 of 126
// active LLM steps with no configured budget, 64 truncations before anything
// said so. There the budget was never set; here it was set and then dropped on
// the floor, which is harder to see, because the config is right there.
//
// SCOPE, DELIBERATELY SMALL
// `ai_actions.go:358-372` still holds its own copy of this logic and is NOT
// changed here. That path serves 127 live steps across 55 agents, and rewriting
// a shared hot path to fix two actions is the wrong trade under a live outage.
// This is therefore the SECOND copy of the rule, knowingly. Per this estate's
// own doctrine (LANDMINES, "two vocabularies, one algorithm"): **a THIRD caller
// should be the extraction, not another paste** — at that point make
// `ExecuteAIStepAction` call this and delete its inline block.
//
// > **UPDATE 2026-09-03 — the paragraph above was ignored twice, and the count
// > it warned about is now settled the other way.**
// > Two actions written AFTER this file (`rewrite_negations`, 2026-08-20;
// > `repair_ordering_register`, 2026-08-31) each pasted the rule again and each
// > ended in a hardcoded `2000`. A literal is worse than a paste: an explicitly
// > supplied option wins at the wire (`aiservice/anthropic.go:307`), so a caller
// > that always sends a number can never inherit the configured one — the
// > literal DEFEATS bugs_open/257 Path A rather than being covered by it. And in
// > `offer-analyser` the literal was numerically EQUAL to the configured value,
// > so `llm_call_log.max_tokens` read `2000` whether the config was honoured or
// > dropped, and no query over the fleet's own instrument could separate the two.
// >
// > Both pastes are deleted, and `execute_vision_prompt` and `ch_llm_review` —
// > which hand-read the key and passed an empty map respectively — now call this
// > helper too. Every model call in this package outside `ExecuteAIStepAction`
// > resolves its budget HERE, and `llm_budget_call_sites_test.go` fails the build
// > if a sixth spelling appears. The extraction the note above asked for is
// > therefore done for the direct callers; what remains is `ai_actions.go`'s own
// > inline block, which is bugs_open/257 candidate 2 and blocked on an import
// > cycle, not on willingness.

package actions

import (
	"go.uber.org/zap"
)

// llmOptionsFromConfig builds the options map the provider clients actually read.
//
// > **CORRECTED 2026-08-10, same day, by the council's `llm_reliability` seat —
// > this comment and the submission that shipped it BOTH claimed the precedence
// > "mirrors ExecuteAIStepAction", and it does not.**
// > That function's outer key is `agentConfig`, which is
// > `CollectedData["agent_config"]` falling back to `agentDef.DefaultConfig`
// > (`ai_actions.go:180,219`) — the AGENT's whole default config, not the step's.
// > So its rule is *agent-level beats ai_service*; this one is *step-level beats
// > ai_service*. Different levels, and I asserted the equivalence from the shape
// > of the code rather than from reading what the variable held. The seat
// > objected on exactly that ("asserted, not verified") and was right.
// >
// > The behaviour here is still the one this call site wants, and the ai_service
// > fallback — the arm that actually carries the live config — is identical in
// > both. What was wrong was the claim, not the code. Recorded rather than
// > quietly edited, because "mirrors X" is the kind of sentence a later reader
// > relies on without re-deriving.
//
// Precedence: the step's own config wins over the `ai_service` block, so a step
// can raise its budget without editing a shared service definition. Values arrive
// from jsonb as float64; int is accepted too so a Go caller constructing config by
// hand behaves the same way.
//
// `where` names the call site in the warning, because the failure this guards
// against is diagnosed from logs by someone who does not yet know which step is
// truncating.
func llmOptionsFromConfig(stepCfg, aiCfg map[string]interface{}, logger *zap.Logger, where string) map[string]interface{} {
	opts := make(map[string]interface{})

	if mt, ok := intFromConfig(stepCfg, aiCfg, "max_tokens"); ok {
		opts["max_tokens"] = mt
	} else if logger != nil {
		// Not an error: the call will work, at 2048. Worth saying out loud
		// because the first oversized reply meets a cliff with no other warning,
		// and the step config will look correct when someone reads it.
		logger.Warn("no max_tokens configured at step or ai_service level; "+
			"the provider's hardcoded fallback applies (anthropic: 2048)",
			zap.String("call_site", where))
	}

	// Extended thinking. Only forwarded when present — the client enables
	// thinking on the presence of this key, so a zero or absent value must not
	// become an explicit request for it.
	if bt, ok := intFromConfig(stepCfg, aiCfg, "budget_tokens"); ok && bt > 0 {
		opts["budget_tokens"] = bt
	}

	return opts
}

// intFromConfig reads a numeric key from the step config, falling back to the
// ai_service block. Returns false when neither holds a usable positive number.
func intFromConfig(stepCfg, aiCfg map[string]interface{}, key string) (int, bool) {
	for _, m := range []map[string]interface{}{stepCfg, aiCfg} {
		if m == nil {
			continue
		}
		switch v := m[key].(type) {
		case float64:
			if v > 0 {
				return int(v), true
			}
		case int:
			if v > 0 {
				return v, true
			}
		}
	}
	return 0, false
}
