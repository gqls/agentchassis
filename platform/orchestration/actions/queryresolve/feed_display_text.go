// FILE: platform/orchestration/actions/queryresolve/feed_display_text.go
//
// The display projection for `content_feed_items` text — ONE seam for the three
// readers that put a third-party news title or summary in front of a visitor.
//
// WHY THIS EXISTS (bugs_open/332, 2026-09-03). Three readers of one column each
// decided independently what "display" means, and only one of them sanitised:
//
//	queryresolve/news_items.go     projectNewsItems  the news page HTML  — stripped
//	render_news_section_action.go  loadNewsItems     /data/*.json        — did NOT
//	render_rss_feed_action.go      loadRSSItems      feed.xml            — did NOT
//
// Nobody chose that. It is what happens when the judgement is not in one place —
// 016b §9, "one call site of a shared judgement gets the rigorous fix; the
// sibling stays heuristic", and there were two siblings. The cost, measured at
// the served artefact 2026-09-03: /data/news-archive.json on a PAID customer
// site carried 7 ATX headings, 4 complete markdown links, 5 truncated links, a
// list marker, an image and a bold marker, while the server-rendered HTML of the
// SAME query carried zero headings. And every news page loads a published
// /tools/assets/news-listing.js that fetches that JSON and assigns it into
// innerHTML unconditionally, so for a JS-enabled visitor the unstripped copy is
// the one that gets read.
//
// After this file, "read feed text for display" and "apply the display
// discipline" are the same act, and a fourth reader inherits it. Same shape and
// same reason as list_item_text.go (bugs_open/425, written 2026-09-02): "two
// producers derive the same listing for the same component, and a hand-copied
// rule is how a deliberate split becomes accidental drift on a tree this many
// sessions share."
//
// WHAT DELIBERATELY DOES **NOT** MOVE HERE
// ----------------------------------------
// The three readers legitimately disagree on six things, and unifying any of
// them would be a regression rather than a tidy-up. Each stays with its caller:
//
//	escaping    html.EscapeString (text/template escapes nothing) · encoding/json
//	            (the browser escapes at render) · xml.Marshal (the marshaller
//	            escapes). Moving escaping in here would DOUBLE-escape two of the
//	            three — render_rss_feed_test.go pins the XML case explicitly.
//	budget      200 / 200 / 500. RSS deliberately carries more.
//	attribution the RSS " (Fuente: X)" suffix — appended by loadRSSItems AFTER
//	            this cut, so it can never itself be truncated away.
//	selection   relevance-first with a per-tool cap (the page) vs chronological
//	            with URL dedup (the feed). A relevance-ordered RSS feed re-orders
//	            on every rebuild and breaks reader-side positional dedup.
//	dates       RFC1123Z · compact "3d ago" for client expansion · long-form.
//	topics      joined string · escaped slice · absent.
//
// So this file holds exactly two things: the STRIP, and the CUT.

package queryresolve

import (
	"os"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// newsMarkdownStripDisabledEnv is the kill switch, MOVED HERE from
// news_items.go on 2026-09-03. It ships ARMED (the owner has ruled against
// default-OFF switches that rot unexercised) and exists because council
// 060bcc0a r5 (guardian) objected that an unconditional, fleet-wide lossy
// transform with no off-switch short of a deploy is a posture this estate does
// not accept.
//
// Moving it is the point, not a side effect: bugs_open/332's own fix candidate
// asks for "the SAME DISABLE_NEWS_MARKDOWN_STRIP switch so one lever disarms
// both producers", and while the switch lived inside projectNewsItems it could
// not reach the other two readers however it was set. One lever, three
// producers — and that is checkable: set it, re-render, and if the JSON comes
// back CLEAN the promise is false.
const newsMarkdownStripDisabledEnv = "DISABLE_NEWS_MARKDOWN_STRIP"

// feedStripEnabled reads the switch. A func, not a package var, so a test can
// set the env and see the change without a process restart.
func feedStripEnabled() bool { return os.Getenv(newsMarkdownStripDisabledEnv) == "" }

// FeedSummaryMaxBytes is the display budget for a news summary on the page and
// in the JSON. BYTES, matching datahelpers.SafeCut and every other cut site in
// this estate. The RSS reader passes its own larger budget.
const FeedSummaryMaxBytes = 200

// FeedDisplayTitle returns a feed item's title as display text.
//
// Titles are not truncated by any of the three readers, so this is the strip
// alone. It matters anyway: 2 of 834 RSS-sourced rows carried markdown in their
// title [MEASURED 2026-09-03], and loadRSSItems emitted title.String verbatim.
func FeedDisplayTitle(s string) string {
	if !feedStripEnabled() {
		return s
	}
	out, _ := datahelpers.StripFeedDisplayMarkdown(s, !datahelpers.HTMLMarkupRe.MatchString(s))
	return strings.TrimSpace(out)
}

// FeedDisplaySummary returns a feed item's summary as display text, stripped
// and then cut to maxBytes at a word boundary.
//
// STRIP BEFORE TRUNCATE, ALWAYS, and the order is the load-bearing part: a link
// cut mid-URL leaves a half-pattern nothing downstream can match. That rule was
// established for one of these three readers in August
// (TestProjectNewsItemsStripsBeforeTruncating) and is now structural for all
// three — there is one cut, and it strips first.
//
// It uses StripFeedDisplayMarkdown, the TIER 2 strip, not StripLiteralMarkdown:
// a feed snippet is a fragment of a scraped markdown DOCUMENT cut mid-token by
// our own 197-byte snippet truncation, which is a different population from the
// prose our own writers produce. Read the WHY TWO TIERS block in
// datahelpers/literal_markdown.go before changing that.
func FeedDisplaySummary(s string, maxBytes int) string {
	if feedStripEnabled() {
		cleaned, _ := datahelpers.StripFeedDisplayMarkdown(s, !datahelpers.HTMLMarkupRe.MatchString(s))
		s = strings.TrimSpace(cleaned)
	}
	return feedSummaryCut(s, maxBytes)
}

// feedSummaryCut trims to maxBytes at a word boundary with an ellipsis.
//
// This replaces TWO functions with identical bodies in two packages —
// news_items.go's truncateSummary and render_news_section_action.go's
// truncateNewsSummary — and fixes the defect both of them carried: they sliced
// BYTES. A cut through a multi-byte rune emits invalid UTF-8, which Postgres
// refuses on the way back in (bugs_open/423's failure). That is not
// hypothetical here: 2 rows of content_feed_items already carry U+FFFD
// [MEASURED 2026-09-03], from the same byte-slicing defect one layer upstream.
//
// It delegates the rune-safety rather than implementing it. datahelpers.SafeCut
// has been "the one truncation primitive in this codebase since 2026-07-20"
// (bugs_open/027 §4b), and list_item_text.go records the council's reuse_agent
// seat catching exactly this reinvention 24 hours ago: "a second truncation
// primitive is exactly how the two spellings this file exists to unify get
// recreated one layer down."
//
// The word-boundary back-off is the two originals' semantics, kept verbatim so
// this is a refactor and not a redesign: back off to the last space, unless that
// would discard more than half the budget, in which case take the hard cut.
func feedSummaryCut(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	cut := datahelpers.SafeCut(s, maxBytes)
	if idx := strings.LastIndex(cut, " "); idx >= maxBytes/2 {
		cut = cut[:idx]
	}
	return cut + "..."
}
