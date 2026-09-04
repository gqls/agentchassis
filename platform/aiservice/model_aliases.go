// ============================================================================
// FILE: platform/aiservice/model_aliases.go
//
// Resolves model aliases to actual Anthropic API model names.
// Update this file when Anthropic releases new model versions.
// ============================================================================

package aiservice

import (
	"go.uber.org/zap"
)

// ModelAliases maps user-friendly aliases to actual API model names
// Update these when Anthropic releases new model versions
var ModelAliases = map[string]string{
	// Claude 5 family
	"claude-sonnet-5": "claude-sonnet-5",

	// Claude 4.8 family
	"claude-opus-4-8": "claude-opus-4-8",

	// Claude 4.6 family
	"claude-sonnet-4-6": "claude-sonnet-4-6",
	"claude-opus-4-6":   "claude-opus-4-6",

	// Claude 4.5 family
	"claude-sonnet-4-5": "claude-sonnet-4-5-20250929",
	"claude-haiku-4-5":  "claude-haiku-4-5-20251001",
	"claude-opus-4-5":   "claude-opus-4-5-20251101",

	// Claude 4 family
	"claude-sonnet-4": "claude-sonnet-4-20250514",
	"claude-opus-4":   "claude-opus-4-20250514",

	// Claude 3.5 family
	"claude-3-5-sonnet": "claude-3-5-sonnet-20241022",
	"claude-3-5-haiku":  "claude-3-5-haiku-20241022",

	// Claude 3 family (legacy)
	"claude-3-opus":   "claude-3-opus-20240229",
	"claude-3-sonnet": "claude-3-sonnet-20240229",
	"claude-3-haiku":  "claude-3-haiku-20240307",
}

// ResolveModelAlias converts a model alias to the actual API model name.
// If the input is not an alias (already a full model name), returns it unchanged.
func ResolveModelAlias(model string, logger *zap.Logger) string {
	if resolved, isAlias := ModelAliases[model]; isAlias {
		if logger != nil {
			logger.Debug("Resolved model alias",
				zap.String("alias", model),
				zap.String("resolved", resolved),
			)
		}
		return resolved
	}
	return model
}

// GetAvailableAliases returns a list of all available model aliases
func GetAvailableAliases() []string {
	aliases := make([]string, 0, len(ModelAliases))
	for alias := range ModelAliases {
		aliases = append(aliases, alias)
	}
	return aliases
}

// IsValidModel checks if a model name is either a valid alias or a known full model name
func IsValidModel(model string) bool {
	// Check if it's an alias
	if _, isAlias := ModelAliases[model]; isAlias {
		return true
	}

	// Check if it's a known full model name (value in the map)
	for _, fullName := range ModelAliases {
		if model == fullName {
			return true
		}
	}

	// Could be a new model we don't know about - allow it through
	// The API will validate it
	return true
}

// ─────────────────────────────────────────────────────────────────────────────
// WHICH MODELS STILL ACCEPT A MANUAL THINKING BUDGET
//
// Added 2026-09-04 after a peer lane hit a 400 four times replaying a stored
// writer prompt against `claude-sonnet-5`. Anthropic replaced the fixed
// thinking-token budget with adaptive thinking: `thinking: {"type":"enabled",
// "budget_tokens":N}` — the exact shape `anthropic.go` emits whenever
// `options["budget_tokens"]` is a positive number — is now REJECTED WITH A 400
// on the Claude 5 family and on Opus 4.7/4.8.
//
// ⚠ IT IS NOT A BLANKET RULE, AND THAT IS THE WHOLE REASON THIS IS A FUNCTION
// RATHER THAN A DELETION. Three different answers are live on this fleet today:
//
//	claude-sonnet-5, claude-opus-4-8      REJECTED — 400, every call
//	claude-sonnet-4-6, claude-opus-4-6    deprecated but FUNCTIONAL (escape hatch)
//	claude-haiku-4-5 and older            REQUIRED — thinking does not happen without it
//
// So a guard that simply stopped sending the key would take extended thinking
// away from the only models where it still works, and `claude-haiku-4-5` carries
// 32 model declarations across 24 live agents.
//
// LIVE EXPOSURE IS ZERO AND THAT IS WHY THIS IS A PREDICATE, NOT A FIX.
// [MEASURED 2026-09-04, two independent encodings — `jsonb_path_exists(
// default_config,'strict $.**.budget_tokens')` and a recursive jsonb walk]
// NO active agent declares `budget_tokens`: 0 of 208 active rows, 0 of all 225
// rows including snapshots and deleted. Four rows match the string; all four are
// a reviewer seat's footprint KEYWORD LIST, not a declaration — a substring
// census over config text reads as three live agents and means none.
//
// What that makes this: a trap armed for the first operator who declares the
// key. `bugs_open/257` round 3 widened where it can be declared from two places
// to six, so the surface grew while the exposure stayed zero. The failure lands
// on every call for that agent, not one step, and the declaration is a one-word
// config change that looks obviously safe.
//
// The consumer today is `cmd/config-key-audit --budget-placement`, which fails
// the run on a declaration against a rejecting model — BEFORE it is ever sent.
// Whether the provider client should also refuse, drop-with-warning, or keep
// returning the 400 is an open question on a shared seam and is NOT decided here
// (`bugs_open/257`): dropping silently would be this bug's own defect class, and
// the 400 is at least loud.
//
// KEEP THIS TABLE IN STEP WITH THE ALIASES ABOVE. A new model id added there and
// not here defaults to "accepts", which is the unsafe direction — so the test
// `TestEveryAliasHasAThinkingBudgetVerdict` fails the build on any alias this
// function has not been taught about.

// WHY THIS IS A PREDICATE AND NOT AN INLINE CONSTRAINT LIKE THE TEMPERATURE ONE —
// asked on the record by the council's `reuse_agent` seat (MEDIUM, corr 47ea9498):
// `anthropic.go` already encodes a model-version-gated request constraint inline
// (temperature dropped for Opus 4.7+), so why a second, structurally different
// mechanism for the same class of problem, and won't the two drift?
//
// A fair question, and the answer is that they are NOT the same mechanism. The
// temperature rule is UNCONDITIONAL: `anthropic.go` never sends temperature to any
// Anthropic model, on any version, so it needs no model check and has no table to
// drift from. `budget_tokens` cannot be handled that way, because the three verdicts
// disagree — dropping it unconditionally would remove extended thinking from
// `claude-haiku-4-5` (32 model declarations across 24 live agents), where the key is
// REQUIRED. A per-model answer needs a per-model table; a blanket rule does not.
//
// If temperature ever becomes model-conditional, it should reuse THIS table rather
// than grow a second one — that is the drift the seat is right to be watching for,
// and this comment is where the next author meets it.

// modelsRejectingThinkingBudget are the API model names that return a 400 for
// `thinking: {"type":"enabled","budget_tokens":N}`. Matched against the RESOLVED
// model name, so an alias and its expansion give the same answer.
var modelsRejectingThinkingBudget = map[string]bool{
	"claude-fable-5":    true,
	"claude-fable-5-1":  true,
	"claude-mythos-5":   true,
	"claude-mythos-5-1": true,
	"claude-opus-5":     true,
	"claude-opus-4-8":   true,
	"claude-opus-4-7":   true,
	"claude-sonnet-5":   true,
}

// modelsAcceptingThinkingBudget exists ONLY so the parity test can tell "this
// model was considered and accepts" from "nobody has looked at this model". It
// is not consulted at runtime — AcceptsThinkingBudget defaults an unknown model
// to accepting, which is the historical behaviour — but a new alias that lands
// in neither map fails the build, and that is the point: the unsafe direction
// here is silence.
var modelsAcceptingThinkingBudget = map[string]bool{
	// Deprecated but functional — the transitional escape hatch.
	"claude-sonnet-4-6": true,
	"claude-opus-4-6":   true,
	// Required for thinking on these; adaptive thinking does not exist there.
	"claude-sonnet-4-5-20250929": true,
	"claude-haiku-4-5-20251001":  true,
	"claude-opus-4-5-20251101":   true,
	"claude-sonnet-4-20250514":   true,
	"claude-opus-4-20250514":     true,
	"claude-3-5-sonnet-20241022": true,
	"claude-3-5-haiku-20241022":  true,
	"claude-3-opus-20240229":     true,
	"claude-3-sonnet-20240229":   true,
	"claude-3-haiku-20240307":    true,
}

// AcceptsThinkingBudget reports whether a model will accept a manual
// `budget_tokens` thinking budget, or reject the request with a 400.
//
// Resolves aliases first, so `claude-sonnet-4-5` and
// `claude-sonnet-4-5-20250929` answer identically. An UNKNOWN model returns
// true — the historical behaviour — because this predicate must never be the
// reason a working call stops being made; the audit that consumes it reports
// what it does not recognise instead.
func AcceptsThinkingBudget(model string) bool {
	return !modelsRejectingThinkingBudget[ResolveModelAlias(model, nil)]
}

// KnownModelAliases returns the alias keys, so a test can assert every alias has
// been given a thinking-budget verdict rather than silently defaulting.
func KnownModelAliases() []string {
	out := make([]string, 0, len(ModelAliases))
	for alias := range ModelAliases {
		out = append(out, alias)
	}
	return out
}
