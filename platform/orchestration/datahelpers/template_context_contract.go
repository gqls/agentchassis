// FILE: platform/orchestration/datahelpers/template_context_contract.go
//
// What ExtractFields makes visible to a prompt template, stated as symbols
// instead of as folklore.
//
// WHY THIS FILE EXISTS. A step renders its prompt_template against
// ExtractFields(CollectedData, input_fields) — a SUBSET of the collected data.
// A template variable whose root is not in that subset is simply absent at
// render: Go templates render a guarded or ranged absent key as NOTHING and an
// unguarded one as "<no value>". No error, no verdict. bugs_open/453 records
// four separate catches of that class, every one of them found by a fixture
// somebody happened to write.
//
// Its fix candidate 1 asks for a lint over the pair, and states the condition
// that makes the lint worth having: "the extractor's speciallyHandled set must
// be read from ONE place or the lint inherits the classifier-gap problem." This
// file is that one place. cmd/config-key-audit --template-input-fields calls
// these functions; it does not carry a list.
//
// The distinction matters more than it looks. A copied list is not wrong when it
// is written — it is wrong LATER, silently, and in the direction that makes the
// check noisier: a field added here that the copy never learns about turns every
// template using it into a false finding, and noise is what gets a check
// switched off.
//
// ⚠ WHAT THIS FILE DOES NOT COVER. Roots injected AFTER ExtractFields returns,
// by the calling action, are the action's contract and not the extractor's —
// execute_llm_prompt adds the platform prompt blocks, execute_vision_prompt adds
// its image manifest. Those are declared in
// platform/orchestration/actions/template_context_contract.go. Asking only this
// file gives an INCOMPLETE root set, which convicts a correct template; the
// first cut of 453's own sizing script did exactly that and reported both vision
// steps as broken.
package datahelpers

import (
	"strings"
	"text/template"
)

// speciallyHandledInputFields are the input_fields names ExtractFields resolves
// with dedicated logic above its generic loop, rather than through
// extractSingleField. Each one is stored under its own name, so each is also the
// template root it provides.
//
// Unexported deliberately: a package-level exported MAP is writable by every
// importer, and a check that can silently mutate the contract it is auditing is
// not a check. Ask through the two functions below.
var speciallyHandledInputFields = map[string]bool{
	"input_data":      true,
	"reviewed_brief":  true,
	"site_record":     true,
	"current_page":    true,
	"current_section": true,
}

// IsSpeciallyHandledInputField reports whether ExtractFields resolves this
// input_fields entry with dedicated logic. ExtractFields itself calls this, so
// the answer cannot drift from the behaviour.
func IsSpeciallyHandledInputField(name string) bool {
	return speciallyHandledInputFields[name]
}

// SpeciallyHandledInputFields returns the set as a fresh slice, unordered.
// Callers that need a stable order sort it themselves.
func SpeciallyHandledInputFields() []string {
	out := make([]string, 0, len(speciallyHandledInputFields))
	for k := range speciallyHandledInputFields {
		out = append(out, k)
	}
	return out
}

// TemplateRootForInputField returns the template root an input_fields entry
// makes available — the LAST dotted segment, not the first.
//
// This is the single most misread line of the extractor, and it is misread in
// the expensive direction. `input_fields: ["sections_for_render.sections_ready"]`
// makes {{.sections_ready}} work and {{.sections_for_render.sections_ready}}
// NOT work: the value is stored under the leaf name, and the path that named it
// is gone by render time. A reader who assumes "root = first segment" concludes
// the template is fine when it is dead, and — the same bug from the other side —
// validateTemplateData in ai_actions.go splits on the first segment and so calls
// a SUCCESSFUL dotted extraction "missing".
//
// ExtractFields calls this, so the rule and the behaviour are one thing.
func TemplateRootForInputField(field string) string {
	parts := strings.Split(field, ".")
	return parts[len(parts)-1]
}

// AlwaysEnsuredTemplateRoots are the roots ExtractFields supplies REGARDLESS of
// what input_fields asks for, via ensureCoreFields: domain, objective and model
// are each searched for aggressively and written to the result when found.
//
// A check that omits these reports a false finding on every template that says
// {{.domain}} — which is most of them.
//
// Pinned to the behaviour by TestAlwaysEnsuredTemplateRootsMatchesExtractFields,
// which drives ExtractFields with an EMPTY input_fields list and compares the
// keys it produced against this list. That is a behavioural pin rather than a
// source scan: it fails if ensureCoreFields stops supplying one of these, and it
// fails if ensureCoreFields starts supplying a fourth.
//
// ⚠ current_page, current_section and render_context are deliberately NOT here.
// ensureCoreFields recovers them ONLY when input_fields requests them (the
// `requested(...)` gate), so they are ordinary requested fields, not always-on
// ones.
func AlwaysEnsuredTemplateRoots() []string {
	return []string{"domain", "objective", "model"}
}

// TemplateRootsFor returns every root the EXTRACTOR alone makes available for a
// step declaring these input_fields — the always-ensured set plus one root per
// declared field.
//
// inputDataPromoted is the part a static caller cannot resolve and must not
// pretend to. When "input_data" is among the fields, ExtractFields copies every
// key of the runtime input_data map to the ROOT level:
//
//	for k, v := range existingInputMap { if _, exists := result[k]; !exists { result[k] = v } }
//
// so the root set for that step depends on a row, and no check over config can
// enumerate it. Callers are expected to report findings on such steps as
// undecidable rather than as defects — see the conditional_root class in
// cmd/config-key-audit/templateinputfields.go.
func TemplateRootsFor(inputFields []string) (roots map[string]bool, inputDataPromoted bool) {
	roots = make(map[string]bool, len(inputFields)+3)
	for _, r := range AlwaysEnsuredTemplateRoots() {
		roots[r] = true
	}
	for _, f := range inputFields {
		if f == "" {
			continue
		}
		roots[TemplateRootForInputField(f)] = true
		if f == "input_data" {
			inputDataPromoted = true
		}
	}
	return roots, inputDataPromoted
}

// PromptTemplateFuncs is the function map every prompt template is parsed and
// executed with. RenderPromptTemplate calls it, so this IS the fleet's template
// dialect rather than a description of it.
//
// Exported because text/template checks function names at PARSE time, not at
// execution: a template saying {{toJSON .x}} fails to parse for anyone who did
// not register toJSON. An offline check that built its own empty func map would
// therefore report a parse failure on every template using a helper — a
// fleet-wide false alarm that looks like corrupted config. Adding a helper here
// keeps the check parsing exactly what production parses.
func PromptTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"toJSON":      templateToJSON,
		"placeholder": templatePlaceholder,
		"rangeStart":  templateRangeStart,
		"rangeEnd":    templateRangeEnd,
	}
}
