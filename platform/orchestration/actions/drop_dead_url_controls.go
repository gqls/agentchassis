package actions

import (
	"regexp"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// DropDeadURLControls and its regexes live in their own file (bugs_open/054)
// deliberately: the render-safety helper is small, self-contained, and was added
// while another session was actively editing component_library.go — keeping it
// here avoids a same-file commit passenger. It is called from
// render_site_components_action.go's chrome render path.

// deadAnchorRe matches an anchor whose href rendered EMPTY (an empty double- or
// single-quoted href, any whitespace around the =) through to its closing tag.
// The \b after <a means
// <area>/<abbr> are not matched; [^>]* is confined to one tag's attributes so a
// sibling anchor with a real href is never swept in; anchors cannot nest, so the
// non-greedy .*? closes at this anchor's own </a>. (?is): dot spans newlines,
// tag/attr names case-insensitive.
var deadAnchorRe = regexp.MustCompile(`(?is)<a\b[^>]*\shref\s*=\s*(?:""|'')[^>]*>.*?</a>`)

// deadImgRe matches an <img> whose src rendered empty (empty double- or
// single-quoted src). The whole void element is DROPPED, not just its attribute:
// an empty src="" makes the browser re-request the current document, and a bare
// <img> left behind is itself a residual dead control (council editquality note,
// 2026-07-22) — symmetry with the anchor drop.
var deadImgRe = regexp.MustCompile(`(?is)<img\b[^>]*\ssrc\s*=\s*(?:""|'')[^>]*>`)

// deadSrcAttrRe blanks an empty src on any OTHER element (<source>, <script>, …)
// where dropping the whole element is not obviously safe: the element then loads
// nothing rather than re-requesting the page.
var deadSrcAttrRe = regexp.MustCompile(`(?i)\ssrc\s*=\s*(?:""|'')`)

// DropDeadURLControls removes interactive controls whose URL attribute rendered
// empty from already-rendered site-chrome HTML. bugs_open/054: after 018's
// schema-driven fill an empty href/src in chrome is never legitimate (nav toggles
// use href="#", not href=""), so it is a DEAD CONTROL. Rather than ship it
// (idea.uk shipped 30 empty-href nav links), the whole anchor is dropped
// (LNK-005: correct-or-absent), a dead <img> is dropped whole, and any other
// empty src is blanked. Callers gate this on a non-empty deadURLFields set from
// RenderTemplateReportingMissing, so a clean render is never scanned and
// byte-identical output is preserved on the happy path.
// The three replacements go through datahelpers.ReplaceAllInMarkup rather than
// ReplaceAllString because these regexes have the blindness bugs_open/180 was
// filed about: `href=""` written inside a <script> string, a <style> comment, a
// <textarea> or an HTML comment is TEXT, and this function DELETES the whole
// element it thinks it has found — a strictly worse outcome than the unlink that
// 180 was reported for. Measured over all 509 assembled pages on 2026-08-02:
// 13 dead-anchor and 37 dead-src matches fleet-wide and ZERO of them inside a
// non-markup region, so this is byte-for-byte the same function on today's fleet
// and the hole is closed before something writes the markup that opens it.
func DropDeadURLControls(html string) string {
	html = datahelpers.ReplaceAllInMarkup(deadAnchorRe, html, "")
	html = datahelpers.ReplaceAllInMarkup(deadImgRe, html, "")
	html = datahelpers.ReplaceAllInMarkup(deadSrcAttrRe, html, "")
	return html
}
