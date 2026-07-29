// FILE: platform/orchestration/datahelpers/element_refs.go
//
// OrphanElementRefs — the element ids a page's own JavaScript addresses that
// exist nowhere in that page's HTML and are never created by the script.
//
// WHY THIS EXISTS. webdesign.co.uk's 63 tools were PORTED from two other sites
// as HTML blobs rather than born through the generator. Whatever trimmed them
// dropped markup the scripts need, and the result ships a page whose script
// throws on the first line that touches a missing element:
//
//	TypeError: Cannot read properties of null (reading 'getContext')
//	TypeError: Cannot set properties of null (setting 'innerHTML')
//
// The visitor sees a tool with no inputs. Nothing in the platform saw anything:
// the page renders, deploys, returns 200, and every existing check passes it.
// Measured 2026-07-29 — 10 live pages fleet-wide are in this state and the
// oldest has been broken since it was ported.
//
// WHY IT IS DECIDABLE STATICALLY, unlike its neighbours. check_dead_controls
// deliberately refuses to judge "<button> with no handler" from source, because
// JS binds handlers at runtime and absence of evidence is not evidence. This is
// the opposite case: the script names the id as a STRING LITERAL, the id is
// either in the document text or it is not, and no runtime behaviour can
// conjure one that the page never writes. The claim is about the artefact, not
// about what the artefact will do.
//
// EVERY RULE BELOW EXISTS TO AVOID A FALSE POSITIVE, because a wrong finding
// here sends a fixer at a working tool:
//
//   - present ids are harvested from the WHOLE page text, INCLUDING inside
//     script string literals. A tool that builds its own markup with
//     `out.innerHTML = '<input id="qty">'` has that id counted as present. This
//     over-counts on purpose: a name that appears anywhere is never reported.
//   - an id assigned dynamically (`el.id = 'x'`, `setAttribute('id','x')`)
//     counts as present.
//   - only bare `#id` selectors are read from querySelector. A compound or
//     descendant selector is ignored rather than half-parsed.
//   - the caller passes the WHOLE page — every component plus the site chrome —
//     so a script legitimately reaching a nav element is not flagged.
//
// MEASURED PRECISION, before this shipped (2026-07-29, live DB):
//
//	98 deployed script-carrying pages fleet-wide → 10 flagged, all 10 on
//	webdesign.co.uk, all 10 confirmed broken in a real browser.
//	32 library tool templates → 0 flagged.
//
// The 0-of-32 is what makes a hard pre-deploy refusal safe (deploy_tool_action):
// it cannot fire on anything that exists today.
//
// KNOWN AND DELIBERATE FALSE NEGATIVES. This finds one defect class, not all of
// them. Two tools that throw the same null-reference error are NOT flagged
// (blueprint-compiler, micro-cms) because they address elements by class or by
// a compound selector. Widening the selector parse would trade the precision
// above for coverage; the browser tier is where that coverage belongs.

package datahelpers

import (
	"regexp"
	"sort"
)

var (
	// document.getElementById('foo') — the id is a string literal.
	refByIDRe = regexp.MustCompile(`getElementById\(\s*['"]([A-Za-z0-9_\-:.]+)['"]\s*\)`)

	// querySelector('#foo') / querySelectorAll('#foo') — bare id selectors only.
	// A compound selector ('#foo .bar', '#foo, #baz') does not match and is
	// therefore never judged.
	refQuerySelRe = regexp.MustCompile(`querySelector(?:All)?\(\s*['"]#([A-Za-z0-9_\-]+)['"]\s*\)`)

	// id="foo" anywhere in the page text, markup or string literal alike.
	presentIDRe = regexp.MustCompile(`\bid\s*=\s*["']([^"']+)["']`)

	// el.id = 'foo' / setAttribute('id', 'foo') — created at runtime.
	dynamicIDRe = regexp.MustCompile(`(?:\.id\s*=|setAttribute\(\s*["']id["']\s*,)\s*["']([A-Za-z0-9_\-:.]+)["']`)
)

// OrphanElementRefs returns the sorted, de-duplicated ids that pageHTML's
// scripts address but that pageHTML never contains or creates. An empty result
// means the check found nothing to say — which is the common case and the only
// safe default.
//
// pageHTML must be the WHOLE page: every component's rendered HTML plus the
// site chrome. Passing one component in isolation will report ids that live in
// its siblings, which is a false positive by construction.
func OrphanElementRefs(pageHTML string) []string {
	referenced := map[string]struct{}{}
	for _, m := range refByIDRe.FindAllStringSubmatch(pageHTML, -1) {
		referenced[m[1]] = struct{}{}
	}
	for _, m := range refQuerySelRe.FindAllStringSubmatch(pageHTML, -1) {
		referenced[m[1]] = struct{}{}
	}
	if len(referenced) == 0 {
		return nil
	}

	for _, m := range presentIDRe.FindAllStringSubmatch(pageHTML, -1) {
		delete(referenced, m[1])
	}
	for _, m := range dynamicIDRe.FindAllStringSubmatch(pageHTML, -1) {
		delete(referenced, m[1])
	}
	if len(referenced) == 0 {
		return nil
	}

	orphans := make([]string, 0, len(referenced))
	for id := range referenced {
		orphans = append(orphans, id)
	}
	sort.Strings(orphans)
	return orphans
}
