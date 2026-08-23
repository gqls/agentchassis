// FILE: platform/orchestration/datahelpers/link_suppress.go
//
// OUTBOUND suppression of links to pages that exist but would 404 — bugs_open/328.
//
// links.go answers "is this href an internal page link". link_repair.go answers
// "the target does not exist, so what do we do before the page ships". This file
// answers a third question that neither could: **the target row exists and is
// perfectly valid, and it still would not serve.**
//
// It is a separate file for the reason link_repair.go's own header gives for
// being separate from links.go: a repair pass is a behaviour change that should
// not arrive as a passenger on an edit to the shared vocabulary.
//
// WHY THIS IS NOT A WIDENING OF RepairPageLinks' INDEX. That index is loaded by
// ONE function (loadValidPagePaths) and consumed by four seams with different
// write surfaces — the deploy gate, the outbound rerender repair, the
// single-component repair, and the persistence chokepoint, which writes both
// page_components.rendered_html AND content_data. Teaching that index about
// deployment would change all four in one edit, including the two that write to
// the database. So this is a separate, strictly narrower set, consumed only by
// the outbound seams. The owner's 2026-08-02 ruling (RFC_010 §2) is the reason
// it is also opt-in at the caller rather than on by construction.
//
// WHY OUTBOUND ONLY, AND WHAT IT BUYS. Nothing here touches content_data. The
// author's href survives in the source the platform re-renders from, so on the
// first render after the target ships the anchor comes back by itself — no
// cascade, no repair queue, no work item, and no second mechanism that has to
// keep pace. That is the whole difference between this and bugs_open/079 / 097,
// whose targets genuinely do not exist and are correctly repaired at source.
// It also means check_phantom_internal_links, which reads STORED rendered_html,
// still sees the authored anchor: this fix silences no detector.
//
// THE TWO ARMS, AND WHY A PLAIN UNLINK WOULD NOT DO. There is a standing
// landmine on RepairPageLinks' unlink arm: "a dead internal link is REPAIRED
// into orphaned prose … where the anchor's inner content is the link LABEL —
// which is every card, CTA and 'read more' control on this platform — the served
// page shows the label and the arrow glyph as bare text in the middle of the
// card." Measured over all 36 affected anchors fleet-wide on 2026-08-23, the
// population splits: **28 are classless prose anchors** inside a sentence, where
// unlinking is correct and invisible, and **8 are classed template controls**
// whose inner content is a short label plus an arrow glyph. So:
//
//	prose   (no class attribute)          -> unlink: drop the <a>, keep the text
//	control (class + label <= 60 chars)   -> drop the whole element
//
// The control arm is DropDeadURLControls' rule for chrome (LNK-005,
// correct-or-absent), applied to the one body case that has the same shape: the
// card keeps its headline and body and loses its action, and label and arrow go
// together because both live inside the anchor. Requiring BOTH a class and a
// short label means an ambiguous anchor falls back to unlink, which is the
// fail-safe direction link_repair.go names for a writer — under-repair leaves a
// visible, still-flagged defect; over-repair deletes prose nobody can recover.
package datahelpers

import (
	"strings"
)

// LinkRepairSuppress and LinkRepairDropControl are this file's actions, beside
// LinkRepairRewrite / LinkRepairUnlink in link_repair.go. They are DISTINCT
// values rather than a reuse of LinkRepairUnlink because they answer a different
// question about the same page, and the recorder rows are read by people asking
// "why is this link not on the page" — "the target does not exist" and "the
// target exists and has never shipped" want opposite follow-up actions.
const (
	LinkRepairSuppress      = "suppress_unshipped"     // <a> dropped, inner text kept (prose)
	LinkRepairDropControl   = "drop_control_unshipped" // whole element dropped (a labelled control)
	suppressControlLabelMax = 60                       // chars of stripped inner text; above this, treat as prose
)

// SuppressRefusedPageLinks removes the internal page anchors in html whose target
// is a member of refused, returning the rewritten markup and an account of every
// change. Output is byte-identical to the input when nothing matched.
//
// refused holds ONLY targets the caller has established would 404 and are not
// arriving (PageLinkRefusedPredicateFor). An empty or nil set is a no-op, and
// that is load-bearing: an empty set means "nothing was refused", and the caller
// is required to have already turned a FAILED lookup into a nil set plus a
// logged skip rather than an empty one — the same contract loadValidPagePaths
// states for its bool.
//
// Only LinkScopePage hrefs are considered. External, mailto:/tel:, #fragment,
// asset and empty hrefs are returned untouched: empty hrefs belong to
// RepairPageLinks, which has already run at every caller of this function, and
// judging them twice would double-count the repair record.
func SuppressRefusedPageLinks(html string, refused PageURLSet) (string, []LinkRepair) {
	if html == "" || len(refused) == 0 {
		return html, nil
	}

	// Whole-input, deliberately, and for the reason link_repair.go argues at
	// length: this is a WRITER, not a judge. A runtime-fill shell's hrefs are
	// hydrated client-side, so the set cannot judge them, and the safe direction
	// for a writer is to touch nothing.
	if HasRuntimeFillMarker(html) {
		return html, nil
	}

	// MarkupMatches, never FindAllStringSubmatchIndex — an <a ...> quoted inside
	// a <script>, <style>, <textarea> or comment is TEXT, and both arms below
	// delete markup (bugs_open/180, which cost a live tool the anchor it builds
	// at runtime).
	matches := MarkupMatches(repairAnchorRe, html)
	if len(matches) == 0 {
		return html, nil
	}

	var out strings.Builder
	var repairs []LinkRepair
	last := 0

	for _, m := range matches {
		fullStart, fullEnd := m[0], m[1]
		hrefStart, hrefEnd := m[2], m[3]
		innerStart, innerEnd := m[4], m[5]
		href := html[hrefStart:hrefEnd]

		if ClassifyLinkScope(href) != LinkScopePage || !refused.Contains(href) {
			continue // not a page link, or the target serves — leave byte-identical
		}

		out.WriteString(html[last:fullStart])
		action := LinkRepairSuppress
		if anchorIsLabelledControl(html[fullStart:fullEnd], html[innerStart:innerEnd]) {
			// Drop the whole element: label and arrow glyph together.
			action = LinkRepairDropControl
		} else {
			out.WriteString(html[innerStart:innerEnd])
		}
		last = fullEnd
		repairs = append(repairs, LinkRepair{Href: href, Action: action})
	}

	if len(repairs) == 0 {
		return html, nil
	}
	out.WriteString(html[last:])
	return out.String(), repairs
}

// anchorIsLabelledControl reports whether this anchor is a template control (a
// card link, a CTA button) rather than a link inside prose.
//
// Both conditions must hold. A class attribute alone is not enough: a template
// could legitimately class a prose link, and dropping that would delete the
// sentence's words. A short inner text alone is not enough either: "see the
// guide" mid-paragraph is short and is prose. Requiring both means the only
// anchors dropped whole are the shape that was actually measured —
// `<a class="info-card-grid__card-link" href="…">Read your rights <em
// class="…-arrow">→</em></a>` — and anything else falls back to unlink.
//
// openingTag is the whole matched anchor; only the bytes up to the first '>'
// are its attributes, which is exact because repairAnchorRe confines its
// attribute run to [^>]*.
func anchorIsLabelledControl(anchor, inner string) bool {
	tagEnd := strings.IndexByte(anchor, '>')
	if tagEnd < 0 {
		return false
	}
	if !strings.Contains(strings.ToLower(anchor[:tagEnd]), "class=") {
		return false
	}
	// The same normalisation ExtractAnchors uses, so "what the link says" has
	// one definition: inner markup stripped, whitespace collapsed. An arrow
	// glyph inside an <em>/<span> therefore does not inflate the length.
	return len([]rune(normaliseAnchorText(inner))) <= suppressControlLabelMax
}
