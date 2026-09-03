// FILE: platform/orchestration/actions/queryresolve/news_items.go
//
// News feed item resolution for `query.latest_news` / `query.news_archive`
// (bugs_open/027, reworked after council REVISE 4b91237a).
//
// WHY HERE: news items previously reached the page only as JSON files fetched
// client-side, so every news page served zero news to non-JS consumers. The
// first fix injected HTML into page_components.rendered_html directly — the
// mechanism 003 explicitly rejects ("HTML patching was rejected as an edit
// mechanism"): a scoped rerender regenerates rendered_html from html_template
// + content_data, wiping anything that lives in neither. The contract-
// compliant route is this one: items are a `query.*`-sourced schema field, so
// they land in content_data like every other section field, the html_template
// renders them, and a rerender REFRESHES them instead of wiping them.
//
// ONE QUERY, TWO PROJECTIONS: QueryNewsItems is the single selection —
// shared with RenderNewsSectionAction's JSON path (loadNewsItems delegates
// here) so the server-rendered HTML and the client-fetched JSON can never
// disagree about which items exist. Each consumer projects its own shape:
//   - the JSON path emits raw text and compact dates ("3d ago") that the
//     client scripts expand;
//   - the resolvers here emit HTML-ESCAPED text and display-ready dates,
//     because component templates render through text/template
//     (component_library.go RenderTemplate), which does NOT auto-escape, and
//     feed titles/summaries are third-party content.
// Escaped values in content_data are correct, not drift: content_data is the
// render source, so it holds what the template should emit.

package queryresolve

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NewsItem is one feed item as selected, raw and unescaped. Consumers project.
type NewsItem struct {
	Title       string
	Summary     string
	URL         string
	Source      string
	PublishedAt time.Time // zero when the feed carried no date
	Topics      []string
}

// Age windows. These mirror the live render_news_section seed config
// (max_age_hours=720 since d3c2f95db, archive = 4x that, same as the action's
// maxAgeHours*4). If the seed window changes, change these WITH it — the
// resolver and the JSON must select the same items or server HTML and client
// refresh will disagree.
const (
	latestNewsAgeHours  = 720
	newsArchiveAgeHours = 720 * 4
)

// newsItemsPerToolCap: at most this many items about any one tool/brand in a
// rendered set (owner ruling D14, 2026-07-29: "no more than two articles for
// one tool, so we keep the usefulness of the site high" —
// webdesign_couk/PLAN_2026-07-25 §D14). Bounds the observed shape where one
// story or product release crowds the feed because dedup keys on source_url:
// five outlets on one Coca-Cola rebrand passed as five items, and Firefox
// 153/154 coverage took four slots. Applied here, inside the single shared
// selection, so the server-rendered HTML and the client JSON stay in
// agreement about which items exist.
const newsItemsPerToolCap = 2

// QueryNewsItems is the single news-item selection, shared by the query.*
// resolvers and RenderNewsSectionAction's JSON path. The WHERE/ORDER BY are
// the JSON path's semantics verbatim: relevant-first, then relevance score,
// then recency; future-dated items (misdated feeds) excluded beyond one day.
// The result obeys newsItemsPerToolCap: the query over-fetches, the cap drops
// lower-ranked items sharing a tool key with two better-ranked ones, and the
// set is trimmed back to maxItems.
func QueryNewsItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxAgeHours, maxItems int, logger *zap.Logger) ([]NewsItem, error) {
	// Over-fetch so capped items can be back-filled from the next-ranked ones
	// rather than shrinking the page below maxItems.
	fetchLimit := maxItems * 3
	if fetchLimit > 150 {
		fetchLimit = 150
	}
	if fetchLimit < maxItems {
		fetchLimit = maxItems
	}
	rows, err := db.QueryContext(ctx, `
		SELECT
			COALESCE(cfi.source_title, '')          AS title,
			COALESCE(cfi.source_summary, '')        AS summary,
			COALESCE(cfi.source_url, '')            AS url,
			cfi.source_published_at,
			COALESCE(cs.name, '')                   AS source_name,
			COALESCE(cfi.topics::text, '[]')        AS topics_json
		FROM content_feed_items cfi
		LEFT JOIN content_sources cs ON cs.id = cfi.source_id
		WHERE cfi.site_id = $1
		  AND cfi.status IN ('relevant', 'ingested')
		  AND cfi.created_at > NOW() - make_interval(hours => $2)
		  AND (cfi.source_published_at IS NULL
		       OR (cfi.source_published_at > NOW() - make_interval(hours => $2)
		           AND cfi.source_published_at <= NOW() + INTERVAL '1 day'))
		ORDER BY
			CASE WHEN cfi.status = 'relevant' THEN 0 ELSE 1 END,
			cfi.relevance_score DESC NULLS LAST,
			cfi.source_published_at DESC NULLS LAST,
			cfi.created_at DESC
		LIMIT $3
	`, siteID, maxAgeHours, fetchLimit)
	if err != nil {
		return nil, fmt.Errorf("QueryNewsItems: %w", err)
	}
	defer rows.Close()

	items := make([]NewsItem, 0)
	for rows.Next() {
		var title, summary, url, sourceName string
		var publishedAt sql.NullTime
		var topicsJSON string
		if err := rows.Scan(&title, &summary, &url, &publishedAt, &sourceName, &topicsJSON); err != nil {
			logger.Warn("QueryNewsItems: scan failed", zap.Error(err))
			continue
		}
		it := NewsItem{Title: title, Summary: summary, URL: url, Source: sourceName}
		if publishedAt.Valid {
			it.PublishedAt = publishedAt.Time
		}
		var topics []string
		if err := json.Unmarshal([]byte(topicsJSON), &topics); err == nil {
			it.Topics = topics
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("QueryNewsItems rows: %w", err)
	}
	topical := newsTopicalTokens(siteSourceQueries(ctx, db, siteID, logger), items)
	items = capNewsItemsPerTool(items, newsItemsPerToolCap, topical)
	if len(items) > maxItems {
		items = items[:maxItems]
	}
	return items, nil
}

// siteSourceQueries returns the site's active content_sources query strings.
// Best-effort: on error the cap still runs with frequency-derived topics only.
func siteSourceQueries(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) []string {
	rows, err := db.QueryContext(ctx, `
		SELECT COALESCE(config->>'query', '') FROM content_sources
		WHERE site_id = $1 AND is_active
	`, siteID)
	if err != nil {
		logger.Warn("siteSourceQueries: query failed, cap runs without query topics", zap.Error(err))
		return nil
	}
	defer rows.Close()
	var queries []string
	for rows.Next() {
		var q string
		if rows.Scan(&q) == nil && q != "" {
			queries = append(queries, q)
		}
	}
	return queries
}

// newsTopicalTokens builds the per-site set of TOPIC words the cap must never
// treat as tool keys. Two sources, both derived rather than configured:
//   - the site's own source queries: what a site asks its feed for IS its
//     subject vocabulary ("CSS new features browser support" makes css and
//     browser topics for webdesign.co.uk, never tools);
//   - document frequency over the fetched pool: a token in >=25% of titles
//     (min 3) is the site's subject matter even when no query names it —
//     gaswholesalers' query says "gas" but its pool is oil-market coverage,
//     and "Oil" must not be a tool key there.
//
// Derivation matters because a hardcoded list is one site's topics in
// disguise: the first cut of this cap hardcoded css/html/design and the
// pre-submission simulation showed it dropping 13 of gaswholesalers' top 20.
func newsTopicalTokens(queries []string, items []NewsItem) map[string]bool {
	topical := make(map[string]bool)
	for _, q := range queries {
		for _, w := range strings.Fields(strings.ToLower(q)) {
			w = strings.Trim(w, "\"'.,:;!?()[]")
			if len(w) >= 2 {
				topical[w] = true
			}
		}
	}
	// Frequency derivation needs a pool large enough that "appears a lot"
	// means "is the subject matter", not "is one well-covered story". Below
	// 12 items a genuine tool cluster (four Firefox headlines) would cross
	// any workable threshold, so small pools rely on query topics alone.
	if len(items) >= 12 {
		df := make(map[string]int)
		for _, it := range items {
			for _, k := range titleTokens(it.Title) {
				df[k]++
			}
		}
		threshold := len(items) / 4
		if threshold < 5 {
			threshold = 5
		}
		for k, n := range df {
			if n >= threshold {
				topical[k] = true
			}
		}
	}
	return topical
}

// capNewsItemsPerTool walks the ranked items and drops any item whose title
// shares a tool key with maxPer already-kept items. Order is preserved, so the
// query's ranking (relevant-first, score, recency) decides WHICH two survive.
func capNewsItemsPerTool(items []NewsItem, maxPer int, topical map[string]bool) []NewsItem {
	if maxPer <= 0 || len(items) == 0 {
		return items
	}
	counts := make(map[string]int)
	out := make([]NewsItem, 0, len(items))
	for _, it := range items {
		keys := titleToolKeys(it.Title, topical)
		blocked := false
		for _, k := range keys {
			if counts[k] >= maxPer {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		for _, k := range keys {
			counts[k]++
		}
		out = append(out, it)
	}
	return out
}

// titleTokens returns the normalised capitalised tokens of a headline —
// tool-key CANDIDATES, before topical filtering. Intra-word hyphens are kept
// so "Coca-Cola" is one token; possessives are stripped.
func titleTokens(title string) []string {
	var toks []string
	seen := make(map[string]bool)
	for _, raw := range strings.Fields(title) {
		tok := strings.Trim(raw, "\"'“”‘’.,:;!?()[]…")
		tok = strings.TrimSuffix(tok, "'s")
		tok = strings.TrimSuffix(tok, "’s")
		if len(tok) < 3 {
			continue
		}
		r := []rune(tok)
		if r[0] < 'A' || r[0] > 'Z' {
			continue // tool/brand names in headlines are capitalised
		}
		key := strings.ToLower(tok)
		if seen[key] {
			continue
		}
		seen[key] = true
		toks = append(toks, key)
	}
	return toks
}

// titleToolKeys extracts the tool/brand keys a headline is about: capitalised
// tokens that are neither generic headline vocabulary nor the site's own
// topics. Deliberately a heuristic, not entity resolution — it only has to
// make same-tool headlines COLLIDE ("Firefox 153 Officially Released" /
// "Mozilla Firefox 154 Enters Beta" share "firefox"), and its failure modes
// are bounded: a spurious collision costs one surplus item its slot, a missed
// one leaves the feed no worse than before the cap.
func titleToolKeys(title string, topical map[string]bool) []string {
	var keys []string
	for _, key := range titleTokens(title) {
		if headlineStopwords[key] || topical[key] {
			continue
		}
		keys = append(keys, key)
	}
	return keys
}

// headlineStopwords are capitalised-in-headlines words that are not names:
// articles, question openers, common verbs/adjectives that headline case or
// sentence position capitalises, plus cross-domain tech acronyms. Site
// SUBJECT vocabulary does not belong here — that is newsTopicalTokens' job,
// derived per site.
var headlineStopwords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "from": true,
	"how": true, "why": true, "what": true, "when": true, "where": true,
	"who": true, "this": true, "that": true, "these": true, "those": true,
	"new": true, "your": true, "you": true, "its": true, "it's": true,
	"are": true, "was": true, "will": true, "can": true, "not": true,
	"all": true, "top": true, "best": true, "here": true, "there": true,
	"after": true, "before": true, "into": true, "over": true, "under": true,
	"released": true, "release": true, "releases": true, "official": true,
	"officially": true, "unveils": true, "launches": true, "launch": true,
	"introducing": true, "announces": true, "announced": true, "gets": true,
	"enters": true, "arrives": true, "brings": true, "adds": true,
	"update": true, "updates": true, "version": true, "beta": true,
	"stable": true, "preview": true, "guide": true, "review": true,
	"look": true, "first": true, "market": true, "industry": true,
	"global": true, "report": true, "analysis": true,
	"api": true, "pdf": true, "http": true, "https": true,
	"json": true, "svg": true,
}

// resolveLatestNews backs `source: "query.latest_news"` — the homepage
// latest-news card grid. Default 6 items (the card row), cap 12.
func resolveLatestNews(ctx context.Context, db *sql.DB, siteID uuid.UUID, limit int, logger *zap.Logger) (interface{}, error) {
	if limit <= 0 {
		limit = 6
	}
	if limit > 12 {
		limit = 12
	}
	items, err := QueryNewsItems(ctx, db, siteID, latestNewsAgeHours, limit, logger)
	if err != nil {
		return nil, err
	}
	logger.Info("queryresolve: resolved latest_news", zap.Int("items", len(items)))
	return projectNewsItems(items, false, logger), nil
}

// resolveNewsArchive backs `source: "query.news_archive"` — the news-index
// listing page. Default 20 items (the archive page depth), cap 50.
func resolveNewsArchive(ctx context.Context, db *sql.DB, siteID uuid.UUID, limit int, logger *zap.Logger) (interface{}, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	items, err := QueryNewsItems(ctx, db, siteID, newsArchiveAgeHours, limit, logger)
	if err != nil {
		return nil, err
	}
	logger.Info("queryresolve: resolved news_archive", zap.Int("items", len(items)))
	return projectNewsItems(items, true, logger), nil
}

// The kill switch DISABLE_NEWS_MARKDOWN_STRIP and its reader MOVED 2026-09-03
// to feed_display_text.go (bugs_open/332). Its reasoning is unchanged and lives
// there: council 060bcc0a r5 (guardian) required an off-switch for an
// unconditional fleet-wide lossy transform, and it ships ARMED. What moving it
// buys is that it now reaches all THREE display readers — while it lived here it
// could not touch the JSON or RSS producers, so "one lever disarms both
// producers", which 332's own fix candidate asks for, was simply not true.

// projectNewsItems shapes raw items for template rendering: HTML-escaped
// text, truncated summaries, display-ready dates. includeTopics mirrors the
// JSON path's archive-only topics.
//
// LITERAL MARKDOWN IS STRIPPED HERE, by default and independent of any step
// flag (bugs_open/184, canary finding 2026-08-19): content_feed_items.
// source_summary is a faithful INGEST record and legitimately carries the
// source's raw markdown (~700 of 10,855 rows measured) — but this
// projection's output is a PLAIN-TEXT render value fed to text/template,
// where `# headings` and `[text](url)` reach the visitor verbatim.
//
// WHY AT THE PRODUCER, NOT ONLY BEHIND THE STEP FLAG (council 060bcc0a r5/r6):
// the step-flag strips (render_component's strip_literal_markdown, the
// reason-gated rerender) only run where a step opted in. page-content-writer
// has TWO render_component steps over the same merge_with overlay —
// render_section (flagged ON by migration 474) and render_from_template
// (template-only sections; flag UNSET, measured live 2026-08-19) — so a
// template-only section listing news would still receive raw markdown from
// this resolver. The producer-local strip is what covers the unflagged
// caller, and it is the only layer that knows the value is display text.
//
// PRECEDENT, stated precisely (reuse_agent/architecture seats, r5): the
// posture — a query resolver sanitises its own display output — is the one
// this file's EscapeString calls and directory_items.go's projection already
// take. The BEHAVIOUR is not the same: those escape HTML (loss-free) and do
// not strip markdown (lossy by design: marker characters are deleted). That
// asymmetry is why this strip has the kill switch above and the escaping
// does not.
//
// ORDER, verified in code: the tool-key cap and topical dedup
// (newsTopicalTokens / capNewsItemsPerTool / titleToolKeys) run inside
// QueryNewsItems on the RAW titles, before this projection is called, so the
// strip changes no clustering decision. Strip runs BEFORE truncation —
// truncating "[Luke Littler](https://…" mid-URL leaves a half-pattern
// nothing can match. The raw record stays raw; the JSON path's projection is
// untouched (client scripts own their own rendering).
//
// OTHER READERS of source_summary, named rather than implied (bug_historian
// r5): feed_triage_actions.go feeds it to the triage LLM as INPUT — raw is
// correct there, markdown is data not display; render_rss_feed_action.go
// emits it on the RSS surface, independently of this projection — a different
// output with its own (XML) escaping and no literal_markdown detector over
// it, so it is left alone here and named as the feed-quality follow-up
// alongside the scraped markdown tables (bugs_closed/184 residuals).
//
// OBSERVABILITY: every projection that stripped anything logs the count at
// Info (the strip is otherwise silent — bug_historian r5). The page-level
// record is the verifier's: a served page that still scans dirty fails the
// item honestly.
func projectNewsItems(items []NewsItem, includeTopics bool, logger *zap.Logger) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	stripped := 0
	for _, it := range items {
		// The strip, the truncation and the kill switch all moved to
		// feed_display_text.go on 2026-09-03 (bugs_open/332) so that this
		// resolver, the JSON producer and the RSS producer cannot disagree about
		// what display means — the two of them that never stripped at all are
		// what made 332 live. Escaping stays HERE, because it is the one thing
		// the three readers legitimately differ on.
		title := FeedDisplayTitle(it.Title)
		summary := FeedDisplaySummary(it.Summary, FeedSummaryMaxBytes)
		if title != it.Title || summary != it.Summary {
			stripped++
		}
		m := map[string]interface{}{
			"title":   html.EscapeString(title),
			"summary": html.EscapeString(summary),
			"url":     html.EscapeString(it.URL),
			"source":  html.EscapeString(it.Source),
			"date":    newsDisplayDate(it.PublishedAt),
		}
		if includeTopics {
			topics := make([]string, 0, len(it.Topics))
			for _, t := range it.Topics {
				if t = strings.TrimSpace(t); t != "" {
					topics = append(topics, html.EscapeString(t))
				}
			}
			m["topics"] = topics
		}
		out = append(out, m)
	}
	if stripped > 0 && logger != nil {
		logger.Info("queryresolve: stripped literal markdown from news items",
			zap.Int("items_stripped", stripped),
			zap.Int("items", len(items)))
	}
	return out
}

// newsDisplayDate renders a published time the way the client scripts do
// after their formatNewsDate expansion — long-form relative for recent items,
// absolute beyond a week — so server-rendered and client-refreshed text agree.
// Zero time renders empty (template {{if .date}} omits it).
func newsDisplayDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	hours := time.Since(t).Hours()
	switch {
	case hours < 1:
		return "Just now"
	case hours < 24:
		n := int(hours)
		if n == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", n)
	case hours < 48:
		return "Yesterday"
	case hours < 168:
		n := int(hours / 24)
		if n == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", n)
	default:
		return t.Format("2 Jan 2006")
	}
}

// truncateSummary MOVED 2026-09-03 to feed_display_text.go as feedSummaryCut
// (bugs_open/332). It was byte-identical to render_news_section_action.go's
// truncateNewsSummary, and both sliced BYTES — a cut through a multi-byte rune
// emits invalid UTF-8. The replacement delegates to datahelpers.SafeCut.
