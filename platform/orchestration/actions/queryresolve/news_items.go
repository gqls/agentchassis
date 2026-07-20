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

// QueryNewsItems is the single news-item selection, shared by the query.*
// resolvers and RenderNewsSectionAction's JSON path. The WHERE/ORDER BY are
// the JSON path's semantics verbatim: relevant-first, then relevance score,
// then recency; future-dated items (misdated feeds) excluded beyond one day.
func QueryNewsItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxAgeHours, maxItems int, logger *zap.Logger) ([]NewsItem, error) {
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
	`, siteID, maxAgeHours, maxItems)
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
	return items, nil
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
	return projectNewsItems(items, false), nil
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
	return projectNewsItems(items, true), nil
}

// projectNewsItems shapes raw items for template rendering: HTML-escaped
// text, truncated summaries, display-ready dates. includeTopics mirrors the
// JSON path's archive-only topics.
func projectNewsItems(items []NewsItem, includeTopics bool) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		m := map[string]interface{}{
			"title":   html.EscapeString(it.Title),
			"summary": html.EscapeString(truncateSummary(it.Summary, 200)),
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

// truncateSummary trims to maxLen at a word boundary with an ellipsis —
// the JSON path's truncation semantics, so both projections agree.
func truncateSummary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	idx := strings.LastIndex(s[:maxLen], " ")
	if idx < maxLen/2 {
		idx = maxLen
	}
	return s[:idx] + "..."
}
