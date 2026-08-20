// FILE: platform/orchestration/datahelpers/template_reproducibility.go
//
// "Can this component's content_data reproduce its rendered_html?" — migration
// 499's routing test, as code. Written 2026-08-20 (bugs_open/277 §5, the
// LANDMINES entry "A component whose content_data CANNOT REPRODUCE its own
// rendered_html"): three repair routes in two days were costed against the
// wrong property (ownership) because nothing asked this question mechanically.
// The discovery check now asks it at filing time to pick a repair route; ask it
// yourself before costing any route that regenerates from content_data.
//
// The test is deliberately COARSE and errs toward "can fill" (= keep the
// regenerate-from-source route, today's default): a false "can fill" costs a
// refused rerender and a human escalation — today's behaviour — while a false
// "cannot fill" would route a regenerable component to an HTML-surface edit
// whose transform then converts what the detector flagged; still correct
// output, but content_data would keep its defect and the next regeneration
// would reprint it. Coarse in the safe direction, stated.
//
// THIS IS A ROUTING TEST, NOT A RENDER-COMPLETENESS TEST — the council's
// bug_historian seat objected (corr b72a4029 r1) that ANY-field-present is too
// weak for "will this render complete", and it is: a template with {{.A}} and
// {{.B}} where only A is populated still blanks B under missingkey=zero (the
// 004/007 blanking family). That question already has a per-field, scope-aware
// instrument: actions.missingBareFields (component_library.go), which walks
// the parsed template tree. It is NOT reused here because `actions` imports
// this package's consumer (`discovery_checks`) — four live imports, so the
// reverse import is a cycle — and moving that live guard here is its own
// refactor. The judgements also genuinely differ: partially-fillable
// components MUST keep the regenerate route (their content_data is real and
// an HTML-surface edit cannot fix its defects), so for ROUTING, any-field-
// present is the correct bar, not a weak one. A page classified "can fill"
// gets an item byte-identical to what this detector minted before this
// helper existed — the partial-blanking exposure there is pre-existing and
// unchanged. If you change the field-extraction rules here, check
// missingBareFields' walk for the same case, and vice versa — the two must
// not silently disagree about what a template reads.

package datahelpers

import "regexp"

// templateTopLevelFieldRe matches the top-level content_data key a Go template
// construct reads: {{.name}}, {{.name.sub}}, {{range .items}}, {{with .x}},
// {{if .flag}} — with optional "-" trim markers. It deliberately does NOT
// match fields reached through functions or variables ({{if eq .a .b}},
// {{$x := .y}}): missing those shrinks the field list, and a smaller list can
// only push the verdict toward "can fill" — the safe direction above.
var templateTopLevelFieldRe = regexp.MustCompile(`\{\{-?\s*(?:(?:range|with|if)\s+)?\.([A-Za-z_][A-Za-z0-9_]*)`)

// TemplateTopLevelFields returns the distinct top-level content_data keys a
// template reads, in first-appearance order.
func TemplateTopLevelFields(tpl string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range templateTopLevelFieldRe.FindAllStringSubmatch(tpl, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// ContentDataCanFillTemplate reports whether content_data holds ANY of the
// template's top-level fields, i.e. whether re-rendering this template from
// this content_data can reproduce any of its content. A template that names no
// fields is static and trivially reproducible (returns true). The worked case
// this exists for: Ported Page's template reads only {{.body}} while 100 of
// 115 instances hold none of it — those render to an empty body with err=nil
// under missingkey=zero, so no error path will ever tell you this.
func ContentDataCanFillTemplate(tpl string, contentData map[string]interface{}) bool {
	fields := TemplateTopLevelFields(tpl)
	if len(fields) == 0 {
		return true
	}
	for _, f := range fields {
		if v, ok := contentData[f]; ok && v != nil && v != "" {
			return true
		}
	}
	return false
}
