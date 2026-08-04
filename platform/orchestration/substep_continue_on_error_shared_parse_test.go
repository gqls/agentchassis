// FILE: platform/orchestration/substep_continue_on_error_shared_parse_test.go
//
// bugs_open/193 — the substep-level read of continue_on_error and the loop-level
// read of the same key were TWO implementations, and they diverged: bugs_open/173
// made this one loud and left the loop-level twin silent.
//
// The five tests in substep_continue_on_error_test.go pin every VALUE the
// resolution can produce, and they pass unchanged across this refactor — which is
// the inertness proof. What none of them can see is WHICH CODE produced the
// value: a private copy of the parse satisfies them exactly as well as the shared
// helper does. So they cannot detect the divergence coming back.
//
// This file adds the missing assertion: that the substep site actually routes
// through datahelpers.GetBoolFieldLoud, observed through the warning's shape.
package orchestration

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestSubstepMalformedDeclarationWarnsThroughTheSharedHelper.
//
// TestSubstepContinueOnErrorMalformedDeclarationFallsBack already proves the
// VALUE is right for a malformed declaration. It cannot prove anything about the
// warning, because that harness passes zap.NewNop(). This one observes the log.
//
// The `config_key` field is the discriminator, and it is deliberately chosen:
// the pre-193 hand-rolled substep warning named the substep and the declared
// type but had NO config_key field — only the shared helper emits that, because
// only the shared helper is generic over the key. So:
//
//	MUTATION: re-inline the old private parse in resolveSubstepContinueOnError
//	(the exact pre-193 body, warning included). Every value assertion still
//	passes, and the config_key assertion goes RED — which is the divergence this
//	bug exists to prevent, caught mechanically.
func TestSubstepMalformedDeclarationWarnsThroughTheSharedHelper(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	logger := zap.New(core)

	got := resolveSubstepContinueOnError(
		map[string]interface{}{"continue_on_error": "true"}, // the string mistake
		true, // the loop is tolerant; a malformed substep declaration inherits that
		"write_page",
		logger,
	)

	if got != true {
		t.Errorf("malformed declaration returned %v; it must inherit the loop's true", got)
	}
	if logs.Len() != 1 {
		t.Fatalf("want exactly 1 warning for a malformed declaration, got %d", logs.Len())
	}

	fields := logs.All()[0].ContextMap()
	if fields["substep"] != "write_page" {
		t.Errorf("the warning does not name the substep, so a reader cannot find the declaration: %v", fields)
	}
	if fields["config_key"] != "continue_on_error" {
		t.Errorf("no config_key field — this site is NOT going through datahelpers.GetBoolFieldLoud, "+
			"so the loop-level and substep-level parses have diverged again (bugs_open/193): %v", fields)
	}
}

// TestSubstepWellFormedDeclarationsStaySilent is the control. A guard proved only
// on its firing branch is satisfied by warning about everything, which would bury
// the real warning in noise on every expansion.
//
// MUTATION: warn unconditionally in GetBoolFieldLoud (before the isBool check). RED.
func TestSubstepWellFormedDeclarationsStaySilent(t *testing.T) {
	for _, declared := range []interface{}{true, false} {
		core, logs := observer.New(zap.WarnLevel)
		resolveSubstepContinueOnError(
			map[string]interface{}{"continue_on_error": declared}, false, "write_page", zap.New(core))
		if logs.Len() != 0 {
			t.Errorf("a well-formed declaration (%v) produced %d warning(s)", declared, logs.Len())
		}
	}
	// And an absent declaration — the commonest case by far, on every substep of
	// every loop. If this warned, the log would be unreadable.
	core, logs := observer.New(zap.WarnLevel)
	resolveSubstepContinueOnError(map[string]interface{}{}, true, "write_page", zap.New(core))
	if logs.Len() != 0 {
		t.Errorf("an absent declaration produced %d warning(s); silence is correct", logs.Len())
	}
}

// TestSubstepInheritanceIsTheFallbackNotAConstant pins the reason the shared
// helper takes the fallback as a PARAMETER rather than defaulting to false: at
// this site a silent substep INHERITS the loop's value, it does not opt out.
//
// MUTATION: pass `false` instead of loopContinueOnError as the helper's fallback.
// The inherit-true case goes RED. (This is also what would happen if someone
// "simplified" the helper by dropping the parameter.)
func TestSubstepInheritanceIsTheFallbackNotAConstant(t *testing.T) {
	for _, loopFlag := range []bool{true, false} {
		got := resolveSubstepContinueOnError(map[string]interface{}{}, loopFlag, "write_page", zap.NewNop())
		if got != loopFlag {
			t.Errorf("silent substep in a loop with continue_on_error=%v resolved to %v; it must inherit", loopFlag, got)
		}
		// And the same for a malformed declaration, which also inherits.
		got = resolveSubstepContinueOnError(
			map[string]interface{}{"continue_on_error": 1}, loopFlag, "write_page", zap.NewNop())
		if got != loopFlag {
			t.Errorf("malformed substep declaration in a loop with continue_on_error=%v resolved to %v; it must inherit", loopFlag, got)
		}
	}
}

// The end-to-end control is NOT duplicated here: the five existing tests in
// substep_continue_on_error_test.go already drive the real expansion path through
// handleLoopExpansion and assert what reaches the injected plan, and they pass
// unchanged across this refactor. Re-asserting that here would be a second copy
// of a guard that already exists — the exact habit bugs_open/193 is about.
