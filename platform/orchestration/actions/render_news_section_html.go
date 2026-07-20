// FILE: platform/orchestration/actions/render_news_section_html.go
//
// Server-side rendering of news items into page_components.rendered_html
// (bugs_open/027).
//
// WHY: RenderNewsSectionAction writes data/latest-news.json and
// data/news-archive.json, and the deployed components fetch those files in the
// browser. That means every news page on the platform serves ZERO news to any
// consumer that does not execute JavaScript — crawlers that skip JS, link
// unfurlers, feed previewers. Confirmed live on relojistas.com,
// gaswholesalers.com and robot-hands.com: HTTP 200, loading placeholder
// present, no <article> elements.
//
// WHAT THIS DOES: the same items that go into the JSON are also rendered to
// HTML and injected into the component's own container, so the page ships
// complete. The JSON and the client script are KEPT — they remain the
// freshness path, because rendered_html only reaches the live site on the next
// rerender+deploy whereas the JSON deploys as a file. Server HTML for
// correctness, client fetch for currency.
//
// ORDERING CONSTRAINT (do not reorder): migration 178 must be applied first.
// The news-listing script used to overwrite its container on an empty feed AND
// on a fetch error, which would have destroyed server-rendered items on any
// blip. 178 guards both paths on "is there a server-rendered article here
// already". latest-news never had that defect. Injecting before 178 is applied
// is a regression, not a fix.
//
// The emitted markup is deliberately byte-compatible with what each script
// produces, so a JS-enabled visitor sees no change when the script swaps the
// container contents. One deliberate divergence: this code HTML-escapes, and
// the scripts do not. Feed titles and summaries are third-party text, so
// escaping is correct; the difference is invisible except on content that
// contains markup characters, where the server version is the safe one.

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// newsContainerAnchors locates the fillable container inside a component's
// rendered_html. Start is matched first; the injection ends at the last
// "</div>" occurring before End. Both components follow the same shape:
// a container div, then a sibling footer div.
type newsContainerAnchors struct {
	Start string
	End   string
}

var (
	// <div class="news-grid" id="news-container"> ... </div><div id="news-footer">
	latestNewsAnchors = newsContainerAnchors{
		Start: `id="news-container">`,
		End:   `<div id="news-footer"`,
	}
	// <div class="news-listing-items" id="news-listing-items"> ... </div>
	// <div class="news-listing-footer" ...>
	newsListingAnchors = newsContainerAnchors{
		Start: `id="news-listing-items">`,
		End:   `<div class="news-listing-footer"`,
	}
)

// renderLatestNewsCardsHTML mirrors latest-news.js's card markup.
func renderLatestNewsCardsHTML(items []newsJSONItem) string {
	var b strings.Builder
	for _, it := range items {
		b.WriteString(`<article class="news-card"><div class="news-card-content">`)
		b.WriteString(fmt.Sprintf(
			`<h3 class="news-card-title"><a href="%s" target="_blank" rel="noopener noreferrer">%s</a></h3>`,
			html.EscapeString(it.URL), html.EscapeString(it.Title)))
		if it.Summary != "" {
			b.WriteString(fmt.Sprintf(`<p class="news-card-summary">%s</p>`, html.EscapeString(it.Summary)))
		}
		b.WriteString(`<div class="news-card-meta">`)
		if it.Source != "" {
			b.WriteString(fmt.Sprintf(`<span class="news-source">%s</span>`, html.EscapeString(it.Source)))
		}
		if it.Date != "" {
			b.WriteString(fmt.Sprintf(`<time class="news-date">%s</time>`, html.EscapeString(expandRelativeNewsDate(it.Date))))
		}
		b.WriteString(`</div></div></article>`)
	}
	return b.String()
}

// renderNewsListingItemsHTML mirrors news-listing.js's item markup.
func renderNewsListingItemsHTML(items []newsJSONItem) string {
	var b strings.Builder
	for _, it := range items {
		b.WriteString(`<article class="news-list-item"><div class="news-list-item-content">`)
		b.WriteString(fmt.Sprintf(
			`<h3 class="news-list-item-title"><a href="%s" target="_blank" rel="noopener noreferrer">%s</a></h3>`,
			html.EscapeString(it.URL), html.EscapeString(it.Title)))
		if it.Summary != "" {
			b.WriteString(fmt.Sprintf(`<p class="news-list-item-summary">%s</p>`, html.EscapeString(it.Summary)))
		}
		b.WriteString(`<div class="news-list-item-meta">`)
		if it.Source != "" {
			b.WriteString(fmt.Sprintf(`<span class="news-list-item-source">%s</span>`, html.EscapeString(it.Source)))
		}
		if it.Date != "" {
			b.WriteString(fmt.Sprintf(`<span class="news-list-item-date">%s</span>`, html.EscapeString(expandRelativeNewsDate(it.Date))))
		}
		b.WriteString(`</div>`)
		if it.Topics != "" {
			b.WriteString(`<div class="news-list-item-topics">`)
			for _, tag := range strings.Split(it.Topics, ", ") {
				if tag == "" {
					continue
				}
				b.WriteString(fmt.Sprintf(`<span class="news-list-tag">%s</span>`, html.EscapeString(tag)))
			}
			b.WriteString(`</div>`)
		}
		b.WriteString(`</div></article>`)
	}
	return b.String()
}

// expandRelativeNewsDate reproduces the scripts' formatNewsDate(), which turns
// the compact forms produced by the Go formatNewsDate ("3d ago", "5h ago") into
// long forms ("3 days ago", "5 hours ago"). Kept here so server and client text
// agree; absolute dates ("2 Jan 2006") and "Just now"/"Yesterday" pass through.
func expandRelativeNewsDate(s string) string {
	units := []struct {
		suffix   string
		singular string
		plural   string
	}{
		{"d ago", " day ago", " days ago"},
		{"h ago", " hour ago", " hours ago"},
		{"m ago", " minute ago", " minutes ago"},
		{"w ago", " week ago", " weeks ago"},
	}
	for _, u := range units {
		if !strings.HasSuffix(s, u.suffix) {
			continue
		}
		n := strings.TrimSuffix(s, u.suffix)
		if n == "" || strings.ContainsFunc(n, func(r rune) bool { return r < '0' || r > '9' }) {
			continue
		}
		if n == "1" {
			return n + u.singular
		}
		return n + u.plural
	}
	return s
}

// injectNewsItems replaces the contents of the anchored container in
// rendered_html with inner. It is idempotent: on the first run it displaces the
// loading placeholder (or the <noscript> block), on later runs it displaces the
// previous items.
//
// Returns ok=false and leaves the HTML untouched when either anchor is missing
// or they appear in the wrong order — a component whose template has drifted
// must be left alone rather than corrupted.
func injectNewsItems(renderedHTML string, a newsContainerAnchors, inner string) (string, bool) {
	if renderedHTML == "" || inner == "" {
		return renderedHTML, false
	}
	si := strings.Index(renderedHTML, a.Start)
	if si < 0 {
		return renderedHTML, false
	}
	openEnd := si + len(a.Start)

	ei := strings.Index(renderedHTML[openEnd:], a.End)
	if ei < 0 {
		return renderedHTML, false
	}
	ei += openEnd

	// The container's own closing tag is the last </div> before the sibling.
	closeIdx := strings.LastIndex(renderedHTML[openEnd:ei], "</div>")
	if closeIdx < 0 {
		return renderedHTML, false
	}
	closeIdx += openEnd

	return renderedHTML[:openEnd] + inner + renderedHTML[closeIdx:], true
}

// persistNewsSectionHTML injects rendered items into every matching
// page_components row for a site and writes them back.
//
// pageFilter is applied to pages: either a name ("index") or, when nameIsType
// is true, a page_type ("news-index"). That mirrors how the action already
// locates each component's content_data for its headline.
//
// LOCKED COMPONENTS ARE SKIPPED, not overwritten. page_components carries
// lock_type/locked_at/lock_expires_at, and a locked row represents a deliberate
// human or agent decision to freeze that markup. Silently rewriting one would
// be exactly the class of defect this change exists to fix.
//
// Failures are logged and counted, never fatal: server-side rendering is an
// enhancement to a JSON pipeline that already works, so it must not be able to
// fail the feed render.
func persistNewsSectionHTML(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageFilter string,
	nameIsType bool,
	function string,
	anchors newsContainerAnchors,
	inner string,
	logger *zap.Logger,
) int {
	if db == nil || inner == "" {
		return 0
	}

	pageClause := "p.name = $2"
	if nameIsType {
		pageClause = "p.page_type = $2"
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT pc.id, COALESCE(pc.rendered_html, '')
		FROM page_components pc
		JOIN content_components cc ON cc.id = pc.component_id
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND %s
		  AND cc.function = $3
		  AND (pc.lock_type IS NULL
		       OR (pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW()))
	`, pageClause), siteID, pageFilter, function)
	if err != nil {
		logger.Warn("persistNewsSectionHTML: select failed",
			zap.String("function", function), zap.Error(err))
		return 0
	}
	defer rows.Close()

	type target struct {
		id   string
		html string
	}
	var targets []target
	for rows.Next() {
		var t target
		if err := rows.Scan(&t.id, &t.html); err != nil {
			logger.Warn("persistNewsSectionHTML: scan failed", zap.Error(err))
			continue
		}
		targets = append(targets, t)
	}

	updated := 0
	for _, t := range targets {
		newHTML, ok := injectNewsItems(t.html, anchors, inner)
		if !ok {
			// Anchors absent — the component template has drifted from the one
			// this renderer knows. Leave it alone and say so; a silent skip here
			// would look identical to success.
			logger.Warn("persistNewsSectionHTML: container anchors not found, skipping",
				zap.String("function", function),
				zap.String("page_component_id", t.id))
			continue
		}
		if newHTML == t.html {
			continue
		}
		if _, err := db.ExecContext(ctx,
			`UPDATE page_components SET rendered_html = $1, updated_at = NOW() WHERE id = $2`,
			newHTML, t.id); err != nil {
			logger.Warn("persistNewsSectionHTML: update failed",
				zap.String("page_component_id", t.id), zap.Error(err))
			continue
		}
		updated++
	}

	if updated > 0 {
		logger.Info("persistNewsSectionHTML: server-rendered news into components",
			zap.String("function", function),
			zap.Int("components_updated", updated))
	}
	return updated
}
