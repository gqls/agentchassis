// FILE: platform/orchestration/actions/section_visible_text.go
//
// visibleTextLength — how much text a READER would see in a stored section,
// with stylesheet and script CONTENT excluded.
//
// WHY THIS EXISTS (bugs_closed/285's delivery half, measured 2026-08-17).
// The per-slot text floor (bugs_open/178, wired into the section editor by
// bugs_open/253's council round) measured `<[^>]*>` -stripped length. That
// strips TAGS but not what is INSIDE <style> and <script>, so CSS declarations
// and JavaScript source count as "text". On the write that emptied
// webdesign.co.uk/learn/ai-builders/content-first.html, that axis read:
//
//	                     tag-stripped        visible (this helper)
//	existing (article)    3,245 chars         2,754 chars
//	incoming (poison)     8,492 chars ↑262%       68 chars ↓2%
//
// The floor is 50%, so the tag-stripped axis reported 262% retained and the
// write proceeded: the article body was replaced by a wrapper stylesheet plus
// an empty <article>, and the page served that for ~23 hours. The class-attribute
// floor (bugs_open/253) also passed — 3 → 4 class attributes, 133%. Nothing on
// the path could see it, because both axes grew.
//
// CALIBRATION — every archived overwrite pair, not a hand-picked case.
// `page_component_history` (migration 357's archive, live since 2026-08-09)
// paired each overwritten row with what replaced it: 117 pairs.
//
//	axis                        pairs refused   what it refused
//	tag-stripped (the old one)        1          a legitimate REPAIR (see below)
//	visible text (this helper)        1          the poisoning write, and only it
//
// The two are DIFFERENT pairs — the axes disagree in both directions, which is
// why this is a correction and not an additional floor:
//
//   - the poisoning write (2026-08-14 18:51) passes the old axis at 262% and is
//     refused by this one at 2%;
//   - putting the article BACK (2026-08-15 18:18, seed 431) reads 38% on the old
//     axis — i.e. the old axis would have REFUSED THE REPAIR had it gone through
//     the editor rather than a direct restore — and reads 4,050% on this one.
//
// A guard that refuses good work gets switched off, and then it protects
// nothing (component_write_guard.go's header, the same lesson).
//
// THE MINIMUM IS ALREADY THERE, AND IT IS 500 — not a number this file adds.
// evaluateSectionShrink skips any slot whose EXISTING side is under
// minShrinkGuardChars (500), so short captions and param blobs are out of scope
// without a second rule. An earlier draft of this change added its own 120-char
// minimum; a mutation run (delete the clause, see what fails) showed it was dead
// code, because 500 dominates it. It was deleted rather than kept as decoration.
//
// WHAT THE AXIS CHANGE LOOSENS, measured rather than waved at. Applying the 500
// rule to VISIBLE text instead of tag-stripped text takes 31 of the 117 archived
// pairs (26%) OUT of the text floor's scope: slots with ≥500 tag-stripped chars
// but <500 visible ones — i.e. mostly stylesheet and script. Two facts make that
// acceptable, and both are checkable:
//   - the old axis refused 0 of those 31, and on a CSS-dominated slot its ratio
//     is dominated by the stylesheet, so it cannot see prose deletion there —
//     which is this bug's whole finding;
//   - structural collapse on those slots is still covered by the class-attribute
//     floor (bugs_open/253, unchanged, its own minimum of 10 class attributes).
//
// Of the 198 live rows this protects, 180 have ≥500 visible chars.
// If a future defect turns up in that 26%, the fix is a lower minimum for this
// axis — which needs evaluateSectionShrink to take it as a parameter, not a
// second copy of the ratio rule here.
//
// SCOPE, deliberately narrow. This is used by the SECTION-EDITOR call site
// (single_slot_floors.go) only. The whole-page save path
// (save_sections_shrink_guard.go) still measures the tag-stripped axis: its
// writes are DELETE+INSERT rebuilds, which the archive records as delete rows
// with no successor to pair against — 3,603 of them against the 281 overwrite
// rows — so there is no pair evidence for that path, and changing an axis
// fleet-wide on evidence that does not cover it is how a guard starts refusing
// good work. Recorded as an open question in bugs_open/293, with the query.
package actions

import (
	"regexp"
	"strings"
)

var (
	// Non-greedy, dot-matches-newline, case-insensitive: <style>/<script>
	// bodies are removed WITH their content, which is the whole point.
	visibleTextStyleBlock  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	visibleTextScriptBlock = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	// Comments go too: an HTML comment is not reader-visible, and the poisoning
	// write carried one ("<!-- .asset-row anchors are seeded via JS below -->").
	visibleTextComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	visibleTextTag     = regexp.MustCompile(`<[^>]*>`)
	visibleTextEntity  = regexp.MustCompile(`&[a-zA-Z#0-9]+;`)
	visibleTextSpace   = regexp.MustCompile(`\s+`)
)

// visibleTextLength returns the number of non-whitespace characters a reader
// would see: style, script and comment blocks removed with their content, then
// tags, then entities, then whitespace collapsed away.
//
// Entities are dropped rather than decoded because this is a QUANTITY, compared
// against another measured the same way — `&nbsp;` decoding to one space or to
// nothing changes both sides identically, and decoding would need an HTML
// parser this package does not otherwise depend on.
func visibleTextLength(html string) int {
	s := visibleTextStyleBlock.ReplaceAllString(html, "")
	s = visibleTextScriptBlock.ReplaceAllString(s, "")
	s = visibleTextComment.ReplaceAllString(s, "")
	s = visibleTextTag.ReplaceAllString(s, "")
	s = visibleTextEntity.ReplaceAllString(s, "")
	s = visibleTextSpace.ReplaceAllString(s, "")
	return len(strings.TrimSpace(s))
}
