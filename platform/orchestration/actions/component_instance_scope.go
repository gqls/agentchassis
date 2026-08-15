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
// THE TOKEN IS DERIVED FROM position, AND THAT CHOICE IS LOAD-BEARING:
//
//   - page_components.slot_name is NOT unique per page — 13 active pages carry
//     duplicate slot names, and the only unique index
//     (uq_page_components_no_byte_identical_duplicate) merely forbids
//     BYTE-IDENTICAL rows, so two instances with different copy are explicitly
//     allowed to share a slot. slot_name therefore cannot namespace anything.
//   - page_components.position IS unique per page — measured across every
//     active page fleet-wide, zero duplicates — and it is already loaded by
//     the rerender path, so no schema or query change is needed.
//   - position is STABLE across re-renders (it is stored, not recomputed),
//     which matters because an unstable token would change the page bytes on
//     every rerender and destroy the byte-identical property the site lanes
//     verify against.
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

// InstanceToken returns the per-instance namespace token for a component at
// `position` on a page. Prefixed with a letter deliberately: an id beginning
// with a digit is a valid HTML id and getElementById finds it, but it is NOT a
// valid CSS identifier, so querySelector("#3-foo") throws. Keeping the token
// selector-safe means a component may use either lookup.
func InstanceToken(position int) string {
	return fmt.Sprintf("c%d", position)
}

var reNotSelectorSafe = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

// InstanceTokenFromSlot derives a token where the page position is NOT in
// scope — currently only RenderComponentAction, which renders one section at a
// time and knows its slot name but not its index on the page.
//
// BEST-EFFORT, AND THE WEAKNESS IS THE POINT: slot names are positional in
// practice ("prose-0", "tool-1") and then this is as good as InstanceToken, but
// slot_name is NOT unique per page — 13 active pages carry repeated slot names
// (all of them a component repeated under one slot). On those, this returns the
// same token for every instance and namespaces nothing. That case is caught by
// DetectInstanceCollisions at the guard, not here: a token function cannot know
// what else is on the page. Never present this as a uniqueness guarantee.
//
// It exists so the value is never ABSENT: templates render with missingkey=zero,
// so an unset InstanceID would silently become "" and put every instance back on
// identical ids — failing in exactly the invisible way this whole change exists
// to remove. A wrong-but-present token fails loudly at the guard; an empty one
// does not fail at all.
func InstanceTokenFromSlot(slot string) string {
	s := reNotSelectorSafe.ReplaceAllString(strings.TrimSpace(slot), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "c0"
	}
	return "c-" + s
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
)

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
		if !reIIFE.MatchString(body) {
			out.UnscopedInlineScripts++
		}
	}
	return out
}
