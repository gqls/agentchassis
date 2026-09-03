// FILE: platform/orchestration/actions/template_context_contract_test.go
//
// bugs_open/453. The declaration in template_context_contract.go is only worth
// reading if it matches what the actions really do, so the load-bearing test
// here drives injectPlatformBlocks for real and compares the keys it WROTE
// against the keys the contract CLAIMS. A list checked against itself would pass
// on a stale declaration, which is the failure this file exists to prevent.
package actions

import (
	"context"
	"testing"

	"github.com/gqls/agentchassis/platform/voicestyle"
)

// The behavioural pin: what injectPlatformBlocks writes IS what
// TemplateRootsInjectedBy("execute_llm_prompt") declares. Adding a third
// platform block without extending the contract fails here.
func TestDeclaredLLMInjectedRootsMatchInjectPlatformBlocks(t *testing.T) {
	td := map[string]interface{}{}
	injectPlatformBlocks(context.Background(), td, func(_ context.Context, name string) (string, bool) {
		// Every configured block resolves, so the map ends up holding exactly
		// the keys the injector is capable of writing.
		return "block text for " + name, true
	})

	written := map[string]bool{}
	for k := range td {
		written[k] = true
	}
	declared := TemplateRootsInjectedBy("execute_llm_prompt")

	if len(declared) != len(written) {
		t.Fatalf("injectPlatformBlocks wrote %d root(s) %v; the contract declares %d %v.\n"+
			"An undeclared injected root makes the offline check report a false finding on every "+
			"template that reads it.", len(written), td, len(declared), declared)
	}
	for _, d := range declared {
		if !written[d] {
			t.Errorf("contract declares %q but injectPlatformBlocks never writes it", d)
		}
	}

	// Sanity: the injector really is driven by platformPromptBlocks, so the
	// comparison above is not two empty sets agreeing.
	if len(written) == 0 {
		t.Fatal("no blocks written — this test would pass vacuously; " +
			"platformPromptBlocks is empty or the stub getter refused everything")
	}
	if _, ok := td[voicestyle.ConfigName]; ok {
		t.Error("the template root is the map KEY (voice_style), not the config NAME — " +
			"if these ever coincide, this test stops discriminating")
	}
}

// The vision manifest is declared, and the constant is the one the action uses
// at its injection site — so a rename is a compile error rather than a silent
// divergence.
func TestVisionInjectedRootIsDeclared(t *testing.T) {
	got := TemplateRootsInjectedBy("execute_vision_prompt")
	if len(got) != 1 || got[0] != VisionImageManifestKey {
		t.Fatalf("TemplateRootsInjectedBy(execute_vision_prompt) = %v, want [%s]", got, VisionImageManifestKey)
	}
	if VisionImageManifestKey != "vision_image_manifest" {
		t.Errorf("VisionImageManifestKey = %q; live templates write {{.vision_image_manifest}} "+
			"and renaming the constant does not rename them", VisionImageManifestKey)
	}
}

// The two sets must NOT be a union. This is the false positive the lane's first
// sizing pass shipped, in both directions.
func TestInjectedRootsAreNotSharedBetweenTheTwoActions(t *testing.T) {
	llm := map[string]bool{}
	for _, r := range TemplateRootsInjectedBy("execute_llm_prompt") {
		llm[r] = true
	}
	vision := map[string]bool{}
	for _, r := range TemplateRootsInjectedBy("execute_vision_prompt") {
		vision[r] = true
	}
	if llm[VisionImageManifestKey] {
		t.Error("execute_llm_prompt does not set the vision manifest")
	}
	if vision["voice_style"] {
		t.Error("execute_vision_prompt does not call injectPlatformBlocks, so voice_style is NOT available to it")
	}
}

func TestRendersPromptTemplateScope(t *testing.T) {
	for _, a := range []string{"execute_llm_prompt", "execute_vision_prompt"} {
		if !RendersPromptTemplate(a) {
			t.Errorf("%s renders a prompt_template through extractDataForAiAgent", a)
		}
	}
	for _, a := range []string{"create_work_item", "loop", "call_agent", "", "execute_llm_promptx"} {
		if RendersPromptTemplate(a) {
			t.Errorf("%s does not render a prompt_template; a prompt_template in its config is inert", a)
		}
	}
	if got := PromptTemplateRenderingActions(); len(got) != 2 {
		t.Errorf("PromptTemplateRenderingActions() = %v, want the two rendering actions", got)
	}
	if TemplateRootsInjectedBy("loop") != nil {
		t.Error("an action that renders nothing injects nothing")
	}
}

// The composition: the extractor's half plus the action's half, with the
// undecidable flag carried through unchanged.
func TestTemplateRootsAvailableTo(t *testing.T) {
	roots, promoted := TemplateRootsAvailableTo("execute_llm_prompt", []string{"current_page"})
	if promoted {
		t.Error("no input_data declared")
	}
	for _, want := range []string{"current_page", "domain", "objective", "model", "voice_style", "build_standard"} {
		if !roots[want] {
			t.Errorf("missing root %q", want)
		}
	}
	if roots[VisionImageManifestKey] {
		t.Error("the vision manifest is not available to an LLM prompt")
	}

	if _, promoted := TemplateRootsAvailableTo("execute_llm_prompt", []string{"input_data"}); !promoted {
		t.Error("input_data promotion must survive the composition — it is what makes a finding advisory")
	}
}
