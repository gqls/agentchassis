// FILE: platform/validation/workflow_config_keys_test.go
//
// bugs_open/101 — the workflow validator is the only place that sees an action
// name and its step config together on every run, so it is where an inert config
// key stops being invisible.

package validation

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func planWithStep(action string, config map[string]interface{}) models.WorkflowPlan {
	return models.WorkflowPlan{
		StartStep: "only",
		Steps: map[string]models.Step{
			"only": {
				Action: action,
				// Non-local actions need a topic, and that check runs first —
				// without this the test would fail for an unrelated reason.
				Topic:  "system.test.requests",
				Config: config,
			},
		},
	}
}

// TestUnknownConfigKeysWarnByDefault: a declared-but-not-strict action reports
// its unknown keys and the workflow still runs. Failing here instead would take
// the fleet down over a key list that is merely incomplete.
func TestUnknownConfigKeysWarnByDefault(t *testing.T) {
	datahelpers.RegisterActionInputSpec("test_validator_warn", datahelpers.ActionInputSpec{
		ConfigKeys: []string{"url_field"},
	})

	core, logs := observer.New(zapcore.WarnLevel)
	v := NewWorkflowValidator(zap.New(core))

	plan := planWithStep("test_validator_warn", map[string]interface{}{
		"url_field":    "input_data.url",
		"follow_links": []interface{}{"fees"},
		"max_pages":    3,
	})

	if err := v.ValidateWorkflow(plan); err != nil {
		t.Fatalf("validation must not fail for a non-strict action: %v", err)
	}

	entries := logs.FilterMessageSnippet("does not read").All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 warning, got %d", len(entries))
	}
	fields := entries[0].ContextMap()
	keys, _ := fields["unrecognised_keys"].([]interface{})
	if len(keys) != 2 {
		t.Errorf("expected both unknown keys named, got %v", fields["unrecognised_keys"])
	}
	if !strings.Contains(entries[0].Message, "silently ignored") {
		t.Errorf("warning should say what actually happens to the key: %q", entries[0].Message)
	}
}

// TestUnknownConfigKeyWarningIsDedupedPerCombination: ValidateWorkflow runs on
// EVERY message. An undeduped warning would repeat for the life of the pod, and a
// warning that scrolls is one nobody reads — leaving the key as invisible as the
// silent ignore it replaced.
func TestUnknownConfigKeyWarningIsDedupedPerCombination(t *testing.T) {
	datahelpers.RegisterActionInputSpec("test_validator_dedupe", datahelpers.ActionInputSpec{
		ConfigKeys: []string{"url_field"},
	})

	core, logs := observer.New(zapcore.WarnLevel)
	v := NewWorkflowValidator(zap.New(core))
	plan := planWithStep("test_validator_dedupe", map[string]interface{}{"bogus": 1})

	for i := 0; i < 5; i++ {
		if err := v.ValidateWorkflow(plan); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}

	if n := logs.FilterMessageSnippet("does not read").Len(); n != 1 {
		t.Errorf("warned %d times across 5 validations, want 1", n)
	}
}

// TestStrictConfigActionFailsValidation: an action that has declared its contract
// COMPLETE gets the door closed — an unknown key is a definition error, not a
// no-op. This is the end state bugs_open/101 candidate 1 asks for; scrape_web is
// not here yet because two live definitions would have to be corrected first.
func TestStrictConfigActionFailsValidation(t *testing.T) {
	datahelpers.RegisterActionInputSpec("test_validator_strict", datahelpers.ActionInputSpec{
		ConfigKeys:   []string{"url_field"},
		StrictConfig: true,
	})

	v := NewWorkflowValidator(zap.NewNop())

	t.Run("unknown key is refused", func(t *testing.T) {
		err := v.ValidateWorkflow(planWithStep("test_validator_strict", map[string]interface{}{
			"url_field":   "input_data.url",
			"folow_links": []interface{}{"fees"}, // typo — the realistic case
		}))
		if err == nil {
			t.Fatal("strict action accepted an unknown config key — a silent accept is the bug restating itself")
		}
		if !strings.Contains(err.Error(), "folow_links") {
			t.Errorf("error must name the offending key, got: %v", err)
		}
	})

	t.Run("clean config passes", func(t *testing.T) {
		// The positive control. Without it, a validator that rejected everything
		// would pass the test above for entirely the wrong reason.
		if err := v.ValidateWorkflow(planWithStep("test_validator_strict", map[string]interface{}{
			"url_field": "input_data.url",
		})); err != nil {
			t.Errorf("clean config rejected: %v", err)
		}
	})
}

// TestRemovedConfigKeyFailsValidation (bugs_open/234): a RETIRED key is a hard
// error with the replacement in the message — on any action that declares it,
// strict or not, opted into unknown-key detection or not. Warn-only is what let
// improvement-loop's refresh_site_components flag stay silently dropped for
// months, so the one thing this test must never tolerate is the removed key
// downgrading to a warning.
func TestRemovedConfigKeyFailsValidation(t *testing.T) {
	// Deliberately NO ConfigKeys and NO CheckConfig: the removed-key check must
	// fire independently of the unknown-key opt-in, or a declaration on a
	// not-yet-opted-in action is silently inert — the "declared but never
	// consulted" shape.
	datahelpers.RegisterActionInputSpec("test_validator_removed_only", datahelpers.ActionInputSpec{
		RemovedConfigKeys: map[string]string{
			"spec": "never read; use spec_literal — bugs_open/234",
		},
	})

	v := NewWorkflowValidator(zap.NewNop())

	t.Run("removed key is refused with the replacement named", func(t *testing.T) {
		err := v.ValidateWorkflow(planWithStep("test_validator_removed_only", map[string]interface{}{
			"spec": map[string]interface{}{"refresh_site_components": true},
		}))
		if err == nil {
			t.Fatal("a declared-removed key validated clean — the key is dead and the " +
				"behaviour it describes silently does not happen, which is bugs_open/234 restated")
		}
		if !strings.Contains(err.Error(), `"spec"`) {
			t.Errorf("error must name the offending key, got: %v", err)
		}
		if !strings.Contains(err.Error(), "spec_literal") {
			t.Errorf("error must carry the replacement message, got: %v", err)
		}
	})

	t.Run("clean config passes", func(t *testing.T) {
		// Positive control: a validator rejecting everything would pass the
		// subtest above for the wrong reason.
		if err := v.ValidateWorkflow(planWithStep("test_validator_removed_only", map[string]interface{}{
			"spec_literal": map[string]interface{}{"refresh_site_components": true},
		})); err != nil {
			t.Errorf("clean config rejected: %v", err)
		}
	})
}

// TestRemovedKeySpecificErrorBeatsStrictGenericOne: when a step carries BOTH a
// removed key and a novel unknown key on a strict action, the removed-key error
// (which names the fix) must be the one returned — checking it first is what the
// validator promises, and this is the input where the ordering is observable.
func TestRemovedKeySpecificErrorBeatsStrictGenericOne(t *testing.T) {
	datahelpers.RegisterActionInputSpec("test_validator_removed_strict", datahelpers.ActionInputSpec{
		ConfigKeys:   []string{"url_field"},
		StrictConfig: true,
		RemovedConfigKeys: map[string]string{
			"old_key": "retired; use url_field",
		},
	})

	v := NewWorkflowValidator(zap.NewNop())
	err := v.ValidateWorkflow(planWithStep("test_validator_removed_strict", map[string]interface{}{
		"old_key":     "x",
		"novel_bogus": 1,
	}))
	if err == nil {
		t.Fatal("strict action with a removed key validated clean")
	}
	if !strings.Contains(err.Error(), "REMOVED") || !strings.Contains(err.Error(), "retired; use url_field") {
		t.Errorf("the removed-key message (with its replacement) must win over the generic strict one, got: %v", err)
	}
}

// TestRemovedKeyIsNotAlsoUnknown: the removed key must not double-report under
// the softer UNKNOWN label — it is its own category with its own consequence.
// (The datahelpers half of this pin lives in unknown_config_keys_test.go; this
// one proves the two checks compose in the validator without the removed key
// leaking into the strict error's key list.)
func TestRemovedKeyIsNotAlsoUnknown(t *testing.T) {
	datahelpers.RegisterActionInputSpec("test_validator_removed_unknown_split", datahelpers.ActionInputSpec{
		ConfigKeys: []string{"url_field"},
		RemovedConfigKeys: map[string]string{
			"old_key": "retired; use url_field",
		},
	})

	unknown, checked := datahelpers.UnknownConfigKeys("test_validator_removed_unknown_split",
		map[string]interface{}{"old_key": "x", "genuinely_unknown": 1})
	if !checked {
		t.Fatal("action with ConfigKeys should be checked")
	}
	if len(unknown) != 1 || unknown[0] != "genuinely_unknown" {
		t.Errorf("unknown = %v, want exactly [genuinely_unknown] — a removed key is "+
			"known-dead, not unknown", unknown)
	}
}

// TestUndeclaredActionIsUntouched: the opt-in guarantee. 208 live actions have not
// declared anything; none of them may gain a warning or an error from this change.
func TestUndeclaredActionIsUntouched(t *testing.T) {
	core, logs := observer.New(zapcore.WarnLevel)
	v := NewWorkflowValidator(zap.New(core))

	err := v.ValidateWorkflow(planWithStep("test_validator_never_registered", map[string]interface{}{
		"anything": 1, "at_all": 2,
	}))
	if err != nil {
		t.Fatalf("undeclared action must validate exactly as before: %v", err)
	}
	if n := logs.FilterMessageSnippet("does not read").Len(); n != 0 {
		t.Errorf("undeclared action produced %d warnings, want 0", n)
	}
}
