// FILE: platform/orchestration/actions/section_visible_text.go
//
// visibleTextLength — how much text a READER would see in a stored section,
// with stylesheet and script CONTENT excluded. A thin count over the estate's
// existing parser-based extractor (datahelpers.VisibleTextFromHTML); the value
// of this file is the CALIBRATION below, not the extraction.
//
// WHY THIS EXISTS (bugs_closed/285's delivery half, measured 2026-08-17).
// The per-slot text floor (bugs_open/178, wired into the section editor by
// bugs_open/253's council round) measured `<[^>]*>` -stripped length. That
// strips TAGS but not what is INSIDE <style> and <script>, so CSS declarations
// and JavaScript source count as "text". On the write that emptied
// webdesign.co.uk/learn/ai-builders/content-first.html, that axis read:
//
//	                     tag-stripped        visible (this measure)
//	existing (article)    3,236 chars         2,143 chars
//	incoming (poison)     8,491 chars ↑262%       16 chars ↓0.7%
//
// The floor is 50%, so the tag-stripped axis reported 262% retained and the
// write proceeded: the article body was replaced by a wrapper stylesheet plus
// an empty <article>, and the page served that for ~23 hours. The class-attribute
// floor (bugs_open/253) also passed — 3 → 4 class attributes, 133%. Nothing on
// the path could see it, because both axes grew.
//
// CALIBRATION — every archived overwrite pair, run through THIS implementation
// (cmd-style harness over `page_component_history`, migration 357's archive,
// live since 2026-08-09): 117 pairs.
//
//	axis                        pairs refused
//	tag-stripped (the old one)        1   — and it is a legitimate REPAIR
//	visible text (this measure)       3   — three real hollowings, listed below
//	both                              0   — the axes never agree on a pair
//
// The three the visible axis refuses, all independently known incidents:
//
//	2026-08-11 14:05  idea.uk / tool-ab-test-calculator      visible  684 → 0
//	                  (tag-stripped GREW 10,399 → 12,929)
//	2026-08-14 18:51  webdesign.co.uk / learn-ai-builders-…  visible 2,143 → 16
//	                  (tag-stripped GREW 3,236 → 8,491 — bugs_closed/285)
//	2026-08-15 18:24  webdesign.co.uk / tool-ab-test-calc…   visible  684 → 0
//	                  (bugs_open/286's hollow fork: 47 raw tags served)
//
// Two of the three are the webdesign_tool_rebuilds lane's tool-hollowing case,
// which asked for exactly this measure from the other end — so the axis is not
// tuned to one incident.
//
// The ONE the old axis refuses is putting the article BACK (2026-08-15 18:18,
// seed 431): 16 → 2,143 visible, but 8,491 → 3,236 tag-stripped = 38% kept, so
// the live floor would have REFUSED THE REPAIR had it gone through the editor
// rather than a direct restore. That is why this is a correction of the axis and
// not an additional floor with an OR: enforcing both would block the fix.
// "A guard that refuses good work gets switched off, and then it protects
// nothing" (component_write_guard.go's header, the same lesson).
//
// THE MINIMUM IS ALREADY THERE — not a number this file adds. evaluateSectionShrink
// skips any slot whose EXISTING side is under the minimum its caller passes, so
// short captions and param blobs are out of scope without a second rule. An earlier
// draft added its own 120-char minimum; a mutation run (delete the clause, see what
// fails) showed it was dead code, because the shared minimum dominated it. It was
// deleted rather than kept as decoration.
//
// > CORRECTED 2026-08-17 (bugs_open/293): that minimum was 500 and a hardcoded
// > constant when this was written. It is now minShrinkGuardVisibleChars (200),
// > passed as a parameter — this file's own closing paragraph predicted both the
// > need and the shape ("the fix is a lower minimum for this axis — which needs
// > evaluateSectionShrink to take it as a parameter"), and 293's sweep supplied the
// > number. 500 was calibrated against TAG-STRIPPED lengths, so carrying it onto
// > this axis quietly excluded ordinary prose: on the rebuild population it left
// > 587 of 1,079 slots unjudged, and on this path 110 of 263.
//
// WHAT THE AXIS CHANGE LOOSENS, measured rather than waved at. Applying the 500
// rule to VISIBLE text leaves 74 of the 117 pairs in scope, so 43 (37%) fall
// out — slots that are mostly stylesheet, script or code. Two facts make that
// acceptable, and both are checkable:
//   - the old axis refused 0 of those 43, and on a CSS-dominated slot its ratio
//     is dominated by the stylesheet, so it cannot see prose deletion there —
//     which is this bug's whole finding;
//   - structural collapse on those slots is still covered by the class-attribute
//     floor (bugs_open/253, unchanged, its own minimum of 10 class attributes).
//
// Of the 198 live rows this protects, 180 have ≥500 visible chars by the SQL
// approximation; the parser measure is stricter, so treat 180 as an upper bound.
// If a defect ever turns up in that 37%, the fix is a lower minimum for this
// axis — which needs evaluateSectionShrink to take it as a parameter, not a
// second copy of the ratio rule here.
//
// SCOPE — no longer narrow, as of 2026-08-17.
//
// > CORRECTED 2026-08-17 (bugs_open/293), and the correction is the whole point of
// > that bug. This paragraph used to read: "This is used by the SECTION-EDITOR call
// > site (single_slot_floors.go) only. The whole-page save path
// > (save_sections_shrink_guard.go) still measures the tag-stripped axis: its writes
// > are DELETE+INSERT rebuilds, which the archive records as delete rows with no
// > successor to pair against — 3,603 of them against the 281 overwrite rows — so
// > there is no pair evidence for that path."
// >
// > **The successor was never missing.** It is the LIVE page_components row, and
// > page_components.created_at proves the re-insert belongs to the rebuild that had
// > just deleted its predecessor: 1,123 of 1,254 within 60 s, 1,109 within 5 s, and
// > not one live row older than its own last delete. That yields 1,079 exactly-paired
// > rebuild writes — nine times this file's original evidence base — and the same
// > rule reproduces the three refusals above with identical figures when applied to
// > the overwrite rows, which is the control that earned it.
//
// All three text floors now measure with this function: the per-slot whole-page
// guard and the page-total floor (both in save_sections_shrink_guard.go) and the
// section editor (single_slot_floors.go). What holds a fourth caller to it is
// shrink_axis_coverage_test.go, not this comment.
//
// One consumer of the RETIRED axis remains, deliberately:
// load_current_section_content_action.go:262 uses shrinkGuardTagStripper with
// minShrinkGuardChars (500) to judge whether an unclaimed stored slot is
// "prose-sized" enough to be a plausible body-text match. It refuses nothing, so it
// is a pairing heuristic rather than a floor, and re-tuning it would change which
// slots get paired with no calibration for that decision.
package actions

import (
	"unicode"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// visibleTextLength returns the number of non-whitespace characters a reader
// would see in a stored section.
//
// REUSE, not a new stripper (council 3279156b, reuse_agent seat — it asked
// whether the estate already had a "reduce HTML to what a reader sees" helper
// before this file added one, and it was right that it does).
// datahelpers.VisibleTextFromHTML → ExtractAssertionText walks a real parsed
// document (golang.org/x/net/html) and drops script, style, noscript, template,
// code, pre, svg, iframe, textarea, select, option and head with their content,
// decoding entities on the way. That is the same question the claims and
// voice-tell scans ask of a page, so the floor and those scans now agree by
// construction instead of by coincidence — and a regex chain that had to be
// kept in step with them is gone.
//
// The one behavioural difference worth knowing: it also excludes <code>/<pre>,
// so a section whose content IS a code sample measures small and falls under
// evaluateSectionShrink's 500-char minimum, i.e. out of scope. Symmetric on both
// sides, so it can never manufacture a refusal — it can only decline to judge.
//
// Whitespace is dropped rather than collapsed so the count is a quantity of
// CHARACTERS a reader sees, not of layout: the two sides are measured
// identically, which is all a ratio needs.
func visibleTextLength(html string) int {
	n := 0
	for _, r := range datahelpers.VisibleTextFromHTML(html) {
		if !unicode.IsSpace(r) {
			n++
		}
	}
	return n
}
