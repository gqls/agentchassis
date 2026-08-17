// FILE: platform/orchestration/actions/component_instance_scope.go
//
// Per-instance scoping for components that carry element ids and scripts.
//
// WHY THIS EXISTS (bugs_open/283). An HTML id must be unique across the whole
// document: document.getElementById returns the FIRST match and silently
// ignores the rest. Most of our interactive components hardcode their ids as
// literal text and bind them with getElementById, so placing two of them on one
// page does not fail — it renders perfectly, accepts input, responds to its
// button, and reads the FIRST instance's fields while writing into the FIRST
// instance's results. On a consumer-credit site that is a page which can show a
// visitor a repayment figure computed from numbers they never entered: a
// plausible wrong answer, not an error.
//
// Measured 2026-08-15: 173 of 240 active components hardcode at least one
// element id (166 literal-only), 100 bind by getElementById, and across the 22
// active calculators one id — btn-calculate — is shared by NINE different
// components. So "list all the calculators on one page" collides between
// DIFFERENT components, not merely between two copies of one.
//
// THREE COLLISION CLASSES, not one. Namespacing ids fixes only the first:
//
//  1. element ids                — getElementById resolves to the first match
//  2. global function names      — the second <script>'s `function calc()`
//     replaces the first; both buttons then run
//     the same component's logic
//  3. window.onload assignment   — a SINGLE slot; the last assignment wins and
//     every earlier component never initialises
//
// Classes 2 and 3 have nothing to do with ids, which is why the fix is
// "scope the component", not "prefix the ids".
//
// THE TOKEN IS component FUNCTION + OCCURRENCE ON THE PAGE, AND THE CHOICE IS
// LOAD-BEARING. `c-mortgages-repayment`, then `c-mortgages-repayment-2` for a
// second copy of the same component on the same page.
//
//   - UNIQUE WITHIN A PAGE. The function separates two DIFFERENT components
//     (the class that actually bites: one id, btn-calculate, is shared by nine
//     different calculators), the occurrence separates two copies of one.
//   - THE SAME ON EVERY PAGE for a single-instance component, which is the
//     property every hand-written selector needs and the only one that is
//     expensive to get wrong. loanandmortgagecalculator's oracle.py addresses
//     all 170 of its checks by literal CSS id; with this rule it needs one
//     prefix per tool, and with any page-varying rule it needs per-page
//     knowledge of every tool.
//   - STABLE ACROSS RE-RENDERS, because it is derived from stored, ordered
//     data rather than recomputed. An unstable token would change the page
//     bytes on every rerender and destroy the byte-identical property the site
//     lanes verify against.
//
// TWO RULES WERE TRIED AND REJECTED — both are in the council trail, and both
// look correct until you ask what a selector must know:
//
//   - `position` (shipped 2026-08-15, superseded 2026-08-16 before any
//     template consumed it). It IS unique per page — measured, zero duplicates
//     fleet-wide — but it is not the same across pages: measured, the LMC tool
//     slot sits at position 0 on 7 pages and position 1 on the other 16, so
//     one component would answer to two different ids depending on the page,
//     and every selector would be coupled to section order.
//   - `page_components.data_uuid` (raised by the council's prior_art seat;
//     1,580 rows, 1,580 distinct, so uniqueness is provable rather than
//     derived). Rejected for the same reason, harder: the id then differs per
//     page AND is opaque, so no selector can be written without a lookup.
//
// The cost of this rule is that uniqueness is DERIVED (from the page's ordered
// section list) rather than read off a unique column. That is why
// DetectInstanceCollisions exists and why the paths that cannot see the whole
// page say so — see InstanceCounter.
//
// DO NOT reuse {{.ComponentID}} for this. It is bound to the CONTENT COMPONENT
// row id on two of the three render paths (rerender_page_sections_action.go and
// v3_site_actions.go), i.e. the same value for every instance of a component,
// while assemble_from_library.go builds a per-instance string. Same placeholder
// name, different guarantees; adopting it for reuse namespaces nothing on the
// path that actually serves re-rendered pages. Full trap in LANDMINES.md.
package actions

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// InstanceContentKey is the render-context key every path binds the token to.
// Named once so no call site spells it, and so a grep for the key finds every
// producer.
const InstanceContentKey = "InstanceID"

var reNotSelectorSafe = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

// InstanceToken returns the namespace token for the `occurrence`-th instance of
// the component whose function is `function` on one page. Occurrence is
// zero-based; the first instance takes the bare token so the common case — a
// component appearing once, which today is EVERY interactive component on every
// live page — reads as `c-mortgages-repayment` rather than `c-…-1`.
//
// Prefixed with a letter deliberately: an id beginning with a digit is a valid
// HTML id and getElementById finds it, but it is NOT a valid CSS identifier, so
// querySelector("#3-foo") throws. Keeping the token selector-safe means a
// component may use either lookup.
func InstanceToken(function string, occurrence int) string {
	s := strings.ToLower(strings.TrimSpace(function))
	s = strings.Trim(reNotSelectorSafe.ReplaceAllString(s, "-"), "-")
	if s == "" {
		// A component with no function is not addressable by name; fall back to
		// the occurrence alone rather than emitting a bare "c-", which would be
		// the SAME token for every such component on the page.
		s = fmt.Sprintf("anon-%d", occurrence)
	}
	if occurrence <= 0 {
		return "c-" + s
	}
	return fmt.Sprintf("c-%s-%d", s, occurrence+1)
}

// InstanceCounter assigns occurrence indices as a render loop walks a page's
// sections in position order. It is THE canonical derivation: every path that
// can see the whole page uses this, so two paths rendering the same page agree
// on every token rather than agreeing only by coincidence.
//
// A path that CANNOT see the whole page — RenderComponentAction and the section
// editor each render one section, and during a build the page's rows may not
// exist yet — must not invent a second rule. It calls InstanceToken with
// occurrence 0 and says so at the call site. That is a possibly-wrong INPUT to
// one rule, not a second rule with a weaker guarantee, and the difference is
// what makes it detectable: a wrong occurrence collides, and a collision is
// exactly what DetectInstanceCollisions reports.
type InstanceCounter struct {
	seen map[string]int
}

// NewInstanceCounter returns a counter for one page's render pass.
func NewInstanceCounter() *InstanceCounter {
	return &InstanceCounter{seen: make(map[string]int)}
}

// Next returns the token for the next instance of `function` on this page and
// advances the count. Call it once per rendered section, in position order.
func (c *InstanceCounter) Next(function string) string {
	if c.seen == nil {
		c.seen = make(map[string]int)
	}
	key := strings.ToLower(strings.TrimSpace(function))
	tok := InstanceToken(function, c.seen[key])
	c.seen[key]++
	return tok
}

// InstanceTokensForPage assigns tokens to an ordered list of component
// functions. Same rule as InstanceCounter, for callers holding the whole list
// at once (and the shape the tests assert against).
func InstanceTokensForPage(functions []string) []string {
	c := NewInstanceCounter()
	out := make([]string, len(functions))
	for i, fn := range functions {
		out[i] = c.Next(fn)
	}
	return out
}

// BindInstanceToken puts `token` on the render context under the one key every
// template reads. Every producer goes through here so that a grep for the
// binding finds all of them, and so no call site spells the key.
func BindInstanceToken(rc *RenderContext, token string) {
	if rc == nil {
		return
	}
	if rc.ContentData == nil {
		rc.ContentData = make(map[string]interface{})
	}
	rc.ContentData[InstanceContentKey] = token
}

// BindSingleSectionInstanceToken binds the token for a path that renders ONE
// section and cannot see the rest of the page — the section editor, and
// RenderComponentAction during a build where the page's rows may not exist yet.
//
// It supplies occurrence 0 to the canonical rule rather than inventing a second
// rule for the single-section case. Right whenever the component appears once on
// the page, which is every interactive component on every live page today
// (measured 2026-08-15: no component that binds by getElementById is
// instantiated twice anywhere). Where it is wrong, the two instances take the
// same token and DetectInstanceCollisions reports it at assembly — a detectable
// wrong answer, not a silent one, and not a second guarantee.
func BindSingleSectionInstanceToken(rc *RenderContext, function string) {
	BindInstanceToken(rc, InstanceToken(function, 0))
}

// reTemplateInstanceID matches a template's reference to the per-instance
// token, in either of Go's spacings and with or without a trim marker.
var reTemplateInstanceID = regexp.MustCompile(`\{\{-?\s*\.` + InstanceContentKey + `\b`)

// TemplateNeedsInstanceID reports whether a template namespaces anything with
// the per-instance token, i.e. whether rendering it without one is a defect.
//
// This is the predicate the SHARED render layer uses, and it is why the fix is
// not "patch the three call sites the council saw". Measured 2026-08-16: EIGHT
// non-test files hold FOURTEEN calls to a RenderTemplate* helper. Three files
// bound a token before this change; of the five that did not, four are
// single-instance by construction (chrome, <head>, and an offline lint) and one
// — the section editor, on nobody's list including the council's — renders a
// page-embedded component and bound nothing at all. A template rendered by that
// path gets missingkey=zero, an empty string, so every instance lands back on
// identical ids, silently, which is the precise failure this seam removes.
// scripts/pattern-check.py's check_unscoped_component_render is the durable
// half: the set of call sites grows, and a census written by reading does not.
func TemplateNeedsInstanceID(templateStr string) bool {
	return reTemplateInstanceID.MatchString(templateStr)
}

var (
	// Attribute ids only. Deliberately does not match {{...}} expressions: an
	// unrendered template is not evidence of a collision.
	reElementID = regexp.MustCompile(`\sid="([^"{}]+)"`)
	// Inline scripts only — a <script src=...> is a shared file, not a
	// per-instance body, and is expected to be present once per component.
	reInlineScript = regexp.MustCompile(`(?s)<script(?:\s[^>]*)?>(.*?)</script>`)
	reScriptHasSrc = regexp.MustCompile(`(?i)<script[^>]*\ssrc=`)
	reWindowOnload = regexp.MustCompile(`window\.onload\s*=`)
	// A component whose script body is wrapped in an IIFE or a block-scoped
	// listener declares nothing globally. These are the accepted wrappers.
	reIIFE = regexp.MustCompile(`(?s)^\s*(?:!|\(|void\s)?\s*(?:function\s*\(|\(\s*\)\s*=>|async\s+function\s*\()`)
	// Leading comments before the wrapper. reIIFE anchors at the body's start,
	// and the estate's tool templates conventionally open with a /* tool-doc */
	// block — measured 2026-08-17 (cmd/instanceaudit over the 91 live
	// getElementById templates), 62 of them are IIFE-wrapped behind exactly that
	// comment and were being reported unscoped. A false flag here is not
	// harmless conservatism once the guard is armed: it would refuse renders of
	// correct components, and at conversion time it would send 62 sound scripts
	// into the judged-rewrite pool. Only LEADING comments are skipped — comments
	// elsewhere cannot sit between "start of body" and "the wrapper".
	reLeadLineComment  = regexp.MustCompile(`^\s*//[^\n]*(?:\n|$)`)
	reLeadBlockComment = regexp.MustCompile(`(?s)^\s*/\*.*?\*/`)
)

// stripLeadingJSComments removes line and block comments from the FRONT of a
// script body, so the accepted-wrapper test judges the first statement rather
// than the first byte.
func stripLeadingJSComments(body string) string {
	for {
		if m := reLeadLineComment.FindString(body); m != "" {
			body = body[len(m):]
			continue
		}
		if m := reLeadBlockComment.FindString(body); m != "" {
			body = body[len(m):]
			continue
		}
		return body
	}
}

// instanceScopeEnforceConfigKey arms the refusal. Shape and naming mirror
// dead_url_guard.go's record/refuse pair deliberately — one idiom for "this
// guard is armed", not a second one.
const instanceScopeEnforceConfigKey = "enforce_instance_scope"

// enforceInstanceScope reports whether a collision should REFUSE the render
// rather than merely be recorded.
//
// OPT-IN, AND THE UNSAFE DEFAULT IS DELIBERATE (owner ruling 2026-08-02 §2: new
// authority on a shared seam ships as an opt-in field with the unsafe side as
// the default). Two reasons it cannot default on:
//
//   - Pages ALREADY collide today. `generic-text-block` resolves its one id
//     through {{.ComponentID}}, which is the shared component row id on this
//     path, so the 13 active pages carrying it two or three times already emit
//     duplicate ids. Defaulting to refuse would fail their next re-render —
//     turning a latent defect into an outage, on pages nobody has asked us to
//     change.
//   - The detector is a regex scan and errs toward reporting. A false positive
//     that writes a line in a result is cheap; one that refuses a re-render is
//     not.
//
// So the sequence is: record everywhere, read the numbers, convert the
// components, then arm per-workflow. Recording is unconditional precisely so
// the arming decision is made against measurements rather than a guess.
func enforceInstanceScope(config map[string]interface{}) bool {
	armed, _ := config[instanceScopeEnforceConfigKey].(bool)
	return armed
}

// InstanceCollisions reports why a rendered page is not safe to carry two
// instances of a component. Each field names a distinct failure class; a page
// can be clean on ids and still broken on window.onload.
type InstanceCollisions struct {
	// DuplicateElementIDs are ids appearing more than once in the document,
	// sorted. Every getElementById for one of these resolves to the first.
	DuplicateElementIDs []string
	// WindowOnloadAssignments counts `window.onload =` across inline scripts.
	// More than one means all but the last component never initialises.
	WindowOnloadAssignments int
	// UnscopedInlineScripts counts inline script bodies that are NOT wrapped in
	// an IIFE, i.e. whose declarations land in global scope where a second
	// instance (or a different component) can replace them by name.
	UnscopedInlineScripts int
}

// Clean reports whether the document is safe to carry repeated components.
func (c InstanceCollisions) Clean() bool {
	return len(c.DuplicateElementIDs) == 0 &&
		c.WindowOnloadAssignments <= 1 &&
		c.UnscopedInlineScripts == 0
}

// Summary renders the report for a log line or a work-item result. Empty string
// when clean, so callers can treat it as the "why it was refused" message.
func (c InstanceCollisions) Summary() string {
	if c.Clean() {
		return ""
	}
	var parts []string
	if n := len(c.DuplicateElementIDs); n > 0 {
		shown := c.DuplicateElementIDs
		if len(shown) > 8 {
			shown = shown[:8]
		}
		parts = append(parts, fmt.Sprintf("%d duplicate element id(s): %s",
			n, strings.Join(shown, ", ")))
	}
	if c.WindowOnloadAssignments > 1 {
		parts = append(parts, fmt.Sprintf(
			"%d window.onload assignments (only the last one runs)",
			c.WindowOnloadAssignments))
	}
	if c.UnscopedInlineScripts > 0 {
		parts = append(parts, fmt.Sprintf(
			"%d inline script(s) declaring into global scope", c.UnscopedInlineScripts))
	}
	return strings.Join(parts, "; ")
}

// DetectInstanceCollisions scans an assembled page (or any HTML fragment) for
// the three classes above.
//
// LIMITATION, stated rather than hidden: this is a regex scan, not an HTML
// parse. An id inside a comment or a string literal counts, and a script body
// whose wrapper this does not recognise reads as unscoped. Both err toward
// REPORTING a collision rather than missing one, which is the safe direction
// for a guard — but it means a non-empty report is a reason to look, and a
// caller must not treat it as proof on its own.
func DetectInstanceCollisions(html string) InstanceCollisions {
	var out InstanceCollisions

	seen := make(map[string]int)
	for _, m := range reElementID.FindAllStringSubmatch(html, -1) {
		seen[m[1]]++
	}
	for id, n := range seen {
		if n > 1 {
			out.DuplicateElementIDs = append(out.DuplicateElementIDs, id)
		}
	}
	sort.Strings(out.DuplicateElementIDs)

	for _, m := range reInlineScript.FindAllStringSubmatch(html, -1) {
		whole, body := m[0], strings.TrimSpace(m[1])
		if reScriptHasSrc.MatchString(whole) || body == "" {
			continue
		}
		out.WindowOnloadAssignments += len(reWindowOnload.FindAllString(body, -1))
		if !reIIFE.MatchString(stripLeadingJSComments(body)) {
			out.UnscopedInlineScripts++
		}
	}
	return out
}
