// FILE: platform/orchestration/actions/llm_options.go
//
// The options map an action must build if it calls an AI client DIRECTLY.
//
// WHY THIS EXISTS
// `ai_service` config looks like it configures the call. For most steps it does,
// because `ExecuteLLMPromptAction` reads it and builds the options map that the
// provider clients actually consult (`ai_actions.go`, the budget ladder). An action that
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
// `ai_actions.go` still holds its own copy of this logic and is NOT
// changed here. That path serves 127 live steps across 55 agents, and rewriting
// a shared hot path to fix two actions is the wrong trade under a live outage.
// This is therefore the SECOND copy of the rule, knowingly. Per this estate's
// own doctrine (LANDMINES, "two vocabularies, one algorithm"): **a THIRD caller
// should be the extraction, not another paste** — at that point make
// `ExecuteLLMPromptAction` call this and delete its inline block.
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
// > helper too. Every model call in this package outside `ExecuteLLMPromptAction`
// > resolves its budget HERE, and `llm_budget_call_sites_test.go` fails the build
// > if a sixth spelling appears. The extraction the note above asked for is
// > therefore done for the direct callers; what remains is `ai_actions.go`'s own
// > inline block, which is bugs_open/257 candidate 2 and blocked on an import
// > cycle, not on willingness.

package actions

import (
	"encoding/json"

	"go.uber.org/zap"
)

// llmOptionsFromConfig builds the options map the provider clients actually read.
//
// > **CORRECTED 2026-08-10, same day, by the council's `llm_reliability` seat —
// > this comment and the submission that shipped it BOTH claimed the precedence
// > "mirrors ExecuteLLMPromptAction", and it does not.**
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

// ─────────────────────────────────────────────────────────────────────────────
// THE LADDER: WHERE A REQUEST-SIZING KEY MAY BE WRITTEN, AND WHICH ONE WINS
//
// Added 2026-09-04 for `bugs_open/257` round 3 (owner decision 4: "the limits
// are set in each individual's config, sometimes it has been set in the wrong
// place and sometimes the agent reads the wrong place, please fix it properly").
//
// An operator can write `max_tokens` in four legitimate-looking places, and
// until this ladder existed the canonical reader looked at TWO of them, in an
// order that put the LEAST specific first. Measured against live
// `agent_definitions` on 2026-09-04 (208 active, non-snapshot):
//
//	.workflow.steps.<step>.config.ai_service.max_tokens   149 decls, 74 agents
//	.max_tokens                          (agent, top)      10 decls,  9 agents
//	.workflow.steps.<step>.config.max_tokens               7 decls,  2 agents
//	.ai_service.max_tokens               (agent, service)   3 decls,  3 agents
//	…<nested loop step>.config.ai_service.max_tokens        2 decls,  1 agent
//
// and the two failures that census exposes are BOTH silent:
//
//  1. SHADOWED. `ExecuteLLMPromptAction` read the AGENT's top-level key first,
//     so an agent carrying one capped every step at it. All ten such agents
//     declare 8000 on their single LLM step and were pinned to 500–2000 by a
//     leftover root key. Nothing logged it; the config reads correctly.
//  2. UNREAD. Nothing looked at a STEP's top-level key at all, so
//     `site-adoption-agent`'s four steps asked for 32000/6000/4000/4000 and every
//     one of them ran at the root 16000. The key sits one brace outside the
//     `ai_service` block it was meant for, next to a `model` and a `provider`
//     that ARE read — which is exactly why it looks right.
//
// The rule is now: MOST SPECIFIC LEVEL WINS, and at each level the canonical
// `ai_service` spelling is consulted before the bare one. Both orderings are
// unobservable on today's fleet — no agent declares both spellings at one level,
// checked 2026-09-04 — so this fixes the two failures above without silently
// re-deciding anything an operator has already written.
//
// A bare key is READ rather than refused, deliberately. Refusing it would turn
// today's seventeen misplaced declarations into seventeen unconfigured steps at
// the 2048 floor, which is the failure `bugs_open/205` counted. The push towards
// one canonical spelling belongs in a report an operator can act on
// (`config-key-audit --budget-placement`), not in a reader that goes quiet.

// budgetLevel is one place a request-sizing key may be written, paired with the
// name that goes in the log line. `where` is not decoration: "which place did
// this number come from" was previously answerable only by re-deriving the
// precedence rule from source, which is how the two failures above survived.
type budgetLevel struct {
	cfg   map[string]interface{}
	where string
}

// subConfigMap returns a nested config block, or nil. nil is a legal level — the
// walker skips it — so a missing `ai_service` block needs no caller-side branch.
func subConfigMap(m map[string]interface{}, key string) map[string]interface{} {
	if m == nil {
		return nil
	}
	sub, _ := m[key].(map[string]interface{})
	return sub
}

// budgetLevelsForStep is the canonical ladder used by `ExecuteLLMPromptAction`,
// most specific first. It mirrors the order `resolveAIServiceConfig` overlays the
// `ai_service` blocks in (runtime step over workflow step over root) and
// interleaves the bare spelling at each level.
//
// `workflowStepCfg` is nil for a NESTED step: `currentStep` for a loop body is a
// synthesised name ("process_sections_loop_iter_0_rewrite_negations") that does
// not appear in `workflow.steps`, so the nested step's config arrives instead as
// the runtime `StepConfig`, which is the level above it here. That is existing
// behaviour, not a new limitation — it is written down because a reader checking
// why a nested budget resolves will otherwise look for a level that cannot fire.
func budgetLevelsForStep(agentCfg, workflowStepCfg, runtimeStepCfg map[string]interface{}) []budgetLevel {
	return []budgetLevel{
		{subConfigMap(runtimeStepCfg, "ai_service"), "step_config.ai_service"},
		{runtimeStepCfg, "step_config"},
		{subConfigMap(workflowStepCfg, "ai_service"), "workflow_step.ai_service"},
		{workflowStepCfg, "workflow_step"},
		{subConfigMap(agentCfg, "ai_service"), "root.ai_service"},
		{agentCfg, "root"},
	}
}

// resolveBudgetKey walks the levels in order and returns the first usable
// positive number, the level it came from, and every OTHER level that declares
// the key — the shadowed ones. The third return exists so a caller can say out
// loud that a declaration was overridden: an operator who writes a number and
// gets a different one has no other way to find out.
func resolveBudgetKey(key string, levels []budgetLevel) (value int, from string, shadowed []string) {
	for _, level := range levels {
		v, ok := numericConfigValue(level.cfg, key)
		if !ok {
			continue
		}
		if from == "" {
			value, from = v, level.where
			continue
		}
		shadowed = append(shadowed, level.where)
	}
	return value, from, shadowed
}

// numericConfigValue reads one positive number out of a config map.
//
// It accepts int, int64, float64 and json.Number because THE SAME KEY ARRIVES AS
// DIFFERENT GO TYPES DEPENDING ON THE PATH: `agent_definitions.default_config` is
// jsonb and decodes to float64, `configs/*.yaml` is read by viper and decodes to
// int. The reader this replaces in `ai_actions.go` was float64-ONLY, so it
// silently dropped the YAML shape — the same hole `platform/aiservice`'s
// `configMaxTokens` was written to close, and the two now agree on what counts as
// configured.
//
// Zero is treated as unset: `max_tokens: 0` is a hard 400 from Anthropic, not a
// request for no limit. Negative is meaningless.
func numericConfigValue(m map[string]interface{}, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	switch v := m[key].(type) {
	case int:
		if v > 0 {
			return v, true
		}
	case int64:
		if v > 0 {
			return int(v), true
		}
	case float64:
		if v > 0 {
			return int(v), true
		}
	case json.Number:
		if n, err := v.Int64(); err == nil && n > 0 {
			return int(n), true
		}
	}
	return 0, false
}

// intFromConfig reads a numeric key from the step config, falling back to the
// ai_service block. Returns false when neither holds a usable positive number.
//
// The two-level ladder a DIRECT caller walks, and it is deliberately NOT the
// six-level one above: a direct caller is handed the step's config and an
// already-merged `ai_service` map, and has no agent-level config to consult.
// Keeping the two lists separate rather than merging the functions is
// `bugs_open/257` candidate 2, ruled by the owner on 2026-09-04 as option (c) —
// the two rules differ permanently, and this comment is the contract that says
// so. What they now share is the WALKER, so a third divergence in how a number
// is read out of a map cannot happen; what they do not share is which levels
// exist, which is a real difference and not an accident.
func intFromConfig(stepCfg, aiCfg map[string]interface{}, key string) (int, bool) {
	value, from, _ := resolveBudgetKey(key, []budgetLevel{
		{stepCfg, "step"},
		{aiCfg, "ai_service"},
	})
	return value, from != ""
}
