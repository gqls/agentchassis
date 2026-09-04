// FILE: platform/aiservice/thinking_budget_test.go
//
// Keeps the thinking-budget verdicts in step with the alias table beside them.
// A new model id added to ModelAliases and not classified here defaults to
// "accepts", which is the UNSAFE direction: it would let a `budget_tokens`
// declaration through to a model that answers with a 400 on every call.
package aiservice

import "testing"

func TestEveryAliasHasAThinkingBudgetVerdict(t *testing.T) {
	for _, alias := range KnownModelAliases() {
		resolved := ResolveModelAlias(alias, nil)
		rejects := modelsRejectingThinkingBudget[resolved]
		accepts := modelsAcceptingThinkingBudget[resolved]
		switch {
		case rejects && accepts:
			t.Errorf("%s (-> %s) is in BOTH verdict maps", alias, resolved)
		case !rejects && !accepts:
			t.Errorf("%s (-> %s) has no thinking-budget verdict.\n"+
				"\tAdd it to modelsRejectingThinkingBudget (400 on `thinking.budget_tokens`) or to\n"+
				"\tmodelsAcceptingThinkingBudget. Defaulting is the unsafe direction: an unclassified\n"+
				"\tmodel is treated as accepting, and the failure is a 400 on EVERY call for that agent.",
				alias, resolved)
		}
	}
	if len(KnownModelAliases()) == 0 {
		t.Fatal("ModelAliases is empty — this parity test is watching nothing")
	}
}

func TestAcceptsThinkingBudget(t *testing.T) {
	// The three live answers on this fleet, measured 2026-09-04.
	for model, want := range map[string]bool{
		"claude-sonnet-5":            false, // 87 declarations / 21 agents
		"claude-opus-4-8":            false, // 1 / 1
		"claude-opus-5":              false,
		"claude-fable-5-1":           false,
		"claude-sonnet-4-6":          true, // 47 / 33 — deprecated, still functional
		"claude-opus-4-6":            true, // 6 / 3
		"claude-haiku-4-5":           true, // 32 / 24 — REQUIRED for thinking
		"claude-haiku-4-5-20251001":  true, // the resolved form answers identically
		"claude-3-5-sonnet-20241022": true,
	} {
		if got := AcceptsThinkingBudget(model); got != want {
			t.Errorf("AcceptsThinkingBudget(%q) = %v, want %v", model, got, want)
		}
	}

	// An unknown model must NOT be reported as rejecting: this predicate is never
	// allowed to be the reason a working call stops being made.
	if !AcceptsThinkingBudget("some-model-released-after-this-file") {
		t.Error("an unrecognised model was reported as rejecting a thinking budget; " +
			"the default must be permissive, and the audit reports what it does not recognise")
	}
}
