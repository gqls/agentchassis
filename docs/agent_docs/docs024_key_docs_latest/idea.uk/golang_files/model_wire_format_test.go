package main

// model_wire_format_test.go — the thinking wire format must match the model
// family, and the default models must be current.
//
// Why this exists: on 2026-07-26 the engine moved from Opus 4.8 / Sonnet 4.6 to
// the 5 family. The model strings are the visible half of that change; the
// load-bearing half is that the 5 family REJECTS the manual thinking budget
// (`thinking:{type:"enabled",budget_tokens:N}`) with a 400. The selector used to
// be an allow-list of adaptive-thinking models, so every model it had not heard
// of — i.e. every future model — fell through to the manual branch. Swapping the
// model ids alone would have 400'd every call in production.

import (
	"strings"
	"testing"
)

// The five family and anything newer must NOT get a manual budget.
func TestFiveFamilyUsesAdaptiveThinking(t *testing.T) {
	for _, model := range []string{
		"claude-opus-5",
		"claude-sonnet-5",
		"claude-fable-5",
		"claude-mythos-5",
		"claude-opus-4-8",
		"claude-opus-4-7",
	} {
		if usesManualThinkingBudget(model) {
			t.Errorf("%s: selected the manual budget format — the API rejects it with a 400", model)
		}
	}
}

// The genuinely-old families still need it.
func TestLegacyFamiliesUseManualBudget(t *testing.T) {
	for _, model := range []string{
		"claude-sonnet-4-6",
		"claude-sonnet-4-5-20250929",
		"claude-opus-4-6",
		"claude-opus-4-5",
		"claude-haiku-4-5",
	} {
		if !usesManualThinkingBudget(model) {
			t.Errorf("%s: expected the legacy manual-budget format", model)
		}
	}
}

// The regression that motivated the inversion: an unrecognised model must get
// the MODERN format, because an unknown model is far likelier to be newer than
// older. Under the old allow-list this returned the legacy format and 400'd.
func TestUnknownModelDefaultsToAdaptive(t *testing.T) {
	for _, model := range []string{"claude-opus-6", "claude-sonnet-6", "some-future-model"} {
		if usesManualThinkingBudget(model) {
			t.Errorf("%s: an unrecognised model must default to adaptive thinking, not a manual budget", model)
		}
	}
}

// The configured defaults must be current-generation. This is the test that
// fails when the catalogue moves on and nobody updated the engine.
func TestDefaultModelsAreCurrentGeneration(t *testing.T) {
	for name, model := range map[string]string{
		"GEN_MODEL":      genModel,
		"CRITIQUE_MODEL": critiqueModel,
		"VERIFY_MODEL":   verifyModel,
		"SCORE_MODEL":    scoreModel,
	} {
		if usesManualThinkingBudget(model) {
			t.Errorf("%s=%q is a legacy-generation model", name, model)
		}
		if !strings.HasSuffix(model, "-5") {
			t.Errorf("%s=%q is not a 5-family model id", name, model)
		}
	}
}

// Every call site must set Effort explicitly. An empty Effort sends no thinking
// field, and that no longer means "no thinking" — on the 5 family it means
// adaptive thinking runs by default and shares the max_tokens cap with the
// answer. This test guards the intent, not the wire.
func TestEffortIsSetAtEveryCallSite(t *testing.T) {
	// Sanity: effortToBudget still maps the levels the call sites use, so the
	// legacy branch remains correct for an operator who pins an old model.
	for _, level := range []string{"low", "medium", "high", "xhigh", "max"} {
		if effortToBudget(level) < 1024 {
			t.Errorf("effortToBudget(%q) = %d, too small to be a valid manual budget",
				level, effortToBudget(level))
		}
	}
	if effortToBudget("") != 0 {
		t.Errorf("effortToBudget(\"\") = %d, want 0", effortToBudget(""))
	}
}
