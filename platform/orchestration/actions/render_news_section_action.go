// FILE: platform/orchestration/actions/render_news_section_action.go
//
// RenderNewsSectionAction queries recent relevant feed items for a site
// and produces a JSON file ready for git commit. The homepage loads this
// JSON client-side via fetch() — no page rerender needed for news updates.
//
// Output: {files: {"data/latest-news.json": "<json>"}, domain: "...", item_count: N}
// The git_commit step in the workflow reads files_field and commits to the repo.
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
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenderNewsSectionInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"page_name", "max_items", "max_age_hours"},
}

func init() {
	datahelpers.RegisterActionInputSpec("render_news_section", RenderNewsSectionInputSpec)
}

// newsJSONOutput is the shape of /data/latest-news.json
type newsJSONOutput struct {
	Headline      string         `json:"headline"`
	Subheadline   string         `json:"subheadline,omitempty"`
	Items         []newsJSONItem `json:"items"`
	InsightsURL   string         `json:"insights_url,omitempty"`
	InsightsLabel string         `json:"insights_label,omitempty"`
	UpdatedAt     string         `json:"updated_at"`
	ItemCount     int            `json:"item_count"`
}

type newsJSONItem struct {
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	URL     string `json:"url"`
	Source  string `json:"source,omitempty"`
	Date    string `json:"date,omitempty"`
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
	// 4. Load relevant feed items
	// -----------------------------------------------------------------------
	rows, err := params.DB.QueryContext(ctx, `
		SELECT 
			cfi.source_title,
			cfi.source_summary,
			cfi.source_url,
			cfi.source_published_at,
			COALESCE(cs.name, '') as source_name
		FROM content_feed_items cfi
		LEFT JOIN content_sources cs ON cs.id = cfi.source_id
		WHERE cfi.site_id = $1 
		  AND cfi.status IN ('relevant', 'ingested')
		  AND cfi.created_at > NOW() - make_interval(hours => $2)
		  AND (cfi.source_published_at IS NULL 
		       OR (cfi.source_published_at > NOW() - make_interval(hours => $2)
		           AND cfi.source_published_at <= NOW() + INTERVAL '1 day'))
		ORDER BY cfi.source_published_at DESC NULLS LAST, cfi.created_at DESC
		LIMIT $3
	`, siteID, maxAgeHours, maxItems)
	if err != nil {
		return nil, fmt.Errorf("query feed items: %w", err)
	}
	defer rows.Close()

	var items []newsJSONItem
	for rows.Next() {
		var title, summary, url, sourceName sql.NullString
		var publishedAt sql.NullTime

		if err := rows.Scan(&title, &summary, &url, &publishedAt, &sourceName); err != nil {
			logger.Warn("RenderNewsSectionAction: scan error", zap.Error(err))
			continue
		}

		item := newsJSONItem{
			Title:   title.String,
			Summary: truncateNewsSummary(summary.String, 200),
			URL:     url.String,
			Source:  sourceName.String,
		}

		if publishedAt.Valid {
			item.Date = formatNewsDate(publishedAt.Time)
		}

		items = append(items, item)
	}

	logger.Info("RenderNewsSectionAction: loaded items",
		zap.Int("count", len(items)),
		zap.String("domain", domain))

	// If no items at all, still produce a valid JSON (empty array)
	// so the homepage shows gracefully
	if items == nil {
		items = []newsJSONItem{}
	}

	// -----------------------------------------------------------------------
	// 5. Check if an insights listing page exists (for the "More" link)
	// -----------------------------------------------------------------------
	var insightsURL string
	_ = params.DB.QueryRowContext(ctx, `
		SELECT url FROM pages
		WHERE site_id = $1 AND page_type = 'news-index' AND status = 'active'
		LIMIT 1
	`, siteID).Scan(&insightsURL)

	insightsLabel := ""
	if insightsURL != "" {
		insightsLabel = "More insights →"
	}

	// -----------------------------------------------------------------------
	// 6. Build JSON output
	// -----------------------------------------------------------------------
	output := newsJSONOutput{
		Headline:      headline,
		Subheadline:   subheadline,
		Items:         items,
		InsightsURL:   insightsURL,
		InsightsLabel: insightsLabel,
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
		ItemCount:     len(items),
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal news JSON: %w", err)
	}

	// -----------------------------------------------------------------------
	// 7. Return files map for git_commit step
	// -----------------------------------------------------------------------
	filesMap := map[string]interface{}{
		"data/latest-news.json": string(jsonBytes),
	}

	logger.Info("RenderNewsSectionAction: JSON produced",
		zap.Int("item_count", len(items)),
		zap.Int("json_bytes", len(jsonBytes)),
		zap.String("domain", domain))

	return map[string]interface{}{
		"files":      filesMap,
		"domain":     domain,
		"item_count": len(items),
		"headline":   headline,
		"file_path":  "data/latest-news.json",
		"rendered":   true,
	}, nil
}

// truncateNewsSummary truncates at a word boundary.
func truncateNewsSummary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	idx := strings.LastIndex(s[:maxLen], " ")
	if idx < maxLen/2 {
		idx = maxLen
	}
	return s[:idx] + "..."
}

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
