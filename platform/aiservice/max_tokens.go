// FILE: platform/aiservice/max_tokens.go
//
// WHERE THE OUTPUT-TOKEN BUDGET IS RESOLVED (bugs_open/257)
//
// Until this file existed, the budget was resolved exactly ONE layer up, in
// `ExecuteLLMPromptAction` (`platform/orchestration/actions/ai_actions.go:358-364`),
// which translated `ai_service.max_tokens` into the `options` map that the
// provider clients actually consult. The contract was therefore "call through
// ExecuteLLMPromptAction, or your configuration is inert" — and nothing stated it,
// nothing checked it, and the compiler was happy either way, because
// `GenerateText` accepts a nil options map.
//
// That made a configured budget silently unreachable for any caller that built
// a client itself. Measured 2026-08-10 (WRONG_CALLS.md, provocation lane): a
// step was given `max_tokens: 8000` by migration 372 and the very next run still
// died at `output_tokens=2048`, so the config was changed a SECOND time against
// a key nothing read. A config key nothing reads looks exactly like one that
// works, and the error echoes the old number back at you.
//
// The fix is to resolve the budget where the request is built, from the config
// the client was ALREADY constructed with. Every provider constructor receives
// the whole `ai_service` map and already reads `model` out of it this way; this
// is one more key from a map already in hand, not new plumbing.
//
// PRECEDENCE AT THE WIRE, after this change:
//
//  1. options["max_tokens"]  — per-call, wins. UNCHANGED.
//  2. ai_service.max_tokens  — from the construction config. NEW.
//  3. the provider fallback  — DefaultMaxOutputTokens for anthropic/gemini;
//     ollama sends no num_predict at all. UNCHANGED.
//
// OPT-IN, WITH THE UNSAFE SIDE AS THE DEFAULT (owner ruling 2026-08-02 §2,
// RFC_010; the shape bugs_open/244 shipped into this same client). Rule 2 fires
// only when an operator has explicitly put `max_tokens` in the `ai_service`
// block a client is built from. A caller whose config does not name it runs the
// identical code path as before, byte for byte. The opt-in signal is the config
// key itself — deliberately NOT a second key, because the whole defect is that
// `ai_service.max_tokens` did not mean what it said, and curing that with a
// different spelling would leave two keys for one budget.
package aiservice

import "encoding/json"

// DefaultMaxOutputTokens is the transport-level floor applied when nobody chose
// a budget: not a sizing decision, a safety net. It is the smallest number in
// the estate, and a call reaching it means no operator sized this step — the
// first oversized reply then meets a silent cliff (bugs_open/205: 8 of 126
// active LLM steps ran unconfigured; 64 truncations before anything said so).
//
// Changing this constant changes the floor for every unconfigured call on
// anthropic AND gemini. Raise a specific step's `ai_service.max_tokens`
// instead; that is now a live lever on every path, which is the point of 257.
const DefaultMaxOutputTokens = 2048

// WHY THIS IS A NEW HELPER AND NOT A REUSE — checked, not assumed, after the
// council's `reuse_agent` seat objected (HIGH, corr 366efae9) that this estate
// already has numeric-config helpers and the submission never said whether they
// were considered. It was right that I had not checked. Having checked:
//
//   - `datahelpers.GetIntField(m, key, default) int` handles float64 and int
//     only — no int64, no json.Number — and returns a bare int with a default.
//   - `datahelpers.ExtractNestedFieldInt(data, path) int` handles float64, int
//     and int64, returns 0 for missing.
//
// Both collapse "the operator chose nothing" and "the operator chose 0" into the
// same int. That distinction is the whole point of the (int, bool) pair below:
// it is what lets anthropic/gemini apply a floor while OLLAMA OMITS THE FIELD
// ENTIRELY — two opposite policies over one answer, which a helper returning
// `int` cannot express. Neither helper rejects a negative or a zero either, and
// `max_tokens: 0` is a hard 400 from Anthropic.
//
// AND THE LAYERING IS DECISIVE, independently of the above. The seat's second
// objection was that this is a THIRD resolver alongside `ai_actions.go:358-364`
// and `llmOptionsFromConfig`. Both of those live in `package actions`, and
// `platform/orchestration/actions` IMPORTS `platform/aiservice`
// (`ai_actions.go:13`). So `aiservice` importing either is a genuine import
// CYCLE — Go refuses to compile it. Sharing the existing resolver in that
// direction is not a design preference that was declined; it is not available.
// `datahelpers` would not cycle, but `aiservice` today has EXACTLY ZERO internal
// dependencies (`go list -deps ./platform/aiservice` returns only itself), and
// giving the estate's lowest-level provider package its first dependency on the
// orchestration layer is a bigger change than the one this file is making.
//
// The honest residual: there are now three places that read this key. What this
// change removes is the CONSEQUENCE of them disagreeing (a silent 2048), not the
// duplication itself. Unifying them is bugs_open/257 candidate 2, still open,
// with its three behavioural differences named in the bug file.

// configMaxTokens reads `max_tokens` from an `ai_service` construction config.
//
// It accepts int, int64, float64 and json.Number because THE SAME KEY ARRIVES
// AS DIFFERENT GO TYPES DEPENDING ON THE PATH, and a client cannot know which:
// `agent_definitions.default_config` is jsonb, so it decodes to **float64**;
// `configs/*.yaml` is read by viper, so it decodes to **int**. A float64-only
// read — which is what `ai_actions.go:361` does — silently drops the YAML case,
// and that is precisely the config `reasoning-agent` carries.
//
// A value must be > 0. Zero is not "no budget", it is a 400 from Anthropic, so
// it is treated as unset and the fallback applies. Negative is meaningless.
// (This matches `llmOptionsFromConfig`'s `intFromConfig`, deliberately: two
// readers of one key that disagree about what counts as configured is the
// drift class bugs_open/257 §3 is about.)
//
// Returns (0, false) when the key is absent, unusable, or non-positive, so the
// caller can distinguish "operator chose nothing" from "operator chose a
// number" — a distinction `int` alone cannot carry.
func configMaxTokens(config map[string]interface{}) (int, bool) {
	// Redundant in Go — reading a key from a nil map yields the zero value, so
	// the switch below would fall through anyway (confirmed by mutation on
	// 2026-08-16: deleting this branch changes no test, i.e. it is an equivalent
	// mutant, not an untested one). Kept because `resolvedMaxTokens(nil)` is a
	// real call shape and an explicit branch says "nil is legal input" where
	// silence would leave the next reader to derive it.
	if config == nil {
		return 0, false
	}
	switch v := config["max_tokens"].(type) {
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
		// A decoder configured with UseNumber() hands numbers over unconverted.
		// Not on any path measured today, but it costs three lines and the
		// failure it prevents is silent.
		if n, err := v.Int64(); err == nil && n > 0 {
			return int(n), true
		}
	}
	return 0, false
}

// resolvedMaxTokens is the client-construction helper: the configured budget if
// there is one, otherwise the shared fallback. Used by the providers that must
// always send a number (anthropic, gemini). Ollama does NOT use it — it omits
// the field entirely when unconfigured, which is its existing behaviour and is
// not this change's business to alter.
func resolvedMaxTokens(config map[string]interface{}) int {
	if mt, ok := configMaxTokens(config); ok {
		return mt
	}
	return DefaultMaxOutputTokens
}
