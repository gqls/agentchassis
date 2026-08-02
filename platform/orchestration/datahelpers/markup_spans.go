// FILE: platform/orchestration/datahelpers/markup_spans.go
//
// ONE definition of "these bytes are not markup", for every regex in this estate
// that REWRITES HTML — bugs_open/180.
//
// THE DEFECT THIS FIXES. Every anchor regex here is a byte-level match over the
// whole input (`repairAnchorRe`, `anchorRe`, `deadAnchorRe` — the header of each
// one explains its `\b` and its non-greedy tail, and none of them can explain
// this). A regex that reads MARKUP cannot tell markup from a string literal that
// CONTAINS markup, so a tool that builds its own anchor at runtime:
//
//	<script>h = '<p>' + t + ' <a href="' + q.link + '">See guide</a>.</p>';</script>
//
// is read as an anchor whose href is EMPTY — `[^"']*` cannot cross the `'` that
// follows `href="` — and RepairPageLinks' unlink arm deletes it. The output is
// still valid JavaScript and still reads sensibly, so nothing throws and nobody
// sees it; the visitor simply cannot click. Measured on 2026-08-02: one live
// tool component had a working link removed from its own source this way.
//
// THE SPELLING IS NOT THE BUG, so this is not a denylist of `' +` and `${`. The
// same function unlinks a template literal (`href="${q.link}"` is a NON-empty
// href, so it takes the phantom arm instead), rewrites markup quoted inside a
// <style> comment, edits the visible text of a <textarea>, and repairs anchors
// that are COMMENTED OUT — four different arms, one cause. Fixing the cause once
// fixes every spelling, including the ones nobody has written yet.
//
// WHAT COUNTS AS NOT-MARKUP. The contents of the raw-text elements (script,
// style, textarea, title) and HTML comments. These are exactly the regions the
// HTML spec says a parser does not look for tags in, which is why the definition
// can be shared rather than argued per caller — and runtime_fill.go's scanner
// already had to know all of them, because a `<` inside a script string would
// otherwise open a phantom element and put every later span wrong. So this is
// one more question asked of a walk that was already being done, not a second
// parser (WHY A HAND-ROLLED SCANNER, and WHY NOT A DOM, are argued there).
//
// MASK, DO NOT MERELY FILTER, and this is the part that is easy to get subtly
// wrong. The obvious fix — match as before, then drop the matches that begin
// inside a span — leaves a second defect standing:
//
//	<script>var t = '<a href="/gone">';</script><a href="/gone">Pricing</a>
//
// The regex's non-greedy `</a>` closes at the REAL anchor, so ONE match spans
// from inside the script to the end of a genuine phantom link. Dropping it
// silently drops the real anchor with it, and `FindAll` never revisits those
// bytes. Masking the spans BEFORE matching removes the decoy and leaves the real
// anchor to be found on its own, with offsets that still index the original.
//
// WHY THE FILLER IS NUL AND NOT A SPACE. A same-length filler is what keeps the
// offsets usable, but the CHOICE of byte is load-bearing and a space is the
// wrong one: it is whitespace, so a pattern that opens with `\s` — like
// drop_dead_url_controls.go's `\ssrc\s*=\s*""` — can begin ON THE MASK. Masking
// `<style>x</style>src=""` with spaces manufactures the leading whitespace that
// the original did not have, and a mask that can invent a match is worse than no
// mask at all. NUL can begin none of these patterns: they all open with `<` or
// with `\s`, and NUL is neither. Newlines are preserved so a line-anchored
// caller behaves identically.
//
// The offset filter is kept anyway, as the INVARIANT rather than an argument
// about fillers: no match that begins inside a non-markup span is ever returned.
// For today's patterns it never fires — that is the point of choosing NUL — but
// it is what a future caller's pattern is held to, and it does not depend on
// anyone re-deriving the filler argument above.
//
// SCOPE, stated so the next adopter meets the RFC threshold rather than
// rediscovering it (the same test runtime_fill.go states for itself). This
// governs WRITERS: it is applied to RepairPageLinks, whose unlink arm destroys
// content. It is deliberately NOT applied to the detectors (links.go's phantom
// scan, check_dead_controls) in the same change: those escalate to a human, a
// false finding costs attention rather than content, and narrowing a judge is a
// judgement about what we want reported — 137 ruled on that direction for the
// runtime-fill marker and the same argument does not transfer automatically.
// The gap is real and is recorded in bugs_open/180 for the detection lane
// (bugs_open/097, bugs_open/116) rather than closed here by assumption.

package datahelpers

import (
	"regexp"
	"strings"
)

// NonMarkupSpans returns the byte ranges of html whose contents a browser never
// parses as markup: raw-text elements (script, style, textarea, title) and HTML
// comments, each taken WHOLE — start tag through close tag. Merged and sorted.
//
// It degrades WIDE on markup it cannot parse: an unclosed raw-text element, an
// unterminated comment or an unterminated quote inside a start tag makes the
// remainder of the input non-markup. For a writer that is the fail-safe
// direction — declining to rewrite bytes it cannot locate — and it is the
// opposite of what the same scan does for the runtime-fill exemption, for the
// reason argued at scanSpans.
func NonMarkupSpans(html string) ByteSpanSet {
	_, nonMarkup := scanSpans(html)
	return nonMarkup
}

// maskFiller is the byte non-markup bytes become. NUL, not a space: see the
// header — a whitespace filler can supply the `\s` an attribute pattern opens
// with and so invent a match the original document did not contain.
const maskFiller = 0x00

// maskNonMarkup returns a copy of html with every non-markup byte replaced by
// maskFiller, SAME LENGTH, so an offset into the result indexes the original.
// Newlines are preserved so a line-anchored pattern behaves identically.
//
// Returns html itself (no copy) when there is nothing to mask, so the common
// case pays one scan and no allocation.
func maskNonMarkup(html string) (masked string, spans ByteSpanSet) {
	spans = NonMarkupSpans(html)
	if len(spans) == 0 {
		return html, nil
	}
	b := []byte(html)
	for _, s := range spans {
		for i := s.Start; i < s.End && i < len(b); i++ {
			if b[i] != '\n' {
				b[i] = maskFiller
			}
		}
	}
	return string(b), spans
}

// MarkupMatches is FindAllStringSubmatchIndex for a pattern that is meant to
// match ELEMENTS: the returned offsets index html, but the matching is done
// against a copy in which the non-markup regions are masked, so markup quoted
// inside a script, a style, a textarea, a title or a comment can neither be
// matched nor swallow a real element that follows it. No returned match begins
// inside a non-markup span.
//
// Drop-in for the regex call it replaces. A caller slicing html by these offsets
// gets the ORIGINAL bytes, including any comment that sat inside an element's
// content — so a repair that keeps inner markup keeps it verbatim.
func MarkupMatches(re *regexp.Regexp, html string) [][]int {
	masked, spans := maskNonMarkup(html)
	ms := re.FindAllStringSubmatchIndex(masked, -1)
	return dropMatchesInSpans(ms, spans)
}

// dropMatchesInSpans enforces the invariant the filler choice is only expected
// to make unreachable: a match beginning inside a masked region is discarded.
func dropMatchesInSpans(ms [][]int, spans ByteSpanSet) [][]int {
	if len(spans) == 0 || len(ms) == 0 {
		return ms
	}
	out := ms[:0]
	for _, m := range ms {
		if spans.Contains(m[0]) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// ReplaceAllInMarkup is ReplaceAllString for a pattern meant to match ELEMENTS,
// with the same rule: matches that begin inside a non-markup region are left
// alone. repl is a literal replacement — no $1 expansion — because every caller
// here deletes what it matches, and expanding a template against masked bytes
// would emit spaces.
//
// Byte-identical output when nothing matches outside the masked regions, so a
// clean input is never perturbed.
func ReplaceAllInMarkup(re *regexp.Regexp, html, repl string) string {
	masked, spans := maskNonMarkup(html)
	ms := dropMatchesInSpans(re.FindAllStringIndex(masked, -1), spans)
	if len(ms) == 0 {
		return html
	}
	var out strings.Builder
	last := 0
	for _, m := range ms {
		out.WriteString(html[last:m[0]])
		out.WriteString(repl)
		last = m[1]
	}
	out.WriteString(html[last:])
	return out.String()
}
