// FILE: platform/orchestration/actions/render_news_section_action.go
//
// RenderNewsSectionAction queries recent relevant feed items for a site
// and produces JSON files ready for git commit. The homepage loads the
// snippet JSON client-side via fetch() — no page rerender needed for
// news updates.
//
// Output: {files: {"data/latest-news.json": "<json>", "data/news-archive.json": "<json>"}, domain: "...", item_count: N}
// The git_commit step in the workflow reads files_field and commits to the repo.
//
// Two JSON files are produced:
//   - data/latest-news.json  — 6 items for the homepage snippet component
//   - data/news-archive.json — 20 items for the dedicated /news.html listing page
//
// The archive file is only produced if a news-index page exists (checked via
// pages table). If no listing page exists, only the snippet JSON is produced.
//
// Registration:
//   "render_news_section": {
//       Handler:     RenderNewsSectionAction,
//       Category:    "feed",
//       Description: "Produce latest-news JSON for git commit",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "render_news_json": {
//       "action": "render_news_section",
//       "config": {
//           "site_id": "input_data.site_id",
//           "max_items": 6,
//           "max_age_hours": 72
//       },
//       "output_field": "news_render_result",
//       "next_step": "commit_news"
//   },
//   "commit_news": {
//       "action": "git_commit",
//       "config": {
//           "files_field": "news_render_result.files",
//           "domain_field": "news_render_result.domain",
//           "commit_message": "Update latest news feed"
//       },
//       "output_field": "news_commit_result",
//       "next_step": "complete"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenderNewsSectionInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"page_name", "max_items", "max_age_hours", "archive_max_items"},
}

func init() {
	datahelpers.RegisterActionInputSpec("render_news_section", RenderNewsSectionInputSpec)
}

// newsJSONOutput is the shape of /data/latest-news.json and /data/news-archive.json
type newsJSONOutput struct {
	Headline      string         `json:"headline"`
	Subheadline   string         `json:"subheadline,omitempty"`
	Items         []newsJSONItem `json:"items"`
	InsightsURL   string         `json:"insights_url,omitempty"`
	InsightsLabel string         `json:"insights_label,omitempty"`
	UpdatedAt     string         `json:"updated_at"`
	ItemCount     int            `json:"item_count"`
	ItemsTotal    int            `json:"items_total,omitempty"` // total available (archive only)
}

type newsJSONItem struct {
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	URL     string `json:"url"`
	Source  string `json:"source,omitempty"`
	Date    string `json:"date,omitempty"`
	Topics  string `json:"topics,omitempty"` // comma-separated tags (archive only)
}

func RenderNewsSectionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "render_news_section"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		RenderNewsSectionInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	maxItems := inputs.GetInt("max_items", 6)
	maxAgeHours := inputs.GetInt("max_age_hours", 72)
	archiveMaxItems := inputs.GetInt("archive_max_items", 20)

	// -----------------------------------------------------------------------
	// 1. Load site domain (needed for git commit path)
	// -----------------------------------------------------------------------
	var domain string
	err = params.DB.QueryRowContext(ctx, `
		SELECT domain FROM sites WHERE id = $1
	`, siteID).Scan(&domain)
	if err != nil {
		return nil, fmt.Errorf("query site domain: %w", err)
	}

	// -----------------------------------------------------------------------
	// 2. Load headline from page_component content_data (set by content writer)
	// -----------------------------------------------------------------------
	headline := "Latest News"
	subheadline := ""

	pageName := inputs.Get("page_name")
	if pageName == "" {
		pageName = "index"
	}

	var existingContentData sql.NullString
	_ = params.DB.QueryRowContext(ctx, `
		SELECT pc.content_data::text 
		FROM page_components pc
		JOIN content_components cc ON cc.id = pc.component_id
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1 AND p.name = $2 AND cc.function = 'latest-news'
		LIMIT 1
	`, siteID, pageName).Scan(&existingContentData)

	if existingContentData.Valid {
		var cd map[string]interface{}
		if json.Unmarshal([]byte(existingContentData.String), &cd) == nil {
			if h, ok := cd["headline"].(string); ok && h != "" {
				headline = h
			}
			if sh, ok := cd["subheadline"].(string); ok {
				subheadline = sh
			}
		}
	}

	// -----------------------------------------------------------------------
	// 3. Expire stale items — runs every cycle as part of normal maintenance
	// -----------------------------------------------------------------------
	// Items with source_published_at > 30 days old, or ingested > 7 days ago
	// without being triaged, get marked expired so they don't accumulate.
	expireResult, err := params.DB.ExecContext(ctx, `
		UPDATE content_feed_items
		SET status = 'expired', updated_at = NOW()
		WHERE site_id = $1
		  AND status IN ('ingested', 'relevant', 'review')
		  AND (
		      (source_published_at IS NOT NULL AND source_published_at < NOW() - INTERVAL '30 days')
		      OR (source_published_at IS NULL AND created_at < NOW() - INTERVAL '7 days' AND status = 'ingested')
		  )
	`, siteID)
	if err != nil {
		logger.Warn("RenderNewsSectionAction: expire query failed", zap.Error(err))
	} else if rows, _ := expireResult.RowsAffected(); rows > 0 {
		logger.Info("RenderNewsSectionAction: expired stale items",
			zap.Int64("count", rows))
	}

	// -----------------------------------------------------------------------
	// 4. Load relevant feed items for homepage snippet
	// -----------------------------------------------------------------------
	// CHANGED: sort order now prefers triaged 'relevant' items over unscored
	// 'ingested' items, and uses relevance_score as a tiebreaker.
	snippetItems, err := loadNewsItems(ctx, params.DB, siteID, maxAgeHours, maxItems, false, logger)
	if err != nil {
		return nil, fmt.Errorf("query feed items: %w", err)
	}

	logger.Info("RenderNewsSectionAction: loaded snippet items",
		zap.Int("count", len(snippetItems)),
		zap.String("domain", domain))

	// If no items at all, still produce a valid JSON (empty array)
	// so the homepage shows gracefully
	if snippetItems == nil {
		snippetItems = []newsJSONItem{}
	}

	// -----------------------------------------------------------------------
	// 5. Check if a news listing page exists (for the "More" link + archive)
	// -----------------------------------------------------------------------
	var insightsURL sql.NullString
	_ = params.DB.QueryRowContext(ctx, `
		SELECT url FROM pages
		WHERE site_id = $1 AND page_type = 'news-index' AND `+datahelpers.PageWantedLivePredicateFor("")+`
		LIMIT 1
	`, siteID).Scan(&insightsURL)

	insightsLabel := ""
	hasListingPage := insightsURL.Valid && insightsURL.String != ""
	if hasListingPage {
		insightsLabel = "More insights →"
	}

	// -----------------------------------------------------------------------
	// 6. Build homepage snippet JSON
	// -----------------------------------------------------------------------
	snippetOutput := newsJSONOutput{
		Headline:      headline,
		Subheadline:   subheadline,
		Items:         snippetItems,
		InsightsURL:   insightsURL.String,
		InsightsLabel: insightsLabel,
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
		ItemCount:     len(snippetItems),
	}

	snippetJSON, err := json.MarshalIndent(snippetOutput, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal snippet JSON: %w", err)
	}

	// -----------------------------------------------------------------------
	// 7. Build files map (start with snippet, add archive if listing page exists)
	// -----------------------------------------------------------------------
	filesMap := map[string]interface{}{
		"data/latest-news.json": string(snippetJSON),
	}

	totalItemCount := len(snippetItems)

	// Only produce archive JSON if a news listing page exists
	if hasListingPage {
		archiveItems, err := loadNewsItems(ctx, params.DB, siteID, maxAgeHours*4, archiveMaxItems, true, logger)
		if err != nil {
			logger.Warn("RenderNewsSectionAction: archive query failed, skipping archive",
				zap.Error(err))
		} else {
			if archiveItems == nil {
				archiveItems = []newsJSONItem{}
			}

			// Count total available items (for "Showing X of Y")
			var totalAvailable int
			_ = params.DB.QueryRowContext(ctx, `
				SELECT COUNT(*) FROM content_feed_items
				WHERE site_id = $1
				  AND status IN ('relevant', 'ingested')
				  AND created_at > NOW() - make_interval(hours => $2)
			`, siteID, maxAgeHours*4).Scan(&totalAvailable)

			// Load headline for the listing page
			archiveHeadline := "Industry News & Insights"
			archiveSubheadline := ""

			var listingContentData sql.NullString
			_ = params.DB.QueryRowContext(ctx, `
				SELECT pc.content_data::text 
				FROM page_components pc
				JOIN content_components cc ON cc.id = pc.component_id
				JOIN pages p ON p.id = pc.page_id
				WHERE p.site_id = $1 AND p.page_type = 'news-index' AND cc.function = 'news-listing'
				LIMIT 1
			`, siteID).Scan(&listingContentData)

			if listingContentData.Valid {
				var cd map[string]interface{}
				if json.Unmarshal([]byte(listingContentData.String), &cd) == nil {
					if h, ok := cd["headline"].(string); ok && h != "" {
						archiveHeadline = h
					}
					if sh, ok := cd["subheadline"].(string); ok {
						archiveSubheadline = sh
					}
				}
			}

			archiveOutput := newsJSONOutput{
				Headline:    archiveHeadline,
				Subheadline: archiveSubheadline,
				Items:       archiveItems,
				UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
				ItemCount:   len(archiveItems),
				ItemsTotal:  totalAvailable,
			}

			archiveJSON, err := json.MarshalIndent(archiveOutput, "", "  ")
			if err != nil {
				logger.Warn("RenderNewsSectionAction: marshal archive JSON failed",
					zap.Error(err))
			} else {
				filesMap["data/news-archive.json"] = string(archiveJSON)
				totalItemCount += len(archiveItems)

				logger.Info("RenderNewsSectionAction: archive JSON produced",
					zap.Int("archive_items", len(archiveItems)),
					zap.Int("total_available", totalAvailable))
			}
		}
	}

	// -----------------------------------------------------------------------
	// 7b. Queue a scoped re-render of the news pages (bugs_open/027)
	// -----------------------------------------------------------------------
	// The news components now declare their items as query.latest_news /
	// query.news_archive schema fields, so the server-rendered HTML comes from
	// the normal template path (plan_sections → queryresolve → content_data →
	// RenderComponentAction) — NOT from patching rendered_html here; 003
	// rejects HTML patching, and a scoped rerender would wipe it. What this
	// step does is deliver freshness: fresh feed items ARE "section data
	// resolved", so emit the same item reconcile_section_data emits, and the
	// scoped path re-resolves the queries and re-renders from stored
	// content_data (no LLM). item_key page_rerender:<page> collapses with
	// concurrent image/data triggers via idx_swi_dedup. Best-effort: a failed
	// emit must not fail the feed render the JSON path just completed.
	rerenderQueued := 0
	if totalItemCount > 0 {
		rerenderQueued = queueNewsPageRerenders(ctx, params.DB, siteID, logger)
	}

	logger.Info("RenderNewsSectionAction: JSON produced",
		zap.Int("snippet_items", len(snippetItems)),
		zap.Int("files_count", len(filesMap)),
		zap.String("domain", domain),
		zap.Bool("has_archive", hasListingPage))

	return map[string]interface{}{
		"files":           filesMap,
		"domain":          domain,
		"item_count":      totalItemCount,
		"headline":        headline,
		"file_path":       "data/latest-news.json",
		"has_archive":     hasListingPage,
		"rendered":        true,
		"rerender_queued": rerenderQueued,
	}, nil
}

// loadNewsItems queries feed items with the improved sort order.
// includeTopics controls whether topic tags are included (for archive view).
// loadNewsItems projects the shared selection (queryresolve.QueryNewsItems —
// ONE query for both the JSON files and the query.latest_news/news_archive
// template fields, so the two can never disagree about which items exist)
// into the JSON shape: raw text, compact dates the client scripts expand,
// comma-joined topics.
func loadNewsItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, maxAgeHours, maxItems int, includeTopics bool, logger *zap.Logger) ([]newsJSONItem, error) {
	raw, err := queryresolve.QueryNewsItems(ctx, db, siteID, maxAgeHours, maxItems, logger)
	if err != nil {
		return nil, err
	}

	var items []newsJSONItem
	for _, r := range raw {
		// STRIPPED 2026-09-03 (bugs_open/332). This projection took title and
		// summary RAW, so /data/latest-news.json and /data/news-archive.json —
		// both public, both fetched and assigned into innerHTML by the published
		// news-listing script — served the markdown the page resolver had just
		// removed. Measured at the artefact that day: the archive JSON on a paid
		// customer site carried 7 ATX headings while the server-rendered HTML of
		// the SAME query carried zero.
		//
		// Escaping deliberately stays absent here: the JSON is DATA, and its
		// consumer escapes at render. HTML-escaping it would double-escape for
		// any correct consumer. (The published script's unescaped innerHTML
		// insertion is a separate defect in content_components.js_content, filed
		// on its own — an escaping bug, not a markdown one.)
		item := newsJSONItem{
			Title:   queryresolve.FeedDisplayTitle(r.Title),
			Summary: queryresolve.FeedDisplaySummary(r.Summary, queryresolve.FeedSummaryMaxBytes),
			URL:     r.URL,
			Source:  r.Source,
		}
		if !r.PublishedAt.IsZero() {
			item.Date = formatNewsDate(r.PublishedAt)
		}
		if includeTopics && len(r.Topics) > 0 {
			item.Topics = strings.Join(r.Topics, ", ")
		}
		items = append(items, item)
	}
	return items, nil
}

// truncateNewsSummary MOVED 2026-09-03 to queryresolve.feedSummaryCut
// (bugs_open/332). It was byte-identical to news_items.go's truncateSummary —
// the same function in two packages — and both sliced BYTES, which cuts a
// multi-byte rune in half; 2 content_feed_items rows already carry U+FFFD from
// that defect one layer upstream. The replacement delegates to
// datahelpers.SafeCut, the estate's one truncation primitive since 2026-07-20.

// formatNewsDate formats a time for display. Relative for recent items.
func formatNewsDate(t time.Time) string {
	hours := time.Since(t).Hours()
	if hours < 1 {
		return "Just now"
	}
	if hours < 24 {
		return fmt.Sprintf("%dh ago", int(hours))
	}
	if hours < 48 {
		return "Yesterday"
	}
	if hours < 168 {
		return fmt.Sprintf("%dd ago", int(hours/24))
	}
	return t.Format("2 Jan 2006")
}
