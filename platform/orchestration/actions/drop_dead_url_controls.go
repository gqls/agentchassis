package actions

import "regexp"

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

// deadSrcAttrRe matches an empty src attribute (empty double- or single-quoted
// src). An empty src on an <img>/<script>/<source> makes the browser re-request
// the current document —
// a real defect — so it is stripped, leaving a source-less element that loads
// nothing rather than the page itself.
var deadSrcAttrRe = regexp.MustCompile(`(?i)\ssrc\s*=\s*(?:""|'')`)

// DropDeadURLControls removes interactive controls whose URL attribute rendered
// empty from already-rendered site-chrome HTML. bugs_open/054: after 018's
// schema-driven fill an empty href/src in chrome is never legitimate (nav toggles
// use href="#", not href=""), so it is a DEAD CONTROL. Rather than ship it
// (idea.uk shipped 30 empty-href nav links), the whole anchor is dropped
// (LNK-005: correct-or-absent) and a dead src is blanked. Callers gate this on a
// non-empty deadURLFields set from RenderTemplateReportingMissing, so a clean
// render is never scanned and byte-identical output is preserved on the happy path.
func DropDeadURLControls(html string) string {
	html = deadAnchorRe.ReplaceAllString(html, "")
	html = deadSrcAttrRe.ReplaceAllString(html, "")
	return html
}
