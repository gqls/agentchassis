// FILE: platform/orchestration/actions/generate_image_refusal_test.go
//
// bugs_open/210 (needs_logo slug) — the framework guard.
//
// GenerateImageAction must REFUSE when no caller supplied a prompt, rather than
// paint from getImagePromptWithPriority's generic fallback ("Generate content
// based on the provided context.") and let store_logo_asset save the result as
// the site's logo. A silently wrong brand asset is strictly worse than a loud
// failure, and this is the single point every generated image passes through —
// so the guard does not depend on any of the three producers remembering
// anything.
//
// MUTATION-PROVEN 2026-08-09, with the result recorded as observed rather than
// as intended. Disabling the guard IN PLACE (`if false && promptSource == …`,
// so the package still compiles — a deletion instead breaks the build, and a
// build failure is not evidence that the TEST catches anything)
// TestGenerateImage_RefusesWhenNoCallerSuppliedAPrompt FAILS, and it fails by
// running on to the Kafka publish and panicking on the absent producer. That
// is precisely the real-world behaviour the guard prevents: execution
// continues, an image is generated from a meaningless sentence, and the caller
// stores it as a brand asset.
//
// This is the check `mutate-the-code-to-prove-the-guard` asks for, and it also
// answers `a-mutation-that-passes-may-have-hit-a-guard-in-series`: the test is
// evidence only because disabling THIS guard, and nothing else, changes the
// outcome.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// The exact live configuration: no prompt in the step config, none in collected
// data, no input_data.prompt, and an agent whose default_config carries no
// prompt_template. Measured 2026-08-09 — that is not a contrived fixture, it is
// how image-generator is configured on this fleet, which is why Priorities 2
// and 3 of the documented chain are unreachable and the generic fallback is the
// only remaining tier.
func generateImageParamsWithNoPrompt() ActionParams {
	return ActionParams{
		DB:               nil, // getImageryStyleGuideForSite returns nil on a nil db
		Logger:           zap.NewNop(),
		AgentType:        "image-generator",
		ExecutionContext: &types.ExecutionContext{StepName: "generate"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			// Supplied so the action does not go to the database for it. The
			// EMPTY map is the point: image-generator's live default_config has
			// no prompt_template, so Priority 2 is absent — verified 2026-08-09
			// with `default_config ? 'prompt_template'` → false.
			"agent_config": map[string]interface{}{},
			"input_data": map[string]interface{}{
				"kind": "logo",
			},
		},
	}
}

func TestGenerateImage_RefusesWhenNoCallerSuppliedAPrompt(t *testing.T) {
	out, err := GenerateImageAction(context.Background(), generateImageParamsWithNoPrompt())

	if err == nil {
		t.Fatalf("expected a refusal, got out=%#v err=nil — an image would be generated from "+
			"the generic fallback and stored as the site's logo", out)
	}
	if !strings.Contains(err.Error(), "refused") {
		t.Fatalf("expected the refusal error, got a different failure: %v", err)
	}
	// The message must tell an operator what to fix. A bare "refused" sends the
	// next reader back into the code to find out which mapping is missing.
	if !strings.Contains(err.Error(), "input_data.prompt") {
		t.Errorf("refusal does not name the field a caller must map: %v", err)
	}
	if !strings.Contains(err.Error(), "logo") {
		t.Errorf("refusal does not name the kind, so it cannot be triaged: %v", err)
	}
}

// Pins WHY the guard is reachable, so that if a prompt_template is ever added to
// image-generator's config this test is the one that changes and tells the next
// author which rung moved. Without it, a future config change could make the
// guard dead and nothing would say so.
func TestGetImagePrompt_GenericFallbackIsTheOnlyRemainingTier(t *testing.T) {
	params := generateImageParamsWithNoPrompt()

	prompt, source := getImagePromptWithPriority(params, map[string]interface{}{})

	if source != imagePromptSourceGenericFallback {
		t.Fatalf("expected the generic fallback tier, got source=%q prompt=%q — if a "+
			"prompt_template has been added to image-generator, the refusal guard in "+
			"GenerateImageAction may now be unreachable; re-read bugs_open/210 §4 before "+
			"assuming that is safe", source, prompt)
	}
	if prompt != "Generate content based on the provided context." {
		t.Errorf("the generic fallback text changed to %q — the bug file, the LANDMINES entry "+
			"and the assets census all quote the old string", prompt)
	}
}

// A caller that DOES supply a prompt must not be refused. Without this the
// guard could be over-broad — refusing everything — and the test above would
// still pass, which is exactly the `check-the-no-op-case-not-only-the-damage-case`
// failure.
func TestGenerateImage_DoesNotRefuseWhenAPromptIsSupplied(t *testing.T) {
	params := generateImageParamsWithNoPrompt()
	params.StepConfig.Config["prompt"] = "A minimal geometric logo mark for a UK vet comparison service"

	// This test deliberately does not stand up Kafka, so execution that gets
	// PAST the guard reaches the producer and panics on it. That panic is the
	// positive signal: it can only happen downstream of the refusal branch, so
	// recovering here and asserting "no refusal error" is a real assertion
	// rather than a swallowed failure.
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("reached the publish path and panicked on the absent Kafka producer "+
					"(%v) — which is only possible past the guard", r)
			}
		}()
		_, err = GenerateImageAction(context.Background(), params)
	}()

	if err != nil && strings.Contains(err.Error(), "refused") {
		t.Fatalf("a caller that supplied a real prompt was refused: %v", err)
	}
}
