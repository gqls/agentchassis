// FILE: platform/orchestration/actions/extract_fields_required_test.go
//
// bugs_open/192. ExtractFieldsAction had no test at all, and its defining
// behaviour — a target field whose configured paths ALL miss is omitted from
// the result, and the step still reports success — is what turned an upstream
// shape change into a fleet-wide page-build outage diagnosed two steps away
// from its cause.
//
// These cases pin both halves: the lenient default is frozen deliberately (it
// is the historical contract of two live steps and must not change under
// anyone), and the new opt-in "required" list is proven to fail, to name the
// field, and to name what was actually in scope.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func extractFieldsParams(config map[string]interface{}, collected map[string]interface{}) ActionParams {
	return ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: config},
		CollectedData:    collected,
	}
}

// The wrapper shape from bugs_open/192: the plan demoted one level by an
// action that reused section_plan as its output_field.
func wrappedSectionPlan() map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"section_plan": map[string]interface{}{
				"applied": false,
				"reason":  "not_edit_live",
				"section_plan": map[string]interface{}{
					"sections_ready": []interface{}{map[string]interface{}{"name": "hero"}},
				},
			},
		},
		"resolved_links": map[string]interface{}{
			"response": map[string]interface{}{
				"link_resolution": map[string]interface{}{"sections_ready": nil},
			},
		},
	}
}

// Default OFF. Without "required", the pre-192 behaviour is preserved exactly:
// nothing resolves, the key is omitted, and the action still succeeds. This is
// the frozen legacy contract — if this test ever fails, a shared seam changed
// meaning for every existing caller.
func TestExtractFields_NoRequired_OmitsAndSucceeds(t *testing.T) {
	config := map[string]interface{}{
		"fields": map[string]interface{}{
			"sections_ready": []interface{}{
				"resolved_links.response.link_resolution.sections_ready",
				"input_data.section_plan.sections_ready",
			},
		},
	}

	got, err := ExtractFieldsAction(context.Background(), extractFieldsParams(config, wrappedSectionPlan()))
	if err != nil {
		t.Fatalf("lenient mode must not error, got: %v", err)
	}
	result, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map: %T", got)
	}
	if _, present := result["sections_ready"]; present {
		t.Errorf("no configured path resolves here, so the key must be absent; got %v", result["sections_ready"])
	}
}

// Opted in, and nothing resolves: the step fails HERE, and the message carries
// the three things needed to act on it without opening the database — which
// field, which paths were tried, and what was in scope.
func TestExtractFields_Required_FailsNamingTheCause(t *testing.T) {
	config := map[string]interface{}{
		"fields": map[string]interface{}{
			"sections_ready": []interface{}{
				"resolved_links.response.link_resolution.sections_ready",
				"input_data.section_plan.sections_ready",
			},
		},
		"required": []interface{}{"sections_ready"},
	}

	got, err := ExtractFieldsAction(context.Background(), extractFieldsParams(config, wrappedSectionPlan()))
	if err == nil {
		t.Fatalf("required field resolved via no path must fail the step, got success: %v", got)
	}
	msg := err.Error()
	for _, want := range []string{
		"sections_ready",                         // the field
		"input_data.section_plan.sections_ready", // a path that was tried
		"input_data",                             // what was actually in scope
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error must name %q so the cause is actionable; got: %s", want, msg)
		}
	}
}

// Opted in and satisfied by the SECOND path: fallbacks still work, and
// "required" costs nothing when the data is there. Uses the flat plan — i.e.
// the shape the 192 fix restores.
func TestExtractFields_Required_SatisfiedByFallbackPath(t *testing.T) {
	collected := map[string]interface{}{
		"input_data": map[string]interface{}{
			"section_plan": map[string]interface{}{
				"sections_ready": []interface{}{map[string]interface{}{"name": "hero"}},
			},
		},
		"resolved_links": map[string]interface{}{
			"response": map[string]interface{}{
				"link_resolution": map[string]interface{}{"sections_ready": nil},
			},
		},
	}
	config := map[string]interface{}{
		"fields": map[string]interface{}{
			"sections_ready": []interface{}{
				"resolved_links.response.link_resolution.sections_ready",
				"input_data.section_plan.sections_ready",
			},
		},
		"required": []interface{}{"sections_ready"},
	}

	got, err := ExtractFieldsAction(context.Background(), extractFieldsParams(config, collected))
	if err != nil {
		t.Fatalf("second path resolves, so the step must succeed; got: %v", err)
	}
	sections, ok := got.(map[string]interface{})["sections_ready"].([]interface{})
	if !ok || len(sections) != 1 {
		t.Fatalf("sections_ready = %v, want the 1-element array from the fallback path", got)
	}
}

// A default satisfies "required" — which is why the check runs after defaults
// are applied, not before.
func TestExtractFields_Required_SatisfiedByDefault(t *testing.T) {
	config := map[string]interface{}{
		"fields":   map[string]interface{}{"topic": []interface{}{"input_data.nothing_here"}},
		"defaults": map[string]interface{}{"topic": "fallback-topic"},
		"required": []interface{}{"topic"},
	}

	got, err := ExtractFieldsAction(context.Background(), extractFieldsParams(config, map[string]interface{}{}))
	if err != nil {
		t.Fatalf("a default satisfies required, so the step must succeed; got: %v", err)
	}
	if got.(map[string]interface{})["topic"] != "fallback-topic" {
		t.Errorf("topic = %v, want fallback-topic", got.(map[string]interface{})["topic"])
	}
}
