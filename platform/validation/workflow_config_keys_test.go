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
