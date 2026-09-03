// FILE: platform/orchestration/actions/template_context_contract.go
//
// Which actions render a prompt_template, and what each one adds to the template
// context AFTER datahelpers.ExtractFields has run.
//
// The extractor's half of the contract is
// platform/orchestration/datahelpers/template_context_contract.go. This is the
// other half, and it lives here because it is NOT uniform: what a template can
// say depends on which action renders it.
//
//	execute_llm_prompt    ExtractFields + injectPlatformBlocks  -> voice_style, build_standard
//	execute_vision_prompt ExtractFields + its own image manifest -> vision_image_manifest
//
// Note that neither is a superset of the other. A vision prompt writing
// {{.voice_style}} gets nothing, because execute_vision_prompt does not call
// injectPlatformBlocks; an LLM prompt writing {{.vision_image_manifest}} gets
// nothing, for the mirror reason.
//
// WHY IT IS DECLARED RATHER THAN DERIVED. bugs_open/453's fix candidate asks for
// a lint over template variables and input_fields, and warns that an un-sourced
// copy of the extractor's special-case list would give the lint the same
// classifier gap it exists to close. The same argument applies with more force
// here, and this lane proved it on itself: the first sizing pass used ONE global
// injected-roots set and reported both live execute_vision_prompt steps as
// broken — two false positives out of twelve findings, on a set small enough to
// read by eye.
//
// The keys below are used AT the injection sites, not merely described here, so
// renaming one is a compile error rather than a silent divergence. That closes
// the rename case completely. It does NOT close the ADDITION case: an action
// that starts injecting a brand-new key with a fresh string literal will not
// appear here, and the check will then report a false finding on any template
// that reads it. There is no mechanism that notices that automatically — so if
// you add a key to a template context, add it here in the same commit. The
// blast radius is one JSON field in an advisory report, which is why this is a
// stated limitation and not a guard.
package actions

import (
	"sort"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// VisionImageManifestKey is the template root execute_vision_prompt writes into
// the context itself, listing the images it attached so the model can cite the
// page each one came from. Used at the injection site in
// execute_vision_prompt_action.go.
const VisionImageManifestKey = "vision_image_manifest"

// promptTemplateRenderingActions maps each action that renders a step's
// prompt_template through extractDataForAiAgent to the extra template roots it
// injects afterwards.
//
// Membership of this map is the definition of "in scope for the
// template-variable lint": an action that is absent renders no prompt_template,
// and a step carrying one under that action has inert config rather than a
// broken template.
//
// execute_llm_prompt's entry is nil because its injected roots are not a literal
// — they are the keys of platformPromptBlocks, read live below, so adding a
// third platform-wide block extends this contract with no edit here.
var promptTemplateRenderingActions = map[string][]string{
	"execute_llm_prompt":    nil,
	"execute_vision_prompt": {VisionImageManifestKey},
}

// RendersPromptTemplate reports whether this action renders a step's
// prompt_template against ExtractFields(input_fields).
func RendersPromptTemplate(action string) bool {
	_, ok := promptTemplateRenderingActions[action]
	return ok
}

// PromptTemplateRenderingActions lists those actions, sorted.
func PromptTemplateRenderingActions() []string {
	out := make([]string, 0, len(promptTemplateRenderingActions))
	for a := range promptTemplateRenderingActions {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

// TemplateRootsInjectedBy returns the roots this action adds to the template
// context after ExtractFields — sorted, empty for an action that renders no
// template.
func TemplateRootsInjectedBy(action string) []string {
	extra, ok := promptTemplateRenderingActions[action]
	if !ok {
		return nil
	}
	out := append([]string(nil), extra...)
	if action == "execute_llm_prompt" {
		for key := range platformPromptBlocks {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

// TemplateRootsAvailableTo answers the whole question for one step: every root a
// template rendered by this action, with these input_fields, can resolve.
//
// inputDataPromoted carries the extractor's undecidable case outward unchanged —
// when it is true, ExtractFields also promotes every key of the runtime
// input_data map to the root level, so the returned set is a LOWER BOUND rather
// than the whole truth and a caller must not convict a variable outside it.
func TemplateRootsAvailableTo(action string, inputFields []string) (roots map[string]bool, inputDataPromoted bool) {
	roots, inputDataPromoted = datahelpers.TemplateRootsFor(inputFields)
	for _, r := range TemplateRootsInjectedBy(action) {
		roots[r] = true
	}
	return roots, inputDataPromoted
}
