// FILE: platform/orchestration/actions/queryresolve/list_item_text.go
//
// The TEXT half of the standard list-item shape, shared for the same reason
// ListedPageEligibilitySQL and PageImageProjectionSQL are shared: two producers
// derive the same listing for the same component, and a hand-copied rule is how
// a deliberate split becomes accidental drift on a tree this many sessions
// share.
//
// WHAT WENT WRONG WITHOUT IT (bugs_open/425, boxingonline.com, 2026-09-02).
// `rebuild_blog_listing`'s scanBlogArticles stripped the trailing site-name
// suffix from pages.title and projected an excerpt from meta_description.
// resolvePagesWhereType — the resolver behind `query.blog_posts`, which feeds
// the SAME content-listing component on the home page — did neither. So one
// site served two spellings of one card: the blog index showed
// "Cruiserweight Is Boxing's Best-Kept Secret", the home page showed
// "Cruiserweight Is Boxing's Best-Kept Secret — And It Won't Stay That Way |
// Boxing Online", and the home page's deck was an empty <p>.
//
// The resolver's own doc-comment already documented the intended shape —
// `"title": "Jump Physics"`, unsuffixed — so this is the projection being made
// to keep the contract it publishes, not a new guarantee.

package queryresolve

import (
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// listItemTitleSeparator is the separator apply_gap_plan_action.go uses when it
// composes a document title as "<page title> | <site>". Kept as a named
// constant so the producer and the stripper name one thing.
const listItemTitleSeparator = " | "

// ListItemExcerptMaxBytes bounds a card deck. Cards are a fixed-height grid
// cell; an unbounded deck pushes the meta row out of the card on the narrowest
// breakpoint. 200 is the bound rebuild_blog_listing has applied since it was
// written — adopted here rather than re-chosen, because the point of this file
// is that the two producers agree.
//
// BYTES, matching datahelpers.TruncateString and the estate's other cut sites
// (render_site_components uses SafeCut(s, 247)). The first cut of this file
// bounded by RUNES and hand-rolled the rune-safety, which the council's
// reuse_agent seat caught: datahelpers.SafeCut has been the one truncation
// primitive in this codebase since 2026-07-20 (bugs_open/027 §4b), and
// TruncateString is SafeCut plus the ellipsis — exactly this function's body.
// Practically identical on live data: [MEASURED 2026-09-02] the mean
// meta_description across the sites carrying this component is 116–150 chars,
// so neither bound truncates a typical deck at all.
const ListItemExcerptMaxBytes = 200

// ListItemTitle returns the DISPLAY headline for a listing card, stripping the
// trailing " | <site name>" suffix that the page's document <title> carries.
//
// LastIndex, not Index, and that is load-bearing: a headline may legitimately
// contain the separator ("Rules, Scoring | What Changed | Boxing Online" —
// [MEASURED 2026-09-02] 24 of 1,172 live page titles carry two or more), and
// only the LAST segment is the site name.
//
// It is deliberately the same rule scanBlogArticles already shipped rather than
// a better one. A stricter rule (strip only when the suffix matches the site's
// own name) would be more precise and would ALSO put the two producers back
// into disagreement until every caller adopted it, which is the defect this
// file exists to close. Tighten it here, once, if it is ever worth tightening.
//
// A title that is nothing but a suffix is left alone: stripping would return
// the empty string, and a blank headline is worse than a suffixed one.
func ListItemTitle(title string) string {
	idx := strings.LastIndex(title, listItemTitleSeparator)
	if idx <= 0 {
		return title
	}
	return title[:idx]
}

// ListItemExcerpt returns the one-sentence deck for a listing card, projected
// from the page's meta_description and bounded at ListItemExcerptMaxBytes.
//
// RUNE-SAFE, via the shared primitive. Slicing a UTF-8 string by byte offset
// can cut a multi-byte character in half and yield invalid UTF-8, which
// Postgres then refuses on the way back in — the failure bugs_open/423 records,
// where a sliced 0x80 made a store fail and the failure branch swallowed it.
// Meta descriptions on this estate routinely carry em-dashes and curly quotes,
// so this is the live case, not a hypothetical one.
//
// It delegates rather than implementing that: datahelpers.SafeCut backs off to
// the last rune start and TruncateString adds the ellipsis. Reinventing it here
// was a real finding of the council's reuse_agent seat, not a style note — a
// second truncation primitive is exactly how the two spellings this file exists
// to unify get recreated one layer down.
func ListItemExcerpt(metaDescription string) string {
	return datahelpers.TruncateString(metaDescription, ListItemExcerptMaxBytes)
}
